package abc

// Hand-checked skeleton of src/aig/gia/giaSweeper.c.

// SwpMan is Swp_Man_t.
type SwpMan struct {
	pGia           *GiaManager // GIA manager under construction
	nConfMax       int         // conflict limit in seconds
	nTimeOut       int         // runtime limit in seconds
	vProbes        *IntVector  // probes
	vCondProbes    *IntVector  // conditions as probes
	vCondAssump    *IntVector  // conditions as SAT solver literals
	pSat           *SatSolver  // SAT solver
	vId2Lit        *IntVector  // mapping of Obj IDs into SAT literal
	vFront         *IntVector  // temporary frontier
	vFanins        *IntVector  // temporary fanins
	vCexSwp        *IntVector  // sweeper counter-example
	vCexUser       *IntVector  // user-visible counter-example
	nSatVars       int         // counter of SAT variables
	nSatCalls      int
	nSatCallsSat   int
	nSatCallsUnsat int
	nSatCallsUndec int
	nSatProofs     int
	timeStart      int64 // abctime
	timeTotal      int64 // abctime
	timeCnf        int64 // abctime
	timeSat        int64 // abctime
	timeSatSat     int64 // abctime
	timeSatUnsat   int64 // abctime
	timeSatUndec   int64 // abctime
}

// Swp_ManObj2Lit (static inline)
// C: int Swp_ManObj2Lit( Swp_Man_t * p, int Id )
func (p *SwpMan) swpManObj2Lit(Id int) int {
	panic("TODO, stubbed")
}

// Swp_ManLit2Lit (static inline)
// C: int Swp_ManLit2Lit( Swp_Man_t * p, int Lit )
func (p *SwpMan) swpManLit2Lit(Lit int) int {
	panic("TODO, stubbed")
}

// Swp_ManSetObj2Lit (static inline)
// C: void Swp_ManSetObj2Lit( Swp_Man_t * p, int Id, int Lit )
func (p *SwpMan) swpManSetObj2Lit(Id int, Lit int) {
	panic("TODO, stubbed")
}

// Swp_ManStart (static inline)
// C: Swp_Man_t * Swp_ManStart( Gia_Man_t * pGia )
func (pGia *GiaManager) swpManStart() *SwpMan {
	panic("TODO, stubbed")
}

// Swp_ManStop (static inline)
// C: void Swp_ManStop( Gia_Man_t * pGia )
func (pGia *GiaManager) swpManStop() {
	panic("TODO, stubbed")
}

// Gia_SweeperStart
// C: Gia_Man_t * Gia_SweeperStart( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaSweeperStart() *GiaManager {
	panic("TODO, stubbed")
}

// Gia_SweeperStop
// C: void Gia_SweeperStop( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaSweeperStop() {
	panic("TODO, stubbed")
}

// Gia_SweeperIsRunning
// C: int Gia_SweeperIsRunning( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaSweeperIsRunning() int {
	panic("TODO, stubbed")
}

// Gia_SweeperMemUsage
// C: double Gia_SweeperMemUsage( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaSweeperMemUsage() float64 {
	panic("TODO, stubbed")
}

// Gia_SweeperPrintStats
// C: void Gia_SweeperPrintStats( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaSweeperPrintStats() {
	panic("TODO, stubbed")
}

// Gia_SweeperSetConflictLimit
// C: void Gia_SweeperSetConflictLimit( Gia_Man_t * p, int nConfMax )
func (p *GiaManager) GiaSweeperSetConflictLimit(nConfMax int) {
	panic("TODO, stubbed")
}

// Gia_SweeperSetRuntimeLimit
// C: void Gia_SweeperSetRuntimeLimit( Gia_Man_t * p, int nSeconds )
func (p *GiaManager) GiaSweeperSetRuntimeLimit(nSeconds int) {
	panic("TODO, stubbed")
}

// Gia_SweeperGetCex
// C: Vec_Int_t * Gia_SweeperGetCex( Gia_Man_t * p )
func (p *GiaManager) GiaSweeperGetCex() *IntVector {
	panic("TODO, stubbed")
}

// Gia_SweeperProbeCreate
// C: int Gia_SweeperProbeCreate( Gia_Man_t * p, int iLit )
func (p *GiaManager) GiaSweeperProbeCreate(iLit int) int {
	panic("TODO, stubbed")
}

// Gia_SweeperProbeDelete
// C: int Gia_SweeperProbeDelete( Gia_Man_t * p, int ProbeId )
func (p *GiaManager) GiaSweeperProbeDelete(ProbeId int) int {
	panic("TODO, stubbed")
}

// Gia_SweeperProbeUpdate
// C: int Gia_SweeperProbeUpdate( Gia_Man_t * p, int ProbeId, int iLitNew )
func (p *GiaManager) GiaSweeperProbeUpdate(ProbeId int, iLitNew int) int {
	panic("TODO, stubbed")
}

// Gia_SweeperProbeLit
// C: int Gia_SweeperProbeLit( Gia_Man_t * p, int ProbeId )
func (p *GiaManager) GiaSweeperProbeLit(ProbeId int) int {
	panic("TODO, stubbed")
}

// Gia_SweeperCollectValidProbeIds
// C: Vec_Int_t * Gia_SweeperCollectValidProbeIds( Gia_Man_t * p )
func (p *GiaManager) GiaSweeperCollectValidProbeIds() *IntVector {
	panic("TODO, stubbed")
}

// Gia_SweeperCondPush
// C: void Gia_SweeperCondPush( Gia_Man_t * p, int ProbeId )
func (p *GiaManager) GiaSweeperCondPush(ProbeId int) {
	panic("TODO, stubbed")
}

