package abc

// Hand-checked skeleton of src/map/scl/sclBuffer.c.

// BUF_SCALE is 1000.

// BufMan is Buf_Man_t.
type BufMan struct {
	// parameters
	nFanMin int // the smallest fanout count to consider
	nFanMax int // the largest fanout count allowed off CP
	fBufPis int // enables buffing of the combinational inputs
	// internal deta
	pNtk      *ABCNetwork         // logic network
	vOffsets  *IntVector          // offsets into edge delays
	vEdges    *IntVector          // edge delays
	vArr      *IntVector          // arrival times
	vDep      *IntVector          // departure times
	vCounts   *FloatVector        // fanout counts
	vQue      *FloatPriorityQueue // Vec_Que_t * : queue by fanout count
	nObjStart int                 // the number of starting objects
	nObjAlloc int                 // the number of allocated objects
	DelayMax  int                 // maximum delay (percentage of inverter delay)
	DelayInv  float32             // inverter delay
	// sorting fanouts
	vOrder   *IntVector                       // ordering of fanouts
	vDelays  *IntVector                       // fanout delays
	vNonCrit *IntVector                       // non-critical fanouts
	vTfCone  *IntVector                       // TFI/TFO cone of the node including the node
	vFanouts *PointerVector[ABCNetworkObject] // temp storage for fanouts
	// statistics
	nSeparate   int
	nDuplicate  int
	nBranch0    int
	nBranch1    int
	nBranchCrit int
}

// Abc_BufNodeArr (static)
// C: int Abc_BufNodeArr( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) abcBufNodeArr(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_BufNodeDep (static)
// C: int Abc_BufNodeDep( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) abcBufNodeDep(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_BufSetNodeArr (static)
// C: void Abc_BufSetNodeArr( Buf_Man_t * p, Abc_Obj_t * pObj, int f )
func (p *BufMan) abcBufSetNodeArr(pObj *ABCNetworkObject, f int) {
	panic("TODO, stubbed")
}

// Abc_BufSetNodeDep (static)
// C: void Abc_BufSetNodeDep( Buf_Man_t * p, Abc_Obj_t * pObj, int f )
func (p *BufMan) abcBufSetNodeDep(pObj *ABCNetworkObject, f int) {
	panic("TODO, stubbed")
}

// Abc_BufEdgeDelay (static)
// C: int Abc_BufEdgeDelay( Buf_Man_t * p, Abc_Obj_t * pObj, int i )
func (p *BufMan) abcBufEdgeDelay(pObj *ABCNetworkObject, i int) int {
	panic("TODO, stubbed")
}

// Abc_BufSetEdgeDelay (static)
// C: void Abc_BufSetEdgeDelay( Buf_Man_t * p, Abc_Obj_t * pObj, int i, int f )
func (p *BufMan) abcBufSetEdgeDelay(pObj *ABCNetworkObject, i int, f int) {
	panic("TODO, stubbed")
}

