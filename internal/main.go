package internal

import (
	"bufio"
	"context"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pohzipohzi/piper/internal/cmd"
)

func Main() {
	stderrLogger := log.New(os.Stderr, "", log.LstdFlags)
	stdoutLogger := log.New(os.Stdout, "", 0)
	go func() {
		sigChan := make(chan os.Signal, 2)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigChan
		stdoutLogger.Println("Received signal:", sig)
		os.Exit(0)
	}()

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		stderrLogger.Println("No command provided")
		os.Exit(0)
	}
	cmdFactory := cmd.NewFactory(args[0], args[1:], []func(io.WriteCloser) io.WriteCloser{withLog(stderrLogger)})

	piper{
		stderr:     stderrLogger,
		stdout:     stdoutLogger,
		cmdFactory: cmdFactory,
	}.Run()
}

type piper struct {
	stderr     *log.Logger
	stdout     *log.Logger
	cmdFactory cmd.Factory
}

func (p piper) Run() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	stdinChan := make(chan string)
	wg := &sync.WaitGroup{}
	go func() {
		<-ctx.Done()
		wg.Wait()
		p.stderr.Println("Execution cancelled")
		os.Exit(0)
	}()
	go func() {
		p.receive(stdinChan, wg)
		cancelFunc()
	}()

	for ctx.Err() == nil {
		f, err := p.cmdFactory.New()
		if err != nil {
			p.stderr.Println("Error creating command:", err)
		}
		p.stderr.Println("Awaiting input")
		s := <-stdinChan
		res, err := f([]byte(s))
		if err != nil {
			p.stderr.Println("Error running command:", err)
			continue
		}
		p.stderr.Println("Received result")
		p.stdout.Print(string(res))
		wg.Done()
	}
}

// receive continually reads from os.Stdin, incrementing the number of results
// we expect to eventually obtain, and sends line-separated strings to a
// channel for processing
func (p piper) receive(stdinChan chan<- string, wg *sync.WaitGroup) {
	scanner := bufio.NewScanner(os.Stdin)
	toPipe := ""
	for scanner.Scan() {
		s := scanner.Text()
		if s == "" && len(toPipe) > 0 {
			wg.Add(1)
			stdinChan <- toPipe
			toPipe = ""
			continue
		}
		if s != "" {
			toPipe += s + "\n"
		}
	}
	if len(toPipe) > 0 {
		wg.Add(1)
		stdinChan <- toPipe
	}
}

// withLog decorates an io.WriteCloser, logging data that passes through it
func withLog(stderrLogger *log.Logger) func(next io.WriteCloser) io.WriteCloser {
	return func(next io.WriteCloser) io.WriteCloser {
		r, w := io.Pipe()
		go func() {
			defer next.Close()
			b, err := ioutil.ReadAll(r)
			if err != nil {
				stderrLogger.Println("Failed to read in intermediate pipe:", err)
				return
			}
			stderrLogger.Printf("Received input\n%s", string(b))
			_, err = next.Write(b)
			if err != nil {
				stderrLogger.Println("Failed to write in intermediate pipe:", err)
			}
		}()
		return w
	}
}
