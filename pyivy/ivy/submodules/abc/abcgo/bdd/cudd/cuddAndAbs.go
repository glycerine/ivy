package cudd

import core "github.com/glycerine/abcgo/bdd/cudd/core"

/**CFile***********************************************************************

  FileName    [cuddAndAbs.c]

  PackageName [cudd]

  Synopsis    [Combined AND and existential abstraction for BDDs]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_bddAndAbstract()
                <li> Cudd_bddAndAbstractLimit()
                </ul>
            Internal procedures included in this module:
                <ul>
                <li> cuddBddAndAbstractRecur()
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

/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/*
*Function********************************************************************

	Synopsis [Takes the AND of two BDDs and simultaneously abstracts the
	variables in cube.]

	Description [Takes the AND of two BDDs and simultaneously abstracts
	the variables in cube. The variables are existentially abstracted.
	Returns a pointer to the result is successful; NULL otherwise.
	Cudd_bddAndAbstract implements the semiring matrix multiplication
	algorithm for the boolean semiring.]

	SideEffects [None]

	SeeAlso     [Cudd_addMatrixMultiply Cudd_addTriangle Cudd_bddAnd]

*****************************************************************************
*/
func Cudd_bddAndAbstract(manager *DdManager, f, g, cube *DdNode) *DdNode {
	return core.Cudd_bddAndAbstract(manager, f, g, cube)
}

func Cudd_bddAndAbstractLimit(manager *DdManager, f, g, cube *DdNode, limit uint32) *DdNode {
	return core.Cudd_bddAndAbstractLimit(manager, f, g, cube, limit)
}

func cuddBddAndAbstractRecur(manager *DdManager, f, g, cube *DdNode) *DdNode {
	return core.Cudd_bddAndAbstract(manager, f, g, cube)
}
