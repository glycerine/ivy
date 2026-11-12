package extrab

import "fmt"

/**CFile****************************************************************

  FileName    [extraBddSymm.c]

  PackageName [extra]

  Synopsis    [Efficient methods to compute the information about
  symmetric variables using the algorithm presented in the paper:
  A. Mishchenko. Fast Computation of Symmetries in Boolean Functions. 
  Transactions on CAD, Nov. 2003.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - September 1, 2003.]

  Revision    [$Id: extraBddSymm.c,v 1.0 2003/09/01 00:00:00 alanmi Exp $]

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

const DD_GET_SYMM_VARS_TAG =          0x0a /* former DD_BDD_XOR_EXIST_ABSTRACT_TAG */

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

  Description [Returns the symmetry information in the form of Extra_SymmInfo_t structure.]

  SideEffects [If the ZDD variables are not derived from BDD variables with
  multiplicity 2, this function may derive them in a wrong way.]

  SeeAlso     []

******************************************************************************/
Extra_SymmInfo_t * Extra_SymmPairsCompute( 
  DdManager * dd,   /* the manager */
  DdNode * bFunc)   /* the function whose symmetries are computed */
{
    var bSupp *DdNode;
    var zRes *DdNode;
    var p *Extra_SymmInfo_t;

    bSupp = Cudd_Support( dd, bFunc );                      Cudd_Ref( bSupp );
    zRes  = Extra_zddSymmPairsCompute( dd, bFunc, bSupp );  Cudd_Ref( zRes );

    p = Extra_SymmPairsCreateFromZdd( dd, zRes, bSupp );

    Cudd_RecursiveDeref( dd, bSupp );
    Cudd_RecursiveDerefZdd( dd, zRes );

    return p;

} /* end of Extra_SymmPairsCompute */


