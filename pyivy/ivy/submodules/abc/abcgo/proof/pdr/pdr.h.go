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
	NRecycle         int32                              // limit on vars for recycling
	NFrameMax        int32                              // limit on frame count
	NConfLimit       int32                              // limit on SAT solver conflicts
	NConfGenLimit    int32                              // limit on SAT solver conflicts during generalization
	NRestLimit       int32                              // limit on the number of proof-obligations
	NTimeOut         int32                              // timeout in seconds
	NTimeOutGap      int32                              // approximate timeout in seconds since the last change
	NTimeOutOne      int32                              // approximate timeout in seconds per one output
	NRandomSeed      int32                              // value to seed the SAT solver with
	FTwoRounds       int32                              // use two rounds for generalization
	FMonoCnf         int32                              // monolythic CNF
	FNewXSim         int32                              // updated X-valued simulation
	FFlopPrio        int32                              // use structural flop priorities
	FFlopOrder       int32                              // order flops for 'analyze_final' during generalization
	FDumpInv         int32                              // dump inductive invariant
	FUseSupp         int32                              // use support in the invariant
	FShortest        int32                              // forces bug traces to be shortest
	FShiftStart      int32                              // allows clause pushing to start from an intermediate frame
	FReuseProofOblig int32                              // reuses proof-obligationgs in the last timeframe
	FSimpleGeneral   int32                              // simplified generalization
	FSkipGeneral     int32                              // skips expensive generalization step
	FSkipDown        int32                              // skips the application of down
	FCtgs            int32                              // handle CTGs in down
	FUseAbs          int32                              // use abstraction
	FUseSimpleRef    int32                              // simplified CEX refinement
	FVerbose         int32                              // verbose output`
	FVeryVerbose     int32                              // very verbose output
	FNotVerbose      int32                              // not printing line by line progress
	FSilent          int32                              // totally silent execution
	FSolveAll        int32                              // do not stop when found a SAT output
	FStoreCex        int32                              // enable storing counter-examples in MO mode
	FUseBridge       int32                              // use bridge interface
	FUsePropOut      int32                              // use property output
	NFailOuts        int32                              // the number of failed outputs
	NDropOuts        int32                              // the number of timed out outputs
	NProveOuts       int32                              // the number of proved outputs
	IFrame           int32                              // explored up to this frame
	RunId            int32                              // PDR id in this run
	pFuncStop        func(int32) int32                  // callback to terminate
	pFuncOnFail      func(int32, *util.Abc_Cex_t) int32 // called for a failed output in MO mode
	timeLastSolved   abctime                            // the time when the last output was solved
	VOutMap          *Vec_Int_t                         // in the multi-output mode, contains status for each PO (0 = sat; 1 = unsat; negative = undecided)
	PInvFileName     string                             // invariable file name
}

// SetVerbose updates the main PDR verbosity flag from dependent packages.
func (p *Pdr_Par_t) SetVerbose(value int32)  { p.FVerbose = value }
func (p *Pdr_Par_t) SetUseAbs(value int32)   { p.FUseAbs = value }
func (p *Pdr_Par_t) SetCtgs(value int32)     { p.FCtgs = value }
func (p *Pdr_Par_t) SetSkipDown(value int32) { p.FSkipDown = value }

// SetStopFunc installs the caller's cancellation callback and run identifier.
func (p *Pdr_Par_t) SetStopFunc(runID int32, stop func(int32) int32) {
	p.RunId = runID
	p.pFuncStop = stop
}

// ConfigureWlc applies the word-level abstraction client's PDR options.
func (p *Pdr_Par_t) ConfigureWlc(verbose, runID int32, stop func(int32) int32, abstraction int32) {
	p.FVerbose = verbose
	p.FVeryVerbose = 0
	p.RunId, p.pFuncStop = runID, stop
	if abstraction != 0 {
		p.FUseAbs, p.FCtgs, p.FSkipDown, p.NRestLimit = 1, 1, 0, 500
	}
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
