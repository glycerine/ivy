package abc

// Hand-checked skeleton of src/proof/live/liveness.c.

// Code inside "#if 0" regions is dead and is not represented.

// nodeName_starts_with (static)
// C: int nodeName_starts_with( Abc_Obj_t *pNode, const char *prefix )
func nodeName_starts_with(pNode *ABCNetworkObject, prefix string) int {
	panic("TODO, stubbed")
}

// printVecPtrOfString (the Vec_Ptr_t holds char * strings)
// C: void printVecPtrOfString( Vec_Ptr_t *vec )
func printVecPtrOfString(vec *PointerVector[string]) {
	panic("TODO, stubbed")
}

// getPoIndex
// C: int getPoIndex( Aig_Man_t *pAig, Aig_Obj_t *pPivot )
func (pAig *AigManager) getPoIndex(pPivot *AigObject) int {
	panic("TODO, stubbed")
}

// retrieveTruePiName
// C: char * retrieveTruePiName( Abc_Ntk_t *pNtkOld, Aig_Man_t *pAigOld, Aig_Man_t *pAigNew, Aig_Obj_t *pObjPivot )
func retrieveTruePiName(pNtkOld *ABCNetwork, pAigOld *AigManager, pAigNew *AigManager, pObjPivot *AigObject) string {
	panic("TODO, stubbed")
}

// retrieveLOName
// C: char * retrieveLOName( Abc_Ntk_t *pNtkOld, Aig_Man_t *pAigOld, Aig_Man_t *pAigNew, Aig_Obj_t *pObjPivot, Vec_Ptr_t *vLive, Vec_Ptr_t * vFair )
func retrieveLOName(pNtkOld *ABCNetwork, pAigOld *AigManager, pAigNew *AigManager, pObjPivot *AigObject, vLive *PointerVector[AigObject], vFair *PointerVector[AigObject]) string {
	panic("TODO, stubbed")
}

// Aig_ManCiCleanupBiere
// C: int Aig_ManCiCleanupBiere( Aig_Man_t * p )
func (p *AigManager) AigManCiCleanupBiere() int {
	panic("TODO, stubbed")
}

// Aig_ManCoCleanupBiere
// C: int Aig_ManCoCleanupBiere( Aig_Man_t * p )
func (p *AigManager) AigManCoCleanupBiere() int {
	panic("TODO, stubbed")
}

// LivenessToSafetyTransformation
// C: Aig_Man_t * LivenessToSafetyTransformation( int mode, Abc_Ntk_t * pNtk, Aig_Man_t * p, Vec_Ptr_t *vLive, Vec_Ptr_t *vFair, Vec_Ptr_t *vAssertSafety, Vec_Ptr_t *vAssumeSafety )
func LivenessToSafetyTransformation(mode int, pNtk *ABCNetwork, p *AigManager, vLive *PointerVector[AigObject], vFair *PointerVector[AigObject], vAssertSafety *PointerVector[AigObject], vAssumeSafety *PointerVector[AigObject]) *AigManager {
	panic("TODO, stubbed")
}

// LivenessToSafetyTransformationAbs
// C: Aig_Man_t * LivenessToSafetyTransformationAbs( int mode, Abc_Ntk_t * pNtk, Aig_Man_t * p, Vec_Int_t *vFlops, Vec_Ptr_t *vLive, Vec_Ptr_t *vFair, Vec_Ptr_t *vAssertSafety, Vec_Ptr_t *vAssumeSafety )
func LivenessToSafetyTransformationAbs(mode int, pNtk *ABCNetwork, p *AigManager, vFlops *IntVector, vLive *PointerVector[AigObject], vFair *PointerVector[AigObject], vAssertSafety *PointerVector[AigObject], vAssumeSafety *PointerVector[AigObject]) *AigManager {
	panic("TODO, stubbed")
}

