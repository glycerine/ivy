/**CFile***********************************************************************

  FileName    [cuddBddAbs.c]

  PackageName [cudd]

  Synopsis    [Quantification functions for BDDs.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_bddExistAbstract()
                <li> Cudd_bddXorExistAbstract()
                <li> Cudd_bddUnivAbstract()
                <li> Cudd_bddBooleanDiff()
                <li> Cudd_bddVarIsDependent()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddBddExistAbstractRecur()
                <li> cuddBddXorExistAbstractRecur()
                <li> cuddBddBooleanDiffRecur()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> bddCheckPositiveCube()
                </ul>
                ]

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

func Cudd_bddExistAbstract(dd *DdManager, f, cube *DdNode) *DdNode {
	return dd.abstract(f, cube, false)
}

func Cudd_bddUnivAbstract(dd *DdManager, f, cube *DdNode) *DdNode { return dd.abstract(f, cube, true) }

func (dd *DdManager) positiveCube(cube *DdNode) bool {
	for cube != dd.One {
		if cube == nil || Cudd_IsConstant(cube) || Cudd_IsComplement(cube) || cube.E != dd.LogicZero {
			return false
		}
		cube = cube.T
	}
	return true
}
func Cudd_bddXorExistAbstract(dd *DdManager, f, g, cube *DdNode) *DdNode {
	if !dd.valid(f, g, cube) || !dd.positiveCube(cube) {
		return nil
	}
	return dd.abstract(Cudd_bddXor(dd, f, g), cube, false)
}
func Cudd_bddBooleanDiff(dd *DdManager, f *DdNode, index int) *DdNode {
	if !dd.valid(f) || index < 0 {
		return nil
	}
	if index >= len(dd.Vars) {
		return dd.LogicZero
	}
	variable := dd.Vars[index]
	return Cudd_bddXor(dd, Cudd_Cofactor(dd, f, variable), Cudd_Cofactor(dd, f, Cudd_Not(variable)))
}

// disjoint tests intersection emptiness without allocating DD nodes.
func (dd *DdManager) disjoint(f, g *DdNode, memo map[[2]*DdNode]bool, known map[[2]*DdNode]bool) bool {
	if f == dd.LogicZero || g == dd.LogicZero || f == Cudd_Not(g) {
		return true
	}
	if f == dd.One || g == dd.One || f == g {
		return false
	}
	key := [2]*DdNode{f, g}
	if known[key] {
		return memo[key]
	}
	level := min(dd.level(f), dd.level(g))
	ft, fe := dd.cofactors(f, level)
	gt, ge := dd.cofactors(g, level)
	result := dd.disjoint(ft, gt, memo, known) && dd.disjoint(fe, ge, memo, known)
	known[key] = true
	memo[key] = result
	return result
}
func Cudd_bddVarIsDependent(dd *DdManager, f, variable *DdNode) bool {
	if !dd.valid(f, variable) || Cudd_IsConstant(variable) {
		return false
	}
	level := dd.level(variable)
	memo := map[*DdNode]bool{}
	known := map[*DdNode]bool{}
	disjointMemo := map[[2]*DdNode]bool{}
	disjointKnown := map[[2]*DdNode]bool{}
	var visit func(*DdNode) bool
	visit = func(f *DdNode) bool {
		if Cudd_IsConstant(f) {
			return f == dd.LogicZero
		}
		if dd.level(f) > level {
			return false
		}
		if known[f] {
			return memo[f]
		}
		t, e := dd.cofactors(f, dd.level(f))
		var result bool
		if dd.level(f) == level {
			result = dd.disjoint(t, e, disjointMemo, disjointKnown)
		} else {
			result = visit(t) && visit(e)
		}
		known[f] = true
		memo[f] = result
		return result
	}
	return visit(f)
}
func cuddBddExistAbstractRecur(dd *DdManager, f, cube *DdNode) *DdNode {
	return Cudd_bddExistAbstract(dd, f, cube)
}
func cuddBddXorExistAbstractRecur(dd *DdManager, f, g, cube *DdNode) *DdNode {
	return Cudd_bddXorExistAbstract(dd, f, g, cube)
}
func cuddBddBooleanDiffRecur(dd *DdManager, f, variable *DdNode) *DdNode {
	if !dd.valid(variable) || Cudd_IsConstant(variable) {
		return nil
	}
	return Cudd_bddBooleanDiff(dd, f, int(Cudd_Regular(variable).Index))
}
