package abc

// Hand-checked skeleton of src/base/wlc/wlcMem.c.

// Wlc_NtkMemBlast_rec
// C: void Wlc_NtkMemBlast_rec( Wlc_Ntk_t * pNew, Wlc_Ntk_t * p, int iObj, Vec_Int_t * vFanins )
func WlcNtkMemBlast_rec(pNew *WLCNetwork, p *WLCNetwork, iObj int, vFanins *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkMemBlast
// C: Wlc_Ntk_t * Wlc_NtkMemBlast( Wlc_Ntk_t * p )
func WlcNtkMemBlast(p *WLCNetwork) *WLCNetwork {
	panic("TODO, stubbed")
}

// Wlc_NtkCollectMemSizes
// C: Vec_Int_t * Wlc_NtkCollectMemSizes( Wlc_Ntk_t * p )
func WlcNtkCollectMemSizes(p *WLCNetwork) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_ObjCheckIsEmpty_rec
// C: int Wlc_ObjCheckIsEmpty_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pObj )
func WlcObjCheckIsEmpty_rec(p *WLCNetwork, pObj *WLCObject) int {
	panic("TODO, stubbed")
}

// Wlc_NtkCleanObjects
// C: Vec_Int_t * Wlc_NtkCleanObjects( Wlc_Ntk_t * p, Vec_Int_t * vObjs )
func WlcNtkCleanObjects(p *WLCNetwork, vObjs *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkCollectMemory
// C: Vec_Int_t * Wlc_NtkCollectMemory( Wlc_Ntk_t * p, int fClean )
func WlcNtkCollectMemory(p *WLCNetwork, fClean int) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkPrintMemory
// C: void Wlc_NtkPrintMemory( Wlc_Ntk_t * p )
func WlcNtkPrintMemory(p *WLCNetwork) {
	panic("TODO, stubbed")
}

// Wlc_NtkCollectMemFanins
// C: Vec_Int_t * Wlc_NtkCollectMemFanins( Wlc_Ntk_t * p, Vec_Int_t * vMemObjs )
func WlcNtkCollectMemFanins(p *WLCNetwork, vMemObjs *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_CountDcs
// C: int Wlc_CountDcs( char * pInit )
func WlcCountDcs(pInit string) int {
	panic("TODO, stubbed")
}

// Wlc_NtkDupOneObject
// C: int Wlc_NtkDupOneObject( Wlc_Ntk_t * pNew, Wlc_Ntk_t * p, Wlc_Obj_t * pObj, int TypeNew, Vec_Int_t * vFanins )
func WlcNtkDupOneObject(pNew *WLCNetwork, p *WLCNetwork, pObj *WLCObject, TypeNew int, vFanins *IntVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkDupOneBuffer
// C: void Wlc_NtkDupOneBuffer( Wlc_Ntk_t * pNew, Wlc_Ntk_t * p, Wlc_Obj_t * pObj, int iFanin, Vec_Int_t * vFanins, int fIsFi )
func WlcNtkDupOneBuffer(pNew *WLCNetwork, p *WLCNetwork, pObj *WLCObject, iFanin int, vFanins *IntVector, fIsFi int) {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsAddToNodeFrames
// C: void Wlc_NtkAbsAddToNodeFrames( Vec_Int_t * vNodeFrames, Vec_Int_t * vLevel )
func WlcNtkAbsAddToNodeFrames(vNodeFrames *IntVector, vLevel *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsCreateFlopOutputs
// C: Vec_Int_t * Wlc_NtkAbsCreateFlopOutputs( Wlc_Ntk_t * pNew, Wlc_Ntk_t * p, Vec_Int_t * vNodeFrames, Vec_Int_t * vFanins )
func WlcNtkAbsCreateFlopOutputs(pNew *WLCNetwork, p *WLCNetwork, vNodeFrames *IntVector, vFanins *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsCreateFlopInputs
// C: void Wlc_NtkAbsCreateFlopInputs( Wlc_Ntk_t * pNew, Wlc_Ntk_t * p, Vec_Int_t * vNodeFrames, Vec_Int_t * vFanins, Vec_Int_t * vNewObjs, Wlc_Obj_t * pCounter, int AdderBits )
func WlcNtkAbsCreateFlopInputs(pNew *WLCNetwork, p *WLCNetwork, vNodeFrames *IntVector, vFanins *IntVector, vNewObjs *IntVector, pCounter *WLCObject, AdderBits int) {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsCreateLogic
// C: void Wlc_NtkAbsCreateLogic( Wlc_Ntk_t * pNew, Wlc_Ntk_t * p, Vec_Int_t * vNodeFrames, Vec_Int_t * vFanins, Vec_Int_t * vNewObjs, Vec_Wec_t * vConstrs, Wlc_Obj_t * pConstrOut )
func WlcNtkAbsCreateLogic(pNew *WLCNetwork, p *WLCNetwork, vNodeFrames *IntVector, vFanins *IntVector, vNewObjs *IntVector, vConstrs *IntLevelVector, pConstrOut *WLCObject) {
	panic("TODO, stubbed")
}

// Wlc_NtkAbstractMemory (piFirstMemPi, piFirstCi and piFirstMemCi are out parameters)
// C: Wlc_Ntk_t * Wlc_NtkAbstractMemory( Wlc_Ntk_t * p, Vec_Int_t * vMemObjs, Vec_Int_t * vMemFanins, int * piFirstMemPi, int * piFirstCi, int * piFirstMemCi, Vec_Wec_t * vConstrs, Vec_Int_t * vNodeFrames )
func WlcNtkAbstractMemory(p *WLCNetwork, vMemObjs *IntVector, vMemFanins *IntVector, piFirstMemPi *int, piFirstCi *int, piFirstMemCi *int, vConstrs *IntLevelVector, vNodeFrames *IntVector) *WLCNetwork {
	panic("TODO, stubbed")
}

// Wlc_NtkDeriveFirstTotal
// C: Vec_Int_t * Wlc_NtkDeriveFirstTotal( Wlc_Ntk_t * p, Vec_Int_t * vMemObjs, Vec_Int_t * vMemFanins, int iFirstMemPi, int iFirstMemCi, int fVerbose )
func WlcNtkDeriveFirstTotal(p *WLCNetwork, vMemObjs *IntVector, vMemFanins *IntVector, iFirstMemPi int, iFirstMemCi int, fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkCexResim
// C: int Wlc_NtkCexResim( Gia_Man_t * pAbs, Abc_Cex_t * p, Vec_Int_t * vFirstTotal, int iBit, Vec_Wrd_t * vRes, int iFrame )
func WlcNtkCexResim(pAbs *GiaManager, p *Counterexample, vFirstTotal *IntVector, iBit int, vRes *WordVector, iFrame int) int {
	panic("TODO, stubbed")
}

// Wlc_NtkConvertCex
// C: Vec_Wrd_t * Wlc_NtkConvertCex( Vec_Int_t * vFirstTotal, Gia_Man_t * pAbs, Abc_Cex_t * pCex, int fVerbose )
func WlcNtkConvertCex(vFirstTotal *IntVector, pAbs *GiaManager, pCex *Counterexample, fVerbose int) *WordVector {
	panic("TODO, stubbed")
}

// Wlc_NtkTrace_rec
// C: void Wlc_NtkTrace_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, int iFrame, Vec_Int_t * vMemObjs, Vec_Wrd_t * vValues, word ValueA, Vec_Int_t * vRes )
func WlcNtkTrace_rec(p *WLCNetwork, pObj *WLCObject, iFrame int, vMemObjs *IntVector, vValues *WordVector, ValueA uint64, vRes *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkTrace
// C: Vec_Int_t * Wlc_NtkTrace( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, int iFrame, Vec_Int_t * vMemObjs, Vec_Wrd_t * vValues )
func WlcNtkTrace(p *WLCNetwork, pObj *WLCObject, iFrame int, vMemObjs *IntVector, vValues *WordVector) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkTraceCheckConfict
// C: int Wlc_NtkTraceCheckConfict( Wlc_Ntk_t * p, Vec_Int_t * vTrace, Vec_Int_t * vMemObjs, Vec_Wrd_t * vValues )
func WlcNtkTraceCheckConfict(p *WLCNetwork, vTrace *IntVector, vMemObjs *IntVector, vValues *WordVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkFindConflict
// C: Vec_Int_t * Wlc_NtkFindConflict( Wlc_Ntk_t * p, Vec_Int_t * vMemObjs, Vec_Wrd_t * vValues, int nFrames )
func WlcNtkFindConflict(p *WLCNetwork, vMemObjs *IntVector, vValues *WordVector, nFrames int) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkPrintConflict
// C: void Wlc_NtkPrintConflict( Wlc_Ntk_t * p, Vec_Int_t * vTrace )
func WlcNtkPrintConflict(p *WLCNetwork, vTrace *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkPrintCex
// C: void Wlc_NtkPrintCex( Wlc_Ntk_t * p, Wlc_Ntk_t * pAbs, Abc_Cex_t * pCex )
func WlcNtkPrintCex(p *WLCNetwork, pAbs *WLCNetwork, pCex *Counterexample) {
	panic("TODO, stubbed")
}

// Wlc_NtkMemAbstractTest
// C: Wlc_Ntk_t * Wlc_NtkMemAbstractTest( Wlc_Ntk_t * p )
func WlcNtkMemAbstractTest(p *WLCNetwork) *WLCNetwork {
	panic("TODO, stubbed")
}

// Wlc_NtkMemAbstract
// C: int Wlc_NtkMemAbstract( Wlc_Ntk_t * p, int nIterMax, int fDumpAbs, int fPdrVerbose, int fVerbose )
func WlcNtkMemAbstract(p *WLCNetwork, nIterMax int, fDumpAbs int, fPdrVerbose int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Wlc_NtkExploreMem2_rec
// C: int Wlc_NtkExploreMem2_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, Vec_Int_t * vCollect, int nFrames )
func WlcNtkExploreMem2_rec(p *WLCNetwork, pObj *WLCObject, vCollect *IntVector, nFrames int) int {
	panic("TODO, stubbed")
}

// Wlc_NtkExploreMem2
// C: void Wlc_NtkExploreMem2( Wlc_Ntk_t * p, int nFrames )
func WlcNtkExploreMem2(p *WLCNetwork, nFrames int) {
	panic("TODO, stubbed")
}

// Wlc_NtkExploreMem_rec
// C: void Wlc_NtkExploreMem_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, Vec_Int_t * vCollect, int nFrames )
func WlcNtkExploreMem_rec(p *WLCNetwork, pObj *WLCObject, vCollect *IntVector, nFrames int) {
	panic("TODO, stubbed")
}

// Wlc_NtkExploreMem
// C: void Wlc_NtkExploreMem( Wlc_Ntk_t * p, int nFrames )
func WlcNtkExploreMem(p *WLCNetwork, nFrames int) {
	panic("TODO, stubbed")
}

// Wlc_NtkFindReachablePiFo
// C: Vec_Int_t * Wlc_NtkFindReachablePiFo( Wlc_Ntk_t * p, Vec_Int_t * vMemObjsClean, int nFrames )
func WlcNtkFindReachablePiFo(p *WLCNetwork, vMemObjsClean *IntVector, nFrames int) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkExtractCisForThisRead
// C: Vec_Int_t * Wlc_NtkExtractCisForThisRead( Vec_Int_t * vReachReadCi, int iRead )
func WlcNtkExtractCisForThisRead(vReachReadCi *IntVector, iRead int) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkCollectOneType
// C: Vec_Int_t * Wlc_NtkCollectOneType( Wlc_Ntk_t * p, Vec_Int_t * vMemObjsClean, int Type1, int Type2 )
func WlcNtkCollectOneType(p *WLCNetwork, vMemObjsClean *IntVector, Type1 int, Type2 int) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkCreateMemoryConstr
// C: void Wlc_NtkCreateMemoryConstr( Wlc_Ntk_t * pNew, Wlc_Ntk_t * p, Vec_Int_t * vMemObjsClean, Vec_Int_t * vReachReadCi )
func WlcNtkCreateMemoryConstr(pNew *WLCNetwork, p *WLCNetwork, vMemObjsClean *IntVector, vReachReadCi *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkAbstractMem
// C: Wlc_Ntk_t * Wlc_NtkAbstractMem( Wlc_Ntk_t * p, int nFrames, int fVerbose )
func WlcNtkAbstractMem(p *WLCNetwork, nFrames int, fVerbose int) *WLCNetwork {
	panic("TODO, stubbed")
}