// LivenessToSafetyTransformationOneStepLoop
// C: Aig_Man_t * LivenessToSafetyTransformationOneStepLoop( int mode, Abc_Ntk_t * pNtk, Aig_Man_t * p, Vec_Ptr_t *vLive, Vec_Ptr_t *vFair, Vec_Ptr_t *vAssertSafety, Vec_Ptr_t *vAssumeSafety )
func LivenessToSafetyTransformationOneStepLoop(mode int, pNtk *ABCNetwork, p *AigManager, vLive *PointerVector[AigObject], vFair *PointerVector[AigObject], vAssertSafety *PointerVector[AigObject], vAssumeSafety *PointerVector[AigObject]) *AigManager {
	panic("TODO, stubbed")
}

// populateLivenessVector (holds Aig_Obj_t * COs)
// C: Vec_Ptr_t * populateLivenessVector( Abc_Ntk_t *pNtk, Aig_Man_t *pAig )
func populateLivenessVector(pNtk *ABCNetwork, pAig *AigManager) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// populateFairnessVector (holds Aig_Obj_t * COs)
// C: Vec_Ptr_t * populateFairnessVector( Abc_Ntk_t *pNtk, Aig_Man_t *pAig )
func populateFairnessVector(pNtk *ABCNetwork, pAig *AigManager) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// populateSafetyAssertionVector (holds Aig_Obj_t * COs)
// C: Vec_Ptr_t * populateSafetyAssertionVector( Abc_Ntk_t *pNtk, Aig_Man_t *pAig )
func populateSafetyAssertionVector(pNtk *ABCNetwork, pAig *AigManager) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// populateSafetyAssumptionVector (holds Aig_Obj_t * COs)
// C: Vec_Ptr_t * populateSafetyAssumptionVector( Abc_Ntk_t *pNtk, Aig_Man_t *pAig )
func populateSafetyAssumptionVector(pNtk *ABCNetwork, pAig *AigManager) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// updateNewNetworkNameManager (the Vec_Ptr_t arguments hold char * names)
// C: void updateNewNetworkNameManager( Abc_Ntk_t *pNtk, Aig_Man_t *pAig, Vec_Ptr_t *vPiNames, Vec_Ptr_t *vLoNames )
func updateNewNetworkNameManager(pNtk *ABCNetwork, pAig *AigManager, vPiNames *PointerVector[string], vLoNames *PointerVector[string]) {
	panic("TODO, stubbed")
}

// Abc_CommandAbcLivenessToSafety
// C: int Abc_CommandAbcLivenessToSafety( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) AbcCommandAbcLivenessToSafety(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// prepareFlopVector
// C: Vec_Int_t * prepareFlopVector( Aig_Man_t * pAig, int vectorLength )
func (pAig *AigManager) prepareFlopVector(vectorLength int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_CommandAbcLivenessToSafetyAbstraction
// C: int Abc_CommandAbcLivenessToSafetyAbstraction( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) AbcCommandAbcLivenessToSafetyAbstraction(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// LivenessToSafetyTransformationWithLTL (numLtlProcessed is an out parameter; ltlBuffer holds char * formulas)
// C: Aig_Man_t * LivenessToSafetyTransformationWithLTL( int mode, Abc_Ntk_t * pNtk, Aig_Man_t * p, Vec_Ptr_t *vLive, Vec_Ptr_t *vFair, Vec_Ptr_t *vAssertSafety, Vec_Ptr_t *vAssumeSafety, int *numLtlProcessed, Vec_Ptr_t *ltlBuffer )
func LivenessToSafetyTransformationWithLTL(mode int, pNtk *ABCNetwork, p *AigManager, vLive *PointerVector[AigObject], vFair *PointerVector[AigObject], vAssertSafety *PointerVector[AigObject], vAssumeSafety *PointerVector[AigObject], numLtlProcessed *int, ltlBuffer *PointerVector[string]) *AigManager {
	panic("TODO, stubbed")
}

// Abc_CommandAbcLivenessToSafetyWithLTL
// C: int Abc_CommandAbcLivenessToSafetyWithLTL( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) AbcCommandAbcLivenessToSafetyWithLTL(argc int, argv []string) int {
	panic("TODO, stubbed")
}
