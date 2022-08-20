package encoder

import (
	"context"
	console_parser "encode-box/pkg/encoder/console-parser"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Encoder struct {
	// FFMpeg command line to execute
	cmd string
	// Channel to send progress into
	PChan chan *console_parser.EncodingProgress
	// Channel to send errors into
	EChan chan error
	// Encoder context
	Ctx context.Context
	// Function to execute to Cancel the encoding process
	Cancel context.CancelFunc
}

// NewEncoder Build a new FFMpeg encoder initialized with cmd as a command
func NewEncoder(ctx *context.Context, cmd string) *Encoder {
	eCtx, cancel := context.WithCancel(*ctx)
	return &Encoder{
		cmd:    cmd,
		PChan:  make(chan *console_parser.EncodingProgress),
		EChan:  make(chan error),
		Ctx:    eCtx,
		Cancel: cancel,
	}
}

func (e *Encoder) Start() {
	defer e.Cancel()
	sanitizedCmd := regexp.MustCompile(`\s+`).ReplaceAllString(e.cmd, " ")
	// Handle a special case. An FFmpeg command can have a -filter_complex option.
	// This option can contains spaces, and if it does, the exec.Command won't work.
	// But we must split all the argument for command exec
	// So if the command contains a "-filter_complex", we extract it before splitting the command,
	// and inserting it back later
	matches := regexp.MustCompile(`-filter_complex "(.+?)"`).FindAllStringSubmatch(sanitizedCmd, 1)
	hasComplexFilter := len(matches) > 0 && len(matches[0]) == 2
	var filterGraphString string
	if hasComplexFilter {
		filterGraphString = matches[0][1]
		sanitizedCmd = strings.Replace(sanitizedCmd, filterGraphString, "fg", 1)
	}
	arguments := strings.Split(sanitizedCmd, " ")[1:]
	if hasComplexFilter {
		for i, arg := range arguments {
			if strings.HasPrefix(arg, "-filter_complex") {
				arguments[i+1] = filterGraphString
			}
		}
	}
	cmd := exec.Command("ffmpeg", arguments...)

	// FFMpeg pipe output in stderr for some reason
	stderr, err := cmd.StderrPipe()
	if err != nil {
		e.EChan <- err
	}

	err = cmd.Start()
	if err != nil {
		e.EChan <- err
	}
	line := console_parser.ParseOutput(&e.Ctx, &stderr, e.PChan, e.EChan)
	err = cmd.Wait()
	if err != nil {
		e.EChan <- fmt.Errorf(line)
	}
}

func (e *Encoder) GetCommandLine() string {
	return e.cmd
}

////ffmpeg -i ./part0.ogg -i part1.ogg  -filter_complex '[0][1]concat=n=2:v=0:a=1[out]' -map [out] output.ogg

// "[0:a]loudnorm=I=-16:TP=-1.5:LRA=11, aformat=sample_fmts=fltp:sample_rates=44100:channel_layouts=stereo[r1];[1]loudnorm=I=-16:TP=-1.5:LRA=11,asplit=2[sc][v1];[r1][sc]sidechaincompress=threshold=0.05:ratio=5:level_sc=0.8[bg];[bg][v1]amix=weights=0.2 1[a3]"
/*
   this.streamProcess
     .input("color=black:s=1280x720:r=25")
     .inputFormat("lavfi")
     .inputFormat("-i anullsrc=r=48000:cl=stereo")
     .map("-map 1:a")
     .map("-map 0:v")
     .outputOption("-c:v libx264rgb");

*/

/*
   this.streamProcess
     .addOption("-analyzeduration 0")
     .outputOption("-preset ultrafast")
     .Format("flv")
     .on("error", (err) => {
       if (this.allowSubProcessToBeKilled) {
         this.allowSubProcessToBeKilled = false;
       }
       console.error(err);
     })
     .on("start", (cmdLine) => console.log(cmdLine))
     .output(output);
   this.streamProcess.run();

*/
