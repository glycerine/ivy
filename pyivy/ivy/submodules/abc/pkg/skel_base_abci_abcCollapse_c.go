package abc

// Hand-checked skeleton of src/base/abci/abcCollapse.c.

// Abc_NodeMinimumBase2
// C: int Abc_NodeMinimumBase2( Abc_Obj_t * pNode )
func AbcNodeMinimumBase2(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMinimumBase2
// C: int Abc_NtkMinimumBase2( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkMinimumBase2() int {
	panic("TODO, stubbed")
}

// Abc_NodeFromGlobalBdds
// C: Abc_Obj_t * Abc_NodeFromGlobalBdds( Abc_Ntk_t * pNtkNew, DdManager * dd, DdNode * bFunc, int fReverse )
func (pNtkNew *ABCNetwork) AbcNodeFromGlobalBdds(dd *CuddFoundation, bFunc *DdNode, fReverse int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkFromGlobalBdds
// C: Abc_Ntk_t * Abc_NtkFromGlobalBdds( Abc_Ntk_t * pNtk, int fReverse )
func (pNtk *ABCNetwork) AbcNtkFromGlobalBdds(fReverse int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDumpVariableOrder
// C: void Abc_NtkDumpVariableOrder( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkDumpVariableOrder() {
	panic("TODO, stubbed")
}

// Abc_NtkCollapse
// C: Abc_Ntk_t * Abc_NtkCollapse( Abc_Ntk_t * pNtk, int fBddSizeMax, int fDualRail, int fReorder, int fReverse, int fDumpOrder, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkCollapse(fBddSizeMax int, fDualRail int, fReorder int, fReverse int, fDumpOrder int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkClpGia_rec
// C: int Abc_NtkClpGia_rec( Gia_Man_t * pNew, Abc_Obj_t * pNode )
func (pNew *GiaManager) AbcNtkClpGia_rec(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkClpGia
// C: Gia_Man_t * Abc_NtkClpGia( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkClpGia() *GiaManager {
	panic("TODO, stubbed")
}

// Abc_NtkCollapseReduce
// C: int Abc_NtkCollapseReduce( Vec_Str_t * vSop, Vec_Int_t * vSupp, Vec_Int_t * vClass, Vec_Wec_t * vSupps )
func AbcNtkCollapseReduce(vSop *ByteVector, vSupp *IntVector, vClass *IntVector, vSupps *IntLevelVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkClpDeriveSatSolver
// C: sat_solver * Abc_NtkClpDeriveSatSolver( Cnf_Dat_t * pCnf, int iCoObjId, Vec_Int_t * vSupp, Vec_Int_t * vAnds, Vec_Int_t * vMap, sat_solver ** ppSat1, sat_solver ** ppSat2, sat_solver ** ppSat3 )
func (pCnf *CnfDat) AbcNtkClpDeriveSatSolver(iCoObjId int, vSupp *IntVector, vAnds *IntVector, vMap *IntVector, ppSat1 **SatSolver, ppSat2 **SatSolver, ppSat3 **SatSolver) *SatSolver {
	panic("TODO, stubbed")
}

// Abc_NtkClpGiaOne
// C: Vec_Str_t * Abc_NtkClpGiaOne( Gia_Man_t * p, int iCo, int nCubeLim, int nBTLimit, int fCanon, int fReverse, Vec_Int_t * vSupp, int fVerbose, Vec_Int_t * vClass, Vec_Wec_t * vSupps )
func (p *GiaManager) AbcNtkClpGiaOne(iCo int, nCubeLim int, nBTLimit int, fCanon int, fReverse int, vSupp *IntVector, fVerbose int, vClass *IntVector, vSupps *IntLevelVector) *ByteVector {
	panic("TODO, stubbed")
}

// Abc_NtkClpGiaOne2
// C: Vec_Str_t * Abc_NtkClpGiaOne2( Cnf_Dat_t * pCnf, Gia_Man_t * p, int iCo, int nCubeLim, int nBTLimit, int fCanon, int fReverse, Vec_Int_t * vSupp, Vec_Int_t * vMap, int fVerbose, Vec_Int_t * vClass, Vec_Wec_t * vSupps )
func (pCnf *CnfDat) AbcNtkClpGiaOne2(p *GiaManager, iCo int, nCubeLim int, nBTLimit int, fCanon int, fReverse int, vSupp *IntVector, vMap *IntVector, fVerbose int, vClass *IntVector, vSupps *IntLevelVector) *ByteVector {
	panic("TODO, stubbed")
}

// Abc_GiaDeriveSops
// C: Vec_Ptr_t * Abc_GiaDeriveSops( Abc_Ntk_t * pNtkNew, Gia_Man_t * p, Vec_Wec_t * vSupps, int nCubeLim, int nBTLimit, int nCostMax, int fCanon, int fReverse, int fCnfShared, int fVerbose )
func (pNtkNew *ABCNetwork) AbcGiaDeriveSops(p *GiaManager, vSupps *IntLevelVector, nCubeLim int, nBTLimit int, nCostMax int, fCanon int, fReverse int, fCnfShared int, fVerbose int) *PointerVector[string] {
	panic("TODO, stubbed")
}

// Abc_NtkFromSopsInt
// C: Abc_Ntk_t * Abc_NtkFromSopsInt( Abc_Ntk_t * pNtk, int nCubeLim, int nBTLimit, int nCostMax, int fCanon, int fReverse, int fCnfShared, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkFromSopsInt(nCubeLim int, nBTLimit int, nCostMax int, fCanon int, fReverse int, fCnfShared int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkCollapseSat
// C: Abc_Ntk_t * Abc_NtkCollapseSat( Abc_Ntk_t * pNtk, int nCubeLim, int nBTLimit, int nCostMax, int fCanon, int fReverse, int fCnfShared, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkCollapseSat(nCubeLim int, nBTLimit int, nCostMax int, fCanon int, fReverse int, fCnfShared int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}
