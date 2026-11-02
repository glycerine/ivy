package abc

// Hand-checked skeleton of src/map/fpga/fpgaCut.c.

// FpgaCutTable is Fpga_CutTable_t (struct Fpga_CutTableStrutct_t; the C tag
// spells "Strutct").
type FpgaCutTable struct {
	pBins  []*FpgaCut // Fpga_Cut_t ** : the table used for linear probing
	nBins  int        // the size of the table
	pCuts  []int      // int * : the array of cuts currently stored
	nCuts  int        // the number of cuts currently stored
	pArray []*FpgaCut // Fpga_Cut_t ** : the temporary array of cuts
	pCuts1 []*FpgaCut // Fpga_Cut_t ** : the temporary array of cuts
	pCuts2 []*FpgaCut // Fpga_Cut_t ** : the temporary array of cuts
}

// Fpga_MappingCuts
// C: void Fpga_MappingCuts( Fpga_Man_t * p )
func (p *FpgaMan) FpgaMappingCuts() {
	panic("TODO, stubbed")
}

// Fpga_MappingCreatePiCuts
// C: void Fpga_MappingCreatePiCuts( Fpga_Man_t * p )
func (p *FpgaMan) FpgaMappingCreatePiCuts() {
	panic("TODO, stubbed")
}

// Fpga_CutCompute (static)
// C: Fpga_Cut_t * Fpga_CutCompute( Fpga_Man_t * p, Fpga_CutTable_t * pTable, Fpga_Node_t * pNode )
func (p *FpgaMan) fpgaCutCompute(pTable *FpgaCutTable, pNode *FpgaNode) *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_CutFilter (static)
// C: void Fpga_CutFilter( Fpga_Man_t * p, Fpga_Node_t * pNode )
func (p *FpgaMan) fpgaCutFilter(pNode *FpgaNode) {
	panic("TODO, stubbed")
}

// Fpga_CutMergeLists (static)
// C: Fpga_Cut_t * Fpga_CutMergeLists( Fpga_Man_t * p, Fpga_CutTable_t * pTable, Fpga_Cut_t * pList1, Fpga_Cut_t * pList2, int fComp1, int fComp2, int fPivot1, int fPivot2 )
func (p *FpgaMan) fpgaCutMergeLists(pTable *FpgaCutTable, pList1 *FpgaCut, pList2 *FpgaCut, fComp1 int, fComp2 int, fPivot1 int, fPivot2 int) *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_CutMergeLists2
// C: Fpga_Cut_t * Fpga_CutMergeLists2( Fpga_Man_t * p, Fpga_CutTable_t * pTable, Fpga_Cut_t * pList1, Fpga_Cut_t * pList2, int fComp1, int fComp2, int fPivot1, int fPivot2 )
func (p *FpgaMan) FpgaCutMergeLists2(pTable *FpgaCutTable, pList1 *FpgaCut, pList2 *FpgaCut, fComp1 int, fComp2 int, fPivot1 int, fPivot2 int) *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_CutMergeTwo (static)
// C: int Fpga_CutMergeTwo( Fpga_Cut_t * pCut1, Fpga_Cut_t * pCut2, Fpga_Node_t * ppNodes[], int nNodesMax )
func fpgaCutMergeTwo(pCut1 *FpgaCut, pCut2 *FpgaCut, ppNodes []*FpgaNode, nNodesMax int) int {
	panic("TODO, stubbed")
}

// Fpga_CutUnionLists (static)
// C: Fpga_Cut_t * Fpga_CutUnionLists( Fpga_Cut_t * pList1, Fpga_Cut_t * pList2 )
func fpgaCutUnionLists(pList1 *FpgaCut, pList2 *FpgaCut) *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_CutBelongsToList (static)
// C: int Fpga_CutBelongsToList( Fpga_Cut_t * pList, Fpga_Node_t * ppNodes[], int nNodes )
func fpgaCutBelongsToList(pList *FpgaCut, ppNodes []*FpgaNode, nNodes int) int {
	panic("TODO, stubbed")
}

// Fpga_CutCountAll
// C: int Fpga_CutCountAll( Fpga_Man_t * pMan )
func (pMan *FpgaMan) FpgaCutCountAll() int {
	panic("TODO, stubbed")
}

