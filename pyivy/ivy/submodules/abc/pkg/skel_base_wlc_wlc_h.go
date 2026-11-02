package abc

// Hand-written skeleton of types in src/base/wlc/wlc.h that have no literal Go
// counterpart (declarations only). Wlc_Ntk_t is the validated port WLCNetwork
// and Wlc_Obj_t is WLCObject; the port WLCParameters is only the defaults
// subset of Wlc_Par_t (no callbacks, booleans), so Wlc_Par_t is declared here.

// WlcPar is Wlc_Par_t.
type WlcPar struct {
	nBitsAdd        int           // adder bit-width
	nBitsMul        int           // multiplier bit-widht
	nBitsMux        int           // MUX bit-width
	nBitsFlop       int           // flop bit-width
	nIterMax        int           // the max number of iterations
	nLimit          int           // the max number of signals
	fXorOutput      int           // XOR outputs of word-level miter
	fCheckClauses   int           // Check clauses in the reloaded trace
	fPushClauses    int           // Push clauses in the reloaded trace
	fMFFC           int           // Refine the entire MFFC of a PPI
	fPdra           int           // Use pdr -nct
	fLoadTrace      int           // Load previous traces if any
	fProofRefine    int           // Use proof-based refinement
	fHybrid         int           // Use a hybrid of CBR and PBR
	fCheckCombUnsat int           // Check if ABS becomes comb. unsat
	fAbs2           int           // Use UFAR style createAbs
	fProofUsePPI    int           // Use PPI values in PBR
	fUseBmc3        int           // Run BMC3 in parallel
	fShrinkAbs      int           // Shrink Abs with BMC
	fShrinkScratch  int           // Restart pdr from scratch after shrinking
	fVerbose        int           // verbose output
	fPdrVerbose     int           // verbose output
	RunId           int           // id in this run
	pFuncStop       func(int) int // int (*pFuncStop)(int) : callback to terminate
}

// WlcBstPar is Wlc_BstPar_t.
type WlcBstPar struct {
	iOutput          int
	nOutputRange     int
	nAdderLimit      int
	nMultLimit       int
	fGiaSimple       int
	fAddOutputs      int
	fMulti           int
	fBooth           int
	fNonRest         int
	fCla             int
	fDivBy0          int
	fNoCleanup       int
	fCreateMiter     int
	fCreateWordMiter int
	fDecMuxes        int
	fSaveFfNames     int
	fVerbose         int
	vBoxIds          *IntVector
}

// WlaMan is Wla_Man_t.
type WlaMan struct {
	p         *WLCNetwork // Wlc_Ntk_t *
	pPars     *WlcPar
	vClauses  *NestedVector[*PdrSet] // Vec_Vec_t *
	vBlacks   *IntVector
	vSignals  *IntVector
	pCex      *Counterexample
	pGia      *GiaManager
	vUnmark   *BitVector
	pPdrPars  *PdrPar // void *
	pThread   any     // void *
	iCexFrame int
	fNewAbs   int
	nIters    int
	nTotalCla int
	nDisj     int
	nNDisj    int
	tPdr      int64 // abctime
	tCbr      int64 // abctime
	tPbr      int64 // abctime
}