/**Function********************************************************************

  Synopsis    [Computes the classical symmetry information as a ZDD.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * Extra_zddSymmPairsCompute( 
  DdManager * dd,   /* the DD manager */
  DdNode * bF,
  DdNode * bVars) 
{
    var res *DdNode;
    for {
        dd.reordered = 0;
        res = extraZddSymmPairsCompute( dd, bF, bVars );
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddSymmPairsCompute */

/**Function********************************************************************

  Synopsis    [Returns a singleton-set ZDD containing all variables that are symmetric with the given one.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * Extra_zddGetSymmetricVars( 
  DdManager * dd,    /* the DD manager */
  DdNode * bF,       /* the first function  - originally, the positive cofactor */
  DdNode * bG,       /* the second fucntion - originally, the negative cofactor */
  DdNode * bVars)    /* the set of variables, on which F and G depend */
{
    var res *DdNode;
    for {
        dd.reordered = 0;
        res = extraZddGetSymmetricVars( dd, bF, bG, bVars );
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddGetSymmetricVars */


/**Function********************************************************************

  Synopsis    [Converts a set of variables into a set of singleton subsets.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * Extra_zddGetSingletons( 
  DdManager * dd,    /* the DD manager */
  DdNode * bVars)    /* the set of variables */
{
    var res *DdNode;
    for {
        dd.reordered = 0;
        res = extraZddGetSingletons( dd, bVars );
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddGetSingletons */

/**Function********************************************************************

  Synopsis    [Filters the set of variables using the support of the function.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * Extra_bddReduceVarSet( 
  DdManager * dd,    /* the DD manager */
  DdNode * bVars,    /* the set of variables to be reduced */
  DdNode * bF)       /* the function whose support is used for reduction */
{
    var res *DdNode;
    for {
        dd.reordered = 0;
        res = extraBddReduceVarSet( dd, bVars, bF );
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_bddReduceVarSet */


/**Function********************************************************************

  Synopsis    [Allocates symmetry information structure.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_SymmPairsAllocate(nVars int32) *Extra_SymmInfo_t {
    var i int32;
    var p *Extra_SymmInfo_t;

    // allocate and clean the storage for symmetry info
    p = new(Extra_SymmInfo_t);
    memset( p, 0, sizeof(Extra_SymmInfo_t) );
    p.nVars     = nVars;
    p.pVars     = ABC_ALLOC( int, nVars );  
    p.pSymms    = ABC_ALLOC( char *, nVars );  
    p.pSymms[0] = ABC_ALLOC( char  , nVars * nVars );
    memset( p.pSymms[0], 0, nVars * nVars * sizeof(char) );

    for  i = 1; i < nVars; i++  {
        p.pSymms[i] = p.pSymms[i-1] + nVars;
    }

    return p;
} /* end of Extra_SymmPairsAllocate */

/**Function********************************************************************

  Synopsis    [Deallocates symmetry information structure.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_SymmPairsDissolve(p *Extra_SymmInfo_t) {
    ABC_FREE( p.pVars );
    ABC_FREE( p.pSymms[0] );
    ABC_FREE( p.pSymms    );
    ABC_FREE( p );
} /* end of Extra_SymmPairsDissolve */

/**Function********************************************************************

  Synopsis    [Allocates symmetry information structure.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_SymmPairsPrint(p *Extra_SymmInfo_t) {
    var i, k int32;
    fmt.Printf("\n");
    for  i = 0; i < p.nVars; i++  {
        for  k = 0; k <= i; k++  {
            fmt.Printf(" ");
        }
        for  k = i+1; k < p.nVars; k++  {
            if p.pSymms[i][k] {
                fmt.Printf("1");
            } else {
                fmt.Printf("."); }
        }
        fmt.Printf("\n");
    }
} /* end of Extra_SymmPairsPrint */


/**Function********************************************************************

  Synopsis    [Creates the symmetry information structure from ZDD.]

  Description [ZDD representation of symmetries is the set of cubes, each
  of which has two variables in the positive polarity. These variables correspond
  to the symmetric variable pair.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_SymmPairsCreateFromZdd(dd *DdManager, zPairs *DdNode, bSupp *DdNode) *Extra_SymmInfo_t {
    var i int32;
    var nSuppSize int32;
    var p *Extra_SymmInfo_t;
    var pMapVars2Nums *int32;
    var bTemp *DdNode;
    var zSet, zCube, zTemp *DdNode;
    var iVar1, iVar2 int32;

    nSuppSize = Extra_bddSuppSize( dd, bSupp );

    // allocate and clean the storage for symmetry info
    p = Extra_SymmPairsAllocate( nSuppSize );

    // allocate the storage for the temporary map
    pMapVars2Nums = ABC_ALLOC( int, dd.size );
    memset( pMapVars2Nums, 0, dd.size * sizeof(int) );

    // assign the variables
    p.nVarsMax = dd.size;
//  p->nNodes   = Cudd_DagSize( zPairs );
    p.nNodes   = 0;
    // inlining of macro cuddT
    for  i, bTemp = 0, bSupp; bTemp != b1; bTemp = ((bTemp).type.kids.T), i++  {
        p.pVars[i] = bTemp.index;
        pMapVars2Nums[bTemp.index] = i;
    }

    // write the symmetry info into the structure
    zSet = zPairs;   Cudd_Ref( zSet );
    for  zSet != z0  {
        // get the next cube
        zCube  = Extra_zddSelectOneSubset( dd, zSet );    Cudd_Ref( zCube );

        // add these two variables to the data structure
        // inlining of macro cuddT
        assert(((((zCube).type.kids.T)).type.kids.T) == z1, "/*__starting_point_inline_macro:cuddT__*/((/*__starting_point_inline_macro:cuddT__*/((zCube).type.kids.T)).type.kids.T) == z1");
        iVar1 = zCube.index/2;
        // inlining of macro cuddT
        iVar2 = ((zCube).type.kids.T).index/2;
        if pMapVars2Nums[iVar1] < pMapVars2Nums[iVar2] {
            p.pSymms[ pMapVars2Nums[iVar1] ][ pMapVars2Nums[iVar2] ] = 1;
        } else {
            p.pSymms[ pMapVars2Nums[iVar2] ][ pMapVars2Nums[iVar1] ] = 1; }
        // count the symmetric pairs
        p.nSymms ++;

        // update the cuver and deref the cube
        zSet = Cudd_zddDiff( dd, zTemp = zSet, zCube );     Cudd_Ref( zSet );
        Cudd_RecursiveDerefZdd( dd, zTemp );
        Cudd_RecursiveDerefZdd( dd, zCube );

    } // for each cube 
    Cudd_RecursiveDerefZdd( dd, zSet );

    ABC_FREE( pMapVars2Nums );
    return p;

} /* end of Extra_SymmPairsCreateFromZdd */


/**Function********************************************************************

  Synopsis    [Checks the possibility of two variables being symmetric.]

  Description [Returns 0 if vars are not symmetric. Return 1 if vars can be symmetric.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
int Extra_bddCheckVarsSymmetric( 
  DdManager * dd,   /* the DD manager */
  DdNode * bF,
  int iVar1,
  int iVar2) 
{
    var bVars *DdNode;
    var Res int32;

//  return 1;

    assert(iVar1 != iVar2, "iVar1 != iVar2");
    assert(iVar1 < dd.size, "iVar1 < dd.size");
    assert(iVar2 < dd.size, "iVar2 < dd.size");

    bVars = Cudd_bddAnd( dd, dd.vars[iVar1], dd.vars[iVar2] );   Cudd_Ref( bVars );

    Res = int32(( extraBddCheckVarsSymmetric( dd, bF, bVars ) == b1 ));

    Cudd_RecursiveDeref( dd, bVars );

    return Res;
} /* end of Extra_bddCheckVarsSymmetric */


/**Function********************************************************************

  Synopsis    [Computes the classical symmetry information for the function.]

  Description [Uses the naive way of comparing cofactors.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_SymmPairsComputeNaive(dd *DdManager, bFunc *DdNode) *Extra_SymmInfo_t {
    var bSupp, bTemp *DdNode;
    var nSuppSize int32;
    var p *Extra_SymmInfo_t;
    var i, k int32;

    // compute the support
    bSupp = Cudd_Support( dd, bFunc );   Cudd_Ref( bSupp );
    nSuppSize = Extra_bddSuppSize( dd, bSupp );
//printf( "Support = %d. ", nSuppSize );
//Extra_bddPrint( dd, bSupp );
//printf( "%d ", nSuppSize );

    // allocate the storage for symmetry info
    p = Extra_SymmPairsAllocate( nSuppSize );

    // assign the variables
    p.nVarsMax = dd.size;
    // inlining of macro cuddT
    for  i, bTemp = 0, bSupp; bTemp != b1; bTemp = ((bTemp).type.kids.T), i++  {
        p.pVars[i] = bTemp.index;
    }

    // go through the candidate pairs and check using Idea1
    for  i = 0; i < nSuppSize; i++  {
    for  k = i+1; k < nSuppSize; k++  {
        p.pSymms[k][i] = p.pSymms[i][k] = Extra_bddCheckVarsSymmetricNaive( dd, bFunc, p.pVars[i], p.pVars[k] );
        if p.pSymms[i][k] {
             p.nSymms++; }
    }
    }

    Cudd_RecursiveDeref( dd, bSupp );
    return p;

} /* end of Extra_SymmPairsComputeNaive */

/**Function********************************************************************

  Synopsis    [Checks if the two variables are symmetric.]

  Description [Returns 0 if vars are not symmetric. Return 1 if vars are symmetric.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
int Extra_bddCheckVarsSymmetricNaive( 
  DdManager * dd,   /* the DD manager */
  DdNode * bF,
  int iVar1,
  int iVar2) 
{
    var bCube1, bCube2 *DdNode;
    var bCof01, bCof10 *DdNode;
    var Res int32;

    assert(iVar1 != iVar2, "iVar1 != iVar2");
    assert(iVar1 < dd.size, "iVar1 < dd.size");
    assert(iVar2 < dd.size, "iVar2 < dd.size");

    // inlining of macro Cudd_Not
    bCube1 = Cudd_bddAnd( dd, ((DdNode *)((ptrint)(dd.vars[iVar1]) ^ 01)), dd.vars[iVar2] );   Cudd_Ref( bCube1 );
    // inlining of macro Cudd_Not
    bCube2 = Cudd_bddAnd( dd, ((DdNode *)((ptrint)(dd.vars[iVar2]) ^ 01)), dd.vars[iVar1] );   Cudd_Ref( bCube2 );

    bCof01 = Cudd_Cofactor( dd, bF, bCube1 );  Cudd_Ref( bCof01 );
    bCof10 = Cudd_Cofactor( dd, bF, bCube2 );  Cudd_Ref( bCof10 );

    Res = int32(( bCof10 == bCof01 ));

    Cudd_RecursiveDeref( dd, bCof01 );
    Cudd_RecursiveDeref( dd, bCof10 );
    Cudd_RecursiveDeref( dd, bCube1 );
    Cudd_RecursiveDeref( dd, bCube2 );

    return Res;
} /* end of Extra_bddCheckVarsSymmetricNaive */


/**Function********************************************************************

  Synopsis    [Builds ZDD representing the set of fixed-size variable tuples.]

  Description [Creates ZDD of all combinations of variables in Support that
  is represented by a BDD.]

  SideEffects [New ZDD variables are created if indices of the variables
               present in the combination are larger than the currently
               allocated number of ZDD variables.]

  SeeAlso     []

******************************************************************************/
DdNode* Extra_zddTuplesFromBdd( 
  DdManager * dd,   /* the DD manager */
  int K,            /* the number of variables in tuples */
  DdNode * bVarsN)   /* the set of all variables represented as a BDD */
{
    var zRes *DdNode;
    var autoDynZ int32;

    autoDynZ = dd.autoDynZ;
    dd.autoDynZ = 0;

    for {
        /* transform the numeric arguments (K) into a DdNode* argument;
         * this allows us to use the standard internal CUDD cache */
        var bVarSet *DdNode = bVarsN;
        var bVarsK *DdNode = bVarsN;
        var nVars, i int32;

        /* determine the number of variables in VarSet */
        for  bVarSet != b1  {
            nVars++;
            /* make sure that the VarSet is a cube */
            // inlining of macro cuddE
            if  ((bVarSet).type.kids.E) != b0 {
                return nil; }
            // inlining of macro cuddT
            bVarSet = ((bVarSet).type.kids.T);
        }
        /* make sure that the number of variables in VarSet is less or equal 
           that the number of variables that should be present in the tuples
        */
        if K > nVars {
            return nil; }

        /* the second argument in the recursive call stannds for <n>;
         * reate the first argument, which stands for <k>
         * as when we are talking about the tuple of <k> out of <n> */
        for  i = 0; i < nVars-K; i++  {
            // inlining of macro cuddT
            bVarsK = ((bVarsK).type.kids.T);
        }

        dd.reordered = 0;
        zRes = extraZddTuplesFromBdd(dd, bVarsK, bVarsN );

    
if !(dd.reordered == 1) { break }
} 
    dd.autoDynZ = autoDynZ;
    return zRes;

} /* end of Extra_zddTuplesFromBdd */

/**Function********************************************************************

  Synopsis    [Selects one subset from the set of subsets represented by a ZDD.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
DdNode* Extra_zddSelectOneSubset( 
  DdManager * dd,   /* the DD manager */
  DdNode * zS)      /* the ZDD */
{
    var res *DdNode;
    for {
        dd.reordered = 0;
        res = extraZddSelectOneSubset(dd, zS);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddSelectOneSubset */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Performs a recursive step of Extra_SymmPairsCompute.]

  Description [Returns the set of symmetric variable pairs represented as a set 
  of two-literal ZDD cubes. Both variables always appear in the positive polarity
  in the cubes. This function works without building new BDD nodes. Some relatively 
  small number of ZDD nodes may be built to ensure proper bookkeeping of the 
  symmetry information.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * 
extraZddSymmPairsCompute( 
  DdManager * dd,   /* the manager */
  DdNode * bFunc,   /* the function whose symmetries are computed */
  DdNode * bVars )  /* the set of variables on which this function depends */
{
    var zRes *DdNode;
    // inlining of macro Cudd_Regular
    var bFR *DdNode  = ((DdNode *)((ptruint)(bFunc) & ~01)); 

    // inlining of macro cuddIsConstant
    if  (bFR).index == CUDD_CONST_INDEX {
        var nVars, i int32;

        // determine how many vars are in the bVars
        nVars = Extra_bddSuppSize( dd, bVars );
        if nVars < 2 {
            return z0;
        } else {
            var bVarsK *DdNode;

            // create the BDD bVarsK corresponding to K = 2;
            bVarsK = bVars;
            for  i = 0; i < nVars-2; i++  {
                // inlining of macro cuddT
                bVarsK = ((bVarsK).type.kids.T);
            }
            return extraZddTuplesFromBdd( dd, bVarsK, bVars );
        }
    }
    assert(bVars != b1, "bVars != b1");

    if zRes = cuddCacheLookup2Zdd(dd, extraZddSymmPairsCompute, bFunc, bVars) {
        return zRes;
    } else {
        var zRes0, zRes1 *DdNode;
        var zTemp, zPlus, zSymmVars *DdNode;             
        var bF0, bF1 *DdNode;             
        var bVarsNew *DdNode;
        var nVarsExtra int32;
        var LevelF int32;

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
        zRes0 = extraZddSymmPairsCompute( dd, bF0, ((bVarsNew).type.kids.T) );
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
            zRes1 = extraZddSymmPairsCompute( dd, bF1, ((bVarsNew).type.kids.T) );
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

        // consider the current top-most variable and find all the vars
        // that are pairwise symmetric with it
        // these variables are returned as a set of ZDD singletons
        // inlining of macro cuddT
        zSymmVars = extraZddGetSymmetricVars( dd, bF1, bF0, ((bVarsNew).type.kids.T) );
        if zSymmVars == nil {
            Cudd_RecursiveDerefZdd( dd, zRes );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zSymmVars).ref);

        // attach the topmost variable to the set, to get the variable pairs
        // use the positive polarity ZDD variable for the purpose

        // there is no need to do so, if zSymmVars is empty
        if zSymmVars == z0 {
            Cudd_RecursiveDerefZdd( dd, zSymmVars );
        } else {
            zPlus = cuddZddGetNode( dd, 2*bFR.index, zSymmVars, z0 );
            if zPlus == nil {
                Cudd_RecursiveDerefZdd( dd, zRes );
                Cudd_RecursiveDerefZdd( dd, zSymmVars );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zPlus).ref);
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(zSymmVars).ref);

            // add these variable pairs to the result
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
        if nVarsExtra != 0 {
            // it is possible to improve this step:
            // (1) there is no need to enter here, if nVarsExtra < 2

            // create the set of topmost nVarsExtra in bVars
            var bVarsExtra *DdNode;
            var nVars int32;

            // remove from bVars all the variable that are in the support of bFunc
            bVarsExtra = extraBddReduceVarSet( dd, bVars, bFunc );  
            if bVarsExtra == nil {
                Cudd_RecursiveDerefZdd( dd, zRes );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(bVarsExtra).ref);

            // determine how many vars are in the bVarsExtra
            nVars = Extra_bddSuppSize( dd, bVarsExtra );
            if nVars < 2 {
                Cudd_RecursiveDeref( dd, bVarsExtra );
            
            } else {
                var i int32;
                var bVarsK *DdNode;

                // create the BDD bVarsK corresponding to K = 2;
                bVarsK = bVarsExtra;
                for  i = 0; i < nVars-2; i++  {
                    // inlining of macro cuddT
                    bVarsK = ((bVarsK).type.kids.T);
                }

                // create the 2 variable tuples
                zPlus = extraZddTuplesFromBdd( dd, bVarsK, bVarsExtra );
                if zPlus == nil {
                    Cudd_RecursiveDeref( dd, bVarsExtra );
                    Cudd_RecursiveDerefZdd( dd, zRes );
                    return nil;
                }
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(zPlus).ref);
                Cudd_RecursiveDeref( dd, bVarsExtra );

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
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(zRes).ref);


        /* insert the result into cache */
        cuddCacheInsert2(dd, extraZddSymmPairsCompute, bFunc, bVars, zRes);
        return zRes;
    }
} /* end of extraZddSymmPairsCompute */

