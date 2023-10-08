//go:build integration
// +build integration

package encode_box

import (
	"context"
	object_storage "encode-box/pkg/object-storage"
	"encoding/base64"
	"fmt"
	"github.com/dapr/go-sdk/client"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

const (
	ObjStoreComponent = "object-store"
	ResDir            = "../../resources/test"
	b64               = false
)

// These are integration test, using all real components
// Dapr and the backend storage should be booted up for this to work

func SetupInt(t *testing.T) (string, *EncodeBox) {
	ctx := context.Background()

	daprClient, err := client.NewClientWithPort("50010")
	if err != nil {
		t.Fatal(err)
	}
	downloader := object_storage.NewObjectStorage(&ctx, daprClient, ObjStoreComponent, b64)
	if err != nil {
		t.Fatal(err)
	}
	err = cpyToStorage(&daprClient, filepath.Join(ResDir, "audio.m4a"), "audio", b64)
	if err != nil {
		t.Fatal(err)
	}
	err = cpyToStorage(&daprClient, filepath.Join(ResDir, "audio.m4a"), "audio2", b64)
	if err != nil {
		t.Fatal(err)
	}
	err = cpyToStorage(&daprClient, filepath.Join(ResDir, "audio.m4a"), "audio3", b64)
	if err != nil {
		t.Fatal(err)
	}
	err = cpyToStorage(&daprClient, filepath.Join(ResDir, "video.mp4"), "video", b64)
	if err != nil {
		t.Fatal(err)
	}
	err = cpyToStorage(&daprClient, filepath.Join(ResDir, "dialog.mp3"), "dialog", b64)
	if err != nil {
		t.Fatal(err)
	}

	dir, err := os.MkdirTemp("", "test-encode-int")
	if err != nil {
		t.Fatal(err)
	}
	eBox := NewEncodeBox(&ctx, downloader, &EncodeBoxOptions{})
	return dir, eBox
}

func TestNewEncodeBox_Int_Encode_SingleAudio(t *testing.T) {
	dir, eBox := SetupInt(t)
	request := EncodingRequest{
		VideoKey:   "video",
		AudiosKeys: []string{"dialog"},
		ImageKey:   "",
		Options:    EncodingOptions{},
	}
	go eBox.Encode(&request, filepath.Join(dir, "out.mp4"))

Loop:
	for {
		select {
		case e := <-eBox.EChan:
			t.Fatal(e)
		case p := <-eBox.PChan:
			fmt.Printf("%+v", p)
			assert.Equal(t, 22.0, p.TargetDuration.Seconds())
		case <-eBox.Ctx.Done():
			fmt.Println("All Done")
			break Loop
		}
	}
	fmt.Printf("Output dir : %s", dir)
}

func TestNewEncodeBox_Int_Encode_MultipleAudio(t *testing.T) {
	dir, eBox := SetupInt(t)
	request := EncodingRequest{
		VideoKey:   "video",
		AudiosKeys: []string{"audio", "audio2", "audio3"},
		ImageKey:   "",
		Options:    EncodingOptions{},
	}
	go eBox.Encode(&request, filepath.Join(dir, "out.mp4"))

Loop:
	for {
		select {
		case e := <-eBox.EChan:
			t.Fatal(e)
		case p := <-eBox.PChan:
			fmt.Printf("%+v", p)
		case <-eBox.Ctx.Done():
			fmt.Println("All Done")
			break Loop
		}
	}
	fmt.Printf("Output dir : %s", dir)
}

func TestNewEncodeBox_Int_Encode_MultipleAudio_LongerThanVideo(t *testing.T) {
	dir, eBox := SetupInt(t)
	request := EncodingRequest{
		VideoKey:   "video",
		AudiosKeys: []string{"dialog", "audio2", "audio3"},
		ImageKey:   "",
		Options:    EncodingOptions{},
	}
	go eBox.Encode(&request, filepath.Join(dir, "out.mp4"))

Loop:
	for {
		select {
		case e := <-eBox.EChan:
			t.Fatal(e)
		case p := <-eBox.PChan:
			assert.Equal(t, 22.0+3+3, p.TargetDuration.Seconds())
			fmt.Printf("%+v", p)
		case <-eBox.Ctx.Done():
			fmt.Println("All Done")
			break Loop
		}
	}
	fmt.Printf("Output dir : %s", dir)
}

func cpyToStorage(daprClient *client.Client, src string, keyName string, b64 bool) error {

	rawContent, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	b64Content := make([]byte, base64.StdEncoding.EncodedLen(len(rawContent)))
	base64.StdEncoding.Encode(b64Content, rawContent)

	content := rawContent
	if b64 {
		content = b64Content
	}

	_, err = (*daprClient).InvokeBinding(context.Background(), &client.InvokeBindingRequest{
		Name:      ObjStoreComponent,
		Operation: "create",
		Data:      content,
		Metadata: map[string]string{
			"key":      keyName,
			"fileName": keyName,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
