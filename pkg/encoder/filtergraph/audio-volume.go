package filtergraph

import "fmt"

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
	return fmt.Sprintf("[%s]volume=%.2f[%s];", avf.children[0].Id(), avf.targetVolume, avf.Id())
}

func (avf *AudioVolumeFilter) Id() string {
	return avf.name
}
