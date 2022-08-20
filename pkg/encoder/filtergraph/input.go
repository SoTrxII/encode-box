package filtergraph

// InputFilter a NO-OP filter representing an -i FFMPEG input
type InputFilter struct {
	Node
}

func (i *InputFilter) Build() string {
	// An input does nothing
	return ""
}

func (i *InputFilter) Id() string {
	return i.name
}

func NewInput(name string) *InputFilter {
	return &InputFilter{Node{
		// We have to identify an input by its index, named input won't work with -i FFMPEG option
		name: name,
		// An input cannot have children, it must be a leaf node
		children: nil,
	}}
}
