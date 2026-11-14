package cudd

import (
	"sort"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// Variable decomposition uses f = (f OR x) AND (f OR !x). The API returns
// one factor when the split is trivial and two factors otherwise.
func Cudd_bddVarConjDecomp(dd *DdManager, f *DdNode, factors *[]*DdNode) int32 {
	if dd == nil || f == nil || factors == nil {
		return 0
	}
	variables := bddSupportIndices(f)
	if len(variables) == 0 {
		*factors = []*DdNode{f}
		Cudd_Ref(f)
		return 1
	}
	best := -1
	bestSize := int32(1 << 30)
	for _, index := range variables {
		est1 := Cudd_EstimateCofactor(dd, f, int32(index), 1)
		est0 := Cudd_EstimateCofactor(dd, f, int32(index), 0)
		size := est1
		if est0 > size {
			size = est0
		}
		if size < bestSize {
			best, bestSize = index, size
		}
	}
	if best < 0 {
		return 0
	}
	x := core.Cudd_bddIthVar(dd, best)
	glocal := core.Cudd_bddOr(dd, f, x)
	hlocal := core.Cudd_bddOr(dd, f, Cudd_Not(x))
	if glocal == nil || hlocal == nil {
		return 0
	}
	if glocal == Cudd_ReadOne(dd) {
		*factors = []*DdNode{hlocal}
		Cudd_Ref(hlocal)
		return 1
	}
	if hlocal == Cudd_ReadOne(dd) {
		*factors = []*DdNode{glocal}
		Cudd_Ref(glocal)
		return 1
	}
	bestA, bestB := glocal, hlocal
	Cudd_Ref(bestA)
	Cudd_Ref(bestB)
	*factors = []*DdNode{bestA, bestB}
	return 2
}

func Cudd_bddVarDisjDecomp(dd *DdManager, f *DdNode, disjuncts *[]*DdNode) int32 {
	if dd == nil || f == nil || disjuncts == nil {
		return 0
	}
	var factors []*DdNode
	n := Cudd_bddVarConjDecomp(dd, Cudd_Not(f), &factors)
	if n == 0 {
		return 0
	}
	for i := range factors {
		factors[i] = Cudd_Not(factors[i])
	}
	*disjuncts = factors
	return n
}

func Cudd_bddApproxConjDecomp(dd *DdManager, f *DdNode, factors *[]*DdNode) int32 {
	return Cudd_bddVarConjDecomp(dd, f, factors)
}

func Cudd_bddApproxDisjDecomp(dd *DdManager, f *DdNode, disjuncts *[]*DdNode) int32 {
	return Cudd_bddVarDisjDecomp(dd, f, disjuncts)
}

func Cudd_bddIterConjDecomp(dd *DdManager, f *DdNode, factors *[]*DdNode) int32 {
	return Cudd_bddGenConjDecomp(dd, f, factors)
}

func Cudd_bddIterDisjDecomp(dd *DdManager, f *DdNode, disjuncts *[]*DdNode) int32 {
	return Cudd_bddGenDisjDecomp(dd, f, disjuncts)
}

func Cudd_bddGenConjDecomp(dd *DdManager, f *DdNode, factors *[]*DdNode) int32 {
	return Cudd_bddVarConjDecomp(dd, f, factors)
}

func Cudd_bddGenDisjDecomp(dd *DdManager, f *DdNode, disjuncts *[]*DdNode) int32 {
	return Cudd_bddVarDisjDecomp(dd, f, disjuncts)
}

func bddSupportIndices(root *DdNode) []int {
	seenNodes := make(map[*DdNode]bool)
	seenVars := make(map[int]bool)
	var visit func(*DdNode)
	visit = func(node *DdNode) {
		if node == nil || Cudd_IsConstant(node) {
			return
		}
		regular := Cudd_Regular(node)
		if seenNodes[regular] {
			return
		}
		seenNodes[regular] = true
		seenVars[int(regular.Index)] = true
		visit(Cudd_T(regular))
		visit(Cudd_E(regular))
	}
	visit(root)
	result := make([]int, 0, len(seenVars))
	for index := range seenVars {
		result = append(result, index)
	}
	sort.Ints(result)
	return result
}
