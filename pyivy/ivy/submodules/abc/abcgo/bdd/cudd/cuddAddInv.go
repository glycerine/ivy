package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddAddInv.c]

  PackageName [cudd]

  Synopsis    [Function to compute the scalar inverse of an ADD.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addScalarInverse()
                </ul>
            Internal procedures included in this module:
                <ul>
                <li> cuddAddScalarInverseRecur()
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

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddAddInv.c,v 1.9 2004/08/13 18:04:45 fabio Exp $";
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

  Synopsis    [Computes the scalar inverse of an ADD.]
  
  Description [Computes an n ADD where the discriminants are the
  multiplicative inverses of the corresponding discriminants of the
  argument ADD.  Returns a pointer to the resulting ADD in case of
  success. Returns NULL if any discriminants smaller than epsilon is
  encountered.]

  SideEffects [None]

******************************************************************************/
func Cudd_addScalarInverse(dd *DdManager, f *DdNode, epsilon *DdNode) *DdNode {
    var res *DdNode;

    // inlining of macro cuddIsConstant
    if !((epsilon).index == CUDD_CONST_INDEX) {
        (void) fmt.Fprintf(dd.err, "Invalid epsilon\n");
        return(nil);
    }
    for {
        dd.reordered = 0;
        res  = cuddAddScalarInverseRecur(dd,f,epsilon);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_addScalarInverse */

/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of addScalarInverse.]

  Description [Returns a pointer to the resulting ADD in case of
  success. Returns NULL if any discriminants smaller than epsilon is
  encountered.]

  SideEffects [None]

******************************************************************************/
func cuddAddScalarInverseRecur(dd *DdManager, f *DdNode, epsilon *DdNode) *DdNode {
    var t, e, res *DdNode;
    CUDD_VALUE_TYPE value;

    statLine(dd);
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        // inlining of macro ddAbs
        // inlining of macro cuddV
        if (((((f).type.value))<0) ? -(((f).type.value)) : (((f).type.value))) < ((epsilon).type.value) { return(nil); }
        // inlining of macro cuddV
        value = 1.0 / ((f).type.value);
        res = cuddUniqueConst(dd,value);
        return(res);
    }

    res = cuddCacheLookup2(dd,Cudd_addScalarInverse,f,epsilon);
    if res != nil { return(res); }

    // inlining of macro cuddT
    t = cuddAddScalarInverseRecur(dd,((f).type.kids.T),epsilon);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    // inlining of macro cuddE
    e = cuddAddScalarInverseRecur(dd,((f).type.kids.E),epsilon);
    if e == nil {
        Cudd_RecursiveDeref(dd, t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    if t == e {
        res = t;
    } else {
        res = cuddUniqueInter(dd,int32(f.index),t,e);
    }
    if res == nil {
        Cudd_RecursiveDeref(dd, t);
        Cudd_RecursiveDeref(dd, e);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);

    cuddCacheInsert2(dd,Cudd_addScalarInverse,f,epsilon,res);

    return(res);

} /* end of cuddAddScalarInverseRecur */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


//ABC_NAMESPACE_IMPL_END


