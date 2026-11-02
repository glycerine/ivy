package abc

// Hand-checked skeleton of src/aig/gia/giaBalMap.c.

// Gia_ManHighlight_rec
// C: void Gia_ManHighlight_rec( Gia_Man_t * p, int iObj )
func (p *GiaManager) GiaManHighlight_rec(iObj int) {
	panic("TODO, stubbed")
}

// Gia_ManPrepareWin
// C: void Gia_ManPrepareWin( Gia_Man_t * p, Vec_Int_t * vOuts, Vec_Int_t ** pvPis, Vec_Int_t ** pvPos, Vec_Int_t ** pvAnds, int fPoOnly )
func (p *GiaManager) GiaManPrepareWin(vOuts *IntVector, pvPis **IntVector, pvPos **IntVector, pvAnds **IntVector, fPoOnly int) {
	panic("TODO, stubbed")
}

// Gia_ManExtractWin
// C: Gia_Man_t * Gia_ManExtractWin( Gia_Man_t * p, Vec_Int_t * vOuts, int fPoOnly )
func (p *GiaManager) GiaManExtractWin(vOuts *IntVector, fPoOnly int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManInsertWin
// C: Gia_Man_t * Gia_ManInsertWin( Gia_Man_t * p, Vec_Int_t * vOuts, Gia_Man_t * pWin )
func (p *GiaManager) GiaManInsertWin(vOuts *IntVector, pWin *GiaManager) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManFindLatest
// C: Vec_Int_t * Gia_ManFindLatest( Gia_Man_t * p, int LevelMax, int nTimeWindow )
func (p *GiaManager) GiaManFindLatest(LevelMax int, nTimeWindow int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManExtractWindow
// C: Gia_Man_t * Gia_ManExtractWindow( Gia_Man_t * p, int LevelMax, int nTimeWindow, int fVerbose )
func (p *GiaManager) GiaManExtractWindow(LevelMax int, nTimeWindow int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManPerformSopBalanceWin
// C: Gia_Man_t * Gia_ManPerformSopBalanceWin( Gia_Man_t * p, int LevelMax, int nTimeWindow, int nCutNum, int nRelaxRatio, int fVerbose )
func (p *GiaManager) GiaManPerformSopBalanceWin(LevelMax int, nTimeWindow int, nCutNum int, nRelaxRatio int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManPerformDsdBalanceWin
// C: Gia_Man_t * Gia_ManPerformDsdBalanceWin( Gia_Man_t * p, int LevelMax, int nTimeWindow, int nLutSize, int nCutNum, int nRelaxRatio, int fVerbose )
func (p *GiaManager) GiaManPerformDsdBalanceWin(LevelMax int, nTimeWindow int, nLutSize int, nCutNum int, nRelaxRatio int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
