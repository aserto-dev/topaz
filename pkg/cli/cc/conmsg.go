package cc

import (
	"io"
	"strings"

	"github.com/fatih/color"
)

// ConMsg - console message, send either StdErr or StdOut.
type ConMsg struct {
	out   io.Writer
	color *color.Color
}

// Con() - console message send to StdErr, default no-color.
func (c *CommonCtx) Con() *ConMsg {
	return &ConMsg{
		out:   color.Error,
		color: color.New(),
	}
}

// Out() - console output message send to StdOut, default no-color.
func (c *CommonCtx) Out() *ConMsg {
	return &ConMsg{
		out:   color.Output,
		color: color.New(),
	}
}

// Info() - info console message (green).
func (cm *ConMsg) Info() *ConMsg {
	cm.color.Add(color.FgGreen)
	return cm
}

// Warn() - warning console message (yellow).
func (cm *ConMsg) Warn() *ConMsg {
	cm.color.Add(color.FgYellow)
	return cm
}

// Error() - error console message (red).
func (cm *ConMsg) Error() *ConMsg {
	cm.color.Add(color.FgRed)
	return cm
}

// Msg() - sends the con|out message, by default adds a CrLr when not present.
func (cm *ConMsg) Msg(message string, args ...interface{}) {
	color.NoColor = NoColor()

	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}

	if len(args) == 0 {
		cm.color.Fprint(color.Error, message)
		return
	}

	cm.color.Fprintf(color.Error, message, args...)
}
