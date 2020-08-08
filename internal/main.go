package internal

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pohzipohzi/piper/internal/cmd"
	"github.com/pohzipohzi/piper/internal/piper"
)

func Main() {
	var (
		flagC string
		flagD string
		flagO bool
	)
	flag.StringVar(&flagC, "c", "", "the command to run")
	flag.StringVar(&flagD, "d", "", "(optional) the command to diff against")
	flag.BoolVar(&flagO, "o", false, "(optional) show output only")
	flag.Parse()
	if flagC == "" {
		flag.Usage()
		return
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	cmdStdinChan := make(chan string)
	go func() {
		piper.New(os.Stdin, cmdStdinChan).Start()
		cancelFunc()
	}()

	var (
		cmdFactory  cmd.Factory
		diffFactory cmd.Factory
	)
	cmdArgs := strings.Split(flagC, " ")
	cmdFactory = cmd.NewFactory(cmdArgs[0], cmdArgs[1:])
	if flagD != "" {
		diffArgs := strings.Split(flagD, " ")
		diffFactory = cmd.NewFactory(diffArgs[0], diffArgs[1:])
	}

	stdout := bufio.NewWriter(os.Stdout)

	for {
		select {
		case <-ctx.Done():
			return
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
