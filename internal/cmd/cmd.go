package cmd

import (
	"bytes"
	"os/exec"
)

type Factory interface {
	New() (func([]byte) ([]byte, []byte, error), error)
}

func NewFactory(name string, args []string) Factory {
	return &cmdFactoryImpl{
		name: name,
		args: args,
	}
}

type cmdFactoryImpl struct {
	name string
	args []string
}

func (i *cmdFactoryImpl) New() (func(b []byte) ([]byte, []byte, error), error) {
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
