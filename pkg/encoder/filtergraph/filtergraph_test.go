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
func TestAudioMixFilterWithModulation(t *testing.T) {
	a1 := NewInput("0")
	a2 := NewInput("1")
	mix := NewAudioMixFilter(a1, a2, WithModulation, [2]float32{1, 0.2})
	builtFilter := mix.Build()
	// This is a 3 steps pipeline (with a ";" at the end)
	assert.Equal(t, 4, len(strings.Split(builtFilter, ";")))
	// The final mixed name should only appear once
	assert.Equal(t, strings.Count(builtFilter, fmt.Sprintf("[%s]", mix.Id())), 1)
	// And both the main and side channel should only be used once
	assert.Equal(t, 1, strings.Count(builtFilter, "[0]"))
	assert.Equal(t, 1, strings.Count(builtFilter, "[1]"))
}

// Testing mix filter in isolation
func TestAudioMixFilterWithoutModulation(t *testing.T) {
	a1 := NewInput("0")
	a2 := NewInput("1")
	mix := NewAudioMixFilter(a1, a2, WithoutModulation, [2]float32{1, 0.2})
	builtFilter := mix.Build()
	// 1 and 0.2, as shares of the whole
	assert.Equal(t, fmt.Sprintf("[0][1]amix=weights=0.833 0.167:normalize=0:duration=first[%s];", mix.Id()), builtFilter)
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

func TestVolumeFilter(t *testing.T) {
	a1 := NewInput("0")
	vol := NewAudioVolumeFilter(a1, 1.5)
	builtFilter := vol.Build()
	assert.Equal(t, fmt.Sprintf("[0]volume=1.50[%s];", vol.Id()), builtFilter)
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
	mix := NewAudioMixFilter(concat, a3, WithoutModulation, [2]float32{0.2, 1})
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

// Each track is fitted to the background, then laid over what is under it
func TestVideoTimelineFilter(t *testing.T) {
	timeline := NewVideoTimelineFilter(NewInput("0"), NewInput("3:v:0"), NewInput("4:v:1"))
	id := timeline.Id()
	fit := "fps=15,scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:-1:-1,setsar=1"
	assert.Equal(t, fmt.Sprintf("[3:v:0]%s[%s_t0];[0][%s_t0]overlay=eof_action=pass[%s_o0];"+
		"[4:v:1]%s[%s_t1];[%s_o0][%s_t1]overlay=eof_action=pass[%s];", fit, id, id, id, fit, id, id, id, id), timeline.Build())
}
