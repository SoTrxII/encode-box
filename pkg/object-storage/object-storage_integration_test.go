//go:build integration
// +build integration

// Be sure to run "make dapr_test" before running this in your IDE
package object_storage

import (
	"context"
	test_utils "encode-box/test-utils"
	"fmt"
	"github.com/dapr/go-sdk/client"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

const (
	// This is bound to local storage
	DEFAULT_COMPONENT = "object-store"
	// This is bound to minio
	MINIO_COMPONENT = "object-store-minio"
	// This is not the real default port, this is the one defined in the Makefile
	DEFAULT_DAPR_PORT = 50010
	STORAGE_PATH      = "../../resources/storage/"
)

func setup(t *testing.T) *ObjectStorage {
	err := os.MkdirAll(STORAGE_PATH, os.ModePerm)
	assert.NoError(t, err)
	daprPort := DEFAULT_DAPR_PORT
	if envPort, err := strconv.ParseInt(os.Getenv("DAPR_GRPC_PORT"), 10, 32); err == nil && envPort != 0 {
		daprPort = int(envPort)
	}
	daprClient, err := client.NewClientWithPort(strconv.Itoa(daprPort))
	assert.NoError(t, err)
	ctx := context.Background()
	// The base64 property must be set to true if using minio !
	return NewObjectStorage(&ctx, daprClient, DEFAULT_COMPONENT, false)
}

func teardown(t *testing.T) {
	err := os.RemoveAll(STORAGE_PATH)
	if err != nil {
		fmt.Println(err)
	}
}
func TestUpload(t *testing.T) {
	objStore := setup(t)
	defer teardown(t)
	err := objStore.Upload(test_utils.GetResAbsolutePath(t, test_utils.Text), test_utils.Text)
	assert.NoError(t, err)
}

func TestDownload(t *testing.T) {
	objStore := setup(t)
	defer teardown(t)
	// Make a temp dir to receive files
	dir, err := os.MkdirTemp("", "obj-store-test")
	assert.NoError(t, err)
	// Attempt to download a non-existing resource
	destPath := filepath.Join(dir, test_utils.Text)
	err = objStore.Download(test_utils.Text, destPath)
	assert.Error(t, err)

	// Attempt to download an existing resource
	copyToStorage(test_utils.GetResAbsolutePath(t, test_utils.Text))
	err = objStore.Download(test_utils.Text, destPath)
	assert.NoError(t, err)
	_, err = os.Stat(destPath)
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
	objStore := setup(t)
	defer teardown(t)
	// Attempt to delete a non-existing resource
	err := objStore.Delete(test_utils.Text)
	assert.Error(t, err)

	// Attempt to delete an existing resource
	copyToStorage(test_utils.GetResAbsolutePath(t, test_utils.Text))
	err = objStore.Delete(test_utils.Text)
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(STORAGE_PATH, test_utils.Text))
	assert.True(t, os.IsNotExist(err))
}

func TestUploadAndDelete(t *testing.T) {
	objStore := setup(t)
	defer teardown(t)
	err := objStore.Upload(test_utils.GetResAbsolutePath(t, test_utils.Text), test_utils.Text)
	assert.NoError(t, err)
	err = objStore.Delete(test_utils.Text)
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(STORAGE_PATH, test_utils.Text))
	assert.True(t, os.IsNotExist(err))
}

func TestUploadAndDownload(t *testing.T) {
	objStore := setup(t)
	defer teardown(t)
	err := objStore.Upload(test_utils.GetResAbsolutePath(t, test_utils.Text), test_utils.Text)
	assert.NoError(t, err)
	// Make a temp dir to receive files
	dir, err := os.MkdirTemp("", "obj-store-test")
	assert.NoError(t, err)
	// Attempt to download an existing resource
	destPath := filepath.Join(dir, test_utils.Text)
	err = objStore.Download(test_utils.Text, destPath)
	assert.NoError(t, err)
	_, err = os.Stat(destPath)
	assert.NoError(t, err)

	// Ensure both file contents are the same
	srcSum, err := test_utils.GetChecksum(test_utils.GetResAbsolutePath(t, test_utils.Text))
	assert.NoError(t, err)
	dstSum, err := test_utils.GetChecksum(destPath)
	assert.NoError(t, err)
	assert.Equal(t, srcSum, dstSum)
}

// Manually copy a file to the object store
func copyToStorage(src string) {
	input, err := os.ReadFile(src)
	if err != nil {
		fmt.Println(err)
		return
	}
	fileName := filepath.Base(src)
	err = os.WriteFile(filepath.Join(STORAGE_PATH, fileName), input, 0644)
	if err != nil {
		fmt.Println("Error creating", STORAGE_PATH)
		fmt.Println(err)
		return
	}
}
