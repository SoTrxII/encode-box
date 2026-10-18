package encoder

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"testing"
)

const (
	ResPath = "../../resources/test/"
)

var (
	TestVideo      = path.Join(ResPath, "/video.mp4")
	TestImage      = path.Join(ResPath, "/test.jpg")
	TestAudio1     = path.Join(ResPath, "/audio.m4a")
	TestAudio2     = path.Join(ResPath, "/audio.m4a")
	TestAudio3     = path.Join(ResPath, "/audio.m4a")
	TestBackground = path.Join(ResPath, "/background-music.mp3")
	TestDialog     = path.Join(ResPath, "/dialog.mp3")
)

// Creates a temp directory for test to run into
func Setup(t *testing.T) (string, string) {
	dir, err := os.MkdirTemp("", "test-enc")
	out := path.Join(dir, "out.mp4")
	if err != nil {
		t.Fatal(err)
	}
	return dir, out
}

// Testing an encoding with a single video track and a single audio track
func TestEncodeBox_GetAudiosVideo_SingleAudio(t *testing.T) {
	dir, out := Setup(t)
	defer Teardown(t, dir)
	ctx := context.Background()
	enc, err := GetAudiosVideoEnc(&ctx, TestVideo, []string{TestAudio1}, out)
	assert.Nil(t, err)
	err = runEncoding(t, enc)
	assert.Nil(t, err)
}

// Testing an encoding with a single video track and multiple audio track
func TestEncodeBox_GetAudiosVideo_MultipleAudios(t *testing.T) {
	dir, out := Setup(t)
	defer Teardown(t, dir)
	ctx := context.Background()
	enc, err := GetAudiosVideoEnc(&ctx, TestVideo, []string{TestAudio1, TestAudio2, TestAudio3}, out)
	err = runEncoding(t, enc)
	assert.Nil(t, err)
}

// Testing an encoding with an image used as a video track and a single audio track
func TestEncodeBox_getAudiosImage_SingleAudio(t *testing.T) {
	dir, out := Setup(t)
	defer Teardown(t, dir)
	ctx := context.Background()
	enc, err := GetAudiosImageEnc(&ctx, TestImage, []string{TestAudio1}, out)
	err = runEncoding(t, enc)
	assert.Nil(t, err)
}

// Testing an encoding with an image used as a video track and multiples audio tracks
func TestEncodeBox_getAudiosImage_MultipleAudios(t *testing.T) {
	dir, out := Setup(t)
	defer Teardown(t, dir)
	ctx := context.Background()
	enc, err := GetAudiosImageEnc(&ctx, TestImage, []string{TestAudio1, TestAudio2, TestAudio3}, out)
	err = runEncoding(t, enc)
	assert.Nil(t, err)
}

// Testing an encoding with a placeholder used as a video track and a single audio track
func TestEncodeBox_getAudiosOnly_SingleAudio(t *testing.T) {
	dir, out := Setup(t)
	defer Teardown(t, dir)
	ctx := context.Background()
	enc, err := GetAudiosOnlyEnc(&ctx, []string{TestAudio1}, "", 0, out)
	err = runEncoding(t, enc)
	assert.Nil(t, err)
}

// Testing an encoding with a placeholder used as a video track and multiples audio tracks
func TestEncodeBox_getAudiosOnly_MultipleAudios(t *testing.T) {
	dir, out := Setup(t)
	defer Teardown(t, dir)
	ctx := context.Background()
	enc, err := GetAudiosOnlyEnc(&ctx, []string{TestAudio1, TestAudio2, TestAudio3}, "", 0, out)
	err = runEncoding(t, enc)
	assert.Nil(t, err)
}

func TestEncodeBox_getAudiosOnly_SingleAudio_SideChannel(t *testing.T) {
	dir, out := Setup(t)
	defer Teardown(t, dir)
	ctx := context.Background()
	enc, err := GetAudiosOnlyEnc(&ctx, []string{TestDialog}, TestBackground, 0, out)
	err = runEncoding(t, enc)
	assert.Nil(t, err)
}

