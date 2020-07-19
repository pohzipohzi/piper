package internal

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pohzipohzi/piper/internal/cmd"
	"github.com/pohzipohzi/piper/internal/piper"
)

func Main() {
	go func() {
		sigChan := make(chan os.Signal, 2)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigChan
		fmt.Fprintln(os.Stderr, "Received signal:", sig)
		os.Exit(0)
	}()

	var (
		help    bool
		command string
		diff    string
	)
	flag.BoolVar(&help, "h", false, "print usage")
	flag.StringVar(&command, "c", "", "the command to run")
	flag.StringVar(&diff, "d", "", "(optional) the command to diff against")
	flag.Parse()
	if help || command == "" {
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
	cmdArgs := strings.Split(command, " ")
	cmdFactory = cmd.NewFactory(cmdArgs[0], cmdArgs[1:])
	if diff != "" {
		diffArgs := strings.Split(diff, " ")
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
			res, err := f(b)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error running command:", err)
				continue
			}
			if diff == "" {
				stdout.WriteString("(input)\n")
				stdout.Write(b)
				stdout.WriteString("(output)\n")
				stdout.Write(res)
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
			res2, err := f2(b)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error running command:", err)
				continue
			}
			if bytes.Equal(res, res2) {
				continue
			}
			stdout.WriteString("(input)\n")
			stdout.Write(b)
			stdout.WriteString("(output: " + command + ")\n")
			stdout.Write(res)
			stdout.WriteString("(output: " + diff + ")\n")
			stdout.Write(res2)
			stdout.WriteByte('\n')
			stdout.Flush()
		}
	}
}
