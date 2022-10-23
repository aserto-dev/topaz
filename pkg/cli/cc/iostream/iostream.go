package iostream

import (
	"bytes"
	"io"
	"os"

	"github.com/aserto-dev/clui"
)

type IO interface {
	Input() io.Reader
	Output() io.Writer
	Error() io.Writer
}

type StdIO struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

func (i *StdIO) Input() io.Reader {
	return i.In
}

func (i *StdIO) Output() io.Writer {
	return i.Out
}

func (i *StdIO) Error() io.Writer {
	return i.Err
}

type BufferIO struct {
	In  *bytes.Buffer
	Out *bytes.Buffer
	Err *bytes.Buffer
}

func (i *BufferIO) Input() io.Reader {
	return i.In
}

func (i *BufferIO) Output() io.Writer {
	return i.Out
}

func (i *BufferIO) Error() io.Writer {
	return i.Err
}

func DefaultIO() *StdIO {
	return &StdIO{os.Stdin, os.Stdout, os.Stderr}
}

func BytesIO() *BufferIO {
	return &BufferIO{&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}}
}

func NewUI(ios IO) *clui.UI {
	return clui.NewUIWithOutputErrorAndInput(ios.Output(), ios.Error(), ios.Input())
}
