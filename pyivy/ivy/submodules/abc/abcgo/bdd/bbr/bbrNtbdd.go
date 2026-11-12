// Global AIG BDD construction, following bbrNtbdd.c (Alan Mishchenko).
package bbr

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

// Bbr_NodeGlobalBdds_rec computes a regular object's global function. Progress
// is a callback invoked with the number of newly constructed internal nodes.
func Bbr_NodeGlobalBdds_rec(dd *DdManager, p *Aig_Man_t, e aig.Edge, limit int, drop bool, progress func(int), counter *int, verbose bool) *DdNode {
	if dd == nil || p == nil || !p.ValidEdge(e) || e.Complemented() || counter == nil {
		return nil
	}
	if int(cudd.Cudd_ReadKeys(dd)-cudd.Cudd_ReadDead(dd)) > limit {
		if verbose {
			fmt.Printf("The number of live nodes reached %d.\n", limit)
		}
		return nil
	}
	f := Aig_ObjGlobalBdd(p, e)
	if f == nil {
		node := p.Node(e)
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
			return nil
		}
		child := func(edge aig.Edge) *DdNode {
			f := Bbr_NodeGlobalBdds_rec(dd, p, aig.Aig_Regular(edge), limit, drop, progress, counter, verbose)
			return cudd.Cudd_NotCond(f, edge.Complemented())
		}
		left := child(node.Fanins[0])
		if left == nil {
			return nil
		}
		cudd.Cudd_Ref(left)
		if node.Kind == aig.Buffer {
			f = left
		} else {
			right := child(node.Fanins[1])
			if right == nil {
				cudd.Cudd_RecursiveDeref(dd, left)
				return nil
			}
			cudd.Cudd_Ref(right)
			if node.Kind == aig.ExclusiveOr {
				f = cudd.Cudd_bddXor(dd, left, right)
			} else {
				f = cudd.Cudd_bddAnd(dd, left, right)
			}
			cudd.Cudd_Ref(f)
			cudd.Cudd_RecursiveDeref(dd, left)
			cudd.Cudd_RecursiveDeref(dd, right)
		}
		Aig_ObjSetGlobalBdd(p, e, f)
		*counter++
		if progress != nil {
			progress(*counter)
		}
	}
	// Mirror consumption without altering graph topology; reset refs on exit.
	if aig.Aig_ObjRefs(p, e) > 0 {
		aig.Aig_ObjDeref(p, e)
	}
	if drop && aig.Aig_ObjRefs(p, e) == 0 {
		cudd.Cudd_Deref(f)
		Aig_ObjSetGlobalBdd(p, e, nil)
	}
	return f
}
func Aig_ManFreeGlobalBdds(p *Aig_Man_t, dd *DdManager) {
	if p == nil {
		return
	}
	for id := range p.Nodes() {
		e := aig.Edge(id) << 1
		if p.ValidEdge(e) {
			Aig_ObjCleanGlobalBdd(dd, p, e)
		}
	}
}
func Aig_ManSizeOfGlobalBdds(p *Aig_Man_t) int {
	if p == nil {
		return 0
	}
	seen := map[*DdNode]bool{}
	var visit func(*DdNode)
	visit = func(f *DdNode) {
		if f == nil {
			return
		}
		f = cudd.Cudd_Regular(f)
		if seen[f] {
			return
		}
		seen[f] = true
		if !cudd.Cudd_IsConstant(f) {
			visit(f.T)
			visit(f.E)
		}
	}
	for _, e := range p.Outputs() {
		visit(Aig_ObjGlobalBdd(p, e))
	}
	return len(seen)
}
func Aig_ManComputeGlobalBdds(p *Aig_Man_t, limit int, drop, reorder, verbose bool) *DdManager {
	if p == nil || p.Check() != nil || limit < 0 {
		return nil
	}
	dd := cudd.NewBDDManager(p.InputCount())
	aig.Aig_ManCleanData(p)
	aig.Aig_ManResetRefs(p)
	defer aig.Aig_ManResetRefs(p)
	Aig_ObjSetGlobalBdd(p, p.Const1(), dd.One)
	cudd.Cudd_Ref(dd.One)
	for i, e := range p.Inputs() {
		Aig_ObjSetGlobalBdd(p, e, dd.Vars[i])
		cudd.Cudd_Ref(dd.Vars[i])
	}
	count := 0
	for _, e := range p.Outputs() {
		driver := p.Node(e).Fanins[0]
		f := Bbr_NodeGlobalBdds_rec(dd, p, aig.Aig_Regular(driver), limit, drop, nil, &count, verbose)
		if f == nil {
			Aig_ManFreeGlobalBdds(p, dd)
			cudd.Cudd_Quit(dd)
			return nil
		}
		f = cudd.Cudd_NotCond(f, driver.Complemented())
		cudd.Cudd_Ref(f)
		Aig_ObjSetGlobalBdd(p, e, f)
	}
	if reorder {
		cudd.Cudd_ReduceHeap(dd, cudd.CUDD_REORDER_SYMM_SIFT, 1)
	}
	return dd
}
