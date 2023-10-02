package filtergraph

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// Testing concat filter in isolation
func TestAudioConcatFilter(t *testing.T) {
	a1 := NewInput("0")
	a2 := NewInput("1")
	concat := NewAudioConcatFilter(a1, a2)
	builtFilter := concat.Build()
	assert.Equal(t, fmt.Sprintf("[0][1]concat=n=2:v=0:a=1[%s];", concat.Id()), builtFilter)
}

// Testing mix filter in isolation
func TestAudioMixFilter(t *testing.T) {
	a1 := NewInput("0")
	a2 := NewInput("1")
	mix := NewAudioMixFilter(a1, a2, [2]float32{0.2, 1})
	builtFilter := mix.Build()
	// This is a 3 steps pipeline (with a ";" at the end)
	assert.Equal(t, 4, len(strings.Split(builtFilter, ";")))
	// The final mixed name should only appear once
	assert.Equal(t, strings.Count(builtFilter, fmt.Sprintf("[%s]", mix.Id())), 1)
	// And both the main and side channel should only be used once
	assert.Equal(t, 1, strings.Count(builtFilter, "[0]"))
	assert.Equal(t, 1, strings.Count(builtFilter, "[1]"))
}

// Testing normalization filter in isolation
func TestNormalizationFilterLoudnorm(t *testing.T) {
	a1 := NewInput("0")
	norm := NewAudioNormalizationFilter(a1, Loudnorm)
	builtFilter := norm.Build()
	assert.Equal(t, fmt.Sprintf("[0]loudnorm=I=-16:TP=-1.5:LRA=11[%s];", norm.Id()), builtFilter)
}

func TestNormalizationFilterDynaudnorm(t *testing.T) {
	a1 := NewInput("0")
	norm := NewAudioNormalizationFilter(a1, Dynaudnorm)
	builtFilter := norm.Build()
	assert.Equal(t, fmt.Sprintf("[0]dynaudnorm[%s];", norm.Id()), builtFilter)
}

func TestNormalizationFilterSpeechnorm(t *testing.T) {
	a1 := NewInput("0")
	norm := NewAudioNormalizationFilter(a1, Speechnorm)
	builtFilter := norm.Build()
	assert.Equal(t, fmt.Sprintf("[0]speechnorm[%s];", norm.Id()), builtFilter)
}

// Testing resample filter in isolation
func TestResampleFilter(t *testing.T) {
	a1 := NewInput("0")
	norm := NewAudioResampleFilter(a1, K44)
	builtFilter := norm.Build()
	assert.Equal(t,
		fmt.Sprintf("[0]aformat=sample_fmts=fltp:sample_rates=%s:channel_layouts=stereo[%s];", "44100", norm.Id()),
		builtFilter)
}

// Testing concat and mix chained
func TestCompositeAudioConcatAndMix(t *testing.T) {
	a1 := NewInput("0")
	a2 := NewInput("1")
	// First, concat a1 and a2, the result will be our main channel
	concat := NewAudioConcatFilter(a1, a2)
	// Then, define a side channel
	a3 := NewInput("2")
	// And mix them together
	mix := NewAudioMixFilter(concat, a3, [2]float32{0.2, 1})
	builtFilter := mix.Build()
	fmt.Println(builtFilter)
	// Concat must be executed before Mix
	concatIndex := strings.Index(builtFilter, "concat")
	mixIndex := strings.Index(builtFilter, "mix")
	assert.Greater(t, mixIndex, concatIndex)
	// concat Id should appear twice. Once as a result for the concat filter, and one more time as an input for
	// the mix filter
	assert.Equal(t, 2, strings.Count(builtFilter, concat.Id()))
}
