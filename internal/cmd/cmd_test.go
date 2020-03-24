package cmd

import (
	"bufio"
	"bytes"
	"log"
	"testing"
)

func Test_Cmd(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(bufio.NewWriter(&buf), "", 0)
	echoFactory := NewFactory("echo", []string{"echo"}, WithLog(logger))
	f, err := echoFactory.New()
	if err != nil {
		t.Error()
	}
	res, err := f(nil)
	if err != nil || string(res) != "echo\n" {
		t.Error()
	}
}
