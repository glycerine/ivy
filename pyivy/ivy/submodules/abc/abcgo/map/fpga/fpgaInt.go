package fpga

/**CFile****************************************************************

  FileName    [fpgaInt.h]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Technology mapping for variable-size-LUT FPGAs.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - August 18, 2004.]

  Revision    [$Id: fpgaInt.h,v 1.8 2004/09/30 21:18:10 satrajit Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include "misc/extra/extra.h"
//#include "fpga.h"

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////
 
// uncomment to have fanouts represented in the mapping graph
//#define FPGA_ALLOCATE_FANOUT  1

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

#ifdef _WIN32
#define inline __inline // compatible with MS VS 6.0
#endif

// the maximum number of cut leaves (currently does not work for 7)
const FPGA_MAX_LEAVES =           6   

// the bit masks
// Original macro:
// #define FPGA_MASK(n)             ((~((unsigned)0)) >> (32-(n)))

#define FPGA_FULL                 (~((unsigned)0))
#define FPGA_NO_VAR               (-9999.0)
// Original macro:
// #define FPGA_NUM_BYTES(n)         (((n)/16 + (((n)%16) > 0))*16)


// maximum/minimum operators
// Original macro:
// #define FPGA_MIN(a,b)             (((a) < (b))? (a) : (b))

// Original macro:
// #define FPGA_MAX(a,b)             (((a) > (b))? (a) : (b))


// the small and large numbers (min/max float are 1.17e-38/3.40e+38)
#define FPGA_FLOAT_LARGE          ((float)1.0e+20)
#define FPGA_FLOAT_SMALL          ((float)1.0e-20)
#define FPGA_INT_LARGE            (10000000)

// the macro to compute the signature
// Original macro:
// #define FPGA_SEQ_SIGN(p)        (1 << (((ABC_PTRUINT_T)p)%31));


// internal macros to work with cuts
// Original macro:
// #define Fpga_CutIsComplement(p)  (((int)((ABC_PTRUINT_T)(p) & 01)))

// Original macro:
// #define Fpga_CutRegular(p)       ((Fpga_Cut_t *)((ABC_PTRUINT_T)(p) & ~01)) 

// Original macro:
// #define Fpga_CutNot(p)           ((Fpga_Cut_t *)((ABC_PTRUINT_T)(p) ^ 01)) 

// Original macro:
// #define Fpga_CutNotCond(p,c)     ((Fpga_Cut_t *)((ABC_PTRUINT_T)(p) ^ (c)))


// the cut nodes
// Original macro:
// #define Fpga_SeqIsComplement( p )      (((int)((ABC_PTRUINT_T) (p) & 01)))

// Original macro:
// #define Fpga_SeqRegular( p )           ((Fpga_Node_t *)((ABC_PTRUINT_T)(p) & ~015))

// Original macro:
// #define Fpga_SeqIndex( p )             ((((ABC_PTRUINT_T)(p)) >> 1) & 07)

// Original macro:
// #define Fpga_SeqIndexCreate( p, Ind )  (((ABC_PTRUINT_T)(p)) | (1 << (((ABC_PTRUINT_T)(Ind)) & 07)))


// internal macros for referencing of nodes
// Original macro:
// #define Fpga_NodeReadRef(p)      ((Fpga_Regular(p))->nRefs)

// Original macro:
// #define Fpga_NodeRef(p)          ((Fpga_Regular(p))->nRefs++)


// returns the complemented attribute of the node
// Original macro:
// #define Fpga_NodeIsSimComplement(p) (Fpga_IsComplement(p)? !(Fpga_Regular(p)->fInv) : (p)->fInv)


// generating random unsigned (#define RAND_MAX 0x7fff)
#define FPGA_RANDOM_UNSIGNED   ((((unsigned)rand()) << 24) ^ (((unsigned)rand()) << 12) ^ ((unsigned)rand()))

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

// the mapping manager
type Fpga_ManStruct_t_ struct {
    // the mapping graph
    pBins       **Fpga_Node_t;         // the table of nodes hashed by their children
    nBins                 int32;         // the size of the table
    pInputs       **Fpga_Node_t;       // the array of inputs
    nInputs                 int32;       // the number of inputs
    pOutputs       **Fpga_Node_t;      // the array of outputs
    nOutputs                 int32;      // the number of outputs
    nNodes                 int32;        // the total number of nodes
    nLatches                 int32;      // the number of latches in the circuit
    pConst1        *Fpga_Node_t;       // the constant 1 node
    vNodesAll     *Fpga_NodeVec_t;     // the nodes by number
    vAnds     *Fpga_NodeVec_t;         // the nodes reachable from COs
    vMapping     *Fpga_NodeVec_t;      // the nodes used in the current mapping

    // info about the original circuit
    pFileName               string;     // the file name
    ppOutputNames              *string; // the primary output names
    pInputArrivals              *float32;// the PI arrival times

    // mapping parameters
    nVarsMax                 int32;      // the max number of variables
    fAreaRecovery                 int32; // the flag to use area flow as the first parameter
    fVerbose                 int32;      // the verbosiness flag
    fSwitching                 int32;    // minimize the switching activity (instead of area)
    fLatchPaths                 int32;   // optimize latch paths for delay, other paths for area
    nTravIds                 int32;      // the counter of traversal IDs
    DelayTarget               float32;   // the target required times

    // support of choice nodes
    nChoiceNodes                 int32;  // the number of choice nodes
    nChoices                 int32;      // the number of all choices

    nCanons                 int32;
    nMatches                 int32;
 
    // the supergate library
    pLutLib      *Fpga_LutLib_t;       // the current LUT library

    // the memory managers
    mmNodes    *Extra_MmFixed_t;       // the memory manager for nodes
    mmCuts    *Extra_MmFixed_t;        // the memory manager for cuts

    // resynthesis parameters
    fResynthesis                 int32;  // the resynthesis flag
    fRequiredGlo               float32;  // the global required times
    fRequiredShift               float32;// the shift of the required times
    fRequiredStart               float32;// the starting global required times
    fRequiredGain               float32; // the reduction in delay
    fAreaGlo               float32;      // the total area
    fAreaGain               float32;     // the reduction in area
    fEpsilon               float32;      // the epsilon used to compare floats
    fDelayWindow               float32;  // the delay window for delay-oriented resynthesis
    DelayLimit               float32;    // for resynthesis
    AreaLimit               float32;     // for resynthesis
    TimeLimit               float32;     // for resynthesis

    // runtime statistics
    timeToMap             int64;     // time to transfer to the mapping structure
    timeCuts             int64;      // time to compute k-feasible cuts
    timeTruth             int64;     // time to compute the truth table for each cut
    timeMatch             int64;     // time to perform matching for each node
    timeRecover             int64;   // time to perform area recovery
    timeToNet             int64;     // time to transfer back to the network
    timeTotal             int64;     // the total mapping time
    time1             int64;         // time to transfer to the mapping structure
    time2             int64;         // time to transfer to the mapping structure
}

// the LUT library
type Fpga_LutLibStruct_t_ struct {
    pName               string;         // the name of the LUT library
    LutMax                 int32;        // the maximum LUT size 
    fVarPinDelays                 int32; // set to 1 if variable pin delays are specified
    pLutAreas [FPGA_MAX_LUTSIZE+1]float32; // the areas of LUTs
    pLutDelays [FPGA_MAX_LUTSIZE+1][FPGA_MAX_LUTSIZE+1]float32;// the delays of LUTs
}

// the mapping node
type Fpga_NodeStruct_t_ struct {
    // general information about the node
    pNext        *Fpga_Node_t;         // the next node in the hash table
    pLevel        *Fpga_Node_t;        // the next node in the linked list by level
    Num                 int32;           // the unique number of this node
    NumA                 int32;          // the unique number of this node
    Num2                 int32;          // the temporary number of this node
    nRefs                 int32;         // the number of references (fanouts) of the given node
    fMark0 uint32; /* C bitfield: 1 bits; review packing and truncation */    // the mark used for traversals
    fMark1 uint32; /* C bitfield: 1 bits; review packing and truncation */    // the mark used for traversals
    fInv uint32; /* C bitfield: 1 bits; review packing and truncation */    // the complemented attribute for the equivalent nodes
    Value uint32; /* C bitfield: 2 bits; review packing and truncation */    // the value of the nodes
    fUsed uint32; /* C bitfield: 1 bits; review packing and truncation */    // the flag indicating that the node is used in the mapping
    fTemp uint32; /* C bitfield: 1 bits; review packing and truncation */    // unused
    Level uint32; /* C bitfield: 11 bits; review packing and truncation */    // the level of the given node
    uData uint32; /* C bitfield: 14 bits; review packing and truncation */    // used to mark the fanins, for which resynthesis was tried
    TravId                 int32;

    // the successors of this node     
    p1        *Fpga_Node_t;            // the first child
    p2        *Fpga_Node_t;            // the second child
    pNextE        *Fpga_Node_t;        // the next functionally equivalent node
    pRepr        *Fpga_Node_t;         // the representative of the functionally equivalent class

