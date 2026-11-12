package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

// cuddSetInteract marks two distinct variables as interacting.
func cuddSetInteract(table *DdManager, x, y int32) {
	core.Cudd_SetInteraction(table, int(x), int(y))
}

// cuddTestInteract reports whether two variables share a DD support.
func cuddTestInteract(table *DdManager, x, y int32) int32 {
	if core.Cudd_TestInteraction(table, int(x), int(y)) {
		return 1
	}
	return 0
}

// cuddInitInteract computes the matrix used by linear reordering.
func cuddInitInteract(table *DdManager) int32 {
	if table == nil {
		return 0
	}
	core.Cudd_ComputeInteractions(table)
	return 1
}
