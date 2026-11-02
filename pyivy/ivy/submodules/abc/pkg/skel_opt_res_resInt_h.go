package abc

// Hand-written skeleton of the Res_Sim_t type in src/opt/res/resInt.h.

// ResSim is Res_Sim_t.
type ResSim struct {
	pAig     *ABCNetwork // Abc_Ntk_t * : AIG for simulation
	nTruePis int         // the number of true PIs of the window
	fConst0  int         // the node can be replaced by constant 0
	fConst1  int         // the node can be replaced by constant 0
	// simulation parameters
	nWords    int // the number of simulation words
	nPats     int // the number of patterns
	nWordsIn  int // the number of simulation words in the input patterns
	nPatsIn   int // the number of patterns in the input patterns
	nBytesIn  int // the number of bytes in the input patterns
	nWordsOut int // the number of simulation words in the output patterns
	nPatsOut  int // the number of patterns in the output patterns
	// simulation info (each Vec_Ptr_t holds unsigned * simulation rows)
	vPats  *PointerVector[uint32] // input simulation patterns
	vPats0 *PointerVector[uint32] // input simulation patterns
	vPats1 *PointerVector[uint32] // input simulation patterns
	vOuts  *PointerVector[uint32] // output simulation info
	nPats0 int                    // the number of 0-patterns accumulated
	nPats1 int                    // the number of 1-patterns accumulated
	// resub candidates (only allocated and freed in resSim_old.c, so the stored element type is not visible)
	vCands *NestedVector[any]
	// statistics
	timeSat int64 // abctime
}
