package filtergraph

import (
	"fmt"
	"strings"
)

type AudioVolumeFilter struct {
	Node
	// From 0 to 1, the volume to apply to the audio
	targetVolume float32
}

func NewAudioVolumeFilter(target Filter, targetVolume float32) *AudioVolumeFilter {
	return &AudioVolumeFilter{
		Node{
			name:     fmt.Sprintf("vol_%s", randString(5)),
			children: []Filter{target},
		},
		targetVolume}
}

func (avf *AudioVolumeFilter) Build() string {
	// Expected format : [0]volume=0.5[r1]
	ss := strings.Builder{}
	// First let the children Build themselves
	for _, c := range avf.children {
		ss.WriteString(c.Build())
	}
	ss.WriteString(fmt.Sprintf("[%s]volume=%.2f[%s];", avf.children[0].Id(), avf.targetVolume, avf.Id()))

	return ss.String()
}

func (avf *AudioVolumeFilter) Id() string {
	return avf.name
}
