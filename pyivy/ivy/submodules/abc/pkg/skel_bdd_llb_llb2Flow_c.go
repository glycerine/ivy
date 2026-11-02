package abc

// Hand-checked skeleton of src/bdd/llb/llb2Flow.c.

// Llb_ObjSetPath (static inline)
// C: int Llb_ObjSetPath( Aig_Obj_t * pObj, Aig_Obj_t * pNext )
func llbObjSetPath(pObj *AigObject, pNext *AigObject) int {
	panic("TODO, stubbed")
}

// Llb_ObjGetPath (static inline)
// C: Aig_Obj_t * Llb_ObjGetPath( Aig_Obj_t * pObj )
func llbObjGetPath(pObj *AigObject) *AigObject {
	panic("TODO, stubbed")
}

// Llb_ObjGetFanoutPath (static inline)
// C: Aig_Obj_t * Llb_ObjGetFanoutPath( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) llbObjGetFanoutPath(pObj *AigObject) *AigObject {
	panic("TODO, stubbed")
}

// Llb_ManCutSupps
// C: Vec_Ptr_t * Llb_ManCutSupps( Aig_Man_t * p, Vec_Ptr_t * vResult )
func (p *AigManager) LlbManCutSupps(vResult *PointerVector[PointerVector[AigObject]]) *PointerVector[PointerVector[AigObject]] {
	panic("TODO, stubbed")
}

// Llb_ManCutMap
// C: Vec_Ptr_t * Llb_ManCutMap( Aig_Man_t * p, Vec_Ptr_t * vResult, Vec_Ptr_t * vSupps )
func (p *AigManager) LlbManCutMap(vResult *PointerVector[PointerVector[AigObject]], vSupps *PointerVector[PointerVector[AigObject]]) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Llb_ManCutPiNum
// C: int Llb_ManCutPiNum( Aig_Man_t * p, Vec_Ptr_t * vMinCut )
func (p *AigManager) LlbManCutPiNum(vMinCut *PointerVector[AigObject]) int {
	panic("TODO, stubbed")
}

// Llb_ManCutLoNum
// C: int Llb_ManCutLoNum( Aig_Man_t * p, Vec_Ptr_t * vMinCut )
func (p *AigManager) LlbManCutLoNum(vMinCut *PointerVector[AigObject]) int {
	panic("TODO, stubbed")
}

// Llb_ManCutLiNum
// C: int Llb_ManCutLiNum( Aig_Man_t * p, Vec_Ptr_t * vMinCut )
func (p *AigManager) LlbManCutLiNum(vMinCut *PointerVector[AigObject]) int {
	panic("TODO, stubbed")
}

// Llb_ManCutVolume_rec
// C: int Llb_ManCutVolume_rec( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) LlbManCutVolume_rec(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Llb_ManCutVolume
// C: int Llb_ManCutVolume( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper )
func (p *AigManager) LlbManCutVolume(vLower *PointerVector[AigObject], vUpper *PointerVector[AigObject]) int {
	panic("TODO, stubbed")
}

