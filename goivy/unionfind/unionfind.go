// Package unionfind provides union-find data structures for stratification checks.
// This is a port of Python's ivy_union_find.py and ivy_union_find2.py.
package unionfind

import "sync/atomic"

// ufIDCounter is a global counter for unique node IDs.
var ufIDCounter int64

// UFNode is a union-find node used in stratification checking.
// It supports path compression and can carry auxiliary data.
type UFNode struct {
	parent *UFNode
	rank   int
	ID     int64

	// Var optionally holds the variable this node represents (for strat_map).
	Var interface{}
}

// NewUFNode creates a new UFNode with a unique ID.
func NewUFNode() *UFNode {
	id := atomic.AddInt64(&ufIDCounter, 1) - 1
	n := &UFNode{ID: id}
	n.parent = n
	return n
}

// Find returns the representative (root) of the set containing x.
// Uses path compression for efficiency.
func Find(x *UFNode) *UFNode {
	if x == nil {
		return nil
	}
	// Find root
	root := x
	for root.parent != root {
		root = root.parent
	}
	// Path compression
	for x.parent != root {
		parent := x.parent
		x.parent = root
		x = parent
	}
	return root
}

// Unify merges the sets containing x and y.
// Uses union by rank.
func Unify(x, y *UFNode) {
	if x == nil || y == nil {
		return
	}
	x = Find(x)
	y = Find(y)
	// No early return when x == y — Python's unify() doesn't check,
	// causing a rank increment when both are already in the same set.
	// We must match Python's behavior for identical union-find trees.
	if x.rank < y.rank {
		x, y = y, x
	}
	y.parent = x
	if x.rank == y.rank {
		x.rank++
	}
}

// ResetCounter resets the global ID counter (useful for testing).
func ResetCounter() {
	atomic.StoreInt64(&ufIDCounter, 0)
}
