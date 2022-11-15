package aubio

import (
	"strings"
	"testing"

	log "github.com/schollz/logger"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	log.SetLevel("trace")
	stdout, stderr, err := run("aubio", "--help")
	assert.Nil(t, err)
	assert.True(t, strings.Contains(stdout, "usage"))
	assert.Empty(t, stderr)
}

func TestPitch(t *testing.T) {
	log.SetLevel("trace")
	midi, err := Pitch("40hz.wav")
	assert.Nil(t, err)
	assert.Equal(t, 27.5, midi)
	midi, err = Pitch("c4.flac")
	assert.Nil(t, err)
	assert.Equal(t, 60.0, midi)

}
