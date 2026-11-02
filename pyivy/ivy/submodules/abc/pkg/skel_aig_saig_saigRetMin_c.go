package abc

// Hand-checked skeleton of src/aig/saig/saigRetMin.c.

// Saig_ManRetimeInitState
// C: Vec_Int_t * Saig_ManRetimeInitState( Aig_Man_t * p )
func (p *AigManager) SaigManRetimeInitState() *IntVector {
	panic("TODO, stubbed")
}

// Saig_ManRetimeUnsatCore
// C: int Saig_ManRetimeUnsatCore( Aig_Man_t * p, int fVerbose )
func (p *AigManager) SaigManRetimeUnsatCore(fVerbose int) int {
	panic("TODO, stubbed")
}

// Saig_ManMarkCone_rec
// C: void Saig_ManMarkCone_rec( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) SaigManMarkCone_rec(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Saig_ManRetimeCountCut
// C: int Saig_ManRetimeCountCut( Aig_Man_t * p, Vec_Ptr_t * vCut )
func (p *AigManager) SaigManRetimeCountCut(vCut *PointerVector[AigObject]) int {
	panic("TODO, stubbed")
}

// Saig_ManRetimeDup_rec
// C: void Saig_ManRetimeDup_rec( Aig_Man_t * pNew, Aig_Obj_t * pObj )
func (pNew *AigManager) SaigManRetimeDup_rec(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Saig_ManRetimeDupForward
// C: Aig_Man_t * Saig_ManRetimeDupForward( Aig_Man_t * p, Vec_Ptr_t * vCut )
func (p *AigManager) SaigManRetimeDupForward(vCut *PointerVector[AigObject]) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManRetimeDupBackward
// C: Aig_Man_t * Saig_ManRetimeDupBackward( Aig_Man_t * p, Vec_Ptr_t * vCut, Vec_Int_t * vInit )
func (p *AigManager) SaigManRetimeDupBackward(vCut *PointerVector[AigObject], vInit *IntVector) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManRetimeDupInitState
// C: Aig_Man_t * Saig_ManRetimeDupInitState( Aig_Man_t * p, Vec_Ptr_t * vCut )
func (p *AigManager) SaigManRetimeDupInitState(vCut *PointerVector[AigObject]) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManGetRegistersToExclude
// C: Vec_Ptr_t * Saig_ManGetRegistersToExclude( Aig_Man_t * p )
func (p *AigManager) SaigManGetRegistersToExclude() *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Saig_ManHideBadRegs
// C: int Saig_ManHideBadRegs( Aig_Man_t * p, Vec_Ptr_t * vBadRegs )
func (p *AigManager) SaigManHideBadRegs(vBadRegs *PointerVector[AigObject]) int {
	panic("TODO, stubbed")
}

// Saig_ManExposeBadRegs
// C: void Saig_ManExposeBadRegs( Aig_Man_t * p, int nBadRegs )
func (p *AigManager) SaigManExposeBadRegs(nBadRegs int) {
	panic("TODO, stubbed")
}

// Saig_ManRetimeMinAreaBackward
// C: Aig_Man_t * Saig_ManRetimeMinAreaBackward( Aig_Man_t * pNew, int fVerbose )
func (pNew *AigManager) SaigManRetimeMinAreaBackward(fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManRetimeMinArea
// C: Aig_Man_t * Saig_ManRetimeMinArea( Aig_Man_t * p, int nMaxIters, int fForwardOnly, int fBackwardOnly, int fInitial, int fVerbose )
func (p *AigManager) SaigManRetimeMinArea(nMaxIters int, fForwardOnly int, fBackwardOnly int, fInitial int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}
