package abc

// Hand-checked skeleton of src/aig/gia/giaSweep.c.

// Gia_ManMarkSeqGiaWithBoxes_rec
// C: void Gia_ManMarkSeqGiaWithBoxes_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vRoots )
func (p *GiaManager) GiaManMarkSeqGiaWithBoxes_rec(pObj *GiaObject, vRoots *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManMarkSeqGiaWithBoxes
// C: void Gia_ManMarkSeqGiaWithBoxes( Gia_Man_t * p, int fSeq )
func (p *GiaManager) GiaManMarkSeqGiaWithBoxes(fSeq int) {
	panic("TODO, stubbed")
}

// Gia_ManDupWithBoxes
// C: Gia_Man_t * Gia_ManDupWithBoxes( Gia_Man_t * p, int fSeq )
func (p *GiaManager) GiaManDupWithBoxes(fSeq int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManFraigCheckCis
// C: int Gia_ManFraigCheckCis( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) GiaManFraigCheckCis(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Gia_ManFraigMarkCis
// C: Gia_Obj_t * Gia_ManFraigMarkCis( Gia_Man_t * p, Gia_Obj_t * pObj, int fMark )
func (p *GiaManager) GiaManFraigMarkCis(pObj *GiaObject, fMark int) *GiaObject {
	panic("TODO, stubbed")
}

// Gia_ManFraigMarkCos
// C: Gia_Obj_t * Gia_ManFraigMarkCos( Gia_Man_t * p, Gia_Obj_t * pObj, int fMark )
func (p *GiaManager) GiaManFraigMarkCos(pObj *GiaObject, fMark int) *GiaObject {
	panic("TODO, stubbed")
}

// Gia_ManFraigMarkAnd
// C: Gia_Obj_t * Gia_ManFraigMarkAnd( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) GiaManFraigMarkAnd(pObj *GiaObject) *GiaObject {
	panic("TODO, stubbed")
}

// Gia_ManFraigCreateGia
// C: Gia_Man_t * Gia_ManFraigCreateGia( Gia_Man_t * p )
func (p *GiaManager) GiaManFraigCreateGia() *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ObjFanin0CopyRepr
// C: int Gia_ObjFanin0CopyRepr( Gia_Man_t * p, Gia_Obj_t * pObj, int * pReprs )
func (p *GiaManager) GiaObjFanin0CopyRepr(pObj *GiaObject, pReprs []int) int {
	panic("TODO, stubbed")
}

// Gia_ObjFanin1CopyRepr
// C: int Gia_ObjFanin1CopyRepr( Gia_Man_t * p, Gia_Obj_t * pObj, int * pReprs )
func (p *GiaManager) GiaObjFanin1CopyRepr(pObj *GiaObject, pReprs []int) int {
	panic("TODO, stubbed")
}

// Gia_ManFraigReduceGia
// C: Gia_Man_t * Gia_ManFraigReduceGia( Gia_Man_t * p, int * pReprs )
func (p *GiaManager) GiaManFraigReduceGia(pReprs []int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManComputeCarryOuts
// C: Vec_Int_t * Gia_ManComputeCarryOuts( Gia_Man_t * p )
func (p *GiaManager) GiaManComputeCarryOuts() *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManCheckIntegrityWithBoxes
// C: void Gia_ManCheckIntegrityWithBoxes( Gia_Man_t * p )
func (p *GiaManager) GiaManCheckIntegrityWithBoxes() {
	panic("TODO, stubbed")
}

// Gia_ManFraigSelectReprs
// C: int * Gia_ManFraigSelectReprs( Gia_Man_t * p, Gia_Man_t * pClp, int fVerbose, int pFlopTypes[3] )
func (p *GiaManager) GiaManFraigSelectReprs(pClp *GiaManager, fVerbose int, pFlopTypes []int) []int {
	panic("TODO, stubbed")
}

// Gia_ManFraigSweepPerform (pPars is a void * that holds a Cec_ParFra_t *)
// C: void Gia_ManFraigSweepPerform( Gia_Man_t * p, void * pPars )
func (p *GiaManager) GiaManFraigSweepPerform(pPars *CecParFra) {
	panic("TODO, stubbed")
}

// Gia_ManFraigSweepSimple (pPars is a void * that holds a Cec_ParFra_t *)
// C: Gia_Man_t * Gia_ManFraigSweepSimple( Gia_Man_t * p, void * pPars )
func (p *GiaManager) GiaManFraigSweepSimple(pPars *CecParFra) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManSweepComputeOneDomainEquivs (pParsS is a void * that holds a Cec_ParFra_t *)
// C: void Gia_ManSweepComputeOneDomainEquivs( Gia_Man_t * p, Vec_Int_t * vRegClasses, int iDom, void * pParsS, int fConst, int fEquiv, int fVerbose )
func (p *GiaManager) GiaManSweepComputeOneDomainEquivs(vRegClasses *IntVector, iDom int, pParsS *CecParFra, fConst int, fEquiv int, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManSweepWithBoxesAndDomains (pParsS is a void * that holds a Cec_ParFra_t *)
// C: Gia_Man_t * Gia_ManSweepWithBoxesAndDomains( Gia_Man_t * p, void * pParsS, int fConst, int fEquiv, int fVerbose, int fVerbEquivs )
func (p *GiaManager) GiaManSweepWithBoxesAndDomains(pParsS *CecParFra, fConst int, fEquiv int, fVerbose int, fVerbEquivs int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManSweepWithBoxes (pParsC and pParsS are void * that hold Cec_ParCec_t * / Cec_ParFra_t *)
// C: Gia_Man_t * Gia_ManSweepWithBoxes( Gia_Man_t * p, void * pParsC, void * pParsS, int fConst, int fEquiv, int fVerbose, int fVerbEquivs )
func (p *GiaManager) GiaManSweepWithBoxes(pParsC *CecParCec, pParsS *CecParFra, fConst int, fEquiv int, fVerbose int, fVerbEquivs int) *GiaManager {
	panic("TODO, stubbed")
}
