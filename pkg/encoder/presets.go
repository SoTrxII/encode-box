package encoder

import (
	"context"
	"encode-box/pkg/encoder/filtergraph"
	"fmt"
)

// GetAudiosVideoEnc Return an initialized encoder with a video and one/multiple audio
// If multiple audios are specified, they will be concatenated
// The resulting video will have the normalized audio overlaid over the video audio track
func GetAudiosVideoEnc(ctx *context.Context, videoPath string, audioPaths []string, output string) (*Encoder, error) {
	builder := Builder{}
	// video track
	builder.AddInput(&FileInput{Path: videoPath})
	videoTrack := filtergraph.NewInput("0")

	var graphRoot filtergraph.Filter

	// audio tracks
	if len(audioPaths) == 1 {
		// Only one audio track, NO-OP
		builder.AddInput(&FileInput{Path: audioPaths[0]})
		graphRoot = filtergraph.NewInput("1")
	} else {
		// If multiple tracks are specified, concat them
		var aFilterInput []filtergraph.Filter
		for i, aPath := range audioPaths {
			builder.AddInput(&FileInput{Path: aPath})
			// Add a new audio track input indexed by one (0 being the video track)
			aTrack := filtergraph.NewInput(fmt.Sprintf("%d", i+1))
			aFilterInput = append(aFilterInput, aTrack)
		}
		graphRoot = filtergraph.NewAudioConcatFilter(aFilterInput...)

	}
	// In any way, normalize...
	graphRoot = filtergraph.NewAudioNormalizationFilter(graphRoot, filtergraph.Speechnorm)

	// ... and resample the resulting audio
	graphRoot = filtergraph.NewAudioResampleFilter(graphRoot, filtergraph.K44)

	// Finally, mix the video track with the combined audio track
	graphRoot = filtergraph.NewAudioMixFilter(videoTrack, graphRoot, filtergraph.WithoutModulation, [2]float32{1, 0.2})

	// And assign the graph to the command
	builder.SetFilterGraph(graphRoot)

	// Map the output -> Take the video from the only video source and the audio from the normalized audio track
	builder.AddOutputOption("-map 0:v").AddOutputOption(fmt.Sprintf("-map [%s]", graphRoot.Id()))

	// Set the output of the encoder
	builder.SetOutput(output)

	return builder.Build(ctx)
}

// GetAudiosVideoEnc Return an initialized encoder with a video and one/multiple audio
// If multiple audios are specified, they will be concatenated
// The resulting video will have the normalized audio overlaid over the video audio track
func GetAudiosImageEnc(ctx *context.Context, imagePath string, audioPaths []string, output string) (*Encoder, error) {
	builder := Builder{}
	// video track
	builder.AddInput(&FileInput{Path: imagePath, Options: []string{"-loop 1"}})
	var graphRoot filtergraph.Filter

	// audio tracks
	if len(audioPaths) == 1 {
		// Only one audio track, NO-OP
		builder.AddInput(&FileInput{Path: audioPaths[0]})
		graphRoot = filtergraph.NewInput("1")
	} else {
		// If multiple tracks are specified, concat them
		var aFilterInput []filtergraph.Filter
		for i, aPath := range audioPaths {
			builder.AddInput(&FileInput{Path: aPath})
			// Add a new audio track input indexed by one (0 being the video track)
			aTrack := filtergraph.NewInput(fmt.Sprintf("%d", i+1))
			aFilterInput = append(aFilterInput, aTrack)
		}
		graphRoot = filtergraph.NewAudioConcatFilter(aFilterInput...)

	}
	// In any way, normalize...
	graphRoot = filtergraph.NewAudioNormalizationFilter(graphRoot, filtergraph.Speechnorm)

	// ... and resample the resulting audio
	graphRoot = filtergraph.NewAudioResampleFilter(graphRoot, filtergraph.K44)

	// And assign the graph to the command
	builder.SetFilterGraph(graphRoot)

	// Add general options
	builder.
		// Set the pixel space
		AddOutputOption("-pix_fmt yuv420p").
		// Set the codec to be used
		AddOutputOption("-c:v libx264").
		// As we are using a static image, we can configure x264 to optimize for it
		AddOutputOption("-tune stillimage").
		// And end the video at the shortest input (the audio)
		AddOutputOption("-shortest")
	// Map the output -> Take the video from the only video source and the audio from the normalized audio track
	builder.AddOutputOption("-map 0:v").AddOutputOption(fmt.Sprintf("-map [%s]", graphRoot.Id()))

	// Set the output of the encoder
	builder.SetOutput(output)

	return builder.Build(ctx)
}

