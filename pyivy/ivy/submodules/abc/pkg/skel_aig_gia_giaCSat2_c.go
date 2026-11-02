package abc

// Hand-checked skeleton of src/aig/gia/giaCSat2.c.

// Cbs2Par is Cbs2_Par_t.
type Cbs2Par struct {
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
	fUseFanout  int // use node with the largest fanin fanout
	// other
	fVerbose int
}

// Cbs2Que is Cbs2_Que_t.
type Cbs2Que struct {
	iHead int   // beginning of the queue
	iTail int   // end of the queue
	nSize int   // allocated size
	pData []int // int * : nodes stored in the queue
}

// Cbs2Man is Cbs2_Man_t.
type Cbs2Man struct {
	Pars     Cbs2Par     // parameters
	pAig     *GiaManager // AIG manager
	pProp    Cbs2Que     // propagation queue
	pJust    Cbs2Que     // justification queue
	pClauses Cbs2Que     // clause queue
	vModel   *IntVector  // satisfying assignment
	vTemp    *IntVector  // temporary storage
	// internal data
	vAssign    ByteVector // Vec_Str_t (by value)
	vMark      ByteVector // Vec_Str_t (by value)
	vLevReason IntVector  // Vec_Int_t (by value)
	vWatches   IntVector  // Vec_Int_t (by value)
	vWatchUpds IntVector  // Vec_Int_t (by value)
	vFanoutN   IntVector  // Vec_Int_t (by value)
	vFanout0   IntVector  // Vec_Int_t (by value)
	vActivity  IntVector  // Vec_Int_t (by value)
	vActStore  IntVector  // Vec_Int_t (by value)
	vJStore    IntVector  // Vec_Int_t (by value)
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
	timeJFront   int64 // abctime
	timeSatUnsat int64 // abctime : unsat
	timeSatSat   int64 // abctime : sat
	timeSatUndec int64 // abctime : undecided
	timeTotal    int64 // abctime : total runtime
	// other statistics
	nPropCalls  [3]int
	nFails      [2]int
	nClauseConf int
}

// Cbs2_VarUnused (static inline)
// C: int Cbs2_VarUnused( Cbs2_Man_t * p, int iVar )
func (p *Cbs2Man) cbs2VarUnused(iVar int) int {
	panic("TODO, stubbed")
}

// Cbs2_VarSetUnused (static inline)
// C: void Cbs2_VarSetUnused( Cbs2_Man_t * p, int iVar )
func (p *Cbs2Man) cbs2VarSetUnused(iVar int) {
	panic("TODO, stubbed")
}

// Cbs2_VarMark0 (static inline)
// C: int Cbs2_VarMark0( Cbs2_Man_t * p, int iVar )
func (p *Cbs2Man) cbs2VarMark0(iVar int) int {
	panic("TODO, stubbed")
}

// Cbs2_VarSetMark0 (static inline)
// C: void Cbs2_VarSetMark0( Cbs2_Man_t * p, int iVar, int Value )
func (p *Cbs2Man) cbs2VarSetMark0(iVar int, Value int) {
	panic("TODO, stubbed")
}

// Cbs2_VarIsAssigned (static inline)
// C: int Cbs2_VarIsAssigned( Cbs2_Man_t * p, int iVar )
func (p *Cbs2Man) cbs2VarIsAssigned(iVar int) int {
	panic("TODO, stubbed")
}

// Cbs2_VarUnassign (static inline)
// C: void Cbs2_VarUnassign( Cbs2_Man_t * p, int iVar )
func (p *Cbs2Man) cbs2VarUnassign(iVar int) {
	panic("TODO, stubbed")
}

// Cbs2_VarValue (static inline)
// C: int Cbs2_VarValue( Cbs2_Man_t * p, int iVar )
func (p *Cbs2Man) cbs2VarValue(iVar int) int {
	panic("TODO, stubbed")
}

// Cbs2_VarSetValue (static inline)
// C: void Cbs2_VarSetValue( Cbs2_Man_t * p, int iVar, int v )
func (p *Cbs2Man) cbs2VarSetValue(iVar int, v int) {
	panic("TODO, stubbed")
}

// Cbs2_VarIsJust (static inline)
// C: int Cbs2_VarIsJust( Cbs2_Man_t * p, Gia_Obj_t * pVar, int iVar )
func (p *Cbs2Man) cbs2VarIsJust(pVar *GiaObject, iVar int) int {
	panic("TODO, stubbed")
}

