package cudd

/**CFile***********************************************************************

  FileName    [cuddZddSetop.c]

  PackageName [cudd]

  Synopsis    [Set operations on ZDDs.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_zddIte()
                    <li> Cudd_zddUnion()
                    <li> Cudd_zddIntersect()
                    <li> Cudd_zddDiff()
                    <li> Cudd_zddDiffConst()
                    <li> Cudd_zddSubset1()
                    <li> Cudd_zddSubset0()
                    <li> Cudd_zddChange()
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    <li> cuddZddIte()
                    <li> cuddZddUnion()
                    <li> cuddZddIntersect()
                    <li> cuddZddDiff()
                    <li> cuddZddChangeAux()
                    <li> cuddZddSubset1()
                    <li> cuddZddSubset0()
                    </ul>
               Static procedures included in this module:
                    <ul>
                    <li> zdd_subset1_aux()
                    <li> zdd_subset0_aux()
                    <li> zddVarToConst()
                    </ul>
              ]

  SeeAlso     []

  Author      [Hyong-Kyoon Shin, In-Ho Moon]

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
static char rcsid[] DD_UNUSED = "$Id: cuddZddSetop.c,v 1.25 2004/08/13 18:04:54 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static DdNode * zdd_subset1_aux (DdManager *zdd, DdNode *P, DdNode *zvar);

// static DdNode * zdd_subset0_aux (DdManager *zdd, DdNode *P, DdNode *zvar);

// static void zddVarToConst (DdNode *f, DdNode **gp, DdNode **hp, DdNode *base, DdNode *empty);


/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Computes the ITE of three ZDDs.]

  Description [Computes the ITE of three ZDDs. Returns a pointer to the
  result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_zddIte(dd *DdManager, f *DdNode, g *DdNode, h *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddZddIte(dd, f, g, h);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_zddIte */


/**Function********************************************************************

  Synopsis [Computes the union of two ZDDs.]

  Description [Computes the union of two ZDDs. Returns a pointer to the
  result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_zddUnion(dd *DdManager, P *DdNode, Q *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddZddUnion(dd, P, Q);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_zddUnion */


/**Function********************************************************************

  Synopsis [Computes the intersection of two ZDDs.]

  Description [Computes the intersection of two ZDDs. Returns a pointer to
  the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_zddIntersect(dd *DdManager, P *DdNode, Q *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddZddIntersect(dd, P, Q);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_zddIntersect */


