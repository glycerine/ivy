package abc

// Hand-checked skeleton of src/bool/kit/kit.h.

// The iterator and cube-walking macros (Kit_DsdNtkForEachObj,
// Kit_DsdObjForEachFanin, Kit_PlaForEachCube, ...) and KIT_MIN/KIT_MAX/
// KIT_INFINITY are macros, not functions, and are not represented. Go has no
// bitfields; each is a uint32 whose C width is in the comment.

// KitSop is Kit_Sop_t.
type KitSop struct {
	nLits  int      // the number of literals
	nCubes int      // the number of cubes
	pCubes []uint32 // unsigned * : the storage for cubes
}

// Kit_Edge_t is the validated port KitEdge and Kit_Graph_t is the validated
// port KitGraph (bool_kit_graph.go); neither is redeclared here.

// KitNode is Kit_Node_t. The anonymous union { int iFunc; void * pFunc; } is
// represented by both members; only one is meaningful at a time.
type KitNode struct {
	eEdge0 KitEdge // the left child of the node
	eEdge1 KitEdge // the right child of the node
	// other info
	iFunc int    // union member : the function of the node (BDD or AIG)
	pFunc any    // union member (void *) : the function of the node (BDD or AIG)
	Level uint32 // unsigned : 14, the level of this node in the global AIG
	// printing info
	fNodeOr uint32 // unsigned : 1, marks the original OR node
	fCompl0 uint32 // unsigned : 1, marks the original complemented edge
	fCompl1 uint32 // unsigned : 1, marks the original complemented edge
	// latch info
	nLat0 uint32 // unsigned : 5, the number of latches on the first edge
	nLat1 uint32 // unsigned : 5, the number of latches on the second edge
	nLat2 uint32 // unsigned : 5, the number of latches on the output edge
}

// KitDsd is the Kit_Dsd_t enum: DSD node types.
type KitDsd int

const (
	KIT_DSD_NONE   KitDsd = 0 // 0: unknown
	KIT_DSD_CONST1        = 1 // 1: constant 1
	KIT_DSD_VAR           = 2 // 2: elementary variable
	KIT_DSD_AND           = 3 // 3: multi-input AND
	KIT_DSD_XOR           = 4 // 4: multi-input XOR
	KIT_DSD_PRIME         = 5 // 5: arbitrary function of 3+ variables
)

// KitDsdObj is Kit_DsdObj_t: a DSD node.
type KitDsdObj struct {
	Id     uint32   // unsigned : 6, the number of this node
	Type   uint32   // unsigned : 3, none, const, var, AND, XOR, MUX, PRIME
	fMark  uint32   // unsigned : 1, finished checking output
	Offset uint32   // unsigned : 8, offset to the truth table
	nRefs  uint32   // unsigned : 8, offset to the truth table
	nFans  uint32   // unsigned : 6, the number of fanins of this node
	pFans  []uint16 // unsigned short pFans[0] : flexible array member, the fanin literals
}

// KitDsdNtk is Kit_DsdNtk_t: a DSD network.
type KitDsdNtk struct {
	nVars       uint16       // at most 16 (perhaps 18?)
	nNodesAlloc uint16       // the number of allocated nodes (at most nVars)
	nNodes      uint16       // the number of nodes
	Root        uint16       // the root of the tree
	pMem        []uint32     // unsigned * : memory for the truth tables (memory manager?)
	pSupps      []uint32     // unsigned * : supports of the nodes
	pNodes      []*KitDsdObj // Kit_DsdObj_t ** : the nodes
}

// KitDsdMan is Kit_DsdMan_t: a DSD manager.
type KitDsdMan struct {
	nVars    int                    // the maximum number of variables
	nWords   int                    // the number of words in TTs
	vTtElems *PointerVector[uint32] // elementary truth tables
	vTtNodes *PointerVector[uint32] // the node truth tables
	// BDD representation
	dd      *CloudManager          // BDD package
	vTtBdds *PointerVector[uint32] // the node truth tables
	vNodes  *IntVector             // temporary array for BDD nodes
}