// Cbs2_VarFanin0Value (static inline)
// C: int Cbs2_VarFanin0Value( Cbs2_Man_t * p, Gia_Obj_t * pVar, int iVar )
func (p *Cbs2Man) cbs2VarFanin0Value(pVar *GiaObject, iVar int) int {
	panic("TODO, stubbed")
}

// Cbs2_VarFanin1Value (static inline)
// C: int Cbs2_VarFanin1Value( Cbs2_Man_t * p, Gia_Obj_t * pVar, int iVar )
func (p *Cbs2Man) cbs2VarFanin1Value(pVar *GiaObject, iVar int) int {
	panic("TODO, stubbed")
}

// Cbs2_VarDecLevel (static inline)
// C: int Cbs2_VarDecLevel( Cbs2_Man_t * p, int iVar )
func (p *Cbs2Man) cbs2VarDecLevel(iVar int) int {
	panic("TODO, stubbed")
}

// Cbs2_VarReason0 (static inline)
// C: int Cbs2_VarReason0( Cbs2_Man_t * p, int iVar )
func (p *Cbs2Man) cbs2VarReason0(iVar int) int {
	panic("TODO, stubbed")
}

// Cbs2_VarReason1 (static inline)
// C: int Cbs2_VarReason1( Cbs2_Man_t * p, int iVar )
func (p *Cbs2Man) cbs2VarReason1(iVar int) int {
	panic("TODO, stubbed")
}

// Cbs2_VarReasonP (static inline)
// C: int * Cbs2_VarReasonP( Cbs2_Man_t * p, int iVar )
func (p *Cbs2Man) cbs2VarReasonP(iVar int) []int {
	panic("TODO, stubbed")
}

// Cbs2_ClauseSize (static inline)
// C: int Cbs2_ClauseSize( Cbs2_Man_t * p, int hClause )
func (p *Cbs2Man) cbs2ClauseSize(hClause int) int {
	panic("TODO, stubbed")
}

// Cbs2_ClauseLits (static inline)
// C: int * Cbs2_ClauseLits( Cbs2_Man_t * p, int hClause )
func (p *Cbs2Man) cbs2ClauseLits(hClause int) []int {
	panic("TODO, stubbed")
}

// Cbs2_ClauseLit (static inline)
// C: int Cbs2_ClauseLit( Cbs2_Man_t * p, int hClause, int i )
func (p *Cbs2Man) cbs2ClauseLit(hClause int, i int) int {
	panic("TODO, stubbed")
}

// Cbs2_ClauseNext1p (static inline)
// C: int * Cbs2_ClauseNext1p( Cbs2_Man_t * p, int hClause )
func (p *Cbs2Man) cbs2ClauseNext1p(hClause int) []int {
	panic("TODO, stubbed")
}

// Cbs2_ClauseSetSize (static inline)
// C: void Cbs2_ClauseSetSize( Cbs2_Man_t * p, int hClause, int x )
func (p *Cbs2Man) cbs2ClauseSetSize(hClause int, x int) {
	panic("TODO, stubbed")
}

// Cbs2_ClauseSetLit (static inline)
// C: void Cbs2_ClauseSetLit( Cbs2_Man_t * p, int hClause, int i, int x )
func (p *Cbs2Man) cbs2ClauseSetLit(hClause int, i int, x int) {
	panic("TODO, stubbed")
}

// Cbs2_ClauseSetNext (static inline)
// C: void Cbs2_ClauseSetNext( Cbs2_Man_t * p, int hClause, int n, int x )
func (p *Cbs2Man) cbs2ClauseSetNext(hClause int, n int, x int) {
	panic("TODO, stubbed")
}

// Cbs2_SetDefaultParams
// C: void Cbs2_SetDefaultParams( Cbs2_Par_t * pPars )
func Cbs2SetDefaultParams(pPars *Cbs2Par) {
	panic("TODO, stubbed")
}

// Cbs2_ManSetConflictNum
// C: void Cbs2_ManSetConflictNum( Cbs2_Man_t * p, int Num )
func (p *Cbs2Man) Cbs2ManSetConflictNum(Num int) {
	panic("TODO, stubbed")
}

// Cbs2_ManAlloc
// C: Cbs2_Man_t * Cbs2_ManAlloc( Gia_Man_t * pGia )
func (pGia *GiaManager) Cbs2ManAlloc() *Cbs2Man {
	panic("TODO, stubbed")
}

