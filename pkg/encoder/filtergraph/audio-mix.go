package filtergraph

import (
	"fmt"
	"strings"
)

// AudioMixFilter mixing two audio tracks, with a modulated volume
// Documentation : https://ffmpeg.org/ffmpeg-filters.html#amix
type AudioMixFilter struct {
	Node
	// Relative volume of [main, side] channels
	weights [2]float32
	// Main channel index to find in both weights and children
	mainChannelIndex uint8
	// Side channel index to find in both weights and children
	sideChannelIndex uint8
}

func NewAudioMixFilter(main Filter, side Filter, weights [2]float32) *AudioMixFilter {
	return &AudioMixFilter{
		Node: Node{
			name:     fmt.Sprintf("mixed_%s", randString(5)),
			children: []Filter{main, side},
		},
		weights:          weights,
		mainChannelIndex: 0,
		sideChannelIndex: 1,
	}
}

func (amf *AudioMixFilter) Build() string {
	// Expected format :
	// nolint:lll  "[1]asplit=2[sc][v1];[r1][sc]sidechaincompress=threshold=0.05:ratio=5:level_sc=0.8[bg];[bg][v1]amix=weights=0.2 1[a3]"
	ss := strings.Builder{}
	// First let the children Build themselves
	for _, c := range amf.children {
		ss.WriteString(c.Build())
	}
	var (
		SideChannelModulate  = fmt.Sprintf("scm_%s", amf.Id())
		SideChannelOriginal  = fmt.Sprintf("sco_%s", amf.Id())
		ModulatedMainChannel = fmt.Sprintf("mmc_%s", amf.Id())
	)
	// Next, duplicate the side channel using asplit
	// One will be used to modulate the main channel volume, and the other mixed with the modulated
	// main channel
	// https://ffmpeg.org/ffmpeg-filters.html#split_002c-asplit
	// scm -> Side Channel Modulate, sco -> Side Channel Original
	ss.WriteString(
		fmt.Sprintf("[%s]asplit=2[%s][%s];",
			amf.children[amf.sideChannelIndex].Id(),
			SideChannelModulate,
			SideChannelOriginal))

	// Use the first duplicate of the side channel to modulate main channel volume
	ss.WriteString(
		fmt.Sprintf("[%s][%s]sidechaincompress=threshold=0.05:ratio=5:level_sc=0.8[%s];",
			amf.children[amf.mainChannelIndex].Id(),
			SideChannelModulate,
			ModulatedMainChannel))

	// Finally, mix the modulated main channel with the original main channel
	// 0.2 1
	ss.WriteString(
		fmt.Sprintf("[%s][%s]amix=weights=%.1f %.1f[%s];",
			ModulatedMainChannel,
			SideChannelOriginal,
			amf.weights[amf.mainChannelIndex],
			amf.weights[amf.sideChannelIndex],
			amf.Id()))

	return ss.String()

}
func (amf *AudioMixFilter) Id() string {
	return amf.name
}
