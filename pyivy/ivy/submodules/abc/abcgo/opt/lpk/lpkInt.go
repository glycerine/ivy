package lpk

import "github.com/glycerine/abcgo/bool/kit"

/**CFile****************************************************************

  FileName    [lpkInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkInt.h,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "base/abc/abc.h"
//#include "bool/kit/kit.h"
//#include "map/if/if.h"
//#include "lpk.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

const LPK_SIZE_MAX = 100   // the largest size of the function
const LPK_CUTS_MAX = 10000 // the largest number of cuts considered

type Lpk_Man_t = Lpk_Man_t_

type Lpk_Cut_t struct {
	nLeaves   uint32              /* C bitfield: 6 bits; review packing and truncation */ // (L) the number of leaves
	nNodes    uint32              /* C bitfield: 6 bits; review packing and truncation */ // (M) the number of nodes
	nNodesDup uint32              /* C bitfield: 6 bits; review packing and truncation */ // (Q) nodes outside of MFFC
	nLuts     uint32              /* C bitfield: 6 bits; review packing and truncation */ // (N) the number of LUTs to try
	unused    uint32              /* C bitfield: 6 bits; review packing and truncation */ // unused
	fHasDsd   uint32              /* C bitfield: 1 bits; review packing and truncation */ // set to 1 if the cut has structural DSD (and so cannot be used)
	fMark     uint32              /* C bitfield: 1 bits; review packing and truncation */ // multipurpose mark
	uSign     [2]uint32           // the signature
	Weight    float32             // the weight of the cut: (M - Q)/N(V)   (the larger the better)
	Gain      int32               // the gain achieved using this cut
	pLeaves   [LPK_SIZE_MAX]int32 // the leaves of the cut
	pNodes    [LPK_SIZE_MAX]int32 // the nodes of the cut
}

type Lpk_Man_t_ struct {
	// parameters
	pPars   *Lpk_Par_t // the set of parameters
	pAbcAPI *Lpk_AbcAPI
	// current representation
	pNtk *Abc_Ntk_t // the network
	pObj *Abc_Obj_t // the node to resynthesize
	// cut representation
	nMffc    int32                   // the size of MFFC of the node
	nCuts    int32                   // the total number of cuts
	nCutsMax int32                   // the largest possible number of cuts
	nEvals   int32                   // the number of good cuts
	pCuts    [LPK_CUTS_MAX]Lpk_Cut_t // the storage for cuts
	pEvals   [LPK_CUTS_MAX]int32     // the good cuts
	// visited nodes
	vVisited *Vec_Vec_t
	// mapping manager
	pIfMan  *If_Man_t
	vCover  *Vec_Int_t
	vLevels *Vec_Vec_t
	// temporary variables
	fCofactoring int32               // working in the cofactoring mode
	fCalledOnce  int32               // limits the depth of MUX cofactoring
	nCalledSRed  int32               // the number of called to SRed
	pRefs        [LPK_SIZE_MAX]int32 // fanin reference counters
	pCands       [LPK_SIZE_MAX]int32 // internal nodes pointing only to the leaves
	vLeaves      *Vec_Ptr_t
	vTemp        *Vec_Ptr_t
	// truth table representation
	vTtElems    *Vec_Ptr_t // elementary truth tables
	vTtNodes    *Vec_Ptr_t // storage for temporary truth tables of the nodes
	graphMemory []uint32   // reusable Kit_TruthToGraph scratch (C's vCover workspace)
	vMemory     *Vec_Int_t
	vBddDir     *Vec_Int_t
	vBddInv     *Vec_Int_t
	puSupps     [32]uint32 // the supports of the cofactors
	ppTruths    [5][16][]uint32
	// variable sets
	vSets   [8]*Vec_Int_t
	pDsdMan *Kit_DsdMan_t
	// statistics
	nNodesTotal  int32 // total number of nodes
	nNodesOver   int32 // nodes with cuts over the limit
	nCutsTotal   int32 // total number of cuts
	nCutsUseful  int32 // useful cuts
	nGainTotal   int32 // the gain in LUTs
	nChanges     int32 // the number of changed nodes
	nBenefited   int32 // the number of gainful that benefited from decomposition
	nMuxes       int32
	nDsds        int32
	nTotalNets   int32
	nTotalNets2  int32
	nTotalNodes  int32
	nTotalNodes2 int32
	// counter of non-DSD blocks
	nBlocks [17]int32
	// runtime
	timeCuts   abctime
	timeTruth  abctime
	timeSupps  abctime
	timeTruth2 abctime
	timeTruth3 abctime
	timeEval   abctime
	timeMap    abctime
	timeOther  abctime
	timeTotal  abctime
	// runtime of eval
	timeEvalMuxAn abctime
	timeEvalMuxSp abctime
	timeEvalDsdAn abctime
	timeEvalDsdSp abctime
}

