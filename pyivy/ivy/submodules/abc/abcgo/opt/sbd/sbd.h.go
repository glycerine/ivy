package sbd

// Port of src/opt/sbd/sbd.h.

type Sbd_Par_t struct {
	NLutSize     int32 // target LUT size
	NLutNum      int32 // target LUT count
	NCutSize     int32 // target cut size
	NCutNum      int32 // target cut count
	NTfoLevels   int32 // the number of TFO levels (windowing)
	NTfoFanMax   int32 // the max number of fanouts (windowing)
	NWinSizeMax  int32 // maximum window size (windowing)
	NBTLimit     int32 // maximum number of SAT conflicts
	NWords       int32 // simulation word count
	FMapping     int32 // generate mapping
	FMoreCuts    int32 // use several cuts
	FFindDivs    int32 // perform divisor search
	FUsePath     int32 // optimize only critical path
	FArea        int32 // area-oriented optimization
	FCover       int32 // use complete cover procedure
	FVerbose     int32 // verbose flag
	FVeryVerbose int32 // verbose flag
}
