package cudd

/**CFile***********************************************************************

  FileName    [cuddBddIte.c]

  PackageName [cudd]

  Synopsis    [BDD ITE function and satellites.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_bddIte()
                <li> Cudd_bddIteConstant()
                <li> Cudd_bddIntersect()
                <li> Cudd_bddAnd()
                <li> Cudd_bddAndLimit()
                <li> Cudd_bddOr()
                <li> Cudd_bddNand()
                <li> Cudd_bddNor()
                <li> Cudd_bddXor()
                <li> Cudd_bddXnor()
                <li> Cudd_bddLeq()
                </ul>
       Internal procedures included in this module:
                <ul>
                <li> cuddBddIteRecur()
                <li> cuddBddIntersectRecur()
                <li> cuddBddAndRecur()
                <li> cuddBddXorRecur()
                </ul>
       Static procedures included in this module:
                <ul>
                <li> bddVarToConst()
                <li> bddVarToCanonical()
                <li> bddVarToCanonicalSimple()
                </ul>]

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
static char rcsid[] DD_UNUSED = "$Id: cuddBddIte.c,v 1.24 2004/08/13 18:04:46 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static void bddVarToConst (DdNode *f, DdNode **gp, DdNode **hp, DdNode *one);

// static int bddVarToCanonical (DdManager *dd, DdNode **fp, DdNode **gp, DdNode **hp, unsigned int *topfp, unsigned int *topgp, unsigned int *tophp);

// static int bddVarToCanonicalSimple (DdManager *dd, DdNode **fp, DdNode **gp, DdNode **hp, unsigned int *topfp, unsigned int *topgp, unsigned int *tophp);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Implements ITE(f,g,h).]

  Description [Implements ITE(f,g,h). Returns a pointer to the
  resulting BDD if successful; NULL if the intermediate result blows
  up.]

  SideEffects [None]

  SeeAlso     [Cudd_addIte Cudd_bddIteConstant Cudd_bddIntersect]

******************************************************************************/
func Cudd_bddIte(dd *DdManager, f *DdNode, g *DdNode, h *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddIteRecur(dd,f,g,h);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddIte */


/**Function********************************************************************

  Synopsis    [Implements ITEconstant(f,g,h).]

  Description [Implements ITEconstant(f,g,h). Returns a pointer to the
  resulting BDD (which may or may not be constant) or DD_NON_CONSTANT.
  No new nodes are created.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIte Cudd_bddIntersect Cudd_bddLeq Cudd_addIteConstant]

******************************************************************************/
func Cudd_bddIteConstant(dd *DdManager, f *DdNode, g *DdNode, h *DdNode) *DdNode {
    var r, Fv, Fnv, Gv, Gnv, H, Hv, Hnv, t, e *DdNode;
    // inlining of macro DD_ONE
    var one *DdNode  = ((dd).one);
    // inlining of macro Cudd_Not
    var zero *DdNode  = ((DdNode *)((ptrint)(one) ^ 01));
    var comple int32;
    unsigned int topf, topg, toph, v;

    statLine(dd);
    /* Trivial cases. */
    if f == one {                       /* ITE(1,G,H) => G */
        return(g); }
    
    if f == zero {                      /* ITE(0,G,H) => H */
        return(h); }
    
    /* f now not a constant. */
    bddVarToConst(f, &g, &h, one);      /* possibly convert g or h */
                                        /* to constants */

    if g == h {                         /* ITE(F,G,G) => G */
        return(g); }

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if ((((DdNode *)((ptruint)(g) & ~01))).index == CUDD_CONST_INDEX) && ((((DdNode *)((ptruint)(h) & ~01))).index == CUDD_CONST_INDEX) { 
        return(DD_NON_CONSTANT); }        /* ITE(F,1,0) or ITE(F,0,1) */
                                        /* => DD_NON_CONSTANT */
    
    // inlining of macro Cudd_Not
    if g == ((DdNode *)((ptrint)(h) ^ 01)) {
        return(DD_NON_CONSTANT); }        /* ITE(F,G,G') => DD_NON_CONSTANT */
                                        /* if F != G and F != G' */
    
    comple = bddVarToCanonical(dd, &f, &g, &h, &topf, &topg, &toph);

    /* Cache lookup. */
    r = cuddConstantLookup(dd, DD_BDD_ITE_CONSTANT_TAG, f, g, h);
    if r != nil {
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple && r != DD_NON_CONSTANT))));
    }

    // inlining of macro ddMin
    
    if (toph) < (topg) {
        v = (toph);
    } else {
        v = (topg);
    }

    /* ITE(F,G,H) = (v,G,H) (non constant) if F = (v,1,0), v < top(G,H). */
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topf < v && ((f).type.kids.T) == one && ((f).type.kids.E) == zero {
        return(DD_NON_CONSTANT);
    }

    /* Compute cofactors. */
    if topf <= v {
        // inlining of macro ddMin
        
        if (v) < (topf) {
            v = (v);
        } else {
            v = (topf);
        }             /* v = top_var(F,G,H) */
        // inlining of macro cuddT
        // inlining of macro cuddE
        Fv = ((f).type.kids.T); Fnv = ((f).type.kids.E);
    } else {
        Fv = Fnv = f;
    }

    if topg == v {
        // inlining of macro cuddT
        // inlining of macro cuddE
        Gv = ((g).type.kids.T); Gnv = ((g).type.kids.E);
    } else {
        Gv = Gnv = g;
    }

    if toph == v {
        // inlining of macro Cudd_Regular
        H = ((DdNode *)((ptruint)(h) & ~01));
        // inlining of macro cuddT
        // inlining of macro cuddE
        Hv = ((H).type.kids.T); Hnv = ((H).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (h) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Hv = ((DdNode *)((ptrint)(Hv) ^ 01));
            // inlining of macro Cudd_Not
            Hnv = ((DdNode *)((ptrint)(Hnv) ^ 01));
        }
    } else {
        Hv = Hnv = h;
    }

    /* Recursion. */
    t = Cudd_bddIteConstant(dd, Fv, Gv, Hv);
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if t == DD_NON_CONSTANT || !((((DdNode *)((ptruint)(t) & ~01))).index == CUDD_CONST_INDEX) {
        cuddCacheInsert(dd, DD_BDD_ITE_CONSTANT_TAG, f, g, h, DD_NON_CONSTANT);
        return(DD_NON_CONSTANT);
    }
    e = Cudd_bddIteConstant(dd, Fnv, Gnv, Hnv);
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if e == DD_NON_CONSTANT || !((((DdNode *)((ptruint)(e) & ~01))).index == CUDD_CONST_INDEX) || t != e {
        cuddCacheInsert(dd, DD_BDD_ITE_CONSTANT_TAG, f, g, h, DD_NON_CONSTANT);
        return(DD_NON_CONSTANT);
    }
    cuddCacheInsert(dd, DD_BDD_ITE_CONSTANT_TAG, f, g, h, t);
    // inlining of macro Cudd_NotCond
    return(((DdNode *)((ptrint)(t) ^ (comple))));

} /* end of Cudd_bddIteConstant */


