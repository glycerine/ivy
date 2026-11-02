package abc

// Hand-checked skeleton of src/opt/lpk/lpkCore.c.

// Abc_NtkDecFromTruth
// C: Abc_Ntk_t * Abc_NtkDecFromTruth( word * pTruth, int nVars, int nLutSize )
func AbcNtkDecFromTruth(pTruth []uint64, nVars int, nLutSize int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkLutMinDecompose
// C: Abc_Obj_t * Abc_NtkLutMinDecompose( Abc_Ntk_t * pNtk, Vec_Ptr_t * vLeaves, word * pTruth, int nLutSize, int Required )
func AbcNtkLutMinDecompose(pNtk *ABCNetwork, vLeaves *PointerVector[ABCNetworkObject], pTruth []uint64, nLutSize int, Required int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Lpk_IfManStart
// C: void Lpk_IfManStart( Lpk_Man_t * p )
func (p *LpkMan) LpkIfManStart() {
	panic("TODO, stubbed")
}

// Lpk_NodeHasChanged
// C: int Lpk_NodeHasChanged( Lpk_Man_t * p, int iNode )
func (p *LpkMan) LpkNodeHasChanged(iNode int) int {
	panic("TODO, stubbed")
}

// Lpk_ExploreCut
// C: int Lpk_ExploreCut( Lpk_Man_t * p, Lpk_Cut_t * pCut, Kit_DsdNtk_t * pNtk )
func (p *LpkMan) LpkExploreCut(pCut *LpkCut, pNtk *KitDsdNtk) int {
	panic("TODO, stubbed")
}

// Lpk_ResynthesizeNode
// C: int Lpk_ResynthesizeNode( Lpk_Man_t * p )
func (p *LpkMan) LpkResynthesizeNode() int {
	panic("TODO, stubbed")
}

// Lpk_ComputeSupports
// C: void Lpk_ComputeSupports( Lpk_Man_t * p, Lpk_Cut_t * pCut, unsigned * pTruth )
func (p *LpkMan) LpkComputeSupports(pCut *LpkCut, pTruth []uint32) {
	panic("TODO, stubbed")
}

// Lpk_ResynthesizeNodeNew
// C: int Lpk_ResynthesizeNodeNew( Lpk_Man_t * p )
func (p *LpkMan) LpkResynthesizeNodeNew() int {
	panic("TODO, stubbed")
}

// Lpk_Resynthesize
// C: int Lpk_Resynthesize( Abc_Ntk_t * pNtk, Lpk_Par_t * pPars )
func LpkResynthesize(pNtk *ABCNetwork, pPars *LpkPar) int {
	panic("TODO, stubbed")
}
