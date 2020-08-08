package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/pohzipohzi/piper/internal/cmd"
	"github.com/pohzipohzi/piper/internal/piper"
)

func Run(flagC string, flagD string, flagO bool) int {
	done := make(chan struct{})
	cmdStdinChan := make(chan string)
	go func() {
		piper.New(os.Stdin, cmdStdinChan).Start()
		done <- struct{}{}
	}()

	cmdFactory := cmd.FactoryFromString(flagC)
	diffFactory := cmd.FactoryFromString(flagD)

	stdout := bufio.NewWriter(os.Stdout)

	for {
		select {
		case <-done:
			return 0
		case s := <-cmdStdinChan:
			b := []byte(s)

			// run command
			f, err := cmdFactory.New()
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
			if flagD == "" {
				if !flagO {
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
			f2, err := diffFactory.New()
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
			if !flagO {
				stdout.WriteString("(input)\n")
				stdout.Write(b)
			}
			stdout.WriteString("(output: " + flagC + ")\n")
			stdout.Write(fstdout)
			stdout.WriteString("(output: " + flagD + ")\n")
			stdout.Write(f2stdout)
			stdout.WriteByte('\n')
			stdout.Flush()
		}
	}
}