// GetAudiosOnlyEnc Return an initialized encoder with a black background and multiple audio tracks
// If multiple audios are specified, they will be concatenated
// The resulting video will have a black background over the video audio track.
//
// sidePictures is the number of picture tracks in the side file (the Roll20
// tab, when the GM recorded its picture too): they are shown over the black
// background, each where it was recorded
func GetAudiosOnlyEnc(ctx *context.Context, audioPaths []string, sideAudioPath string, sidePictures int, output string) (*Encoder, error) {
	builder := Builder{}
	// video track
	background := "color=black:s=1280x720:r=25"
	if sidePictures > 0 {
		background = fmt.Sprintf("color=black:s=%dx%d:r=%d", filtergraph.TimelineWidth, filtergraph.TimelineHeight, filtergraph.TimelineFps)
	}
	builder.AddInput(&FileInput{Path: background, Format: "lavfi"})
	var graphRoot filtergraph.Filter

	// audio tracks
	if len(audioPaths) == 1 {
		// Only one audio track, NO-OP
		builder.AddInput(&FileInput{Path: audioPaths[0]})
		graphRoot = filtergraph.NewInput("1")
	} else {
		// If multiple tracks are specified, concat them
		var aFilterInput []filtergraph.Filter
		for i, aPath := range audioPaths {
			builder.AddInput(&FileInput{Path: aPath})
			// Add a new audio track input indexed by one (0 being the video track)
			aTrack := filtergraph.NewInput(fmt.Sprintf("%d", i+1))
			aFilterInput = append(aFilterInput, aTrack)
		}
		graphRoot = filtergraph.NewAudioConcatFilter(aFilterInput...)

	}
	// In any way, normalize...
	graphRoot = filtergraph.NewAudioNormalizationFilter(graphRoot, filtergraph.Speechnorm)

	// If a side audio track is specified, add it to the mix
	if sideAudioPath != "" {
		var sideTrack filtergraph.Filter
		var sideOptions []string
		if sidePictures > 0 {
			sideOptions = []string{"-vn"}
		}
		builder.AddInput(&FileInput{Path: sideAudioPath, Options: sideOptions})
		sideTrack = filtergraph.NewInput(fmt.Sprintf("%d", len(audioPaths)+1))
		// Normalize it...
		sideTrack = filtergraph.NewAudioNormalizationFilter(sideTrack, filtergraph.Dynaudnorm)
		// And reduce its volume to properly stay in the background
		sideTrack = filtergraph.NewAudioVolumeFilter(sideTrack, 0.22)
		// ... and mix it with the main audio track
		graphRoot = filtergraph.NewAudioMixFilter(graphRoot, sideTrack, filtergraph.WithoutModulation, [2]float32{1, 0.85})
	}

	builder.
		AddOutputOption("-pix_fmt yuv420p").
		AddOutputOption("-c:v libx264").
		// And end the video at the shortest input (the audio)
		AddOutputOption("-shortest")
	if sidePictures == 0 {
		builder.SetFilterGraph(graphRoot)
		// As we are using a static image, we can configure x264 to optimize for it
		builder.AddOutputOption("-tune stillimage").AddOutputOption("-map 0:v")
	} else {
		// The side file again, once per picture track. Read once, ffmpeg decodes
		// and holds every frame of the file until the latest track's first one:
		// 35 GB for 20 min with a track starting at 13 min, against 0.84 GB this way
		var tracks []filtergraph.Filter
		for i := 0; i < sidePictures; i++ {
			builder.AddInput(&FileInput{Path: sideAudioPath, Options: []string{"-an"}})
			tracks = append(tracks, filtergraph.NewInput(fmt.Sprintf("%d:v:%d", len(audioPaths)+2+i, i)))
		}
		picture := filtergraph.NewVideoTimelineFilter(filtergraph.NewInput("0"), tracks...)
		builder.SetFilterGraph(graphRoot, picture)
		// A moving picture is the costly part: 0.35 CPU-second per second of
		// video on the VPS against 0.63 for the default (medium), so ~25 min of
		// 3-4 cores for a 4h session, next to whisper
		builder.AddOutputOption("-preset veryfast").AddOutputOption(fmt.Sprintf("-map [%s]", picture.Id()))
	}
	builder.AddOutputOption(fmt.Sprintf("-map [%s]", graphRoot.Id()))
	builder.SetOutput(output)
	return builder.Build(ctx)
}
