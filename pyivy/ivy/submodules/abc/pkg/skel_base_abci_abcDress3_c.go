package abc

// Hand-checked skeleton of src/base/abci/abcDress3.c.

// Abc_ConvertHopToGia_rec1
// C: void Abc_ConvertHopToGia_rec1( Gia_Man_t * p, Hop_Obj_t * pObj )
func (p *GiaManager) AbcConvertHopToGia_rec1(pObj *HopObject) {
	panic("TODO, stubbed")
}

// Abc_ConvertHopToGia_rec2
// C: void Abc_ConvertHopToGia_rec2( Hop_Obj_t * pObj )
func AbcConvertHopToGia_rec2(pObj *HopObject) {
	panic("TODO, stubbed")
}

// Abc_ConvertHopToGia
// C: int Abc_ConvertHopToGia( Gia_Man_t * p, Hop_Obj_t * pRoot )
func (p *GiaManager) AbcConvertHopToGia(pRoot *HopObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkAigToGiaOne
// C: void Abc_NtkAigToGiaOne( Gia_Man_t * p, Abc_Ntk_t * pNtk, Vec_Int_t * vMap )
func (p *GiaManager) AbcNtkAigToGiaOne(pNtk *ABCNetwork, vMap *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkAigToGiaTwo
// C: Gia_Man_t * Abc_NtkAigToGiaTwo( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int fByName )
func (pNtk1 *ABCNetwork) AbcNtkAigToGiaTwo(pNtk2 *ABCNetwork, fByName int) *GiaManager {
	panic("TODO, stubbed")
}

// Abc_NtkCollectAddOne (static)
// C: void Abc_NtkCollectAddOne( int iNtk, int iObj, int iGiaLit, Gia_Man_t * pGia, Vec_Int_t * vGia2Cla, Vec_Int_t * vNexts[2] )
func abcNtkCollectAddOne(iNtk int, iObj int, iGiaLit int, pGia *GiaManager, vGia2Cla *IntVector, vNexts []*IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkCollectEquivClasses
// C: Vec_Int_t * Abc_NtkCollectEquivClasses( Abc_Ntk_t * pNtks[2], Gia_Man_t * pGia )
func AbcNtkCollectEquivClasses(pNtks []*ABCNetwork, pGia *GiaManager) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkDumpEquivFile
// C: void Abc_NtkDumpEquivFile( char * pFileName, Vec_Int_t * vClasses, Abc_Ntk_t * pNtks[2] )
func AbcNtkDumpEquivFile(pFileName string, vClasses *IntVector, pNtks []*ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NtkDumpEquiv
// C: void Abc_NtkDumpEquiv( Abc_Ntk_t * pNtks[2], char * pFileName, int nConfs, int fByName, int fVerbose )
func AbcNtkDumpEquiv(pNtks []*ABCNetwork, pFileName string, nConfs int, fByName int, fVerbose int) {
	panic("TODO, stubbed")
}
