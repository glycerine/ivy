package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

// cuddZddLinearSifting is the internal CUDD entry point retained for callers
// in the compatibility package. The graph transformation and sifting logic
// live beside the manager implementation in core.
func cuddZddLinearSifting(table *DdManager, lower, upper int32) int32 {
	return int32(core.Cudd_zddLinearSifting(table, int(lower), int(upper)))
}
