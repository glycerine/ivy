package abc

// Hand-checked skeleton of src/base/wlc/wlcSim.c.

// The statics here share names with statics in wlcGraft.c; wlcGraft.c carries the
// _graft suffix, these are the unsuffixed ones.

// Wlc_ObjSim (static; a pointer into the manager's simulation words)
// C: word * Wlc_ObjSim( Gia_Man_t * p, int iObj )
func wlcObjSim(p *GiaManager, iObj int) []uint64 {
	panic("TODO, stubbed")
}

// Wlc_ObjSimPi (static)
// C: void Wlc_ObjSimPi( Gia_Man_t * p, int iObj )
func wlcObjSimPi(p *GiaManager, iObj int) {
	panic("TODO, stubbed")
}

// Wlc_ObjSimRo (static)
// C: void Wlc_ObjSimRo( Gia_Man_t * p, int iObj )
func wlcObjSimRo(p *GiaManager, iObj int) {
	panic("TODO, stubbed")
}

// Wlc_ObjSimCo (static)
// C: void Wlc_ObjSimCo( Gia_Man_t * p, int iObj )
func wlcObjSimCo(p *GiaManager, iObj int) {
	panic("TODO, stubbed")
}

// Wlc_ObjSimAnd (static)
// C: void Wlc_ObjSimAnd( Gia_Man_t * p, int iObj )
func wlcObjSimAnd(p *GiaManager, iObj int) {
	panic("TODO, stubbed")
}

// Wlc_NtkDeleteSim (the Vec_Ptr_t holds word * simulation rows)
// C: void Wlc_NtkDeleteSim( Vec_Ptr_t * p )
func WlcNtkDeleteSim(p *PointerVector[uint64]) {
	panic("TODO, stubbed")
}

// Wlc_NtkSimulate (returns word * simulation rows)
// C: Vec_Ptr_t * Wlc_NtkSimulate( Wlc_Ntk_t * p, Vec_Int_t * vNodes, int nWords, int nFrames )
func WlcNtkSimulate(p *WLCNetwork, vNodes *IntVector, nWords int, nFrames int) *PointerVector[uint64] {
	panic("TODO, stubbed")
}

// Wlc_NtkSimulatePrint
// C: void Wlc_NtkSimulatePrint( Wlc_Ntk_t * p, Vec_Int_t * vNodes, Vec_Ptr_t * vRes, int nWords, int nFrames )
func WlcNtkSimulatePrint(p *WLCNetwork, vNodes *IntVector, vRes *PointerVector[uint64], nWords int, nFrames int) {
	panic("TODO, stubbed")
}

// Wlc_NtkSimulateTest
// C: void Wlc_NtkSimulateTest( Wlc_Ntk_t * p )
func WlcNtkSimulateTest(p *WLCNetwork) {
	panic("TODO, stubbed")
}
