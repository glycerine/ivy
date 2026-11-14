package cudd

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/st"
)

/**CFile***********************************************************************

  FileName    [cuddPriority.c]

  PackageName [cudd]

  Synopsis    [Priority functions.]

  Description [External procedures included in this file:
            <ul>
            <li> Cudd_PrioritySelect()
            <li> Cudd_Xgty()
            <li> Cudd_Xeqy()
            <li> Cudd_addXeqy()
            <li> Cudd_Dxygtdxz()
            <li> Cudd_Dxygtdyz()
            <li> Cudd_Inequality()
            <li> Cudd_Disequality()
            <li> Cudd_bddInterval()
            <li> Cudd_CProjection()
            <li> Cudd_addHamming()
            <li> Cudd_MinHammingDist()
            <li> Cudd_bddClosestCube()
            </ul>
        Internal procedures included in this module:
            <ul>
            <li> cuddCProjectionRecur()
            <li> cuddBddClosestCube()
            </ul>
        Static procedures included in this module:
            <ul>
            <li> cuddMinHammingDistRecur()
            <li> separateCube()
            <li> createResult()
            </ul>
            ]

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

//#include "misc/util/util_hack.h"
//#include "cuddInt.h"

//ABC_NAMESPACE_IMPL_START

/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

const DD_DEBUG = 1

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/
// static int cuddMinHammingDistRecur (DdNode * f, int *minterm, DdHashTable * table, int upperBound);

// static DdNode * separateCube (DdManager *dd, DdNode *f, CUDD_VALUE_TYPE *distance);

// static DdNode * createResult (DdManager *dd, unsigned int index, unsigned int phase, DdNode *cube, CUDD_VALUE_TYPE distance);

/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/*
*Function********************************************************************

	Synopsis    [Selects pairs from R using a priority function.]

	Description [Selects pairs from a relation R(x,y) (given as a BDD)
	in such a way that a given x appears in one pair only. Uses a
	priority function to determine which y should be paired to a given x.
	Cudd_PrioritySelect returns a pointer to
	the selected function if successful; NULL otherwise.
	Three of the arguments--x, y, and z--are vectors of BDD variables.
	The first two are the variables on which R depends. The third vectore
	is a vector of auxiliary variables, used during the computation. This
	vector is optional. If a NULL value is passed instead,
	Cudd_PrioritySelect will create the working variables on the fly.
	The sizes of x and y (and z if it is not NULL) should equal n.
	The priority function Pi can be passed as a BDD, or can be built by
	Cudd_PrioritySelect. If NULL is passed instead of a DdNode *,
	parameter Pifunc is used by Cudd_PrioritySelect to build a BDD for the
	priority function. (Pifunc is a pointer to a C function.) If Pi is not
	NULL, then Pifunc is ignored. Pifunc should have the same interface as
	the standard priority functions (e.g., Cudd_Dxygtdxz).
	Cudd_PrioritySelect and Cudd_CProjection can sometimes be used
	interchangeably. Specifically, calling Cudd_PrioritySelect with
	Cudd_Xgty as Pifunc produces the same result as calling
	Cudd_CProjection with the all-zero minterm as reference minterm.
	However, depending on the application, one or the other may be
	preferable:
	<ul>
	<li> When extracting representatives from an equivalence relation,
	Cudd_CProjection has the advantage of nor requiring the auxiliary
	variables.
	<li> When computing matchings in general bipartite graphs,
	Cudd_PrioritySelect normally obtains better results because it can use
	more powerful matching schemes (e.g., Cudd_Dxygtdxz).
	</ul>
	]

	SideEffects [If called with z == NULL, will create new variables in
	the manager.]

	SeeAlso     [Cudd_Dxygtdxz Cudd_Dxygtdyz Cudd_Xgty
	Cudd_bddAdjPermuteX Cudd_CProjection]

*****************************************************************************
*/
func Cudd_PrioritySelect(dd *DdManager, r *DdNode, x, y, z []*DdNode, pi *DdNode, n int, pifunc DD_PRFP) *DdNode {
	if dd == nil || r == nil || n < 0 || len(x) < n || len(y) < n {
		return nil
	}
	createdZ := z == nil
	if createdZ {
		if pi != nil {
			return nil
		}
		z = make([]*DdNode, n)
		for i := 0; i < n; i++ {
			z[i] = core.Cudd_bddIthVar(dd, len(dd.Vars))
			if z[i] == nil {
				return nil
			}
		}
	} else if len(z) < n {
		return nil
	}
	if pi == nil {
		if pifunc == nil {
			return nil
		}
		pi = pifunc(dd, n, x, y, z)
		if pi == nil {
			return nil
		}
		Cudd_Ref(pi)
		defer Cudd_RecursiveDeref(dd, pi)
	}
	zcube := dd.One
	Cudd_Ref(zcube)
	for i := n - 1; i >= 0; i-- {
		next := core.Cudd_bddAnd(dd, z[i], zcube)
		if next == nil {
			Cudd_RecursiveDeref(dd, zcube)
			return nil
		}
		Cudd_Ref(next)
		Cudd_RecursiveDeref(dd, zcube)
		zcube = next
	}
	_ = createdZ
	vector := append([]*DdNode(nil), dd.Vars...)
	for i := 0; i < n; i++ {
		if x[i] == nil || y[i] == nil || z[i] == nil {
			Cudd_RecursiveDeref(dd, zcube)
			return nil
		}
		yi, zi := int(Cudd_Regular(y[i]).Index), int(Cudd_Regular(z[i]).Index)
		if yi < 0 || yi >= len(vector) || zi < 0 || zi >= len(vector) {
			Cudd_RecursiveDeref(dd, zcube)
			return nil
		}
		vector[yi], vector[zi] = z[i], y[i]
	}
	rxz := core.Cudd_bddVectorCompose(dd, r, vector)
	if rxz == nil {
		Cudd_RecursiveDeref(dd, zcube)
		return nil
	}
	Cudd_Ref(rxz)
	q := core.Cudd_bddAndAbstract(dd, rxz, pi, zcube)
	Cudd_RecursiveDeref(dd, rxz)
	Cudd_RecursiveDeref(dd, zcube)
	if q == nil {
		return nil
	}
	Cudd_Ref(q)
	result := core.Cudd_bddAnd(dd, r, Cudd_Not(q))
	Cudd_RecursiveDeref(dd, q)
	return result
} /* Cudd_PrioritySelect */

