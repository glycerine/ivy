package abc

// Hand-checked skeleton of src/map/fpga/fpgaCreate.c.

// Fpga_HashKey2 (static)
// C: unsigned Fpga_HashKey2( Fpga_Node_t * p0, Fpga_Node_t * p1, int TableSize )
func fpgaHashKey2(p0 *FpgaNode, p1 *FpgaNode, TableSize int) uint32 {
	panic("TODO, stubbed")
}

// Fpga_ManReadInputNum
// C: int Fpga_ManReadInputNum( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManReadInputNum() int {
	panic("TODO, stubbed")
}

// Fpga_ManReadOutputNum
// C: int Fpga_ManReadOutputNum( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManReadOutputNum() int {
	panic("TODO, stubbed")
}

// Fpga_ManReadOutputs
// C: Fpga_Node_t ** Fpga_ManReadOutputs( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManReadOutputs() []*FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_ManReadInputArrivals
// C: float * Fpga_ManReadInputArrivals( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManReadInputArrivals() []float32 {
	panic("TODO, stubbed")
}

// Fpga_ManReadVerbose
// C: int Fpga_ManReadVerbose( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManReadVerbose() int {
	panic("TODO, stubbed")
}

// Fpga_ManReadVarMax
// C: int Fpga_ManReadVarMax( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManReadVarMax() int {
	panic("TODO, stubbed")
}

// Fpga_ManReadLutAreas
// C: float * Fpga_ManReadLutAreas( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManReadLutAreas() []float32 {
	panic("TODO, stubbed")
}

// Fpga_ManReadMapping
// C: Fpga_NodeVec_t* Fpga_ManReadMapping( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManReadMapping() *FpgaNodeVec {
	panic("TODO, stubbed")
}

// Fpga_ManSetOutputNames
// C: void Fpga_ManSetOutputNames( Fpga_Man_t * p, char ** ppNames )
func (p *FpgaMan) FpgaManSetOutputNames(ppNames []string) {
	panic("TODO, stubbed")
}

// Fpga_ManSetInputArrivals
// C: void Fpga_ManSetInputArrivals( Fpga_Man_t * p, float * pArrivals )
func (p *FpgaMan) FpgaManSetInputArrivals(pArrivals []float32) {
	panic("TODO, stubbed")
}

// Fpga_ManSetAreaRecovery
// C: void Fpga_ManSetAreaRecovery( Fpga_Man_t * p, int fAreaRecovery )
func (p *FpgaMan) FpgaManSetAreaRecovery(fAreaRecovery int) {
	panic("TODO, stubbed")
}

// Fpga_ManSetDelayLimit
// C: void Fpga_ManSetDelayLimit( Fpga_Man_t * p, float DelayLimit )
func (p *FpgaMan) FpgaManSetDelayLimit(DelayLimit float32) {
	panic("TODO, stubbed")
}

// Fpga_ManSetAreaLimit
// C: void Fpga_ManSetAreaLimit( Fpga_Man_t * p, float AreaLimit )
func (p *FpgaMan) FpgaManSetAreaLimit(AreaLimit float32) {
	panic("TODO, stubbed")
}

// Fpga_ManSetChoiceNodeNum
// C: void Fpga_ManSetChoiceNodeNum( Fpga_Man_t * p, int nChoiceNodes )
func (p *FpgaMan) FpgaManSetChoiceNodeNum(nChoiceNodes int) {
	panic("TODO, stubbed")
}

// Fpga_ManSetChoiceNum
// C: void Fpga_ManSetChoiceNum( Fpga_Man_t * p, int nChoices )
func (p *FpgaMan) FpgaManSetChoiceNum(nChoices int) {
	panic("TODO, stubbed")
}

// Fpga_ManSetVerbose
// C: void Fpga_ManSetVerbose( Fpga_Man_t * p, int fVerbose )
func (p *FpgaMan) FpgaManSetVerbose(fVerbose int) {
	panic("TODO, stubbed")
}

// Fpga_ManSetSwitching
// C: void Fpga_ManSetSwitching( Fpga_Man_t * p, int fSwitching )
func (p *FpgaMan) FpgaManSetSwitching(fSwitching int) {
	panic("TODO, stubbed")
}

