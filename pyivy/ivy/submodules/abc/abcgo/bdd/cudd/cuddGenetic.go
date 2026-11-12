package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

func cuddGa(dd *DdManager, lower, upper int32) int32 {
	if core.Cudd_GeneticReorder(dd, int(lower), int(upper)) {
		return 1
	}
	return 0
}
