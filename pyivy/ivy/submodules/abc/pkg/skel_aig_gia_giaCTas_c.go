package abc

// Hand-checked skeleton of src/aig/gia/giaCTas.c.

// TasPar is Tas_Par_t.
type TasPar struct {
	// conflict limits
	nBTLimit   int // limit on the number of conflicts
	nJustLimit int // limit on the size of justification queue
	// current parameters
	nBTThis    int // number of conflicts
	nBTThisNc  int // number of conflicts
	nJustThis  int // max size of the frontier
	nBTTotal   int // total number of conflicts
	nJustTotal int // total size of the frontier
	// activity
	VarDecay float32 // variable activity decay
	VarInc   int     // variable increment
	// decision heuristics
	fUseActive  int // use most active
	fUseHighest int // use node with the highest ID
	fUseLowest  int // use node with the highest ID
	fUseMaxFF   int // use node with the largest fanin fanout
	// other
	fVerbose int
}

// TasCls is Tas_Cls_t.
type TasCls struct {
	iNext [2]int // beginning of the queue
	nLits int    // the number of literals
	pLits []int  // int pLits[0] : clause literals (flexible array member)
}

// TasSto is Tas_Sto_t.
type TasSto struct {
	iCur  int   // current position
	nSize int   // allocated size
	pData []int // int * : clause information
}

// TasQue is Tas_Que_t.
type TasQue struct {
	iHead int          // beginning of the queue
	iTail int          // end of the queue
	nSize int          // allocated size
	pData []*GiaObject // Gia_Obj_t ** : nodes stored in the queue
}

