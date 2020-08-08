package internal

import (
	"bufio"
	"bytes"
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
	done := make(chan struct{})
	go func() {
		piper.New(os.Stdin, toPipe).Start()
		done <- struct{}{}
	}()
	factoryC := cmd.NewFactory(h.flagC)
	factoryD := cmd.NewFactory(h.flagD)
	for {
		select {
		case <-done:
			return 0
		case s := <-toPipe:
			input := []byte(s)
			stdoutC, err := h.run(factoryC, input)
			if err != nil {
				continue
			}
			if h.flagD == "" {
				h.outputC(input, stdoutC)
				continue
			}
			stdoutD, err := h.run(factoryD, input)
			if err != nil {
				continue
			}
			if bytes.Equal(stdoutC, stdoutD) {
				continue
			}
			h.outputD(input, stdoutC, stdoutD)
		}
	}
}

func (h Handler) run(f cmd.Factory, input []byte) ([]byte, error) {
	defer h.stderr.Flush()
	stdout, stderr, err := f.Run(input)
	if len(stderr) > 0 {
		_, _ = h.stderr.Write(stderr)
		_ = h.stderr.WriteByte('\n')
	}
	if err != nil {
		_, _ = h.stderr.WriteString("error running command: " + err.Error())
		_ = h.stderr.WriteByte('\n')
	}
	return stdout, err
}

func (h Handler) outputC(input, output []byte) {
	defer h.stdout.Flush()
	if !h.flagO {
		_, _ = h.stdout.WriteString("(input)\n")
		_, _ = h.stdout.Write(input)
	}
	_, _ = h.stdout.WriteString("(output)\n")
	_, _ = h.stdout.Write(output)
	_ = h.stdout.WriteByte('\n')
}

func (h Handler) outputD(input, outputC, outputD []byte) {
	defer h.stdout.Flush()
	if !h.flagO {
		_, _ = h.stdout.WriteString("(input)\n")
		_, _ = h.stdout.Write(input)
	}
	_, _ = h.stdout.WriteString("(output: " + h.flagC + ")\n")
	_, _ = h.stdout.Write(outputC)
	_, _ = h.stdout.WriteString("(output: " + h.flagD + ")\n")
	_, _ = h.stdout.Write(outputD)
	_ = h.stdout.WriteByte('\n')
}