// Fpga_ManSetLatchPaths
// C: void Fpga_ManSetLatchPaths( Fpga_Man_t * p, int fLatchPaths )
func (p *FpgaMan) FpgaManSetLatchPaths(fLatchPaths int) {
	panic("TODO, stubbed")
}

// Fpga_ManSetLatchNum
// C: void Fpga_ManSetLatchNum( Fpga_Man_t * p, int nLatches )
func (p *FpgaMan) FpgaManSetLatchNum(nLatches int) {
	panic("TODO, stubbed")
}

// Fpga_ManSetDelayTarget
// C: void Fpga_ManSetDelayTarget( Fpga_Man_t * p, float DelayTarget )
func (p *FpgaMan) FpgaManSetDelayTarget(DelayTarget float32) {
	panic("TODO, stubbed")
}

// Fpga_ManSetName
// C: void Fpga_ManSetName( Fpga_Man_t * p, char * pFileName )
func (p *FpgaMan) FpgaManSetName(pFileName string) {
	panic("TODO, stubbed")
}

// Fpga_LibReadLutMax
// C: int Fpga_LibReadLutMax( Fpga_LutLib_t * pLib )
func (pLib *FpgaLutLib) FpgaLibReadLutMax() int {
	panic("TODO, stubbed")
}

// Fpga_NodeReadData0 (char * used as an opaque pointer)
// C: char * Fpga_NodeReadData0( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeReadData0() any {
	panic("TODO, stubbed")
}

// Fpga_NodeReadData1
// C: Fpga_Node_t * Fpga_NodeReadData1( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeReadData1() *FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_NodeReadRefs
// C: int Fpga_NodeReadRefs( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeReadRefs() int {
	panic("TODO, stubbed")
}

// Fpga_NodeReadNum
// C: int Fpga_NodeReadNum( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeReadNum() int {
	panic("TODO, stubbed")
}

// Fpga_NodeReadLevel
// C: int Fpga_NodeReadLevel( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeReadLevel() int {
	panic("TODO, stubbed")
}

// Fpga_NodeReadCuts
// C: Fpga_Cut_t * Fpga_NodeReadCuts( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeReadCuts() *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_NodeReadCutBest
// C: Fpga_Cut_t * Fpga_NodeReadCutBest( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeReadCutBest() *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_NodeReadOne
// C: Fpga_Node_t * Fpga_NodeReadOne( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeReadOne() *FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_NodeReadTwo
// C: Fpga_Node_t * Fpga_NodeReadTwo( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeReadTwo() *FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_NodeSetData0 (char * used as an opaque pointer)
// C: void Fpga_NodeSetData0( Fpga_Node_t * p, char * pData )
func (p *FpgaNode) FpgaNodeSetData0(pData any) {
	panic("TODO, stubbed")
}

// Fpga_NodeSetData1
// C: void Fpga_NodeSetData1( Fpga_Node_t * p, Fpga_Node_t * pNode )
func (p *FpgaNode) FpgaNodeSetData1(pNode *FpgaNode) {
	panic("TODO, stubbed")
}

// Fpga_NodeSetNextE
// C: void Fpga_NodeSetNextE( Fpga_Node_t * p, Fpga_Node_t * pNextE )
func (p *FpgaNode) FpgaNodeSetNextE(pNextE *FpgaNode) {
	panic("TODO, stubbed")
}

// Fpga_NodeSetRepr
// C: void Fpga_NodeSetRepr( Fpga_Node_t * p, Fpga_Node_t * pRepr )
func (p *FpgaNode) FpgaNodeSetRepr(pRepr *FpgaNode) {
	panic("TODO, stubbed")
}

// Fpga_NodeSetSwitching
// C: void Fpga_NodeSetSwitching( Fpga_Node_t * p, float Switching )
func (p *FpgaNode) FpgaNodeSetSwitching(Switching float32) {
	panic("TODO, stubbed")
}

// Fpga_NodeIsConst
// C: int Fpga_NodeIsConst( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeIsConst() int {
	panic("TODO, stubbed")
}

// Fpga_NodeIsVar
// C: int Fpga_NodeIsVar( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeIsVar() int {
	panic("TODO, stubbed")
}

