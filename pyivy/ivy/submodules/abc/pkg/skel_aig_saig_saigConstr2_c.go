package abc

// Hand-checked skeleton of src/aig/saig/saigConstr2.c.

// Aig_ObjFrames (static inline)
// C: Aig_Obj_t * Aig_ObjFrames( Aig_Obj_t ** pObjMap, int nFs, Aig_Obj_t * pObj, int i )
func aigObjFrames(pObjMap []*AigObject, nFs int, pObj *AigObject, i int) *AigObject {
	panic("TODO, stubbed")
}

// Aig_ObjSetFrames (static inline)
// C: void Aig_ObjSetFrames( Aig_Obj_t ** pObjMap, int nFs, Aig_Obj_t * pObj, int i, Aig_Obj_t * pNode )
func aigObjSetFrames(pObjMap []*AigObject, nFs int, pObj *AigObject, i int, pNode *AigObject) {
	panic("TODO, stubbed")
}

// Aig_ObjChild0Frames (static inline)
// C: Aig_Obj_t * Aig_ObjChild0Frames( Aig_Obj_t ** pObjMap, int nFs, Aig_Obj_t * pObj, int i )
func aigObjChild0Frames(pObjMap []*AigObject, nFs int, pObj *AigObject, i int) *AigObject {
	panic("TODO, stubbed")
}

// Aig_ObjChild1Frames (static inline)
// C: Aig_Obj_t * Aig_ObjChild1Frames( Aig_Obj_t ** pObjMap, int nFs, Aig_Obj_t * pObj, int i )
func aigObjChild1Frames(pObjMap []*AigObject, nFs int, pObj *AigObject, i int) *AigObject {
	panic("TODO, stubbed")
}

// Ssw_ManProfileConstraints
// C: int Ssw_ManProfileConstraints( Aig_Man_t * p, int nWords, int nFrames, int fVerbose )
func (p *AigManager) SswManProfileConstraints(nWords int, nFrames int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Saig_ManCreateIndMiter
// C: Aig_Man_t * Saig_ManCreateIndMiter( Aig_Man_t * pAig, Vec_Vec_t * vCands )
func (pAig *AigManager) SaigManCreateIndMiter(vCands *NestedVector[*AigObject]) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManFilterUsingIndOne_new
// C: int Saig_ManFilterUsingIndOne_new( Aig_Man_t * p, Aig_Man_t * pFrame, sat_solver * pSat, Cnf_Dat_t * pCnf, int nConfs, int nProps, int Counter )
func (p *AigManager) SaigManFilterUsingIndOne_new(pFrame *AigManager, pSat *SatSolver, pCnf *CnfDat, nConfs int, nProps int, Counter int) int {
	panic("TODO, stubbed")
}

// Saig_ManFilterUsingInd
// C: void Saig_ManFilterUsingInd( Aig_Man_t * p, Vec_Vec_t * vCands, int nConfs, int nProps, int fVerbose )
func (p *AigManager) SaigManFilterUsingInd(vCands *NestedVector[*AigObject], nConfs int, nProps int, fVerbose int) {
	panic("TODO, stubbed")
}

// Saig_ManUnrollCOI_
// C: Aig_Man_t * Saig_ManUnrollCOI_( Aig_Man_t * p, int nFrames )
func (p *AigManager) SaigManUnrollCOI_(nFrames int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManUnrollCOI
// C: Aig_Man_t * Saig_ManUnrollCOI( Aig_Man_t * pAig, int nFrames )
func (pAig *AigManager) SaigManUnrollCOI(nFrames int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_CollectSatValues
// C: void Saig_CollectSatValues( sat_solver * pSat, Cnf_Dat_t * pCnf, Vec_Ptr_t * vInfo, int * piPat )
func (pSat *SatSolver) SaigCollectSatValues(pCnf *CnfDat, vInfo *PointerVector[uint32], piPat *int) {
	panic("TODO, stubbed")
}

// Saig_DetectTryPolarity
// C: int Saig_DetectTryPolarity( sat_solver * pSat, int nConfs, int nProps, Cnf_Dat_t * pCnf, Aig_Obj_t * pObj, int iPol, Vec_Ptr_t * vInfo, int * piPat, int fVerbose )
func (pSat *SatSolver) SaigDetectTryPolarity(nConfs int, nProps int, pCnf *CnfDat, pObj *AigObject, iPol int, vInfo *PointerVector[uint32], piPat *int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Ssw_ManFindDirectImplications
// C: Vec_Vec_t * Ssw_ManFindDirectImplications( Aig_Man_t * p, int nFrames, int nConfs, int nProps, int fVerbose )
func (p *AigManager) SswManFindDirectImplications(nFrames int, nConfs int, nProps int, fVerbose int) *NestedVector[*AigObject] {
	panic("TODO, stubbed")
}

// Saig_ManDetectConstrFunc
// C: Vec_Vec_t * Saig_ManDetectConstrFunc( Aig_Man_t * p, int nFrames, int nConfs, int nProps, int fVerbose )
func (p *AigManager) SaigManDetectConstrFunc(nFrames int, nConfs int, nProps int, fVerbose int) *NestedVector[*AigObject] {
	panic("TODO, stubbed")
}

// Saig_ManDetectConstrFuncTest
// C: void Saig_ManDetectConstrFuncTest( Aig_Man_t * p, int nFrames, int nConfs, int nProps, int fOldAlgo, int fVerbose )
func (p *AigManager) SaigManDetectConstrFuncTest(nFrames int, nConfs int, nProps int, fOldAlgo int, fVerbose int) {
	panic("TODO, stubbed")
}

// Saig_ManDupUnfoldConstrsFunc
// C: Aig_Man_t * Saig_ManDupUnfoldConstrsFunc( Aig_Man_t * pAig, int nFrames, int nConfs, int nProps, int fOldAlgo, int fVerbose )
func (pAig *AigManager) SaigManDupUnfoldConstrsFunc(nFrames int, nConfs int, nProps int, fOldAlgo int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManDupFoldConstrsFunc
// C: Aig_Man_t * Saig_ManDupFoldConstrsFunc( Aig_Man_t * pAig, int fCompl, int fVerbose )
func (pAig *AigManager) SaigManDupFoldConstrsFunc(fCompl int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}
