package table

import (
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func New(w io.Writer) *tablewriter.Table {
	opts := []tablewriter.Option{
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.Border{
				Top:    tw.BorderNone.Top,
				Bottom: tw.BorderNone.Bottom,
				Left:   tw.BorderNone.Left,
				Right:  tw.BorderNone.Right,
			},
		}),
		tablewriter.WithAlignment(tw.Alignment{tw.AlignLeft}),
		tablewriter.WithPadding(tw.PaddingDefault),
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.BorderNone,
			Symbols: tw.NewSymbols(tw.StyleASCII),
			Settings: tw.Settings{
				Separators: tw.SeparatorsNone,
				Lines:      tw.LinesNone,
			},
			Streaming: false,
		})),
	}

	return tablewriter.NewTable(w, opts...)
}
