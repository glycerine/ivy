package abc

// Hand-checked skeleton of src/base/wlc/wlcAbs2.c.

// Every function here is static and shares its name with a static in wlcAbs.c,
// so each carries an _abs2 suffix, except the non-static Wlc_NtkAbsCore2.

// Wlc_NtkAbsMarkOpers (static)
// C: Vec_Bit_t * Wlc_NtkAbsMarkOpers( Wlc_Ntk_t * p, Wlc_Par_t * pPars, Vec_Bit_t * vUnmark, int fVerbose )
func wlcNtkAbsMarkOpers_abs2(p *WLCNetwork, pPars *WlcPar, vUnmark *BitVector, fVerbose int) *BitVector {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsMarkNodes_rec (static)
// C: void Wlc_NtkAbsMarkNodes_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, Vec_Bit_t * vLeaves, Vec_Int_t * vPisOld, Vec_Int_t * vPisNew, Vec_Int_t * vFlops )
func wlcNtkAbsMarkNodes_rec_abs2(p *WLCNetwork, pObj *WLCObject, vLeaves *BitVector, vPisOld *IntVector, vPisNew *IntVector, vFlops *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsMarkNodes (static)
// C: void Wlc_NtkAbsMarkNodes( Wlc_Ntk_t * p, Vec_Bit_t * vLeaves, Vec_Int_t * vPisOld, Vec_Int_t * vPisNew, Vec_Int_t * vFlops )
func wlcNtkAbsMarkNodes_abs2(p *WLCNetwork, vLeaves *BitVector, vPisOld *IntVector, vPisNew *IntVector, vFlops *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkAbs (static; pvPisNew is an out parameter)
// C: Wlc_Ntk_t * Wlc_NtkAbs( Wlc_Ntk_t * p, Wlc_Par_t * pPars, Vec_Bit_t * vUnmark, Vec_Int_t ** pvPisNew, int fVerbose )
func wlcNtkAbs_abs2(p *WLCNetwork, pPars *WlcPar, vUnmark *BitVector, pvPisNew **IntVector, fVerbose int) *WLCNetwork {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsRefinement (static)
// C: Vec_Int_t * Wlc_NtkAbsRefinement( Wlc_Ntk_t * p, Gia_Man_t * pGia, Abc_Cex_t * pCex, Vec_Int_t * vPisNew )
func wlcNtkAbsRefinement_abs2(p *WLCNetwork, pGia *GiaManager, pCex *Counterexample, vPisNew *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkNodeDeref_rec (static)
// C: int Wlc_NtkNodeDeref_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pNode, Vec_Bit_t * vUnmark )
func wlcNtkNodeDeref_rec_abs2(p *WLCNetwork, pNode *WLCObject, vUnmark *BitVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkNodeRef_rec (static)
// C: int Wlc_NtkNodeRef_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pNode )
func wlcNtkNodeRef_rec_abs2(p *WLCNetwork, pNode *WLCObject) int {
	panic("TODO, stubbed")
}

// Wlc_NtkMarkMffc (static)
// C: int Wlc_NtkMarkMffc( Wlc_Ntk_t * p, Wlc_Obj_t * pNode, Vec_Bit_t * vUnmark )
func wlcNtkMarkMffc_abs2(p *WLCNetwork, pNode *WLCObject, vUnmark *BitVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkRemoveFromAbstraction (static)
// C: int Wlc_NtkRemoveFromAbstraction( Wlc_Ntk_t * p, Vec_Int_t * vRefine, Vec_Bit_t * vUnmark )
func wlcNtkRemoveFromAbstraction_abs2(p *WLCNetwork, vRefine *IntVector, vUnmark *BitVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsCore2
// C: int Wlc_NtkAbsCore2( Wlc_Ntk_t * p, Wlc_Par_t * pPars )
func WlcNtkAbsCore2(p *WLCNetwork, pPars *WlcPar) int {
	panic("TODO, stubbed")
}
