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
	stdout *bufio.Writer
	stderr *bufio.Writer
}

func NewHandler(flagC string, flagD string, flagO bool) Handler {
	return Handler{
		flagC:  flagC,
		flagD:  flagD,
		flagO:  flagO,
		stdout: bufio.NewWriter(os.Stdout),
		stderr: bufio.NewWriter(os.Stderr),
	}
}

func (h Handler) Run() int {
	toPipe := make(chan string)
	done := h.startPiping(os.Stdin, toPipe)
	cmdFactory := cmd.NewFactory(h.flagC)
	diffFactory := cmd.NewFactory(h.flagD)
	for {
		select {
		case <-done:
			return 0
		case s := <-toPipe:
			input := []byte(s)
			cmdStdout, err := h.run(cmdFactory, input)
			if err != nil {
				continue
			}
			if h.flagD == "" {
				if !h.flagO {
					h.stdout.WriteString("(input)\n")
					h.stdout.Write(input)
				}
				h.stdout.WriteString("(output)\n")
				h.stdout.Write(cmdStdout)
				h.stdout.WriteByte('\n')
				h.stdout.Flush()
				continue
			}
			diffStdout, err := h.run(diffFactory, input)
			if err != nil {
				continue
			}
			if bytes.Equal(cmdStdout, diffStdout) {
				continue
			}
			if !h.flagO {
				h.stdout.WriteString("(input)\n")
				h.stdout.Write(input)
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

func (h Handler) run(f cmd.Factory, input []byte) ([]byte, error) {
	stdout, stderr, err := f.Run(input)
	if len(stderr) > 0 {
		h.stderr.Write(stderr)
		h.stderr.WriteByte('\n')
		h.stderr.Flush()
	}
	if err != nil {
		h.stderr.WriteString("error running command: " + err.Error())
		h.stderr.WriteByte('\n')
		h.stderr.Flush()
	}
	return stdout, err
}
