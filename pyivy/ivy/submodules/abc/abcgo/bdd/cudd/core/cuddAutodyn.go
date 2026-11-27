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
		dd.liveBDDKeys() >= int(dd.nextDyn) {
		// reorder here, with the recursion's pending results referenced as
		// in C, then unwind and restart the operation
		pending := append([]*DdNode(nil), dd.inflight...)
		for _, n := range pending {
			Cudd_Ref(n)
		}
		dd.inAutoReorder = true
		Cudd_ReduceHeap(dd, dd.autoMethod, 10)
		dd.inAutoReorder = false
		for _, n := range pending {
			Cudd_RecursiveDeref(dd, n)
		}
		if dd.liveBDDKeys() >= int(dd.nextDyn) {
			dd.updateNextDyn() // a skipped reordering must not loop
		}
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
		dd.inflight = dd.inflight[:0]
		r, restart := dd.guardedOp(op)
		if !restart {
			return r
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
	// nextDyn = (keys - constants.keys + 1) * DD_DYN_RATIO + constants.keys
	next := uint32((dd.unique.count+dd.addUnique.count+1)*DD_DYN_RATIO + len(dd.constants))
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
