package abc

// Hand-checked skeleton of src/base/abci/abcReconv.c.

// AbcManCut is Abc_ManCut_t.
type AbcManCut struct {
	nNodeSizeMax int                              // the limit on the size of the supernode
	nConeSizeMax int                              // the limit on the size of the containing cone
	nNodeFanStop int                              // the limit on the size of the supernode
	nConeFanStop int                              // the limit on the size of the containing cone
	vNodeLeaves  *PointerVector[ABCNetworkObject] // fanins of the collapsed node (the cut)
	vConeLeaves  *PointerVector[ABCNetworkObject] // fanins of the containing cone
	vVisited     *PointerVector[ABCNetworkObject] // the visited nodes
	vLevels      *NestedVector[*ABCNetworkObject] // the data structure to compute TFO nodes
	vNodesTfo    *PointerVector[ABCNetworkObject] // the nodes in the TFO of the cut
}

// Abc_NodesMark (static)
// C: void Abc_NodesMark( Vec_Ptr_t * vVisited )
func abcNodesMark(vVisited *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NodesUnmark (static)
// C: void Abc_NodesUnmark( Vec_Ptr_t * vVisited )
func abcNodesUnmark(vVisited *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NodesUnmarkB (static)
// C: void Abc_NodesUnmarkB( Vec_Ptr_t * vVisited )
func abcNodesUnmarkB(vVisited *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NodeGetLeafCostOne (static)
// C: int Abc_NodeGetLeafCostOne( Abc_Obj_t * pNode, int nFaninLimit )
func abcNodeGetLeafCostOne(pNode *ABCNetworkObject, nFaninLimit int) int {
	panic("TODO, stubbed")
}

// Abc_NodeGetLeafCostTwo (static)
// C: int Abc_NodeGetLeafCostTwo( Abc_Obj_t * pNode, int nFaninLimit, Abc_Obj_t ** ppLeafToAdd, Abc_Obj_t ** pNodeToMark1, Abc_Obj_t ** pNodeToMark2 )
func abcNodeGetLeafCostTwo(pNode *ABCNetworkObject, nFaninLimit int, ppLeafToAdd **ABCNetworkObject, pNodeToMark1 **ABCNetworkObject, pNodeToMark2 **ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NodeFindCut
// C: Vec_Ptr_t * Abc_NodeFindCut( Abc_ManCut_t * p, Abc_Obj_t * pRoot, int fContain )
func (p *AbcManCut) AbcNodeFindCut(pRoot *ABCNetworkObject, fContain int) *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NodeBuildCutLevelOne_int (static)
// C: int Abc_NodeBuildCutLevelOne_int( Vec_Ptr_t * vVisited, Vec_Ptr_t * vLeaves, int nSizeLimit, int nFaninLimit )
func abcNodeBuildCutLevelOne_int(vVisited *PointerVector[ABCNetworkObject], vLeaves *PointerVector[ABCNetworkObject], nSizeLimit int, nFaninLimit int) int {
	panic("TODO, stubbed")
}

// Abc_NodeBuildCutLevelTwo_int (static)
// C: int Abc_NodeBuildCutLevelTwo_int( Vec_Ptr_t * vVisited, Vec_Ptr_t * vLeaves, int nFaninLimit )
func abcNodeBuildCutLevelTwo_int(vVisited *PointerVector[ABCNetworkObject], vLeaves *PointerVector[ABCNetworkObject], nFaninLimit int) int {
	panic("TODO, stubbed")
}

// Abc_NodeConeCollect
// C: void Abc_NodeConeCollect( Abc_Obj_t ** ppRoots, int nRoots, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vVisited, int fIncludeFanins )
func AbcNodeConeCollect(ppRoots []*ABCNetworkObject, nRoots int, vLeaves *PointerVector[ABCNetworkObject], vVisited *PointerVector[ABCNetworkObject], fIncludeFanins int) {
	panic("TODO, stubbed")
}

// Abc_NodeConeMarkCollect_rec (static)
// C: void Abc_NodeConeMarkCollect_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vVisited )
func abcNodeConeMarkCollect_rec(pNode *ABCNetworkObject, vVisited *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NodeConeBdd
// C: DdNode * Abc_NodeConeBdd( DdManager * dd, DdNode ** pbVars, Abc_Obj_t * pRoot, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vVisited )
func AbcNodeConeBdd(dd *CuddFoundation, pbVars []*DdNode, pRoot *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], vVisited *PointerVector[ABCNetworkObject]) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NodeConeDcs
// C: DdNode * Abc_NodeConeDcs( DdManager * dd, DdNode ** pbVarsX, DdNode ** pbVarsY, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vRoots, Vec_Ptr_t * vVisited )
func AbcNodeConeDcs(dd *CuddFoundation, pbVarsX []*DdNode, pbVarsY []*DdNode, vLeaves *PointerVector[ABCNetworkObject], vRoots *PointerVector[ABCNetworkObject], vVisited *PointerVector[ABCNetworkObject]) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkManCutStart
// C: Abc_ManCut_t * Abc_NtkManCutStart( int nNodeSizeMax, int nConeSizeMax, int nNodeFanStop, int nConeFanStop )
func AbcNtkManCutStart(nNodeSizeMax int, nConeSizeMax int, nNodeFanStop int, nConeFanStop int) *AbcManCut {
	panic("TODO, stubbed")
}

// Abc_NtkManCutStop
// C: void Abc_NtkManCutStop( Abc_ManCut_t * p )
func (p *AbcManCut) AbcNtkManCutStop() {
	panic("TODO, stubbed")
}

// Abc_NtkManCutReadCutLarge
// C: Vec_Ptr_t * Abc_NtkManCutReadCutLarge( Abc_ManCut_t * p )
func (p *AbcManCut) AbcNtkManCutReadCutLarge() *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkManCutReadCutSmall
// C: Vec_Ptr_t * Abc_NtkManCutReadCutSmall( Abc_ManCut_t * p )
func (p *AbcManCut) AbcNtkManCutReadCutSmall() *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkManCutReadVisited
// C: Vec_Ptr_t * Abc_NtkManCutReadVisited( Abc_ManCut_t * p )
func (p *AbcManCut) AbcNtkManCutReadVisited() *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NodeCollectTfoCands
// C: Vec_Ptr_t * Abc_NodeCollectTfoCands( Abc_ManCut_t * p, Abc_Obj_t * pRoot, Vec_Ptr_t * vLeaves, int LevelMax )
func (p *AbcManCut) AbcNodeCollectTfoCands(pRoot *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], LevelMax int) *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}
