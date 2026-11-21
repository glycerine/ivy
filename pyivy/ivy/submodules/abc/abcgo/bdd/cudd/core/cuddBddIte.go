/**CFile***********************************************************************

  FileName    [cuddBddIte.c]

  PackageName [cudd]

  Synopsis    [BDD ITE function and satellites.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_bddIte()
                <li> Cudd_bddIteConstant()
                <li> Cudd_bddIntersect()
                <li> Cudd_bddAnd()
                <li> Cudd_bddAndLimit()
                <li> Cudd_bddOr()
                <li> Cudd_bddNand()
                <li> Cudd_bddNor()
                <li> Cudd_bddXor()
                <li> Cudd_bddXnor()
                <li> Cudd_bddLeq()
                </ul>
       Internal procedures included in this module:
                <ul>
                <li> cuddBddIteRecur()
                <li> cuddBddIntersectRecur()
                <li> cuddBddAndRecur()
                <li> cuddBddXorRecur()
                </ul>
       Static procedures included in this module:
                <ul>
                <li> bddVarToConst()
                <li> bddVarToCanonical()
                <li> bddVarToCanonicalSimple()
                </ul>]

  SeeAlso     []

  Author      [Fabio Somenzi]

  Copyright   [Copyright (c) 1995-2004, Regents of the University of Colorado

  All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions
  are met:

  Redistributions of source code must retain the above copyright
  notice, this list of conditions and the following disclaimer.

  Redistributions in binary form must reproduce the above copyright
  notice, this list of conditions and the following disclaimer in the
  documentation and/or other materials provided with the distribution.

  Neither the name of the University of Colorado nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
  "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
  LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
  FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
  COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
  INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
  BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
  LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
  CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
  LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN
  ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
  POSSIBILITY OF SUCH DAMAGE.]

******************************************************************************/

package core

func Cudd_bddIte(dd *DdManager, f, g, h *DdNode) *DdNode {
	if !dd.valid(f, g, h) {
		return nil
	}
	return dd.ite(f, g, h)
}

func Cudd_bddAnd(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.valid(f, g) {
		return nil
	}
	return cuddBddAndRecur(dd, f, g)
}

func Cudd_bddOr(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.valid(f, g) {
		return nil
	}
	return Cudd_Not(cuddBddAndRecur(dd, Cudd_Not(f), Cudd_Not(g)))
}

func Cudd_bddXor(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.valid(f, g) {
		return nil
	}
	return dd.ite(f, Cudd_Not(g), g)
}

func Cudd_bddXnor(dd *DdManager, f, g *DdNode) *DdNode { return Cudd_Not(Cudd_bddXor(dd, f, g)) }

func Cudd_bddNand(dd *DdManager, f, g *DdNode) *DdNode { return Cudd_Not(Cudd_bddAnd(dd, f, g)) }

func Cudd_bddNor(dd *DdManager, f, g *DdNode) *DdNode { return Cudd_Not(Cudd_bddOr(dd, f, g)) }

func Cudd_bddLeq(dd *DdManager, f, g *DdNode) bool {
	if !dd.valid(f, g) {
		return false
	}
	return dd.disjoint(f, Cudd_Not(g), map[[2]*DdNode]bool{}, map[[2]*DdNode]bool{})
}

// DD_NON_CONSTANT is the original ITE-constant query's distinguished result.
// It is not an edge belonging to any manager.
var DD_NON_CONSTANT = &DdNode{Index: bddConstantIndex - 1}

func Cudd_bddIteConstant(dd *DdManager, f, g, h *DdNode) *DdNode {
	if !dd.valid(f, g, h) {
		return nil
	}
	memo := map[bddITEKey]*DdNode{}
	var visit func(*DdNode, *DdNode, *DdNode) *DdNode
	visit = func(f, g, h *DdNode) *DdNode {
		if f == dd.One {
			return g
		}
		if f == dd.LogicZero {
			return h
		}
		if g == h {
			return g
		}
		key := bddITEKey{f, g, h}
		if v := memo[key]; v != nil {
			return v
		}
		level := min(dd.level(f), min(dd.level(g), dd.level(h)))
		ft, fe := dd.cofactors(f, level)
		gt, ge := dd.cofactors(g, level)
		ht, he := dd.cofactors(h, level)
		t := visit(ft, gt, ht)
		if t == DD_NON_CONSTANT || !Cudd_IsConstant(t) {
			memo[key] = DD_NON_CONSTANT
			return DD_NON_CONSTANT
		}
		e := visit(fe, ge, he)
		result := DD_NON_CONSTANT
		if t == e {
			result = t
		}
		memo[key] = result
		return result
	}
	return visit(f, g, h)
}
func Cudd_bddIntersect(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.valid(f, g) {
		return nil
	}
	memo := map[[2]*DdNode]*DdNode{}
	var visit func(*DdNode, *DdNode) *DdNode
	visit = func(f, g *DdNode) *DdNode {
		if f == dd.LogicZero || g == dd.LogicZero || f == Cudd_Not(g) {
			return dd.LogicZero
		}
		if f == g || g == dd.One {
			return f
		}
		if f == dd.One {
			return g
		}
		key := [2]*DdNode{f, g}
		if v := memo[key]; v != nil {
			return v
		}
		level := min(dd.level(f), dd.level(g))
		ft, fe := dd.cofactors(f, level)
		gt, ge := dd.cofactors(g, level)
		t := visit(ft, gt)
		var v *DdNode
		if t != dd.LogicZero {
			v = dd.node(uint32(dd.InvPerm[level]), t, dd.LogicZero)
		} else {
			v = dd.node(uint32(dd.InvPerm[level]), dd.LogicZero, visit(fe, ge))
		}
		memo[key] = v
		return v
	}
	return visit(f, g)
}
func cuddBddIteRecur(dd *DdManager, f, g, h *DdNode) *DdNode { return Cudd_bddIte(dd, f, g, h) }
func cuddBddAndRecur(dd *DdManager, f, g *DdNode) *DdNode    { return dd.andChecked(f, g, nil) }

