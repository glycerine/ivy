package aig

// ObjectData returns the auxiliary payload associated with a graph node.
func (g *Graph) ObjectData(e Edge) any {
	if !g.valid(e) {
		panic("aig: invalid object")
	}
	return g.objectData[e.NodeID()]
}

// SetObjectData associates an auxiliary payload with a graph node.
func (g *Graph) SetObjectData(e Edge, value any) {
	if !g.valid(e) {
		panic("aig: invalid object")
	}
	if value == nil {
		delete(g.objectData, e.NodeID())
		return
	}
	if g.objectData == nil {
		g.objectData = make(map[int]any)
	}
	g.objectData[e.NodeID()] = value
}
