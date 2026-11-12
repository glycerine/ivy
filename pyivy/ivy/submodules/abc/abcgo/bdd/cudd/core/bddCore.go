package core

import (
	"fmt"
	"math"
	"sort"
)

// DdNode is an edge handle. Complemented edges have their own canonical Go
// handle rather than a tag hidden in a pointer. Inspect children on Regular().
// Every regular nonterminal has an uncomplemented then edge.
type DdNode struct {
	Index            uint32
	Ref              uint32
	T, E             *DdNode
	Value            float64
	ID               uint64
	manager          *DdManager
	regular, inverse *DdNode
	family           uint8
}
type bddUniqueKey struct {
	index uint32
	t, e  *DdNode
}
type bddITEKey struct{ f, g, h *DdNode }

// DdManager owns canonical ROBDD nodes and the ITE computed table. Like CUDD,
// a manager must not be modified concurrently without external synchronization.
type DdManager struct {
	One, LogicZero                    *DdNode
	Zero, PlusInfinity, MinusInfinity *DdNode
	Background                        *DdNode
	ZddSize                           int
	Univ                              []*DdNode
	zddUnique                         map[bddUniqueKey]*DdNode
	constants                         map[uint64]*DdNode
	Vars                              []*DdNode
	Perm, InvPerm                     []int
	unique                            map[bddUniqueKey]*DdNode
	cache                             map[bddITEKey]*DdNode
	nextID                            uint64
	closed                            bool
	allocationLimited                 bool
	allocationCeiling                 int
}

const bddConstantIndex uint32 = CUDD_CONST_INDEX

