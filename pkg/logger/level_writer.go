package logger

import (
	"io"

	"github.com/rs/zerolog"
)

type LevelWriter struct {
	io.Writer

	ErrorWriter io.Writer
}

func (lw *LevelWriter) WriteLevel(l zerolog.Level, p []byte) (int, error) {
	w := lw.Writer
	if l > zerolog.InfoLevel {
		w = lw.ErrorWriter
	}

	return w.Write(p)
}
