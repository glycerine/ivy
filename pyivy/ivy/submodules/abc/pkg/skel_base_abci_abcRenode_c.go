package abc

// Hand-checked skeleton of src/base/abci/abcRenode.c.

// Abc_NtkRenode
// C: Abc_Ntk_t * Abc_NtkRenode( Abc_Ntk_t * pNtk, int nFaninMax, int nCubeMax, int nFlowIters, int nAreaIters, int fArea, int fUseBdds, int fUseSops, int fUseCnfs, int fUseMv, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkRenode(nFaninMax int, nCubeMax int, nFlowIters int, nAreaIters int, fArea int, fUseBdds int, fUseSops int, fUseCnfs int, fUseMv int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkRenodeEvalAig (static)
// C: int Abc_NtkRenodeEvalAig( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) abcNtkRenodeEvalAig(pCut *IfCut) int {
	panic("TODO, stubbed")
}

// Abc_NtkRenodeEvalBdd (static)
// C: int Abc_NtkRenodeEvalBdd( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) abcNtkRenodeEvalBdd(pCut *IfCut) int {
	panic("TODO, stubbed")
}

// Abc_NtkRenodeEvalSop (static)
// C: int Abc_NtkRenodeEvalSop( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) abcNtkRenodeEvalSop(pCut *IfCut) int {
	panic("TODO, stubbed")
}

// Abc_NtkRenodeEvalCnf (static)
// C: int Abc_NtkRenodeEvalCnf( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) abcNtkRenodeEvalCnf(pCut *IfCut) int {
	panic("TODO, stubbed")
}

// Abc_NtkRenodeEvalMv (static)
// C: int Abc_NtkRenodeEvalMv( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) abcNtkRenodeEvalMv(pCut *IfCut) int {
	panic("TODO, stubbed")
}
