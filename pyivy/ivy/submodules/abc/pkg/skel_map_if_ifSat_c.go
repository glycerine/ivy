package abc

// Hand-checked skeleton of src/map/if/ifSat.c.

// If_ManSatBuildXY (returns a void * that holds a sat_solver *)
// C: void * If_ManSatBuildXY( int nLutSize )
func IfManSatBuildXY(nLutSize int) *SatSolver {
	panic("TODO, stubbed")
}

// If_ManSatBuildXYZ (returns a void * that holds a sat_solver *)
// C: void * If_ManSatBuildXYZ( int nLutSize )
func IfManSatBuildXYZ(nLutSize int) *SatSolver {
	panic("TODO, stubbed")
}

// If_ManSatUnbuild (p is a void * that holds a sat_solver *)
// C: void If_ManSatUnbuild( void * p )
func IfManSatUnbuild(p *SatSolver) {
	panic("TODO, stubbed")
}

// If_ManSat6ComposeLut4
// C: word If_ManSat6ComposeLut4( int t, word f[4], int k )
func ifManSat6ComposeLut4(t int, f []uint64, k int) uint64 {
	panic("TODO, stubbed")
}

// If_ManSat6Truth
// C: word If_ManSat6Truth( word uBound, word uFree, int * pBSet, int nBSet, int * pSSet, int nSSet, int * pFSet, int nFSet )
func IfManSat6Truth(uBound uint64, uFree uint64, pBSet []int, nBSet int, pSSet []int, nSSet int, pFSet []int, nFSet int) uint64 {
	panic("TODO, stubbed")
}

// If_ManSatCheckXY (pSat is a void * that holds a sat_solver *)
// C: int If_ManSatCheckXY( void * pSat, int nLutSize, word * pTruth, int nVars, unsigned uSet, word * pTBound, word * pTFree, Vec_Int_t * vLits )
func IfManSatCheckXY(pSat *SatSolver, nLutSize int, pTruth []uint64, nVars int, uSet uint32, pTBound *uint64, pTFree *uint64, vLits *IntVector) int {
	panic("TODO, stubbed")
}

// If_ManSatCheckXYall_int (pSat is a void * that holds a sat_solver *)
// C: unsigned If_ManSatCheckXYall_int( void * pSat, int nLutSize, word * pTruth, int nVars, Vec_Int_t * vLits )
func IfManSatCheckXYall_int(pSat *SatSolver, nLutSize int, pTruth []uint64, nVars int, vLits *IntVector) uint32 {
	panic("TODO, stubbed")
}

// If_ManSatCheckXYall (pSat is a void * that holds a sat_solver *)
// C: unsigned If_ManSatCheckXYall( void * pSat, int nLutSize, word * pTruth, int nVars, Vec_Int_t * vLits )
func IfManSatCheckXYall(pSat *SatSolver, nLutSize int, pTruth []uint64, nVars int, vLits *IntVector) uint32 {
	panic("TODO, stubbed")
}

// If_ManSatTest2
// C: void If_ManSatTest2()
func IfManSatTest2() {
	panic("TODO, stubbed")
}

// If_ManSatTest3
// C: void If_ManSatTest3()
func IfManSatTest3() {
	panic("TODO, stubbed")
}
