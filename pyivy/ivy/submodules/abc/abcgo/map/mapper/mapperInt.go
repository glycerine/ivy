package mapper

/**CFile****************************************************************

  FileName    [mapperInt.h]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperInt.h,v 1.8 2004/09/30 21:18:10 satrajit Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
 
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <float.h>
//#include <math.h>

//#include "base/main/main.h"
//#include "map/mio/mio.h"
//#include "mapper.h"

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

// uncomment to have fanouts represented in the mapping graph
//#define MAP_ALLOCATE_FANOUT  1

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

// the bit masks
// Original macro:
// #define MAP_MASK(n)             ((~((unsigned)0)) >> (32-(n)))

#define MAP_FULL                 (~((unsigned)0))
#define MAP_NO_VAR               (-9999.0)

// maximum/minimum operators
// Original macro:
// #define MAP_MIN(a,b)             (((a) < (b))? (a) : (b))

// Original macro:
// #define MAP_MAX(a,b)             (((a) > (b))? (a) : (b))


// the small and large numbers (min/max float are 1.17e-38/3.40e+38)
#define MAP_FLOAT_LARGE          ((float)(FLT_MAX/10))
#define MAP_FLOAT_SMALL          ((float)1.0e-03)

// generating random unsigned (#define RAND_MAX 0x7fff)
#define MAP_RANDOM_UNSIGNED   ((((unsigned)rand()) << 24) ^ (((unsigned)rand()) << 12) ^ ((unsigned)rand()))

// internal macros to work with cuts
// Original macro:
// #define Map_CutIsComplement(p)  (((int)((ABC_PTRUINT_T) (p) & 01)))

// Original macro:
// #define Map_CutRegular(p)       ((Map_Cut_t *)((ABC_PTRUINT_T)(p) & ~01)) 

// Original macro:
// #define Map_CutNot(p)           ((Map_Cut_t *)((ABC_PTRUINT_T)(p) ^ 01)) 

// Original macro:
// #define Map_CutNotCond(p,c)     ((Map_Cut_t *)((ABC_PTRUINT_T)(p) ^ (c)))


// internal macros for referencing of nodes
// Original macro:
// #define Map_NodeReadRef(p)      ((Map_Regular(p))->nRefs)

// Original macro:
// #define Map_NodeRef(p)          ((Map_Regular(p))->nRefs++)


// macros to get hold of the bits in the support info
// Original macro:
// #define Map_InfoSetVar(p,i)     (p[(i)>>5] |=  (1<<((i) & 31)))

// Original macro:
// #define Map_InfoRemVar(p,i)     (p[(i)>>5] &= ~(1<<((i) & 31)))

// Original macro:
// #define Map_InfoFlipVar(p,i)    (p[(i)>>5] ^=  (1<<((i) & 31)))

// Original macro:
// #define Map_InfoReadVar(p,i)   ((p[(i)>>5]  &  (1<<((i) & 31))) > 0)


// returns the complemented attribute of the node
// Original macro:
// #define Map_NodeIsSimComplement(p) (Map_IsComplement(p)? !(Map_Regular(p)->fInv) : (p)->fInv)


////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

// the mapping manager
type Map_ManStruct_t_ struct {
    // the mapping graph
    pBins        **Map_Node_t;         // the table of nodes hashed by their children
    nBins                 int32;         // the size of the table
    pInputs        **Map_Node_t;       // the array of inputs
    nInputs                 int32;       // the number of inputs
    pOutputs        **Map_Node_t;      // the array of outputs
    nOutputs                 int32;      // the number of outputs
    nNodes                 int32;        // the total number of nodes
    pConst1         *Map_Node_t;       // the constant 1 node
    vMapObjs      *Map_NodeVec_t;      // the array of all nodes
    vMapBufs      *Map_NodeVec_t;      // the array of all nodes
    pNodeDelays              *float32;   // the array of node delays

    // info about the original circuit
    ppOutputNames              *string; // the primary output names
    pInputArrivals         *Map_Time_t;// the PI arrival times
    pOutputRequireds         *Map_Time_t;// the PI arrival times

    // mapping parameters
    nVarsMax                 int32;      // the max number of variables
    fAreaRecovery                 int32; // the flag to enable area recovery
    fVerbose                 int32;      // the verbosiness flag
    fMappingMode                 int32;  // set to 1 when doing area
    fRequiredGlo               float32;  // the global required times
    fEpsilon               float32;      // the epsilon used to compare floats
    AreaBase               float32;      // the area after delay-oriented mapping
    AreaFinal               float32;     // the area after delay-oriented mapping
    nIterations                 int32;   // How many matching passes to do
    fObeyFanoutLimits                 int32;// Should mapper try to obey fanout limits or not
    DelayTarget               float32;   // the required times set by the user
    nTravIds                 int32;      // the traversal counter
    fSwitching                 int32;    // use switching activity
    fSkipFanout                 int32;   // skip large gates when mapping high-fanout nodes
    fUseProfile                 int32;   // use standard-cell profile

    // the supergate library
    pSuperLib     *Map_SuperLib_t;     // the current supergate library
    uTruths [6][2]uint32; // the elementary truth tables
    uTruthsLarge [10][32]uint32; // the elementary truth tables
    nCounts [32]int32;   // the counter of minterms
    nCountsBest [32]int32;// the counter of minterms
    vVisited      *Map_NodeVec_t;      // the visited cuts during cut computation

    // the memory managers
    mmNodes    *Extra_MmFixed_t;       // the memory manager for nodes
    mmCuts    *Extra_MmFixed_t;        // the memory manager for cuts

    // precomputed N-canonical forms
    unsigned short *    uCanons;       // N-canonical forms
    uPhases              *string;       // N-canonical phases
    pCounters               string;     // counters of phases

    // various statistical variables
    nChoiceNodes                 int32;  // the number of choice nodes
    nChoices                 int32;      // the number of all choices
    nCanons                 int32;       // the number of times N-canonical form was computed
    nMatches                 int32;      // the number of times supergate matching was performed
    nPhases                 int32;       // the number of phases considered during matching
    nFanoutViolations                 int32;  // the number of nodes in mapped circuit violating fanout

    // runtime statistics
    timeToMap             abctime;     // time to transfer to the mapping structure
    timeCuts             abctime;      // time to compute k-feasible cuts
    timeTruth             abctime;     // time to compute the truth table for each cut
    timeMatch             abctime;     // time to perform matching for each node
    timeArea             abctime;      // time to recover area after delay oriented mapping
    timeSweep             abctime;     // time to perform technology dependent sweep
    timeToNet             abctime;     // time to transfer back to the network
    timeTotal             abctime;     // the total mapping time
    time1             abctime;         // time to transfer to the mapping structure
    time2             abctime;         // time to transfer to the mapping structure
    time3             abctime;         // time to transfer to the mapping structure
}

// the supergate library
type Map_SuperLibStruct_t_ struct {
    // general info
    pName               string;         // the name of the supergate library
    pGenlib      *Mio_Library_t;       // the generic library

    // other info
    nVarsMax                 int32;      // the max number of variables
    nSupersAll                 int32;    // the total number of supergates
    nSupersReal                 int32;   // the total number of supergates
    nLines                 int32;        // the total number of lines in the supergate file
    fVerbose                 int32;      // the verbosity flag

    // hash tables
    ppSupers       **Map_Super_t;      // the array of supergates
    tTableC    *Map_HashTable_t;       // the table mapping N-canonical forms into supergates
    tTable    *Map_HashTable_t;        // the table mapping truth tables into supergates

    // data structures for N-canonical form computation
    uTruths [6][2]uint32; // the elementary truth tables
    uMask [2]uint32;      // the mask for the truth table

    // the invertor
    pGateInv         *Mio_Gate_t;      // the pointer to the intertor gate
    tDelayInv          Map_Time_t;     // the delay of the inverter
    AreaInv               float32;       // the area of the inverter
    AreaBuf               float32;       // the area of the buffer
    pSuperInv        *Map_Super_t;     // the supergate representing the inverter

    // the memory manager for the internal table
    mmSupers    *Extra_MmFixed_t;      // the mamory manager for supergates
    mmEntries    *Extra_MmFixed_t;     // the memory manager for the entries
    mmForms     *Extra_MmFlex_t;       // the memory manager for formulas
}

// the mapping node
type Map_NodeStruct_t_ struct {
    // general information about the node
    p          *Map_Man_t;             // the mapping manager
    pNext         *Map_Node_t;         // the next node in the hash table
    Num                 int32;           // the unique number of this node
    TravId                 int32;        // the traversal ID (use to avoid cleaning marks)
    nRefs                 int32;         // the number of references (fanouts) of the given node
    fMark0 uint32; /* C bitfield: 1 bits; review packing and truncation */    // the mark used for traversals
    fMark1 uint32; /* C bitfield: 1 bits; review packing and truncation */    // the mark used for traversals
    fUsed uint32; /* C bitfield: 1 bits; review packing and truncation */    // the mark to mark the node or its fanins
    fInv uint32; /* C bitfield: 1 bits; review packing and truncation */    // the complemented attribute for the equivalent nodes
    fInvert uint32; /* C bitfield: 1 bits; review packing and truncation */    // the flag to denote the use of interter
    Level uint32; /* C bitfield: 16 bits; review packing and truncation */    // the level of the given node
    NumTemp uint32; /* C bitfield: 10 bits; review packing and truncation */    // the level of the given node
    nRefAct [3]int32;    // estimated fanout for current covering phase, neg and pos and sum
    nRefEst [3]float32;    // actual fanout for previous covering phase, neg and pos and sum
    Switching               float32;     // the probability of switching

    // connectivity
    p1         *Map_Node_t;            // the first child
    p2         *Map_Node_t;            // the second child
    pNextE         *Map_Node_t;        // the next functionally equivalent node
    pRepr         *Map_Node_t;         // the representative of the functionally equivalent class

