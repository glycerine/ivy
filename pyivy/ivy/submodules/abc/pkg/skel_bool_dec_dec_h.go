package abc

// Hand-checked skeleton of src/bool/dec/dec.h.

// DecEdge is Dec_Edge_t.
type DecEdge struct {
	fCompl uint32 // unsigned : 1 : the complemented bit
	Node   uint32 // unsigned : 30 : the decomposition node pointed by the edge
}

// DecNode is Dec_Node_t.
type DecNode struct {
	eEdge0 DecEdge // the left child of the node
	eEdge1 DecEdge // the right child of the node
	// other info
	// union { int iFunc; void * pFunc; }
	iFunc int    // the literal of the node (AIG)
	pFunc any    // void * : the function of the node (BDD or AIG)
	Level uint32 // unsigned : 14 : the level of this node in the global AIG
	// printing info
	fNodeOr uint32 // unsigned : 1 : marks the original OR node
	fCompl0 uint32 // unsigned : 1 : marks the original complemented edge
	fCompl1 uint32 // unsigned : 1 : marks the original complemented edge
	// latch info
	nLat0 uint32 // unsigned : 5 : the number of latches on the first edge
	nLat1 uint32 // unsigned : 5 : the number of latches on the second edge
	nLat2 uint32 // unsigned : 5 : the number of latches on the output edge
}

// DecGraph is Dec_Graph_t.
type DecGraph struct {
	fConst  int       // marks the constant 1 graph
	nLeaves int       // the number of leaves
	nSize   int       // the number of nodes (including the leaves)
	nCap    int       // the number of allocated nodes
	pNodes  []DecNode // Dec_Node_t * : the array of leaves and internal nodes
	eRoot   DecEdge   // the pointer to the topmost node
}
