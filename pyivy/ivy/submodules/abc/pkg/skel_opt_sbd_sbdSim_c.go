package abc

// Hand-checked skeleton of src/opt/sbd/sbdSim.c.

// Sbd_ObjSims (static; a pointer into the manager's simulation words)
// C: word * Sbd_ObjSims( Gia_Man_t * p, int i )
func sbdObjSims(p *GiaManager, i int) []uint64 {
	panic("TODO, stubbed")
}

// Sbd_ObjCtrl (static; a pointer into the manager's control words)
// C: word * Sbd_ObjCtrl( Gia_Man_t * p, int i )
func sbdObjCtrl(p *GiaManager, i int) []uint64 {
	panic("TODO, stubbed")
}

// Sbd_GiaSimRoundBack2
// C: void Sbd_GiaSimRoundBack2( Gia_Man_t * p )
func (p *GiaManager) SbdGiaSimRoundBack2() {
	panic("TODO, stubbed")
}

// Sbd_GiaSatOne_rec
// C: int Sbd_GiaSatOne_rec( Gia_Man_t * p, Gia_Obj_t * pObj, int Value, int fFirst, int iPat )
func (p *GiaManager) SbdGiaSatOne_rec(pObj *GiaObject, Value int, fFirst int, iPat int) int {
	panic("TODO, stubbed")
}

// Sbd_GiaSatOne
// C: int Sbd_GiaSatOne( Gia_Man_t * p, Vec_Int_t * vPairs )
func (p *GiaManager) SbdGiaSatOne(vPairs *IntVector) int {
	panic("TODO, stubbed")
}

// Sbd_GiaSimRoundBack
// C: void Sbd_GiaSimRoundBack( Gia_Man_t * p )
func (p *GiaManager) SbdGiaSimRoundBack() {
	panic("TODO, stubbed")
}

// Sbd_GiaSimRound (pvMap is an out parameter)
// C: void Sbd_GiaSimRound( Gia_Man_t * p, int fTry, Vec_Int_t ** pvMap )
func (p *GiaManager) SbdGiaSimRound(fTry int, pvMap **IntVector) {
	panic("TODO, stubbed")
}

// Sbd_GiaSimTest
// C: void Sbd_GiaSimTest( Gia_Man_t * pGia )
func (pGia *GiaManager) SbdGiaSimTest() {
	panic("TODO, stubbed")
}
