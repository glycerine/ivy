package core

// Automatic dynamic reordering (cuddUniqueInter, Cudd_ReduceHeap): when a
// new BDD node brings the live keys to nextDyn inside an operation, the
// operation is abandoned, the heap is reordered with the automatic method,
// and the operation restarts from the top, as C returns NULL with
// reordered = 1 and its callers loop.

// DD_FIRST_REORDER is the first auto-reordering threshold; DD_DYN_RATIO
// scales the heap size into the next one.
const (
	DD_FIRST_REORDER = 4004
	DD_DYN_RATIO     = 2
)

type bddAutoReorder struct{}

// autoReorderCheck is called when a BDD node is created.
func (dd *DdManager) autoReorderCheck() {
	if dd.autoDyn && dd.autoGuard > 0 && dd.autoDynSafe > 0 && !dd.inAutoReorder &&
		dd.unique.count+dd.addUnique.count-dd.dead >= int(dd.nextDyn) {
		panic(bddAutoReorder{})
	}
}

// withAutoReorder runs a top-level BDD operation, restarting it after an
// automatic reordering; nested operations run inside the outer guard.
func (dd *DdManager) withAutoReorder(op func() *DdNode) *DdNode {
	if !dd.autoDyn || dd.autoDynSafe == 0 || dd.autoGuard > 0 || dd.inAutoReorder {
		return op()
	}
	for {
		r, restart := dd.guardedOp(op)
		if !restart {
			return r
		}
		dd.inAutoReorder = true
		Cudd_ReduceHeap(dd, dd.autoMethod, 10)
		dd.inAutoReorder = false
		// a reordering skipped below minsize leaves the threshold; raise it
		// so the restarted operation does not trigger again at once
		if dd.unique.count+dd.addUnique.count-dd.dead >= int(dd.nextDyn) {
			dd.updateNextDyn()
		}
	}
}

func (dd *DdManager) guardedOp(op func() *DdNode) (r *DdNode, restart bool) {
	dd.autoGuard++
	defer func() {
		dd.autoGuard--
		if x := recover(); x != nil {
			if _, ok := x.(bddAutoReorder); ok {
				r, restart = nil, true
				return
			}
			panic(x)
		}
	}()
	return op(), false
}

// updateNextDyn is Cudd_ReduceHeap's threshold update after a reordering.
func (dd *DdManager) updateNextDyn() {
	next := uint32((dd.unique.count + dd.addUnique.count + 1) * DD_DYN_RATIO)
	if dd.reorderings < 20 || next > dd.nextDyn {
		dd.nextDyn = next
	} else {
		dd.nextDyn += 20
	}
}

// Cudd_AutodynSafeBegin opens a region whose callers keep every BDD they
// hold referenced (Cudd_Ref), as C code does; only there does automatic
// reordering trigger. Go ports elsewhere may hold unreferenced results
// between calls, which a reordering would invalidate. Regions nest.
func Cudd_AutodynSafeBegin(dd *DdManager) { dd.autoDynSafe++ }

// Cudd_AutodynSafeEnd closes a region opened by Cudd_AutodynSafeBegin.
func Cudd_AutodynSafeEnd(dd *DdManager) {
	if dd.autoDynSafe > 0 {
		dd.autoDynSafe--
	}
}
