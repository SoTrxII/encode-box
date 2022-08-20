package filtergraph

import (
	"fmt"
	"strings"
)

type AudioResampleFilter struct {
	Node
	// Sampling rate to resample the audio into
	targetFormat Sampling
}

func NewAudioResampleFilter(target Filter, targetFormat Sampling) *AudioResampleFilter {
	return &AudioResampleFilter{Node{
		name:     fmt.Sprintf("norm_%s", randString(5)),
		children: []Filter{target},
	}, targetFormat}
}

func (arf *AudioResampleFilter) Build() string {
	// Expected format : [0]aformat=sample_fmts=fltp:sample_rates=44100:channel_layouts=stereo[r1]
	ss := strings.Builder{}
	// First let the children Build themselves
	for _, c := range arf.children {
		ss.WriteString(c.Build())
	}
	ss.WriteString(
		fmt.Sprintf("[%s]aformat=sample_fmts=fltp:sample_rates=%s:channel_layouts=stereo[%s];",
			arf.children[0].Id(),
			arf.targetFormat,
			arf.Id()))

	return ss.String()
}

func (arf *AudioResampleFilter) Id() string {
	return arf.name
}

// Sampling rates
type Sampling string

const (
	K44 Sampling = "44100"
	K48          = "48000"
)
