package cudd

import "math"

/**CFile***********************************************************************

  FileName    [cuddAddNeg.c]

  PackageName [cudd]

  Synopsis    [Function to compute the negation of an ADD.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addNegate()
                <li> Cudd_addRoundOff()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddAddNegateRecur()
                <li> cuddAddRoundOffRecur()
                </ul> ]

  Author      [Fabio Somenzi, Balakrishna Kumthekar]

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

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddAddNeg.c,v 1.12 2009/02/20 02:14:58 fabio Exp $";
#endif


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

/**Function********************************************************************

  Synopsis    [Computes the additive inverse of an ADD.]

  Description [Computes the additive inverse of an ADD. Returns a pointer
  to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addCmpl]

******************************************************************************/
func Cudd_addNegate(dd *DdManager, f *DdNode) *DdNode {
    var res *DdNode;

    for {
        res = cuddAddNegateRecur(dd,f);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_addNegate */


/**Function********************************************************************

  Synopsis    [Rounds off the discriminants of an ADD.]

  Description [Rounds off the discriminants of an ADD. The discriminants are
  rounded off to N digits after the decimal. Returns a pointer to the result
  ADD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_addRoundOff(dd *DdManager, f *DdNode, N int32) *DdNode {
    var res *DdNode;
    var trunc float64  = math.Pow(float64(10.0), float64(float64(N)));

    for {
        res = cuddAddRoundOffRecur(dd,f,trunc);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_addRoundOff */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Implements the recursive step of Cudd_addNegate.]

  Description [Implements the recursive step of Cudd_addNegate.
  Returns a pointer to the result.]

  SideEffects [None]

******************************************************************************/
func cuddAddNegateRecur(dd *DdManager, f *DdNode) *DdNode {
    var res, fv, fvn, T, E *DdNode;

    statLine(dd);
    /* Check terminal cases. */
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        // inlining of macro cuddV
        res = cuddUniqueConst(dd,-((f).type.value));
        return(res);
    }

    /* Check cache */
    res = cuddCacheLookup1(dd,Cudd_addNegate,f);
    if res != nil { return(res); }

    /* Recursive Step */
    // inlining of macro cuddT
    fv = ((f).type.kids.T);
    // inlining of macro cuddE
    fvn = ((f).type.kids.E);
    T = cuddAddNegateRecur(dd,fv);
    if T == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);

    E = cuddAddNegateRecur(dd,fvn);
    if E == nil {
        Cudd_RecursiveDeref(dd,T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);
    if T == E {
        res = T;
    } else {
        res = cuddUniqueInter(dd,int32(f.index),T,E);
    }
    if res == nil {
        Cudd_RecursiveDeref(dd, T);
        Cudd_RecursiveDeref(dd, E);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(T).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(E).ref);

    /* Store result. */
    cuddCacheInsert1(dd,Cudd_addNegate,f,res);

    return(res);

} /* end of cuddAddNegateRecur */


/**Function********************************************************************

  Synopsis    [Implements the recursive step of Cudd_addRoundOff.]

  Description [Implements the recursive step of Cudd_addRoundOff.
  Returns a pointer to the result.]

  SideEffects [None]

******************************************************************************/
func cuddAddRoundOffRecur(dd *DdManager, f *DdNode, trunc float64) *DdNode {

    var res, fv, fvn, T, E *DdNode;
    var n float64;
    DD_CTFP1 cacheOp;

    statLine(dd);
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        // inlining of macro cuddV
        n = math.Ceil(float64(((f).type.value)*trunc))/trunc;
        res = cuddUniqueConst(dd,n);
        return(res);
    }
    cacheOp = (DD_CTFP1) Cudd_addRoundOff;
    res = cuddCacheLookup1(dd,cacheOp,f);
    if res != nil {
        return(res);
    }
    /* Recursive Step */
    // inlining of macro cuddT
    fv = ((f).type.kids.T);
    // inlining of macro cuddE
    fvn = ((f).type.kids.E);
    T = cuddAddRoundOffRecur(dd,fv,trunc);
    if T == nil {
       return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);
    E = cuddAddRoundOffRecur(dd,fvn,trunc);
    if E == nil {
        Cudd_RecursiveDeref(dd,T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);
    if T == E {
        res = T;
    } else {
        res = cuddUniqueInter(dd,int32(f.index),T,E);
    }
    if res == nil {
        Cudd_RecursiveDeref(dd,T);
        Cudd_RecursiveDeref(dd,E);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(T).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(E).ref);

    /* Store result. */
    cuddCacheInsert1(dd,cacheOp,f,res);
    return(res);

} /* end of cuddAddRoundOffRecur */

/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


//ABC_NAMESPACE_IMPL_END

