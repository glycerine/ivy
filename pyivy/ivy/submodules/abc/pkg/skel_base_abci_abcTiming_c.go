package abc

// Hand-checked skeleton of src/base/abci/abcTiming.c.

// AbcManTime is Abc_ManTime_t.
type AbcManTime struct {
	tArrDef     AbcTime                 // Abc_Time_t
	tReqDef     AbcTime                 // Abc_Time_t
	vArrs       *PointerVector[AbcTime] // Vec_Ptr_t * of Abc_Time_t *
	vReqs       *PointerVector[AbcTime] // Vec_Ptr_t * of Abc_Time_t *
	tInDriveDef AbcTime                 // Abc_Time_t
	tOutLoadDef AbcTime                 // Abc_Time_t
	tInDrive    []AbcTime               // Abc_Time_t * : array
	tOutLoad    []AbcTime               // Abc_Time_t * : array
}

// Abc_FloatEqual (static)
// C: int Abc_FloatEqual( float x, float y )
func abcFloatEqual(x float32, y float32) int {
	panic("TODO, stubbed")
}

// Abc_NodeArrival (static)
// C: Abc_Time_t * Abc_NodeArrival( Abc_Obj_t * pNode )
func abcNodeArrival(pNode *ABCNetworkObject) *AbcTime {
	panic("TODO, stubbed")
}

// Abc_NodeRequired (static)
// C: Abc_Time_t * Abc_NodeRequired( Abc_Obj_t * pNode )
func abcNodeRequired(pNode *ABCNetworkObject) *AbcTime {
	panic("TODO, stubbed")
}

// Abc_NtkReadDefaultArrival
// C: Abc_Time_t * Abc_NtkReadDefaultArrival( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkReadDefaultArrival() *AbcTime {
	panic("TODO, stubbed")
}

// Abc_NtkReadDefaultRequired
// C: Abc_Time_t * Abc_NtkReadDefaultRequired( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkReadDefaultRequired() *AbcTime {
	panic("TODO, stubbed")
}

// Abc_NodeReadArrival
// C: Abc_Time_t * Abc_NodeReadArrival( Abc_Obj_t * pNode )
func AbcNodeReadArrival(pNode *ABCNetworkObject) *AbcTime {
	panic("TODO, stubbed")
}

// Abc_NodeReadRequired
// C: Abc_Time_t * Abc_NodeReadRequired( Abc_Obj_t * pNode )
func AbcNodeReadRequired(pNode *ABCNetworkObject) *AbcTime {
	panic("TODO, stubbed")
}

// Abc_NtkReadDefaultArrivalWorst
// C: float Abc_NtkReadDefaultArrivalWorst( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkReadDefaultArrivalWorst() float32 {
	panic("TODO, stubbed")
}

// Abc_NtkReadDefaultRequiredWorst
// C: float Abc_NtkReadDefaultRequiredWorst( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkReadDefaultRequiredWorst() float32 {
	panic("TODO, stubbed")
}

