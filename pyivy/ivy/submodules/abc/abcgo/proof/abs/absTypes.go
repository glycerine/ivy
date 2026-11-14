package abs

// Abstraction parameter records, translated from abs.h.

type Abs_Par_t struct {
	NFramesMax         int32  // maximum frames
	NFramesStart       int32  // starting frame
	NFramesPast        int32  // overlap frames
	NConfLimit         int32  // conflict limit
	nLearnedMax        int32  // max number of learned clauses
	NLearnedStart      int32  // max number of learned clauses
	NLearnedDelta      int32  // delta increase of learned clauses
	NLearnedPerce      int32  // percentage of clauses to leave
	NTimeOut           int32  // timeout in seconds
	NRatioMin          int32  // stop when less than this % of object is unabstracted
	NRatioMin2         int32  // stop when less than this % of object is unabstracted during refinement
	NRatioMax          int32  // restart when the number of abstracted object is more than this
	FUseTermVars       int32  // use terminal variables
	FUseRollback       int32  // use rollback to the starting number of frames
	FPropFanout        int32  // propagate fanout implications
	FAddLayer          int32  // refinement strategy by adding layers
	FNewRefine         int32  // uses new refinement heuristics
	FUseSkip           int32  // skip proving intermediate timeframes
	FUseSimple         int32  // use simple CNF construction
	FSkipHash          int32  // skip hashing CNF while unrolling
	FUseFullProof      int32  // use full proof for UNSAT cores
	FDumpVabs          int32  // dumps the abstracted model
	FDumpMabs          int32  // dumps the original AIG with abstraction map
	FCallProver        int32  // calls the prover
	FSimpProver        int32  // calls simplification before prover
	PFileVabs          string // dumps the abstracted model into this file
	FVerbose           int32  // verbose flag
	FVeryVerbose       int32  // print additional information
	IFrame             int32  // the number of frames covered
	iFrameProved       int32  // the number of frames proved
	nFramesNoChange    int32  // the number of last frames without changes
	NFramesNoChangeLim int32  // the number of last frames without changes to dump abstraction
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
