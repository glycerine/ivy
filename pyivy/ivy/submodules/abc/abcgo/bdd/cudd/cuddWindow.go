package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

// cuddWindowReorder applies the selected sliding-window variable ordering
// method. The manager implementation lives in core, where it can rebuild the
// canonical tables while preserving referenced node handles.
func cuddWindowReorder(table *DdManager, low, high int, submethod Cudd_ReorderingType) int {
	return core.Cudd_WindowReorder(table, low, high, submethod)
}
