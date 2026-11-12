package res

import "math/rand"

/**CFile****************************************************************

  FileName    [resInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resynthesis package.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 15, 2007.]

  Revision    [$Id: resInt.h,v 1.00 2007/01/15 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "res.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Res_Win_t struct {
	// windowing parameters
	pNode      *Abc_Obj_t // the node in the center
	nWinTfiMax int32      // the fanin levels
	nWinTfoMax int32      // the fanout levels
	nLevDivMax int32      // the maximum divisor level
	// internal windowing parameters
	nFanoutLimit int32 // the limit on the fanout count of a TFO node (if more, the node is treated as a root)
	nLevTfiMinus int32 // the number of additional levels to search from TFO below the level of leaves
	// derived windowing parameters
	nLevLeafMin int32 // the minimum level of a leaf
	nLevTravMin int32 // the minimum level to search from TFO
	nDivsPlus   int32 // the number of additional divisors
	// the window data
	vRoots    *Vec_Ptr_t // outputs of the window
	vLeaves   *Vec_Ptr_t // inputs of the window
	vBranches *Vec_Ptr_t // side nodes of the window
	vNodes    *Vec_Ptr_t // internal nodes of the window
	vDivs     *Vec_Ptr_t // candidate divisors of the node
	// temporary data
	vMatrix *Vec_Vec_t // TFI nodes below the given node
}

type Res_Sim_t struct {
	pAig     *Abc_Ntk_t // AIG for simulation
	nTruePis int32      // the number of true PIs of the window
	fConst0  int32      // the node can be replaced by constant 0
	fConst1  int32      // the node can be replaced by constant 0
	// simulation parameters
	nWords    int32 // the number of simulation words
	nPats     int32 // the number of patterns
	nWordsIn  int32 // the number of simulation words in the input patterns
	nPatsIn   int32 // the number of patterns in the input patterns
	nBytesIn  int32 // the number of bytes in the input patterns
	nWordsOut int32 // the number of simulation words in the output patterns
	nPatsOut  int32 // the number of patterns in the output patterns
	// simulation info
	vPats  *Vec_Ptr_t // input simulation patterns
	vPats0 *Vec_Ptr_t // input simulation patterns
	vPats1 *Vec_Ptr_t // input simulation patterns
	vOuts  *Vec_Ptr_t // output simulation info
	nPats0 int32      // the number of 0-patterns accumulated
	nPats1 int32      // the number of 1-patterns accumulated
	// resub candidates
	vCands *Vec_Vec_t // resubstitution candidates
	// statistics
	timeSat abctime
	rng *rand.Rand
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== resDivs.c ==========================================================*/
// extern void          Res_WinDivisors( Res_Win_t * p, int nLevDivMax );

// extern void          Res_WinSweepLeafTfo_rec( Abc_Obj_t * pObj, int nLevelLimit );

// extern int           Res_WinVisitMffc( Abc_Obj_t * pNode );

/*=== resFilter.c ==========================================================*/
// extern int           Res_FilterCandidates( Res_Win_t * pWin, Abc_Ntk_t * pAig, Res_Sim_t * pSim, Vec_Vec_t * vResubs, Vec_Vec_t * vResubsW, int nFaninsMax, int fArea );

// extern int           Res_FilterCandidatesArea( Res_Win_t * pWin, Abc_Ntk_t * pAig, Res_Sim_t * pSim, Vec_Vec_t * vResubs, Vec_Vec_t * vResubsW, int nFaninsMax );

/*=== resSat.c ==========================================================*/
// extern void *        Res_SatProveUnsat( Abc_Ntk_t * pAig, Vec_Ptr_t * vFanins );

// extern int           Res_SatSimulate( Res_Sim_t * p, int nPats, int fOnSet );

/*=== resSim.c ==========================================================*/
// extern Res_Sim_t *   Res_SimAlloc( int nWords );

// extern void          Res_SimFree( Res_Sim_t * p );

// extern int           Res_SimPrepare( Res_Sim_t * p, Abc_Ntk_t * pAig, int nTruePis, int fVerbose );

/*=== resStrash.c ==========================================================*/
// extern Abc_Ntk_t *   Res_WndStrash( Res_Win_t * p );

/*=== resWnd.c ==========================================================*/
// extern void          Res_UpdateNetwork( Abc_Obj_t * pObj, Vec_Ptr_t * vFanins, Hop_Obj_t * pFunc, Vec_Vec_t * vLevels );

/*=== resWnd.c ==========================================================*/
// extern Res_Win_t *   Res_WinAlloc();

// extern void          Res_WinFree( Res_Win_t * p );

// extern int           Res_WinIsTrivial( Res_Win_t * p );

// extern int           Res_WinCompute( Abc_Obj_t * pNode, int nWinTfiMax, int nWinTfoMax, Res_Win_t * p );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
