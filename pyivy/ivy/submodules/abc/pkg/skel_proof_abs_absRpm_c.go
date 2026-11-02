package abc

// Hand-checked skeleton of src/proof/abs/absRpm.c.

// Gia_ObjDom (static inline)
// C: int Gia_ObjDom( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) giaObjDom(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Gia_ObjSetDom (static inline)
// C: void Gia_ObjSetDom( Gia_Man_t * p, Gia_Obj_t * pObj, int d )
func (p *GiaManager) giaObjSetDom(pObj *GiaObject, d int) {
	panic("TODO, stubbed")
}

// Gia_ManAddDom
// C: void Gia_ManAddDom( Gia_Man_t * p, Gia_Obj_t * pObj, int iDom0 )
func (p *GiaManager) GiaManAddDom(pObj *GiaObject, iDom0 int) {
	panic("TODO, stubbed")
}

// Gia_ManComputeDoms
// C: void Gia_ManComputeDoms( Gia_Man_t * p )
func (p *GiaManager) GiaManComputeDoms() {
	panic("TODO, stubbed")
}

// Gia_ManCreateSupps
// C: Vec_Wec_t * Gia_ManCreateSupps( Gia_Man_t * p, int fVerbose )
func (p *GiaManager) GiaManCreateSupps(fVerbose int) *IntLevelVector {
	panic("TODO, stubbed")
}

// Gia_ManDomTest
// C: void Gia_ManDomTest( Gia_Man_t * p )
func (p *GiaManager) GiaManDomTest() {
	panic("TODO, stubbed")
}

// Gia_ManTestDoms2
// C: void Gia_ManTestDoms2( Gia_Man_t * p )
func (p *GiaManager) GiaManTestDoms2() {
	panic("TODO, stubbed")
}

// Gia_ManCollectDoms
// C: Vec_Int_t * Gia_ManCollectDoms( Gia_Man_t * p )
func (p *GiaManager) GiaManCollectDoms() *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManComputePiDoms
// C: Vec_Int_t * Gia_ManComputePiDoms( Gia_Man_t * p )
func (p *GiaManager) GiaManComputePiDoms() *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManTestDoms
// C: void Gia_ManTestDoms( Gia_Man_t * p )
func (p *GiaManager) GiaManTestDoms() {
	panic("TODO, stubbed")
}

// Gia_ManCountFanoutlessFlops
// C: void Gia_ManCountFanoutlessFlops( Gia_Man_t * p )
func (p *GiaManager) GiaManCountFanoutlessFlops() {
	panic("TODO, stubbed")
}

// Gia_ManCountPisNodes_rec
// C: void Gia_ManCountPisNodes_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vPis, Vec_Int_t * vAnds )
func (p *GiaManager) GiaManCountPisNodes_rec(pObj *GiaObject, vPis *IntVector, vAnds *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManCountPisNodes
// C: void Gia_ManCountPisNodes( Gia_Man_t * p, Vec_Int_t * vPis, Vec_Int_t * vAnds )
func (p *GiaManager) GiaManCountPisNodes(vPis *IntVector, vAnds *IntVector) {
	panic("TODO, stubbed")
}

// Abs_GiaObjDeref_rec (static (forward-declared static))
// C: int Abs_GiaObjDeref_rec( Gia_Man_t * p, Gia_Obj_t * pNode )
func (p *GiaManager) absGiaObjDeref_rec(pNode *GiaObject) int {
	panic("TODO, stubbed")
}

// Abs_GiaObjRef_rec (static (forward-declared static))
// C: int Abs_GiaObjRef_rec( Gia_Man_t * p, Gia_Obj_t * pNode )
func (p *GiaManager) absGiaObjRef_rec(pNode *GiaObject) int {
	panic("TODO, stubbed")
}

// Abs_GiaSortNodes
// C: int Abs_GiaSortNodes( Gia_Man_t * p, Vec_Int_t * vSupp )
func (p *GiaManager) AbsGiaSortNodes(vSupp *IntVector) int {
	panic("TODO, stubbed")
}

// Abs_ManSupport1_rec
// C: void Abs_ManSupport1_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSupp )
func (p *GiaManager) AbsManSupport1_rec(pObj *GiaObject, vSupp *IntVector) {
	panic("TODO, stubbed")
}

// Abs_ManSupport1
// C: int Abs_ManSupport1( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSupp )
func (p *GiaManager) AbsManSupport1(pObj *GiaObject, vSupp *IntVector) int {
	panic("TODO, stubbed")
}

// Abs_ManSupport2_rec
// C: void Abs_ManSupport2_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSupp )
func (p *GiaManager) AbsManSupport2_rec(pObj *GiaObject, vSupp *IntVector) {
	panic("TODO, stubbed")
}

// Abs_ManSupport2 (static (forward-declared static))
// C: int Abs_ManSupport2( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSupp )
func (p *GiaManager) absManSupport2(pObj *GiaObject, vSupp *IntVector) int {
	panic("TODO, stubbed")
}

// Abs_ManSupport3
// C: int Abs_ManSupport3( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSupp )
func (p *GiaManager) AbsManSupport3(pObj *GiaObject, vSupp *IntVector) int {
	panic("TODO, stubbed")
}

// Abs_GiaCofPrint
// C: int Abs_GiaCofPrint( word * pTruth, int nSize, int nSize0, int Res )
func AbsGiaCofPrint(pTruth []uint64, nSize int, nSize0 int, Res int) int {
	panic("TODO, stubbed")
}

// Abs_GiaCheckTruth
// C: int Abs_GiaCheckTruth( word * pTruth, int nSize, int nSize0 )
func AbsGiaCheckTruth(pTruth []uint64, nSize int, nSize0 int) int {
	panic("TODO, stubbed")
}

// Abs_RpmPerformMark
// C: void Abs_RpmPerformMark( Gia_Man_t * p, int nCutMax, int fVerbose, int fVeryVerbose )
func (p *GiaManager) AbsRpmPerformMark(nCutMax int, fVerbose int, fVeryVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManDupRpm
// C: Gia_Man_t * Gia_ManDupRpm( Gia_Man_t * p )
func (p *GiaManager) GiaManDupRpm() *GiaManager {
	panic("TODO, stubbed")
}

// Abs_RpmPerform
// C: Gia_Man_t * Abs_RpmPerform( Gia_Man_t * p, int nCutMax, int fVerbose, int fVeryVerbose )
func (p *GiaManager) AbsRpmPerform(nCutMax int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
