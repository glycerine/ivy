package abc

// Hand-checked skeleton of src/aig/gia/giaBidec.c.

// Gia_ManConvertAigToTruth_rec
// C: unsigned * Gia_ManConvertAigToTruth_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vTruth, int nWords, Vec_Int_t * vVisited )
func (p *GiaManager) GiaManConvertAigToTruth_rec(pObj *GiaObject, vTruth *IntVector, nWords int, vVisited *IntVector) []uint32 {
	panic("TODO, stubbed")
}

// Gia_ManConvertAigToTruth
// C: unsigned * Gia_ManConvertAigToTruth( Gia_Man_t * p, Gia_Obj_t * pRoot, Vec_Int_t * vLeaves, Vec_Int_t * vTruth, Vec_Int_t * vVisited )
func (p *GiaManager) GiaManConvertAigToTruth(pRoot *GiaObject, vLeaves *IntVector, vTruth *IntVector, vVisited *IntVector) []uint32 {
	panic("TODO, stubbed")
}

// Gia_ObjPerformBidec
// C: int Gia_ObjPerformBidec( Bdc_Man_t * pManDec, Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pRoot, Vec_Int_t * vLeaves, Vec_Int_t * vTruth, Vec_Int_t * vVisited )
func GiaObjPerformBidec(pManDec *BdcTable, pNew *GiaManager, p *GiaManager, pRoot *GiaObject, vLeaves *IntVector, vTruth *IntVector, vVisited *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManPerformBidec
// C: Gia_Man_t * Gia_ManPerformBidec( Gia_Man_t * p, int fVerbose )
func (p *GiaManager) GiaManPerformBidec(fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
