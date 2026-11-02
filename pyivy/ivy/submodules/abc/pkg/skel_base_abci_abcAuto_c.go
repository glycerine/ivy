package abc

// Hand-checked skeleton of src/base/abci/abcAuto.c.

// Abc_NtkAutoPrint
// C: void Abc_NtkAutoPrint( Abc_Ntk_t * pNtk, int Output, int fNaive, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkAutoPrint(Output int, fNaive int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkAutoPrintAll (static)
// C: void Abc_NtkAutoPrintAll( DdManager * dd, int nInputs, DdNode * pbOutputs[], int nOutputs, char * pInputNames[], char * pOutputNames[], int fNaive )
func abcNtkAutoPrintAll(dd *CuddFoundation, nInputs int, pbOutputs []*DdNode, nOutputs int, pInputNames []string, pOutputNames []string, fNaive int) {
	panic("TODO, stubbed")
}

// Abc_NtkAutoPrintOne (static)
// C: void Abc_NtkAutoPrintOne( DdManager * dd, int nInputs, DdNode * pbOutputs[], int Output, char * pInputNames[], char * pOutputNames[], int fNaive )
func abcNtkAutoPrintOne(dd *CuddFoundation, nInputs int, pbOutputs []*DdNode, Output int, pInputNames []string, pOutputNames []string, fNaive int) {
	panic("TODO, stubbed")
}
