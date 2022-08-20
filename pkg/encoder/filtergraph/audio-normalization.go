package filtergraph

import (
	"fmt"
	"strings"
)

// AudioNormalizationFilter : Normalize audio loudness
// Documentation : https://ffmpeg.org/ffmpeg-filters.html#loudnorm
type AudioNormalizationFilter struct {
	Node
}

func NewAudioNormalizationFilter(target Filter) *AudioNormalizationFilter {
	return &AudioNormalizationFilter{Node{
		name:     fmt.Sprintf("norm_%s", randString(5)),
		children: []Filter{target},
	}}
}

func (amf *AudioNormalizationFilter) Build() string {
	// Expected format : [0:a]loudnorm=I=-16:TP=-1.5:LRA=11[norm]
	ss := strings.Builder{}
	// First let the children Build themselves
	for _, c := range amf.children {
		ss.WriteString(c.Build())
	}
	ss.WriteString(fmt.Sprintf("[%s]loudnorm=I=-16:TP=-1.5:LRA=11[%s];", amf.children[0].Id(), amf.Id()))
	return ss.String()
}

func (amf *AudioNormalizationFilter) Id() string {
	return amf.name
}
