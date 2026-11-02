package abc

// Hand-checked skeleton of src/proof/ssw/sswRarity.c.

// SswRarMan is Ssw_RarMan_t as defined in sswRarity.c. (sswRarity2.c defines a
// different struct of the same C name; see SswRarMan2.)
type SswRarMan struct {
	// parameters
	pPars     *SswRarPars
	nGroups   int // the number of flop groups
	nWordsReg int // the number of words in the registers
	// internal data
	pAig      *AigManager // AIG with equivalence classes
	ppClasses *SswCla     // equivalence classes
	vInits    *IntVector  // initial state
	// simulation data
	pObjData []uint64 // word * : simulation info for each obj
	pPatData []uint64 // word * : pattern data for each reg
	// candidates to update
	vUpdConst *PointerVector[AigObject] // constant 1 candidates
	vUpdClass *PointerVector[AigObject] // class representatives
	// rarity data
	pRarity   []int     // int * : occur counts for patterns in groups
	pPatCosts []float64 // double * : pattern costs
	// best patterns
	vPatBests *IntVector // best patterns
	iFailPo   int        // failed primary output
	iFailPat  int        // failed pattern
	// counter-examples
	vCexes *PointerVector[any] // entries are NULL or the integer 1 cast to a pointer
}

// Ssw_RarGetBinPat (static inline)
// C: int Ssw_RarGetBinPat( Ssw_RarMan_t * p, int iBin, int iPat )
func (p *SswRarMan) sswRarGetBinPat(iBin int, iPat int) int {
	panic("TODO, stubbed")
}

// Ssw_RarSetBinPat (static inline)
// C: void Ssw_RarSetBinPat( Ssw_RarMan_t * p, int iBin, int iPat, int Value )
func (p *SswRarMan) sswRarSetBinPat(iBin int, iPat int, Value int) {
	panic("TODO, stubbed")
}

// Ssw_RarAddToBinPat (static inline)
// C: void Ssw_RarAddToBinPat( Ssw_RarMan_t * p, int iBin, int iPat )
func (p *SswRarMan) sswRarAddToBinPat(iBin int, iPat int) {
	panic("TODO, stubbed")
}

// Ssw_RarBitWordNum (static inline)
// C: int Ssw_RarBitWordNum( int nBits )
func sswRarBitWordNum(nBits int) int {
	panic("TODO, stubbed")
}

// Ssw_RarObjSim (static inline; returns a pointer into pObjData)
// C: word * Ssw_RarObjSim( Ssw_RarMan_t * p, int Id )
func (p *SswRarMan) sswRarObjSim(Id int) []uint64 {
	panic("TODO, stubbed")
}

// Ssw_RarPatSim (static inline; returns a pointer into pPatData)
// C: word * Ssw_RarPatSim( Ssw_RarMan_t * p, int Id )
func (p *SswRarMan) sswRarPatSim(Id int) []uint64 {
	panic("TODO, stubbed")
}

// Ssw_RarSetDefaultParams
// C: void Ssw_RarSetDefaultParams( Ssw_RarPars_t * p )
func (p *SswRarPars) SswRarSetDefaultParams() {
	panic("TODO, stubbed")
}

// Ssw_RarManPrepareRandom
// C: void Ssw_RarManPrepareRandom( int nRandSeed )
func SswRarManPrepareRandom(nRandSeed int) {
	panic("TODO, stubbed")
}

// Ssw_RarManAssingRandomPis
// C: void Ssw_RarManAssingRandomPis( Ssw_RarMan_t * p )
func (p *SswRarMan) SswRarManAssingRandomPis() {
	panic("TODO, stubbed")
}

// Ssw_RarDeriveCex
// C: Abc_Cex_t * Ssw_RarDeriveCex( Ssw_RarMan_t * p, int iFrame, int iPo, int iPatFinal, int fVerbose )
func (p *SswRarMan) SswRarDeriveCex(iFrame int, iPo int, iPatFinal int, fVerbose int) *Counterexample {
	panic("TODO, stubbed")
}

// transpose32 (array parameter transposed in place)
// C: void transpose32( unsigned A[32] )
func transpose32(A *[32]uint32) {
	panic("TODO, stubbed")
}

// transpose64 (array parameter transposed in place)
// C: void transpose64( word A[64] )
func transpose64(A *[64]uint64) {
	panic("TODO, stubbed")
}

// transpose64Simple
// C: void transpose64Simple( word A[64], word B[64] )
func transpose64Simple(A *[64]uint64, B *[64]uint64) {
	panic("TODO, stubbed")
}

// TransposeTest
// C: void TransposeTest()
func TransposeTest() {
	panic("TODO, stubbed")
}

// Ssw_RarTranspose
// C: void Ssw_RarTranspose( Ssw_RarMan_t * p )
func (p *SswRarMan) SswRarTranspose() {
	panic("TODO, stubbed")
}

