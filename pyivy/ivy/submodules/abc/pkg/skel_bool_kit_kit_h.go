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
// C: unsigned Kit_DsdObjOffset( int nFans )
func kitDsdObjOffset(nFans int) uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdObjTruth (static inline; returns a pointer into the object, or NULL)
// C: unsigned * Kit_DsdObjTruth( Kit_DsdObj_t * pObj )
func kitDsdObjTruth(pObj *KitDsdObj) []uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdNtkObjNum (static inline)
// C: int Kit_DsdNtkObjNum( Kit_DsdNtk_t * pNtk )
func kitDsdNtkObjNum(pNtk *KitDsdNtk) int {
	panic("TODO, stubbed")
}

// Kit_DsdNtkObj (static inline)
// C: Kit_DsdObj_t * Kit_DsdNtkObj( Kit_DsdNtk_t * pNtk, int Id )
func kitDsdNtkObj(pNtk *KitDsdNtk, Id int) *KitDsdObj {
	panic("TODO, stubbed")
}

// Kit_DsdNtkRoot (static inline)
// C: Kit_DsdObj_t * Kit_DsdNtkRoot( Kit_DsdNtk_t * pNtk )
func kitDsdNtkRoot(pNtk *KitDsdNtk) *KitDsdObj {
	panic("TODO, stubbed")
}

// Kit_DsdLitIsLeaf (static inline)
// C: int Kit_DsdLitIsLeaf( Kit_DsdNtk_t * pNtk, int Lit )
func kitDsdLitIsLeaf(pNtk *KitDsdNtk, Lit int) int {
	panic("TODO, stubbed")
}

// Kit_DsdLitSupport (static inline)
// C: unsigned Kit_DsdLitSupport( Kit_DsdNtk_t * pNtk, int Lit )
func kitDsdLitSupport(pNtk *KitDsdNtk, Lit int) uint32 {
	panic("TODO, stubbed")
}

// Kit_CubeHasLit (static inline)
// C: int Kit_CubeHasLit( unsigned uCube, int i )
func kitCubeHasLit(uCube uint32, i int) int {
	panic("TODO, stubbed")
}

// Kit_CubeSetLit (static inline)
// C: unsigned Kit_CubeSetLit( unsigned uCube, int i )
func kitCubeSetLit(uCube uint32, i int) uint32 {
	panic("TODO, stubbed")
}

// Kit_CubeXorLit (static inline)
// C: unsigned Kit_CubeXorLit( unsigned uCube, int i )
func kitCubeXorLit(uCube uint32, i int) uint32 {
	panic("TODO, stubbed")
}

// Kit_CubeRemLit (static inline)
// C: unsigned Kit_CubeRemLit( unsigned uCube, int i )
func kitCubeRemLit(uCube uint32, i int) uint32 {
	panic("TODO, stubbed")
}

// Kit_CubeContains (static inline)
// C: int Kit_CubeContains( unsigned uLarge, unsigned uSmall )
func kitCubeContains(uLarge uint32, uSmall uint32) int {
	panic("TODO, stubbed")
}

// Kit_CubeSharp (static inline)
// C: unsigned Kit_CubeSharp( unsigned uCube, unsigned uMask )
func kitCubeSharp(uCube uint32, uMask uint32) uint32 {
	panic("TODO, stubbed")
}

// Kit_CubeMask (static inline)
// C: unsigned Kit_CubeMask( int nVar )
func kitCubeMask(nVar int) uint32 {
	panic("TODO, stubbed")
}

// Kit_CubeIsMarked (static inline)
// C: int Kit_CubeIsMarked( unsigned uCube )
func kitCubeIsMarked(uCube uint32) int {
	panic("TODO, stubbed")
}

// Kit_CubeMark (static inline)
// C: unsigned Kit_CubeMark( unsigned uCube )
func kitCubeMark(uCube uint32) uint32 {
	panic("TODO, stubbed")
}

// Kit_CubeUnmark (static inline)
// C: unsigned Kit_CubeUnmark( unsigned uCube )
func kitCubeUnmark(uCube uint32) uint32 {
	panic("TODO, stubbed")
}

// Kit_SopCubeNum (static inline)
// C: int Kit_SopCubeNum( Kit_Sop_t * cSop )
func kitSopCubeNum(cSop *KitSop) int {
	panic("TODO, stubbed")
}

