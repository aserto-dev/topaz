package directory

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/mattn/go-isatty"
)

const (
	objects   = "objects"
	relations = "relations"
	skipped   = "skipped"
)

type Counter struct {
	objects   *Item
	relations *Item
}

// func New() *Counter {
// 	return &Counter{}
// }

type Item struct {
	Name    string
	value   int64
	skipped int64
}

func (c *Item) Incr() *Item {
	atomic.AddInt64(&c.value, 1)
	return c
}

func (c *Item) Skip() *Item {
	atomic.AddInt64(&c.skipped, 1)
	return c
}

func (c *Item) Print(w io.Writer) {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Fprintf(w, "\033[2K\r%15s: %d", c.Name, c.value)
	}
}

func (c *Counter) Objects() *Item {
	if c.objects == nil {
		c.objects = &Item{Name: objects}
	}
	return c.objects
}

func (c *Counter) Relations() *Item {
	if c.relations == nil {
		c.relations = &Item{Name: relations}
	}
	return c.relations
}

const unknownFieldsMsg string = " WARNING data contained unknown fields"

func (c *Counter) Print(w io.Writer) {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Fprintf(w, "\033[2K\r")
	}

	if c.objects != nil {
		msg := ""
		if c.objects.skipped > 0 {
			msg = unknownFieldsMsg
		}
		fmt.Fprintf(w, "%15s %d%s\n", objects, c.objects.value, msg)
	} else {
		fmt.Fprintf(w, "%15s %s\n", objects, skipped)
	}

	if c.relations != nil {
		msg := ""
		if c.relations.skipped > 0 {
			msg = unknownFieldsMsg
		}
		fmt.Fprintf(w, "%15s %d%s\n", relations, c.relations.value, msg)
	} else {
		fmt.Fprintf(w, "%15s %s\n", relations, skipped)
	}
}
