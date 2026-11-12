package reo

/**CFile****************************************************************

  FileName    [reo.h]

  PackageName [REO: A specialized DD reordering engine.]

  Synopsis    [External and internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 15, 2002.]

  Revision    [$Id: reo.h,v 1.0 2002/15/10 03:00:00 alanmi Exp $]

***********************************************************************/





//#include <stdio.h>
//#include <stdlib.h>
//#include "bdd/extrab/extraBdd.h"

////////////////////////////////////////////////////////////////////////
///                     MACRO DEFINITIONS                            ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


// reordering parameters
const REO_REORDER_LIMIT =      1.15  // determines the quality/runtime trade-off
const REO_QUAL_PAR =              3  // the quality [1 = simple lower bound, 2 = strict, larger = heuristic]
// internal parameters
const REO_CONST_LEVEL =       30000  // the number of the constant level
const REO_TOPREF_UNDEF =      30000  // the undefined top reference
const REO_CHUNK_SIZE =         5000  // the number of units allocated at one time
const REO_COST_EPSILON =  0.0000001  // difference in cost large enough so that it counted as an error
const REO_HIGH_VALUE =     10000000  // a large value used to initialize some variables
// interface parameters
const REO_ENABLE =                1  // the value of the enable flag
const REO_DISABLE =               0  // the value of the disable flag

// the types of minimization currently supported
 type reo_min_type int32
const (
    REO_MINIMIZE_NODES reo_min_type = iota

    REO_MINIMIZE_WIDTH
   // may not work for BDDs with complemented edges
    REO_MINIMIZE_APL

)

 

////////////////////////////////////////////////////////////////////////
///                      DATA STRUCTURES                             ///
////////////////////////////////////////////////////////////////////////

type reo_unit =     _reo_unit;    // the unit representing one DD node during reordering
type reo_plane =    _reo_plane;   // the set of nodes on one level
type reo_hash =     _reo_hash;    // the entry in the hash table
type reo_man =      _reo_man;     // the reordering manager
type reo_test =     _reo_test;    // 

type _reo_unit struct {
    lev       int16;             // the level of this node at the beginning
    TopRef       int16;          // the top level from which this node is refed (used to update BDD width)
    TopRefNew       int16;       // the new top level from which this node is refed (used to update BDD width)
    n       int16;               // the number of incoming edges (similar to ref count in the BDD)
    Sign         int32;            // the signature

    pE   *reo_unit;              // the pointer to the "else" branch
    pT   *reo_unit;              // the pointer to the "then" branch
    Next   *reo_unit;            // the link to the next one in the list
    Weight      float64;          // the probability of traversing this node
}

type _reo_plane struct {
    fSifted         int32;         // to mark the sifted variables
    statsNodes         int32;      // the number of nodes in the current level
    statsWidth         int32;      // the width on the current level
    statsApl      float64;        // the sum of node probabilities on this level
    statsCost      float64;       // the current cost is stored here
    statsCostAbove      float64;  // the current cost is stored here
    statsCostBelow      float64;  // the current cost is stored here

    pHead   *reo_unit;           // the pointer to the beginning of the unit list
}

type _reo_hash struct {
    Sign         int32;            // signature of the current cache operation
    Arg1   *reo_unit;            // the first argument
    Arg2   *reo_unit;            // the second argument
    Arg3   *reo_unit;            // the third argument
}

type _reo_man struct {
    // these paramaters can be set by the API functions
    fMinWidth         int32;       // the flag to enable reordering for minimum width
    fMinApl         int32;         // the flag to enable reordering for minimum APL
    fVerbose         int32;        // the verbosity level
    fVerify         int32;         // the flag toggling verification
    fRemapUp         int32;        // the flag to enable remapping   
    nIters         int32;          // the number of iterations of sifting to perform

    // parameters given by the user when reordering is called
    dd  *DdManager;              // the CUDD BDD manager
    pOrder        *int32;          // the resulting variable order will be returned here

    // derived parameters
    fThisIsAdd         int32;      // this flag is one if the function is the ADD 
    pSupp        *int32;           // the support of the given function
    nSuppAlloc         int32;      // the max allowed number of support variables
    nSupp         int32;           // the number of support variables
    pOrderInt        *int32;       // the array storing the internal variable permutation
    pVarCosts     *float64;       // other arrays
    pLevelOrder        *int32;     // other arrays
    pWidthCofs  **reo_unit;      // temporary storage for cofactors used during reordering for width

    // parameters related to cost
    nNodesBeg         int32;
    nNodesCur         int32;
    nNodesEnd         int32;
    nWidthCur         int32;
    nWidthBeg         int32;
    nWidthEnd         int32;
    nAplCur      float64;
    nAplBeg      float64;
    nAplEnd      float64;

    // mapping of the function into planes and back
    pMapToPlanes        *int32;    // the mapping of var indexes into plane levels
    pMapToDdVarsOrig        *int32;// the mapping of plane levels into the original indexes
    pMapToDdVarsFinal        *int32;// the mapping of plane levels into the final indexes

    // the planes table
    pPlanes  *reo_plane;
    nPlanes         int32;
    pTops  **reo_unit;
    nTops         int32;
    nTopsAlloc         int32;

    // the hash table
    HTable   *reo_hash;           // the table itself
    nTableSize         int32;       // the size of the hash table
    Signature         int32;        // the signature counter

    // the referenced node list
    nNodesMaxAlloc         int32;   // this parameters determins how much memory is allocated
    pRefNodes    **DdNode;
    nRefNodes         int32;
    nRefNodesAlloc         int32;

    // unit memory management
    pUnitFreeList   *reo_unit;
    pMemChunks  **reo_unit;
    nMemChunks         int32;
    nMemChunksAlloc         int32;
    nUnitsUsed         int32;

    // statistic variables
    HashSuccess         int32;
    HashFailure         int32;
    nSwaps         int32;            // the number of swaps
    nNISwaps         int32;          // the number of swaps without interaction
}

