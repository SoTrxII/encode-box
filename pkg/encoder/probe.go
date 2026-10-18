package encoder

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func GetDuration(path string) (time.Duration, error) {
	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 input.mp4
	arg := strings.Split("-v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1", " ")
	arg = append(arg, path)
	cmd := exec.Command("ffprobe", arg...)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	dur, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(int(dur)) * time.Second, nil
}

// CountVideoStreams tells how many picture tracks a file holds: the jukebox
// recording carries one per uninterrupted capture when the GM records the tab's
// picture too, none when only its sound
func CountVideoStreams(path string) (int, error) {
	out, err := exec.Command("ffprobe", "-v", "error", "-select_streams", "v",
		"-show_entries", "stream=index", "-of", "csv=p=0", path).Output()
	if err != nil {
		return 0, err
	}
	return len(strings.Fields(string(out))), nil
}
