package testing

import (
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
)

type LogDebugger struct {
	t          *testing.T
	outLogPath string
	outLog     *os.File
	buffer     *strings.Builder
	locker     *sync.Mutex
}

func NewLogDebugger(t *testing.T, logName string) *LogDebugger {
	outLog, err := ioutil.TempFile("", logName+"-test-log-*.log")
	if err != nil {
		t.Error(err)
	}
	outLogPath := outLog.Name()

	t.Logf(">> Log output for %s: %s\n", logName, outLogPath)

	return &LogDebugger{
		t:          t,
		outLog:     outLog,
		outLogPath: outLogPath,
		buffer:     &strings.Builder{},
		locker:     &sync.Mutex{},
	}
}

func (d *LogDebugger) Write(p []byte) (int, error) {
	d.locker.Lock()
	defer d.locker.Unlock()
	_, err := d.buffer.Write(p)
	if err != nil {
		d.t.Error(err)
	}

	_, err = d.outLog.Write(p)
	if err != nil {
		d.t.Error(err)
	}

	_, err = os.Stdout.Write(p)
	if err != nil {
		d.t.Error(err)
	}

	return len(p), nil
}

func (d *LogDebugger) Contains(message string) bool {
	return strings.Contains(d.buffer.String(), message)
}
