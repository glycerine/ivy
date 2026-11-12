package cudd

/**CFile***********************************************************************

  FileName    [cuddBridge.c]

  PackageName [cudd]

  Synopsis    [Translation from BDD to ADD and vice versa and transfer between
  different managers.]

  Description [External procedures included in this file:
            <ul>
            <li> Cudd_addBddThreshold()
            <li> Cudd_addBddStrictThreshold()
            <li> Cudd_addBddInterval()
            <li> Cudd_addBddIthBit()
            <li> Cudd_BddToAdd()
            <li> Cudd_addBddPattern()
            <li> Cudd_bddTransfer()
            </ul>
        Internal procedures included in this file:
            <ul>
            <li> cuddBddTransfer()
            <li> cuddAddBddDoPattern()
            </ul>
        Static procedures included in this file:
            <ul>
            <li> addBddDoThreshold()
            <li> addBddDoStrictThreshold()
            <li> addBddDoInterval()
            <li> addBddDoIthBit()
            <li> ddBddToAddRecur()
            <li> cuddBddTransferRecur()
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
static char rcsid[] DD_UNUSED = "$Id: cuddBridge.c,v 1.19 2008/04/25 06:42:55 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static DdNode * addBddDoThreshold (DdManager *dd, DdNode *f, DdNode *val);

// static DdNode * addBddDoStrictThreshold (DdManager *dd, DdNode *f, DdNode *val);

// static DdNode * addBddDoInterval (DdManager *dd, DdNode *f, DdNode *l, DdNode *u);

// static DdNode * addBddDoIthBit (DdManager *dd, DdNode *f, DdNode *index);

// static DdNode * ddBddToAddRecur (DdManager *dd, DdNode *B);

// static DdNode * cuddBddTransferRecur (DdManager *ddS, DdManager *ddD, DdNode *f, st__table *table);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Converts an ADD to a BDD.]

  Description [Converts an ADD to a BDD by replacing all
  discriminants greater than or equal to value with 1, and all other
  discriminants with 0. Returns a pointer to the resulting BDD if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addBddInterval Cudd_addBddPattern Cudd_BddToAdd
  Cudd_addBddStrictThreshold]

******************************************************************************/
func Cudd_addBddThreshold(dd *DdManager, f *DdNode, value CUDD_VALUE_TYPE) *DdNode {
    var res *DdNode;
    var val *DdNode;
    
    val = cuddUniqueConst(dd,value);
    if val == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(val).ref);

    for {
        dd.reordered = 0;
        res = addBddDoThreshold(dd, f, val);
    
if !(dd.reordered == 1) { break }
} 

    if res == nil {
        Cudd_RecursiveDeref(dd, val);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref(dd, val);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of Cudd_addBddThreshold */


/**Function********************************************************************

  Synopsis    [Converts an ADD to a BDD.]

  Description [Converts an ADD to a BDD by replacing all
  discriminants STRICTLY greater than value with 1, and all other
  discriminants with 0. Returns a pointer to the resulting BDD if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addBddInterval Cudd_addBddPattern Cudd_BddToAdd 
  Cudd_addBddThreshold]

******************************************************************************/
func Cudd_addBddStrictThreshold(dd *DdManager, f *DdNode, value CUDD_VALUE_TYPE) *DdNode {
    var res *DdNode;
    var val *DdNode;
    
    val = cuddUniqueConst(dd,value);
    if val == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(val).ref);

    for {
        dd.reordered = 0;
        res = addBddDoStrictThreshold(dd, f, val);
    
if !(dd.reordered == 1) { break }
} 

    if res == nil {
        Cudd_RecursiveDeref(dd, val);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref(dd, val);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of Cudd_addBddStrictThreshold */


/**Function********************************************************************

  Synopsis    [Converts an ADD to a BDD.]

  Description [Converts an ADD to a BDD by replacing all
  discriminants greater than or equal to lower and less than or equal to
  upper with 1, and all other discriminants with 0. Returns a pointer to
  the resulting BDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addBddThreshold Cudd_addBddStrictThreshold 
  Cudd_addBddPattern Cudd_BddToAdd]

******************************************************************************/
func Cudd_addBddInterval(dd *DdManager, f *DdNode, lower CUDD_VALUE_TYPE, upper CUDD_VALUE_TYPE) *DdNode {
    var res *DdNode;
    var l *DdNode;
    var u *DdNode;
    
    /* Create constant nodes for the interval bounds, so that we can use
    ** the global cache.
    */
    l = cuddUniqueConst(dd,lower);
    if l == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(l).ref);
    u = cuddUniqueConst(dd,upper);
    if u == nil {
        Cudd_RecursiveDeref(dd,l);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(u).ref);

    for {
        dd.reordered = 0;
        res = addBddDoInterval(dd, f, l, u);
    
if !(dd.reordered == 1) { break }
} 

    if res == nil {
        Cudd_RecursiveDeref(dd, l);
        Cudd_RecursiveDeref(dd, u);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref(dd, l);
    Cudd_RecursiveDeref(dd, u);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of Cudd_addBddInterval */


/**Function********************************************************************

  Synopsis    [Converts an ADD to a BDD by extracting the i-th bit from
  the leaves.]

  Description [Converts an ADD to a BDD by replacing all
  discriminants whose i-th bit is equal to 1 with 1, and all other
  discriminants with 0. The i-th bit refers to the integer
  representation of the leaf value. If the value is has a fractional
  part, it is ignored. Repeated calls to this procedure allow one to
  transform an integer-valued ADD into an array of BDDs, one for each
  bit of the leaf values. Returns a pointer to the resulting BDD if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addBddInterval Cudd_addBddPattern Cudd_BddToAdd]

******************************************************************************/
func Cudd_addBddIthBit(dd *DdManager, f *DdNode, bit int32) *DdNode {
    var res *DdNode;
    var index *DdNode;
    
    index = cuddUniqueConst(dd,(CUDD_VALUE_TYPE) bit);
    if index == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(index).ref);

    for {
        dd.reordered = 0;
        res = addBddDoIthBit(dd, f, index);
    
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

} /* end of Cudd_addBddIthBit */


/**Function********************************************************************

  Synopsis    [Converts a BDD to a 0-1 ADD.]

  Description [Converts a BDD to a 0-1 ADD. Returns a pointer to the
  resulting ADD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addBddPattern Cudd_addBddThreshold Cudd_addBddInterval
  Cudd_addBddStrictThreshold]

******************************************************************************/
func Cudd_BddToAdd(dd *DdManager, B *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = ddBddToAddRecur(dd, B);
    
if !(dd.reordered ==1) { break }
} 
    return(res);

} /* end of Cudd_BddToAdd */


/**Function********************************************************************

  Synopsis    [Converts an ADD to a BDD.]

  Description [Converts an ADD to a BDD by replacing all
  discriminants different from 0 with 1. Returns a pointer to the
  resulting BDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_BddToAdd Cudd_addBddThreshold Cudd_addBddInterval
  Cudd_addBddStrictThreshold]

******************************************************************************/
func Cudd_addBddPattern(dd *DdManager, f *DdNode) *DdNode {
    var res *DdNode;
    
    for {
        dd.reordered = 0;
        res = cuddAddBddDoPattern(dd, f);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_addBddPattern */


/**Function********************************************************************

  Synopsis    [Convert a BDD from a manager to another one.]

  Description [Convert a BDD from a manager to another one. The orders of the
  variables in the two managers may be different. Returns a
  pointer to the BDD in the destination manager if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_bddTransfer(ddSource *DdManager, ddDestination *DdManager, f *DdNode) *DdNode {
    var res *DdNode;
    for {
        ddDestination.reordered = 0;
        res = cuddBddTransfer(ddSource, ddDestination, f);
    
if !(ddDestination.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddTransfer */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Convert a BDD from a manager to another one.]

  Description [Convert a BDD from a manager to another one. Returns a
  pointer to the BDD in the destination manager if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddTransfer]

******************************************************************************/
func cuddBddTransfer(ddS *DdManager, ddD *DdManager, f *DdNode) *DdNode {
    var res *DdNode;
    var table *st__table  = nil;
    var gen *st__generator  = nil;
    var key, value *DdNode;

    table = st__init_table( st__ptrcmp, st__ptrhash);
    if table == nil { goto failure; }
    res = cuddBddTransferRecur(ddS, ddD, f, table);
    // inlining of macro cuddRef
    if res != nil { cuddSatInc(Cudd_Regular(res).ref); }

    /* Dereference all elements in the table and dispose of the table.
    ** This must be done also if res is NULL to avoid leaks in case of
    ** reordering. */
    gen = st__init_gen(table);
    if gen == nil { goto failure; }
    for  st__gen(gen, (char **)&key, (char **)&value) {
        Cudd_RecursiveDeref(ddD, value);
    }
    st__free_gen(gen); gen = nil;
    st__free_table(table); table = nil;

    // inlining of macro cuddDeref
    if res != nil { cuddSatDec(Cudd_Regular(res).ref); }
    return(res);

failure:
    /* No need to free gen because it is always NULL here. */
    if table != nil { st__free_table(table); }
    return(nil);

} /* end of cuddBddTransfer */


/**Function********************************************************************

  Synopsis    [Performs the recursive step for Cudd_addBddPattern.]

  Description [Performs the recursive step for Cudd_addBddPattern. Returns a
  pointer to the resulting BDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddAddBddDoPattern(dd *DdManager, f *DdNode) *DdNode {
    var res, T, E *DdNode;
    var fv, fvn *DdNode;
    var v int32;

    statLine(dd);
    /* Check terminal case. */
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        // inlining of macro Cudd_NotCond
        // inlining of macro DD_ONE
        // inlining of macro DD_ZERO
        return(((DdNode *)((ptrint)(((dd).one)) ^ (f == ((dd).zero)))));
    }

    /* Check cache. */
    res = cuddCacheLookup1(dd,Cudd_addBddPattern,f);
    if res != nil { return(res); }

    /* Recursive step. */
    v = f.index;
    // inlining of macro cuddT
    // inlining of macro cuddE
    fv = ((f).type.kids.T); fvn = ((f).type.kids.E);

    T = cuddAddBddDoPattern(dd,fv);
    if T == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);

    E = cuddAddBddDoPattern(dd,fvn);
    if E == nil {
        Cudd_RecursiveDeref(dd, T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (T) & 01)) != 0 {
        if T == E {
            // inlining of macro Cudd_Not
            res = ((DdNode *)((ptrint)(T) ^ 01));
        } else {
            // inlining of macro Cudd_Not
            res = cuddUniqueInter(dd,v,((DdNode *)((ptrint)(T) ^ 01)),((DdNode *)((ptrint)(E) ^ 01)));
        }
        if res == nil {
            Cudd_RecursiveDeref(dd, T);
            Cudd_RecursiveDeref(dd, E);
            return(nil);
        }
        // inlining of macro Cudd_Not
        res = ((DdNode *)((ptrint)(res) ^ 01));
    } else {
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
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(T).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(E).ref);

    /* Store result. */
    cuddCacheInsert1(dd,Cudd_addBddPattern,f,res);

    return(res);

} /* end of cuddAddBddDoPattern */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step for Cudd_addBddThreshold.]

  Description [Performs the recursive step for Cudd_addBddThreshold.
  Returns a pointer to the BDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [addBddDoStrictThreshold]

******************************************************************************/
func addBddDoThreshold(dd *DdManager, f *DdNode, val *DdNode) *DdNode {
    var res, T, E *DdNode;
    var fv, fvn *DdNode;
    var v int32;

    statLine(dd);
    /* Check terminal case. */
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        // inlining of macro Cudd_NotCond
        // inlining of macro DD_ONE
        // inlining of macro cuddV
        return(((DdNode *)((ptrint)(((dd).one)) ^ (((f).type.value) < ((val).type.value)))));
    }

    /* Check cache. */
    res = cuddCacheLookup2(dd,addBddDoThreshold,f,val);
    if res != nil { return(res); }

    /* Recursive step. */
    v = f.index;
    // inlining of macro cuddT
    // inlining of macro cuddE
    fv = ((f).type.kids.T); fvn = ((f).type.kids.E);

    T = addBddDoThreshold(dd,fv,val);
    if T == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);

    E = addBddDoThreshold(dd,fvn,val);
    if E == nil {
        Cudd_RecursiveDeref(dd, T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (T) & 01)) != 0 {
        if T == E {
            // inlining of macro Cudd_Not
            res = ((DdNode *)((ptrint)(T) ^ 01));
        } else {
            // inlining of macro Cudd_Not
            res = cuddUniqueInter(dd,v,((DdNode *)((ptrint)(T) ^ 01)),((DdNode *)((ptrint)(E) ^ 01)));
        }
        if res == nil {
            Cudd_RecursiveDeref(dd, T);
            Cudd_RecursiveDeref(dd, E);
            return(nil);
        }
        // inlining of macro Cudd_Not
        res = ((DdNode *)((ptrint)(res) ^ 01));
    } else {
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
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(T).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(E).ref);

    /* Store result. */
    cuddCacheInsert2(dd,addBddDoThreshold,f,val,res);

    return(res);

} /* end of addBddDoThreshold */


