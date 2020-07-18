package internal

import (
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

	for {
		select {
		case <-ctx.Done():
			return
		case s := <-cmdStdinChan:
			fmt.Fprintln(os.Stderr, "INPUT")
			fmt.Fprint(os.Stderr, s)
			b := []byte(s)

			// run command
			f, err := cmdFactory.New()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error creating command:", err)
				continue
			}
			res, err := f(b)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error running command:", err)
				continue
			}
			if diff == "" {
				fmt.Fprintln(os.Stderr, "OUTPUT")
				fmt.Fprint(os.Stdout, string(res))
				continue
			}

			// run diff
			f2, err := diffFactory.New()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error creating command:", err)
				continue
			}
			res2, err := f2(b)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error running command:", err)
				continue
			}
			if bytes.Equal(res, res2) {
				fmt.Fprintln(os.Stderr, "EQUAL")
				continue
			}
			fmt.Fprintln(os.Stdout, "NOT EQUAL")
			fmt.Fprintln(os.Stderr, "Output for \""+command+"\"")
			fmt.Fprint(os.Stdout, string(res))
			fmt.Fprintln(os.Stderr, "Output for \""+diff+"\"")
			fmt.Fprint(os.Stdout, string(res2))
		}
	}
}
