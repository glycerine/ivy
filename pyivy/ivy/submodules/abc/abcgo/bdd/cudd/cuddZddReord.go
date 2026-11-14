package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

// ZDD dynamic ordering is implemented in core, where it can update the unique
// table and preserve manager-owned node handles atomically.
var (
	Cudd_zddShuffleHeap = core.Cudd_zddShuffleHeap
	Cudd_zddReduceHeap  = core.Cudd_zddReduceHeap
)
