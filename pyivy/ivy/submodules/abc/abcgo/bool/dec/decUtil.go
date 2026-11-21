package dec

// Dec_GraphDeriveTruth evaluates all 32 assignments together and preserves
// caller-owned node payloads.
func Dec_GraphDeriveTruth(g *Dec_Graph_t) uint32 {
	root := Edge{int(g.eRoot.Node), g.eRoot.fCompl != 0}
	return graphTruth32(g, int(g.size), int(g.nLeaves), root, g.fConst != 0)
}

func (g *Dec_Graph_t) Native() *Graph {
	native := &Graph{LeafCount: int(g.nLeaves), Constant: g.fConst != 0, Root: Edge{int(g.eRoot.Node), g.eRoot.fCompl != 0}, Nodes: make([]Node, int(g.size))}
	for i := range g.size {
		n := g.node(uint32(i))
		native.Nodes[i] = Node{Edge{int(n.eEdge0.Node), n.eEdge0.fCompl != 0}, Edge{int(n.eEdge1.Node), n.eEdge1.fCompl != 0}}
	}
	return native
}
