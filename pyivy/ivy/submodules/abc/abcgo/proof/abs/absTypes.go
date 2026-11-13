package abs

// Abstraction parameter records, translated from abs.h.

type Abs_Par_t struct {
	nFramesMax         int32  // maximum frames
	nFramesStart       int32  // starting frame
	nFramesPast        int32  // overlap frames
	nConfLimit         int32  // conflict limit
	nLearnedMax        int32  // max number of learned clauses
	nLearnedStart      int32  // max number of learned clauses
	nLearnedDelta      int32  // delta increase of learned clauses
	nLearnedPerce      int32  // percentage of clauses to leave
	nTimeOut           int32  // timeout in seconds
	nRatioMin          int32  // stop when less than this % of object is unabstracted
	nRatioMin2         int32  // stop when less than this % of object is unabstracted during refinement
	nRatioMax          int32  // restart when the number of abstracted object is more than this
	fUseTermVars       int32  // use terminal variables
	fUseRollback       int32  // use rollback to the starting number of frames
	fPropFanout        int32  // propagate fanout implications
	fAddLayer          int32  // refinement strategy by adding layers
	fNewRefine         int32  // uses new refinement heuristics
	fUseSkip           int32  // skip proving intermediate timeframes
	fUseSimple         int32  // use simple CNF construction
	fSkipHash          int32  // skip hashing CNF while unrolling
	fUseFullProof      int32  // use full proof for UNSAT cores
	fDumpVabs          int32  // dumps the abstracted model
	fDumpMabs          int32  // dumps the original AIG with abstraction map
	fCallProver        int32  // calls the prover
	fSimpProver        int32  // calls simplification before prover
	pFileVabs          string // dumps the abstracted model into this file
	fVerbose           int32  // verbose flag
	fVeryVerbose       int32  // print additional information
	iFrame             int32  // the number of frames covered
	iFrameProved       int32  // the number of frames proved
	nFramesNoChange    int32  // the number of last frames without changes
	nFramesNoChangeLim int32  // the number of last frames without changes to dump abstraction
}

// old abstraction parameters

type Gia_ParAbs_t struct {
	Algo         int32 // the algorithm to be used
	nFramesMax   int32 // timeframes for PBA
	nConfMax     int32 // conflicts for PBA
	fDynamic     int32 // dynamic unfolding for PBA
	fConstr      int32 // use constraints
	nFramesBmc   int32 // timeframes for BMC
	nConfMaxBmc  int32 // conflicts for BMC
	nStableMax   int32 // the number of stable frames to quit
	nRatio       int32 // ratio of flops to quit
	TimeOut      int32 // approximate timeout in seconds
	TimeOutVT    int32 // approximate timeout in seconds
	nBobPar      int32 // Bob's parameter
	fUseBdds     int32 // use BDDs to refine abstraction
	fUseDprove   int32 // use 'dprove' to refine abstraction
	fUseStart    int32 // use starting frame
	fVerbose     int32 // verbose output
	fVeryVerbose int32 // printing additional information
	Status       int32 // the problem status
	nFramesDone  int32 // the number of frames covered
}
