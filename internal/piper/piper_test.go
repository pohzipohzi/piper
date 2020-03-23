package piper

import (
	"bytes"
	"testing"
)

func Test_New(t *testing.T) {
	t.Parallel()
	if New(nil, nil) == nil {
		t.Error()
	}
}

func Test_piperImpl_Start(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		input      []byte
		outputChan chan string
		expect     []string
	}{
		{
			input:      []byte("\r\n"),
			outputChan: make(chan string),
			expect:     []string{},
		},
		{
			input:      []byte("\n"),
			outputChan: make(chan string),
			expect:     []string{},
		},
		{
			input:      []byte(""),
			outputChan: make(chan string),
			expect:     []string{},
		},
		{
			input:      []byte("1\r\n"),
			outputChan: make(chan string),
			expect:     []string{"1\n"},
		},
		{
			input:      []byte("1\n"),
			outputChan: make(chan string),
			expect:     []string{"1\n"},
		},
		{
			input:      []byte("1"),
			outputChan: make(chan string),
			expect:     []string{"1\n"},
		},
		{
			input:      []byte("\n1"),
			outputChan: make(chan string),
			expect:     []string{"1\n"},
		},
		{
			input:      []byte("\r\n1"),
			outputChan: make(chan string),
			expect:     []string{"1\n"},
		},
		{
			input:      []byte("\n1\n"),
			outputChan: make(chan string),
			expect:     []string{"1\n"},
		},
		{
			input:      []byte("\r\n1\r\n"),
			outputChan: make(chan string),
			expect:     []string{"1\n"},
		},
		{
			input:      []byte("\r\n1\n"),
			outputChan: make(chan string),
			expect:     []string{"1\n"},
		},
		{
			input:      []byte("\n1\r\n"),
			outputChan: make(chan string),
			expect:     []string{"1\n"},
		},
		{
			input:      []byte("1\n2"),
			outputChan: make(chan string),
			expect:     []string{"1\n2\n"},
		},
		{
			input:      []byte("1\n\n2"),
			outputChan: make(chan string),
			expect:     []string{"1\n", "2\n"},
		},
		{
			input:      []byte("1\n\n\n2"),
			outputChan: make(chan string),
			expect:     []string{"1\n", "2\n"},
		},
		{
			input:      []byte("1\n\n2\n3"),
			outputChan: make(chan string),
			expect:     []string{"1\n", "2\n3\n"},
		},
		{
			input:      []byte("1\n2\n\n3"),
			outputChan: make(chan string),
			expect:     []string{"1\n2\n", "3\n"},
		},
		{
			input:      []byte("1\n\n2\n\n3"),
			outputChan: make(chan string),
			expect:     []string{"1\n", "2\n", "3\n"},
		},
		{
			input:      []byte("1\n\n\n2\n\n\n3"),
			outputChan: make(chan string),
			expect:     []string{"1\n", "2\n", "3\n"},
		},
	} {
		reader := bytes.NewReader(test.input)
		pi := &piperImpl{
			r: reader,
			w: test.outputChan,
		}
		go pi.Start()
		for _, v := range test.expect {
			recv := <-test.outputChan
			if v != recv {
				t.Error()
			}
		}
	}
}
