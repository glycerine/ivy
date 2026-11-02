package abc

// Hand-written skeleton of the types in src/proof/pdr/pdr.h.

// PdrPar is Pdr_Par_t.
type PdrPar struct {
	//    int iOutput;          // zero-based number of primary output to solve   (commented out in the C source)
	nRecycle         int                            // limit on vars for recycling
	nFrameMax        int                            // limit on frame count
	nConfLimit       int                            // limit on SAT solver conflicts
	nConfGenLimit    int                            // limit on SAT solver conflicts during generalization
	nRestLimit       int                            // limit on the number of proof-obligations
	nTimeOut         int                            // timeout in seconds
	nTimeOutGap      int                            // approximate timeout in seconds since the last change
	nTimeOutOne      int                            // approximate timeout in seconds per one output
	nRandomSeed      int                            // value to seed the SAT solver with
	fTwoRounds       int                            // use two rounds for generalization
	fMonoCnf         int                            // monolythic CNF
	fNewXSim         int                            // updated X-valued simulation
	fFlopPrio        int                            // use structural flop priorities
	fFlopOrder       int                            // order flops for 'analyze_final' during generalization
	fDumpInv         int                            // dump inductive invariant
	fUseSupp         int                            // use support in the invariant
	fShortest        int                            // forces bug traces to be shortest
	fShiftStart      int                            // allows clause pushing to start from an intermediate frame
	fReuseProofOblig int                            // reuses proof-obligationgs in the last timeframe
	fSimpleGeneral   int                            // simplified generalization
	fSkipGeneral     int                            // skips expensive generalization step
	fSkipDown        int                            // skips the application of down
	fCtgs            int                            // handle CTGs in down
	fUseAbs          int                            // use abstraction
	fUseSimpleRef    int                            // simplified CEX refinement
	fVerbose         int                            // verbose output`
	fVeryVerbose     int                            // very verbose output
	fNotVerbose      int                            // not printing line by line progress
	fSilent          int                            // totally silent execution
	fSolveAll        int                            // do not stop when found a SAT output
	fStoreCex        int                            // enable storing counter-examples in MO mode
	fUseBridge       int                            // use bridge interface
	fUsePropOut      int                            // use property output
	nFailOuts        int                            // the number of failed outputs
	nDropOuts        int                            // the number of timed out outputs
	nProveOuts       int                            // the number of proved outputs
	iFrame           int                            // explored up to this frame
	RunId            int                            // PDR id in this run
	pFuncStop        func(int) int                  // int(*)(int) : callback to terminate
	pFuncOnFail      func(int, *Counterexample) int // int(*)(int,Abc_Cex_t*) : called for a failed output in MO mode
	timeLastSolved   int64                          // the time when the last output was solved
	vOutMap          *IntVector                     // in the multi-output mode, contains status for each PO (0 = sat; 1 = unsat; negative = undecided)
	pInvFileName     string                         // char * : invariable file name
}
