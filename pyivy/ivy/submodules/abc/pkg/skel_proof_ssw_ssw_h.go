package abc

// Hand-written skeleton of the types in src/proof/ssw/ssw.h.
// void * members are written any.

// SswPars is Ssw_Pars_t.
type SswPars struct {
	nPartSize      int // size of the partition
	nOverSize      int // size of the overlap between partitions
	nFramesK       int // the induction depth
	nFramesAddSim  int // the number of additional frames to simulate
	fConstrs       int // treat the last nConstrs POs as seq constraints
	fMergeFull     int // enables full merge when constraints are used
	nMaxLevs       int // the max number of levels of nodes to consider
	nBTLimit       int // conflict limit at a node
	nBTLimitGlobal int // conflict limit for multiple runs
	nMinDomSize    int // min clock domain considered for optimization
	nItersStop     int // stop after the given number of iterations
	fDumpSRInit    int // dumps speculative reduction
	nResimDelta    int // the number of nodes to resimulate
	nStepsMax      int // (scorr only) the max number of induction steps
	TimeLimit      int // time out in seconds
	fPolarFlip     int // uses polarity adjustment
	fLatchCorr     int // perform register correspondence
	fConstCorr     int // perform constant correspondence
	fOutputCorr    int // perform 'PO correspondence'
	fSemiFormal    int // enable semiformal filtering
	//    int              fUniqueness;   // enable uniqueness constraints   (commented out in the C source)
	fDynamic      int // enable dynamic addition of constraints
	fLocalSim     int // enable local simulation simulation
	fPartSigCorr  int // uses partial signal correspondence
	nIsleDist     int // extends islands by the given distance
	fScorrGia     int // new signal correspondence implementation
	fUseCSat      int // new SAT solver using when fScorrGia is selected
	fVerbose      int // verbose stats
	fFlopVerbose  int // verbose printout of redundant flops
	fEquivDump    int // enables dumping equivalences
	fEquivDump2   int // enables dumping equivalences
	fStopWhenGone int // stop when PO output is not a candidate constant
	// optimized latch correspondence
	fLatchCorrOpt int // perform register correspondence (optimized)
	nSatVarMax    int // max number of SAT vars before recycling SAT solver (optimized latch corr only)
	nRecycleCalls int // calls to perform before recycling SAT solver (optimized latch corr only)
	// optimized signal correspondence
	nSatVarMax2    int // max number of SAT vars before recycling SAT solver (optimized latch corr only)
	nRecycleCalls2 int // calls to perform before recycling SAT solver (optimized latch corr only)
	// internal parameters
	nIters     int // the number of iterations performed
	nConflicts int // the total number of conflicts performed
	// callback
	pData any // void *
	pFunc any // void *
}

// SswRarPars is Ssw_RarPars_t.
type SswRarPars struct {
	nFrames        int
	nWords         int
	nBinSize       int
	nRounds        int
	nRestart       int
	nRandSeed      int
	TimeOut        int
	TimeOutGap     int
	fSolveAll      int
	fSetLastState  int
	fVerbose       int
	fNotVerbose    int
	fSilent        int
	fDropSatOuts   int
	fMiter         int
	fUseCex        int
	fLatchOnly     int
	fUseFfGrouping int
	nSolved        int
	pCex           *Counterexample                // Abc_Cex_t *
	pFuncOnFail    func(int, *Counterexample) int // int(*)(int,Abc_Cex_t*) : called for a failed output in MO mode
}