/**Function********************************************************************

  Synopsis    [Performs the recursive step for Cudd_addBddStrictThreshold.]

  Description [Performs the recursive step for Cudd_addBddStrictThreshold.
  Returns a pointer to the BDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [addBddDoThreshold]

******************************************************************************/
func addBddDoStrictThreshold(dd *DdManager, f *DdNode, val *DdNode) *DdNode {
    var res, T, E *DdNode;
    var fv, fvn *DdNode;
    var v int32;

    statLine(dd);
    /* Check terminal case. */
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        // inlining of macro Cudd_NotCond
        // inlining of macro DD_ONE
        // inlining of macro cuddV
        return(((DdNode *)((ptrint)(((dd).one)) ^ (((f).type.value) <= ((val).type.value)))));
    }

    /* Check cache. */
    res = cuddCacheLookup2(dd,addBddDoStrictThreshold,f,val);
    if res != nil { return(res); }

    /* Recursive step. */
    v = f.index;
    // inlining of macro cuddT
    // inlining of macro cuddE
    fv = ((f).type.kids.T); fvn = ((f).type.kids.E);

    T = addBddDoStrictThreshold(dd,fv,val);
    if T == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);

    E = addBddDoStrictThreshold(dd,fvn,val);
    if E == nil {
        Cudd_RecursiveDeref(dd, T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (T) & 01)) != 0 {
        if T == E {
            // inlining of macro Cudd_Not
            res = ((DdNode *)((ptrint)(T) ^ 01));
        } else {
            // inlining of macro Cudd_Not
            res = cuddUniqueInter(dd,v,((DdNode *)((ptrint)(T) ^ 01)),((DdNode *)((ptrint)(E) ^ 01)));
        }
        if res == nil {
            Cudd_RecursiveDeref(dd, T);
            Cudd_RecursiveDeref(dd, E);
            return(nil);
        }
        // inlining of macro Cudd_Not
        res = ((DdNode *)((ptrint)(res) ^ 01));
    } else {
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
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(T).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(E).ref);

    /* Store result. */
    cuddCacheInsert2(dd,addBddDoStrictThreshold,f,val,res);

    return(res);

} /* end of addBddDoStrictThreshold */


