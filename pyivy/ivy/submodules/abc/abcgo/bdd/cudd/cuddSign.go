package cudd

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"math"
)

/**CFile***********************************************************************

  FileName    [cuddSign.c]

  PackageName [cudd]

  Synopsis    [Computation of signatures.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_CofMinterm();
                    </ul>
                Static procedures included in this module:
                    <ul>
                    <li> ddCofMintermAux()
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

//#include "misc/util/util_hack.h"
//#include "cuddInt.h"

//ABC_NAMESPACE_IMPL_START

// Cudd_CofMinterm returns the positive-cofactor onset fractions in variable
// index order, followed by the onset fraction of node itself. The C source
// indexes outside its one-element allocation for constant roots unless the
// manager has exactly one variable; Go detects this invalid source case.
func Cudd_CofMinterm(dd *DdManager, node *DdNode) []float64 {
	size := len(dd.Vars)
	values := ddCofMintermAux(dd, node, make(map[*DdNode][]float64))
	firstLevel := cuddSignLevel(dd, node)
	if core.Cudd_IsConstant(node) {
		firstLevel = 1
	}
	result := make([]float64, size+1)
	for i := 0; i < size; i++ {
		if i >= cuddSignLevel(dd, node) {
			result[dd.InvPerm[i]] = values[i-firstLevel]
		} else {
			result[dd.InvPerm[i]] = values[size-firstLevel]
		}
	}
	result[size] = values[size-firstLevel]
	return result
}

func cuddSignLevel(dd *DdManager, n *DdNode) int {
	n = core.Cudd_Regular(n)
	if core.Cudd_IsConstant(n) {
		return math.MaxInt
	}
	return dd.Perm[n.Index]
}

func ddCofMintermAux(dd *DdManager, node *DdNode, memo map[*DdNode][]float64) []float64 {
	if v, ok := memo[node]; ok {
		return v
	}
	n := core.Cudd_Regular(node)
	if core.Cudd_IsConstant(n) {
		if node == dd.Zero || node == dd.LogicZero {
			return []float64{0}
		}
		return []float64{1}
	}
	level := cuddSignLevel(dd, node)
	t := core.Cudd_NotCond(n.T, core.Cudd_IsComplement(node))
	e := core.Cudd_NotCond(n.E, core.Cudd_IsComplement(node))
	vt, ve := ddCofMintermAux(dd, t, memo), ddCofMintermAux(dd, e, memo)
	result := make([]float64, len(dd.Vars)-level+1)
	result[0] = vt[len(vt)-1]
	for i := 1; i < len(result); i++ {
		a, b := vt[len(vt)-1], ve[len(ve)-1]
		if delta := cuddSignLevel(dd, t) - level; i >= delta {
			a = vt[i-delta]
		}
		if delta := cuddSignLevel(dd, e) - level; i >= delta {
			b = ve[i-delta]
		}
		result[i] = (a + b) / 2
	}
	if n.Ref > 1 {
		memo[node] = result
	}
	return result
}