// Kit_SopCube (static inline)
// C: unsigned Kit_SopCube( Kit_Sop_t * cSop, int i )
func kitSopCube(cSop *KitSop, i int) uint32 {
	panic("TODO, stubbed")
}

// Kit_SopShrink (static inline)
// C: void Kit_SopShrink( Kit_Sop_t * cSop, int nCubesNew )
func kitSopShrink(cSop *KitSop, nCubesNew int) {
	panic("TODO, stubbed")
}

// Kit_SopPushCube (static inline)
// C: void Kit_SopPushCube( Kit_Sop_t * cSop, unsigned uCube )
func kitSopPushCube(cSop *KitSop, uCube uint32) {
	panic("TODO, stubbed")
}

// Kit_SopWriteCube (static inline)
// C: void Kit_SopWriteCube( Kit_Sop_t * cSop, unsigned uCube, int i )
func kitSopWriteCube(cSop *KitSop, uCube uint32, i int) {
	panic("TODO, stubbed")
}

// Kit_EdgeCreate (static inline)
// C: Kit_Edge_t Kit_EdgeCreate( int Node, int fCompl )
func kitEdgeCreate(Node int, fCompl int) KitEdge {
	panic("TODO, stubbed")
}

// Kit_EdgeToInt (static inline)
// C: unsigned Kit_EdgeToInt( Kit_Edge_t eEdge )
func kitEdgeToInt(eEdge KitEdge) uint32 {
	panic("TODO, stubbed")
}

// Kit_IntToEdge (static inline)
// C: Kit_Edge_t Kit_IntToEdge( unsigned Edge )
func kitIntToEdge(Edge uint32) KitEdge {
	panic("TODO, stubbed")
}

// Kit_EdgeToInt_ (static inline; reinterprets the edge through a union)
// C: unsigned Kit_EdgeToInt_( Kit_Edge_t m )
func kitEdgeToInt_(m KitEdge) uint32 {
	panic("TODO, stubbed")
}

// Kit_IntToEdge_ (static inline; reinterprets the integer through a union)
// C: Kit_Edge_t Kit_IntToEdge_( unsigned m )
func kitIntToEdge_(m uint32) KitEdge {
	panic("TODO, stubbed")
}

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
// C: int Kit_SuppIsMinBase( int Supp )
func kitSuppIsMinBase(Supp int) int {
	panic("TODO, stubbed")
}

// Kit_BitWordNum (static inline)
// C: int Kit_BitWordNum( int nBits )
func kitBitWordNum(nBits int) int {
	panic("TODO, stubbed")
}

// Kit_TruthWordNum (static inline)
// C: int Kit_TruthWordNum( int nVars )
func kitTruthWordNum(nVars int) int {
	panic("TODO, stubbed")
}

// Kit_BitMask (static inline)
// C: unsigned Kit_BitMask( int nBits )
func kitBitMask(nBits int) uint32 {
	panic("TODO, stubbed")
}

// Kit_TruthSetBit (static inline)
// C: void Kit_TruthSetBit( unsigned * p, int Bit )
func kitTruthSetBit(p []uint32, Bit int) {
	panic("TODO, stubbed")
}

// Kit_TruthXorBit (static inline)
// C: void Kit_TruthXorBit( unsigned * p, int Bit )
func kitTruthXorBit(p []uint32, Bit int) {
	panic("TODO, stubbed")
}

// Kit_TruthHasBit (static inline)
// C: int Kit_TruthHasBit( unsigned * p, int Bit )
func kitTruthHasBit(p []uint32, Bit int) int {
	panic("TODO, stubbed")
}

// Kit_WordFindFirstBit (static inline)
// C: int Kit_WordFindFirstBit( unsigned uWord )
func kitWordFindFirstBit(uWord uint32) int {
	panic("TODO, stubbed")
}

// Kit_WordHasOneBit (static inline)
// C: int Kit_WordHasOneBit( unsigned uWord )
func kitWordHasOneBit(uWord uint32) int {
	panic("TODO, stubbed")
}

// Kit_WordCountOnes (static inline)
// C: int Kit_WordCountOnes( unsigned uWord )
func kitWordCountOnes(uWord uint32) int {
	panic("TODO, stubbed")
}