/*
*Function********************************************************************

	Synopsis    [Generates a BDD for the function x &gt; y.]

	Description [This function generates a BDD for the function x &gt; y.
	Both x and y are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\] and
	y\[0\] y\[1\] ...  y\[N-1\], with 0 the most significant bit.
	The BDD is built bottom-up.
	It has 3*N-1 internal nodes, if the variables are ordered as follows:
	x\[0\] y\[0\] x\[1\] y\[1\] ... x\[N-1\] y\[N-1\].
	Argument z is not used by Cudd_Xgty: it is included to make it
	call-compatible to Cudd_Dxygtdxz and Cudd_Dxygtdyz.]

	SideEffects [None]

	SeeAlso     [Cudd_PrioritySelect Cudd_Dxygtdxz Cudd_Dxygtdyz]

*****************************************************************************
*/
func Cudd_Xgty(dd *DdManager, n int, z, x, y []*DdNode) *DdNode {
	_ = z // z is present only for compatibility with DD_PRFP.
	if n <= 0 || len(x) < n || len(y) < n {
		return nil
	}
	u := core.Cudd_bddAnd(dd, x[n-1], core.Cudd_Not(y[n-1]))
	if u == nil {
		return nil
	}
	Cudd_Ref(u)
	for i := n - 2; i >= 0; i-- {
		v := core.Cudd_bddAnd(dd, y[i], core.Cudd_Not(u))
		if v == nil {
			Cudd_RecursiveDeref(dd, u)
			return nil
		}
		Cudd_Ref(v)
		w := core.Cudd_bddAnd(dd, core.Cudd_Not(y[i]), u)
		if w == nil {
			Cudd_RecursiveDeref(dd, u)
			Cudd_RecursiveDeref(dd, v)
			return nil
		}
		Cudd_Ref(w)
		Cudd_RecursiveDeref(dd, u)
		u = core.Cudd_bddIte(dd, x[i], core.Cudd_Not(v), w)
		if u == nil {
			Cudd_RecursiveDeref(dd, v)
			Cudd_RecursiveDeref(dd, w)
			return nil
		}
		Cudd_Ref(u)
		Cudd_RecursiveDeref(dd, v)
		Cudd_RecursiveDeref(dd, w)
	}
	Cudd_Deref(u)
	return u
} /* end of Cudd_Xgty */

