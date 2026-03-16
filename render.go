package cli

import (
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"
)

type Align int

const (
	AlignLeft Align = iota
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
	align map[int]Align
}

func NewTableRenderer(w io.Writer) *TableRenderer {
	return &TableRenderer{
		out:   w,
		align: make(map[int]Align),
	}
}

func (r *TableRenderer) SetAlignment(col int, align Align) {
	r.align[col] = align
}

func (r *TableRenderer) alignmentFor(col int, str string) Align {
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
	}
	wt := tabwriter.NewWriter(r.out, 0, 0, 2, ' ', 0)

	if len(t.Headers) > 0 {
		for i := range t.Headers {
			if i > 0 {
				fmt.Fprint(wt, "\t")
			}
			fmt.Fprint(wt, t.Headers[i])
		}
		fmt.Fprintln(wt)
	}

	for _, row := range t.Rows {
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
