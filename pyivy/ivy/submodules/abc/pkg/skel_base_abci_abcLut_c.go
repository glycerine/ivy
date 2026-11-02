package abc

// Hand-checked skeleton of src/base/abci/abcLut.c.

const SCL_VARS_MAX = 15 // the maximum number of variables

// AbcManScl is Abc_ManScl_t.
type AbcManScl struct {
	nLutSize    int                              // the LUT size
	nCutSizeMax int                              // the max number of leaves of the cone
	nNodesMax   int                              // the max number of divisors in the cone
	nWords      int                              // the number of machine words in sim info
	vLeaves     *PointerVector[ABCNetworkObject] // leaves of the cut
	vVolume     *PointerVector[ABCNetworkObject] // volume of the cut
	pBSet       [SCL_VARS_MAX]int                // bound set
	uTruth      []uint32                         // unsigned * : truth table of the cone
	uVars       [][]uint32                       // unsigned ** : elementary truth tables
	uSims       [][]uint32                       // unsigned ** : truth tables of the nodes
	uCofs       [][]uint32                       // unsigned ** : truth tables of the cofactors
}

// Abc_NtkSuperChoiceLut
// C: int Abc_NtkSuperChoiceLut( Abc_Ntk_t * pNtk, int nLutSize, int nCutSizeMax, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkSuperChoiceLut(nLutSize int, nCutSizeMax int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NodeLutMap (static)
// C: void Abc_NodeLutMap( Cut_Man_t * pManCuts, Abc_Obj_t * pObj )
func abcNodeLutMap(pManCuts *CutMan, pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NtkStartCutManForScl (static)
// C: Cut_Man_t * Abc_NtkStartCutManForScl( Abc_Ntk_t * pNtk, int nLutSize )
func (pNtk *ABCNetwork) abcNtkStartCutManForScl(nLutSize int) *CutMan {
	panic("TODO, stubbed")
}

// Abc_ManSclStart (static)
// C: Abc_ManScl_t * Abc_ManSclStart( int nLutSize, int nCutSizeMax, int nNodesMax )
func abcManSclStart(nLutSize int, nCutSizeMax int, nNodesMax int) *AbcManScl {
	panic("TODO, stubbed")
}

// Abc_ManSclStop (static)
// C: void Abc_ManSclStop( Abc_ManScl_t * p )
func (p *AbcManScl) abcManSclStop() {
	panic("TODO, stubbed")
}

// Abc_NodeSuperChoiceTruth
// C: unsigned * Abc_NodeSuperChoiceTruth( Abc_ManScl_t * pManScl )
func (pManScl *AbcManScl) AbcNodeSuperChoiceTruth() []uint32 {
	panic("TODO, stubbed")
}

// Abc_NodeSuperChoiceCollect2_rec
// C: void Abc_NodeSuperChoiceCollect2_rec( Abc_Obj_t * pObj, Vec_Ptr_t * vVolume )
func AbcNodeSuperChoiceCollect2_rec(pObj *ABCNetworkObject, vVolume *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NodeSuperChoiceCollect2
// C: void Abc_NodeSuperChoiceCollect2( Abc_Obj_t * pRoot, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vVolume )
func AbcNodeSuperChoiceCollect2(pRoot *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], vVolume *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NodeSuperChoiceCollect_rec
// C: void Abc_NodeSuperChoiceCollect_rec( Abc_Obj_t * pObj, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vVolume )
func AbcNodeSuperChoiceCollect_rec(pObj *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], vVolume *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NodeSuperChoiceCollect
// C: void Abc_NodeSuperChoiceCollect( Abc_Obj_t * pRoot, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vVolume )
func AbcNodeSuperChoiceCollect(pRoot *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], vVolume *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NodeLeavesRemove
// C: void Abc_NodeLeavesRemove( Vec_Ptr_t * vLeaves, unsigned uPhase, int nVars )
func AbcNodeLeavesRemove(vLeaves *PointerVector[ABCNetworkObject], uPhase uint32, nVars int) {
	panic("TODO, stubbed")
}

// Abc_NodeGetLevel
// C: int Abc_NodeGetLevel( Abc_Obj_t * pObj )
func AbcNodeGetLevel(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NodeSuperChoiceLut (static)
// C: Abc_Obj_t * Abc_NodeSuperChoiceLut( Abc_ManScl_t * p, Abc_Obj_t * pObj )
func (p *AbcManScl) abcNodeSuperChoiceLut(pObj *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NodeCompareLevelsInc
// C: int Abc_NodeCompareLevelsInc( int * pp1, int * pp2 )
func AbcNodeCompareLevelsInc(pp1 *int, pp2 *int) int {
	panic("TODO, stubbed")
}

// Abc_NodeDecomposeSort
// C: void Abc_NodeDecomposeSort( Abc_Obj_t ** pLeaves, int nVars, int * pBSet, int nLutSize )
func AbcNodeDecomposeSort(pLeaves []*ABCNetworkObject, nVars int, pBSet []int, nLutSize int) {
	panic("TODO, stubbed")
}

// Abc_NodeDecomposeStep (static)
// C: int Abc_NodeDecomposeStep( Abc_ManScl_t * p )
func (p *AbcManScl) abcNodeDecomposeStep() int {
	panic("TODO, stubbed")
}