// Fpga_CutsCleanSign
// C: void Fpga_CutsCleanSign( Fpga_Man_t * pMan )
func (pMan *FpgaMan) FpgaCutsCleanSign() {
	panic("TODO, stubbed")
}

// Fpga_CutsCleanRoot
// C: void Fpga_CutsCleanRoot( Fpga_Man_t * pMan )
func (pMan *FpgaMan) FpgaCutsCleanRoot() {
	panic("TODO, stubbed")
}

// Fpga_CutListPrint (static)
// C: void Fpga_CutListPrint( Fpga_Man_t * pMan, Fpga_Node_t * pRoot )
func (pMan *FpgaMan) fpgaCutListPrint(pRoot *FpgaNode) {
	panic("TODO, stubbed")
}

// Fpga_CutListPrint2 (static)
// C: void Fpga_CutListPrint2( Fpga_Man_t * pMan, Fpga_Node_t * pRoot )
func (pMan *FpgaMan) fpgaCutListPrint2(pRoot *FpgaNode) {
	panic("TODO, stubbed")
}

// Fpga_CutPrint_ (static)
// C: void Fpga_CutPrint_( Fpga_Man_t * pMan, Fpga_Cut_t * pCut, Fpga_Node_t * pRoot )
func (pMan *FpgaMan) fpgaCutPrint_(pCut *FpgaCut, pRoot *FpgaNode) {
	panic("TODO, stubbed")
}

// Fpga_CutTableStart (static)
// C: Fpga_CutTable_t * Fpga_CutTableStart( Fpga_Man_t * pMan )
func (pMan *FpgaMan) fpgaCutTableStart() *FpgaCutTable {
	panic("TODO, stubbed")
}

// Fpga_CutTableStop (static)
// C: void Fpga_CutTableStop( Fpga_CutTable_t * p )
func (p *FpgaCutTable) fpgaCutTableStop() {
	panic("TODO, stubbed")
}

// Fpga_CutTableHash (static)
// C: unsigned Fpga_CutTableHash( Fpga_Node_t * ppNodes[], int nNodes )
func fpgaCutTableHash(ppNodes []*FpgaNode, nNodes int) uint32 {
	panic("TODO, stubbed")
}

// Fpga_CutTableLookup (static)
// C: int Fpga_CutTableLookup( Fpga_CutTable_t * p, Fpga_Node_t * ppNodes[], int nNodes )
func (p *FpgaCutTable) fpgaCutTableLookup(ppNodes []*FpgaNode, nNodes int) int {
	panic("TODO, stubbed")
}

// Fpga_CutTableConsider (static)
// C: Fpga_Cut_t * Fpga_CutTableConsider( Fpga_Man_t * pMan, Fpga_CutTable_t * p, Fpga_Node_t * ppNodes[], int nNodes )
func (pMan *FpgaMan) fpgaCutTableConsider(p *FpgaCutTable, ppNodes []*FpgaNode, nNodes int) *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_CutTableRestart (static)
// C: void Fpga_CutTableRestart( Fpga_CutTable_t * p )
func (p *FpgaCutTable) fpgaCutTableRestart() {
	panic("TODO, stubbed")
}

// Fpga_CutSortCutsCompare (static)
// C: int Fpga_CutSortCutsCompare( Fpga_Cut_t ** pC1, Fpga_Cut_t ** pC2 )
func fpgaCutSortCutsCompare(pC1 **FpgaCut, pC2 **FpgaCut) int {
	panic("TODO, stubbed")
}

// Fpga_CutSortCuts (static)
// C: Fpga_Cut_t * Fpga_CutSortCuts( Fpga_Man_t * pMan, Fpga_CutTable_t * p, Fpga_Cut_t * pList )
func (pMan *FpgaMan) fpgaCutSortCuts(p *FpgaCutTable, pList *FpgaCut) *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_CutList2Array (static)
// C: int Fpga_CutList2Array( Fpga_Cut_t ** pArray, Fpga_Cut_t * pList )
func fpgaCutList2Array(pArray []*FpgaCut, pList *FpgaCut) int {
	panic("TODO, stubbed")
}

// Fpga_CutArray2List (static)
// C: Fpga_Cut_t * Fpga_CutArray2List( Fpga_Cut_t ** pArray, int nCuts )
func fpgaCutArray2List(pArray []*FpgaCut, nCuts int) *FpgaCut {
	panic("TODO, stubbed")
}
