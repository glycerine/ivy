package abc

// Hand-checked skeleton of src/bool/kit/kitHop.c.

// Kit_GraphToGiaInternal
// C: int Kit_GraphToGiaInternal( Gia_Man_t * pMan, Kit_Graph_t * pGraph, int fHash )
func (pMan *GiaManager) KitGraphToGiaInternal(pGraph *KitGraph, fHash int) int {
	panic("TODO, stubbed")
}

// Kit_GraphToGia
// C: int Kit_GraphToGia( Gia_Man_t * pMan, Kit_Graph_t * pGraph, Vec_Int_t * vLeaves, int fHash )
func (pMan *GiaManager) KitGraphToGia(pGraph *KitGraph, vLeaves *IntVector, fHash int) int {
	panic("TODO, stubbed")
}

// Kit_TruthToGia
// C: int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash )
func (pMan *GiaManager) KitTruthToGia(pTruth []uint32, nVars int, vMemory *IntVector, vLeaves *IntVector, fHash int) int {
	panic("TODO, stubbed")
}

// Kit_TruthToGia2
// C: int Kit_TruthToGia2( Gia_Man_t * pMan, unsigned * pTruth0, unsigned * pTruth1, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash )
func (pMan *GiaManager) KitTruthToGia2(pTruth0 []uint32, pTruth1 []uint32, nVars int, vMemory *IntVector, vLeaves *IntVector, fHash int) int {
	panic("TODO, stubbed")
}

// Kit_GraphToHopInternal
// C: Hop_Obj_t * Kit_GraphToHopInternal( Hop_Man_t * pMan, Kit_Graph_t * pGraph )
func (pMan *HopManager) KitGraphToHopInternal(pGraph *KitGraph) *HopObject {
	panic("TODO, stubbed")
}

// Kit_GraphToHop
// C: Hop_Obj_t * Kit_GraphToHop( Hop_Man_t * pMan, Kit_Graph_t * pGraph )
func (pMan *HopManager) KitGraphToHop(pGraph *KitGraph) *HopObject {
	panic("TODO, stubbed")
}

// Kit_TruthToHop
// C: Hop_Obj_t * Kit_TruthToHop( Hop_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory )
func (pMan *HopManager) KitTruthToHop(pTruth []uint32, nVars int, vMemory *IntVector) *HopObject {
	panic("TODO, stubbed")
}

// Kit_CoverToHop
// C: Hop_Obj_t * Kit_CoverToHop( Hop_Man_t * pMan, Vec_Int_t * vCover, int nVars, Vec_Int_t * vMemory )
func (pMan *HopManager) KitCoverToHop(vCover *IntVector, nVars int, vMemory *IntVector) *HopObject {
	panic("TODO, stubbed")
}
