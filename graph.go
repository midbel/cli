package cli

import (
	"bufio"
	"io"
)

func RenderHorizontal(w io.Writer, tree *Tree, opts *TreeRenderOptions) error {
	if opts == nil {
		opts = defaultTreeRenderOptions.clone()
	}
	var (
		maker       = makeLayout(opts.VerticalGap, opts.Position)
		layout      = maker.Make(tree.Root)
		borderWidth int
	)
	if opts.Border {
		borderWidth++
	}
	if opts.Width == 0 || opts.Height == 0 {
		for _, x := range layout {
			n := len(x.Value) + (2*opts.Padding)
			opts.Width = max(opts.Width, n)
		}
		opts.Width = (opts.Width + opts.HorizontalGap) * maker.HorizontalDepth()
		opts.Height = maker.VerticalDepth() * opts.VerticalGap
	}

	var (
		sWidth  = (opts.Width / maker.HorizontalDepth())
		sHeight = (opts.Height / maker.VerticalDepth())
		vOffset = sHeight / 2
	)
	if w := sWidth * maker.HorizontalDepth(); w != opts.Width {
		opts.Width = w
	}
	if h := sHeight * maker.VerticalDepth(); h != opts.Height {
		opts.Height = h
	}
	for _, x := range layout {
		x.X = (x.X * opts.Width) / maker.HorizontalDepth()
		x.Y = (x.Y * opts.Height) / maker.VerticalDepth()
	}

	var (
		grid    = makeCanvas(opts.Width, opts.Height, opts.Border)
		connect = makeConnector(sWidth)
	)
	// draw horizontal connectors
	for _, x := range layout {
		var (
			tmp    = make([]byte, len(connect))
			size   = len(x.Value) + (2 * opts.Padding)
			offset = getOffsetX(opts.Align, sWidth, size)
			start  = x.X + borderWidth
		)
		copy(tmp, connect)
		if x.Leaf() {
			tmp = tmp[:sWidth-offset-size]
		} else if x.Root {
			if len(x.Children) == 1 {
				tmp[len(tmp)-1] = horizontalBarAscii
			}
			tmp = tmp[offset:]
			start += offset
		}
		grid.Put(start, x.Y+borderWidth+vOffset, tmp)
	}
	// draw vertical connectors
	for _, x := range layout {
		if x.Leaf() {
			continue
		}
		for i := 0; i < len(x.Children)-1; i++ {
			for j := x.Children[i].Y + vOffset + borderWidth; j < x.Children[i+1].Y+vOffset+borderWidth; j++ {
				at := x.X + borderWidth + sWidth
				if grid.GetByte(at, j) == connectBarAscii {
					continue
				}
				grid.PutByte(at, j, verticalBarAscii)
			}
		}
	}
	// draw values
	for _, x := range layout {
		var (
			value  = x.Get(opts.Padding)
			size   = len(value)
			offset = getOffsetX(opts.Align, sWidth, size)
			start  = x.X + borderWidth + offset
		)
		grid.Put(start, x.Y+borderWidth+vOffset, value)
	}
	return grid.Render(w)
}

func RenderVertical(w io.Writer, tree *Tree, opts *TreeRenderOptions) error {
	if opts == nil {
		opts = defaultTreeRenderOptions.clone()
	}
	return nil
}

func RenderCompact(w io.Writer, tree *Tree, opts *TreeRenderOptions) error {
	if opts == nil {
		opts = defaultTreeRenderOptions.clone()
	}
	return nil
}

type ParentPosition uint8

const (
	ParentAlignCenter ParentPosition = iota
	ParentAlignFirst
	ParentAlignLast
)

type ConnectorStyle uint8

const (
	ConnectorAscii ConnectorStyle = iota
	ConnectorUnicode
)

type TreeRenderOptions struct {
	Width         int
	Height        int
	HorizontalGap int
	VerticalGap   int
	Padding       int
	Connector     ConnectorStyle
	Align         Alignment
	Position      ParentPosition
	Border        bool
}

var defaultTreeRenderOptions = &TreeRenderOptions{
	VerticalGap:   DefaultVerticalGapSize,
	HorizontalGap: DefaultHorizontalGapSize,
	Padding:       1,
	Connector:     ConnectorAscii,
	Align:         AlignCenter,
	Position:      ParentAlignCenter,
	Border:        true,
}

func (t *TreeRenderOptions) clone() *TreeRenderOptions {
	x := *t
	return &x
}

type Node struct {
	Value    string
	Children []*Node
}

func NewNode(value string) *Node {
	return &Node{
		Value: value,
	}
}

