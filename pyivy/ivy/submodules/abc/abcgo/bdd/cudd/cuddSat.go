package cudd

import (
	"math"
	"unsafe"

	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/st"
)

const DD_BIGGY int32 = 1000000

type cuddPathPair struct{ pos, neg int32 }

// Cudd_Eval follows the assignment through complemented edges and returns an
// existing terminal node.
func Cudd_Eval(dd *DdManager, f *DdNode, inputs *int32) *DdNode {
	if dd == nil || f == nil || (len(dd.Vars) != 0 && inputs == nil) {
		return nil
	}
	inputValues := unsafe.Slice(inputs, len(dd.Vars))
	for !core.Cudd_IsConstant(f) {
		n := core.Cudd_Regular(f)
		child := core.Cudd_E(n)
		if inputValues[n.Index] == 1 {
			child = core.Cudd_T(n)
		}
		f = core.Cudd_NotCond(child, core.Cudd_IsComplement(f))
	}
	return f
}

func satPair(table *st.Table, node *DdNode, value *cuddPathPair) {
	table.Insert(node, value)
}

// getShortest memoizes the two terminal distances for regular nodes. The
// complemented handle exchanges the distances, as in CUDD's tagged edges.
func getShortest(root *DdNode, weight, support []int32, visited *st.Table, zero *DdNode) cuddPathPair {
	r := core.Cudd_Regular(root)
	if v, ok := visited.Lookup(r); ok {
		p := *v.(*cuddPathPair)
		if core.Cudd_IsComplement(root) {
			p.pos, p.neg = p.neg, p.pos
		}
		return p
	}
	var p cuddPathPair
	if core.Cudd_IsConstant(r) {
		if r != zero {
			p = cuddPathPair{0, DD_BIGGY}
		} else {
			p = cuddPathPair{DD_BIGGY, 0}
		}
	} else {
		t, e := core.Cudd_T(r), core.Cudd_E(r)
		pt := getShortest(t, weight, support, visited, zero)
		pe := getShortest(e, weight, support, visited, zero)
		w := int32(1)
		if weight != nil {
			w = weight[r.Index]
		}
		p.pos = minPath(pe.pos, pt.pos+w)
		p.neg = minPath(pe.neg, pt.neg+w)
		if support != nil {
			support[r.Index] = 1
		}
	}
	satPair(visited, r, &cuddPathPair{p.pos, p.neg})
	if core.Cudd_IsComplement(root) {
		p.pos, p.neg = p.neg, p.pos
	}
	return p
}

