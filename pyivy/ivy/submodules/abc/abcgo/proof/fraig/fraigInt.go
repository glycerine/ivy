package fraig

import "github.com/glycerine/abcgo/sat/msat"

/**CFile****************************************************************

  FileName    [fraigInt.h]

  PackageName [FRAIG: Functionally reduced AND-INV graphs.]

  Synopsis    [Internal declarations of the FRAIG package.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - October 1, 2004]

  Revision    [$Id: fraigInt.h,v 1.15 2005/07/08 01:01:31 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/util/abc_global.h"
//#include "fraig.h"
//#include "sat/msat/msat.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

/*
   The AIG node policy:
   - Each node has its main number (pNode->Num)
     This is the number of this node in the array of all nodes and its SAT variable number
   - The PI nodes are stored along with other nodes
     Additionally, PI nodes have a PI number, by which they are stored in the PI node array
   - The constant node is has number 0 and is also stored in the array
*/

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

// enable this macro to support the fanouts
const FRAIG_ENABLE_FANOUTS = true
const FRAIG_PATTERNS_RANDOM = 2048  // should not be less than 128 and more than 32768 (2^15)
const FRAIG_PATTERNS_DYNAMIC = 2048 // should not be less than 256 and more than 32768 (2^15)
const FRAIG_MAX_PRIMES = 1024       // the maximum number of primes used for hashing

// this parameter determines when simulation info is extended
// it will be extended when the free storage in the dynamic simulation
// info is less or equal to this number of words (FRAIG_WORDS_STORE)
// this is done because if the free storage for dynamic simulation info
// is not sufficient, computation becomes inefficient
const FRAIG_WORDS_STORE = 5

// the bit masks
// Original macro:
// #define FRAIG_MASK(n)             ((~((unsigned)0)) >> (32-(n)))

const FRAIG_FULL uint32 = ^uint32(0)

// Original macro:
// #define FRAIG_NUM_WORDS(n)         (((n)>>5) + (((n)&31) > 0))

// generating random unsigned (#define RAND_MAX 0x7fff)
//#define FRAIG_RANDOM_UNSIGNED   ((((unsigned)rand()) << 24) ^ (((unsigned)rand()) << 12) ^ ((unsigned)rand()))
// FRAIG_RANDOM_UNSIGNED is supplied by the manager's AIG random generator.

// macros to get hold of the bits in a bit string
// Original macro:
// #define Fraig_BitStringSetBit(p,i)  ((p)[(i)>>5] |= (1<<((i) & 31)))

// Original macro:
// #define Fraig_BitStringXorBit(p,i)  ((p)[(i)>>5] ^= (1<<((i) & 31)))

// Original macro:
// #define Fraig_BitStringHasBit(p,i) (((p)[(i)>>5]  & (1<<((i) & 31))) > 0)

// macros to get hold of the bits in the support info
//#define Fraig_NodeSetVarStr(p,i)      (Fraig_Regular(p)->pSuppStr[((i)%FRAIG_SUPP_SIGN)>>5] |= (1<<(((i)%FRAIG_SUPP_SIGN) & 31)))
//#define Fraig_NodeHasVarStr(p,i)     ((Fraig_Regular(p)->pSuppStr[((i)%FRAIG_SUPP_SIGN)>>5]  & (1<<(((i)%FRAIG_SUPP_SIGN) & 31))) > 0)
// Original macro:
// #define Fraig_NodeSetVarStr(p,i)     Fraig_BitStringSetBit(Fraig_Regular(p)->pSuppStr,i)

// Original macro:
// #define Fraig_NodeHasVarStr(p,i)     Fraig_BitStringHasBit(Fraig_Regular(p)->pSuppStr,i)

// copied from "extra.h" for stand-aloneness
// Original macro:
// #define Fraig_PrintTime(a,t)      printf( "%s = ", (a) ); printf( "%6.2f sec\n", (float)(t)/(float)(CLOCKS_PER_SEC) )

// Original macro:
// #define Fraig_HashKey2(a,b,TSIZE) (((ABC_PTRUINT_T)(a) + (ABC_PTRUINT_T)(b) * 12582917) % TSIZE)