/**Function********************************************************************

  Synopsis    [Performs a recursive step of Extra_zddGetSymmetricVars.]

  Description [Returns the set of ZDD singletons, containing those positive
  ZDD variables that correspond to BDD variables x, for which it is true 
  that bF(x=0) == bG(x=1).]

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * extraZddGetSymmetricVars( 
  DdManager * dd,    /* the DD manager */
  DdNode * bF,       /* the first function  - originally, the positive cofactor */
  DdNode * bG,       /* the second function - originally, the negative cofactor */
  DdNode * bVars)    /* the set of variables, on which F and G depend */
{
    var zRes *DdNode;
    // inlining of macro Cudd_Regular
    var bFR *DdNode  = ((DdNode *)((ptruint)(bF) & ~01)); 
    // inlining of macro Cudd_Regular
    var bGR *DdNode  = ((DdNode *)((ptruint)(bG) & ~01)); 

    // inlining of macro cuddIsConstant
    if  ((bFR).index == CUDD_CONST_INDEX) && ((bGR).index == CUDD_CONST_INDEX) {
        if bF == bG {
            return extraZddGetSingletons( dd, bVars );
        } else { 
            return z0; }
    }
    assert(bVars != b1, "bVars != b1");

    if zRes = cuddCacheLookupZdd(dd, DD_GET_SYMM_VARS_TAG, bF, bG, bVars) {
        return zRes;
    } else {
        var zRes0, zRes1 *DdNode;             
        var zPlus, zTemp *DdNode;
        var bF0, bF1 *DdNode;             
        var bG0, bG1 *DdNode;             
        var bVarsNew *DdNode;
    
        var LevelF int32;
// inlining of macro cuddI

        if (bFR.index)==CUDD_CONST_INDEX {
            LevelF = int32((bFR.index));
        } else {
            LevelF = (dd).perm[(bFR.index)];
        }
        var LevelG int32;
// inlining of macro cuddI

        if (bGR.index)==CUDD_CONST_INDEX {
            LevelG = int32((bGR.index));
        } else {
            LevelG = (dd).perm[(bGR.index)];
        }
        var LevelFG int32;

        if LevelF < LevelG {
            LevelFG = LevelF;
        } else {
            LevelFG = LevelG; }

        // at least one of the arguments is not a constant
        assert(LevelFG < dd.size, "LevelFG < dd.size");

        // every variable in bF and bG should be also in bVars, therefore LevelFG cannot be above LevelV
        // if LevelFG is below LevelV, scroll through the vars in bVars to the same level as LevelFG
        // inlining of macro cuddT
        for  bVarsNew = bVars; LevelFG > dd.perm[bVarsNew.index]; bVarsNew = ((bVarsNew).type.kids.T)  {}
        assert(LevelFG == dd.perm[bVarsNew.index], "LevelFG == dd.perm[bVarsNew.index]");

        // cofactor the functions
        if LevelF == LevelFG {
            if bFR != bF { // bF is complemented 
            
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
        
        } else { 
            bF0 = bF1 = bF; }

        if LevelG == LevelFG {
            if bGR != bG { // bG is complemented 
            
                // inlining of macro Cudd_Not
                // inlining of macro cuddE
                bG0 = ((DdNode *)((ptrint)(((bGR).type.kids.E)) ^ 01));
                // inlining of macro Cudd_Not
                // inlining of macro cuddT
                bG1 = ((DdNode *)((ptrint)(((bGR).type.kids.T)) ^ 01));
            
            } else {
                // inlining of macro cuddE
                bG0 = ((bGR).type.kids.E);
                // inlining of macro cuddT
                bG1 = ((bGR).type.kids.T);
            }
        
        } else { 
            bG0 = bG1 = bG; }

        // solve subproblems
        // inlining of macro cuddT
        zRes0 = extraZddGetSymmetricVars( dd, bF0, bG0, ((bVarsNew).type.kids.T) );
        if zRes0 == nil { 
            return nil; }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zRes0).ref);

        // if there is not symmetries in the negative cofactor
        // there is no need to test the positive cofactor
        if zRes0 == z0 {
            zRes = zRes0;  // zRes takes reference
        } else {
            // inlining of macro cuddT
            zRes1 = extraZddGetSymmetricVars( dd, bF1, bG1, ((bVarsNew).type.kids.T) );
            if zRes1 == nil {
                Cudd_RecursiveDerefZdd( dd, zRes0 );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zRes1).ref);

            // only those variables should belong to the resulting set 
            // for which the property is true for both cofactors
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

        // add one more singleton if the property is true for this variable
        if bF0 == bG1 {
            zPlus = cuddZddGetNode( dd, 2*bVarsNew.index, z1, z0 );
            if zPlus == nil {
                Cudd_RecursiveDerefZdd( dd, zRes );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zPlus).ref);

            // add these variable pairs to the result
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

        if bF == bG && bVars != bVarsNew { 
            // if the functions are equal, so are their cofactors
            // add those variables from V that are above F and G 

            var bVarsExtra *DdNode;

            assert(LevelFG > dd.perm[bVars.index], "LevelFG > dd.perm[bVars.index]");

            // create the BDD of the extra variables
            bVarsExtra = cuddBddExistAbstractRecur( dd, bVars, bVarsNew );
            if bVarsExtra == nil {
                Cudd_RecursiveDerefZdd( dd, zRes );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(bVarsExtra).ref);

            zPlus = extraZddGetSingletons( dd, bVarsExtra );
            if zPlus == nil {
                Cudd_RecursiveDeref( dd, bVarsExtra );
                Cudd_RecursiveDerefZdd( dd, zRes );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zPlus).ref);
            Cudd_RecursiveDeref( dd, bVarsExtra );

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

        cuddCacheInsert( dd, DD_GET_SYMM_VARS_TAG, bF, bG, bVars, zRes );
        return zRes;
    }
}   /* end of extraZddGetSymmetricVars */