#ifdef MAP_ALLOCATE_FANOUT
    // representation of node's fanouts
    Map_Node_t *        pFanPivot;     // the first fanout of this node
    Map_Node_t *        pFanFanin1;    // the next fanout of p1
    Map_Node_t *        pFanFanin2;    // the next fanout of p2
//    Map_NodeVec_t *     vFanouts;      // the array of fanouts of the gate
#endif

    // the delay information
    Map_Time_t          tArrival[2];   // the best arrival time of the neg (0) and pos (1) phases
    Map_Time_t          tRequired[2];  // the required time of the neg (0) and pos (1) phases

    // misc information  
    Map_Cut_t *         pCutBest[2];   // the best mapping for neg and pos phase
    pCuts          *Map_Cut_t;         // mapping choices for the node (elementary comes first)
    pData0               string;        // temporary storage for the corresponding network node
    pData1               string;        // temporary storage for the corresponding network node
} 

// the match of the cut
type Map_MatchStruct_t_ struct {
    // information used for matching
    pSupers        *Map_Super_t;     
    uPhase            uint32;  
    // information about the best selected match
    uPhaseBest            uint32;    // the best phase (the EXOR of match's phase and gate's phase)
    pSuperBest        *Map_Super_t;    // the best supergate matched
    // the parameters of the match
    tArrive          Map_Time_t;       // the arrival time of this match
    AreaFlow               float32;      // the area flow or area of this match
}
  