//#define Fraig_HashKey2(a,b,TSIZE) (( ((unsigned)(a)->Num * 19) ^ ((unsigned)(b)->Num * 1999) ) % TSIZE)
//#define Fraig_HashKey2(a,b,TSIZE) ( ((unsigned)((a)->Num + (b)->Num) * ((a)->Num + (b)->Num + 1) / 2) % TSIZE)
// the other two hash functions give bad distribution of hash chain lengths (not clear why)

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

type ABC_INT64_T = int64
type abctime = int64
type Msat_Solver_t = msat.Msat_Solver_t
type Msat_IntVec_t = msat.Msat_IntVec_t
type Msat_ClauseVec_t = msat.Msat_ClauseVec_t

// the mapping manager
type Fraig_ManStruct_t_ struct {
	// the AIG nodes
	vInputs  *Fraig_NodeVec_t // the array of primary inputs
	vNodes   *Fraig_NodeVec_t // the array of all nodes, including primary inputs
	vOutputs *Fraig_NodeVec_t // the array of primary outputs (some internal nodes)
	pConst1  *Fraig_Node_t    // the pointer to the constant node (vNodes->pArray[0])

	// info about the original circuit
	ppInputNames  []string // the primary input names
	ppOutputNames []string // the primary output names

	// various hash-tables
	pTableS  *Fraig_HashTable_t // hashing by structure
	pTableF  *Fraig_HashTable_t // hashing by simulation info
	pTableF0 *Fraig_HashTable_t // hashing by simulation info (sparse functions)

	// parameters
	nWordsRand int32       // the number of words of random simulation info
	nWordsDyna int32       // the number of words of dynamic simulation info
	nBTLimit   int32       // the max number of backtracks to perform
	nSeconds   int32       // the runtime limit for the miter proof
	fFuncRed   int32       // performs only one level hashing
	fFeedBack  int32       // enables solver feedback
	fDist1Pats int32       // enables solver feedback
	fDoSparse  int32       // performs equiv tests for sparse functions
	fChoicing  int32       // enables recording structural choices
	fTryProve  int32       // tries to solve the final miter
	fVerbose   int32       // the verbosiness flag
	fVerboseP  int32       // the verbosiness flag
	nInspLimit ABC_INT64_T // the inspection limit

	nTravIds  int32 // the traversal counter
	nTravIds2 int32 // the traversal counter

	// info related to the solver feedback
	iWordStart int32            // the first word to use for simulation
	iWordPerm  int32            // the number of words stored permanently
	iPatsPerm  int32            // the number of patterns stored permanently
	vCones     *Fraig_NodeVec_t // the temporary array of internal variables
	vPatsReal  *Msat_IntVec_t   // the array of real pattern numbers
	pSimsReal  []uint32         // used for simulation patterns
	pSimsDiff  []uint32         // used for simulation patterns
	pSimsTemp  []uint32         // used for simulation patterns

	// the support information
	nSuppWords int32
	pSuppS     [][]uint32
	pSuppF     [][]uint32

	// the memory managers
	mmNodes *fraigNodePool    // typed chunks for pointer-bearing Go nodes
	mmSims  *Fraig_MemFixed_t // the memory manager for simulation info

	// solving the SAT problem
	pSat     *Msat_Solver_t // the SAT solver
	vProj    *Msat_IntVec_t // the temporary array of projection vars
	nSatNums int32          // the counter of SAT variables
	pModel   []int32        // the assignment, which satisfies the miter
	// these arrays belong to the solver
	vVarsInt   *Msat_IntVec_t   // the temporary array of variables
	vAdjacents []*Msat_IntVec_t // the temporary storage for connectivity
	vVarsUsed  *Msat_IntVec_t   // the array marking vars appearing in the cone

	// various statistic variables
	nSatCalls     int32 // the number of times equivalence checking was called
	nSatProof     int32 // the number of times a proof was found
	nSatCounter   int32 // the number of times a counter example was found
	nSatFails     int32 // the number of times the SAT solver failed to complete due to resource limit or prediction
	nSatFailsReal int32 // the number of times the SAT solver failed to complete due to resource limit

	nSatCallsImp   int32 // the number of times equivalence checking was called
	nSatProofImp   int32 // the number of times a proof was found
	nSatCounterImp int32 // the number of times a counter example was found
	nSatFailsImp   int32 // the number of times the SAT solver failed to complete

	nSatZeros    int32 // the number of times the simulation vector is zero
	nSatSupps    int32 // the number of times the support info was useful
	nRefErrors   int32 // the number of ref counting errors
	nImplies     int32 // the number of implication cases
	nSatImpls    int32 // the number of implication SAT calls
	nVarsClauses int32 // the number of variables with clauses
	nSimplifies0 int32
	nSimplifies1 int32
	nImplies0    int32
	nImplies1    int32

	// runtime statistics
	timeToAig abctime // time to transfer to the mapping structure
	timeSims  abctime // time to compute k-feasible cuts
	timeTrav  abctime // time to traverse the network
	timeFeed  abctime // time for solver feedback (recording and resimulating)
	timeImply abctime // time to analyze implications
	timeSat   abctime // time to compute the truth table for each cut
	timeToNet abctime // time to transfer back to the network
	timeTotal abctime // the total mapping time
	time1     abctime // time to perform one task
	time2     abctime // time to perform another task
	time3     abctime // time to perform another task
	time4     abctime // time to perform another task
}