// The projection's implicit manager pin contributes one native reference.
// AND calls this only for regular BDD decisions.
func (dd *DdManager) singleBDDRef(n *DdNode) bool {
	if n.Ref > 1 {
		return false
	}
	if n == dd.Vars[n.Index] {
		return n.Ref == 0
	}
	return n.Ref == 1
}

// Use the CUDD two-operand recursion for conjunction and its De Morgan dual.
// A nil third operand distinguishes its computed entries from general ITEs.
func (dd *DdManager) andChecked(f, g *DdNode, check func() bool) *DdNode {
	if check != nil && !check() {
		return nil
	}
	fr, gr := Cudd_Regular(f), Cudd_Regular(g)
	switch {
	case f == g, g == dd.One:
		return f
	case f == dd.One:
		return g
	case fr == gr, f == dd.LogicZero, g == dd.LogicZero:
		return dd.LogicZero
	}
	if f.ID > g.ID {
		f, g = g, f
		fr, gr = gr, fr
	}
	key := computedKey{f: f, g: g, kind: cacheBDDAnd}
	hash := key.hash()
	if !dd.singleBDDRef(fr) || !dd.singleBDDRef(gr) {
		if r := dd.cacheLookup(key, hash, true); r != nil {
			return r
		}
	}

	lf, lg := dd.Perm[fr.Index], dd.Perm[gr.Index]
	index := gr.Index
	ft, fe, gt, ge := f, f, g, g
	if lf <= lg {
		index = fr.Index
		ft, fe = fr.T, fr.E
		if f != fr {
			ft, fe = Cudd_Not(ft), Cudd_Not(fe)
		}
	}
	if lg <= lf {
		gt, ge = gr.T, gr.E
		if g != gr {
			gt, ge = Cudd_Not(gt), Cudd_Not(ge)
		}
	}
	t := dd.andChecked(ft, gt, check)
	if t == nil {
		return nil
	}
	e := dd.andChecked(fe, ge, check)
	if e == nil {
		return nil
	}
	r := dd.node(index, t, e)
	if !dd.singleBDDRef(fr) || !dd.singleBDDRef(gr) {
		dd.cacheInsert(key, r, hash)
	}
	return r
}

func cuddBddXorRecur(dd *DdManager, f, g *DdNode) *DdNode       { return Cudd_bddXor(dd, f, g) }
func cuddBddIntersectRecur(dd *DdManager, f, g *DdNode) *DdNode { return Cudd_bddIntersect(dd, f, g) }

type bddAllocationLimit struct{}

// A failed bounded operation removes its new nodes and releases exactly the
// child references acquired while constructing or reviving intermediate nodes.
type bddAllocationAttempt struct {
	dd         *DdManager
	parent     *bddAllocationAttempt
	firstID    uint64
	oldLimited bool
	oldCeiling int
	revived    []*DdNode
}

func (dd *DdManager) beginAllocationAttempt(limit uint64) *bddAllocationAttempt {
	a := &bddAllocationAttempt{
		dd: dd, parent: dd.allocationAttempt, firstID: dd.nextID,
		oldLimited: dd.allocationLimited, oldCeiling: dd.allocationCeiling,
	}
	used := dd.unique.count
	ceiling := used + int(min(limit, uint64(int(^uint(0)>>1)-used)))
	if dd.allocationLimited {
		ceiling = min(ceiling, dd.allocationCeiling)
	}
	dd.allocationAttempt, dd.allocationLimited, dd.allocationCeiling = a, true, ceiling
	return a
}

func (a *bddAllocationAttempt) finish(result **DdNode) {
	dd := a.dd
	dd.allocationAttempt, dd.allocationLimited, dd.allocationCeiling = a.parent, a.oldLimited, a.oldCeiling
	if failure := recover(); failure != nil {
		if _, limited := failure.(bddAllocationLimit); !limited {
			panic(failure)
		}
		*result = nil
	}
	if *result != nil {
		if a.parent != nil {
			a.parent.revived = append(a.parent.revived, a.revived...)
		}
		return
	}
	for key, n := range dd.unique.entries {
		if n.ID >= a.firstID {
			if n.childrenReleased {
				dd.dead--
			}
			if !n.childrenReleased {
				Cudd_Deref(n.T)
				Cudd_Deref(n.E)
			}
			n.detachHandle()
			dd.unique.remove(key)
		}
	}
	for _, n := range a.revived {
		if n.childrenReleased {
			continue
		}
		n.childrenReleased = true
		if n.inBDDTable {
			dd.dead++
		} else if n.family == 1 {
			dd.deadZ++
		}
		Cudd_Deref(n.T)
		Cudd_Deref(n.E)
	}
	Cudd_CacheFlush(dd)
	dd.nextID = a.firstID
}

// Cudd_bddAndLimit bounds newly allocated unique nodes. A failed operation rolls
// back its unique nodes and computed entries, preserving prior edge identities.
func Cudd_bddAndLimit(dd *DdManager, f, g *DdNode, limit uint32) (result *DdNode) {
	if !dd.valid(f, g) {
		return nil
	}
	defer dd.beginAllocationAttempt(uint64(limit)).finish(&result)
	return Cudd_bddAnd(dd, f, g)
}
