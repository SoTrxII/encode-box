//go:build integration
// +build integration

package encode_box

import (
	"context"
	object_storage "encode-box/pkg/object-storage"
	"encoding/base64"
	"fmt"
	"github.com/dapr/go-sdk/client"
	"os"
	"path/filepath"
	"testing"
)

const (
	ObjStoreComponent = "object-store"
	ResDir            = "../../resources/test"
)

// These are integration test, using all real components
// Dapr and the backend storage should be booted up for this to work

func SetupInt(t *testing.T) (string, *EncodeBox[client.Client]) {
	ctx := context.Background()
	downloader, err := object_storage.NewDaprObjectStorage(&ctx, ObjStoreComponent)
	if err != nil {
		t.Fatal(err)
	}
	daprClient, err := client.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	err = copy(&daprClient, filepath.Join(ResDir, "audio.m4a"), "audio")
	if err != nil {
		t.Fatal(err)
	}
	err = copy(&daprClient, filepath.Join(ResDir, "audio.m4a"), "audio2")
	if err != nil {
		t.Fatal(err)
	}
	err = copy(&daprClient, filepath.Join(ResDir, "audio.m4a"), "audio3")
	if err != nil {
		t.Fatal(err)
	}
	err = copy(&daprClient, filepath.Join(ResDir, "video.mp4"), "video")
	if err != nil {
		t.Fatal(err)
	}

	dir, err := os.MkdirTemp("", "test-encode-int")
	if err != nil {
		t.Fatal(err)
	}
	eBox := NewEncodeBox[client.Client](&ctx, downloader)
	return dir, eBox
}

func TestNewEncodeBox_Int_Encode_SingleAudio(t *testing.T) {
	dir, eBox := SetupInt(t)
	request := EncodingRequest{
		VideoKey:   "video",
		AudiosKeys: []string{"audio"},
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

func copy(daprClient *client.Client, src string, keyName string) error {
	rawContent, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	b64Content := make([]byte, base64.StdEncoding.EncodedLen(len(rawContent)))
	base64.StdEncoding.Encode(b64Content, rawContent)
	_, err = (*daprClient).InvokeBinding(context.Background(), &client.InvokeBindingRequest{
		Name:      ObjStoreComponent,
		Operation: "create",
		Data:      b64Content,
		Metadata: map[string]string{
			"key": keyName,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
