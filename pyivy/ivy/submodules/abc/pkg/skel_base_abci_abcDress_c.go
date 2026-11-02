package abc

// Hand-checked skeleton of src/base/abci/abcDress.c.

// Abc_NtkDress
// C: void Abc_NtkDress( Abc_Ntk_t * pNtkLogic, char * pFileName, int fVerbose )
func (pNtkLogic *ABCNetwork) AbcNtkDress(pFileName string, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkDressDeriveMapping (static; stmm_table * mapping names)
// C: stmm_table * Abc_NtkDressDeriveMapping( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkDressDeriveMapping() *StableTable[string, string] {
	panic("TODO, stubbed")
}

// Abc_NtkDressTransferNames (static)
// C: void Abc_NtkDressTransferNames( Abc_Ntk_t * pNtk, stmm_table * tMapping, int fVerbose )
func (pNtk *ABCNetwork) abcNtkDressTransferNames(tMapping *StableTable[string, string], fVerbose int) {
	panic("TODO, stubbed")
}