func NewBDDManager(variables int) *DdManager {
	if variables < 0 {
		return nil
	}
	dd := &DdManager{unique: map[bddUniqueKey]*DdNode{}, cache: map[bddITEKey]*DdNode{}, nextID: 2}
	dd.One = &DdNode{Index: bddConstantIndex, Value: 1, ID: 0, manager: dd}
	dd.LogicZero = &DdNode{Index: bddConstantIndex, Value: 0, ID: 1, manager: dd, regular: dd.One, inverse: dd.One}
	dd.One.inverse = dd.LogicZero
	dd.constants = map[uint64]*DdNode{math.Float64bits(1): dd.One}
	dd.zddUnique = map[bddUniqueKey]*DdNode{}
	dd.Zero = dd.constant(0)
	dd.Background = dd.Zero
	dd.PlusInfinity = dd.constant(math.Inf(1))
	dd.MinusInfinity = dd.constant(math.Inf(-1))
	for i := 0; i < variables; i++ {
		Cudd_bddIthVar(dd, i)
	}
	return dd
}
func Cudd_Regular(n *DdNode) *DdNode {
	if n == nil {
		return nil
	}
	if n.regular != nil {
		return n.regular
	}
	return n
}
func Cudd_IsComplement(n *DdNode) bool { return n != nil && n.regular != nil }
func Cudd_Not(n *DdNode) *DdNode {
	if n == nil {
		return nil
	}
	return n.inverse
}
func Cudd_NotCond(n *DdNode, condition bool) *DdNode {
	if condition {
		return Cudd_Not(n)
	}
	return n
}
func Cudd_IsConstant(n *DdNode) bool           { return n != nil && Cudd_Regular(n).Index == bddConstantIndex }
func Cudd_T(n *DdNode) *DdNode                 { return Cudd_Regular(n).T }
func Cudd_E(n *DdNode) *DdNode                 { return Cudd_Regular(n).E }
func Cudd_NodeReadIndex(n *DdNode) uint32      { return Cudd_Regular(n).Index }
func Cudd_ReadOne(dd *DdManager) *DdNode       { return dd.One }
func Cudd_ReadLogicZero(dd *DdManager) *DdNode { return dd.LogicZero }
func Cudd_ReadSize(dd *DdManager) int          { return len(dd.Vars) }
func Cudd_ReadPerm(dd *DdManager, index int) int {
	if index < 0 || index >= len(dd.Perm) {
		return -1
	}
	return dd.Perm[index]
}
func Cudd_ReadInvPerm(dd *DdManager, level int) int {
	if level < 0 || level >= len(dd.InvPerm) {
		return -1
	}
	return dd.InvPerm[level]
}
func Cudd_ReadNodeCount(dd *DdManager) int {
	return len(dd.unique) + len(dd.zddUnique) + len(dd.constants)
}
func (dd *DdManager) valid(nodes ...*DdNode) bool {
	if dd == nil || dd.closed {
		return false
	}
	for _, n := range nodes {
		if n == nil || n.manager != dd || Cudd_Regular(n).family != 0 || (Cudd_IsConstant(n) && Cudd_Regular(n) != dd.One) {
			return false
		}
	}
	return true
}
func (dd *DdManager) node(index uint32, t, e *DdNode) *DdNode {
	if t == e {
		return t
	}
	complemented := Cudd_IsComplement(t)
	if complemented {
		t = Cudd_Not(t)
		e = Cudd_Not(e)
	}
	key := bddUniqueKey{index, t, e}
	n := dd.unique[key]
	if n == nil {
		if dd.allocationLimited && len(dd.unique) >= dd.allocationCeiling {
			panic(bddAllocationLimit{})
		}
		n = &DdNode{Index: index, T: t, E: e, ID: dd.nextID, manager: dd}
		n.inverse = &DdNode{Index: index, ID: dd.nextID + 1, manager: dd, regular: n, inverse: n}
		dd.nextID += 2
		dd.unique[key] = n
	}
	return Cudd_NotCond(n, complemented)
}
func Cudd_bddIthVar(dd *DdManager, index int) *DdNode {
	if !dd.valid() || index < 0 || uint64(index) >= uint64(bddConstantIndex) {
		return nil
	}
	for len(dd.Vars) <= index {
		i := len(dd.Vars)
		dd.Perm = append(dd.Perm, i)
		dd.InvPerm = append(dd.InvPerm, i)
		dd.Vars = append(dd.Vars, dd.node(uint32(i), dd.One, dd.LogicZero))
	}
	return dd.Vars[index]
}
func Cudd_bddNewVar(dd *DdManager) *DdNode { return Cudd_bddIthVar(dd, len(dd.Vars)) }
func (dd *DdManager) level(n *DdNode) int {
	r := Cudd_Regular(n)
	if r.Index == bddConstantIndex {
		return math.MaxInt
	}
	return dd.Perm[r.Index]
}
func (dd *DdManager) cofactors(n *DdNode, level int) (*DdNode, *DdNode) {
	if dd.level(n) != level {
		return n, n
	}
	r := Cudd_Regular(n)
	return Cudd_NotCond(r.T, Cudd_IsComplement(n)), Cudd_NotCond(r.E, Cudd_IsComplement(n))
}
func (dd *DdManager) ite(f, g, h *DdNode) *DdNode {
	if f == dd.One {
		return g
	}
	if f == dd.LogicZero {
		return h
	}
	if g == h {
		return g
	}
	if g == dd.One && h == dd.LogicZero {
		return f
	}
	if g == dd.LogicZero && h == dd.One {
		return Cudd_Not(f)
	}
	if Cudd_IsComplement(f) {
		f = Cudd_Not(f)
		g, h = h, g
	}
	invert := Cudd_IsComplement(g)
	if invert {
		g = Cudd_Not(g)
		h = Cudd_Not(h)
	}
	key := bddITEKey{f, g, h}
	if n := dd.cache[key]; n != nil {
		return Cudd_NotCond(n, invert)
	}
	level := min(dd.level(f), min(dd.level(g), dd.level(h)))
	ft, fe := dd.cofactors(f, level)
	gt, ge := dd.cofactors(g, level)
	ht, he := dd.cofactors(h, level)
	t, e := dd.ite(ft, gt, ht), dd.ite(fe, ge, he)
	n := dd.node(uint32(dd.InvPerm[level]), t, e)
	dd.cache[key] = n
	return Cudd_NotCond(n, invert)
}

func Cudd_Ref(n *DdNode) {
	n = Cudd_Regular(n)
	if n != nil && n.Ref < math.MaxUint32 {
		n.Ref++
	}
}
func Cudd_Deref(n *DdNode) {
	n = Cudd_Regular(n)
	if n != nil && n.Ref > 0 && n.Ref < math.MaxUint32 {
		n.Ref--
	}
}
func Cudd_RecursiveDeref(dd *DdManager, n *DdNode) {
	if dd != nil && !dd.closed && n != nil && n.manager == dd {
		Cudd_Deref(n)
	}
}

