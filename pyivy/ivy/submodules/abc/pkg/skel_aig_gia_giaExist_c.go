package abc

// Hand-checked skeleton of src/aig/gia/giaExist.c.

// Gia_ManQuantInfoId (static inline)
// C: word * Gia_ManQuantInfoId( Gia_Man_t * p, int iObj )
func (p *GiaManager) giaManQuantInfoId(iObj int) []uint64 {
	panic("TODO, stubbed")
}

// Gia_ManQuantInfo (static inline)
// C: word * Gia_ManQuantInfo( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) giaManQuantInfo(pObj *GiaObject) []uint64 {
	panic("TODO, stubbed")
}

// Gia_ObjCopyGetTwoArray (static inline)
// C: void Gia_ObjCopyGetTwoArray( Gia_Man_t * p, int iObj, int LitsOut[2] )
func (p *GiaManager) giaObjCopyGetTwoArray(iObj int, LitsOut []int) {
	panic("TODO, stubbed")
}

// Gia_ObjCopySetTwoArray (static inline)
// C: void Gia_ObjCopySetTwoArray( Gia_Man_t * p, int iObj, int LitsIn[2] )
func (p *GiaManager) giaObjCopySetTwoArray(iObj int, LitsIn []int) {
	panic("TODO, stubbed")
}

// Gia_ManQuantVerify_rec
// C: int Gia_ManQuantVerify_rec( Gia_Man_t * p, int iObj, int CiId )
func (p *GiaManager) GiaManQuantVerify_rec(iObj int, CiId int) int {
	panic("TODO, stubbed")
}

// Gia_ManQuantVerify
// C: void Gia_ManQuantVerify( Gia_Man_t * p, int iObj )
func (p *GiaManager) GiaManQuantVerify(iObj int) {
	panic("TODO, stubbed")
}

// Gia_ManQuantSetSuppStart
// C: void Gia_ManQuantSetSuppStart( Gia_Man_t * p )
func (p *GiaManager) GiaManQuantSetSuppStart() {
	panic("TODO, stubbed")
}

// Gia_ManQuantSetSuppZero
// C: void Gia_ManQuantSetSuppZero( Gia_Man_t * p )
func (p *GiaManager) GiaManQuantSetSuppZero() {
	panic("TODO, stubbed")
}

// Gia_ManQuantSetSuppCi
// C: void Gia_ManQuantSetSuppCi( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) GiaManQuantSetSuppCi(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManQuantSetSuppAnd
// C: void Gia_ManQuantSetSuppAnd( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) GiaManQuantSetSuppAnd(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManQuantCheckSupp
// C: int Gia_ManQuantCheckSupp( Gia_Man_t * p, int iObj, int iSupp )
func (p *GiaManager) GiaManQuantCheckSupp(iObj int, iSupp int) int {
	panic("TODO, stubbed")
}

// Gia_ManQuantUpdateCiSupp
// C: void Gia_ManQuantUpdateCiSupp( Gia_Man_t * p, int iObj )
func (p *GiaManager) GiaManQuantUpdateCiSupp(iObj int) {
	panic("TODO, stubbed")
}

// Gia_ManQuantCheckOverlap
// C: int Gia_ManQuantCheckOverlap( Gia_Man_t * p, int iObj )
func (p *GiaManager) GiaManQuantCheckOverlap(iObj int) int {
	panic("TODO, stubbed")
}

// Gia_ManQuantMarkUsedCis (pData is a void *)
// C: void Gia_ManQuantMarkUsedCis( Gia_Man_t * p, int(*pFuncCiToKeep)(void *, int), void * pData )
func (p *GiaManager) GiaManQuantMarkUsedCis(pFuncCiToKeep func(any, int) int, pData any) {
	panic("TODO, stubbed")
}

// Gia_ManQuantCountUsed_rec
// C: int Gia_ManQuantCountUsed_rec( Gia_Man_t * p, int iObj )
func (p *GiaManager) GiaManQuantCountUsed_rec(iObj int) int {
	panic("TODO, stubbed")
}

