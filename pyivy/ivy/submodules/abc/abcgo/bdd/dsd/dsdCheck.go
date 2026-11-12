package dsd

/**CFile****************************************************************

  FileName    [dsdCheck.c]

  PackageName [DSD: Disjoint-support decomposition package.]

  Synopsis    [Procedures to check the identity of root functions.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 8.0. Started - September 22, 2003.]

  Revision    [$Id: dsdCheck.c,v 1.0 2002/22/09 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dsdInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Dds_Cache_t =  Dsd_Cache_t_;
type Dsd_Entry_t =  Dsd_Entry_t_;

type Dsd_Cache_t_ struct {
    pTable      *Dsd_Entry_t;
    nTableSize               int32;
    nSuccess               int32;
    nFailure               int32;
}

type Dsd_Entry_t_ struct {
    DdNode * bX[5];
}

var pCache *Dds_Cache_t;

// static int Dsd_CheckRootFunctionIdentity_rec( DdManager * dd, DdNode * bF1, DdNode * bF2, DdNode * bC1, DdNode * bC2 );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function********************************************************************

  Synopsis    [(Re)allocates the local cache.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Dsd_CheckCacheAllocate(nEntries int32) {
    var nRequested int32;

    pCache = new(Dds_Cache_t);
    *pCache = Dds_Cache_t{};

    // check what is the size of the current cache
    nRequested = Abc_PrimeCudd( nEntries );
    if pCache.nTableSize != nRequested { // the current size is different
        // deallocate the old, allocate the new
        if pCache.nTableSize {
            Dsd_CheckCacheDeallocate(); }
        // allocate memory for the hash table
        pCache.nTableSize = nRequested;
        pCache.pTable = ABC_ALLOC( Dsd_Entry_t, nRequested );
    }
    // otherwise, there is no need to allocate, just clean
    Dsd_CheckCacheClear();
//  printf( "\nThe number of allocated cache entries = %d.\n\n", pCache->nTableSize );
}

/**Function********************************************************************

  Synopsis    [Deallocates the local cache.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Dsd_CheckCacheDeallocate() {
    ABC_FREE( pCache.pTable );
    ABC_FREE( pCache );
}

/**Function********************************************************************

  Synopsis    [Clears the local cache.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Dsd_CheckCacheClear() {
    var i int32;
    for  i = 0; i < pCache.nTableSize; i++  {
        pCache.pTable[0].bX[0] = nil;
    }
}


/**Function********************************************************************

  Synopsis    [Checks whether it is true that bF1(bC1=0) == bF2(bC2=0).]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Dsd_CheckRootFunctionIdentity(dd *DdManager, bF1 *DdNode, bF2 *DdNode, bC1 *DdNode, bC2 *DdNode) int32 {
    var RetValue int32;
//  pCache->nSuccess = 0;
//  pCache->nFailure = 0;
    RetValue = Dsd_CheckRootFunctionIdentity_rec(dd, bF1, bF2, bC1, bC2);
//  printf( "Cache success = %d. Cache failure = %d.\n", pCache->nSuccess, pCache->nFailure );
    return RetValue;
}

/**Function********************************************************************

  Synopsis    [Performs the recursive step of Dsd_CheckRootFunctionIdentity().]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Dsd_CheckRootFunctionIdentity_rec(dd *DdManager, bF1 *DdNode, bF2 *DdNode, bC1 *DdNode, bC2 *DdNode) int32 {
    var HKey uint32;

    // if either bC1 or bC2 is zero, the test is true
//  if ( bC1 == b0 || bC2 == b0 )  return 1;
    assert(bC1 != b0, "bC1 != b0");
    assert(bC2 != b0, "bC2 != b0");

    // if both bC1 and bC2 are one - perform comparison
    if bC1 == b1 && bC2 == b1 {  return int32(( bF1 == bF2 )); }

    if bF1 == b0 {
        // inlining of macro Cudd_Not
        return Cudd_bddLeq( dd, bC2, ((DdNode *)((ptrint)(bF2) ^ 01)) ); }

    if bF1 == b1 {
        return Cudd_bddLeq( dd, bC2, bF2 ); }

    if bF2 == b0 {
        // inlining of macro Cudd_Not
        return Cudd_bddLeq( dd, bC1, ((DdNode *)((ptrint)(bF1) ^ 01)) ); }

    if bF2 == b1 {
        return Cudd_bddLeq( dd, bC1, bF1 ); }

    // otherwise, keep expanding

    // check cache
//  HKey = _Hash( ((unsigned)bF1), ((unsigned)bF2), ((unsigned)bC1), ((unsigned)bC2) );
    // inlining of macro hashKey4
    HKey = ((((((ABC_PTRUINT_T)(bF1) + (ABC_PTRUINT_T)(bF2)) * DD_P1 + (ABC_PTRUINT_T)(bC1)) * DD_P2 +    (ABC_PTRUINT_T)(bC2)) * DD_P3) % pCache.nTableSize);
    if pCache.pTable[HKey].bX[0] == bF1 && 
         pCache.pTable[HKey].bX[1] == bF2 && 
         pCache.pTable[HKey].bX[2] == bC1 && 
         pCache.pTable[HKey].bX[3] == bC2 {
        pCache.nSuccess++;
        return int32((ABC_PTRUINT_T))pCache.pTable[HKey].bX[4]; // the last bit records the result (yes/no)
    
    } else {

        // determine the top variables
        var RetValue int32;
        var bA [4]*DdNode = [4]*DdNode{ bF1, bF2, bC1, bC2 }; // arguments
        // inlining of macro Cudd_Regular
        DdNode * bAR[4] = { ((DdNode *)((ptruint)(bF1) & ~01)), ((DdNode *)((ptruint)(bF2) & ~01)), ((DdNode *)((ptruint)(bC1) & ~01)), ((DdNode *)((ptruint)(bC2) & ~01)) }; // regular arguments
        // inlining of macro cuddI
        int CurLevel[4] = { (((bAR[0].index)==CUDD_CONST_INDEX)?int32((bAR[0].index)):(dd).perm[(bAR[0].index)]), (((bAR[1].index)==CUDD_CONST_INDEX)?int32((bAR[1].index)):(dd).perm[(bAR[1].index)]), (((bAR[2].index)==CUDD_CONST_INDEX)?int32((bAR[2].index)):(dd).perm[(bAR[2].index)]), (((bAR[3].index)==CUDD_CONST_INDEX)?int32((bAR[3].index)):(dd).perm[(bAR[3].index)]) };
        var TopLevel int32  = CUDD_CONST_INDEX;
        var i int32;
        var bE, bT [4]*DdNode;
        var bF1next, bF2next, bC1next, bC2next *DdNode;

        pCache.nFailure++;

        // determine the top level
        for  i = 0; i < 4; i++  {
            if TopLevel > CurLevel[i] {
                 TopLevel = CurLevel[i]; }
        }

        // compute the cofactors
        for  i = 0; i < 4; i++  {
        if TopLevel == CurLevel[i] {
            if bA[i] != bAR[i] { // complemented
            
                // inlining of macro Cudd_Not
                // inlining of macro cuddE
                bE[i] = ((DdNode *)((ptrint)(((bAR[i]).type.kids.E)) ^ 01));
                // inlining of macro Cudd_Not
                // inlining of macro cuddT
                bT[i] = ((DdNode *)((ptrint)(((bAR[i]).type.kids.T)) ^ 01));
            
            } else {
                // inlining of macro cuddE
                bE[i] = ((bAR[i]).type.kids.E);
                // inlining of macro cuddT
                bT[i] = ((bAR[i]).type.kids.T);
            }
        
        } else {
            bE[i] = bT[i] = bA[i]; }
        }

        // solve subproblems
        // three cases are possible

        // (1) the top var belongs to both C1 and C2
        //     in this case, any cofactor of F1 and F2 will do, 
        //     as long as the corresponding cofactor of C1 and C2 is not equal to 0
        if TopLevel == CurLevel[2] && TopLevel == CurLevel[3] {
            if bE[2] != b0 { // C1
            
                bF1next = bE[0];
                bC1next = bE[2];
            
            } else {
                bF1next = bT[0];
                bC1next = bT[2];
            }
            if bE[3] != b0 { // C2
            
                bF2next = bE[1];
                bC2next = bE[3];
            
            } else {
                bF2next = bT[1];
                bC2next = bT[3];
            }
            RetValue = Dsd_CheckRootFunctionIdentity_rec( dd, bF1next, bF2next, bC1next, bC2next );
        
        // (2) the top var belongs to either C1 or C2
        //     in this case normal splitting of cofactors
        } else if TopLevel == CurLevel[2] && TopLevel != CurLevel[3] {
            if bE[2] != b0 { // C1
            
                bF1next = bE[0];
                bC1next = bE[2];
            
            } else {
                bF1next = bT[0];
                bC1next = bT[2];
            }
            // split around this variable
            RetValue = Dsd_CheckRootFunctionIdentity_rec( dd, bF1next, bE[1], bC1next, bE[3] );
            if RetValue == 1 { // test another branch; otherwise, there is no need to test
                RetValue = Dsd_CheckRootFunctionIdentity_rec( dd, bF1next, bT[1], bC1next, bT[3] ); }
        
        } else if TopLevel != CurLevel[2] && TopLevel == CurLevel[3] {
            if bE[3] != b0 { // C2
            
                bF2next = bE[1];
                bC2next = bE[3];
            
            } else {
                bF2next = bT[1];
                bC2next = bT[3];
            }
            // split around this variable
            RetValue = Dsd_CheckRootFunctionIdentity_rec( dd, bE[0], bF2next, bE[2], bC2next );
            if RetValue == 1 { // test another branch; otherwise, there is no need to test
                RetValue = Dsd_CheckRootFunctionIdentity_rec( dd, bT[0], bF2next, bT[2], bC2next ); }
        
        // (3) the top var does not belong to C1 and C2
        //     in this case normal splitting of cofactors
        } else { // if ( TopLevel != CurLevel[2] && TopLevel != CurLevel[3] )
        
            // split around this variable
            RetValue = Dsd_CheckRootFunctionIdentity_rec( dd, bE[0], bE[1], bE[2], bE[3] );
            if RetValue == 1 { // test another branch; otherwise, there is no need to test
                RetValue = Dsd_CheckRootFunctionIdentity_rec( dd, bT[0], bT[1], bT[2], bT[3] ); }
        }

        // set cache
        for  i = 0; i < 4; i++  {
            pCache.pTable[HKey].bX[i] = bA[i];
        }
        pCache.pTable[HKey].bX[4] = (DdNode*)(ABC_PTRUINT_T)RetValue;

        return RetValue;
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

