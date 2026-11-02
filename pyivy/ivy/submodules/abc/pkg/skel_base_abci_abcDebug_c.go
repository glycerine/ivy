package abc

// Hand-checked skeleton of src/base/abci/abcDebug.c.

// Abc_NtkAutoDebug
// C: void Abc_NtkAutoDebug( Abc_Ntk_t * pNtk, int (*pFuncError) (Abc_Ntk_t *) )
func (pNtk *ABCNetwork) AbcNtkAutoDebug(pFuncError func(*ABCNetwork) int) {
	panic("TODO, stubbed")
}

// Abc_NtkCountFaninsTotal (static)
// C: int Abc_NtkCountFaninsTotal( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkCountFaninsTotal() int {
	panic("TODO, stubbed")
}

// Abc_NtkFindGivenFanin
// C: int Abc_NtkFindGivenFanin( Abc_Ntk_t * pNtk, int Step, Abc_Obj_t ** ppObj, Abc_Obj_t ** ppFanin )
func (pNtk *ABCNetwork) AbcNtkFindGivenFanin(Step int, ppObj **ABCNetworkObject, ppFanin **ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkAutoDebugModify (static)
// C: Abc_Ntk_t * Abc_NtkAutoDebugModify( Abc_Ntk_t * pNtkInit, int Step, int fConst1 )
func (pNtkInit *ABCNetwork) abcNtkAutoDebugModify(Step int, fConst1 int) *ABCNetwork {
	panic("TODO, stubbed")
}
