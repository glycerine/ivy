package abc

// Hand-checked skeleton of src/map/if/ifDec10.c.

// The static data tables BitCount8, PMasks, Truth6 and Truth10 are not represented.

// If_Dec10WordNum
// C: int If_Dec10WordNum( int nVars )
func ifDec10WordNum(nVars int) int {
	panic("TODO, stubbed")
}

// If_Dec10PrintConfigOne
// C: void If_Dec10PrintConfigOne( unsigned z )
func ifDec10PrintConfigOne(z uint32) {
	panic("TODO, stubbed")
}

// If_Dec10PrintConfig
// C: void If_Dec10PrintConfig( unsigned * pZ )
func IfDec10PrintConfig(pZ []uint32) {
	panic("TODO, stubbed")
}

// If_Dec10ComposeLut4
// C: void If_Dec10ComposeLut4( int t, word ** pF, word * pR, int nVars )
func ifDec10ComposeLut4(t int, pF [][]uint64, pR []uint64, nVars int) {
	panic("TODO, stubbed")
}

// If_Dec10Verify
// C: void If_Dec10Verify( word * pF, int nVars, unsigned * pZ )
func IfDec10Verify(pF []uint64, nVars int, pZ []uint32) {
	panic("TODO, stubbed")
}

// If_Dec10CofCount2
// C: int If_Dec10CofCount2( word * pF, int nVars )
func ifDec10CofCount2(pF []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_Dec10CofCount
// C: int If_Dec10CofCount( word * pF, int nVars )
func ifDec10CofCount(pF []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_Dec10Copy
// C: void If_Dec10Copy( word * pOut, word * pIn, int nVars )
func ifDec10Copy(pOut []uint64, pIn []uint64, nVars int) {
	panic("TODO, stubbed")
}

// If_Dec10SwapAdjacent
// C: void If_Dec10SwapAdjacent( word * pOut, word * pIn, int iVar, int nVars )
func ifDec10SwapAdjacent(pOut []uint64, pIn []uint64, iVar int, nVars int) {
	panic("TODO, stubbed")
}

// If_Dec10MoveTo
// C: void If_Dec10MoveTo( word * pF, int nVars, int v, int p, int Pla2Var[], int Var2Pla[] )
func ifDec10MoveTo(pF []uint64, nVars int, v int, p int, Pla2Var []int, Var2Pla []int) {
	panic("TODO, stubbed")
}

// If_DecSuppIsMinBase (duplicate static name in ifDec07.c/ifDec08.c/ifDec10.c)
// C: int If_DecSuppIsMinBase( int Supp )
func ifDecSuppIsMinBase_10(Supp int) int {
	panic("TODO, stubbed")
}

// If_Dec10HasVar
// C: int If_Dec10HasVar( word * t, int nVars, int iVar )
func ifDec10HasVar(t []uint64, nVars int, iVar int) int {
	panic("TODO, stubbed")
}

// If_Dec10Support
// C: int If_Dec10Support( word * t, int nVars )
func ifDec10Support(t []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_Dec10Cofactors
// C: void If_Dec10Cofactors( word * pF, int nVars, int iVar, word * pCof0, word * pCof1 )
func IfDec10Cofactors(pF []uint64, nVars int, iVar int, pCof0 []uint64, pCof1 []uint64) {
	panic("TODO, stubbed")
}

// If_Dec10Count16
// C: int If_Dec10Count16( int Num16 )
func ifDec10Count16(Num16 int) int {
	panic("TODO, stubbed")
}

// If_DecVerifyPerm (duplicate static name in ifDec07.c/ifDec08.c/ifDec10.c)
// C: void If_DecVerifyPerm( int Pla2Var[10], int Var2Pla[10], int nVars )
func ifDecVerifyPerm_10(Pla2Var []int, Var2Pla []int, nVars int) {
	panic("TODO, stubbed")
}

// If_Dec10Perform
// C: int If_Dec10Perform( word * pF, int nVars, int fDerive )
func IfDec10Perform(pF []uint64, nVars int, fDerive int) int {
	panic("TODO, stubbed")
}

// If_CutPerformCheck10
// C: int If_CutPerformCheck10( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr )
func (p *IfMan) IfCutPerformCheck10(pTruth []uint32, nVars int, nLeaves int, pStr string) int {
	panic("TODO, stubbed")
}
