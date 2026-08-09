package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
)

type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
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
	out   io.Writer
	align map[int]Alignment

	WithLineNumbers bool
	WithUnderline   bool
	Separator       rune
}

func NewTableRenderer(w io.Writer) *TableRenderer {
	return &TableRenderer{
		out:       w,
		align:     make(map[int]Alignment),
		Separator: ' ',
	}
}

func (r *TableRenderer) SetAlignment(col int, align Alignment) {
	r.align[col] = align
}

func (r *TableRenderer) alignmentFor(col int, str string) Alignment {
	a, ok := r.align[col]
	if ok {
		return a
	}
	if _, err := strconv.ParseFloat(str, 64); err == nil {
		return AlignRight
	}
	if _, err := strconv.ParseBool(str); err == nil {
		return AlignCenter
	}
	if str == crossMarker || str == checkMarker {
		return AlignCenter
	}
	return AlignLeft
}

func (r *TableRenderer) Empty() {
	fmt.Fprintln(r.out)
}

func (r *TableRenderer) Render(t Table) error {
	if len(t.Rows) == 0 {
		return nil
	}
	if t.Title != "" {
		fmt.Fprintln(r.out, t.Title)
		if r.WithUnderline {
			fmt.Fprintln(r.out, strings.Repeat("-", len(t.Title)))
		}
	}
	wt := tabwriter.NewWriter(r.out, 0, 0, 2, byte(r.Separator), 0)

	if len(t.Headers) > 0 {
		if r.WithLineNumbers {
			fmt.Fprint(wt, "#")
			fmt.Fprint(wt, "\t")
		}
		for i := range t.Headers {
			if i > 0 {
				fmt.Fprint(wt, "\t")
			}
			fmt.Fprint(wt, t.Headers[i])
		}
		fmt.Fprintln(wt)
	}

	for i, row := range t.Rows {
		if r.WithLineNumbers {
			fmt.Fprintf(wt, "%-d", i+1)
			fmt.Fprint(wt, "\t")
		}
		for i := range row {
			if i > 0 {
				fmt.Fprint(wt, "\t")
			}
			var (
				str   = row[i]
				align = r.alignmentFor(i+1, str)
			)
			w := len(str)
			if i < len(t.Headers) {
				w = len(t.Headers[i])
			}
			switch align {
			case AlignRight:
				str = Right(str, w)
			case AlignCenter:
				str = Center(str, len(t.Headers[i]))
			default:
			}
			fmt.Fprint(wt, str)
		}
		fmt.Fprintln(wt)
	}

	return wt.Flush()
}