/**Function********************************************************************

  Synopsis    [Returns a function included in the intersection of f and g.]

  Description [Computes a function included in the intersection of f and
  g. (That is, a witness that the intersection is not empty.)
  Cudd_bddIntersect tries to build as few new nodes as possible. If the
  only result of interest is whether f and g intersect,
  Cudd_bddLeq should be used instead.]

  SideEffects [None]

  SeeAlso     [Cudd_bddLeq Cudd_bddIteConstant]

******************************************************************************/
DdNode *
Cudd_bddIntersect(
  DdManager * dd /* manager */,
  DdNode * f /* first operand */,
  DdNode * g /* second operand */)
{
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddIntersectRecur(dd,f,g);
    
if !(dd.reordered == 1) { break }
} 

    return(res);

} /* end of Cudd_bddIntersect */


/**Function********************************************************************

  Synopsis    [Computes the conjunction of two BDDs f and g.]

  Description [Computes the conjunction of two BDDs f and g. Returns a
  pointer to the resulting BDD if successful; NULL if the intermediate
  result blows up.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIte Cudd_addApply Cudd_bddAndAbstract Cudd_bddIntersect
  Cudd_bddOr Cudd_bddNand Cudd_bddNor Cudd_bddXor Cudd_bddXnor]

******************************************************************************/
func Cudd_bddAnd(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddAndRecur(dd,f,g);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddAnd */


/**Function********************************************************************

  Synopsis    [Computes the conjunction of two BDDs f and g.  Returns
  NULL if too many nodes are required.]

  Description [Computes the conjunction of two BDDs f and g. Returns a
  pointer to the resulting BDD if successful; NULL if the intermediate
  result blows up or more new nodes than <code>limit</code> are
  required.]

  SideEffects [None]

  SeeAlso     [Cudd_bddAnd]

******************************************************************************/
DdNode *
Cudd_bddAndLimit(
  DdManager * dd,
  DdNode * f,
  DdNode * g,
  unsigned int limit)
{
    var res *DdNode;
    unsigned int saveLimit = dd.maxLive;

    dd.maxLive = (dd.keys - dd.dead) + (dd.keysZ - dd.deadZ) + limit;
    for {
        dd.reordered = 0;
        res = cuddBddAndRecur(dd,f,g);
    
if !(dd.reordered == 1) { break }
} 
    dd.maxLive = saveLimit;
    return(res);

} /* end of Cudd_bddAndLimit */


/**Function********************************************************************

  Synopsis    [Computes the disjunction of two BDDs f and g.]

  Description [Computes the disjunction of two BDDs f and g. Returns a
  pointer to the resulting BDD if successful; NULL if the intermediate
  result blows up.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIte Cudd_addApply Cudd_bddAnd Cudd_bddNand Cudd_bddNor
  Cudd_bddXor Cudd_bddXnor]

******************************************************************************/
func Cudd_bddOr(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        // inlining of macro Cudd_Not
        res = cuddBddAndRecur(dd,((DdNode *)((ptrint)(f) ^ 01)),((DdNode *)((ptrint)(g) ^ 01)));
    
if !(dd.reordered == 1) { break }
} 
    // inlining of macro Cudd_NotCond
    res = ((DdNode *)((ptrint)(res) ^ (res != nil)));
    return(res);

} /* end of Cudd_bddOr */


