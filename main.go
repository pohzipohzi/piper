package main

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

func main() {
	if len(os.Args) < 2 {
		log.Println("No command provided")
		return
	}

	// kill the program immediately on receiving signal
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Println("Received signal:", sig)
		os.Exit(0)
	}()

	// set up graceful shutdown for terminating on EOF
	ctx, cancelFunc := context.WithCancel(context.Background())
	stdinChan := make(chan string)
	wg := &sync.WaitGroup{}
	go func() {
		<-ctx.Done()
		wg.Wait()
		log.Println("Execution cancelled")
		os.Exit(0)
	}()
	go receive(cancelFunc, stdinChan, wg)

	// keep running the command as long as context is not cancelled
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
		log.Printf("Received result\n%s", string(bytes))
		wg.Done()
	}
}

// receive continually reads from os.Stdin, incrementing the number of results
// we expect to eventually obtain, and sends the read bytes to a byte channel
// for processing
//
// also handles program termination via file EOF
func receive(cancelFunc func(), stdinChan chan<- string, wg *sync.WaitGroup) {
	reader := bufio.NewReader(os.Stdin)
	justPiped := false
	for {
		s, err := reader.ReadString('\n')
		sp := shouldPipe(s)
		if justPiped && sp {
			continue
		}
		if sp {
			wg.Add(1)
		}
		if err == io.EOF {
			log.Println("Received EOF")
			if !justPiped {
				wg.Add(1)
				stdinChan <- "\n"
			}
			cancelFunc()
			return
		}
		stdinChan <- s
		justPiped = sp
	}
}

// pipe continually reads and writes bytes from a channel to a buffered writer,
// eventually piping the buffer (by closing the writer), returning errors if
// any
func pipe(stdinChan <-chan string, cmdStdin io.WriteCloser) error {
	writeCloser := withLog(cmdStdin)
	defer writeCloser.Close()
	for {
		s := <-stdinChan
		if shouldPipe(s) {
			return nil
		}
		_, err := writeCloser.Write([]byte(s))
		if err != nil {
			return err
		}
	}
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

// shouldPipe determines if a buffer should be piped
func shouldPipe(s string) bool {
	return s == "\n"
}
