// Package cascade ports the LUT-cascade helpers from extraBddCas.c and
// Extra_bddBitsToCube from extraBddMisc.c (Alan Mishchenko, ABC).
package cascade

import (
	"math/bits"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

type DdManager = cudd.DdManager
type DdNode = cudd.DdNode

func Extra_bddBitsToCube(dd *DdManager, code, width int, vars []*DdNode, msbFirst bool) *DdNode {
	if dd == nil || width < 0 || width > bits.UintSize-1 || vars != nil && len(vars) < width {
		return nil
	}
	r := dd.One
	for i := 0; i < width; i++ {
		var v *DdNode
		if vars == nil {
			v = cudd.Cudd_bddIthVar(dd, i)
		} else {
			v = vars[i]
		}
		shift := i
		if msbFirst {
			shift = width - 1 - i
		}
		r = cudd.Cudd_bddAnd(dd, r, cudd.Cudd_NotCond(v, code&(1<<shift) == 0))
	}
	return r
}
func Extra_bddEncodingBinary(dd *DdManager, funcs []*DdNode, vars []*DdNode) *DdNode {
	if len(funcs) > 1 && len(vars) < bits.Len(uint(len(funcs)-1)) {
		return nil
	}
	r := dd.LogicZero
	for i, f := range funcs {
		r = cudd.Cudd_bddOr(dd, r, cudd.Cudd_bddAnd(dd, Extra_bddBitsToCube(dd, i, len(vars), vars, true), f))
	}
	return r
}

// Paths is a deterministic first-visit list of distinct cofactors and the sums
// of input paths reaching each. It supports ADDs and complemented-edge BDDs.
type Paths struct{ Nodes, Columns []*DdNode }

func Extra_bddNodePathsUnderCutArray(dd *DdManager, nodes, codes []*DdNode, cut int) Paths {
	if len(nodes) != len(codes) {
		return Paths{}
	}
	memo := make(map[*DdNode]Paths)
	var visit func(*DdNode) Paths
	visit = func(n *DdNode) Paths {
		if p, ok := memo[n]; ok {
			return p
		}
		reg := cudd.Cudd_Regular(n)
		if cudd.Cudd_IsConstant(reg) || dd.Perm[reg.Index] >= cut {
			return Paths{[]*DdNode{n}, []*DdNode{dd.One}}
		}
		t, e := reg.T, reg.E
		if cudd.Cudd_IsComplement(n) {
			t, e = cudd.Cudd_Not(t), cudd.Cudd_Not(e)
		}
		result := Paths{}
		positions := make(map[*DdNode]int)
		v := cudd.Cudd_bddIthVar(dd, int(reg.Index))
		for _, b := range []struct{ node, lit *DdNode }{{t, v}, {e, cudd.Cudd_Not(v)}} {
			p := visit(b.node)
			for i, n := range p.Nodes {
				path := cudd.Cudd_bddAnd(dd, b.lit, p.Columns[i])
				if j, ok := positions[n]; ok {
					result.Columns[j] = cudd.Cudd_bddOr(dd, result.Columns[j], path)
				} else {
					positions[n] = len(result.Nodes)
					result.Nodes = append(result.Nodes, n)
					result.Columns = append(result.Columns, path)
				}
			}
		}
		memo[n] = result
		return result
	}
	result := Paths{}
	positions := make(map[*DdNode]int)
	for i, n := range nodes {
		p := visit(n)
		for j, k := range p.Nodes {
			path := cudd.Cudd_bddAnd(dd, codes[i], p.Columns[j])
			if path == dd.LogicZero {
				continue
			}
			if pos, ok := positions[k]; ok {
				result.Columns[pos] = cudd.Cudd_bddOr(dd, result.Columns[pos], path)
			} else {
				positions[k] = len(result.Nodes)
				result.Nodes = append(result.Nodes, k)
				result.Columns = append(result.Columns, path)
			}
		}
	}
	return result
}

// Extra_ProfileWidth counts each cut frontier by accumulating the level
// interval spanned by every node's earliest incoming edge. It does not build
// path diagrams, and runs in time proportional to graph size plus variables.
func Extra_ProfileWidth(dd *DdManager, f *DdNode) []int {
	if dd == nil || f == nil {
		return nil
	}
	levels, count := dd.Perm, len(dd.Vars)
	if !cudd.Cudd_bddLeq(dd, f, f) {
		if cudd.Cudd_IsAdd(f) {
			if cudd.Cudd_addFindMin(dd, f) == nil {
				return nil
			}
		} else {
			if cudd.Cudd_zddUnion(dd, f, dd.Zero) == nil {
				return nil
			}
			levels, count = dd.PermZ, dd.ZddSize
		}
	}
	earliest := make(map[*DdNode]int)
	var visit func(*DdNode, int)
	visit = func(n *DdNode, start int) {
		if old, seen := earliest[n]; seen {
			if start < old {
				earliest[n] = start
			}
			return
		}
		earliest[n] = start
		r := cudd.Cudd_Regular(n)
		if !cudd.Cudd_IsConstant(r) {
			t, e := r.T, r.E
			if cudd.Cudd_IsComplement(n) {
				t, e = cudd.Cudd_Not(t), cudd.Cudd_Not(e)
			}
			next := levels[r.Index] + 1
			visit(t, next)
			visit(e, next)
		}
	}
	visit(f, 0)
	difference := make([]int, count+2)
	for n, start := range earliest {
		r := cudd.Cudd_Regular(n)
		end := count
		if !cudd.Cudd_IsConstant(r) {
			end = levels[r.Index]
		}
		difference[start]++
		difference[end+1]--
	}
	profile := make([]int, count+1)
	width := 0
	for cut := range profile {
		width += difference[cut]
		profile[cut] = width
	}
	return profile
}

// Extra_bddEncodingNonStrict uses the source's bounded search (five direct
// input bits, 100 backtracks) before re-encoding each remaining cofactor.
// The returned relation maps each column to a code; simple counts code bits
// that can be wired directly from a LUT input.
func Extra_bddEncodingNonStrict(dd *DdManager, columns []*DdNode, varsCol *DdNode, codes []*DdNode) (relation *DdNode, simple int) {
	encoded := Extra_bddEncodingBinary(dd, columns, codes)
	if encoded == nil {
		return nil, 0
	}
	support := dd.Support(varsCol)
	best := []int{}
	backtracks := 0
	cut := len(dd.Vars)
	if len(codes) > 0 {
		cut = dd.Perm[codes[0].Index]
	}
	// The original counts distinct reachable code subgraphs, summing their
	// minterms rather than taking their union. Overlapping code subgraphs can
	// conservatively reject an otherwise legal direct-input encoding.
	costOf := func(cube *DdNode) float64 {
		assigned := map[uint32]bool{}
		for cube != dd.One {
			n := cudd.Cudd_Regular(cube)
			t, e := n.T, n.E
			if cudd.Cudd_IsComplement(cube) {
				t, e = cudd.Cudd_Not(t), cudd.Cudd_Not(e)
			}
			if t == dd.LogicZero {
				assigned[n.Index] = false
				cube = e
			} else {
				assigned[n.Index] = true
				cube = t
			}
		}
		seen := map[*DdNode]bool{}
		var walk func(*DdNode) float64
		walk = func(n *DdNode) float64 {
			if n == dd.LogicZero || seen[n] {
				return 0
			}
			seen[n] = true
			r := cudd.Cudd_Regular(n)
			if cudd.Cudd_IsConstant(r) || dd.Perm[r.Index] >= cut {
				return cudd.Cudd_CountMinterm(dd, n, len(codes))
			}
			t, e := r.T, r.E
			if cudd.Cudd_IsComplement(n) {
				t, e = cudd.Cudd_Not(t), cudd.Cudd_Not(e)
			}
			if value, ok := assigned[r.Index]; ok {
				if value {
					return walk(t)
				}
				return walk(e)
			}
			return walk(e) + walk(t)
		}
		return walk(encoded)
	}
	// Each cube describes one assignment to already selected direct inputs.
	var search func([]int, []*DdNode, []int)
	search = func(remaining []int, cubes []*DdNode, order []int) {
		depth := len(order) + 1
		if depth > 5 || len(remaining) == 0 || backtracks > 100 || len(order) >= len(codes) {
			return
		}
		backtracks++
		for pos, index := range remaining {
			next := make([]*DdNode, 0, 2*len(cubes))
			v := cudd.Cudd_bddIthVar(dd, index)
			valid := true
			for _, cube := range cubes {
				for _, lit := range []*DdNode{cudd.Cudd_Not(v), v} {
					c := cudd.Cudd_bddAnd(dd, cube, lit)
					cost := costOf(c)
					if cost > float64(uint64(1)<<uint(len(codes)-depth)) {
						valid = false
						break
					}
					next = append(next, c)
				}
				if !valid {
					break
				}
			}
			if valid {
				current := append(append([]int(nil), order...), index)
				if len(current) > len(best) {
					best = append([]int(nil), current...)
				}
				rest := append(append([]int(nil), remaining[:pos]...), remaining[pos+1:]...)
				search(rest, next, current)
			}
			if len(best) == 5 || len(best) == len(codes) {
				return
			}
		}
	}
	if len(columns) > 2 {
		search(support, []*DdNode{dd.One}, nil)
	}
	var build func(*DdNode, int) *DdNode
	build = func(f *DdNode, depth int) *DdNode {
		if depth == len(best) {
			p := Extra_bddNodePathsUnderCutArray(dd, []*DdNode{f}, []*DdNode{dd.One}, cut)
			cols := make([]*DdNode, 0, len(p.Nodes))
			for i, n := range p.Nodes {
				if n != dd.LogicZero {
					cols = append(cols, p.Columns[i])
				}
			}
			return Extra_bddEncodingBinary(dd, cols, codes[depth:])
		}
		v := cudd.Cudd_bddIthVar(dd, best[depth])
		lo := build(cudd.Cudd_Cofactor(dd, f, cudd.Cudd_Not(v)), depth+1)
		hi := build(cudd.Cudd_Cofactor(dd, f, v), depth+1)
		lo = cudd.Cudd_bddAnd(dd, lo, cudd.Cudd_bddAnd(dd, cudd.Cudd_Not(v), cudd.Cudd_Not(codes[depth])))
		hi = cudd.Cudd_bddAnd(dd, hi, cudd.Cudd_bddAnd(dd, v, codes[depth]))
		return cudd.Cudd_bddOr(dd, lo, hi)
	}
	return build(encoded, 0), len(best)
}
