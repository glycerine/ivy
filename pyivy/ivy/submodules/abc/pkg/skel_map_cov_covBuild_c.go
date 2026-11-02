package abc

// Hand-checked skeleton of src/map/cov/covBuild.c.

// Abc_NtkCovDeriveCube
// C: Abc_Obj_t * Abc_NtkCovDeriveCube( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pObj, Min_Cube_t * pCube, Vec_Int_t * vSupp, int fCompl )
func AbcNtkCovDeriveCube(pNtkNew *ABCNetwork, pObj *ABCNetworkObject, pCube *MinCube, vSupp *IntVector, fCompl int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkCovDeriveNode_rec
// C: Abc_Obj_t * Abc_NtkCovDeriveNode_rec( Cov_Man_t * p, Abc_Ntk_t * pNtkNew, Abc_Obj_t * pObj, int Level )
func (p *CovMan) AbcNtkCovDeriveNode_rec(pNtkNew *ABCNetwork, pObj *ABCNetworkObject, Level int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkCovDerive
// C: Abc_Ntk_t * Abc_NtkCovDerive( Cov_Man_t * p, Abc_Ntk_t * pNtk )
func (p *CovMan) AbcNtkCovDerive(pNtk *ABCNetwork) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkCovDeriveInv
// C: Abc_Obj_t * Abc_NtkCovDeriveInv( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pObj, int fCompl )
func AbcNtkCovDeriveInv(pNtkNew *ABCNetwork, pObj *ABCNetworkObject, fCompl int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkCovDeriveCubeInv
// C: Abc_Obj_t * Abc_NtkCovDeriveCubeInv( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pObj, Min_Cube_t * pCube, Vec_Int_t * vSupp )
func AbcNtkCovDeriveCubeInv(pNtkNew *ABCNetwork, pObj *ABCNetworkObject, pCube *MinCube, vSupp *IntVector) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkCovDeriveNodeInv_rec
// C: Abc_Obj_t * Abc_NtkCovDeriveNodeInv_rec( Cov_Man_t * p, Abc_Ntk_t * pNtkNew, Abc_Obj_t * pObj, int fCompl )
func (p *CovMan) AbcNtkCovDeriveNodeInv_rec(pNtkNew *ABCNetwork, pObj *ABCNetworkObject, fCompl int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkCovDeriveClean
// C: Abc_Ntk_t * Abc_NtkCovDeriveClean( Cov_Man_t * p, Abc_Ntk_t * pNtk )
func (p *CovMan) AbcNtkCovDeriveClean(pNtk *ABCNetwork) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkCovDerive_rec
// C: Abc_Obj_t * Abc_NtkCovDerive_rec( Cov_Man_t * p, Abc_Ntk_t * pNtkNew, Abc_Obj_t * pObj )
func (p *CovMan) AbcNtkCovDerive_rec(pNtkNew *ABCNetwork, pObj *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkCovDeriveRegular
// C: Abc_Ntk_t * Abc_NtkCovDeriveRegular( Cov_Man_t * p, Abc_Ntk_t * pNtk )
func (p *CovMan) AbcNtkCovDeriveRegular(pNtk *ABCNetwork) *ABCNetwork {
	panic("TODO, stubbed")
}
