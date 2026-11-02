package abc

// Hand-checked skeleton of src/opt/fret/fretInit.c.

// Abc_FlowRetime_InitState
// C: void Abc_FlowRetime_InitState( Abc_Ntk_t * pNtk )
func AbcFlowRetime_InitState(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_ObjFirstNonLatchBox (static inline; pResult is an out parameter)
// C: int Abc_FlowRetime_ObjFirstNonLatchBox( Abc_Obj_t * pOrigObj, Abc_Obj_t ** pResult )
func abcFlowRetime_ObjFirstNonLatchBox(pOrigObj *ABCNetworkObject, pResult **ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_PrintInitStateInfo
// C: void Abc_FlowRetime_PrintInitStateInfo( Abc_Ntk_t * pNtk )
func AbcFlowRetime_PrintInitStateInfo(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_UpdateForwardInit
// C: void Abc_FlowRetime_UpdateForwardInit( Abc_Ntk_t * pNtk )
func AbcFlowRetime_UpdateForwardInit(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_UpdateForwardInit_rec (static)
// C: void Abc_FlowRetime_UpdateForwardInit_rec( Abc_Obj_t * pObj )
func abcFlowRetime_UpdateForwardInit_rec(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_EvalHop_rec (static; f and dc are out parameters)
// C: void Abc_FlowRetime_EvalHop_rec( Hop_Man_t *pHop, Hop_Obj_t *pObj, int *f, int *dc )
func abcFlowRetime_EvalHop_rec(pHop *HopManager, pObj *HopObject, f *int, dc *int) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_SetInitValue (static inline)
// C: void Abc_FlowRetime_SetInitValue( Abc_Obj_t * pObj, int val, int dc )
func abcFlowRetime_SetInitValue(pObj *ABCNetworkObject, val int, dc int) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_SimulateNode (static)
// C: void Abc_FlowRetime_SimulateNode( Abc_Obj_t * pObj )
func abcFlowRetime_SimulateNode(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_SimulateSop (static)
// C: void Abc_FlowRetime_SimulateSop( Abc_Obj_t * pObj, char *pSop )
func abcFlowRetime_SimulateSop(pObj *ABCNetworkObject, pSop string) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_SetupBackwardInit
// C: void Abc_FlowRetime_SetupBackwardInit( Abc_Ntk_t * pNtk )
func AbcFlowRetime_SetupBackwardInit(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_SolveBackwardInit
// C: int Abc_FlowRetime_SolveBackwardInit( Abc_Ntk_t * pNtk )
func AbcFlowRetime_SolveBackwardInit(pNtk *ABCNetwork) int {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_UpdateBackwardInit
// C: void Abc_FlowRetime_UpdateBackwardInit( Abc_Ntk_t * pNtk )
func AbcFlowRetime_UpdateBackwardInit(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_CopyNodeToInitNtk
// C: Abc_Obj_t * Abc_FlowRetime_CopyNodeToInitNtk( Abc_Obj_t *pOrigObj )
func AbcFlowRetime_CopyNodeToInitNtk(pOrigObj *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_UpdateBackwardInit_rec (static)
// C: Abc_Obj_t* Abc_FlowRetime_UpdateBackwardInit_rec( Abc_Obj_t *pOrigObj)
func abcFlowRetime_UpdateBackwardInit_rec(pOrigObj *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_VerifyBackwardInit (static)
// C: void Abc_FlowRetime_VerifyBackwardInit( Abc_Ntk_t * pNtk )
func abcFlowRetime_VerifyBackwardInit(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_VerifyBackwardInit_rec (static)
// C: void Abc_FlowRetime_VerifyBackwardInit_rec( Abc_Obj_t * pObj )
func abcFlowRetime_VerifyBackwardInit_rec(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_PartialSat (the Vec_Ptr_t holds Abc_Obj_t *)
// C: int Abc_FlowRetime_PartialSat(Vec_Ptr_t *vNodes, int cut)
func AbcFlowRetime_PartialSat(vNodes *PointerVector[ABCNetworkObject], cut int) int {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_ConstrainInit
// C: void Abc_FlowRetime_ConstrainInit( )
func AbcFlowRetime_ConstrainInit() {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_RemoveInitBias
// C: void Abc_FlowRetime_RemoveInitBias( )
func AbcFlowRetime_RemoveInitBias() {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_ConnectBiasNode (static)
// C: void Abc_FlowRetime_ConnectBiasNode(Abc_Obj_t *pBiasNode, Abc_Obj_t *pObj, int biasLag)
func abcFlowRetime_ConnectBiasNode(pBiasNode *ABCNetworkObject, pObj *ABCNetworkObject, biasLag int) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_AddInitBias
// C: void Abc_FlowRetime_AddInitBias( )
func AbcFlowRetime_AddInitBias() {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_ClearInitToOrig (static)
// C: void Abc_FlowRetime_ClearInitToOrig( Abc_Obj_t *pInit )
func abcFlowRetime_ClearInitToOrig(pInit *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_SetInitToOrig (static)
// C: void Abc_FlowRetime_SetInitToOrig( Abc_Obj_t *pInit, Abc_Obj_t *pOrig)
func abcFlowRetime_SetInitToOrig(pInit *ABCNetworkObject, pOrig *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_FlowRetime_GetInitToOrig (static; pOrig and lag are out parameters)
// C: void Abc_FlowRetime_GetInitToOrig( Abc_Obj_t *pInit, Abc_Obj_t **pOrig, int *lag )
func abcFlowRetime_GetInitToOrig(pInit *ABCNetworkObject, pOrig **ABCNetworkObject, lag *int) {
	panic("TODO, stubbed")
}
