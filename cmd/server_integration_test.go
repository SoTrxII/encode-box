//go:build integration
// +build integration

package main

import (
	"context"
	encode_box "encode-box/pkg/encode-box"
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
)

// These are integration test, using all real components
// Dapr and the backend storage should be booted up for this to work

func SetupInt(t *testing.T) (string, *encode_box.EncodeBox[client.Client]) {
	ctx := context.Background()
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
	objStore, err := object_storage.NewDaprObjectStorage(&ctx, &daprClient, ObjStoreComponent)
	eBox := encode_box.NewEncodeBox(&ctx, objStore)
	return dir, eBox
}

func TestMain_Encode_AudioVideo_Int(t *testing.T) {
	dir, eBox := SetupInt(t)
	req := encode_box.EncodingRequest{
		VideoKey:   "video",
		AudiosKeys: []string{"audio", "audio2", "audio3"},
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{},
	}
	// Make temp dir for output
	dir, err := os.MkdirTemp("", "encode-instance")
	if err != nil {
		t.Fatal(err)
	}
	err, _ = encode[client.Client](eBox, &req, filepath.Join(dir, "out.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dir)
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