// the mapping node
type Fraig_NodeStruct_t_ struct {
	// Complemented C pointers become canonical paired Go objects. Only the
	// regular node owns graph data; a complemented handle points back to it.
	complementOf *Fraig_Node_t
	complement   *Fraig_Node_t
	// various numbers associated with the node
	Num     int32 // the unique number (SAT var number) of this node
	NumPi   int32 // if the node is a PI, this is its variable number
	Level   int32 // the level of the node
	nRefs   int32 // the number of references of the node
	TravId  int32 // the traversal ID (use to avoid cleaning marks)
	TravId2 int32 // the traversal ID (use to avoid cleaning marks)

	// general information about the node
	fInv     uint32 /* C bitfield: 1 bits; review packing and truncation */  // the mark to show that simulation info is complemented
	fNodePo  uint32 /* C bitfield: 1 bits; review packing and truncation */  // the mark used for primary outputs
	fClauses uint32 /* C bitfield: 1 bits; review packing and truncation */  // the clauses for this node are loaded
	fMark0   uint32 /* C bitfield: 1 bits; review packing and truncation */  // the mark used for traversals
	fMark1   uint32 /* C bitfield: 1 bits; review packing and truncation */  // the mark used for traversals
	fMark2   uint32 /* C bitfield: 1 bits; review packing and truncation */  // the mark used for traversals
	fMark3   uint32 /* C bitfield: 1 bits; review packing and truncation */  // the mark used for traversals
	fFeedUse uint32 /* C bitfield: 1 bits; review packing and truncation */  // the presence of the variable in the feedback
	fFeedVal uint32 /* C bitfield: 1 bits; review packing and truncation */  // the value of the variable in the feedback
	fFailTfo uint32 /* C bitfield: 1 bits; review packing and truncation */  // the node is in the TFO of the failed SAT run
	nFanouts uint32 /* C bitfield: 2 bits; review packing and truncation */  // the indicator of fanouts (none, one, or many)
	nOnes    uint32 /* C bitfield: 20 bits; review packing and truncation */ // the number of 1's in the random sim info

	// the children of the node
	p1      *Fraig_Node_t    // the first child
	p2      *Fraig_Node_t    // the second child
	vFanins *Fraig_NodeVec_t // the fanins of the supergate rooted at this node
	//    Fraig_NodeVec_t *     vFanouts;      // the fanouts of the supergate rooted at this node

	// various linked lists
	pNextS *Fraig_Node_t // the next node in the structural hash table
	pNextF *Fraig_Node_t // the next node in the functional (simulation) hash table
	pNextD *Fraig_Node_t // the next node in the list of nodes based on dynamic simulation
	pNextE *Fraig_Node_t // the next structural choice (functionally-equivalent node)
	pRepr  *Fraig_Node_t // the canonical functional representative of the node

	// simulation data
	uHashR uint32   // the hash value for random information
	uHashD uint32   // the hash value for dynamic information
	puSimR []uint32 // the simulation information (random)
	puSimD []uint32 // the simulation information (dynamic)

	// misc information
	pData0 *Fraig_Node_t // temporary storage for the corresponding network node
	pData1 *Fraig_Node_t // temporary storage for the corresponding network node

	// representation of node's fanouts
	pFanPivot  *Fraig_Node_t // the first fanout of this node
	pFanFanin1 *Fraig_Node_t // the next fanout of p1
	pFanFanin2 *Fraig_Node_t // the next fanout of p2

}

