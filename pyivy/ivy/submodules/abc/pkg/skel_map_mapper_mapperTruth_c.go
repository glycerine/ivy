package abc

// Hand-checked skeleton of src/map/mapper/mapperTruth.c.

// Map_MappingTruths
// C: void Map_MappingTruths( Map_Man_t * pMan )
func (pMan *MapMan) MapMappingTruths() {
	panic("TODO, stubbed")
}

// Map_TruthsCut (static)
// C: void Map_TruthsCut( Map_Man_t * p, Map_Cut_t * pCut )
func (p *MapMan) mapTruthsCut(pCut *MapCut) {
	panic("TODO, stubbed")
}

// Map_TruthsCutOne
// C: void Map_TruthsCutOne( Map_Man_t * p, Map_Cut_t * pCut, unsigned uTruth[] )
func (p *MapMan) MapTruthsCutOne(pCut *MapCut, uTruth []uint32) {
	panic("TODO, stubbed")
}

// Map_CutsCollect_rec (static)
// C: void Map_CutsCollect_rec( Map_Cut_t * pCut, Map_NodeVec_t * vVisited )
func mapCutsCollect_rec(pCut *MapCut, vVisited *MapNodeVec) {
	panic("TODO, stubbed")
}