/*
*Function********************************************************************

	Synopsis    [Generates a BDD for the function x==y.]

	Description [This function generates a BDD for the function x==y.
	Both x and y are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\] and
	y\[0\] y\[1\] ...  y\[N-1\], with 0 the most significant bit.
	The BDD is built bottom-up.
	It has 3*N-1 internal nodes, if the variables are ordered as follows:
	x\[0\] y\[0\] x\[1\] y\[1\] ... x\[N-1\] y\[N-1\]. ]

	SideEffects [None]

	SeeAlso     [Cudd_addXeqy]

*****************************************************************************
*/
func Cudd_Xeqy(dd *DdManager, n int, x, y []*DdNode) *DdNode {
	if n <= 0 || len(x) < n || len(y) < n {
		return nil
	}
	u := core.Cudd_bddIte(dd, x[n-1], y[n-1], core.Cudd_Not(y[n-1]))
	if u == nil {
		return nil
	}
	Cudd_Ref(u)
	for i := n - 2; i >= 0; i-- {
		v := core.Cudd_bddAnd(dd, y[i], u)
		if v == nil {
			Cudd_RecursiveDeref(dd, u)
			return nil
		}
		Cudd_Ref(v)
		w := core.Cudd_bddAnd(dd, core.Cudd_Not(y[i]), u)
		if w == nil {
			Cudd_RecursiveDeref(dd, u)
			Cudd_RecursiveDeref(dd, v)
			return nil
		}
		Cudd_Ref(w)
		Cudd_RecursiveDeref(dd, u)
		u = core.Cudd_bddIte(dd, x[i], v, w)
		if u == nil {
			Cudd_RecursiveDeref(dd, v)
			Cudd_RecursiveDeref(dd, w)
			return nil
		}
		Cudd_Ref(u)
		Cudd_RecursiveDeref(dd, v)
		Cudd_RecursiveDeref(dd, w)
	}
	Cudd_Deref(u)
	return u
} /* end of Cudd_Xeqy */

/*
*Function********************************************************************

	Synopsis    [Generates an ADD for the function x==y.]

	Description [This function generates an ADD for the function x==y.
	Both x and y are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\] and
	y\[0\] y\[1\] ...  y\[N-1\], with 0 the most significant bit.
	The ADD is built bottom-up.
	It has 3*N-1 internal nodes, if the variables are ordered as follows:
	x\[0\] y\[0\] x\[1\] y\[1\] ... x\[N-1\] y\[N-1\]. ]

	SideEffects [None]

	SeeAlso     [Cudd_Xeqy]

*****************************************************************************
*/
func Cudd_addXeqy(dd *DdManager, n int, x, y []*DdNode) *DdNode {
	if n <= 0 || len(x) < n || len(y) < n {
		return nil
	}
	one, zero := dd.One, dd.Zero
	v := core.Cudd_addIte(dd, y[n-1], one, zero)
	if v == nil {
		return nil
	}
	Cudd_Ref(v)
	w := core.Cudd_addIte(dd, y[n-1], zero, one)
	if w == nil {
		Cudd_RecursiveDeref(dd, v)
		return nil
	}
	Cudd_Ref(w)
	u := core.Cudd_addIte(dd, x[n-1], v, w)
	if u == nil {
		Cudd_RecursiveDeref(dd, v)
		Cudd_RecursiveDeref(dd, w)
		return nil
	}
	Cudd_Ref(u)
	Cudd_RecursiveDeref(dd, v)
	Cudd_RecursiveDeref(dd, w)
	for i := n - 2; i >= 0; i-- {
		v = core.Cudd_addIte(dd, y[i], u, zero)
		if v == nil {
			Cudd_RecursiveDeref(dd, u)
			return nil
		}
		Cudd_Ref(v)
		w = core.Cudd_addIte(dd, y[i], zero, u)
		if w == nil {
			Cudd_RecursiveDeref(dd, u)
			Cudd_RecursiveDeref(dd, v)
			return nil
		}
		Cudd_Ref(w)
		Cudd_RecursiveDeref(dd, u)
		u = core.Cudd_addIte(dd, x[i], v, w)
		if u == nil {
			Cudd_RecursiveDeref(dd, v)
			Cudd_RecursiveDeref(dd, w)
			return nil
		}
		Cudd_Ref(u)
		Cudd_RecursiveDeref(dd, v)
		Cudd_RecursiveDeref(dd, w)
	}
	Cudd_Deref(u)
	return u
} /* end of Cudd_addXeqy */

/*
*Function********************************************************************

	Synopsis    [Generates a BDD for the function d(x,y) &gt; d(x,z).]

	Description [This function generates a BDD for the function d(x,y)
	&gt; d(x,z);
	x, y, and z are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\],
	y\[0\] y\[1\] ...  y\[N-1\], and z\[0\] z\[1\] ...  z\[N-1\],
	with 0 the most significant bit.
	The distance d(x,y) is defined as:
	      \sum_{i=0}^{N-1}(|x_i - y_i| \cdot 2^{N-i-1}).
	The BDD is built bottom-up.
	It has 7*N-3 internal nodes, if the variables are ordered as follows:
	x\[0\] y\[0\] z\[0\] x\[1\] y\[1\] z\[1\] ... x\[N-1\] y\[N-1\] z\[N-1\]. ]

	SideEffects [None]

	SeeAlso     [Cudd_PrioritySelect Cudd_Dxygtdyz Cudd_Xgty Cudd_bddAdjPermuteX]

*****************************************************************************
*/
func bddDistanceGreater(dd *DdManager, n int, a, b, c, d []*DdNode) *DdNode {
	result := dd.LogicZero
	for i := n - 1; i >= 0; i-- {
		left := core.Cudd_bddXor(dd, a[i], b[i])
		right := core.Cudd_bddXor(dd, c[i], d[i])
		if left == nil || right == nil {
			return nil
		}
		whenLeft := core.Cudd_bddIte(dd, right, result, dd.One)
		whenNotLeft := core.Cudd_bddIte(dd, right, dd.LogicZero, result)
		if whenLeft == nil || whenNotLeft == nil {
			return nil
		}
		result = core.Cudd_bddIte(dd, left, whenLeft, whenNotLeft)
		if result == nil {
			return nil
		}
	}
	return result
}