// internal representation of the function to be decomposed

type Lpk_Fun_t struct {
	vNodes    *Vec_Ptr_t // the array of leaves and decomposition nodes
	Id        uint32     /* C bitfield: 7 bits; review packing and truncation */  // the ID of this node
	nVars     uint32     /* C bitfield: 5 bits; review packing and truncation */  // the number of variables
	nLutK     uint32     /* C bitfield: 4 bits; review packing and truncation */  // the number of LUT inputs
	nAreaLim  uint32     /* C bitfield: 14 bits; review packing and truncation */ // the area limit (the largest allowed)
	fSupports uint32     /* C bitfield: 1 bits; review packing and truncation */  // supports of cofactors were precomputed
	fMark     uint32     /* C bitfield: 1 bits; review packing and truncation */  // marks the MUX-based dec
	uSupp     uint32     // the support of this component
	puSupps   [32]uint32 // the supports of the cofactors
	nDelayLim uint32     // the delay limit (the largest allowed)
	pDelays   [16]int32  // the delays of the inputs
	pFanins   [16]byte   // the fanins of this function
	truth     []uint32   // the truth table and two cofactor workspaces
}

// preliminary decomposition result

type Lpk_Res_t struct {
	nBSVars    int32   // the number of bound set variables
	BSVars     uint32  // the bound set
	nCofVars   int32   // the number of cofactoring variables
	pCofVars   [4]byte // the cofactoring variables
	nSuppSizeS int32   // support size of the smaller (decomposed) function
	nSuppSizeL int32   // support size of the larger (composition) function
	DelayEst   int32   // estimated delay of the decomposition
	AreaEst    int32   // estimated area of the decomposition
	Variable   int32   // variable in MUX decomposition
	Polarity   int32   // polarity in MUX decomposition
}

func Lpk_LutNumVars(nLutsLim, nLutK int32) int32 {
	if nLutsLim <= 0 || nLutK < 2 {
		return 0
	}
	return nLutsLim*(nLutK-1) + 1
}

