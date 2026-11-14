package cudd

import core "github.com/glycerine/abcgo/bdd/cudd/core"

// Cudd_zddSymmProfile prints the symmetry classes for the requested ZDD
// levels, in terms of the variable indexes stored in each class.
func Cudd_zddSymmProfile(table *DdManager, lower, upper int32) {
	if table == nil {
		return
	}
	core.Cudd_zddSymmProfileTo(table, int(lower), int(upper), core.Cudd_zddOutput(table))
}

// cuddZddSymmCheck is the source-compatible internal predicate. CUDD returns
// one exactly when adjacent ZDD levels are symmetric.
func cuddZddSymmCheck(table *DdManager, x, y int32) int32 {
	return int32(core.Cudd_zddSymmCheck(table, int(x), int(y)))
}

// cuddZddSymmSifting and cuddZddSymmSiftingConv preserve the C internal
// result convention: zero on failure, otherwise one plus symmetric-variable
// count. The core owns all variable movement and circular group state.
func cuddZddSymmSifting(table *DdManager, lower, upper int32) int32 {
	return int32(core.Cudd_zddSymmSifting(table, int(lower), int(upper), false))
}

func cuddZddSymmSiftingConv(table *DdManager, lower, upper int32) int32 {
	return int32(core.Cudd_zddSymmSifting(table, int(lower), int(upper), true))
}