/**Function********************************************************************

  Synopsis    [Computes the NAND of two BDDs f and g.]

  Description [Computes the NAND of two BDDs f and g. Returns a
  pointer to the resulting BDD if successful; NULL if the intermediate
  result blows up.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIte Cudd_addApply Cudd_bddAnd Cudd_bddOr Cudd_bddNor
  Cudd_bddXor Cudd_bddXnor]

******************************************************************************/
func Cudd_bddNand(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddAndRecur(dd,f,g);
    
if !(dd.reordered == 1) { break }
} 
    // inlining of macro Cudd_NotCond
    res = ((DdNode *)((ptrint)(res) ^ (res != nil)));
    return(res);

} /* end of Cudd_bddNand */


/**Function********************************************************************

  Synopsis    [Computes the NOR of two BDDs f and g.]

  Description [Computes the NOR of two BDDs f and g. Returns a
  pointer to the resulting BDD if successful; NULL if the intermediate
  result blows up.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIte Cudd_addApply Cudd_bddAnd Cudd_bddOr Cudd_bddNand
  Cudd_bddXor Cudd_bddXnor]

******************************************************************************/
func Cudd_bddNor(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        // inlining of macro Cudd_Not
        res = cuddBddAndRecur(dd,((DdNode *)((ptrint)(f) ^ 01)),((DdNode *)((ptrint)(g) ^ 01)));
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddNor */


/**Function********************************************************************

  Synopsis    [Computes the exclusive OR of two BDDs f and g.]

  Description [Computes the exclusive OR of two BDDs f and g. Returns a
  pointer to the resulting BDD if successful; NULL if the intermediate
  result blows up.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIte Cudd_addApply Cudd_bddAnd Cudd_bddOr
  Cudd_bddNand Cudd_bddNor Cudd_bddXnor]

******************************************************************************/
func Cudd_bddXor(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddXorRecur(dd,f,g);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddXor */


/**Function********************************************************************

  Synopsis    [Computes the exclusive NOR of two BDDs f and g.]

  Description [Computes the exclusive NOR of two BDDs f and g. Returns a
  pointer to the resulting BDD if successful; NULL if the intermediate
  result blows up.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIte Cudd_addApply Cudd_bddAnd Cudd_bddOr
  Cudd_bddNand Cudd_bddNor Cudd_bddXor]

******************************************************************************/
func Cudd_bddXnor(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        // inlining of macro Cudd_Not
        res = cuddBddXorRecur(dd,f,((DdNode *)((ptrint)(g) ^ 01)));
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddXnor */


/**Function********************************************************************

  Synopsis    [Determines whether f is less than or equal to g.]

  Description [Returns 1 if f is less than or equal to g; 0 otherwise.
  No new nodes are created.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIteConstant Cudd_addEvalConst]

******************************************************************************/
func Cudd_bddLeq(dd *DdManager, f *DdNode, g *DdNode) int32 {
    var one, zero, tmp, F, fv, fvn, gv, gvn *DdNode;
    unsigned int topf, topg, res;

    statLine(dd);
    /* Terminal cases and normalization. */
    if f == g { return(1); }

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (g) & 01)) != 0 {
        /* Special case: if f is regular and g is complemented,
        ** f(1,...,1) = 1 > 0 = g(1,...,1).
        */
        // inlining of macro Cudd_IsComplement
        if !((int32( ((ptrint) (f) & 01)) != 0)) { return(0); }
        /* Both are complemented: Swap and complement because
        ** f <= g <=> g' <= f' and we want the second argument to be regular.
        */
        tmp = g;
        // inlining of macro Cudd_Not
        g = ((DdNode *)((ptrint)(f) ^ 01));
        // inlining of macro Cudd_Not
        f = ((DdNode *)((ptrint)(tmp) ^ 01));
    // inlining of macro Cudd_IsComplement
    // inlining of macro cuddF2L
    // inlining of macro Cudd_Regular
    } else if (int32( ((ptrint) (f) & 01))) && ((((DdNode *)((ptruint)(g) & ~01)).Id << 1) | (int32( ((ptrint) (g) & 01)))) < ((((DdNode *)((ptruint)(f) & ~01)).Id << 1) | (int32( ((ptrint) (f) & 01)))) {
        tmp = g;
        // inlining of macro Cudd_Not
        g = ((DdNode *)((ptrint)(f) ^ 01));
        // inlining of macro Cudd_Not
        f = ((DdNode *)((ptrint)(tmp) ^ 01));
    }

    /* Now g is regular and, if f is not regular, f < g. */
    // inlining of macro DD_ONE
    one = ((dd).one);
    if g == one { return(1); }    /* no need to test against zero */
    if f == one { return(0); }    /* since at this point g != one */
    // inlining of macro Cudd_Not
    if ((DdNode *)((ptrint)(f) ^ 01)) == g { return(0); } /* because neither is constant */
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));
    if f == zero { return(1); }

    /* Here neither f nor g is constant. */

    /* Check cache. */
    tmp = cuddCacheLookup2(dd,(DD_CTFP)Cudd_bddLeq,f,g);
    if tmp != nil {
        return(tmp == one);
    }

    /* Compute cofactors. */
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    topf = dd.perm[F.index];
    topg = dd.perm[g.index];
    if topf <= topg {
        // inlining of macro cuddT
        // inlining of macro cuddE
        fv = ((F).type.kids.T); fvn = ((F).type.kids.E);
        if f != F {
            // inlining of macro Cudd_Not
            fv = ((DdNode *)((ptrint)(fv) ^ 01));
            // inlining of macro Cudd_Not
            fvn = ((DdNode *)((ptrint)(fvn) ^ 01));
        }
    } else {
        fv = fvn = f;
    }
    if topg <= topf {
        // inlining of macro cuddT
        // inlining of macro cuddE
        gv = ((g).type.kids.T); gvn = ((g).type.kids.E);
    } else {
        gv = gvn = g;
    }

    /* Recursive calls. Since we want to maximize the probability of
    ** the special case f(1,...,1) > g(1,...,1), we consider the negative
    ** cofactors first. Indeed, the complementation parity of the positive
    ** cofactors is the same as the one of the parent functions.
    */
    res = Cudd_bddLeq(dd,fvn,gvn) && Cudd_bddLeq(dd,fv,gv);

    /* Store result in cache and return. */
    cuddCacheInsert2(dd,(DD_CTFP)Cudd_bddLeq,f,g,(res ? one : zero));
    return(res);

} /* end of Cudd_bddLeq */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Implements the recursive step of Cudd_bddIte.]

  Description [Implements the recursive step of Cudd_bddIte. Returns a
  pointer to the resulting BDD. NULL if the intermediate result blows
  up or if reordering occurs.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddBddIteRecur(dd *DdManager, f *DdNode, g *DdNode, h *DdNode) *DdNode {
    var one, zero, res *DdNode;
    var r, Fv, Fnv, Gv, Gnv, H, Hv, Hnv, t, e *DdNode;
    unsigned int topf, topg, toph, v;
    var index int32  = -1;
    var comple int32;

    statLine(dd);
    /* Terminal cases. */

    /* One variable cases. */
    // inlining of macro DD_ONE
    if f == (one = ((dd).one)) {        /* ITE(1,G,H) = G */
        return(g); }
    
    // inlining of macro Cudd_Not
    if f == (zero = ((DdNode *)((ptrint)(one) ^ 01))) {    /* ITE(0,G,H) = H */
        return(h); }
    
    /* From now on, f is known not to be a constant. */
    if g == one || f == g {   /* ITE(F,F,H) = ITE(F,1,H) = F + H */
        if h == zero {        /* ITE(F,1,0) = F */
            return(f);
        } else {
            // inlining of macro Cudd_Not
            res = cuddBddAndRecur(dd,((DdNode *)((ptrint)(f) ^ 01)),((DdNode *)((ptrint)(h) ^ 01)));
            // inlining of macro Cudd_NotCond
            return(((DdNode *)((ptrint)(res) ^ (res != nil))));
        }
    // inlining of macro Cudd_Not
    } else if g == zero || f == ((DdNode *)((ptrint)(g) ^ 01)) { /* ITE(F,!F,H) = ITE(F,0,H) = !F * H */
        if h == one {         /* ITE(F,0,1) = !F */
            // inlining of macro Cudd_Not
            return(((DdNode *)((ptrint)(f) ^ 01)));
        } else {
            // inlining of macro Cudd_Not
            res = cuddBddAndRecur(dd,((DdNode *)((ptrint)(f) ^ 01)),h);
            return(res);
        }
    }
    if h == zero || f == h {    /* ITE(F,G,F) = ITE(F,G,0) = F * G */
        res = cuddBddAndRecur(dd,f,g);
        return(res);
    // inlining of macro Cudd_Not
    } else if h == one || f == ((DdNode *)((ptrint)(h) ^ 01)) { /* ITE(F,G,!F) = ITE(F,G,1) = !F + G */
        // inlining of macro Cudd_Not
        res = cuddBddAndRecur(dd,f,((DdNode *)((ptrint)(g) ^ 01)));
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(res) ^ (res != nil))));
    }

    /* Check remaining one variable case. */
    if g == h {               /* ITE(F,G,G) = G */
        return(g);
    // inlining of macro Cudd_Not
    } else if g == ((DdNode *)((ptrint)(h) ^ 01)) { /* ITE(F,G,!G) = F <-> G */
        res = cuddBddXorRecur(dd,f,h);
        return(res);
    }
    
    /* From here, there are no constants. */
    comple = bddVarToCanonicalSimple(dd, &f, &g, &h, &topf, &topg, &toph);

    /* f & g are now regular pointers */

    // inlining of macro ddMin
    
    if (toph) < (topg) {
        v = (toph);
    } else {
        v = (topg);
    }

    /* A shortcut: ITE(F,G,H) = (v,G,H) if F = (v,1,0), v < top(G,H). */
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topf < v && ((f).type.kids.T) == one && ((f).type.kids.E) == zero {
        r = cuddUniqueInter(dd, int32( f.index), g, h);
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple && r != nil))));
    }

    /* Check cache. */
    r = cuddCacheLookup(dd, DD_BDD_ITE_TAG, f, g, h);
    if r != nil {
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    }

    /* Compute cofactors. */
    if topf <= v {
        // inlining of macro ddMin
        
        if (v) < (topf) {
            v = (v);
        } else {
            v = (topf);
        }     /* v = top_var(F,G,H) */
        index = f.index;
        // inlining of macro cuddT
        // inlining of macro cuddE
        Fv = ((f).type.kids.T); Fnv = ((f).type.kids.E);
    } else {
        Fv = Fnv = f;
    }
    if topg == v {
        index = g.index;
        // inlining of macro cuddT
        // inlining of macro cuddE
        Gv = ((g).type.kids.T); Gnv = ((g).type.kids.E);
    } else {
        Gv = Gnv = g;
    }
    if toph == v {
        // inlining of macro Cudd_Regular
        H = ((DdNode *)((ptruint)(h) & ~01));
        index = H.index;
        // inlining of macro cuddT
        // inlining of macro cuddE
        Hv = ((H).type.kids.T); Hnv = ((H).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (h) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Hv = ((DdNode *)((ptrint)(Hv) ^ 01));
            // inlining of macro Cudd_Not
            Hnv = ((DdNode *)((ptrint)(Hnv) ^ 01));
        }
    } else {
        Hv = Hnv = h;
    }

    /* Recursive step. */
    t = cuddBddIteRecur(dd,Fv,Gv,Hv);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = cuddBddIteRecur(dd,Fnv,Gnv,Hnv);
    if e == nil {
        Cudd_IterDerefBdd(dd,t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    if t == e {
        r = t;
    } else {
        r = cuddUniqueInter(dd,index,t,e);
    }
    if r == nil {
        Cudd_IterDerefBdd(dd,t);
        Cudd_IterDerefBdd(dd,e);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);

    cuddCacheInsert(dd, DD_BDD_ITE_TAG, f, g, h, r);
    // inlining of macro Cudd_NotCond
    return(((DdNode *)((ptrint)(r) ^ (comple))));

} /* end of cuddBddIteRecur */


/**Function********************************************************************

  Synopsis    [Implements the recursive step of Cudd_bddIntersect.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_bddIntersect]

******************************************************************************/
func cuddBddIntersectRecur(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var res *DdNode;
    var F, G, t, e *DdNode;
    var fv, fnv, gv, gnv *DdNode;
    var one, zero *DdNode;
    unsigned int index, topf, topg;

    statLine(dd);
    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Terminal cases. */
    // inlining of macro Cudd_Not
    if f == zero || g == zero || f == ((DdNode *)((ptrint)(g) ^ 01)) { return(zero); }
    if f == g || g == one { return(f); }
    if f == one { return(g); }

    /* At this point f and g are not constant. */
    // inlining of macro cuddF2L
    // inlining of macro Cudd_Regular
    // inlining of macro Cudd_IsComplement
    if ((((DdNode *)((ptruint)(f) & ~01)).Id << 1) | (int32( ((ptrint) (f) & 01)))) > ((((DdNode *)((ptruint)(g) & ~01)).Id << 1) | (int32( ((ptrint) (g) & 01)))) { var tmp *DdNode  = f; f = g; g = tmp; }
    res = cuddCacheLookup2(dd,Cudd_bddIntersect,f,g);
    if res != nil { return(res); }

    /* Find splitting variable. Here we can skip the use of cuddI,
    ** because the operands are known to be non-constant.
    */
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    topf = dd.perm[F.index];
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    topg = dd.perm[G.index];

    /* Compute cofactors. */
    if topf <= topg {
        index = F.index;
        // inlining of macro cuddT
        fv = ((F).type.kids.T);
        // inlining of macro cuddE
        fnv = ((F).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (f) & 01)) != 0 {
            // inlining of macro Cudd_Not
            fv = ((DdNode *)((ptrint)(fv) ^ 01));
            // inlining of macro Cudd_Not
            fnv = ((DdNode *)((ptrint)(fnv) ^ 01));
        }
    } else {
        index = G.index;
        fv = fnv = f;
    }

    if topg <= topf {
        // inlining of macro cuddT
        gv = ((G).type.kids.T);
        // inlining of macro cuddE
        gnv = ((G).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (g) & 01)) != 0 {
            // inlining of macro Cudd_Not
            gv = ((DdNode *)((ptrint)(gv) ^ 01));
            // inlining of macro Cudd_Not
            gnv = ((DdNode *)((ptrint)(gnv) ^ 01));
        }
    } else {
        gv = gnv = g;
    }

    /* Compute partial results. */
    t = cuddBddIntersectRecur(dd,fv,gv);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);
    if t != zero {
        e = zero;
    } else {
        e = cuddBddIntersectRecur(dd,fnv,gnv);
        if e == nil {
            Cudd_IterDerefBdd(dd, t);
            return(nil);
        }
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    if t == e { /* both equal zero */
        res = t;
    // inlining of macro Cudd_IsComplement
    } else if int32( ((ptrint) (t) & 01)) != 0 {
        // inlining of macro Cudd_Not
        res = cuddUniqueInter(dd,int32(index),((DdNode *)((ptrint)(t) ^ 01)),((DdNode *)((ptrint)(e) ^ 01)));
        if res == nil {
            Cudd_IterDerefBdd(dd, t);
            Cudd_IterDerefBdd(dd, e);
            return(nil);
        }
        // inlining of macro Cudd_Not
        res = ((DdNode *)((ptrint)(res) ^ 01));
    } else {
        res = cuddUniqueInter(dd,int32(index),t,e);
        if res == nil {
            Cudd_IterDerefBdd(dd, t);
            Cudd_IterDerefBdd(dd, e);
            return(nil);
        }
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);

    cuddCacheInsert2(dd,Cudd_bddIntersect,f,g,res);

    return(res);

} /* end of cuddBddIntersectRecur */


