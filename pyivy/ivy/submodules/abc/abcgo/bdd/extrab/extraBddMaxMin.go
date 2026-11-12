package extrab

/**CFile***********************************************************************

  FileName    [zMaxMin.c]

  PackageName [extra]

  Synopsis    [Experimental version of some ZDD operators.]

  Description [External procedures included in this module:
				<ul>
				<li> Extra_zddMaximal();
				<li> Extra_zddMinimal();
				<li> Extra_zddMaxUnion();
				<li> Extra_zddMinUnion();
				<li> Extra_zddDotProduct();
				<li> Extra_zddCrossProduct();
				<li> Extra_zddMaxDotProduct();
				</ul>
			Internal procedures included in this module:
				<ul>
				<li> extraZddMaximal();
				<li> extraZddMinimal();
				<li> extraZddMaxUnion();
				<li> extraZddMinUnion();
				<li> extraZddDotProduct();
				<li> extraZddCrossProduct();
				<li> extraZddMaxDotProduct();
				</ul>
			StaTc procedures included in this module:
				<ul>
				</ul>
				
          DotProduct and MaxDotProduct were introduced 
		  by O.Coudert to solve problems arising in two-level planar routing
          See O. Coudert, C.-J. R. Shi. Exact Multi-Layer Topological Planar 
		  Routing. Proc. of IEEE Custom Integrated Circuit Conference '96, 
		  pp. 179-182.
	      ]

  SeeAlso     []

  Author      [Alan Mishchenko]

  Copyright   []

  ReviSon    [$zMaxMin.c, v.1.2, November 26, 2000, alanmi $]

******************************************************************************/

//#include "extraBdd.h"

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

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaTcStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* StaTc Function prototypes                                                */
/*---------------------------------------------------------------------------*/

/**AutomaTcEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported Functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Computes the maximal of a set represented by its ZDD.]

  Description []

  SideEffects []

  SeeAlso     [Extra_zddMinimal]

******************************************************************************/
// Extra_zddMaximal is ported in extraBddSet.go (authoritative extrab/sets).