// Kit_DsdObjOffset (static inline)
// Kit_DsdObjTruth (static inline; returns a pointer into the object, or NULL)
// C: unsigned * Kit_DsdObjTruth( Kit_DsdObj_t * pObj )
func kitDsdObjTruth(pObj *KitDsdObj) []uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdNtkObjNum (static inline)
// Kit_DsdNtkObj (static inline)
// Kit_DsdNtkRoot (static inline)
// Kit_DsdLitIsLeaf (static inline)
// Kit_DsdLitSupport (static inline)
// Kit_CubeHasLit (static inline)
// Kit_CubeSetLit (static inline)
// Kit_CubeXorLit (static inline)
// Kit_CubeRemLit (static inline)
// Kit_CubeContains (static inline)
// Kit_CubeSharp (static inline)
// Kit_CubeMask (static inline)
// Kit_CubeIsMarked (static inline)
// Kit_CubeMark (static inline)
// Kit_CubeUnmark (static inline)
// Kit_SopCubeNum (static inline)
// Kit_SopCube (static inline)
// Kit_SopShrink (static inline)
// Kit_SopPushCube (static inline)
// Kit_SopWriteCube (static inline)
// Kit_EdgeCreate (static inline)
// Kit_EdgeToInt (static inline)
// Kit_IntToEdge (static inline)
// Kit_EdgeToInt_ (static inline; reinterprets the edge through a union)
// Kit_IntToEdge_ (static inline; reinterprets the integer through a union)
// Kit_GraphIsConst (static inline)
// C: int Kit_GraphIsConst( Kit_Graph_t * pGraph )
func kitGraphIsConst(pGraph *KitGraph) int {
	panic("TODO, stubbed")
}

// Kit_GraphIsConst0 (static inline)
// C: int Kit_GraphIsConst0( Kit_Graph_t * pGraph )
func kitGraphIsConst0(pGraph *KitGraph) int {
	panic("TODO, stubbed")
}

// Kit_GraphIsConst1 (static inline)
// C: int Kit_GraphIsConst1( Kit_Graph_t * pGraph )
func kitGraphIsConst1(pGraph *KitGraph) int {
	panic("TODO, stubbed")
}

// Kit_GraphIsComplement (static inline)
// C: int Kit_GraphIsComplement( Kit_Graph_t * pGraph )
func kitGraphIsComplement(pGraph *KitGraph) int {
	panic("TODO, stubbed")
}

// Kit_GraphIsVar (static inline)
// C: int Kit_GraphIsVar( Kit_Graph_t * pGraph )
func kitGraphIsVar(pGraph *KitGraph) int {
	panic("TODO, stubbed")
}

// Kit_GraphComplement (static inline)
// C: void Kit_GraphComplement( Kit_Graph_t * pGraph )
func kitGraphComplement(pGraph *KitGraph) {
	panic("TODO, stubbed")
}

// Kit_GraphSetRoot (static inline)
// C: void Kit_GraphSetRoot( Kit_Graph_t * pGraph, Kit_Edge_t eRoot )
func kitGraphSetRoot(pGraph *KitGraph, eRoot KitEdge) {
	panic("TODO, stubbed")
}

// Kit_GraphLeaveNum (static inline)
// C: int Kit_GraphLeaveNum( Kit_Graph_t * pGraph )
func kitGraphLeaveNum(pGraph *KitGraph) int {
	panic("TODO, stubbed")
}

// Kit_GraphNodeNum (static inline)
// C: int Kit_GraphNodeNum( Kit_Graph_t * pGraph )
func kitGraphNodeNum(pGraph *KitGraph) int {
	panic("TODO, stubbed")
}

// Kit_GraphNode (static inline)
// C: Kit_Node_t * Kit_GraphNode( Kit_Graph_t * pGraph, int i )
func kitGraphNode_inline(pGraph *KitGraph, i int) *KitNode {
	panic("TODO, stubbed")
}

