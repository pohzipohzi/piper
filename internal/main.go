package internal

import (
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
	)
	flag.BoolVar(&help, "h", false, "print usage")
	flag.StringVar(&command, "c", "", "the command to run")
	flag.Parse()
	if help || command == "" {
		flag.Usage()
		return
	}

	args := strings.Split(command, " ")
	cmdFactory := cmd.NewFactory(args[0], args[1:])

	ctx, cancelFunc := context.WithCancel(context.Background())
	cmdStdinChan := make(chan string)
	go func() {
		piper.New(os.Stdin, cmdStdinChan).Start()
		cancelFunc()
	}()

	for {
		f, err := cmdFactory.New()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating command:", err)
		}
		select {
		case <-ctx.Done():
			return
		case s := <-cmdStdinChan:
			fmt.Fprintln(os.Stderr, "INPUT")
			fmt.Fprint(os.Stderr, s)
			res, err := f([]byte(s))
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error running command:", err)
				continue
			}
			fmt.Fprintln(os.Stderr, "OUTPUT")
			fmt.Fprint(os.Stdout, string(res))
		}
	}
}
