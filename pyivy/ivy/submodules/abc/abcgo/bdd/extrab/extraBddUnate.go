package extrab

import "fmt"

/**CFile****************************************************************

  FileName    [extraBddUnate.c]

  PackageName [extra]

  Synopsis    [Efficient methods to compute the information about
  unate variables using an algorithm that is conceptually similar to
  the algorithm for two-variable symmetry computation presented in:
  A. Mishchenko. Fast Computation of Symmetries in Boolean Functions. 
  Transactions on CAD, Nov. 2003.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - September 1, 2003.]

  Revision    [$Id: extraBddUnate.c,v 1.0 2003/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

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

  Synopsis    [Computes the classical symmetry information for the function.]

  Description [Returns the symmetry information in the form of Extra_UnateInfo_t structure.]

  SideEffects [If the ZDD variables are not derived from BDD variables with
  multiplicity 2, this function may derive them in a wrong way.]

  SeeAlso     []

******************************************************************************/
Extra_UnateInfo_t * Extra_UnateComputeFast( 
  DdManager * dd,   /* the manager */
  DdNode * bFunc)   /* the function whose symmetries are computed */
{
    var bSupp *DdNode;
    var zRes *DdNode;
    var p *Extra_UnateInfo_t;

    bSupp = Cudd_Support( dd, bFunc );                      Cudd_Ref( bSupp );
    zRes  = Extra_zddUnateInfoCompute( dd, bFunc, bSupp );  Cudd_Ref( zRes );

    p = Extra_UnateInfoCreateFromZdd( dd, zRes, bSupp );

    Cudd_RecursiveDeref( dd, bSupp );
    Cudd_RecursiveDerefZdd( dd, zRes );

    return p;

} /* end of Extra_UnateInfoCompute */


