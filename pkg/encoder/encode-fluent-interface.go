package encoder

import (
	"context"
	"encode-box/pkg/encoder/filtergraph"
	"fmt"
	"strings"
)

type Builder struct {
	// All "-i" inputs
	inputs []*FileInput
	// All general options on the input
	inputOptions []string
	// All options on the output file
	outputOptions []string
	// Output file name
	output string
	// Root of the filter graph to be used
	filterGraph filtergraph.Filter
}

// AddInput Add a new input to the encoder
func (eb *Builder) AddInput(input *FileInput) *Builder {
	eb.inputs = append(eb.inputs, input)
	return eb
}

// AddOutputOption Add a new output option to the encoder
func (eb *Builder) AddOutputOption(opt string) *Builder {
	eb.outputOptions = append(eb.outputOptions, opt)
	return eb
}

// SetOutput Set the result file Path
func (eb *Builder) SetOutput(path string) *Builder {
	eb.output = path
	return eb
}

// SetFilterGraph Set the complex filters to be used
func (eb *Builder) SetFilterGraph(graph filtergraph.Filter) *Builder {
	eb.filterGraph = graph
	return eb
}

// Collapses the whole builder into a command-line string
func (eb *Builder) getFFmpegCmd() string {
	// ffmpeg (-i [inputs])* [inputOptions] [filters] [outputOptions] [output_name]
	ss := strings.Builder{}
	ss.WriteString("ffmpeg")

	// Inputs
	for _, input := range eb.inputs {
		ss.WriteString(fmt.Sprintf(" %s", input.String()))
	}

	// Filters
	if eb.filterGraph != nil {
		fGraph := eb.filterGraph.Build()
		// Remove the last ";" in the last step of the filtergraph. That's how FFmpeg knows it's complete
		ss.WriteString(fmt.Sprintf(` -filter_complex "%s"`, strings.TrimSuffix(fGraph, ";")))
	}

	// Output options
	for _, outputOpt := range eb.outputOptions {
		ss.WriteString(fmt.Sprintf(" %s", outputOpt))
	}

	// Output name
	ss.WriteString(fmt.Sprintf(` %s`, eb.output))

	return ss.String()
}

// Build Return a new initialized encoder ready to be started
func (eb *Builder) Build(ctx *context.Context) (*Encoder, error) {
	if len(eb.inputs) == 0 {
		return nil, fmt.Errorf("no inputs specified")
	}
	if eb.output == "" {
		return nil, fmt.Errorf("no output file Path specified")
	}
	return NewEncoder(ctx, eb.getFFmpegCmd()), nil
}

// FileInput Any valid -i input
type FileInput struct {
	// Path to file in the filesystem
	Path string
	// Format of the Path to input
	Format string
	// Specific options to be applied to this input
	Options []string
}

// Convert the input into a FFMPEG compatible cmd
func (ei *FileInput) String() string {
	ss := strings.Builder{}

	// Options
	if len(ei.Options) != 0 {
		ss.WriteString(fmt.Sprintf("%s ", strings.Join(ei.Options, " ")))
	}
	// Only specify Format if explicitly specified
	if ei.Format != "" {
		ss.WriteString(fmt.Sprintf("-f %s ", ei.Format))
	}
	ss.WriteString(fmt.Sprintf(`-i %s`, ei.Path))
	return ss.String()
}
