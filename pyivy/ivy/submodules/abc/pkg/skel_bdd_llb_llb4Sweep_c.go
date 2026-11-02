package abc

// Hand-checked skeleton of src/bdd/llb/llb4Sweep.c.

// Llb_Nonlin4SweepOrder_rec
// C: void Llb_Nonlin4SweepOrder_rec( Aig_Man_t * pAig, Aig_Obj_t * pObj, Vec_Int_t * vOrder, int * pCounter, int fSaveAll )
func (pAig *AigManager) LlbNonlin4SweepOrder_rec(pObj *AigObject, vOrder *IntVector, pCounter *int, fSaveAll int) {
	panic("TODO, stubbed")
}

// Llb_Nonlin4SweepOrder
// C: Vec_Int_t * Llb_Nonlin4SweepOrder( Aig_Man_t * pAig, int * pCounter, int fSaveAll )
func (pAig *AigManager) LlbNonlin4SweepOrder(pCounter *int, fSaveAll int) *IntVector {
	panic("TODO, stubbed")
}

// Llb4_Nonlin4SweepCutpoints
// C: int Llb4_Nonlin4SweepCutpoints( Aig_Man_t * pAig, Vec_Int_t * vOrder, int nBddLimit, int fVerbose )
func (pAig *AigManager) Llb4Nonlin4SweepCutpoints(vOrder *IntVector, nBddLimit int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Llb_Nonlin4SweepPartitions_rec
// C: DdNode * Llb_Nonlin4SweepPartitions_rec( DdManager * dd, Aig_Obj_t * pObj, Vec_Int_t * vOrder, Vec_Ptr_t * vRoots )
func LlbNonlin4SweepPartitions_rec(dd *CuddFoundation, pObj *AigObject, vOrder *IntVector, vRoots *PointerVector[DdNode]) *DdNode {
	panic("TODO, stubbed")
}

// Llb_Nonlin4SweepPartitions
// C: Vec_Ptr_t * Llb_Nonlin4SweepPartitions( DdManager * dd, Aig_Man_t * pAig, Vec_Int_t * vOrder, int fTransition )
func LlbNonlin4SweepPartitions(dd *CuddFoundation, pAig *AigManager, vOrder *IntVector, fTransition int) *PointerVector[DdNode] {
	panic("TODO, stubbed")
}

// Llb4_Nonlin4SweepBadMonitor
// C: DdNode * Llb4_Nonlin4SweepBadMonitor( Aig_Man_t * pAig, Vec_Int_t * vOrder, DdManager * dd )
func (pAig *AigManager) Llb4Nonlin4SweepBadMonitor(vOrder *IntVector, dd *CuddFoundation) *DdNode {
	panic("TODO, stubbed")
}

// Llb_Nonlin4SweepVars2Q
// C: Vec_Int_t * Llb_Nonlin4SweepVars2Q( Aig_Man_t * pAig, Vec_Int_t * vOrder, int fAddLis )
func (pAig *AigManager) LlbNonlin4SweepVars2Q(vOrder *IntVector, fAddLis int) *IntVector {
	panic("TODO, stubbed")
}

// Llb_Nonlin4SweepDeref
// C: void Llb_Nonlin4SweepDeref( DdManager * dd, Vec_Ptr_t * vParts )
func LlbNonlin4SweepDeref(dd *CuddFoundation, vParts *PointerVector[DdNode]) {
	panic("TODO, stubbed")
}

// Llb_Nonlin4SweepPrint
// C: void Llb_Nonlin4SweepPrint( Vec_Ptr_t * vFuncs )
func LlbNonlin4SweepPrint(vFuncs *PointerVector[DdNode]) {
	panic("TODO, stubbed")
}

// Llb4_Nonlin4SweepBadStates
// C: DdManager * Llb4_Nonlin4SweepBadStates( Aig_Man_t * pAig, Vec_Int_t * vOrder, int nVars )
func (pAig *AigManager) Llb4Nonlin4SweepBadStates(vOrder *IntVector, nVars int) *CuddFoundation {
	panic("TODO, stubbed")
}

// Llb4_Nonlin4SweepGroups
// C: DdManager * Llb4_Nonlin4SweepGroups( Aig_Man_t * pAig, Vec_Int_t * vOrder, int nVars, Vec_Ptr_t ** pvGroups, int nBddLimitClp, int fVerbose )
func (pAig *AigManager) Llb4Nonlin4SweepGroups(vOrder *IntVector, nVars int, pvGroups **PointerVector[DdNode], nBddLimitClp int, fVerbose int) *CuddFoundation {
	panic("TODO, stubbed")
}

// Llb_Nonlin4SweepPrintSuppProfile
// C: void Llb_Nonlin4SweepPrintSuppProfile( DdManager * dd, Aig_Man_t * pAig, Vec_Int_t * vOrder, Vec_Ptr_t * vGroups, int fVerbose )
func LlbNonlin4SweepPrintSuppProfile(dd *CuddFoundation, pAig *AigManager, vOrder *IntVector, vGroups *PointerVector[DdNode], fVerbose int) {
	panic("TODO, stubbed")
}

// Llb4_Nonlin4Sweep
// C: void Llb4_Nonlin4Sweep( Aig_Man_t * pAig, int nSweepMax, int nClusterMax, DdManager ** pdd, Vec_Int_t ** pvOrder, Vec_Ptr_t ** pvGroups, int fVerbose )
func (pAig *AigManager) Llb4Nonlin4Sweep(nSweepMax int, nClusterMax int, pdd **CuddFoundation, pvOrder **IntVector, pvGroups **PointerVector[DdNode], fVerbose int) {
	panic("TODO, stubbed")
}

// Llb4_Nonlin4SweepExperiment
// C: void Llb4_Nonlin4SweepExperiment( Aig_Man_t * pAig )
func (pAig *AigManager) Llb4Nonlin4SweepExperiment() {
	panic("TODO, stubbed")
}
