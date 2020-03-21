package cmd

import (
	"io"
	"os/exec"
)

type Factory interface {
	New() (func([]byte) ([]byte, error), error)
}

func NewFactory(name string, args []string, opts []func(io.WriteCloser) io.WriteCloser) Factory {
	return &cmdFactoryImpl{
		name: name,
		args: args,
		opts: opts,
	}
}

type cmdFactoryImpl struct {
	name string
	args []string
	opts []func(io.WriteCloser) io.WriteCloser
}

func (i *cmdFactoryImpl) New() (func(b []byte) ([]byte, error), error) {
	cmd := exec.Command(i.name, i.args...)
	wc, err := cmd.StdinPipe()
	for _, o := range i.opts {
		wc = o(wc)
	}
	if err != nil {
		return nil, err
	}
	return func(b []byte) ([]byte, error) {
		_, err := wc.Write(b)
		if err != nil {
			return nil, err
		}
		err = wc.Close()
		if err != nil {
			return nil, err
		}
		return cmd.Output()
	}, nil
}