func Cudd_Dxygtdxz(dd *DdManager, n int, x, y, z []*DdNode) *DdNode {
	if n <= 0 || len(x) < n || len(y) < n || len(z) < n {
		return nil
	}
	return bddDistanceGreater(dd, n, x, y, x, z)
} /* end of Cudd_Dxygtdxz */

/*
*Function********************************************************************

	Synopsis    [Generates a BDD for the function d(x,y) &gt; d(y,z).]

	Description [This function generates a BDD for the function d(x,y)
	&gt; d(y,z);
	x, y, and z are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\],
	y\[0\] y\[1\] ...  y\[N-1\], and z\[0\] z\[1\] ...  z\[N-1\],
	with 0 the most significant bit.
	The distance d(x,y) is defined as:
	      \sum_{i=0}^{N-1}(|x_i - y_i| \cdot 2^{N-i-1}).
	The BDD is built bottom-up.
	It has 7*N-3 internal nodes, if the variables are ordered as follows:
	x\[0\] y\[0\] z\[0\] x\[1\] y\[1\] z\[1\] ... x\[N-1\] y\[N-1\] z\[N-1\]. ]

	SideEffects [None]

	SeeAlso     [Cudd_PrioritySelect Cudd_Dxygtdxz Cudd_Xgty Cudd_bddAdjPermuteX]

*****************************************************************************
*/
func Cudd_Dxygtdyz(dd *DdManager, n int, x, y, z []*DdNode) *DdNode {
	if n <= 0 || len(x) < n || len(y) < n || len(z) < n {
		return nil
	}
	return bddDistanceGreater(dd, n, x, y, y, z)
} /* end of Cudd_Dxygtdyz */

/*
*Function********************************************************************

	Synopsis    [Generates a BDD for the function x - y &ge; c.]

	Description [This function generates a BDD for the function x -y &ge; c.
	Both x and y are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\] and
	y\[0\] y\[1\] ...  y\[N-1\], with 0 the most significant bit.
	The BDD is built bottom-up.
	It has a linear number of nodes if the variables are ordered as follows:
	x\[0\] y\[0\] x\[1\] y\[1\] ... x\[N-1\] y\[N-1\].]

	SideEffects [None]

	SeeAlso     [Cudd_Xgty]

*****************************************************************************
*/
func Cudd_Inequality(dd *DdManager, n, c int, x, y []*DdNode) *DdNode {
	if n < 0 || len(x) < n || len(y) < n {
		return nil
	}
	if n == 0 {
		if c <= 0 {
			return dd.One
		}
		return dd.LogicZero
	}
	if n >= 62 {
		return nil
	}
	maxDiff := (int64(1) << n) - 1
	if int64(c) > maxDiff {
		return dd.LogicZero
	}
	if int64(c) <= -maxDiff {
		return dd.One
	}
	type state struct {
		bit    int
		prefix int64
	}
	memo := make(map[state]*DdNode)
	var build func(int, int64) *DdNode
	build = func(bit int, prefix int64) *DdNode {
		remaining := (int64(1) << (n - bit)) - 1
		if prefix-remaining >= int64(c) {
			return dd.One
		}
		if prefix+remaining < int64(c) {
			return dd.LogicZero
		}
		if bit == n {
			if prefix >= int64(c) {
				return dd.One
			}
			return dd.LogicZero
		}
		key := state{bit, prefix}
		if old := memo[key]; old != nil {
			return old
		}
		w := int64(1) << (n - bit - 1)
		f00 := build(bit+1, prefix)
		f01 := build(bit+1, prefix-w)
		f10 := build(bit+1, prefix+w)
		if f00 == nil || f01 == nil || f10 == nil {
			return nil
		}
		whenX0 := core.Cudd_bddIte(dd, y[bit], f01, f00)
		whenX1 := core.Cudd_bddIte(dd, y[bit], f00, f10)
		if whenX0 == nil || whenX1 == nil {
			return nil
		}
		result := core.Cudd_bddIte(dd, x[bit], whenX1, whenX0)
		if result != nil {
			memo[key] = result
		}
		return result
	}
	return build(0, 0)
} /* end of Cudd_Inequality */