/**Function********************************************************************

  Synopsis    [Computes the classical symmetry information as a ZDD.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * Extra_zddUnateInfoCompute( 
  DdManager * dd,   /* the DD manager */
  DdNode * bF,
  DdNode * bVars) 
{
    var res *DdNode;
    for {
        dd.reordered = 0;
        res = extraZddUnateInfoCompute( dd, bF, bVars );
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddUnateInfoCompute */


/**Function********************************************************************

  Synopsis    [Converts a set of variables into a set of singleton subsets.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * Extra_zddGetSingletonsBoth( 
  DdManager * dd,    /* the DD manager */
  DdNode * bVars)    /* the set of variables */
{
    var res *DdNode;
    for {
        dd.reordered = 0;
        res = extraZddGetSingletonsBoth( dd, bVars );
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddGetSingletonsBoth */

/**Function********************************************************************

  Synopsis    [Allocates unateness information structure.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_UnateInfoAllocate(nVars int32) *Extra_UnateInfo_t {
    var p *Extra_UnateInfo_t;
    // allocate and clean the storage for unateness info
    p = new(Extra_UnateInfo_t);
    memset( p, 0, sizeof(Extra_UnateInfo_t) );
    p.nVars     = nVars;
    p.pVars     = ABC_ALLOC( Extra_UnateVar_t, nVars );  
    memset( p.pVars, 0, nVars * sizeof(Extra_UnateVar_t) );
    return p;
} /* end of Extra_UnateInfoAllocate */

/**Function********************************************************************

  Synopsis    [Deallocates symmetry information structure.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_UnateInfoDissolve(p *Extra_UnateInfo_t) {
    ABC_FREE( p.pVars );
    ABC_FREE( p );
} /* end of Extra_UnateInfoDissolve */

/**Function********************************************************************

  Synopsis    [Allocates symmetry information structure.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_UnateInfoPrint(p *Extra_UnateInfo_t) {
    var pBuffer string;
    var i int32;
    pBuffer = ABC_ALLOC( char, p.nVarsMax+1 );
    memset( pBuffer, ' ', (size_t)p.nVarsMax );
    pBuffer[p.nVarsMax] = 0;
    for  i = 0; i < p.nVars; i++  {
        if p.pVars[i].Neg {
            pBuffer[ p.pVars[i].iVar ] = 'n';
        } else if p.pVars[i].Pos {
            pBuffer[ p.pVars[i].iVar ] = 'p';
        } else {
            pBuffer[ p.pVars[i].iVar ] = '.'; }
    }
    fmt.Printf("%s\n", pBuffer);
    ABC_FREE( pBuffer );
} /* end of Extra_UnateInfoPrint */


/**Function********************************************************************

  Synopsis    [Creates the symmetry information structure from ZDD.]

  Description [ZDD representation of symmetries is the set of cubes, each
  of which has two variables in the positive polarity. These variables correspond
  to the symmetric variable pair.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_UnateInfoCreateFromZdd(dd *DdManager, zPairs *DdNode, bSupp *DdNode) *Extra_UnateInfo_t {
    var p *Extra_UnateInfo_t;
    var bTemp, zSet, zCube, zTemp *DdNode;
    var pMapVars2Nums *int32;
    var i, nSuppSize int32;

    nSuppSize = Extra_bddSuppSize( dd, bSupp );

    // allocate and clean the storage for symmetry info
    p = Extra_UnateInfoAllocate( nSuppSize );

    // allocate the storage for the temporary map
    pMapVars2Nums = ABC_ALLOC( int, dd.size );
    memset( pMapVars2Nums, 0, dd.size * sizeof(int) );

    // assign the variables
    p.nVarsMax = dd.size;
    // inlining of macro cuddT
    for  i, bTemp = 0, bSupp; bTemp != b1; bTemp = ((bTemp).type.kids.T), i++  {
        p.pVars[i].iVar = bTemp.index;
        pMapVars2Nums[bTemp.index] = i;
    }

    // write the symmetry info into the structure
    zSet = zPairs;   Cudd_Ref( zSet );
//    Cudd_zddPrintCover( dd, zPairs );    printf( "\n" );
    for  zSet != z0  {
        // get the next cube
        zCube  = Extra_zddSelectOneSubset( dd, zSet );    Cudd_Ref( zCube );

        // add this var to the data structure
        // inlining of macro cuddT
        // inlining of macro cuddE
        assert(((zCube).type.kids.T) == z1 && ((zCube).type.kids.E) == z0, "/*__starting_point_inline_macro:cuddT__*/((zCube).type.kids.T) == z1 && /*__starting_point_inline_macro:cuddE__*/((zCube).type.kids.E) == z0");
        if zCube.index & 1 { // neg
            p.pVars[ pMapVars2Nums[zCube.index/2] ].Neg = 1;
        } else {
            p.pVars[ pMapVars2Nums[zCube.index/2] ].Pos = 1; }
        // count the unate vars
        p.nUnate++;

        // update the cuver and deref the cube
        zSet = Cudd_zddDiff( dd, zTemp = zSet, zCube );     Cudd_Ref( zSet );
        Cudd_RecursiveDerefZdd( dd, zTemp );
        Cudd_RecursiveDerefZdd( dd, zCube );

    } // for each cube 
    Cudd_RecursiveDerefZdd( dd, zSet );
    ABC_FREE( pMapVars2Nums );
    return p;

} /* end of Extra_UnateInfoCreateFromZdd */



/**Function********************************************************************

  Synopsis    [Computes the classical unateness information for the function.]

  Description [Uses the naive way of comparing cofactors.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_UnateComputeSlow(dd *DdManager, bFunc *DdNode) *Extra_UnateInfo_t {
    var nSuppSize int32;
    var bSupp, bTemp *DdNode;
    var p *Extra_UnateInfo_t;
    var i, Res int32;

    // compute the support
    bSupp = Cudd_Support( dd, bFunc );   Cudd_Ref( bSupp );
    nSuppSize = Extra_bddSuppSize( dd, bSupp );
//printf( "Support = %d. ", nSuppSize );
//Extra_bddPrint( dd, bSupp );
//printf( "%d ", nSuppSize );

    // allocate the storage for symmetry info
    p = Extra_UnateInfoAllocate( nSuppSize );

    // assign the variables
    p.nVarsMax = dd.size;
    // inlining of macro cuddT
    for  i, bTemp = 0, bSupp; bTemp != b1; bTemp = ((bTemp).type.kids.T), i++  {
        Res = Extra_bddCheckUnateNaive( dd, bFunc, bTemp.index );
        p.pVars[i].iVar = bTemp.index;
        if Res == -1 {
            p.pVars[i].Neg = 1;
        } else if Res == 1 {
            p.pVars[i].Pos = 1; }
        p.nUnate += (Res != 0);
    }
    Cudd_RecursiveDeref( dd, bSupp );
    return p;

} /* end of Extra_UnateComputeSlow */

/**Function********************************************************************

  Synopsis    [Checks if the two variables are symmetric.]

  Description [Returns 0 if vars are not unate. Return -1/+1 if the var is neg/pos unate.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
int Extra_bddCheckUnateNaive( 
  DdManager * dd,   /* the DD manager */
  DdNode * bF,
  int iVar) 
{
    var bCof0, bCof1 *DdNode;
    var Res int32;

    assert(iVar < dd.size, "iVar < dd.size");

    // inlining of macro Cudd_Not
    bCof0 = Cudd_Cofactor( dd, bF, ((DdNode *)((ptrint)(Cudd_bddIthVar(dd,iVar)) ^ 01)) );  Cudd_Ref( bCof0 );
    bCof1 = Cudd_Cofactor( dd, bF, Cudd_bddIthVar(dd,iVar) );            Cudd_Ref( bCof1 );

    if Cudd_bddLeq( dd, bCof0, bCof1 ) {
        Res = 1;
    } else if Cudd_bddLeq( dd, bCof1, bCof0 ) {
        Res =-1;
    } else {
        Res = 0; }

    Cudd_RecursiveDeref( dd, bCof0 );
    Cudd_RecursiveDeref( dd, bCof1 );
    return Res;
} /* end of Extra_bddCheckUnateNaive */