#ifdef FPGA_ALLOCATE_FANOUT
    // representation of node's fanouts
    Fpga_Node_t *       pFanPivot;     // the first fanout of this node
    Fpga_Node_t *       pFanFanin1;    // the next fanout of p1
    Fpga_Node_t *       pFanFanin2;    // the next fanout of p2
//    Fpga_NodeVec_t *    vFanouts;      // the array of fanouts of the gate
#endif

    // the delay information
    tRequired               float32;     // the best area flow 
    aEstFanouts               float32;   // the fanout estimation
    Switching               float32;     // the probability of switching
    LValue                 int32;        // the l-value of the node
    nLatches1               int16;     // the number of latches on the first edge
    nLatches2               int16;     // the number of latches on the second edge

    // cut information
    pCutBest         *Fpga_Cut_t;      // the best mapping
    pCutOld         *Fpga_Cut_t;       // the old mapping
    pCuts         *Fpga_Cut_t;         // mapping choices for the node (elementary comes first)
    pCutsN         *Fpga_Cut_t;        // mapping choices for the node (elementary comes first)

    // misc information  
    pData0               string;        // temporary storage for the corresponding network node
} 
  
// the cuts used for matching
type Fpga_CutStruct_t_ struct {
    pOne         *Fpga_Cut_t;          // the father of this cut
    pTwo         *Fpga_Cut_t;          // the mother of this cut
    pRoot        *Fpga_Node_t;         // the root of the cut
    Fpga_Node_t *       ppLeaves[FPGA_MAX_LEAVES+1];   // the leaves of this cut
    fLevel               float32;        // the average level of the fanins
    uSign            uint32;         // signature for quick comparison
    fMark                byte;         // the mark to denote visited cut
    Phase                byte;         // the mark to denote complemented cut
    nLeaves                byte;       // the number of leaves of this cut
    nVolume                byte;       // the volume of this cut
    tArrival               float32;      // the arrival time 
    aFlow               float32;         // the area flow of the cut
    pNext         *Fpga_Cut_t;         // the pointer to the next cut in the list
}

