package cudd

import core "github.com/glycerine/abcgo/bdd/cudd/core"

/**CFile***********************************************************************

  FileName    [cuddAddWalsh.c]

  PackageName [cudd]

  Synopsis    [Functions that generate Walsh matrices and residue
  functions in ADD form.]

  Description [External procedures included in this module:
            <ul>
            <li> Cudd_addWalsh()
            <li> Cudd_addResidue()
            </ul>
        Static procedures included in this module:
            <ul>
            <li> addWalshInt()
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

//#include "misc/util/util_hack.h"
//#include "cuddInt.h"

//ABC_NAMESPACE_IMPL_START

/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

// C source revision identifier omitted.

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static DdNode * addWalshInt (DdManager *dd, DdNode **x, DdNode **y, int n);

/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/*
*Function********************************************************************

	Synopsis    [Generates a Walsh matrix in ADD form.]

	Description [Generates a Walsh matrix in ADD form. Returns a pointer
	to the matrixi if successful; NULL otherwise.]

	SideEffects [None]

*****************************************************************************
*/
func Cudd_addWalsh(dd *DdManager, x **DdNode, y **DdNode, n int32) *DdNode {
	var res *DdNode

	for {
		dd.reordered = 0
		res = addWalshInt(dd, x, y, n)

		if !(dd.reordered == 1) {
			break
		}
	}
	return (res)

} /* end of Cudd_addWalsh */

/*
*Function********************************************************************

	Synopsis    [Builds an ADD for the residue modulo m of an n-bit
	number.]

	Description [Builds an ADD for the residue modulo m of an n-bit
	number. The modulus must be at least 2, and the number of bits at
	least 1. Parameter options specifies whether the MSB should be on top
	or the LSB; and whther the number whose residue is computed is in
	two's complement notation or not. The macro CUDD_RESIDUE_DEFAULT
	specifies LSB on top and unsigned number. The macro CUDD_RESIDUE_MSB
	specifies MSB on top, and the macro CUDD_RESIDUE_TC specifies two's
	complement residue. To request MSB on top and two's complement residue
	simultaneously, one can OR the two macros:
	CUDD_RESIDUE_MSB | CUDD_RESIDUE_TC.
	Cudd_addResidue returns a pointer to the resulting ADD if successful;
	NULL otherwise.]

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
func Cudd_addResidue(dd *DdManager, n, modulus, options, top int32) *DdNode {
	return core.Cudd_addResidue(dd, int(n), int(modulus), int(options), int(top))
}

/*
*Function********************************************************************

	Synopsis    [Implements the recursive step of Cudd_addWalsh.]

	Description [Generates a Walsh matrix in ADD form. Returns a pointer
	to the matrixi if successful; NULL otherwise.]

	SideEffects [None]

*****************************************************************************
*/
func addWalshInt(dd *DdManager, x **DdNode, y **DdNode, n int32) *DdNode {
	var one, minusone *DdNode
	var t *DdNode = nil
	var u *DdNode = nil
	var t1 *DdNode
	var u1 *DdNode
	var v *DdNode
	var w *DdNode
	var i int32

	// inlining of macro DD_ONE
	one = ((dd).one)
	if n == 0 {
		return (one)
	}

	/* Build bottom part of ADD outside loop */
	minusone = cuddUniqueConst(dd, (CUDD_VALUE_TYPE)-1)
	if minusone == nil {
		return (nil)
	}
	// inlining of macro cuddRef
	cuddSatInc(Cudd_Regular(minusone).ref)
	v = Cudd_addIte(dd, y[n-1], minusone, one)
	if v == nil {
		Cudd_RecursiveDeref(dd, minusone)
		return (nil)
	}
	// inlining of macro cuddRef
	cuddSatInc(Cudd_Regular(v).ref)
	u = Cudd_addIte(dd, x[n-1], v, one)
	if u == nil {
		Cudd_RecursiveDeref(dd, minusone)
		Cudd_RecursiveDeref(dd, v)
		return (nil)
	}
	// inlining of macro cuddRef
	cuddSatInc(Cudd_Regular(u).ref)
	Cudd_RecursiveDeref(dd, v)
	if n > 1 {
		w = Cudd_addIte(dd, y[n-1], one, minusone)
		if w == nil {
			Cudd_RecursiveDeref(dd, minusone)
			Cudd_RecursiveDeref(dd, u)
			return (nil)
		}
		// inlining of macro cuddRef
		cuddSatInc(Cudd_Regular(w).ref)
		t = Cudd_addIte(dd, x[n-1], w, minusone)
		if t == nil {
			Cudd_RecursiveDeref(dd, minusone)
			Cudd_RecursiveDeref(dd, u)
			Cudd_RecursiveDeref(dd, w)
			return (nil)
		}
		// inlining of macro cuddRef
		cuddSatInc(Cudd_Regular(t).ref)
		Cudd_RecursiveDeref(dd, w)
	}
	// inlining of macro cuddDeref
	cuddSatDec(Cudd_Regular(minusone).ref) /* minusone is in the result; it won't die */

	/* Loop to build the rest of the ADD */
	for i = n - 2; i >= 0; i-- {
		t1 = t
		u1 = u
		v = Cudd_addIte(dd, y[i], t1, u1)
		if v == nil {
			Cudd_RecursiveDeref(dd, u1)
			Cudd_RecursiveDeref(dd, t1)
			return (nil)
		}
		// inlining of macro cuddRef
		cuddSatInc(Cudd_Regular(v).ref)
		u = Cudd_addIte(dd, x[i], v, u1)
		if u == nil {
			Cudd_RecursiveDeref(dd, u1)
			Cudd_RecursiveDeref(dd, t1)
			Cudd_RecursiveDeref(dd, v)
			return (nil)
		}
		// inlining of macro cuddRef
		cuddSatInc(Cudd_Regular(u).ref)
		Cudd_RecursiveDeref(dd, v)
		if i > 0 {
			w = Cudd_addIte(dd, y[i], u1, t1)
			if w == nil {
				Cudd_RecursiveDeref(dd, u1)
				Cudd_RecursiveDeref(dd, t1)
				Cudd_RecursiveDeref(dd, u)
				return (nil)
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(w).ref)
			t = Cudd_addIte(dd, x[i], w, t1)
			if u == nil {
				Cudd_RecursiveDeref(dd, u1)
				Cudd_RecursiveDeref(dd, t1)
				Cudd_RecursiveDeref(dd, u)
				Cudd_RecursiveDeref(dd, w)
				return (nil)
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(t).ref)
			Cudd_RecursiveDeref(dd, w)
		}
		Cudd_RecursiveDeref(dd, u1)
		Cudd_RecursiveDeref(dd, t1)
	}

	// inlining of macro cuddDeref
	cuddSatDec(Cudd_Regular(u).ref)
	return (u)

} /* end of addWalshInt */

//ABC_NAMESPACE_IMPL_END