// the cuts used for matching
type Map_CutStruct_t_ struct {
    pNext          *Map_Cut_t;         // the pointer to the next cut in the list
    pOne          *Map_Cut_t;          // the father of this cut
    pTwo          *Map_Cut_t;          // the mother of this cut
    Map_Node_t *        ppLeaves[6];   // the leaves of this cut
    uTruth            uint32;        // truth table for five-input cuts
    nLeaves                byte;       // the number of leaves
    nVolume                byte;       // the volume of this cut
    fMark                byte;         // the mark to denote visited cut
    Phase                byte;         // the mark to denote complemented cut
    Map_Match_t         M[2];          // the matches for positive/negative phase
}

// the supergate internally represented
type Map_SuperStruct_t_ struct {
    Num                 int32;           // the ID of the supergate
    fSuper uint32; /* C bitfield: 1 bits; review packing and truncation */  // the flag to distinquish a real super from a fake one
    fExclude uint32; /* C bitfield: 1 bits; review packing and truncation */  // the flag if set causes gate to be excluded from being used for mapping
    nFanins uint32; /* C bitfield: 3 bits; review packing and truncation */  // the number of inputs
    nGates uint32; /* C bitfield: 3 bits; review packing and truncation */  // the number of gates inside this supergate
    nFanLimit uint32; /* C bitfield: 4 bits; review packing and truncation */  // the max number of fanout count
    nSupers uint32; /* C bitfield: 16 bits; review packing and truncation */  // the number of supergates in the list
    nPhases uint32; /* C bitfield: 4 bits; review packing and truncation */  // the number of phases for matching with canonical form
    unsigned char       uPhases[4];    // the maximum of 4 phases for matching with canonical form
    nUsed                 int32;         // the number of times the supergate is used
    Map_Super_t *       pFanins[6];    // the fanins of the gate
    pRoot         *Mio_Gate_t;         // the root gate
    uTruth [2]uint32;     // the truth table
    Map_Time_t          tDelaysR[6];   // the pin-to-pin delay constraints for the rise of the output
    Map_Time_t          tDelaysF[6];   // the pin-to-pin delay constraints for the rise of the output
    tDelayMax          Map_Time_t;     // the maximum delay
    Area               float32;          // the area
    pFormula               string;      // the symbolic formula
    pNext        *Map_Super_t;         // the pointer to the next super in the list
}