// the vector of nodes
type Fpga_NodeVecStruct_t_ struct {
    pArray       **Fpga_Node_t;        // the array of nodes
    nSize                 int32;         // the number of entries in the array
    nCap                 int32;          // the number of allocated entries
}

// getting hold of the next fanout of the node
// Original macro:
// #define Fpga_NodeReadNextFanout( pNode, pFanout )                \
//     ( ( pFanout == NULL )? NULL :                                \
//         ((Fpga_Regular((pFanout)->p1) == (pNode))?               \
//              (pFanout)->pFanFanin1 : (pFanout)->pFanFanin2) )


// getting hold of the place where the next fanout will be attached
// Original macro:
// #define Fpga_NodeReadNextFanoutPlace( pNode, pFanout )           \
//     ( (Fpga_Regular((pFanout)->p1) == (pNode))?                  \
//          &(pFanout)->pFanFanin1 : &(pFanout)->pFanFanin2 )


// iterator through the fanouts of the node
// Original macro:
// #define Fpga_NodeForEachFanout( pNode, pFanout )                 \
//     for ( pFanout = (pNode)->pFanPivot; pFanout;                 \
//           pFanout = Fpga_NodeReadNextFanout(pNode, pFanout) )


// safe iterator through the fanouts of the node
// Original macro:
// #define Fpga_NodeForEachFanoutSafe( pNode, pFanout, pFanout2 )   \
//     for ( pFanout  = (pNode)->pFanPivot,                         \
//           pFanout2 = Fpga_NodeReadNextFanout(pNode, pFanout);    \
//           pFanout;                                               \
//           pFanout  = pFanout2,                                   \
//           pFanout2 = Fpga_NodeReadNextFanout(pNode, pFanout) )


func Fpga_FloatMoreThan(p *Fpga_Man_t, Arg1 float32, Arg2 float32) int32 { return Arg1 > Arg2 + p.fEpsilon; }
func Fpga_FloatLessThan(p *Fpga_Man_t, Arg1 float32, Arg2 float32) int32 { return Arg1 < Arg2 - p.fEpsilon; }
func Fpga_FloatEqual(p *Fpga_Man_t, Arg1 float32, Arg2 float32) int32 { return Arg1 > Arg2 - p.fEpsilon && Arg1 < Arg2 + p.fEpsilon; }

////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== fpgaCut.c ===============================================================*/
// extern void              Fpga_MappingCuts( Fpga_Man_t * p );

// extern void              Fpga_MappingCreatePiCuts( Fpga_Man_t * p );

// extern int               Fpga_CutCountAll( Fpga_Man_t * pMan );

