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
	"time"

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
	start := time.Now()
	defer func() {
		log.Tracef("Pitch took %+v", time.Since(start))
	}()

	vals := make([]float64, 10000)
	vi := 0
	rounded := make([]float64, 10000)
	ri := 0
	for _, algo := range []string{"schmitt", "mcomb"} {
		stdout, _, errP := run("aubio", "pitch", "-i", fname, "-m", algo, "-u", "midi")
		if errP != nil {
			log.Error(errP)
			continue
		}
		for _, line := range strings.Split(stdout, "\n") {
			if vi == len(vals) {
				break
			}
			foo := strings.Fields(line)
			if len(foo) < 2 {
				continue
			}
			val, _ := strconv.ParseFloat(foo[1], 64)
			if val < 0.1 {
				continue
			}

			vals[vi] = val
			vi++
			rounded[ri] = math.Round(val)
			ri++
		}
	}
	vals = vals[:vi]
	rounded = rounded[:ri]
	mode, _ := stats.Mode(rounded)
	toavg := make([]float64, len(vals))
	vi = 0
	for _, val := range vals {
		if vi == len(toavg) {
			break
		}
		if val > mode[0]-2 && val < mode[0]+2 {
			toavg[vi] = val
			vi++
		}
	}
	toavg = toavg[:vi]
	midi, _ = stats.Mean(toavg)
	midi, _ = stats.Round(midi, 1)
	return
}
