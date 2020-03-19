package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("No command provided")
		return
	}

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Println("Received signal:", sig)
		os.Exit(0)
	}()

	stdinChan := make(chan string, 1)
	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for scanner.Scan() {
			stdinChan <- scanner.Text()
		}
	}()

	for {
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmdStdin, err := cmd.StdinPipe()
		if err != nil {
			log.Println("Error obtaining stdin:", err)
			continue
		}
		log.Println("Awaiting input")
		err = bufferedPipe(stdinChan, cmdStdin, func(s string) bool { return s == "" })
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
	}
}

func bufferedPipe(stdinChan <-chan string, cmdStdin io.WriteCloser, shouldPipe func(string) bool) error {
	writeCloser := withLog(cmdStdin)
	defer writeCloser.Close()
	for {
		s := <-stdinChan
		if shouldPipe(s) {
			return nil
		}
		_, err := writeCloser.Write([]byte(s + "\n"))
		if err != nil {
			return err
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
