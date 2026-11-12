package extrab

/**CFile***********************************************************************

  FileName    [zSubSet.c]

  PackageName [extra]

  Synopsis    [Experimental version of some ZDD operators.]

  Description [External procedures included in this module:
		    <ul>
		    <li> Extra_zddSubSet();
		    <li> Extra_zddSupSet();
		    <li> Extra_zddNotSubSet();
		    <li> Extra_zddNotSupSet();
		    <li> Extra_zddMaxNotSupSet();
			<li> Extra_zddEmptyBelongs();
			<li> Extra_zddIsOneSubset();
		    </ul>
	       Internal procedures included in this module:
		    <ul>
		    <li> extraZddSubSet();
		    <li> extraZddSupSet();
		    <li> extraZddNotSubSet();
		    <li> extraZddNotSupSet();
		    <li> extraZddMaxNotSupSet();
		    </ul>
	       Static procedures included in this module:
		    <ul>
		    </ul>

          SubSet, SupSet, NotSubSet, NotSupSet were introduced 
		  by O.Coudert to solve problems arising in two-level SOP 
		  minimization. See O. Coudert, "Two-Level Logic Minimization: 
		  An Overview", Integration. Vol. 17, No. 2, pp. 97-140, Oct 1994.
	      ]

  SeeAlso     []

  Author      [Alan Mishchenko]

  Copyright   []

  Revision    [$zSubSet.c, v.1.2, November 16, 2000, alanmi $]

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


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Computes subsets in X that are contained in some of the subsets of Y.]

  Description []

  SideEffects []

  SeeAlso     [Extra_zddNotSubSet, Extra_zddSupSet, Extra_zddNotSupSet]

******************************************************************************/
func Extra_zddSubSet(dd *DdManager, X *DdNode, Y *DdNode) *DdNode {
    var res *DdNode;
    var autoDynZ int32;

    autoDynZ = dd.autoDynZ;
    dd.autoDynZ = 0;

    for {
	dd.reordered = 0;
	res = extraZddSubSet(dd, X, Y);
    
if !(dd.reordered == 1) { break }
} 
    dd.autoDynZ = autoDynZ;
    return(res);

} /* end of Extra_zddSubSet */


/**Function********************************************************************

  Synopsis    [Computes subsets in X that contain some of the subsets of Y.]

  Description []

  SideEffects []

  SeeAlso     [Extra_zddSubSet, Extra_zddNotSubSet, Extra_zddNotSupSet]

******************************************************************************/
func Extra_zddSupSet(dd *DdManager, X *DdNode, Y *DdNode) *DdNode {
    var res *DdNode;
    var autoDynZ int32;

    autoDynZ = dd.autoDynZ;
    dd.autoDynZ = 0;

    for {
	dd.reordered = 0;
	res = extraZddSupSet(dd, X, Y);
    
if !(dd.reordered == 1) { break }
} 
    dd.autoDynZ = autoDynZ;
    return(res);

} /* end of Extra_zddSupSet */

/**Function********************************************************************

  Synopsis    [Computes subsets in X that are not contained in any of the subsets of Y.]

  Description []

  SideEffects []

  SeeAlso     [Extra_zddSubSet, Extra_zddSupSet, Extra_zddNotSupSet]

******************************************************************************/
func Extra_zddNotSubSet(dd *DdManager, X *DdNode, Y *DdNode) *DdNode {
    var res *DdNode;
    var autoDynZ int32;

    autoDynZ = dd.autoDynZ;
    dd.autoDynZ = 0;

    for {
	dd.reordered = 0;
	res = extraZddNotSubSet(dd, X, Y);
    
if !(dd.reordered == 1) { break }
} 
    dd.autoDynZ = autoDynZ;
    return(res);

} /* end of Extra_zddNotSubSet */


