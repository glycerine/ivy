package abc

// Hand-checked skeleton of src/map/fpga/fpgaVec.c.

// Fpga_NodeVecAlloc
// C: Fpga_NodeVec_t * Fpga_NodeVecAlloc( int nCap )
func FpgaNodeVecAlloc(nCap int) *FpgaNodeVec {
	panic("TODO, stubbed")
}

// Fpga_NodeVecFree
// C: void Fpga_NodeVecFree( Fpga_NodeVec_t * p )
func (p *FpgaNodeVec) FpgaNodeVecFree() {
	panic("TODO, stubbed")
}

// Fpga_NodeVecReadArray
// C: Fpga_Node_t ** Fpga_NodeVecReadArray( Fpga_NodeVec_t * p )
func (p *FpgaNodeVec) FpgaNodeVecReadArray() []*FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_NodeVecReadSize
// C: int Fpga_NodeVecReadSize( Fpga_NodeVec_t * p )
func (p *FpgaNodeVec) FpgaNodeVecReadSize() int {
	panic("TODO, stubbed")
}

// Fpga_NodeVecGrow
// C: void Fpga_NodeVecGrow( Fpga_NodeVec_t * p, int nCapMin )
func (p *FpgaNodeVec) FpgaNodeVecGrow(nCapMin int) {
	panic("TODO, stubbed")
}

// Fpga_NodeVecShrink
// C: void Fpga_NodeVecShrink( Fpga_NodeVec_t * p, int nSizeNew )
func (p *FpgaNodeVec) FpgaNodeVecShrink(nSizeNew int) {
	panic("TODO, stubbed")
}

// Fpga_NodeVecClear
// C: void Fpga_NodeVecClear( Fpga_NodeVec_t * p )
func (p *FpgaNodeVec) FpgaNodeVecClear() {
	panic("TODO, stubbed")
}

// Fpga_NodeVecPush
// C: void Fpga_NodeVecPush( Fpga_NodeVec_t * p, Fpga_Node_t * Entry )
func (p *FpgaNodeVec) FpgaNodeVecPush(Entry *FpgaNode) {
	panic("TODO, stubbed")
}

// Fpga_NodeVecPushUnique
// C: int Fpga_NodeVecPushUnique( Fpga_NodeVec_t * p, Fpga_Node_t * Entry )
func (p *FpgaNodeVec) FpgaNodeVecPushUnique(Entry *FpgaNode) int {
	panic("TODO, stubbed")
}

// Fpga_NodeVecPop
// C: Fpga_Node_t * Fpga_NodeVecPop( Fpga_NodeVec_t * p )
func (p *FpgaNodeVec) FpgaNodeVecPop() *FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_NodeVecWriteEntry
// C: void Fpga_NodeVecWriteEntry( Fpga_NodeVec_t * p, int i, Fpga_Node_t * Entry )
func (p *FpgaNodeVec) FpgaNodeVecWriteEntry(i int, Entry *FpgaNode) {
	panic("TODO, stubbed")
}

// Fpga_NodeVecReadEntry
// C: Fpga_Node_t * Fpga_NodeVecReadEntry( Fpga_NodeVec_t * p, int i )
func (p *FpgaNodeVec) FpgaNodeVecReadEntry(i int) *FpgaNode {
	panic("TODO, stubbed")
}

// Fpga_NodeVecCompareLevels (static)
// C: int Fpga_NodeVecCompareLevels( Fpga_Node_t ** pp1, Fpga_Node_t ** pp2 )
func fpgaNodeVecCompareLevels(pp1 **FpgaNode, pp2 **FpgaNode) int {
	panic("TODO, stubbed")
}

// Fpga_NodeVecSortByLevel
// C: void Fpga_NodeVecSortByLevel( Fpga_NodeVec_t * p )
func (p *FpgaNodeVec) FpgaNodeVecSortByLevel() {
	panic("TODO, stubbed")
}

// Fpga_NodeVecCompareArrivals
// C: int Fpga_NodeVecCompareArrivals( Fpga_Node_t ** ppS1, Fpga_Node_t ** ppS2 )
func FpgaNodeVecCompareArrivals(ppS1 **FpgaNode, ppS2 **FpgaNode) int {
	panic("TODO, stubbed")
}

// Fpga_SortNodesByArrivalTimes
// C: void Fpga_SortNodesByArrivalTimes( Fpga_NodeVec_t * p )
func (p *FpgaNodeVec) FpgaSortNodesByArrivalTimes() {
	panic("TODO, stubbed")
}

// Fpga_NodeVecUnion
// C: void Fpga_NodeVecUnion( Fpga_NodeVec_t * p, Fpga_NodeVec_t * p1, Fpga_NodeVec_t * p2 )
func (p *FpgaNodeVec) FpgaNodeVecUnion(p1 *FpgaNodeVec, p2 *FpgaNodeVec) {
	panic("TODO, stubbed")
}

// Fpga_NodeVecPushOrder
// C: void Fpga_NodeVecPushOrder( Fpga_NodeVec_t * vNodes, Fpga_Node_t * pNode, int fIncreasing )
func (vNodes *FpgaNodeVec) FpgaNodeVecPushOrder(pNode *FpgaNode, fIncreasing int) {
	panic("TODO, stubbed")
}

// Fpga_NodeVecReverse
// C: void Fpga_NodeVecReverse( Fpga_NodeVec_t * vNodes )
func (vNodes *FpgaNodeVec) FpgaNodeVecReverse() {
	panic("TODO, stubbed")
}
