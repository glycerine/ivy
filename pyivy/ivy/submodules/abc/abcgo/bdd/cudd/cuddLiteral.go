package cudd

/**CFile***********************************************************************

  FileName    [cuddLiteral.c]

  PackageName [cudd]

  Synopsis    [Functions for manipulation of literal sets represented by
  BDDs.]

  Description [External procedures included in this file:
                <ul>
                <li> Cudd_bddLiteralSetIntersection()
                </ul>
            Internal procedures included in this file:
                <ul>
                <li> cuddBddLiteralSetIntersectionRecur()
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
static char rcsid[] DD_UNUSED = "$Id: cuddLiteral.c,v 1.8 2004/08/13 18:04:50 fabio Exp $";
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

  Synopsis    [Computes the intesection of two sets of literals
  represented as BDDs.]

  Description [Computes the intesection of two sets of literals
  represented as BDDs. Each set is represented as a cube of the
  literals in the set. The empty set is represented by the constant 1.
  No variable can be simultaneously present in both phases in a set.
  Returns a pointer to the BDD representing the intersected sets, if
  successful; NULL otherwise.]

  SideEffects [None]

******************************************************************************/
func Cudd_bddLiteralSetIntersection(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddLiteralSetIntersectionRecur(dd,f,g);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddLiteralSetIntersection */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of
  Cudd_bddLiteralSetIntersection.]

  Description [Performs the recursive step of
  Cudd_bddLiteralSetIntersection. Scans the cubes for common variables,
  and checks whether they agree in phase.  Returns a pointer to the
  resulting cube if successful; NULL otherwise.]

  SideEffects [None]

******************************************************************************/
func cuddBddLiteralSetIntersectionRecur(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var res, tmp *DdNode;
    var F, G *DdNode;
    var fc, gc *DdNode;
    var one *DdNode;
    var zero *DdNode;
    unsigned int topf, topg, comple;
    var phasef, phaseg int32;

    statLine(dd);
    if f == g { return(f); }

    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    // inlining of macro DD_ONE
    one = ((dd).one);

    /* Here f != g. If F == G, then f and g are complementary.
    ** Since they are two cubes, this case only occurs when f == v,
    ** g == v', and v is a variable or its complement.
    */
    if F == G { return(one); }

    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));
    // inlining of macro cuddI
    
    if (F.index)==CUDD_CONST_INDEX {
        topf = int32((F.index));
    } else {
        topf = (dd).perm[(F.index)];
    }
    // inlining of macro cuddI
    
    if (G.index)==CUDD_CONST_INDEX {
        topg = int32((G.index));
    } else {
        topg = (dd).perm[(G.index)];
    }
    /* Look for a variable common to both cubes. If there are none, this
    ** loop will stop when the constant node is reached in both cubes.
    */
    for topf != topg {
        if topf < topg {      /* move down on f */
            comple = f != F;
            // inlining of macro cuddT
            f = ((F).type.kids.T);
            // inlining of macro Cudd_Not
            if comple { f = ((DdNode *)((ptrint)(f) ^ 01)); }
            if f == zero {
                // inlining of macro cuddE
                f = ((F).type.kids.E);
                // inlining of macro Cudd_Not
                if comple { f = ((DdNode *)((ptrint)(f) ^ 01)); }
            }
            // inlining of macro Cudd_Regular
            F = ((DdNode *)((ptruint)(f) & ~01));
            // inlining of macro cuddI
            
            if (F.index)==CUDD_CONST_INDEX {
                topf = int32((F.index));
            } else {
                topf = (dd).perm[(F.index)];
            }
        } else if topg < topf {
            comple = g != G;
            // inlining of macro cuddT
            g = ((G).type.kids.T);
            // inlining of macro Cudd_Not
            if comple { g = ((DdNode *)((ptrint)(g) ^ 01)); }
            if g == zero {
                // inlining of macro cuddE
                g = ((G).type.kids.E);
                // inlining of macro Cudd_Not
                if comple { g = ((DdNode *)((ptrint)(g) ^ 01)); }
            }
            // inlining of macro Cudd_Regular
            G = ((DdNode *)((ptruint)(g) & ~01));
            // inlining of macro cuddI
            
            if (G.index)==CUDD_CONST_INDEX {
                topg = int32((G.index));
            } else {
                topg = (dd).perm[(G.index)];
            }
        }
    }

    /* At this point, f == one <=> g == 1. It suffices to test one of them. */
    if f == one { return(one); }

    res = cuddCacheLookup2(dd,Cudd_bddLiteralSetIntersection,f,g);
    if res != nil {
        return(res);
    }

    /* Here f and g are both non constant and have the same top variable. */
    comple = f != F;
    // inlining of macro cuddT
    fc = ((F).type.kids.T);
    phasef = 1;
    // inlining of macro Cudd_Not
    if comple { fc = ((DdNode *)((ptrint)(fc) ^ 01)); }
    if fc == zero {
        // inlining of macro cuddE
        fc = ((F).type.kids.E);
        phasef = 0;
        // inlining of macro Cudd_Not
        if comple { fc = ((DdNode *)((ptrint)(fc) ^ 01)); }
    }
    comple = g != G;
    // inlining of macro cuddT
    gc = ((G).type.kids.T);
    phaseg = 1;
    // inlining of macro Cudd_Not
    if comple { gc = ((DdNode *)((ptrint)(gc) ^ 01)); }
    if gc == zero {
        // inlining of macro cuddE
        gc = ((G).type.kids.E);
        phaseg = 0;
        // inlining of macro Cudd_Not
        if comple { gc = ((DdNode *)((ptrint)(gc) ^ 01)); }
    }

    tmp = cuddBddLiteralSetIntersectionRecur(dd,fc,gc);
    if tmp == nil {
        return(nil);
    }

    if phasef != phaseg {
        res = tmp;
    } else {
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(tmp).ref);
        if phasef == 0 {
            // inlining of macro Cudd_Not
            res = cuddBddAndRecur(dd,((DdNode *)((ptrint)(dd.vars[F.index]) ^ 01)),tmp);
        } else {
            res = cuddBddAndRecur(dd,dd.vars[F.index],tmp);
        }
        if res == nil {
            Cudd_RecursiveDeref(dd,tmp);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(tmp).ref); /* Just cuddDeref, because it is included in result */
    }

    cuddCacheInsert2(dd,Cudd_bddLiteralSetIntersection,f,g,res);

    return(res);

} /* end of cuddBddLiteralSetIntersectionRecur */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


//ABC_NAMESPACE_IMPL_END


