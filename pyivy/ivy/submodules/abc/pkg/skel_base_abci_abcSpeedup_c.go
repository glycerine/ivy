package abc

// Hand-checked skeleton of src/base/abci/abcSpeedup.c.

// Abc_ObjArrival (static inline)
// C: float Abc_ObjArrival( Abc_Obj_t * pNode )
func abcObjArrival(pNode *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_ObjRequired (static inline)
// C: float Abc_ObjRequired( Abc_Obj_t * pNode )
func abcObjRequired(pNode *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_ObjSlack (static inline)
// C: float Abc_ObjSlack( Abc_Obj_t * pNode )
func abcObjSlack(pNode *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_ObjSetArrival (static inline)
// C: void Abc_ObjSetArrival( Abc_Obj_t * pNode, float Time )
func abcObjSetArrival(pNode *ABCNetworkObject, Time float32) {
	panic("TODO, stubbed")
}

// Abc_ObjSetRequired (static inline)
// C: void Abc_ObjSetRequired( Abc_Obj_t * pNode, float Time )
func abcObjSetRequired(pNode *ABCNetworkObject, Time float32) {
	panic("TODO, stubbed")
}

// Abc_ObjSetSlack (static inline)
// C: void Abc_ObjSetSlack( Abc_Obj_t * pNode, float Time )
func abcObjSetSlack(pNode *ABCNetworkObject, Time float32) {
	panic("TODO, stubbed")
}

// Abc_NtkDelayTraceSortPins
// C: void Abc_NtkDelayTraceSortPins( Abc_Obj_t * pNode, int * pPinPerm, float * pPinDelays )
func AbcNtkDelayTraceSortPins(pNode *ABCNetworkObject, pPinPerm []int, pPinDelays []float32) {
	panic("TODO, stubbed")
}

// Abc_NtkDelayTraceLut
// C: float Abc_NtkDelayTraceLut( Abc_Ntk_t * pNtk, int fUseLutLib )
func (pNtk *ABCNetwork) AbcNtkDelayTraceLut(fUseLutLib int) float32 {
	panic("TODO, stubbed")
}

// Abc_NtkDelayTracePrint
// C: void Abc_NtkDelayTracePrint( Abc_Ntk_t * pNtk, int fUseLutLib, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDelayTracePrint(fUseLutLib int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_AigCheckTfi_rec
// C: int Abc_AigCheckTfi_rec( Abc_Obj_t * pNode, Abc_Obj_t * pOld )
func AbcAigCheckTfi_rec(pNode *ABCNetworkObject, pOld *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_AigCheckTfi
// C: int Abc_AigCheckTfi( Abc_Obj_t * pNew, Abc_Obj_t * pOld )
func AbcAigCheckTfi(pNew *ABCNetworkObject, pOld *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkSpeedupNode_rec
// C: int Abc_NtkSpeedupNode_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vNodes )
func AbcNtkSpeedupNode_rec(pNode *ABCNetworkObject, vNodes *PointerVector[ABCNetworkObject]) int {
	panic("TODO, stubbed")
}

// Abc_NtkSpeedupNode (vTimes holds float arrival times encoded as pointer-sized values)
// C: void Abc_NtkSpeedupNode( Abc_Ntk_t * pNtk, Abc_Ntk_t * pAig, Abc_Obj_t * pNode, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vTimes )
func (pNtk *ABCNetwork) AbcNtkSpeedupNode(pAig *ABCNetwork, pNode *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], vTimes *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkDelayTraceTCEdges
// C: unsigned Abc_NtkDelayTraceTCEdges( Abc_Ntk_t * pNtk, Abc_Obj_t * pNode, float tDelta, int fUseLutLib )
func (pNtk *ABCNetwork) AbcNtkDelayTraceTCEdges(pNode *ABCNetworkObject, tDelta float32, fUseLutLib int) uint32 {
	panic("TODO, stubbed")
}

// Abc_NtkSpeedup
// C: Abc_Ntk_t * Abc_NtkSpeedup( Abc_Ntk_t * pNtk, int fUseLutLib, int Percentage, int Degree, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) AbcNtkSpeedup(fUseLutLib int, Percentage int, Degree int, fVerbose int, fVeryVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkPowerEstimate
// C: Vec_Int_t * Abc_NtkPowerEstimate( Abc_Ntk_t * pNtk, int fProbOne )
func (pNtk *ABCNetwork) AbcNtkPowerEstimate(fProbOne int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkPowerPrint
// C: void Abc_NtkPowerPrint( Abc_Ntk_t * pNtk, Vec_Int_t * vProbs )
func (pNtk *ABCNetwork) AbcNtkPowerPrint(vProbs *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkPowerCriticalEdges
// C: unsigned Abc_NtkPowerCriticalEdges( Abc_Ntk_t * pNtk, Abc_Obj_t * pNode, float Limit, Vec_Int_t * vProbs )
func (pNtk *ABCNetwork) AbcNtkPowerCriticalEdges(pNode *ABCNetworkObject, Limit float32, vProbs *IntVector) uint32 {
	panic("TODO, stubbed")
}

// Abc_NtkPowerdown
// C: Abc_Ntk_t * Abc_NtkPowerdown( Abc_Ntk_t * pNtk, int fUseLutLib, int Percentage, int Degree, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) AbcNtkPowerdown(fUseLutLib int, Percentage int, Degree int, fVerbose int, fVeryVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}
