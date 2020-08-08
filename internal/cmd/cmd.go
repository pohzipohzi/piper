package cmd

import (
	"bytes"
	"os/exec"
	"strings"
)

type Factory interface {
	New() (func([]byte) ([]byte, []byte, error), error)
	String() string
}

func NewFactory(s string) Factory {
	if s == "" {
		return nil
	}
	args := strings.Split(s, " ")
	return &factoryImpl{
		s:    s,
		name: args[0],
		args: args[1:],
	}
}

type factoryImpl struct {
	s    string
	name string
	args []string
}

func (i *factoryImpl) New() (func(b []byte) ([]byte, []byte, error), error) {
	cmd := exec.Command(i.name, i.args...)
	wc, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	return func(b []byte) ([]byte, []byte, error) {
		_, err := wc.Write(b)
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
	}, nil
}

func (i *factoryImpl) String() string {
	return i.s
}
