package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("No command provided")
		return
	}

	go exitOnTerminationSignal()

	ctx, cancelFunc := context.WithCancel(context.Background())
	stdinChan := make(chan string)
	wg := &sync.WaitGroup{}
	go func() {
		<-ctx.Done()
		wg.Wait()
		log.Println("Execution cancelled")
		os.Exit(0)
	}()
	go func() {
		receive(stdinChan, wg)
		cancelFunc()
	}()

	for ctx.Err() == nil {
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmdStdin, err := cmd.StdinPipe()
		if err != nil {
			log.Println("Error obtaining stdin:", err)
			continue
		}
		log.Println("Awaiting input")
		err = pipe(stdinChan, cmdStdin)
		if err != nil {
			log.Println("Error piping to command:", err)
			continue
		}
		bytes, err := cmd.Output()
		if err != nil {
			log.Println("Error running command:", err)
			continue
		}
		log.Print("Received result")
		fmt.Print(string(bytes))
		wg.Done()
	}
}

// exitOnTerminationSignal prepares the program to exit on receiving
// termination signals
func exitOnTerminationSignal() {
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigChan
	log.Println("Received signal:", sig)
	os.Exit(0)
}

// receive continually reads from os.Stdin, incrementing the number of results
// we expect to eventually obtain, and sends line-separated strings to a
// channel for processing
func receive(stdinChan chan<- string, wg *sync.WaitGroup) {
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
func pipe(stdinChan <-chan string, cmdStdin io.WriteCloser) error {
	writeCloser := withLog(cmdStdin)
	defer writeCloser.Close()
	s := <-stdinChan
	_, err := writeCloser.Write([]byte(s))
	return err
}

// withLog decorates an io.WriteCloser, logging data that passes through it
func withLog(next io.WriteCloser) io.WriteCloser {
	r, w := io.Pipe()
	go func() {
		defer next.Close()
		p, err := ioutil.ReadAll(r)
		if err != nil {
			log.Println("Failed to read in intermediate pipe:", err)
			return
		}
		log.Printf("Received input\n%s", string(p))
		_, err = next.Write(p)
		if err != nil {
			log.Println("Failed to write in intermediate pipe:", err)
		}
	}()
	return w
}
