package cec

import "time"

/**CFile****************************************************************

  FileName    [cec.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cec.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

// dynamic SAT parameters

type Cec_ParSat_t struct {
	deadline      time.Time
	SolverType    int32 // SAT solver type
	NBTLimit      int32 // conflict limit at a node
	NSatVarMax    int32 // the max number of SAT variables
	NCallsRecycle int32 // calls to perform before recycling SAT solver
	FNonChrono    int32 // use non-chronological backtracling (for circuit SAT only)
	fPolarFlip    int32 // flops polarity of variables
	FCheckMiter   int32 // the circuit is the miter
	//    int              fFirstStop;    // stop on the first sat output
	FLearnCls  int32 // perform clause learning
	FSaveCexes int32 // saves counter-examples
	FVerbose   int32 // verbose stats
}

// simulation parameters

type Cec_ParSim_t struct {
	deadline    time.Time
	NWords      int32 // the number of simulation words
	NFrames     int32 // the number of simulation frames
	NRounds     int32 // the number of simulation rounds
	NNonRefines int32 // the max number of rounds without refinement
	TimeLimit   int32 // the runtime limit in seconds
	FDualOut    int32 // miter with separate outputs
	FCheckMiter int32 // the circuit is the miter
	//    int              fFirstStop;    // stop on the first sat output
	FSeqSimulate int32 // performs sequential simulation
	fLatchCorr   int32 // consider only latch outputs
	fConstCorr   int32 // consider only constants
	fVeryVerbose int32 // verbose stats
	FVerbose     int32 // verbose stats
}

// semiformal parameters

type Cec_ParSmf_t struct {
	NWords      int32 // the number of simulation words
	NRounds     int32 // the number of simulation rounds
	NFrames     int32 // the max number of time frames
	NNonRefines int32 // the max number of rounds without refinement
	NMinOutputs int32 // the min outputs to accumulate
	NBTLimit    int32 // conflict limit at a node
	TimeLimit   int32 // the runtime limit in seconds
	FDualOut    int32 // miter with separate outputs
	FCheckMiter int32 // the circuit is the miter
	//    int              fFirstStop;    // stop on the first sat output
	FVerbose int32 // verbose stats
}

// combinational SAT sweeping parameters

type Cec_ParFra_t struct {
	Error         error // native driver validation/model errors
	JType         int32 // solver type
	NWords        int32 // the number of simulation words
	NRounds       int32 // the number of simulation rounds
	NItersMax     int32 // the maximum number of iterations of SAT sweeping
	NBTLimit      int32 // conflict limit at a node
	TimeLimit     int32 // the runtime limit in seconds
	NLevelMax     int32 // restriction on the level nodes to be swept
	NDepthMax     int32 // the depth in terms of steps of speculative reduction
	NCallsRecycle int32 // calls to perform before recycling SAT solver
	nSatVarMax    int32 // the max number of SAT variables
	NGenIters     int32 // pattern generation iterations
	FRewriting    int32 // enables AIG rewriting
	FCheckMiter   int32 // the circuit is the miter
	//    int              fFirstStop;    // stop on the first sat output
	FDualOut     int32 // miter with separate outputs
	fColorDiff   int32 // miter with separate outputs
	fSatSweeping int32 // enable SAT sweeping
	FRunCSat     int32 // enable another solver
	FUseCones    int32 // use cones
	fUseOrigIds  int32 // enable recording of original IDs
	FVeryVerbose int32 // verbose stats
	FVerbose     int32 // verbose stats
	iOutFail     int32 // the failed output
}

// combinational equivalence checking parameters

type Cec_ParCec_t struct {
	// Error reports invalid input or model validation failures from a verification run.
	Error     error
	NBTLimit  int32 // conflict limit at a node
	TimeLimit int32 // the runtime limit in seconds
	//    int              fFirstStop;    // stop on the first sat output
	fUseSmartCnf int32 // use smart CNF computation
	fRewriting   int32 // enables AIG rewriting
	FNaive       int32 // performs naive SAT-based checking
	FSilent      int32 // print no messages
	FVeryVerbose int32 // verbose stats
	FVerbose     int32 // verbose stats
	iOutFail     int32 // the number of failed output
}

// sequential register correspodence parameters

type Cec_ParCor_t struct {
	deadline     time.Time
	nWords       int32 // the number of simulation words
	nRounds      int32 // the number of simulation rounds
	NFrames      int32 // the number of time frames
	NPrefix      int32 // the number of time frames in the prefix
	NBTLimit     int32 // conflict limit at a node
	nLevelMax    int32 // (scorr only) the max number of levels
	nStepsMax    int32 // (scorr only) the max number of induction steps
	FLatchCorr   int32 // consider only latch outputs
	FConstCorr   int32 // consider only constants
	FUseRings    int32 // use rings
	FMakeChoices int32 // use equilvaences as choices
	FUseCSat     int32 // use circuit-based solver
	//    int              fFirstStop;    // stop on the first sat output
	fUseSmartCnf  int32 // use smart CNF computation
	FStopWhenGone int32 // quit when PO is not a candidate constant
	FVerboseFlops int32 // verbose stats
	fVeryVerbose  int32 // verbose stats
	FVerbose      int32 // verbose stats
	// callback
	pData any
	pFunc any
}

// sequential register correspodence parameters

type Cec_ParChc_t struct {
	Error        error // native choice validation/proof errors
	nWords       int32 // the number of simulation words
	nRounds      int32 // the number of simulation rounds
	NBTLimit     int32 // conflict limit at a node
	fUseRings    int32 // use rings
	FUseCSat     int32 // use circuit-based solver
	fVeryVerbose int32 // verbose stats
	FVerbose     int32 // verbose stats
}

// sequential synthesis parameters

type Cec_ParSeq_t struct {
	fUseLcorr    int32 // enables latch correspondence
	fUseScorr    int32 // enables signal correspondence
	nBTLimit     int32 // (scorr/lcorr) conflict limit at a node
	nFrames      int32 // (scorr/lcorr) the number of timeframes
	nLevelMax    int32 // (scorr only) the max number of levels
	fConsts      int32 // (scl only) merging constants
	fEquivs      int32 // (scl only) merging equivalences
	fUseMiniSat  int32 // enables MiniSat in lcorr/scorr
	nMinDomSize  int32 // the size of minimum clock domain
	fVeryVerbose int32 // verbose stats
	fVerbose     int32 // verbose stats
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== cecCec.c ==========================================================*/
// extern int           Cec_ManVerify( Gia_Man_t * p, Cec_ParCec_t * pPars );

