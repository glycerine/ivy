package dsd

import (
	"fmt"

	bdd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func Dsd_Decompose(m *Dsd_Manager_t, functions []*DdNode, nFuncs int) {
	if nFuncs < 0 || nFuncs > len(functions) {
		panic("dsd: invalid function count")
	}
	m.roots = make([]*Dsd_Node_t, nFuncs)
	for i, f := range functions[:nFuncs] {
		m.roots[i] = Dsd_DecomposeOne(m, f)
		if m.verbose {
			depth, gate := 0, 0
			Dsd_TreeNodeGetInfoOne(m.roots[i], &depth, &gate)
			fmt.Printf("#%02d: Ins=%2d. Gts=%3d. Pri=%3d. Max=%3d. Csc=%2d.\n", i, bdd.Cudd_SupportSize(m.dd, f), Dsd_TreeCountNonTerminalNodesOne(m.roots[i]), Dsd_TreeCountPrimeNodesOne(m.roots[i]), gate, depth)
		}
	}
}
func Dsd_DecomposeOne(m *Dsd_Manager_t, f *DdNode) *Dsd_Node_t {
	if m == nil || m.table == nil || f == nil || bdd.Cudd_Support(m.dd, f) == nil {
		panic("dsd: invalid manager or function")
	}
	for _, index := range m.dd.Support(f) {
		if index >= len(m.inputs) {
			panic("dsd: function uses an input outside the manager's declared support")
		}
	}
	return m.decompose(f)
}

// makeNode stores the global function of the regular decomposition node. The
// cache value may be complemented even though its BDD key is regular.
func (m *Dsd_Manager_t) makeNode(f *DdNode, kind Dsd_Type_t, children []*Dsd_Node_t, complement bool) *Dsd_Node_t {
	p := Dsd_TreeNodeCreate(int(kind), len(children), 0)
	copy(p.pDecs, children)
	p.G = bdd.Cudd_NotCond(f, complement)
	p.S = bdd.Cudd_Support(m.dd, f)
	bdd.Cudd_Ref(p.G)
	bdd.Cudd_Ref(p.S)
	result := Dsd_NotCond(p, complement)
	m.table[f] = result
	return result
}
func prepend(first *Dsd_Node_t, rest []*Dsd_Node_t) []*Dsd_Node_t {
	out := make([]*Dsd_Node_t, 1, len(rest)+1)
	out[0] = first
	return append(out, rest...)
}
func replace(first *Dsd_Node_t, rest []*Dsd_Node_t, skip int) []*Dsd_Node_t {
	out := []*Dsd_Node_t{first}
	for i, p := range rest {
		if i != skip {
			out = append(out, p)
		}
	}
	return out
}
func commonComponents(m *Dsd_Manager_t, l, h *Dsd_Node_t) (common []*Dsd_Node_t, lastL, lastH *Dsd_Node_t) {
	i, j := 0, 0
	for i < len(l.pDecs) && j < len(h.pDecs) {
		a, b := l.pDecs[i], h.pDecs[j]
		la, lb := level(m.dd, Dsd_Regular(a).S), level(m.dd, Dsd_Regular(b).S)
		if la == lb {
			if a == b {
				common = append(common, a)
			} else {
				lastL, lastH = a, b
			}
			i++
			j++
		} else if la < lb {
			lastL = a
			i++
		} else {
			lastH = b
			j++
		}
	}
	if i < len(l.pDecs) {
		lastL = l.pDecs[i]
	}
	if j < len(h.pDecs) {
		lastH = h.pDecs[j]
	}
	return
}
func (m *Dsd_Manager_t) sum(children []*Dsd_Node_t, xor bool) (*DdNode, *DdNode) {
	f, s := m.dd.LogicZero, m.dd.One
	for _, p := range children {
		if xor {
			f = bdd.Cudd_bddXor(m.dd, f, function(p))
		} else {
			f = bdd.Cudd_bddOr(m.dd, f, function(p))
		}
		s = bdd.Cudd_bddAnd(m.dd, s, Dsd_Regular(p).S)
	}
	return f, s
}
func containing(m *Dsd_Manager_t, p *Dsd_Node_t, v *DdNode) (*Dsd_Node_t, bool) {
	if p.Type == DSD_NODE_BUF || p.Type == DSD_NODE_CONST1 {
		return nil, false
	}
	for _, child := range p.pDecs {
		r := Dsd_Regular(child)
		if suppHas(m.dd, r.S, v) {
			return r, Dsd_IsComplement(child)
		}
	}
	panic("dsd: missing containing component")
}

// decompose follows the source's four cofactor cases: terminals/parity,
// containment, common components, and overlap-refined prime components.
func (m *Dsd_Manager_t) decompose(input *DdNode) *Dsd_Node_t {
	dd := m.dd
	f := bdd.Cudd_Regular(input)
	phase := bdd.Cudd_IsComplement(input)
	if hit, ok := m.table[f]; ok {
		return Dsd_NotCond(hit, phase)
	}
	lo, hi := f.E, f.T
	variable := m.inputs[f.Index]
	finish := func(kind Dsd_Type_t, children []*Dsd_Node_t, comp bool) *Dsd_Node_t {
		return Dsd_NotCond(m.makeNode(f, kind, children, comp), phase)
	}
	if hi == dd.One {
		l := m.decompose(lo)
		lr := Dsd_Regular(l)
		if lr.Type == DSD_NODE_OR && !Dsd_IsComplement(l) {
			return finish(DSD_NODE_OR, prepend(variable, lr.pDecs), false)
		}
		return finish(DSD_NODE_OR, []*Dsd_Node_t{variable, l}, false)
	}
	if bdd.Cudd_IsConstant(lo) {
		h := m.decompose(hi)
		hr := Dsd_Regular(h)
		if lo == dd.LogicZero {
			if hr.Type == DSD_NODE_OR && Dsd_IsComplement(h) {
				return finish(DSD_NODE_OR, prepend(Dsd_Not(variable), hr.pDecs), true)
			}
			return finish(DSD_NODE_OR, []*Dsd_Node_t{Dsd_Not(variable), Dsd_Not(h)}, true)
		}
		if hr.Type == DSD_NODE_OR && !Dsd_IsComplement(h) {
			return finish(DSD_NODE_OR, prepend(Dsd_Not(variable), hr.pDecs), false)
		}
		return finish(DSD_NODE_OR, []*Dsd_Node_t{Dsd_Not(variable), h}, false)
	}
	if bdd.Cudd_Regular(lo) == hi {
		l := m.decompose(lo)
		lr := Dsd_Regular(l)
		if lr.Type == DSD_NODE_EXOR {
			return finish(DSD_NODE_EXOR, prepend(variable, lr.pDecs), Dsd_IsComplement(l))
		}
		return finish(DSD_NODE_EXOR, []*Dsd_Node_t{variable, lr}, Dsd_IsComplement(l))
	}
	l, h := m.decompose(lo), m.decompose(hi)
	lr, hr := Dsd_Regular(l), Dsd_Regular(h)
	suppLH := bdd.Cudd_bddAnd(dd, lr.S, hr.S)
	nL, nH, nLH := suppSize(dd, lr.S), suppSize(dd, hr.S), suppSize(dd, suppLH)
	// Case 3b: one cofactor/component contains the support of the other.
	var large, small, largeR, smallR *Dsd_Node_t
	if hr.S != lr.S && suppContains(dd, hr.S, lr.S) {
		large, small, largeR, smallR = h, l, hr, lr
	} else if hr.S != lr.S && suppContains(dd, lr.S, hr.S) {
		large, small, largeR, smallR = l, h, lr, hr
	}
	if largeR != nil {
		lowLarge := smallR != lr
		if largeR.Type == DSD_NODE_PRIME {
			bLarge, bSmall := hi, lo
			if lowLarge {
				bLarge, bSmall = lo, hi
			}
			for i, component := range largeR.pDecs {
				// Prime components have uncomplemented handles.
				g := Dsd_Regular(component).G
				foundOne := m.ident(bLarge, bSmall, g, dd.One)
				foundZero := false
				if !foundOne {
					foundZero = m.ident(bLarge, bSmall, bdd.Cudd_Not(g), dd.One)
				}
				if !foundOne && !foundZero {
					continue
				}
				var merged *DdNode
				if foundOne {
					v := dd.Vars[f.Index]
					if !lowLarge {
						v = bdd.Cudd_Not(v)
					}
					merged = bdd.Cudd_bddOr(dd, v, g)
				} else {
					v := dd.Vars[f.Index]
					if lowLarge {
						v = bdd.Cudd_Not(v)
					}
					merged = bdd.Cudd_bddAnd(dd, v, g)
				}
				replacement := Dsd_Regular(m.decompose(merged))
				return finish(DSD_NODE_PRIME, replace(replacement, largeR.pDecs, i), false)
			}
		}
		var component *Dsd_Node_t
		count := 0
		for _, c := range largeR.pDecs {
			if c == small || c == Dsd_Not(small) {
				component, count = c, 1
				break
			}
		}
		if component == nil && largeR.Type == smallR.Type && (largeR.Type == DSD_NODE_EXOR || (smallR.Type == DSD_NODE_OR && Dsd_IsComplement(large) == Dsd_IsComplement(small))) {
			common, _, _ := commonComponents(m, largeR, smallR)
			if len(common) == len(smallR.pDecs) {
				component, count = smallR, len(smallR.pDecs)
			}
		}
		if component != nil {
			cr := Dsd_Regular(component)
			c1, c2, c3 := Dsd_IsComplement(large), Dsd_IsComplement(component), Dsd_IsComplement(small)
			if largeR.Type == DSD_NODE_OR && (c1 != c2) == c3 {
				target := bdd.Cudd_NotCond(f, c1)
				componentF := bdd.Cudd_NotCond(cr.G, c2)
				remainder := bdd.Cudd_bddAndAbstract(dd, target, bdd.Cudd_Not(componentF), cr.S)
				replacement := m.decompose(remainder)
				children := []*Dsd_Node_t{replacement, component}
				if count != 1 {
					children = prepend(replacement, cr.pDecs)
				}
				return finish(DSD_NODE_OR, children, c1)
			}
			if largeR.Type == DSD_NODE_EXOR {
				target := bdd.Cudd_NotCond(f, !c3)
				remainder := bdd.Cudd_bddXor(dd, target, cr.G)
				replacement := m.decompose(remainder)
				children := []*Dsd_Node_t{replacement, component}
				if count != 1 {
					children = prepend(replacement, cr.pDecs)
				}
				return finish(DSD_NODE_EXOR, children, !c3)
			}
		}
	}
	// Case 3a: disjoint cofactor support forms a three-input prime (MUX).
	if nLH == nL+nH {
		a, b := lr, hr
		if level(dd, a.S) > level(dd, b.S) {
			a, b = b, a
		}
		return finish(DSD_NODE_PRIME, []*Dsd_Node_t{variable, a, b}, false)
	}
	// Case 3c: factor common OR/EXOR components, or merge matching primes.
	if lr.Type == hr.Type && lr.Type != DSD_NODE_BUF && (lr.Type != DSD_NODE_OR || Dsd_IsComplement(l) == Dsd_IsComplement(h)) && (lr.Type != DSD_NODE_PRIME || len(lr.pDecs) == len(hr.pDecs)) {
		common, lastL, lastH := commonComponents(m, lr, hr)
		if len(common) > 0 {
			if lr.Type == DSD_NODE_OR {
				commonF, commonS := m.sum(common, false)
				target := bdd.Cudd_NotCond(f, Dsd_IsComplement(l))
				remainder := bdd.Cudd_bddAndAbstract(dd, target, bdd.Cudd_Not(commonF), commonS)
				return finish(DSD_NODE_OR, prepend(m.decompose(remainder), common), Dsd_IsComplement(l))
			}
			if lr.Type == DSD_NODE_EXOR {
				commonF, _ := m.sum(common, true)
				replacement := m.decompose(bdd.Cudd_bddXor(dd, f, commonF))
				return finish(DSD_NODE_EXOR, prepend(Dsd_Regular(replacement), common), Dsd_IsComplement(replacement))
			}
			if lr.Type == DSD_NODE_PRIME && (len(common) == len(lr.pDecs)-1 || len(common) == len(lr.pDecs)) {
				mode := 0
				if len(common) == len(lr.pDecs) {
					for i, a := range lr.pDecs {
						b := hr.pDecs[i]
						if m.ident(lo, hi, a.G, bdd.Cudd_Not(b.G)) && m.ident(lo, hi, bdd.Cudd_Not(a.G), b.G) {
							lastL, lastH = a, b
							mode = 2
							common = nil
							for _, c := range lr.pDecs {
								if c != a {
									common = append(common, c)
								}
							}
							break
						}
					}
				} else {
					if m.ident(lo, hi, bdd.Cudd_Not(lastL.G), bdd.Cudd_Not(lastH.G)) && m.ident(lo, hi, lastL.G, lastH.G) {
						mode = 1
					} else if m.ident(lo, hi, lastL.G, bdd.Cudd_Not(lastH.G)) && m.ident(lo, hi, bdd.Cudd_Not(lastL.G), lastH.G) {
						mode = 2
					}
				}
				if mode != 0 {
					high := lastH.G
					if mode == 2 {
						high = bdd.Cudd_Not(high)
					}
					merged := bdd.Cudd_bddIte(dd, dd.Vars[f.Index], high, lastL.G)
					return finish(DSD_NODE_PRIME, prepend(Dsd_Regular(m.decompose(merged)), common), false)
				}
			}
		}
	}
	// Case 4: refine overlapping cofactor components to their common blocks.
	children := []*Dsd_Node_t{variable}
	atomicL := lr.Type == DSD_NODE_PRIME && nL == len(lr.pDecs)
	atomicH := hr.Type == DSD_NODE_PRIME && nH == len(hr.pDecs)
	if ((atomicL || atomicH) && lr.S == hr.S) || (atomicL && atomicH) {
		a, b := lr.S, hr.S
		for a != dd.One || b != dd.One {
			la, lb := level(dd, a), level(dd, b)
			top := min(la, lb)
			if la <= lb {
				a = a.T
			}
			if lb <= la {
				b = b.T
			}
			children = append(children, m.inputs[dd.InvPerm[top]])
		}
	} else {
		remainingL, remainingH := lr.S, hr.S
		for remainingL != dd.One || remainingH != dd.One {
			levL, levH := level(dd, remainingL), level(dd, remainingH)
			top := min(levL, levH)
			v := dd.Vars[dd.InvPerm[top]]
			var chosen *Dsd_Node_t
			if levL != levH {
				higher, lowerSupport := lr, hr.S
				if levL > levH {
					higher, lowerSupport = hr, lr.S
				}
				current := higher
				var previous *Dsd_Node_t
				for suppOverlap(dd, current.S, lowerSupport) {
					previous = current
					current, _ = containing(m, current, v)
				}
				if previous == nil || previous.Type == DSD_NODE_PRIME {
					chosen = current
				} else {
					nonoverlap := []*Dsd_Node_t{}
					for _, c := range previous.pDecs {
						if !suppOverlap(dd, Dsd_Regular(c).S, lowerSupport) {
							nonoverlap = append(nonoverlap, c)
						}
					}
					if len(nonoverlap) == 1 {
						chosen = current
					} else {
						sum, _ := m.sum(nonoverlap, previous.Type == DSD_NODE_EXOR)
						chosen = m.decompose(sum)
					}
				}
			} else {
				// Record the complete left containment path with edge polarities; find
				// its first shared node on the right and inspect the preceding parents.
				type step struct {
					node     *Dsd_Node_t
					polarity bool
				}
				path := []step{}
				leftPath := make(map[*Dsd_Node_t]bool)
				cur, pol := lr, false
				for cur != nil {
					leftPath[cur] = true
					path = append(path, step{cur, pol})
					cur, pol = containing(m, cur, v)
				}
				right := hr
				var prevR *Dsd_Node_t
				polR := false
				for !leftPath[right] {
					prevR = right
					right, polR = containing(m, right, v)
					if right == nil {
						panic("dsd: containment paths do not meet")
					}
				}
				var prevL *Dsd_Node_t
				polL := false
				for i, s := range path {
					if s.node == right {
						polL = s.polarity
						if i > 0 {
							prevL = path[i-1].node
						}
						break
					}
				}
				if prevL == nil || prevR == nil || prevL.Type != prevR.Type || prevL.Type == DSD_NODE_PRIME || polL != polR {
					chosen = right
				} else {
					common, _, _ := commonComponents(m, prevL, prevR)
					if len(common) <= 1 {
						chosen = right
					} else {
						sum, _ := m.sum(common, prevL.Type == DSD_NODE_EXOR)
						chosen = m.decompose(sum)
					}
				}
			}
			chosen = Dsd_Regular(chosen)
			children = append(children, chosen)
			if levL <= levH {
				remainingL = bdd.Cudd_bddExistAbstract(dd, remainingL, chosen.S)
			}
			if levH <= levL {
				remainingH = bdd.Cudd_bddExistAbstract(dd, remainingH, chosen.S)
			}
		}
	}
	return finish(DSD_NODE_PRIME, children, false)
}
