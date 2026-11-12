package cudd

/**CFile***********************************************************************

  FileName    [cuddClip.c]

  PackageName [cudd]

  Synopsis    [Clipping functions.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_bddClippingAnd()
                <li> Cudd_bddClippingAndAbstract()
                </ul>
       Internal procedures included in this module:
                <ul>
                <li> cuddBddClippingAnd()
                <li> cuddBddClippingAndAbstract()
                </ul>
       Static procedures included in this module:
                <ul>
                <li> cuddBddClippingAndRecur()
                <li> cuddBddClipAndAbsRecur()
                </ul>

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
static char rcsid[] DD_UNUSED = "$Id: cuddClip.c,v 1.8 2004/08/13 18:04:47 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static DdNode * cuddBddClippingAndRecur (DdManager *manager, DdNode *f, DdNode *g, int distance, int direction);

// static DdNode * cuddBddClipAndAbsRecur (DdManager *manager, DdNode *f, DdNode *g, DdNode *cube, int distance, int direction);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Approximates the conjunction of two BDDs f and g.]

  Description [Approximates the conjunction of two BDDs f and g. Returns a
  pointer to the resulting BDD if successful; NULL if the intermediate
  result blows up.]

  SideEffects [None]

  SeeAlso     [Cudd_bddAnd]

******************************************************************************/
DdNode *
Cudd_bddClippingAnd(
  DdManager * dd /* manager */,
  DdNode * f /* first conjunct */,
  DdNode * g /* second conjunct */,
  int  maxDepth /* maximum recursion depth */,
  int  direction /* under (0) or over (1) approximation */)
{
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddClippingAnd(dd,f,g,maxDepth,direction);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddClippingAnd */


/**Function********************************************************************

  Synopsis    [Approximates the conjunction of two BDDs f and g and
  simultaneously abstracts the variables in cube.]

  Description [Approximates the conjunction of two BDDs f and g and
  simultaneously abstracts the variables in cube. The variables are
  existentially abstracted. Returns a pointer to the resulting BDD if
  successful; NULL if the intermediate result blows up.]

  SideEffects [None]

  SeeAlso     [Cudd_bddAndAbstract Cudd_bddClippingAnd]

******************************************************************************/
DdNode *
Cudd_bddClippingAndAbstract(
  DdManager * dd /* manager */,
  DdNode * f /* first conjunct */,
  DdNode * g /* second conjunct */,
  DdNode * cube /* cube of variables to be abstracted */,
  int  maxDepth /* maximum recursion depth */,
  int  direction /* under (0) or over (1) approximation */)
{
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddClippingAndAbstract(dd,f,g,cube,maxDepth,direction);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddClippingAndAbstract */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Approximates the conjunction of two BDDs f and g.]

  Description [Approximates the conjunction of two BDDs f and g. Returns a
  pointer to the resulting BDD if successful; NULL if the intermediate
  result blows up.]

  SideEffects [None]

  SeeAlso     [Cudd_bddClippingAnd]

******************************************************************************/
DdNode *
cuddBddClippingAnd(
  DdManager * dd /* manager */,
  DdNode * f /* first conjunct */,
  DdNode * g /* second conjunct */,
  int  maxDepth /* maximum recursion depth */,
  int  direction /* under (0) or over (1) approximation */)
{
    var res *DdNode;

    res = cuddBddClippingAndRecur(dd,f,g,maxDepth,direction);

    return(res);

} /* end of cuddBddClippingAnd */


/**Function********************************************************************

  Synopsis    [Approximates the conjunction of two BDDs f and g and
  simultaneously abstracts the variables in cube.]

  Description [Approximates the conjunction of two BDDs f and g and
  simultaneously abstracts the variables in cube. Returns a
  pointer to the resulting BDD if successful; NULL if the intermediate
  result blows up.]

  SideEffects [None]

  SeeAlso     [Cudd_bddClippingAndAbstract]

******************************************************************************/
DdNode *
cuddBddClippingAndAbstract(
  DdManager * dd /* manager */,
  DdNode * f /* first conjunct */,
  DdNode * g /* second conjunct */,
  DdNode * cube /* cube of variables to be abstracted */,
  int  maxDepth /* maximum recursion depth */,
  int  direction /* under (0) or over (1) approximation */)
{
    var res *DdNode;

    res = cuddBddClipAndAbsRecur(dd,f,g,cube,maxDepth,direction);

    return(res);

} /* end of cuddBddClippingAndAbstract */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Implements the recursive step of Cudd_bddClippingAnd.]

  Description [Implements the recursive step of Cudd_bddClippingAnd by taking
  the conjunction of two BDDs.  Returns a pointer to the result is
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [cuddBddClippingAnd]

******************************************************************************/
func cuddBddClippingAndRecur(manager *DdManager, f *DdNode, g *DdNode, distance int32, direction int32) *DdNode {
    var F, ft, fe, G, gt, ge *DdNode;
    var one, zero, r, t, e *DdNode;
    unsigned int topf, topg, index;
    DD_CTFP cacheOp;

    statLine(manager);
    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Terminal cases. */
    // inlining of macro Cudd_Not
    if f == zero || g == zero || f == ((DdNode *)((ptrint)(g) ^ 01)) { return(zero); }
    if f == g || g == one { return(f); }
    if f == one { return(g); }
    if distance == 0 {
        /* One last attempt at returning the right result. We sort of
        ** cheat by calling Cudd_bddLeq. */
        if Cudd_bddLeq(manager,f,g) { return(f); }
        if Cudd_bddLeq(manager,g,f) { return(g); }
        if direction == 1 {
            // inlining of macro Cudd_Not
            if Cudd_bddLeq(manager,f,((DdNode *)((ptrint)(g) ^ 01))) ||
                // inlining of macro Cudd_Not
                Cudd_bddLeq(manager,g,((DdNode *)((ptrint)(f) ^ 01))) { return(zero); }
        }
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(one) ^ ((direction == 0)))));
    }

    /* At this point f and g are not constant. */
    distance--;

    /* Check cache. Try to increase cache efficiency by sorting the
    ** pointers. */
    if f > g {
        var tmp *DdNode  = f;
        f = g; g = tmp;
    }
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    cacheOp = (DD_CTFP)
        (direction ? Cudd_bddClippingAnd : cuddBddClippingAnd);
    if F.ref != 1 || G.ref != 1 {
        r = cuddCacheLookup2(manager, cacheOp, f, g);
        if r != nil { return(r); }
    }

    /* Here we can skip the use of cuddI, because the operands are known
    ** to be non-constant.
    */
    topf = manager.perm[F.index];
    topg = manager.perm[G.index];

    /* Compute cofactors. */
    if topf <= topg {
        index = F.index;
        // inlining of macro cuddT
        ft = ((F).type.kids.T);
        // inlining of macro cuddE
        fe = ((F).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (f) & 01)) != 0 {
            // inlining of macro Cudd_Not
            ft = ((DdNode *)((ptrint)(ft) ^ 01));
            // inlining of macro Cudd_Not
            fe = ((DdNode *)((ptrint)(fe) ^ 01));
        }
    } else {
        index = G.index;
        ft = fe = f;
    }

    if topg <= topf {
        // inlining of macro cuddT
        gt = ((G).type.kids.T);
        // inlining of macro cuddE
        ge = ((G).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (g) & 01)) != 0 {
            // inlining of macro Cudd_Not
            gt = ((DdNode *)((ptrint)(gt) ^ 01));
            // inlining of macro Cudd_Not
            ge = ((DdNode *)((ptrint)(ge) ^ 01));
        }
    } else {
        gt = ge = g;
    }

    t = cuddBddClippingAndRecur(manager, ft, gt, distance, direction);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);
    e = cuddBddClippingAndRecur(manager, fe, ge, distance, direction);
    if e == nil {
        Cudd_RecursiveDeref(manager, t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    if t == e {
        r = t;
    } else {
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (t) & 01)) != 0 {
            // inlining of macro Cudd_Not
            r = cuddUniqueInter(manager,int32(index),((DdNode *)((ptrint)(t) ^ 01)),((DdNode *)((ptrint)(e) ^ 01)));
            if r == nil {
                Cudd_RecursiveDeref(manager, t);
                Cudd_RecursiveDeref(manager, e);
                return(nil);
            }
            // inlining of macro Cudd_Not
            r = ((DdNode *)((ptrint)(r) ^ 01));
        } else {
            r = cuddUniqueInter(manager,int32(index),t,e);
            if r == nil {
                Cudd_RecursiveDeref(manager, t);
                Cudd_RecursiveDeref(manager, e);
                return(nil);
            }
        }
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    if F.ref != 1 || G.ref != 1 {
        cuddCacheInsert2(manager, cacheOp, f, g, r); }
    return(r);

} /* end of cuddBddClippingAndRecur */