// extern int           Cec_ManVerifyTwo( Gia_Man_t * p0, Gia_Man_t * p1, int fVerbose );

// extern int           Cec_ManVerifySimple( Gia_Man_t * p );

/*=== cecChoice.c ==========================================================*/
// extern Gia_Man_t *   Cec_ManChoiceComputation( Gia_Man_t * pAig, Cec_ParChc_t * pPars );

/*=== cecCorr.c ==========================================================*/
// extern int           Cec_ManLSCorrespondenceClasses( Gia_Man_t * pAig, Cec_ParCor_t * pPars );

// extern Gia_Man_t *   Cec_ManLSCorrespondence( Gia_Man_t * pAig, Cec_ParCor_t * pPars );

/*=== cecCore.c ==========================================================*/
// extern void          Cec_ManSatSetDefaultParams( Cec_ParSat_t * p );

// extern void          Cec_ManSimSetDefaultParams( Cec_ParSim_t * p );

// extern void          Cec_ManSmfSetDefaultParams( Cec_ParSmf_t * p );

// extern void          Cec_ManFraSetDefaultParams( Cec_ParFra_t * p );

// extern void          Cec_ManCecSetDefaultParams( Cec_ParCec_t * p );

// extern void          Cec_ManCorSetDefaultParams( Cec_ParCor_t * p );

// extern void          Cec_ManChcSetDefaultParams( Cec_ParChc_t * p );

// extern Gia_Man_t *   Cec_ManSatSweeping( Gia_Man_t * pAig, Cec_ParFra_t * pPars, int fSilent );

// extern Gia_Man_t *   Cec_ManSatSolving( Gia_Man_t * pAig, Cec_ParSat_t * pPars, int f0Proved );

// extern void          Cec_ManSimulation( Gia_Man_t * pAig, Cec_ParSim_t * pPars );

/*=== cecSeq.c ==========================================================*/
// extern int           Cec_ManSeqResimulateCounter( Gia_Man_t * pAig, Cec_ParSim_t * pPars, Abc_Cex_t * pCex );

// extern int           Cec_ManSeqSemiformal( Gia_Man_t * pAig, Cec_ParSmf_t * pPars );

// extern int           Cec_ManCheckNonTrivialCands( Gia_Man_t * pAig );

/*=== cecSynth.c ==========================================================*/
// extern int           Cec_SeqReadMinDomSize( Cec_ParSeq_t * p );

// extern int           Cec_SeqReadVerbose( Cec_ParSeq_t * p );

// extern void          Cec_SeqSynthesisSetDefaultParams( Cec_ParSeq_t * pPars );

// extern int           Cec_SequentialSynthesisPart( Gia_Man_t * p, Cec_ParSeq_t * pPars );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