// The voices decide how long the audio is, whatever the length of the jukebox
// recording : it is stopped a few seconds after Discord on /session end (longer),
// or earlier when the GM stops recording before the end (shorter)
func TestEncodeBox_getAudiosOnly_SideChannelNeverChangesTheLength(t *testing.T) {
	// The audio track, not the file : the picture always runs ~3s past the
	// audio, ffmpeg's -shortest stopping late (25.92s of video for 22.81s of
	// audio, with or without a side track)
	audioSeconds := func(file string) float64 { return streamSeconds(t, file, "a:0") }
	voices := audioSeconds(TestDialog)
	for name, side := range map[string]string{"longer": TestBackground, "shorter": TestAudio1} {
		dir, out := Setup(t)
		ctx := context.Background()
		enc, err := GetAudiosOnlyEnc(&ctx, []string{TestDialog}, side, 0, out)
		assert.Nil(t, err)
		assert.Nil(t, runEncoding(t, enc))
		// Encoder padding and an audio frame of slack
		assert.InDelta(t, voices, audioSeconds(out), 0.1, "with a %s side track", name)
		Teardown(t, dir)
	}
}

// Testing an encoding with a placeholder used as a video track and multiples audio tracks
func TestEncodeBox_getAudiosOnly_MultipleAudios_SideChannel(t *testing.T) {
	dir, out := Setup(t)
	defer Teardown(t, dir)
	ctx := context.Background()
	enc, err := GetAudiosOnlyEnc(&ctx, []string{TestAudio1, TestAudio2, TestAudio3}, TestAudio2, 0, out)
	err = runEncoding(t, enc)
	assert.Nil(t, err)
}

// Deletes the created temp directory
func Teardown(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Logf("Unable to remove created temp dir %s", dir)
	}
}

// Start the specified encoder and waits fo it to finish running.
// Return an error if the encoder did not succeed, nil otherwise
func runEncoding(t *testing.T, enc *Encoder) error {
	cmd := enc.GetCommandLine()
	fmt.Println(cmd)
	go enc.Start()
	for {
		select {
		case err := <-enc.EChan:
			return err
		case p := <-enc.PChan:
			fmt.Printf("%+v", p)
		case <-enc.Ctx.Done():
			fmt.Println("All Done !")
			return nil
		}
	}
}

// The Roll20 tab's picture, as roll20-audio-sync hands it over: one picture
// track per capture, at its place on the Discord timeline, next to the jukebox
// sound. Each shows where it was recorded, the black background elsewhere,
// and the voices still decide the length
func TestEncodeBox_getAudiosOnly_SidePictures(t *testing.T) {
	dir, out := Setup(t)
	defer Teardown(t, dir)
	voices := 22.81 // TestDialog
	// A capture from 3s to 7s, and one from 18s, resized, running past the voices
	side := path.Join(dir, "capture.mkv")
	assert.Nil(t, exec.Command("ffmpeg", "-v", "error",
		"-f", "lavfi", "-i", "sine=d=24",
		"-itsoffset", "3", "-f", "lavfi", "-i", "testsrc=s=1280x720:r=15:d=4",
		"-itsoffset", "18", "-f", "lavfi", "-i", "testsrc=s=960x540:r=15:d=8",
		"-map", "0", "-map", "1", "-map", "2", "-c:v", "libx264", "-c:a", "libopus", side).Run())

	ctx := context.Background()
	enc, err := GetAudiosOnlyEnc(&ctx, []string{TestDialog}, side, 2, out)
	assert.Nil(t, err)
	assert.Nil(t, runEncoding(t, enc))

	assert.InDelta(t, voices, streamSeconds(t, out, "a:0"), 0.1)
	assert.InDelta(t, voices, streamSeconds(t, out, "v:0"), 0.5)

	// Mean luma of the frame at t: 16 is black
	luma := func(at float64) float64 {
		o, _ := exec.Command("ffmpeg", "-ss", fmt.Sprint(at), "-i", out, "-frames:v", "1",
			"-vf", "signalstats,metadata=print:key=lavfi.signalstats.YAVG", "-f", "null", "-").CombinedOutput()
		_, value, _ := strings.Cut(string(o), "YAVG=")
		y, _ := strconv.ParseFloat(strings.Fields(value)[0], 64)
		return y
	}
	for at, black := range map[float64]bool{1: true, 4: false, 6.5: false, 10: true, 19: false, 22: false} {
		if black {
			assert.InDelta(t, 16, luma(at), 1, "black at %.1fs", at)
		} else {
			assert.Greater(t, luma(at), 50.0, "picture at %.1fs", at)
		}
	}
}

// The duration of one stream of a file
func streamSeconds(t *testing.T, file, stream string) float64 {
	out, err := exec.Command("ffprobe", "-v", "error", "-select_streams", stream,
		"-show_entries", "stream=duration", "-of", "csv=p=0", file).Output()
	assert.Nil(t, err)
	d, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	assert.Nil(t, err)
	return d
}