func (n *Node) Leaf() bool {
	return len(n.Children) == 0
}

type Tree struct {
	Root *Node
}

func NewTree(node *Node) *Tree {
	return &Tree{
		Root: node,
	}
}

type layoutNode struct {
	*Node
	Root     bool
	X        int
	Y        int
	Children []*layoutNode
}

func (n *layoutNode) Get(padding int) []byte {
	var (
		value = []byte(n.Value)
		size  = len(value)
	)
	if padding > 0 {
		size += 2 * padding
		tmp := make([]byte, size)
		for i := range tmp {
			tmp[i] = ' '
		}
		copy(tmp[padding:], value)
		value = tmp
	}
	return value
}

const (
	DefaultVerticalGapSize   = 2
	DefaultHorizontalGapSize = 5
)

const (
	connectBarAscii    = '+'
	verticalBarAscii   = '|'
	horizontalBarAscii = '-'
	// verticalBarUnicode = '\u0007C'
	// horizontalBarUnicode = '\u02015'
)

func makeConnector(width int) []byte {
	connect := make([]byte, width+1)
	for i := range connect {
		connect[i] = horizontalBarAscii
		if i == 0 || i == len(connect)-1 {
			connect[i] = connectBarAscii
		}
	}
	return connect
}

type canvas struct {
	Width  int
	Height int
	cells  [][]byte
}

func (c *canvas) GetByte(x, y int) byte {
	return c.cells[y][x]
}

func (c *canvas) PutByte(x, y int, char byte) {
	c.cells[y][x] = char
}

func (c *canvas) Put(x, y int, chars []byte) {
	copy(c.cells[y][x:], chars)
}

func (c *canvas) Render(w io.Writer) error {
	ws := bufio.NewWriter(w)
	defer ws.Flush()
	for i := range c.cells {
		_, err := ws.Write(c.cells[i])
		if err != nil {
			return err
		}
		ws.WriteByte('\n')
	}
	return nil
}

func makeCanvas(width, height int, border bool) *canvas {
	if border {
		width += 2
		height += 2
	}

	var (
		cells  = make([][]byte, height)
		blank = make([]byte, width)
	)
	for i := range blank {
		blank[i] = ' '
	}
	if border {
		blank[0] = verticalBarAscii
		blank[width-1] = verticalBarAscii
	}
	for i := range cells {
		cells[i] = make([]byte, width)
		if border && (i == 0 || i == height-1) {
			for j := range cells[i] {
				cells[i][j] = horizontalBarAscii
				if j == 0 || j == width-1 {
					cells[i][j] = connectBarAscii
				}
			}
		} else {
			copy(cells[i], blank)
		}
	}
	c := &canvas{
		cells: cells,
		Width: width,
		Height: height,
	}
	return c
}

type layoutMaker struct {
	nextLeafPosition int
	gapSize          int
	maxDepth         int
	align            ParentPosition
}

func makeLayout(gap int, align ParentPosition) *layoutMaker {
	return &layoutMaker{
		gapSize: gap,
		align:   align,
	}
}

func (m *layoutMaker) Make(node *Node) []*layoutNode {
	return m.makeLayout(node, 0)
}

func (m *layoutMaker) VerticalDepth() int {
	return m.nextLeafPosition
}

func (m *layoutMaker) HorizontalDepth() int {
	return m.maxDepth + 1
}

func (m *layoutMaker) makeLayout(node *Node, depth int) []*layoutNode {
	var (
		all []*layoutNode
		sub = layoutNode{
			Node: node,
			Root: depth == 0,
			X:    depth,
		}
	)
	for _, c := range node.Children {
		all = append(all, m.makeLayout(c, depth+1)...)
		for i := range all {
			if all[i].X != depth+1 {
				continue
			}
			sub.Children = append(sub.Children, all[i])
		}
	}
	if node.Leaf() {
		sub.Y = m.nextLeafPosition
		m.nextLeafPosition += m.gapSize
	} else {
		if m.align == ParentAlignFirst {
			sub.Y = all[0].Y
		} else if m.align == ParentAlignLast {
			sub.Y = all[len(all)-1].Y
		} else {
			var sum int
			for i := range all {
				sum += all[i].Y
			}
			sub.Y = sum / (len(all))
		}
	}
	m.maxDepth = max(depth, m.maxDepth)
	return append(all, &sub)
}

func getOffsetX(align Alignment, width, size int) int {
	switch align {
	case AlignLeft:
		return 0
	case AlignRight:
		return width - size
	default:
		return (width - size) / 2
	}
}
