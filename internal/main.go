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
}

func NewHandler(flagC string, flagD string, flagO bool) Handler {
	return Handler{
		isOutputOnly: flagO,
		cmd:          cmd.NewFactory(flagC),
		diff:         cmd.NewFactory(flagD),
	}
}

func (h Handler) Run() int {
	stdout := bufio.NewWriter(os.Stdout)
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
					stdout.WriteString("(input)\n")
					stdout.Write(b)
				}
				stdout.WriteString("(output)\n")
				stdout.Write(fstdout)
				stdout.WriteByte('\n')
				stdout.Flush()
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
				stdout.WriteString("(input)\n")
				stdout.Write(b)
			}
			stdout.WriteString("(output: " + h.cmd.String() + ")\n")
			stdout.Write(fstdout)
			stdout.WriteString("(output: " + h.diff.String() + ")\n")
			stdout.Write(f2stdout)
			stdout.WriteByte('\n')
			stdout.Flush()
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
