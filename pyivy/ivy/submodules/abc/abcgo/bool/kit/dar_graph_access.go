package kit

// Kit_GraphRootEdge and Kit_NodeEdges expose the decomposition graph edges
// needed by the DAR refactoring evaluator without exposing mutable graph data.
func Kit_GraphRootEdge(g *Kit_Graph_t) Kit_Edge_t          { return g.eRoot }
func Kit_NodeEdges(n *Kit_Node_t) (Kit_Edge_t, Kit_Edge_t) { return n.eEdge0, n.eEdge1 }