/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Performs a recursive step of Extra_UnateInfoCompute.]

  Description [Returns the set of symmetric variable pairs represented as a set 
  of two-literal ZDD cubes. Both variables always appear in the positive polarity
  in the cubes. This function works without building new BDD nodes. Some relatively 
  small number of ZDD nodes may be built to ensure proper bookkeeping of the 
  symmetry information.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * 
extraZddUnateInfoCompute( 
  DdManager * dd,   /* the manager */
  DdNode * bFunc,   /* the function whose symmetries are computed */
  DdNode * bVars )  /* the set of variables on which this function depends */
{
    var zRes *DdNode;
    // inlining of macro Cudd_Regular
    var bFR *DdNode  = ((DdNode *)((ptruint)(bFunc) & ~01)); 

    // inlining of macro cuddIsConstant
    if  (bFR).index == CUDD_CONST_INDEX {
        // inlining of macro cuddIsConstant
        if  (bVars).index == CUDD_CONST_INDEX {
            return z0; }
        return extraZddGetSingletonsBoth( dd, bVars );
    }
    assert(bVars != b1, "bVars != b1");

    if zRes = cuddCacheLookup2Zdd(dd, extraZddUnateInfoCompute, bFunc, bVars) {
        return zRes;
    } else {
        var zRes0, zRes1 *DdNode;
        var zTemp, zPlus *DdNode;             
        var bF0, bF1 *DdNode;             
        var bVarsNew *DdNode;
        var nVarsExtra int32;
        var LevelF int32;
        var AddVar int32;

        // every variable in bF should be also in bVars, therefore LevelF cannot be above LevelV
        // if LevelF is below LevelV, scroll through the vars in bVars to the same level as F
        // count how many extra vars are there in bVars
        nVarsExtra = 0;
        LevelF = dd.perm[bFR.index];
        // inlining of macro cuddT
        for  bVarsNew = bVars; LevelF > dd.perm[bVarsNew.index]; bVarsNew = ((bVarsNew).type.kids.T)  {
            nVarsExtra++;
        } 
        // the indexes (level) of variables should be synchronized now
        assert(bFR.index == bVarsNew.index, "bFR.index == bVarsNew.index");

        // cofactor the function
        if bFR != bFunc { // bFunc is complemented 
        
            // inlining of macro Cudd_Not
            // inlining of macro cuddE
            bF0 = ((DdNode *)((ptrint)(((bFR).type.kids.E)) ^ 01));
            // inlining of macro Cudd_Not
            // inlining of macro cuddT
            bF1 = ((DdNode *)((ptrint)(((bFR).type.kids.T)) ^ 01));
        
        } else {
            // inlining of macro cuddE
            bF0 = ((bFR).type.kids.E);
            // inlining of macro cuddT
            bF1 = ((bFR).type.kids.T);
        }

        // solve subproblems
        // inlining of macro cuddT
        zRes0 = extraZddUnateInfoCompute( dd, bF0, ((bVarsNew).type.kids.T) );
        if zRes0 == nil {
            return nil; }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zRes0).ref);

        // if there is no symmetries in the negative cofactor
        // there is no need to test the positive cofactor
        if zRes0 == z0 {
            zRes = zRes0;  // zRes takes reference
        } else {
            // inlining of macro cuddT
            zRes1 = extraZddUnateInfoCompute( dd, bF1, ((bVarsNew).type.kids.T) );
            if zRes1 == nil {
                Cudd_RecursiveDerefZdd( dd, zRes0 );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zRes1).ref);

            // only those variables are pair-wise symmetric 
            // that are pair-wise symmetric in both cofactors
            // therefore, intersect the solutions
            zRes = cuddZddIntersect( dd, zRes0, zRes1 );
            if zRes == nil {
                Cudd_RecursiveDerefZdd( dd, zRes0 );
                Cudd_RecursiveDerefZdd( dd, zRes1 );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zRes).ref);
            Cudd_RecursiveDerefZdd( dd, zRes0 );
            Cudd_RecursiveDerefZdd( dd, zRes1 );
        }

        // consider the current top-most variable
        AddVar = -1;
        if Cudd_bddLeq( dd, bF0, bF1 ) { // pos
            AddVar = 0;
        } else if Cudd_bddLeq( dd, bF1, bF0 ) { // neg
            AddVar = 1; }
        if AddVar >= 0 {
            // create the singleton
            zPlus = cuddZddGetNode( dd, 2*bFR.index + AddVar, z1, z0 );
            if zPlus == nil {
                Cudd_RecursiveDerefZdd( dd, zRes );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zPlus).ref);

            // add these to the result
            zRes = cuddZddUnion( dd, zTemp = zRes, zPlus );
            if zRes == nil {
                Cudd_RecursiveDerefZdd( dd, zTemp );
                Cudd_RecursiveDerefZdd( dd, zPlus );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zRes).ref);
            Cudd_RecursiveDerefZdd( dd, zTemp );
            Cudd_RecursiveDerefZdd( dd, zPlus );
        }
        // only zRes is referenced at this point

        // if we skipped some variables, these variables cannot be symmetric with
        // any variables that are currently in the support of bF, but they can be 
        // symmetric with the variables that are in bVars but not in the support of bF
        // inlining of macro cuddT
        for  bVarsNew = bVars; LevelF > dd.perm[bVarsNew.index]; bVarsNew = ((bVarsNew).type.kids.T)  {
            // create the negative singleton
            zPlus = cuddZddGetNode( dd, 2*bVarsNew.index+1, z1, z0 );
            if zPlus == nil {
                Cudd_RecursiveDerefZdd( dd, zRes );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zPlus).ref);

            // add these to the result
            zRes = cuddZddUnion( dd, zTemp = zRes, zPlus );
            if zRes == nil {
                Cudd_RecursiveDerefZdd( dd, zTemp );
                Cudd_RecursiveDerefZdd( dd, zPlus );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zRes).ref);
            Cudd_RecursiveDerefZdd( dd, zTemp );
            Cudd_RecursiveDerefZdd( dd, zPlus );
        

            // create the positive singleton
            zPlus = cuddZddGetNode( dd, 2*bVarsNew.index, z1, z0 );
            if zPlus == nil {
                Cudd_RecursiveDerefZdd( dd, zRes );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zPlus).ref);

            // add these to the result
            zRes = cuddZddUnion( dd, zTemp = zRes, zPlus );
            if zRes == nil {
                Cudd_RecursiveDerefZdd( dd, zTemp );
                Cudd_RecursiveDerefZdd( dd, zPlus );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zRes).ref);
            Cudd_RecursiveDerefZdd( dd, zTemp );
            Cudd_RecursiveDerefZdd( dd, zPlus );
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(zRes).ref);

        /* insert the result into cache */
        cuddCacheInsert2(dd, extraZddUnateInfoCompute, bFunc, bVars, zRes);
        return zRes;
    }
} /* end of extraZddUnateInfoCompute */


