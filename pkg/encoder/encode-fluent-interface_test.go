package encoder

import (
	"context"
	"encode-box/pkg/encoder/filtergraph"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestEncoderBuilder_getCmd(t *testing.T) {
	builder := &Builder{}
	// Build a graph for normalizaing audio
	fInput := filtergraph.NewInput("0")
	graph := filtergraph.NewAudioNormalizationFilter(fInput, filtergraph.Dynaudnorm)

	const inputPath = "/tmp/test"
	const outputPath = "/tmp/testOut"
	outputOptions := []string{"-c:v libx264rgb", "-b:v 192k "}
	builder = builder.
		AddInput(&FileInput{Path: inputPath}).
		SetFilterGraph(graph)

	for _, out := range outputOptions {
		builder.AddOutputOption(out)
	}
	cmd := builder.
		SetOutput(outputPath).
		getFFmpegCmd()

	expectedCmd := fmt.Sprintf("ffmpeg -i %s -filter_complex \"%s\" %s %s %s", inputPath, strings.TrimSuffix(graph.Build(), ";"), outputOptions[0], outputOptions[1], outputPath)
	assert.Equal(t, expectedCmd, cmd)
}

func TestEncoderBuilder_BuildNoInput(t *testing.T) {
	builder := &Builder{}
	ctx := context.Background()
	_, err := builder.SetOutput("test").Build(&ctx)
	assert.Error(t, fmt.Errorf("no input file specified"), err)
}
func TestEncoderBuilder_BuildNoOutput(t *testing.T) {
	builder := &Builder{}
	ctx := context.Background()
	_, err := builder.AddInput(&FileInput{Path: "test"}).Build(&ctx)
	assert.Error(t, fmt.Errorf("no output file Path specified"), err)
}

// Test collapse input when all options are specified
func TestInput_StringWithoutFormat(t *testing.T) {
	const path = "color=black:s=1280x720:r=25"
	input1 := &FileInput{
		Path: path,
	}
	assert.Equal(t, fmt.Sprintf("-i %s", path), input1.String())
}

// Test collapse input when all options are specified
func TestInput_StringWithFormat(t *testing.T) {
	const path = "color=black:s=1280x720:r=25"
	const format = "lavfi"
	input1 := &FileInput{
		Path:   path,
		Format: format,
	}
	assert.Equal(t, fmt.Sprintf("-f %s -i %s", format, path), input1.String())
}

// Test collapse input when all options are specified
func TestInput_StringWithFormatAndOptions(t *testing.T) {
	const path = "color=black:s=1280x720:r=25"
	const format = "lavfi"
	input1 := &FileInput{
		Path:    path,
		Format:  format,
		Options: []string{"test"},
	}
	assert.Equal(t, fmt.Sprintf("test -f %s -i %s", format, path), input1.String())
}
