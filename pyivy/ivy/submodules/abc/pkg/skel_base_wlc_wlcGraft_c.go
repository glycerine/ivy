package abc

// Hand-checked skeleton of src/base/wlc/wlcGraft.c.

// Wlc_ObjSim (static; shares its name with a static in wlcSim.c, so this one carries a _graft suffix; a pointer into the manager's simulation words)
// C: word * Wlc_ObjSim( Gia_Man_t * p, int iObj )
func wlcObjSim_graft(p *GiaManager, iObj int) []uint64 {
	panic("TODO, stubbed")
}

// Wlc_ObjSimPi (static; _graft suffix as above)
// C: void Wlc_ObjSimPi( Gia_Man_t * p, int iObj )
func wlcObjSimPi_graft(p *GiaManager, iObj int) {
	panic("TODO, stubbed")
}

// Wlc_ObjSimRo (static; _graft suffix as above)
// C: void Wlc_ObjSimRo( Gia_Man_t * p, int iObj )
func wlcObjSimRo_graft(p *GiaManager, iObj int) {
	panic("TODO, stubbed")
}

// Wlc_ObjSimCo (static; _graft suffix as above)
// C: void Wlc_ObjSimCo( Gia_Man_t * p, int iObj )
func wlcObjSimCo_graft(p *GiaManager, iObj int) {
	panic("TODO, stubbed")
}

// Wlc_ObjSimAnd (static; _graft suffix as above)
// C: void Wlc_ObjSimAnd( Gia_Man_t * p, int iObj )
func wlcObjSimAnd_graft(p *GiaManager, iObj int) {
	panic("TODO, stubbed")
}

// Wlc_NtkCollectObjs_rec
// C: int Wlc_NtkCollectObjs_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, Vec_Int_t * vObjs )
func WlcNtkCollectObjs_rec(p *WLCNetwork, pObj *WLCObject, vObjs *IntVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkCollectObjs (pCount is an out parameter)
// C: Vec_Int_t * Wlc_NtkCollectObjs( Wlc_Ntk_t * p, int fEven, int * pCount )
func WlcNtkCollectObjs(p *WLCNetwork, fEven int, pCount *int) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkSaveOneNode
// C: void Wlc_NtkSaveOneNode( Wlc_Ntk_t * p, Wlc_Obj_t * pWlcObj, Gia_Man_t * pGia, Vec_Mem_t * vTtMem )
func WlcNtkSaveOneNode(p *WLCNetwork, pWlcObj *WLCObject, pGia *GiaManager, vTtMem *WordMemory) {
	panic("TODO, stubbed")
}

// Wlc_NtkFindOneNode
// C: void Wlc_NtkFindOneNode( Wlc_Ntk_t * p, Wlc_Obj_t * pWlcObj, Gia_Man_t * pGia, Vec_Mem_t * vTtMem )
func WlcNtkFindOneNode(p *WLCNetwork, pWlcObj *WLCObject, pGia *GiaManager, vTtMem *WordMemory) {
	panic("TODO, stubbed")
}

// Wlc_NtkGraftMulti
// C: Wlc_Ntk_t * Wlc_NtkGraftMulti( Wlc_Ntk_t * p, int fVerbose )
func WlcNtkGraftMulti(p *WLCNetwork, fVerbose int) *WLCNetwork {
	panic("TODO, stubbed")
}

// Sbc_Mult (r is filled in)
// C: void Sbc_Mult( word a, word b, word r[2] )
func SbcMult(a uint64, b uint64, r *[2]uint64) {
	panic("TODO, stubbed")
}

// Sbc_SimMult (R is filled in)
// C: void Sbc_SimMult( word A[64], word B[64], word R[128], int nIns )
func SbcSimMult(A *[64]uint64, B *[64]uint64, R *[128]uint64, nIns int) {
	panic("TODO, stubbed")
}

// Sbc_ManDetectMult
// C: Vec_Int_t * Sbc_ManDetectMult( Gia_Man_t * p, Vec_Int_t * vIns )
func (p *GiaManager) SbcManDetectMult(vIns *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Sbc_ManWlcNodes2
// C: Vec_Int_t * Sbc_ManWlcNodes2( Wlc_Ntk_t * pNtk, Gia_Man_t * p, Vec_Int_t * vGiaLits )
func SbcManWlcNodes2(pNtk *WLCNetwork, p *GiaManager, vGiaLits *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Sbc_ManWlcNodes
// C: int Sbc_ManWlcNodes( Wlc_Ntk_t * pNtk, Gia_Man_t * p, Vec_Int_t * vGia2Out, int nOuts )
func SbcManWlcNodes(pNtk *WLCNetwork, p *GiaManager, vGia2Out *IntVector, nOuts int) int {
	panic("TODO, stubbed")
}

// Sbc_ManDetectMultTest
// C: void Sbc_ManDetectMultTest( Wlc_Ntk_t * pNtk, int fVerbose )
func SbcManDetectMultTest(pNtk *WLCNetwork, fVerbose int) {
	panic("TODO, stubbed")
}
