package abc

// Hand-checked skeleton of src/map/fpga/fpgaUtils.c.

// Fpga_MappingDfs
// C: Fpga_NodeVec_t * Fpga_MappingDfs( Fpga_Man_t * pMan, int fCollectEquiv )
func (pMan *FpgaMan) FpgaMappingDfs(fCollectEquiv int) *FpgaNodeVec {
	panic("TODO, stubbed")
}

// Fpga_MappingDfs_rec (static)
// C: void Fpga_MappingDfs_rec( Fpga_Node_t * pNode, Fpga_NodeVec_t * vNodes, int fCollectEquiv )
func fpgaMappingDfs_rec(pNode *FpgaNode, vNodes *FpgaNodeVec, fCollectEquiv int) {
	panic("TODO, stubbed")
}

// Fpga_MappingDfsNodes
// C: Fpga_NodeVec_t * Fpga_MappingDfsNodes( Fpga_Man_t * pMan, Fpga_Node_t ** ppNodes, int nNodes, int fEquiv )
func (pMan *FpgaMan) FpgaMappingDfsNodes(ppNodes []*FpgaNode, nNodes int, fEquiv int) *FpgaNodeVec {
	panic("TODO, stubbed")
}

// Fpga_MappingGetAreaFlow
// C: float Fpga_MappingGetAreaFlow( Fpga_Man_t * p )
func (p *FpgaMan) FpgaMappingGetAreaFlow() float32 {
	panic("TODO, stubbed")
}

// Fpga_MappingArea
// C: float Fpga_MappingArea( Fpga_Man_t * pMan )
func (pMan *FpgaMan) FpgaMappingArea() float32 {
	panic("TODO, stubbed")
}

// Fpga_MappingArea_rec
// C: float Fpga_MappingArea_rec( Fpga_Man_t * pMan, Fpga_Node_t * pNode, Fpga_NodeVec_t * vNodes )
func (pMan *FpgaMan) FpgaMappingArea_rec(pNode *FpgaNode, vNodes *FpgaNodeVec) float32 {
	panic("TODO, stubbed")
}

// Fpga_MappingAreaTrav
// C: float Fpga_MappingAreaTrav( Fpga_Man_t * pMan )
func (pMan *FpgaMan) FpgaMappingAreaTrav() float32 {
	panic("TODO, stubbed")
}

// Fpga_MappingSetRefsAndArea_rec
// C: float Fpga_MappingSetRefsAndArea_rec( Fpga_Man_t * pMan, Fpga_Node_t * pNode, Fpga_Node_t ** ppStore )
func (pMan *FpgaMan) FpgaMappingSetRefsAndArea_rec(pNode *FpgaNode, ppStore []*FpgaNode) float32 {
	panic("TODO, stubbed")
}

// Fpga_MappingSetRefsAndArea
// C: float Fpga_MappingSetRefsAndArea( Fpga_Man_t * pMan )
func (pMan *FpgaMan) FpgaMappingSetRefsAndArea() float32 {
	panic("TODO, stubbed")
}

// Fpga_MappingCompareOutputDelay (static)
// C: int Fpga_MappingCompareOutputDelay( Fpga_Node_t ** ppNode1, Fpga_Node_t ** ppNode2 )
func fpgaMappingCompareOutputDelay(ppNode1 **FpgaNode, ppNode2 **FpgaNode) int {
	panic("TODO, stubbed")
}

// Fpga_MappingFindLatest (static)
// C: void Fpga_MappingFindLatest( Fpga_Man_t * p, int * pNodes, int nNodesMax )
func (p *FpgaMan) fpgaMappingFindLatest(pNodes []int, nNodesMax int) {
	panic("TODO, stubbed")
}

// Fpga_MappingPrintOutputArrivals
// C: void Fpga_MappingPrintOutputArrivals( Fpga_Man_t * p )
func (p *FpgaMan) FpgaMappingPrintOutputArrivals() {
	panic("TODO, stubbed")
}

// Fpga_MappingSetupTruthTables
// C: void Fpga_MappingSetupTruthTables( unsigned uTruths[][2] )
func FpgaMappingSetupTruthTables(uTruths [][2]uint32) {
	panic("TODO, stubbed")
}

// Fpga_MappingSetupMask
// C: void Fpga_MappingSetupMask( unsigned uMask[], int nVarsMax )
func FpgaMappingSetupMask(uMask []uint32, nVarsMax int) {
	panic("TODO, stubbed")
}

// Fpga_ManCheckConsistency
// C: int Fpga_ManCheckConsistency( Fpga_Man_t * p )
func (p *FpgaMan) FpgaManCheckConsistency() int {
	panic("TODO, stubbed")
}