// Abc_NodeReadArrivalAve
// C: float Abc_NodeReadArrivalAve( Abc_Obj_t * pNode )
func AbcNodeReadArrivalAve(pNode *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_NodeReadRequiredAve
// C: float Abc_NodeReadRequiredAve( Abc_Obj_t * pNode )
func AbcNodeReadRequiredAve(pNode *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_NodeReadArrivalWorst
// C: float Abc_NodeReadArrivalWorst( Abc_Obj_t * pNode )
func AbcNodeReadArrivalWorst(pNode *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_NodeReadRequiredWorst
// C: float Abc_NodeReadRequiredWorst( Abc_Obj_t * pNode )
func AbcNodeReadRequiredWorst(pNode *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_NtkReadDefaultInputDrive
// C: Abc_Time_t * Abc_NtkReadDefaultInputDrive( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkReadDefaultInputDrive() *AbcTime {
	panic("TODO, stubbed")
}

// Abc_NtkReadDefaultOutputLoad
// C: Abc_Time_t * Abc_NtkReadDefaultOutputLoad( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkReadDefaultOutputLoad() *AbcTime {
	panic("TODO, stubbed")
}

// Abc_NodeReadInputDrive
// C: Abc_Time_t * Abc_NodeReadInputDrive( Abc_Ntk_t * pNtk, int iPi )
func (pNtk *ABCNetwork) AbcNodeReadInputDrive(iPi int) *AbcTime {
	panic("TODO, stubbed")
}

// Abc_NodeReadOutputLoad
// C: Abc_Time_t * Abc_NodeReadOutputLoad( Abc_Ntk_t * pNtk, int iPo )
func (pNtk *ABCNetwork) AbcNodeReadOutputLoad(iPo int) *AbcTime {
	panic("TODO, stubbed")
}

// Abc_NodeReadInputDriveWorst
// C: float Abc_NodeReadInputDriveWorst( Abc_Ntk_t * pNtk, int iPi )
func (pNtk *ABCNetwork) AbcNodeReadInputDriveWorst(iPi int) float32 {
	panic("TODO, stubbed")
}

// Abc_NodeReadOutputLoadWorst
// C: float Abc_NodeReadOutputLoadWorst( Abc_Ntk_t * pNtk, int iPo )
func (pNtk *ABCNetwork) AbcNodeReadOutputLoadWorst(iPo int) float32 {
	panic("TODO, stubbed")
}

// Abc_NtkTimeSetDefaultArrival
// C: void Abc_NtkTimeSetDefaultArrival( Abc_Ntk_t * pNtk, float Rise, float Fall )
func (pNtk *ABCNetwork) AbcNtkTimeSetDefaultArrival(Rise float32, Fall float32) {
	panic("TODO, stubbed")
}

// Abc_NtkTimeSetDefaultRequired
// C: void Abc_NtkTimeSetDefaultRequired( Abc_Ntk_t * pNtk, float Rise, float Fall )
func (pNtk *ABCNetwork) AbcNtkTimeSetDefaultRequired(Rise float32, Fall float32) {
	panic("TODO, stubbed")
}

// Abc_NtkTimeSetArrival
// C: void Abc_NtkTimeSetArrival( Abc_Ntk_t * pNtk, int ObjId, float Rise, float Fall )
func (pNtk *ABCNetwork) AbcNtkTimeSetArrival(ObjId int, Rise float32, Fall float32) {
	panic("TODO, stubbed")
}

// Abc_NtkTimeSetRequired
// C: void Abc_NtkTimeSetRequired( Abc_Ntk_t * pNtk, int ObjId, float Rise, float Fall )
func (pNtk *ABCNetwork) AbcNtkTimeSetRequired(ObjId int, Rise float32, Fall float32) {
	panic("TODO, stubbed")
}

// Abc_NtkTimeSetDefaultInputDrive
// C: void Abc_NtkTimeSetDefaultInputDrive( Abc_Ntk_t * pNtk, float Rise, float Fall )
func (pNtk *ABCNetwork) AbcNtkTimeSetDefaultInputDrive(Rise float32, Fall float32) {
	panic("TODO, stubbed")
}

// Abc_NtkTimeSetDefaultOutputLoad
// C: void Abc_NtkTimeSetDefaultOutputLoad( Abc_Ntk_t * pNtk, float Rise, float Fall )
func (pNtk *ABCNetwork) AbcNtkTimeSetDefaultOutputLoad(Rise float32, Fall float32) {
	panic("TODO, stubbed")
}

// Abc_NtkTimeSetInputDrive
// C: void Abc_NtkTimeSetInputDrive( Abc_Ntk_t * pNtk, int PiNum, float Rise, float Fall )
func (pNtk *ABCNetwork) AbcNtkTimeSetInputDrive(PiNum int, Rise float32, Fall float32) {
	panic("TODO, stubbed")
}

// Abc_NtkTimeSetOutputLoad
// C: void Abc_NtkTimeSetOutputLoad( Abc_Ntk_t * pNtk, int PoNum, float Rise, float Fall )
func (pNtk *ABCNetwork) AbcNtkTimeSetOutputLoad(PoNum int, Rise float32, Fall float32) {
	panic("TODO, stubbed")
}

// Abc_NtkTimeInitialize
// C: void Abc_NtkTimeInitialize( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkOld )
func (pNtk *ABCNetwork) AbcNtkTimeInitialize(pNtkOld *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NtkTimeScale
// C: void Abc_NtkTimeScale( Abc_Ntk_t * pNtk, float Scale )
func (pNtk *ABCNetwork) AbcNtkTimeScale(Scale float32) {
	panic("TODO, stubbed")
}

// Abc_NtkTimePrepare
// C: void Abc_NtkTimePrepare( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkTimePrepare() {
	panic("TODO, stubbed")
}

// Abc_ManTimeStart (static)
// C: Abc_ManTime_t * Abc_ManTimeStart( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcManTimeStart() *AbcManTime {
	panic("TODO, stubbed")
}

// Abc_ManTimeStop
// C: void Abc_ManTimeStop( Abc_ManTime_t * p )
func (p *AbcManTime) AbcManTimeStop() {
	panic("TODO, stubbed")
}

// Abc_ManTimeDup
// C: void Abc_ManTimeDup( Abc_Ntk_t * pNtkOld, Abc_Ntk_t * pNtkNew )
func (pNtkOld *ABCNetwork) AbcManTimeDup(pNtkNew *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NtkTimePrint
// C: void Abc_NtkTimePrint( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkTimePrint() {
	panic("TODO, stubbed")
}

// Abc_ManTimeExpand (static)
// C: void Abc_ManTimeExpand( Abc_ManTime_t * p, int nSize, int fProgressive )
func (p *AbcManTime) abcManTimeExpand(nSize int, fProgressive int) {
	panic("TODO, stubbed")
}

// Abc_NtkSetNodeLevelsArrival
// C: void Abc_NtkSetNodeLevelsArrival( Abc_Ntk_t * pNtkOld )
func (pNtkOld *ABCNetwork) AbcNtkSetNodeLevelsArrival() {
	panic("TODO, stubbed")
}

// Abc_NtkGetCiArrivalTimes
// C: Abc_Time_t * Abc_NtkGetCiArrivalTimes( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkGetCiArrivalTimes() []AbcTime {
	panic("TODO, stubbed")
}