/**Function********************************************************************

  Synopsis [Approximates the AND of two BDDs and simultaneously abstracts the
  variables in cube.]

  Description [Approximates the AND of two BDDs and simultaneously
  abstracts the variables in cube. The variables are existentially
  abstracted.  Returns a pointer to the result is successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddClippingAndAbstract]

******************************************************************************/
func cuddBddClipAndAbsRecur(manager *DdManager, f *DdNode, g *DdNode, cube *DdNode, distance int32, direction int32) *DdNode {
    var F, ft, fe, G, gt, ge *DdNode;
    var one, zero, r, t, e, Cube *DdNode;
    unsigned int topf, topg, topcube, top, index;
    var cacheTag ptruint;

    statLine(manager);
    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Terminal cases. */
    // inlining of macro Cudd_Not
    if f == zero || g == zero || f == ((DdNode *)((ptrint)(g) ^ 01)) { return(zero); }
    if f == one && g == one {   return(one); }
    if cube == one {
        return(cuddBddClippingAndRecur(manager, f, g, distance, direction));
    }
    if f == one || f == g {
        return (cuddBddExistAbstractRecur(manager, g, cube));
    }
    if g == one {
        return (cuddBddExistAbstractRecur(manager, f, cube));
    }
    // inlining of macro Cudd_NotCond
    if distance == 0 { return(((DdNode *)((ptrint)(one) ^ ((direction == 0))))); }

    /* At this point f, g, and cube are not constant. */
    distance--;

    /* Check cache. */
    if f > g { /* Try to increase cache efficiency. */
        var tmp *DdNode  = f;
        f = g; g = tmp;
    }
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    if direction != 0 {
        cacheTag = DD_BDD_CLIPPING_AND_ABSTRACT_UP_TAG;
    } else {
        cacheTag = DD_BDD_CLIPPING_AND_ABSTRACT_DOWN_TAG;
    }
    if F.ref != 1 || G.ref != 1 {
        r = cuddCacheLookup(manager, cacheTag,
                            f, g, cube);
        if r != nil {
            return(r);
        }
    }

    /* Here we can skip the use of cuddI, because the operands are known
    ** to be non-constant.
    */
    topf = manager.perm[F.index];
    topg = manager.perm[G.index];
    // inlining of macro ddMin
    
    if (topg) < (topf) {
        top = (topg);
    } else {
        top = (topf);
    }
    topcube = manager.perm[cube.index];

    if topcube < top {
        // inlining of macro cuddT
        return(cuddBddClipAndAbsRecur(manager, f, g, ((cube).type.kids.T),
                                      distance, direction));
    }
    /* Now, topcube >= top. */

    if topf == top {
        index = F.index;
        // inlining of macro cuddT
        ft = ((F).type.kids.T);
        // inlining of macro cuddE
        fe = ((F).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (f) & 01)) != 0 {
            // inlining of macro Cudd_Not
            ft = ((DdNode *)((ptrint)(ft) ^ 01));
            // inlining of macro Cudd_Not
            fe = ((DdNode *)((ptrint)(fe) ^ 01));
        }
    } else {
        index = G.index;
        ft = fe = f;
    }

    if topg == top {
        // inlining of macro cuddT
        gt = ((G).type.kids.T);
        // inlining of macro cuddE
        ge = ((G).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (g) & 01)) != 0 {
            // inlining of macro Cudd_Not
            gt = ((DdNode *)((ptrint)(gt) ^ 01));
            // inlining of macro Cudd_Not
            ge = ((DdNode *)((ptrint)(ge) ^ 01));
        }
    } else {
        gt = ge = g;
    }

    if topcube == top {
        // inlining of macro cuddT
        Cube = ((cube).type.kids.T);
    } else {
        Cube = cube;
    }

    t = cuddBddClipAndAbsRecur(manager, ft, gt, Cube, distance, direction);
    if t == nil { return(nil); }

    /* Special case: 1 OR anything = 1. Hence, no need to compute
    ** the else branch if t is 1.
    */
    if t == one && topcube == top {
        if F.ref != 1 || G.ref != 1 {
            cuddCacheInsert(manager, cacheTag, f, g, cube, one); }
        return(one);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = cuddBddClipAndAbsRecur(manager, fe, ge, Cube, distance, direction);
    if e == nil {
        Cudd_RecursiveDeref(manager, t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    if topcube == top {       /* abstract */
        // inlining of macro Cudd_Not
        r = cuddBddClippingAndRecur(manager, ((DdNode *)((ptrint)(t) ^ 01)), ((DdNode *)((ptrint)(e) ^ 01)),
                                    distance, (direction == 0));
        if r == nil {
            Cudd_RecursiveDeref(manager, t);
            Cudd_RecursiveDeref(manager, e);
            return(nil);
        }
        // inlining of macro Cudd_Not
        r = ((DdNode *)((ptrint)(r) ^ 01));
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(r).ref);
        Cudd_RecursiveDeref(manager, t);
        Cudd_RecursiveDeref(manager, e);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(r).ref);
    } else if t == e {
        r = t;
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    } else {
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (t) & 01)) != 0 {
            // inlining of macro Cudd_Not
            r = cuddUniqueInter(manager,int32(index),((DdNode *)((ptrint)(t) ^ 01)),((DdNode *)((ptrint)(e) ^ 01)));
            if r == nil {
                Cudd_RecursiveDeref(manager, t);
                Cudd_RecursiveDeref(manager, e);
                return(nil);
            }
            // inlining of macro Cudd_Not
            r = ((DdNode *)((ptrint)(r) ^ 01));
        } else {
            r = cuddUniqueInter(manager,int32(index),t,e);
            if r == nil {
                Cudd_RecursiveDeref(manager, t);
                Cudd_RecursiveDeref(manager, e);
                return(nil);
            }
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
    }
    if F.ref != 1 || G.ref != 1 {
        cuddCacheInsert(manager, cacheTag, f, g, cube, r); }
    return (r);

} /* end of cuddBddClipAndAbsRecur */


//ABC_NAMESPACE_IMPL_END


