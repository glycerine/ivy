package abc

// Hand-checked skeleton of src/aig/gia/giaCSat.c.

// CbsPar is Cbs_Par_t.
type CbsPar struct {
	// conflict limits
	nBTLimit   int // limit on the number of conflicts
	nJustLimit int // limit on the size of justification queue
	// current parameters
	nBTThis    int // number of conflicts
	nBTThisNc  int // number of conflicts
	nJustThis  int // max size of the frontier
	nBTTotal   int // total number of conflicts
	nJustTotal int // total size of the frontier
	// decision heuristics
	fUseHighest int // use node with the highest ID
	fUseLowest  int // use node with the highest ID
	fUseMaxFF   int // use node with the largest fanin fanout
	// other
	fVerbose int
}

// CbsQue is Cbs_Que_t.
type CbsQue struct {
	iHead int          // beginning of the queue
	iTail int          // end of the queue
	nSize int          // allocated size
	pData []*GiaObject // Gia_Obj_t ** : nodes stored in the queue
}

// CbsMan is Cbs_Man_t.
type CbsMan struct {
	Pars     CbsPar                    // parameters
	pAig     *GiaManager               // AIG manager
	pProp    CbsQue                    // propagation queue
	pJust    CbsQue                    // justification queue
	pClauses CbsQue                    // clause queue
	pIter    []*GiaObject              // Gia_Obj_t ** : iterator through clause vars
	vLevReas *IntVector                // levels and decisions
	vModel   *IntVector                // satisfying assignment
	vTemp    *PointerVector[GiaObject] // temporary storage
	// SAT calls statistics
	nSatUnsat int // the number of proofs
	nSatSat   int // the number of failure
	nSatUndec int // the number of timeouts
	nSatTotal int // the number of calls
	// conflicts
	nConfUnsat int // conflicts in unsat problems
	nConfSat   int // conflicts in sat problems
	nConfUndec int // conflicts in undec problems
	// runtime stats
	timeSatUnsat int64 // abctime : unsat
	timeSatSat   int64 // abctime : sat
	timeSatUndec int64 // abctime : undecided
	timeTotal    int64 // abctime : total runtime
}

