package abc

// Hand-checked skeleton of src/base/abci/abcNtbdd.c.

// Abc_NtkDeriveFromBdd (dd0 and bFunc are void * that hold a DdManager * and a DdNode *)
// C: Abc_Ntk_t * Abc_NtkDeriveFromBdd( void * dd0, void * bFunc, char * pNamePo, Vec_Ptr_t * vNamesPi )
func AbcNtkDeriveFromBdd(dd0 *CuddFoundation, bFunc *DdNode, pNamePo string, vNamesPi *PointerVector[string]) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkBddToMuxes
// C: Abc_Ntk_t * Abc_NtkBddToMuxes( Abc_Ntk_t * pNtk, int fGlobal, int Limit )
func (pNtk *ABCNetwork) AbcNtkBddToMuxes(fGlobal int, Limit int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkBddToMuxesPerform (static)
// C: void Abc_NtkBddToMuxesPerform( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkNew )
func (pNtk *ABCNetwork) abcNtkBddToMuxesPerform(pNtkNew *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NodeBddToMuxes (static)
// C: Abc_Obj_t * Abc_NodeBddToMuxes( Abc_Obj_t * pNodeOld, Abc_Ntk_t * pNtkNew )
func abcNodeBddToMuxes(pNodeOld *ABCNetworkObject, pNtkNew *ABCNetwork) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NodeBddToMuxes_rec (static)
// C: Abc_Obj_t * Abc_NodeBddToMuxes_rec( DdManager * dd, DdNode * bFunc, Abc_Ntk_t * pNtkNew, st__table * tBdd2Node )
func abcNodeBddToMuxes_rec(dd *CuddFoundation, bFunc *DdNode, pNtkNew *ABCNetwork, tBdd2Node *StableTable[*DdNode, *ABCNetworkObject]) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkBddToMuxesPerformGlo (static)
// C: int Abc_NtkBddToMuxesPerformGlo( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkNew, int Limit )
func (pNtk *ABCNetwork) abcNtkBddToMuxesPerformGlo(pNtkNew *ABCNetwork, Limit int) int {
	panic("TODO, stubbed")
}

// Abc_NtkBuildGlobalBdds (returns a void * that holds a DdManager *)
// C: void * Abc_NtkBuildGlobalBdds( Abc_Ntk_t * pNtk, int nBddSizeMax, int fDropInternal, int fReorder, int fReverse, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkBuildGlobalBdds(nBddSizeMax int, fDropInternal int, fReorder int, fReverse int, fVerbose int) *CuddFoundation {
	panic("TODO, stubbed")
}

// Abc_NodeGlobalBdds_rec (static; pProgress is a ProgressBar *)
// C: DdNode * Abc_NodeGlobalBdds_rec( DdManager * dd, Abc_Obj_t * pNode, int nBddSizeMax, int fDropInternal, ProgressBar * pProgress, int * pCounter, int fVerbose )
func abcNodeGlobalBdds_rec(dd *CuddFoundation, pNode *ABCNetworkObject, nBddSizeMax int, fDropInternal int, pProgress *ExtraProgressBar, pCounter *int, fVerbose int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkFreeGlobalBdds (returns a void * that holds a DdManager *)
// C: void * Abc_NtkFreeGlobalBdds( Abc_Ntk_t * pNtk, int fFreeMan )
func (pNtk *ABCNetwork) AbcNtkFreeGlobalBdds(fFreeMan int) *CuddFoundation {
	panic("TODO, stubbed")
}

// Abc_NtkSizeOfGlobalBdds
// C: int Abc_NtkSizeOfGlobalBdds( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkSizeOfGlobalBdds() int {
	panic("TODO, stubbed")
}

// Abc_NtkSpacePercentage
// C: double Abc_NtkSpacePercentage( Abc_Obj_t * pNode )
func AbcNtkSpacePercentage(pNode *ABCNetworkObject) float64 {
	panic("TODO, stubbed")
}

// Abc_NtkBddImplicationTest
// C: void Abc_NtkBddImplicationTest()
func AbcNtkBddImplicationTest() {
	panic("TODO, stubbed")
}
