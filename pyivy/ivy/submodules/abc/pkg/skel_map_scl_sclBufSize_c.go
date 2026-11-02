package abc

// Hand-checked skeleton of src/map/scl/sclBufSize.c.

// BusMan is Bus_Man_t.
type BusMan struct {
	// user data
	pPars    *SCBusPars  // parameters
	pNtk     *ABCNetwork // user's network
	pPiDrive *SCCell     // PI driver
	// library
	pLib       *SCLib       // cell library
	pInv       *SCCell      // base interter (largest/average/???)
	pWLoadUsed *SCWireLoad  // name of the used WireLoad model
	vWireCaps  *FloatVector // estimated wire loads
	// internal
	vCins    *FloatVector                     // input cap for fanouts
	vETimes  *FloatVector                     // fanout edge departures
	vLoads   *FloatVector                     // loads for all nodes
	vDepts   *FloatVector                     // departure times
	vFanouts *PointerVector[ABCNetworkObject] // fanout array
}

// Bus_SclObjMan (static)
// C: Bus_Man_t * Bus_SclObjMan( Abc_Obj_t * p )
func busSclObjMan(p *ABCNetworkObject) *BusMan {
	panic("TODO, stubbed")
}

// Bus_SclObjCin (static)
// C: float Bus_SclObjCin( Abc_Obj_t * p )
func busSclObjCin(p *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Bus_SclObjSetCin (static)
// C: void Bus_SclObjSetCin( Abc_Obj_t * p, float cap )
func busSclObjSetCin(p *ABCNetworkObject, cap float32) {
	panic("TODO, stubbed")
}

// Bus_SclObjETime (static)
// C: float Bus_SclObjETime( Abc_Obj_t * p )
func busSclObjETime(p *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Bus_SclObjSetETime (static)
// C: void Bus_SclObjSetETime( Abc_Obj_t * p, float time )
func busSclObjSetETime(p *ABCNetworkObject, time float32) {
	panic("TODO, stubbed")
}

// Bus_SclObjLoad (static)
// C: float Bus_SclObjLoad( Abc_Obj_t * p )
func busSclObjLoad(p *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Bus_SclObjSetLoad (static)
// C: void Bus_SclObjSetLoad( Abc_Obj_t * p, float cap )
func busSclObjSetLoad(p *ABCNetworkObject, cap float32) {
	panic("TODO, stubbed")
}

// Bus_SclObjDept (static)
// C: float Bus_SclObjDept( Abc_Obj_t * p )
func busSclObjDept(p *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Bus_SclObjUpdateDept (static)
// C: void Bus_SclObjUpdateDept( Abc_Obj_t * p, float time )
func busSclObjUpdateDept(p *ABCNetworkObject, time float32) {
	panic("TODO, stubbed")
}

// Bus_ManStart
// C: Bus_Man_t * Bus_ManStart( Abc_Ntk_t * pNtk, SC_Lib * pLib, SC_BusPars * pPars )
func BusManStart(pNtk *ABCNetwork, pLib *SCLib, pPars *SCBusPars) *BusMan {
	panic("TODO, stubbed")
}

// Bus_ManStop
// C: void Bus_ManStop( Bus_Man_t * p )
func (p *BusMan) BusManStop() {
	panic("TODO, stubbed")
}

// Bus_ManReadInOutLoads
// C: void Bus_ManReadInOutLoads( Bus_Man_t * p )
func (p *BusMan) BusManReadInOutLoads() {
	panic("TODO, stubbed")
}

// Abc_NtkComputeEdgeDept (static)
// C: float Abc_NtkComputeEdgeDept( Abc_Obj_t * pFanout, int iFanin, float Slew )
func abcNtkComputeEdgeDept(pFanout *ABCNetworkObject, iFanin int, Slew float32) float32 {
	panic("TODO, stubbed")
}

// Abc_NtkComputeNodeDeparture
// C: float Abc_NtkComputeNodeDeparture( Abc_Obj_t * pObj, float Slew )
func AbcNtkComputeNodeDeparture(pObj *ABCNetworkObject, Slew float32) float32 {
	panic("TODO, stubbed")
}

// Abc_NtkComputeFanoutInfo
// C: void Abc_NtkComputeFanoutInfo( Abc_Obj_t * pObj, float Slew )
func AbcNtkComputeFanoutInfo(pObj *ABCNetworkObject, Slew float32) {
	panic("TODO, stubbed")
}

// Abc_NtkComputeNodeLoad
// C: float Abc_NtkComputeNodeLoad( Bus_Man_t * p, Abc_Obj_t * pObj )
func (p *BusMan) AbcNtkComputeNodeLoad(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_NtkComputeFanoutLoad
// C: float Abc_NtkComputeFanoutLoad( Bus_Man_t * p, Vec_Ptr_t * vFanouts )
func (p *BusMan) AbcNtkComputeFanoutLoad(vFanouts *PointerVector[ABCNetworkObject]) float32 {
	panic("TODO, stubbed")
}

// Abc_NtkPrintFanoutProfile
// C: void Abc_NtkPrintFanoutProfile( Abc_Obj_t * pObj )
func AbcNtkPrintFanoutProfile(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintFanoutProfileVec
// C: void Abc_NtkPrintFanoutProfileVec( Abc_Obj_t * pObj, Vec_Ptr_t * vFanouts )
func AbcNtkPrintFanoutProfileVec(pObj *ABCNetworkObject, vFanouts *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Bus_SclCompareFanouts
// C: int Bus_SclCompareFanouts( Abc_Obj_t ** pp1, Abc_Obj_t ** pp2 )
func BusSclCompareFanouts(pp1 **ABCNetworkObject, pp2 **ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Bus_SclInsertFanout
// C: void Bus_SclInsertFanout( Vec_Ptr_t * vFanouts, Abc_Obj_t * pObj )
func BusSclInsertFanout(vFanouts *PointerVector[ABCNetworkObject], pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Bus_SclCheckSortedFanout
// C: void Bus_SclCheckSortedFanout( Vec_Ptr_t * vFanouts )
func BusSclCheckSortedFanout(vFanouts *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_SclOneNodePrint
// C: void Abc_SclOneNodePrint( Bus_Man_t * p, Abc_Obj_t * pObj )
func (p *BusMan) AbcSclOneNodePrint(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclAddOneInv
// C: Abc_Obj_t * Abc_SclAddOneInv( Bus_Man_t * p, Abc_Obj_t * pObj, Vec_Ptr_t * vFanouts, float Gain )
func (p *BusMan) AbcSclAddOneInv(pObj *ABCNetworkObject, vFanouts *PointerVector[ABCNetworkObject], Gain float32) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_SclBufSize
// C: void Abc_SclBufSize( Bus_Man_t * p, float Gain )
func (p *BusMan) AbcSclBufSize(Gain float32) {
	panic("TODO, stubbed")
}

// Abc_SclBufferingPerform
// C: Abc_Ntk_t * Abc_SclBufferingPerform( Abc_Ntk_t * pNtk, SC_Lib * pLib, SC_BusPars * pPars )
func AbcSclBufferingPerform(pNtk *ABCNetwork, pLib *SCLib, pPars *SCBusPars) *ABCNetwork {
	panic("TODO, stubbed")
}