/**Function********************************************************************

  Synopsis    [Computes the minimal of a set represented by its ZDD.]

  Description []

  SideEffects []

  SeeAlso     [Extra_zddMaximal]

******************************************************************************/
func Extra_zddMinimal(dd *DdManager, S *DdNode) *DdNode {
    var res *DdNode;
    for {
	dd.reordered = 0;
	res = extraZddMinimal(dd, S);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddMinimal */


/**Function********************************************************************

  Synopsis    [Computes the maximal of the union of two sets represented by ZDDs.]

  Description []

  SideEffects []

  SeeAlso     [Extra_zddMaximal Extra_zddMimimal Extra_zddMinUnion]

******************************************************************************/
func Extra_zddMaxUnion(dd *DdManager, S *DdNode, T *DdNode) *DdNode {
    var res *DdNode;
    for {
	dd.reordered = 0;
	res = extraZddMaxUnion(dd, S, T);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddMaxUnion */


/**Function********************************************************************

  Synopsis    [Computes the minimal of the union of two sets represented by ZDDs.]

  Description []

  SideEffects []

  SeeAlso     [Extra_zddMaximal Extra_zddMimimal Extra_zddMaxUnion]

******************************************************************************/
func Extra_zddMinUnion(dd *DdManager, S *DdNode, T *DdNode) *DdNode {
    var res *DdNode;
    for {
	dd.reordered = 0;
	res = extraZddMinUnion(dd, S, T);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddMinUnion */


/**Function********************************************************************

  Synopsis    [Computes the dot product of two sets of subsets represented by ZDDs.]

  Description [The dot product is defined as a set of pair-wise unions of subsets 
  belonging to the arguments.]

  SideEffects []

  SeeAlso     [Extra_zddCrossProduct]

******************************************************************************/
func Extra_zddDotProduct(dd *DdManager, S *DdNode, T *DdNode) *DdNode {
    var res *DdNode;
    for {
	dd.reordered = 0;
	res = extraZddDotProduct(dd, S, T);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddDotProduct */


/**Function********************************************************************

  Synopsis    [Computes the cross product of two sets of subsets represented by ZDDs.]

  Description [The cross product is defined as a set of pair-wise intersections of subsets 
  belonging to the arguments.]

  SideEffects []

  SeeAlso     [Extra_zddDotProduct]

******************************************************************************/
func Extra_zddCrossProduct(dd *DdManager, S *DdNode, T *DdNode) *DdNode {
    var res *DdNode;
    for {
	dd.reordered = 0;
	res = extraZddCrossProduct(dd, S, T);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddCrossProduct */


/**Function********************************************************************

  Synopsis    [Computes the maximal of the DotProduct of the union of two sets represented by ZDDs.]

  Description []

  SideEffects []

  SeeAlso     [Extra_zddDotProduct Extra_zddMaximal Extra_zddMinCrossProduct]

******************************************************************************/
func Extra_zddMaxDotProduct(dd *DdManager, S *DdNode, T *DdNode) *DdNode {
    var res *DdNode;
    for {
	dd.reordered = 0;
	res = extraZddMaxDotProduct(dd, S, T);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddMaxDotProduct */


/*---------------------------------------------------------------------------*/
/* Definition of internal Functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddMaximal.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
// extraZddMaximal is ported in extraBddSet.go (authoritative extrab/sets).



/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddMinimal.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func extraZddMinimal(dd *DdManager, zSet *DdNode) *DdNode {
	var zRes *DdNode;
    statLine(dd); 

	/* consider terminal cases */
	if zSet == z0 {
		return zSet; }
	/* the empty combinaTon, if present, is the only minimal combinaTon */
	if Extra_zddEmptyBelongs(dd, zSet) {
		return z1; } 

    /* check cache */
	zRes = cuddCacheLookup1Zdd(dd, extraZddMinimal, zSet);
	if zRes != nil {
		return(zRes);
	} else {
		var zSet0, zSet1, zRes0, zRes1 *DdNode;

		/* compute minimal for subsets without the top-most element */
		// inlining of macro cuddE
		zSet0 = extraZddMinimal(dd, ((zSet).type.kids.E));
		if zSet0 == nil {
			return nil; }
		// inlining of macro cuddRef
		cuddSatInc(Cudd_Regular(zSet0).ref);

		/* compute minimal for subsets with the top-most element */
		// inlining of macro cuddT
		zSet1 = extraZddMinimal(dd, ((zSet).type.kids.T));
		if zSet1 == nil {
			Cudd_RecursiveDerefZdd(dd, zSet0);
			return nil;
		}
		// inlining of macro cuddRef
		cuddSatInc(Cudd_Regular(zSet1).ref);

		/* subset without this element remains unchanged */
		zRes0 = zSet0;

		/* remove subsets with this element that contain subsets without this element */
		zRes1 = extraZddNotSupSet(dd, zSet1, zSet0);
		if zRes1 == nil {
			Cudd_RecursiveDerefZdd(dd, zSet0);
			Cudd_RecursiveDerefZdd(dd, zSet1);
			return nil;
		}
		// inlining of macro cuddRef
		cuddSatInc(Cudd_Regular(zRes1).ref);
		Cudd_RecursiveDerefZdd(dd, zSet1);

		/* create the new node */
		zRes = cuddZddGetNode( dd, zSet.index, zRes1, zRes0 );
		if zRes == nil {
			Cudd_RecursiveDerefZdd( dd, zRes0 );
			Cudd_RecursiveDerefZdd( dd, zRes1 );
			return nil;
		}
		// inlining of macro cuddDeref
		cuddSatDec(Cudd_Regular(zRes0).ref);
		// inlining of macro cuddDeref
		cuddSatDec(Cudd_Regular(zRes1).ref);

		/* insert the result into cache */
		cuddCacheInsert1(dd, extraZddMinimal, zSet, zRes);
		return zRes;
	}
} /* end of extraZddMinimal */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddMaxUnion.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func extraZddMaxUnion(dd *DdManager, S *DdNode, T *DdNode) *DdNode {
	var zRes *DdNode;
	var TopS, TopT int32;
    statLine(dd); 

	/* consider terminal cases */
	if S == z0 {
		return T; }
	if T == z0 {
		return S; }
	if S == T {
		return S; }
	if S == z1 {
		return T; }
	if T == z1 {
		return S; }

	/* the operation is commutative - normalize the problem */
	TopS = dd.permZ[S.index];
	TopT = dd.permZ[T.index];

	if TopS > TopT || (TopS == TopT && S > T) {
		return extraZddMaxUnion(dd, T, S); }

    /* check cache */
	zRes = cuddCacheLookup2Zdd(dd, extraZddMaxUnion, S, T);
	if zRes != nil {
		return zRes;
	} else {
		var zSet0, zSet1, zRes0, zRes1 *DdNode;

		if TopS == TopT {
			/* compute maximal for subsets without the top-most element */
			// inlining of macro cuddE
			zSet0 = extraZddMaxUnion(dd, ((S).type.kids.E), ((T).type.kids.E) );
			if zSet0 == nil {
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet0).ref);

			/* compute maximal for subsets with the top-most element */
			// inlining of macro cuddT
			zSet1 = extraZddMaxUnion(dd, ((S).type.kids.T), ((T).type.kids.T) );
			if zSet1 == nil {
				Cudd_RecursiveDerefZdd(dd, zSet0);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet1).ref);
		
		} else { /* if ( TopS < TopT ) */
		
			/* compute maximal for subsets without the top-most element */
			// inlining of macro cuddE
			zSet0 = extraZddMaxUnion(dd, ((S).type.kids.E), T );
			if zSet0 == nil {
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet0).ref);

			/* subset with this element is just the cofactor of S */
			// inlining of macro cuddT
			zSet1 = ((S).type.kids.T);
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet1).ref);
		}

		/* remove subsets without this element covered by subsets with this element */
		zRes0 = extraZddNotSubSet(dd, zSet0, zSet1);
		if zRes0 == nil {
			Cudd_RecursiveDerefZdd(dd, zSet0);
			Cudd_RecursiveDerefZdd(dd, zSet1);
			return nil;
		}
		// inlining of macro cuddRef
		cuddSatInc(Cudd_Regular(zRes0).ref);
		Cudd_RecursiveDerefZdd(dd, zSet0);

		/* subset with this element remains unchanged */
		zRes1 = zSet1;

		/* create the new node */
		zRes = cuddZddGetNode( dd, S.index, zRes1, zRes0 );
		if zRes == nil {
			Cudd_RecursiveDerefZdd( dd, zRes0 );
			Cudd_RecursiveDerefZdd( dd, zRes1 );
			return nil;
		}
		// inlining of macro cuddDeref
		cuddSatDec(Cudd_Regular(zRes0).ref);
		// inlining of macro cuddDeref
		cuddSatDec(Cudd_Regular(zRes1).ref);

		/* insert the result into cache */
		cuddCacheInsert2(dd, extraZddMaxUnion, S, T, zRes);
		return zRes;
	}
} /* end of extraZddMaxUnion */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddMinUnion.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func extraZddMinUnion(dd *DdManager, S *DdNode, T *DdNode) *DdNode {
	var zRes *DdNode;
	var TopS, TopT int32;
    statLine(dd); 

	/* consider terminal cases */
	if S == z0 {
		return T; }
	if T == z0 {
		return S; }
	if S == T {
		return S; }
	/* the empty combination, if present, is the only minimal combination */
	if Extra_zddEmptyBelongs(dd, S) || Extra_zddEmptyBelongs(dd, T) {
		return z1; } 

	/* the operation is commutative - normalize the problem */
	TopS = dd.permZ[S.index];
	TopT = dd.permZ[T.index];

	if TopS > TopT || (TopS == TopT && S > T) {
		return extraZddMinUnion(dd, T, S); }

    /* check cache */
	zRes = cuddCacheLookup2Zdd(dd, extraZddMinUnion, S, T);
	if zRes != nil {
		return(zRes);
	} else {
		var zSet0, zSet1, zRes0, zRes1 *DdNode;
		if TopS == TopT {
			/* compute maximal for subsets without the top-most element */
			// inlining of macro cuddE
			zSet0 = extraZddMinUnion(dd, ((S).type.kids.E), ((T).type.kids.E) );
			if zSet0 == nil {
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet0).ref);

			/* compute maximal for subsets with the top-most element */
			// inlining of macro cuddT
			zSet1 = extraZddMinUnion(dd, ((S).type.kids.T), ((T).type.kids.T) );
			if zSet1 == nil {
				Cudd_RecursiveDerefZdd(dd, zSet0);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet1).ref);
		
		} else { /* if ( TopS < TopT ) */
		
			/* compute maximal for subsets without the top-most element */
			// inlining of macro cuddE
			zSet0 = extraZddMinUnion(dd, ((S).type.kids.E), T );
			if zSet0 == nil {
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet0).ref);

			/* subset with this element is just the cofactor of S */
			// inlining of macro cuddT
			zSet1 = ((S).type.kids.T);
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet1).ref);
		}

		/* subset without this element remains unchanged */
		zRes0 = zSet0;

		/* remove subsets with this element that contain subsets without this element */
		zRes1 = extraZddNotSupSet(dd, zSet1, zSet0);
		if zRes1 == nil {
			Cudd_RecursiveDerefZdd(dd, zSet0);
			Cudd_RecursiveDerefZdd(dd, zSet1);
			return nil;
		}
		// inlining of macro cuddRef
		cuddSatInc(Cudd_Regular(zRes1).ref);
		Cudd_RecursiveDerefZdd(dd, zSet1);

		/* create the new node */
		zRes = cuddZddGetNode( dd, S.index, zRes1, zRes0 );
		if zRes == nil {
			Cudd_RecursiveDerefZdd( dd, zRes0 );
			Cudd_RecursiveDerefZdd( dd, zRes1 );
			return nil;
		}
		// inlining of macro cuddDeref
		cuddSatDec(Cudd_Regular(zRes0).ref);
		// inlining of macro cuddDeref
		cuddSatDec(Cudd_Regular(zRes1).ref);

		/* insert the result into cache */
		cuddCacheInsert2(dd, extraZddMinUnion, S, T, zRes);
		return zRes;
	}
} /* end of extraZddMinUnion */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddDotProduct.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func extraZddDotProduct(dd *DdManager, S *DdNode, T *DdNode) *DdNode {
	var zRes *DdNode;
	var TopS, TopT int32;
    statLine(dd); 

	/* consider terminal cases */
	if S == z0 || T == z0 {
		return z0; }
	if S == z1 {
		return T; }
	if T == z1 {
		return S; }

	/* the operation is commutative - normalize the problem */
	TopS = dd.permZ[S.index];
	TopT = dd.permZ[T.index];

	if TopS > TopT || (TopS == TopT && S > T) {
		return extraZddDotProduct(dd, T, S); }

    /* check cache */
	zRes = cuddCacheLookup2Zdd(dd, extraZddDotProduct, S, T);
	if zRes != nil {
		return zRes;
	} else {
		var zSet0, zSet1, zRes0, zRes1, zTemp *DdNode;
		if TopS == TopT {
			/* compute the union of two cofactors of T (T0+T1) */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zTemp = cuddZddUnion(dd, ((T).type.kids.E), ((T).type.kids.T) );
			if zTemp == nil {
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zTemp).ref);

			/* compute DotProduct with the top element for subsets (S1, T0+T1) */
			// inlining of macro cuddT
			zSet0 = extraZddDotProduct(dd, ((S).type.kids.T), zTemp );
			if zSet0 == nil {
				Cudd_RecursiveDerefZdd(dd, zTemp);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet0).ref);
			Cudd_RecursiveDerefZdd(dd, zTemp);

			/* compute DotProduct with the top element for subsets (S0, T1) */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zSet1 = extraZddDotProduct(dd, ((S).type.kids.E), ((T).type.kids.T) );
			if zSet1 == nil {
				Cudd_RecursiveDerefZdd(dd, zSet0);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet1).ref);

			/* compute the union of these two partial results (zSet0 + zSet1) */
			zRes1 = cuddZddUnion(dd, zSet0, zSet1 );
			if zRes1 == nil {
				Cudd_RecursiveDerefZdd(dd, zSet0);
				Cudd_RecursiveDerefZdd(dd, zSet1);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);
			Cudd_RecursiveDerefZdd(dd, zSet0);
			Cudd_RecursiveDerefZdd(dd, zSet1);

			/* compute DotProduct for subsets without the top-most element */
			// inlining of macro cuddE
			zRes0 = extraZddDotProduct(dd, ((S).type.kids.E), ((T).type.kids.E) );
			if zRes0 == nil {
				Cudd_RecursiveDerefZdd(dd, zRes1);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);
		
		} else { /* if ( TopS < TopT ) */
		
			/* compute DotProduct with the top element for subsets (S1, T) */
			// inlining of macro cuddT
			zRes1 = extraZddDotProduct(dd, ((S).type.kids.T), T );
			if zRes1 == nil {
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);

			/* compute DotProduct for subsets without the top-most element */
			// inlining of macro cuddE
			zRes0 = extraZddDotProduct(dd, ((S).type.kids.E), T );
			if zRes0 == nil {
				Cudd_RecursiveDerefZdd(dd, zRes1);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);
		}

		/* create the new node */
		zRes = cuddZddGetNode( dd, S.index, zRes1, zRes0 );
		if zRes == nil {
			Cudd_RecursiveDerefZdd( dd, zRes0 );
			Cudd_RecursiveDerefZdd( dd, zRes1 );
			return nil;
		}
		// inlining of macro cuddDeref
		cuddSatDec(Cudd_Regular(zRes0).ref);
		// inlining of macro cuddDeref
		cuddSatDec(Cudd_Regular(zRes1).ref);

		/* insert the result into cache */
		cuddCacheInsert2(dd, extraZddDotProduct, S, T, zRes);
		return zRes;
	}
} /* end of extraZddDotProduct */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddCrossProduct.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func extraZddCrossProduct(dd *DdManager, S *DdNode, T *DdNode) *DdNode {
	var zRes *DdNode;
	var TopS, TopT int32;
    statLine(dd); 

	/* consider terminal cases */
	if S == z0 || T == z0 {
		return z0; }
	if S == z1 || T == z1 {
		return z1; }

	/* the operation is commutative - normalize the problem */
	TopS = dd.permZ[S.index];
	TopT = dd.permZ[T.index];

	if TopS > TopT || (TopS == TopT && S > T) {
		return extraZddCrossProduct(dd, T, S); }

    /* check cache */
	zRes = cuddCacheLookup2Zdd(dd, extraZddCrossProduct, S, T);
	if zRes != nil {
		return zRes;
	} else {
		var zSet0, zSet1, zRes0, zRes1, zTemp *DdNode;

		if TopS == TopT {
			/* compute the union of two cofactors of T (T0+T1) */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zTemp = cuddZddUnion(dd, ((T).type.kids.E), ((T).type.kids.T) );
			if zTemp == nil {
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zTemp).ref);

			/* compute CrossProduct without the top element for subsets (S0, T0+T1) */
			// inlining of macro cuddE
			zSet0 = extraZddCrossProduct(dd, ((S).type.kids.E), zTemp );
			if zSet0 == nil {
				Cudd_RecursiveDerefZdd(dd, zTemp);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet0).ref);
			Cudd_RecursiveDerefZdd(dd, zTemp);

			/* compute CrossProduct without the top element for subsets (S1, T0) */
			// inlining of macro cuddT
			// inlining of macro cuddE
			zSet1 = extraZddCrossProduct(dd, ((S).type.kids.T), ((T).type.kids.E) );
			if zSet1 == nil {
				Cudd_RecursiveDerefZdd(dd, zSet0);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet1).ref);

			/* compute the union of these two partial results (zSet0 + zSet1) */
			zRes0 = cuddZddUnion(dd, zSet0, zSet1 );
			if zRes0 == nil {
				Cudd_RecursiveDerefZdd(dd, zSet0);
				Cudd_RecursiveDerefZdd(dd, zSet1);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);
			Cudd_RecursiveDerefZdd(dd, zSet0);
			Cudd_RecursiveDerefZdd(dd, zSet1);

			/* compute CrossProduct for subsets with the top-most element */
			// inlining of macro cuddT
			zRes1 = extraZddCrossProduct(dd, ((S).type.kids.T), ((T).type.kids.T) );
			if zRes1 == nil {
				Cudd_RecursiveDerefZdd(dd, zRes0);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);

			/* create the new node */
			zRes = cuddZddGetNode( dd, S.index, zRes1, zRes0 );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zRes1 );
				return nil;
			}
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes0).ref);
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes1).ref);
		
		} else { /* if ( TopS < TopT ) */
		
			/* compute CrossProduct without the top element (S0, T) */
			// inlining of macro cuddE
			zSet0 = extraZddCrossProduct(dd, ((S).type.kids.E), T );
			if zSet0 == nil {
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet0).ref);

			/* compute CrossProduct without the top element (S1, T) */
			// inlining of macro cuddT
			zSet1 = extraZddCrossProduct(dd, ((S).type.kids.T), T );
			if zSet1 == nil {
				Cudd_RecursiveDerefZdd(dd, zSet0);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet1).ref);

			/* compute the union of these two partial results (zSet0 + zSet1) */
			zRes = cuddZddUnion(dd, zSet0, zSet1 );
			if zRes == nil {
				Cudd_RecursiveDerefZdd(dd, zSet0);
				Cudd_RecursiveDerefZdd(dd, zSet1);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes).ref);
			Cudd_RecursiveDerefZdd(dd, zSet0);
			Cudd_RecursiveDerefZdd(dd, zSet1);
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes).ref);
		}

		/* insert the result into cache */
		cuddCacheInsert2(dd, extraZddCrossProduct, S, T, zRes);
		return zRes;
	}
} /* end of extraZddCrossProduct */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddMaxDotProduct.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func extraZddMaxDotProduct(dd *DdManager, S *DdNode, T *DdNode) *DdNode {
	var zRes *DdNode;
	var TopS, TopT int32;
    statLine(dd); 

	/* consider terminal cases */
	if S == z0 || T == z0 {
		return z0; }
	if S == z1 {
		return T; }
	if T == z1 {
		return S; }

	/* the operation is commutative - normalize the problem */
	TopS = dd.permZ[S.index];
	TopT = dd.permZ[T.index];

	if TopS > TopT || (TopS == TopT && S > T) {
		return extraZddMaxDotProduct(dd, T, S); }

    /* check cache */
	zRes = cuddCacheLookup2Zdd(dd, extraZddMaxDotProduct, S, T);
	if zRes != nil {
		return zRes;
	} else {
		var zSet0, zSet1, zRes0, zRes1, zTemp *DdNode;
		if TopS == TopT {
			/* compute the union of two cofactors of T (T0+T1) */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zTemp = extraZddMaxUnion(dd, ((T).type.kids.E), ((T).type.kids.T) );
			if zTemp == nil {
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zTemp).ref);

			/* compute MaxDotProduct with the top element for subsets (S1, T0+T1) */
			// inlining of macro cuddT
			zSet0 = extraZddMaxDotProduct(dd, ((S).type.kids.T), zTemp );
			if zSet0 == nil {
				Cudd_RecursiveDerefZdd(dd, zTemp);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet0).ref);
			Cudd_RecursiveDerefZdd(dd, zTemp);

			/* compute MaxDotProduct with the top element for subsets (S0, T1) */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zSet1 = extraZddMaxDotProduct(dd, ((S).type.kids.E), ((T).type.kids.T) );
			if zSet1 == nil {
				Cudd_RecursiveDerefZdd(dd, zSet0);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zSet1).ref);

			/* compute the union of these two partial results (zSet0 + zSet1) */
			zRes1 = extraZddMaxUnion(dd, zSet0, zSet1 );
			if zRes1 == nil {
				Cudd_RecursiveDerefZdd(dd, zSet0);
				Cudd_RecursiveDerefZdd(dd, zSet1);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);
			Cudd_RecursiveDerefZdd(dd, zSet0);
			Cudd_RecursiveDerefZdd(dd, zSet1);

			/* compute MaxDotProduct for subsets without the top-most element */
			// inlining of macro cuddE
			zRes0 = extraZddMaxDotProduct(dd, ((S).type.kids.E), ((T).type.kids.E) );
			if zRes0 == nil {
				Cudd_RecursiveDerefZdd(dd, zRes1);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);
		
		} else { /* if ( TopS < TopT ) */
		
			/* compute MaxDotProduct with the top element for subsets (S1, T) */
			// inlining of macro cuddT
			zRes1 = extraZddMaxDotProduct(dd, ((S).type.kids.T), T );
			if zRes1 == nil {
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);

			/* compute MaxDotProduct for subsets without the top-most element */
			// inlining of macro cuddE
			zRes0 = extraZddMaxDotProduct(dd, ((S).type.kids.E), T );
			if zRes0 == nil {
				Cudd_RecursiveDerefZdd(dd, zRes1);
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);
		}

		/* remove subsets without this element covered by subsets with this element */
		zRes0 = extraZddNotSubSet(dd, zTemp = zRes0, zRes1);
		if zRes0 == nil {
			Cudd_RecursiveDerefZdd(dd, zTemp);
			Cudd_RecursiveDerefZdd(dd, zRes1);
			return nil;
		}
		// inlining of macro cuddRef
		cuddSatInc(Cudd_Regular(zRes0).ref);
		Cudd_RecursiveDerefZdd(dd, zTemp);

		/* create the new node */
		zRes = cuddZddGetNode( dd, S.index, zRes1, zRes0 );
		if zRes == nil {
			Cudd_RecursiveDerefZdd( dd, zRes0 );
			Cudd_RecursiveDerefZdd( dd, zRes1 );
			return nil;
		}
		// inlining of macro cuddDeref
		cuddSatDec(Cudd_Regular(zRes0).ref);
		// inlining of macro cuddDeref
		cuddSatDec(Cudd_Regular(zRes1).ref);

		/* insert the result into cache */
		cuddCacheInsert2(dd, extraZddMaxDotProduct, S, T, zRes);
		return zRes;
	}
} /* end of extraZddMaxDotProduct */

/*---------------------------------------------------------------------------*/
/* Definition of staTc Functions                                            */
/*---------------------------------------------------------------------------*/

//ABC_NAMESPACE_IMPL_END
