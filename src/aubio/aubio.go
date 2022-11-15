package aubio

import (
	"bytes"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/montanaflynn/stats"
	log "github.com/schollz/logger"
)

// TempDir is where the temporary intermediate files are held
var TempDir = os.TempDir()

// TempPrefix is a unique indicator of the temporary files
var TempPrefix = "aubio"

// TempType is the type of file to be generated (should be "wav")
var TempType = "wav"

var GarbageCollection = false

func Tmpfile() string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(TempDir, TempPrefix+hex.EncodeToString(randBytes)+"."+TempType)
}

func init() {
	log.SetLevel("info")
	stdout, _, _ := run("aubio", "--help")
	if !strings.Contains(stdout, "usage") {
		panic("need to install aubio")
	}

}

func run(args ...string) (string, string, error) {
	log.Trace(strings.Join(args, " "))
	baseCmd := args[0]
	cmdArgs := args[1:]
	cmd := exec.Command(baseCmd, cmdArgs...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Errorf("%s -> '%s'", strings.Join(args, " "), err.Error())
		log.Error(outb.String())
		log.Error(errb.String())
	}
	return outb.String(), errb.String(), err
}

// MustString returns only the first argument of any function, as a string
func MustString(t ...interface{}) string {
	if len(t) > 0 {
		return t[0].(string)
	}
	return ""
}

// MustFloat returns only the first argument of any function, as a float
func MustFloat(t ...interface{}) float64 {
	if len(t) > 0 {
		return t[0].(float64)
	}
	return 0.0
}

func Pitch(fname string) (midi float64, err error) {
	vals := []float64{}
	rounded := []float64{}

	for _, algo := range []string{"schmitt", "mcomb"} {
		stdout, _, errP := run("aubio", "pitch", "-i", fname, "-m", algo, "-u", "midi")
		if errP != nil {
			log.Error(errP)
			continue
		}
		for _, line := range strings.Split(stdout, "\n") {
			foo := strings.Fields(line)
			if len(foo) < 2 {
				continue
			}
			val, _ := strconv.ParseFloat(foo[1], 64)
			if val < 0.1 {
				continue
			}
			vals = append(vals, val)
			rounded = append(rounded, math.Round(val))
		}
	}
	mode, _ := stats.Mode(rounded)
	toavg := []float64{}
	for _, val := range vals {
		if val > mode[0]-2 && val < mode[0]+2 {
			toavg = append(toavg, val)
		}
	}
	midi, _ = stats.Mean(toavg)
	midi, _ = stats.Round(midi, 1)
	return
}
