package filtergraph

import (
	"fmt"
	"strings"
)

// AudioConcatFilter Put one or more audio one after another
// /!\ The audio must use the same codec /!\
// Documentation : https://ffmpeg.org/ffmpeg-filters.html#concat
type AudioConcatFilter struct {
	Node
}

func NewAudioConcatFilter(inputs ...Filter) *AudioConcatFilter {
	return &AudioConcatFilter{Node{
		name:     fmt.Sprintf("concat_%s", randString(5)),
		children: inputs,
	}}
}

func (cfn *AudioConcatFilter) Build() string {
	// Expected format : children_build;[children_id_1][children_id_2]concat=n=2:v=0:a=1[input_id]
	ss := strings.Builder{}
	// First let the children Build themselves
	for _, c := range cfn.children {
		ss.WriteString(c.Build())
	}
	for _, c := range cfn.children {
		ss.WriteString(fmt.Sprintf("[%s]", c.Id()))
	}
	ss.WriteString(fmt.Sprintf("concat=n=%d:v=0:a=1[%s];", len(cfn.children), cfn.Id()))
	return ss.String()
}

func (cfn *AudioConcatFilter) Id() string {
	return cfn.name
}
