/**CFile***********************************************************************

  FileName    [cuddAddApply.c]

  PackageName [cudd]

  Synopsis    [Apply functions for ADDs and their operators.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addApply()
                <li> Cudd_addMonadicApply()
                <li> Cudd_addPlus()
                <li> Cudd_addTimes()
                <li> Cudd_addThreshold()
                <li> Cudd_addSetNZ()
                <li> Cudd_addDivide()
                <li> Cudd_addMinus()
                <li> Cudd_addMinimum()
                <li> Cudd_addMaximum()
                <li> Cudd_addOneZeroMaximum()
                <li> Cudd_addDiff()
                <li> Cudd_addAgreement()
                <li> Cudd_addOr()
                <li> Cudd_addNand()
                <li> Cudd_addNor()
                <li> Cudd_addXor()
                <li> Cudd_addXnor()
                </ul>
            Internal procedures included in this module:
                <ul>
                <li> cuddAddApplyRecur()
                <li> cuddAddMonadicApplyRecur()
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

import "math"

// Apply callbacks retain CUDD's protocol: return a simplified result or nil
// when Shannon recursion is required. Callback-local operand swaps are honored.
func Cudd_addApply(dd *DdManager, op DD_AOP, f, g *DdNode) *DdNode {
	if !dd.validAdd(f, g) || op == nil {
		return nil
	}
	memo := map[[2]*DdNode]*DdNode{}
	var visit func(*DdNode, *DdNode) *DdNode
	visit = func(f, g *DdNode) *DdNode {
		if result := op(dd, &f, &g); result != nil {
			if dd.validAdd(result) {
				return result
			}
			return nil
		}
		if !dd.validAdd(f, g) || Cudd_IsConstant(f) && Cudd_IsConstant(g) {
			return nil
		}
		key := [2]*DdNode{f, g}
		if n := memo[key]; n != nil {
			return n
		}
		level := min(dd.level(f), dd.level(g))
		ft, fe := dd.addCofactors(f, level)
		gt, ge := dd.addCofactors(g, level)
		t := visit(ft, gt)
		if t == nil {
			return nil
		}
		e := visit(fe, ge)
		if e == nil {
			return nil
		}
		n := dd.addNode(uint32(dd.InvPerm[level]), t, e)
		memo[key] = n
		return n
	}
	return visit(f, g)
}
func Cudd_addMonadicApply(dd *DdManager, op DD_MAOP, f *DdNode) *DdNode {
	if !dd.validAdd(f) || op == nil {
		return nil
	}
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(f *DdNode) *DdNode {
		if n := op(dd, f); n != nil {
			if dd.validAdd(n) {
				return n
			}
			return nil
		}
		if Cudd_IsConstant(f) {
			return nil
		}
		if n := memo[f]; n != nil {
			return n
		}
		t, e := visit(f.T), visit(f.E)
		if t == nil || e == nil {
			return nil
		}
		n := dd.addNode(f.Index, t, e)
		memo[f] = n
		return n
	}
	return visit(f)
}
func addTerminal(dd *DdManager, f, g **DdNode, op func(float64, float64) float64) *DdNode {
	if !dd.validAdd(*f, *g) || !Cudd_IsConstant(*f) || !Cudd_IsConstant(*g) {
		return nil
	}
	return Cudd_addConst(dd, op((*f).Value, (*g).Value))
}
func addBoolean(dd *DdManager, value bool) *DdNode {
	if value {
		return dd.One
	}
	return dd.Zero
}
func Cudd_addPlus(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == dd.Zero {
		return *g
	}
	if *g == dd.Zero {
		return *f
	}
	return addTerminal(dd, f, g, func(a, b float64) float64 { return a + b })
}
func Cudd_addTimes(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == dd.Zero || *g == dd.Zero {
		return dd.Zero
	}
	if *f == dd.One {
		return *g
	}
	if *g == dd.One {
		return *f
	}
	return addTerminal(dd, f, g, func(a, b float64) float64 { return a * b })
}
func Cudd_addDivide(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == dd.Zero {
		return dd.Zero
	}
	if *g == dd.One {
		return *f
	}
	return addTerminal(dd, f, g, func(a, b float64) float64 { return a / b })
}
func Cudd_addMinus(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == *g {
		return dd.Zero
	}
	if *g == dd.Zero {
		return *f
	}
	return addTerminal(dd, f, g, func(a, b float64) float64 { return a - b })
}
func Cudd_addMinimum(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == *g {
		return *f
	}
	if *f == dd.PlusInfinity {
		return *g
	}
	if *g == dd.PlusInfinity {
		return *f
	}
	return addTerminal(dd, f, g, math.Min)
}
func Cudd_addMaximum(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == *g {
		return *f
	}
	if *f == dd.MinusInfinity {
		return *g
	}
	if *g == dd.MinusInfinity {
		return *f
	}
	return addTerminal(dd, f, g, math.Max)
}
func Cudd_addThreshold(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == *g || *f == dd.PlusInfinity {
		return *f
	}
	if Cudd_IsConstant(*f) && Cudd_IsConstant(*g) {
		if (*f).Value >= (*g).Value {
			return *f
		}
		return dd.Zero
	}
	return nil
}
func Cudd_addSetNZ(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == *g || *g == dd.Zero {
		return *f
	}
	if *f == dd.Zero || Cudd_IsConstant(*g) {
		return *g
	}
	return nil
}
func Cudd_addOneZeroMaximum(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == *g || *g == dd.PlusInfinity {
		return dd.Zero
	}
	if Cudd_IsConstant(*f) && Cudd_IsConstant(*g) {
		return addBoolean(dd, (*f).Value > (*g).Value)
	}
	return nil
}
func Cudd_addDiff(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == *g {
		return dd.PlusInfinity
	}
	return Cudd_addMinimum(dd, f, g)
}
func Cudd_addAgreement(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == *g || *f == dd.Background {
		return *f
	}
	if *g == dd.Background {
		return *g
	}
	if Cudd_IsConstant(*f) && Cudd_IsConstant(*g) {
		return dd.Background
	}
	return nil
}

// The following operations have the original 0/1 ADD domain contract.
func Cudd_addOr(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == dd.One || *g == dd.One {
		return dd.One
	}
	if *f == *g {
		return *f
	}
	if Cudd_IsConstant(*f) {
		return *g
	}
	if Cudd_IsConstant(*g) {
		return *f
	}
	return nil
}
func Cudd_addNand(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == dd.Zero || *g == dd.Zero {
		return dd.One
	}
	if Cudd_IsConstant(*f) && Cudd_IsConstant(*g) {
		return dd.Zero
	}
	return nil
}
func Cudd_addNor(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == dd.One || *g == dd.One {
		return dd.Zero
	}
	if Cudd_IsConstant(*f) && Cudd_IsConstant(*g) {
		return dd.One
	}
	return nil
}
func Cudd_addXor(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == *g {
		return dd.Zero
	}
	if Cudd_IsConstant(*f) && Cudd_IsConstant(*g) {
		return addBoolean(dd, (*f).Value != (*g).Value)
	}
	return nil
}
func Cudd_addXnor(dd *DdManager, f, g **DdNode) *DdNode {
	if *f == *g {
		return dd.One
	}
	if Cudd_IsConstant(*f) && Cudd_IsConstant(*g) {
		return addBoolean(dd, (*f).Value == (*g).Value)
	}
	return nil
}
func Cudd_addLog(dd *DdManager, f *DdNode) *DdNode {
	if Cudd_IsConstant(f) {
		return Cudd_addConst(dd, math.Log(f.Value))
	}
	return nil
}
func Cudd_addNegate(dd *DdManager, f *DdNode) *DdNode {
	return Cudd_addMonadicApply(dd, func(dd *DdManager, n *DdNode) *DdNode {
		if Cudd_IsConstant(n) {
			return Cudd_addConst(dd, -n.Value)
		}
		return nil
	}, f)
}
