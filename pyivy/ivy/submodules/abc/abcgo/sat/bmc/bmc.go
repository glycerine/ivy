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
	NVars       int32
	NNodes      int32
	NLutSize    int32
	NMajSupp    int32
	FMajority   int32
	FUseIncr    int32
	FOnlyAnd    int32
	FGlucose    int32
	FOrderNodes int32
	FEnumSols   int32
	FFewerVars  int32
	RuntimeLim  int32
	FVerbose    int32
	PTtStr      string
}

func Bmc_EsParSetDefault(pPars *Bmc_EsPar_t) {
	*pPars = Bmc_EsPar_t{}
	pPars.NVars = 0
	pPars.NNodes = 0
	pPars.NLutSize = 2
	pPars.NMajSupp = 0
	pPars.FMajority = 0
	pPars.FUseIncr = 0
	pPars.FOnlyAnd = 0
	pPars.FGlucose = 0
	pPars.FOrderNodes = 0
	pPars.FEnumSols = 0
	pPars.FFewerVars = 0
	pPars.RuntimeLim = 0
	pPars.FVerbose = 1
}

// unrolling manager
type Unr_Man_t = Unr_Man_t_

type Bmc_AndPar_t struct {
	NStart        int32 // starting timeframe
	NFramesMax    int32 // maximum number of timeframes
	NFramesAdd    int32 // the number of additional frames
	NConfLimit    int32 // maximum number of conflicts at a node
	NTimeOut      int32 // timeout in seconds
	NLutSize      int32 // LUT size for cut computation
	NProcs        int32 // the number of parallel solvers
	FLoadCnf      int32 // dynamic CNF loading
	FDumpFrames   int32 // dump unrolled timeframes
	FUseSynth     int32 // use synthesis
	FUseOldCnf    int32 // use old CNF construction
	FUseGlucose   int32 // use Glucose 3.0 as the default solver
	FUseEliminate int32 // use variable elimination
	FVerbose      int32 // verbose
	FVeryVerbose  int32 // very verbose
	FNotVerbose   int32 // skip line-by-line print-out
	IFrame        int32 // explored up to this frame
	NFailOuts     int32 // the number of failed outputs
	NDropOuts     int32 // the number of dropped outputs

	PFuncOnFrameDone func(int32, int32, int32) // callback on each frame status (frame, po, status)
}

type Bmc_BCorePar_t struct {
	IFrame      int32  // timeframe
	IOutput     int32  // property output
	NTimeOut    int32  // timeout in seconds
	PFilePivots string // file name with AIG IDs of pivot objects
	PFileProof  string // file name to write the resulting proof
	FVerbose    int32  // verbose output
}

type Bmc_MulPar_t struct {
	TimeOutGlo   int32
	TimeOutLoc   int32
	TimeOutInc   int32
	TimeOutGap   int32
	TimePerOut   int32
	FUseSyn      int32
	FDumpFinal   int32
	FVerbose     int32
	FVeryVerbose int32
}

type Bmc_ParFf_t struct {
	PFileName      string
	PFormStr       string
	Algo           int32
	fComplVars     int32
	FStartPats     int32
	NTimeOut       int32
	NIterCheck     int32
	NCardConstr    int32
	FNonStrict     int32
	FBasic         int32
	FFfOnly        int32
	FCheckUntest   int32
	FDump          int32
	FDumpDelay     int32
	FDumpUntest    int32
	FDumpNewFaults int32
	FVerbose       int32
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