// Evaluate follows complemented branches without creating nodes.
func (dd *DdManager) Evaluate(n *DdNode, assignment []bool) (bool, error) {
	if !dd.valid(n) {
		return false, fmt.Errorf("cudd: foreign, nil or closed-manager edge")
	}
	for !Cudd_IsConstant(n) {
		r := Cudd_Regular(n)
		if int(r.Index) >= len(assignment) {
			return false, fmt.Errorf("cudd: missing variable %d", r.Index)
		}
		next := r.E
		if assignment[r.Index] {
			next = r.T
		}
		n = Cudd_NotCond(next, Cudd_IsComplement(n))
	}
	return n == dd.One, nil
}

// Support returns variable indices in manager level order.
func (dd *DdManager) Support(n *DdNode) []int {
	if !dd.valid(n) {
		return nil
	}
	seen := map[*DdNode]bool{}
	indices := map[int]bool{}
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		n = Cudd_Regular(n)
		if seen[n] || Cudd_IsConstant(n) {
			return
		}
		seen[n] = true
		indices[int(n.Index)] = true
		visit(n.T)
		visit(n.E)
	}
	visit(n)
	result := make([]int, 0, len(indices))
	for i := range indices {
		result = append(result, i)
	}
	sort.Slice(result, func(i, j int) bool { return dd.Perm[result[i]] < dd.Perm[result[j]] })
	return result
}
func Cudd_Support(dd *DdManager, f *DdNode) *DdNode {
	if !dd.valid(f) {
		return nil
	}
	cube := dd.One
	support := dd.Support(f)
	for i := len(support) - 1; i >= 0; i-- {
		cube = Cudd_bddAnd(dd, dd.Vars[support[i]], cube)
	}
	return cube
}
func Cudd_SupportSize(dd *DdManager, f *DdNode) int { return len(dd.Support(f)) }

// Cudd_bddPermute performs simultaneous substitution. ITE reconstructs the
// canonical order even when the supplied permutation reverses variable levels.
func Cudd_bddPermute(dd *DdManager, f *DdNode, permutation []int) *DdNode {
	if !dd.valid(f) {
		return nil
	}
	highest := -1
	for _, i := range dd.Support(f) {
		if i >= len(permutation) || permutation[i] < 0 || uint64(permutation[i]) >= uint64(bddConstantIndex) {
			return nil
		}
		highest = max(highest, permutation[i])
	}
	if highest >= 0 && Cudd_bddIthVar(dd, highest) == nil {
		return nil
	}
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if Cudd_IsConstant(n) {
			return n
		}
		if v := memo[n]; v != nil {
			return v
		}
		r := Cudd_Regular(n)
		t, e := visit(r.T), visit(r.E)
		v := dd.ite(dd.Vars[permutation[r.Index]], t, e)
		v = Cudd_NotCond(v, Cudd_IsComplement(n))
		memo[n] = v
		return v
	}
	return visit(f)
}
func Cudd_bddCompose(dd *DdManager, f, g *DdNode, index int) *DdNode {
	if !dd.valid(f, g) || index < 0 {
		return nil
	}
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if Cudd_IsConstant(n) {
			return n
		}
		if v := memo[n]; v != nil {
			return v
		}
		r := Cudd_Regular(n)
		t, e := visit(r.T), visit(r.E)
		variable := dd.Vars[r.Index]
		if int(r.Index) == index {
			variable = g
		}
		v := Cudd_NotCond(dd.ite(variable, t, e), Cudd_IsComplement(n))
		memo[n] = v
		return v
	}
	return visit(f)
}
func Cudd_Cofactor(dd *DdManager, f, cube *DdNode) *DdNode {
	if !dd.valid(f, cube) || cube == dd.LogicZero {
		return nil
	}
	assignments := map[uint32]bool{}
	for cube != dd.One {
		if Cudd_IsConstant(cube) {
			return nil
		}
		r := Cudd_Regular(cube)
		t, e := dd.cofactors(cube, dd.level(cube))
		if e == dd.LogicZero {
			assignments[r.Index] = true
			cube = t
		} else if t == dd.LogicZero {
			assignments[r.Index] = false
			cube = e
		} else {
			return nil
		}
	}
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if Cudd_IsConstant(n) {
			return n
		}
		if v := memo[n]; v != nil {
			return v
		}
		r := Cudd_Regular(n)
		var v *DdNode
		if value, ok := assignments[r.Index]; ok {
			next := r.E
			if value {
				next = r.T
			}
			v = visit(next)
		} else {
			v = dd.node(r.Index, visit(r.T), visit(r.E))
		}
		v = Cudd_NotCond(v, Cudd_IsComplement(n))
		memo[n] = v
		return v
	}
	return visit(f)
}
func (dd *DdManager) abstract(f, cube *DdNode, universal bool) *DdNode {
	if !dd.valid(f, cube) {
		return nil
	}
	if !dd.positiveCube(cube) {
		return nil
	}
	quantified := map[uint32]bool{}
	for _, i := range dd.Support(cube) {
		quantified[uint32(i)] = true
	}
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if Cudd_IsConstant(n) {
			return n
		}
		if v := memo[n]; v != nil {
			return v
		}
		r := Cudd_Regular(n)
		t, e := dd.cofactors(n, dd.level(n))
		t, e = visit(t), visit(e)
		var v *DdNode
		if quantified[r.Index] {
			if universal {
				v = Cudd_bddAnd(dd, t, e)
			} else {
				v = Cudd_bddOr(dd, t, e)
			}
		} else {
			v = dd.node(r.Index, t, e)
		}
		memo[n] = v
		return v
	}
	return visit(f)
}

