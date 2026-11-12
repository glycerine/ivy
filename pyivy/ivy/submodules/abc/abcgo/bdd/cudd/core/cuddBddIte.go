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
	return dd.ite(f, g, dd.LogicZero)
}

func Cudd_bddOr(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.valid(f, g) {
		return nil
	}
	return dd.ite(f, dd.One, g)
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
func cuddBddIteRecur(dd *DdManager, f, g, h *DdNode) *DdNode    { return Cudd_bddIte(dd, f, g, h) }
func cuddBddAndRecur(dd *DdManager, f, g *DdNode) *DdNode       { return Cudd_bddAnd(dd, f, g) }
func cuddBddXorRecur(dd *DdManager, f, g *DdNode) *DdNode       { return Cudd_bddXor(dd, f, g) }
func cuddBddIntersectRecur(dd *DdManager, f, g *DdNode) *DdNode { return Cudd_bddIntersect(dd, f, g) }

type bddAllocationLimit struct{}

// Cudd_bddAndLimit bounds newly allocated unique nodes. A failed operation rolls
// back its unique nodes and computed entries, preserving prior edge identities.
func Cudd_bddAndLimit(dd *DdManager, f, g *DdNode, limit uint32) (result *DdNode) {
	if !dd.valid(f, g) {
		return nil
	}
	oldLimited, oldCeiling := dd.allocationLimited, dd.allocationCeiling
	firstID := dd.nextID
	dd.allocationLimited = true
	dd.allocationCeiling = len(dd.unique) + int(limit)
	defer func() {
		dd.allocationLimited, dd.allocationCeiling = oldLimited, oldCeiling
		if failure := recover(); failure != nil {
			if _, limited := failure.(bddAllocationLimit); !limited {
				panic(failure)
			}
			for key, n := range dd.unique {
				if n.ID >= firstID {
					delete(dd.unique, key)
				}
			}
			clear(dd.cache)
			dd.nextID = firstID
			result = nil
		}
	}()
	return Cudd_bddAnd(dd, f, g)
}
