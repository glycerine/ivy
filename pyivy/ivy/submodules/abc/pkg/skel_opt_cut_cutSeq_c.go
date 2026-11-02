package abc

// Hand-checked skeleton of src/opt/cut/cutSeq.c.

// Cut_NodeShiftCutLeaves (static inline)
// C: void Cut_NodeShiftCutLeaves( Cut_Cut_t * pList, int nLat )
func cutNodeShiftCutLeaves(pList *CutCut, nLat int) {
	panic("TODO, stubbed")
}

// Cut_NodeComputeCutsSeq
// C: void Cut_NodeComputeCutsSeq( Cut_Man_t * p, int Node, int Node0, int Node1, int fCompl0, int fCompl1, int nLat0, int nLat1, int fTriv, int CutSetNum )
func (p *CutMan) CutNodeComputeCutsSeq(Node int, Node0 int, Node1 int, fCompl0 int, fCompl1 int, nLat0 int, nLat1 int, fTriv int, CutSetNum int) {
	panic("TODO, stubbed")
}

// Cut_NodeNewMergeWithOld
// C: void Cut_NodeNewMergeWithOld( Cut_Man_t * p, int Node )
func (p *CutMan) CutNodeNewMergeWithOld(Node int) {
	panic("TODO, stubbed")
}

// Cut_NodeTempTransferToNew
// C: int Cut_NodeTempTransferToNew( Cut_Man_t * p, int Node, int CutSetNum )
func (p *CutMan) CutNodeTempTransferToNew(Node int, CutSetNum int) int {
	panic("TODO, stubbed")
}

// Cut_NodeOldTransferToNew
// C: void Cut_NodeOldTransferToNew( Cut_Man_t * p, int Node )
func (p *CutMan) CutNodeOldTransferToNew(Node int) {
	panic("TODO, stubbed")
}
