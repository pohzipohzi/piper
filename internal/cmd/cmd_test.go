package cmd

import (
	"testing"
)

func Test_Cmd(t *testing.T) {
	echoFactory := NewFactory("echo", []string{"echo"})
	f, err := echoFactory.New()
	if err != nil {
		t.Error()
	}
	res, err := f(nil)
	if err != nil || string(res) != "echo\n" {
		t.Error()
	}
}
