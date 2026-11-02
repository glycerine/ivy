package abc

// Hand-written skeleton of the types in src/opt/dar/dar.h.

// DarRwrPar is Dar_RwrPar_t.
type DarRwrPar struct {
	nCutsMax     int // the maximum number of cuts to try
	nSubgMax     int // the maximum number of subgraphs to try
	fFanout      int // support fanout representation
	fUpdateLevel int // update level
	fUseZeros    int // performs zero-cost replacement
	fPower       int // enables power-aware rewriting
	fRecycle     int // enables cut recycling
	fVerbose     int // enables verbose output
	fVeryVerbose int // enables very verbose output
}

// DarRefPar is Dar_RefPar_t.
type DarRefPar struct {
	nMffcMin     int // the min MFFC size for which refactoring is used
	nLeafMax     int // the max number of leaves of a cut
	nCutsMax     int // the max number of cuts to consider
	fExtend      int // extends the cut below MFFC
	fUpdateLevel int // updates the level after each move
	fUseZeros    int // perform zero-cost replacements
	fVerbose     int // verbosity level
	fVeryVerbose int // enables very verbose output
}