/**Function********************************************************************

  Synopsis    [Performs a recursive step of Extra_zddGetSingletons.]

  Description [Returns the set of ZDD singletons, containing those pos/neg 
  polarity ZDD variables that correspond to the BDD variables in bVars.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * extraZddGetSingletonsBoth( 
  DdManager * dd,    /* the DD manager */
  DdNode * bVars)    /* the set of variables */
{
    var zRes *DdNode;

    if bVars == b1 {
        return z1; }

    if zRes = cuddCacheLookup1Zdd(dd, extraZddGetSingletonsBoth, bVars) {
        return zRes;
    } else {
        var zTemp, zPlus *DdNode;          

        // solve subproblem
        // inlining of macro cuddT
        zRes = extraZddGetSingletonsBoth( dd, ((bVars).type.kids.T) );
        if zRes == nil { 
            return nil; }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zRes).ref);

        
        // create the negative singleton
        zPlus = cuddZddGetNode( dd, 2*bVars.index+1, z1, z0 );
        if zPlus == nil {
            Cudd_RecursiveDerefZdd( dd, zRes );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zPlus).ref);

        // add these to the result
        zRes = cuddZddUnion( dd, zTemp = zRes, zPlus );
        if zRes == nil {
            Cudd_RecursiveDerefZdd( dd, zTemp );
            Cudd_RecursiveDerefZdd( dd, zPlus );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zRes).ref);
        Cudd_RecursiveDerefZdd( dd, zTemp );
        Cudd_RecursiveDerefZdd( dd, zPlus );
        

        // create the positive singleton
        zPlus = cuddZddGetNode( dd, 2*bVars.index, z1, z0 );
        if zPlus == nil {
            Cudd_RecursiveDerefZdd( dd, zRes );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zPlus).ref);

        // add these to the result
        zRes = cuddZddUnion( dd, zTemp = zRes, zPlus );
        if zRes == nil {
            Cudd_RecursiveDerefZdd( dd, zTemp );
            Cudd_RecursiveDerefZdd( dd, zPlus );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zRes).ref);
        Cudd_RecursiveDerefZdd( dd, zTemp );
        Cudd_RecursiveDerefZdd( dd, zPlus );

        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(zRes).ref);
        cuddCacheInsert1( dd, extraZddGetSingletonsBoth, bVars, zRes );
        return zRes;
    }
}   /* end of extraZddGetSingletonsBoth */


/*---------------------------------------------------------------------------*/
/* Definition of static Functions                                            */
/*---------------------------------------------------------------------------*/
//ABC_NAMESPACE_IMPL_END

