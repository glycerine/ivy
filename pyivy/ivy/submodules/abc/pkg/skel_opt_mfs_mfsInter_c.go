package abc

// Hand-checked skeleton of src/opt/mfs/mfsInter.c.

// Abc_MfsSatAddXor
// C: int Abc_MfsSatAddXor( sat_solver * pSat, int iVarA, int iVarB, int iVarC )
func AbcMfsSatAddXor(pSat *SatSolver, iVarA int, iVarB int, iVarC int) int {
	panic("TODO, stubbed")
}

// Abc_MfsCreateSolverResub
// C: sat_solver * Abc_MfsCreateSolverResub( Mfs_Man_t * p, int * pCands, int nCands, int fInvert )
func (p *MfsMan) AbcMfsCreateSolverResub(pCands []int, nCands int, fInvert int) *SatSolver {
	panic("TODO, stubbed")
}

// Abc_NtkMfsInterplateTruth
// C: unsigned * Abc_NtkMfsInterplateTruth( Mfs_Man_t * p, int * pCands, int nCands, int fInvert )
func (p *MfsMan) AbcNtkMfsInterplateTruth(pCands []int, nCands int, fInvert int) []uint32 {
	panic("TODO, stubbed")
}

// Abc_NtkMfsInterplateEval
// C: int Abc_NtkMfsInterplateEval( Mfs_Man_t * p, int * pCands, int nCands )
func (p *MfsMan) AbcNtkMfsInterplateEval(pCands []int, nCands int) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsInterplate
// C: Hop_Obj_t * Abc_NtkMfsInterplate( Mfs_Man_t * p, int * pCands, int nCands )
func (p *MfsMan) AbcNtkMfsInterplate(pCands []int, nCands int) *HopObject {
	panic("TODO, stubbed")
}
