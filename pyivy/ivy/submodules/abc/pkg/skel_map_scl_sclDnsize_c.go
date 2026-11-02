package abc

// Hand-checked skeleton of src/map/scl/sclDnsize.c.

// Abc_SclFindWindow (pvNodes and pvEvals are out parameters)
// C: void Abc_SclFindWindow( Abc_Obj_t * pPivot, Vec_Int_t ** pvNodes, Vec_Int_t ** pvEvals )
func AbcSclFindWindow(pPivot *ABCNetworkObject, pvNodes **IntVector, pvEvals **IntVector) {
	panic("TODO, stubbed")
}

// Abc_SclCheckImprovement
// C: int Abc_SclCheckImprovement( SC_Man * p, Abc_Obj_t * pObj, Vec_Int_t * vNodes, Vec_Int_t * vEvals, int Notches, int DelayGap )
func (p *SCMan) AbcSclCheckImprovement(pObj *ABCNetworkObject, vNodes *IntVector, vEvals *IntVector, Notches int, DelayGap int) int {
	panic("TODO, stubbed")
}

// Abc_NtkCollectNodesByArea
// C: void Abc_NtkCollectNodesByArea( SC_Man * p, Abc_Ntk_t * pNtk )
func (p *SCMan) AbcNtkCollectNodesByArea(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_SclCheckOverlap
// C: int Abc_SclCheckOverlap( Abc_Ntk_t * pNtk, Vec_Int_t * vNodes )
func AbcSclCheckOverlap(pNtk *ABCNetwork, vNodes *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_SclDnsizePrint
// C: void Abc_SclDnsizePrint( SC_Man * p, int Iter, int nAttempts, int nOverlaps, int nChanges, int fVerbose )
func (p *SCMan) AbcSclDnsizePrint(Iter int, nAttempts int, nOverlaps int, nChanges int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_SclDnsizePerformInt
// C: void Abc_SclDnsizePerformInt( SC_Lib * pLib, Abc_Ntk_t * pNtk, SC_SizePars * pPars )
func (pLib *SCLib) AbcSclDnsizePerformInt(pNtk *ABCNetwork, pPars *SCSizePars) {
	panic("TODO, stubbed")
}

// Abc_SclDnsizePerform
// C: void Abc_SclDnsizePerform( SC_Lib * pLib, Abc_Ntk_t * pNtk, SC_SizePars * pPars )
func (pLib *SCLib) AbcSclDnsizePerform(pNtk *ABCNetwork, pPars *SCSizePars) {
	panic("TODO, stubbed")
}
