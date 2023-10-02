package object_storage

import (
	"bufio"
	"bytes"
	"context"
	"encode-box/internal/utils"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

type ObjectStore interface {
	// Download a file from the backend storage
	Download(key, path string) error
	// Buffer the content of a file in memory
	Buffer(key string) (data *io.Reader, err error)
	// Upload Uploads a file on the backend storage
	Upload(path string, key string) error
	// Delete a file in the remote object storage
	Delete(key string) error
}

// ObjectStorage any S3-like storage solution
type ObjectStorage struct {
	// Name of the Dapr component to use
	componentName string
	// Client to query the backend storage
	client utils.Binder
	// Current running context
	ctx *context.Context
	// Are file stored in base 64 ? True for most components
	isBase64 bool
}

// NewObjectStorage Prod ready constructor for an object-storage using Dapr
func NewObjectStorage(ctx *context.Context, client utils.Binder, component string, base64 bool) *ObjectStorage {
	return &ObjectStorage{
		componentName: component,
		client:        client,
		ctx:           ctx,
		isBase64:      base64,
	}
}

// Download a file from the backend storage
func (od *ObjectStorage) Download(key, path string) error {
	reader, err := od.Buffer(key)
	if err != nil {
		return err
	}
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, *reader)
	return err
}

// Buffer the content of a file in memory
func (od *ObjectStorage) Buffer(key string) (data *io.Reader, err error) {
	res, err := od.client.InvokeBinding(*od.ctx, &utils.InvokeBindingRequest{
		Name:      od.componentName,
		Operation: "get",
		Data:      nil,
		Metadata: map[string]string{ // The key used to change the name of the file isn't consistent across component. Weird
			// https://docs.dapr.io/reference/components-reference/supported-bindings/s3/
			"key": key,
			// https://docs.dapr.io/reference/components-reference/supported-bindings/localstorage/
			"fileName": key,
		},
	})
	if err != nil {
		return nil, err
	}

	// If it's not base64, just return the data
	if !od.isBase64 {
		var reader io.Reader = bytes.NewReader(res.Data)
		return &reader, nil
	}

	// Else, decode the data
	input := bytes.NewBuffer(res.Data)
	decoder := base64.NewDecoder(base64.StdEncoding, input)
	return &decoder, nil
}

// Upload Uploads a file on the backend storage
func (od *ObjectStorage) Upload(path string, key string) error {
	b64bytes, err := readFileToB64(path)
	if err != nil {
		return err
	}
	_, err = od.client.InvokeBinding(*od.ctx, &utils.InvokeBindingRequest{
		Name:      od.componentName,
		Operation: "create",
		Data:      b64bytes,
		Metadata: map[string]string{
			// The key used to change the name of the file isn't consistent across component. Weird
			// https://docs.dapr.io/reference/components-reference/supported-bindings/s3/
			"key": key,
			// https://docs.dapr.io/reference/components-reference/supported-bindings/localstorage/
			"fileName": key,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// Delete a file in the remote object storage
func (od *ObjectStorage) Delete(key string) error {
	_, err := od.client.InvokeBinding(*od.ctx, &utils.InvokeBindingRequest{
		Name:      od.componentName,
		Operation: "delete",
		Data:      nil,
		Metadata: map[string]string{
			// The key used to change the name of the file isn't consistent across component. Weird
			// https://docs.dapr.io/reference/components-reference/supported-bindings/s3/
			"key": key,
			// https://docs.dapr.io/reference/components-reference/supported-bindings/localstorage/
			"fileName": key,
		},
	})
	return err
}

// Read a file into a base64 bytes-array
func readFileToB64(path string) ([]byte, error) {
	var buf bytes.Buffer
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	b64enc := base64.NewEncoder(base64.StdEncoding, &buf)
	// Make a 10 MB buffer
	p := make([]byte, 10*1024*1024)
	for {
		n, err := reader.Read(p)
		if err != nil {
			if err == io.EOF {
				b64enc.Write(p[:n])
				fmt.Println(string(p[:n]))
				break
			}
			fmt.Println(err)
		}
		b64enc.Write(p[:n])
	}
	err = b64enc.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
