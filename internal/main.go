package internal

import (
	"bufio"
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
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

	piper{
		stderr: stderrLogger,
		stdout: stdoutLogger,
	}.Run()
}

type piper struct {
	stderr *log.Logger
	stdout *log.Logger
}

func (p piper) Run() {
	if len(os.Args) < 2 {
		p.stderr.Println("No command provided")
		return
	}

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
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmdStdin, err := cmd.StdinPipe()
		if err != nil {
			p.stderr.Println("Error obtaining stdin:", err)
			continue
		}
		p.stderr.Println("Awaiting input")
		err = p.pipe(stdinChan, cmdStdin)
		if err != nil {
			p.stderr.Println("Error piping to command:", err)
			continue
		}
		bytes, err := cmd.Output()
		if err != nil {
			p.stderr.Println("Error running command:", err)
			continue
		}
		p.stderr.Println("Received result")
		p.stdout.Print(string(bytes))
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

// pipe continually reads and writes bytes from a channel to a buffered writer,
// eventually piping the buffer (by closing the writer), returning errors if
// any
func (p piper) pipe(stdinChan <-chan string, cmdStdin io.WriteCloser) error {
	writeCloser := p.withLog(cmdStdin)
	defer writeCloser.Close()
	s := <-stdinChan
	_, err := writeCloser.Write([]byte(s))
	return err
}

// withLog decorates an io.WriteCloser, logging data that passes through it
func (p piper) withLog(next io.WriteCloser) io.WriteCloser {
	r, w := io.Pipe()
	go func() {
		defer next.Close()
		b, err := ioutil.ReadAll(r)
		if err != nil {
			p.stderr.Println("Failed to read in intermediate pipe:", err)
			return
		}
		p.stderr.Printf("Received input\n%s", string(b))
		_, err = next.Write(b)
		if err != nil {
			p.stderr.Println("Failed to write in intermediate pipe:", err)
		}
	}()
	return w
}
