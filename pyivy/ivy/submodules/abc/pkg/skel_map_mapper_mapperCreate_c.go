package abc

// Hand-checked skeleton of src/map/mapper/mapperCreate.c.

// Map_HashKey2 (static)
// C: unsigned Map_HashKey2( Map_Node_t * p0, Map_Node_t * p1, int TableSize )
func mapHashKey2(p0 *MapNode, p1 *MapNode, TableSize int) uint32 {
	panic("TODO, stubbed")
}

// Map_ManReadInputNum
// C: int Map_ManReadInputNum( Map_Man_t * p )
func (p *MapMan) MapManReadInputNum() int {
	panic("TODO, stubbed")
}

// Map_ManReadOutputNum
// C: int Map_ManReadOutputNum( Map_Man_t * p )
func (p *MapMan) MapManReadOutputNum() int {
	panic("TODO, stubbed")
}

// Map_ManReadBufNum
// C: int Map_ManReadBufNum( Map_Man_t * p )
func (p *MapMan) MapManReadBufNum() int {
	panic("TODO, stubbed")
}

// Map_ManReadOutputs
// C: Map_Node_t ** Map_ManReadOutputs( Map_Man_t * p )
func (p *MapMan) MapManReadOutputs() []*MapNode {
	panic("TODO, stubbed")
}

// Map_ManReadBufs
// C: Map_Node_t ** Map_ManReadBufs( Map_Man_t * p )
func (p *MapMan) MapManReadBufs() []*MapNode {
	panic("TODO, stubbed")
}

// Map_ManReadBufDriver
// C: Map_Node_t * Map_ManReadBufDriver( Map_Man_t * p, int i )
func (p *MapMan) MapManReadBufDriver(i int) *MapNode {
	panic("TODO, stubbed")
}

// Map_ManReadInputArrivals
// C: Map_Time_t * Map_ManReadInputArrivals( Map_Man_t * p )
func (p *MapMan) MapManReadInputArrivals() []MapTime {
	panic("TODO, stubbed")
}

// Map_ManReadOutputRequireds
// C: Map_Time_t * Map_ManReadOutputRequireds( Map_Man_t * p )
func (p *MapMan) MapManReadOutputRequireds() []MapTime {
	panic("TODO, stubbed")
}

// Map_ManReadVerbose
// C: int Map_ManReadVerbose( Map_Man_t * p )
func (p *MapMan) MapManReadVerbose() int {
	panic("TODO, stubbed")
}

// Map_ManReadAreaFinal
// C: float Map_ManReadAreaFinal( Map_Man_t * p )
func (p *MapMan) MapManReadAreaFinal() float32 {
	panic("TODO, stubbed")
}

// Map_ManReadRequiredGlo
// C: float Map_ManReadRequiredGlo( Map_Man_t * p )
func (p *MapMan) MapManReadRequiredGlo() float32 {
	panic("TODO, stubbed")
}

// Map_ManSetOutputNames
// C: void Map_ManSetOutputNames( Map_Man_t * p, char ** ppNames )
func (p *MapMan) MapManSetOutputNames(ppNames []string) {
	panic("TODO, stubbed")
}

// Map_ManSetAreaRecovery
// C: void Map_ManSetAreaRecovery( Map_Man_t * p, int fAreaRecovery )
func (p *MapMan) MapManSetAreaRecovery(fAreaRecovery int) {
	panic("TODO, stubbed")
}

// Map_ManSetDelayTarget
// C: void Map_ManSetDelayTarget( Map_Man_t * p, float DelayTarget )
func (p *MapMan) MapManSetDelayTarget(DelayTarget float32) {
	panic("TODO, stubbed")
}

// Map_ManSetInputArrivals
// C: void Map_ManSetInputArrivals( Map_Man_t * p, Map_Time_t * pArrivals )
func (p *MapMan) MapManSetInputArrivals(pArrivals []MapTime) {
	panic("TODO, stubbed")
}

// Map_ManSetOutputRequireds
// C: void Map_ManSetOutputRequireds( Map_Man_t * p, Map_Time_t * pRequireds )
func (p *MapMan) MapManSetOutputRequireds(pRequireds []MapTime) {
	panic("TODO, stubbed")
}

// Map_ManSetObeyFanoutLimits
// C: void Map_ManSetObeyFanoutLimits( Map_Man_t * p, int fObeyFanoutLimits )
func (p *MapMan) MapManSetObeyFanoutLimits(fObeyFanoutLimits int) {
	panic("TODO, stubbed")
}