// Cbs_VarIsAssigned (static inline)
// C: int Cbs_VarIsAssigned( Gia_Obj_t * pVar )
func cbsVarIsAssigned(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs_VarAssign (static inline)
// C: void Cbs_VarAssign( Gia_Obj_t * pVar )
func cbsVarAssign(pVar *GiaObject) {
	panic("TODO, stubbed")
}

// Cbs_VarUnassign (static inline)
// C: void Cbs_VarUnassign( Gia_Obj_t * pVar )
func cbsVarUnassign(pVar *GiaObject) {
	panic("TODO, stubbed")
}

// Cbs_VarValue (static inline)
// C: int Cbs_VarValue( Gia_Obj_t * pVar )
func cbsVarValue(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs_VarSetValue (static inline)
// C: void Cbs_VarSetValue( Gia_Obj_t * pVar, int v )
func cbsVarSetValue(pVar *GiaObject, v int) {
	panic("TODO, stubbed")
}

// Cbs_VarIsJust (static inline)
// C: int Cbs_VarIsJust( Gia_Obj_t * pVar )
func cbsVarIsJust(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs_VarFanin0Value (static inline)
// C: int Cbs_VarFanin0Value( Gia_Obj_t * pVar )
func cbsVarFanin0Value(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs_VarFanin1Value (static inline)
// C: int Cbs_VarFanin1Value( Gia_Obj_t * pVar )
func cbsVarFanin1Value(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs_VarDecLevel (static inline)
// C: int Cbs_VarDecLevel( Cbs_Man_t * p, Gia_Obj_t * pVar )
func (p *CbsMan) cbsVarDecLevel(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs_VarReason0 (static inline)
// C: Gia_Obj_t * Cbs_VarReason0( Cbs_Man_t * p, Gia_Obj_t * pVar )
func (p *CbsMan) cbsVarReason0(pVar *GiaObject) *GiaObject {
	panic("TODO, stubbed")
}

// Cbs_VarReason1 (static inline)
// C: Gia_Obj_t * Cbs_VarReason1( Cbs_Man_t * p, Gia_Obj_t * pVar )
func (p *CbsMan) cbsVarReason1(pVar *GiaObject) *GiaObject {
	panic("TODO, stubbed")
}

// Cbs_ClauseDecLevel (static inline)
// C: int Cbs_ClauseDecLevel( Cbs_Man_t * p, int hClause )
func (p *CbsMan) cbsClauseDecLevel(hClause int) int {
	panic("TODO, stubbed")
}

// Cbs_SetDefaultParams
// C: void Cbs_SetDefaultParams( Cbs_Par_t * pPars )
func CbsSetDefaultParams(pPars *CbsPar) {
	panic("TODO, stubbed")
}

// Cbs_ManSetConflictNum
// C: void Cbs_ManSetConflictNum( Cbs_Man_t * p, int Num )
func (p *CbsMan) CbsManSetConflictNum(Num int) {
	panic("TODO, stubbed")
}

// Cbs_ManAlloc
// C: Cbs_Man_t * Cbs_ManAlloc( Gia_Man_t * pGia )
func (pGia *GiaManager) CbsManAlloc() *CbsMan {
	panic("TODO, stubbed")
}

// Cbs_ManStop
// C: void Cbs_ManStop( Cbs_Man_t * p )
func (p *CbsMan) CbsManStop() {
	panic("TODO, stubbed")
}

// Cbs_ReadModel
// C: Vec_Int_t * Cbs_ReadModel( Cbs_Man_t * p )
func (p *CbsMan) CbsReadModel() *IntVector {
	panic("TODO, stubbed")
}

// Cbs_ManCheckLimits (static inline)
// C: int Cbs_ManCheckLimits( Cbs_Man_t * p )
func (p *CbsMan) cbsManCheckLimits() int {
	panic("TODO, stubbed")
}

// Cbs_ManSaveModel (static inline)
// C: void Cbs_ManSaveModel( Cbs_Man_t * p, Vec_Int_t * vCex )
func (p *CbsMan) cbsManSaveModel(vCex *IntVector) {
	panic("TODO, stubbed")
}

// Cbs_ManSaveModelAll (static inline)
// C: void Cbs_ManSaveModelAll( Cbs_Man_t * p, Vec_Int_t * vCex )
func (p *CbsMan) cbsManSaveModelAll(vCex *IntVector) {
	panic("TODO, stubbed")
}

// Cbs_QueIsEmpty (static inline)
// C: int Cbs_QueIsEmpty( Cbs_Que_t * p )
func (p *CbsQue) cbsQueIsEmpty() int {
	panic("TODO, stubbed")
}

// Cbs_QuePush (static inline)
// C: void Cbs_QuePush( Cbs_Que_t * p, Gia_Obj_t * pObj )
func (p *CbsQue) cbsQuePush(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Cbs_QueHasNode (static inline)
// C: int Cbs_QueHasNode( Cbs_Que_t * p, Gia_Obj_t * pObj )
func (p *CbsQue) cbsQueHasNode(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs_QueStore (static inline)
// C: void Cbs_QueStore( Cbs_Que_t * p, int * piHeadOld, int * piTailOld )
func (p *CbsQue) cbsQueStore(piHeadOld *int, piTailOld *int) {
	panic("TODO, stubbed")
}

// Cbs_QueRestore (static inline)
// C: void Cbs_QueRestore( Cbs_Que_t * p, int iHeadOld, int iTailOld )
func (p *CbsQue) cbsQueRestore(iHeadOld int, iTailOld int) {
	panic("TODO, stubbed")
}

// Cbs_QueFinish (static inline)
// C: int Cbs_QueFinish( Cbs_Que_t * p )
func (p *CbsQue) cbsQueFinish() int {
	panic("TODO, stubbed")
}

// Cbs_VarFaninFanoutMax (static inline)
// C: int Cbs_VarFaninFanoutMax( Cbs_Man_t * p, Gia_Obj_t * pObj )
func (p *CbsMan) cbsVarFaninFanoutMax(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs_ManDecideHighest (static inline)
// C: Gia_Obj_t * Cbs_ManDecideHighest( Cbs_Man_t * p )
func (p *CbsMan) cbsManDecideHighest() *GiaObject {
	panic("TODO, stubbed")
}

// Cbs_ManDecideLowest (static inline)
// C: Gia_Obj_t * Cbs_ManDecideLowest( Cbs_Man_t * p )
func (p *CbsMan) cbsManDecideLowest() *GiaObject {
	panic("TODO, stubbed")
}

// Cbs_ManDecideMaxFF (static inline)
// C: Gia_Obj_t * Cbs_ManDecideMaxFF( Cbs_Man_t * p )
func (p *CbsMan) cbsManDecideMaxFF() *GiaObject {
	panic("TODO, stubbed")
}

// Cbs_ManCancelUntil (static inline)
// C: void Cbs_ManCancelUntil( Cbs_Man_t * p, int iBound )
func (p *CbsMan) cbsManCancelUntil(iBound int) {
	panic("TODO, stubbed")
}

// Cbs_ManAssign (static inline)
// C: void Cbs_ManAssign( Cbs_Man_t * p, Gia_Obj_t * pObj, int Level, Gia_Obj_t * pRes0, Gia_Obj_t * pRes1 )
func (p *CbsMan) cbsManAssign(pObj *GiaObject, Level int, pRes0 *GiaObject, pRes1 *GiaObject) {
	panic("TODO, stubbed")
}

// Cbs_ManClauseSize (static inline)
// C: int Cbs_ManClauseSize( Cbs_Man_t * p, int hClause )
func (p *CbsMan) cbsManClauseSize(hClause int) int {
	panic("TODO, stubbed")
}

// Cbs_ManPrintClause (static inline)
// C: void Cbs_ManPrintClause( Cbs_Man_t * p, int Level, int hClause )
func (p *CbsMan) cbsManPrintClause(Level int, hClause int) {
	panic("TODO, stubbed")
}

// Cbs_ManPrintClauseNew (static inline)
// C: void Cbs_ManPrintClauseNew( Cbs_Man_t * p, int Level, int hClause )
func (p *CbsMan) cbsManPrintClauseNew(Level int, hClause int) {
	panic("TODO, stubbed")
}

// Cbs_ManDeriveReason (static inline)
// C: void Cbs_ManDeriveReason( Cbs_Man_t * p, int Level )
func (p *CbsMan) cbsManDeriveReason(Level int) {
	panic("TODO, stubbed")
}

// Cbs_ManAnalyze (static inline)
// C: int Cbs_ManAnalyze( Cbs_Man_t * p, int Level, Gia_Obj_t * pVar, Gia_Obj_t * pFan0, Gia_Obj_t * pFan1 )
func (p *CbsMan) cbsManAnalyze(Level int, pVar *GiaObject, pFan0 *GiaObject, pFan1 *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs_ManResolve (static inline)
// C: int Cbs_ManResolve( Cbs_Man_t * p, int Level, int hClause0, int hClause1 )
func (p *CbsMan) cbsManResolve(Level int, hClause0 int, hClause1 int) int {
	panic("TODO, stubbed")
}

// Cbs_ManPropagateOne (static inline)
// C: int Cbs_ManPropagateOne( Cbs_Man_t * p, Gia_Obj_t * pVar, int Level )
func (p *CbsMan) cbsManPropagateOne(pVar *GiaObject, Level int) int {
	panic("TODO, stubbed")
}

// Cbs_ManPropagateTwo (static inline)
// C: int Cbs_ManPropagateTwo( Cbs_Man_t * p, Gia_Obj_t * pVar, int Level )
func (p *CbsMan) cbsManPropagateTwo(pVar *GiaObject, Level int) int {
	panic("TODO, stubbed")
}

// Cbs_ManPropagate
// C: int Cbs_ManPropagate( Cbs_Man_t * p, int Level )
func (p *CbsMan) CbsManPropagate(Level int) int {
	panic("TODO, stubbed")
}

// Cbs_ManSolve_rec
// C: int Cbs_ManSolve_rec( Cbs_Man_t * p, int Level )
func (p *CbsMan) CbsManSolve_rec(Level int) int {
	panic("TODO, stubbed")
}

// Cbs_ManSolve
// C: int Cbs_ManSolve( Cbs_Man_t * p, Gia_Obj_t * pObj )
func (p *CbsMan) CbsManSolve(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs_ManSolve2
// C: int Cbs_ManSolve2( Cbs_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pObj2 )
func (p *CbsMan) CbsManSolve2(pObj *GiaObject, pObj2 *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs_ManSatPrintStats
// C: void Cbs_ManSatPrintStats( Cbs_Man_t * p )
func (p *CbsMan) CbsManSatPrintStats() {
	panic("TODO, stubbed")
}

// Cbs_ManSolveMiterNc
// C: Vec_Int_t * Cbs_ManSolveMiterNc( Gia_Man_t * pAig, int nConfs, Vec_Str_t ** pvStatus, int f0Proved, int fVerbose )
func (pAig *GiaManager) CbsManSolveMiterNc(nConfs int, pvStatus **ByteVector, f0Proved int, fVerbose int) *IntVector {
	panic("TODO, stubbed")
}