// Kit_TruthCountOnes (static inline)
// C: int Kit_TruthCountOnes( unsigned * pIn, int nVars )
func kitTruthCountOnes(pIn []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthFindFirstBit (static inline)
// C: int Kit_TruthFindFirstBit( unsigned * pIn, int nVars )
func kitTruthFindFirstBit(pIn []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthFindFirstZero (static inline)
// C: int Kit_TruthFindFirstZero( unsigned * pIn, int nVars )
func kitTruthFindFirstZero(pIn []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthIsEqual (static inline)
// C: int Kit_TruthIsEqual( unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthIsEqual(pIn0 []uint32, pIn1 []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthIsEqualWithCare (static inline)
// C: int Kit_TruthIsEqualWithCare( unsigned * pIn0, unsigned * pIn1, unsigned * pCare, int nVars )
func kitTruthIsEqualWithCare(pIn0 []uint32, pIn1 []uint32, pCare []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthIsOpposite (static inline)
// C: int Kit_TruthIsOpposite( unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthIsOpposite(pIn0 []uint32, pIn1 []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthIsEqualWithPhase (static inline)
// C: int Kit_TruthIsEqualWithPhase( unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthIsEqualWithPhase(pIn0 []uint32, pIn1 []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthIsConst0 (static inline)
// C: int Kit_TruthIsConst0( unsigned * pIn, int nVars )
func kitTruthIsConst0(pIn []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthIsConst1 (static inline)
// C: int Kit_TruthIsConst1( unsigned * pIn, int nVars )
func kitTruthIsConst1(pIn []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthIsImply (static inline)
// C: int Kit_TruthIsImply( unsigned * pIn1, unsigned * pIn2, int nVars )
func kitTruthIsImply(pIn1 []uint32, pIn2 []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthIsDisjoint (static inline)
// C: int Kit_TruthIsDisjoint( unsigned * pIn1, unsigned * pIn2, int nVars )
func kitTruthIsDisjoint(pIn1 []uint32, pIn2 []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthIsDisjoint3 (static inline)
// C: int Kit_TruthIsDisjoint3( unsigned * pIn1, unsigned * pIn2, unsigned * pIn3, int nVars )
func kitTruthIsDisjoint3(pIn1 []uint32, pIn2 []uint32, pIn3 []uint32, nVars int) int {
	panic("TODO, stubbed")
}

// Kit_TruthCopy (static inline)
// C: void Kit_TruthCopy( unsigned * pOut, unsigned * pIn, int nVars )
func kitTruthCopy(pOut []uint32, pIn []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_TruthClear (static inline)
// C: void Kit_TruthClear( unsigned * pOut, int nVars )
func kitTruthClear(pOut []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_TruthFill (static inline)
// C: void Kit_TruthFill( unsigned * pOut, int nVars )
func kitTruthFill(pOut []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_TruthNot (static inline)
// C: void Kit_TruthNot( unsigned * pOut, unsigned * pIn, int nVars )
func kitTruthNot(pOut []uint32, pIn []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_TruthAnd (static inline)
// C: void Kit_TruthAnd( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthAnd(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_TruthOr (static inline)
// C: void Kit_TruthOr( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthOr(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_TruthXor (static inline)
// C: void Kit_TruthXor( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthXor(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_TruthSharp (static inline)
// C: void Kit_TruthSharp( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthSharp(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_TruthNand (static inline)
// C: void Kit_TruthNand( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthNand(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_TruthAndPhase (static inline)
// C: void Kit_TruthAndPhase( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars, int fCompl0, int fCompl1 )
func kitTruthAndPhase(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int, fCompl0 int, fCompl1 int) {
	panic("TODO, stubbed")
}

// Kit_TruthOrPhase (static inline)
// C: void Kit_TruthOrPhase( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars, int fCompl0, int fCompl1 )
func kitTruthOrPhase(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int, fCompl0 int, fCompl1 int) {
	panic("TODO, stubbed")
}

// Kit_TruthMux (static inline)
// C: void Kit_TruthMux( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, unsigned * pCtrl, int nVars )
func kitTruthMux(pOut []uint32, pIn0 []uint32, pIn1 []uint32, pCtrl []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_TruthMuxPhase (static inline)
// C: void Kit_TruthMuxPhase( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, unsigned * pCtrl, int nVars, int fComp0 )
func kitTruthMuxPhase(pOut []uint32, pIn0 []uint32, pIn1 []uint32, pCtrl []uint32, nVars int, fComp0 int) {
	panic("TODO, stubbed")
}

// Kit_TruthIthVar (static inline)
// C: void Kit_TruthIthVar( unsigned * pTruth, int nVars, int iVar )
func kitTruthIthVar(pTruth []uint32, nVars int, iVar int) {
	panic("TODO, stubbed")
}
