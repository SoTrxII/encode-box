package test_utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type Resource string

const (
	Text  = "test.txt"
	Audio = "audio.m4a"
	Image = "test.jpg"
	Video = "video.mp4"
)
const (
	ResPath = "../resources/test"
)

func GetResAbsolutePath(t *testing.T, r Resource) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Couldn't get current file path")
	}
	dir := filepath.Dir(filename)
	return filepath.Join(dir, ResPath, string(r))
}

func GetAssetContent(t *testing.T, r Resource, b64 bool) []byte {
	content, err := os.ReadFile(GetResAbsolutePath(t, r))
	if err != nil {
		t.Fatal(err)
	}
	if !b64 {
		return content
	}
	return []byte(base64.StdEncoding.EncodeToString(content))
}

// Checksum returns the SHA-256 checksum of the specified file
func GetChecksum(filename string) (string, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// Create a new SHA-256 hasher
	hasher := sha256.New()
	// Copy the file contents to the hasher
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}
	// Get the checksum as a byte slice
	checksumBytes := hasher.Sum(nil)
	// Convert the checksum to a hexadecimal string
	checksum := hex.EncodeToString(checksumBytes)
	return checksum, nil
}
