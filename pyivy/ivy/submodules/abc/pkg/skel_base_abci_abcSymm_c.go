package abc

// Hand-checked skeleton of src/base/abci/abcSymm.c.

// ExtraSymmInfo is Extra_SymmInfo_t (src/bdd/extrab/extraBdd.h).
type ExtraSymmInfo struct {
	nVars    int      // the number of variables in the support
	nVarsMax int      // the number of variables in the DD manager
	nSymms   int      // the number of pair-wise symmetries
	nNodes   int      // the number of nodes in a ZDD (if applicable)
	pVars    []int    // int * : the list of all variables present in the support
	pSymms   [][]int8 // char ** : the symmetry information
}

// Abc_NtkSymmetries
// C: void Abc_NtkSymmetries( Abc_Ntk_t * pNtk, int fUseBdds, int fNaive, int fReorder, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkSymmetries(fUseBdds int, fNaive int, fReorder int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkSymmetriesUsingSandS (static)
// C: void Abc_NtkSymmetriesUsingSandS( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) abcNtkSymmetriesUsingSandS(fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkSymmetriesUsingBdds (static)
// C: void Abc_NtkSymmetriesUsingBdds( Abc_Ntk_t * pNtk, int fNaive, int fReorder, int fVerbose )
func (pNtk *ABCNetwork) abcNtkSymmetriesUsingBdds(fNaive int, fReorder int, fVerbose int) {
	panic("TODO, stubbed")
}

// Ntk_NetworkSymmsBdd (static)
// C: void Ntk_NetworkSymmsBdd( DdManager * dd, Abc_Ntk_t * pNtk, int fNaive, int fVerbose )
func ntkNetworkSymmsBdd(dd *CuddFoundation, pNtk *ABCNetwork, fNaive int, fVerbose int) {
	panic("TODO, stubbed")
}

// Ntk_NetworkSymmsPrint (static)
// C: void Ntk_NetworkSymmsPrint( Abc_Ntk_t * pNtk, Extra_SymmInfo_t * pSymms )
func (pNtk *ABCNetwork) ntkNetworkSymmsPrint(pSymms *ExtraSymmInfo) {
	panic("TODO, stubbed")
}

// Ntk_SymTryRandomFlips
// C: void Ntk_SymTryRandomFlips( word * pFun, word * pNpn, int nVars )
func NtkSymTryRandomFlips(pFun []uint64, pNpn []uint64, nVars int) {
	panic("TODO, stubbed")
}

// Ntk_SymFunDeriveNpn
// C: void Ntk_SymFunDeriveNpn( word * pFun, int nVars, int * pComp )
func NtkSymFunDeriveNpn(pFun []uint64, nVars int, pComp []int) {
	panic("TODO, stubbed")
}

// Ntk_SymFunGenerate
// C: void Ntk_SymFunGenerate( int nVars, int fVerbose )
func NtkSymFunGenerate(nVars int, fVerbose int) {
	panic("TODO, stubbed")
}
