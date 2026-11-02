package abc

// Hand-checked skeleton of src/proof/live/liveness_sim.c.

// All of these are static in the C source; the names that clash with the
// non-static functions of liveness.c carry a _sim suffix here.

// printVecPtrOfString (static)
// C: void printVecPtrOfString( Vec_Ptr_t *vec )
func printVecPtrOfString_sim(vec *PointerVector[string]) {
	panic("TODO, stubbed")
}

// getPoIndex (static)
// C: int getPoIndex( Aig_Man_t *pAig, Aig_Obj_t *pPivot )
func (pAig *AigManager) getPoIndex_sim(pPivot *AigObject) int {
	panic("TODO, stubbed")
}

// retrieveTruePiName (static)
// C: char * retrieveTruePiName( Abc_Ntk_t *pNtkOld, Aig_Man_t *pAigOld, Aig_Man_t *pAigNew, Aig_Obj_t *pObjPivot )
func retrieveTruePiName_sim(pNtkOld *ABCNetwork, pAigOld *AigManager, pAigNew *AigManager, pObjPivot *AigObject) string {
	panic("TODO, stubbed")
}

// retrieveLOName (static)
// C: char * retrieveLOName( Abc_Ntk_t *pNtkOld, Aig_Man_t *pAigOld, Aig_Man_t *pAigNew, Aig_Obj_t *pObjPivot, Vec_Ptr_t *vLive, Vec_Ptr_t * vFair )
func retrieveLOName_sim(pNtkOld *ABCNetwork, pAigOld *AigManager, pAigNew *AigManager, pObjPivot *AigObject, vLive *PointerVector[AigObject], vFair *PointerVector[AigObject]) string {
	panic("TODO, stubbed")
}

// Aig_ManCiCleanupBiere (static)
// C: int Aig_ManCiCleanupBiere( Aig_Man_t * p )
func (p *AigManager) aigManCiCleanupBiere_sim() int {
	panic("TODO, stubbed")
}

// Aig_ManCoCleanupBiere (static)
// C: int Aig_ManCoCleanupBiere( Aig_Man_t * p )
func (p *AigManager) aigManCoCleanupBiere_sim() int {
	panic("TODO, stubbed")
}

// LivenessToSafetyTransformationSim (static)
// C: Aig_Man_t * LivenessToSafetyTransformationSim( Abc_Ntk_t * pNtk, Aig_Man_t * p, Vec_Ptr_t *vLive, Vec_Ptr_t *vFair )
func livenessToSafetyTransformationSim(pNtk *ABCNetwork, p *AigManager, vLive *PointerVector[AigObject], vFair *PointerVector[AigObject]) *AigManager {
	panic("TODO, stubbed")
}

// LivenessToSafetyTransformationOneStepLoopSim (static)
// C: Aig_Man_t * LivenessToSafetyTransformationOneStepLoopSim( Abc_Ntk_t * pNtk, Aig_Man_t * p, Vec_Ptr_t *vLive, Vec_Ptr_t *vFair )
func livenessToSafetyTransformationOneStepLoopSim(pNtk *ABCNetwork, p *AigManager, vLive *PointerVector[AigObject], vFair *PointerVector[AigObject]) *AigManager {
	panic("TODO, stubbed")
}

// populateLivenessVector (static; holds Aig_Obj_t * COs)
// C: Vec_Ptr_t * populateLivenessVector( Abc_Ntk_t *pNtk, Aig_Man_t *pAig )
func populateLivenessVector_sim(pNtk *ABCNetwork, pAig *AigManager) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// populateFairnessVector (static; holds Aig_Obj_t * COs)
// C: Vec_Ptr_t * populateFairnessVector( Abc_Ntk_t *pNtk, Aig_Man_t *pAig )
func populateFairnessVector_sim(pNtk *ABCNetwork, pAig *AigManager) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// updateNewNetworkNameManager (static)
// C: void updateNewNetworkNameManager( Abc_Ntk_t *pNtk, Aig_Man_t *pAig, Vec_Ptr_t *vPiNames, Vec_Ptr_t *vLoNames )
func updateNewNetworkNameManager_sim(pNtk *ABCNetwork, pAig *AigManager, vPiNames *PointerVector[string], vLoNames *PointerVector[string]) {
	panic("TODO, stubbed")
}

// Abc_CommandAbcLivenessToSafetySim
// C: int Abc_CommandAbcLivenessToSafetySim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) AbcCommandAbcLivenessToSafetySim(argc int, argv []string) int {
	panic("TODO, stubbed")
}