/**Function********************************************************************

  Synopsis [Implements the recursive step of Cudd_bddAnd.]

  Description [Implements the recursive step of Cudd_bddAnd by taking
  the conjunction of two BDDs.  Returns a pointer to the result is
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddAnd]

******************************************************************************/
func cuddBddAndRecur(manager *DdManager, f *DdNode, g *DdNode) *DdNode {
    var F, fv, fnv, G, gv, gnv *DdNode;
    var one, r, t, e *DdNode;
    unsigned int topf, topg, index;

    statLine(manager);
    // inlining of macro DD_ONE
    one = ((manager).one);

    /* Terminal cases. */
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    if F == G {
        if f == g { return(f);
        // inlining of macro Cudd_Not
        } else { return(((DdNode *)((ptrint)(one) ^ 01))); }
    }
    if F == one {
        if f == one { return(g);
        } else { return(f); }
    }
    if G == one {
        if g == one { return(f);
        } else { return(g); }
    }

    /* At this point f and g are not constant. */
    // inlining of macro cuddF2L
    // inlining of macro Cudd_Regular
    // inlining of macro Cudd_IsComplement
    if ((((DdNode *)((ptruint)(f) & ~01)).Id << 1) | (int32( ((ptrint) (f) & 01)))) > ((((DdNode *)((ptruint)(g) & ~01)).Id << 1) | (int32( ((ptrint) (g) & 01)))) { /* Try to increase cache efficiency. */
        var tmp *DdNode  = f;
        f = g;
        g = tmp;
        // inlining of macro Cudd_Regular
        F = ((DdNode *)((ptruint)(f) & ~01));
        // inlining of macro Cudd_Regular
        G = ((DdNode *)((ptruint)(g) & ~01));
    }

    /* Check cache. */
    if F.ref != 1 || G.ref != 1 {
        r = cuddCacheLookup2(manager, Cudd_bddAnd, f, g);
        if r != nil { return(r); }
    }

    if manager.TimeStop && Abc_Clock() > manager.TimeStop {
        return nil; }

    /* Here we can skip the use of cuddI, because the operands are known
    ** to be non-constant.
    */
    topf = manager.perm[F.index];
    topg = manager.perm[G.index];

    /* Compute cofactors. */
    if topf <= topg {
        index = F.index;
        // inlining of macro cuddT
        fv = ((F).type.kids.T);
        // inlining of macro cuddE
        fnv = ((F).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (f) & 01)) != 0 {
            // inlining of macro Cudd_Not
            fv = ((DdNode *)((ptrint)(fv) ^ 01));
            // inlining of macro Cudd_Not
            fnv = ((DdNode *)((ptrint)(fnv) ^ 01));
        }
    } else {
        index = G.index;
        fv = fnv = f;
    }

    if topg <= topf {
        // inlining of macro cuddT
        gv = ((G).type.kids.T);
        // inlining of macro cuddE
        gnv = ((G).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (g) & 01)) != 0 {
            // inlining of macro Cudd_Not
            gv = ((DdNode *)((ptrint)(gv) ^ 01));
            // inlining of macro Cudd_Not
            gnv = ((DdNode *)((ptrint)(gnv) ^ 01));
        }
    } else {
        gv = gnv = g;
    }

    t = cuddBddAndRecur(manager, fv, gv);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = cuddBddAndRecur(manager, fnv, gnv);
    if e == nil {
        Cudd_IterDerefBdd(manager, t);
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
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    if F.ref != 1 || G.ref != 1 {
        cuddCacheInsert2(manager, Cudd_bddAnd, f, g, r); }
    return(r);

} /* end of cuddBddAndRecur */