// the vector of nodes
type Fraig_NodeVecStruct_t_ struct {
	// The slice is authoritative for both size and capacity.
	pArray []*Fraig_Node_t
}

// the hash table
type Fraig_HashTableStruct_t_ struct {
	pBins    []*Fraig_Node_t // the table bins
	nBins    int32           // the size of the table
	nEntries int32           // the total number of entries in the table
}

// getting hold of the next fanout of the node
// Original macro:
// #define Fraig_NodeReadNextFanout( pNode, pFanout )                \
//     ( ( pFanout == NULL )? NULL :                                 \
//         ((Fraig_Regular((pFanout)->p1) == (pNode))?               \
//              (pFanout)->pFanFanin1 : (pFanout)->pFanFanin2) )

// getting hold of the place where the next fanout will be attached
// Original macro:
// #define Fraig_NodeReadNextFanoutPlace( pNode, pFanout )           \
//     ( (Fraig_Regular((pFanout)->p1) == (pNode))?                  \
//          &(pFanout)->pFanFanin1 : &(pFanout)->pFanFanin2 )

// iterator through the fanouts of the node
// Original macro:
// #define Fraig_NodeForEachFanout( pNode, pFanout )                 \
//     for ( pFanout = (pNode)->pFanPivot; pFanout;                  \
//           pFanout = Fraig_NodeReadNextFanout(pNode, pFanout) )

// safe iterator through the fanouts of the node
// Original macro:
// #define Fraig_NodeForEachFanoutSafe( pNode, pFanout, pFanout2 )   \
//     for ( pFanout  = (pNode)->pFanPivot,                          \
//           pFanout2 = Fraig_NodeReadNextFanout(pNode, pFanout);    \
//           pFanout;                                                \
//           pFanout  = pFanout2,                                    \
//           pFanout2 = Fraig_NodeReadNextFanout(pNode, pFanout) )

// iterators through the entries in the linked lists of nodes
// the list of nodes in the structural hash table
// Original macro:
// #define Fraig_TableBinForEachEntryS( pBin, pEnt )                 \
//     for ( pEnt = pBin;                                            \
//           pEnt;                                                   \
//           pEnt = pEnt->pNextS )

// Original macro:
// #define Fraig_TableBinForEachEntrySafeS( pBin, pEnt, pEnt2 )      \
//     for ( pEnt = pBin,                                            \
//           pEnt2 = pEnt? pEnt->pNextS: NULL;                       \
//           pEnt;                                                   \
//           pEnt = pEnt2,                                           \
//           pEnt2 = pEnt? pEnt->pNextS: NULL )

// the list of nodes in the functional (simulation) hash table
// Original macro:
// #define Fraig_TableBinForEachEntryF( pBin, pEnt )                 \
//     for ( pEnt = pBin;                                            \
//           pEnt;                                                   \
//           pEnt = pEnt->pNextF )

// Original macro:
// #define Fraig_TableBinForEachEntrySafeF( pBin, pEnt, pEnt2 )      \
//     for ( pEnt = pBin,                                            \
//           pEnt2 = pEnt? pEnt->pNextF: NULL;                       \
//           pEnt;                                                   \
//           pEnt = pEnt2,                                           \
//           pEnt2 = pEnt? pEnt->pNextF: NULL )

// the list of nodes with the same simulation and different functionality
// Original macro:
// #define Fraig_TableBinForEachEntryD( pBin, pEnt )                 \
//     for ( pEnt = pBin;                                            \
//           pEnt;                                                   \
//           pEnt = pEnt->pNextD )

// Original macro:
// #define Fraig_TableBinForEachEntrySafeD( pBin, pEnt, pEnt2 )      \
//     for ( pEnt = pBin,                                            \
//           pEnt2 = pEnt? pEnt->pNextD: NULL;                       \
//           pEnt;                                                   \
//           pEnt = pEnt2,                                           \
//           pEnt2 = pEnt? pEnt->pNextD: NULL )

