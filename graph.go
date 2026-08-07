package cli

import (
	"io"
	"fmt"
)

type Node struct {
	Value fmt.Stringer
	Children []*Node
}

type Tree struct {
	Root *Node
}

func (t *Tree) Render(w io.Writer) error {
	layout := createLayout(t.Root)
	_ = layout
	return nil
}

type layoutNode struct {
	*Node
	X int
	Y int
	Width int
	Height int
}

func createLayout(node *Node) []*layoutNode {
	return nil
}