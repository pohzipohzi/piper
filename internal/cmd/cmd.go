package cmd

import (
	"bytes"
	"os/exec"
	"strings"
)

type Factory interface {
	Run(input []byte) (stdout []byte, stderr []byte, err error)
}

func NewFactory(s string) Factory {
	if s == "" {
		return nil
	}
	args := strings.Split(s, " ")
	return &factoryImpl{
		name: args[0],
		args: args[1:],
	}
}

type factoryImpl struct {
	name string
	args []string
}

func (i *factoryImpl) Run(b []byte) ([]byte, []byte, error) {
	cmd := exec.Command(i.name, i.args...)
	wc, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	_, err = wc.Write(b)
	if err != nil {
		return nil, nil, err
	}
	err = wc.Close()
	if err != nil {
		return nil, nil, err
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}