// Fpga_NodeIsAnd
// C: int Fpga_NodeIsAnd( Fpga_Node_t * p )
func (p *FpgaNode) FpgaNodeIsAnd() int {
	panic("TODO, stubbed")
}

// Fpga_NodeComparePhase
// C: int Fpga_NodeComparePhase( Fpga_Node_t * p1, Fpga_Node_t * p2 )
func (p1 *FpgaNode) FpgaNodeComparePhase(p2 *FpgaNode) int {
	panic("TODO, stubbed")
}

// Fpga_CutReadLeavesNum
// C: int Fpga_CutReadLeavesNum( Fpga_Cut_t * p )
func (p *FpgaCut) FpgaCutReadLeavesNum() int {
	panic("TODO, stubbed")
}

// Fpga_CutReadLeaves
// C: Fpga_Node_t ** Fpga_CutReadLeaves( Fpga_Cut_t * p )
func (p *FpgaCut) FpgaCutReadLeaves() []*FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_ManCreate
// C: Fpga_Man_t * Fpga_ManCreate( int nInputs, int nOutputs, int fVerbose )
func FpgaManCreate(nInputs int, nOutputs int, fVerbose int) *FpgaMan {
	panic("TODO, stubbed")
}

// Fpga_ManFree
// C: void Fpga_ManFree( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManFree() {
	panic("TODO, stubbed")
}

// Fpga_ManPrintTimeStats
// C: void Fpga_ManPrintTimeStats( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManPrintTimeStats() {
	panic("TODO, stubbed")
}

// Fpga_NodeCreate
// C: Fpga_Node_t * Fpga_NodeCreate( Fpga_Man_t * p, Fpga_Node_t * p1, Fpga_Node_t * p2 )
func (p *FpgaMan) FpgaNodeCreate(p1 *FpgaNode, p2 *FpgaNode) *FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_TableCreate (static)
// C: void Fpga_TableCreate( Fpga_Man_t * pMan )
func (pMan *FpgaMan) fpgaTableCreate() {
	panic("TODO, stubbed")
}

// Fpga_TableLookup (static)
// C: Fpga_Node_t * Fpga_TableLookup( Fpga_Man_t * pMan, Fpga_Node_t * p1, Fpga_Node_t * p2 )
func (pMan *FpgaMan) fpgaTableLookup(p1 *FpgaNode, p2 *FpgaNode) *FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_TableResize (static)
// C: void Fpga_TableResize( Fpga_Man_t * pMan )
func (pMan *FpgaMan) fpgaTableResize() {
	panic("TODO, stubbed")
}

// Fpga_NodeAnd
// C: Fpga_Node_t * Fpga_NodeAnd( Fpga_Man_t * p, Fpga_Node_t * p1, Fpga_Node_t * p2 )
func (p *FpgaMan) FpgaNodeAnd(p1 *FpgaNode, p2 *FpgaNode) *FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_NodeOr
// C: Fpga_Node_t * Fpga_NodeOr( Fpga_Man_t * p, Fpga_Node_t * p1, Fpga_Node_t * p2 )
func (p *FpgaMan) FpgaNodeOr(p1 *FpgaNode, p2 *FpgaNode) *FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_NodeExor
// C: Fpga_Node_t * Fpga_NodeExor( Fpga_Man_t * p, Fpga_Node_t * p1, Fpga_Node_t * p2 )
func (p *FpgaMan) FpgaNodeExor(p1 *FpgaNode, p2 *FpgaNode) *FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_NodeMux
// C: Fpga_Node_t * Fpga_NodeMux( Fpga_Man_t * p, Fpga_Node_t * pC, Fpga_Node_t * pT, Fpga_Node_t * pE )
func (p *FpgaMan) FpgaNodeMux(pC *FpgaNode, pT *FpgaNode, pE *FpgaNode) *FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_NodeSetChoice
// C: void Fpga_NodeSetChoice( Fpga_Man_t * pMan, Fpga_Node_t * pNodeOld, Fpga_Node_t * pNodeNew )
func (pMan *FpgaMan) FpgaNodeSetChoice(pNodeOld *FpgaNode, pNodeNew *FpgaNode) {
	panic("TODO, stubbed")
}

// Fpga_ManStats
// C: void Fpga_ManStats( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManStats() {
	panic("TODO, stubbed")
}