// TasMan is Tas_Man_t.
type TasMan struct {
	Pars     TasPar                    // parameters
	pAig     *GiaManager               // AIG manager
	pProp    TasQue                    // propagation queue
	pJust    TasQue                    // justification queue
	pClauses TasQue                    // clause queue
	pIter    []*GiaObject              // Gia_Obj_t ** : iterator through clause vars
	vLevReas *IntVector                // levels and decisions
	vModel   *IntVector                // satisfying assignment
	vTemp    *PointerVector[GiaObject] // temporary storage
	// watched clauses
	pStore     TasSto     // storage for watched clauses
	pWatches   []int      // int * : watched lists for each literal
	vWatchLits *IntVector // lits whose watched are assigned
	nClauses   int        // the counter of clauses
	// activity
	pActivity   []float32  // float * : variable activity
	vActiveVars *IntVector // variables with activity
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

// Tas_VarIsAssigned (static inline)
// C: int Tas_VarIsAssigned( Gia_Obj_t * pVar )
func tasVarIsAssigned(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_VarAssign (static inline)
// C: void Tas_VarAssign( Gia_Obj_t * pVar )
func tasVarAssign(pVar *GiaObject) {
	panic("TODO, stubbed")
}

// Tas_VarUnassign (static inline)
// C: void Tas_VarUnassign( Gia_Obj_t * pVar )
func tasVarUnassign(pVar *GiaObject) {
	panic("TODO, stubbed")
}

// Tas_VarValue (static inline)
// C: int Tas_VarValue( Gia_Obj_t * pVar )
func tasVarValue(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_VarSetValue (static inline)
// C: void Tas_VarSetValue( Gia_Obj_t * pVar, int v )
func tasVarSetValue(pVar *GiaObject, v int) {
	panic("TODO, stubbed")
}

// Tas_VarIsJust (static inline)
// C: int Tas_VarIsJust( Gia_Obj_t * pVar )
func tasVarIsJust(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_VarFanin0Value (static inline)
// C: int Tas_VarFanin0Value( Gia_Obj_t * pVar )
func tasVarFanin0Value(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_VarFanin1Value (static inline)
// C: int Tas_VarFanin1Value( Gia_Obj_t * pVar )
func tasVarFanin1Value(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_VarToLit (static inline)
// C: int Tas_VarToLit( Tas_Man_t * p, Gia_Obj_t * pObj )
func (p *TasMan) tasVarToLit(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_LitIsTrue (static inline)
// C: int Tas_LitIsTrue( Gia_Obj_t * pObj, int Lit )
func tasLitIsTrue(pObj *GiaObject, Lit int) int {
	panic("TODO, stubbed")
}

// Tas_ClsHandle (static inline)
// C: int Tas_ClsHandle( Tas_Man_t * p, Tas_Cls_t * pClause )
func (p *TasMan) tasClsHandle(pClause *TasCls) int {
	panic("TODO, stubbed")
}

// Tas_ClsFromHandle (static inline)
// C: Tas_Cls_t * Tas_ClsFromHandle( Tas_Man_t * p, int h )
func (p *TasMan) tasClsFromHandle(h int) *TasCls {
	panic("TODO, stubbed")
}

// Tas_VarDecLevel (static inline)
// C: int Tas_VarDecLevel( Tas_Man_t * p, Gia_Obj_t * pVar )
func (p *TasMan) tasVarDecLevel(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_VarReason0 (static inline)
// C: Gia_Obj_t * Tas_VarReason0( Tas_Man_t * p, Gia_Obj_t * pVar )
func (p *TasMan) tasVarReason0(pVar *GiaObject) *GiaObject {
	panic("TODO, stubbed")
}

// Tas_VarReason1 (static inline)
// C: Gia_Obj_t * Tas_VarReason1( Tas_Man_t * p, Gia_Obj_t * pVar )
func (p *TasMan) tasVarReason1(pVar *GiaObject) *GiaObject {
	panic("TODO, stubbed")
}

// Tas_ClauseDecLevel (static inline)
// C: int Tas_ClauseDecLevel( Tas_Man_t * p, int hClause )
func (p *TasMan) tasClauseDecLevel(hClause int) int {
	panic("TODO, stubbed")
}

// Tas_VarHasReasonCls (static inline)
// C: int Tas_VarHasReasonCls( Tas_Man_t * p, Gia_Obj_t * pVar )
func (p *TasMan) tasVarHasReasonCls(pVar *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_VarReasonCls (static inline)
// C: Tas_Cls_t * Tas_VarReasonCls( Tas_Man_t * p, Gia_Obj_t * pVar )
func (p *TasMan) tasVarReasonCls(pVar *GiaObject) *TasCls {
	panic("TODO, stubbed")
}

// Tas_SetDefaultParams
// C: void Tas_SetDefaultParams( Tas_Par_t * pPars )
func TasSetDefaultParams(pPars *TasPar) {
	panic("TODO, stubbed")
}

// Tas_ManAlloc
// C: Tas_Man_t * Tas_ManAlloc( Gia_Man_t * pAig, int nBTLimit )
func (pAig *GiaManager) TasManAlloc(nBTLimit int) *TasMan {
	panic("TODO, stubbed")
}

// Tas_ManStop
// C: void Tas_ManStop( Tas_Man_t * p )
func (p *TasMan) TasManStop() {
	panic("TODO, stubbed")
}

// Tas_ReadModel
// C: Vec_Int_t * Tas_ReadModel( Tas_Man_t * p )
func (p *TasMan) TasReadModel() *IntVector {
	panic("TODO, stubbed")
}

// Tas_ManCheckLimits (static inline)
// C: int Tas_ManCheckLimits( Tas_Man_t * p )
func (p *TasMan) tasManCheckLimits() int {
	panic("TODO, stubbed")
}

// Tas_ManSaveModel (static inline)
// C: void Tas_ManSaveModel( Tas_Man_t * p, Vec_Int_t * vCex )
func (p *TasMan) tasManSaveModel(vCex *IntVector) {
	panic("TODO, stubbed")
}

// Tas_QueIsEmpty (static inline)
// C: int Tas_QueIsEmpty( Tas_Que_t * p )
func (p *TasQue) tasQueIsEmpty() int {
	panic("TODO, stubbed")
}

// Tas_QuePush (static inline)
// C: void Tas_QuePush( Tas_Que_t * p, Gia_Obj_t * pObj )
func (p *TasQue) tasQuePush(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Tas_QueHasNode (static inline)
// C: int Tas_QueHasNode( Tas_Que_t * p, Gia_Obj_t * pObj )
func (p *TasQue) tasQueHasNode(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_QueStore (static inline)
// C: void Tas_QueStore( Tas_Que_t * p, int * piHeadOld, int * piTailOld )
func (p *TasQue) tasQueStore(piHeadOld *int, piTailOld *int) {
	panic("TODO, stubbed")
}

// Tas_QueRestore (static inline)
// C: void Tas_QueRestore( Tas_Que_t * p, int iHeadOld, int iTailOld )
func (p *TasQue) tasQueRestore(iHeadOld int, iTailOld int) {
	panic("TODO, stubbed")
}

// Tas_QueFinish (static inline)
// C: int Tas_QueFinish( Tas_Que_t * p )
func (p *TasQue) tasQueFinish() int {
	panic("TODO, stubbed")
}

// Tas_VarFaninFanoutMax (static inline)
// C: int Tas_VarFaninFanoutMax( Tas_Man_t * p, Gia_Obj_t * pObj )
func (p *TasMan) tasVarFaninFanoutMax(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_ManFindActive (static inline)
// C: Gia_Obj_t * Tas_ManFindActive( Tas_Man_t * p )
func (p *TasMan) tasManFindActive() *GiaObject {
	panic("TODO, stubbed")
}

// Tas_ManDecideHighestFanin (static inline)
// C: Gia_Obj_t * Tas_ManDecideHighestFanin( Tas_Man_t * p )
func (p *TasMan) tasManDecideHighestFanin() *GiaObject {
	panic("TODO, stubbed")
}

// Tas_ManDecideHighest (static inline)
// C: Gia_Obj_t * Tas_ManDecideHighest( Tas_Man_t * p )
func (p *TasMan) tasManDecideHighest() *GiaObject {
	panic("TODO, stubbed")
}

// Tas_ManDecideHighestA (static inline)
// C: Gia_Obj_t * Tas_ManDecideHighestA( Tas_Man_t * p )
func (p *TasMan) tasManDecideHighestA() *GiaObject {
	panic("TODO, stubbed")
}

// Tas_ManDecideLowest (static inline)
// C: Gia_Obj_t * Tas_ManDecideLowest( Tas_Man_t * p )
func (p *TasMan) tasManDecideLowest() *GiaObject {
	panic("TODO, stubbed")
}

// Tas_ManDecideMaxFF (static inline)
// C: Gia_Obj_t * Tas_ManDecideMaxFF( Tas_Man_t * p )
func (p *TasMan) tasManDecideMaxFF() *GiaObject {
	panic("TODO, stubbed")
}

// Tas_ManCancelUntil (static inline)
// C: void Tas_ManCancelUntil( Tas_Man_t * p, int iBound )
func (p *TasMan) tasManCancelUntil(iBound int) {
	panic("TODO, stubbed")
}

// Tas_ManAssign (static inline)
// C: void Tas_ManAssign( Tas_Man_t * p, Gia_Obj_t * pObj, int Level, Gia_Obj_t * pRes0, Gia_Obj_t * pRes1 )
func (p *TasMan) tasManAssign(pObj *GiaObject, Level int, pRes0 *GiaObject, pRes1 *GiaObject) {
	panic("TODO, stubbed")
}

// Tas_ManClauseSize (static inline)
// C: int Tas_ManClauseSize( Tas_Man_t * p, int hClause )
func (p *TasMan) tasManClauseSize(hClause int) int {
	panic("TODO, stubbed")
}

// Tas_ManPrintClause (static inline)
// C: void Tas_ManPrintClause( Tas_Man_t * p, int Level, int hClause )
func (p *TasMan) tasManPrintClause(Level int, hClause int) {
	panic("TODO, stubbed")
}

// Tas_ManPrintClauseNew (static inline)
// C: void Tas_ManPrintClauseNew( Tas_Man_t * p, int Level, int hClause )
func (p *TasMan) tasManPrintClauseNew(Level int, hClause int) {
	panic("TODO, stubbed")
}

// Tas_ManDeriveReason (static inline)
// C: void Tas_ManDeriveReason( Tas_Man_t * p, int Level )
func (p *TasMan) tasManDeriveReason(Level int) {
	panic("TODO, stubbed")
}

// Tas_ManAnalyze (static inline)
// C: int Tas_ManAnalyze( Tas_Man_t * p, int Level, Gia_Obj_t * pVar, Gia_Obj_t * pFan0, Gia_Obj_t * pFan1 )
func (p *TasMan) tasManAnalyze(Level int, pVar *GiaObject, pFan0 *GiaObject, pFan1 *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_ManResolve (static inline)
// C: int Tas_ManResolve( Tas_Man_t * p, int Level, int hClause0, int hClause1 )
func (p *TasMan) tasManResolve(Level int, hClause0 int, hClause1 int) int {
	panic("TODO, stubbed")
}

// Tas_ManAllocCls (static inline)
// C: Tas_Cls_t * Tas_ManAllocCls( Tas_Man_t * p, int nSize )
func (p *TasMan) tasManAllocCls(nSize int) *TasCls {
	panic("TODO, stubbed")
}

// Tas_ManWatchClause (static inline)
// C: void Tas_ManWatchClause( Tas_Man_t * p, Tas_Cls_t * pClause, int Lit )
func (p *TasMan) tasManWatchClause(pClause *TasCls, Lit int) {
	panic("TODO, stubbed")
}

// Tas_ManCreateCls (static inline)
// C: Tas_Cls_t * Tas_ManCreateCls( Tas_Man_t * p, int hClause )
func (p *TasMan) tasManCreateCls(hClause int) *TasCls {
	panic("TODO, stubbed")
}

// Tas_ManCreateFromCls (static inline)
// C: int Tas_ManCreateFromCls( Tas_Man_t * p, Tas_Cls_t * pCls, int Level )
func (p *TasMan) tasManCreateFromCls(pCls *TasCls, Level int) int {
	panic("TODO, stubbed")
}

// Tas_ManPropagateWatch (static inline)
// C: int Tas_ManPropagateWatch( Tas_Man_t * p, int Level, int Lit )
func (p *TasMan) tasManPropagateWatch(Level int, Lit int) int {
	panic("TODO, stubbed")
}

// Tas_ManPropagateOne (static inline)
// C: int Tas_ManPropagateOne( Tas_Man_t * p, Gia_Obj_t * pVar, int Level )
func (p *TasMan) tasManPropagateOne(pVar *GiaObject, Level int) int {
	panic("TODO, stubbed")
}

// Tas_ManPropagateTwo (static inline)
// C: int Tas_ManPropagateTwo( Tas_Man_t * p, Gia_Obj_t * pVar, int Level )
func (p *TasMan) tasManPropagateTwo(pVar *GiaObject, Level int) int {
	panic("TODO, stubbed")
}

// Tas_ManPropagate
// C: int Tas_ManPropagate( Tas_Man_t * p, int Level )
func (p *TasMan) TasManPropagate(Level int) int {
	panic("TODO, stubbed")
}

// Tas_ManSolve_rec
// C: int Tas_ManSolve_rec( Tas_Man_t * p, int Level )
func (p *TasMan) TasManSolve_rec(Level int) int {
	panic("TODO, stubbed")
}

// Tas_ManSolve
// C: int Tas_ManSolve( Tas_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pObj2 )
func (p *TasMan) TasManSolve(pObj *GiaObject, pObj2 *GiaObject) int {
	panic("TODO, stubbed")
}

// Tas_ManSolveArray
// C: int Tas_ManSolveArray( Tas_Man_t * p, Vec_Ptr_t * vObjs )
func (p *TasMan) TasManSolveArray(vObjs *PointerVector[GiaObject]) int {
	panic("TODO, stubbed")
}

// Tas_ManSatPrintStats
// C: void Tas_ManSatPrintStats( Tas_Man_t * p )
func (p *TasMan) TasManSatPrintStats() {
	panic("TODO, stubbed")
}

// Tas_ManSolveMiterNc
// C: Vec_Int_t * Tas_ManSolveMiterNc( Gia_Man_t * pAig, int nConfs, Vec_Str_t ** pvStatus, int fVerbose )
func (pAig *GiaManager) TasManSolveMiterNc(nConfs int, pvStatus **ByteVector, fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Tas_StorePatternTry
// C: int Tas_StorePatternTry( Vec_Ptr_t * vInfo, Vec_Ptr_t * vPres, int iBit, int * pLits, int nLits )
func TasStorePatternTry(vInfo *PointerVector[uint32], vPres *PointerVector[uint32], iBit int, pLits []int, nLits int) int {
	panic("TODO, stubbed")
}

// Tas_StorePattern
// C: int Tas_StorePattern( Vec_Ptr_t * vSimInfo, Vec_Ptr_t * vPres, Vec_Int_t * vCex )
func TasStorePattern(vSimInfo *PointerVector[uint32], vPres *PointerVector[uint32], vCex *IntVector) int {
	panic("TODO, stubbed")
}

// Tas_ManSolveMiterNc2
// C: void Tas_ManSolveMiterNc2( Gia_Man_t * pAig, int nConfs, Gia_Man_t * pAigOld, Vec_Ptr_t * vOldRoots, Vec_Ptr_t * vSimInfo )
func (pAig *GiaManager) TasManSolveMiterNc2(nConfs int, pAigOld *GiaManager, vOldRoots *PointerVector[GiaObject], vSimInfo *PointerVector[uint32]) {
	panic("TODO, stubbed")
}