/**Function********************************************************************

  Synopsis [Implements the recursive step of Cudd_bddXor.]

  Description [Implements the recursive step of Cudd_bddXor by taking
  the exclusive OR of two BDDs.  Returns a pointer to the result is
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddXor]

******************************************************************************/
func cuddBddXorRecur(manager *DdManager, f *DdNode, g *DdNode) *DdNode {
    var fv, fnv, G, gv, gnv *DdNode;
    var one, zero, r, t, e *DdNode;
    unsigned int topf, topg, index;

    statLine(manager);
    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Terminal cases. */
    if f == g { return(zero); }
    // inlining of macro Cudd_Not
    if f == ((DdNode *)((ptrint)(g) ^ 01)) { return(one); }
    // inlining of macro cuddF2L
    // inlining of macro Cudd_Regular
    // inlining of macro Cudd_IsComplement
    if ((((DdNode *)((ptruint)(f) & ~01)).Id << 1) | (int32( ((ptrint) (f) & 01)))) > ((((DdNode *)((ptruint)(g) & ~01)).Id << 1) | (int32( ((ptrint) (g) & 01)))) { /* Try to increase cache efficiency and simplify tests. */
        var tmp *DdNode  = f;
        f = g;
        g = tmp;
    }
    if g == zero { return(f); }
    // inlining of macro Cudd_Not
    if g == one { return(((DdNode *)((ptrint)(f) ^ 01))); }
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (f) & 01)) != 0 {
        // inlining of macro Cudd_Not
        f = ((DdNode *)((ptrint)(f) ^ 01));
        // inlining of macro Cudd_Not
        g = ((DdNode *)((ptrint)(g) ^ 01));
    }
    /* Now the first argument is regular. */
    // inlining of macro Cudd_Not
    if f == one { return(((DdNode *)((ptrint)(g) ^ 01))); }

    /* At this point f and g are not constant. */

    /* Check cache. */
    r = cuddCacheLookup2(manager, Cudd_bddXor, f, g);
    if r != nil { return(r); }

    /* Here we can skip the use of cuddI, because the operands are known
    ** to be non-constant.
    */
    topf = manager.perm[f.index];
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    topg = manager.perm[G.index];

    /* Compute cofactors. */
    if topf <= topg {
        index = f.index;
        // inlining of macro cuddT
        fv = ((f).type.kids.T);
        // inlining of macro cuddE
        fnv = ((f).type.kids.E);
    } else {
        index = G.index;
        fv = fnv = f;
    }

    if topg <= topf {
        // inlining of macro cuddT
        gv = ((G).type.kids.T);
        // inlining of macro cuddE
        gnv = ((G).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (g) & 01)) != 0 {
            // inlining of macro Cudd_Not
            gv = ((DdNode *)((ptrint)(gv) ^ 01));
            // inlining of macro Cudd_Not
            gnv = ((DdNode *)((ptrint)(gnv) ^ 01));
        }
    } else {
        gv = gnv = g;
    }

    t = cuddBddXorRecur(manager, fv, gv);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = cuddBddXorRecur(manager, fnv, gnv);
    if e == nil {
        Cudd_IterDerefBdd(manager, t);
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
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    cuddCacheInsert2(manager, Cudd_bddXor, f, g, r);
    return(r);

} /* end of cuddBddXorRecur */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Replaces variables with constants if possible.]

  Description [This function performs part of the transformation to
  standard form by replacing variables with constants if possible.]

  SideEffects [None]

  SeeAlso     [bddVarToCanonical bddVarToCanonicalSimple]