// Abc_NtkGetCoRequiredTimes
// C: Abc_Time_t * Abc_NtkGetCoRequiredTimes( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkGetCoRequiredTimes() []AbcTime {
	panic("TODO, stubbed")
}

// Abc_NtkGetCiArrivalFloats
// C: float * Abc_NtkGetCiArrivalFloats( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkGetCiArrivalFloats() []float32 {
	panic("TODO, stubbed")
}

// Abc_NtkGetCoRequiredFloats
// C: float * Abc_NtkGetCoRequiredFloats( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkGetCoRequiredFloats() []float32 {
	panic("TODO, stubbed")
}

// Abc_NtkDelayTraceSlackStart
// C: Vec_Int_t * Abc_NtkDelayTraceSlackStart( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkDelayTraceSlackStart() *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkDelayTraceSlack (static)
// C: float Abc_NtkDelayTraceSlack( Vec_Int_t * vSlacks, Abc_Obj_t * pObj, int iFanin )
func abcNtkDelayTraceSlack(vSlacks *IntVector, pObj *ABCNetworkObject, iFanin int) float32 {
	panic("TODO, stubbed")
}

// Abc_NtkDelayTraceSetSlack (static)
// C: void Abc_NtkDelayTraceSetSlack( Vec_Int_t * vSlacks, Abc_Obj_t * pObj, int iFanin, float Num )
func abcNtkDelayTraceSetSlack(vSlacks *IntVector, pObj *ABCNetworkObject, iFanin int, Num float32) {
	panic("TODO, stubbed")
}

