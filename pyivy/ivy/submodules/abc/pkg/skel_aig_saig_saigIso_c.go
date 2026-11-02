package abc

// Hand-checked skeleton of src/aig/saig/saigIso.c.

// Saig_ManFindIsoPermCos
// C: Vec_Int_t * Saig_ManFindIsoPermCos( Aig_Man_t * pAig, Vec_Int_t * vPermCis )
func (pAig *AigManager) SaigManFindIsoPermCos(vPermCis *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Saig_ManDupIsoCanonical_rec
// C: void Saig_ManDupIsoCanonical_rec( Aig_Man_t * pNew, Aig_Man_t * pAig, Aig_Obj_t * pObj )
func (pNew *AigManager) SaigManDupIsoCanonical_rec(pAig *AigManager, pObj *AigObject) {
	panic("TODO, stubbed")
}

// Saig_ManDupIsoCanonical
// C: Aig_Man_t * Saig_ManDupIsoCanonical( Aig_Man_t * pAig, int fVerbose )
func (pAig *AigManager) SaigManDupIsoCanonical(fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Iso_ManCheckMapping
// C: int Iso_ManCheckMapping( Aig_Man_t * pAig1, Aig_Man_t * pAig2, Vec_Int_t * vMap2to1, int fVerbose )
func (pAig1 *AigManager) IsoManCheckMapping(pAig2 *AigManager, vMap2to1 *IntVector, fVerbose int) int {
	panic("TODO, stubbed")
}

// Iso_ManNegEdgeNum
// C: int Iso_ManNegEdgeNum( Aig_Man_t * pAig )
func (pAig *AigManager) IsoManNegEdgeNum() int {
	panic("TODO, stubbed")
}

// Iso_ManFindMapping
// C: Vec_Int_t * Iso_ManFindMapping( Aig_Man_t * pAig1, Aig_Man_t * pAig2, Vec_Int_t * vPerm1_, Vec_Int_t * vPerm2_, int fVerbose )
func (pAig1 *AigManager) IsoManFindMapping(pAig2 *AigManager, vPerm1_ *IntVector, vPerm2_ *IntVector, fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Iso_ManFilterPos_old
// C: Aig_Man_t * Iso_ManFilterPos_old( Aig_Man_t * pAig, int fVerbose )
func (pAig *AigManager) IsoManFilterPos_old(fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Iso_StoCompareVecStr
// C: int Iso_StoCompareVecStr( Vec_Str_t ** p1, Vec_Str_t ** p2 )
func IsoStoCompareVecStr(p1 **ByteVector, p2 **ByteVector) int {
	panic("TODO, stubbed")
}

// Iso_ManFilterPos
// C: Aig_Man_t * Iso_ManFilterPos( Aig_Man_t * pAig, Vec_Ptr_t ** pvPosEquivs, int fVerbose )
func (pAig *AigManager) IsoManFilterPos(pvPosEquivs **PointerVector[IntVector], fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Iso_ManTest
// C: Aig_Man_t * Iso_ManTest( Aig_Man_t * pAig, int fVerbose )
func (pAig *AigManager) IsoManTest(fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManIsoReduce
// C: Aig_Man_t * Saig_ManIsoReduce( Aig_Man_t * pAig, Vec_Ptr_t ** pvPosEquivs, int fVerbose )
func (pAig *AigManager) SaigManIsoReduce(pvPosEquivs **PointerVector[IntVector], fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Iso_ManTest888
// C: Aig_Man_t * Iso_ManTest888( Aig_Man_t * pAig1, int fVerbose )
func (pAig1 *AigManager) IsoManTest888(fVerbose int) *AigManager {
	panic("TODO, stubbed")
}
