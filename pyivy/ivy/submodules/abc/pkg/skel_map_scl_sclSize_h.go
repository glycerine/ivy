package abc

// Hand-checked skeleton of src/map/scl/sclSize.h.

// SCMan is SC_Man.
type SCMan struct {
	pLib  *SCLib      // library
	pNtk  *ABCNetwork // network
	nObjs int         // allocated size
	// get assignment
	vGatesBest *IntVector // best gate sizes found so far
	vUpdates   *IntVector // sizing updates in this round
	vUpdates2  *IntVector // sizing updates in this round
	// timing information
	pWLoadUsed *SCWireLoad         // name of the used WireLoad model
	vWireCaps  *FloatVector        // wire capacitances
	pLoads     []SCPair            // SC_Pair * : loads for each gate
	pDepts     []SCPair            // SC_Pair * : departures for each gate
	pTimes     []SCPair            // SC_Pair * : arrivals for each gate
	pSlews     []SCPair            // SC_Pair * : slews for each gate
	vInDrive   *FloatVector        // maximum input drive strength
	vTimesOut  *FloatVector        // output arrival times
	vQue       *FloatPriorityQueue // Vec_Que_t * : outputs by their time
	pPiDrive   *SCCell             // cell driving primary inputs
	// backup information
	vLoads2 *FloatVector // backup storage for loads
	vLoads3 *FloatVector // backup storage for loads
	vTimes2 *FloatVector // backup storage for times
	vTimes3 *FloatVector // backup storage for slews
	// buffer trees
	EstLoadMax float32 // max ratio of Cout/Cin when this kicks in
	EstLoadAve float32 // average load of the gate
	EstLinear  float32 // linear coefficient
	nEstNodes  int     // the number of estimations
	// intermediate data
	vNodeByGain *FloatPriorityQueue // Vec_Que_t * : nodes by gain
	vNode2Gain  *FloatVector        // mapping node into its gain
	vNode2Gate  *IntVector          // mapping node into its best gate
	vNodeIter   *IntVector          // the last iteration the node was upsized
	vBestFans   *IntVector          // best fanouts
	// incremental timing update
	vLevels     *IntLevelVector
	vChanged    *IntVector
	nIncUpdates int
	// optimization parameters
	SumArea     float32 // total area
	MaxDelay    float32 // max delay
	SumArea0    float32 // total area at the begining
	MaxDelay0   float32 // max delay at the begining
	BestDelay   float32 // best delay in the middle
	ReportDelay float32 // delay to report
	// runtime statistics
	timeTotal int64 // abctime : starting/total time
	timeCone  int64 // abctime : critical path selection
	timeSize  int64 // abctime : incremental sizing
	timeTime  int64 // abctime : timing update
	timeOther int64 // abctime : everything else
}

// Abc_SclObjLib (static inline)
// C: SC_Lib * Abc_SclObjLib( Abc_Obj_t * p )
func abcSclObjLib(p *ABCNetworkObject) *SCLib {
	panic("TODO, stubbed")
}

// Abc_SclObjCellId (static inline)
// C: int Abc_SclObjCellId( Abc_Obj_t * p )
func abcSclObjCellId(p *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclObjCell (static inline)
// C: SC_Cell * Abc_SclObjCell( Abc_Obj_t * p )
func abcSclObjCell(p *ABCNetworkObject) *SCCell {
	panic("TODO, stubbed")
}

// Abc_SclObjSetCell (static inline)
// C: void Abc_SclObjSetCell( Abc_Obj_t * p, SC_Cell * pCell )
func abcSclObjSetCell(p *ABCNetworkObject, pCell *SCCell) {
	panic("TODO, stubbed")
}

// Abc_SclObjLoad (static inline)
// C: SC_Pair * Abc_SclObjLoad( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclObjLoad(pObj *ABCNetworkObject) *SCPair {
	panic("TODO, stubbed")
}

// Abc_SclObjDept (static inline)
// C: SC_Pair * Abc_SclObjDept( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclObjDept(pObj *ABCNetworkObject) *SCPair {
	panic("TODO, stubbed")
}

