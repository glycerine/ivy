package abc

// Hand-checked skeleton of src/map/if/ifDec07.c.

// The static data tables BitCount8, PMasks, Truth6 and Truth7 are not represented.

// If_DecPrintConfig
// C: void If_DecPrintConfig( word z )
func IfDecPrintConfig(z uint64) {
	panic("TODO, stubbed")
}

// If_Dec6ComposeLut4
// C: word If_Dec6ComposeLut4( int t, word f[4] )
func ifDec6ComposeLut4(t int, f []uint64) uint64 {
	panic("TODO, stubbed")
}

// If_Dec6Truth
// C: word If_Dec6Truth( word z )
func IfDec6Truth(z uint64) uint64 {
	panic("TODO, stubbed")
}

// If_Dec6Verify
// C: void If_Dec6Verify( word t, word z )
func IfDec6Verify(t uint64, z uint64) {
	panic("TODO, stubbed")
}

// If_Dec7ComposeLut4
// C: void If_Dec7ComposeLut4( int t, word f[4][2], word r[2] )
func ifDec7ComposeLut4(t int, f *[4][2]uint64, r []uint64) {
	panic("TODO, stubbed")
}

// If_Dec7Verify
// C: void If_Dec7Verify( word t[2], word z )
func IfDec7Verify(t []uint64, z uint64) {
	panic("TODO, stubbed")
}

// If_Dec6CofCount2
// C: int If_Dec6CofCount2( word t )
func ifDec6CofCount2(t uint64) int {
	panic("TODO, stubbed")
}

// If_Dec7CofCount3
// C: int If_Dec7CofCount3( word t[2] )
func ifDec7CofCount3(t []uint64) int {
	panic("TODO, stubbed")
}

// If_Dec6SwapAdjacent
// C: word If_Dec6SwapAdjacent( word t, int v )
func ifDec6SwapAdjacent(t uint64, v int) uint64 {
	panic("TODO, stubbed")
}

// If_Dec6MoveTo
// C: word If_Dec6MoveTo( word t, int v, int p, int Pla2Var[6], int Var2Pla[6] )
func ifDec6MoveTo(t uint64, v int, p int, Pla2Var []int, Var2Pla []int) uint64 {
	panic("TODO, stubbed")
}

// If_Dec7SwapAdjacent
// C: void If_Dec7SwapAdjacent( word t[2], int v )
func ifDec7SwapAdjacent(t []uint64, v int) {
	panic("TODO, stubbed")
}

// If_Dec7MoveTo
// C: void If_Dec7MoveTo( word t[2], int v, int p, int Pla2Var[7], int Var2Pla[7] )
func ifDec7MoveTo(t []uint64, v int, p int, Pla2Var []int, Var2Pla []int) {
	panic("TODO, stubbed")
}

// If_Dec6DeriveCount2
// C: int If_Dec6DeriveCount2( word t, int * pCof0, int * pCof1 )
func ifDec6DeriveCount2(t uint64, pCof0 *int, pCof1 *int) int {
	panic("TODO, stubbed")
}

// If_Dec7DeriveCount3
// C: int If_Dec7DeriveCount3( word t[2], int * pCof0, int * pCof1 )
func ifDec7DeriveCount3(t []uint64, pCof0 *int, pCof1 *int) int {
	panic("TODO, stubbed")
}

// If_Dec6Cofactor
// C: word If_Dec6Cofactor( word t, int iVar, int fCof1 )
func ifDec6Cofactor(t uint64, iVar int, fCof1 int) uint64 {
	panic("TODO, stubbed")
}

// If_Dec6DeriveDisjoint
// C: word If_Dec6DeriveDisjoint( word t, int Pla2Var[6], int Var2Pla[6] )
func ifDec6DeriveDisjoint(t uint64, Pla2Var []int, Var2Pla []int) uint64 {
	panic("TODO, stubbed")
}

// If_Dec6DeriveNonDisjoint
// C: word If_Dec6DeriveNonDisjoint( word t, int s, int Pla2Var0[6], int Var2Pla0[6] )
func ifDec6DeriveNonDisjoint(t uint64, s int, Pla2Var0 []int, Var2Pla0 []int) uint64 {
	panic("TODO, stubbed")
}

// If_Dec7DeriveDisjoint
// C: word If_Dec7DeriveDisjoint( word t[2], int Pla2Var[7], int Var2Pla[7] )
func ifDec7DeriveDisjoint(t []uint64, Pla2Var []int, Var2Pla []int) uint64 {
	panic("TODO, stubbed")
}

// If_Dec6CountOnes
// C: int If_Dec6CountOnes( word t )
func ifDec6CountOnes(t uint64) int {
	panic("TODO, stubbed")
}

// If_Dec6HasVar
// C: int If_Dec6HasVar( word t, int v )
func ifDec6HasVar(t uint64, v int) int {
	panic("TODO, stubbed")
}

// If_Dec7HasVar
// C: int If_Dec7HasVar( word t[2], int v )
func ifDec7HasVar(t []uint64, v int) int {
	panic("TODO, stubbed")
}