// the vector of nodes
type Map_NodeVecStruct_t_ struct {
    pArray        **Map_Node_t;        // the array of nodes
    nSize                 int32;         // the number of entries in the array
    nCap                 int32;          // the number of allocated entries
}

// the hash table 
type Map_HashTableStruct_t_ struct {
    pBins   **Map_HashEntry_t;         // the table bins
    nBins                 int32;         // the size of the table
    nEntries                 int32;      // the total number of entries in the table
    mmMan    *Extra_MmFixed_t;         // the memory manager for entries
}

// the entry in the hash table
type Map_HashEntryStruct_t_ struct {
    uTruth [2]uint32;     // the truth table for 6-var function
    uPhase            uint32;        // the phase to tranform it into the canonical form
    pGates        *Map_Super_t;        // the linked list of matching supergates
    pNext    *Map_HashEntry_t;         // the next entry in the hash table
}

// getting hold of the next fanout of the node
// Original macro:
// #define Map_NodeReadNextFanout( pNode, pFanout )                \
//     ( ( pFanout == NULL )? NULL :                               \
//         ((Map_Regular((pFanout)->p1) == (pNode))?               \
//              (pFanout)->pFanFanin1 : (pFanout)->pFanFanin2) )


// getting hold of the place where the next fanout will be attached
// Original macro:
// #define Map_NodeReadNextFanoutPlace( pNode, pFanout )           \
//     ( (Map_Regular((pFanout)->p1) == (pNode))?                  \
//          &(pFanout)->pFanFanin1 : &(pFanout)->pFanFanin2 )


// iterator through the fanouts of the node
// Original macro:
// #define Map_NodeForEachFanout( pNode, pFanout )                 \
//     for ( pFanout = (pNode)->pFanPivot; pFanout;                \
//           pFanout = Map_NodeReadNextFanout(pNode, pFanout) )


// safe iterator through the fanouts of the node
// Original macro:
// #define Map_NodeForEachFanoutSafe( pNode, pFanout, pFanout2 )   \
//     for ( pFanout  = (pNode)->pFanPivot,                        \
//           pFanout2 = Map_NodeReadNextFanout(pNode, pFanout);    \
//           pFanout;                                              \
//           pFanout  = pFanout2,                                  \
//           pFanout2 = Map_NodeReadNextFanout(pNode, pFanout) )


////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== mapperCanon.c =============================================================*/
/*=== mapperCut.c ===============================================================*/
// extern void              Map_MappingCuts( Map_Man_t * p );

/*=== mapperCutUtils.c ===============================================================*/
// extern Map_Cut_t *       Map_CutAlloc( Map_Man_t * p );

// extern void              Map_CutFree( Map_Man_t * p, Map_Cut_t * pCut );

// extern void              Map_CutPrint( Map_Man_t * p, Map_Node_t * pRoot, Map_Cut_t * pCut, int fPhase );

// extern float             Map_CutGetRootArea( Map_Cut_t * pCut, int fPhase );

// extern int               Map_CutGetLeafPhase( Map_Cut_t * pCut, int fPhase, int iLeaf );

// extern int               Map_NodeGetLeafPhase( Map_Node_t * pNode, int fPhase, int iLeaf );

// extern Map_Cut_t *       Map_CutListAppend( Map_Cut_t * pSetAll, Map_Cut_t * pSets );