func Cudd_bddAndAbstract(dd *DdManager, f, g, cube *DdNode) *DdNode {
	v := Cudd_bddAnd(dd, f, g)
	if v == nil {
		return nil
	}
	return dd.abstract(v, cube, false)
}
func Cudd_DagSize(f *DdNode) int {
	if f == nil {
		return 0
	}
	seen := map[*DdNode]bool{}
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		n = Cudd_Regular(n)
		if seen[n] {
			return
		}
		seen[n] = true
		if !Cudd_IsConstant(n) {
			visit(n.T)
			visit(n.E)
		}
	}
	visit(f)
	return len(seen)
}
func Cudd_CountMinterm(dd *DdManager, f *DdNode, nvars int) float64 {
	if !dd.valid(f) || nvars < 0 {
		return math.NaN()
	}
	memo := map[*DdNode]float64{}
	var probability func(*DdNode) float64
	probability = func(n *DdNode) float64 {
		if n == dd.One {
			return 1
		}
		if n == dd.LogicZero {
			return 0
		}
		if p, ok := memo[n]; ok {
			return p
		}
		r := Cudd_Regular(n)
		p := (probability(r.T) + probability(r.E)) / 2
		if Cudd_IsComplement(n) {
			p = 1 - p
		}
		memo[n] = p
		return p
	}
	return math.Ldexp(probability(f), nvars)
}

func (dd *DdManager) constant(value float64) *DdNode {
	if value == 0 {
		value = 0
	} // Canonicalize IEEE negative zero.
	key := math.Float64bits(value)
	if n := dd.constants[key]; n != nil {
		return n
	}
	n := &DdNode{Index: bddConstantIndex, Value: value, ID: dd.nextID, manager: dd}
	n.inverse = &DdNode{Index: bddConstantIndex, ID: dd.nextID + 1, manager: dd, regular: n, inverse: n}
	dd.nextID += 2
	dd.constants[key] = n
	return n
}
func (dd *DdManager) zddNode(index uint32, t, e *DdNode) *DdNode {
	if t == dd.Zero {
		return e
	}
	key := bddUniqueKey{index, t, e}
	if n := dd.zddUnique[key]; n != nil {
		return n
	}
	n := &DdNode{Index: index, T: t, E: e, ID: dd.nextID, manager: dd, family: 1}
	n.inverse = &DdNode{Index: index, ID: dd.nextID + 1, manager: dd, regular: n, inverse: n}
	dd.nextID += 2
	dd.zddUnique[key] = n
	return n
}
