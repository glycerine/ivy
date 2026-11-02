package abc

// Hand-checked skeleton of src/proof/ssw/sswDyn.c.

// Ssw_ManLabelPiNodes
// C: void Ssw_ManLabelPiNodes( Ssw_Man_t * p )
func (p *SswMan) SswManLabelPiNodes() {
	panic("TODO, stubbed")
}

// Ssw_ManCollectPis_rec
// C: void Ssw_ManCollectPis_rec( Aig_Obj_t * pObj, Vec_Ptr_t * vNewPis )
func SswManCollectPis_rec(pObj *AigObject, vNewPis *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Ssw_ManCollectPos_rec
// C: void Ssw_ManCollectPos_rec( Ssw_Man_t * p, Aig_Obj_t * pObj, Vec_Int_t * vNewPos )
func (p *SswMan) SswManCollectPos_rec(pObj *AigObject, vNewPos *IntVector) {
	panic("TODO, stubbed")
}

// Ssw_ManLoadSolver
// C: void Ssw_ManLoadSolver( Ssw_Man_t * p, Aig_Obj_t * pRepr, Aig_Obj_t * pObj )
func (p *SswMan) SswManLoadSolver(pRepr *AigObject, pObj *AigObject) {
	panic("TODO, stubbed")
}

// Ssw_ManSweepTransferDyn
// C: void Ssw_ManSweepTransferDyn( Ssw_Man_t * p )
func (p *SswMan) SswManSweepTransferDyn() {
	panic("TODO, stubbed")
}

// Ssw_ManSweepResimulateDyn
// C: int Ssw_ManSweepResimulateDyn( Ssw_Man_t * p, int f )
func (p *SswMan) SswManSweepResimulateDyn(f int) int {
	panic("TODO, stubbed")
}

// Ssw_ManSweepResimulateDynLocal
// C: int Ssw_ManSweepResimulateDynLocal( Ssw_Man_t * p, int f )
func (p *SswMan) SswManSweepResimulateDynLocal(f int) int {
	panic("TODO, stubbed")
}

// Ssw_ManSweepDyn
// C: int Ssw_ManSweepDyn( Ssw_Man_t * p )
func (p *SswMan) SswManSweepDyn() int {
	panic("TODO, stubbed")
}
