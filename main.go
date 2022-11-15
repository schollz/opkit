package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/schollz/logger"
	"github.com/schollz/postsolarpunk/src/aubio"
	"github.com/schollz/postsolarpunk/src/sox"
)

var flagConvert, flagDebug bool
var flagKit string
var flagTune float64

func init() {
	flag.BoolVar(&flagDebug, "debug", false, "debug mode")
	flag.BoolVar(&flagConvert, "convert", false, "convert all the samples")
	flag.StringVar(&flagKit, "kit", "", "make kit (Kick, Snare, Hat, Bass)")
	flag.Float64Var(&flagTune, "tune", 30, "midi note to tune kit")
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
}

type Sample struct {
	Path     string
	Duration float64
	Midi     float64
}

func makeKit(kind string, note float64) (err error) {
	log.Infof("making kit from '%s' tuned to %f", kind, note)
	samples := make([]Sample, 1000)
	i := 0
	err = filepath.Walk("psp",
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(p) == ".wav" && strings.Contains(p, kind) {
				_, fname := filepath.Split(p)
				foo := strings.Split(fname, "_")
				if len(foo) > 1 {
					samples[i] = Sample{
						Path:     p,
						Duration: sox.MustFloat(sox.Length(p)),
						Midi:     sox.MustFloat(strconv.ParseFloat(foo[1], 64)),
					}
					i++
				}
			}
			return nil
		})
	if err != nil {
		log.Error(err)
	}

	samples = samples[:i]
	log.Infof("found %d samples for '%s'", len(samples), kind)
	log.Tracef("samples: %+v", samples[0])
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
