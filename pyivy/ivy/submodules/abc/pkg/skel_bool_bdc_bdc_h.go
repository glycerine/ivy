package abc

// Hand-written skeleton of the Bdc_Par_t type in src/bool/bdc/bdc.h.

// BdcPar is Bdc_Par_t.
type BdcPar struct {
	// general parameters
	nVarsMax     int // the maximum support
	fVerbose     int // enable basic stats
	fVeryVerbose int // enable detailed stats
}
