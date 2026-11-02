package abc

// Hand-checked skeleton of src/base/abci/abcRewrite.c.

// Abc_NtkRewrite
// C: int Abc_NtkRewrite( Abc_Ntk_t * pNtk, int fUpdateLevel, int fUseZeros, int fVerbose, int fVeryVerbose, int fPlaceEnable )
func (pNtk *ABCNetwork) AbcNtkRewrite(fUpdateLevel int, fUseZeros int, fVerbose int, fVeryVerbose int, fPlaceEnable int) int {
	panic("TODO, stubbed")
}

// Abc_NtkStartCutManForRewrite (static)
// C: Cut_Man_t * Abc_NtkStartCutManForRewrite( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkStartCutManForRewrite() *CutMan {
	panic("TODO, stubbed")
}

// Abc_NodePrintCuts (static)
// C: void Abc_NodePrintCuts( Abc_Obj_t * pNode )
func abcNodePrintCuts(pNode *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_ManRewritePrintDivs
// C: void Abc_ManRewritePrintDivs( Vec_Ptr_t * vDivs, int nLeaves )
func AbcManRewritePrintDivs(vDivs *PointerVector[ABCNetworkObject], nLeaves int) {
	panic("TODO, stubbed")
}

// Abc_ManShowCutCone_rec
// C: void Abc_ManShowCutCone_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vDivs )
func AbcManShowCutCone_rec(pNode *ABCNetworkObject, vDivs *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_ManShowCutCone (static)
// C: void Abc_ManShowCutCone( Abc_Obj_t * pNode, Vec_Ptr_t * vLeaves )
func abcManShowCutCone(pNode *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_RwrExpWithCut_rec
// C: void Abc_RwrExpWithCut_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vLeaves, int fUseA )
func AbcRwrExpWithCut_rec(pNode *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], fUseA int) {
	panic("TODO, stubbed")
}

// Abc_RwrExpWithCut
// C: void Abc_RwrExpWithCut( Abc_Obj_t * pNode, Vec_Ptr_t * vLeaves )
func AbcRwrExpWithCut(pNode *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}
