package iostream

import (
	"os"
)

type StdIO struct {
	in  *os.File
	out *os.File
	err *os.File
}

func (i *StdIO) StdIn() *os.File {
	return i.in
}

func (i *StdIO) StdOut() *os.File {
	return i.out
}

func (i *StdIO) StdErr() *os.File {
	return i.err
}

func DefaultIO() *StdIO {
	return &StdIO{os.Stdin, os.Stdout, os.Stderr}
}
