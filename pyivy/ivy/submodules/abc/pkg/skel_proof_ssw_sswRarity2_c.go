package abc

// Hand-checked skeleton of src/proof/ssw/sswRarity2.c.

// SswRarMan2 is Ssw_RarMan_t as defined in sswRarity2.c. (sswRarity.c defines
// a different struct of the same C name; see SswRarMan.)
type SswRarMan2 struct {
	// parameters
	nWords   int // the number of words to simulate
	nFrames  int // the number of frames to simulate
	nBinSize int // the number of flops in one group
	fVerbose int // the verbosiness flag
	nGroups  int // the number of flop groups
	// internal data
	pAig      *AigManager            // AIG with equivalence classes
	ppClasses *SswCla                // equivalence classes
	pSml      *SswSml                // simulation manager
	vSimInfo  *PointerVector[uint32] // simulation info from pSml manager
	vInits    *IntVector             // initial state
	// rarity data
	pRarity      []int     // int * : occur counts for patterns in groups
	pGroupValues []int     // int * : occur counts in each group
	pPatCosts    []float64 // double * : pattern costs
}

// Ssw_RarGetBinPat (static inline)
// C: int Ssw_RarGetBinPat( Ssw_RarMan_t * p, int iBin, int iPat )
func (p *SswRarMan2) sswRarGetBinPat(iBin int, iPat int) int {
	panic("TODO, stubbed")
}

// Ssw_RarSetBinPat (static inline)
// C: void Ssw_RarSetBinPat( Ssw_RarMan_t * p, int iBin, int iPat, int Value )
func (p *SswRarMan2) sswRarSetBinPat(iBin int, iPat int, Value int) {
	panic("TODO, stubbed")
}

// Ssw_RarAddToBinPat (static inline)
// C: void Ssw_RarAddToBinPat( Ssw_RarMan_t * p, int iBin, int iPat )
func (p *SswRarMan2) sswRarAddToBinPat(iBin int, iPat int) {
	panic("TODO, stubbed")
}

// Ssw_RarManStart (static; named with suffix 2 to avoid clashing with the static of the same name in sswRarity.c)
// C: Ssw_RarMan_t * Ssw_RarManStart( Aig_Man_t * pAig, int nWords, int nFrames, int nBinSize, int fVerbose )
func (pAig *AigManager) sswRarManStart2(nWords int, nFrames int, nBinSize int, fVerbose int) *SswRarMan2 {
	panic("TODO, stubbed")
}

// Ssw_RarManStop (static)
// C: void Ssw_RarManStop( Ssw_RarMan_t * p )
func (p *SswRarMan2) sswRarManStop() {
	panic("TODO, stubbed")
}

// Ssw_RarUpdateCounters (static)
// C: void Ssw_RarUpdateCounters( Ssw_RarMan_t * p )
func (p *SswRarMan2) sswRarUpdateCounters() {
	panic("TODO, stubbed")
}

// Ssw_RarTransferPatterns (static)
// C: void Ssw_RarTransferPatterns( Ssw_RarMan_t * p, Vec_Int_t * vInits )
func (p *SswRarMan2) sswRarTransferPatterns(vInits *IntVector) {
	panic("TODO, stubbed")
}

// Ssw_RarFindStartingState (static; suffix 2 as above)
// C: Vec_Int_t * Ssw_RarFindStartingState( Aig_Man_t * pAig, Abc_Cex_t * pCex )
func (pAig *AigManager) sswRarFindStartingState2(pCex *Counterexample) *IntVector {
	panic("TODO, stubbed")
}

// Ssw_RarSimulate2
// C: int Ssw_RarSimulate2( Aig_Man_t * pAig, int nFrames, int nWords, int nBinSize, int nRounds, int TimeOut, int fVerbose )
func (pAig *AigManager) SswRarSimulate2(nFrames int, nWords int, nBinSize int, nRounds int, TimeOut int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Ssw_RarSignalFilter2
// C: int Ssw_RarSignalFilter2( Aig_Man_t * pAig, int nFrames, int nWords, int nBinSize, int nRounds, int TimeOut, Abc_Cex_t * pCex, int fLatchOnly, int fVerbose )
func (pAig *AigManager) SswRarSignalFilter2(nFrames int, nWords int, nBinSize int, nRounds int, TimeOut int, pCex *Counterexample, fLatchOnly int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Ssw_RarSignalFilterGia2
// C: int Ssw_RarSignalFilterGia2( Gia_Man_t * p, int nFrames, int nWords, int nBinSize, int nRounds, int TimeOut, Abc_Cex_t * pCex, int fLatchOnly, int fVerbose )
func (p *GiaManager) SswRarSignalFilterGia2(nFrames int, nWords int, nBinSize int, nRounds int, TimeOut int, pCex *Counterexample, fLatchOnly int, fVerbose int) int {
	panic("TODO, stubbed")
}
