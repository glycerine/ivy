package abc

// Hand-checked skeleton of src/base/abci/abcReach.c.

// Abc_NtkInitStateVarMap
// C: DdNode * Abc_NtkInitStateVarMap( DdManager * dd, Abc_Ntk_t * pNtk, int fVerbose )
func AbcNtkInitStateVarMap(dd *CuddFoundation, pNtk *ABCNetwork, fVerbose int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkCreatePartitions
// C: DdNode ** Abc_NtkCreatePartitions( DdManager * dd, Abc_Ntk_t * pNtk, int fReorder, int fVerbose )
func AbcNtkCreatePartitions(dd *CuddFoundation, pNtk *ABCNetwork, fReorder int, fVerbose int) []*DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkComputeReachable
// C: DdNode * Abc_NtkComputeReachable( DdManager * dd, Abc_Ntk_t * pNtk, DdNode ** pbParts, DdNode * bInitial, DdNode * bOutput, int nBddMax, int nIterMax, int fPartition, int fReorder, int fVerbose )
func AbcNtkComputeReachable(dd *CuddFoundation, pNtk *ABCNetwork, pbParts []*DdNode, bInitial *DdNode, bOutput *DdNode, nBddMax int, nIterMax int, fPartition int, fReorder int, fVerbose int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkVerifyUsingBdds
// C: void Abc_NtkVerifyUsingBdds( Abc_Ntk_t * pNtk, int nBddMax, int nIterMax, int fPartition, int fReorder, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkVerifyUsingBdds(nBddMax int, nIterMax int, fPartition int, fReorder int, fVerbose int) {
	panic("TODO, stubbed")
}
