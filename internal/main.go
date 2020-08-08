package internal

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"github.com/pohzipohzi/piper/internal/cmd"
	"github.com/pohzipohzi/piper/internal/piper"
)

type Handler struct {
	flagC  string
	flagD  string
	flagO  bool
	cmd    cmd.Factory
	diff   cmd.Factory
	stdout *bufio.Writer
	stderr *bufio.Writer
}

func NewHandler(flagC string, flagD string, flagO bool) Handler {
	return Handler{
		flagC:  flagC,
		flagD:  flagD,
		flagO:  flagO,
		cmd:    cmd.NewFactory(flagC),
		diff:   cmd.NewFactory(flagD),
		stdout: bufio.NewWriter(os.Stdout),
		stderr: bufio.NewWriter(os.Stderr),
	}
}

func (h Handler) Run() int {
	toPipe := make(chan string)
	done := h.startPiping(os.Stdin, toPipe)
	for {
		select {
		case <-done:
			return 0
		case s := <-toPipe:
			b := []byte(s)

			cmdStdout, cmdStderr, err := h.cmd.Run(b)
			if len(cmdStderr) > 0 {
				h.stderr.Write(cmdStderr)
				h.stderr.WriteByte('\n')
				h.stderr.Flush()
			}
			if err != nil {
				h.stderr.WriteString("error running command: " + err.Error())
				h.stderr.WriteByte('\n')
				h.stderr.Flush()
				continue
			}
			if h.flagD == "" {
				if !h.flagO {
					h.stdout.WriteString("(input)\n")
					h.stdout.Write(b)
				}
				h.stdout.WriteString("(output)\n")
				h.stdout.Write(cmdStdout)
				h.stdout.WriteByte('\n')
				h.stdout.Flush()
				continue
			}

			// run diff
			diffStdout, diffStderr, err := h.diff.Run(b)
			if len(diffStderr) > 0 {
				h.stderr.Write(cmdStderr)
				h.stderr.WriteByte('\n')
				h.stderr.Flush()
			}
			if err != nil {
				h.stderr.WriteString("error running command: " + err.Error())
				h.stderr.WriteByte('\n')
				h.stderr.Flush()
				continue
			}
			if bytes.Equal(cmdStdout, diffStdout) {
				continue
			}
			if !h.flagO {
				h.stdout.WriteString("(input)\n")
				h.stdout.Write(b)
			}
			h.stdout.WriteString("(output: " + h.flagC + ")\n")
			h.stdout.Write(cmdStdout)
			h.stdout.WriteString("(output: " + h.flagD + ")\n")
			h.stdout.Write(diffStdout)
			h.stdout.WriteByte('\n')
			h.stdout.Flush()
		}
	}
}

func (h Handler) startPiping(in io.Reader, out chan<- string) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		piper.New(in, out).Start()
		done <- struct{}{}
	}()
	return done
}
