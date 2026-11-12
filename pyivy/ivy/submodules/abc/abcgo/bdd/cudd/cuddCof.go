package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddCof.c]

  PackageName [cudd]

  Synopsis    [Cofactoring functions.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_Cofactor()
                </ul>
               Internal procedures included in this module:
                <ul>
                <li> cuddGetBranches()
                <li> cuddCheckCube()
                <li> cuddCofactorRecur()
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
static char rcsid[] DD_UNUSED = "$Id: cuddCof.c,v 1.9 2004/08/13 18:04:47 fabio Exp $";
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

  Synopsis    [Computes the cofactor of f with respect to g.]

  Description [Computes the cofactor of f with respect to g; g must be
  the BDD or the ADD of a cube. Returns a pointer to the cofactor if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddConstrain Cudd_bddRestrict]

******************************************************************************/
func Cudd_Cofactor(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var res, zero *DdNode;

    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    zero = ((DdNode *)((ptrint)(((dd).one)) ^ 01));
    // inlining of macro DD_ZERO
    if g == zero || g == ((dd).zero) {
        (void) fmt.Fprintf(dd.err, "Cudd_Cofactor: Invalid restriction 1\n");
        dd.errorCode = CUDD_INVALID_ARG;
        return(nil);
    }
    for {
        dd.reordered = 0;
        res = cuddCofactorRecur(dd,f,g);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_Cofactor */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Computes the children of g.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddGetBranches(g *DdNode, g1 **DdNode, g0 **DdNode) {
    // inlining of macro Cudd_Regular
    var G *DdNode  = ((DdNode *)((ptruint)(g) & ~01));

    // inlining of macro cuddT
    *g1 = ((G).type.kids.T);
    // inlining of macro cuddE
    *g0 = ((G).type.kids.E);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (g) & 01)) != 0 {
        // inlining of macro Cudd_Not
        *g1 = ((DdNode *)((ptrint)(*g1) ^ 01));
        // inlining of macro Cudd_Not
        *g0 = ((DdNode *)((ptrint)(*g0) ^ 01));
    }

} /* end of cuddGetBranches */


/**Function********************************************************************

  Synopsis    [Checks whether g is the BDD of a cube.]

  Description [Checks whether g is the BDD of a cube. Returns 1 in case
  of success; 0 otherwise. The constant 1 is a valid cube, but all other
  constant functions cause cuddCheckCube to return 0.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddCheckCube(dd *DdManager, g *DdNode) int32 {
    var g1, g0, one, zero *DdNode;
    
    // inlining of macro DD_ONE
    one = ((dd).one);
    if g == one { return(1); }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(g) & ~01))).index == CUDD_CONST_INDEX { return(0); }

    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));
    cuddGetBranches(g,&g1,&g0);

    if g0 == zero {
        return(cuddCheckCube(dd, g1));
    }
    if g1 == zero {
        return(cuddCheckCube(dd, g0));
    }
    return(0);

} /* end of cuddCheckCube */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_Cofactor.]

  Description [Performs the recursive step of Cudd_Cofactor. Returns a
  pointer to the cofactor if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_Cofactor]

******************************************************************************/
func cuddCofactorRecur(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var one, zero, F, G, g1, g0, f1, f0, t, e, r *DdNode;
    unsigned int topf,topg;
    var comple int32;

    statLine(dd);
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro cuddIsConstant
    if (F).index == CUDD_CONST_INDEX { return(f); }

    // inlining of macro DD_ONE
    one = ((dd).one);

    /* The invariant g != 0 is true on entry to this procedure and is
    ** recursively maintained by it. Therefore it suffices to test g
    ** against one to make sure it is not constant.
    */
    if g == one { return(f); }
    /* From now on, f and g are known not to be constants. */

    comple = f != F;
    r = cuddCacheLookup2(dd,Cudd_Cofactor,F,g);
    if r != nil {
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    }

    topf = dd.perm[F.index];
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    topg = dd.perm[G.index];

    /* We take the cofactors of F because we are going to rely on
    ** the fact that the cofactors of the complement are the complements
    ** of the cofactors to better utilize the cache. Variable comple
    ** remembers whether we have to complement the result or not.
    */
    if topf <= topg {
        // inlining of macro cuddT
        // inlining of macro cuddE
        f1 = ((F).type.kids.T); f0 = ((F).type.kids.E);
    } else {
        f1 = f0 = F;
    }
    if topg <= topf {
        // inlining of macro cuddT
        // inlining of macro cuddE
        g1 = ((G).type.kids.T); g0 = ((G).type.kids.E);
        // inlining of macro Cudd_Not
        if g != G { g1 = ((DdNode *)((ptrint)(g1) ^ 01)); g0 = ((DdNode *)((ptrint)(g0) ^ 01)); }
    } else {
        g1 = g0 = g;
    }

    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));
    if topf >= topg {
        // inlining of macro DD_ZERO
        if g0 == zero || g0 == ((dd).zero) {
            r = cuddCofactorRecur(dd, f1, g1);
        // inlining of macro DD_ZERO
        } else if g1 == zero || g1 == ((dd).zero) {
            r = cuddCofactorRecur(dd, f0, g0);
        } else {
            (void) fmt.Fprintf(dd.out, "Cudd_Cofactor: Invalid restriction 2\n");
            dd.errorCode = CUDD_INVALID_ARG;
            return(nil);
        }
        if r == nil { return(nil); }
    } else /* if (topf < topg) */ {
        t = cuddCofactorRecur(dd, f1, g);
        if t == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(t).ref);
        e = cuddCofactorRecur(dd, f0, g);
        if e == nil {
            Cudd_RecursiveDeref(dd, t);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);

        if t == e {
            r = t;
        // inlining of macro Cudd_IsComplement
        } else if int32( ((ptrint) (t) & 01)) != 0 {
            // inlining of macro Cudd_Not
            r = cuddUniqueInter(dd,int32(F.index),((DdNode *)((ptrint)(t) ^ 01)),((DdNode *)((ptrint)(e) ^ 01)));
            if r != nil {
                // inlining of macro Cudd_Not
                r = ((DdNode *)((ptrint)(r) ^ 01)); }
        } else {
            r = cuddUniqueInter(dd,int32(F.index),t,e);
        }
        if r == nil {
            Cudd_RecursiveDeref(dd ,e);
            Cudd_RecursiveDeref(dd ,t);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    }

    cuddCacheInsert2(dd,Cudd_Cofactor,F,g,r);

    // inlining of macro Cudd_NotCond
    return(((DdNode *)((ptrint)(r) ^ (comple))));

} /* end of cuddCofactorRecur */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


//ABC_NAMESPACE_IMPL_END


