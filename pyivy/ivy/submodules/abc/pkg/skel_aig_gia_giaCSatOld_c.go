package abc

// Hand-checked skeleton of src/aig/gia/giaCSatOld.c.

// Cbs0Par is Cbs0_Par_t.
type Cbs0Par struct {
	// conflict limits
	nBTLimit   int // limit on the number of conflicts
	nJustLimit int // limit on the size of justification queue
	// current parameters
	nBTThis    int // number of conflicts
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

// Cbs0Que is Cbs0_Que_t.
type Cbs0Que struct {
	iHead int          // beginning of the queue
	iTail int          // end of the queue
	nSize int          // allocated size
	pData []*GiaObject // Gia_Obj_t ** : nodes stored in the queue
}

// Cbs0Man is Cbs0_Man_t.
type Cbs0Man struct {
	Pars   Cbs0Par     // parameters
	pAig   *GiaManager // AIG manager
	pProp  Cbs0Que     // propagation queue
	pJust  Cbs0Que     // justification queue
	vModel *IntVector  // satisfying assignment
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

// Cbs0_VarIsAssigned (static inline)
// C: int Cbs0_VarIsAssigned( Gia_Obj_t * pVar )
func cbs0VarIsAssigned(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs0_VarAssign (static inline)
// C: void Cbs0_VarAssign( Gia_Obj_t * pVar )
func cbs0VarAssign(pVar *GiaObject) {
	panic("TODO, stubbed")
}

// Cbs0_VarUnassign (static inline)
// C: void Cbs0_VarUnassign( Gia_Obj_t * pVar )
func cbs0VarUnassign(pVar *GiaObject) {
	panic("TODO, stubbed")
}

// Cbs0_VarValue (static inline)
// C: int Cbs0_VarValue( Gia_Obj_t * pVar )
func cbs0VarValue(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs0_VarSetValue (static inline)
// C: void Cbs0_VarSetValue( Gia_Obj_t * pVar, int v )
func cbs0VarSetValue(pVar *GiaObject, v int) {
	panic("TODO, stubbed")
}

// Cbs0_VarIsJust (static inline)
// C: int Cbs0_VarIsJust( Gia_Obj_t * pVar )
func cbs0VarIsJust(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs0_VarFanin0Value (static inline)
// C: int Cbs0_VarFanin0Value( Gia_Obj_t * pVar )
func cbs0VarFanin0Value(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs0_VarFanin1Value (static inline)
// C: int Cbs0_VarFanin1Value( Gia_Obj_t * pVar )
func cbs0VarFanin1Value(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs0_SetDefaultParams
// C: void Cbs0_SetDefaultParams( Cbs0_Par_t * pPars )
func Cbs0SetDefaultParams(pPars *Cbs0Par) {
	panic("TODO, stubbed")
}

// Cbs0_ManAlloc
// C: Cbs0_Man_t * Cbs0_ManAlloc()
func Cbs0ManAlloc() *Cbs0Man {
	panic("TODO, stubbed")
}

// Cbs0_ManStop
// C: void Cbs0_ManStop( Cbs0_Man_t * p )
func (p *Cbs0Man) Cbs0ManStop() {
	panic("TODO, stubbed")
}

// Cbs0_ReadModel
// C: Vec_Int_t * Cbs0_ReadModel( Cbs0_Man_t * p )
func (p *Cbs0Man) Cbs0ReadModel() *IntVector {
	panic("TODO, stubbed")
}

// Cbs0_ManCheckLimits (static inline)
// C: int Cbs0_ManCheckLimits( Cbs0_Man_t * p )
func (p *Cbs0Man) cbs0ManCheckLimits() int {
	panic("TODO, stubbed")
}

// Cbs0_ManSaveModel (static inline)
// C: void Cbs0_ManSaveModel( Cbs0_Man_t * p, Vec_Int_t * vCex )
func (p *Cbs0Man) cbs0ManSaveModel(vCex *IntVector) {
	panic("TODO, stubbed")
}

// Cbs0_QueIsEmpty (static inline)
// C: int Cbs0_QueIsEmpty( Cbs0_Que_t * p )
func (p *Cbs0Que) cbs0QueIsEmpty() int {
	panic("TODO, stubbed")
}

// Cbs0_QuePush (static inline)
// C: void Cbs0_QuePush( Cbs0_Que_t * p, Gia_Obj_t * pObj )
func (p *Cbs0Que) cbs0QuePush(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Cbs0_QueHasNode (static inline)
// C: int Cbs0_QueHasNode( Cbs0_Que_t * p, Gia_Obj_t * pObj )
func (p *Cbs0Que) cbs0QueHasNode(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs0_QueStore (static inline)
// C: void Cbs0_QueStore( Cbs0_Que_t * p, int * piHeadOld, int * piTailOld )
func (p *Cbs0Que) cbs0QueStore(piHeadOld *int, piTailOld *int) {
	panic("TODO, stubbed")
}

// Cbs0_QueRestore (static inline)
// C: void Cbs0_QueRestore( Cbs0_Que_t * p, int iHeadOld, int iTailOld )
func (p *Cbs0Que) cbs0QueRestore(iHeadOld int, iTailOld int) {
	panic("TODO, stubbed")
}

// Cbs0_VarFaninFanoutMax (static inline)
// C: int Cbs0_VarFaninFanoutMax( Cbs0_Man_t * p, Gia_Obj_t * pObj )
func (p *Cbs0Man) cbs0VarFaninFanoutMax(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs0_ManDecideHighest (static inline)
// C: Gia_Obj_t * Cbs0_ManDecideHighest( Cbs0_Man_t * p )
func (p *Cbs0Man) cbs0ManDecideHighest() *GiaObject {
	panic("TODO, stubbed")
}

// Cbs0_ManDecideLowest (static inline)
// C: Gia_Obj_t * Cbs0_ManDecideLowest( Cbs0_Man_t * p )
func (p *Cbs0Man) cbs0ManDecideLowest() *GiaObject {
	panic("TODO, stubbed")
}

// Cbs0_ManDecideMaxFF (static inline)
// C: Gia_Obj_t * Cbs0_ManDecideMaxFF( Cbs0_Man_t * p )
func (p *Cbs0Man) cbs0ManDecideMaxFF() *GiaObject {
	panic("TODO, stubbed")
}

// Cbs0_ManCancelUntil (static inline)
// C: void Cbs0_ManCancelUntil( Cbs0_Man_t * p, int iBound )
func (p *Cbs0Man) cbs0ManCancelUntil(iBound int) {
	panic("TODO, stubbed")
}

// Cbs0_ManAssign (static inline)
// C: void Cbs0_ManAssign( Cbs0_Man_t * p, Gia_Obj_t * pObj )
func (p *Cbs0Man) cbs0ManAssign(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Cbs0_ManPropagateOne (static inline)
// C: int Cbs0_ManPropagateOne( Cbs0_Man_t * p, Gia_Obj_t * pVar )
func (p *Cbs0Man) cbs0ManPropagateOne(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs0_ManPropagateTwo (static inline)
// C: int Cbs0_ManPropagateTwo( Cbs0_Man_t * p, Gia_Obj_t * pVar )
func (p *Cbs0Man) cbs0ManPropagateTwo(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs0_ManPropagate
// C: int Cbs0_ManPropagate( Cbs0_Man_t * p )
func (p *Cbs0Man) Cbs0ManPropagate() int {
	panic("TODO, stubbed")
}

// Cbs0_ManSolve_rec
// C: int Cbs0_ManSolve_rec( Cbs0_Man_t * p )
func (p *Cbs0Man) Cbs0ManSolve_rec() int {
	panic("TODO, stubbed")
}

// Cbs0_ManSolve
// C: int Cbs0_ManSolve( Cbs0_Man_t * p, Gia_Obj_t * pObj )
func (p *Cbs0Man) Cbs0ManSolve(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Cbs0_ManSatPrintStats
// C: void Cbs0_ManSatPrintStats( Cbs0_Man_t * p )
func (p *Cbs0Man) Cbs0ManSatPrintStats() {
	panic("TODO, stubbed")
}

// Cbs_ManSolveMiter
// C: Vec_Int_t * Cbs_ManSolveMiter( Gia_Man_t * pAig, int nConfs, Vec_Str_t ** pvStatus, int fVerbose )
func (pAig *GiaManager) CbsManSolveMiter(nConfs int, pvStatus **ByteVector, fVerbose int) *IntVector {
	panic("TODO, stubbed")
}
