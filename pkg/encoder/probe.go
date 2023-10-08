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
