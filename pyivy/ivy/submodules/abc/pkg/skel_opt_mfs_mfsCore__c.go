package abc

// Hand-checked skeleton of src/opt/mfs/mfsCore_.c.

// A stale alternate copy of mfsCore.c that defines the same function names;
// every name here carries a trailing underscore to stay distinct.

// Abc_NtkMfsParsDefault (alternate copy; see mfsCore.c)
// C: void Abc_NtkMfsParsDefault( Mfs_Par_t * pPars )
func (pPars *MfsPar) AbcNtkMfsParsDefault_() {
	panic("TODO, stubbed")
}

// Abc_NtkMfsResub (alternate copy; see mfsCore.c)
// C: int Abc_NtkMfsResub( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsResub_(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsNode (alternate copy; see mfsCore.c)
// C: int Abc_NtkMfsNode( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsNode_(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfs (alternate copy; see mfsCore.c)
// C: int Abc_NtkMfs( Abc_Ntk_t * pNtk, Mfs_Par_t * pPars )
func AbcNtkMfs_(pNtk *ABCNetwork, pPars *MfsPar) int {
	panic("TODO, stubbed")
}
