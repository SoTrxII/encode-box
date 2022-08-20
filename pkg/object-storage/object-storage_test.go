package object_storage

import (
	"context"
	mock_client "encode-box/internal/mock/dapr"
	"encoding/base64"
	"github.com/dapr/go-sdk/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const (
	ResPath = "../../resources/test/"
)

// Create a new empty temp dir
func Setup(t *testing.T) string {
	dir, err := ioutil.TempDir("", "prefix")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestObjectStorage_Download(t *testing.T) {
	dir := Setup(t)
	defer Teardown(t, dir)
	// And creates a new resource
	ctrl := gomock.NewController(t)
	daprClient := mock_client.NewMockClient(ctrl)
	//testFile := path.Join(ResPath, "test.txt")
	testFileContent, err := ioutil.ReadFile(path.Join(ResPath, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	// Dapr returns b64
	b64Content := base64.StdEncoding.EncodeToString(testFileContent)
	daprClient.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte(b64Content)}, nil)

	//
	ctx := context.Background()
	od := ObjectStorage[*mock_client.MockClient]{
		assetsPath:    dir,
		componentName: "test",
		client:        &daprClient,
		ctx:           &ctx,
	}
	path, _ := od.Download("test.txt")
	writtenFileContent, err := ioutil.ReadFile(*path)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, string(testFileContent), string(writtenFileContent))
}

func TestObjectStorage_Upload(t *testing.T) {
	dir := Setup(t)
	defer Teardown(t, dir)

	// And creates a new resource
	ctrl := gomock.NewController(t)
	daprClient := mock_client.NewMockClient(ctrl)
	daprClient.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: nil}, nil)

	//
	ctx := context.Background()
	od := ObjectStorage[*mock_client.MockClient]{
		assetsPath:    dir,
		componentName: "test",
		client:        &daprClient,
		ctx:           &ctx,
	}
	err := od.Upload(path.Join(ResPath, "test.txt"), "key")
	assert.Nil(t, err)
}

// Check that the streamijng way to build the B64 signature is identical to the
// non-streaming way
func TestObjectStorage_readFileToB64(t *testing.T) {
	path := path.Join(ResPath, "test.txt")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	control := base64.StdEncoding.EncodeToString(content)
	expected, err := readFileToB64(path)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, control, string(expected))
}

// Remove created dir
func Teardown(t *testing.T, dir string) {
	defer os.RemoveAll(dir)
}
