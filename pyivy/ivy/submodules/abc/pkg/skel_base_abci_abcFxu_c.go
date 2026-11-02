package abc

// Hand-checked skeleton of src/base/abci/abcFxu.c.

// FxuData is Fxu_Data_t (struct FxuDataStruct of src/opt/fxu/fxu.h).
type FxuData struct {
	// user specified parameters
	fOnlyS       int // set to 1 to have only single-cube divs
	fOnlyD       int // set to 1 to have only double-cube divs
	fUse0        int // set to 1 to have 0-weight also extracted
	fUseCompl    int // set to 1 to have complement taken into account
	fVerbose     int // set to 1 to have verbose output
	fVeryVerbose int // set to 1 to have more verbose output
	nNodesExt    int // the number of divisors to extract
	nSingleMax   int // the max number of single-cube divisors to consider
	nPairsMax    int // the max number of double-cube divisors to consider
	WeightMin    int // the min weight of a divisor to extract
	LitCountMax  int // the max literal count of a divisor to consider
	fCanonDivs   int // use only canonical divisors (AND/XOR/MUX)
	// the input information
	vSops   *PointerVector[string]    // the SOPs for each node in the network
	vFanins *PointerVector[IntVector] // the fanins of each node in the network
	// output information
	vSopsNew   *PointerVector[string]    // the SOPs for each node in the network after extraction
	vFaninsNew *PointerVector[IntVector] // the fanins of each node in the network after extraction
	// the SOP manager
	pManSop *ExtraFlexArena // Mem_Flex_t *
	// statistics
	nNodesOld int // the old number of nodes
	nNodesNew int // the number of divisors actually extracted
}

// Abc_NtkSetDefaultFxParams (the C name is Abc_NtkSetDefaultFxParams)
// C: void Abc_NtkSetDefaultFxParams( Fxu_Data_t * p )
func (p *FxuData) AbcNtkSetDefaultFxParams() {
	panic("TODO, stubbed")
}

// Abc_NtkFastExtract
// C: int Abc_NtkFastExtract( Abc_Ntk_t * pNtk, Fxu_Data_t * p )
func (pNtk *ABCNetwork) AbcNtkFastExtract(p *FxuData) int {
	panic("TODO, stubbed")
}

// Abc_NtkFxuCheck (static)
// C: int Abc_NtkFxuCheck( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkFxuCheck() int {
	panic("TODO, stubbed")
}

// Abc_NtkFxuCollectInfo (static)
// C: void Abc_NtkFxuCollectInfo( Abc_Ntk_t * pNtk, Fxu_Data_t * p )
func (pNtk *ABCNetwork) abcNtkFxuCollectInfo(p *FxuData) {
	panic("TODO, stubbed")
}

// Abc_NtkFxuFreeInfo
// C: void Abc_NtkFxuFreeInfo( Fxu_Data_t * p )
func (p *FxuData) AbcNtkFxuFreeInfo() {
	panic("TODO, stubbed")
}

// Abc_NtkFxuReconstruct (static)
// C: void Abc_NtkFxuReconstruct( Abc_Ntk_t * pNtk, Fxu_Data_t * p )
func (pNtk *ABCNetwork) abcNtkFxuReconstruct(p *FxuData) {
	panic("TODO, stubbed")
}
