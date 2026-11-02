package abc

// Hand-checked skeleton of src/map/if/ifDec08.c.

// The static data tables BitCount8, PMasks, Truth6 and Truth10 are not represented.

// If_Dec08WordNum
// C: int If_Dec08WordNum( int nVars )
func ifDec08WordNum(nVars int) int {
	panic("TODO, stubbed")
}

// If_Dec08PrintConfigOne
// C: void If_Dec08PrintConfigOne( unsigned z )
func ifDec08PrintConfigOne(z uint32) {
	panic("TODO, stubbed")
}

// If_Dec08PrintConfig
// C: void If_Dec08PrintConfig( unsigned * pZ )
func IfDec08PrintConfig(pZ []uint32) {
	panic("TODO, stubbed")
}

// If_Dec08ComposeLut4
// C: void If_Dec08ComposeLut4( int t, word ** pF, word * pR, int nVars )
func ifDec08ComposeLut4(t int, pF [][]uint64, pR []uint64, nVars int) {
	panic("TODO, stubbed")
}

// If_Dec08Verify
// C: void If_Dec08Verify( word * pF, int nVars, unsigned * pZ )
func IfDec08Verify(pF []uint64, nVars int, pZ []uint32) {
	panic("TODO, stubbed")
}

// If_Dec08CofCount2
// C: int If_Dec08CofCount2( word * pF, int nVars )
func ifDec08CofCount2(pF []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_Dec08CofCount
// C: int If_Dec08CofCount( word * pF, int nVars )
func ifDec08CofCount(pF []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_Dec08Copy
// C: void If_Dec08Copy( word * pOut, word * pIn, int nVars )
func ifDec08Copy(pOut []uint64, pIn []uint64, nVars int) {
	panic("TODO, stubbed")
}

// If_Dec08SwapAdjacent
// C: void If_Dec08SwapAdjacent( word * pOut, word * pIn, int iVar, int nVars )
func ifDec08SwapAdjacent(pOut []uint64, pIn []uint64, iVar int, nVars int) {
	panic("TODO, stubbed")
}

// If_Dec08MoveTo
// C: void If_Dec08MoveTo( word * pF, int nVars, int v, int p, int Pla2Var[], int Var2Pla[] )
func ifDec08MoveTo(pF []uint64, nVars int, v int, p int, Pla2Var []int, Var2Pla []int) {
	panic("TODO, stubbed")
}

// If_DecSuppIsMinBase (duplicate static name in ifDec07.c/ifDec08.c/ifDec10.c)
// C: int If_DecSuppIsMinBase( int Supp )
func ifDecSuppIsMinBase_08(Supp int) int {
	panic("TODO, stubbed")
}

// If_Dec08HasVar
// C: int If_Dec08HasVar( word * t, int nVars, int iVar )
func ifDec08HasVar(t []uint64, nVars int, iVar int) int {
	panic("TODO, stubbed")
}

// If_Dec08Support
// C: int If_Dec08Support( word * t, int nVars )
func ifDec08Support(t []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_Dec08Cofactors
// C: void If_Dec08Cofactors( word * pF, int nVars, int iVar, word * pCof0, word * pCof1 )
func IfDec08Cofactors(pF []uint64, nVars int, iVar int, pCof0 []uint64, pCof1 []uint64) {
	panic("TODO, stubbed")
}

// If_Dec08Count16
// C: int If_Dec08Count16( int Num16 )
func ifDec08Count16(Num16 int) int {
	panic("TODO, stubbed")
}

// If_DecVerifyPerm (duplicate static name in ifDec07.c/ifDec08.c/ifDec10.c)
// C: void If_DecVerifyPerm( int Pla2Var[10], int Var2Pla[10], int nVars )
func ifDecVerifyPerm_08(Pla2Var []int, Var2Pla []int, nVars int) {
	panic("TODO, stubbed")
}

// If_Dec08Perform
// C: int If_Dec08Perform( word * pF, int nVars, int fDerive )
func IfDec08Perform(pF []uint64, nVars int, fDerive int) int {
	panic("TODO, stubbed")
}

// If_CutPerformCheck08
// C: int If_CutPerformCheck08( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr )
func (p *IfMan) IfCutPerformCheck08(pTruth []uint32, nVars int, nLeaves int, pStr string) int {
	panic("TODO, stubbed")
}