/**Function********************************************************************

  Synopsis    [Computes subsets in X that do not contain any of the subsets of Y.]

  Description []

  SideEffects []

  SeeAlso     [Extra_zddSubSet, Extra_zddSupSet, Extra_zddNotSubSet]

******************************************************************************/
func Extra_zddNotSupSet(dd *DdManager, X *DdNode, Y *DdNode) *DdNode {
    var res *DdNode;
    var autoDynZ int32;

    autoDynZ = dd.autoDynZ;
    dd.autoDynZ = 0;

    for {
	dd.reordered = 0;
	res = extraZddNotSupSet(dd, X, Y);
    
if !(dd.reordered == 1) { break }
} 
    dd.autoDynZ = autoDynZ;
    return(res);

} /* end of Extra_zddNotSupSet */



/**Function********************************************************************

  Synopsis    [Computes the maximal of subsets in X not contained in any of the subsets of Y.]

  Description []

  SideEffects []

  SeeAlso     [Extra_zddSubSet, Extra_zddSupSet, Extra_zddSubSet, Extra_zddNotSupSet]

******************************************************************************/
func Extra_zddMaxNotSupSet(dd *DdManager, X *DdNode, Y *DdNode) *DdNode {
    var res *DdNode;
    var autoDynZ int32;

    autoDynZ = dd.autoDynZ;
    dd.autoDynZ = 0;

    for {
	dd.reordered = 0;
	res = extraZddMaxNotSupSet(dd, X, Y);
    
if !(dd.reordered == 1) { break }
} 
    dd.autoDynZ = autoDynZ;
    return(res);

} /* end of Extra_zddMaxNotSupSet */


/**Function********************************************************************

  Synopsis    [Returns 1 if ZDD contains the empty combination; 0 otherwise.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_zddEmptyBelongs(dd *DdManager, zS *DdNode) int32 {
	for  zS.index != CUDD_MAXINDEX  {
		// inlining of macro cuddE
		zS = ((zS).type.kids.E);
	}
	return int32(( zS == z1 ));

} /* end of Extra_zddEmptyBelongs */