// Gia_ManQuantCountUsed
// C: int Gia_ManQuantCountUsed( Gia_Man_t * p, int iObj )
func (p *GiaManager) GiaManQuantCountUsed(iObj int) int {
	panic("TODO, stubbed")
}

// Gia_ManQuantDupConeSupp_rec (pData is a void *)
// C: void Gia_ManQuantDupConeSupp_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vCis, Vec_Int_t * vObjs, int(*pFuncCiToKeep)(void *, int), void * pData )
func (pNew *GiaManager) GiaManQuantDupConeSupp_rec(p *GiaManager, pObj *GiaObject, vCis *IntVector, vObjs *IntVector, pFuncCiToKeep func(any, int) int, pData any) {
	panic("TODO, stubbed")
}

// Gia_ManQuantDupConeSupp (pData is a void *)
// C: Gia_Man_t * Gia_ManQuantDupConeSupp( Gia_Man_t * p, int iLit, int(*pFuncCiToKeep)(void *, int), void * pData, Vec_Int_t ** pvCis, int * pOutLit )
func (p *GiaManager) GiaManQuantDupConeSupp(iLit int, pFuncCiToKeep func(any, int) int, pData any, pvCis **IntVector, pOutLit *int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManQuantExist_rec
// C: void Gia_ManQuantExist_rec( Gia_Man_t * p, int iObj, int pRes[2] )
func (p *GiaManager) GiaManQuantExist_rec(iObj int, pRes []int) {
	panic("TODO, stubbed")
}

// Gia_ManQuantExist2 (pData is a void *)
// C: int Gia_ManQuantExist2( Gia_Man_t * p0, int iLit, int(*pFuncCiToKeep)(void *, int), void * pData )
func (p0 *GiaManager) GiaManQuantExist2(iLit int, pFuncCiToKeep func(any, int) int, pData any) int {
	panic("TODO, stubbed")
}

// Gia_ManQuantCollect_rec
// C: void Gia_ManQuantCollect_rec( Gia_Man_t * p, int iObj, Vec_Int_t * vQuantCis, Vec_Int_t * vQuantSide, Vec_Int_t * vQuantAnds )
func (p *GiaManager) GiaManQuantCollect_rec(iObj int, vQuantCis *IntVector, vQuantSide *IntVector, vQuantAnds *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManQuantCollect (pData is a void *)
// C: void Gia_ManQuantCollect( Gia_Man_t * p, int iObj, int(*pFuncCiToKeep)(void *, int), void * pData, Vec_Int_t * vQuantCis, Vec_Int_t * vQuantSide, Vec_Int_t * vQuantAnds )
func (p *GiaManager) GiaManQuantCollect(iObj int, pFuncCiToKeep func(any, int) int, pData any, vQuantCis *IntVector, vQuantSide *IntVector, vQuantAnds *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManQuantExist2Dup
// C: Gia_Man_t * Gia_ManQuantExist2Dup( Gia_Man_t * p, int iLit, Vec_Int_t * vCis, Vec_Int_t * vSide, Vec_Int_t * vAnds, int * pOutLit )
func (p *GiaManager) GiaManQuantExist2Dup(iLit int, vCis *IntVector, vSide *IntVector, vAnds *IntVector, pOutLit *int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManQuantExistInt
// C: int Gia_ManQuantExistInt( Gia_Man_t * p0, int iLit, Vec_Int_t * vCis, Vec_Int_t * vSide, Vec_Int_t * vAnds )
func (p0 *GiaManager) GiaManQuantExistInt(iLit int, vCis *IntVector, vSide *IntVector, vAnds *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManQuantExist (pData is a void *)
// C: int Gia_ManQuantExist( Gia_Man_t * p0, int iLit, int(*pFuncCiToKeep)(void *, int), void * pData )
func (p0 *GiaManager) GiaManQuantExist(iLit int, pFuncCiToKeep func(any, int) int, pData any) int {
	panic("TODO, stubbed")
}
