package cudd

/**CFile***********************************************************************

  FileName    [cuddAddIte.c]

  PackageName [cudd]

  Synopsis    [ADD ITE function and satellites.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addIte()
                <li> Cudd_addIteConstant()
                <li> Cudd_addEvalConst()
                <li> Cudd_addCmpl()
                <li> Cudd_addLeq()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddAddIteRecur()
                <li> cuddAddCmplRecur()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> addVarToConst()
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
static char rcsid[] DD_UNUSED = "$Id: cuddAddIte.c,v 1.15 2004/08/13 18:04:45 fabio Exp $";
#endif


/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static void addVarToConst (DdNode *f, DdNode **gp, DdNode **hp, DdNode *one, DdNode *zero);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Implements ITE(f,g,h).]

  Description [Implements ITE(f,g,h). This procedure assumes that f is
  a 0-1 ADD.  Returns a pointer to the resulting ADD if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIte Cudd_addIteConstant Cudd_addApply]

******************************************************************************/
func Cudd_addIte(dd *DdManager, f *DdNode, g *DdNode, h *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddAddIteRecur(dd,f,g,h);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_addIte */


/**Function********************************************************************

  Synopsis    [Implements ITEconstant for ADDs.]

  Description [Implements ITEconstant for ADDs.  f must be a 0-1 ADD.
  Returns a pointer to the resulting ADD (which may or may not be
  constant) or DD_NON_CONSTANT. No new nodes are created. This function
  can be used, for instance, to check that g has a constant value
  (specified by h) whenever f is 1. If the constant value is unknown,
  then one should use Cudd_addEvalConst.]

  SideEffects [None]

  SeeAlso     [Cudd_addIte Cudd_addEvalConst Cudd_bddIteConstant]

******************************************************************************/
func Cudd_addIteConstant(dd *DdManager, f *DdNode, g *DdNode, h *DdNode) *DdNode {
    var one, zero *DdNode;
    var Fv, Fnv, Gv, Gnv, Hv, Hnv, r, t, e *DdNode;
    unsigned int topf,topg,toph,v;

    statLine(dd);
    /* Trivial cases. */
    // inlining of macro DD_ONE
    if f == (one = ((dd).one)) {      /* ITE(1,G,H) = G */
        return(g);
    }
    // inlining of macro DD_ZERO
    if f == (zero = ((dd).zero)) {    /* ITE(0,G,H) = H */
        return(h);
    }

    /* From now on, f is known not to be a constant. */
    addVarToConst(f,&g,&h,one,zero);

    /* Check remaining one variable cases. */
    if g == h {                       /* ITE(F,G,G) = G */
        return(g);
    }
    // inlining of macro cuddIsConstant
    if ((g).index == CUDD_CONST_INDEX) && ((h).index == CUDD_CONST_INDEX) {
        return(DD_NON_CONSTANT);
    }

    // inlining of macro cuddI
    
    if (f.index)==CUDD_CONST_INDEX {
        topf = int32((f.index));
    } else {
        topf = (dd).perm[(f.index)];
    }
    // inlining of macro cuddI
    
    if (g.index)==CUDD_CONST_INDEX {
        topg = int32((g.index));
    } else {
        topg = (dd).perm[(g.index)];
    }
    // inlining of macro cuddI
    
    if (h.index)==CUDD_CONST_INDEX {
        toph = int32((h.index));
    } else {
        toph = (dd).perm[(h.index)];
    }
    // inlining of macro ddMin
    
    if (toph) < (topg) {
        v = (toph);
    } else {
        v = (topg);
    }

    /* ITE(F,G,H) = (x,G,H) (non constant) if F = (x,1,0), x < top(G,H). */
    // inlining of macro cuddIsConstant
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topf < v && ((((f).type.kids.T)).index == CUDD_CONST_INDEX) && ((((f).type.kids.E)).index == CUDD_CONST_INDEX) {
        return(DD_NON_CONSTANT);
    }

    /* Check cache. */
    r = cuddConstantLookup(dd,DD_ADD_ITE_CONSTANT_TAG,f,g,h);
    if r != nil {
        return(r);
    }

    /* Compute cofactors. */
    if topf <= v {
        // inlining of macro ddMin
        
        if (v) < (topf) {
            v = (v);
        } else {
            v = (topf);
        }      /* v = top_var(F,G,H) */
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
        // inlining of macro cuddT
        // inlining of macro cuddE
        Hv = ((h).type.kids.T); Hnv = ((h).type.kids.E);
    } else {
        Hv = Hnv = h;
    }
    
    /* Recursive step. */
    t = Cudd_addIteConstant(dd,Fv,Gv,Hv);
    // inlining of macro cuddIsConstant
    if t == DD_NON_CONSTANT || !((t).index == CUDD_CONST_INDEX) {
        cuddCacheInsert(dd, DD_ADD_ITE_CONSTANT_TAG, f, g, h, DD_NON_CONSTANT);
        return(DD_NON_CONSTANT);
    }
    e = Cudd_addIteConstant(dd,Fnv,Gnv,Hnv);
    // inlining of macro cuddIsConstant
    if e == DD_NON_CONSTANT || !((e).index == CUDD_CONST_INDEX) || t != e {
        cuddCacheInsert(dd, DD_ADD_ITE_CONSTANT_TAG, f, g, h, DD_NON_CONSTANT);
        return(DD_NON_CONSTANT);
    }
    cuddCacheInsert(dd, DD_ADD_ITE_CONSTANT_TAG, f, g, h, t);
    return(t);

} /* end of Cudd_addIteConstant */


/**Function********************************************************************

  Synopsis    [Checks whether ADD g is constant whenever ADD f is 1.]

  Description [Checks whether ADD g is constant whenever ADD f is 1.  f
  must be a 0-1 ADD.  Returns a pointer to the resulting ADD (which may
  or may not be constant) or DD_NON_CONSTANT. If f is identically 0,
  the check is assumed to be successful, and the background value is
  returned. No new nodes are created.]

  SideEffects [None]

  SeeAlso     [Cudd_addIteConstant Cudd_addLeq]

******************************************************************************/
func Cudd_addEvalConst(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var zero *DdNode;
    var Fv, Fnv, Gv, Gnv, r, t, e *DdNode;
    unsigned int topf,topg;

#ifdef DD_DEBUG
    assert(!Cudd_IsComplement(f));
#endif

    statLine(dd);
    /* Terminal cases. */
    // inlining of macro DD_ONE
    // inlining of macro cuddIsConstant
    if f == ((dd).one) || ((g).index == CUDD_CONST_INDEX) {
        return(g);
    }
    // inlining of macro DD_ZERO
    if f == (zero = ((dd).zero)) {
        return(dd.background);
    }

#ifdef DD_DEBUG
    assert(!cuddIsConstant(f));
#endif
    /* From now on, f and g are known not to be constants. */

    // inlining of macro cuddI
    
    if (f.index)==CUDD_CONST_INDEX {
        topf = int32((f.index));
    } else {
        topf = (dd).perm[(f.index)];
    }
    // inlining of macro cuddI
    
    if (g.index)==CUDD_CONST_INDEX {
        topg = int32((g.index));
    } else {
        topg = (dd).perm[(g.index)];
    }

    /* Check cache. */
    r = cuddConstantLookup(dd,DD_ADD_EVAL_CONST_TAG,f,g,g);
    if r != nil {
        return(r);
    }

    /* Compute cofactors. */
    if topf <= topg {
        // inlining of macro cuddT
        // inlining of macro cuddE
        Fv = ((f).type.kids.T); Fnv = ((f).type.kids.E);
    } else {
        Fv = Fnv = f;
    }
    if topg <= topf {
        // inlining of macro cuddT
        // inlining of macro cuddE
        Gv = ((g).type.kids.T); Gnv = ((g).type.kids.E);
    } else {
        Gv = Gnv = g;
    }
    
    /* Recursive step. */
    if Fv != zero {
        t = Cudd_addEvalConst(dd,Fv,Gv);
        // inlining of macro cuddIsConstant
        if t == DD_NON_CONSTANT || !((t).index == CUDD_CONST_INDEX) {
            cuddCacheInsert2(dd, Cudd_addEvalConst, f, g, DD_NON_CONSTANT);
            return(DD_NON_CONSTANT);
        }
        if Fnv != zero {
            e = Cudd_addEvalConst(dd,Fnv,Gnv);
            // inlining of macro cuddIsConstant
            if e == DD_NON_CONSTANT || !((e).index == CUDD_CONST_INDEX) || t != e {
                cuddCacheInsert2(dd, Cudd_addEvalConst, f, g, DD_NON_CONSTANT);
                return(DD_NON_CONSTANT);
            }
        }
        cuddCacheInsert2(dd,Cudd_addEvalConst,f,g,t);
        return(t);
    } else { /* Fnv must be != zero */
        e = Cudd_addEvalConst(dd,Fnv,Gnv);
        cuddCacheInsert2(dd, Cudd_addEvalConst, f, g, e);
        return(e);
    }

} /* end of Cudd_addEvalConst */


/**Function********************************************************************

  Synopsis    [Computes the complement of an ADD a la C language.]

  Description [Computes the complement of an ADD a la C language: The
  complement of 0 is 1 and the complement of everything else is 0.
  Returns a pointer to the resulting ADD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addNegate]

******************************************************************************/
func Cudd_addCmpl(dd *DdManager, f *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddAddCmplRecur(dd,f);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_addCmpl */


/**Function********************************************************************

  Synopsis    [Determines whether f is less than or equal to g.]

  Description [Returns 1 if f is less than or equal to g; 0 otherwise.
  No new nodes are created. This procedure works for arbitrary ADDs.
  For 0-1 ADDs Cudd_addEvalConst is more efficient.]

  SideEffects [None]

  SeeAlso     [Cudd_addIteConstant Cudd_addEvalConst Cudd_bddLeq]

******************************************************************************/
func Cudd_addLeq(dd *DdManager, f *DdNode, g *DdNode) int32 {
    var tmp, fv, fvn, gv, gvn *DdNode;
    unsigned int topf, topg, res;

    /* Terminal cases. */
    if f == g { return(1); }

    statLine(dd);
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        // inlining of macro cuddIsConstant
        // inlining of macro cuddV
        if (g).index == CUDD_CONST_INDEX { return(((f).type.value) <= ((g).type.value)); }
        // inlining of macro DD_MINUS_INFINITY
        if f == ((dd).minusinfinity) { return(1); }
        // inlining of macro DD_PLUS_INFINITY
        if f == ((dd).plusinfinity) { return(0); } /* since f != g */
    }
    // inlining of macro DD_PLUS_INFINITY
    if g == ((dd).plusinfinity) { return(1); }
    // inlining of macro DD_MINUS_INFINITY
    if g == ((dd).minusinfinity) { return(0); } /* since f != g */

    /* Check cache. */
    tmp = cuddCacheLookup2(dd,(DD_CTFP)Cudd_addLeq,f,g);
    if tmp != nil {
        // inlining of macro DD_ONE
        return(tmp == ((dd).one));
    }

    /* Compute cofactors. One of f and g is not constant. */
    // inlining of macro cuddI
    
    if (f.index)==CUDD_CONST_INDEX {
        topf = int32((f.index));
    } else {
        topf = (dd).perm[(f.index)];
    }
    // inlining of macro cuddI
    
    if (g.index)==CUDD_CONST_INDEX {
        topg = int32((g.index));
    } else {
        topg = (dd).perm[(g.index)];
    }
    if topf <= topg {
        // inlining of macro cuddT
        // inlining of macro cuddE
        fv = ((f).type.kids.T); fvn = ((f).type.kids.E);
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

    res = Cudd_addLeq(dd,fvn,gvn) && Cudd_addLeq(dd,fv,gv);

    /* Store result in cache and return. */
    cuddCacheInsert2(dd,(DD_CTFP) Cudd_addLeq,f,g,
                     // inlining of macro Cudd_NotCond
                     // inlining of macro DD_ONE
                     ((DdNode *)((ptrint)(((dd).one)) ^ (res==0))));
    return(res);

} /* end of Cudd_addLeq */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Implements the recursive step of Cudd_addIte(f,g,h).]

  Description [Implements the recursive step of Cudd_addIte(f,g,h).
  Returns a pointer to the resulting ADD if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addIte]

******************************************************************************/
func cuddAddIteRecur(dd *DdManager, f *DdNode, g *DdNode, h *DdNode) *DdNode {
    var one, zero *DdNode;
    var r, Fv, Fnv, Gv, Gnv, Hv, Hnv, t, e *DdNode;
    unsigned int topf,topg,toph,v;
    var index int32  = -1;

    statLine(dd);
    /* Trivial cases. */

    /* One variable cases. */
    // inlining of macro DD_ONE
    if f == (one = ((dd).one)) {      /* ITE(1,G,H) = G */
        return(g);
    }
    // inlining of macro DD_ZERO
    if f == (zero = ((dd).zero)) {    /* ITE(0,G,H) = H */
        return(h);
    }

    /* From now on, f is known to not be a constant. */
    addVarToConst(f,&g,&h,one,zero);

    /* Check remaining one variable cases. */
    if g == h {                       /* ITE(F,G,G) = G */
        return(g);
    }

    if g == one {                     /* ITE(F,1,0) = F */
        if h == zero { return(f); }
    }

    // inlining of macro cuddI
    
    if (f.index)==CUDD_CONST_INDEX {
        topf = int32((f.index));
    } else {
        topf = (dd).perm[(f.index)];
    }
    // inlining of macro cuddI
    
    if (g.index)==CUDD_CONST_INDEX {
        topg = int32((g.index));
    } else {
        topg = (dd).perm[(g.index)];
    }
    // inlining of macro cuddI
    
    if (h.index)==CUDD_CONST_INDEX {
        toph = int32((h.index));
    } else {
        toph = (dd).perm[(h.index)];
    }
    // inlining of macro ddMin
    
    if (toph) < (topg) {
        v = (toph);
    } else {
        v = (topg);
    }

    /* A shortcut: ITE(F,G,H) = (x,G,H) if F=(x,1,0), x < top(G,H). */
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topf < v && ((f).type.kids.T) == one && ((f).type.kids.E) == zero {
        r = cuddUniqueInter(dd,int32(f.index),g,h);
        return(r);
    }
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topf < v && ((f).type.kids.T) == zero && ((f).type.kids.E) == one {
        r = cuddUniqueInter(dd,int32(f.index),h,g);
        return(r);
    }

    /* Check cache. */
    r = cuddCacheLookup(dd,DD_ADD_ITE_TAG,f,g,h);
    if r != nil {
        return(r);
    }

    /* Compute cofactors. */
    if topf <= v {
        // inlining of macro ddMin
        
        if (v) < (topf) {
            v = (v);
        } else {
            v = (topf);
        }      /* v = top_var(F,G,H) */
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
        index = h.index;
        // inlining of macro cuddT
        // inlining of macro cuddE
        Hv = ((h).type.kids.T); Hnv = ((h).type.kids.E);
    } else {
        Hv = Hnv = h;
    }
    
    /* Recursive step. */
    t = cuddAddIteRecur(dd,Fv,Gv,Hv);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = cuddAddIteRecur(dd,Fnv,Gnv,Hnv);
    if e == nil {
        Cudd_RecursiveDeref(dd,t);
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
        Cudd_RecursiveDeref(dd,t);
        Cudd_RecursiveDeref(dd,e);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);

    cuddCacheInsert(dd,DD_ADD_ITE_TAG,f,g,h,r);

    return(r);

} /* end of cuddAddIteRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_addCmpl.]

  Description [Performs the recursive step of Cudd_addCmpl. Returns a
  pointer to the resulting ADD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addCmpl]

******************************************************************************/
func cuddAddCmplRecur(dd *DdManager, f *DdNode) *DdNode {
    var one, zero *DdNode;
    var r, Fv, Fnv, t, e *DdNode;

    statLine(dd);
    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro DD_ZERO
    zero = ((dd).zero); 

    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        if f == zero {
            return(one);
        } else {
            return(zero);
        }
    }
    r = cuddCacheLookup1(dd,Cudd_addCmpl,f);
    if r != nil {
        return(r);
    }
    // inlining of macro cuddT
    Fv = ((f).type.kids.T);
    // inlining of macro cuddE
    Fnv = ((f).type.kids.E);
    t = cuddAddCmplRecur(dd,Fv);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);
    e = cuddAddCmplRecur(dd,Fnv);
    if e == nil {
        Cudd_RecursiveDeref(dd,t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);
    if t == e {
        r = t;
    } else {
        r = cuddUniqueInter(dd,int32(f.index),t,e);
    }
    if r == nil {
        Cudd_RecursiveDeref(dd, t);
        Cudd_RecursiveDeref(dd, e);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);
    cuddCacheInsert1(dd,Cudd_addCmpl,f,r);
    return(r);

} /* end of cuddAddCmplRecur */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Replaces variables with constants if possible (part of
  canonical form).]

  Description []

  SideEffects [None]

******************************************************************************/
func addVarToConst(f *DdNode, gp **DdNode, hp **DdNode, one *DdNode, zero *DdNode) {
    var g *DdNode  = *gp;
    var h *DdNode  = *hp;

    if f == g { /* ITE(F,F,H) = ITE(F,1,H) = F + H */
        *gp = one;
    }

    if f == h { /* ITE(F,G,F) = ITE(F,G,0) = F * G */
        *hp = zero;
    }

} /* end of addVarToConst */


//ABC_NAMESPACE_IMPL_END

