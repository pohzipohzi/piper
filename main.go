package main

import (
	"bufio"
	"io"
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

	stdinChan := make(chan string, 1)
	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for scanner.Scan() {
			stdinChan <- scanner.Text()
		}
	}()

	for {
		log.Println("Initializing new command")
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmdStdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatalln("Error obtaining stdin:", err)
		}
		runWrite(sigChan, stdinChan, cmdStdin)
		log.Println("Running command")
		bytes, err := cmd.Output()
		if err != nil {
			log.Fatalln("Error running command:", err)
		}
		log.Printf("Received result:\n%s\n", string(bytes))
	}
}

func runWrite(sigChan <-chan os.Signal, stdinChan <-chan string, cmdStdin io.WriteCloser) {
	defer cmdStdin.Close()
	writer := bufio.NewWriter(cmdStdin)
	defer writer.Flush()
	for {
		select {
		case sig := <-sigChan:
			log.Println("Received signal:", sig)
			os.Exit(0)
		case s := <-stdinChan:
			writer.WriteString(s + "\n")
			if s == "" {
				return
			}
		}
	}
}