func minPath(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func satDistance(table *st.Table, n *DdNode, complement bool) (int32, bool) {
	v, ok := table.Lookup(core.Cudd_Regular(n))
	if !ok {
		return 0, false
	}
	p := v.(*cuddPathPair)
	if complement {
		return p.neg, true
	}
	return p.pos, true
}

func buildPath(dd *DdManager, visited *st.Table, f *DdNode, weight []int32, cost int32, largest bool) *DdNode {
	sol := dd.One
	for !core.Cudd_IsConstant(f) {
		r := core.Cudd_Regular(f)
		t, e := core.Cudd_NotCond(core.Cudd_T(r), core.Cudd_IsComplement(f)), core.Cudd_NotCond(core.Cudd_E(r), core.Cudd_IsComplement(f))
		w := int32(1)
		if weight != nil {
			w = weight[r.Index]
		}
		tc, ec := cost-w, cost
		if largest {
			tc, ec = cost-1, cost-1
		}
		if c, ok := satDistance(visited, t, core.Cudd_IsComplement(t)); ok && c == tc {
			sol = core.Cudd_bddAnd(dd, dd.Vars[r.Index], sol)
			f = t
			cost = tc
		} else if c, ok := satDistance(visited, e, core.Cudd_IsComplement(e)); ok && c == ec {
			sol = core.Cudd_bddAnd(dd, core.Cudd_Not(dd.Vars[r.Index]), sol)
			f = e
			cost = ec
		} else {
			return nil
		}
		if sol == nil {
			return nil
		}
	}
	return sol
}

func runPath(dd *DdManager, f *DdNode, weight, support []int32, length *int32, largest bool) *DdNode {
	if support != nil {
		for i := range dd.Vars {
			support[i] = 0
		}
	}
	if f == dd.LogicZero {
		if length != nil {
			*length = DD_BIGGY
		}
		return core.Cudd_Not(dd.One)
	}
	table := st.New(nil, nil)
	var pair cuddPathPair
	if largest {
		pair = getLargest(f, table, dd.LogicZero)
	} else {
		pair = getShortest(f, weight, support, table, dd.LogicZero)
	}
	cost := pair.pos
	sol := buildPath(dd, table, f, weight, cost, largest)
	if length != nil {
		*length = cost
	}
	return sol
}

func Cudd_ShortestPath(manager *DdManager, f *DdNode, weight, support, length *int32) *DdNode {
	if manager == nil {
		return nil
	}
	var weights, supportValues []int32
	if weight != nil {
		weights = unsafe.Slice(weight, len(manager.Vars))
	}
	if support != nil {
		supportValues = unsafe.Slice(support, len(manager.Vars))
	}
	return runPath(manager, f, weights, supportValues, length, false)
}
func Cudd_LargestCube(manager *DdManager, f *DdNode, length *int32) *DdNode {
	return runPath(manager, f, nil, nil, length, true)
}
func Cudd_ShortestLength(manager *DdManager, f *DdNode, weight *int32) int32 {
	if manager == nil || f == nil {
		return DD_BIGGY
	}
	if f == manager.LogicZero {
		return DD_BIGGY
	}
	var weights []int32
	if weight != nil {
		weights = unsafe.Slice(weight, len(manager.Vars))
	}
	p := getShortest(f, weights, nil, st.New(nil, nil), manager.LogicZero)
	return p.pos
}

func getLargest(root *DdNode, visited *st.Table, zero *DdNode) cuddPathPair {
	r := core.Cudd_Regular(root)
	if v, ok := visited.Lookup(r); ok {
		p := *v.(*cuddPathPair)
		if core.Cudd_IsComplement(root) {
			p.pos, p.neg = p.neg, p.pos
		}
		return p
	}
	var p cuddPathPair
	if core.Cudd_IsConstant(r) {
		if r != zero {
			p = cuddPathPair{0, DD_BIGGY}
		} else {
			p = cuddPathPair{DD_BIGGY, 0}
		}
	} else {
		t, e := getLargest(core.Cudd_T(r), visited, zero), getLargest(core.Cudd_E(r), visited, zero)
		p = cuddPathPair{minPath(t.pos, e.pos) + 1, minPath(t.neg, e.neg) + 1}
	}
	satPair(visited, r, &cuddPathPair{p.pos, p.neg})
	if core.Cudd_IsComplement(root) {
		p.pos, p.neg = p.neg, p.pos
	}
	return p
}

func Cudd_Decreasing(dd *DdManager, f *DdNode, i int32) *DdNode {
	if i < 0 || int(i) >= len(dd.Vars) {
		return core.Cudd_Not(dd.One)
	}
	// f is decreasing in x iff its positive and negative cofactors are ordered.
	var check func(*DdNode) bool
	check = func(n *DdNode) bool {
		if core.Cudd_IsConstant(n) || int(core.Cudd_Regular(n).Index) >= len(dd.Perm) || dd.Perm[core.Cudd_Regular(n).Index] > dd.Perm[i] {
			return true
		}
		r := core.Cudd_Regular(n)
		t, e := core.Cudd_NotCond(core.Cudd_T(r), core.Cudd_IsComplement(n)), core.Cudd_NotCond(core.Cudd_E(r), core.Cudd_IsComplement(n))
		if r.Index == uint32(i) {
			return core.Cudd_bddLeq(dd, t, e)
		}
		return check(t) && check(e)
	}
	if check(f) {
		return dd.One
	}
	return core.Cudd_Not(dd.One)
}
func Cudd_Increasing(dd *DdManager, f *DdNode, i int32) *DdNode {
	return Cudd_Decreasing(dd, core.Cudd_Not(f), i)
}

func Cudd_EquivDC(dd *DdManager, f, g, d *DdNode) int32 {
	if d == dd.One || f == g {
		return 1
	}
	if d == dd.LogicZero || f == core.Cudd_Not(g) {
		return 0
	}
	// Equivalence on the care set is equivalent to the absence of a care-set
	// counterexample in either direction.
	a := core.Cudd_bddAnd(dd, d, core.Cudd_bddXor(dd, f, g))
	if a == nil || a != dd.LogicZero {
		return 0
	}
	return 1
}
func Cudd_bddLeqUnless(dd *DdManager, f, g, d *DdNode) int32 {
	bad := core.Cudd_bddAnd(dd, core.Cudd_Not(d), core.Cudd_bddAnd(dd, f, core.Cudd_Not(g)))
	if bad == nil || bad != dd.LogicZero {
		return 0
	}
	return 1
}

func Cudd_EqualSupNorm(dd *DdManager, f, g *DdNode, tolerance float64, pr int) int32 {
	if f == g {
		return 1
	}
	seen := map[[2]*DdNode]bool{}
	var eq func(*DdNode, *DdNode) bool
	eq = func(a, b *DdNode) bool {
		if a == b {
			return true
		}
		key := [2]*DdNode{a, b}
		if seen[key] {
			return true
		}
		seen[key] = true
		if core.Cudd_IsConstant(a) && core.Cudd_IsConstant(b) {
			return math.Abs(a.Value-b.Value) <= tolerance
		}
		ia, ib := core.Cudd_Regular(a).Index, core.Cudd_Regular(b).Index
		var at, ae, bt, be *DdNode
		if ia <= ib {
			at, ae = core.Cudd_T(a), core.Cudd_E(a)
		} else {
			at, ae = a, a
		}
		if ib <= ia {
			bt, be = core.Cudd_T(b), core.Cudd_E(b)
		} else {
			bt, be = b, b
		}
		return eq(at, bt) && eq(ae, be)
	}
	if eq(f, g) {
		return 1
	}
	return 0
}

func Cudd_bddMakePrime(dd *DdManager, cube, f *DdNode) *DdNode {
	if !core.Cudd_bddLeq(dd, cube, f) {
		return nil
	}
	res := cube
	for !core.Cudd_IsConstant(cube) {
		r := core.Cudd_Regular(cube)
		v := dd.Vars[r.Index]
		expanded := core.Cudd_bddExistAbstract(dd, res, v)
		if expanded == nil {
			return nil
		}
		if core.Cudd_bddLeq(dd, expanded, f) {
			res = expanded
		}
		t, e := core.Cudd_NotCond(core.Cudd_T(r), core.Cudd_IsComplement(cube)), core.Cudd_NotCond(core.Cudd_E(r), core.Cudd_IsComplement(cube))
		if t == dd.LogicZero {
			cube = e
		} else if e == dd.LogicZero {
			cube = t
		} else {
			return nil
		}
	}
	if cube != dd.One {
		return nil
	}
	return res
}

func cuddBddMakePrime(dd *DdManager, cube, f *DdNode) *DdNode {
	return Cudd_bddMakePrime(dd, cube, f)
}
