package abc

// Hand-checked skeleton of src/opt/lpk/lpkAbcDec.c.

// Lpk_ImplementFun
// C: Abc_Obj_t * Lpk_ImplementFun( Lpk_Man_t * pMan, Abc_Ntk_t * pNtk, Vec_Ptr_t * vLeaves, Lpk_Fun_t * p )
func (pMan *LpkMan) LpkImplementFun(pNtk *ABCNetwork, vLeaves *PointerVector[ABCNetworkObject], p *LpkFun) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Lpk_Implement_rec
// C: Abc_Obj_t * Lpk_Implement_rec( Lpk_Man_t * pMan, Abc_Ntk_t * pNtk, Vec_Ptr_t * vLeaves, Lpk_Fun_t * pFun )
func (pMan *LpkMan) LpkImplement_rec(pNtk *ABCNetwork, vLeaves *PointerVector[ABCNetworkObject], pFun *LpkFun) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Lpk_Implement
// C: Abc_Obj_t * Lpk_Implement( Lpk_Man_t * pMan, Abc_Ntk_t * pNtk, Vec_Ptr_t * vLeaves, int nLeavesOld )
func (pMan *LpkMan) LpkImplement(pNtk *ABCNetwork, vLeaves *PointerVector[ABCNetworkObject], nLeavesOld int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Lpk_Decompose_rec
// C: int Lpk_Decompose_rec( Lpk_Man_t * pMan, Lpk_Fun_t * p )
func (pMan *LpkMan) LpkDecompose_rec(p *LpkFun) int {
	panic("TODO, stubbed")
}

// Lpk_DecomposeClean
// C: void Lpk_DecomposeClean( Vec_Ptr_t * vLeaves, int nLeavesOld )
func LpkDecomposeClean(vLeaves *PointerVector[ABCNetworkObject], nLeavesOld int) {
	panic("TODO, stubbed")
}

// Lpk_Decompose
// C: Abc_Obj_t * Lpk_Decompose( Lpk_Man_t * p, Abc_Ntk_t * pNtk, Vec_Ptr_t * vLeaves, unsigned * pTruth, unsigned * puSupps, int nLutK, int AreaLim, int DelayLim )
func (p *LpkMan) LpkDecompose(pNtk *ABCNetwork, vLeaves *PointerVector[ABCNetworkObject], pTruth []uint32, puSupps []uint32, nLutK int, AreaLim int, DelayLim int) *ABCNetworkObject {
	panic("TODO, stubbed")
}
