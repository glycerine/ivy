package cudd

// Matrix multiplication and shortest-path helpers for ADDs. The recursion
// follows the variable order of the manager, so it remains correct after a
// variable reorder.

import (
	"unsafe"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// cuddNodeArray accepts both Go slices and the pointer-to-first-element form
// used by translated C callers. The latter is only valid when it points to n
// consecutive *DdNode values, as in the original C API.
func cuddMatNodeArray(array any, n int32) []*DdNode {
	if n <= 0 || array == nil {
		return nil
	}
	if nodes, ok := array.([]*DdNode); ok {
		if int32(len(nodes)) < n {
			return nil
		}
		return nodes[:n]
	}
	if first, ok := array.(**DdNode); ok && first != nil {
		return unsafe.Slice(first, int(n))
	}
	return nil
}

func cuddMatTop(dd *DdManager, n *DdNode) int {
	if n == nil || Cudd_IsConstant(n) {
		return int(Cudd_ReadSize(dd))
	}
	return int(Cudd_ReadPerm(dd, int32(n.Index)))
}

func cuddMatCofactors(dd *DdManager, n *DdNode, level int) (*DdNode, *DdNode) {
	if cuddMatTop(dd, n) != level {
		return n, n
	}
	return n.T, n.E
}

func cuddMatNode(dd *DdManager, level int, then, otherwise *DdNode) *DdNode {
	if then == nil || otherwise == nil {
		return nil
	}
	if then == otherwise {
		return then
	}
	index := Cudd_ReadInvPerm(dd, int32(level))
	if index < 0 {
		return nil
	}
	// ADD cofactors must be recombined with an ADD variable. A BDD variable
	// is a complemented-edge node and is not a valid ADD ITE condition.
	v := Cudd_addIthVar(dd, int32(index))
	if v == nil {
		return nil
	}
	return Cudd_addIte(dd, v, then, otherwise)
}

// Cudd_addMatrixMultiply computes the quasiring product and sums over z.
// z may be a []*DdNode or a pointer to the first item of a C-style array.
func Cudd_addMatrixMultiply(dd *DdManager, A, B *DdNode, z any, nz int32) *DdNode {
	vars := cuddMatNodeArray(z, nz)
	if nz > 0 && vars == nil {
		return nil
	}
	sum := make(map[int]bool, nz)
	for _, v := range vars {
		if v == nil || Cudd_IsConstant(v) {
			return nil
		}
		sum[int(v.Index)] = true
	}
	cache := make(map[matPair]*DdNode)
	return cuddAddMMRecur(dd, A, B, -1, sum, cache)
}

type matPair struct{ a, b *DdNode }

func cuddAddMMRecur(dd *DdManager, a, b *DdNode, top int, sum map[int]bool, cache map[matPair]*DdNode) *DdNode {
	if a == nil || b == nil {
		return nil
	}
	if a == dd.Zero || b == dd.Zero {
		return dd.Zero
	}
	if Cudd_IsConstant(a) && Cudd_IsConstant(b) {
		value := core.Cudd_V(a) * core.Cudd_V(b)
		for index := range sum {
			if int(Cudd_ReadPerm(dd, int32(index))) > top {
				value *= 2
			}
		}
		return Cudd_addConst(dd, value)
	}
	if a.ID > b.ID {
		a, b = b, a
	}
	key := matPair{a, b}
	ta, tb := cuddMatTop(dd, a), cuddMatTop(dd, b)
	level := ta
	if tb < level {
		level = tb
	}
	if cached := cache[key]; cached != nil {
		return cuddMatScale(dd, cached, top, level, sum)
	}
	at, ae := cuddMatCofactors(dd, a, level)
	bt, be := cuddMatCofactors(dd, b, level)
	t := cuddAddMMRecur(dd, at, bt, level, sum, cache)
	if t == nil {
		return nil
	}
	e := cuddAddMMRecur(dd, ae, be, level, sum, cache)
	if e == nil {
		return nil
	}
	index := int(Cudd_ReadInvPerm(dd, int32(level)))
	var res *DdNode
	if sum[index] {
		res = Cudd_addApply(dd, Cudd_addPlus, t, e)
	} else {
		res = cuddMatNode(dd, level, t, e)
	}
	if res == nil {
		return nil
	}
	cache[key] = res
	// The cache stores the minimally scaled result. Account for summation
	// variables skipped between the parent and this node.
	return cuddMatScale(dd, res, top, level, sum)
}

func cuddMatScale(dd *DdManager, res *DdNode, top, level int, sum map[int]bool) *DdNode {
	scale := 1.0
	for i := range sum {
		p := int(Cudd_ReadPerm(dd, int32(i)))
		if p > top && p < level {
			scale *= 2
		}
	}
	if scale != 1 {
		res = Cudd_addApply(dd, Cudd_addTimes, res, Cudd_addConst(dd, scale))
	}
	return res
}

// Cudd_addTimesPlus computes sum_z(A*B) using ADD existential abstraction.
func Cudd_addTimesPlus(dd *DdManager, A, B *DdNode, z any, nz int32) *DdNode {
	vars := cuddMatNodeArray(z, nz)
	if nz > 0 && vars == nil {
		return nil
	}
	tmp := Cudd_addApply(dd, Cudd_addTimes, A, B)
	if tmp == nil {
		return nil
	}
	cube := dd.One
	Cudd_Ref(tmp)
	Cudd_Ref(cube)
	for i := len(vars) - 1; i >= 0; i-- {
		if vars[i] == nil {
			Cudd_RecursiveDeref(dd, tmp)
			Cudd_RecursiveDeref(dd, cube)
			return nil
		}
		previous := cube
		next := Cudd_addIte(dd, vars[i], previous, dd.Zero)
		if next == nil {
			Cudd_RecursiveDeref(dd, tmp)
			Cudd_RecursiveDeref(dd, previous)
			return nil
		}
		Cudd_Ref(next)
		Cudd_RecursiveDeref(dd, previous)
		cube = next
	}
	res := Cudd_addExistAbstract(dd, tmp, cube)
	if res == nil {
		Cudd_RecursiveDeref(dd, tmp)
		Cudd_RecursiveDeref(dd, cube)
		return nil
	}
	Cudd_Ref(res)
	Cudd_RecursiveDeref(dd, cube)
	Cudd_RecursiveDeref(dd, tmp)
	Cudd_Deref(res)
	return res
}

// Cudd_addTriangle performs the shortest-path semiring product and abstracts
// z by pointwise minimum.
func Cudd_addTriangle(dd *DdManager, f, g *DdNode, z any, nz int32) *DdNode {
	vars := cuddMatNodeArray(z, nz)
	if nz > 0 && vars == nil {
		return nil
	}
	abstract := make(map[int]bool, nz)
	for _, v := range vars {
		if v == nil || Cudd_IsConstant(v) {
			return nil
		}
		abstract[int(v.Index)] = true
	}
	cache := make(map[matPair]*DdNode)
	return cuddAddTriangleRecur(dd, f, g, abstract, cache)
}

func cuddAddTriangleRecur(dd *DdManager, f, g *DdNode, abstract map[int]bool, cache map[matPair]*DdNode) *DdNode {
	if f == nil || g == nil {
		return nil
	}
	if f == dd.PlusInfinity || g == dd.PlusInfinity {
		return dd.PlusInfinity
	}
	if Cudd_IsConstant(f) && Cudd_IsConstant(g) {
		return Cudd_addConst(dd, core.Cudd_V(f)+core.Cudd_V(g))
	}
	if f.ID < g.ID {
		f, g = g, f
	}
	key := matPair{f, g}
	if r := cache[key]; r != nil {
		return r
	}
	level := cuddMatTop(dd, f)
	if other := cuddMatTop(dd, g); other < level {
		level = other
	}
	ft, fe := cuddMatCofactors(dd, f, level)
	gt, ge := cuddMatCofactors(dd, g, level)
	t := cuddAddTriangleRecur(dd, ft, gt, abstract, cache)
	if t == nil {
		return nil
	}
	e := cuddAddTriangleRecur(dd, fe, ge, abstract, cache)
	if e == nil {
		return nil
	}
	var result *DdNode
	if abstract[int(Cudd_ReadInvPerm(dd, int32(level)))] {
		result = Cudd_addApply(dd, Cudd_addMinimum, t, e)
	} else {
		result = cuddMatNode(dd, level, t, e)
	}
	cache[key] = result
	return result
}

// Cudd_addOuterSum returns min(M, r+c) pointwise.
func Cudd_addOuterSum(dd *DdManager, M, r, c *DdNode) *DdNode {
	cache := make(map[outerSumKey]*DdNode)
	return cuddAddOuterSumRecur(dd, M, r, c, cache)
}

type outerSumKey struct{ m, r, c *DdNode }

func cuddAddOuterSumRecur(dd *DdManager, m, r, c *DdNode, cache map[outerSumKey]*DdNode) *DdNode {
	if m == nil || r == nil || c == nil {
		return nil
	}
	if r == dd.PlusInfinity || c == dd.PlusInfinity {
		return m
	}
	if Cudd_IsConstant(r) && Cudd_IsConstant(c) {
		sum := Cudd_addConst(dd, core.Cudd_V(r)+core.Cudd_V(c))
		if Cudd_IsConstant(m) {
			if core.Cudd_V(sum) <= core.Cudd_V(m) {
				return sum
			}
			return m
		}
		return Cudd_addApply(dd, Cudd_addMinimum, sum, m)
	}
	key := outerSumKey{m, r, c}
	if result := cache[key]; result != nil {
		return result
	}
	level := cuddMatTop(dd, m)
	for _, n := range []*DdNode{r, c} {
		if l := cuddMatTop(dd, n); l < level {
			level = l
		}
	}
	mt, me := cuddMatCofactors(dd, m, level)
	rt, re := cuddMatCofactors(dd, r, level)
	ct, ce := cuddMatCofactors(dd, c, level)
	t := cuddAddOuterSumRecur(dd, mt, rt, ct, cache)
	if t == nil {
		return nil
	}
	e := cuddAddOuterSumRecur(dd, me, re, ce, cache)
	if e == nil {
		return nil
	}
	result := cuddMatNode(dd, level, t, e)
	cache[key] = result
	return result
}