/**Function********************************************************************

  Synopsis    [Returns 1 if the set is empty or consists of one subset only.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_zddIsOneSubset(dd *DdManager, zS *DdNode) int32 {
	for  zS.index != CUDD_MAXINDEX  {
		// inlining of macro cuddT
		assert(((zS).type.kids.T) != z0, "/*__starting_point_inline_macro:cuddT__*/((zS).type.kids.T) != z0");
		// inlining of macro cuddE
		if  ((zS).type.kids.E) != z0 {
			return 0; }
		// inlining of macro cuddT
		zS = ((zS).type.kids.T);
	}
	return int32(( zS == z1 ));

} /* end of Extra_zddEmptyBelongs */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddSubSet.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func extraZddSubSet(dd *DdManager, X *DdNode, Y *DdNode) *DdNode {	
	var zRes *DdNode;
    statLine(dd); 
	/* any comb is a subset of itself */
	if X == Y { 
		return X; }
	/* if X is empty, the result is empty */
	if X == z0 { 
		return z0; }
	/* combs in X are notsubsets of non-existant combs in Y */
	if Y == z0 { 
		return z0; }
	/* the empty comb is contained in all combs of Y */
	if X == z1 { 
		return z1; }
	/* only {()} is the subset of {()} */
	if Y == z1 { /* check whether the empty combination is present in X */ 
		if Extra_zddEmptyBelongs( dd, X ) != 0 {
			return z1;
		} else {
			return z0;
		} }

    /* check cache */
    zRes = cuddCacheLookup2Zdd(dd, extraZddSubSet, X, Y);
    if zRes != nil {
    	return(zRes);
	} else {
		var zRes0, zRes1, zTemp *DdNode;
		var TopLevelX int32  = dd.permZ[ X.index ];
		var TopLevelY int32  = dd.permZ[ Y.index ];

		if TopLevelX < TopLevelY {
			/* compute combs of X without var that are notsubsets of combs with Y */
			// inlining of macro cuddE
			zRes = extraZddSubSet( dd, ((X).type.kids.E), Y );
			if zRes == nil {  return nil; }
		
		} else if TopLevelX == TopLevelY {
			/* merge combs of Y with and without var */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zTemp = cuddZddUnion( dd, ((Y).type.kids.E), ((Y).type.kids.T) );
			if zTemp == nil {
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zTemp).ref);

			/* compute combs of X without var that are notsubsets of combs is Temp */
			// inlining of macro cuddE
			zRes0 = extraZddSubSet( dd, ((X).type.kids.E), zTemp );
			if zRes0 == nil {
				Cudd_RecursiveDerefZdd( dd, zTemp );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);
			Cudd_RecursiveDerefZdd( dd, zTemp );

			/* combs of X with var that are notsubsets of combs in Y with var */
			// inlining of macro cuddT
			zRes1 = extraZddSubSet( dd, ((X).type.kids.T), ((Y).type.kids.T) );
			if zRes1 == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);

			/* compose Res0 and Res1 with the given ZDD variable */
			zRes = cuddZddGetNode( dd, X.index, zRes1, zRes0 );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zRes1 );
				return nil;
			}
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes0).ref);
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes1).ref);
		
		} else { /* if ( TopLevelX > TopLevelY ) */
		
			/* merge combs of Y with and without var */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zTemp = cuddZddUnion( dd, ((Y).type.kids.E), ((Y).type.kids.T) );
			if zTemp == nil {  return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zTemp).ref);

			/* compute combs that are notsubsets of Temp */
			zRes = extraZddSubSet( dd, X, zTemp );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zTemp );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes).ref);
			Cudd_RecursiveDerefZdd( dd, zTemp );
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes).ref);
		}

		/* insert the result into cache */
	    cuddCacheInsert2(dd, extraZddSubSet, X, Y, zRes);
		return zRes;
	}
} /* end of extraZddSubSet */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddSupSet.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func extraZddSupSet(dd *DdManager, X *DdNode, Y *DdNode) *DdNode {	
	var zRes *DdNode;
    statLine(dd); 
	/* any comb is a superset of itself */
	if X == Y { 
		return X; }
	/* no comb in X is superset of non-existing combs */
	if Y == z0 { 
		return z0; }	
	/* any comb in X is the superset of the empty comb */
	if Extra_zddEmptyBelongs( dd, Y ) != 0 { 
		return X; }
	/* if X is empty, the result is empty */
	if X == z0 { 
		return z0; }
	/* if X is the empty comb (and Y does not contain it!), return empty */
	if X == z1 { 
		return z0; }

    /* check cache */
    zRes = cuddCacheLookup2Zdd(dd, extraZddSupSet, X, Y);
    if zRes != nil {
    	return(zRes);
	} else {
		var zRes0, zRes1, zTemp *DdNode;
		var TopLevelX int32  = dd.permZ[ X.index ];
		var TopLevelY int32  = dd.permZ[ Y.index ];

		if TopLevelX < TopLevelY {
			/* combinations of X without label that are supersets of combinations with Y */
			// inlining of macro cuddE
			zRes0 = extraZddSupSet( dd, ((X).type.kids.E), Y );
			if zRes0 == nil {  return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);

			/* combinations of X with label that are supersets of combinations with Y */
			// inlining of macro cuddT
			zRes1 = extraZddSupSet( dd, ((X).type.kids.T), Y );
			if zRes1 == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);

			/* compose Res0 and Res1 with the given ZDD variable */
			zRes = cuddZddGetNode( dd, X.index, zRes1, zRes0 );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zRes1 );
				return nil;
			}
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes0).ref);
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes1).ref);
		
		} else if TopLevelX == TopLevelY {
			/* combs of X without var that are supersets of combs of Y without var */
			// inlining of macro cuddE
			zRes0 = extraZddSupSet( dd, ((X).type.kids.E), ((Y).type.kids.E) );
			if zRes0 == nil {  return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);

			/* merge combs of Y with and without var */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zTemp = cuddZddUnion( dd, ((Y).type.kids.E), ((Y).type.kids.T) );
			if zTemp == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zTemp).ref);

			/* combs of X with label that are supersets of combs in Temp */
			// inlining of macro cuddT
			zRes1 = extraZddSupSet( dd, ((X).type.kids.T), zTemp );
			if zRes1 == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zTemp );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);
			Cudd_RecursiveDerefZdd( dd, zTemp );

			/* compose Res0 and Res1 with the given ZDD variable */
			zRes = cuddZddGetNode( dd, X.index, zRes1, zRes0 );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zRes1 );
				return nil;
			}
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes0).ref);
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes1).ref);
		
		} else { /* if ( TopLevelX > TopLevelY ) */
		
			/* combs of X that are supersets of combs of Y without label */
			// inlining of macro cuddE
			zRes = extraZddSupSet( dd, X, ((Y).type.kids.E) );
			if zRes == nil {  return nil; }
		}

		/* insert the result into cache */
	    cuddCacheInsert2(dd, extraZddSupSet, X, Y, zRes);
		return zRes;
	}
} /* end of extraZddSupSet */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddNotSubSet.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func extraZddNotSubSet(dd *DdManager, X *DdNode, Y *DdNode) *DdNode {	
	var zRes *DdNode;
    statLine(dd); 
	/* any comb is a subset of itself */
	if X == Y { 
		return z0; }
	/* combs in X are notsubsets of non-existant combs in Y */
	if Y == z0 { 
		return X; }	
	/* only {()} is the subset of {()} */
	if Y == z1 { /* remove empty combination from X */
		return cuddZddDiff( dd, X, z1 ); }
	/* if X is empty, the result is empty */
	if X == z0 { 
		return z0; }
	/* the empty comb is contained in all combs of Y */
	if X == z1 { 
		return z0; }

    /* check cache */
    zRes = cuddCacheLookup2Zdd(dd, extraZddNotSubSet, X, Y);
    if zRes != nil {
    	return(zRes);
	} else {
		var zRes0, zRes1, zTemp *DdNode;
		var TopLevelX int32  = dd.permZ[ X.index ];
		var TopLevelY int32  = dd.permZ[ Y.index ];

		if TopLevelX < TopLevelY {
			/* compute combs of X without var that are notsubsets of combs with Y */
			// inlining of macro cuddE
			zRes0 = extraZddNotSubSet( dd, ((X).type.kids.E), Y );
			if zRes0 == nil {  
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);

			/* combs of X with var cannot be subsets of combs without var in Y */
			// inlining of macro cuddT
			zRes1 = ((X).type.kids.T);

			/* compose Res0 and Res1 with the given ZDD variable */
			zRes = cuddZddGetNode( dd, X.index, zRes1, zRes0 );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				return nil;
			}
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes0).ref);
		
		} else if TopLevelX == TopLevelY {
			/* merge combs of Y with and without var */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zTemp = cuddZddUnion( dd, ((Y).type.kids.E), ((Y).type.kids.T) );
			if zTemp == nil {  
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zTemp).ref);

			/* compute combs of X without var that are notsubsets of combs is Temp */
			// inlining of macro cuddE
			zRes0 = extraZddNotSubSet( dd, ((X).type.kids.E), zTemp );
			if zRes0 == nil {
				Cudd_RecursiveDerefZdd( dd, zTemp );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);
			Cudd_RecursiveDerefZdd( dd, zTemp );

			/* combs of X with var that are notsubsets of combs in Y with var */
			// inlining of macro cuddT
			zRes1 = extraZddNotSubSet( dd, ((X).type.kids.T), ((Y).type.kids.T) );
			if zRes1 == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);

			/* compose Res0 and Res1 with the given ZDD variable */
			zRes = cuddZddGetNode( dd, X.index, zRes1, zRes0 );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zRes1 );
				return nil;
			}
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes0).ref);
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes1).ref);
		
		} else { /* if ( TopLevelX > TopLevelY ) */
		
			/* merge combs of Y with and without var */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zTemp = cuddZddUnion( dd, ((Y).type.kids.E), ((Y).type.kids.T) );
			if zTemp == nil {  
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zTemp).ref);

			/* compute combs that are notsubsets of Temp */
			zRes = extraZddNotSubSet( dd, X, zTemp );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zTemp );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes).ref);
			Cudd_RecursiveDerefZdd( dd, zTemp );
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes).ref);
		}

		/* insert the result into cache */
	    cuddCacheInsert2(dd, extraZddNotSubSet, X, Y, zRes);
		return zRes;
	}
} /* end of extraZddNotSubSet */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddNotSupSet.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func extraZddNotSupSet(dd *DdManager, X *DdNode, Y *DdNode) *DdNode {	
	var zRes *DdNode;
    statLine(dd); 
	/* any comb is a superset of itself */
	if X == Y { 
		return z0; }
	/* no comb in X is superset of non-existing combs */
	if Y == z0 { 
		return X; }	
	/* any comb in X is the superset of the empty comb */
	if Extra_zddEmptyBelongs( dd, Y ) != 0 { 
		return z0; }
	/* if X is empty, the result is empty */
	if X == z0 { 
		return z0; }
	/* if X is the empty comb (and Y does not contain it!), return it */
	if X == z1 { 
		return z1; }

    /* check cache */
    zRes = cuddCacheLookup2Zdd(dd, extraZddNotSupSet, X, Y);
    if zRes != nil {
    	return(zRes);
	} else {
		var zRes0, zRes1, zTemp *DdNode;
		var TopLevelX int32  = dd.permZ[ X.index ];
		var TopLevelY int32  = dd.permZ[ Y.index ];

		if TopLevelX < TopLevelY {
			/* combinations of X without label that are supersets of combinations of Y */
			// inlining of macro cuddE
			zRes0 = extraZddNotSupSet( dd, ((X).type.kids.E), Y );
			if zRes0 == nil {  
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);

			/* combinations of X with label that are supersets of combinations of Y */
			// inlining of macro cuddT
			zRes1 = extraZddNotSupSet( dd, ((X).type.kids.T), Y );
			if zRes1 == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);

			/* compose Res0 and Res1 with the given ZDD variable */
			zRes = cuddZddGetNode( dd, X.index, zRes1, zRes0 );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zRes1 );
				return nil;
			}
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes0).ref);
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes1).ref);
		
		} else if TopLevelX == TopLevelY {
			/* combs of X without var that are not supersets of combs of Y without var */
			// inlining of macro cuddE
			zRes0 = extraZddNotSupSet( dd, ((X).type.kids.E), ((Y).type.kids.E) );
			if zRes0 == nil {  
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);

			/* merge combs of Y with and without var */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zTemp = cuddZddUnion( dd, ((Y).type.kids.E), ((Y).type.kids.T) );
			if zTemp == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zTemp).ref);

			/* combs of X with label that are supersets of combs in Temp */
			// inlining of macro cuddT
			zRes1 = extraZddNotSupSet( dd, ((X).type.kids.T), zTemp );
			if zRes1 == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zTemp );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);
			Cudd_RecursiveDerefZdd( dd, zTemp );

			/* compose Res0 and Res1 with the given ZDD variable */
			zRes = cuddZddGetNode( dd, X.index, zRes1, zRes0 );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zRes1 );
				return nil;
			}
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes0).ref);
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes1).ref);
		
		} else { /* if ( TopLevelX > TopLevelY ) */
		
			/* combs of X that are supersets of combs of Y without label */
			// inlining of macro cuddE
			zRes = extraZddNotSupSet( dd, X, ((Y).type.kids.E) );
			if zRes == nil {  return nil; }
		}

		/* insert the result into cache */
	    cuddCacheInsert2(dd, extraZddNotSupSet, X, Y, zRes);
		return zRes;
	}
} /* end of extraZddNotSupSet */