// Abc_NtkDelayTraceCritPath_rec
// C: int Abc_NtkDelayTraceCritPath_rec( Vec_Int_t * vSlacks, Abc_Obj_t * pNode, Abc_Obj_t * pLeaf, Vec_Int_t * vBest )
func AbcNtkDelayTraceCritPath_rec(vSlacks *IntVector, pNode *ABCNetworkObject, pLeaf *ABCNetworkObject, vBest *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkDelayTraceCritPathCollect_rec
// C: void Abc_NtkDelayTraceCritPathCollect_rec( Vec_Int_t * vSlacks, Abc_Obj_t * pNode, Vec_Int_t * vBest, Vec_Ptr_t * vPath )
func AbcNtkDelayTraceCritPathCollect_rec(vSlacks *IntVector, pNode *ABCNetworkObject, vBest *IntVector, vPath *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NodeDelayTraceArrival
// C: void Abc_NodeDelayTraceArrival( Abc_Obj_t * pNode, Vec_Int_t * vSlacks )
func AbcNodeDelayTraceArrival(pNode *ABCNetworkObject, vSlacks *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkDelayTrace
// C: float Abc_NtkDelayTrace( Abc_Ntk_t * pNtk, Abc_Obj_t * pOut, Abc_Obj_t * pIn, int fPrint )
func (pNtk *ABCNetwork) AbcNtkDelayTrace(pOut *ABCNetworkObject, pIn *ABCNetworkObject, fPrint int) float32 {
	panic("TODO, stubbed")
}

// Abc_ObjLevelNew
// C: int Abc_ObjLevelNew( Abc_Obj_t * pObj )
func AbcObjLevelNew(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_ObjReverseLevelNew
// C: int Abc_ObjReverseLevelNew( Abc_Obj_t * pObj )
func AbcObjReverseLevelNew(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_ObjRequiredLevel
// C: int Abc_ObjRequiredLevel( Abc_Obj_t * pObj )
func AbcObjRequiredLevel(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_ObjReverseLevel
// C: int Abc_ObjReverseLevel( Abc_Obj_t * pObj )
func AbcObjReverseLevel(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_ObjSetReverseLevel
// C: void Abc_ObjSetReverseLevel( Abc_Obj_t * pObj, int LevelR )
func AbcObjSetReverseLevel(pObj *ABCNetworkObject, LevelR int) {
	panic("TODO, stubbed")
}

// Abc_NtkStartReverseLevels
// C: void Abc_NtkStartReverseLevels( Abc_Ntk_t * pNtk, int nMaxLevelIncrease )
func (pNtk *ABCNetwork) AbcNtkStartReverseLevels(nMaxLevelIncrease int) {
	panic("TODO, stubbed")
}

// Abc_NtkStopReverseLevels
// C: void Abc_NtkStopReverseLevels( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkStopReverseLevels() {
	panic("TODO, stubbed")
}

// Abc_NtkUpdateLevel
// C: void Abc_NtkUpdateLevel( Abc_Obj_t * pObjNew, Vec_Vec_t * vLevels )
func AbcNtkUpdateLevel(pObjNew *ABCNetworkObject, vLevels *NestedVector[*ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkUpdateReverseLevel
// C: void Abc_NtkUpdateReverseLevel( Abc_Obj_t * pObjNew, Vec_Vec_t * vLevels )
func AbcNtkUpdateReverseLevel(pObjNew *ABCNetworkObject, vLevels *NestedVector[*ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkUpdate
// C: void Abc_NtkUpdate( Abc_Obj_t * pObj, Abc_Obj_t * pObjNew, Vec_Vec_t * vLevels )
func AbcNtkUpdate(pObj *ABCNetworkObject, pObjNew *ABCNetworkObject, vLevels *NestedVector[*ABCNetworkObject]) {
	panic("TODO, stubbed")
}