******************************************************************************/
func bddVarToConst(f *DdNode, gp **DdNode, hp **DdNode, one *DdNode) {
    var g *DdNode  = *gp;
    var h *DdNode  = *hp;

    if f == g {    /* ITE(F,F,H) = ITE(F,1,H) = F + H */
        *gp = one;
    // inlining of macro Cudd_Not
    } else if f == ((DdNode *)((ptrint)(g) ^ 01)) {    /* ITE(F,!F,H) = ITE(F,0,H) = !F * H */
        // inlining of macro Cudd_Not
        *gp = ((DdNode *)((ptrint)(one) ^ 01));
    }
    if f == h {    /* ITE(F,G,F) = ITE(F,G,0) = F * G */
        // inlining of macro Cudd_Not
        *hp = ((DdNode *)((ptrint)(one) ^ 01));
    // inlining of macro Cudd_Not
    } else if f == ((DdNode *)((ptrint)(h) ^ 01)) {    /* ITE(F,G,!F) = ITE(F,G,1) = !F + G */
        *hp = one;
    }

} /* end of bddVarToConst */


/**Function********************************************************************

  Synopsis [Picks unique member from equiv expressions.]

  Description [Reduces 2 variable expressions to canonical form.]

  SideEffects [None]

  SeeAlso     [bddVarToConst bddVarToCanonicalSimple]

******************************************************************************/
static int
bddVarToCanonical(
  DdManager * dd,
  DdNode ** fp,
  DdNode ** gp,
  DdNode ** hp,
  unsigned int * topfp,
  unsigned int * topgp,
  unsigned int * tophp)
{
    register DdNode             *F, *G, *H, *r, *f, *g, *h;
    register unsigned int       topf, topg, toph;
    var one *DdNode  = dd.one;
    var comple, change int32;

    f = *fp;
    g = *gp;
    h = *hp;
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    // inlining of macro Cudd_Regular
    H = ((DdNode *)((ptruint)(h) & ~01));
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
    // inlining of macro cuddI
    
    if (H.index)==CUDD_CONST_INDEX {
        toph = int32((H.index));
    } else {
        toph = (dd).perm[(H.index)];
    }

    change = 0;

    if G == one {                     /* ITE(F,c,H) */
        // inlining of macro cuddF2L
        // inlining of macro Cudd_Regular
        // inlining of macro Cudd_IsComplement
        if (topf > toph) || (topf == toph && ((((DdNode *)((ptruint)(f) & ~01)).Id << 1) | (int32( ((ptrint) (f) & 01)))) > ((((DdNode *)((ptruint)(h) & ~01)).Id << 1) | (int32( ((ptrint) (h) & 01))))) {
            r = h;
            h = f;
            f = r;                      /* ITE(F,1,H) = ITE(H,1,F) */
            if g != one {     /* g == zero */
                // inlining of macro Cudd_Not
                f = ((DdNode *)((ptrint)(f) ^ 01));                /* ITE(F,0,H) = ITE(!H,0,!F) */
                // inlining of macro Cudd_Not
                h = ((DdNode *)((ptrint)(h) ^ 01));
            }
            change = 1;
        }
    } else if H == one {              /* ITE(F,G,c) */
        // inlining of macro cuddF2L
        // inlining of macro Cudd_Regular
        // inlining of macro Cudd_IsComplement
        if (topf > topg) || (topf == topg && ((((DdNode *)((ptruint)(f) & ~01)).Id << 1) | (int32( ((ptrint) (f) & 01)))) > ((((DdNode *)((ptruint)(g) & ~01)).Id << 1) | (int32( ((ptrint) (g) & 01))))) {
            r = g;
            g = f;
            f = r;                      /* ITE(F,G,0) = ITE(G,F,0) */
            if h == one {
                // inlining of macro Cudd_Not
                f = ((DdNode *)((ptrint)(f) ^ 01));                /* ITE(F,G,1) = ITE(!G,!F,1) */
                // inlining of macro Cudd_Not
                g = ((DdNode *)((ptrint)(g) ^ 01));
            }
            change = 1;
        }
    // inlining of macro Cudd_Not
    } else if g == ((DdNode *)((ptrint)(h) ^ 01)) {      /* ITE(F,G,!G) = ITE(G,F,!F) */
        // inlining of macro cuddF2L
        // inlining of macro Cudd_Regular
        // inlining of macro Cudd_IsComplement
        if (topf > topg) || (topf == topg && ((((DdNode *)((ptruint)(f) & ~01)).Id << 1) | (int32( ((ptrint) (f) & 01)))) > ((((DdNode *)((ptruint)(g) & ~01)).Id << 1) | (int32( ((ptrint) (g) & 01))))) {
            r = f;
            f = g;
            g = r;
            // inlining of macro Cudd_Not
            h = ((DdNode *)((ptrint)(r) ^ 01));
            change = 1;
        }
    }
    /* adjust pointers so that the first 2 arguments to ITE are regular */
    // inlining of macro Cudd_IsComplement
    if (int32( ((ptrint) (f) & 01))) != 0 {    /* ITE(!F,G,H) = ITE(F,H,G) */
        // inlining of macro Cudd_Not
        f = ((DdNode *)((ptrint)(f) ^ 01));
        r = g;
        g = h;
        h = r;
        change = 1;
    }
    comple = 0;
    // inlining of macro Cudd_IsComplement
    if (int32( ((ptrint) (g) & 01))) != 0 {    /* ITE(F,!G,H) = !ITE(F,G,!H) */
        // inlining of macro Cudd_Not
        g = ((DdNode *)((ptrint)(g) ^ 01));
        // inlining of macro Cudd_Not
        h = ((DdNode *)((ptrint)(h) ^ 01));
        change = 1;
        comple = 1;
    }
    if change != 0 {
        *fp = f;
        *gp = g;
        *hp = h;
    }
    // inlining of macro cuddI
    *topfp = (((f.index)==CUDD_CONST_INDEX)?int32((f.index)):(dd).perm[(f.index)]);
    // inlining of macro cuddI
    *topgp = (((g.index)==CUDD_CONST_INDEX)?int32((g.index)):(dd).perm[(g.index)]);
    // inlining of macro cuddI
    // inlining of macro Cudd_Regular
    *tophp = (((((DdNode *)((ptruint)(h) & ~01)).index)==CUDD_CONST_INDEX)?int32((((DdNode *)((ptruint)(h) & ~01)).index)):(dd).perm[(((DdNode *)((ptruint)(h) & ~01)).index)]);

    return(comple);

} /* end of bddVarToCanonical */