// extern void              Map_CutListRecycle( Map_Man_t * p, Map_Cut_t * pSetList, Map_Cut_t * pSave );

// extern int               Map_CutListCount( Map_Cut_t * pSets );

// extern void              Map_CutRemoveFanouts( Map_Node_t * pNode, Map_Cut_t * pCut, int fPhase );

// extern void              Map_CutInsertFanouts( Map_Node_t * pNode, Map_Cut_t * pCut, int fPhase );

/*=== mapperFanout.c =============================================================*/
// extern void              Map_NodeAddFaninFanout( Map_Node_t * pFanin, Map_Node_t * pFanout );

// extern void              Map_NodeRemoveFaninFanout( Map_Node_t * pFanin, Map_Node_t * pFanoutToRemove );

// extern int               Map_NodeGetFanoutNum( Map_Node_t * pNode );

/*=== mapperLib.c ============================================================*/
// extern Map_SuperLib_t *  Map_SuperLibCreate( Mio_Library_t * pGenlib, Vec_Str_t * vStr, char * pFileName, char * pExcludeFile, int  fAlgorithm, int  fVerbose );

// extern void              Map_SuperLibFree( Map_SuperLib_t * p );

/*=== mapperMatch.c ===============================================================*/
// extern int               Map_MappingMatches( Map_Man_t * p );

/*=== mapperRefs.c =============================================================*/
// extern void              Map_MappingEstimateRefsInit( Map_Man_t * p );

// extern void              Map_MappingEstimateRefs( Map_Man_t * p );

// extern float             Map_CutGetAreaFlow( Map_Cut_t * pCut, int fPhase );

// extern float             Map_CutGetAreaRefed( Map_Cut_t * pCut, int fPhase );

// extern float             Map_CutGetAreaDerefed( Map_Cut_t * pCut, int fPhase );

// extern float             Map_CutRef( Map_Cut_t * pCut, int fPhase, int fProfile );

// extern float             Map_CutDeref( Map_Cut_t * pCut, int fPhase, int fProfile );

// extern void              Map_MappingSetRefs( Map_Man_t * pMan );

// extern float             Map_MappingGetArea( Map_Man_t * pMan );

/*=== mapperSwitch.c =============================================================*/
// extern float             Map_SwitchCutGetDerefed( Map_Node_t * pNode, Map_Cut_t * pCut, int fPhase );

// extern float             Map_SwitchCutRef( Map_Node_t * pNode, Map_Cut_t * pCut, int fPhase );

// extern float             Map_SwitchCutDeref( Map_Node_t * pNode, Map_Cut_t * pCut, int fPhase );

// extern float             Map_MappingGetSwitching( Map_Man_t * pMan );

/*=== mapperTree.c ===============================================================*/
// extern int               Map_LibraryDeriveGateInfo( Map_SuperLib_t * pLib, st__table * tExcludeGate );

// extern int               Map_LibraryReadFileTreeStr( Map_SuperLib_t * pLib, Mio_Library_t * pGenlib, Vec_Str_t * vStr, char * pFileName );

// extern int               Map_LibraryReadTree( Map_SuperLib_t * pLib, Mio_Library_t * pGenlib, char * pFileName, char * pExcludeFile );

// extern void              Map_LibraryPrintTree( Map_SuperLib_t * pLib );

/*=== mapperSuper.c ===============================================================*/
// extern int               Map_LibraryRead( Map_SuperLib_t * p, char * pFileName );

// extern void              Map_LibraryPrintSupergate( Map_Super_t * pGate );

/*=== mapperTable.c ============================================================*/
// extern Map_HashTable_t * Map_SuperTableCreate( Map_SuperLib_t * pLib );

// extern void              Map_SuperTableFree( Map_HashTable_t * p );

// extern int               Map_SuperTableInsertC( Map_HashTable_t * pLib, unsigned uTruthC[], Map_Super_t * pGate );

// extern int               Map_SuperTableInsert( Map_HashTable_t * pLib, unsigned uTruth[], Map_Super_t * pGate, unsigned uPhase );

// extern Map_Super_t *     Map_SuperTableLookup( Map_HashTable_t * p, unsigned uTruth[], unsigned * puPhase );

// extern void              Map_SuperTableSortSupergates( Map_HashTable_t * p, int nSupersMax );

