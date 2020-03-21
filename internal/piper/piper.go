package piper

import (
	"bufio"
	"io"
)

// Piper continually reads from a reader into a buffer, sending the buffer as a
// line-separated string to a channel upon receiving an empty line
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
	toPipe := ""
	for scanner.Scan() {
		s := scanner.Text()
		if s == "" && len(toPipe) > 0 {
			i.w <- toPipe
			toPipe = ""
			continue
		}
		if s != "" {
			toPipe += s + "\n"
		}
	}
	if len(toPipe) > 0 {
		i.w <- toPipe
	}
}
