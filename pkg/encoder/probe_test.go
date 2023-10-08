package encoder

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProbe_GetDuration_Video(t *testing.T) {
	duration, err := GetDuration(TestVideo)
	assert.NoError(t, err)
	assert.Equal(t, 10.0, duration.Seconds())
}

func TestProbe_GetDuration_AudioShort(t *testing.T) {
	duration, err := GetDuration(TestAudio1)
	assert.NoError(t, err)
	assert.Equal(t, 3.0, duration.Seconds())
}

func TestProbe_GetDuration_AudioLong(t *testing.T) {
	duration, err := GetDuration(TestDialog)
	assert.NoError(t, err)
	assert.Equal(t, 22.0, duration.Seconds())
}

func TestProbe_GetDuration_AudioVeryLong(t *testing.T) {
	duration, err := GetDuration(TestBackground)
	assert.NoError(t, err)
	assert.Equal(t, 10.0*60, duration.Seconds())
}

// Ffprobe actually return a value for images, it should be 0
func TestProbe_GetDuration_Image(t *testing.T) {
	duration, err := GetDuration(TestImage)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, duration.Seconds())
}

func TestProbe_GetDuration_NonExisting(t *testing.T) {
	duration, err := GetDuration("non-existing")
	assert.Error(t, err)
	assert.Zero(t, duration)
}
