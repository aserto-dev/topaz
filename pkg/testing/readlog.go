package testing

import (
	"bufio"
	"io"
	"time"
)

// LogReadLine reads and returns a single line from the given file.
// If a given timeout has passed, an error is returned.
func LogReadLine(reader *bufio.Reader, timeout time.Duration) (string, error) {
	s := make(chan string)
	e := make(chan error)

	go func() {
		b, _, err := reader.ReadLine()

		line := string(b)

		if err != nil {
			e <- err
		} else {
			s <- line
		}
		close(s)
		close(e)
	}()

	select {
	case line := <-s:
		return line, nil
	case err := <-e:
		if err == io.EOF {
			return "", nil
		}
		return "", err
	case <-time.After(timeout):
		return "", nil
	}
}
