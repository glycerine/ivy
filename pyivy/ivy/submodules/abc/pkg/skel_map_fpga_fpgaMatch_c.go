package abc

// Hand-checked skeleton of src/map/fpga/fpgaMatch.c.

// Fpga_MappingMatches
// C: int Fpga_MappingMatches( Fpga_Man_t * p, int fDelayOriented )
func (p *FpgaMan) FpgaMappingMatches(fDelayOriented int) int {
	panic("TODO, stubbed")
}

// Fpga_MatchNode (static)
// C: int Fpga_MatchNode( Fpga_Man_t * p, Fpga_Node_t * pNode, int fDelayOriented )
func (p *FpgaMan) fpgaMatchNode(pNode *FpgaNode, fDelayOriented int) int {
	panic("TODO, stubbed")
}

// Fpga_MappingMatchesArea
// C: int Fpga_MappingMatchesArea( Fpga_Man_t * p )
func (p *FpgaMan) FpgaMappingMatchesArea() int {
	panic("TODO, stubbed")
}

// Fpga_MappingMatchesAreaArray (static)
// C: int Fpga_MappingMatchesAreaArray( Fpga_Man_t * p, Fpga_NodeVec_t * vNodes )
func (p *FpgaMan) fpgaMappingMatchesAreaArray(vNodes *FpgaNodeVec) int {
	panic("TODO, stubbed")
}

// Fpga_MatchNodeArea (static)
// C: int Fpga_MatchNodeArea( Fpga_Man_t * p, Fpga_Node_t * pNode )
func (p *FpgaMan) fpgaMatchNodeArea(pNode *FpgaNode) int {
	panic("TODO, stubbed")
}

// Fpga_MappingMatchesSwitch
// C: int Fpga_MappingMatchesSwitch( Fpga_Man_t * p )
func (p *FpgaMan) FpgaMappingMatchesSwitch() int {
	panic("TODO, stubbed")
}

// Fpga_MatchNodeSwitch (static)
// C: int Fpga_MatchNodeSwitch( Fpga_Man_t * p, Fpga_Node_t * pNode )
func (p *FpgaMan) fpgaMatchNodeSwitch(pNode *FpgaNode) int {
	panic("TODO, stubbed")
}

// Fpga_FindBestNode (ppNode and ppCutBest are out parameters)
// C: float Fpga_FindBestNode( Fpga_Man_t * p, Fpga_NodeVec_t * vNodes, Fpga_Node_t ** ppNode, Fpga_Cut_t ** ppCutBest )
func (p *FpgaMan) FpgaFindBestNode(vNodes *FpgaNodeVec, ppNode **FpgaNode, ppCutBest **FpgaCut) float32 {
	panic("TODO, stubbed")
}
