package cmd

import (
	"io"
	"io/ioutil"
	"log"
	"os/exec"
)

type Factory interface {
	New() (func([]byte) ([]byte, error), error)
}

type Opt func(io.WriteCloser) io.WriteCloser

func NewFactory(name string, args []string, opts ...Opt) Factory {
	return &cmdFactoryImpl{
		name: name,
		args: args,
		opts: opts,
	}
}

type cmdFactoryImpl struct {
	name string
	args []string
	opts []Opt
}

func (i *cmdFactoryImpl) New() (func(b []byte) ([]byte, error), error) {
	cmd := exec.Command(i.name, i.args...)
	wc, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	for _, o := range i.opts {
		wc = o(wc)
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

// WithLog sets up an intermediary pipe that logs data passing through a writer
func WithLog(stderrLogger *log.Logger) Opt {
	return func(next io.WriteCloser) io.WriteCloser {
		r, w := io.Pipe()
		go func() {
			defer next.Close()
			b, err := ioutil.ReadAll(r)
			if err != nil {
				stderrLogger.Println("Failed to read in WithLog pipe:", err)
				return
			}
			stderrLogger.Printf("Received input\n%s", string(b))
			_, err = next.Write(b)
			if err != nil {
				stderrLogger.Println("Failed to write in WithLog pipe:", err)
			}
		}()
		return w
	}
}
