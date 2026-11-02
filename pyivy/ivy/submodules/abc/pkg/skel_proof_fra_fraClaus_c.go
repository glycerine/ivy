package abc

// Hand-checked skeleton of src/proof/fra/fraClaus.c.

// CluMan is Clu_Man_t.
type CluMan struct {
	// parameters
	nFrames      int // the K of the K-step induction
	nPref        int // the number of timeframes to skip
	nClausesMax  int // the max number of 4-clauses to consider
	nLutSize     int // the max cut size
	nLevels      int // the number of levels for cut computation
	nCutsMax     int // the maximum number of cuts to compute at a node
	nBatches     int // the number of clause batches to use
	fStepUp      int // increase cut size for each batch
	fTarget      int // tries to prove the property
	fVerbose     int
	fVeryVerbose int
	// internal parameters
	nSimWords     int // the number of simulation words
	nSimWordsPref int // the number of simulation words in the prefix
	nSimFrames    int // the number of frames to simulate
	nBTLimit      int // the largest number of backtracks (0 = infinite)
	// the network
	pAig *AigManager
	// SAT solvers
	pSatMain *SatSolver
	pSatBmc  *SatSolver
	// CNF for the test solver
	pCnf        *CnfDat
	fFail       int
	fFiltering  int
	fNothingNew int
	// clauses
	vLits          *IntVector
	vClauses       *IntVector
	vCosts         *IntVector
	nClauses       int
	nCuts          int
	nOneHots       int
	nOneHotsProven int
	// clauses proven
	vLitsProven    *IntVector
	vClausesProven *IntVector
	// counter-examples
	vCexes      *PointerVector[uint32] // Vec_Ptr_t of unsigned * simulation-info rows
	nCexes      int
	nCexesAlloc int
}

// Fra_ClausRunBmc
// C: int Fra_ClausRunBmc( Clu_Man_t * p )
func (p *CluMan) FraClausRunBmc() int {
	panic("TODO, stubbed")
}

// Fra_ClausRunSat
// C: int Fra_ClausRunSat( Clu_Man_t * p )
func (p *CluMan) FraClausRunSat() int {
	panic("TODO, stubbed")
}

// Fra_ClausRunSat0
// C: int Fra_ClausRunSat0( Clu_Man_t * p )
func (p *CluMan) FraClausRunSat0() int {
	panic("TODO, stubbed")
}

// transpose32a (non-static in the C source)
// C: void transpose32a( unsigned a[32] )
// The C array parameter decays to a pointer and is transposed in place.
func transpose32a(a *[32]uint32) {
	panic("TODO, stubbed")
}

// Fra_ClausProcessClausesCut
// C: int Fra_ClausProcessClausesCut( Clu_Man_t * p, Fra_Sml_t * pSimMan, Dar_Cut_t * pCut, int * pScores )
func (p *CluMan) FraClausProcessClausesCut(pSimMan *FraSml, pCut *DarCut, pScores []int) int {
	panic("TODO, stubbed")
}

// Fra_ClausProcessClausesCut2
// C: int Fra_ClausProcessClausesCut2( Clu_Man_t * p, Fra_Sml_t * pSimMan, Dar_Cut_t * pCut, int * pScores )
func (p *CluMan) FraClausProcessClausesCut2(pSimMan *FraSml, pCut *DarCut, pScores []int) int {
	panic("TODO, stubbed")
}

// Fra_ClausProcessClausesCut3
// C: void Fra_ClausProcessClausesCut3( Clu_Man_t * p, Fra_Sml_t * pSimMan, Aig_Cut_t * pCut, int * pScores )
func (p *CluMan) FraClausProcessClausesCut3(pSimMan *FraSml, pCut *AigCut, pScores []int) {
	panic("TODO, stubbed")
}

// Fra_ClausSelectClauses
// C: int Fra_ClausSelectClauses( Clu_Man_t * p )
func (p *CluMan) FraClausSelectClauses() int {
	panic("TODO, stubbed")
}

// Fra_ClausRecordClause
// C: void Fra_ClausRecordClause( Clu_Man_t * p, Dar_Cut_t * pCut, int iMint, int Cost )
func (p *CluMan) FraClausRecordClause(pCut *DarCut, iMint int, Cost int) {
	panic("TODO, stubbed")
}

// Fra_ClausRecordClause2
// C: void Fra_ClausRecordClause2( Clu_Man_t * p, Aig_Cut_t * pCut, int iMint, int Cost )
func (p *CluMan) FraClausRecordClause2(pCut *AigCut, iMint int, Cost int) {
	panic("TODO, stubbed")
}

