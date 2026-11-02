package abc

// Hand-written skeleton of the types in src/opt/mfs/mfs.h.

// MfsPar is Mfs_Par_t.
type MfsPar struct {
	// general parameters
	nWinTfoLevs  int // the maximum fanout levels
	nFanoutsMax  int // the maximum number of fanouts
	nDepthMax    int // the maximum number of logic levels
	nWinMax      int // the maximum size of the window
	nGrowthLevel int // the maximum allowed growth in level
	nBTLimit     int // the maximum number of conflicts in one SAT run
	fRrOnly      int // perform redundance removal
	fResub       int // performs resubstitution
	fArea        int // performs optimization for area
	fMoreEffort  int // performs high-affort minimization
	fSwapEdge    int // performs edge swapping
	fOneHotness  int // adds one-hotness conditions
	fDelay       int // performs optimization for delay
	fPower       int // performs power-aware optimization
	fGiaSat      int // use new SAT solver
	fVerbose     int // enable basic stats
	fVeryVerbose int // enable detailed stats
}
