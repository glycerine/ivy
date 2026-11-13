package ssw

import (
	"github.com/glycerine/abcgo/misc/util"
)

/**CFile****************************************************************

  FileName    [ssw.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: ssw.h,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// choicing parameters

type Ssw_Pars_t struct {
	nPartSize      int32 // size of the partition
	nOverSize      int32 // size of the overlap between partitions
	nFramesK       int32 // the induction depth
	nFramesAddSim  int32 // the number of additional frames to simulate
	fConstrs       int32 // treat the last nConstrs POs as seq constraints
	fMergeFull     int32 // enables full merge when constraints are used
	nMaxLevs       int32 // the max number of levels of nodes to consider
	nBTLimit       int32 // conflict limit at a node
	nBTLimitGlobal int32 // conflict limit for multiple runs
	nMinDomSize    int32 // min clock domain considered for optimization
	nItersStop     int32 // stop after the given number of iterations
	fDumpSRInit    int32 // dumps speculative reduction
	nResimDelta    int32 // the number of nodes to resimulate
	nStepsMax      int32 // (scorr only) the max number of induction steps
	TimeLimit      int32 // time out in seconds
	fPolarFlip     int32 // uses polarity adjustment
	fLatchCorr     int32 // perform register correspondence
	fConstCorr     int32 // perform constant correspondence
	fOutputCorr    int32 // perform 'PO correspondence'
	fSemiFormal    int32 // enable semiformal filtering
	//    int              fUniqueness;   // enable uniqueness constraints
	fDynamic      int32 // enable dynamic addition of constraints
	fLocalSim     int32 // enable local simulation simulation
	fPartSigCorr  int32 // uses partial signal correspondence
	nIsleDist     int32 // extends islands by the given distance
	fScorrGia     int32 // new signal correspondence implementation
	fUseCSat      int32 // new SAT solver using when fScorrGia is selected
	fVerbose      int32 // verbose stats
	fFlopVerbose  int32 // verbose printout of redundant flops
	fEquivDump    int32 // enables dumping equivalences
	fEquivDump2   int32 // enables dumping equivalences
	fStopWhenGone int32 // stop when PO output is not a candidate constant
	// optimized latch correspondence
	fLatchCorrOpt int32 // perform register correspondence (optimized)
	nSatVarMax    int32 // max number of SAT vars before recycling SAT solver (optimized latch corr only)
	nRecycleCalls int32 // calls to perform before recycling SAT solver (optimized latch corr only)
	// optimized signal correspondence
	nSatVarMax2    int32 // max number of SAT vars before recycling SAT solver (optimized latch corr only)
	nRecycleCalls2 int32 // calls to perform before recycling SAT solver (optimized latch corr only)
	// internal parameters
	nIters     int32 // the number of iterations performed
	nConflicts int32 // the total number of conflicts performed
	// callback
	pData any
	pFunc any
}

type Ssw_RarPars_t struct {
	nFrames        int32
	nWords         int32
	nBinSize       int32
	nRounds        int32
	nRestart       int32
	nRandSeed      int32
	TimeOut        int32
	TimeOutGap     int32
	fSolveAll      int32
	fSetLastState  int32
	fVerbose       int32
	fNotVerbose    int32
	fSilent        int32
	fDropSatOuts   int32
	fMiter         int32
	fUseCex        int32
	fLatchOnly     int32
	fUseFfGrouping int32
	nSolved        int32
	pCex           *util.Abc_Cex_t
	pFuncOnFail    func(int32, *util.Abc_Cex_t) int32 // called for a failed output in MO mode
}

type Ssw_Sml_t = Ssw_Sml_t_ // sequential simulation manager

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== sswBmc.c ==========================================================*/
// extern int           Ssw_BmcDynamic( Aig_Man_t * pAig, int nFramesMax, int nConfLimit, int fVerbose, int * piFrame );

/*=== sswConstr.c ==========================================================*/
// extern int           Ssw_ManSetConstrPhases( Aig_Man_t * p, int nFrames, Vec_Int_t ** pvInits );

/*=== sswCore.c ==========================================================*/
// extern void          Ssw_ManSetDefaultParams( Ssw_Pars_t * p );

// extern void          Ssw_ManSetDefaultParamsLcorr( Ssw_Pars_t * p );

// extern Aig_Man_t *   Ssw_SignalCorrespondence( Aig_Man_t * pAig, Ssw_Pars_t * pPars );

// extern Aig_Man_t *   Ssw_LatchCorrespondence( Aig_Man_t * pAig, Ssw_Pars_t * pPars );

/*=== sswIslands.c ==========================================================*/
// extern int           Ssw_SecWithSimilarityPairs( Aig_Man_t * p0, Aig_Man_t * p1, Vec_Int_t * vPairs, Ssw_Pars_t * pPars );

// extern int           Ssw_SecWithSimilarity( Aig_Man_t * p0, Aig_Man_t * p1, Ssw_Pars_t * pPars );

/*=== sswMiter.c ===================================================*/
/*=== sswPart.c ==========================================================*/
// extern Aig_Man_t *   Ssw_SignalCorrespondencePart( Aig_Man_t * pAig, Ssw_Pars_t * pPars );

/*=== sswPairs.c ===================================================*/
// extern int           Ssw_MiterStatus( Aig_Man_t * p, int fVerbose );

// extern int           Ssw_SecWithPairs( Aig_Man_t * pAig1, Aig_Man_t * pAig2, Vec_Int_t * vIds1, Vec_Int_t * vIds2, Ssw_Pars_t * pPars );

// extern int           Ssw_SecGeneral( Aig_Man_t * pAig1, Aig_Man_t * pAig2, Ssw_Pars_t * pPars );

// extern int           Ssw_SecGeneralMiter( Aig_Man_t * pMiter, Ssw_Pars_t * pPars );

/*=== sswRarity.c ===================================================*/
// extern void          Ssw_RarSetDefaultParams( Ssw_RarPars_t * p );

// extern int           Ssw_RarSignalFilter( Aig_Man_t * pAig, Ssw_RarPars_t * pPars );

// extern int           Ssw_RarSimulate( Aig_Man_t * pAig, Ssw_RarPars_t * pPars );

/*=== sswSim.c ===================================================*/
// extern Ssw_Sml_t *   Ssw_SmlSimulateComb( Aig_Man_t * pAig, int nWords );

// extern Ssw_Sml_t *   Ssw_SmlSimulateSeq( Aig_Man_t * pAig, int nPref, int nFrames, int nWords );

// extern void          Ssw_SmlUnnormalize( Ssw_Sml_t * p );

// extern void          Ssw_SmlStop( Ssw_Sml_t * p );

// extern int           Ssw_SmlNumFrames( Ssw_Sml_t * p );

// extern int           Ssw_SmlNumWordsTotal( Ssw_Sml_t * p );

// extern unsigned *    Ssw_SmlSimInfo( Ssw_Sml_t * p, Aig_Obj_t * pObj );

// extern int           Ssw_SmlObjsAreEqualWord( Ssw_Sml_t * p, Aig_Obj_t * pObj0, Aig_Obj_t * pObj1 );

// extern void          Ssw_SmlInitializeSpecial( Ssw_Sml_t * p, Vec_Int_t * vInit );

// extern int           Ssw_SmlCheckNonConstOutputs( Ssw_Sml_t * p );

// extern Vec_Ptr_t *   Ssw_SmlSimDataPointers( Ssw_Sml_t * p );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
