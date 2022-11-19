package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/schollz/logger"
	"github.com/schollz/opkit/src/aubio"
	"github.com/schollz/opkit/src/sox"
	"github.com/schollz/progressbar/v3"
	"github.com/schollz/teoperator/src/convert"
)

var flagConvert, flagDebug bool
var flagKit, flagFilenameOut, flagFolderConvert, flagMix1, flagMix2, flagDurations string
var flagTune, flagMinLength, flagMaxLength float64

func init() {
	flag.BoolVar(&flagDebug, "debug", false, "debug mode")
	flag.StringVar(&flagDurations, "durations", "", "create cache of durations of folder")
	flag.StringVar(&flagKit, "kit", "", "make kit (Kick, Snare, Hat, Bass)")
	flag.StringVar(&flagFilenameOut, "out", "out.aif", "output file name or folder to convert to")
	flag.StringVar(&flagFolderConvert, "convert", "", "input folder to convert")
	flag.StringVar(&flagMix1, "mix1", "", "input folder to mix")
	flag.StringVar(&flagMix2, "mix2", "", "input folder to mix")
	flag.Float64Var(&flagTune, "tune", -1, "midi note to tune kit")
	flag.Float64Var(&flagMinLength, "min", 0, "min length for kit")
	flag.Float64Var(&flagMaxLength, "max", 2, " max length for kit")
}

func main() {
	flag.Parse()
	log.SetLevel("info")
	if flagDebug {
		log.SetLevel("trace")
	}
	var err error
	if flagDurations != "" {
		createDurations()
	}
	if flagFolderConvert != "" {
		convertAllFiles()
	}
	if flagMix1 != "" && flagMix2 != "" {
		makeMix()
	}
	if flagKit != "" {
		err = makeKit(flagKit, flagTune)
		if err != nil {
			log.Error(err)
			return
		}
	}
	fmt.Println(closestNote(33, 60))
	fmt.Println(repitchedLength(33, 60, 1))
}

func makeMix() (err error) {
	parts := strings.Split(flagMix1, "/")
	durations := make(map[string]float64)
	b, err := ioutil.ReadFile(path.Join(parts[0], "cache_durations.json"))
	if err != nil {
		log.Error(err)
		return
	}
	err = json.Unmarshal(b, &durations)
	if err != nil {
		log.Error(err)
		return
	}
	fileList1 := make([]string, len(durations))
	i := 0
	for k := range durations {
		if strings.Contains(k, flagMix1) {
			fileList1[i] = k
			i++
		}
	}
	fileList1 = fileList1[:i]

	fileList2 := make([]string, len(durations))
	i = 0
	for k := range durations {
		if strings.Contains(k, flagMix2) {
			fileList2[i] = k
			i++
		}
	}
	fileList2 = fileList2[:i]
	fmt.Println(fileList2[0])

	os.MkdirAll(flagFilenameOut, os.ModePerm)

	rand.Seed(time.Now().UTC().UnixNano())
	bar := progressbar.Default(500)
	for i := 0; i < 500; i++ {
		bar.Add(1)
		f1 := fileList1[rand.Intn(len(fileList1))]
		f2 := fileList2[rand.Intn(len(fileList1))]
		sox.Mixer(f1, f2, path.Join(flagFilenameOut, fmt.Sprintf("mix_%d.wav", i)))
	}
	return
}

type Sample struct {
	Path     string
	Duration float64
	Midi     float64
	Ratio    float64
}

