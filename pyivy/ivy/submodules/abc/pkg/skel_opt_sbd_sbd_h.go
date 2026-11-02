package abc

// Hand-written skeleton of the types in src/opt/sbd/sbd.h.

// SbdPar is Sbd_Par_t.
type SbdPar struct {
	nLutSize     int // target LUT size
	nLutNum      int // target LUT count
	nCutSize     int // target cut size
	nCutNum      int // target cut count
	nTfoLevels   int // the number of TFO levels (windowing)
	nTfoFanMax   int // the max number of fanouts (windowing)
	nWinSizeMax  int // maximum window size (windowing)
	nBTLimit     int // maximum number of SAT conflicts
	nWords       int // simulation word count
	fMapping     int // generate mapping
	fMoreCuts    int // use several cuts
	fFindDivs    int // perform divisor search
	fUsePath     int // optimize only critical path
	fArea        int // area-oriented optimization
	fCover       int // use complete cover procedure
	fVerbose     int // verbose flag
	fVeryVerbose int // verbose flag
}
