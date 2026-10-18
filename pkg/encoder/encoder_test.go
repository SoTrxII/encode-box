package encoder

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
	"time"
)

const (
	ResDir = "../../resources"
	// Handy ffmpeg command with an infinite audio source and video source, this allows for a never-ending encoding.
	// Read in real time and capped at a minute all the same: on Windows, ffmpeg is a Chocolatey shim, and
	// killing it leaves the real ffmpeg running
	InfiniteCommand = "ffmpeg -re -f lavfi -i color=black:s=1280x720:r=25 -f lavfi -i anullsrc=r=48000:cl=stereo -map 0:v -map 1:a -c:v libx264rgb -t 60 %s"
	WithFilters     = "ffmpeg -i ../../resources/test/video.mp4 -i ../../resources/test/audio.m4a -filter_complex \"[1]loudnorm=I=-16:TP=-1.5:LRA=11[norm_spDdr];[norm_spDdr]aformat=sample_fmts=fltp:sample_rates=44100:channel_layouts=stereo[norm_yEuGz];[norm_yEuGz]asplit=2[scm_mixed_eNulz][sco_mixed_eNulz];[0][scm_mixed_eNulz]sidechaincompress=threshold=0.05:ratio=5:level_sc=0.8[mmc_mixed_eNulz];[mmc_mixed_eNulz][sco_mixed_eNulz]amix=weights=0.2 1.0[mixed_eNulz]\" -map 0:v -map [mixed_eNulz] %s"
	InputNotExists  = "ffmpeg -i ./meh.mp4 -map 0:v -map 1:a -c:v libx264rgb %s"
)

// tempDir is removed after the test if it can be. On Windows, the ffmpeg a
// killed Chocolatey shim leaves behind still holds its output for a while
func tempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "enc-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func TestEncoder_StartInfinite(t *testing.T) {
	dir := tempDir(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//testImage := Path.Join(ResDir, "/test/test.jpg")
	enc := NewEncoder(&ctx, fmt.Sprintf(InfiniteCommand, path.Join(dir, "test.mp4")))
	defer cancel()
	go enc.Start()
	for {
		select {
		case p := <-enc.PChan:
			fmt.Printf("%+v\n", p)
		case e := <-enc.EChan:
			t.Fatal(e)
		case <-ctx.Done():
			fmt.Printf("All done \n")
			return
		}
	}
}

func TestEncoder_CancelStopsFFmpeg(t *testing.T) {
	ctx := context.Background()
	enc := NewEncoder(&ctx, fmt.Sprintf(InfiniteCommand, path.Join(tempDir(t), "test.mp4")))
	stopped := make(chan struct{})
	go func() {
		enc.Start()
		close(stopped)
	}()

	// Cancel once ffmpeg is encoding
	select {
	case <-enc.PChan:
	case e := <-enc.EChan:
		t.Fatal(e)
	case <-time.After(30 * time.Second):
		t.Fatal("ffmpeg never reported any progress")
	}
	enc.Cancel()

	for {
		select {
		case <-stopped:
			return
		case <-enc.PChan:
		case <-enc.EChan:
		case <-time.After(10 * time.Second):
			t.Fatal("ffmpeg kept encoding after the encoder was cancelled")
		}
	}
}

func TestEncoder_StartWithFilters(t *testing.T) {
	dir := tempDir(t)
	//testImage := Path.Join(ResDir, "/test/test.jpg")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	enc := NewEncoder(&ctx, fmt.Sprintf(WithFilters, path.Join(dir, "test.mp4")))
	defer cancel()
	go enc.Start()
	for {
		select {
		case p := <-enc.PChan:
			fmt.Printf("%+v\n", p)
		case e := <-enc.EChan:
			t.Fatal(e)
		case <-ctx.Done():
			fmt.Printf("All done \n")
			return
		}
	}
}

func TestEncoder_StartError(t *testing.T) {
	dir := tempDir(t)
	//testImage := Path.Join(ResDir, "/test/test.jpg")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	enc := NewEncoder(&ctx, fmt.Sprintf(InputNotExists, path.Join(dir, "test.mp4")))
	defer cancel()
	go enc.Start()
	errorTriggered := false
	for {
		select {
		case p := <-enc.PChan:
			fmt.Printf("%+v\n", p)
		case e := <-enc.EChan:
			assert.ErrorContains(t, e, "No such file or directory")
			errorTriggered = true
		case <-ctx.Done():
			if !errorTriggered {
				t.Fatal(fmt.Sprintf("Expected error but not triggered"))
			}
			fmt.Printf("All done \n")
			return
		}
	}
}
