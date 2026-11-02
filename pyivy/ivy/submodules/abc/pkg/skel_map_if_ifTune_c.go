package abc

// Hand-checked skeleton of src/map/if/ifTune.c.

const (
	IFN_PAR = 1024
	IFN_INS = 11
	IFN_WRD = 1 << (IFN_INS - 6) // (IFN_INS > 6 ? 1 << (IFN_INS-6) : 1)
)

// Ifn_DsdType_t: network types.
const (
	IFN_DSD_NONE   = 0 // 0:  unknown
	IFN_DSD_CONST0     // 1:  constant
	IFN_DSD_VAR        // 2:  variable
	IFN_DSD_AND        // 3:  AND
	IFN_DSD_XOR        // 4:  XOR
	IFN_DSD_MUX        // 5:  MUX
	IFN_DSD_PRIME      // 6:  PRIME
)

// The static table Ifn_Symbs[16] = { NULL, "const", "var", "()", "[]", "<>", "{}" } is not represented.

// IfnObj is Ifn_Obj_t.
type IfnObj struct {
	Type    uint32       // unsigned : 3 : node type
	nFanins uint32       // unsigned : 5 : fanin counter
	iFirst  uint32       // unsigned : 8 : first parameter
	Var     uint32       // unsigned : 16 : current variable
	Fanins  [IFN_INS]int // fanin IDs
}

// IfnNtk is Ifn_Ntk_t.
type IfnNtk struct {
	// cell structure
	nInps int                 // inputs
	nObjs int                 // objects
	Nodes [2 * IFN_INS]IfnObj // nodes
	// constraints
	pConstr [IFN_INS * IFN_INS]int // constraint pairs
	nConstr int                    // number of pairs
	// user data
	nVars     int      // variables
	nWords    int      // truth table words
	nParsVNum int      // selection parameters per variable
	nParsVIni int      // first selection parameter
	nPars     int      // total parameters
	pTruth    []uint64 // word * : user truth table
	// matching procedures
	Values   [IFN_PAR]int                  // variable values
	pTtElems [IFN_INS * IFN_WRD]uint64     // elementary truth tables
	pTtObjs  [2 * IFN_INS * IFN_WRD]uint64 // object truth tables
}

// Ifn_ElemTruth (static inline)
// C: word * Ifn_ElemTruth( Ifn_Ntk_t * p, int i )
func (p *IfnNtk) ifnElemTruth(i int) []uint64 {
	panic("TODO, stubbed")
}

// Ifn_ObjTruth (static inline)
// C: word * Ifn_ObjTruth( Ifn_Ntk_t * p, int i )
func (p *IfnNtk) ifnObjTruth(i int) []uint64 {
	panic("TODO, stubbed")
}