// Map_ManSetNumIterations
// C: void Map_ManSetNumIterations( Map_Man_t * p, int nIterations )
func (p *MapMan) MapManSetNumIterations(nIterations int) {
	panic("TODO, stubbed")
}

// Map_ManReadFanoutViolations
// C: int Map_ManReadFanoutViolations( Map_Man_t * p )
func (p *MapMan) MapManReadFanoutViolations() int {
	panic("TODO, stubbed")
}

// Map_ManSetFanoutViolations
// C: void Map_ManSetFanoutViolations( Map_Man_t * p, int nVio )
func (p *MapMan) MapManSetFanoutViolations(nVio int) {
	panic("TODO, stubbed")
}

// Map_ManSetChoiceNodeNum
// C: void Map_ManSetChoiceNodeNum( Map_Man_t * p, int nChoiceNodes )
func (p *MapMan) MapManSetChoiceNodeNum(nChoiceNodes int) {
	panic("TODO, stubbed")
}

// Map_ManSetChoiceNum
// C: void Map_ManSetChoiceNum( Map_Man_t * p, int nChoices )
func (p *MapMan) MapManSetChoiceNum(nChoices int) {
	panic("TODO, stubbed")
}

// Map_ManSetVerbose
// C: void Map_ManSetVerbose( Map_Man_t * p, int fVerbose )
func (p *MapMan) MapManSetVerbose(fVerbose int) {
	panic("TODO, stubbed")
}

// Map_ManSetSwitching
// C: void Map_ManSetSwitching( Map_Man_t * p, int fSwitching )
func (p *MapMan) MapManSetSwitching(fSwitching int) {
	panic("TODO, stubbed")
}

// Map_ManSetSkipFanout
// C: void Map_ManSetSkipFanout( Map_Man_t * p, int fSkipFanout )
func (p *MapMan) MapManSetSkipFanout(fSkipFanout int) {
	panic("TODO, stubbed")
}

// Map_ManSetUseProfile
// C: void Map_ManSetUseProfile( Map_Man_t * p )
func (p *MapMan) MapManSetUseProfile() {
	panic("TODO, stubbed")
}

// Map_NodeReadMan
// C: Map_Man_t * Map_NodeReadMan( Map_Node_t * p )
func (p *MapNode) MapNodeReadMan() *MapMan {
	panic("TODO, stubbed")
}

// Map_NodeReadData (char * used as an opaque pointer)
// C: char * Map_NodeReadData( Map_Node_t * p, int fPhase )
func (p *MapNode) MapNodeReadData(fPhase int) any {
	panic("TODO, stubbed")
}

// Map_NodeReadNum
// C: int Map_NodeReadNum( Map_Node_t * p )
func (p *MapNode) MapNodeReadNum() int {
	panic("TODO, stubbed")
}

// Map_NodeReadLevel
// C: int Map_NodeReadLevel( Map_Node_t * p )
func (p *MapNode) MapNodeReadLevel() int {
	panic("TODO, stubbed")
}

// Map_NodeReadCuts
// C: Map_Cut_t * Map_NodeReadCuts( Map_Node_t * p )
func (p *MapNode) MapNodeReadCuts() *MapCut {
	panic("TODO, stubbed")
}

// Map_NodeReadCutBest
// C: Map_Cut_t * Map_NodeReadCutBest( Map_Node_t * p, int fPhase )
func (p *MapNode) MapNodeReadCutBest(fPhase int) *MapCut {
	panic("TODO, stubbed")
}

// Map_NodeReadOne
// C: Map_Node_t * Map_NodeReadOne( Map_Node_t * p )
func (p *MapNode) MapNodeReadOne() *MapNode {
	panic("TODO, stubbed")
}

// Map_NodeReadTwo
// C: Map_Node_t * Map_NodeReadTwo( Map_Node_t * p )
func (p *MapNode) MapNodeReadTwo() *MapNode {
	panic("TODO, stubbed")
}

// Map_NodeSetData (char * used as an opaque pointer)
// C: void Map_NodeSetData( Map_Node_t * p, int fPhase, char * pData )
func (p *MapNode) MapNodeSetData(fPhase int, pData any) {
	panic("TODO, stubbed")
}

// Map_NodeSetNextE
// C: void Map_NodeSetNextE( Map_Node_t * p, Map_Node_t * pNextE )
func (p *MapNode) MapNodeSetNextE(pNextE *MapNode) {
	panic("TODO, stubbed")
}

// Map_NodeSetRepr
// C: void Map_NodeSetRepr( Map_Node_t * p, Map_Node_t * pRepr )
func (p *MapNode) MapNodeSetRepr(pRepr *MapNode) {
	panic("TODO, stubbed")
}

