package abc

// Hand-checked skeleton of src/base/abci/abcIf.c.

// If_ManComputeSwitching
// C: void If_ManComputeSwitching( If_Man_t * pIfMan )
func (pIfMan *IfMan) IfManComputeSwitching() {
	panic("TODO, stubbed")
}

// Abc_NtkIf
// C: Abc_Ntk_t * Abc_NtkIf( Abc_Ntk_t * pNtk, If_Par_t * pPars )
func (pNtk *ABCNetwork) AbcNtkIf(pPars *IfPar) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_ObjIfCopy (static)
// C: If_Obj_t * Abc_ObjIfCopy( Abc_Obj_t * pNode )
func abcObjIfCopy(pNode *ABCNetworkObject) *IfObj {
	panic("TODO, stubbed")
}

// Abc_NtkToIf
// C: If_Man_t * Abc_NtkToIf( Abc_Ntk_t * pNtk, If_Par_t * pPars )
func (pNtk *ABCNetwork) AbcNtkToIf(pPars *IfPar) *IfMan {
	panic("TODO, stubbed")
}

// Abc_MapBoxSetPrevNext (static)
// C: void Abc_MapBoxSetPrevNext( Vec_Ptr_t * vDrivers, Vec_Int_t * vMapIn, Vec_Int_t * vMapOut, int Id )
func abcMapBoxSetPrevNext_abcIf(vDrivers *PointerVector[ABCNetworkObject], vMapIn *IntVector, vMapOut *IntVector, Id int) {
	panic("TODO, stubbed")
}

// Abc_MapBox2Next (static)
// C: int Abc_MapBox2Next( Vec_Ptr_t * vDrivers, Vec_Int_t * vMapIn, Vec_Int_t * vMapOut, int Id )
func abcMapBox2Next_abcIf(vDrivers *PointerVector[ABCNetworkObject], vMapIn *IntVector, vMapOut *IntVector, Id int) int {
	panic("TODO, stubbed")
}

// Abc_MapBox2Prev (static)
// C: int Abc_MapBox2Prev( Vec_Ptr_t * vDrivers, Vec_Int_t * vMapIn, Vec_Int_t * vMapOut, int Id )
func abcMapBox2Prev_abcIf(vDrivers *PointerVector[ABCNetworkObject], vMapIn *IntVector, vMapOut *IntVector, Id int) int {
	panic("TODO, stubbed")
}

// Abc_NtkFromIf (static)
// C: Abc_Ntk_t * Abc_NtkFromIf( If_Man_t * pIfMan, Abc_Ntk_t * pNtk )
func (pIfMan *IfMan) abcNtkFromIf(pNtk *ABCNetwork) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NodeBuildFromMiniInt
// C: Hop_Obj_t * Abc_NodeBuildFromMiniInt( Hop_Man_t * pMan, Vec_Int_t * vAig, int nLeaves )
func (pMan *HopManager) AbcNodeBuildFromMiniInt(vAig *IntVector, nLeaves int) *HopObject {
	panic("TODO, stubbed")
}

// Abc_NodeBuildFromMini
// C: Hop_Obj_t * Abc_NodeBuildFromMini( Hop_Man_t * pMan, If_Man_t * p, If_Cut_t * pCut, int fUseDsd )
func (pMan *HopManager) AbcNodeBuildFromMini(p *IfMan, pCut *IfCut, fUseDsd int) *HopObject {
	panic("TODO, stubbed")
}

// Abc_NodeFromIf_rec
// C: Abc_Obj_t * Abc_NodeFromIf_rec( Abc_Ntk_t * pNtkNew, If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Int_t * vCover )
func (pNtkNew *ABCNetwork) AbcNodeFromIf_rec(pIfMan *IfMan, pIfObj *IfObj, vCover *IntVector) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NodeIfToHop_rec
// C: Hop_Obj_t * Abc_NodeIfToHop_rec( Hop_Man_t * pHopMan, If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Ptr_t * vVisited )
func (pHopMan *HopManager) AbcNodeIfToHop_rec(pIfMan *IfMan, pIfObj *IfObj, vVisited *PointerVector[IfObj]) *HopObject {
	panic("TODO, stubbed")
}

// Abc_NodeIfToHop2_rec
// C: Hop_Obj_t * Abc_NodeIfToHop2_rec( Hop_Man_t * pHopMan, If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Ptr_t * vVisited )
func (pHopMan *HopManager) AbcNodeIfToHop2_rec(pIfMan *IfMan, pIfObj *IfObj, vVisited *PointerVector[IfObj]) *HopObject {
	panic("TODO, stubbed")
}

// Abc_NodeIfToHop (static)
// C: Hop_Obj_t * Abc_NodeIfToHop( Hop_Man_t * pHopMan, If_Man_t * pIfMan, If_Obj_t * pIfObj )
func (pHopMan *HopManager) abcNodeIfToHop(pIfMan *IfMan, pIfObj *IfObj) *HopObject {
	panic("TODO, stubbed")
}

// Abc_ObjCompareFlow
// C: int Abc_ObjCompareFlow( Abc_Obj_t ** ppNode0, Abc_Obj_t ** ppNode1 )
func AbcObjCompareFlow(ppNode0 **ABCNetworkObject, ppNode1 **ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkFindGoodOrder_rec
// C: void Abc_NtkFindGoodOrder_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vNodes )
func AbcNtkFindGoodOrder_rec(pNode *ABCNetworkObject, vNodes *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkFindGoodOrder (static)
// C: Vec_Ptr_t * Abc_NtkFindGoodOrder( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkFindGoodOrder() *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkMarkMux
// C: void Abc_NtkMarkMux( Abc_Obj_t * pDriver, Abc_Obj_t ** ppNode1, Abc_Obj_t ** ppNode2 )
func AbcNtkMarkMux(pDriver *ABCNetworkObject, ppNode1 **ABCNetworkObject, ppNode2 **ABCNetworkObject) {
	panic("TODO, stubbed")
}
