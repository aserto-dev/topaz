package table

import (
	"io"

	tw "github.com/olekukonko/tablewriter"
)

type TableWriter struct {
	w     io.Writer
	table *table
}

type table struct {
	columns    [][]string
	data       [][][]string
	noAutoWrap bool
}

func New(w io.Writer) *TableWriter {
	return &TableWriter{
		w:     w,
		table: &table{},
	}
}

// WithColumns prints a new table.
func (u *TableWriter) WithColumns(columns ...string) *TableWriter {
	u.table.columns = append(u.table.columns, columns)
	u.table.data = append(u.table.data, [][]string{})
	return u
}

// WithRow adds a row in the latest table.
func (u *TableWriter) WithRow(values ...string) *TableWriter {
	if len(u.table.columns) < 1 {
		return u.WithColumns(make([]string, len(values))...).WithRow(values...)
	}

	u.table.data[len(u.table.data)-1] = append(u.table.data[len(u.table.data)-1], values)

	return u
}

func (u *TableWriter) WithTableNoAutoWrapText() *TableWriter {
	u.table.noAutoWrap = true
	return u
}

func (u *TableWriter) Do() {
	for idx, headers := range u.table.columns {
		table := tw.NewWriter(u.w)
		table.SetHeader(headers)
		table.SetBorders(tw.Border{Left: false, Top: false, Right: false, Bottom: false})
		table.SetBorder(false)
		table.SetAlignment(tw.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetHeaderLine(false)
		table.SetHeaderAlignment(tw.ALIGN_LEFT)
		table.SetColumnSeparator("")
		table.SetAutoWrapText(!u.table.noAutoWrap)

		if idx < len(u.table.data) {
			table.AppendBulk(u.table.data[idx])
		}

		table.Render()
	}
}
