package abc

// Hand-checked skeleton of src/aig/gia/giaQbf.c.

// QbfMan is Qbf_Man_t.
type QbfMan struct {
	pGia       *GiaManager   // original miter
	nPars      int           // parameter variables
	nVars      int           // functional variables
	fVerbose   int           // verbose flag
	iParVarBeg int           // SAT var ID of the first par variable in the ver solver
	pSatVer    *SatSolver    // verification instance
	pSatSyn    *SatSolver    // synthesis instance
	pSatSynG   BmcgSatSolver // bmcg_sat_solver * : synthesis instance
	vValues    *IntVector    // variable values
	vParMap    *IntVector    // parameter mapping
	vLits      *IntVector    // literals for the SAT solver
	clkStart   int64         // abctime : global timeout
	clkSat     int64         // abctime : SAT solver time
}

// Gia_GenCollectFlopIndexes
// C: Vec_Int_t * Gia_GenCollectFlopIndexes( char * pStr, int nLutNum, int nLutSize, int nFlops )
func GiaGenCollectFlopIndexes(pStr string, nLutNum int, nLutSize int, nFlops int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_GenCreateMux_rec
// C: int Gia_GenCreateMux_rec( Gia_Man_t * pNew, int * pCtrl, int nCtrl, Vec_Int_t * vData, int Shift )
func (pNew *GiaManager) GiaGenCreateMux_rec(pCtrl []int, nCtrl int, vData *IntVector, Shift int) int {
	panic("TODO, stubbed")
}

// Gia_GenCreateMuxes
// C: Vec_Int_t * Gia_GenCreateMuxes( Gia_Man_t * p, Gia_Man_t * pNew, Vec_Int_t * vFlops, int nLutNum, int nLutSize, Vec_Int_t * vParLits, int fUseRi )
func (p *GiaManager) GiaGenCreateMuxes(pNew *GiaManager, vFlops *IntVector, nLutNum int, nLutSize int, vParLits *IntVector, fUseRi int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_GenQbfMiter
// C: Gia_Man_t * Gia_GenQbfMiter( Gia_Man_t * p, int nFrames, int nLutNum, int nLutSize, char * pStr, int fUseOut, int fVerbose )
func (p *GiaManager) GiaGenQbfMiter(nFrames int, nLutNum int, nLutSize int, pStr string, fUseOut int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_Gen2CreateMux_rec
// C: int Gia_Gen2CreateMux_rec( Gia_Man_t * pNew, int * pCtrl, int nCtrl, Vec_Int_t * vData, int Shift )
func (pNew *GiaManager) GiaGen2CreateMux_rec(pCtrl []int, nCtrl int, vData *IntVector, Shift int) int {
	panic("TODO, stubbed")
}

// Gia_Gen2CreateMuxes
// C: Vec_Int_t * Gia_Gen2CreateMuxes( Gia_Man_t * pNew, int nLutSize, int nLutNum, Vec_Int_t * vPLits, Vec_Int_t * vXLits )
func (pNew *GiaManager) GiaGen2CreateMuxes(nLutSize int, nLutNum int, vPLits *IntVector, vXLits *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_Gen2CreateMiter
// C: Gia_Man_t * Gia_Gen2CreateMiter( int nLutSize, int nLutNum )
func GiaGen2CreateMiter(nLutSize int, nLutNum int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_Gen2CodeOne
// C: int Gia_Gen2CodeOne( int nLutSize, int nLutNum, Vec_Int_t * vCode, int x )
func GiaGen2CodeOne(nLutSize int, nLutNum int, vCode *IntVector, x int) int {
	panic("TODO, stubbed")
}

// Gia_Gen2CodeOneP
// C: word * Gia_Gen2CodeOneP( int nLutSize, int nLutNum, Vec_Int_t * vCode, int x )
func GiaGen2CodeOneP(nLutSize int, nLutNum int, vCode *IntVector, x int) []uint64 {
	panic("TODO, stubbed")
}

// Gia_Gen2CodePrint
// C: void Gia_Gen2CodePrint( int nLutSize, int nLutNum, Vec_Int_t * vCode )
func GiaGen2CodePrint(nLutSize int, nLutNum int, vCode *IntVector) {
	panic("TODO, stubbed")
}

// Gia_Gen2CodeTest
// C: void Gia_Gen2CodeTest()
func GiaGen2CodeTest() {
	panic("TODO, stubbed")
}

// Gia_ManSatEnum
// C: int Gia_ManSatEnum( Gia_Man_t * pGia, int nConfLimit, int nTimeOut, int fVerbose )
func (pGia *GiaManager) GiaManSatEnum(nConfLimit int, nTimeOut int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_QbfDumpFile
// C: void Gia_QbfDumpFile( Gia_Man_t * pGia, int nPars )
func (pGia *GiaManager) GiaQbfDumpFile(nPars int) {
	panic("TODO, stubbed")
}

// Gia_QbfAlloc
// C: Qbf_Man_t * Gia_QbfAlloc( Gia_Man_t * pGia, int nPars, int fGlucose, int fVerbose )
func (pGia *GiaManager) GiaQbfAlloc(nPars int, fGlucose int, fVerbose int) *QbfMan {
	panic("TODO, stubbed")
}

// Gia_QbfFree
// C: void Gia_QbfFree( Qbf_Man_t * p )
func (p *QbfMan) GiaQbfFree() {
	panic("TODO, stubbed")
}

// Gia_QbfQuantifyOne
// C: Gia_Man_t * Gia_QbfQuantifyOne( Gia_Man_t * p, int iVar, int fAndAll, int fOrAll )
func (p *GiaManager) GiaQbfQuantifyOne(iVar int, fAndAll int, fOrAll int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_QbfQuantifyAll
// C: Gia_Man_t * Gia_QbfQuantifyAll( Gia_Man_t * p, int nPars, int fAndAll, int fOrAll )
func (p *GiaManager) GiaQbfQuantifyAll(nPars int, fAndAll int, fOrAll int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_QbfCofactor
// C: Gia_Man_t * Gia_QbfCofactor( Gia_Man_t * p, int nPars, Vec_Int_t * vValues, Vec_Int_t * vParMap )
func (p *GiaManager) GiaQbfCofactor(nPars int, vValues *IntVector, vParMap *IntVector) *GiaManager {
	panic("TODO, stubbed")
}

// Cnf_SpecialDataLift
// C: void Cnf_SpecialDataLift( Cnf_Dat_t * p, int nVarsPlus, int firstPiVar, int lastPiVar)
func CnfSpecialDataLift(p *CnfDat, nVarsPlus int, firstPiVar int, lastPiVar int) {
	panic("TODO, stubbed")
}

// Gia_QbfAddCofactor
// C: int Gia_QbfAddCofactor( Qbf_Man_t * p, Gia_Man_t * pCof )
func (p *QbfMan) GiaQbfAddCofactor(pCof *GiaManager) int {
	panic("TODO, stubbed")
}

// Gia_QbfAddCofactorG
// C: int Gia_QbfAddCofactorG( Qbf_Man_t * p, Gia_Man_t * pCof )
func (p *QbfMan) GiaQbfAddCofactorG(pCof *GiaManager) int {
	panic("TODO, stubbed")
}

// Gia_QbfOnePattern
// C: void Gia_QbfOnePattern( Qbf_Man_t * p, Vec_Int_t * vValues )
func (p *QbfMan) GiaQbfOnePattern(vValues *IntVector) {
	panic("TODO, stubbed")
}

// Gia_QbfPrint
// C: void Gia_QbfPrint( Qbf_Man_t * p, Vec_Int_t * vValues, int Iter )
func (p *QbfMan) GiaQbfPrint(vValues *IntVector, Iter int) {
	panic("TODO, stubbed")
}

// Gia_QbfVerify
// C: int Gia_QbfVerify( Qbf_Man_t * p, Vec_Int_t * vValues )
func (p *QbfMan) GiaQbfVerify(vValues *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_QbfAddSpecialConstr
// C: void Gia_QbfAddSpecialConstr( Qbf_Man_t * p )
func (p *QbfMan) GiaQbfAddSpecialConstr() {
	panic("TODO, stubbed")
}

// Gia_QbfLearnConstraint
// C: void Gia_QbfLearnConstraint( Qbf_Man_t * p, Vec_Int_t * vValues )
func (p *QbfMan) GiaQbfLearnConstraint(vValues *IntVector) {
	panic("TODO, stubbed")
}

// Gia_QbfSolve
// C: int Gia_QbfSolve( Gia_Man_t * pGia, int nPars, int nIterLimit, int nConfLimit, int nTimeOut, int nEncVars, int fGlucose, int fVerbose )
func (pGia *GiaManager) GiaQbfSolve(nPars int, nIterLimit int, nConfLimit int, nTimeOut int, nEncVars int, fGlucose int, fVerbose int) int {
	panic("TODO, stubbed")
}
