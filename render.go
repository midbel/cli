package cli

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type Table struct {
	Title   string
	Headers []string
	Rows    [][]string
}

type Renderer interface {
	Render(Table) error
}

type TableRenderer struct {
	out io.Writer
}

func NewTableRenderer(w io.Writer) Renderer {
	return &TableRenderer{
		out: w,
	}
}

func (r *TableRenderer) Render(t Table) error {
	if t.Title != "" {
		fmt.Fprintln(r.out, t.Title)
	}
	wt := tabwriter.NewWriter(r.out, 0, 0, 2, ' ', 0)

	for i := range t.Headers {
		if i > 0 {
			fmt.Fprint(wt, "\t")
		}
		fmt.Fprint(wt, t.Headers[i])
	}
	fmt.Fprintln(wt)
	for i := range t.Headers {
		if i > 0 {
			fmt.Fprint(wt, "\t")
		}
		fmt.Fprint(wt, strings.Repeat("-", len(t.Headers[i])))
	}
	fmt.Fprintln(wt)

	for _, row := range t.Rows {
		for i := range row {
			if i > 0 {
				fmt.Fprint(wt, "\t")
			}
			fmt.Fprint(wt, row[i])
		}
		fmt.Fprintln(wt)
	}

	return wt.Flush()
}
