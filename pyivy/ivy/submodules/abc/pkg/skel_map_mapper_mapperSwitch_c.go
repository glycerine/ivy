package abc

// Hand-checked skeleton of src/map/mapper/mapperSwitch.c.

// Map_SwitchCutGetDerefed
// C: float Map_SwitchCutGetDerefed( Map_Node_t * pNode, Map_Cut_t * pCut, int fPhase )
func (pNode *MapNode) MapSwitchCutGetDerefed(pCut *MapCut, fPhase int) float32 {
	panic("TODO, stubbed")
}

// Map_SwitchCutRef
// C: float Map_SwitchCutRef( Map_Node_t * pNode, Map_Cut_t * pCut, int fPhase )
func (pNode *MapNode) MapSwitchCutRef(pCut *MapCut, fPhase int) float32 {
	panic("TODO, stubbed")
}

// Map_SwitchCutDeref
// C: float Map_SwitchCutDeref( Map_Node_t * pNode, Map_Cut_t * pCut, int fPhase )
func (pNode *MapNode) MapSwitchCutDeref(pCut *MapCut, fPhase int) float32 {
	panic("TODO, stubbed")
}

// Map_SwitchCutRefDeref (static)
// C: float Map_SwitchCutRefDeref( Map_Node_t * pNode, Map_Cut_t * pCut, int fPhase, int fReference )
func (pNode *MapNode) mapSwitchCutRefDeref(pCut *MapCut, fPhase int, fReference int) float32 {
	panic("TODO, stubbed")
}

// Map_MappingGetSwitching
// C: float Map_MappingGetSwitching( Map_Man_t * pMan )
func (pMan *MapMan) MapMappingGetSwitching() float32 {
	panic("TODO, stubbed")
}