func Lpk_LutNumLuts(NVarsMax, nLutK int32) int32 {
	if NVarsMax <= 1 {
		return 0
	}
	if nLutK < 2 {
		return 0
	}
	divisor := nLutK - 1
	count := (NVarsMax - 1) / divisor
	if (NVarsMax-1)%divisor != 0 {
		count++
	}
	return count
}
func Lpk_FunTruth(p *Lpk_Fun_t, Num int32) []uint32 {
	if p == nil || Num < 0 || Num >= 3 {
		panic("lpk: invalid truth table index")
	}
	words := int(kit.Kit_TruthWordNum(int32(p.nVars)))
	start := int(Num) * words
	if start+words > len(p.truth) {
		panic("lpk: truth table storage is undersized")
	}
	return p.truth[start : start+words]
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                           ITERATORS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Lpk_CutForEachLeaf( pNtk, pCut, pObj, i )                                        \
//     for ( i = 0; (i < (pCut)->nLeaves) && (((pObj) = Abc_NtkObj(pNtk, (pCut)->pLeaves[i])), 1); i++ )

// Original macro:
// #define Lpk_CutForEachNode( pNtk, pCut, pObj, i )                                        \
//     for ( i = 0; (i < (pCut)->nNodes) && (((pObj) = Abc_NtkObj(pNtk, (pCut)->pNodes[i])), 1); i++ )

// Original macro:
// #define Lpk_CutForEachNodeReverse( pNtk, pCut, pObj, i )                                 \
//     for ( i = (pCut)->nNodes - 1; (i >= 0) && (((pObj) = Abc_NtkObj(pNtk, (pCut)->pNodes[i])), 1); i-- )

// Original macro:
// #define Lpk_SuppForEachVar( Supp, Var )\
//     for ( Var = 0; Var < 16; Var++ )\
//         if ( !(Supp & (1<<Var)) ) {} else

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== lpkAbcDec.c ============================================================*/
// extern Abc_Obj_t *    Lpk_Decompose( Lpk_Man_t * pMan, Abc_Ntk_t * pNtk, Vec_Ptr_t * vLeaves, unsigned * pTruth, unsigned * puSupps, int nLutK, int AreaLim, int DelayLim );

/*=== lpkAbcDsd.c ============================================================*/
// extern Lpk_Res_t *    Lpk_DsdAnalize( Lpk_Man_t * pMan, Lpk_Fun_t * p, int nShared );

// extern Lpk_Fun_t *    Lpk_DsdSplit( Lpk_Man_t * pMan, Lpk_Fun_t * p, char * pCofVars, int nCofVars, unsigned uBoundSet );

/*=== lpkAbcMux.c ============================================================*/
// extern Lpk_Res_t *    Lpk_MuxAnalize( Lpk_Man_t * pMan, Lpk_Fun_t * p );

// extern Lpk_Fun_t *    Lpk_MuxSplit( Lpk_Man_t * pMan, Lpk_Fun_t * p, int Var, int Pol );

/*=== lpkAbcUtil.c ============================================================*/
// extern Lpk_Fun_t *    Lpk_FunAlloc( int nVars );

// extern void           Lpk_FunFree( Lpk_Fun_t * p );

// extern Lpk_Fun_t *    Lpk_FunCreate( Abc_Ntk_t * pNtk, Vec_Ptr_t * vLeaves, unsigned * pTruth, int nLutK, int AreaLim, int DelayLim );

// extern Lpk_Fun_t *    Lpk_FunDup( Lpk_Fun_t * p, unsigned * pTruth );

// extern int            Lpk_FunSuppMinimize( Lpk_Fun_t * p );

// extern void           Lpk_FunComputeCofSupps( Lpk_Fun_t * p );

// extern int            Lpk_SuppDelay( unsigned uSupp, int * pDelays );

// extern int            Lpk_SuppToVars( unsigned uBoundSet, char * pVars );

/*=== lpkCut.c =========================================================*/
// extern unsigned *     Lpk_CutTruth( Lpk_Man_t * p, Lpk_Cut_t * pCut, int fInv );

// extern int            Lpk_NodeCuts( Lpk_Man_t * p );

/*=== lpkMap.c =========================================================*/
// extern Lpk_Man_t *    Lpk_ManStart( Lpk_Par_t * pPars );

// extern void           Lpk_ManStop( Lpk_Man_t * p );

/*=== lpkMap.c =========================================================*/
// extern If_Obj_t *     Lpk_MapPrime( Lpk_Man_t * p, unsigned * pTruth, int nVars, If_Obj_t ** ppLeaves );

// extern If_Obj_t *     Lpk_MapTree_rec( Lpk_Man_t * p, Kit_DsdNtk_t * pNtk, If_Obj_t ** ppLeaves[:], int iLit, If_Obj_t * pResult );

/*=== lpkMulti.c =======================================================*/
// extern If_Obj_t *     Lpk_MapTreeMulti( Lpk_Man_t * p, unsigned * pTruth, int nLeaves, If_Obj_t ** ppLeaves );

/*=== lpkMux.c =========================================================*/
// extern If_Obj_t *     Lpk_MapTreeMux_rec( Lpk_Man_t * p, unsigned * pTruth, int nVars, If_Obj_t ** ppLeaves );

// extern If_Obj_t *     Lpk_MapSuppRedDec_rec( Lpk_Man_t * p, unsigned * pTruth, int nVars, If_Obj_t ** ppLeaves );

/*=== lpkSets.c =========================================================*/
// extern unsigned       Lpk_MapSuppRedDecSelect( Lpk_Man_t * p, unsigned * pTruth, int nVars, int * piVar, int * piVarReused );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
