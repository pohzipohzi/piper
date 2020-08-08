package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pohzipohzi/piper/internal/cmd"
	"github.com/pohzipohzi/piper/internal/piper"
)

type Handler struct {
	isOutputOnly bool
	cmd          cmd.Factory
	diff         cmd.Factory
	stdout       *bufio.Writer
}

func NewHandler(flagC string, flagD string, flagO bool) Handler {
	return Handler{
		isOutputOnly: flagO,
		cmd:          cmd.NewFactory(flagC),
		diff:         cmd.NewFactory(flagD),
		stdout:       bufio.NewWriter(os.Stdout),
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

			// run command
			f, err := h.cmd.New()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error creating command:", err)
				continue
			}
			fstdout, fstderr, err := f(b)
			if len(fstderr) > 0 {
				fmt.Fprintln(os.Stderr, string(fstderr))
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, "error running command:", err)
				continue
			}
			if h.diff == nil {
				if !h.isOutputOnly {
					h.stdout.WriteString("(input)\n")
					h.stdout.Write(b)
				}
				h.stdout.WriteString("(output)\n")
				h.stdout.Write(fstdout)
				h.stdout.WriteByte('\n')
				h.stdout.Flush()
				continue
			}

			// run diff
			f2, err := h.diff.New()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error creating command:", err)
				continue
			}
			f2stdout, _, err := f2(b)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error running command:", err)
				continue
			}
			if bytes.Equal(fstdout, f2stdout) {
				continue
			}
			if !h.isOutputOnly {
				h.stdout.WriteString("(input)\n")
				h.stdout.Write(b)
			}
			h.stdout.WriteString("(output: " + h.cmd.String() + ")\n")
			h.stdout.Write(fstdout)
			h.stdout.WriteString("(output: " + h.diff.String() + ")\n")
			h.stdout.Write(f2stdout)
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
