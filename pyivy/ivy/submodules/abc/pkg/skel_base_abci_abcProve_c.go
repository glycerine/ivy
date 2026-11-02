package abc

// Hand-checked skeleton of src/base/abci/abcProve.c.

// Abc_NtkMiterProve (pPars is a void * in C that holds a Prove_Params_t *)
// C: int Abc_NtkMiterProve( Abc_Ntk_t ** ppNtk, void * pPars )
func AbcNtkMiterProve(ppNtk **ABCNetwork, pPars *ProveParams) int {
	panic("TODO, stubbed")
}

// Abc_NtkMiterFraig (static)
// C: Abc_Ntk_t * Abc_NtkMiterFraig( Abc_Ntk_t * pNtk, int nBTLimit, ABC_INT64_T nInspLimit, int * pRetValue, int * pNumFails, ABC_INT64_T * pNumConfs, ABC_INT64_T * pNumInspects )
func (pNtk *ABCNetwork) abcNtkMiterFraig(nBTLimit int, nInspLimit int64, pRetValue *int, pNumFails *int, pNumConfs *int64, pNumInspects *int64) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkMiterPrint (static)
// C: void Abc_NtkMiterPrint( Abc_Ntk_t * pNtk, char * pString, abctime clk, int fVerbose )
func (pNtk *ABCNetwork) abcNtkMiterPrint(pString string, clk int64, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkMiterRwsat
// C: Abc_Ntk_t * Abc_NtkMiterRwsat( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkMiterRwsat() *ABCNetwork {
	panic("TODO, stubbed")
}