/**Function********************************************************************

  Synopsis [Performs the recursive step of Extra_zddMaxNotSupSet.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func extraZddMaxNotSupSet(dd *DdManager, X *DdNode, Y *DdNode) *DdNode {	
	var zRes *DdNode;
    statLine(dd); 
	/* any comb is a superset of itself */
	if X == Y { 
		return z0; }
	/* no comb in X is superset of non-existing combs */
	if Y == z0 { 
		return extraZddMaximal( dd, X ); }	
	/* any comb in X is the superset of the empty comb */
	if Extra_zddEmptyBelongs( dd, Y ) != 0 { 
		return z0; }
	/* if X is empty, the result is empty */
	if X == z0 { 
		return z0; }
	/* if X is the empty comb (and Y does not contain it!), return it */
	if X == z1 { 
		return z1; }

    /* check cache */
    zRes = cuddCacheLookup2Zdd(dd, extraZddMaxNotSupSet, X, Y);
    if zRes != nil {
    	return(zRes);
	} else {
		var zRes0, zRes1, zTemp *DdNode;
		var TopLevelX int32  = dd.permZ[ X.index ];
		var TopLevelY int32  = dd.permZ[ Y.index ];

		if TopLevelX < TopLevelY {
			/* combinations of X without label that are supersets of combinations with Y */
			// inlining of macro cuddE
			zRes0 = extraZddMaxNotSupSet( dd, ((X).type.kids.E), Y );
			if zRes0 == nil {  
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);

			/* combinations of X with label that are supersets of combinations with Y */
			// inlining of macro cuddT
			zRes1 = extraZddMaxNotSupSet( dd, ((X).type.kids.T), Y );
			if zRes1 == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);

			/* ---------------------------------------------------- */
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
			/* ---------------------------------------------------- */

			/* compose Res0 and Res1 with the given ZDD variable */
			zRes = cuddZddGetNode( dd, X.index, zRes1, zRes0 );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zRes1 );
				return nil;
			}
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes0).ref);
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes1).ref);
		
		} else if TopLevelX == TopLevelY {
			/* combs of X without var that are supersets of combs of Y without var */
			// inlining of macro cuddE
			zRes0 = extraZddMaxNotSupSet( dd, ((X).type.kids.E), ((Y).type.kids.E) );
			if zRes0 == nil {  
				return nil; }
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes0).ref);

			/* merge combs of Y with and without var */
			// inlining of macro cuddE
			// inlining of macro cuddT
			zTemp = cuddZddUnion( dd, ((Y).type.kids.E), ((Y).type.kids.T) );
			if zTemp == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zTemp).ref);

			/* combs of X with label that are supersets of combs in Temp */
			// inlining of macro cuddT
			zRes1 = extraZddMaxNotSupSet( dd, ((X).type.kids.T), zTemp );
			if zRes1 == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zTemp );
				return nil;
			}
			// inlining of macro cuddRef
			cuddSatInc(Cudd_Regular(zRes1).ref);
			Cudd_RecursiveDerefZdd( dd, zTemp );

			/* ---------------------------------------------------- */
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
			/* ---------------------------------------------------- */

			/* compose Res0 and Res1 with the given ZDD variable */
			zRes = cuddZddGetNode( dd, X.index, zRes1, zRes0 );
			if zRes == nil {
				Cudd_RecursiveDerefZdd( dd, zRes0 );
				Cudd_RecursiveDerefZdd( dd, zRes1 );
				return nil;
			}
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes0).ref);
			// inlining of macro cuddDeref
			cuddSatDec(Cudd_Regular(zRes1).ref);
		
		} else { /* if ( TopLevelX > TopLevelY ) */
		
			/* combs of X that are supersets of combs of Y without label */
			// inlining of macro cuddE
			zRes = extraZddMaxNotSupSet( dd, X, ((Y).type.kids.E) );
			if zRes == nil {  return nil; }
		}

		/* insert the result into cache */
	    cuddCacheInsert2(dd, extraZddMaxNotSupSet, X, Y, zRes);
		return zRes;
	}
} /* end of extraZddMaxNotSupSet */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


//ABC_NAMESPACE_IMPL_END
