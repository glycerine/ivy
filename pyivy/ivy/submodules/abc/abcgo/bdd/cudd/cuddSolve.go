package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

/**CFile***********************************************************************

  FileName    [cuddSolve.c]

  PackageName [cudd]

  Synopsis    [Boolean equation solver and related functions.]

  Description [External functions included in this modoule:
                <ul>
                <li> Cudd_SolveEqn()
                <li> Cudd_VerifySol()
                </ul>
        Internal functions included in this module:
                <ul>
                <li> cuddSolveEqnRecur()
                <li> cuddVerifySol()
                </ul> ]

  SeeAlso     []

  Author      [Balakrishna Kumthekar]

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

//#include "misc/util/util_hack.h"
//#include "cuddInt.h"

//ABC_NAMESPACE_IMPL_START

// Cudd_SolveEqn returns the consistency condition, parametric solutions and
// their corresponding unknown variable indices, in cube order.
func Cudd_SolveEqn(bdd *DdManager, f, y *DdNode, n int) (*DdNode, []*DdNode, []int) {
	g := make([]*DdNode, n)
	indices := make([]int, n)
	result := cuddSolveEqnRecur(bdd, f, y, g, indices, 0)
	return result, g, indices
}

func Cudd_VerifySol(bdd *DdManager, f *DdNode, g []*DdNode, indices []int, n int) *DdNode {
	return cuddVerifySol(bdd, f, g, indices, n)
}

func cuddSolveEqnRecur(bdd *DdManager, f, y *DdNode, g []*DdNode, indices []int, i int) *DdNode {
	if y == bdd.One {
		return f
	}
	indices[i] = int(y.Index)
	variable := bdd.Vars[indices[i]]
	fm1 := core.Cudd_bddExistAbstract(bdd, core.Cudd_Not(f), variable)
	if fm1 == nil {
		return nil
	}
	fm1 = core.Cudd_Not(fm1)
	core.Cudd_Ref(fm1)
	fn := cuddSolveEqnRecur(bdd, fm1, y.T, g, indices, i+1)
	if fn == nil {
		core.Cudd_RecursiveDeref(bdd, fm1)
		return nil
	}
	core.Cudd_Ref(fn)
	fv := core.Cudd_Cofactor(bdd, f, variable)
	if fv == nil {
		core.Cudd_RecursiveDeref(bdd, fm1)
		core.Cudd_RecursiveDeref(bdd, fn)
		return nil
	}
	core.Cudd_Ref(fv)
	fvbar := core.Cudd_Cofactor(bdd, f, core.Cudd_Not(variable))
	if fvbar == nil {
		core.Cudd_RecursiveDeref(bdd, fm1)
		core.Cudd_RecursiveDeref(bdd, fn)
		core.Cudd_RecursiveDeref(bdd, fv)
		return nil
	}
	core.Cudd_Ref(fvbar)
	w := core.Cudd_bddIte(bdd, variable, core.Cudd_Not(fv), fvbar)
	if w == nil {
		for _, v := range []*DdNode{fm1, fn, fv, fvbar} {
			core.Cudd_RecursiveDeref(bdd, v)
		}
		return nil
	}
	core.Cudd_Ref(w)
	t := core.Cudd_bddRestrict(bdd, w, core.Cudd_Not(fm1))
	if t == nil {
		for _, v := range []*DdNode{fm1, fn, fv, fvbar, w} {
			core.Cudd_RecursiveDeref(bdd, v)
		}
		return nil
	}
	core.Cudd_Ref(t)
	for _, v := range []*DdNode{fm1, w, fv, fvbar} {
		core.Cudd_RecursiveDeref(bdd, v)
	}
	for j := len(g) - 1; j > i; j-- {
		w = core.Cudd_bddCompose(bdd, t, g[j], indices[j])
		if w == nil {
			core.Cudd_RecursiveDeref(bdd, fn)
			core.Cudd_RecursiveDeref(bdd, t)
			return nil
		}
		core.Cudd_Ref(w)
		core.Cudd_RecursiveDeref(bdd, t)
		t = w
	}
	g[i] = t
	core.Cudd_Deref(fn)
	return fn
}

func cuddVerifySol(bdd *DdManager, f *DdNode, g []*DdNode, indices []int, n int) *DdNode {
	r := f
	core.Cudd_Ref(r)
	for j := n - 1; j >= 0; j-- {
		w := core.Cudd_bddCompose(bdd, r, g[j], indices[j])
		if w == nil {
			return nil
		}
		core.Cudd_Ref(w)
		core.Cudd_RecursiveDeref(bdd, r)
		r = w
	}
	core.Cudd_Deref(r)
	return r
}