/*
*Function********************************************************************

	Synopsis    [Generates a BDD for the function x - y != c.]

	Description [This function generates a BDD for the function x -y != c.
	Both x and y are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\] and
	y\[0\] y\[1\] ...  y\[N-1\], with 0 the most significant bit.
	The BDD is built bottom-up.
	It has a linear number of nodes if the variables are ordered as follows:
	x\[0\] y\[0\] x\[1\] y\[1\] ... x\[N-1\] y\[N-1\].]

	SideEffects [None]

	SeeAlso     [Cudd_Xgty]

*****************************************************************************
*/
func Cudd_Disequality(dd *DdManager, n, c int, x, y []*DdNode) *DdNode {
	if c == int(^uint(0)>>1) {
		return nil
	}
	geNext := Cudd_Inequality(dd, n, c+1, x, y)
	geCurrent := Cudd_Inequality(dd, n, c, x, y)
	if geNext == nil || geCurrent == nil {
		return nil
	}
	return core.Cudd_bddOr(dd, geNext, Cudd_Not(geCurrent))
} /* end of Cudd_Disequality */

/*
*Function********************************************************************

	Synopsis    [Generates a BDD for the function lowerB &le; x &le; upperB.]

	Description [This function generates a BDD for the function
	lowerB &le; x &le; upperB, where x is an N-bit number,
	x\[0\] x\[1\] ... x\[N-1\], with 0 the most significant bit (important!).
	The number of variables N should be sufficient to represent the bounds;
	otherwise, the bounds are truncated to their N least significant bits.
	Two BDDs are built bottom-up for lowerB &le; x and x &le; upperB, and they
	are finally conjoined.]

	SideEffects [None]

	SeeAlso     [Cudd_Xgty]

*****************************************************************************
*/
func Cudd_bddInterval(dd *DdManager, n int, x []*DdNode, lower, upper uint32) *DdNode {
	if n < 0 || len(x) < n || n >= 32 || lower > upper {
		return nil
	}
	if n == 0 {
		if lower == 0 {
			return dd.One
		}
		return dd.LogicZero
	}
	maxValue := uint64(1)<<n - 1
	if uint64(lower) > maxValue {
		return dd.LogicZero
	}
	if uint64(upper) > maxValue {
		upper = uint32(maxValue)
	}
	type state struct {
		bit    int
		prefix uint64
	}
	memo := make(map[state]*DdNode)
	var build func(int, uint64) *DdNode
	build = func(bit int, prefix uint64) *DdNode {
		remaining := (uint64(1) << (n - bit)) - 1
		if prefix+remaining < uint64(lower) || prefix > uint64(upper) {
			return dd.LogicZero
		}
		if prefix >= uint64(lower) && prefix+remaining <= uint64(upper) {
			return dd.One
		}
		if bit == n {
			if prefix >= uint64(lower) && prefix <= uint64(upper) {
				return dd.One
			}
			return dd.LogicZero
		}
		key := state{bit, prefix}
		if old := memo[key]; old != nil {
			return old
		}
		w := uint64(1) << (n - bit - 1)
		lo, hi := build(bit+1, prefix), build(bit+1, prefix+w)
		if lo == nil || hi == nil {
			return nil
		}
		result := core.Cudd_bddIte(dd, x[bit], hi, lo)
		if result != nil {
			memo[key] = result
		}
		return result
	}
	return build(0, 0)
} /* end of Cudd_bddInterval */

/*
*Function********************************************************************

	Synopsis    [Computes the compatible projection of R w.r.t. cube Y.]

	Description [Computes the compatible projection of relation R with
	respect to cube Y. Returns a pointer to the c-projection if
	successful; NULL otherwise. For a comparison between Cudd_CProjection
	and Cudd_PrioritySelect, see the documentation of the latter.]

	SideEffects [None]

	SeeAlso     [Cudd_PrioritySelect]

*****************************************************************************
*/
func Cudd_CProjection(dd *DdManager, r, cube *DdNode) *DdNode {
	if dd == nil || r == nil || cuddCheckCube(dd, cube) == 0 {
		return nil
	}
	vars, ok := cubeLiterals(dd, cube)
	if !ok {
		return nil
	}
	return projectCompatible(dd, r, vars)
} /* end of Cudd_CProjection */

/*
*Function********************************************************************

	Synopsis    [Computes the Hamming distance ADD.]

	Description [Computes the Hamming distance ADD. Returns an ADD that
	gives the Hamming distance between its two arguments if successful;
	NULL otherwise. The two vectors xVars and yVars identify the variables
	that form the two arguments.]

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
func Cudd_addHamming(dd *DdManager, xVars, yVars []*DdNode, nVars int32) *DdNode {
	if dd == nil || nVars < 0 || int(nVars) > len(xVars) || int(nVars) > len(yVars) {
		return nil
	}
	result := dd.Zero
	for i := int32(0); i < nVars; i++ {
		mismatch := core.Cudd_bddXor(dd, xVars[i], yVars[i])
		if mismatch == nil {
			return nil
		}
		addMismatch := core.Cudd_BddToAdd(dd, mismatch)
		if addMismatch == nil {
			return nil
		}
		result = core.Cudd_addApply(dd, core.Cudd_addPlus, result, addMismatch)
		if result == nil {
			return nil
		}
	}
	return result
} /* end of Cudd_addHamming */

