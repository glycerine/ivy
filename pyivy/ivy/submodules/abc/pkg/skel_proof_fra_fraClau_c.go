package abc

// Hand-checked skeleton of src/proof/fra/fraClau.c.

// ClaMan is Cla_Man_t.
type ClaMan struct {
	// SAT solvers
	pSatMain *SatSolver
	pSatTest *SatSolver
	pSatBmc  *SatSolver
	// CNF for the test solver
	//    Cnf_Dat_t *      pCnfTest;   (commented out in the C source)
	// SAT variables
	vSatVarsMainCs *IntVector
	vSatVarsTestCs *IntVector
	vSatVarsTestNs *IntVector
	vSatVarsBmcNs  *IntVector
	// helper variables
	nSatVarsTestBeg int
	nSatVarsTestCur int
	// counter-examples
	vCexMain0 *IntVector
	vCexMain  *IntVector
	vCexTest  *IntVector
	vCexBase  *IntVector
	vCexAssm  *IntVector
	vCexBmc   *IntVector
	// mapping of CS into NS var numbers
	pMapCsMainToCsTest []int // int *
	pMapCsTestToCsMain []int // int *
	pMapCsTestToNsTest []int // int *
	pMapCsTestToNsBmc  []int // int *
}

// Fra_ClauSaveLatchVars
// C: Vec_Int_t * Fra_ClauSaveLatchVars( Aig_Man_t * pMan, Cnf_Dat_t * pCnf, int fCsVars )
func (pMan *AigManager) FraClauSaveLatchVars(pCnf *CnfDat, fCsVars int) *IntVector {
	panic("TODO, stubbed")
}

// Fra_ClauSaveOutputVars
// C: Vec_Int_t * Fra_ClauSaveOutputVars( Aig_Man_t * pMan, Cnf_Dat_t * pCnf )
func (pMan *AigManager) FraClauSaveOutputVars(pCnf *CnfDat) *IntVector {
	panic("TODO, stubbed")
}

// Fra_ClauSaveInputVars
// C: Vec_Int_t * Fra_ClauSaveInputVars( Aig_Man_t * pMan, Cnf_Dat_t * pCnf, int nStarting )
func (pMan *AigManager) FraClauSaveInputVars(pCnf *CnfDat, nStarting int) *IntVector {
	panic("TODO, stubbed")
}

// Fra_ClauCreateMapping
// C: int * Fra_ClauCreateMapping( Vec_Int_t * vSatVarsFrom, Vec_Int_t * vSatVarsTo, int nVarsMax )
func FraClauCreateMapping(vSatVarsFrom *IntVector, vSatVarsTo *IntVector, nVarsMax int) []int {
	panic("TODO, stubbed")
}

// Fra_ClauStop
// C: void Fra_ClauStop( Cla_Man_t * p )
func (p *ClaMan) FraClauStop() {
	panic("TODO, stubbed")
}

// Fra_ClauStart
// C: Cla_Man_t * Fra_ClauStart( Aig_Man_t * pMan )
func (pMan *AigManager) FraClauStart() *ClaMan {
	panic("TODO, stubbed")
}

// Vec_IntSplitHalf (static in fraClau.c)
// C: Vec_Int_t * Vec_IntSplitHalf( Vec_Int_t * vVec )
func vecIntSplitHalf(vVec *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Vec_IntComplement (static in fraClau.c)
// C: void Vec_IntComplement( Vec_Int_t * vVec )
func vecIntComplement(vVec *IntVector) {
	panic("TODO, stubbed")
}

// Fra_ClauCheckProperty
// C: int Fra_ClauCheckProperty( Cla_Man_t * p, Vec_Int_t * vCex )
func (p *ClaMan) FraClauCheckProperty(vCex *IntVector) int {
	panic("TODO, stubbed")
}

// Fra_ClauCheckBmc
// C: int Fra_ClauCheckBmc( Cla_Man_t * p, Vec_Int_t * vClause )
func (p *ClaMan) FraClauCheckBmc(vClause *IntVector) int {
	panic("TODO, stubbed")
}

// Fra_ClauRemapClause
// C: void Fra_ClauRemapClause( int * pMap, Vec_Int_t * vClause, Vec_Int_t * vRemapped, int fInv )
func FraClauRemapClause(pMap []int, vClause *IntVector, vRemapped *IntVector, fInv int) {
	panic("TODO, stubbed")
}

// Fra_ClauCheckClause
// C: int Fra_ClauCheckClause( Cla_Man_t * p, Vec_Int_t * vClause, Vec_Int_t * vCex )
func (p *ClaMan) FraClauCheckClause(vClause *IntVector, vCex *IntVector) int {
	panic("TODO, stubbed")
}

// Fra_ClauReduceClause
// C: void Fra_ClauReduceClause( Vec_Int_t * vMain, Vec_Int_t * vNew )
func FraClauReduceClause(vMain *IntVector, vNew *IntVector) {
	panic("TODO, stubbed")
}

// Fra_ClauMinimizeClause_rec
// C: void Fra_ClauMinimizeClause_rec( Cla_Man_t * p, Vec_Int_t * vBasis, Vec_Int_t * vExtra )
func (p *ClaMan) FraClauMinimizeClause_rec(vBasis *IntVector, vExtra *IntVector) {
	panic("TODO, stubbed")
}

// Fra_ClauMinimizeClause
// C: void Fra_ClauMinimizeClause( Cla_Man_t * p, Vec_Int_t * vBasis, Vec_Int_t * vExtra )
func (p *ClaMan) FraClauMinimizeClause(vBasis *IntVector, vExtra *IntVector) {
	panic("TODO, stubbed")
}

// Fra_ClauPrintClause
// C: void Fra_ClauPrintClause( Vec_Int_t * vSatCsVars, Vec_Int_t * vCex )
func FraClauPrintClause(vSatCsVars *IntVector, vCex *IntVector) {
	panic("TODO, stubbed")
}

// Fra_Clau
// C: int Fra_Clau( Aig_Man_t * pMan, int nIters, int fVerbose, int fVeryVerbose )
func (pMan *AigManager) FraClau(nIters int, fVerbose int, fVeryVerbose int) int {
	panic("TODO, stubbed")
}
