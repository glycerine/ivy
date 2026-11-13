package bmc

/**CFile****************************************************************

  FileName    [bmc.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmc.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/saig/saig.h"
//#include "aig/gia/gia.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

//#define USE_NODE_ORDER 1

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// exact synthesis parameters

type Bmc_EsPar_t struct {
	nVars       int32
	nNodes      int32
	nLutSize    int32
	nMajSupp    int32
	fMajority   int32
	fUseIncr    int32
	fOnlyAnd    int32
	fGlucose    int32
	fOrderNodes int32
	fEnumSols   int32
	fFewerVars  int32
	RuntimeLim  int32
	fVerbose    int32
	pTtStr      string
}

func Bmc_EsParSetDefault(pPars *Bmc_EsPar_t) {
	*pPars = Bmc_EsPar_t{}
	pPars.nVars = 0
	pPars.nNodes = 0
	pPars.nLutSize = 2
	pPars.nMajSupp = 0
	pPars.fMajority = 0
	pPars.fUseIncr = 0
	pPars.fOnlyAnd = 0
	pPars.fGlucose = 0
	pPars.fOrderNodes = 0
	pPars.fEnumSols = 0
	pPars.fFewerVars = 0
	pPars.RuntimeLim = 0
	pPars.fVerbose = 1
}

// unrolling manager
type Unr_Man_t = Unr_Man_t_

type Bmc_AndPar_t struct {
	nStart        int32 // starting timeframe
	nFramesMax    int32 // maximum number of timeframes
	nFramesAdd    int32 // the number of additional frames
	nConfLimit    int32 // maximum number of conflicts at a node
	nTimeOut      int32 // timeout in seconds
	nLutSize      int32 // LUT size for cut computation
	nProcs        int32 // the number of parallel solvers
	fLoadCnf      int32 // dynamic CNF loading
	fDumpFrames   int32 // dump unrolled timeframes
	fUseSynth     int32 // use synthesis
	fUseOldCnf    int32 // use old CNF construction
	fUseGlucose   int32 // use Glucose 3.0 as the default solver
	fUseEliminate int32 // use variable elimination
	fVerbose      int32 // verbose
	fVeryVerbose  int32 // very verbose
	fNotVerbose   int32 // skip line-by-line print-out
	iFrame        int32 // explored up to this frame
	nFailOuts     int32 // the number of failed outputs
	nDropOuts     int32 // the number of dropped outputs

	pFuncOnFrameDone func(int32, int32, int32) // callback on each frame status (frame, po, status)
}

type Bmc_BCorePar_t struct {
	iFrame      int32  // timeframe
	iOutput     int32  // property output
	nTimeOut    int32  // timeout in seconds
	pFilePivots string // file name with AIG IDs of pivot objects
	pFileProof  string // file name to write the resulting proof
	fVerbose    int32  // verbose output
}

type Bmc_MulPar_t struct {
	TimeOutGlo   int32
	TimeOutLoc   int32
	TimeOutInc   int32
	TimeOutGap   int32
	TimePerOut   int32
	fUseSyn      int32
	fDumpFinal   int32
	fVerbose     int32
	fVeryVerbose int32
}

type Bmc_ParFf_t struct {
	pFileName      string
	pFormStr       string
	Algo           int32
	fComplVars     int32
	fStartPats     int32
	nTimeOut       int32
	nIterCheck     int32
	nCardConstr    int32
	fNonStrict     int32
	fBasic         int32
	fFfOnly        int32
	fCheckUntest   int32
	fDump          int32
	fDumpDelay     int32
	fDumpUntest    int32
	fDumpNewFaults int32
	fVerbose       int32
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== bmcBCore.c ==========================================================*/
// extern void              Bmc_ManBCorePerform( Gia_Man_t * pGia, Bmc_BCorePar_t * pPars );

/*=== bmcBmc.c ==========================================================*/
// extern int               Saig_ManBmcSimple( Aig_Man_t * pAig, int nFrames, int nSizeMax, int nBTLimit, int fRewrite, int fVerbose, int * piFrame, int nCofFanLit, int fUseSatoko );

