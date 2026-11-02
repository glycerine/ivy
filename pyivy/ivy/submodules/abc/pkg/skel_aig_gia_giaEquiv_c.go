package abc

// Hand-checked skeleton of src/aig/gia/giaEquiv.c.

// Gia_ManOrigIdsInit
// C: void Gia_ManOrigIdsInit( Gia_Man_t * p )
func (p *GiaManager) GiaManOrigIdsInit() {
	panic("TODO, stubbed")
}

// Gia_ManOrigIdsStart
// C: void Gia_ManOrigIdsStart( Gia_Man_t * p )
func (p *GiaManager) GiaManOrigIdsStart() {
	panic("TODO, stubbed")
}

// Gia_ManOrigIdsRemap
// C: void Gia_ManOrigIdsRemap( Gia_Man_t * p, Gia_Man_t * pNew )
func (p *GiaManager) GiaManOrigIdsRemap(pNew *GiaManager) {
	panic("TODO, stubbed")
}

// Gia_ManOrigIdsRemapPairsInsert
// C: void Gia_ManOrigIdsRemapPairsInsert( Vec_Int_t * vMap, int One, int Two )
func GiaManOrigIdsRemapPairsInsert(vMap *IntVector, One int, Two int) {
	panic("TODO, stubbed")
}

// Gia_ManOrigIdsRemapPairsExtract
// C: int Gia_ManOrigIdsRemapPairsExtract( Vec_Int_t * vMap, int One )
func GiaManOrigIdsRemapPairsExtract(vMap *IntVector, One int) int {
	panic("TODO, stubbed")
}