/*=== fpgaCutUtils.c ===============================================================*/
// extern Fpga_Cut_t *      Fpga_CutAlloc( Fpga_Man_t * p );

// extern Fpga_Cut_t *      Fpga_CutDup( Fpga_Man_t * p, Fpga_Cut_t * pCutOld );

// extern void              Fpga_CutFree( Fpga_Man_t * p, Fpga_Cut_t * pCut );

// extern void              Fpga_CutPrint( Fpga_Man_t * p, Fpga_Node_t * pRoot, Fpga_Cut_t * pCut );

// extern Fpga_Cut_t *      Fpga_CutCreateSimple( Fpga_Man_t * p, Fpga_Node_t * pNode );

// extern float             Fpga_CutGetRootArea( Fpga_Man_t * p, Fpga_Cut_t * pCut );

// extern Fpga_Cut_t *      Fpga_CutListAppend( Fpga_Cut_t * pSetAll, Fpga_Cut_t * pSets );

// extern void              Fpga_CutListRecycle( Fpga_Man_t * p, Fpga_Cut_t * pSetList, Fpga_Cut_t * pSave );

// extern int               Fpga_CutListCount( Fpga_Cut_t * pSets );

// extern void              Fpga_CutRemoveFanouts( Fpga_Man_t * p, Fpga_Node_t * pNode, Fpga_Cut_t * pCut );

// extern void              Fpga_CutInsertFanouts( Fpga_Man_t * p, Fpga_Node_t * pNode, Fpga_Cut_t * pCut );

// extern float             Fpga_CutGetAreaRefed( Fpga_Man_t * pMan, Fpga_Cut_t * pCut );

// extern float             Fpga_CutGetAreaDerefed( Fpga_Man_t * pMan, Fpga_Cut_t * pCut );

// extern float             Fpga_CutRef( Fpga_Man_t * pMan, Fpga_Node_t * pNode, Fpga_Cut_t * pCut, int fFanouts );

// extern float             Fpga_CutDeref( Fpga_Man_t * pMan, Fpga_Node_t * pNode, Fpga_Cut_t * pCut, int fFanouts );

// extern float             Fpga_CutGetAreaFlow( Fpga_Man_t * pMan, Fpga_Cut_t * pCut );

// extern void              Fpga_CutGetParameters( Fpga_Man_t * pMan, Fpga_Cut_t * pCut );

/*=== fraigFanout.c =============================================================*/
// extern void              Fpga_NodeAddFaninFanout( Fpga_Node_t * pFanin, Fpga_Node_t * pFanout );

// extern void              Fpga_NodeRemoveFaninFanout( Fpga_Node_t * pFanin, Fpga_Node_t * pFanoutToRemove );

// extern int               Fpga_NodeGetFanoutNum( Fpga_Node_t * pNode );

/*=== fpgaLib.c ============================================================*/
// extern Fpga_LutLib_t *   Fpga_LutLibRead( char * FileName, int fVerbose );

// extern void              Fpga_LutLibFree( Fpga_LutLib_t * p );

// extern void              Fpga_LutLibPrint( Fpga_LutLib_t * pLutLib );

// extern int               Fpga_LutLibDelaysAreDiscrete( Fpga_LutLib_t * pLutLib );

/*=== fpgaMatch.c ===============================================================*/
// extern int               Fpga_MappingMatches( Fpga_Man_t * p, int fDelayOriented );

// extern int               Fpga_MappingMatchesArea( Fpga_Man_t * p );

// extern int               Fpga_MappingMatchesSwitch( Fpga_Man_t * p );

/*=== fpgaShow.c =============================================================*/
// extern void              Fpga_MappingShow( Fpga_Man_t * pMan, char * pFileName );

// extern void              Fpga_MappingShowNodes( Fpga_Man_t * pMan, Fpga_Node_t ** ppRoots, int nRoots, char * pFileName );

/*=== fpgaSwitch.c =============================================================*/
// extern float             Fpga_CutGetSwitchDerefed( Fpga_Man_t * pMan, Fpga_Node_t * pNode, Fpga_Cut_t * pCut );

// extern float             Fpga_CutRefSwitch( Fpga_Man_t * pMan, Fpga_Node_t * pNode, Fpga_Cut_t * pCut, int fFanouts );

// extern float             Fpga_CutDerefSwitch( Fpga_Man_t * pMan, Fpga_Node_t * pNode, Fpga_Cut_t * pCut, int fFanouts );

// extern float             Fpga_MappingGetSwitching( Fpga_Man_t * pMan, Fpga_NodeVec_t * vMapping );