// If_DecVerifyPerm (duplicate static name in ifDec08.c/ifDec10.c)
// C: void If_DecVerifyPerm( int Pla2Var[6], int Var2Pla[6] )
func ifDecVerifyPerm_07(Pla2Var []int, Var2Pla []int) {
	panic("TODO, stubbed")
}

// If_Dec6Perform
// C: word If_Dec6Perform( word t, int fDerive )
func IfDec6Perform(t uint64, fDerive int) uint64 {
	panic("TODO, stubbed")
}

// If_Dec7Perform
// C: word If_Dec7Perform( word t0[2], int fDerive )
func IfDec7Perform(t0 []uint64, fDerive int) uint64 {
	panic("TODO, stubbed")
}

// If_DecSuppIsMinBase (duplicate static name in ifDec08.c/ifDec10.c)
// C: int If_DecSuppIsMinBase( int Supp )
func ifDecSuppIsMinBase_07(Supp int) int {
	panic("TODO, stubbed")
}

// If_Dec6TruthShrink
// C: word If_Dec6TruthShrink( word uTruth, int nVars, int nVarsAll, unsigned Phase )
func ifDec6TruthShrink(uTruth uint64, nVars int, nVarsAll int, Phase uint32) uint64 {
	panic("TODO, stubbed")
}

// If_Dec6MinimumBase
// C: word If_Dec6MinimumBase( word uTruth, int * pSupp, int nVarsAll, int * pnVars )
func IfDec6MinimumBase(uTruth uint64, pSupp []int, nVarsAll int, pnVars *int) uint64 {
	panic("TODO, stubbed")
}

// If_Dec7TruthShrink
// C: void If_Dec7TruthShrink( word uTruth[2], int nVars, int nVarsAll, unsigned Phase )
func ifDec7TruthShrink(uTruth []uint64, nVars int, nVarsAll int, Phase uint32) {
	panic("TODO, stubbed")
}

// If_Dec7MinimumBase
// C: void If_Dec7MinimumBase( word uTruth[2], int * pSupp, int nVarsAll, int * pnVars )
func IfDec7MinimumBase(uTruth []uint64, pSupp []int, nVarsAll int, pnVars *int) {
	panic("TODO, stubbed")
}

// If_Dec6SuppSize
// C: int If_Dec6SuppSize( word t )
func ifDec6SuppSize(t uint64) int {
	panic("TODO, stubbed")
}

// If_Dec6CheckMux
// C: int If_Dec6CheckMux( word t )
func ifDec6CheckMux(t uint64) int {
	panic("TODO, stubbed")
}

// If_Dec7Cofactor
// C: void If_Dec7Cofactor( word t[2], int iVar, int fCof1, word r[2] )
func ifDec7Cofactor(t []uint64, iVar int, fCof1 int, r []uint64) {
	panic("TODO, stubbed")
}

// If_Dec7SuppSize
// C: int If_Dec7SuppSize( word t[2] )
func ifDec7SuppSize(t []uint64) int {
	panic("TODO, stubbed")
}

// If_Dec7CheckMux
// C: int If_Dec7CheckMux( word t[2] )
func ifDec7CheckMux(t []uint64) int {
	panic("TODO, stubbed")
}

// If_Dec6PickBestMux
// C: int If_Dec6PickBestMux( word t, word Cofs[2] )
func IfDec6PickBestMux(t uint64, Cofs []uint64) int {
	panic("TODO, stubbed")
}

// If_Dec7PickBestMux
// C: int If_Dec7PickBestMux( word t[2], word c0r[2], word c1r[2] )
func IfDec7PickBestMux(t []uint64, c0r []uint64, c1r []uint64) int {
	panic("TODO, stubbed")
}

// If_Dec5CofCount2
// C: word If_Dec5CofCount2( word t, int x, int y, int * Pla2Var, word t0, int fDerive )
func ifDec5CofCount2(t uint64, x int, y int, Pla2Var []int, t0 uint64, fDerive int) uint64 {
	panic("TODO, stubbed")
}

// If_Dec5Perform
// C: word If_Dec5Perform( word t, int fDerive )
func IfDec5Perform(t uint64, fDerive int) uint64 {
	panic("TODO, stubbed")
}

// If_Dec5PerformEx
// C: word If_Dec5PerformEx()
func IfDec5PerformEx() uint64 {
	panic("TODO, stubbed")
}

// If_Dec5PerformTest
// C: void If_Dec5PerformTest()
func IfDec5PerformTest() {
	panic("TODO, stubbed")
}

// If_CutPerformDerive07
// C: word If_CutPerformDerive07( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr )
func (p *IfMan) IfCutPerformDerive07(pTruth []uint32, nVars int, nLeaves int, pStr string) uint64 {
	panic("TODO, stubbed")
}

// If_CutPerformCheck07
// C: int If_CutPerformCheck07( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr )
func (p *IfMan) IfCutPerformCheck07(pTruth []uint32, nVars int, nLeaves int, pStr string) int {
	panic("TODO, stubbed")
}
