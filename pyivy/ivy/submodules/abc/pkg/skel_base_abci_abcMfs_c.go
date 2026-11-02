package abc

// Hand-checked skeleton of src/base/abci/abcMfs.c.

// Abc_NtkAssignIDs
// C: Vec_Ptr_t * Abc_NtkAssignIDs( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkAssignIDs() *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkAssignIDs2
// C: Vec_Ptr_t * Abc_NtkAssignIDs2( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkAssignIDs2() *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkAssignStarts
// C: Vec_Int_t * Abc_NtkAssignStarts( Abc_Ntk_t * pNtk, Vec_Ptr_t * vNodes, int * pnTotal )
func (pNtk *ABCNetwork) AbcNtkAssignStarts(vNodes *PointerVector[ABCNetworkObject], pnTotal *int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkFillTruthStore
// C: void Abc_NtkFillTruthStore( word TruthStore[16][1<<10] )
func AbcNtkFillTruthStore(TruthStore *[16][1 << 10]uint64) {
	panic("TODO, stubbed")
}

// Abc_NtkExtractMfs
// C: Sfm_Ntk_t * Abc_NtkExtractMfs( Abc_Ntk_t * pNtk, int nFirstFixed )
func (pNtk *ABCNetwork) AbcNtkExtractMfs(nFirstFixed int) *SfmNtk {
	panic("TODO, stubbed")
}

// Abc_NtkExtractMfs2
// C: Sfm_Ntk_t * Abc_NtkExtractMfs2( Abc_Ntk_t * pNtk, int iPivot )
func (pNtk *ABCNetwork) AbcNtkExtractMfs2(iPivot int) *SfmNtk {
	panic("TODO, stubbed")
}

// Abc_NtkInsertMfs
// C: void Abc_NtkInsertMfs( Abc_Ntk_t * pNtk, Sfm_Ntk_t * p )
func (pNtk *ABCNetwork) AbcNtkInsertMfs(p *SfmNtk) {
	panic("TODO, stubbed")
}

// Abc_NtkPerformMfs
// C: int Abc_NtkPerformMfs( Abc_Ntk_t * pNtk, Sfm_Par_t * pPars )
func (pNtk *ABCNetwork) AbcNtkPerformMfs(pPars *SfmPar) int {
	panic("TODO, stubbed")
}

// Abc_NtkUnrollAndDrop
// C: Abc_Ntk_t * Abc_NtkUnrollAndDrop( Abc_Ntk_t * p, int nFrames, int nFramesAdd, Vec_Int_t * vFlops, int * piPivot )
func (p *ABCNetwork) AbcNtkUnrollAndDrop(nFrames int, nFramesAdd int, vFlops *IntVector, piPivot *int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkReinsertNodes
// C: void Abc_NtkReinsertNodes( Abc_Ntk_t * p, Abc_Ntk_t * pNtk, int iPivot )
func (p *ABCNetwork) AbcNtkReinsertNodes(pNtk *ABCNetwork, iPivot int) {
	panic("TODO, stubbed")
}

// Abc_NtkMfsAfterICheck
// C: int Abc_NtkMfsAfterICheck( Abc_Ntk_t * p, int nFrames, int nFramesAdd, Vec_Int_t * vFlops, Sfm_Par_t * pPars )
func (p *ABCNetwork) AbcNtkMfsAfterICheck(nFrames int, nFramesAdd int, vFlops *IntVector, pPars *SfmPar) int {
	panic("TODO, stubbed")
}