// extern void              Map_SuperTableSortSupergatesByDelay( Map_HashTable_t * p, int nSupersMax );

/*=== mapperTime.c =============================================================*/
// extern float             Map_TimeCutComputeArrival( Map_Node_t * pNode, Map_Cut_t * pCut, int fPhase, float tWorstCaseLimit );

// extern float             Map_TimeComputeArrivalMax( Map_Man_t * p );

// extern void              Map_TimeComputeRequiredGlobal( Map_Man_t * p );

/*=== mapperTruth.c ===============================================================*/
// extern void              Map_MappingTruths( Map_Man_t * pMan );

// extern int               Map_TruthsCutDontCare( Map_Man_t * pMan, Map_Cut_t * pCut, unsigned * uTruthDc );

// extern int               Map_TruthCountOnes( unsigned * uTruth, int nLeaves );

// extern int               Map_TruthDetectTwoFirst( unsigned * uTruth, int nLeaves );

/*=== mapperUtils.c ===============================================================*/
// extern Map_NodeVec_t *   Map_MappingDfs( Map_Man_t * pMan, int fCollectEquiv );

// extern int               Map_MappingCountLevels( Map_Man_t * pMan );

// extern void              Map_MappingUnmark( Map_Man_t * pMan );

// extern void              Map_MappingMark_rec( Map_Node_t * pNode );

// extern void              Map_MappingUnmark_rec( Map_Node_t * pNode );

// extern void              Map_MappingPrintOutputArrivals( Map_Man_t * p );

// extern void              Map_MappingSetupMask( unsigned uMask[], int nVarsMax );

// extern int               Map_MappingNodeIsViolator( Map_Node_t * pNode, Map_Cut_t * pCut, int fPosPol );

// extern float             Map_MappingGetAreaFlow( Map_Man_t * p );

// extern void              Map_MappingSortByLevel( Map_Man_t * pMan, Map_NodeVec_t * vNodes );

// extern int               Map_MappingCountDoubles( Map_Man_t * pMan, Map_NodeVec_t * vNodes );

// extern void              Map_MappingExpandTruth( unsigned uTruth[2], int nVars );

// extern float             Map_MappingPrintSwitching( Map_Man_t * pMan );

// extern void              Map_MappingSetPlacementInfo( Map_Man_t * p );

// extern float             Map_MappingPrintWirelength( Map_Man_t * p );

// extern void              Map_MappingWireReport( Map_Man_t * p );

// extern float             Map_MappingComputeDelayWithFanouts( Map_Man_t * p );

// extern int               Map_MappingGetMaxLevel( Map_Man_t * pMan );

// extern void              Map_MappingSetChoiceLevels( Map_Man_t * pMan );

// extern void              Map_MappingReportChoices( Map_Man_t * pMan );

/*=== mapperVec.c =============================================================*/
// extern Map_NodeVec_t *   Map_NodeVecAlloc( int nCap );

// extern void              Map_NodeVecFree( Map_NodeVec_t * p );

// extern Map_NodeVec_t *   Map_NodeVecDup( Map_NodeVec_t * p );

// extern Map_Node_t **     Map_NodeVecReadArray( Map_NodeVec_t * p );

// extern int               Map_NodeVecReadSize( Map_NodeVec_t * p );

// extern void              Map_NodeVecGrow( Map_NodeVec_t * p, int nCapMin );

// extern void              Map_NodeVecShrink( Map_NodeVec_t * p, int nSizeNew );

// extern void              Map_NodeVecClear( Map_NodeVec_t * p );

// extern void              Map_NodeVecPush( Map_NodeVec_t * p, Map_Node_t * Entry );

// extern int               Map_NodeVecPushUnique( Map_NodeVec_t * p, Map_Node_t * Entry );

// extern Map_Node_t *      Map_NodeVecPop( Map_NodeVec_t * p );

// extern void              Map_NodeVecRemove( Map_NodeVec_t * p, Map_Node_t * Entry );

// extern void              Map_NodeVecWriteEntry( Map_NodeVec_t * p, int i, Map_Node_t * Entry );

// extern Map_Node_t *      Map_NodeVecReadEntry( Map_NodeVec_t * p, int i );

// extern void              Map_NodeVecSortByLevel( Map_NodeVec_t * p );




//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
