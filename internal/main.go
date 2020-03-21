package internal

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pohzipohzi/piper/internal/cmd"
	"github.com/pohzipohzi/piper/internal/piper"
)

func Main() {
	stderr := log.New(os.Stderr, "", log.LstdFlags)
	stdout := log.New(os.Stdout, "", 0)
	go func() {
		sigChan := make(chan os.Signal, 2)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigChan
		stdout.Println("Received signal:", sig)
		os.Exit(0)
	}()

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		stderr.Println("No command provided")
		return
	}
	cmdFactory := cmd.NewFactory(args[0], args[1:], cmd.WithLog(stderr))

	ctx, cancelFunc := context.WithCancel(context.Background())
	cmdStdinChan := make(chan string)
	go func() {
		piper.New(os.Stdin, cmdStdinChan).Start()
		cancelFunc()
	}()

	for {
		f, err := cmdFactory.New()
		if err != nil {
			stderr.Println("Error creating command:", err)
		}
		stderr.Println("Awaiting input")
		select {
		case <-ctx.Done():
			stderr.Println("Execution cancelled")
			return
		case s := <-cmdStdinChan:
			res, err := f([]byte(s))
			if err != nil {
				stderr.Println("Error running command:", err)
				continue
			}
			stderr.Println("Received result")
			stdout.Print(string(res))
		}
	}
}
