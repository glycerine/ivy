package abc

import "os"

// Hand-checked skeleton of src/proof/live/kliveness.c.

// Functions inside "#if 0" regions of the C source are dead code and are not
// represented.

// readLiveSignal_0
// C: Aig_Obj_t * readLiveSignal_0( Aig_Man_t *pAig, int liveIndex_0 )
func (pAig *AigManager) readLiveSignal_0(liveIndex_0 int) *AigObject {
	panic("TODO, stubbed")
}

// readLiveSignal_k
// C: Aig_Obj_t * readLiveSignal_k( Aig_Man_t *pAig, int liveIndex_k )
func (pAig *AigManager) readLiveSignal_k(liveIndex_k int) *AigObject {
	panic("TODO, stubbed")
}

// introduceAbsorberLogic (pLiveIndex_0 and pLiveIndex_k are in/out scalar parameters)
// C: Aig_Man_t * introduceAbsorberLogic( Aig_Man_t *pAig, int *pLiveIndex_0, int *pLiveIndex_k, int nonFirstIteration )
func (pAig *AigManager) introduceAbsorberLogic(pLiveIndex_0 *int, pLiveIndex_k *int, nonFirstIteration int) *AigManager {
	panic("TODO, stubbed")
}

// modifyAigToApplySafetyInvar
// C: void modifyAigToApplySafetyInvar(Aig_Man_t *pAig, int csTarget, int safetyInvarPO)
func (pAig *AigManager) modifyAigToApplySafetyInvar(csTarget int, safetyInvarPO int) {
	panic("TODO, stubbed")
}

// flipConePdr
// C: int flipConePdr( Aig_Man_t *pAig, int directive, int targetCSPropertyIndex, int safetyInvariantPOIndex, int absorberCount )
func (pAig *AigManager) flipConePdr(directive int, targetCSPropertyIndex int, safetyInvariantPOIndex int, absorberCount int) int {
	panic("TODO, stubbed")
}

// collectSafetyInvariantPOIndex
// C: int collectSafetyInvariantPOIndex(Abc_Ntk_t *pNtk)
func collectSafetyInvariantPOIndex(pNtk *ABCNetwork) int {
	panic("TODO, stubbed")
}

// collectUserGivenDisjunctiveMonotoneSignals (the Vec_Ptr_t holds Vec_Int_t * singletons; returns NULL when empty)
// C: Vec_Ptr_t * collectUserGivenDisjunctiveMonotoneSignals( Abc_Ntk_t *pNtk )
func collectUserGivenDisjunctiveMonotoneSignals(pNtk *ABCNetwork) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// deallocateMasterBarrierDisjunctInt (the Vec_Ptr_t holds Vec_Int_t *)
// C: void deallocateMasterBarrierDisjunctInt(Vec_Ptr_t *vMasterBarrierDisjunctsArg)
func deallocateMasterBarrierDisjunctInt(vMasterBarrierDisjunctsArg *PointerVector[IntVector]) {
	panic("TODO, stubbed")
}

// deallocateMasterBarrierDisjunctVecPtrVecInt (the Vec_Ptr_t holds Vec_Ptr_t * of Vec_Int_t *)
// C: void deallocateMasterBarrierDisjunctVecPtrVecInt(Vec_Ptr_t *vMasterBarrierDisjunctsArg)
func deallocateMasterBarrierDisjunctVecPtrVecInt(vMasterBarrierDisjunctsArg *PointerVector[PointerVector[IntVector]]) {
	panic("TODO, stubbed")
}

// getVecOfVecFairness (returns an empty Vec_Ptr_t; the C never pushes anything, so the element type is untyped void *)
// C: Vec_Ptr_t * getVecOfVecFairness(FILE *fp)
func getVecOfVecFairness(fp *os.File) *PointerVector[any] {
	panic("TODO, stubbed")
}

// Abc_CommandCS_kLiveness
// C: int Abc_CommandCS_kLiveness( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) AbcCommandCS_kLiveness(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandNChooseK
// C: int Abc_CommandNChooseK( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) AbcCommandNChooseK(argc int, argv []string) int {
	panic("TODO, stubbed")
}
