package cmd

import (
	"testing"
)

func Test_Cmd(t *testing.T) {
	echoFactory := NewFactory("echo echo")
	stdout, stderr, err := echoFactory.Run(nil)
	if err != nil || string(stdout) != "echo\n" || string(stderr) != "" {
		t.Error()
	}
}
