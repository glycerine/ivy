package pdr

import "github.com/glycerine/abcgo/misc/util"

/**CFile****************************************************************

  FileName    [pdr.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 20, 2010.]

  Revision    [$Id: pdr.h,v 1.00 2010/11/20 00:00:00 alanmi Exp $]

***********************************************************************/

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Pdr_Par_t struct {
	//    int iOutput;          // zero-based number of primary output to solve
	nRecycle         int32                              // limit on vars for recycling
	nFrameMax        int32                              // limit on frame count
	nConfLimit       int32                              // limit on SAT solver conflicts
	nConfGenLimit    int32                              // limit on SAT solver conflicts during generalization
	nRestLimit       int32                              // limit on the number of proof-obligations
	nTimeOut         int32                              // timeout in seconds
	nTimeOutGap      int32                              // approximate timeout in seconds since the last change
	nTimeOutOne      int32                              // approximate timeout in seconds per one output
	nRandomSeed      int32                              // value to seed the SAT solver with
	fTwoRounds       int32                              // use two rounds for generalization
	fMonoCnf         int32                              // monolythic CNF
	fNewXSim         int32                              // updated X-valued simulation
	fFlopPrio        int32                              // use structural flop priorities
	fFlopOrder       int32                              // order flops for 'analyze_final' during generalization
	fDumpInv         int32                              // dump inductive invariant
	fUseSupp         int32                              // use support in the invariant
	fShortest        int32                              // forces bug traces to be shortest
	fShiftStart      int32                              // allows clause pushing to start from an intermediate frame
	fReuseProofOblig int32                              // reuses proof-obligationgs in the last timeframe
	fSimpleGeneral   int32                              // simplified generalization
	fSkipGeneral     int32                              // skips expensive generalization step
	fSkipDown        int32                              // skips the application of down
	fCtgs            int32                              // handle CTGs in down
	fUseAbs          int32                              // use abstraction
	fUseSimpleRef    int32                              // simplified CEX refinement
	fVerbose         int32                              // verbose output`
	fVeryVerbose     int32                              // very verbose output
	fNotVerbose      int32                              // not printing line by line progress
	fSilent          int32                              // totally silent execution
	fSolveAll        int32                              // do not stop when found a SAT output
	fStoreCex        int32                              // enable storing counter-examples in MO mode
	fUseBridge       int32                              // use bridge interface
	fUsePropOut      int32                              // use property output
	nFailOuts        int32                              // the number of failed outputs
	nDropOuts        int32                              // the number of timed out outputs
	nProveOuts       int32                              // the number of proved outputs
	iFrame           int32                              // explored up to this frame
	RunId            int32                              // PDR id in this run
	pFuncStop        func(int32) int32                  // callback to terminate
	pFuncOnFail      func(int32, *util.Abc_Cex_t) int32 // called for a failed output in MO mode
	timeLastSolved   abctime                            // the time when the last output was solved
	vOutMap          *Vec_Int_t                         // in the multi-output mode, contains status for each PO (0 = sat; 1 = unsat; negative = undecided)
	pInvFileName     string                             // invariable file name
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== pdrCore.c ==========================================================*/
// extern void               Pdr_ManSetDefaultParams( Pdr_Par_t * pPars );

// extern int                Pdr_ManSolve( Aig_Man_t * p, Pdr_Par_t * pPars );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
