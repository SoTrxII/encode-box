// Package filtergraph :: Define ffmpeg filters as an oriented graph-like structure to be later used in
// ffmpeg filter_complex option
package filtergraph

import (
	"math/rand"
)

// Node A single node of an FFMPEG filter Tree
type Node struct {
	// Unique Node Id
	name string
	// All filter to be applied before this one can be compiled
	children []Filter
}
type Filter interface {
	// Build Resolve the graph into a string usable in FFMPEG -filter_complex option
	Build() string
	// Return a unique id for this filter
	Id() string
}

// Random string generation
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
)

// Return a random string containing n characters
func randString(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}
