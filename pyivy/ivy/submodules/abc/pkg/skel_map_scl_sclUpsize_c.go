package abc

// Hand-checked skeleton of src/map/scl/sclUpsize.c.

// Abc_SclFindTFO_rec
// C: void Abc_SclFindTFO_rec( Abc_Obj_t * pObj, Vec_Int_t * vNodes, Vec_Int_t * vCos )
func AbcSclFindTFO_rec(pObj *ABCNetworkObject, vNodes *IntVector, vCos *IntVector) {
	panic("TODO, stubbed")
}

// Abc_SclFindTFO
// C: Vec_Int_t * Abc_SclFindTFO( Abc_Ntk_t * p, Vec_Int_t * vPath )
func AbcSclFindTFO(p *ABCNetwork, vPath *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Abc_SclFindCriticalCoWindow
// C: Vec_Int_t * Abc_SclFindCriticalCoWindow( SC_Man * p, int Window )
func (p *SCMan) AbcSclFindCriticalCoWindow(Window int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_SclFindCriticalNodeWindow_rec
// C: void Abc_SclFindCriticalNodeWindow_rec( SC_Man * p, Abc_Obj_t * pObj, Vec_Int_t * vPath, float fSlack, int fDept )
func (p *SCMan) AbcSclFindCriticalNodeWindow_rec(pObj *ABCNetworkObject, vPath *IntVector, fSlack float32, fDept int) {
	panic("TODO, stubbed")
}

// Abc_SclFindCriticalNodeWindow
// C: Vec_Int_t * Abc_SclFindCriticalNodeWindow( SC_Man * p, Vec_Int_t * vPathCos, int Window, int fDept )
func (p *SCMan) AbcSclFindCriticalNodeWindow(vPathCos *IntVector, Window int, fDept int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_SclUnmarkCriticalNodeWindow
// C: void Abc_SclUnmarkCriticalNodeWindow( SC_Man * p, Vec_Int_t * vPath )
func (p *SCMan) AbcSclUnmarkCriticalNodeWindow(vPath *IntVector) {
	panic("TODO, stubbed")
}

// Abc_SclCountNearCriticalNodes
// C: int Abc_SclCountNearCriticalNodes( SC_Man * p )
func (p *SCMan) AbcSclCountNearCriticalNodes() int {
	panic("TODO, stubbed")
}

// Abc_SclFindNodesToUpdate (pvNodes and pvEvals are out parameters)
// C: void Abc_SclFindNodesToUpdate( Abc_Obj_t * pPivot, Vec_Int_t ** pvNodes, Vec_Int_t ** pvEvals, Abc_Obj_t * pExtra )
func AbcSclFindNodesToUpdate(pPivot *ABCNetworkObject, pvNodes **IntVector, pvEvals **IntVector, pExtra *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_SclFindBestCell (pGainBest is an out parameter)
// C: int Abc_SclFindBestCell( SC_Man * p, Abc_Obj_t * pObj, Vec_Int_t * vRecalcs, Vec_Int_t * vEvals, int Notches, int DelayGap, float * pGainBest )
func (p *SCMan) AbcSclFindBestCell(pObj *ABCNetworkObject, vRecalcs *IntVector, vEvals *IntVector, Notches int, DelayGap int, pGainBest *float32) int {
	panic("TODO, stubbed")
}

// Abc_SclFindBypasses
// C: int Abc_SclFindBypasses( SC_Man * p, Vec_Int_t * vPathNodes, int Ratio, int Notches, int iIter, int DelayGap, int fVeryVerbose )
func (p *SCMan) AbcSclFindBypasses(vPathNodes *IntVector, Ratio int, Notches int, iIter int, DelayGap int, fVeryVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_SclObjCheckMarkedFanFans
// C: int Abc_SclObjCheckMarkedFanFans( Abc_Obj_t * pObj )
func AbcSclObjCheckMarkedFanFans(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclObjMarkFanFans
// C: void Abc_SclObjMarkFanFans( Abc_Obj_t * pObj, Vec_Ptr_t * vNodes )
func AbcSclObjMarkFanFans(pObj *ABCNetworkObject, vNodes *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_SclFindUpsizes
// C: int Abc_SclFindUpsizes( SC_Man * p, Vec_Int_t * vPathNodes, int Ratio, int Notches, int iIter, int DelayGap, int fMoreConserf )
func (p *SCMan) AbcSclFindUpsizes(vPathNodes *IntVector, Ratio int, Notches int, iIter int, DelayGap int, fMoreConserf int) int {
	panic("TODO, stubbed")
}

// Abc_SclApplyUpdateToBest
// C: void Abc_SclApplyUpdateToBest( Vec_Int_t * vGatesBest, Vec_Int_t * vGates, Vec_Int_t * vUpdate )
func AbcSclApplyUpdateToBest(vGatesBest *IntVector, vGates *IntVector, vUpdate *IntVector) {
	panic("TODO, stubbed")
}

// Abc_SclUndoRecentChanges
// C: void Abc_SclUndoRecentChanges( Abc_Ntk_t * pNtk, Vec_Int_t * vTrans )
func AbcSclUndoRecentChanges(pNtk *ABCNetwork, vTrans *IntVector) {
	panic("TODO, stubbed")
}

// Abc_SclUpsizePrintDiffs
// C: void Abc_SclUpsizePrintDiffs( SC_Man * p, SC_Lib * pLib, Abc_Ntk_t * pNtk )
func (p *SCMan) AbcSclUpsizePrintDiffs(pLib *SCLib, pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_SclUpsizePrint
// C: void Abc_SclUpsizePrint( SC_Man * p, int Iter, int win, int nPathPos, int nPathNodes, int nUpsizes, int nTFOs, int fVerbose )
func (p *SCMan) AbcSclUpsizePrint(Iter int, win int, nPathPos int, nPathNodes int, nUpsizes int, nTFOs int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_SclUpsizeRemoveDangling
// C: void Abc_SclUpsizeRemoveDangling( SC_Man * p, Abc_Ntk_t * pNtk )
func (p *SCMan) AbcSclUpsizeRemoveDangling(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_SclUpsizePerformInt
// C: void Abc_SclUpsizePerformInt( SC_Lib * pLib, Abc_Ntk_t * pNtk, SC_SizePars * pPars )
func (pLib *SCLib) AbcSclUpsizePerformInt(pNtk *ABCNetwork, pPars *SCSizePars) {
	panic("TODO, stubbed")
}

// Abc_SclUpsizePerform
// C: void Abc_SclUpsizePerform( SC_Lib * pLib, Abc_Ntk_t * pNtk, SC_SizePars * pPars )
func (pLib *SCLib) AbcSclUpsizePerform(pNtk *ABCNetwork, pPars *SCSizePars) {
	panic("TODO, stubbed")
}
