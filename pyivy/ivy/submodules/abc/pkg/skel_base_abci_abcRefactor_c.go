package abc

// Hand-checked skeleton of src/base/abci/abcRefactor.c.

// AbcManRef is Abc_ManRef_t.
type AbcManRef struct {
	nNodeSizeMax     int                              // the limit on the size of the supernode
	nConeSizeMax     int                              // the limit on the size of the containing cone
	fVerbose         int                              // the verbosity flag
	vVars            *PointerVector[uint64]           // truth tables
	vFuncs           *PointerVector[uint64]           // functions
	vMemory          *IntVector                       // memory
	vCube            *ByteVector                      // temporary
	vForm            *IntVector                       // temporary
	vVisited         *PointerVector[ABCNetworkObject] // temporary
	vLeaves          *PointerVector[ABCNetworkObject] // temporary
	nLastGain        int
	nNodesConsidered int
	nNodesRefactored int
	nNodesGained     int
	nNodesBeg        int
	nNodesEnd        int
	timeCut          int64 // abctime
	timeTru          int64 // abctime
	timeDcs          int64 // abctime
	timeSop          int64 // abctime
	timeFact         int64 // abctime
	timeEval         int64 // abctime
	timeRes          int64 // abctime
	timeNtk          int64 // abctime
	timeTotal        int64 // abctime
}

// Abc_NodeConeTruth
// C: word * Abc_NodeConeTruth( Vec_Ptr_t * vVars, Vec_Ptr_t * vFuncs, int nWordsMax, Abc_Obj_t * pRoot, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vVisited )
func AbcNodeConeTruth(vVars *PointerVector[uint64], vFuncs *PointerVector[uint64], nWordsMax int, pRoot *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], vVisited *PointerVector[ABCNetworkObject]) []uint64 {
	panic("TODO, stubbed")
}

// Abc_NodeConeIsConst0
// C: int Abc_NodeConeIsConst0( word * pTruth, int nVars )
func AbcNodeConeIsConst0(pTruth []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// Abc_NodeConeIsConst1
// C: int Abc_NodeConeIsConst1( word * pTruth, int nVars )
func AbcNodeConeIsConst1(pTruth []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// Abc_NodeRefactor
// C: Dec_Graph_t * Abc_NodeRefactor( Abc_ManRef_t * p, Abc_Obj_t * pNode, Vec_Ptr_t * vFanins, int fUpdateLevel, int fUseZeros, int fUseDcs, int fVerbose )
func (p *AbcManRef) AbcNodeRefactor(pNode *ABCNetworkObject, vFanins *PointerVector[ABCNetworkObject], fUpdateLevel int, fUseZeros int, fUseDcs int, fVerbose int) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_NtkManRefStart
// C: Abc_ManRef_t * Abc_NtkManRefStart( int nNodeSizeMax, int nConeSizeMax, int fUseDcs, int fVerbose )
func AbcNtkManRefStart(nNodeSizeMax int, nConeSizeMax int, fUseDcs int, fVerbose int) *AbcManRef {
	panic("TODO, stubbed")
}

// Abc_NtkManRefStop
// C: void Abc_NtkManRefStop( Abc_ManRef_t * p )
func (p *AbcManRef) AbcNtkManRefStop() {
	panic("TODO, stubbed")
}

// Abc_NtkManRefPrintStats
// C: void Abc_NtkManRefPrintStats( Abc_ManRef_t * p )
func (p *AbcManRef) AbcNtkManRefPrintStats() {
	panic("TODO, stubbed")
}

// Abc_NtkRefactor
// C: int Abc_NtkRefactor( Abc_Ntk_t * pNtk, int nNodeSizeMax, int nConeSizeMax, int fUpdateLevel, int fUseZeros, int fUseDcs, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkRefactor(nNodeSizeMax int, nConeSizeMax int, fUpdateLevel int, fUseZeros int, fUseDcs int, fVerbose int) int {
	panic("TODO, stubbed")
}
