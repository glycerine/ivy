package abc

// Hand-checked skeleton of src/base/abci/abcResub.c.

// AbcManRes is Abc_ManRes_t.
type AbcManRes struct {
	nLeavesMax      int                              // the max number of leaves in the cone
	nDivsMax        int                              // the max number of divisors in the cone
	pRoot           *ABCNetworkObject                // the root of the cone
	nLeaves         int                              // the number of leaves
	nDivs           int                              // the number of all divisor (including leaves)
	nMffc           int                              // the size of MFFC
	nLastGain       int                              // the gain the number of nodes
	vDivs           *PointerVector[ABCNetworkObject] // the divisors
	nBits           int                              // the number of simulation bits
	nWords          int                              // the number of unsigneds for siminfo
	vSims           *PointerVector[uint32]           // simulation info
	pInfo           []uint32                         // unsigned * : pointer to simulation info
	pCareSet        []uint32                         // unsigned *
	vDivs1UP        *PointerVector[ABCNetworkObject] // the single-node unate divisors
	vDivs1UN        *PointerVector[ABCNetworkObject] // the single-node unate divisors
	vDivs1B         *PointerVector[ABCNetworkObject] // the single-node binate divisors
	vDivs2UP0       *PointerVector[ABCNetworkObject] // the double-node unate divisors
	vDivs2UP1       *PointerVector[ABCNetworkObject] // the double-node unate divisors
	vDivs2UN0       *PointerVector[ABCNetworkObject] // the double-node unate divisors
	vDivs2UN1       *PointerVector[ABCNetworkObject] // the double-node unate divisors
	vTemp           *PointerVector[ABCNetworkObject] // temporary array of nodes
	timeCut         int64                            // abctime
	timeTruth       int64                            // abctime
	timeRes         int64                            // abctime
	timeDiv         int64                            // abctime
	timeMffc        int64                            // abctime
	timeSim         int64                            // abctime
	timeRes1        int64                            // abctime
	timeResD        int64                            // abctime
	timeRes2        int64                            // abctime
	timeRes3        int64                            // abctime
	timeNtk         int64                            // abctime
	timeTotal       int64                            // abctime
	nUsedNodeC      int
	nUsedNode0      int
	nUsedNode1Or    int
	nUsedNode1And   int
	nUsedNode2Or    int
	nUsedNode2And   int
	nUsedNode2OrAnd int
	nUsedNode2AndOr int
	nUsedNode3OrAnd int
	nUsedNode3AndOr int
	nUsedNodeTotal  int
	nTotalDivs      int
	nTotalLeaves    int
	nTotalGain      int
	nNodesBeg       int
	nNodesEnd       int
}

// Abc_NtkResubstitute
// C: int Abc_NtkResubstitute( Abc_Ntk_t * pNtk, int nCutMax, int nStepsMax, int nLevelsOdc, int fUpdateLevel, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) AbcNtkResubstitute(nCutMax int, nStepsMax int, nLevelsOdc int, fUpdateLevel int, fVerbose int, fVeryVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_ManResubStart (static)
// C: Abc_ManRes_t * Abc_ManResubStart( int nLeavesMax, int nDivsMax )
func abcManResubStart(nLeavesMax int, nDivsMax int) *AbcManRes {
	panic("TODO, stubbed")
}

// Abc_ManResubStop (static)
// C: void Abc_ManResubStop( Abc_ManRes_t * p )
func (p *AbcManRes) abcManResubStop() {
	panic("TODO, stubbed")
}

// Abc_ManResubPrint (static)
// C: void Abc_ManResubPrint( Abc_ManRes_t * p )
func (p *AbcManRes) abcManResubPrint() {
	panic("TODO, stubbed")
}

// Abc_ManResubCollectDivs_rec
// C: void Abc_ManResubCollectDivs_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vInternal )
func AbcManResubCollectDivs_rec(pNode *ABCNetworkObject, vInternal *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_ManResubCollectDivs (static)
// C: int Abc_ManResubCollectDivs( Abc_ManRes_t * p, Abc_Obj_t * pRoot, Vec_Ptr_t * vLeaves, int Required )
func (p *AbcManRes) abcManResubCollectDivs(pRoot *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], Required int) int {
	panic("TODO, stubbed")
}

