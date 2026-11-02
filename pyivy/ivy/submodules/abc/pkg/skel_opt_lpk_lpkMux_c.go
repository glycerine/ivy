package abc

// Hand-checked skeleton of src/opt/lpk/lpkMux.c.

// Lpk_MapTreeBestCofVar
// C: int Lpk_MapTreeBestCofVar( Lpk_Man_t * p, unsigned * pTruth, int nVars, unsigned * pCof0, unsigned * pCof1 )
func (p *LpkMan) LpkMapTreeBestCofVar(pTruth []uint32, nVars int, pCof0 []uint32, pCof1 []uint32) int {
	panic("TODO, stubbed")
}

// Lpk_MapTreeMux_rec
// C: If_Obj_t * Lpk_MapTreeMux_rec( Lpk_Man_t * p, unsigned * pTruth, int nVars, If_Obj_t ** ppLeaves )
func (p *LpkMan) LpkMapTreeMux_rec(pTruth []uint32, nVars int, ppLeaves []*IfObj) *IfObj {
	panic("TODO, stubbed")
}

// Lpk_MapSuppRedDec_rec
// C: If_Obj_t * Lpk_MapSuppRedDec_rec( Lpk_Man_t * p, unsigned * pTruth, int nVars, If_Obj_t ** ppLeaves )
func (p *LpkMan) LpkMapSuppRedDec_rec(pTruth []uint32, nVars int, ppLeaves []*IfObj) *IfObj {
	panic("TODO, stubbed")
}

// Lpk_TryDec
// C: void Lpk_TryDec( word * pTruth, int nVars )
func LpkTryDec(pTruth []uint64, nVars int) {
	panic("TODO, stubbed")
}

// Lpk_TryDecTest
// C: void Lpk_TryDecTest()
func LpkTryDecTest() {
	panic("TODO, stubbed")
}
