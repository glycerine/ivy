package abc

// Hand-checked skeleton of src/base/abci/abcAttach.c.

// Abc_NtkAttach
// C: int Abc_NtkAttach( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkAttach() int {
	panic("TODO, stubbed")
}

// Abc_NodeAttach (static)
// C: int Abc_NodeAttach( Abc_Obj_t * pNode, Mio_Gate_t ** ppGates, unsigned ** puTruthGates, int nGates, unsigned uTruths[][2] )
func abcNodeAttach(pNode *ABCNetworkObject, ppGates []*MioGate, puTruthGates [][]uint32, nGates int, uTruths [][2]uint32) int {
	panic("TODO, stubbed")
}

// Abc_AttachSetupTruthTables (static)
// C: void Abc_AttachSetupTruthTables( unsigned uTruths[][2] )
func abcAttachSetupTruthTables(uTruths [][2]uint32) {
	panic("TODO, stubbed")
}

// Abc_AttachComputeTruth (static)
// C: void Abc_AttachComputeTruth( char * pSop, unsigned uTruthsIn[][2], unsigned * uTruthRes )
func abcAttachComputeTruth(pSop string, uTruthsIn [][2]uint32, uTruthRes []uint32) {
	panic("TODO, stubbed")
}

// Abc_AttachFind (static)
// C: Mio_Gate_t * Abc_AttachFind( Mio_Gate_t ** ppGates, unsigned ** puTruthGates, int nGates, unsigned * uTruthNode, int * Perm )
func abcAttachFind(ppGates []*MioGate, puTruthGates [][]uint32, nGates int, uTruthNode []uint32, Perm []int) *MioGate {
	panic("TODO, stubbed")
}

// Abc_AttachCompare (static)
// C: int Abc_AttachCompare( unsigned ** puTruthGates, int nGates, unsigned * uTruthNode )
func abcAttachCompare(puTruthGates [][]uint32, nGates int, uTruthNode []uint32) int {
	panic("TODO, stubbed")
}

// Abc_TruthPermute (static)
// C: void Abc_TruthPermute( char * pPerm, int nVars, unsigned * uTruthNode, unsigned * uTruthPerm )
func abcTruthPermute(pPerm []int8, nVars int, uTruthNode []uint32, uTruthPerm []uint32) {
	panic("TODO, stubbed")
}