// Cbs2_ManStop
// C: void Cbs2_ManStop( Cbs2_Man_t * p )
func (p *Cbs2Man) Cbs2ManStop() {
	panic("TODO, stubbed")
}

// Cbs2_ReadModel
// C: Vec_Int_t * Cbs2_ReadModel( Cbs2_Man_t * p )
func (p *Cbs2Man) Cbs2ReadModel() *IntVector {
	panic("TODO, stubbed")
}

// Cbs2_ManCheckLimits (static inline)
// C: int Cbs2_ManCheckLimits( Cbs2_Man_t * p )
func (p *Cbs2Man) cbs2ManCheckLimits() int {
	panic("TODO, stubbed")
}

// Cbs2_ManSaveModel (static inline)
// C: void Cbs2_ManSaveModel( Cbs2_Man_t * p, Vec_Int_t * vCex )
func (p *Cbs2Man) cbs2ManSaveModel(vCex *IntVector) {
	panic("TODO, stubbed")
}

// Cbs2_ManSaveModelAll (static inline)
// C: void Cbs2_ManSaveModelAll( Cbs2_Man_t * p, Vec_Int_t * vCex )
func (p *Cbs2Man) cbs2ManSaveModelAll(vCex *IntVector) {
	panic("TODO, stubbed")
}

// Cbs2_QueIsEmpty (static inline)
// C: int Cbs2_QueIsEmpty( Cbs2_Que_t * p )
func (p *Cbs2Que) cbs2QueIsEmpty() int {
	panic("TODO, stubbed")
}

// Cbs2_QuePush (static inline)
// C: void Cbs2_QuePush( Cbs2_Que_t * p, int iObj )
func (p *Cbs2Que) cbs2QuePush(iObj int) {
	panic("TODO, stubbed")
}

// Cbs2_QueGrow (static inline)
// C: void Cbs2_QueGrow( Cbs2_Que_t * p, int Plus )
func (p *Cbs2Que) cbs2QueGrow(Plus int) {
	panic("TODO, stubbed")
}

// Cbs2_QueHasNode (static inline)
// C: int Cbs2_QueHasNode( Cbs2_Que_t * p, int iObj )
func (p *Cbs2Que) cbs2QueHasNode(iObj int) int {
	panic("TODO, stubbed")
}

// Cbs2_QueStore (static inline)
// C: void Cbs2_QueStore( Cbs2_Que_t * p, int * piHeadOld, int * piTailOld )
func (p *Cbs2Que) cbs2QueStore(piHeadOld *int, piTailOld *int) {
	panic("TODO, stubbed")
}

// Cbs2_QueRestore (static inline)
// C: void Cbs2_QueRestore( Cbs2_Que_t * p, int iHeadOld, int iTailOld )
func (p *Cbs2Que) cbs2QueRestore(iHeadOld int, iTailOld int) {
	panic("TODO, stubbed")
}

