package cudd

/**CFile***********************************************************************

  FileName    [cuddLiteral.c]

  PackageName [cudd]

  Synopsis    [Functions for manipulation of literal sets represented by
  BDDs.]

  Description [External procedures included in this file:
                <ul>
                <li> Cudd_bddLiteralSetIntersection()
                </ul>
            Internal procedures included in this file:
                <ul>
                <li> cuddBddLiteralSetIntersectionRecur()
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

import (
	"math"
	"reflect"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// Cudd_bddLiteralSetIntersection returns the cube containing the literals
// common to f and g with the same phase, following cuddLiteral.c.
func Cudd_bddLiteralSetIntersection(dd *DdManager, f, g *DdNode) *DdNode {
	return cuddBddLiteralSetIntersectionRecur(dd, f, g)
}

func cuddLiteralLevel(dd *DdManager, n *DdNode) int {
	r := Cudd_Regular(n)
	if Cudd_IsConstant(r) {
		return math.MaxInt
	}
	return dd.Perm[r.Index]
}

// cuddLiteralNext follows the nonzero child of a cube and reports its phase.
func cuddLiteralNext(dd *DdManager, n *DdNode) (*DdNode, bool) {
	r := Cudd_Regular(n)
	complement := Cudd_IsComplement(n)
	child := Cudd_NotCond(Cudd_T(r), complement)
	if child != dd.LogicZero {
		return child, true
	}
	return Cudd_NotCond(Cudd_E(r), complement), false
}

func cuddBddLiteralSetIntersectionRecur(dd *DdManager, f, g *DdNode) *DdNode {
	if f == g {
		return f
	}
	F, G := Cudd_Regular(f), Cudd_Regular(g)
	if F == G {
		return dd.One
	}
	topf, topg := cuddLiteralLevel(dd, f), cuddLiteralLevel(dd, g)
	for topf != topg {
		if topf < topg {
			f, _ = cuddLiteralNext(dd, f)
			F = Cudd_Regular(f)
			topf = cuddLiteralLevel(dd, f)
		} else {
			g, _ = cuddLiteralNext(dd, g)
			G = Cudd_Regular(g)
			topg = cuddLiteralLevel(dd, g)
		}
	}
	if f == dd.One {
		return dd.One
	}
	op := reflect.ValueOf(Cudd_bddLiteralSetIntersection).Pointer()
	if cached := core.Cudd_CacheLookup(dd, op, f, g, nil, 0); cached != nil {
		return cached
	}
	fc, phasef := cuddLiteralNext(dd, f)
	gc, phaseg := cuddLiteralNext(dd, g)
	tmp := cuddBddLiteralSetIntersectionRecur(dd, fc, gc)
	if tmp == nil {
		return nil
	}
	result := tmp
	if phasef == phaseg {
		Cudd_Ref(tmp)
		literal := dd.Vars[F.Index]
		if !phasef {
			literal = Cudd_Not(literal)
		}
		result = core.Cudd_bddAnd(dd, literal, tmp)
		if result == nil {
			Cudd_RecursiveDeref(dd, tmp)
			return nil
		}
		Cudd_Deref(tmp)
	}
	core.Cudd_CacheInsert(dd, op, f, g, nil, result, 0)
	return result
}
