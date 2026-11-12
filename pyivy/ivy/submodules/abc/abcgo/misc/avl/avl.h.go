// Package avl implements the ordered duplicate-key table from ABC's misc/avl.
// Keys and values are opaque Go values; callers provide their ordering.
package avl

const (
	AVL_FORWARD = iota
	AVL_BACKWARD
)

type Compare func(a, b any) int

type node struct {
	left, right *node
	key, value  any
	height      int
}

type Tree struct {
	root    *node
	compare Compare
	count   int
}

type Generator struct {
	nodes []*node
	next  int
}

// New returns an empty tree. Equal keys are allowed and retained separately.
func New(compare Compare) *Tree {
	if compare == nil {
		panic("avl: nil comparator")
	}
	return &Tree{compare: compare}
}

// AvlTree and AvlGenerator retain convenient names for ported ABC consumers.
type AvlTree = Tree
type AvlGenerator = Generator
