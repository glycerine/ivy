package abc

// Hand-checked skeleton of src/map/scl/sclSize.c.

// Abc_SclFindCriticalCo (pfRise is an out parameter)
// C: Abc_Obj_t * Abc_SclFindCriticalCo( SC_Man * p, int * pfRise )
func (p *SCMan) AbcSclFindCriticalCo(pfRise *int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_SclFindMostCriticalFanin2 (pfRise is an in/out parameter)
// C: Abc_Obj_t * Abc_SclFindMostCriticalFanin2( SC_Man * p, int * pfRise, Abc_Obj_t * pNode )
func (p *SCMan) AbcSclFindMostCriticalFanin2(pfRise *int, pNode *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_SclFindMostCriticalFanin (pfRise is an in/out parameter)
// C: Abc_Obj_t * Abc_SclFindMostCriticalFanin( SC_Man * p, int * pfRise, Abc_Obj_t * pNode )
func (p *SCMan) AbcSclFindMostCriticalFanin(pfRise *int, pNode *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_SclTimeNodePrint (static)
// C: void Abc_SclTimeNodePrint( SC_Man * p, Abc_Obj_t * pObj, int fRise, int Length, float maxDelay )
func (p *SCMan) abcSclTimeNodePrint(pObj *ABCNetworkObject, fRise int, Length int, maxDelay float32) {
	panic("TODO, stubbed")
}

// Abc_SclTimeNtkPrint
// C: void Abc_SclTimeNtkPrint( SC_Man * p, int fShowAll, int fPrintPath )
func (p *SCMan) AbcSclTimeNtkPrint(fShowAll int, fPrintPath int) {
	panic("TODO, stubbed")
}

// Abc_SclTimeFanin (static)
// C: void Abc_SclTimeFanin( SC_Man * p, SC_Timing * pTime, Abc_Obj_t * pObj, Abc_Obj_t * pFanin )
func (p *SCMan) abcSclTimeFanin(pTime *SCTiming, pObj *ABCNetworkObject, pFanin *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclDeptFanin (static)
// C: void Abc_SclDeptFanin( SC_Man * p, SC_Timing * pTime, Abc_Obj_t * pObj, Abc_Obj_t * pFanin )
func (p *SCMan) abcSclDeptFanin(pTime *SCTiming, pObj *ABCNetworkObject, pFanin *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclDeptObj (static)
// C: void Abc_SclDeptObj( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclDeptObj(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclObjLoadValue (static)
// C: float Abc_SclObjLoadValue( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclObjLoadValue(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_SclTimeCi (static)
// C: void Abc_SclTimeCi( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclTimeCi(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclTimeNode
// C: void Abc_SclTimeNode( SC_Man * p, Abc_Obj_t * pObj, int fDept )
func (p *SCMan) AbcSclTimeNode(pObj *ABCNetworkObject, fDept int) {
	panic("TODO, stubbed")
}

// Abc_SclTimeCone
// C: void Abc_SclTimeCone( SC_Man * p, Vec_Int_t * vCone )
func (p *SCMan) AbcSclTimeCone(vCone *IntVector) {
	panic("TODO, stubbed")
}

// Abc_SclTimeNtkRecompute (pArea and pDelay are out parameters (nullable))
// C: void Abc_SclTimeNtkRecompute( SC_Man * p, float * pArea, float * pDelay, int fReverse, float DUser )
func (p *SCMan) AbcSclTimeNtkRecompute(pArea *float32, pDelay *float32, fReverse int, DUser float32) {
	panic("TODO, stubbed")
}

// Abc_SclTimeIncUpdateClean (static)
// C: void Abc_SclTimeIncUpdateClean( SC_Man * p )
func (p *SCMan) abcSclTimeIncUpdateClean() {
	panic("TODO, stubbed")
}

// Abc_SclTimeIncAddNode (static)
// C: void Abc_SclTimeIncAddNode( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclTimeIncAddNode(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclTimeIncAddFanins (static)
// C: void Abc_SclTimeIncAddFanins( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclTimeIncAddFanins(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclTimeIncAddFanouts (static)
// C: void Abc_SclTimeIncAddFanouts( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclTimeIncAddFanouts(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclTimeIncUpdateArrival (static)
// C: void Abc_SclTimeIncUpdateArrival( SC_Man * p )
func (p *SCMan) abcSclTimeIncUpdateArrival() {
	panic("TODO, stubbed")
}

// Abc_SclTimeIncUpdateDeparture (static)
// C: void Abc_SclTimeIncUpdateDeparture( SC_Man * p )
func (p *SCMan) abcSclTimeIncUpdateDeparture() {
	panic("TODO, stubbed")
}

// Abc_SclTimeIncCheckLevel
// C: void Abc_SclTimeIncCheckLevel( Abc_Ntk_t * pNtk )
func AbcSclTimeIncCheckLevel(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_SclTimeIncUpdate
// C: int Abc_SclTimeIncUpdate( SC_Man * p )
func (p *SCMan) AbcSclTimeIncUpdate() int {
	panic("TODO, stubbed")
}

// Abc_SclTimeIncInsert
// C: void Abc_SclTimeIncInsert( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) AbcSclTimeIncInsert(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclTimeIncUpdateLevel_rec
// C: void Abc_SclTimeIncUpdateLevel_rec( Abc_Obj_t * pObj )
func AbcSclTimeIncUpdateLevel_rec(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclTimeIncUpdateLevel
// C: void Abc_SclTimeIncUpdateLevel( Abc_Obj_t * pObj )
func AbcSclTimeIncUpdateLevel(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclManReadSlewAndLoad
// C: void Abc_SclManReadSlewAndLoad( SC_Man * p, Abc_Ntk_t * pNtk )
func (p *SCMan) AbcSclManReadSlewAndLoad(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_SclManStart
// C: SC_Man * Abc_SclManStart( SC_Lib * pLib, Abc_Ntk_t * pNtk, int fUseWireLoads, int fDept, float DUser, int nTreeCRatio )
func (pLib *SCLib) AbcSclManStart(pNtk *ABCNetwork, fUseWireLoads int, fDept int, DUser float32, nTreeCRatio int) *SCMan {
	panic("TODO, stubbed")
}

// Abc_SclTimePerformInt
// C: void Abc_SclTimePerformInt( SC_Lib * pLib, Abc_Ntk_t * pNtk, int nTreeCRatio, int fUseWireLoads, int fShowAll, int fPrintPath, int fDumpStats )
func (pLib *SCLib) AbcSclTimePerformInt(pNtk *ABCNetwork, nTreeCRatio int, fUseWireLoads int, fShowAll int, fPrintPath int, fDumpStats int) {
	panic("TODO, stubbed")
}

// Abc_SclTimePerform
// C: void Abc_SclTimePerform( SC_Lib * pLib, Abc_Ntk_t * pNtk, int nTreeCRatio, int fUseWireLoads, int fShowAll, int fPrintPath, int fDumpStats )
func (pLib *SCLib) AbcSclTimePerform(pNtk *ABCNetwork, nTreeCRatio int, fUseWireLoads int, fShowAll int, fPrintPath int, fDumpStats int) {
	panic("TODO, stubbed")
}

// Abc_SclCheckCommonInputs
// C: int Abc_SclCheckCommonInputs( Abc_Obj_t * pObj, Abc_Obj_t * pFanin )
func AbcSclCheckCommonInputs(pObj *ABCNetworkObject, pFanin *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclPrintFaninPairs
// C: void Abc_SclPrintFaninPairs( SC_Man * p, Abc_Ntk_t * pNtk )
func (p *SCMan) AbcSclPrintFaninPairs(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_ObjIsBuffer (static)
// C: int Abc_ObjIsBuffer( Abc_Obj_t * pObj )
func abcObjIsBuffer(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclHasBufferFanout
// C: int Abc_SclHasBufferFanout( Abc_Obj_t * pObj )
func AbcSclHasBufferFanout(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclCountBufferFanoutsInt
// C: int Abc_SclCountBufferFanoutsInt( Abc_Obj_t * pObj )
func AbcSclCountBufferFanoutsInt(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclCountBufferFanouts
// C: int Abc_SclCountBufferFanouts( Abc_Obj_t * pObj )
func AbcSclCountBufferFanouts(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclCountNonBufferFanoutsInt
// C: int Abc_SclCountNonBufferFanoutsInt( Abc_Obj_t * pObj )
func AbcSclCountNonBufferFanoutsInt(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclCountNonBufferFanouts
// C: int Abc_SclCountNonBufferFanouts( Abc_Obj_t * pObj )
func AbcSclCountNonBufferFanouts(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclCountNonBufferDelayInt
// C: float Abc_SclCountNonBufferDelayInt( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) AbcSclCountNonBufferDelayInt(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_SclCountNonBufferDelay
// C: float Abc_SclCountNonBufferDelay( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) AbcSclCountNonBufferDelay(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_SclCountNonBufferLoadInt
// C: float Abc_SclCountNonBufferLoadInt( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) AbcSclCountNonBufferLoadInt(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_SclCountNonBufferLoad
// C: float Abc_SclCountNonBufferLoad( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) AbcSclCountNonBufferLoad(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_SclPrintBuffersOne
// C: void Abc_SclPrintBuffersOne( SC_Man * p, Abc_Obj_t * pObj, int nOffset )
func (p *SCMan) AbcSclPrintBuffersOne(pObj *ABCNetworkObject, nOffset int) {
	panic("TODO, stubbed")
}

// Abc_SclPrintBuffersInt
// C: void Abc_SclPrintBuffersInt( SC_Man * p, Abc_Obj_t * pObj, int nOffset )
func (p *SCMan) AbcSclPrintBuffersInt(pObj *ABCNetworkObject, nOffset int) {
	panic("TODO, stubbed")
}

// Abc_SclPrintBufferTrees
// C: void Abc_SclPrintBufferTrees( SC_Man * p, Abc_Ntk_t * pNtk )
func (p *SCMan) AbcSclPrintBufferTrees(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_SclPrintBuffers
// C: void Abc_SclPrintBuffers( SC_Lib * pLib, Abc_Ntk_t * pNtk, int fVerbose )
func (pLib *SCLib) AbcSclPrintBuffers(pNtk *ABCNetwork, fVerbose int) {
	panic("TODO, stubbed")
}
