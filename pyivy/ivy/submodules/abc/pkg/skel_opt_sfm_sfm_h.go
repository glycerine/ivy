package abc

// Hand-written skeleton of the types in src/opt/sfm/sfm.h.

// SfmPar is Sfm_Par_t.
type SfmPar struct {
	nTfoLevMax    int // the maximum fanout levels
	nTfiLevMax    int // the maximum fanin levels
	nFanoutMax    int // the maximum number of fanouts
	nDepthMax     int // the maximum depth to try
	nVarMax       int // the maximum variable count
	nMffcMin      int // the minimum MFFC size
	nMffcMax      int // the maximum MFFC size
	nDecMax       int // the maximum number of decompositions
	nWinSizeMax   int // the maximum window size
	nGrowthLevel  int // the maximum allowed growth in level
	nBTLimit      int // the maximum number of conflicts in one SAT run
	nNodesMax     int // the maximum number of nodes to try
	iNodeOne      int // one particular node to try
	nFirstFixed   int // the number of first nodes to be treated as fixed
	nTimeWin      int // the size of timing window in percents
	DeltaCrit     int // delay delta in picoseconds
	DelAreaRatio  int // delay/area tradeoff (how many ps we trade for a unit of area)
	fRrOnly       int // perform redundance removal
	fArea         int // performs optimization for area
	fAreaRev      int // performs optimization for area in reverse order
	fMoreEffort   int // performs high-affort minimization
	fUseAndOr     int // enable internal detection of AND/OR gates
	fZeroCost     int // enable zero-cost replacement
	fUseSim       int // enable simulation
	fUseDcs       int // enable deriving don't-cares
	fPrintDecs    int // enable printing decompositions
	fAllBoxes     int // enable preserving all boxes
	fLibVerbose   int // enable library stats
	fDelayVerbose int // enable delay stats
	fVerbose      int // enable basic stats
	fVeryVerbose  int // enable detailed stats
}