// Abc_BufNodeSlack (static)
// C: int Abc_BufNodeSlack( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) abcBufNodeSlack(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_BufEdgeSlack (static)
// C: int Abc_BufEdgeSlack( Buf_Man_t * p, Abc_Obj_t * pObj, Abc_Obj_t * pFan )
func (p *BufMan) abcBufEdgeSlack(pObj *ABCNetworkObject, pFan *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclReportDupFanins
// C: void Abc_SclReportDupFanins( Abc_Ntk_t * pNtk )
func AbcSclReportDupFanins(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_SclObjIsBufInv (static)
// C: int Abc_SclObjIsBufInv( Abc_Obj_t * pObj )
func abcSclObjIsBufInv(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclIsInv
// C: int Abc_SclIsInv( Abc_Obj_t * pObj )
func AbcSclIsInv(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclGetRealFaninLit
// C: int Abc_SclGetRealFaninLit( Abc_Obj_t * pObj )
func AbcSclGetRealFaninLit(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclUnBufferPerform
// C: Abc_Ntk_t * Abc_SclUnBufferPerform( Abc_Ntk_t * pNtk, int fVerbose )
func AbcSclUnBufferPerform(pNtk *ABCNetwork, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_SclCountMaxPhases
// C: int Abc_SclCountMaxPhases( Abc_Ntk_t * pNtk )
func AbcSclCountMaxPhases(pNtk *ABCNetwork) int {
	panic("TODO, stubbed")
}

// Abc_SclBufferPhase
// C: Abc_Ntk_t * Abc_SclBufferPhase( Abc_Ntk_t * pNtk, int fVerbose )
func AbcSclBufferPhase(pNtk *ABCNetwork, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_SclUnBufferPhase
// C: Abc_Ntk_t * Abc_SclUnBufferPhase( Abc_Ntk_t * pNtk, int fVerbose )
func AbcSclUnBufferPhase(pNtk *ABCNetwork, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_SclCheckNtk
// C: int Abc_SclCheckNtk( Abc_Ntk_t * p, int fVerbose )
func AbcSclCheckNtk(p *ABCNetwork, fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NodeInvUpdateFanPolarity
// C: void Abc_NodeInvUpdateFanPolarity( Abc_Obj_t * pObj )
func AbcNodeInvUpdateFanPolarity(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NodeInvUpdateObjFanoutPolarity
// C: void Abc_NodeInvUpdateObjFanoutPolarity( Abc_Obj_t * pObj, Abc_Obj_t * pFanout )
func AbcNodeInvUpdateObjFanoutPolarity(pObj *ABCNetworkObject, pFanout *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NodeCompareLevels
// C: int Abc_NodeCompareLevels( Abc_Obj_t ** pp1, Abc_Obj_t ** pp2 )
func AbcNodeCompareLevels(pp1 **ABCNetworkObject, pp2 **ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclComputeReverseLevel
// C: int Abc_SclComputeReverseLevel( Abc_Obj_t * pObj )
func AbcSclComputeReverseLevel(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_SclPerformBufferingOne
// C: Abc_Obj_t * Abc_SclPerformBufferingOne( Abc_Obj_t * pObj, int Degree, int fUseInvs, int fVerbose )
func AbcSclPerformBufferingOne(pObj *ABCNetworkObject, Degree int, fUseInvs int, fVerbose int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_SclPerformBuffering_rec
// C: void Abc_SclPerformBuffering_rec( Abc_Obj_t * pObj, int DegreeR, int Degree, int fUseInvs, int fVerbose )
func AbcSclPerformBuffering_rec(pObj *ABCNetworkObject, DegreeR int, Degree int, fUseInvs int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_SclPerformBuffering
// C: Abc_Ntk_t * Abc_SclPerformBuffering( Abc_Ntk_t * p, int DegreeR, int Degree, int fUseInvs, int fVerbose )
func AbcSclPerformBuffering(p *ABCNetwork, DegreeR int, Degree int, fUseInvs int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_BufComputeArr
// C: float Abc_BufComputeArr( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) AbcBufComputeArr(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_BufComputeDep
// C: float Abc_BufComputeDep( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) AbcBufComputeDep(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_BufUpdateGlobal
// C: void Abc_BufUpdateGlobal( Buf_Man_t * p )
func (p *BufMan) AbcBufUpdateGlobal() {
	panic("TODO, stubbed")
}

// Abc_BufCreateEdges
// C: void Abc_BufCreateEdges( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) AbcBufCreateEdges(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_BufAddToQue
// C: void Abc_BufAddToQue( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) AbcBufAddToQue(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_BufCollectTfoCone_rec
// C: void Abc_BufCollectTfoCone_rec( Abc_Obj_t * pNode, Vec_Int_t * vNodes )
func AbcBufCollectTfoCone_rec(pNode *ABCNetworkObject, vNodes *IntVector) {
	panic("TODO, stubbed")
}

// Abc_BufCollectTfoCone
// C: void Abc_BufCollectTfoCone( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) AbcBufCollectTfoCone(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_BufUpdateArr
// C: void Abc_BufUpdateArr( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) AbcBufUpdateArr(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_BufCollectTfiCone_rec
// C: void Abc_BufCollectTfiCone_rec( Abc_Obj_t * pNode, Vec_Int_t * vNodes )
func AbcBufCollectTfiCone_rec(pNode *ABCNetworkObject, vNodes *IntVector) {
	panic("TODO, stubbed")
}

// Abc_BufCollectTfiCone
// C: void Abc_BufCollectTfiCone( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) AbcBufCollectTfiCone(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_BufUpdateDep
// C: void Abc_BufUpdateDep( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) AbcBufUpdateDep(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Buf_ManStart
// C: Buf_Man_t * Buf_ManStart( Abc_Ntk_t * pNtk, int FanMin, int FanMax, int fBufPis )
func BufManStart(pNtk *ABCNetwork, FanMin int, FanMax int, fBufPis int) *BufMan {
	panic("TODO, stubbed")
}

// Buf_ManStop
// C: void Buf_ManStop( Buf_Man_t * p )
func (p *BufMan) BufManStop() {
	panic("TODO, stubbed")
}

// Abc_BufSortByDelay
// C: Vec_Int_t * Abc_BufSortByDelay( Buf_Man_t * p, int iPivot )
func (p *BufMan) AbcBufSortByDelay(iPivot int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_BufPrintOne
// C: void Abc_BufPrintOne( Buf_Man_t * p, int iPivot )
func (p *BufMan) AbcBufPrintOne(iPivot int) {
	panic("TODO, stubbed")
}

// Abc_BufReplaceBufsByInvs
// C: void Abc_BufReplaceBufsByInvs( Abc_Ntk_t * pNtk )
func AbcBufReplaceBufsByInvs(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_BufComputeAverage
// C: int Abc_BufComputeAverage( Buf_Man_t * p, int iPivot, Vec_Int_t * vOrder )
func (p *BufMan) AbcBufComputeAverage(iPivot int, vOrder *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_BufFindNonBuffDriver
// C: Abc_Obj_t * Abc_BufFindNonBuffDriver( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) AbcBufFindNonBuffDriver(pObj *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_BufCountNonCritical
// C: int Abc_BufCountNonCritical( Buf_Man_t * p, Abc_Obj_t * pObj )
func (p *BufMan) AbcBufCountNonCritical(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_BufPerformOne
// C: void Abc_BufPerformOne( Buf_Man_t * p, int iPivot, int fSkipDup, int fVerbose )
func (p *BufMan) AbcBufPerformOne(iPivot int, fSkipDup int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_SclBufPerform
// C: Abc_Ntk_t * Abc_SclBufPerform( Abc_Ntk_t * pNtk, int FanMin, int FanMax, int fBufPis, int fSkipDup, int fVerbose )
func AbcSclBufPerform(pNtk *ABCNetwork, FanMin int, FanMax int, fBufPis int, fSkipDup int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}
