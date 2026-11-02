package abc

// Hand-written skeleton of the types in src/opt/lpk/lpk.h.

// LpkPar is Lpk_Par_t.
type LpkPar struct {
	// user-controlled parameters
	nLutsMax     int // (N) the maximum number of LUTs in the structure
	nLutsOver    int // (Q) the maximum number of LUTs not in the MFFC
	nVarsShared  int // (S) the maximum number of shared variables (crossbars)
	nGrowthLevel int // (L) the maximum increase in the node level after resynthesis
	fSatur       int // iterate till saturation
	fZeroCost    int // accept zero-cost replacements
	fFirst       int // use root node and first cut only
	fOldAlgo     int // use old algorithm
	fVerbose     int // the verbosiness flag
	fVeryVerbose int // additional verbose info printout
	// internal parameters
	nLutSize int // (K) the LUT size (determined by the input network)
	nVarsMax int // (V) the largest number of variables: V = N * (K-1) + 1
}
