package filtergraph

import (
	"fmt"
	"strings"
)

// AudioMix mixing two audio tracks, with a modulated volume
// Documentation : https://ffmpeg.org/ffmpeg-filters.html#amix
type AudioMix struct {
	Node
	mode AudioMixMode
	// Relative volume of [main, side] channels
	weights [2]float32
	// Main channel index to find in both weights and children
	mainChannelIndex uint8
	// Side channel index to find in both weights and children
	sideChannelIndex uint8
}

type AudioMixMode uint8

const (
	WithoutModulation AudioMixMode = iota
	WithModulation
)

func NewAudioMixFilter(main Filter, side Filter, mode AudioMixMode, weights [2]float32) *AudioMix {
	return &AudioMix{
		Node: Node{
			name:     fmt.Sprintf("mixed_%s", randString(5)),
			children: []Filter{main, side},
		},
		mode:             mode,
		weights:          weights,
		mainChannelIndex: 0,
		sideChannelIndex: 1,
	}
}

func (amf *AudioMix) Build() string {
	switch amf.mode {
	case WithModulation:
		return amf.withModulation()
	case WithoutModulation:
		return amf.withoutModulation()
	default:
		return "invalid mix filter"
	}

}

// The main track alone decides how long the mix is and how loud it stays: the
// mix stops with it (duration=first), and the volumes are fixed (normalize=0)
// at the share each track has while both play.
//
// amix's defaults made the side track's length audible both ways. Measured
// with 60s of voices under a jukebox recording:
//   - a 75s jukebox (it is stopped a few seconds after Discord on /session end)
//     gave a 75s result, the video ending on 15s of music alone
//   - a 20s jukebox (the GM stopped recording early) made the voices 5.6 dB
//     louder from 20s on: amix shares the volume among the inputs still running
//
// With this both give 60s and voices within 0.1 dB from start to end.
// Padding the side track with apad instead stalled ffmpeg 8.0 on the video preset
func (amf *AudioMix) withoutModulation() string {
	// Expected format : [main][side]amix=weights=0.833 0.167:normalize=0:duration=first[res]
	ss := strings.Builder{}
	for _, c := range amf.children {
		ss.WriteString(c.Build())
	}

	main, side := amf.weights[amf.mainChannelIndex], amf.weights[amf.sideChannelIndex]
	ss.WriteString(
		fmt.Sprintf("[%s][%s]amix=weights=%.3f %.3f:normalize=0:duration=first[%s];",
			amf.children[amf.mainChannelIndex].Id(),
			amf.children[amf.sideChannelIndex].Id(),
			main/(main+side),
			side/(main+side),
			amf.Id(),
		),
	)
	return ss.String()
}

func (amf *AudioMix) withModulation() string {
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

func (amf *AudioMix) Id() string {
	return amf.name
}
