package iostream

import (
	"io"
	"os"
)

type StdIO struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

func (i *StdIO) StdIn() io.Reader {
	return i.in
}

func (i *StdIO) StdOut() io.Writer {
	return i.out
}

func (i *StdIO) StdErr() io.Writer {
	return i.err
}

func DefaultIO() *StdIO {
	return &StdIO{os.Stdin, os.Stdout, os.Stderr}
}