// Ifn_Prepare
// C: int Ifn_Prepare( Ifn_Ntk_t * p, word * pTruth, int nVars )
func (p *IfnNtk) IfnPrepare(pTruth []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// Ifn_NtkPrint
// C: void Ifn_NtkPrint( Ifn_Ntk_t * p )
func (p *IfnNtk) IfnNtkPrint() {
	panic("TODO, stubbed")
}

// Ifn_NtkLutSizeMax
// C: int Ifn_NtkLutSizeMax( Ifn_Ntk_t * p )
func (p *IfnNtk) IfnNtkLutSizeMax() int {
	panic("TODO, stubbed")
}

// Ifn_NtkInputNum
// C: int Ifn_NtkInputNum( Ifn_Ntk_t * p )
func (p *IfnNtk) IfnNtkInputNum() int {
	panic("TODO, stubbed")
}

// Ifn_ErrorMessage
// C: int Ifn_ErrorMessage( const char * format, ... )
func IfnErrorMessage(format string, args ...any) int {
	panic("TODO, stubbed")
}

// Inf_ManOpenSymb
// C: int Inf_ManOpenSymb( char * pStr )
func InfManOpenSymb(pStr string) int {
	panic("TODO, stubbed")
}

// Ifn_ManStrCheck
// C: int Ifn_ManStrCheck( char * pStr, int * pnInps, int * pnObjs )
func IfnManStrCheck(pStr string, pnInps *int, pnObjs *int) int {
	panic("TODO, stubbed")
}

// Ifn_NtkParseFindClosingParenthesis (static inline; returns the offset of the closing symbol in pStr, or -1 for NULL)
// C: char * Ifn_NtkParseFindClosingParenthesis( char * pStr, char Open, char Close )
func ifnNtkParseFindClosingParenthesis(pStr string, Open byte, Close byte) int {
	panic("TODO, stubbed")
}

// Ifn_NtkParseInt_rec (ppFinal is a char ** cursor into pStr; here an offset into pStr (-1 for NULL))
// C: int Ifn_NtkParseInt_rec( char * pStr, Ifn_Ntk_t * p, char ** ppFinal, int * piNode )
func IfnNtkParseInt_rec(pStr string, p *IfnNtk, ppFinal *int, piNode *int) int {
	panic("TODO, stubbed")
}

// Ifn_NtkParseInt
// C: int Ifn_NtkParseInt( char * pStr, Ifn_Ntk_t * p )
func IfnNtkParseInt(pStr string, p *IfnNtk) int {
	panic("TODO, stubbed")
}

// Ifn_ManStrType2
// C: int Ifn_ManStrType2( char * pStr )
func IfnManStrType2(pStr string) int {
	panic("TODO, stubbed")
}

// Ifn_ManStrCheck2
// C: int Ifn_ManStrCheck2( char * pStr, int * pnInps, int * pnObjs )
func IfnManStrCheck2(pStr string, pnInps *int, pnObjs *int) int {
	panic("TODO, stubbed")
}

// Ifn_NtkParseInt2
// C: int Ifn_NtkParseInt2( char * pStr, Ifn_Ntk_t * p )
func IfnNtkParseInt2(pStr string, p *IfnNtk) int {
	panic("TODO, stubbed")
}

// Ifn_NtkParseConstraints
// C: void Ifn_NtkParseConstraints( char * pStr, Ifn_Ntk_t * p )
func IfnNtkParseConstraints(pStr string, p *IfnNtk) {
	panic("TODO, stubbed")
}

// Ifn_NtkParse
// C: Ifn_Ntk_t * Ifn_NtkParse( char * pStr )
func IfnNtkParse(pStr string) *IfnNtk {
	panic("TODO, stubbed")
}

// Ifn_NtkTtBits
// C: int Ifn_NtkTtBits( char * pStr )
func IfnNtkTtBits(pStr string) int {
	panic("TODO, stubbed")
}

// Ifn_ManStrFindModel
// C: Gia_Man_t * Ifn_ManStrFindModel( Ifn_Ntk_t * p )
func (p *IfnNtk) IfnManStrFindModel() *GiaManager {
	panic("TODO, stubbed")
}

// Ifn_ManStrFindCofactors
// C: Gia_Man_t * Ifn_ManStrFindCofactors( int nIns, Gia_Man_t * p )
func IfnManStrFindCofactors(nIns int, p *GiaManager) *GiaManager {
	panic("TODO, stubbed")
}

// Cnf_DeriveGiaRemapped
// C: Cnf_Dat_t * Cnf_DeriveGiaRemapped( Gia_Man_t * p )
func (p *GiaManager) cnfDeriveGiaRemapped() *CnfDat {
	panic("TODO, stubbed")
}

// Ifn_ManStrFindSolver
// C: sat_solver * Ifn_ManStrFindSolver( Gia_Man_t * p, Vec_Int_t ** pvPiVars, Vec_Int_t ** pvPoVars )
func (p *GiaManager) IfnManStrFindSolver(pvPiVars **IntVector, pvPoVars **IntVector) *SatSolver {
	panic("TODO, stubbed")
}

// Ifn_ManSatBuild
// C: sat_solver * Ifn_ManSatBuild( Ifn_Ntk_t * p, Vec_Int_t ** pvPiVars, Vec_Int_t ** pvPoVars )
func (p *IfnNtk) IfnManSatBuild(pvPiVars **IntVector, pvPoVars **IntVector) *SatSolver {
	panic("TODO, stubbed")
}

// If_ManSatBuildFromCell (returns a void * that holds a sat_solver *)
// C: void * If_ManSatBuildFromCell( char * pStr, Vec_Int_t ** pvPiVars, Vec_Int_t ** pvPoVars, Ifn_Ntk_t ** ppNtk )
func IfManSatBuildFromCell(pStr string, pvPiVars **IntVector, pvPoVars **IntVector, ppNtk **IfnNtk) *SatSolver {
	panic("TODO, stubbed")
}

// Ifn_ManSatPrintPerm
// C: void Ifn_ManSatPrintPerm( char * pPerms, int nVars )
func IfnManSatPrintPerm(pPerms []int8, nVars int) {
	panic("TODO, stubbed")
}

// Ifn_ManSatCheckOne
// C: int Ifn_ManSatCheckOne( sat_solver * pSat, Vec_Int_t * vPoVars, word * pTruth, int nVars, int * pPerm, int nInps, Vec_Int_t * vLits )
func (pSat *SatSolver) IfnManSatCheckOne(vPoVars *IntVector, pTruth []uint64, nVars int, pPerm []int, nInps int, vLits *IntVector) int {
	panic("TODO, stubbed")
}

// Ifn_ManSatDeriveOne
// C: void Ifn_ManSatDeriveOne( sat_solver * pSat, Vec_Int_t * vPiVars, Vec_Int_t * vValues )
func (pSat *SatSolver) IfnManSatDeriveOne(vPiVars *IntVector, vValues *IntVector) {
	panic("TODO, stubbed")
}

// If_ManSatFindCofigBits (pSat is a void * that holds a sat_solver *)
// C: int If_ManSatFindCofigBits( void * pSat, Vec_Int_t * vPiVars, Vec_Int_t * vPoVars, word * pTruth, int nVars, word Perm, int nInps, Vec_Int_t * vValues )
func IfManSatFindCofigBits(pSat *SatSolver, vPiVars *IntVector, vPoVars *IntVector, pTruth []uint64, nVars int, Perm uint64, nInps int, vValues *IntVector) int {
	panic("TODO, stubbed")
}

// Ifn_ManSatFindCofigBitsTest
// C: int Ifn_ManSatFindCofigBitsTest( Ifn_Ntk_t * p, word * pTruth, int nVars, word Perm )
func (p *IfnNtk) IfnManSatFindCofigBitsTest(pTruth []uint64, nVars int, Perm uint64) int {
	panic("TODO, stubbed")
}

// If_ManSatDeriveGiaFromBits (pGia is a void * that holds a Gia_Man_t *)
// C: int If_ManSatDeriveGiaFromBits( void * pGia, Ifn_Ntk_t * p, word * pConfigData, Vec_Int_t * vLeaves, Vec_Int_t * vCover )
func IfManSatDeriveGiaFromBits(pGia *GiaManager, p *IfnNtk, pConfigData []uint64, vLeaves *IntVector, vCover *IntVector) int {
	panic("TODO, stubbed")
}

// If_ManDeriveGiaFromCells (pGia and the result are void * that hold Gia_Man_t *)
// C: void * If_ManDeriveGiaFromCells( void * pGia )
func IfManDeriveGiaFromCells(pGia *GiaManager) *GiaManager {
	panic("TODO, stubbed")
}

// Ifn_NtkDeriveTruth
// C: word * Ifn_NtkDeriveTruth( Ifn_Ntk_t * p, int * pValues )
func (p *IfnNtk) IfnNtkDeriveTruth(pValues []int) []uint64 {
	panic("TODO, stubbed")
}

// Ifn_TtComparisonConstr
// C: void Ifn_TtComparisonConstr( word * pTruth, int nVars, int fMore, int fEqual )
func IfnTtComparisonConstr(pTruth []uint64, nVars int, fMore int, fEqual int) {
	panic("TODO, stubbed")
}

// Ifn_AddClause (pBeg and pEnd delimit a literal array)
// C: int Ifn_AddClause( sat_solver * pSat, int * pBeg, int * pEnd )
func (pSat *SatSolver) IfnAddClause(pBeg []int, pEnd []int) int {
	panic("TODO, stubbed")
}

// Ifn_NtkAddConstrOne
// C: void Ifn_NtkAddConstrOne( sat_solver * pSat, Vec_Int_t * vCover, int * pVars, int nVars )
func (pSat *SatSolver) IfnNtkAddConstrOne(vCover *IntVector, pVars []int, nVars int) {
	panic("TODO, stubbed")
}

// Ifn_NtkAddConstraints
// C: void Ifn_NtkAddConstraints( Ifn_Ntk_t * p, sat_solver * pSat )
func (p *IfnNtk) IfnNtkAddConstraints(pSat *SatSolver) {
	panic("TODO, stubbed")
}

// Ifn_NtkAddClauses
// C: int Ifn_NtkAddClauses( Ifn_Ntk_t * p, int * pValues, sat_solver * pSat )
func (p *IfnNtk) IfnNtkAddClauses(pValues []int, pSat *SatSolver) int {
	panic("TODO, stubbed")
}

// Ifn_NtkMatchPrintStatus
// C: void Ifn_NtkMatchPrintStatus( sat_solver * p, int Iter, int status, int iMint, int Value, abctime clk )
func (p *SatSolver) IfnNtkMatchPrintStatus(Iter int, status int, iMint int, Value int, clk int64) {
	panic("TODO, stubbed")
}

// Ifn_NtkMatchPrintConfig
// C: void Ifn_NtkMatchPrintConfig( Ifn_Ntk_t * p, sat_solver * pSat )
func (p *IfnNtk) IfnNtkMatchPrintConfig(pSat *SatSolver) {
	panic("TODO, stubbed")
}

// Ifn_NtkMatchCollectPerm
// C: word Ifn_NtkMatchCollectPerm( Ifn_Ntk_t * p, sat_solver * pSat )
func (p *IfnNtk) IfnNtkMatchCollectPerm(pSat *SatSolver) uint64 {
	panic("TODO, stubbed")
}

// Ifn_NtkMatchCollectConfig
// C: void Ifn_NtkMatchCollectConfig( Ifn_Ntk_t * p, sat_solver * pSat, word * pConfig )
func (p *IfnNtk) IfnNtkMatchCollectConfig(pSat *SatSolver, pConfig []uint64) {
	panic("TODO, stubbed")
}

// Ifn_NtkMatchPrintPerm
// C: void Ifn_NtkMatchPrintPerm( word Perm, int nInps )
func IfnNtkMatchPrintPerm(Perm uint64, nInps int) {
	panic("TODO, stubbed")
}

// Ifn_NtkMatch
// C: int Ifn_NtkMatch( Ifn_Ntk_t * p, word * pTruth, int nVars, int nConfls, int fVerbose, int fVeryVerbose, word * pConfig )
func (p *IfnNtk) IfnNtkMatch(pTruth []uint64, nVars int, nConfls int, fVerbose int, fVeryVerbose int, pConfig []uint64) int {
	panic("TODO, stubbed")
}

// Ifn_NtkRead
// C: void Ifn_NtkRead()
func IfnNtkRead() {
	panic("TODO, stubbed")
}
