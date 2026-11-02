package abc

// Hand-checked skeleton of src/map/mapper/mapperRefs.c.

// Map_NodeReadRefPhaseAct
// C: int Map_NodeReadRefPhaseAct( Map_Node_t * pNode, int fPhase )
func (pNode *MapNode) MapNodeReadRefPhaseAct(fPhase int) int {
	panic("TODO, stubbed")
}

// Map_NodeReadRefPhaseEst
// C: float Map_NodeReadRefPhaseEst( Map_Node_t * pNode, int fPhase )
func (pNode *MapNode) MapNodeReadRefPhaseEst(fPhase int) float32 {
	panic("TODO, stubbed")
}

// Map_NodeIncRefPhaseAct
// C: int Map_NodeIncRefPhaseAct( Map_Node_t * pNode, int fPhase )
func (pNode *MapNode) MapNodeIncRefPhaseAct(fPhase int) int {
	panic("TODO, stubbed")
}

// Map_NodeDecRefPhaseAct
// C: int Map_NodeDecRefPhaseAct( Map_Node_t * pNode, int fPhase )
func (pNode *MapNode) MapNodeDecRefPhaseAct(fPhase int) int {
	panic("TODO, stubbed")
}

// Map_MappingEstimateRefsInit
// C: void Map_MappingEstimateRefsInit( Map_Man_t * p )
func (p *MapMan) MapMappingEstimateRefsInit() {
	panic("TODO, stubbed")
}

// Map_MappingEstimateRefs
// C: void Map_MappingEstimateRefs( Map_Man_t * p )
func (p *MapMan) MapMappingEstimateRefs() {
	panic("TODO, stubbed")
}

// Map_CutGetAreaFlow
// C: float Map_CutGetAreaFlow( Map_Cut_t * pCut, int fPhase )
func (pCut *MapCut) MapCutGetAreaFlow(fPhase int) float32 {
	panic("TODO, stubbed")
}

// Map_CutRefDeref
// C: float Map_CutRefDeref( Map_Cut_t * pCut, int fPhase, int fReference, int fUpdateProf )
func (pCut *MapCut) MapCutRefDeref(fPhase int, fReference int, fUpdateProf int) float32 {
	panic("TODO, stubbed")
}

// Map_CutGetAreaRefed
// C: float Map_CutGetAreaRefed( Map_Cut_t * pCut, int fPhase )
func (pCut *MapCut) MapCutGetAreaRefed(fPhase int) float32 {
	panic("TODO, stubbed")
}

// Map_CutGetAreaDerefed
// C: float Map_CutGetAreaDerefed( Map_Cut_t * pCut, int fPhase )
func (pCut *MapCut) MapCutGetAreaDerefed(fPhase int) float32 {
	panic("TODO, stubbed")
}

// Map_CutRef
// C: float Map_CutRef( Map_Cut_t * pCut, int fPhase, int fProfile )
func (pCut *MapCut) MapCutRef(fPhase int, fProfile int) float32 {
	panic("TODO, stubbed")
}

// Map_CutDeref
// C: float Map_CutDeref( Map_Cut_t * pCut, int fPhase, int fProfile )
func (pCut *MapCut) MapCutDeref(fPhase int, fProfile int) float32 {
	panic("TODO, stubbed")
}

// Map_MappingSetRefs_rec
// C: void Map_MappingSetRefs_rec( Map_Man_t * pMan, Map_Node_t * pNode )
func (pMan *MapMan) MapMappingSetRefs_rec(pNode *MapNode) {
	panic("TODO, stubbed")
}

// Map_MappingSetRefs
// C: void Map_MappingSetRefs( Map_Man_t * pMan )
func (pMan *MapMan) MapMappingSetRefs() {
	panic("TODO, stubbed")
}

// Map_MappingGetArea
// C: float Map_MappingGetArea( Map_Man_t * pMan )
func (pMan *MapMan) MapMappingGetArea() float32 {
	panic("TODO, stubbed")
}
