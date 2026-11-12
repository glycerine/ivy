package cudd

/**CFile***********************************************************************

  FileName    [cuddAddFind.c]

  PackageName [cudd]

  Synopsis    [Functions to find maximum and minimum in an ADD and to
  extract the i-th bit.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addFindMax()
                <li> Cudd_addFindMin()
                <li> Cudd_addIthBit()
                </ul>
               Static functions included in this module:
                <ul>
                <li> addDoIthBit()
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
static char rcsid[] DD_UNUSED = "$Id: cuddAddFind.c,v 1.8 2004/08/13 18:04:45 fabio Exp $";
#endif


/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static DdNode * addDoIthBit (DdManager *dd, DdNode *f, DdNode *index);


/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Finds the maximum discriminant of f.]

  Description [Returns a pointer to a constant ADD.]

  SideEffects [None]

******************************************************************************/
func Cudd_addFindMax(dd *DdManager, f *DdNode) *DdNode {
    var t, e, res *DdNode;

    statLine(dd);
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        return(f);
    }

    res = cuddCacheLookup1(dd,Cudd_addFindMax,f);
    if res != nil {
        return(res);
    }

    // inlining of macro cuddT
    t  = Cudd_addFindMax(dd,((f).type.kids.T));
    // inlining of macro DD_PLUS_INFINITY
    if t == ((dd).plusinfinity) { return(t); }

    // inlining of macro cuddE
    e  = Cudd_addFindMax(dd,((f).type.kids.E));

    // inlining of macro cuddV
    if ((t).type.value) >= ((e).type.value) {
        res = t;
    } else {
        res = e;
    }

    cuddCacheInsert1(dd,Cudd_addFindMax,f,res);

    return(res);

} /* end of Cudd_addFindMax */


/**Function********************************************************************

  Synopsis    [Finds the minimum discriminant of f.]

  Description [Returns a pointer to a constant ADD.]

  SideEffects [None]

******************************************************************************/
func Cudd_addFindMin(dd *DdManager, f *DdNode) *DdNode {
    var t, e, res *DdNode;

    statLine(dd);
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        return(f);
    }

    res = cuddCacheLookup1(dd,Cudd_addFindMin,f);
    if res != nil {
        return(res);
    }

    // inlining of macro cuddT
    t  = Cudd_addFindMin(dd,((f).type.kids.T));
    // inlining of macro DD_MINUS_INFINITY
    if t == ((dd).minusinfinity) { return(t); }

    // inlining of macro cuddE
    e  = Cudd_addFindMin(dd,((f).type.kids.E));

    // inlining of macro cuddV
    if ((t).type.value) <= ((e).type.value) {
        res = t;
    } else {
        res = e;
    }

    cuddCacheInsert1(dd,Cudd_addFindMin,f,res);

    return(res);

} /* end of Cudd_addFindMin */


/**Function********************************************************************

  Synopsis    [Extracts the i-th bit from an ADD.]

  Description [Produces an ADD from another ADD by replacing all
  discriminants whose i-th bit is equal to 1 with 1, and all other
  discriminants with 0. The i-th bit refers to the integer
  representation of the leaf value. If the value is has a fractional
  part, it is ignored. Repeated calls to this procedure allow one to
  transform an integer-valued ADD into an array of ADDs, one for each
  bit of the leaf values. Returns a pointer to the resulting ADD if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addBddIthBit]

******************************************************************************/
func Cudd_addIthBit(dd *DdManager, f *DdNode, bit int32) *DdNode {
    var res *DdNode;
    var index *DdNode;
    
    /* Use a constant node to remember the bit, so that we can use the
    ** global cache.
    */
    index = cuddUniqueConst(dd,(CUDD_VALUE_TYPE) bit);
    if index == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(index).ref);

    for {
        dd.reordered = 0;
        res = addDoIthBit(dd, f, index);
    
if !(dd.reordered == 1) { break }
} 

    if res == nil {
        Cudd_RecursiveDeref(dd, index);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref(dd, index);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of Cudd_addIthBit */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step for Cudd_addIthBit.]

  Description [Performs the recursive step for Cudd_addIthBit.
  Returns a pointer to the BDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func addDoIthBit(dd *DdManager, f *DdNode, index *DdNode) *DdNode {
    var res, T, E *DdNode;
    var fv, fvn *DdNode;
    var mask, value int32;
    var v int32;

    statLine(dd);
    /* Check terminal case. */
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        // inlining of macro cuddV
        mask = 1 << (int32( ((index).type.value)));
        // inlining of macro cuddV
        value = int32( ((f).type.value));
        if (value & mask) == 0 {
            // inlining of macro DD_ZERO
            return ((dd).zero);
        } else {
            // inlining of macro DD_ONE
            return ((dd).one);
        }
    }

    /* Check cache. */
    res = cuddCacheLookup2(dd,addDoIthBit,f,index);
    if res != nil { return(res); }

    /* Recursive step. */
    v = f.index;
    // inlining of macro cuddT
    // inlining of macro cuddE
    fv = ((f).type.kids.T); fvn = ((f).type.kids.E);

    T = addDoIthBit(dd,fv,index);
    if T == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);

    E = addDoIthBit(dd,fvn,index);
    if E == nil {
        Cudd_RecursiveDeref(dd, T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);

    if T == E {
        res = T;
    } else {
        res = cuddUniqueInter(dd,v,T,E);
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
    cuddCacheInsert2(dd,addDoIthBit,f,index,res);

    return(res);

} /* end of addDoIthBit */


//ABC_NAMESPACE_IMPL_END