/**Function********************************************************************

  Synopsis    [Performs the recursive step for Cudd_addBddInterval.]

  Description [Performs the recursive step for Cudd_addBddInterval.
  Returns a pointer to the BDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [addBddDoThreshold addBddDoStrictThreshold]

******************************************************************************/
func addBddDoInterval(dd *DdManager, f *DdNode, l *DdNode, u *DdNode) *DdNode {
    var res, T, E *DdNode;
    var fv, fvn *DdNode;
    var v int32;

    statLine(dd);
    /* Check terminal case. */
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        // inlining of macro Cudd_NotCond
        // inlining of macro DD_ONE
        // inlining of macro cuddV
        return(((DdNode *)((ptrint)(((dd).one)) ^ (((f).type.value) < ((l).type.value) || ((f).type.value) > ((u).type.value)))));
    }

    /* Check cache. */
    res = cuddCacheLookup(dd,DD_ADD_BDD_DO_INTERVAL_TAG,f,l,u);
    if res != nil { return(res); }

    /* Recursive step. */
    v = f.index;
    // inlining of macro cuddT
    // inlining of macro cuddE
    fv = ((f).type.kids.T); fvn = ((f).type.kids.E);

    T = addBddDoInterval(dd,fv,l,u);
    if T == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);

    E = addBddDoInterval(dd,fvn,l,u);
    if E == nil {
        Cudd_RecursiveDeref(dd, T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (T) & 01)) != 0 {
        if T == E {
            // inlining of macro Cudd_Not
            res = ((DdNode *)((ptrint)(T) ^ 01));
        } else {
            // inlining of macro Cudd_Not
            res = cuddUniqueInter(dd,v,((DdNode *)((ptrint)(T) ^ 01)),((DdNode *)((ptrint)(E) ^ 01)));
        }
        if res == nil {
            Cudd_RecursiveDeref(dd, T);
            Cudd_RecursiveDeref(dd, E);
            return(nil);
        }
        // inlining of macro Cudd_Not
        res = ((DdNode *)((ptrint)(res) ^ 01));
    } else {
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
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(T).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(E).ref);

    /* Store result. */
    cuddCacheInsert(dd,DD_ADD_BDD_DO_INTERVAL_TAG,f,l,u,res);

    return(res);

} /* end of addBddDoInterval */