// Abc_SclObjTime (static inline)
// C: SC_Pair * Abc_SclObjTime( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclObjTime(pObj *ABCNetworkObject) *SCPair {
	panic("TODO, stubbed")
}

// Abc_SclObjSlew (static inline)
// C: SC_Pair * Abc_SclObjSlew( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclObjSlew(pObj *ABCNetworkObject) *SCPair {
	panic("TODO, stubbed")
}

// Abc_SclObjLoadMax (static inline)
// C: double Abc_SclObjLoadMax( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclObjLoadMax(pObj *ABCNetworkObject) float64 {
	panic("TODO, stubbed")
}

// Abc_SclObjLoadAve (static inline)
// C: float Abc_SclObjLoadAve( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclObjLoadAve(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_SclObjTimeOne (static inline)
// C: double Abc_SclObjTimeOne( SC_Man * p, Abc_Obj_t * pObj, int fRise )
func (p *SCMan) abcSclObjTimeOne(pObj *ABCNetworkObject, fRise int) float64 {
	panic("TODO, stubbed")
}

// Abc_SclObjTimeMax (static inline)
// C: float Abc_SclObjTimeMax( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclObjTimeMax(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_SclObjSlewMax (static inline)
// C: double Abc_SclObjSlewMax( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclObjSlewMax(pObj *ABCNetworkObject) float64 {
	panic("TODO, stubbed")
}

// Abc_SclObjGetSlackR (static inline)
// C: float Abc_SclObjGetSlackR( SC_Man * p, Abc_Obj_t * pObj, float D )
func (p *SCMan) abcSclObjGetSlackR(pObj *ABCNetworkObject, D float32) float32 {
	panic("TODO, stubbed")
}

// Abc_SclObjGetSlackF (static inline)
// C: float Abc_SclObjGetSlackF( SC_Man * p, Abc_Obj_t * pObj, float D )
func (p *SCMan) abcSclObjGetSlackF(pObj *ABCNetworkObject, D float32) float32 {
	panic("TODO, stubbed")
}

// Abc_SclObjGetSlack (static inline)
// C: float Abc_SclObjGetSlack( SC_Man * p, Abc_Obj_t * pObj, float D )
func (p *SCMan) abcSclObjGetSlack(pObj *ABCNetworkObject, D float32) float32 {
	panic("TODO, stubbed")
}

// Abc_SclObjSlackMax (static inline)
// C: double Abc_SclObjSlackMax( SC_Man * p, Abc_Obj_t * pObj, float D )
func (p *SCMan) abcSclObjSlackMax(pObj *ABCNetworkObject, D float32) float64 {
	panic("TODO, stubbed")
}

// Abc_SclObjDupFanin (static inline)
// C: void Abc_SclObjDupFanin( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclObjDupFanin(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclObjInDrive (static inline)
// C: float Abc_SclObjInDrive( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclObjInDrive(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_SclObjSetInDrive (static inline)
// C: void Abc_SclObjSetInDrive( SC_Man * p, Abc_Obj_t * pObj, float c)
func (p *SCMan) abcSclObjSetInDrive(pObj *ABCNetworkObject, c float32) {
	panic("TODO, stubbed")
}

// Abc_SclManAlloc (static inline)
// C: SC_Man * Abc_SclManAlloc( SC_Lib * pLib, Abc_Ntk_t * pNtk )
func (pLib *SCLib) abcSclManAlloc(pNtk *ABCNetwork) *SCMan {
	panic("TODO, stubbed")
}

// Abc_SclManFree (static inline)
// C: void Abc_SclManFree( SC_Man * p )
func (p *SCMan) abcSclManFree() {
	panic("TODO, stubbed")
}

// Abc_SclManCleanTime (static inline)
// C: void Abc_SclManCleanTime( SC_Man * p )
func (p *SCMan) abcSclManCleanTime() {
	panic("TODO, stubbed")
}

