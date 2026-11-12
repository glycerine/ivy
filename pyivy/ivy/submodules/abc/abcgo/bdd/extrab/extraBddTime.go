package extrab

/**CFile****************************************************************

  FileName    [extraBddTime.c]

  PackageName [extra]

  Synopsis    [Procedures to control runtime in BDD operators.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - September 1, 2003.]

  Revision    [$Id: extraBddTime.c,v 1.0 2003/05/21 18:03:50 alanmi Exp $]

***********************************************************************/

//#include "extraBdd.h"

//ABC_NAMESPACE_IMPL_START


/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

const CHECK_FACTOR = 10

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/



/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static DdNode * cuddBddAndRecurTime( DdManager * manager, DdNode * f, DdNode * g, int * pRecCalls, int TimeOut );

// static DdNode * cuddBddAndAbstractRecurTime( DdManager * manager, DdNode * f, DdNode * g, DdNode * cube, int * pRecCalls, int TimeOut );

// static DdNode * extraTransferPermuteTime( DdManager * ddS, DdManager * ddD, DdNode * f, int * Permute, int TimeOut );

// static DdNode * extraTransferPermuteRecurTime( DdManager * ddS, DdManager * ddD, DdNode * f, st__table * table, int * Permute, int TimeOut );


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Computes the conjunction of two BDDs f and g.]

  Description [Computes the conjunction of two BDDs f and g. Returns a
  pointer to the resulting BDD if successful; NULL if the intermediate
  result blows up.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIte Cudd_addApply Cudd_bddAndAbstract Cudd_bddIntersect
  Cudd_bddOr Cudd_bddNand Cudd_bddNor Cudd_bddXor Cudd_bddXnor]

