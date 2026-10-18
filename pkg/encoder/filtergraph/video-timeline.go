package filtergraph

import (
	"fmt"
	"strings"
)

// VideoTimelineFilter lays picture tracks over a background. Each track keeps
// its own timestamps, so it shows where it was recorded on the session's
// timeline, and the background shows wherever no track runs (before the GM
// started recording, or between two captures)
type VideoTimelineFilter struct {
	Node
}

func NewVideoTimelineFilter(background Filter, tracks ...Filter) *VideoTimelineFilter {
	return &VideoTimelineFilter{Node{
		name:     fmt.Sprintf("timeline_%s", randString(5)),
		children: append([]Filter{background}, tracks...),
	}}
}

func (vtf *VideoTimelineFilter) Build() string {
	// Expected format, for two tracks :
	// [2:v:0]fps=15,scale=...,setsar=1[t0];[0][t0]overlay=eof_action=pass[o0];[o0][t1]overlay=...[res];
	ss := strings.Builder{}
	for _, c := range vtf.children {
		ss.WriteString(c.Build())
	}
	under := vtf.children[0].Id()
	for i, track := range vtf.children[1:] {
		label := fmt.Sprintf("%s_t%d", vtf.Id(), i)
		// The background's rate and size: a GM's window can be any size, and
		// can change size mid-capture
		ss.WriteString(fmt.Sprintf("[%s]fps=%d,scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:-1:-1,setsar=1[%s];",
			track.Id(), TimelineFps, TimelineWidth, TimelineHeight, TimelineWidth, TimelineHeight, label))
		out := fmt.Sprintf("%s_o%d", vtf.Id(), i)
		if i == len(vtf.children)-2 {
			out = vtf.Id()
		}
		// eof_action=pass: the background again once the track ends
		ss.WriteString(fmt.Sprintf("[%s][%s]overlay=eof_action=pass[%s];", under, label, out))
		under = out
	}
	return ss.String()
}

func (vtf *VideoTimelineFilter) Id() string {
	return vtf.name
}

// Lullaby captures the tab at 720p and 15 fps: more would only be upscaled
const (
	TimelineFps    = 15
	TimelineWidth  = 1280
	TimelineHeight = 720
)
