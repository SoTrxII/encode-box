package object_storage

import (
	test_utils "encode-box/test-utils"
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

// Check that the streamijng way to build the B64 signature is identical to the
// non-streaming way
func TestObjectStorage_readFileToB64(t *testing.T) {
	path := path.Join(test_utils.GetResAbsolutePath(t, test_utils.Text))
	content, err := os.ReadFile(path)
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
