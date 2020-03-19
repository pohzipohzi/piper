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

	ctx, cancelFunc := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Println("Received signal:", sig)
		cancelFunc()
	}()

	// listen to os.Stdin and pipe bytes to a channel
	// waitgroup is the remaining number of results we expect to obtain
	stdinChan := make(chan []byte)
	wg := sync.WaitGroup{}
	reader := bufio.NewReader(os.Stdin)
	go func() {
		justPiped := false
		for {
			bytes, err := reader.ReadBytes('\n')
			sp := shouldPipe(bytes)
			if justPiped && sp {
				continue
			}
			if sp {
				wg.Add(1)
			}
			stdinChan <- bytes
			if err == io.EOF {
				log.Println("Received EOF")
				if !justPiped {
					stdinChan <- []byte{'\n'}
					wg.Add(1)
				}
				close(stdinChan)
				cancelFunc()
				return
			}
			justPiped = sp
		}
	}()

	// run the command each time we have output to pipe
	for {
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmdStdin, err := cmd.StdinPipe()
		if err != nil {
			log.Println("Error obtaining stdin:", err)
			continue
		}
		log.Println("Awaiting input")

		errChan := make(chan error)
		go func() {
			writeCloser := withLog(cmdStdin)
			defer writeCloser.Close()
			for {
				s := <-stdinChan
				if shouldPipe(s) {
					errChan <- nil
					return
				}
				_, err := writeCloser.Write(s)
				if err != nil {
					errChan <- err
					return
				}
			}
		}()

		select {
		case <-ctx.Done():
			wg.Wait()
			log.Println("Execution cancelled")
			return
		case err := <-errChan:
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
}

func withLog(to io.WriteCloser) io.WriteCloser {
	r, w := io.Pipe()
	go func() {
		defer to.Close()
		p, err := ioutil.ReadAll(r)
		if err != nil {
			log.Println("Failed to read in intermediate pipe:", err)
			return
		}
		log.Printf("Received input\n%s", string(p))
		_, err = to.Write(p)
		if err != nil {
			log.Println("Failed to write in intermediate pipe:", err)
		}
	}()
	return w
}

func shouldPipe(b []byte) bool {
	return len(b) == 1 && b[0] == '\n'
}