/**Function********************************************************************

  Synopsis [Computes the difference of two ZDDs.]

  Description [Computes the difference of two ZDDs. Returns a pointer to the
  result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_zddDiffConst]

******************************************************************************/
func Cudd_zddDiff(dd *DdManager, P *DdNode, Q *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddZddDiff(dd, P, Q);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_zddDiff */


/**Function********************************************************************

  Synopsis [Performs the inclusion test for ZDDs (P implies Q).]

  Description [Inclusion test for ZDDs (P implies Q). No new nodes are
  generated by this procedure. Returns empty if true;
  a valid pointer different from empty or DD_NON_CONSTANT otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_zddDiff]

******************************************************************************/
func Cudd_zddDiffConst(zdd *DdManager, P *DdNode, Q *DdNode) *DdNode {
    var p_top, q_top int32;
    // inlining of macro DD_ZERO
    var empty *DdNode  = ((zdd).zero), *t, *res;
    var table *DdManager  = zdd;

    statLine(zdd);
    if P == empty {
        return(empty); }
    if Q == empty {
        return(P); }
    if P == Q {
        return(empty); }

    /* Check cache.  The cache is shared by cuddZddDiff(). */
    res = cuddCacheLookup2Zdd(table, cuddZddDiff, P, Q);
    if res != nil {
        return(res); }

    // inlining of macro cuddIsConstant
    if (P).index == CUDD_CONST_INDEX {
        p_top = P.index;
    } else {
        p_top = zdd.permZ[P.index]; }
    // inlining of macro cuddIsConstant
    if (Q).index == CUDD_CONST_INDEX {
        q_top = Q.index;
    } else {
        q_top = zdd.permZ[Q.index]; }
    if p_top < q_top {
        res = DD_NON_CONSTANT;
    } else if p_top > q_top {
        // inlining of macro cuddE
        res = Cudd_zddDiffConst(zdd, P, ((Q).type.kids.E));
    } else {
        // inlining of macro cuddT
        t = Cudd_zddDiffConst(zdd, ((P).type.kids.T), ((Q).type.kids.T));
        if t != empty {
            res = DD_NON_CONSTANT;
        } else {
            // inlining of macro cuddE
            res = Cudd_zddDiffConst(zdd, ((P).type.kids.E), ((Q).type.kids.E)); }
    }

    cuddCacheInsert2(table, cuddZddDiff, P, Q, res);

    return(res);

} /* end of Cudd_zddDiffConst */


/**Function********************************************************************

  Synopsis [Computes the positive cofactor of a ZDD w.r.t. a variable.]

  Description [Computes the positive cofactor of a ZDD w.r.t. a
  variable. In terms of combinations, the result is the set of all
  combinations in which the variable is asserted. Returns a pointer to
  the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_zddSubset0]

******************************************************************************/
func Cudd_zddSubset1(dd *DdManager, P *DdNode, var int32) *DdNode {
    var r *DdNode;

    for {
        dd.reordered = 0;
        r = cuddZddSubset1(dd, P, var);
    
if !(dd.reordered == 1) { break }
} 

    return(r);

} /* end of Cudd_zddSubset1 */


/**Function********************************************************************

  Synopsis [Computes the negative cofactor of a ZDD w.r.t. a variable.]

  Description [Computes the negative cofactor of a ZDD w.r.t. a
  variable. In terms of combinations, the result is the set of all
  combinations in which the variable is negated. Returns a pointer to
  the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_zddSubset1]

******************************************************************************/
func Cudd_zddSubset0(dd *DdManager, P *DdNode, var int32) *DdNode {
    var r *DdNode;

    for {
        dd.reordered = 0;
        r = cuddZddSubset0(dd, P, var);
    
if !(dd.reordered == 1) { break }
} 

    return(r);

} /* end of Cudd_zddSubset0 */


/**Function********************************************************************

  Synopsis [Substitutes a variable with its complement in a ZDD.]

  Description [Substitutes a variable with its complement in a ZDD.
  returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_zddChange(dd *DdManager, P *DdNode, var int32) *DdNode {
    var res *DdNode;

    if (unsigned int) var >= CUDD_MAXINDEX - 1 { return(nil); }
    
    for {
        dd.reordered = 0;
        res = cuddZddChange(dd, P, var);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_zddChange */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_zddIte.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddIte(dd *DdManager, f *DdNode, g *DdNode, h *DdNode) *DdNode {
    var tautology, empty *DdNode;
    var r, Gv, Gvn, Hv, Hvn, t, e *DdNode;
    unsigned int topf,topg,toph,v,top;
    var index int32;

    statLine(dd);
    /* Trivial cases. */
    /* One variable cases. */
    // inlining of macro DD_ZERO
    if f == (empty = ((dd).zero)) {   /* ITE(0,G,H) = H */
        return(h);
    }
    // inlining of macro cuddIZ
    
    if (f.index)==CUDD_CONST_INDEX {
        topf = int32((f.index));
    } else {
        topf = (dd).permZ[(f.index)];
    }
    // inlining of macro cuddIZ
    
    if (g.index)==CUDD_CONST_INDEX {
        topg = int32((g.index));
    } else {
        topg = (dd).permZ[(g.index)];
    }
    // inlining of macro cuddIZ
    
    if (h.index)==CUDD_CONST_INDEX {
        toph = int32((h.index));
    } else {
        toph = (dd).permZ[(h.index)];
    }
    // inlining of macro ddMin
    
    if (toph) < (topg) {
        v = (toph);
    } else {
        v = (topg);
    }
    // inlining of macro ddMin
    
    if (v) < (topf) {
        top = (v);
    } else {
        top = (topf);
    }

    if top == CUDD_MAXINDEX {
        // inlining of macro DD_ONE
        tautology = ((dd).one);
    } else {
        tautology = dd.univ[top];
    }
    if f == tautology {                       /* ITE(1,G,H) = G */
        return(g);
    }

    /* From now on, f is known to not be a constant. */
    zddVarToConst(f,&g,&h,tautology,empty);

    /* Check remaining one variable cases. */
    if g == h {                       /* ITE(F,G,G) = G */
        return(g);
    }

    if g == tautology {                       /* ITE(F,1,0) = F */
        if h == empty { return(f); }
    }

    /* Check cache. */
    r = cuddCacheLookupZdd(dd,DD_ZDD_ITE_TAG,f,g,h);
    if r != nil {
        return(r);
    }

    /* Recompute these because they may have changed in zddVarToConst. */
    // inlining of macro cuddIZ
    
    if (g.index)==CUDD_CONST_INDEX {
        topg = int32((g.index));
    } else {
        topg = (dd).permZ[(g.index)];
    }
    // inlining of macro cuddIZ
    
    if (h.index)==CUDD_CONST_INDEX {
        toph = int32((h.index));
    } else {
        toph = (dd).permZ[(h.index)];
    }
    // inlining of macro ddMin
    
    if (toph) < (topg) {
        v = (toph);
    } else {
        v = (topg);
    }

    if topf < v {
        // inlining of macro cuddE
        r = cuddZddIte(dd,((f).type.kids.E),g,h);
        if r == nil { return(nil); }
    } else if topf > v {
        if topg > v {
            Gvn = g;
            index = h.index;
        } else {
            // inlining of macro cuddE
            Gvn = ((g).type.kids.E);
            index = g.index;
        }
        if toph > v {
            Hv = empty; Hvn = h;
        } else {
            // inlining of macro cuddT
            // inlining of macro cuddE
            Hv = ((h).type.kids.T); Hvn = ((h).type.kids.E);
        }
        e = cuddZddIte(dd,f,Gvn,Hvn);
        if e == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        r = cuddZddGetNode(dd,index,Hv,e);
        if r == nil {
            Cudd_RecursiveDerefZdd(dd,e);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    } else {
        index = f.index;
        if topg > v {
            Gv = empty; Gvn = g;
        } else {
            // inlining of macro cuddT
            // inlining of macro cuddE
            Gv = ((g).type.kids.T); Gvn = ((g).type.kids.E);
        }
        if toph > v {
            Hv = empty; Hvn = h;
        } else {
            // inlining of macro cuddT
            // inlining of macro cuddE
            Hv = ((h).type.kids.T); Hvn = ((h).type.kids.E);
        }
        // inlining of macro cuddE
        e = cuddZddIte(dd,((f).type.kids.E),Gvn,Hvn);
        if e == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        // inlining of macro cuddT
        t = cuddZddIte(dd,((f).type.kids.T),Gv,Hv);
        if t == nil {
            Cudd_RecursiveDerefZdd(dd,e);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(t).ref);
        r = cuddZddGetNode(dd,index,t,e);
        if r == nil {
            Cudd_RecursiveDerefZdd(dd,e);
            Cudd_RecursiveDerefZdd(dd,t);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    }

    cuddCacheInsert(dd,DD_ZDD_ITE_TAG,f,g,h,r);

    return(r);

} /* end of cuddZddIte */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_zddUnion.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddUnion(zdd *DdManager, P *DdNode, Q *DdNode) *DdNode {
    var p_top, q_top int32;
    // inlining of macro DD_ZERO
    var empty *DdNode  = ((zdd).zero), *t, *e, *res;
    var table *DdManager  = zdd;

    statLine(zdd);
    if P == empty {
        return(Q); }
    if Q == empty {
        return(P); }
    if P == Q {
        return(P); }

    /* Check cache */
    res = cuddCacheLookup2Zdd(table, cuddZddUnion, P, Q);
    if res != nil {
        return(res); }

    // inlining of macro cuddIsConstant
    if (P).index == CUDD_CONST_INDEX {
        p_top = P.index;
    } else {
        p_top = zdd.permZ[P.index]; }
    // inlining of macro cuddIsConstant
    if (Q).index == CUDD_CONST_INDEX {
        q_top = Q.index;
    } else {
        q_top = zdd.permZ[Q.index]; }
    if p_top < q_top {
        // inlining of macro cuddE
        e = cuddZddUnion(zdd, ((P).type.kids.E), Q);
        if e == nil { return (nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        // inlining of macro cuddT
        res = cuddZddGetNode(zdd, P.index, ((P).type.kids.T), e);
        if res == nil {
            Cudd_RecursiveDerefZdd(table, e);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    } else if p_top > q_top {
        // inlining of macro cuddE
        e = cuddZddUnion(zdd, P, ((Q).type.kids.E));
        if e == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        // inlining of macro cuddT
        res = cuddZddGetNode(zdd, Q.index, ((Q).type.kids.T), e);
        if res == nil {
            Cudd_RecursiveDerefZdd(table, e);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    } else {
        // inlining of macro cuddT
        t = cuddZddUnion(zdd, ((P).type.kids.T), ((Q).type.kids.T));
        if t == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(t).ref);
        // inlining of macro cuddE
        e = cuddZddUnion(zdd, ((P).type.kids.E), ((Q).type.kids.E));
        if e == nil {
            Cudd_RecursiveDerefZdd(table, t);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        res = cuddZddGetNode(zdd, P.index, t, e);
        if res == nil {
            Cudd_RecursiveDerefZdd(table, t);
            Cudd_RecursiveDerefZdd(table, e);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    }

    cuddCacheInsert2(table, cuddZddUnion, P, Q, res);

    return(res);

} /* end of cuddZddUnion */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_zddIntersect.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddIntersect(zdd *DdManager, P *DdNode, Q *DdNode) *DdNode {
    var p_top, q_top int32;
    // inlining of macro DD_ZERO
    var empty *DdNode  = ((zdd).zero), *t, *e, *res;
    var table *DdManager  = zdd;

    statLine(zdd);
    if P == empty {
        return(empty); }
    if Q == empty {
        return(empty); }
    if P == Q {
        return(P); }

    /* Check cache. */
    res = cuddCacheLookup2Zdd(table, cuddZddIntersect, P, Q);
    if res != nil {
        return(res); }

    // inlining of macro cuddIsConstant
    if (P).index == CUDD_CONST_INDEX {
        p_top = P.index;
    } else {
        p_top = zdd.permZ[P.index]; }
    // inlining of macro cuddIsConstant
    if (Q).index == CUDD_CONST_INDEX {
        q_top = Q.index;
    } else {
        q_top = zdd.permZ[Q.index]; }
    if p_top < q_top {
        // inlining of macro cuddE
        res = cuddZddIntersect(zdd, ((P).type.kids.E), Q);
        if res == nil { return(nil); }
    } else if p_top > q_top {
        // inlining of macro cuddE
        res = cuddZddIntersect(zdd, P, ((Q).type.kids.E));
        if res == nil { return(nil); }
    } else {
        // inlining of macro cuddT
        t = cuddZddIntersect(zdd, ((P).type.kids.T), ((Q).type.kids.T));
        if t == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(t).ref);
        // inlining of macro cuddE
        e = cuddZddIntersect(zdd, ((P).type.kids.E), ((Q).type.kids.E));
        if e == nil {
            Cudd_RecursiveDerefZdd(table, t);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        res = cuddZddGetNode(zdd, P.index, t, e);
        if res == nil {
            Cudd_RecursiveDerefZdd(table, t);
            Cudd_RecursiveDerefZdd(table, e);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    }

    cuddCacheInsert2(table, cuddZddIntersect, P, Q, res);

    return(res);

} /* end of cuddZddIntersect */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_zddDiff.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddDiff(zdd *DdManager, P *DdNode, Q *DdNode) *DdNode {
    var p_top, q_top int32;
    // inlining of macro DD_ZERO
    var empty *DdNode  = ((zdd).zero), *t, *e, *res;
    var table *DdManager  = zdd;

    statLine(zdd);
    if P == empty {
        return(empty); }
    if Q == empty {
        return(P); }
    if P == Q {
        return(empty); }

    /* Check cache.  The cache is shared by Cudd_zddDiffConst(). */
    res = cuddCacheLookup2Zdd(table, cuddZddDiff, P, Q);
    if res != nil && res != DD_NON_CONSTANT {
        return(res); }

    // inlining of macro cuddIsConstant
    if (P).index == CUDD_CONST_INDEX {
        p_top = P.index;
    } else {
        p_top = zdd.permZ[P.index]; }
    // inlining of macro cuddIsConstant
    if (Q).index == CUDD_CONST_INDEX {
        q_top = Q.index;
    } else {
        q_top = zdd.permZ[Q.index]; }
    if p_top < q_top {
        // inlining of macro cuddE
        e = cuddZddDiff(zdd, ((P).type.kids.E), Q);
        if e == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        // inlining of macro cuddT
        res = cuddZddGetNode(zdd, P.index, ((P).type.kids.T), e);
        if res == nil {
            Cudd_RecursiveDerefZdd(table, e);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    } else if p_top > q_top {
        // inlining of macro cuddE
        res = cuddZddDiff(zdd, P, ((Q).type.kids.E));
        if res == nil { return(nil); }
    } else {
        // inlining of macro cuddT
        t = cuddZddDiff(zdd, ((P).type.kids.T), ((Q).type.kids.T));
        if t == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(t).ref);
        // inlining of macro cuddE
        e = cuddZddDiff(zdd, ((P).type.kids.E), ((Q).type.kids.E));
        if e == nil {
            Cudd_RecursiveDerefZdd(table, t);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        res = cuddZddGetNode(zdd, P.index, t, e);
        if res == nil {
            Cudd_RecursiveDerefZdd(table, t);
            Cudd_RecursiveDerefZdd(table, e);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    }

    cuddCacheInsert2(table, cuddZddDiff, P, Q, res);

    return(res);

} /* end of cuddZddDiff */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_zddChange.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddChangeAux(zdd *DdManager, P *DdNode, zvar *DdNode) *DdNode {
    var top_var, level int32;
    var res, t, e *DdNode;
    // inlining of macro DD_ONE
    var base *DdNode  = ((zdd).one);
    // inlining of macro DD_ZERO
    var empty *DdNode  = ((zdd).zero);

    statLine(zdd);
    if P == empty {
        return(empty); }
    if P == base {
        return(zvar); }

    /* Check cache. */
    res = cuddCacheLookup2Zdd(zdd, cuddZddChangeAux, P, zvar);
    if res != nil {
        return(res); }

    top_var = zdd.permZ[P.index];
    level = zdd.permZ[zvar.index];

    if top_var > level {
        // inlining of macro DD_ZERO
        res = cuddZddGetNode(zdd, zvar.index, P, ((zdd).zero));
        if res == nil { return(nil); }
    } else if top_var == level {
        // inlining of macro cuddE
        // inlining of macro cuddT
        res = cuddZddGetNode(zdd, zvar.index, ((P).type.kids.E), ((P).type.kids.T));
        if res == nil { return(nil); }
    } else {
        // inlining of macro cuddT
        t = cuddZddChangeAux(zdd, ((P).type.kids.T), zvar);
        if t == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(t).ref);
        // inlining of macro cuddE
        e = cuddZddChangeAux(zdd, ((P).type.kids.E), zvar);
        if e == nil {
            Cudd_RecursiveDerefZdd(zdd, t);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        res = cuddZddGetNode(zdd, P.index, t, e);
        if res == nil {
            Cudd_RecursiveDerefZdd(zdd, t);
            Cudd_RecursiveDerefZdd(zdd, e);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    }

    cuddCacheInsert2(zdd, cuddZddChangeAux, P, zvar, res);

    return(res);

} /* end of cuddZddChangeAux */


/**Function********************************************************************

  Synopsis [Computes the positive cofactor of a ZDD w.r.t. a variable.]

  Description [Computes the positive cofactor of a ZDD w.r.t. a
  variable. In terms of combinations, the result is the set of all
  combinations in which the variable is asserted. Returns a pointer to
  the result if successful; NULL otherwise. cuddZddSubset1 performs
  the same function as Cudd_zddSubset1, but does not restart if
  reordering has taken place. Therefore it can be called from within a
  recursive procedure.]

  SideEffects [None]

  SeeAlso     [cuddZddSubset0 Cudd_zddSubset1]

******************************************************************************/
func cuddZddSubset1(dd *DdManager, P *DdNode, var int32) *DdNode {
    var zvar, r *DdNode;
    var base, empty *DdNode;

    // inlining of macro DD_ONE
    base = ((dd).one);
    // inlining of macro DD_ZERO
    empty = ((dd).zero);

    zvar = cuddUniqueInterZdd(dd, var, base, empty);
    if zvar == nil {
        return(nil);
    } else {
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zvar).ref);
        r = zdd_subset1_aux(dd, P, zvar);
        if r == nil {
            Cudd_RecursiveDerefZdd(dd, zvar);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(r).ref);
        Cudd_RecursiveDerefZdd(dd, zvar);
    }

    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(r).ref);
    return(r);

} /* end of cuddZddSubset1 */


/**Function********************************************************************

  Synopsis [Computes the negative cofactor of a ZDD w.r.t. a variable.]

  Description [Computes the negative cofactor of a ZDD w.r.t. a
  variable. In terms of combinations, the result is the set of all
  combinations in which the variable is negated. Returns a pointer to
  the result if successful; NULL otherwise. cuddZddSubset0 performs
  the same function as Cudd_zddSubset0, but does not restart if
  reordering has taken place. Therefore it can be called from within a
  recursive procedure.]

  SideEffects [None]

  SeeAlso     [cuddZddSubset1 Cudd_zddSubset0]

******************************************************************************/
func cuddZddSubset0(dd *DdManager, P *DdNode, var int32) *DdNode {
    var zvar, r *DdNode;
    var base, empty *DdNode;

    // inlining of macro DD_ONE
    base = ((dd).one);
    // inlining of macro DD_ZERO
    empty = ((dd).zero);

    zvar = cuddUniqueInterZdd(dd, var, base, empty);
    if zvar == nil {
        return(nil);
    } else {
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zvar).ref);
        r = zdd_subset0_aux(dd, P, zvar);
        if r == nil {
            Cudd_RecursiveDerefZdd(dd, zvar);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(r).ref);
        Cudd_RecursiveDerefZdd(dd, zvar);
    }

    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(r).ref);
    return(r);

} /* end of cuddZddSubset0 */


/**Function********************************************************************

  Synopsis [Substitutes a variable with its complement in a ZDD.]

  Description [Substitutes a variable with its complement in a ZDD.
  returns a pointer to the result if successful; NULL
  otherwise. cuddZddChange performs the same function as
  Cudd_zddChange, but does not restart if reordering has taken
  place. Therefore it can be called from within a recursive
  procedure.]

  SideEffects [None]

  SeeAlso     [Cudd_zddChange]

******************************************************************************/
func cuddZddChange(dd *DdManager, P *DdNode, var int32) *DdNode {
    var zvar, res *DdNode;

    // inlining of macro DD_ONE
    // inlining of macro DD_ZERO
    zvar = cuddUniqueInterZdd(dd, var, ((dd).one), ((dd).zero));
    if zvar == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(zvar).ref);

    res = cuddZddChangeAux(dd, P, zvar);
    if res == nil {
        Cudd_RecursiveDerefZdd(dd,zvar);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDerefZdd(dd,zvar);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of cuddZddChange */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_zddSubset1.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func zdd_subset1_aux(zdd *DdManager, P *DdNode, zvar *DdNode) *DdNode {
    var top_var, level int32;
    var res, t, e *DdNode;
    var empty *DdNode;

    statLine(zdd);
    // inlining of macro DD_ZERO
    empty = ((zdd).zero);

    /* Check cache. */
    res = cuddCacheLookup2Zdd(zdd, zdd_subset1_aux, P, zvar);
    if res != nil {
        return(res); }

    // inlining of macro cuddIsConstant
    if (P).index == CUDD_CONST_INDEX {
        res = empty;
        cuddCacheInsert2(zdd, zdd_subset1_aux, P, zvar, res);
        return(res);
    }

    top_var = zdd.permZ[P.index];
    level = zdd.permZ[zvar.index];

    if top_var > level {
        res = empty;
    } else if top_var == level {
        // inlining of macro cuddT
        res = ((P).type.kids.T);
    } else {
        // inlining of macro cuddT
        t = zdd_subset1_aux(zdd, ((P).type.kids.T), zvar);
        if t == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(t).ref);
        // inlining of macro cuddE
        e = zdd_subset1_aux(zdd, ((P).type.kids.E), zvar);
        if e == nil {
            Cudd_RecursiveDerefZdd(zdd, t);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        res = cuddZddGetNode(zdd, P.index, t, e);
        if res == nil {
            Cudd_RecursiveDerefZdd(zdd, t);
            Cudd_RecursiveDerefZdd(zdd, e);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    }

    cuddCacheInsert2(zdd, zdd_subset1_aux, P, zvar, res);

    return(res);

} /* end of zdd_subset1_aux */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_zddSubset0.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func zdd_subset0_aux(zdd *DdManager, P *DdNode, zvar *DdNode) *DdNode {
    var top_var, level int32;
    var res, t, e *DdNode;

    statLine(zdd);

    /* Check cache. */
    res = cuddCacheLookup2Zdd(zdd, zdd_subset0_aux, P, zvar);
    if res != nil {
        return(res); }

    // inlining of macro cuddIsConstant
    if (P).index == CUDD_CONST_INDEX {
        res = P;
        cuddCacheInsert2(zdd, zdd_subset0_aux, P, zvar, res);
        return(res);
    }

    top_var = zdd.permZ[P.index];
    level = zdd.permZ[zvar.index];

    if top_var > level {
        res = P;
    
    } else if top_var == level {
        // inlining of macro cuddE
        res = ((P).type.kids.E);
    
    } else {
        // inlining of macro cuddT
        t = zdd_subset0_aux(zdd, ((P).type.kids.T), zvar);
        if t == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(t).ref);
        // inlining of macro cuddE
        e = zdd_subset0_aux(zdd, ((P).type.kids.E), zvar);
        if e == nil {
            Cudd_RecursiveDerefZdd(zdd, t);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        res = cuddZddGetNode(zdd, P.index, t, e);
        if res == nil {
            Cudd_RecursiveDerefZdd(zdd, t);
            Cudd_RecursiveDerefZdd(zdd, e);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    }

    cuddCacheInsert2(zdd, zdd_subset0_aux, P, zvar, res);

    return(res);

} /* end of zdd_subset0_aux */


/**Function********************************************************************

  Synopsis    [Replaces variables with constants if possible (part of
  canonical form).]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func zddVarToConst(f *DdNode, gp **DdNode, hp **DdNode, base *DdNode, empty *DdNode) {
    var g *DdNode  = *gp;
    var h *DdNode  = *hp;

    if f == g { /* ITE(F,F,H) = ITE(F,1,H) = F + H */
        *gp = base;
    }

    if f == h { /* ITE(F,G,F) = ITE(F,G,0) = F * G */
        *hp = empty;
    }

} /* end of zddVarToConst */


//ABC_NAMESPACE_IMPL_END