// Kit_GraphNodeLast (static inline)
// C: Kit_Node_t * Kit_GraphNodeLast( Kit_Graph_t * pGraph )
func kitGraphNodeLast(pGraph *KitGraph) *KitNode {
	panic("TODO, stubbed")
}

// Kit_GraphNodeInt (static inline)
// C: int Kit_GraphNodeInt( Kit_Graph_t * pGraph, Kit_Node_t * pNode )
func kitGraphNodeInt(pGraph *KitGraph, pNode *KitNode) int {
	panic("TODO, stubbed")
}

// Kit_GraphNodeIsVar (static inline)
// C: int Kit_GraphNodeIsVar( Kit_Graph_t * pGraph, Kit_Node_t * pNode )
func kitGraphNodeIsVar(pGraph *KitGraph, pNode *KitNode) int {
	panic("TODO, stubbed")
}

// Kit_GraphVar (static inline)
// C: Kit_Node_t * Kit_GraphVar( Kit_Graph_t * pGraph )
func kitGraphVar(pGraph *KitGraph) *KitNode {
	panic("TODO, stubbed")
}

// Kit_GraphVarInt (static inline)
// C: int Kit_GraphVarInt( Kit_Graph_t * pGraph )
func kitGraphVarInt(pGraph *KitGraph) int {
	panic("TODO, stubbed")
}

// Kit_GraphNodeFanin0 (static inline)
// C: Kit_Node_t * Kit_GraphNodeFanin0( Kit_Graph_t * pGraph, Kit_Node_t * pNode )
func kitGraphNodeFanin0(pGraph *KitGraph, pNode *KitNode) *KitNode {
	panic("TODO, stubbed")
}

// Kit_GraphNodeFanin1 (static inline)
// C: Kit_Node_t * Kit_GraphNodeFanin1( Kit_Graph_t * pGraph, Kit_Node_t * pNode )
func kitGraphNodeFanin1(pGraph *KitGraph, pNode *KitNode) *KitNode {
	panic("TODO, stubbed")
}

// Kit_GraphRootLevel (static inline)
// C: int Kit_GraphRootLevel( Kit_Graph_t * pGraph )
func kitGraphRootLevel(pGraph *KitGraph) int {
	panic("TODO, stubbed")
}

// Kit_SuppIsMinBase (static inline)
// Kit_BitWordNum (static inline)
// Kit_TruthWordNum (static inline)
// Kit_BitMask (static inline)
// Kit_TruthSetBit (static inline)
// Kit_TruthXorBit (static inline)
// Kit_TruthHasBit (static inline)
// Kit_WordFindFirstBit (static inline)
// Kit_WordHasOneBit (static inline)
// Kit_WordCountOnes (static inline)
// Kit_TruthCountOnes (static inline)
// Kit_TruthFindFirstBit (static inline)
// Kit_TruthFindFirstZero (static inline)
// Kit_TruthIsEqual (static inline)
// Kit_TruthIsEqualWithCare (static inline)
// Kit_TruthIsOpposite (static inline)
// Kit_TruthIsEqualWithPhase (static inline)
// Kit_TruthIsConst0 (static inline)
// Kit_TruthIsConst1 (static inline)
// Kit_TruthIsImply (static inline)
// Kit_TruthIsDisjoint (static inline)
// Kit_TruthIsDisjoint3 (static inline)
// Kit_TruthCopy (static inline)
// Kit_TruthClear (static inline)
// Kit_TruthFill (static inline)
// Kit_TruthNot (static inline)
// Kit_TruthAnd (static inline)
// Kit_TruthOr (static inline)
// Kit_TruthXor (static inline)
// Kit_TruthSharp (static inline)
// Kit_TruthNand (static inline)
// Kit_TruthAndPhase (static inline)
// Kit_TruthOrPhase (static inline)
// Kit_TruthMux (static inline)
// Kit_TruthMuxPhase (static inline)
// Kit_TruthIthVar (static inline)
