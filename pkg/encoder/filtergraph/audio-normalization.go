package filtergraph

import (
	"fmt"
	"strings"
)

// AudioNormalizationFilter : Normalize audio loudness
type AudioNormalizationFilter struct {
	Node
	// Mode of normalization
	mode AudioNormalizationMode
}

type AudioNormalizationMode uint8

const (
	// Default mode, slow but precise
	// Documentation : https://ffmpeg.org/ffmpeg-filters.html#loudnorm
	Loudnorm AudioNormalizationMode = iota
	// Faster than loudnorm, less precise
	// Documentation : https://ffmpeg.org/ffmpeg-filters.html#dynaudnorm
	Dynaudnorm
	// Made for speech normalization
	// Documentation : https://ffmpeg.org/ffmpeg-filters.html#speechnorm
	Speechnorm
)

func NewAudioNormalizationFilter(target Filter, mode AudioNormalizationMode) *AudioNormalizationFilter {
	return &AudioNormalizationFilter{
		Node{
			name:     fmt.Sprintf("norm_%s", randString(5)),
			children: []Filter{target},
		},
		mode}
}

func (amf *AudioNormalizationFilter) Build() string {
	// Expected format : [0:a]loudnorm=I=-16:TP=-1.5:LRA=11[norm]
	ss := strings.Builder{}
	// First let the children Build themselves
	for _, c := range amf.children {
		ss.WriteString(c.Build())
	}
	switch amf.mode {
	case Loudnorm:
		ss.WriteString(fmt.Sprintf("[%s]loudnorm=I=-16:TP=-1.5:LRA=11[%s];", amf.children[0].Id(), amf.Id()))
	case Dynaudnorm:
		ss.WriteString(fmt.Sprintf("[%s]dynaudnorm[%s];", amf.children[0].Id(), amf.Id()))
	case Speechnorm:
		ss.WriteString(fmt.Sprintf("[%s]speechnorm[%s];", amf.children[0].Id(), amf.Id()))
	}
	return ss.String()
}

func (amf *AudioNormalizationFilter) Id() string {
	return amf.name
}
