package abc

// Hand-checked skeleton of src/base/abci/abcUnreach.c.

// Abc_NtkExtractSequentialDcs
// C: int Abc_NtkExtractSequentialDcs( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkExtractSequentialDcs(fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkTransitionRelation (static)
// C: DdNode * Abc_NtkTransitionRelation( DdManager * dd, Abc_Ntk_t * pNtk, int fVerbose )
func abcNtkTransitionRelation(dd *CuddFoundation, pNtk *ABCNetwork, fVerbose int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkInitStateAndVarMap (static)
// C: DdNode * Abc_NtkInitStateAndVarMap( DdManager * dd, Abc_Ntk_t * pNtk, int fVerbose )
func abcNtkInitStateAndVarMap(dd *CuddFoundation, pNtk *ABCNetwork, fVerbose int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkComputeUnreachable (static)
// C: DdNode * Abc_NtkComputeUnreachable( DdManager * dd, Abc_Ntk_t * pNtk, DdNode * bTrans, DdNode * bInitial, int fVerbose )
func abcNtkComputeUnreachable(dd *CuddFoundation, pNtk *ABCNetwork, bTrans *DdNode, bInitial *DdNode, fVerbose int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkConstructExdc (static)
// C: Abc_Ntk_t * Abc_NtkConstructExdc( DdManager * dd, Abc_Ntk_t * pNtk, DdNode * bUnreach )
func abcNtkConstructExdc(dd *CuddFoundation, pNtk *ABCNetwork, bUnreach *DdNode) *ABCNetwork {
	panic("TODO, stubbed")
}