// Ssw_RarManInitialize
// C: void Ssw_RarManInitialize( Ssw_RarMan_t * p, Vec_Int_t * vInit )
func (p *SswRarMan) SswRarManInitialize(vInit *IntVector) {
	panic("TODO, stubbed")
}

// Ssw_RarManPoIsConst0
// C: int Ssw_RarManPoIsConst0( void * pMan, Aig_Obj_t * pObj )
func SswRarManPoIsConst0(pMan *SswRarMan, pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Ssw_RarManObjIsConst
// C: int Ssw_RarManObjIsConst( void * pMan, Aig_Obj_t * pObj )
func SswRarManObjIsConst(pMan *SswRarMan, pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Ssw_RarManObjsAreEqual
// C: int Ssw_RarManObjsAreEqual( void * pMan, Aig_Obj_t * pObj0, Aig_Obj_t * pObj1 )
func SswRarManObjsAreEqual(pMan *SswRarMan, pObj0 *AigObject, pObj1 *AigObject) int {
	panic("TODO, stubbed")
}

// Ssw_RarManObjHashWord
// C: unsigned Ssw_RarManObjHashWord( void * pMan, Aig_Obj_t * pObj )
func SswRarManObjHashWord(pMan *SswRarMan, pObj *AigObject) uint32 {
	panic("TODO, stubbed")
}

// Ssw_RarManObjWhichOne
// C: int Ssw_RarManObjWhichOne( Ssw_RarMan_t * p, Aig_Obj_t * pObj )
func (p *SswRarMan) SswRarManObjWhichOne(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Ssw_RarManCheckNonConstOutputs
// C: int Ssw_RarManCheckNonConstOutputs( Ssw_RarMan_t * p, int iFrame, abctime Time )
func (p *SswRarMan) SswRarManCheckNonConstOutputs(iFrame int, Time int64) int {
	panic("TODO, stubbed")
}

// Ssw_RarManSimulate
// C: void Ssw_RarManSimulate( Ssw_RarMan_t * p, Vec_Int_t * vInit, int fUpdate, int fFirst )
func (p *SswRarMan) SswRarManSimulate(vInit *IntVector, fUpdate int, fFirst int) {
	panic("TODO, stubbed")
}

// Ssw_RarManStart (static)
// C: Ssw_RarMan_t * Ssw_RarManStart( Aig_Man_t * pAig, Ssw_RarPars_t * pPars )
func (pAig *AigManager) sswRarManStart(pPars *SswRarPars) *SswRarMan {
	panic("TODO, stubbed")
}

// Ssw_RarManStop (static)
// C: void Ssw_RarManStop( Ssw_RarMan_t * p )
func (p *SswRarMan) sswRarManStop() {
	panic("TODO, stubbed")
}

// Ssw_RarTransferPatterns (static)
// C: void Ssw_RarTransferPatterns( Ssw_RarMan_t * p, Vec_Int_t * vInits )
func (p *SswRarMan) sswRarTransferPatterns(vInits *IntVector) {
	panic("TODO, stubbed")
}

// Ssw_RarFindStartingState (static)
// C: Vec_Int_t * Ssw_RarFindStartingState( Aig_Man_t * pAig, Abc_Cex_t * pCex )
func (pAig *AigManager) sswRarFindStartingState(pCex *Counterexample) *IntVector {
	panic("TODO, stubbed")
}

// Ssw_RarCheckTrivial
// C: int Ssw_RarCheckTrivial( Aig_Man_t * pAig, int fVerbose )
func (pAig *AigManager) SswRarCheckTrivial(fVerbose int) int {
	panic("TODO, stubbed")
}

// Ssw_RarSimulate
// C: int Ssw_RarSimulate( Aig_Man_t * pAig, Ssw_RarPars_t * pPars )
func (pAig *AigManager) SswRarSimulate(pPars *SswRarPars) int {
	panic("TODO, stubbed")
}

// Ssw_RarRandomPermFlop
// C: Vec_Int_t * Ssw_RarRandomPermFlop( int nFlops, int nUnused )
func SswRarRandomPermFlop(nFlops int, nUnused int) *IntVector {
	panic("TODO, stubbed")
}

// Ssw_RarSimulateGia
// C: int Ssw_RarSimulateGia( Gia_Man_t * p, Ssw_RarPars_t * pPars )
func (p *GiaManager) SswRarSimulateGia(pPars *SswRarPars) int {
	panic("TODO, stubbed")
}

// Ssw_RarSignalFilter
// C: int Ssw_RarSignalFilter( Aig_Man_t * pAig, Ssw_RarPars_t * pPars )
func (pAig *AigManager) SswRarSignalFilter(pPars *SswRarPars) int {
	panic("TODO, stubbed")
}

// Ssw_RarSignalFilterGia
// C: int Ssw_RarSignalFilterGia( Gia_Man_t * p, Ssw_RarPars_t * pPars )
func (p *GiaManager) SswRarSignalFilterGia(pPars *SswRarPars) int {
	panic("TODO, stubbed")
}
