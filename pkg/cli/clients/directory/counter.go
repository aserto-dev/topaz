package directory

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"

	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	"github.com/mattn/go-isatty"
	"google.golang.org/grpc/codes"
)

const (
	objectsCounter   string = "objects"
	relationsCounter string = "relations"
	skipped          string = "skipped"
)

type Counter struct {
	objects   *Item
	relations *Item
}

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
		c.objects = &Item{Name: objectsCounter}
	}
	return c.objects
}

func (c *Counter) Relations() *Item {
	if c.relations == nil {
		c.relations = &Item{Name: relationsCounter}
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
		fmt.Fprintf(w, "%-9s : %d%s\n", objectsCounter, c.objects.value, msg)
	} else {
		fmt.Fprintf(w, "%-9s : %s\n", objectsCounter, skipped)
	}

	if c.relations != nil {
		msg := ""
		if c.relations.skipped > 0 {
			msg = unknownFieldsMsg
		}
		fmt.Fprintf(w, "%-9s : %d%s\n", relationsCounter, c.relations.value, msg)
	} else {
		fmt.Fprintf(w, "%-9s : %s\n", relationsCounter, skipped)
	}
}

func printStatus(w io.Writer, status *dsi3.ImportStatus) {
	fmt.Fprintf(w, "%-9s : %s - %s (%d)\n",
		"error",
		status.Msg,
		codes.Code(status.Code).String(),
		status.Code)
}

func printCounter(w io.Writer, ctr *dsi3.ImportCounter) {
	fmt.Fprintf(w, "%-9s : %d (set:%d delete:%d error:%d)\n",
		ctr.Type,
		ctr.Recv,
		ctr.Set,
		ctr.Delete,
		ctr.Error,
	)
}