// Map_NodeSetSwitching
// C: void Map_NodeSetSwitching( Map_Node_t * p, float Switching )
func (p *MapNode) MapNodeSetSwitching(Switching float32) {
	panic("TODO, stubbed")
}

// Map_NodeIsConst
// C: int Map_NodeIsConst( Map_Node_t * p )
func (p *MapNode) MapNodeIsConst() int {
	panic("TODO, stubbed")
}

// Map_NodeIsVar
// C: int Map_NodeIsVar( Map_Node_t * p )
func (p *MapNode) MapNodeIsVar() int {
	panic("TODO, stubbed")
}

// Map_NodeIsBuf
// C: int Map_NodeIsBuf( Map_Node_t * p )
func (p *MapNode) MapNodeIsBuf() int {
	panic("TODO, stubbed")
}

// Map_NodeIsAnd
// C: int Map_NodeIsAnd( Map_Node_t * p )
func (p *MapNode) MapNodeIsAnd() int {
	panic("TODO, stubbed")
}

// Map_NodeComparePhase
// C: int Map_NodeComparePhase( Map_Node_t * p1, Map_Node_t * p2 )
func (p1 *MapNode) MapNodeComparePhase(p2 *MapNode) int {
	panic("TODO, stubbed")
}

// Map_CutReadSuperBest
// C: Map_Super_t * Map_CutReadSuperBest( Map_Cut_t * p, int fPhase )
func (p *MapCut) MapCutReadSuperBest(fPhase int) *MapSuper {
	panic("TODO, stubbed")
}

// Map_CutReadSuper0
// C: Map_Super_t * Map_CutReadSuper0( Map_Cut_t * p )
func (p *MapCut) MapCutReadSuper0() *MapSuper {
	panic("TODO, stubbed")
}

// Map_CutReadSuper1
// C: Map_Super_t * Map_CutReadSuper1( Map_Cut_t * p )
func (p *MapCut) MapCutReadSuper1() *MapSuper {
	panic("TODO, stubbed")
}

// Map_CutReadLeavesNum
// C: int Map_CutReadLeavesNum( Map_Cut_t * p )
func (p *MapCut) MapCutReadLeavesNum() int {
	panic("TODO, stubbed")
}

// Map_CutReadLeaves
// C: Map_Node_t ** Map_CutReadLeaves( Map_Cut_t * p )
func (p *MapCut) MapCutReadLeaves() []*MapNode {
	panic("TODO, stubbed")
}

// Map_CutReadPhaseBest
// C: unsigned Map_CutReadPhaseBest( Map_Cut_t * p, int fPhase )
func (p *MapCut) MapCutReadPhaseBest(fPhase int) uint32 {
	panic("TODO, stubbed")
}

// Map_CutReadPhase0
// C: unsigned Map_CutReadPhase0( Map_Cut_t * p )
func (p *MapCut) MapCutReadPhase0() uint32 {
	panic("TODO, stubbed")
}

// Map_CutReadPhase1
// C: unsigned Map_CutReadPhase1( Map_Cut_t * p )
func (p *MapCut) MapCutReadPhase1() uint32 {
	panic("TODO, stubbed")
}

// Map_CutReadNext
// C: Map_Cut_t * Map_CutReadNext( Map_Cut_t * p )
func (p *MapCut) MapCutReadNext() *MapCut {
	panic("TODO, stubbed")
}

// Map_SuperReadFormula
// C: char * Map_SuperReadFormula( Map_Super_t * p )
func (p *MapSuper) MapSuperReadFormula() string {
	panic("TODO, stubbed")
}

// Map_SuperReadRoot
// C: Mio_Gate_t * Map_SuperReadRoot( Map_Super_t * p )
func (p *MapSuper) MapSuperReadRoot() *MioGate {
	panic("TODO, stubbed")
}

// Map_SuperReadNum
// C: int Map_SuperReadNum( Map_Super_t * p )
func (p *MapSuper) MapSuperReadNum() int {
	panic("TODO, stubbed")
}

// Map_SuperReadFanins
// C: Map_Super_t ** Map_SuperReadFanins( Map_Super_t * p )
func (p *MapSuper) MapSuperReadFanins() []*MapSuper {
	panic("TODO, stubbed")
}

// Map_SuperReadFaninNum
// C: int Map_SuperReadFaninNum( Map_Super_t * p )
func (p *MapSuper) MapSuperReadFaninNum() int {
	panic("TODO, stubbed")
}