/**Function********************************************************************

  Synopsis    [Performs the recursive step for Cudd_addBddIthBit.]

  Description [Performs the recursive step for Cudd_addBddIthBit.
  Returns a pointer to the BDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func addBddDoIthBit(dd *DdManager, f *DdNode, index *DdNode) *DdNode {
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
        // inlining of macro Cudd_NotCond
        // inlining of macro DD_ONE
        return(((DdNode *)((ptrint)(((dd).one)) ^ ((value & mask) == 0))));
    }

    /* Check cache. */
    res = cuddCacheLookup2(dd,addBddDoIthBit,f,index);
    if res != nil { return(res); }

    /* Recursive step. */
    v = f.index;
    // inlining of macro cuddT
    // inlining of macro cuddE
    fv = ((f).type.kids.T); fvn = ((f).type.kids.E);

    T = addBddDoIthBit(dd,fv,index);
    if T == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);

    E = addBddDoIthBit(dd,fvn,index);
    if E == nil {
        Cudd_RecursiveDeref(dd, T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (T) & 01)) != 0 {
        if T == E {
            // inlining of macro Cudd_Not
            res = ((DdNode *)((ptrint)(T) ^ 01));
        } else {
            // inlining of macro Cudd_Not
            res = cuddUniqueInter(dd,v,((DdNode *)((ptrint)(T) ^ 01)),((DdNode *)((ptrint)(E) ^ 01)));
        }
        if res == nil {
            Cudd_RecursiveDeref(dd, T);
            Cudd_RecursiveDeref(dd, E);
            return(nil);
        }
        // inlining of macro Cudd_Not
        res = ((DdNode *)((ptrint)(res) ^ 01));
    } else {
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
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(T).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(E).ref);

    /* Store result. */
    cuddCacheInsert2(dd,addBddDoIthBit,f,index,res);

    return(res);

} /* end of addBddDoIthBit */


