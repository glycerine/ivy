/**CFile***********************************************************************

  FileName    [cuddAddIte.c]

  PackageName [cudd]

  Synopsis    [ADD ITE function and satellites.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addIte()
                <li> Cudd_addIteConstant()
                <li> Cudd_addEvalConst()
                <li> Cudd_addCmpl()
                <li> Cudd_addLeq()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddAddIteRecur()
                <li> cuddAddCmplRecur()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> addVarToConst()
                </ul>]

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

import (
	"fmt"
)

// Cudd_IsAdd identifies an uncomplemented ADD edge. Constant one is shared by
// BDD and ADD families; other BDD nodes never pass this test.
func Cudd_IsAdd(n *DdNode) bool {
	return n != nil && !Cudd_IsComplement(n) && (Cudd_IsConstant(n) || n.family == 2)
}
func (dd *DdManager) validAdd(nodes ...*DdNode) bool {
	if dd == nil || dd.closed {
		return false
	}
	for _, n := range nodes {
		if n == nil || n.manager != dd || !Cudd_IsAdd(n) {
			return false
		}
	}
	return true
}
func (dd *DdManager) addNode(index uint32, t, e *DdNode) *DdNode {
	if t == e {
		return t
	}
	if dd.addUnique == nil {
		dd.addUnique = map[bddUniqueKey]*DdNode{}
	}
	key := bddUniqueKey{index, t, e}
	if n := dd.addUnique[key]; n != nil {
		reclaimNode(n)
		return n
	}
	n := &DdNode{Index: index, T: t, E: e, ID: dd.nextID, manager: dd, family: 2}
	dd.nextID += 2
	Cudd_Ref(t)
	Cudd_Ref(e)
	dd.addUnique[key] = n
	return n
}
func Cudd_addConst(dd *DdManager, value float64) *DdNode {
	if dd == nil || dd.closed {
		return nil
	}
	return dd.constant(value)
}
func Cudd_addIthVar(dd *DdManager, index int) *DdNode {
	if Cudd_bddIthVar(dd, index) == nil {
		return nil
	}
	return dd.addNode(uint32(index), dd.One, dd.Zero)
}
func Cudd_addNewVar(dd *DdManager) *DdNode {
	if dd == nil {
		return nil
	}
	return Cudd_addIthVar(dd, len(dd.Vars))
}

// Cudd_addNewVarAtLevel creates the next ADD variable and places it at level.
// Existing ADD/BDD/ZDD functions remain valid because their node order follows
// the manager permutation, which is updated by inserting only the new index.
func Cudd_addNewVarAtLevel(dd *DdManager, level int) *DdNode {
	if dd == nil || dd.closed || level < 0 || level > len(dd.Vars) {
		return nil
	}
	v := Cudd_addNewVar(dd)
	if v == nil || level == len(dd.Vars)-1 {
		return v
	}
	order := append([]int(nil), dd.InvPerm...)
	index := int(v.Index)
	copy(order[level+1:], order[level:len(order)-1])
	order[level] = index
	for newLevel, variable := range order {
		dd.InvPerm[newLevel] = variable
		dd.Perm[variable] = newLevel
	}
	dd.symmetryGroups = nil
	return v
}
func (dd *DdManager) addCofactors(n *DdNode, level int) (*DdNode, *DdNode) {
	if dd.level(n) == level {
		return n.T, n.E
	}
	return n, n
}
func Cudd_addIte(dd *DdManager, f, g, h *DdNode) *DdNode {
	if !dd.validAdd(f, g, h) {
		return nil
	}
	if dd.addCache == nil {
		dd.addCache = map[bddITEKey]*DdNode{}
	}
	var visit func(*DdNode, *DdNode, *DdNode) *DdNode
	visit = func(f, g, h *DdNode) *DdNode {
		if Cudd_IsConstant(f) {
			if f.Value == 0 {
				return h
			}
			return g
		}
		if g == h {
			return g
		}
		key := bddITEKey{f, g, h}
		dd.cacheLookups++
		if n := dd.addCache[key]; n != nil {
			dd.cacheHits++
			return n
		}
		level := min(dd.level(f), min(dd.level(g), dd.level(h)))
		ft, fe := dd.addCofactors(f, level)
		gt, ge := dd.addCofactors(g, level)
		ht, he := dd.addCofactors(h, level)
		n := dd.addNode(uint32(dd.InvPerm[level]), visit(ft, gt, ht), visit(fe, ge, he))
		dd.addCache[key] = n
		return n
	}
	return visit(f, g, h)
}
func Cudd_addPermute(dd *DdManager, f *DdNode, permutation []int) *DdNode {
	if !dd.validAdd(f) {
		return nil
	}
	support := dd.Support(f)
	highest := -1
	for _, i := range support {
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
		v := Cudd_addIte(dd, Cudd_addIthVar(dd, permutation[n.Index]), visit(n.T), visit(n.E))
		memo[n] = v
		return v
	}
	return visit(f)
}
func (dd *DdManager) EvaluateADD(f *DdNode, assignment []bool) (float64, error) {
	if !dd.validAdd(f) {
		return 0, fmt.Errorf("cudd: invalid ADD edge")
	}
	for !Cudd_IsConstant(f) {
		if int(f.Index) >= len(assignment) {
			return 0, fmt.Errorf("cudd: missing variable %d", f.Index)
		}
		if assignment[f.Index] {
			f = f.T
		} else {
			f = f.E
		}
	}
	return f.Value, nil
}

func Cudd_addIteConstant(dd *DdManager, f, g, h *DdNode) *DdNode {
	if !dd.validAdd(f, g, h) {
		return nil
	}
	memo := map[bddITEKey]*DdNode{}
	var visit func(*DdNode, *DdNode, *DdNode) *DdNode
	visit = func(f, g, h *DdNode) *DdNode {
		if Cudd_IsConstant(f) {
			if f.Value == 0 {
				return h
			}
			return g
		}
		if g == h {
			return g
		}
		key := bddITEKey{f, g, h}
		if n := memo[key]; n != nil {
			return n
		}
		level := min(dd.level(f), min(dd.level(g), dd.level(h)))
		ft, fe := dd.addCofactors(f, level)
		gt, ge := dd.addCofactors(g, level)
		ht, he := dd.addCofactors(h, level)
		t := visit(ft, gt, ht)
		if t == DD_NON_CONSTANT || !Cudd_IsConstant(t) {
			memo[key] = DD_NON_CONSTANT
			return DD_NON_CONSTANT
		}
		e := visit(fe, ge, he)
		n := DD_NON_CONSTANT
		if t == e {
			n = t
		}
		memo[key] = n
		return n
	}
	return visit(f, g, h)
}
func Cudd_addEvalConst(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.validAdd(f, g) {
		return nil
	}
	memo := map[[2]*DdNode]*DdNode{}
	var visit func(*DdNode, *DdNode) *DdNode
	visit = func(f, g *DdNode) *DdNode {
		if f == dd.One || Cudd_IsConstant(g) {
			return g
		}
		if f == dd.Zero {
			return dd.Background
		}
		key := [2]*DdNode{f, g}
		if n := memo[key]; n != nil {
			return n
		}
		level := min(dd.level(f), dd.level(g))
		ft, fe := dd.addCofactors(f, level)
		gt, ge := dd.addCofactors(g, level)
		if ft == dd.Zero {
			n := visit(fe, ge)
			memo[key] = n
			return n
		}
		t := visit(ft, gt)
		if t == DD_NON_CONSTANT || !Cudd_IsConstant(t) {
			memo[key] = DD_NON_CONSTANT
			return DD_NON_CONSTANT
		}
		if fe != dd.Zero {
			if e := visit(fe, ge); e != t {
				memo[key] = DD_NON_CONSTANT
				return DD_NON_CONSTANT
			}
		}
		memo[key] = t
		return t
	}
	return visit(f, g)
}
func Cudd_addCmpl(dd *DdManager, f *DdNode) *DdNode {
	return Cudd_addMonadicApply(dd, func(dd *DdManager, n *DdNode) *DdNode {
		if Cudd_IsConstant(n) {
			return addBoolean(dd, n == dd.Zero)
		}
		return nil
	}, f)
}
func Cudd_addLeq(dd *DdManager, f, g *DdNode) bool {
	if !dd.validAdd(f, g) {
		return false
	}
	known := map[[2]*DdNode]bool{}
	memo := map[[2]*DdNode]bool{}
	var visit func(*DdNode, *DdNode) bool
	visit = func(f, g *DdNode) bool {
		if f == g {
			return true
		}
		if Cudd_IsConstant(f) && Cudd_IsConstant(g) {
			return f.Value <= g.Value
		}
		key := [2]*DdNode{f, g}
		if known[key] {
			return memo[key]
		}
		level := min(dd.level(f), dd.level(g))
		ft, fe := dd.addCofactors(f, level)
		gt, ge := dd.addCofactors(g, level)
		result := visit(ft, gt) && visit(fe, ge)
		known[key] = true
		memo[key] = result
		return result
	}
	return visit(f, g)
}