/*
*Function********************************************************************

	Synopsis    [Returns the minimum Hamming distance between f and minterm.]

	Description [Returns the minimum Hamming distance between the
	minterms of a function f and a reference minterm. The function is
	given as a BDD; the minterm is given as an array of integers, one
	for each variable in the manager.  Returns the minimum distance if
	it is less than the upper bound; the upper bound if the minimum
	distance is at least as large; CUDD_OUT_OF_MEM in case of failure.]

	SideEffects [None]

	SeeAlso     [Cudd_addHamming Cudd_bddClosestCube]

*****************************************************************************
*/
func Cudd_MinHammingDist(dd *DdManager, f *DdNode, minterm []int32, upperBound int32) int32 {
	if dd == nil || f == nil || upperBound < 0 {
		return CUDD_OUT_OF_MEM
	}
	table := st.New(nil, nil)
	defer table.Clear()
	return cuddMinHammingDistRecur(f, minterm, table, dd, upperBound)
} /* end of Cudd_MinHammingDist */

/*
*Function********************************************************************

	Synopsis    [Finds a cube of f at minimum Hamming distance from g.]

	Description [Finds a cube of f at minimum Hamming distance from the
	minterms of g.  All the minterms of the cube are at the minimum
	distance.  If the distance is 0, the cube belongs to the
	intersection of f and g.  Returns the cube if successful; NULL
	otherwise.]

	SideEffects [The distance is returned as a side effect.]

	SeeAlso     [Cudd_MinHammingDist]

*****************************************************************************
*/
func Cudd_bddClosestCube(dd *DdManager, f, g *DdNode, distance *int32) *DdNode {
	if dd == nil || f == nil || g == nil {
		return nil
	}
	cube, d := closestCubeSearch(dd, f, g)
	if cube == nil {
		return nil
	}
	if distance != nil {
		*distance = int32(d)
	}
	return cube
} /* end of Cudd_bddClosestCube */

/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/*
*Function********************************************************************

	Synopsis    [Performs the recursive step of Cudd_CProjection.]

	Description [Performs the recursive step of Cudd_CProjection. Returns
	the projection if successful; NULL otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_CProjection]

*****************************************************************************
*/
func cuddCProjectionRecur(dd *DdManager, r, cube, _ *DdNode) *DdNode {
	vars, ok := cubeLiterals(dd, cube)
	if !ok {
		return nil
	}
	return projectCompatible(dd, r, vars)
}

func cubeLiterals(dd *DdManager, cube *DdNode) ([]*DdNode, bool) {
	var result []*DdNode
	zero := dd.LogicZero
	for cube != dd.One {
		if cube == nil || core.Cudd_IsConstant(cube) {
			return nil, false
		}
		regular := Cudd_Regular(cube)
		t, e := Cudd_T(regular), Cudd_E(regular)
		if Cudd_IsComplement(cube) {
			t, e = Cudd_Not(t), Cudd_Not(e)
		}
		variable := core.Cudd_bddIthVar(dd, int(regular.Index))
		if variable == nil {
			return nil, false
		}
		if e == zero {
			result = append(result, variable)
			cube = t
		} else if t == zero {
			result = append(result, Cudd_Not(variable))
			cube = e
		} else {
			return nil, false
		}
	}
	return result, true
}

func projectCompatible(dd *DdManager, relation *DdNode, variables []*DdNode) *DdNode {
	if relation == nil {
		return nil
	}
	if len(variables) == 0 {
		return relation
	}
	variable := variables[0]
	low := core.Cudd_bddRestrict(dd, relation, Cudd_Not(variable))
	high := core.Cudd_bddRestrict(dd, relation, variable)
	if low == nil || high == nil {
		return nil
	}
	remaining := dd.One
	for _, v := range variables[1:] {
		remaining = core.Cudd_bddAnd(dd, remaining, v)
		if remaining == nil {
			return nil
		}
	}
	canChooseLow := core.Cudd_bddExistAbstract(dd, low, remaining)
	if canChooseLow == nil {
		return nil
	}
	lowPart := core.Cudd_bddAnd(dd, low, Cudd_Not(variable))
	highPart := core.Cudd_bddAnd(dd, high, variable)
	if lowPart == nil || highPart == nil {
		return nil
	}
	chosen := core.Cudd_bddIte(dd, canChooseLow, lowPart, highPart)
	if chosen == nil {
		return nil
	}
	return projectCompatible(dd, chosen, variables[1:])
} /* end of cuddCProjectionRecur */