/**Function********************************************************************

  Synopsis    [Performs a recursive step of Extra_zddGetSingletons.]

  Description [Returns the set of ZDD singletons, containing those positive 
  polarity ZDD variables that correspond to the BDD variables in bVars.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * extraZddGetSingletons( 
  DdManager * dd,    /* the DD manager */
  DdNode * bVars)    /* the set of variables */
{
    var zRes *DdNode;

    if bVars == b1 {
//    if ( bVars == b0 )  // bug fixed by Jin Zhang, Jan 23, 2004
        return z1; }

    if zRes = cuddCacheLookup1Zdd(dd, extraZddGetSingletons, bVars) {
        return zRes;
    } else {
        var zTemp, zPlus *DdNode;          

        // solve subproblem
        // inlining of macro cuddT
        zRes = extraZddGetSingletons( dd, ((bVars).type.kids.T) );
        if zRes == nil { 
            return nil; }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zRes).ref);
        
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

        cuddCacheInsert1( dd, extraZddGetSingletons, bVars, zRes );
        return zRes;
    }
}   /* end of extraZddGetSingletons */


/**Function********************************************************************

  Synopsis    [Performs a recursive step of Extra_bddReduceVarSet.]

  Description [Returns the set of all variables in the given set that are not in the 
  support of the given function.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * extraBddReduceVarSet( 
  DdManager * dd,    /* the DD manager */
  DdNode * bVars,    /* the set of variables to be reduced */
  DdNode * bF)       /* the function whose support is used for reduction */
{
    var bRes *DdNode;
    // inlining of macro Cudd_Regular
    var bFR *DdNode  = ((DdNode *)((ptruint)(bF) & ~01));

    // inlining of macro cuddIsConstant
    if  ((bFR).index == CUDD_CONST_INDEX) || bVars == b1 {
        return bVars; }

    if bRes = cuddCacheLookup2(dd, extraBddReduceVarSet, bVars, bF) {
        return bRes;
    } else {
        var bF0, bF1 *DdNode;             
        var bVarsThis, bVarsLower, bTemp *DdNode;
        var LevelF int32;

        // if LevelF is below LevelV, scroll through the vars in bVars 
        LevelF = dd.perm[bFR.index];
        // inlining of macro cuddI
        // inlining of macro cuddT
        for  bVarsThis = bVars; LevelF > (((bVarsThis.index)==CUDD_CONST_INDEX)?int32((bVarsThis.index)):(dd).perm[(bVarsThis.index)]); bVarsThis = ((bVarsThis).type.kids.T)  {}
        // scroll also through the current var, because it should be not be added
        // inlining of macro cuddI
        if LevelF == (((bVarsThis.index)==CUDD_CONST_INDEX)?int32((bVarsThis.index)):(dd).perm[(bVarsThis.index)]) {
            // inlining of macro cuddT
            bVarsLower = ((bVarsThis).type.kids.T);
        } else {
            bVarsLower = bVarsThis; }

        // cofactor the function
        if bFR != bF { // bFunc is complemented 
        
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
        bRes = extraBddReduceVarSet( dd, bVarsLower, bF0 );
        if bRes == nil { 
            return nil; }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(bRes).ref);

        bRes = extraBddReduceVarSet( dd, bTemp = bRes, bF1 );
        if bRes == nil {
            Cudd_RecursiveDeref( dd, bTemp );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(bRes).ref);
        Cudd_RecursiveDeref( dd, bTemp );

        // the current var should not be added
        // add the skipped vars
        if bVarsThis != bVars {
            var bVarsExtra *DdNode;

            // extract the skipped variables
            bVarsExtra = cuddBddExistAbstractRecur( dd, bVars, bVarsThis );
            if bVarsExtra == nil {
                Cudd_RecursiveDeref( dd, bRes );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(bVarsExtra).ref);

            // add these variables
            bRes = cuddBddAndRecur( dd, bTemp = bRes, bVarsExtra );
            if bRes == nil {
                Cudd_RecursiveDeref( dd, bTemp );
                Cudd_RecursiveDeref( dd, bVarsExtra );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(bRes).ref);
            Cudd_RecursiveDeref( dd, bTemp );
            Cudd_RecursiveDeref( dd, bVarsExtra );
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(bRes).ref);

        cuddCacheInsert2( dd, extraBddReduceVarSet, bVars, bF, bRes );
        return bRes;
    }
}   /* end of extraBddReduceVarSet */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Extra_bddCheckVarsSymmetric().]

  Description [Returns b0 if the variables are not symmetric. Returns b1 if the
  variables can be symmetric. The variables are represented in the form of a 
  two-variable cube. In case the cube contains one variable (below Var1 level),
  the cube's pointer is complemented if the variable Var1 occurred on the 
  current path; otherwise, the cube's pointer is regular. Uses additional 
  complemented bit (Hash_Not) to mark the result if in the BDD rooted that this
  node there is a branch passing though the node labeled with Var2.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * extraBddCheckVarsSymmetric( 
  DdManager * dd,    /* the DD manager */
  DdNode * bF,
  DdNode * bVars) 
{
    var bRes *DdNode;

    if bF == b0 {
        return b1; }

    assert(bVars != b1, "bVars != b1");
    
    if bRes = cuddCacheLookup2(dd, extraBddCheckVarsSymmetric, bF, bVars) {
        return bRes;
    } else {
        var bRes0, bRes1 *DdNode;
        var bF0, bF1 *DdNode;             
        // inlining of macro Cudd_Regular
        var bFR *DdNode  = ((DdNode *)((ptruint)(bF) & ~01));
        var LevelF int32;
// inlining of macro cuddI

        if (bFR.index)==CUDD_CONST_INDEX {
            LevelF = int32((bFR.index));
        } else {
            LevelF = (dd).perm[(bFR.index)];
        }

        // inlining of macro Cudd_Regular
        var bVarsR *DdNode  = ((DdNode *)((ptruint)(bVars) & ~01));
        var fVar1Pres int32;
        var iLev1 int32;
        var iLev2 int32;

        if bVarsR != bVars { // cube's pointer is complemented
        
            // inlining of macro cuddT
            assert(((bVarsR).type.kids.T) == b1, "/*__starting_point_inline_macro:cuddT__*/((bVarsR).type.kids.T) == b1");
            fVar1Pres = 1;                    // the first var is present on the path
            iLev1 = -1;                       // we are already below the first var level
            iLev2 = dd.perm[bVarsR.index];  // the level of the second var
        
        } else {  // cube's pointer is NOT complemented
        
            fVar1Pres = 0;                    // the first var is absent on the path
            // inlining of macro cuddT
            if  ((bVars).type.kids.T) == b1 {
                iLev1 = -1;                             // we are already below the first var level 
                iLev2 = dd.perm[bVars.index];         // the level of the second var
            
            } else {
                // inlining of macro cuddT
                assert(((((bVars).type.kids.T)).type.kids.T) == b1, "/*__starting_point_inline_macro:cuddT__*/((/*__starting_point_inline_macro:cuddT__*/((bVars).type.kids.T)).type.kids.T) == b1");
                iLev1 = dd.perm[bVars.index];         // the level of the first var 
                // inlining of macro cuddT
                iLev2 = dd.perm[((bVars).type.kids.T).index];  // the level of the second var
            }
        }

        // cofactor the function
        // the cofactors are needed only if we are above the second level
        if LevelF < iLev2 {
            if bFR != bF { // bFunc is complemented 
            
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
        
        } else {
            bF0 = bF1 = nil; }

        // consider five cases:
        // (1) F is above iLev1
        // (2) F is on the level iLev1
        // (3) F is between iLev1 and iLev2
        // (4) F is on the level iLev2
        // (5) F is below iLev2

        // (1) F is above iLev1
        if LevelF < iLev1 {
            // the returned result cannot have the hash attribute
            // because we still did not reach the level of Var1;
            // the attribute never travels above the level of Var1
            bRes0 = extraBddCheckVarsSymmetric( dd, bF0, bVars );
//          assert( !Hash_IsComplement( bRes0 ) );
            assert(bRes0 != z0, "bRes0 != z0");
            if bRes0 == b0 {
                bRes = b0;
            } else {
                bRes = extraBddCheckVarsSymmetric( dd, bF1, bVars ); }
//          assert( !Hash_IsComplement( bRes ) );
            assert(bRes != z0, "bRes != z0");
        
        // (2) F is on the level iLev1
        } else if LevelF == iLev1 {
            // inlining of macro Cudd_Not
            // inlining of macro cuddT
            bRes0 = extraBddCheckVarsSymmetric( dd, bF0, ((DdNode *)((ptrint)(((bVars).type.kids.T)) ^ 01)) );
            if bRes0 == b0 {
                bRes = b0;
            } else {
                // inlining of macro Cudd_Not
                // inlining of macro cuddT
                bRes1 = extraBddCheckVarsSymmetric( dd, bF1, ((DdNode *)((ptrint)(((bVars).type.kids.T)) ^ 01)) );
                if bRes1 == b0 {
                    bRes = b0;
                } else {
//                  if ( Hash_IsComplement( bRes0 ) || Hash_IsComplement( bRes1 ) )
                    if bRes0 == z0 || bRes1 == z0 {
                        bRes = b1;
                    } else {
                        bRes = b0; }
                }
            }
        
        // (3) F is between iLev1 and iLev2
        } else if LevelF < iLev2 {
            bRes0 = extraBddCheckVarsSymmetric( dd, bF0, bVars );
            if bRes0 == b0 {
                bRes = b0;
            } else {
                bRes1 = extraBddCheckVarsSymmetric( dd, bF1, bVars );
                if bRes1 == b0 {
                    bRes = b0;
                } else {
//                  if ( Hash_IsComplement( bRes0 ) || Hash_IsComplement( bRes1 ) )
//                      bRes = Hash_Not( b1 );
                    if bRes0 == z0 || bRes1 == z0 {
                        bRes = z0;
                    } else {
                        bRes = b1; }
                }
            }
        
        // (4) F is on the level iLev2
        } else if LevelF == iLev2 {
            // this is the only place where the hash attribute (Hash_Not) can be added 
            // to the result; it can be added only if the path came through the node
            // lebeled with Var1; therefore, the hash attribute cannot be returned
            // to the caller function
            if fVar1Pres != 0 {
//              bRes = Hash_Not( b1 );
                bRes = z0;
            } else { 
                bRes = b0; }
        
        // (5) F is below iLev2
        } else { // if ( LevelF > iLev2 )
        
            // it is possible that the path goes through the node labeled by Var1
            // and still everything is okay; we do not label with Hash_Not here
            // because the path does not go through node labeled by Var2
            bRes = b1;
        }

        cuddCacheInsert2(dd, extraBddCheckVarsSymmetric, bF, bVars, bRes);
        return bRes;
    }
} /* end of extraBddCheckVarsSymmetric */

/**Function********************************************************************

  Synopsis    [Performs the reordering-sensitive step of Extra_zddTupleFromBdd().]

  Description [Generates in a bottom-up fashion ZDD for all combinations
               composed of k variables out of variables belonging to Support.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode* extraZddTuplesFromBdd( 
  DdManager * dd,   /* the DD manager */
  DdNode * bVarsK,   /* the number of variables in tuples */
  DdNode * bVarsN)   /* the set of all variables */
{
    var zRes, zRes0, zRes1 *DdNode;
    statLine(dd); 

    /* terminal cases */
/*  if ( k < 0 || k > n )
 *      return dd->zero;
 *  if ( n == 0 )
 *      return dd->one; 
 */
    // inlining of macro cuddI
    if  (((bVarsK.index)==CUDD_CONST_INDEX)?int32((bVarsK.index)):(dd).perm[(bVarsK.index)]) < (((bVarsN.index)==CUDD_CONST_INDEX)?int32((bVarsN.index)):(dd).perm[(bVarsN.index)]) {
        return z0; }
    if bVarsN == b1 {
        return z1; }

    /* check cache */
    zRes = cuddCacheLookup2Zdd(dd, extraZddTuplesFromBdd, bVarsK, bVarsN);
    if zRes != nil {
        return(zRes); }

    /* ZDD in which this variable is 0 */
/*  zRes0 = extraZddTuplesFromBdd( dd, k,     n-1 ); */
    // inlining of macro cuddT
    zRes0 = extraZddTuplesFromBdd( dd, bVarsK, ((bVarsN).type.kids.T) );
    if zRes0 == nil { 
        return nil; }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(zRes0).ref);

    /* ZDD in which this variable is 1 */
/*  zRes1 = extraZddTuplesFromBdd( dd, k-1,          n-1 ); */
    if bVarsK == b1 {
        zRes1 = z0;
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zRes1).ref);
    
    } else {
        // inlining of macro cuddT
        zRes1 = extraZddTuplesFromBdd( dd, ((bVarsK).type.kids.T), ((bVarsN).type.kids.T) );
        if zRes1 == nil {
            Cudd_RecursiveDerefZdd( dd, zRes0 );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zRes1).ref);
    }

    /* compose Res0 and Res1 with the given ZDD variable */
    zRes = cuddZddGetNode( dd, 2*bVarsN.index, zRes1, zRes0 );
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
    cuddCacheInsert2(dd, extraZddTuplesFromBdd, bVarsK, bVarsN, zRes);
    return zRes;

} /* end of extraZddTuplesFromBdd */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Extra_zddSelectOneSubset.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
DdNode * extraZddSelectOneSubset( 
  DdManager * dd, 
  DdNode * zS )