// Abc_SclLoadStore (static inline)
// C: void Abc_SclLoadStore( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclLoadStore(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclLoadRestore (static inline)
// C: void Abc_SclLoadRestore( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclLoadRestore(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclLoadStore3 (static inline)
// C: void Abc_SclLoadStore3( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclLoadStore3(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclLoadRestore3 (static inline)
// C: void Abc_SclLoadRestore3( SC_Man * p, Abc_Obj_t * pObj )
func (p *SCMan) abcSclLoadRestore3(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclConeStore (static inline)
// C: void Abc_SclConeStore( SC_Man * p, Vec_Int_t * vCone )
func (p *SCMan) abcSclConeStore(vCone *IntVector) {
	panic("TODO, stubbed")
}

// Abc_SclConeRestore (static inline)
// C: void Abc_SclConeRestore( SC_Man * p, Vec_Int_t * vCone )
func (p *SCMan) abcSclConeRestore(vCone *IntVector) {
	panic("TODO, stubbed")
}

// Abc_SclEvalStore (static inline)
// C: void Abc_SclEvalStore( SC_Man * p, Vec_Int_t * vCone )
func (p *SCMan) abcSclEvalStore(vCone *IntVector) {
	panic("TODO, stubbed")
}

// Abc_SclEvalPerform (static inline)
// C: float Abc_SclEvalPerform( SC_Man * p, Vec_Int_t * vCone )
func (p *SCMan) abcSclEvalPerform(vCone *IntVector) float32 {
	panic("TODO, stubbed")
}

// Abc_SclEvalPerformLegal (static inline)
// C: float Abc_SclEvalPerformLegal( SC_Man * p, Vec_Int_t * vCone, float D )
func (p *SCMan) abcSclEvalPerformLegal(vCone *IntVector, D float32) float32 {
	panic("TODO, stubbed")
}

// Abc_SclConeClean (static inline)
// C: void Abc_SclConeClean( SC_Man * p, Vec_Int_t * vCone )
func (p *SCMan) abcSclConeClean(vCone *IntVector) {
	panic("TODO, stubbed")
}

// Abc_SclGetBufInvCount (static inline)
// C: int Abc_SclGetBufInvCount( Abc_Ntk_t * pNtk )
func abcSclGetBufInvCount(pNtk *ABCNetwork) int {
	panic("TODO, stubbed")
}

// Abc_SclGetAverageSize (static inline)
// C: float Abc_SclGetAverageSize( Abc_Ntk_t * pNtk )
func abcSclGetAverageSize(pNtk *ABCNetwork) float32 {
	panic("TODO, stubbed")
}

// Abc_SclGetTotalArea (static inline)
// C: float Abc_SclGetTotalArea( Abc_Ntk_t * pNtk )
func abcSclGetTotalArea(pNtk *ABCNetwork) float32 {
	panic("TODO, stubbed")
}

// Abc_SclGetMaxDelay (static inline)
// C: float Abc_SclGetMaxDelay( SC_Man * p )
func (p *SCMan) abcSclGetMaxDelay() float32 {
	panic("TODO, stubbed")
}

// Abc_SclGetMaxDelayNodeFanins (static inline)
// C: float Abc_SclGetMaxDelayNodeFanins( SC_Man * p, Abc_Obj_t * pNode )
func (p *SCMan) abcSclGetMaxDelayNodeFanins(pNode *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_SclReadMaxDelay (static inline)
// C: float Abc_SclReadMaxDelay( SC_Man * p )
func (p *SCMan) abcSclReadMaxDelay() float32 {
	panic("TODO, stubbed")
}

// Abc_SclObjResiable (static inline)
// C: SC_Cell * Abc_SclObjResiable( SC_Man * p, Abc_Obj_t * pObj, int fUpsize )
func (p *SCMan) abcSclObjResiable(pObj *ABCNetworkObject, fUpsize int) *SCCell {
	panic("TODO, stubbed")
}

// Abc_SclDumpStats (static inline)
// C: void Abc_SclDumpStats( SC_Man * p, char * pFileName, abctime Time )
func (p *SCMan) abcSclDumpStats(pFileName string, Time int64) {
	panic("TODO, stubbed")
}