// Abc_ManResubPrintDivs (static)
// C: void Abc_ManResubPrintDivs( Abc_ManRes_t * p, Abc_Obj_t * pRoot, Vec_Ptr_t * vLeaves )
func (p *AbcManRes) abcManResubPrintDivs(pRoot *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_ManResubSimulate (static)
// C: void Abc_ManResubSimulate( Vec_Ptr_t * vDivs, int nLeaves, Vec_Ptr_t * vSims, int nLeavesMax, int nWords )
func abcManResubSimulate(vDivs *PointerVector[ABCNetworkObject], nLeaves int, vSims *PointerVector[uint32], nLeavesMax int, nWords int) {
	panic("TODO, stubbed")
}

// Abc_ManResubQuit0
// C: Dec_Graph_t * Abc_ManResubQuit0( Abc_Obj_t * pRoot, Abc_Obj_t * pObj )
func AbcManResubQuit0(pRoot *ABCNetworkObject, pObj *ABCNetworkObject) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_ManResubQuit1
// C: Dec_Graph_t * Abc_ManResubQuit1( Abc_Obj_t * pRoot, Abc_Obj_t * pObj0, Abc_Obj_t * pObj1, int fOrGate )
func AbcManResubQuit1(pRoot *ABCNetworkObject, pObj0 *ABCNetworkObject, pObj1 *ABCNetworkObject, fOrGate int) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_ManResubQuit21
// C: Dec_Graph_t * Abc_ManResubQuit21( Abc_Obj_t * pRoot, Abc_Obj_t * pObj0, Abc_Obj_t * pObj1, Abc_Obj_t * pObj2, int fOrGate )
func AbcManResubQuit21(pRoot *ABCNetworkObject, pObj0 *ABCNetworkObject, pObj1 *ABCNetworkObject, pObj2 *ABCNetworkObject, fOrGate int) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_ManResubQuit2
// C: Dec_Graph_t * Abc_ManResubQuit2( Abc_Obj_t * pRoot, Abc_Obj_t * pObj0, Abc_Obj_t * pObj1, Abc_Obj_t * pObj2, int fOrGate )
func AbcManResubQuit2(pRoot *ABCNetworkObject, pObj0 *ABCNetworkObject, pObj1 *ABCNetworkObject, pObj2 *ABCNetworkObject, fOrGate int) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_ManResubQuit3
// C: Dec_Graph_t * Abc_ManResubQuit3( Abc_Obj_t * pRoot, Abc_Obj_t * pObj0, Abc_Obj_t * pObj1, Abc_Obj_t * pObj2, Abc_Obj_t * pObj3, int fOrGate )
func AbcManResubQuit3(pRoot *ABCNetworkObject, pObj0 *ABCNetworkObject, pObj1 *ABCNetworkObject, pObj2 *ABCNetworkObject, pObj3 *ABCNetworkObject, fOrGate int) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_ManResubDivsS (static)
// C: void Abc_ManResubDivsS( Abc_ManRes_t * p, int Required )
func (p *AbcManRes) abcManResubDivsS(Required int) {
	panic("TODO, stubbed")
}

// Abc_ManResubDivsD (static)
// C: void Abc_ManResubDivsD( Abc_ManRes_t * p, int Required )
func (p *AbcManRes) abcManResubDivsD(Required int) {
	panic("TODO, stubbed")
}

// Abc_ManResubQuit (static)
// C: Dec_Graph_t * Abc_ManResubQuit( Abc_ManRes_t * p )
func (p *AbcManRes) abcManResubQuit() *DecGraph {
	panic("TODO, stubbed")
}

// Abc_ManResubDivs0 (static)
// C: Dec_Graph_t * Abc_ManResubDivs0( Abc_ManRes_t * p )
func (p *AbcManRes) abcManResubDivs0() *DecGraph {
	panic("TODO, stubbed")
}

// Abc_ManResubDivs1 (static)
// C: Dec_Graph_t * Abc_ManResubDivs1( Abc_ManRes_t * p, int Required )
func (p *AbcManRes) abcManResubDivs1(Required int) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_ManResubDivs12 (static)
// C: Dec_Graph_t * Abc_ManResubDivs12( Abc_ManRes_t * p, int Required )
func (p *AbcManRes) abcManResubDivs12(Required int) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_ManResubDivs2 (static)
// C: Dec_Graph_t * Abc_ManResubDivs2( Abc_ManRes_t * p, int Required )
func (p *AbcManRes) abcManResubDivs2(Required int) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_ManResubDivs3 (static)
// C: Dec_Graph_t * Abc_ManResubDivs3( Abc_ManRes_t * p, int Required )
func (p *AbcManRes) abcManResubDivs3(Required int) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_ManResubCleanup (static)
// C: void Abc_ManResubCleanup( Abc_ManRes_t * p )
func (p *AbcManRes) abcManResubCleanup() {
	panic("TODO, stubbed")
}

// Abc_ManResubEval (static)
// C: Dec_Graph_t * Abc_ManResubEval( Abc_ManRes_t * p, Abc_Obj_t * pRoot, Vec_Ptr_t * vLeaves, int nSteps, int fUpdateLevel, int fVerbose )
func (p *AbcManRes) abcManResubEval(pRoot *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], nSteps int, fUpdateLevel int, fVerbose int) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_CutVolumeCheck_rec
// C: int Abc_CutVolumeCheck_rec( Abc_Obj_t * pObj )
func AbcCutVolumeCheck_rec(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_CutVolumeCheck (static)
// C: int Abc_CutVolumeCheck( Abc_Obj_t * pNode, Vec_Ptr_t * vLeaves )
func abcCutVolumeCheck(pNode *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject]) int {
	panic("TODO, stubbed")
}

// Abc_CutFactor_rec
// C: void Abc_CutFactor_rec( Abc_Obj_t * pObj, Vec_Ptr_t * vLeaves )
func AbcCutFactor_rec(pObj *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_CutFactor
// C: Vec_Ptr_t * Abc_CutFactor( Abc_Obj_t * pNode )
func AbcCutFactor(pNode *ABCNetworkObject) *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_CutFactorLarge (static)
// C: Vec_Ptr_t * Abc_CutFactorLarge( Abc_Obj_t * pNode, int nLeavesMax )
func abcCutFactorLarge(pNode *ABCNetworkObject, nLeavesMax int) *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}
