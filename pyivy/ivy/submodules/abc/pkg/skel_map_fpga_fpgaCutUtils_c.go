package abc

// Hand-checked skeleton of src/map/fpga/fpgaCutUtils.c.

// Fpga_CutAlloc
// C: Fpga_Cut_t * Fpga_CutAlloc( Fpga_Man_t * p )
func (p *FpgaMan) FpgaCutAlloc() *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_CutDup
// C: Fpga_Cut_t * Fpga_CutDup( Fpga_Man_t * p, Fpga_Cut_t * pCutOld )
func (p *FpgaMan) FpgaCutDup(pCutOld *FpgaCut) *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_CutFree
// C: void Fpga_CutFree( Fpga_Man_t * p, Fpga_Cut_t * pCut )
func (p *FpgaMan) FpgaCutFree(pCut *FpgaCut) {
	panic("TODO, stubbed")
}

// Fpga_CutPrint
// C: void Fpga_CutPrint( Fpga_Man_t * p, Fpga_Node_t * pRoot, Fpga_Cut_t * pCut )
func (p *FpgaMan) FpgaCutPrint(pRoot *FpgaNode, pCut *FpgaCut) {
	panic("TODO, stubbed")
}

// Fpga_CutCreateSimple
// C: Fpga_Cut_t * Fpga_CutCreateSimple( Fpga_Man_t * p, Fpga_Node_t * pNode )
func (p *FpgaMan) FpgaCutCreateSimple(pNode *FpgaNode) *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_CutGetRootArea
// C: float Fpga_CutGetRootArea( Fpga_Man_t * p, Fpga_Cut_t * pCut )
func (p *FpgaMan) FpgaCutGetRootArea(pCut *FpgaCut) float32 {
	panic("TODO, stubbed")
}

// Fpga_CutListAppend
// C: Fpga_Cut_t * Fpga_CutListAppend( Fpga_Cut_t * pSetAll, Fpga_Cut_t * pSets )
func FpgaCutListAppend(pSetAll *FpgaCut, pSets *FpgaCut) *FpgaCut {
	panic("TODO, stubbed")
}

// Fpga_CutListRecycle
// C: void Fpga_CutListRecycle( Fpga_Man_t * p, Fpga_Cut_t * pSetList, Fpga_Cut_t * pSave )
func (p *FpgaMan) FpgaCutListRecycle(pSetList *FpgaCut, pSave *FpgaCut) {
	panic("TODO, stubbed")
}

// Fpga_CutListCount
// C: int Fpga_CutListCount( Fpga_Cut_t * pSets )
func FpgaCutListCount(pSets *FpgaCut) int {
	panic("TODO, stubbed")
}

// Fpga_CutGetParameters
// C: void Fpga_CutGetParameters( Fpga_Man_t * pMan, Fpga_Cut_t * pCut )
func (pMan *FpgaMan) FpgaCutGetParameters(pCut *FpgaCut) {
	panic("TODO, stubbed")
}

// Fpga_CutGetAreaFlow
// C: float Fpga_CutGetAreaFlow( Fpga_Man_t * pMan, Fpga_Cut_t * pCut )
func (pMan *FpgaMan) FpgaCutGetAreaFlow(pCut *FpgaCut) float32 {
	panic("TODO, stubbed")
}

// Fpga_CutGetAreaRefed
// C: float Fpga_CutGetAreaRefed( Fpga_Man_t * pMan, Fpga_Cut_t * pCut )
func (pMan *FpgaMan) FpgaCutGetAreaRefed(pCut *FpgaCut) float32 {
	panic("TODO, stubbed")
}

// Fpga_CutGetAreaDerefed
// C: float Fpga_CutGetAreaDerefed( Fpga_Man_t * pMan, Fpga_Cut_t * pCut )
func (pMan *FpgaMan) FpgaCutGetAreaDerefed(pCut *FpgaCut) float32 {
	panic("TODO, stubbed")
}

// Fpga_CutRef
// C: float Fpga_CutRef( Fpga_Man_t * pMan, Fpga_Node_t * pNode, Fpga_Cut_t * pCut, int fFanouts )
func (pMan *FpgaMan) FpgaCutRef(pNode *FpgaNode, pCut *FpgaCut, fFanouts int) float32 {
	panic("TODO, stubbed")
}

// Fpga_CutDeref
// C: float Fpga_CutDeref( Fpga_Man_t * pMan, Fpga_Node_t * pNode, Fpga_Cut_t * pCut, int fFanouts )
func (pMan *FpgaMan) FpgaCutDeref(pNode *FpgaNode, pCut *FpgaCut, fFanouts int) float32 {
	panic("TODO, stubbed")
}

// Fpga_MappingSetUsedCuts
// C: void Fpga_MappingSetUsedCuts( Fpga_Man_t * pMan )
func (pMan *FpgaMan) FpgaMappingSetUsedCuts() {
	panic("TODO, stubbed")
}
