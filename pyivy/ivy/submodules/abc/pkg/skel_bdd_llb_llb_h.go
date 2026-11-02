package abc

// Hand-checked skeleton of src/bdd/llb/llb.h.

// GiaParLlb is Gia_ParLlb_t: parameters of the BDD-based reachability.
type GiaParLlb struct {
	nBddMax       int // maximum BDD size
	nIterMax      int // maximum iteration count
	nClusterMax   int // maximum cluster size
	nHintDepth    int // the number of times to cofactor
	HintFirst     int // the number of first hint to use
	fUseFlow      int // use flow computation
	nVolumeMax    int // the largest volume
	nVolumeMin    int // the smallest volume
	nPartValue    int // partitioning value
	fBackward     int // enable backward reachability
	fReorder      int // enable dynamic variable reordering
	fIndConstr    int // extract inductive constraints
	fUsePivots    int // use internal pivot variables
	fCluster      int // use partition clustering
	fSchedule     int // use cluster scheduling
	fDumpReached  int // dump reached states into a file
	fVerbose      int // print verbose information
	fVeryVerbose  int // print dependency matrices
	fSilent       int // do not print any infomation
	fSkipReach    int // skip reachability (preparation phase only)
	fSkipOutCheck int // does not check the property output
	TimeLimit     int // time limit for one reachability run
	TimeLimitGlo  int // time limit for all reachability runs
	// internal parameters
	TimeTarget int64 // abctime : the time to stop
	iFrame     int   // explored up to this frame
}
