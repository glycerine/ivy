package abc

// Hand-checked skeleton of src/bool/kit/kitDec.c.

// KitManDec is Kit_ManDec_t.
type KitManDec struct {
	nVarsMax    int                    // the max number of variables
	nWordsMax   int                    // the max number of words
	vTruthVars  *PointerVector[uint32] // elementary truth tables
	vTruthNodes *PointerVector[uint32] // internal truth tables
	// current problem
	nVarsIn int        // the current number of variables
	vLutsIn *IntVector // LUT truth tables
	vSuppIn *IntVector // LUT supports
	ATimeIn [64]int8   // char ATimeIn[64] : variable arrival times
	// extracted information
	pTruthIn  []uint32 // unsigned * : computed truth table
	pTruthOut []uint32 // unsigned * : computed truth table
	nVarsOut  int      // the current number of variables
	nWordsOut int      // the current number of words
	Order     [32]int8 // char Order[32] : new vars into old vars after supp minimization
	// computed information
	vLutsOut *IntVector // problem decomposition
	vSuppOut *IntVector // problem decomposition
	ATimeOut [64]int8   // char ATimeOut[64] : variable arrival times
}

// Kit_DecOuputArrival (static)
// C: int Kit_DecOuputArrival( int nVars, Vec_Int_t * vLuts, char ATimes[] )
func kitDecOuputArrival(nVars int, vLuts *IntVector, ATimes []int8) int {
	panic("TODO, stubbed")
}

// Kit_ManDecStart
// C: Kit_ManDec_t * Kit_ManDecStart( int nVarsMax )
func KitManDecStart(nVarsMax int) *KitManDec {
	panic("TODO, stubbed")
}

// Kit_ManDecStop
// C: void Kit_ManDecStop( Kit_ManDec_t * p )
func (p *KitManDec) KitManDecStop() {
	panic("TODO, stubbed")
}

// Kit_DecComputeOuputArrival
// C: int Kit_DecComputeOuputArrival( int nVars, Vec_Int_t * vSupps, int LutSize, char ATimesIn[], char ATimesOut[] )
func KitDecComputeOuputArrival(nVars int, vSupps *IntVector, LutSize int, ATimesIn []int8, ATimesOut []int8) int {
	panic("TODO, stubbed")
}

// Kit_DecComputeTruthOne
// C: void Kit_DecComputeTruthOne( int LutSize, unsigned * pTruthLut, int nVars, unsigned * pTruths[], unsigned * pTemp, unsigned * pRes )
func KitDecComputeTruthOne(LutSize int, pTruthLut []uint32, nVars int, pTruths [][]uint32, pTemp []uint32, pRes []uint32) {
	panic("TODO, stubbed")
}

// Kit_DecComputeTruth
// C: void Kit_DecComputeTruth( Kit_ManDec_t * p, int nVars, Vec_Int_t * vSupps, int LutSize, Vec_Int_t * vLuts, unsigned * pRes )
func (p *KitManDec) KitDecComputeTruth(nVars int, vSupps *IntVector, LutSize int, vLuts *IntVector, pRes []uint32) {
	panic("TODO, stubbed")
}

// Kit_DecComputePattern
// C: int Kit_DecComputePattern( int nVars, unsigned * pTruth, int LutSize, int Pattern[] )
func KitDecComputePattern(nVars int, pTruth []uint32, LutSize int, Pattern []int) int {
	panic("TODO, stubbed")
}

// Kit_DecComputeShared_rec
// C: int Kit_DecComputeShared_rec( int Pattern[], int Vars[], int nVars, int Shared[], int iVarTry )
func KitDecComputeShared_rec(Pattern []int, Vars []int, nVars int, Shared []int, iVarTry int) int {
	panic("TODO, stubbed")
}

// Kit_DecComputeShared
// C: int Kit_DecComputeShared( int Pattern[], int LutSize, int Shared[] )
func KitDecComputeShared(Pattern []int, LutSize int, Shared []int) int {
	panic("TODO, stubbed")
}