// the list of nodes with the same functionality
// Original macro:
// #define Fraig_TableBinForEachEntryE( pBin, pEnt )                 \
//     for ( pEnt = pBin;                                            \
//           pEnt;                                                   \
//           pEnt = pEnt->pNextE )

// Original macro:
// #define Fraig_TableBinForEachEntrySafeE( pBin, pEnt, pEnt2 )      \
//     for ( pEnt = pBin,                                            \
//           pEnt2 = pEnt? pEnt->pNextE: NULL;                       \
//           pEnt;                                                   \
//           pEnt = pEnt2,                                           \
//           pEnt2 = pEnt? pEnt->pNextE: NULL )

////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

// random number generator imported from another package
// extern unsigned            Aig_ManRandom( int fReset );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== fraigCanon.c =============================================================*/
// extern Fraig_Node_t *      Fraig_NodeAndCanon( Fraig_Man_t * pMan, Fraig_Node_t * p1, Fraig_Node_t * p2 );

/*=== fraigFanout.c =============================================================*/
// extern void                Fraig_NodeAddFaninFanout( Fraig_Node_t * pFanin, Fraig_Node_t * pFanout );

// extern void                Fraig_NodeRemoveFaninFanout( Fraig_Node_t * pFanin, Fraig_Node_t * pFanoutToRemove );

// extern int                 Fraig_NodeGetFanoutNum( Fraig_Node_t * pNode );

/*=== fraigFeed.c =============================================================*/
// extern void                Fraig_FeedBackInit( Fraig_Man_t * p );

// extern void                Fraig_FeedBack( Fraig_Man_t * p, int * pModel, Msat_IntVec_t * vVars, Fraig_Node_t * pOld, Fraig_Node_t * pNew );

// extern void                Fraig_FeedBackTest( Fraig_Man_t * p );

// extern int                 Fraig_FeedBackCompress( Fraig_Man_t * p );

// extern int *               Fraig_ManAllocCounterExample( Fraig_Man_t * p );

// extern int *               Fraig_ManSaveCounterExample( Fraig_Man_t * p, Fraig_Node_t * pNode );

/*=== fraigMan.c =============================================================*/
// extern void                Fraig_ManCreateSolver( Fraig_Man_t * p );

/*=== fraigMem.c =============================================================*/
// extern Fraig_MemFixed_t *  Fraig_MemFixedStart( int nEntrySize );

// extern void                Fraig_MemFixedStop( Fraig_MemFixed_t * p, int fVerbose );

// extern char *              Fraig_MemFixedEntryFetch( Fraig_MemFixed_t * p );

// extern void                Fraig_MemFixedEntryRecycle( Fraig_MemFixed_t * p, char * pEntry );

// extern void                Fraig_MemFixedRestart( Fraig_MemFixed_t * p );

// extern int                 Fraig_MemFixedReadMemUsage( Fraig_MemFixed_t * p );

/*=== fraigNode.c =============================================================*/
// extern Fraig_Node_t *      Fraig_NodeCreateConst( Fraig_Man_t * p );

// extern Fraig_Node_t *      Fraig_NodeCreatePi( Fraig_Man_t * p );

// extern Fraig_Node_t *      Fraig_NodeCreate( Fraig_Man_t * p, Fraig_Node_t * p1, Fraig_Node_t * p2 );

// extern void                Fraig_NodeSimulate( Fraig_Node_t * pNode, int iWordStart, int iWordStop, int fUseRand );

/*=== fraigPrime.c =============================================================*/

/*=== fraigSat.c ===============================================================*/
// extern int                 Fraig_NodeIsImplication( Fraig_Man_t * p, Fraig_Node_t * pOld, Fraig_Node_t * pNew, int nBTLimit );

/*=== fraigTable.c =============================================================*/
// extern Fraig_HashTable_t * Fraig_HashTableCreate( int nSize );

// extern void                Fraig_HashTableFree( Fraig_HashTable_t * p );

// extern int                 Fraig_HashTableLookupS( Fraig_Man_t * pMan, Fraig_Node_t * p1, Fraig_Node_t * p2, Fraig_Node_t ** ppNodeRes );