/**Function********************************************************************

  Synopsis    [Performs the recursive step for Cudd_BddToAdd.]

  Description [Performs the recursive step for Cudd_BddToAdd. Returns a
  pointer to the resulting ADD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func ddBddToAddRecur(dd *DdManager, B *DdNode) *DdNode {
    var one *DdNode;
    var res, res1, T, E, Bt, Be *DdNode;
    var complement int32;

    statLine(dd);
    // inlining of macro DD_ONE
    one = ((dd).one);

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(B) & ~01))).index == CUDD_CONST_INDEX {
        if B == one {
            res = one;
        } else {
            // inlining of macro DD_ZERO
            res = ((dd).zero);
        }
        return(res);
    }
    /* Check visited table */
    res = cuddCacheLookup1(dd,ddBddToAddRecur,B);
    if res != nil { return(res); }

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (B) & 01)) != 0 {
        complement = 1;
        // inlining of macro cuddT
        // inlining of macro Cudd_Regular
        Bt = ((((DdNode *)((ptruint)(B) & ~01))).type.kids.T);
        // inlining of macro cuddE
        // inlining of macro Cudd_Regular
        Be = ((((DdNode *)((ptruint)(B) & ~01))).type.kids.E);
    } else {
        // inlining of macro cuddT
        Bt = ((B).type.kids.T);
        // inlining of macro cuddE
        Be = ((B).type.kids.E);
    }

    T = ddBddToAddRecur(dd, Bt);
    if T == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);

    E = ddBddToAddRecur(dd, Be);
    if E == nil {
        Cudd_RecursiveDeref(dd, T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);

    /* No need to check for T == E, because it is guaranteed not to happen. */
    // inlining of macro Cudd_Regular
    res = cuddUniqueInter(dd, int32( ((DdNode *)((ptruint)(B) & ~01)).index), T, E);
    if res == nil {
        Cudd_RecursiveDeref(dd ,T);
        Cudd_RecursiveDeref(dd ,E);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(T).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(E).ref);

    if complement != 0 {
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        res1 = cuddAddCmplRecur(dd, res);
        if res1 == nil {
            Cudd_RecursiveDeref(dd, res);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        Cudd_RecursiveDeref(dd, res);
        res = res1;
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref);
    }

    /* Store result. */
    cuddCacheInsert1(dd,ddBddToAddRecur,B,res);

    return(res);

} /* end of ddBddToAddRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddTransfer.]

  Description [Performs the recursive step of Cudd_bddTransfer.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [cuddBddTransfer]

******************************************************************************/
func cuddBddTransferRecur(ddS *DdManager, ddD *DdManager, f *DdNode, table *st__table) *DdNode {
    var ft, fe, t, e, var, res *DdNode;
    var one, zero *DdNode;
    var index int32;
    var comple int32;

    statLine(ddD);
    // inlining of macro DD_ONE
    one = ((ddD).one);
    // inlining of macro Cudd_IsComplement
    comple = (int32( ((ptrint) (f) & 01)));

    /* Trivial cases. */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    // inlining of macro Cudd_NotCond
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX { return(((DdNode *)((ptrint)(one) ^ (comple)))); }

    /* Make canonical to increase the utilization of the cache. */
    // inlining of macro Cudd_NotCond
    f = ((DdNode *)((ptrint)(f) ^ (comple)));
    /* Now f is a regular pointer to a non-constant node. */

    /* Check the cache. */
    if st__lookup(table, (char *)f, (char **)&res) {
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(res) ^ (comple)))); }

    if ddS.TimeStop && Abc_Clock() > ddS.TimeStop {
        return nil; }
    if ddD.TimeStop && Abc_Clock() > ddD.TimeStop {
        return nil; }
    
    /* Recursive step. */
    index = f.index;
    // inlining of macro cuddT
    // inlining of macro cuddE
    ft = ((f).type.kids.T); fe = ((f).type.kids.E);

    t = cuddBddTransferRecur(ddS, ddD, ft, table);
    if t == nil {
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = cuddBddTransferRecur(ddS, ddD, fe, table);
    if e == nil {
        Cudd_RecursiveDeref(ddD, t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));
    var = cuddUniqueInter(ddD,index,one,zero);
    if var == nil {
        Cudd_RecursiveDeref(ddD, t);
        Cudd_RecursiveDeref(ddD, e);
        return(nil);
    }
    res = cuddBddIteRecur(ddD,var,t,e);
    if res == nil {
        Cudd_RecursiveDeref(ddD, t);
        Cudd_RecursiveDeref(ddD, e);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref(ddD, t);
    Cudd_RecursiveDeref(ddD, e);

    if st__add_direct(table, (char *) f, (char *) res) == st__OUT_OF_MEM {
        Cudd_RecursiveDeref(ddD, res);
        return(nil);
    }
    // inlining of macro Cudd_NotCond
    return(((DdNode *)((ptrint)(res) ^ (comple))));

} /* end of cuddBddTransferRecur */


//ABC_NAMESPACE_IMPL_END


