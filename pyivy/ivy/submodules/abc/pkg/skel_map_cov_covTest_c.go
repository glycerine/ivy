package abc

// Hand-checked skeleton of src/map/cov/covTest.c.

// Abc_NtkCovTestSop is declared in cov.h but not defined in covTest.c.

// Abc_NodeDeriveCoverPro
// C: Min_Cube_t * Abc_NodeDeriveCoverPro( Min_Man_t * p, Min_Cube_t * pCover0, Min_Cube_t * pCover1 )
func (p *MinMan) AbcNodeDeriveCoverPro(pCover0 *MinCube, pCover1 *MinCube) *MinCube {
	panic("TODO, stubbed")
}

// Abc_NodeDeriveCoverSum
// C: Min_Cube_t * Abc_NodeDeriveCoverSum( Min_Man_t * p, Min_Cube_t * pCover0, Min_Cube_t * pCover1 )
func (p *MinMan) AbcNodeDeriveCoverSum(pCover0 *MinCube, pCover1 *MinCube) *MinCube {
	panic("TODO, stubbed")
}

// Abc_NodeDeriveSops
// C: int Abc_NodeDeriveSops( Min_Man_t * p, Abc_Obj_t * pRoot, Vec_Ptr_t * vSupp, Vec_Ptr_t * vNodes )
func (p *MinMan) AbcNodeDeriveSops(pRoot *ABCNetworkObject, vSupp *PointerVector[ABCNetworkObject], vNodes *PointerVector[ABCNetworkObject]) int {
	panic("TODO, stubbed")
}

// Abc_NtkTestSop
// C: void Abc_NtkTestSop( Abc_Ntk_t * pNtk )
func AbcNtkTestSop(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NodeDeriveCover
// C: Min_Cube_t * Abc_NodeDeriveCover( Min_Man_t * p, Min_Cube_t * pCov0, Min_Cube_t * pCov1, int fComp0, int fComp1 )
func (p *MinMan) AbcNodeDeriveCover(pCov0 *MinCube, pCov1 *MinCube, fComp0 int, fComp1 int) *MinCube {
	panic("TODO, stubbed")
}

// Abc_NodeDeriveEsops
// C: int Abc_NodeDeriveEsops( Min_Man_t * p, Abc_Obj_t * pRoot, Vec_Ptr_t * vSupp, Vec_Ptr_t * vNodes )
func (p *MinMan) AbcNodeDeriveEsops(pRoot *ABCNetworkObject, vSupp *PointerVector[ABCNetworkObject], vNodes *PointerVector[ABCNetworkObject]) int {
	panic("TODO, stubbed")
}

// Abc_NtkTestEsop
// C: void Abc_NtkTestEsop( Abc_Ntk_t * pNtk )
func AbcNtkTestEsop(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}
