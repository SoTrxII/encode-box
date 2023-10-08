package console_parser

import (
	"bufio"
	"bytes"
	"context"
	log "encode-box/pkg/logger"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var logger = log.Build()

// EncodingProgress A progress as emitted by Ffmpeg
type EncodingProgress struct {
	// Number of total frames processed
	Frames int64 `json:"frames"`
	// Number of frames processed each second
	Fps int `json:"fps"`
	// Quality target. Usually between 20 and 30
	Quality float32 `json:"quality"`
	// Estimated size of the converted file (kb)
	Size int64 `json:"size"`
	// Total processed time
	Time time.Time `json:"time"`
	// Target bitrate
	Bitrate string `json:"bitrate"`
	// Encoding speed. A "2" means 1 second of encoding would be a 2 seconds playback
	Speed float32 `json:"speed"`
	// The duration of the output file
	TargetDuration time.Duration `json:"totalDuration"`
}

func ParseOutput(ctx *context.Context, readStream *io.ReadCloser, progressChan chan *EncodingProgress, errorChan chan error) string {
	scanner := bufio.NewScanner(*readStream)
	scanner.Split(scanFfmpegOutput)
	// Match every spaces or group of spaces not preceded by an "=" or a space
	// this splits a typical FFMPEG progress line
	// frame= 2805 fps=329 q=28.0 size=       0kB time=00:01:49.84 bitrate=   0.0kbits/s speed=12.9x

	// Store the last n ffmpeg lines
	stack := NewRingLogBuffer(5)
	for scanner.Scan() {
		// Check if the operation must be cancelled
		select {
		case <-(*ctx).Done():
			return stack.String()
		default:
			// Continue
		}
		// Retrieve a new line...
		if scanner.Err() != nil {
			errorChan <- scanner.Err()
		}
		line := scanner.Text()
		stack.Push(line)
		//fmt.Println(line)
		// And check if it's the progress line
		// A progress line begins with "frame = xxx". Discard the line otherwise
		if !strings.HasPrefix(line, "f") {
			continue
		}
		progress, err := parseProgress(line)
		//lineBuffer = line
		if err != nil {
			logger.Warnf("[Console parser] :: progress line \"%s\" ignored", line)
			continue
		}
		// Return the parsed progress
		progressChan <- progress

	}
	return stack.String()
}

// scanFfmpegOutput A modified version of a traditional scanLine, allowing to parse FFMpeg output
func scanFfmpegOutput(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// In FFMPEG's case, having a non newline terminated by \r means we have a progress line
		return i + 1, dropCR(data[0:i]), nil
	}

	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

// Parse a progress line in the following format :
// frame=   85 fps=0.0 q=28.0 size=       0kB time=00:00:01.04 bitrate=   0.4kbits/s speed=   2x
func parseProgress(progressLine string) (*EncodingProgress, error) {
	blanks := regexp.MustCompile(`=\s+`)
	components := strings.Split(blanks.ReplaceAllString(strings.TrimSpace(progressLine), "="), " ")
	p := &EncodingProgress{}
	for _, c := range components {
		err, key, value := parseComponentString(c)
		if err != nil {
			return nil, err
		}
		switch key {
		case "frame":
			if n, err := strconv.ParseInt(value, 10, 64); err == nil {
				p.Frames = n
			}
			break
		case "fps":
			if n, err := strconv.ParseInt(value, 10, 32); err == nil {
				p.Fps = int(n)
			}
			break
		case "q":
			if n, err := strconv.ParseFloat(value, 32); err == nil {
				p.Quality = float32(n)
			}
			break
		case "size":
			if s, err := parseSize(value); err == nil {
				p.Size = s
			}
			break
		case "bitrate":
			p.Bitrate = value
			break
		case "time":
			if t, err := time.Parse("15:04:05", value); err == nil {
				p.Time = t
			}
			break
		case "speed":
			if n, err := strconv.ParseFloat(strings.TrimSuffix(value, "x"), 32); err == nil {
				p.Speed = float32(n)
			}
			break
		}
	}
	return p, nil
}

// Parse a string with format "key=value" into a key/value pair
func parseComponentString(str string) (error error, key string, value string) {
	pair := strings.Split(str, "=")
	if len(pair) != 2 {
		return fmt.Errorf("Invalid pair : %s", pair), "", ""
	}
	return nil, pair[0], pair[1]
}

// Parse a "12kb" string into a Kilo based size
func parseSize(rawSize string) (int64, error) {
	sizeRegex := regexp.MustCompile(`(\d+)([a-zA-Z]+)`)
	units := []string{"kb", "mb", "gb"}
	matches := sizeRegex.FindStringSubmatch(rawSize)
	if len(matches) != 3 {
		return 0, fmt.Errorf("Invalid format")
	}
	unit := strings.ToLower(matches[2])
	size, err := strconv.ParseInt(strings.ToLower(matches[1]), 10, 64)
	if err != nil {
		return 0, err
	}
	multiplier := getSizeMultiplier(unit, units)
	if multiplier == -1 {
		return 0, err
	}
	return size * multiplier, nil

}

// Return a multiplier to convert any unit into kb
func getSizeMultiplier(element string, data []string) int64 {
	for k, v := range data {
		if element == v {
			return int64(math.Pow(float64(1024), float64(k)))
		}
	}
	return -1 //not found.
}
