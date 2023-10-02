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
	"strconv"
	"testing"
)

const (
	ObjStoreComponent = "object-store"
)

// These are integration test, using all real components
// Dapr and the backend storage should be booted up for this to work

func SetupInt(t *testing.T) (string, *encode_box.EncodeBox) {
	ctx := context.Background()
	daprClient, err := client.NewClientWithPort(strconv.Itoa(50010))
	if err != nil {
		t.Fatal(err)
	}
	err = copyToStorage(&daprClient, filepath.Join(ResDir, "audio.m4a"), "audio")
	if err != nil {
		t.Fatal(err)
	}
	err = copyToStorage(&daprClient, filepath.Join(ResDir, "audio.m4a"), "audio2")
	if err != nil {
		t.Fatal(err)
	}
	err = copyToStorage(&daprClient, filepath.Join(ResDir, "audio.m4a"), "audio3")
	if err != nil {
		t.Fatal(err)
	}
	err = copyToStorage(&daprClient, filepath.Join(ResDir, "video.mp4"), "video")
	if err != nil {
		t.Fatal(err)
	}
	err = copyToStorage(&daprClient, filepath.Join(ResDir, "test.jpg"), "image")
	if err != nil {
		t.Fatal(err)
	}

	dir, err := os.MkdirTemp("", "test-encode-int")
	if err != nil {
		t.Fatal(err)
	}
	objStore := object_storage.NewObjectStorage(&ctx, daprClient, ObjStoreComponent, b64)
	eBox := encode_box.NewEncodeBox(&ctx, objStore, &encode_box.EncodeBoxOptions{ObjStoreMaxRetry: 3})
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
	err, _ = encode(eBox, &req, filepath.Join(dir, "out.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dir)
}
func TestMain_Encode_AudioImage_Int(t *testing.T) {
	dir, eBox := SetupInt(t)
	req := encode_box.EncodingRequest{
		VideoKey:   "",
		AudiosKeys: []string{"audio", "audio2", "audio3"},
		ImageKey:   "image",
		Options:    encode_box.EncodingOptions{},
	}
	// Make temp dir for output
	dir, err := os.MkdirTemp("", "encode-instance")
	if err != nil {
		t.Fatal(err)
	}
	err, _ = encode(eBox, &req, filepath.Join(dir, "out.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dir)
}
func TestMain_Encode_AudioOnly_Int(t *testing.T) {
	dir, eBox := SetupInt(t)
	req := encode_box.EncodingRequest{
		VideoKey:   "",
		AudiosKeys: []string{"audio", "audio2", "audio3"},
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{},
	}
	// Make temp dir for output
	dir, err := os.MkdirTemp("", "encode-instance")
	if err != nil {
		t.Fatal(err)
	}
	err, _ = encode(eBox, &req, filepath.Join(dir, "out.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dir)
}
func copyToStorage(daprClient *client.Client, src string, keyName string) error {
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
			"key":      keyName,
			"fileName": keyName,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
