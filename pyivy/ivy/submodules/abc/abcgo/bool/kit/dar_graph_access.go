package kit

// Kit_GraphRootEdge and Kit_NodeEdges expose the decomposition graph edges
// needed by the DAR refactoring evaluator without exposing mutable graph data.
func Kit_GraphRootEdge(g *Kit_Graph_t) Kit_Edge_t          { return g.eRoot }
func Kit_NodeEdges(n *Kit_Node_t) (Kit_Edge_t, Kit_Edge_t) { return n.eEdge0, n.eEdge1 }

// PrintingFlags returns the original operation and the two input phases
// preceding the De Morgan transformation used to store an OR as an AND.
func (n *Kit_Node_t) PrintingFlags() [3]uint32 {
	flags := uint32(n.printing)
	return [3]uint32{flags & 1, flags >> 1 & 1, flags >> 2 & 1}
}
