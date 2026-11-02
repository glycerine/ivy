package abc

// Hand-checked skeleton of src/map/cov/covMan.c.

// Cov_ManAlloc
// C: Cov_Man_t * Cov_ManAlloc( Abc_Ntk_t * pNtk, int nFaninMax, int nCubesMax )
func CovManAlloc(pNtk *ABCNetwork, nFaninMax int, nCubesMax int) *CovMan {
	panic("TODO, stubbed")
}

// Cov_ManFree
// C: void Cov_ManFree( Cov_Man_t * p )
func (p *CovMan) CovManFree() {
	panic("TODO, stubbed")
}

// Abc_NodeCovDropData
// C: void Abc_NodeCovDropData( Cov_Man_t * p, Abc_Obj_t * pObj )
func (p *CovMan) AbcNodeCovDropData(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}