******************************************************************************/
func Extra_bddAndTime(dd *DdManager, f *DdNode, g *DdNode, TimeOut int32) *DdNode {
    var res *DdNode;
    var Counter int32;

    for {
    dd.reordered = 0;
    res = cuddBddAndRecurTime(dd,f,g, &Counter, TimeOut);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_bddAndTime */

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
func Extra_bddAndAbstractTime(manager *DdManager, f *DdNode, g *DdNode, cube *DdNode, TimeOut int32) *DdNode {
    var res *DdNode;
    var Counter int32;

    for {
    manager.reordered = 0;
    res = cuddBddAndAbstractRecurTime(manager, f, g, cube, &Counter, TimeOut);
    
if !(manager.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_bddAndAbstractTime */

/**Function********************************************************************

  Synopsis    [Convert a {A,B}DD from a manager to another with variable remapping.]

  Description [Convert a {A,B}DD from a manager to another one. The orders of the
  variables in the two managers may be different. Returns a
  pointer to the {A,B}DD in the destination manager if successful; NULL
  otherwise. The i-th entry in the array Permute tells what is the index
  of the i-th variable from the old manager in the new manager.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Extra_TransferPermuteTime(ddSource *DdManager, ddDestination *DdManager, f *DdNode, Permute *int32, TimeOut int32) *DdNode {
    var bRes *DdNode;
    for
    {
        ddDestination.reordered = 0;
        bRes = extraTransferPermuteTime( ddSource, ddDestination, f, Permute, TimeOut );
    
if !( ddDestination.reordered == 1 ) { break }
}
    
    return ( bRes );

} /* end of Extra_TransferPermuteTime */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis [Implements the recursive step of Cudd_bddAnd.]

  Description [Implements the recursive step of Cudd_bddAnd by taking
  the conjunction of two BDDs.  Returns a pointer to the result is
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddAnd]

******************************************************************************/
func cuddBddAndRecurTime(manager *DdManager, f *DdNode, g *DdNode, pRecCalls *int32, TimeOut int32) *DdNode {
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
    if f > g { /* Try to increase cache efficiency. */
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

//    if ( TimeOut && ((*pRecCalls)++ % CHECK_FACTOR) == 0 && TimeOut < Abc_Clock() )
    if (TimeOut != 0) && Abc_Clock() > TimeOut {
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

    t = cuddBddAndRecurTime(manager, fv, gv, pRecCalls, TimeOut);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = cuddBddAndRecurTime(manager, fnv, gnv, pRecCalls, TimeOut);
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

  Synopsis [Takes the AND of two BDDs and simultaneously abstracts the
  variables in cube.]

  Description [Takes the AND of two BDDs and simultaneously abstracts
  the variables in cube. The variables are existentially abstracted.
  Returns a pointer to the result is successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddAndAbstract]

******************************************************************************/
func cuddBddAndAbstractRecurTime(manager *DdManager, f *DdNode, g *DdNode, cube *DdNode, pRecCalls *int32, TimeOut int32) *DdNode {
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
    if f == one && g == one {    return(one); }

    if cube == one {
    return(cuddBddAndRecurTime(manager, f, g, pRecCalls, TimeOut));
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
        return(cuddBddAndRecurTime(manager, f, g, pRecCalls, TimeOut));
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

//    if ( TimeOut && ((*pRecCalls)++ % CHECK_FACTOR) == 0 && TimeOut < Abc_Clock() )
    if (TimeOut != 0) && Abc_Clock() > TimeOut {
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

    if topcube == top {    /* quantify */
    // inlining of macro cuddT
    var Cube *DdNode  = ((cube).type.kids.T);
    t = cuddBddAndAbstractRecurTime(manager, ft, gt, Cube, pRecCalls, TimeOut);
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
        e = cuddBddAndAbstractRecurTime(manager, fe, ge, Cube, pRecCalls, TimeOut);
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
        r = cuddBddAndRecurTime(manager, ((DdNode *)((ptrint)(t) ^ 01)), ((DdNode *)((ptrint)(e) ^ 01)), pRecCalls, TimeOut);
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
    t = cuddBddAndAbstractRecurTime(manager, ft, gt, cube, pRecCalls, TimeOut);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);
    e = cuddBddAndAbstractRecurTime(manager, fe, ge, cube, pRecCalls, TimeOut);
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

/**Function********************************************************************

  Synopsis    [Convert a BDD from a manager to another one.]

  Description [Convert a BDD from a manager to another one. Returns a
  pointer to the BDD in the destination manager if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Extra_TransferPermute]

******************************************************************************/
func extraTransferPermuteTime(ddS *DdManager, ddD *DdManager, f *DdNode, Permute *int32, TimeOut int32) *DdNode {
    var res *DdNode;
    var table *st__table  = nil;
    var gen *st__generator  = nil;
    var key, value *DdNode;

    table = st__init_table( st__ptrcmp, st__ptrhash );
    if table == nil {
        goto failure; }
    res = extraTransferPermuteRecurTime( ddS, ddD, f, table, Permute, TimeOut );
    if res != nil {
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref); }

    /* Dereference all elements in the table and dispose of the table.
       ** This must be done also if res is NULL to avoid leaks in case of
       ** reordering. */
    gen = st__init_gen( table );
    if gen == nil {
        goto failure; }
    for  st__gen( gen, ( char ** ) &key, ( char ** ) &value )  {
        Cudd_RecursiveDeref( ddD, value );
    }
    st__free_gen( gen );
    gen = nil;
    st__free_table( table );
    table = nil;

    if res != nil {
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref); }
    return ( res );

  failure:
    if table != nil {
        st__free_table( table ); }
    if gen != nil {
        st__free_gen( gen ); }
    return ( nil );

} /* end of extraTransferPermuteTime */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Extra_TransferPermute.]

  Description [Performs the recursive step of Extra_TransferPermute.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [extraTransferPermuteTime]

******************************************************************************/
func extraTransferPermuteRecurTime(ddS *DdManager, ddD *DdManager, f *DdNode, table *st__table, Permute *int32, TimeOut int32) *DdNode {
    var ft, fe, t, e, var, res *DdNode;
    var one, zero *DdNode;
    var index int32;
    var comple int32;

    statLine( ddD );
    // inlining of macro DD_ONE
    one = ((ddD).one);
    // inlining of macro Cudd_IsComplement
    comple = (int32( ((ptrint) (f) & 01)));

    /* Trivial cases. */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if  (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {
        // inlining of macro Cudd_NotCond
        return ( ((DdNode *)((ptrint)(one) ^ (comple))) ); }


    /* Make canonical to increase the utilization of the cache. */
    // inlining of macro Cudd_NotCond
    f = ((DdNode *)((ptrint)(f) ^ (comple)));
    /* Now f is a regular pointer to a non-constant node. */

    /* Check the cache. */
    if st__lookup( table, ( char * ) f, ( char ** ) &res ) {
        // inlining of macro Cudd_NotCond
        return ( ((DdNode *)((ptrint)(res) ^ (comple))) ); }

    if (TimeOut != 0) && Abc_Clock() > TimeOut {
        return nil; }

    /* Recursive step. */
    if Permute != nil {
        index = Permute[f.index];
    } else {
        index = f.index; }

    // inlining of macro cuddT
    ft = ((f).type.kids.T);
    // inlining of macro cuddE
    fe = ((f).type.kids.E);

    t = extraTransferPermuteRecurTime( ddS, ddD, ft, table, Permute, TimeOut );
    if t == nil {
        return ( nil );
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = extraTransferPermuteRecurTime( ddS, ddD, fe, table, Permute, TimeOut );
    if e == nil {
        Cudd_RecursiveDeref( ddD, t );
        return ( nil );
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(ddD.one) ^ 01));
    var = cuddUniqueInter( ddD, index, one, zero );
    if var == nil {
        Cudd_RecursiveDeref( ddD, t );
        Cudd_RecursiveDeref( ddD, e );
        return ( nil );
    }
    res = cuddBddIteRecur( ddD, var, t, e );

    if res == nil {
        Cudd_RecursiveDeref( ddD, t );
        Cudd_RecursiveDeref( ddD, e );
        return ( nil );
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref( ddD, t );
    Cudd_RecursiveDeref( ddD, e );

    if st__add_direct( table, ( char * ) f, ( char * ) res ) ==
         st__OUT_OF_MEM {
        Cudd_RecursiveDeref( ddD, res );
        return ( nil );
    }
    // inlining of macro Cudd_NotCond
    return ( ((DdNode *)((ptrint)(res) ^ (comple))) );

}  /* end of extraTransferPermuteRecurTime */

////////////////////////////////////////////////////////////////////////
///                           END OF FILE                            ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END