// Map_SuperReadNext
// C: Map_Super_t * Map_SuperReadNext( Map_Super_t * p )
func (p *MapSuper) MapSuperReadNext() *MapSuper {
	panic("TODO, stubbed")
}

// Map_SuperReadNumPhases
// C: int Map_SuperReadNumPhases( Map_Super_t * p )
func (p *MapSuper) MapSuperReadNumPhases() int {
	panic("TODO, stubbed")
}

// Map_SuperReadPhases
// C: unsigned char * Map_SuperReadPhases( Map_Super_t * p )
func (p *MapSuper) MapSuperReadPhases() []uint8 {
	panic("TODO, stubbed")
}

// Map_SuperReadFanoutLimit
// C: int Map_SuperReadFanoutLimit( Map_Super_t * p )
func (p *MapSuper) MapSuperReadFanoutLimit() int {
	panic("TODO, stubbed")
}

// Map_SuperLibReadGenLib
// C: Mio_Library_t * Map_SuperLibReadGenLib( Map_SuperLib_t * p )
func (p *MapSuperLib) MapSuperLibReadGenLib() *MioLibrary {
	panic("TODO, stubbed")
}

// Map_SuperLibReadAreaInv
// C: float Map_SuperLibReadAreaInv( Map_SuperLib_t * p )
func (p *MapSuperLib) MapSuperLibReadAreaInv() float32 {
	panic("TODO, stubbed")
}

// Map_SuperLibReadDelayInv
// C: Map_Time_t Map_SuperLibReadDelayInv( Map_SuperLib_t * p )
func (p *MapSuperLib) MapSuperLibReadDelayInv() MapTime {
	panic("TODO, stubbed")
}

// Map_SuperLibReadVarsMax
// C: int Map_SuperLibReadVarsMax( Map_SuperLib_t * p )
func (p *MapSuperLib) MapSuperLibReadVarsMax() int {
	panic("TODO, stubbed")
}

// Map_ManCreate
// C: Map_Man_t * Map_ManCreate( int nInputs, int nOutputs, int fVerbose )
func MapManCreate(nInputs int, nOutputs int, fVerbose int) *MapMan {
	panic("TODO, stubbed")
}

// Map_ManFree
// C: void Map_ManFree( Map_Man_t * p )
func (p *MapMan) MapManFree() {
	panic("TODO, stubbed")
}

// Map_ManCreateNodeDelays
// C: void Map_ManCreateNodeDelays( Map_Man_t * p, int LogFan )
func (p *MapMan) MapManCreateNodeDelays(LogFan int) {
	panic("TODO, stubbed")
}

// Map_ManPrintTimeStats
// C: void Map_ManPrintTimeStats( Map_Man_t * p )
func (p *MapMan) MapManPrintTimeStats() {
	panic("TODO, stubbed")
}

// Map_ManPrintStatsToFile
// C: void Map_ManPrintStatsToFile( char * pName, float Area, float Delay, abctime Time )
func MapManPrintStatsToFile(pName string, Area float32, Delay float32, Time int64) {
	panic("TODO, stubbed")
}

// Map_NodeCreate
// C: Map_Node_t * Map_NodeCreate( Map_Man_t * p, Map_Node_t * p1, Map_Node_t * p2 )
func (p *MapMan) MapNodeCreate(p1 *MapNode, p2 *MapNode) *MapNode {
	panic("TODO, stubbed")
}

// Map_TableCreate (static)
// C: void Map_TableCreate( Map_Man_t * pMan )
func (pMan *MapMan) mapTableCreate() {
	panic("TODO, stubbed")
}

// Map_NodeAnd
// C: Map_Node_t * Map_NodeAnd( Map_Man_t * pMan, Map_Node_t * p1, Map_Node_t * p2 )
func (pMan *MapMan) MapNodeAnd(p1 *MapNode, p2 *MapNode) *MapNode {
	panic("TODO, stubbed")
}

// Map_TableResize (static)
// C: void Map_TableResize( Map_Man_t * pMan )
func (pMan *MapMan) mapTableResize() {
	panic("TODO, stubbed")
}

// Map_NodeBuf
// C: Map_Node_t * Map_NodeBuf( Map_Man_t * p, Map_Node_t * p1 )
func (p *MapMan) MapNodeBuf(p1 *MapNode) *MapNode {
	panic("TODO, stubbed")
}

// Map_NodeSetChoice
// C: void Map_NodeSetChoice( Map_Man_t * pMan, Map_Node_t * pNodeOld, Map_Node_t * pNodeNew )
func (pMan *MapMan) MapNodeSetChoice(pNodeOld *MapNode, pNodeNew *MapNode) {
	panic("TODO, stubbed")
}