/*=== fpgaTime.c ===============================================================*/
// extern float             Fpga_TimeCutComputeArrival( Fpga_Man_t * pMan, Fpga_Cut_t * pCut );

// extern float             Fpga_TimeCutComputeArrival_rec( Fpga_Man_t * pMan, Fpga_Cut_t * pCut );

// extern float             Fpga_TimeComputeArrivalMax( Fpga_Man_t * p );

// extern void              Fpga_TimeComputeRequiredGlobal( Fpga_Man_t * p, int fFirstTime );

// extern void              Fpga_TimeComputeRequired( Fpga_Man_t * p, float fRequired );

// extern void              Fpga_TimePropagateRequired( Fpga_Man_t * p, Fpga_NodeVec_t * vNodes );

// extern void              Fpga_TimePropagateArrival( Fpga_Man_t * p );

/*=== fpgaVec.c =============================================================*/
// extern Fpga_NodeVec_t *  Fpga_NodeVecAlloc( int nCap );

// extern void              Fpga_NodeVecFree( Fpga_NodeVec_t * p );

// extern Fpga_Node_t **    Fpga_NodeVecReadArray( Fpga_NodeVec_t * p );

// extern int               Fpga_NodeVecReadSize( Fpga_NodeVec_t * p );

// extern void              Fpga_NodeVecGrow( Fpga_NodeVec_t * p, int nCapMin );

// extern void              Fpga_NodeVecShrink( Fpga_NodeVec_t * p, int nSizeNew );

// extern void              Fpga_NodeVecClear( Fpga_NodeVec_t * p );

// extern void              Fpga_NodeVecPush( Fpga_NodeVec_t * p, Fpga_Node_t * Entry );

// extern int               Fpga_NodeVecPushUnique( Fpga_NodeVec_t * p, Fpga_Node_t * Entry );

// extern Fpga_Node_t *     Fpga_NodeVecPop( Fpga_NodeVec_t * p );

// extern void              Fpga_NodeVecWriteEntry( Fpga_NodeVec_t * p, int i, Fpga_Node_t * Entry );

// extern Fpga_Node_t *     Fpga_NodeVecReadEntry( Fpga_NodeVec_t * p, int i );

// extern void              Fpga_NodeVecSortByLevel( Fpga_NodeVec_t * p );

// extern void              Fpga_SortNodesByArrivalTimes( Fpga_NodeVec_t * p );

// extern void              Fpga_NodeVecUnion( Fpga_NodeVec_t * p, Fpga_NodeVec_t * p1, Fpga_NodeVec_t * p2 );

// extern void              Fpga_NodeVecPushOrder( Fpga_NodeVec_t * vNodes, Fpga_Node_t * pNode, int fIncreasing );

// extern void              Fpga_NodeVecReverse( Fpga_NodeVec_t * vNodes );


/*=== fpgaUtils.c ===============================================================*/
// extern Fpga_NodeVec_t *  Fpga_MappingDfs( Fpga_Man_t * pMan, int fCollectEquiv );

// extern Fpga_NodeVec_t *  Fpga_MappingDfsNodes( Fpga_Man_t * pMan, Fpga_Node_t ** ppNodes, int nNodes, int fEquiv );

// extern int               Fpga_CountLevels( Fpga_Man_t * pMan );

// extern float             Fpga_MappingGetAreaFlow( Fpga_Man_t * p );

// extern float             Fpga_MappingArea( Fpga_Man_t * pMan );

// extern float             Fpga_MappingAreaTrav( Fpga_Man_t * pMan );

// extern float             Fpga_MappingSetRefsAndArea( Fpga_Man_t * pMan );

// extern void              Fpga_MappingPrintOutputArrivals( Fpga_Man_t * p );

// extern void              Fpga_MappingSetupTruthTables( unsigned uTruths[][2] );

// extern void              Fpga_MappingSetupMask( unsigned uMask[], int nVarsMax );

// extern void              Fpga_MappingSortByLevel( Fpga_Man_t * pMan, Fpga_NodeVec_t * vNodes, int fIncreasing );

// extern Fpga_NodeVec_t *  Fpga_DfsLim( Fpga_Man_t * pMan, Fpga_Node_t * pNode, int nLevels );

// extern Fpga_NodeVec_t *  Fpga_MappingLevelize( Fpga_Man_t * pMan, Fpga_NodeVec_t * vNodes );

// extern int               Fpga_MappingMaxLevel( Fpga_Man_t * pMan );

// extern void              Fpga_ManReportChoices( Fpga_Man_t * pMan );

// extern void              Fpga_MappingSetChoiceLevels( Fpga_Man_t * pMan );




//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