// extern Fraig_Node_t *      Fraig_HashTableLookupF( Fraig_Man_t * pMan, Fraig_Node_t * pNode );

// extern Fraig_Node_t *      Fraig_HashTableLookupF0( Fraig_Man_t * pMan, Fraig_Node_t * pNode );

// extern void                Fraig_HashTableInsertF0( Fraig_Man_t * pMan, Fraig_Node_t * pNode );

// extern int                 Fraig_CompareSimInfo( Fraig_Node_t * pNode1, Fraig_Node_t * pNode2, int iWordLast, int fUseRand );

// extern int                 Fraig_CompareSimInfoUnderMask( Fraig_Node_t * pNode1, Fraig_Node_t * pNode2, int iWordLast, int fUseRand, unsigned * puMask );

// extern int                 Fraig_FindFirstDiff( Fraig_Node_t * pNode1, Fraig_Node_t * pNode2, int fCompl, int iWordLast, int fUseRand );

// extern void                Fraig_CollectXors( Fraig_Node_t * pNode1, Fraig_Node_t * pNode2, int iWordLast, int fUseRand, unsigned * puMask );

// extern void                Fraig_TablePrintStatsS( Fraig_Man_t * pMan );

// extern void                Fraig_TablePrintStatsF( Fraig_Man_t * pMan );

// extern void                Fraig_TablePrintStatsF0( Fraig_Man_t * pMan );

// extern int                 Fraig_TableRehashF0( Fraig_Man_t * pMan, int fLinkEquiv );

/*=== fraigUtil.c ===============================================================*/
// extern int                 Fraig_NodeCountPis( Msat_IntVec_t * vVars, int nVarsPi );

// extern int                 Fraig_NodeCountSuppVars( Fraig_Man_t * p, Fraig_Node_t * pNode, int fSuppStr );

// extern int                 Fraig_NodesCompareSupps( Fraig_Man_t * p, Fraig_Node_t * pOld, Fraig_Node_t * pNew );

// extern int                 Fraig_NodeAndSimpleCase_rec( Fraig_Node_t * pOld, Fraig_Node_t * pNew );

// extern int                 Fraig_NodeIsExorType( Fraig_Node_t * pNode );

// extern void                Fraig_ManSelectBestChoice( Fraig_Man_t * p );

// extern int                 Fraig_BitStringCountOnes( unsigned * pString, int nWords );

// extern void                Fraig_PrintBinary( FILE * pFile, unsigned * pSign, int nBits );

// extern int                 Fraig_NodeIsExorType( Fraig_Node_t * pNode );

// extern int                 Fraig_NodeIsExor( Fraig_Node_t * pNode );

// extern int                 Fraig_NodeIsMuxType( Fraig_Node_t * pNode );

// extern Fraig_Node_t *      Fraig_NodeRecognizeMux( Fraig_Node_t * pNode, Fraig_Node_t ** ppNodeT, Fraig_Node_t ** ppNodeE );

// extern int                 Fraig_ManCountExors( Fraig_Man_t * pMan );

// extern int                 Fraig_ManCountMuxes( Fraig_Man_t * pMan );

// extern int                 Fraig_NodeSimsContained( Fraig_Man_t * pMan, Fraig_Node_t * pNode1, Fraig_Node_t * pNode2 );

// extern int                 Fraig_NodeIsInSupergate( Fraig_Node_t * pOld, Fraig_Node_t * pNew );

// extern Fraig_NodeVec_t *   Fraig_CollectSupergate( Fraig_Node_t * pNode, int fStopAtMux );

// extern int                 Fraig_CountPis( Fraig_Man_t * p, Msat_IntVec_t * vVarNums );

// extern void                Fraig_ManIncrementTravId( Fraig_Man_t * pMan );

// extern void                Fraig_NodeSetTravIdCurrent( Fraig_Man_t * pMan, Fraig_Node_t * pNode );

// extern int                 Fraig_NodeIsTravIdCurrent( Fraig_Man_t * pMan, Fraig_Node_t * pNode );

// extern int                 Fraig_NodeIsTravIdPrevious( Fraig_Man_t * pMan, Fraig_Node_t * pNode );

/*=== fraigVec.c ===============================================================*/
// extern void                Fraig_NodeVecSortByRefCount( Fraig_NodeVec_t * p );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
