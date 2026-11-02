package abc

// Hand-checked skeleton of src/map/mpm/mpmGates.c.

// Mpm_ManFindDsdMatches (pScl is a void * that holds an SC_Lib *)
// C: Vec_Wec_t * Mpm_ManFindDsdMatches( Mpm_Man_t * p, void * pScl )
func (p *MpmMan) MpmManFindDsdMatches(pScl *SCLib) *IntLevelVector {
	panic("TODO, stubbed")
}

// Mpm_ManFindCells
// C: Vec_Ptr_t * Mpm_ManFindCells( Mio_Library_t * pMio, SC_Lib * pScl, Vec_Wec_t * vNpnConfigs )
func (pMio *MioLibrary) MpmManFindCells(pScl *SCLib, vNpnConfigs *IntLevelVector) *PointerVector[MioGate] {
	panic("TODO, stubbed")
}

// Mpm_ManFindMappedNodes
// C: Vec_Int_t * Mpm_ManFindMappedNodes( Mpm_Man_t * p )
func (p *MpmMan) MpmManFindMappedNodes() *IntVector {
	panic("TODO, stubbed")
}

// Mpm_ManGetAbcNode
// C: Abc_Obj_t * Mpm_ManGetAbcNode( Abc_Ntk_t * pNtk, Vec_Int_t * vCopy, int iMigLit )
func MpmManGetAbcNode(pNtk *ABCNetwork, vCopy *IntVector, iMigLit int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Mpm_ManDeriveMappedAbcNtk
// C: Abc_Ntk_t * Mpm_ManDeriveMappedAbcNtk( Mpm_Man_t * p, Mio_Library_t * pMio )
func (p *MpmMan) MpmManDeriveMappedAbcNtk(pMio *MioLibrary) *ABCNetwork {
	panic("TODO, stubbed")
}

// Mpm_ManPerformCellMapping
// C: Abc_Ntk_t * Mpm_ManPerformCellMapping( Mig_Man_t * pMig, Mpm_Par_t * pPars, Mio_Library_t * pMio )
func (pMig *MigMan) MpmManPerformCellMapping(pPars *MpmPar, pMio *MioLibrary) *ABCNetwork {
	panic("TODO, stubbed")
}

// Mpm_ManCellMapping (pMio is a void * that holds a Mio_Library_t *)
// C: Abc_Ntk_t * Mpm_ManCellMapping( Gia_Man_t * pGia, Mpm_Par_t * pPars, void * pMio )
func (pGia *GiaManager) MpmManCellMapping(pPars *MpmPar, pMio *MioLibrary) *ABCNetwork {
	panic("TODO, stubbed")
}