// Gia_ManOrigIdsRemapPairs
// C: Vec_Int_t * Gia_ManOrigIdsRemapPairs( Vec_Int_t * vEquivPairs, int nObjs )
func GiaManOrigIdsRemapPairs(vEquivPairs *IntVector, nObjs int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManOrigIdsReduce
// C: Gia_Man_t * Gia_ManOrigIdsReduce( Gia_Man_t * p, Vec_Int_t * vPairs )
func (p *GiaManager) GiaManOrigIdsReduce(vPairs *IntVector) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManOrigIdsReduceTest
// C: Gia_Man_t * Gia_ManOrigIdsReduceTest( Gia_Man_t * p, Vec_Int_t * vPairs )
func (p *GiaManager) GiaManOrigIdsReduceTest(vPairs *IntVector) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManComputeGiaEquivs
// C: Gia_Man_t * Gia_ManComputeGiaEquivs( Gia_Man_t * pGia, int nConfs, int fVerbose )
func (pGia *GiaManager) GiaManComputeGiaEquivs(nConfs int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManCheckTopoOrder_rec
// C: int Gia_ManCheckTopoOrder_rec( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) GiaManCheckTopoOrder_rec(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Gia_ManCheckTopoOrder
// C: int Gia_ManCheckTopoOrder( Gia_Man_t * p )
func (p *GiaManager) GiaManCheckTopoOrder() int {
	panic("TODO, stubbed")
}

// Gia_ManDeriveNexts
// C: int * Gia_ManDeriveNexts( Gia_Man_t * p )
func (p *GiaManager) GiaManDeriveNexts() []int {
	panic("TODO, stubbed")
}

// Gia_ManDeriveReprs
// C: void Gia_ManDeriveReprs( Gia_Man_t * p )
func (p *GiaManager) GiaManDeriveReprs() {
	panic("TODO, stubbed")
}

// Gia_ManEquivCountLitsAll
// C: int Gia_ManEquivCountLitsAll( Gia_Man_t * p )
func (p *GiaManager) GiaManEquivCountLitsAll() int {
	panic("TODO, stubbed")
}

// Gia_ManEquivCountClasses
// C: int Gia_ManEquivCountClasses( Gia_Man_t * p )
func (p *GiaManager) GiaManEquivCountClasses() int {
	panic("TODO, stubbed")
}

// Gia_ManEquivCheckLits
// C: int Gia_ManEquivCheckLits( Gia_Man_t * p, int nLits )
func (p *GiaManager) GiaManEquivCheckLits(nLits int) int {
	panic("TODO, stubbed")
}

// Gia_ManPrintStatsClasses
// C: void Gia_ManPrintStatsClasses( Gia_Man_t * p )
func (p *GiaManager) GiaManPrintStatsClasses() {
	panic("TODO, stubbed")
}

// Gia_ManEquivCountLits
// C: int Gia_ManEquivCountLits( Gia_Man_t * p )
func (p *GiaManager) GiaManEquivCountLits() int {
	panic("TODO, stubbed")
}

// Gia_ManEquivCountOne
// C: int Gia_ManEquivCountOne( Gia_Man_t * p, int i )
func (p *GiaManager) GiaManEquivCountOne(i int) int {
	panic("TODO, stubbed")
}

// Gia_ManEquivPrintOne
// C: void Gia_ManEquivPrintOne( Gia_Man_t * p, int i, int Counter )
func (p *GiaManager) GiaManEquivPrintOne(i int, Counter int) {
	panic("TODO, stubbed")
}

// Gia_ManEquivPrintClasses
// C: void Gia_ManEquivPrintClasses( Gia_Man_t * p, int fVerbose, float Mem )
func (p *GiaManager) GiaManEquivPrintClasses(fVerbose int, Mem float32) {
	panic("TODO, stubbed")
}

// Gia_ManChoiceMinLevel_rec
// C: int Gia_ManChoiceMinLevel_rec( Gia_Man_t * p, int iPivot, int fDiveIn, Vec_Int_t * vMap )
func (p *GiaManager) GiaManChoiceMinLevel_rec(iPivot int, fDiveIn int, vMap *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManChoiceMinLevel
// C: Vec_Int_t * Gia_ManChoiceMinLevel( Gia_Man_t * p )
func (p *GiaManager) GiaManChoiceMinLevel() *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManEquivRepr (static inline)
// C: Gia_Obj_t * Gia_ManEquivRepr( Gia_Man_t * p, Gia_Obj_t * pObj, int fUseAll, int fDualOut )
func (p *GiaManager) giaManEquivRepr(pObj *GiaObject, fUseAll int, fDualOut int) *GiaObject {
	panic("TODO, stubbed")
}

// Gia_ManEquivReduce_rec
// C: void Gia_ManEquivReduce_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, int fUseAll, int fDualOut )
func (pNew *GiaManager) GiaManEquivReduce_rec(p *GiaManager, pObj *GiaObject, fUseAll int, fDualOut int) {
	panic("TODO, stubbed")
}

// Gia_ManEquivReduce
// C: Gia_Man_t * Gia_ManEquivReduce( Gia_Man_t * p, int fUseAll, int fDualOut, int fSkipPhase, int fVerbose )
func (p *GiaManager) GiaManEquivReduce(fUseAll int, fDualOut int, fSkipPhase int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManEquivReduce2_rec
// C: void Gia_ManEquivReduce2_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vMap, int fDiveIn )
func (pNew *GiaManager) GiaManEquivReduce2_rec(p *GiaManager, pObj *GiaObject, vMap *IntVector, fDiveIn int) {
	panic("TODO, stubbed")
}

// Gia_ManEquivReduce2
// C: Gia_Man_t * Gia_ManEquivReduce2( Gia_Man_t * p )
func (p *GiaManager) GiaManEquivReduce2() *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManEquivFixOutputPairs
// C: void Gia_ManEquivFixOutputPairs( Gia_Man_t * p )
func (p *GiaManager) GiaManEquivFixOutputPairs() {
	panic("TODO, stubbed")
}

// Gia_ManEquivUpdatePointers
// C: void Gia_ManEquivUpdatePointers( Gia_Man_t * p, Gia_Man_t * pNew )
func (p *GiaManager) GiaManEquivUpdatePointers(pNew *GiaManager) {
	panic("TODO, stubbed")
}

// Gia_ManEquivDeriveReprs
// C: void Gia_ManEquivDeriveReprs( Gia_Man_t * p, Gia_Man_t * pNew, Gia_Man_t * pFinal )
func (p *GiaManager) GiaManEquivDeriveReprs(pNew *GiaManager, pFinal *GiaManager) {
	panic("TODO, stubbed")
}

// Gia_ManEquivRemapDfs
// C: Gia_Man_t * Gia_ManEquivRemapDfs( Gia_Man_t * p )
func (p *GiaManager) GiaManEquivRemapDfs() *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManEquivReduceAndRemap
// C: Gia_Man_t * Gia_ManEquivReduceAndRemap( Gia_Man_t * p, int fSeq, int fMiterPairs )
func (p *GiaManager) GiaManEquivReduceAndRemap(fSeq int, fMiterPairs int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManEquivSetColor_rec
// C: int Gia_ManEquivSetColor_rec( Gia_Man_t * p, Gia_Obj_t * pObj, int fOdds )
func (p *GiaManager) GiaManEquivSetColor_rec(pObj *GiaObject, fOdds int) int {
	panic("TODO, stubbed")
}

// Gia_ManEquivSetColors
// C: int Gia_ManEquivSetColors( Gia_Man_t * p, int fVerbose )
func (p *GiaManager) GiaManEquivSetColors(fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_ManSpecBuild (static inline)
// C: void Gia_ManSpecBuild( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vXorLits, int fDualOut, int fSpeculate, Vec_Int_t * vTrace, Vec_Int_t * vGuide, Vec_Int_t * vMap )
func (pNew *GiaManager) giaManSpecBuild(p *GiaManager, pObj *GiaObject, vXorLits *IntVector, fDualOut int, fSpeculate int, vTrace *IntVector, vGuide *IntVector, vMap *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManSpecReduce_rec
// C: void Gia_ManSpecReduce_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vXorLits, int fDualOut, int fSpeculate, Vec_Int_t * vTrace, Vec_Int_t * vGuide, Vec_Int_t * vMap )
func (pNew *GiaManager) GiaManSpecReduce_rec(p *GiaManager, pObj *GiaObject, vXorLits *IntVector, fDualOut int, fSpeculate int, vTrace *IntVector, vGuide *IntVector, vMap *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManSpecReduceTrace
// C: Gia_Man_t * Gia_ManSpecReduceTrace( Gia_Man_t * p, Vec_Int_t * vTrace, Vec_Int_t * vMap )
func (p *GiaManager) GiaManSpecReduceTrace(vTrace *IntVector, vMap *IntVector) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManSpecReduce
// C: Gia_Man_t * Gia_ManSpecReduce( Gia_Man_t * p, int fDualOut, int fSynthesis, int fSpeculate, int fSkipSome, int fVerbose )
func (p *GiaManager) GiaManSpecReduce(fDualOut int, fSynthesis int, fSpeculate int, fSkipSome int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManSpecBuildInit
// C: void Gia_ManSpecBuildInit( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vXorLits, int f, int fDualOut )
func (pNew *GiaManager) GiaManSpecBuildInit(p *GiaManager, pObj *GiaObject, vXorLits *IntVector, f int, fDualOut int) {
	panic("TODO, stubbed")
}

// Gia_ManSpecReduceInit_rec
// C: void Gia_ManSpecReduceInit_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vXorLits, int f, int fDualOut )
func (pNew *GiaManager) GiaManSpecReduceInit_rec(p *GiaManager, pObj *GiaObject, vXorLits *IntVector, f int, fDualOut int) {
	panic("TODO, stubbed")
}

// Gia_ManSpecReduceInit
// C: Gia_Man_t * Gia_ManSpecReduceInit( Gia_Man_t * p, Abc_Cex_t * pInit, int nFrames, int fDualOut )
func (p *GiaManager) GiaManSpecReduceInit(pInit *Counterexample, nFrames int, fDualOut int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManSpecReduceInitFrames
// C: Gia_Man_t * Gia_ManSpecReduceInitFrames( Gia_Man_t * p, Abc_Cex_t * pInit, int nFramesMax, int * pnFrames, int fDualOut, int nMinOutputs )
func (p *GiaManager) GiaManSpecReduceInitFrames(pInit *Counterexample, nFramesMax int, pnFrames *int, fDualOut int, nMinOutputs int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManEquivTransform
// C: void Gia_ManEquivTransform( Gia_Man_t * p, int fVerbose )
func (p *GiaManager) GiaManEquivTransform(fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManEquivMark
// C: void Gia_ManEquivMark( Gia_Man_t * p, char * pFileName, int fSkipSome, int fVerbose )
func (p *GiaManager) GiaManEquivMark(pFileName string, fSkipSome int, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManEquivFilter
// C: void Gia_ManEquivFilter( Gia_Man_t * p, Vec_Int_t * vPoIds, int fVerbose )
func (p *GiaManager) GiaManEquivFilter(vPoIds *IntVector, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManEquivFilterTest
// C: void Gia_ManEquivFilterTest( Gia_Man_t * p )
func (p *GiaManager) GiaManEquivFilterTest() {
	panic("TODO, stubbed")
}

// Gia_ManEquivImprove
// C: void Gia_ManEquivImprove( Gia_Man_t * p )
func (p *GiaManager) GiaManEquivImprove() {
	panic("TODO, stubbed")
}

// Gia_ObjCheckTfi_rec
// C: int Gia_ObjCheckTfi_rec( Gia_Man_t * p, Gia_Obj_t * pOld, Gia_Obj_t * pNode, Vec_Ptr_t * vVisited )
func (p *GiaManager) GiaObjCheckTfi_rec(pOld *GiaObject, pNode *GiaObject, vVisited *PointerVector[GiaObject]) int {
	panic("TODO, stubbed")
}

// Gia_ObjCheckTfi
// C: int Gia_ObjCheckTfi( Gia_Man_t * p, Gia_Obj_t * pOld, Gia_Obj_t * pNode )
func (p *GiaManager) GiaObjCheckTfi(pOld *GiaObject, pNode *GiaObject) int {
	panic("TODO, stubbed")
}

// Gia_ManAddNextEntry_rec
// C: void Gia_ManAddNextEntry_rec( Gia_Man_t * p, Gia_Obj_t * pOld, Gia_Obj_t * pNode )
func (p *GiaManager) GiaManAddNextEntry_rec(pOld *GiaObject, pNode *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManEquivToChoices_rec
// C: void Gia_ManEquivToChoices_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj )
func (pNew *GiaManager) GiaManEquivToChoices_rec(p *GiaManager, pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManRemoveBadChoices
// C: void Gia_ManRemoveBadChoices( Gia_Man_t * p )
func (p *GiaManager) GiaManRemoveBadChoices() {
	panic("TODO, stubbed")
}

// Gia_ManEquivToChoices
// C: Gia_Man_t * Gia_ManEquivToChoices( Gia_Man_t * p, int nSnapshots )
func (p *GiaManager) GiaManEquivToChoices(nSnapshots int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManCountChoiceNodes
// C: int Gia_ManCountChoiceNodes( Gia_Man_t * p )
func (p *GiaManager) GiaManCountChoiceNodes() int {
	panic("TODO, stubbed")
}

// Gia_ManCountChoices
// C: int Gia_ManCountChoices( Gia_Man_t * p )
func (p *GiaManager) GiaManCountChoices() int {
	panic("TODO, stubbed")
}

// Gia_ManHasNoEquivs
// C: int Gia_ManHasNoEquivs( Gia_Man_t * p )
func (p *GiaManager) GiaManHasNoEquivs() int {
	panic("TODO, stubbed")
}

// Gia_CommandSpecI
// C: int Gia_CommandSpecI( Gia_Man_t * pGia, int nFramesInit, int nBTLimitInit, int fStart, int fCheckMiter, int fVerbose )
func (pGia *GiaManager) GiaCommandSpecI(nFramesInit int, nBTLimitInit int, fStart int, fCheckMiter int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_ManFilterEquivsForSpeculation
// C: int Gia_ManFilterEquivsForSpeculation( Gia_Man_t * pGia, char * pName1, char * pName2, int fLatchA, int fLatchB )
func (pGia *GiaManager) GiaManFilterEquivsForSpeculation(pName1 string, pName2 string, fLatchA int, fLatchB int) int {
	panic("TODO, stubbed")
}

// Gia_ManFilterEquivsUsingParts
// C: int Gia_ManFilterEquivsUsingParts( Gia_Man_t * pGia, char * pName1, char * pName2 )
func (pGia *GiaManager) GiaManFilterEquivsUsingParts(pName1 string, pName2 string) int {
	panic("TODO, stubbed")
}

// Gia_ManFilterEquivsUsingLatches
// C: void Gia_ManFilterEquivsUsingLatches( Gia_Man_t * pGia, int fFlopsOnly, int fFlopsWith, int fUseRiDrivers )
func (pGia *GiaManager) GiaManFilterEquivsUsingLatches(fFlopsOnly int, fFlopsWith int, fUseRiDrivers int) {
	panic("TODO, stubbed")
}

// Gia_ManChangeOrder_rec
// C: int Gia_ManChangeOrder_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj )
func (pNew *GiaManager) GiaManChangeOrder_rec(p *GiaManager, pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Gia_ManChangeOrder
// C: Gia_Man_t * Gia_ManChangeOrder( Gia_Man_t * p )
func (p *GiaManager) GiaManChangeOrder() *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManTransferEquivs
// C: void Gia_ManTransferEquivs( Gia_Man_t * p, Gia_Man_t * pNew )
func (p *GiaManager) GiaManTransferEquivs(pNew *GiaManager) {
	panic("TODO, stubbed")
}

// Gia_ManTransferTest
// C: void Gia_ManTransferTest( Gia_Man_t * p )
func (p *GiaManager) GiaManTransferTest() {
	panic("TODO, stubbed")
}

// Cec4_ManMarkIndependentClasses_rec
// C: int Cec4_ManMarkIndependentClasses_rec( Gia_Man_t * p, int iObj )
func (p *GiaManager) Cec4ManMarkIndependentClasses_rec(iObj int) int {
	panic("TODO, stubbed")
}

// Cec4_ManMarkIndependentClasses
// C: int Cec4_ManMarkIndependentClasses( Gia_Man_t * p, Gia_Man_t * pNew )
func (p *GiaManager) Cec4ManMarkIndependentClasses(pNew *GiaManager) int {
	panic("TODO, stubbed")
}

// Cec4_ManSatSolverAnd_rec
// C: int Cec4_ManSatSolverAnd_rec( Gia_Man_t * pCho, Gia_Man_t * p, Gia_Man_t * pNew, int iObj )
func (pCho *GiaManager) Cec4ManSatSolverAnd_rec(p *GiaManager, pNew *GiaManager, iObj int) int {
	panic("TODO, stubbed")
}

// Cec4_ManSatSolverChoices_rec
// C: int Cec4_ManSatSolverChoices_rec( Gia_Man_t * pCho, Gia_Man_t * p, Gia_Man_t * pNew, int iObj )
func (pCho *GiaManager) Cec4ManSatSolverChoices_rec(p *GiaManager, pNew *GiaManager, iObj int) int {
	panic("TODO, stubbed")
}

// Cec4_ManSatSolverChoices
// C: Gia_Man_t * Cec4_ManSatSolverChoices( Gia_Man_t * p, Gia_Man_t * pNew )
func (p *GiaManager) Cec4ManSatSolverChoices(pNew *GiaManager) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManCombSpecReduce
// C: Gia_Man_t * Gia_ManCombSpecReduce( Gia_Man_t * p )
func (p *GiaManager) GiaManCombSpecReduce() *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManCombSpecReduceTest
// C: void Gia_ManCombSpecReduceTest( Gia_Man_t * p, char * pFileName )
func (p *GiaManager) GiaManCombSpecReduceTest(pFileName string) {
	panic("TODO, stubbed")
}