// used to manipulate units
// Original macro:
// #define Unit_Regular(u)     ((reo_unit *)((ABC_PTRUINT_T)(u) & ~01))

// Original macro:
// #define Unit_Not(u)         ((reo_unit *)((ABC_PTRUINT_T)(u) ^ 01))

// Original macro:
// #define Unit_NotCond(u,c)   ((reo_unit *)((ABC_PTRUINT_T)(u) ^ (c)))

// Original macro:
// #define Unit_IsConstant(u)  ((int)((u)->lev == REO_CONST_LEVEL))


////////////////////////////////////////////////////////////////////////
///                   FUNCTION DECLARATIONS                          ///
////////////////////////////////////////////////////////////////////////

// ======================= reoApi.c ========================================
// extern reo_man *  Extra_ReorderInit( int nDdVarsMax, int nNodesMax );

// extern void       Extra_ReorderQuit( reo_man * p );

// extern void       Extra_ReorderSetMinimizationType( reo_man * p, reo_min_type fMinType );

// extern void       Extra_ReorderSetRemapping( reo_man * p, int fRemapUp );

// extern void       Extra_ReorderSetIterations( reo_man * p, int nIters );

// extern void       Extra_ReorderSetVerbosity( reo_man * p, int fVerbose );

// extern void       Extra_ReorderSetVerification( reo_man * p, int fVerify );

// extern DdNode *   Extra_Reorder( reo_man * p, DdManager * dd, DdNode * Func, int * pOrder );

// extern void       Extra_ReorderArray( reo_man * p, DdManager * dd, DdNode * Funcs[], DdNode * FuncsRes[], int nFuncs, int * pOrder );

// ======================= reoCore.c =======================================
// extern void       reoReorderArray( reo_man * p, DdManager * dd, DdNode * Funcs[], DdNode * FuncsRes[], int nFuncs, int * pOrder );

// extern void       reoResizeStructures( reo_man * p, int nDdVarsMax, int nNodesMax, int nFuncs );

// ======================= reoProfile.c ======================================
// extern void       reoProfileNodesStart( reo_man * p );

// extern void       reoProfileAplStart( reo_man * p );

// extern void       reoProfileWidthStart( reo_man * p );

// extern void       reoProfileWidthStart2( reo_man * p );

// extern void       reoProfileAplPrint( reo_man * p );

// extern void       reoProfileNodesPrint( reo_man * p );

// extern void       reoProfileWidthPrint( reo_man * p );

// extern void       reoProfileWidthVerifyLevel( reo_plane * pPlane, int Level );

// ======================= reoSift.c =======================================
// extern void       reoReorderSift( reo_man * p );

// ======================= reoSwap.c =======================================
// extern double     reoReorderSwapAdjacentVars( reo_man * p, int Level, int fMovingUp );

// ======================= reoTransfer.c ===================================
// extern reo_unit * reoTransferNodesToUnits_rec( reo_man * p, DdNode * F );

// extern DdNode *   reoTransferUnitsToNodes_rec( reo_man * p, reo_unit * pUnit );

// ======================= reoUnits.c ======================================
// extern reo_unit * reoUnitsGetNextUnit(reo_man * p );

// extern void       reoUnitsRecycleUnit( reo_man * p, reo_unit * pUnit );

// extern void       reoUnitsRecycleUnitList( reo_man * p, reo_plane * pPlane );

// extern void       reoUnitsAddUnitToPlane( reo_plane * pPlane, reo_unit * pUnit );

// extern void       reoUnitsStopDispenser( reo_man * p );

// ======================= reoTest.c =======================================
// extern void       Extra_ReorderTest( DdManager * dd, DdNode * Func );

// extern DdNode *   Extra_ReorderCudd( DdManager * dd, DdNode * aFunc, int pPermuteReo[] );

// extern int        Extra_bddReorderTest( DdManager * dd, DdNode * bF );
 
// extern int        Extra_addReorderTest( DdManager * dd, DdNode * aF );
 



//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                         END OF FILE                              ///
////////////////////////////////////////////////////////////////////////