// Cbs2_VarFaninFanoutMax (static inline)
// C: int Cbs2_VarFaninFanoutMax( Cbs2_Man_t * p, Gia_Obj_t * pObj )
func (p *Cbs2Man) cbs2VarFaninFanoutMax(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs2_ManDecideHighest (static inline)
// C: int Cbs2_ManDecideHighest( Cbs2_Man_t * p )
func (p *Cbs2Man) cbs2ManDecideHighest() int {
	panic("TODO, stubbed")
}

// Cbs2_ManDecideLowest (static inline)
// C: Gia_Obj_t * Cbs2_ManDecideLowest( Cbs2_Man_t * p )
func (p *Cbs2Man) cbs2ManDecideLowest() *GiaObject {
	panic("TODO, stubbed")
}

// Cbs2_ManDecideMaxFF (static inline)
// C: Gia_Obj_t * Cbs2_ManDecideMaxFF( Cbs2_Man_t * p )
func (p *Cbs2Man) cbs2ManDecideMaxFF() *GiaObject {
	panic("TODO, stubbed")
}

// Cbs2_ManCancelUntil (static inline)
// C: void Cbs2_ManCancelUntil( Cbs2_Man_t * p, int iBound )
func (p *Cbs2Man) cbs2ManCancelUntil(iBound int) {
	panic("TODO, stubbed")
}

// Cbs2_ManAssign (static inline)
// C: void Cbs2_ManAssign( Cbs2_Man_t * p, int iLit, int Level, int iRes0, int iRes1 )
func (p *Cbs2Man) cbs2ManAssign(iLit int, Level int, iRes0 int, iRes1 int) {
	panic("TODO, stubbed")
}

// Cbs2_ManPrintClause (static inline)
// C: void Cbs2_ManPrintClause( Cbs2_Man_t * p, int Level, int hClause )
func (p *Cbs2Man) cbs2ManPrintClause(Level int, hClause int) {
	panic("TODO, stubbed")
}

// Cbs2_ManPrintCube (static inline)
// C: void Cbs2_ManPrintCube( Cbs2_Man_t * p, int Level, int hClause )
func (p *Cbs2Man) cbs2ManPrintCube(Level int, hClause int) {
	panic("TODO, stubbed")
}

// Cbs2_ManBumpClause (static inline)
// C: void Cbs2_ManBumpClause( Cbs2_Man_t * p, int hClause )
func (p *Cbs2Man) cbs2ManBumpClause(hClause int) {
	panic("TODO, stubbed")
}

// Cbs2_ManBumpClean (static inline)
// C: void Cbs2_ManBumpClean( Cbs2_Man_t * p )
func (p *Cbs2Man) cbs2ManBumpClean() {
	panic("TODO, stubbed")
}

// Cbs2_ManCleanWatch (static inline)
// C: void Cbs2_ManCleanWatch( Cbs2_Man_t * p )
func (p *Cbs2Man) cbs2ManCleanWatch() {
	panic("TODO, stubbed")
}

// Cbs2_ManWatchClause (static inline)
// C: void Cbs2_ManWatchClause( Cbs2_Man_t * p, int hClause, int Lit )
func (p *Cbs2Man) cbs2ManWatchClause(hClause int, Lit int) {
	panic("TODO, stubbed")
}

// Cbs2_QueFinish (static inline)
// C: int Cbs2_QueFinish( Cbs2_Man_t * p, int Level )
func (p *Cbs2Man) cbs2QueFinish(Level int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManDeriveReason (static inline)
// C: int Cbs2_ManDeriveReason( Cbs2_Man_t * p, int Level )
func (p *Cbs2Man) cbs2ManDeriveReason(Level int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManAnalyze (static inline)
// C: int Cbs2_ManAnalyze( Cbs2_Man_t * p, int Level, int iVar, int iFan0, int iFan1 )
func (p *Cbs2Man) cbs2ManAnalyze(Level int, iVar int, iFan0 int, iFan1 int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManPropagateClauses (static inline)
// C: int Cbs2_ManPropagateClauses( Cbs2_Man_t * p, int Level, int Lit )
func (p *Cbs2Man) cbs2ManPropagateClauses(Level int, Lit int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManResolve (static inline)
// C: int Cbs2_ManResolve( Cbs2_Man_t * p, int Level, int hClause0, int hClause1 )
func (p *Cbs2Man) cbs2ManResolve(Level int, hClause0 int, hClause1 int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManPropagateOne (static inline)
// C: int Cbs2_ManPropagateOne( Cbs2_Man_t * p, int iVar, int Level )
func (p *Cbs2Man) cbs2ManPropagateOne(iVar int, Level int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManPropagateTwo (static inline)
// C: int Cbs2_ManPropagateTwo( Cbs2_Man_t * p, int iVar, int Level )
func (p *Cbs2Man) cbs2ManPropagateTwo(iVar int, Level int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManPropagateUnassigned (static inline)
// C: void Cbs2_ManPropagateUnassigned( Cbs2_Man_t * p, int iVar, int Level )
func (p *Cbs2Man) cbs2ManPropagateUnassigned(iVar int, Level int) {
	panic("TODO, stubbed")
}

// Cbs2_ManPropagate
// C: int Cbs2_ManPropagate( Cbs2_Man_t * p, int Level )
func (p *Cbs2Man) Cbs2ManPropagate(Level int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManPropagate2
// C: int Cbs2_ManPropagate2( Cbs2_Man_t * p, int Level )
func (p *Cbs2Man) Cbs2ManPropagate2(Level int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManUpdateDecVar2 (static inline)
// C: int Cbs2_ManUpdateDecVar2( Cbs2_Man_t * p, int iObj, int iDecLit )
func (p *Cbs2Man) cbs2ManUpdateDecVar2(iObj int, iDecLit int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManUpdateDecVar3 (static inline)
// C: int Cbs2_ManUpdateDecVar3( Cbs2_Man_t * p, int iObj, int iDecLit )
func (p *Cbs2Man) cbs2ManUpdateDecVar3(iObj int, iDecLit int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManUpdateDecVar (static inline)
// C: int Cbs2_ManUpdateDecVar( Cbs2_Man_t * p, int iObj, int iDecLit )
func (p *Cbs2Man) cbs2ManUpdateDecVar(iObj int, iDecLit int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManUpdateFrontier
// C: int Cbs2_ManUpdateFrontier( Cbs2_Man_t * p, int iPropHeadOld, int * piDecLit )
func (p *Cbs2Man) Cbs2ManUpdateFrontier(iPropHeadOld int, piDecLit *int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManSolve1_rec
// C: int Cbs2_ManSolve1_rec( Cbs2_Man_t * p, int Level )
func (p *Cbs2Man) Cbs2ManSolve1_rec(Level int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManSolve2_rec
// C: int Cbs2_ManSolve2_rec( Cbs2_Man_t * p, int Level )
func (p *Cbs2Man) Cbs2ManSolve2_rec(Level int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManSolve_rec
// C: int Cbs2_ManSolve_rec( Cbs2_Man_t * p, int Level )
func (p *Cbs2Man) Cbs2ManSolve_rec(Level int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManSolve
// C: int Cbs2_ManSolve( Cbs2_Man_t * p, int iLit )
func (p *Cbs2Man) Cbs2ManSolve(iLit int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManSolve2
// C: int Cbs2_ManSolve2( Cbs2_Man_t * p, int iLit, int iLit2 )
func (p *Cbs2Man) Cbs2ManSolve2(iLit int, iLit2 int) int {
	panic("TODO, stubbed")
}

// Cbs2_ManSatPrintStats
// C: void Cbs2_ManSatPrintStats( Cbs2_Man_t * p )
func (p *Cbs2Man) Cbs2ManSatPrintStats() {
	panic("TODO, stubbed")
}

// Cbs2_ObjPrintFanouts
// C: void Cbs2_ObjPrintFanouts( Cbs2_Man_t * p, int iObj )
func (p *Cbs2Man) Cbs2ObjPrintFanouts(iObj int) {
	panic("TODO, stubbed")
}

// Cbs2_ManPrintFanouts
// C: void Cbs2_ManPrintFanouts( Cbs2_Man_t * p )
func (p *Cbs2Man) Cbs2ManPrintFanouts() {
	panic("TODO, stubbed")
}

// Cbs2_ObjCreateFanout
// C: void Cbs2_ObjCreateFanout( Cbs2_Man_t * p, int iObj, int iFan0, int iFan1 )
func (p *Cbs2Man) Cbs2ObjCreateFanout(iObj int, iFan0 int, iFan1 int) {
	panic("TODO, stubbed")
}

// Cbs2_ObjDeleteFanout
// C: void Cbs2_ObjDeleteFanout( Cbs2_Man_t * p, int iObj )
func (p *Cbs2Man) Cbs2ObjDeleteFanout(iObj int) {
	panic("TODO, stubbed")
}

// Cbs2_ManCreateFanout_rec
// C: void Cbs2_ManCreateFanout_rec( Cbs2_Man_t * p, int iObj )
func (p *Cbs2Man) Cbs2ManCreateFanout_rec(iObj int) {
	panic("TODO, stubbed")
}

// Cbs2_ManDeleteFanout_rec
// C: void Cbs2_ManDeleteFanout_rec( Cbs2_Man_t * p, int iObj )
func (p *Cbs2Man) Cbs2ManDeleteFanout_rec(iObj int) {
	panic("TODO, stubbed")
}

// Cbs2_ManCheckFanouts
// C: void Cbs2_ManCheckFanouts( Cbs2_Man_t * p )
func (p *Cbs2Man) Cbs2ManCheckFanouts() {
	panic("TODO, stubbed")
}

// Cbs2_ManSolveMiterNc
// C: Vec_Int_t * Cbs2_ManSolveMiterNc( Gia_Man_t * pAig, int nConfs, Vec_Str_t ** pvStatus, int fVerbose )
func (pAig *GiaManager) Cbs2ManSolveMiterNc(nConfs int, pvStatus **ByteVector, fVerbose int) *IntVector {
	panic("TODO, stubbed")
}
