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