// Fpga_CompareNodesByLevelDecreasing
// C: int Fpga_CompareNodesByLevelDecreasing( Fpga_Node_t ** ppS1, Fpga_Node_t ** ppS2 )
func FpgaCompareNodesByLevelDecreasing(ppS1 **FpgaNode, ppS2 **FpgaNode) int {
	panic("TODO, stubbed")
}

// Fpga_CompareNodesByLevelIncreasing
// C: int Fpga_CompareNodesByLevelIncreasing( Fpga_Node_t ** ppS1, Fpga_Node_t ** ppS2 )
func FpgaCompareNodesByLevelIncreasing(ppS1 **FpgaNode, ppS2 **FpgaNode) int {
	panic("TODO, stubbed")
}

// Fpga_MappingSortByLevel
// C: void Fpga_MappingSortByLevel( Fpga_Man_t * pMan, Fpga_NodeVec_t * vNodes, int fIncreasing )
func (pMan *FpgaMan) FpgaMappingSortByLevel(vNodes *FpgaNodeVec, fIncreasing int) {
	panic("TODO, stubbed")
}

// Fpga_DfsLim
// C: Fpga_NodeVec_t * Fpga_DfsLim( Fpga_Man_t * pMan, Fpga_Node_t * pNode, int nLevels )
func (pMan *FpgaMan) FpgaDfsLim(pNode *FpgaNode, nLevels int) *FpgaNodeVec {
	panic("TODO, stubbed")
}

// Fpga_DfsLim_rec (static)
// C: void Fpga_DfsLim_rec( Fpga_Node_t * pNode, int Level, Fpga_NodeVec_t * vNodes )
func fpgaDfsLim_rec(pNode *FpgaNode, Level int, vNodes *FpgaNodeVec) {
	panic("TODO, stubbed")
}

// Fpga_ManCleanData0
// C: void Fpga_ManCleanData0( Fpga_Man_t * pMan )
func (pMan *FpgaMan) FpgaManCleanData0() {
	panic("TODO, stubbed")
}

// Fpga_CollectNodeTfo
// C: Fpga_NodeVec_t * Fpga_CollectNodeTfo( Fpga_Man_t * pMan, Fpga_Node_t * pNode )
func (pMan *FpgaMan) FpgaCollectNodeTfo(pNode *FpgaNode) *FpgaNodeVec {
	panic("TODO, stubbed")
}

// Fpga_CollectNodeTfo_rec (static)
// C: int Fpga_CollectNodeTfo_rec( Fpga_Node_t * pNode, Fpga_Node_t * pPivot, Fpga_NodeVec_t * vVisited, Fpga_NodeVec_t * vTfo )
func fpgaCollectNodeTfo_rec(pNode *FpgaNode, pPivot *FpgaNode, vVisited *FpgaNodeVec, vTfo *FpgaNodeVec) int {
	panic("TODO, stubbed")
}

// Fpga_MappingLevelize
// C: Fpga_NodeVec_t * Fpga_MappingLevelize( Fpga_Man_t * pMan, Fpga_NodeVec_t * vNodes )
func (pMan *FpgaMan) FpgaMappingLevelize(vNodes *FpgaNodeVec) *FpgaNodeVec {
	panic("TODO, stubbed")
}

// Fpga_MappingMaxLevel
// C: int Fpga_MappingMaxLevel( Fpga_Man_t * pMan )
func (pMan *FpgaMan) FpgaMappingMaxLevel() int {
	panic("TODO, stubbed")
}

// Fpga_MappingUpdateLevel_rec
// C: int Fpga_MappingUpdateLevel_rec( Fpga_Man_t * pMan, Fpga_Node_t * pNode, int fMaximum )
func (pMan *FpgaMan) FpgaMappingUpdateLevel_rec(pNode *FpgaNode, fMaximum int) int {
	panic("TODO, stubbed")
}

// Fpga_MappingSetChoiceLevels
// C: void Fpga_MappingSetChoiceLevels( Fpga_Man_t * pMan )
func (pMan *FpgaMan) FpgaMappingSetChoiceLevels() {
	panic("TODO, stubbed")
}

// Fpga_ManReportChoices
// C: void Fpga_ManReportChoices( Fpga_Man_t * pMan )
func (pMan *FpgaMan) FpgaManReportChoices() {
	panic("TODO, stubbed")
}

// Fpga_MappingOrderCosByLevel (static)
// C: Fpga_NodeVec_t * Fpga_MappingOrderCosByLevel( Fpga_Man_t * pMan )
func (pMan *FpgaMan) fpgaMappingOrderCosByLevel() *FpgaNodeVec {
	panic("TODO, stubbed")
}