/*
*Function********************************************************************

	Synopsis    [Performs the recursive step of Cudd_bddClosestCube.]

	Description [Performs the recursive step of Cudd_bddClosestCube.
	Returns the cube if succesful; NULL otherwise.  The procedure uses a
	four-way recursion to examine all four combinations of cofactors of
	<code>f</code> and <code>g</code> according to the following formula.
	<pre>
	  H(f,g) = min(H(ft,gt), H(fe,ge), H(ft,ge)+1, H(fe,gt)+1)
	</pre>
	Bounding is based on the following observations.
	<ul>
	<li> If we already found two points at distance 0, there is no point in
	     continuing.  Furthermore,
	<li> If F == not(G) then the best we can hope for is a minimum distance
	     of 1.  If we have already found two points at distance 1, there is
	     no point in continuing.  (Indeed, H(F,G) == 1 in this case.  We
	     have to continue, though, to find the cube.)
	</ul>
	The variable <code>bound</code> is set at the largest value of the distance
	that we are still interested in.  Therefore, we desist when
	<pre>
	  (bound == -1) and (F != not(G)) or (bound == 0) and (F == not(G)).
	</pre>
	If we were maximally aggressive in using the bound, we would always
	set the bound to the minimum distance seen thus far minus one.  That
	is, we would maintain the invariant
	<pre>
	  bound < minD,
	</pre>
	except at the very beginning, when we have no value for
	<code>minD</code>.<p>

	However, we do not use <code>bound < minD</code> when examining the
	two negative cofactors, because we try to find a large cube at
	minimum distance.  To do so, we try to find a cube in the negative
	cofactors at the same or smaller distance from the cube found in the
	positive cofactors.<p>

	When we compute <code>H(ft,ge)</code> and <code>H(fe,gt)</code> we
	know that we are going to add 1 to the result of the recursive call
	to account for the difference in the splitting variable.  Therefore,
	we decrease the bound correspondingly.<p>

	Another important observation concerns the need of examining all
	four pairs of cofators only when both <code>f</code> and
	<code>g</code> depend on the top variable.<p>

	Suppose <code>gt == ge == g</code>.  (That is, <code>g</code> does
	not depend on the top variable.)  Then
	<pre>
	  H(f,g) = min(H(ft,g), H(fe,g), H(ft,g)+1, H(fe,g)+1)
	         = min(H(ft,g), H(fe,g)) .
	</pre>
	Therefore, under these circumstances, we skip the two "cross" cases.<p>

	An interesting feature of this function is the scheme used for
	caching the results in the global computed table.  Since we have a
	cube and a distance, we combine them to form an ADD.  The
	combination replaces the zero child of the top node of the cube with
	the negative of the distance.  (The use of the negative is to avoid
	ambiguity with 1.)  The degenerate cases (zero and one) are treated
	specially because the distance is known (0 for one, and infinity for
	zero).]

	SideEffects [None]

	SeeAlso     [Cudd_bddClosestCube]

*****************************************************************************
*/
func cuddBddClosestCube(dd *DdManager, f, g *DdNode, bound CUDD_VALUE_TYPE) *DdNode {
	cube, distance := closestCubeSearch(dd, f, g)
	if cube == nil || CUDD_VALUE_TYPE(distance) > bound {
		return nil
	}
	return cube
}

type closestState struct {
	f, g  *DdNode
	level int
}
type closestResult struct {
	cost        int
	fAssignment []bool
	valid       bool
}

func closestCubeSearch(dd *DdManager, f, g *DdNode) (*DdNode, int) {
	if dd == nil || f == nil || g == nil {
		return nil, 0
	}
	variables := dd.Vars
	memo := make(map[closestState]closestResult)
	var visit func(int, *DdNode, *DdNode) closestResult
	visit = func(level int, left, right *DdNode) closestResult {
		if left == dd.LogicZero || right == dd.LogicZero {
			return closestResult{}
		}
		if level == len(variables) {
			return closestResult{valid: true, fAssignment: []bool{}}
		}
		key := closestState{left, right, level}
		if result, ok := memo[key]; ok {
			return result
		}
		variable := variables[level]
		left0 := core.Cudd_bddRestrict(dd, left, Cudd_Not(variable))
		left1 := core.Cudd_bddRestrict(dd, left, variable)
		right0 := core.Cudd_bddRestrict(dd, right, Cudd_Not(variable))
		right1 := core.Cudd_bddRestrict(dd, right, variable)
		var best closestResult
		for fb := 0; fb <= 1; fb++ {
			for gb := 0; gb <= 1; gb++ {
				lf, rg := left0, right0
				if fb == 1 {
					lf = left1
				}
				if gb == 1 {
					rg = right1
				}
				child := visit(level+1, lf, rg)
				if !child.valid {
					continue
				}
				cost := child.cost
				if fb != gb {
					cost++
				}
				if !best.valid || cost < best.cost {
					best = closestResult{cost: cost, fAssignment: append([]bool{fb == 1}, child.fAssignment...), valid: true}
				}
			}
		}
		memo[key] = best
		return best
	}
	result := visit(0, f, g)
	if !result.valid {
		return nil, 0
	}
	cube := dd.One
	for i, value := range result.fAssignment {
		literal := variables[i]
		if !value {
			literal = Cudd_Not(literal)
		}
		cube = core.Cudd_bddAnd(dd, cube, literal)
		if cube == nil {
			return nil, 0
		}
	}
	return cube, result.cost
} /* end of cuddBddClosestCube */

