package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

// Reference operations delegate to the manager implementation in core. The
// Go manager keeps nodes alive through its canonical tables, so dereferencing
// adjusts the reference count without the C implementation's explicit GC stack.
func Cudd_Ref(n *DdNode) { core.Cudd_Ref(n) }

func Cudd_Deref(n *DdNode) { core.Cudd_Deref(n) }

func Cudd_RecursiveDeref(dd *DdManager, n *DdNode) {
	core.Cudd_RecursiveDeref(dd, n)
}

func Cudd_IterDerefBdd(dd *DdManager, n *DdNode) {
	core.Cudd_RecursiveDeref(dd, n)
}

func Cudd_RecursiveDerefZdd(dd *DdManager, n *DdNode) {
	core.Cudd_RecursiveDeref(dd, n)
}
