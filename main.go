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
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/schollz/logger"
	"github.com/schollz/postsolarpunk/src/aubio"
	"github.com/schollz/postsolarpunk/src/sox"
	"github.com/schollz/teoperator/src/convert"
)

var flagConvert, flagDebug bool
var flagKit string
var flagTune float64

func init() {
	flag.BoolVar(&flagDebug, "debug", false, "debug mode")
	flag.BoolVar(&flagConvert, "convert", false, "convert all the samples")
	flag.StringVar(&flagKit, "kit", "", "make kit (Kick, Snare, Hat, Bass)")
	flag.Float64Var(&flagTune, "tune", -1, "midi note to tune kit")
}

func main() {
	flag.Parse()
	log.SetLevel("info")
	if flagDebug {
		log.SetLevel("trace")
	}
	var err error
	if flagConvert {
		convertAllFiles()

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
	b, err := ioutil.ReadFile("cache_durations.json")
	if err != nil {
		err = createDurations()
		if err != nil {
			log.Error(err)
			return
		}
	}
	b, _ = ioutil.ReadFile("cache_durations.json")
	err = json.Unmarshal(b, &durations)
	if err != nil {
		return
	}
	log.Infof("making kit from '%s' tuned to %f", kind, note)
	samples := make([]Sample, 1000)
	i := 0
	for p, duration := range durations {
		if !strings.Contains(p, kind) || duration < 0.05 {
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
	j := rand.Intn(24)
	s := make([]Sample, 24)
	duration := 0.0
	rand.Seed(time.Now().UTC().UnixNano())
	for tries := 0; tries < 40; tries++ {

		j += len(samples) / 12
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
	fmt.Println(convert.ToDrum2(fnames, 0))
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
	err = filepath.Walk("psp",
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
	err = ioutil.WriteFile("cache_durations.json", b, 0644)
	return
}
func convertAllFiles() {
	pathNew := "psp"
	it := 0
	err := filepath.Walk("pulsar-23 postsolarpunk pack",
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(p) == ".wav" {
				if sox.MustFloat(sox.Length(p)) > 4 {
					return nil
				}
				folder, _ := filepath.Split(p)
				log.Trace(folder, p)
				err = os.MkdirAll(path.Join(pathNew, folder), os.ModePerm)
				if err != nil {
					log.Error(err)
				}
				midi, _ := aubio.Pitch(p)
				log.Trace(midi)
				nameNew := fmt.Sprintf("%04d_%2.1f_.wav", it, midi)
				it++
				sox.SilenceTrimEndMono(p, path.Join(pathNew, folder, nameNew))
			}
			return nil
		})
	if err != nil {
		log.Error(err)
	}
}
