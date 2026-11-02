package abc

// Hand-checked skeleton of src/opt/mfs/mfsGia.c.

// Gia_ObjChild0Copy (static)
// C: int Gia_ObjChild0Copy( Aig_Obj_t * pObj )
func giaObjChild0Copy(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Gia_ObjChild1Copy (static)
// C: int Gia_ObjChild1Copy( Aig_Obj_t * pObj )
func giaObjChild1Copy(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Gia_ManCreateResubMiter
// C: Gia_Man_t * Gia_ManCreateResubMiter( Aig_Man_t * p )
func (p *AigManager) GiaManCreateResubMiter() *GiaManager {
	panic("TODO, stubbed")
}

// Abc_NtkMfsConstructGia
// C: void Abc_NtkMfsConstructGia( Mfs_Man_t * p )
func (p *MfsMan) AbcNtkMfsConstructGia() {
	panic("TODO, stubbed")
}

// Abc_NtkMfsDeconstructGia
// C: void Abc_NtkMfsDeconstructGia( Mfs_Man_t * p )
func (p *MfsMan) AbcNtkMfsDeconstructGia() {
	panic("TODO, stubbed")
}

// Abc_NtkMfsResimulate
// C: void Abc_NtkMfsResimulate( Gia_Man_t * p, Vec_Int_t * vCex )
func (p *GiaManager) AbcNtkMfsResimulate(vCex *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkMfsTryResubOnceGia
// C: int Abc_NtkMfsTryResubOnceGia( Mfs_Man_t * p, int * pCands, int nCands )
func (p *MfsMan) AbcNtkMfsTryResubOnceGia(pCands []int, nCands int) int {
	panic("TODO, stubbed")
}