// Fra_ClausSmlNodeIsConst
// C: int Fra_ClausSmlNodeIsConst( Fra_Sml_t * pSeq, Aig_Obj_t * pObj )
func (pSeq *FraSml) FraClausSmlNodeIsConst(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Fra_ClausSmlNodesAreImp
// C: int Fra_ClausSmlNodesAreImp( Fra_Sml_t * pSeq, Aig_Obj_t * pObj1, Aig_Obj_t * pObj2 )
func (pSeq *FraSml) FraClausSmlNodesAreImp(pObj1 *AigObject, pObj2 *AigObject) int {
	panic("TODO, stubbed")
}

// Fra_ClausSmlNodesAreImpC
// C: int Fra_ClausSmlNodesAreImpC( Fra_Sml_t * pSeq, Aig_Obj_t * pObj1, Aig_Obj_t * pObj2 )
func (pSeq *FraSml) FraClausSmlNodesAreImpC(pObj1 *AigObject, pObj2 *AigObject) int {
	panic("TODO, stubbed")
}

// Fra_ClausCollectLatchClauses
// C: int Fra_ClausCollectLatchClauses( Clu_Man_t * p, Fra_Sml_t * pSeq )
func (p *CluMan) FraClausCollectLatchClauses(pSeq *FraSml) int {
	panic("TODO, stubbed")
}

// Fra_ClausProcessClauses
// C: int Fra_ClausProcessClauses( Clu_Man_t * p, int fRefs )
func (p *CluMan) FraClausProcessClauses(fRefs int) int {
	panic("TODO, stubbed")
}

// Fra_ClausProcessClauses2
// C: int Fra_ClausProcessClauses2( Clu_Man_t * p, int fRefs )
func (p *CluMan) FraClausProcessClauses2(fRefs int) int {
	panic("TODO, stubbed")
}

// Fra_ClausBmcClauses
// C: int Fra_ClausBmcClauses( Clu_Man_t * p )
func (p *CluMan) FraClausBmcClauses() int {
	panic("TODO, stubbed")
}

// Fra_ClausSimInfoClean
// C: void Fra_ClausSimInfoClean( Clu_Man_t * p )
func (p *CluMan) FraClausSimInfoClean() {
	panic("TODO, stubbed")
}

// Fra_ClausSimInfoRealloc
// C: void Fra_ClausSimInfoRealloc( Clu_Man_t * p )
func (p *CluMan) FraClausSimInfoRealloc() {
	panic("TODO, stubbed")
}

// Fra_ClausSimInfoRecord
// C: void Fra_ClausSimInfoRecord( Clu_Man_t * p, int * pModel )
func (p *CluMan) FraClausSimInfoRecord(pModel []int) {
	panic("TODO, stubbed")
}

// Fra_ClausSimInfoCheck
// C: int Fra_ClausSimInfoCheck( Clu_Man_t * p, int * pLits, int nLits )
func (p *CluMan) FraClausSimInfoCheck(pLits []int, nLits int) int {
	panic("TODO, stubbed")
}

// Fra_ClausInductiveClauses
// C: int Fra_ClausInductiveClauses( Clu_Man_t * p )
func (p *CluMan) FraClausInductiveClauses() int {
	panic("TODO, stubbed")
}

// Fra_ClausAlloc
// C: Clu_Man_t * Fra_ClausAlloc( Aig_Man_t * pAig, int nFrames, int nPref, int nClausesMax, int nLutSize, int nLevels, int nCutsMax, int nBatches, int fStepUp, int fTarget, int fVerbose, int fVeryVerbose )
func (pAig *AigManager) FraClausAlloc(nFrames int, nPref int, nClausesMax int, nLutSize int, nLevels int, nCutsMax int, nBatches int, fStepUp int, fTarget int, fVerbose int, fVeryVerbose int) *CluMan {
	panic("TODO, stubbed")
}

// Fra_ClausFree
// C: void Fra_ClausFree( Clu_Man_t * p )
func (p *CluMan) FraClausFree() {
	panic("TODO, stubbed")
}

// Fra_ClausAddToStorage
// C: void Fra_ClausAddToStorage( Clu_Man_t * p )
func (p *CluMan) FraClausAddToStorage() {
	panic("TODO, stubbed")
}

// Fra_ClausPrintIndClauses
// C: void Fra_ClausPrintIndClauses( Clu_Man_t * p )
func (p *CluMan) FraClausPrintIndClauses() {
	panic("TODO, stubbed")
}

// Fra_ClausGetLiteral
// C: Aig_Obj_t * Fra_ClausGetLiteral( Clu_Man_t * p, int * pVar2Id, int Lit )
func (p *CluMan) FraClausGetLiteral(pVar2Id []int, Lit int) *AigObject {
	panic("TODO, stubbed")
}

// Fra_ClausWriteIndClauses
// C: void Fra_ClausWriteIndClauses( Clu_Man_t * p )
func (p *CluMan) FraClausWriteIndClauses() {
	panic("TODO, stubbed")
}

// Fra_ClausEstimateCoverageOne
// C: void Fra_ClausEstimateCoverageOne( Fra_Sml_t * pSim, int * pLits, int nLits, int * pVar2Id, unsigned * pResult )
func (pSim *FraSml) FraClausEstimateCoverageOne(pLits []int, nLits int, pVar2Id []int, pResult []uint32) {
	panic("TODO, stubbed")
}

// Fra_ClausEstimateCoverage
// C: void Fra_ClausEstimateCoverage( Clu_Man_t * p )
func (p *CluMan) FraClausEstimateCoverage() {
	panic("TODO, stubbed")
}

// Fra_Claus
// C: int Fra_Claus( Aig_Man_t * pAig, int nFrames, int nPref, int nClausesMax, int nLutSize, int nLevels, int nCutsMax, int nBatches, int fStepUp, int fBmc, int fRefs, int fTarget, int fVerbose, int fVeryVerbose )
func (pAig *AigManager) FraClaus(nFrames int, nPref int, nClausesMax int, nLutSize int, nLevels int, nCutsMax int, nBatches int, fStepUp int, fBmc int, fRefs int, fTarget int, fVerbose int, fVeryVerbose int) int {
	panic("TODO, stubbed")
}
