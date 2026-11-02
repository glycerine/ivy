package abc

// Hand-checked skeleton of src/base/abci/abcQbf.c.

// Abc_NtkQbf
// C: void Abc_NtkQbf( Abc_Ntk_t * pNtk, int nPars, int nItersMax, int fDumpCnf, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkQbf(nPars int, nItersMax int, fDumpCnf int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkModelToVector (static)
// C: void Abc_NtkModelToVector( Abc_Ntk_t * pNtk, Vec_Int_t * vPiValues )
func (pNtk *ABCNetwork) abcNtkModelToVector(vPiValues *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkVectorClearPars (static)
// C: void Abc_NtkVectorClearPars( Vec_Int_t * vPiValues, int nPars )
func abcNtkVectorClearPars(vPiValues *IntVector, nPars int) {
	panic("TODO, stubbed")
}

// Abc_NtkVectorClearVars (static)
// C: void Abc_NtkVectorClearVars( Abc_Ntk_t * pNtk, Vec_Int_t * vPiValues, int nPars )
func (pNtk *ABCNetwork) abcNtkVectorClearVars(vPiValues *IntVector, nPars int) {
	panic("TODO, stubbed")
}

// Abc_NtkVectorPrintPars (static)
// C: void Abc_NtkVectorPrintPars( Vec_Int_t * vPiValues, int nPars )
func abcNtkVectorPrintPars(vPiValues *IntVector, nPars int) {
	panic("TODO, stubbed")
}

// Abc_NtkVectorPrintVars (static)
// C: void Abc_NtkVectorPrintVars( Abc_Ntk_t * pNtk, Vec_Int_t * vPiValues, int nPars )
func (pNtk *ABCNetwork) abcNtkVectorPrintVars(vPiValues *IntVector, nPars int) {
	panic("TODO, stubbed")
}
