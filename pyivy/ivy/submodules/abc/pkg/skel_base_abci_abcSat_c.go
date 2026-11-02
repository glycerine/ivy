package abc

// Hand-checked skeleton of src/base/abci/abcSat.c.

// Abc_NtkMiterSat
// C: int Abc_NtkMiterSat( Abc_Ntk_t * pNtk, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, int fVerbose, ABC_INT64_T * pNumConfs, ABC_INT64_T * pNumInspects )
func (pNtk *ABCNetwork) AbcNtkMiterSat(nConfLimit int64, nInsLimit int64, fVerbose int, pNumConfs *int64, pNumInspects *int64) int {
	panic("TODO, stubbed")
}

// Abc_NtkGetCiSatVarNums
// C: Vec_Int_t * Abc_NtkGetCiSatVarNums( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkGetCiSatVarNums() *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkClauseTriv
// C: int Abc_NtkClauseTriv( sat_solver * pSat, Abc_Obj_t * pNode, Vec_Int_t * vVars )
func (pSat *SatSolver) AbcNtkClauseTriv(pNode *ABCNetworkObject, vVars *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkClauseTop
// C: int Abc_NtkClauseTop( sat_solver * pSat, Vec_Ptr_t * vNodes, Vec_Int_t * vVars )
func (pSat *SatSolver) AbcNtkClauseTop(vNodes *PointerVector[ABCNetworkObject], vVars *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkClauseAnd
// C: int Abc_NtkClauseAnd( sat_solver * pSat, Abc_Obj_t * pNode, Vec_Ptr_t * vSuper, Vec_Int_t * vVars )
func (pSat *SatSolver) AbcNtkClauseAnd(pNode *ABCNetworkObject, vSuper *PointerVector[ABCNetworkObject], vVars *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkClauseMux
// C: int Abc_NtkClauseMux( sat_solver * pSat, Abc_Obj_t * pNode, Abc_Obj_t * pNodeC, Abc_Obj_t * pNodeT, Abc_Obj_t * pNodeE, Vec_Int_t * vVars )
func (pSat *SatSolver) AbcNtkClauseMux(pNode *ABCNetworkObject, pNodeC *ABCNetworkObject, pNodeT *ABCNetworkObject, pNodeE *ABCNetworkObject, vVars *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkCollectSupergate_rec
// C: int Abc_NtkCollectSupergate_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vSuper, int fFirst, int fStopAtMux )
func AbcNtkCollectSupergate_rec(pNode *ABCNetworkObject, vSuper *PointerVector[ABCNetworkObject], fFirst int, fStopAtMux int) int {
	panic("TODO, stubbed")
}

// Abc_NtkCollectSupergate
// C: void Abc_NtkCollectSupergate( Abc_Obj_t * pNode, int fStopAtMux, Vec_Ptr_t * vNodes )
func AbcNtkCollectSupergate(pNode *ABCNetworkObject, fStopAtMux int, vNodes *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkNodeFactor
// C: int Abc_NtkNodeFactor( Abc_Obj_t * pObj, int nLevelMax )
func AbcNtkNodeFactor(pObj *ABCNetworkObject, nLevelMax int) int {
	panic("TODO, stubbed")
}

// Abc_NtkMiterSatCreateInt
// C: int Abc_NtkMiterSatCreateInt( sat_solver * pSat, Abc_Ntk_t * pNtk )
func (pSat *SatSolver) AbcNtkMiterSatCreateInt(pNtk *ABCNetwork) int {
	panic("TODO, stubbed")
}

// Abc_NtkMiterSatCreate (returns a void * that holds a sat_solver *)
// C: void * Abc_NtkMiterSatCreate( Abc_Ntk_t * pNtk, int fAllPrimes )
func (pNtk *ABCNetwork) AbcNtkMiterSatCreate(fAllPrimes int) *SatSolver {
	panic("TODO, stubbed")
}

// Abc_NodeAddClauses
// C: int Abc_NodeAddClauses( sat_solver * pSat, char * pSop0, char * pSop1, Abc_Obj_t * pNode, Vec_Int_t * vVars )
func (pSat *SatSolver) AbcNodeAddClauses(pSop0 string, pSop1 string, pNode *ABCNetworkObject, vVars *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NodeAddClausesTop
// C: int Abc_NodeAddClausesTop( sat_solver * pSat, Abc_Obj_t * pNode, Vec_Int_t * vVars )
func (pSat *SatSolver) AbcNodeAddClausesTop(pNode *ABCNetworkObject, vVars *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkMiterSatCreateLogic (static)
// C: sat_solver * Abc_NtkMiterSatCreateLogic( Abc_Ntk_t * pNtk, int fAllPrimes )
func (pNtk *ABCNetwork) abcNtkMiterSatCreateLogic(fAllPrimes int) *SatSolver {
	panic("TODO, stubbed")
}

// Abc_NtkWriteSorterCnf
// C: void Abc_NtkWriteSorterCnf( char * pFileName, int nVars, int nQueens )
func AbcNtkWriteSorterCnf(pFileName string, nVars int, nQueens int) {
	panic("TODO, stubbed")
}
