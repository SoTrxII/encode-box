package encoder

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

const (
	ResPath = "../../resources/test/"
)

var (
	TestVideo  = path.Join(ResPath, "/video.mp4")
	TestAudio1 = path.Join(ResPath, "/audio.m4a")
	TestAudio2 = path.Join(ResPath, "/audio.m4a")
	TestAudio3 = path.Join(ResPath, "/audio.m4a")
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
	// TODO
	t.Skip()
}

// Testing an encoding with an image used as a video track and multiples audio tracks
func TestEncodeBox_getAudiosImage_MultipleAudios(t *testing.T) {
	// TODO
	t.Skip()
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