// Gia_SweeperCondPop
// C: int Gia_SweeperCondPop( Gia_Man_t * p )
func (p *GiaManager) GiaSweeperCondPop() int {
	panic("TODO, stubbed")
}

// Gia_SweeperCondVector
// C: Vec_Int_t * Gia_SweeperCondVector( Gia_Man_t * p )
func (p *GiaManager) GiaSweeperCondVector() *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManExtract_rec (static inline)
// C: void Gia_ManExtract_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vObjIds )
func (p *GiaManager) giaManExtract_rec(pObj *GiaObject, vObjIds *IntVector) {
	panic("TODO, stubbed")
}

// Gia_SweeperExtractUserLogic
// C: Gia_Man_t * Gia_SweeperExtractUserLogic( Gia_Man_t * p, Vec_Int_t * vProbeIds, Vec_Ptr_t * vInNames, Vec_Ptr_t * vOutNames )
func (p *GiaManager) GiaSweeperExtractUserLogic(vProbeIds *IntVector, vInNames *PointerVector[string], vOutNames *PointerVector[string]) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_SweeperLogicDump
// C: void Gia_SweeperLogicDump( Gia_Man_t * p, Vec_Int_t * vProbeIds, int fDumpConds, char * pFileName )
func (p *GiaManager) GiaSweeperLogicDump(vProbeIds *IntVector, fDumpConds int, pFileName string) {
	panic("TODO, stubbed")
}

// Gia_SweeperCleanup
// C: Gia_Man_t * Gia_SweeperCleanup( Gia_Man_t * p, char * pCommLime )
func (p *GiaManager) GiaSweeperCleanup(pCommLime string) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManAddClausesMux (static inline)
// C: void Gia_ManAddClausesMux( Swp_Man_t * p, Gia_Obj_t * pNode )
func (p *SwpMan) giaManAddClausesMux(pNode *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManAddClausesSuper (static inline)
// C: void Gia_ManAddClausesSuper( Swp_Man_t * p, Gia_Obj_t * pNode, Vec_Int_t * vSuper )
func (p *SwpMan) giaManAddClausesSuper(pNode *GiaObject, vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManCollectSuper_rec (static inline)
// C: void Gia_ManCollectSuper_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSuper )
func (p *GiaManager) giaManCollectSuper_rec(pObj *GiaObject, vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManCollectSuper (static inline)
// C: void Gia_ManCollectSuper( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSuper )
func (p *GiaManager) giaManCollectSuper(pObj *GiaObject, vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManObjAddToFrontier (static inline)
// C: void Gia_ManObjAddToFrontier( Swp_Man_t * p, int Id, Vec_Int_t * vFront )
func (p *SwpMan) giaManObjAddToFrontier(Id int, vFront *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManCnfNodeAddToSolver (static inline)
// C: void Gia_ManCnfNodeAddToSolver( Swp_Man_t * p, int NodeId )
func (p *SwpMan) giaManCnfNodeAddToSolver(NodeId int) {
	panic("TODO, stubbed")
}

// Gia_ManGetCex (static inline)
// C: Vec_Int_t * Gia_ManGetCex( Gia_Man_t * pGia, Vec_Int_t * vId2Lit, sat_solver * pSat, Vec_Int_t * vCex )
func (pGia *GiaManager) giaManGetCex(vId2Lit *IntVector, pSat *SatSolver, vCex *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_SweeperCheckEquiv
// C: int Gia_SweeperCheckEquiv( Gia_Man_t * pGia, int Probe1, int Probe2 )
func (pGia *GiaManager) GiaSweeperCheckEquiv(Probe1 int, Probe2 int) int {
	panic("TODO, stubbed")
}

// Gia_SweeperCondCheckUnsat
// C: int Gia_SweeperCondCheckUnsat( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaSweeperCondCheckUnsat() int {
	panic("TODO, stubbed")
}

// Gia_SweeperGraft
// C: Vec_Int_t * Gia_SweeperGraft( Gia_Man_t * pDst, Vec_Int_t * vProbes, Gia_Man_t * pSrc )
func (pDst *GiaManager) GiaSweeperGraft(vProbes *IntVector, pSrc *GiaManager) *IntVector {
	panic("TODO, stubbed")
}

// Gia_SweeperSweep
// C: Gia_Man_t * Gia_SweeperSweep( Gia_Man_t * p, Vec_Int_t * vProbeOuts, int nWords, int nConfs, int fVerify, int fVerbose )
func (p *GiaManager) GiaSweeperSweep(vProbeOuts *IntVector, nWords int, nConfs int, fVerify int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_SweeperFraig
// C: int Gia_SweeperFraig( Gia_Man_t * p, Vec_Int_t * vProbeIds, char * pCommLime, int nWords, int nConfs, int fVerify, int fVerbose )
func (p *GiaManager) GiaSweeperFraig(vProbeIds *IntVector, pCommLime string, nWords int, nConfs int, fVerify int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_SweeperRun
// C: int Gia_SweeperRun( Gia_Man_t * p, Vec_Int_t * vProbeIds, char * pCommLime, int fVerbose )
func (p *GiaManager) GiaSweeperRun(vProbeIds *IntVector, pCommLime string, fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_SweeperFraigTest
// C: Gia_Man_t * Gia_SweeperFraigTest( Gia_Man_t * pInit, int nWords, int nConfs, int fVerbose )
func (pInit *GiaManager) GiaSweeperFraigTest(nWords int, nConfs int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
