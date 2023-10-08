package console_parser

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
	"time"
)

func TestConsoleParser_ParseProgress(t *testing.T) {
	const SampleLine = "frame= 2349 fps=335 q=28.0 size=       10kB time=00:01:31.60 bitrate=   0.0kbits/s speed=13.1x"
	e, err := parseProgress(SampleLine)
	assert.Nil(t, err)
	assert.Equal(t, e.Frames, int64(2349))
	assert.Equal(t, e.Fps, 335)
	assert.Equal(t, e.Quality, float32(28.0))
	assert.Equal(t, e.Size, int64(10))
	assert.Equal(t, time.Minute+31*time.Second, e.Time)
	assert.Equal(t, e.Speed, float32(13.1))
}

func TestConsoleParser_ParseProgressError(t *testing.T) {
	const SampleLine = "sdjsdjksjkd=sdjsdhdj=jjsj"
	_, err := parseProgress(SampleLine)
	assert.NotNil(t, err)
}

func TestConsoleParser_ParseSize(t *testing.T) {
	s, err := parseSize("10kb")
	assert.Nil(t, err)
	assert.Equal(t, int64(10), s)
	s, err = parseSize("10Mb")
	assert.Nil(t, err)
	assert.Equal(t, int64(10*1024), s)
	s, err = parseSize("10mb")
	assert.Nil(t, err)
	assert.Equal(t, int64(10*1024), s)
	s, err = parseSize("10Gb")
	assert.Nil(t, err)
	assert.Equal(t, int64(10*1024*1024), s)
	s, err = parseSize("10GB")
	assert.Nil(t, err)
	assert.Equal(t, int64(10*1024*1024), s)
	s, err = parseSize("dshsdhsd")
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), s)
}

func TestConsoleParser_ParseOutputGibberish(t *testing.T) {
	stringReader := strings.NewReader("shiny!")
	stringReadCloser := io.NopCloser(stringReader)
	ctx := context.Background()
	// Make a buffered channel to avoid being blocked while reading the channel length
	progressChan := make(chan *EncodingProgress, 1)
	errorChan := make(chan error)
	// This is random test so..
	ParseOutput(&ctx, &stringReadCloser, progressChan, errorChan)
	// There shouldn't be any progress emitted...
	assert.Equal(t, 0, len(progressChan))
	// Nor any error
	assert.Equal(t, 0, len(errorChan))
}

// Test the parsing of a valid FFMPEG progress event
func TestConsoleParser_ParseOutputProgressLine(t *testing.T) {
	stringReader := strings.NewReader("frame=   85 fps=0.0 q=28.0 size=       0kB time=00:00:01.04 bitrate=   0.4kbits/s speed=   2x    ")
	stringReadCloser := io.NopCloser(stringReader)
	ctx := context.Background()
	// Make a buffered channel to avoid being blocked while reading the channel length
	progressChan := make(chan *EncodingProgress, 1)
	errorChan := make(chan error)
	// This is a valid progress line so...
	ParseOutput(&ctx, &stringReadCloser, progressChan, errorChan)
	// A progress object should have been emitted
	assert.Equal(t, 1, len(progressChan))
	// But no error
	assert.Equal(t, 0, len(errorChan))
}