// selects one subset from the ZDD zS
// returns z0 if and only if zS is an empty set of cubes
{
    var zRes *DdNode;

    if zS == z0 {    return z0; }
    if zS == z1 {    return z1; }
    
    // check cache
    if zRes = cuddCacheLookup1Zdd( dd, extraZddSelectOneSubset, zS ) {
        return zRes;
    } else {
        var zS0, zS1, zTemp *DdNode; 

        // inlining of macro cuddE
        zS0 = ((zS).type.kids.E);
        // inlining of macro cuddT
        zS1 = ((zS).type.kids.T);

        if zS0 != z0 {
            zRes = extraZddSelectOneSubset( dd, zS0 );
            if zRes == nil {
                return nil; }
        
        } else { // if ( zS0 == z0 )
        
            assert(zS1 != z0, "zS1 != z0");
            zRes = extraZddSelectOneSubset( dd, zS1 );
            if zRes == nil {
                return nil; }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zRes).ref);

            zRes = cuddZddGetNode( dd, zS.index, zTemp = zRes, z0 );
            if zRes == nil {
                Cudd_RecursiveDerefZdd( dd, zTemp );
                return nil;
            }
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(zTemp).ref);
        }

        // insert the result into cache
        cuddCacheInsert1( dd, extraZddSelectOneSubset, zS, zRes );
        return zRes;
    }       
} /* end of extraZddSelectOneSubset */


/*---------------------------------------------------------------------------*/
/* Definition of static Functions                                            */
/*---------------------------------------------------------------------------*/
//ABC_NAMESPACE_IMPL_END

