package abc

// Hand-checked skeleton of src/proof/ssw/sswUnique.c.

// Ssw_UniqueRegisterPairInfo
// C: void Ssw_UniqueRegisterPairInfo( Ssw_Man_t * p )
func (p *SswMan) SswUniqueRegisterPairInfo() {
	panic("TODO, stubbed")
}

// Ssw_ManUniqueOne
// C: int Ssw_ManUniqueOne( Ssw_Man_t * p, Aig_Obj_t * pRepr, Aig_Obj_t * pObj, int fVerbose )
func (p *SswMan) SswManUniqueOne(pRepr *AigObject, pObj *AigObject, fVerbose int) int {
	panic("TODO, stubbed")
}

// Ssw_ManUniqueAddConstraint
// C: int Ssw_ManUniqueAddConstraint( Ssw_Man_t * p, Vec_Ptr_t * vCommon, int f1, int f2 )
func (p *SswMan) SswManUniqueAddConstraint(vCommon *PointerVector[AigObject], f1 int, f2 int) int {
	panic("TODO, stubbed")
}
