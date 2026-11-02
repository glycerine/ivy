package abc

// Hand-checked skeleton of src/opt/mfs/mfsCore.c.

// Abc_NtkMfsParsDefault
// C: void Abc_NtkMfsParsDefault( Mfs_Par_t * pPars )
func (pPars *MfsPar) AbcNtkMfsParsDefault() {
	panic("TODO, stubbed")
}

// Abc_WinNode
// C: int Abc_WinNode(Mfs_Man_t * p, Abc_Obj_t *pNode)
func (p *MfsMan) AbcWinNode(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsPowerResub
// C: void Abc_NtkMfsPowerResub( Mfs_Man_t * p, Mfs_Par_t * pPars)
func (p *MfsMan) AbcNtkMfsPowerResub(pPars *MfsPar) {
	panic("TODO, stubbed")
}

// Abc_NtkMfsResub
// C: int Abc_NtkMfsResub( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsResub(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsNode
// C: int Abc_NtkMfsNode( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsNode(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfs
// C: int Abc_NtkMfs( Abc_Ntk_t * pNtk, Mfs_Par_t * pPars )
func AbcNtkMfs(pNtk *ABCNetwork, pPars *MfsPar) int {
	panic("TODO, stubbed")
}
