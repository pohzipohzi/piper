package piper

import (
	"bufio"
	"bytes"
	"io"
)

// Piper continually reads from a reader into a buffer, eventually piping the
// buffer to a channel as a line-separated string.
type Piper interface {
	Start()
}

func New(r io.Reader, w chan<- string) Piper {
	return &piperImpl{
		r: r,
		w: w,
	}
}

type piperImpl struct {
	r io.Reader
	w chan<- string
}

func (i *piperImpl) Start() {
	scanner := bufio.NewScanner(i.r)
	buf := new(bytes.Buffer)
	for scanner.Scan() {
		b := scanner.Bytes()
		if len(b) == 0 && buf.Len() > 0 {
			i.w <- buf.String()
			buf.Reset()
			continue
		}
		if len(b) > 0 {
			_, _ = buf.Write(b)
			_ = buf.WriteByte('\n')
		}
	}
	if buf.Len() > 0 {
		i.w <- buf.String()
	}
}