/*=== bmcBmc2.c ==========================================================*/
// extern int               Saig_BmcPerform( Aig_Man_t * pAig, int nStart, int nFramesMax, int nNodesMax, int nTimeOut, int nConfMaxOne, int nConfMaxAll, int fVerbose, int fVerbOverwrite, int * piFrames, int fSilent, int fUseSatoko );

/*=== bmcBmc3.c ==========================================================*/
// extern void              Saig_ParBmcSetDefaultParams( Saig_ParBmc_t * p );

// extern int               Saig_ManBmcScalable( Aig_Man_t * pAig, Saig_ParBmc_t * pPars );

/*=== bmcBmcAnd.c ==========================================================*/
// extern int               Gia_ManBmcPerform( Gia_Man_t * p, Bmc_AndPar_t * pPars );

/*=== bmcCexCare.c ==========================================================*/
// extern Abc_Cex_t *       Bmc_CexCareExtendToObjects( Gia_Man_t * p, Abc_Cex_t * pCex, Abc_Cex_t * pCexCare );

// extern Abc_Cex_t *       Bmc_CexCareMinimize( Aig_Man_t * p, int nRealPis, Abc_Cex_t * pCex, int nTryCexes, int fCheck, int fVerbose );

// extern Abc_Cex_t *       Bmc_CexCareMinimizeAig( Gia_Man_t * p, int nRealPis, Abc_Cex_t * pCex, int nTryCexes, int fCheck, int fVerbose );

// extern void              Bmc_CexCareVerify( Aig_Man_t * p, Abc_Cex_t * pCex, Abc_Cex_t * pCexMin, int fVerbose );

// extern Abc_Cex_t *       Bmc_CexCareSatBasedMinimize( Aig_Man_t * p, int nRealPis, Abc_Cex_t * pCex, int fHighEffort, int fCheck, int fVerbose );

// extern Abc_Cex_t *       Bmc_CexCareSatBasedMinimizeAig( Gia_Man_t * p, Abc_Cex_t * pCex, int fHighEffort, int fVerbose );

/*=== bmcCexCut.c ==========================================================*/
// extern Gia_Man_t *       Bmc_GiaTargetStates( Gia_Man_t * p, Abc_Cex_t * pCex, int iFrBeg, int iFrEnd, int fCombOnly, int fGenAll, int fAllFrames, int fVerbose );

// extern Aig_Man_t *       Bmc_AigTargetStates( Aig_Man_t * p, Abc_Cex_t * pCex, int iFrBeg, int iFrEnd, int fCombOnly, int fGenAll, int fAllFrames, int fVerbose );

/*=== bmcCexMin.c ==========================================================*/
// extern Abc_Cex_t *       Saig_ManCexMinPerform( Aig_Man_t * pAig, Abc_Cex_t * pCex );

/*=== bmcCexTool.c ==========================================================*/
// extern void              Bmc_CexPrint( Abc_Cex_t * pCex, int nRealPis, int fVerbose );

// extern int               Bmc_CexVerify( Gia_Man_t * p, Abc_Cex_t * pCex, Abc_Cex_t * pCexCare );

/*=== bmcICheck.c ==========================================================*/
// extern void              Bmc_PerformICheck( Gia_Man_t * p, int nFramesMax, int nTimeOut, int fEmpty, int fVerbose );

// extern Vec_Int_t *       Bmc_PerformISearch( Gia_Man_t * p, int nFramesMax, int nTimeOut, int fReverse, int fBackTopo, int fDump, int fVerbose );

/*=== bmcUnroll.c ==========================================================*/
// extern Unr_Man_t *       Unr_ManUnrollStart( Gia_Man_t * pGia, int fVerbose );

// extern Gia_Man_t *       Unr_ManUnrollFrame( Unr_Man_t * p, int f );

// extern void              Unr_ManFree( Unr_Man_t * p );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