/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/

/*
*Function********************************************************************

	Synopsis    [Performs the recursive step of Cudd_MinHammingDist.]

	Description [Performs the recursive step of Cudd_MinHammingDist.
	It is based on the following identity. Let H(f) be the
	minimum Hamming distance of the minterms of f from the reference
	minterm. Then:
	<xmp>
	  H(f) = min(H(f0)+h0,H(f1)+h1)
	</xmp>
	where f0 and f1 are the two cofactors of f with respect to its top
	variable; h0 is 1 if the minterm assigns 1 to the top variable of f;
	h1 is 1 if the minterm assigns 0 to the top variable of f.
	The upper bound on the distance is used to bound the depth of the
	recursion.
	Returns the minimum distance unless it exceeds the upper bound or
	computation fails.]

	SideEffects [None]

	SeeAlso     [Cudd_MinHammingDist]

*****************************************************************************
*/
func cuddMinHammingDistRecur(f *DdNode, minterm []int32, table *st.Table, dd *DdManager, upperBound int32) int32 {
	if f == nil || table == nil || dd == nil || upperBound < 0 {
		return CUDD_OUT_OF_MEM
	}
	if core.Cudd_IsConstant(f) {
		if f == dd.LogicZero {
			return upperBound
		}
		return 0
	}
	if value, ok := table.Lookup(f); ok {
		cached := value.(int32)
		if cached > upperBound {
			return upperBound
		}
		return cached
	}
	index := int(Cudd_Regular(f).Index)
	if index >= len(minterm) {
		return CUDD_OUT_OF_MEM
	}
	t, e := core.Cudd_T(f), core.Cudd_E(f)
	if minterm[index] == 0 {
		t, e = e, t
	}
	best := cuddMinHammingDistRecur(t, minterm, table, dd, upperBound)
	if best > 0 && upperBound > 0 {
		other := cuddMinHammingDistRecur(e, minterm, table, dd, upperBound-1)
		if other < upperBound && other+1 < best {
			best = other + 1
		}
	}
	if best < upperBound {
		table.Insert(f, best)
	}
	return best
} /* end of cuddMinHammingDistRecur */

/*
*Function********************************************************************

	Synopsis    [Separates cube from distance.]

	Description [Separates cube from distance.  Returns the cube if
	successful; NULL otherwise.]

	SideEffects [The distance is returned as a side effect.]

	SeeAlso     [cuddBddClosestCube createResult]

*****************************************************************************
*/
func separateCube(dd *DdManager, f *DdNode, distance *CUDD_VALUE_TYPE) *DdNode {
	if dd == nil || f == nil || distance == nil {
		return nil
	}
	if f == dd.One {
		*distance = 0
		return f
	}
	if f == dd.LogicZero {
		*distance = CUDD_VALUE_TYPE(CUDD_CONST_INDEX + 1)
		return f
	}
	return nil
} /* end of separateCube */

/*
*Function********************************************************************

	Synopsis    [Builds a result for cache storage.]

	Description [Builds a result for cache storage.  Returns a pointer
	to the resulting ADD if successful; NULL otherwise.]

	SideEffects [None]

	SeeAlso     [cuddBddClosestCube separateCube]

*****************************************************************************
*/
func createResult(dd *DdManager, index uint32, phase bool, cube *DdNode, distance CUDD_VALUE_TYPE) *DdNode {
	if dd == nil || cube == nil {
		return nil
	}
	if index == CUDD_CONST_INDEX && core.Cudd_IsConstant(cube) {
		return cube
	}
	constant := core.Cudd_addConst(dd, -distance)
	if constant == nil {
		return nil
	}
	if index == CUDD_CONST_INDEX {
		regular := Cudd_Regular(cube)
		selector := core.Cudd_bddIthVar(dd, int(regular.Index))
		t, e := core.Cudd_T(regular), core.Cudd_E(regular)
		if t == dd.Zero {
			return core.Cudd_addIte(dd, selector, constant, e)
		}
		return core.Cudd_addIte(dd, selector, t, constant)
	}
	selector := core.Cudd_bddIthVar(dd, int(index))
	if selector == nil {
		return nil
	}
	if phase {
		return core.Cudd_addIte(dd, selector, cube, constant)
	}
	return core.Cudd_addIte(dd, selector, constant, cube)
} /* end of createResult */

//ABC_NAMESPACE_IMPL_END