func makeKit(kind string, note float64) (err error) {
	defer func() {
		sox.Clean()
	}()
	durations := make(map[string]float64)
	b, err := ioutil.ReadFile("psp2/cache_durations.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &durations)
	if err != nil {
		return
	}
	log.Infof("making kit from '%s' tuned to %f", kind, note)
	samples := make([]Sample, 1000)
	i := 0
	for p, duration := range durations {
		if !strings.Contains(p, kind) || duration < flagMinLength || duration > flagMaxLength {
			continue
		}
		_, fname := filepath.Split(p)
		foo := strings.Split(fname, "_")
		if len(foo) > 1 {
			midi := sox.MustFloat(strconv.ParseFloat(foo[1], 64))
			ratio := 1.0
			if flagTune > -1 {
				duration = repitchedLength(midi, flagTune, duration)
				ratio = closestRatio(midi, flagTune)
				midi = closestNote(midi, flagTune)
			}
			samples[i] = Sample{
				Path:     p,
				Duration: duration,
				Midi:     midi,
				Ratio:    ratio,
			}
			i++
		}
	}
	if i == 0 {
		err = fmt.Errorf("no files found for '%s'", kind)
		return
	}
	samples = samples[:i]

	sort.Slice(samples, func(i, j int) bool {
		return samples[i].Duration > samples[j].Duration
	})
	log.Infof("found %d samples for '%s'", len(samples), kind)
	log.Tracef("samples: %+v", samples[0])
	i = 0
	j := rand.Intn(12)
	s := make([]Sample, 24)
	duration := 0.0
	rand.Seed(time.Now().UTC().UnixNano())
	for tries := 0; tries < 40; tries++ {
		j += len(samples)/12 + rand.Intn(6)
		j = j % len(samples)
		duration += samples[j].Duration
		if duration > 11.5 {
			duration -= samples[j].Duration
			continue
		}
		s[i] = samples[j]
		i++
		if i > len(s)-1 {
			break
		}
	}
	s = s[:i]
	sort.Slice(s, func(i, j int) bool {
		return s[i].Duration > s[j].Duration
	})
	log.Infof("found %d samples, total duration: %2.1f", len(s), duration)
	fnames := make([]string, len(s))
	for i, v := range s {
		fnames[i], err = sox.Speed(v.Path, v.Ratio)
		d, _ := sox.Length(fnames[i])
		log.Debugf("%s: %2.3f %2.3f", fnames[i], v.Duration, d)
		if err != nil {
			log.Error(err)
		}
	}
	finalName, err := convert.ToDrum2(fnames, 0)
	if err != nil {
		log.Error(err)
		return
	}
	if filepath.Ext(flagFilenameOut) != ".aif" {
		flagFilenameOut += ".aif"
	}
	err = os.Rename(finalName, flagFilenameOut)
	return
}

func repitchedLength(note1, note2, duration float64) float64 {
	cr := closestRatio(note1, note2)
	return duration / cr
}

func closestRatio(note1, note2 float64) (ratio float64) {
	closest := closestNote(note1, note2)
	return noteRatio(closest, note1)
}

// closestNote finds the closest note between note1 and any octave of note2
func closestNote(note1, note2 float64) (closest float64) {
	note0 := math.Mod(note2, 12)
	diff := 10000.0
	for i := 0.0; i < 12; i++ {
		note := note0 + (12.0 * i)
		if math.Abs(note-note1) < diff {
			closest = note
			diff = math.Abs(note - note1)
		}
	}
	return
}

func noteRatio(note1, note2 float64) (ratio float64) {
	ratio = math.Pow(2, (note1-69)/12) / math.Pow(2, (note2-69)/12)
	return
}

func createDurations() (err error) {
	m := make(map[string]float64)
	err = filepath.Walk(flagDurations,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(p) == ".wav" {
				m[p] = sox.MustFloat(sox.Length(p))
			}
			return nil
		})
	if err != nil {
		log.Error(err)
		return
	}
	b, err := json.MarshalIndent(m, " ", " ")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(path.Join(flagDurations, "cache_durations.json"), b, 0644)
	return
}

func convertAllFiles() {
	os.MkdirAll(flagFilenameOut, os.ModePerm)

	files := make([]string, 10000)
	i := 0
	err := filepath.Walk(flagFolderConvert,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(p) == ".wav" {
				files[i] = p
				i++
			}
			return nil
		})
	if err != nil {
		log.Error(err)
	}
	files = files[:i]

	numJobs := len(files)
	type job struct {
		filename string
		it       int
	}
	type result struct {
		filename string
		err      error
	}

	jobs := make(chan job, numJobs)
	results := make(chan result, numJobs)
	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(jobs <-chan job, results chan<- result) {
			for j := range jobs {
				_ = j
				// step 3: specify the work for the worker
				var r result
				p := j.filename
				folder, filename := filepath.Split(p)
				r.filename = filename
				r.err = os.MkdirAll(path.Join(flagFilenameOut, folder), os.ModePerm)
				if r.err == nil {
					midi, _ := aubio.Pitch(p)
					midi_ref := 31.0 // G
					ratio := closestRatio(midi, midi_ref)
					midiClosest := closestNote(midi, midi_ref)
					nameNew := fmt.Sprintf("%04d_%2.1f_.wav", j.it, midiClosest)
					r.err = sox.DrumProcess(p, path.Join(flagFilenameOut, folder, nameNew), ratio)
				}
				results <- r
			}
		}(jobs, results)
	}

	// step 4: send out jobs
	for i := 0; i < numJobs; i++ {
		jobs <- job{files[i], i}
	}
	close(jobs)

	// step 5: do something with results
	bar := progressbar.Default(int64(numJobs))
	for i := 0; i < numJobs; i++ {
		bar.Add(1)
		r := <-results
		if r.err != nil {
			log.Errorf("%s: %s", r.filename, r.err)
		}
	}

}