// Llb_ManCutNodes_rec
// C: void Llb_ManCutNodes_rec( Aig_Man_t * p, Aig_Obj_t * pObj, Vec_Ptr_t * vNodes )
func (p *AigManager) LlbManCutNodes_rec(pObj *AigObject, vNodes *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_ManCutNodes
// C: Vec_Ptr_t * Llb_ManCutNodes( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper )
func (p *AigManager) LlbManCutNodes(vLower *PointerVector[AigObject], vUpper *PointerVector[AigObject]) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_ManCutSupp
// C: Vec_Ptr_t * Llb_ManCutSupp( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper )
func (p *AigManager) LlbManCutSupp(vLower *PointerVector[AigObject], vUpper *PointerVector[AigObject]) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_ManCutRange
// C: Vec_Ptr_t * Llb_ManCutRange( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper )
func (p *AigManager) LlbManCutRange(vLower *PointerVector[AigObject], vUpper *PointerVector[AigObject]) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_ManCutPrint
// C: void Llb_ManCutPrint( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper )
func (p *AigManager) LlbManCutPrint(vLower *PointerVector[AigObject], vUpper *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_ManResultPrint
// C: void Llb_ManResultPrint( Aig_Man_t * p, Vec_Ptr_t * vResult )
func (p *AigManager) LlbManResultPrint(vResult *PointerVector[PointerVector[AigObject]]) {
	panic("TODO, stubbed")
}

// Llb_ManFlowBwdPath2_rec
// C: int Llb_ManFlowBwdPath2_rec( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) LlbManFlowBwdPath2_rec(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Llb_ManFlowLabelTfi_rec
// C: void Llb_ManFlowLabelTfi_rec( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) LlbManFlowLabelTfi_rec(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Llb_ManFlowUpdateCut
// C: void Llb_ManFlowUpdateCut( Aig_Man_t * p, Vec_Ptr_t * vMinCut )
func (p *AigManager) LlbManFlowUpdateCut(vMinCut *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_ManFlowMinCut
// C: Vec_Ptr_t * Llb_ManFlowMinCut( Aig_Man_t * p )
func (p *AigManager) LlbManFlowMinCut() *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_ManFlowVerifyCut_rec
// C: int Llb_ManFlowVerifyCut_rec( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) LlbManFlowVerifyCut_rec(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Llb_ManFlowVerifyCut
// C: int Llb_ManFlowVerifyCut( Aig_Man_t * p, Vec_Ptr_t * vMinCut )
func (p *AigManager) LlbManFlowVerifyCut(vMinCut *PointerVector[AigObject]) int {
	panic("TODO, stubbed")
}

// Llb_ManFlow
// C: Vec_Ptr_t * Llb_ManFlow( Aig_Man_t * p, Vec_Ptr_t * vSources, int * pnFlow )
func (p *AigManager) LlbManFlow(vSources *PointerVector[AigObject], pnFlow *int) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_ManFlowCompute
// C: Vec_Ptr_t * Llb_ManFlowCompute( Aig_Man_t * p )
func (p *AigManager) LlbManFlowCompute() *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_ManFlowCleanMarkB_rec
// C: void Llb_ManFlowCleanMarkB_rec( Aig_Obj_t * pObj )
func LlbManFlowCleanMarkB_rec(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Llb_ManFlowSetMarkA_rec
// C: void Llb_ManFlowSetMarkA_rec( Aig_Obj_t * pObj )
func LlbManFlowSetMarkA_rec(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Llb_ManFlowPrepareCut
// C: void Llb_ManFlowPrepareCut( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper )
func (p *AigManager) LlbManFlowPrepareCut(vLower *PointerVector[AigObject], vUpper *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_ManFlowUnmarkCone
// C: void Llb_ManFlowUnmarkCone( Aig_Man_t * p, Vec_Ptr_t * vCone )
func (p *AigManager) LlbManFlowUnmarkCone(vCone *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_ManFlowCollectAndMarkCone_rec
// C: void Llb_ManFlowCollectAndMarkCone_rec( Aig_Man_t * p, Aig_Obj_t * pObj, Vec_Ptr_t * vCone )
func (p *AigManager) LlbManFlowCollectAndMarkCone_rec(pObj *AigObject, vCone *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_ManFlowCollectAndMarkCone
// C: void Llb_ManFlowCollectAndMarkCone( Aig_Man_t * p, Vec_Ptr_t * vStarts, Vec_Ptr_t * vCone )
func (p *AigManager) LlbManFlowCollectAndMarkCone(vStarts *PointerVector[AigObject], vCone *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_ManComputeCutLo
// C: Vec_Ptr_t * Llb_ManComputeCutLo( Aig_Man_t * p )
func (p *AigManager) LlbManComputeCutLo() *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_ManComputeCutLi
// C: Vec_Ptr_t * Llb_ManComputeCutLi( Aig_Man_t * p )
func (p *AigManager) LlbManComputeCutLi() *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_ManFlowGetObjSet
// C: void Llb_ManFlowGetObjSet( Aig_Man_t * p, Vec_Ptr_t * vLower, int iStart, int nSize, Vec_Ptr_t * vSet )
func (p *AigManager) LlbManFlowGetObjSet(vLower *PointerVector[AigObject], iStart int, nSize int, vSet *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_ManFlowFindBestCut
// C: Vec_Ptr_t * Llb_ManFlowFindBestCut( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper, int Num )
func (p *AigManager) LlbManFlowFindBestCut(vLower *PointerVector[AigObject], vUpper *PointerVector[AigObject], Num int) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_ManComputeCuts
// C: Vec_Ptr_t * Llb_ManComputeCuts( Aig_Man_t * p, int Num, int fVerbose, int fVeryVerbose )
func (p *AigManager) LlbManComputeCuts(Num int, fVerbose int, fVeryVerbose int) *PointerVector[PointerVector[AigObject]] {
	panic("TODO, stubbed")
}

// Llb_BddSetDefaultParams
// C: void Llb_BddSetDefaultParams( Gia_ParLlb_t * p )
func LlbBddSetDefaultParams(p *GiaParLlb) {
	panic("TODO, stubbed")
}

// Llb_ManMinCutTest
// C: void Llb_ManMinCutTest( Aig_Man_t * pAig, int Num )
func (pAig *AigManager) LlbManMinCutTest(Num int) {
	panic("TODO, stubbed")
}