/**Function********************************************************************

  Synopsis [Picks unique member from equiv expressions.]

  Description [Makes sure the first two pointers are regular.  This
  mat require the complementation of the result, which is signaled by
  returning 1 instead of 0.  This function is simpler than the general
  case because it assumes that no two arguments are the same or
  complementary, and no argument is constant.]

  SideEffects [None]

  SeeAlso     [bddVarToConst bddVarToCanonical]

******************************************************************************/
static int
bddVarToCanonicalSimple(
  DdManager * dd,
  DdNode ** fp,
  DdNode ** gp,
  DdNode ** hp,
  unsigned int * topfp,
  unsigned int * topgp,
  unsigned int * tophp)
{
    register DdNode             *r, *f, *g, *h;
    var comple, change int32;

    f = *fp;
    g = *gp;
    h = *hp;

    change = 0;

    /* adjust pointers so that the first 2 arguments to ITE are regular */
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (f) & 01)) != 0 { /* ITE(!F,G,H) = ITE(F,H,G) */
        // inlining of macro Cudd_Not
        f = ((DdNode *)((ptrint)(f) ^ 01));
        r = g;
        g = h;
        h = r;
        change = 1;
    }
    comple = 0;
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (g) & 01)) != 0 { /* ITE(F,!G,H) = !ITE(F,G,!H) */
        // inlining of macro Cudd_Not
        g = ((DdNode *)((ptrint)(g) ^ 01));
        // inlining of macro Cudd_Not
        h = ((DdNode *)((ptrint)(h) ^ 01));
        change = 1;
        comple = 1;
    }
    if change != 0 {
        *fp = f;
        *gp = g;
        *hp = h;
    }

    /* Here we can skip the use of cuddI, because the operands are known
    ** to be non-constant.
    */
    *topfp = dd.perm[f.index];
    *topgp = dd.perm[g.index];
    // inlining of macro Cudd_Regular
    *tophp = dd.perm[((DdNode *)((ptruint)(h) & ~01)).index];

    return(comple);

} /* end of bddVarToCanonicalSimple */


//ABC_NAMESPACE_IMPL_END


