package console_parser

import "strings"

// A string ring buffer, contains the last size lines of FFMPEG log
type ringLogBuffer struct {
	// proper string buffer
	content []string
	// total array size
	size int
	// Index of the last line inserted
	currentIndex int
}

func NewRingLogBuffer(size int) *ringLogBuffer {
	return &ringLogBuffer{size: size, content: make([]string, size)}
}

func (rlb *ringLogBuffer) Push(str string) {
	rlb.content[rlb.currentIndex] = str
	rlb.currentIndex = (rlb.currentIndex + 1) % rlb.size
}

func (rlb *ringLogBuffer) String() string {
	return strings.Join(rlb.content, "\n")
}
