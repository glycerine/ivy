package cudd

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

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddAndAbs.c,v 1.19 2004/08/13 18:04:46 fabio Exp $";
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

  Synopsis [Takes the AND of two BDDs and simultaneously abstracts the
  variables in cube.]

  Description [Takes the AND of two BDDs and simultaneously abstracts
  the variables in cube. The variables are existentially abstracted.
  Returns a pointer to the result is successful; NULL otherwise.
  Cudd_bddAndAbstract implements the semiring matrix multiplication
  algorithm for the boolean semiring.]

  SideEffects [None]

  SeeAlso     [Cudd_addMatrixMultiply Cudd_addTriangle Cudd_bddAnd]

******************************************************************************/
func Cudd_bddAndAbstract(manager *DdManager, f *DdNode, g *DdNode, cube *DdNode) *DdNode {
    var res *DdNode;

    for {
        manager.reordered = 0;
        res = cuddBddAndAbstractRecur(manager, f, g, cube);
    
if !(manager.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddAndAbstract */


/**Function********************************************************************

  Synopsis [Takes the AND of two BDDs and simultaneously abstracts the
  variables in cube.  Returns NULL if too many nodes are required.]

  Description [Takes the AND of two BDDs and simultaneously abstracts
  the variables in cube. The variables are existentially abstracted.
  Returns a pointer to the result is successful; NULL otherwise.
  In particular, if the number of new nodes created exceeds
  <code>limit</code>, this function returns NULL.]

  SideEffects [None]

  SeeAlso     [Cudd_bddAndAbstract]

******************************************************************************/
DdNode *
Cudd_bddAndAbstractLimit(
  DdManager * manager,
  DdNode * f,
  DdNode * g,
  DdNode * cube,
  unsigned int limit)
{
    var res *DdNode;
    unsigned int saveLimit = manager.maxLive;

    manager.maxLive = (manager.keys - manager.dead) +
      (manager.keysZ - manager.deadZ) + limit;
    for {
        manager.reordered = 0;
        res = cuddBddAndAbstractRecur(manager, f, g, cube);
    
if !(manager.reordered == 1) { break }
} 
    manager.maxLive = saveLimit;
    return(res);

} /* end of Cudd_bddAndAbstractLimit */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Takes the AND of two BDDs and simultaneously abstracts the
  variables in cube.]

  Description [Takes the AND of two BDDs and simultaneously abstracts
  the variables in cube. The variables are existentially abstracted.
  Returns a pointer to the result is successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddAndAbstract]

******************************************************************************/
func cuddBddAndAbstractRecur(manager *DdManager, f *DdNode, g *DdNode, cube *DdNode) *DdNode {
    var F, ft, fe, G, gt, ge *DdNode;
    var one, zero, r, t, e *DdNode;
    unsigned int topf, topg, topcube, top, index;

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
        return(cuddBddAndRecur(manager, f, g));
    }
    if f == one || f == g {
        return(cuddBddExistAbstractRecur(manager, g, cube));
    }
    if g == one {
        return(cuddBddExistAbstractRecur(manager, f, cube));
    }
    /* At this point f, g, and cube are not constant. */

    if f > g { /* Try to increase cache efficiency. */
        var tmp *DdNode  = f;
        f = g;
        g = tmp;
    }

    /* Here we can skip the use of cuddI, because the operands are known
    ** to be non-constant.
    */
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    topf = manager.perm[F.index];
    topg = manager.perm[G.index];
    // inlining of macro ddMin
    
    if (topg) < (topf) {
        top = (topg);
    } else {
        top = (topf);
    }
    topcube = manager.perm[cube.index];

    for topcube < top {
        // inlining of macro cuddT
        cube = ((cube).type.kids.T);
        if cube == one {
            return(cuddBddAndRecur(manager, f, g));
        }
        topcube = manager.perm[cube.index];
    }
    /* Now, topcube >= top. */

    /* Check cache. */
    if F.ref != 1 || G.ref != 1 {
        r = cuddCacheLookup(manager, DD_BDD_AND_ABSTRACT_TAG, f, g, cube);
        if r != nil {
            return(r);
        }
    }

    if manager.TimeStop && Abc_Clock() > manager.TimeStop {
        return nil; }

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

    if topcube == top {       /* quantify */
        // inlining of macro cuddT
        var Cube *DdNode  = ((cube).type.kids.T);
        t = cuddBddAndAbstractRecur(manager, ft, gt, Cube);
        if t == nil { return(nil); }
        /* Special case: 1 OR anything = 1. Hence, no need to compute
        ** the else branch if t is 1. Likewise t + t * anything == t.
        ** Notice that t == fe implies that fe does not depend on the
        ** variables in Cube. Likewise for t == ge.
        */
        if t == one || t == fe || t == ge {
            if F.ref != 1 || G.ref != 1 {
                cuddCacheInsert(manager, DD_BDD_AND_ABSTRACT_TAG,
                                f, g, cube, t); }
            return(t);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(t).ref);
        /* Special case: t + !t * anything == t + anything. */
        // inlining of macro Cudd_Not
        if t == ((DdNode *)((ptrint)(fe) ^ 01)) {
            e = cuddBddExistAbstractRecur(manager, ge, Cube);
        // inlining of macro Cudd_Not
        } else if t == ((DdNode *)((ptrint)(ge) ^ 01)) {
            e = cuddBddExistAbstractRecur(manager, fe, Cube);
        } else {
            e = cuddBddAndAbstractRecur(manager, fe, ge, Cube);
        }
        if e == nil {
            Cudd_IterDerefBdd(manager, t);
            return(nil);
        }
        if t == e {
            r = t;
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(t).ref);
        } else {
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(e).ref);
            // inlining of macro Cudd_Not
            r = cuddBddAndRecur(manager, ((DdNode *)((ptrint)(t) ^ 01)), ((DdNode *)((ptrint)(e) ^ 01)));
            if r == nil {
                Cudd_IterDerefBdd(manager, t);
                Cudd_IterDerefBdd(manager, e);
                return(nil);
            }
            // inlining of macro Cudd_Not
            r = ((DdNode *)((ptrint)(r) ^ 01));
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(r).ref);
            Cudd_DelayedDerefBdd(manager, t);
            Cudd_DelayedDerefBdd(manager, e);
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(r).ref);
        }
    } else {
        t = cuddBddAndAbstractRecur(manager, ft, gt, cube);
        if t == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(t).ref);
        e = cuddBddAndAbstractRecur(manager, fe, ge, cube);
        if e == nil {
            Cudd_IterDerefBdd(manager, t);
            return(nil);
        }
        if t == e {
            r = t;
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(t).ref);
        } else {
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(e).ref);
            // inlining of macro Cudd_IsComplement
            if int32( ((ptrint) (t) & 01)) != 0 {
                r = cuddUniqueInter(manager, int32( index),
                                    // inlining of macro Cudd_Not
                                    ((DdNode *)((ptrint)(t) ^ 01)), ((DdNode *)((ptrint)(e) ^ 01)));
                if r == nil {
                    Cudd_IterDerefBdd(manager, t);
                    Cudd_IterDerefBdd(manager, e);
                    return(nil);
                }
                // inlining of macro Cudd_Not
                r = ((DdNode *)((ptrint)(r) ^ 01));
            } else {
                r = cuddUniqueInter(manager,int32(index),t,e);
                if r == nil {
                    Cudd_IterDerefBdd(manager, t);
                    Cudd_IterDerefBdd(manager, e);
                    return(nil);
                }
            }
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(e).ref);
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(t).ref);
        }
    }

    if F.ref != 1 || G.ref != 1 {
        cuddCacheInsert(manager, DD_BDD_AND_ABSTRACT_TAG, f, g, cube, r); }
    return (r);

} /* end of cuddBddAndAbstractRecur */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


//ABC_NAMESPACE_IMPL_END


