package abc

// Hand-checked skeleton of src/base/abci/abcRestruct.c.

// AbcManRst is Abc_ManRst_t.
type AbcManRst struct {
	pNtk               *ABCNetwork                      // the network for restructuring
	nCutMax            int                              // the limit on the size of the supernode
	fUpdateLevel       int                              // turns on watching the number of levels
	fUseZeros          int                              // turns on zero-cost replacements
	fVerbose           int                              // the verbosity flag
	dd                 *CuddFoundation                  // the BDD manager
	pManDsd            *DSDManager                      // the DSD manager
	vVisited           *PointerVector[ABCNetworkObject] // temporary
	vLeaves            *PointerVector[ABCNetworkObject] // temporary
	vDecs              *PointerVector[ABCNetworkObject] // temporary
	vTemp              *PointerVector[ABCNetworkObject] // temporary
	vSims              *IntVector                       // temporary
	vRands             *IntVector                       // temporary
	vOnes              *IntVector                       // temporary
	vBinate            *IntVector                       // temporary
	vTwos              *IntVector                       // temporary
	nLastGain          int
	nCutsConsidered    int
	nCutsExplored      int
	nNodesConsidered   int
	nNodesRestructured int
	nNodesGained       int
	timeCut            int
	timeBdd            int
	timeDsd            int
	timeEval           int
	timeRes            int
	timeNtk            int
	timeTotal          int
}

// Abc_NtkRestructure
// C: int Abc_NtkRestructure( Abc_Ntk_t * pNtk, int nCutMax, int fUpdateLevel, int fUseZeros, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkRestructure(nCutMax int, fUpdateLevel int, fUseZeros int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_RestructNodeDivisors
// C: void Abc_RestructNodeDivisors( Abc_ManRst_t * p, Abc_Obj_t * pRoot, int nNodesSaved )
func (p *AbcManRst) AbcRestructNodeDivisors(pRoot *ABCNetworkObject, nNodesSaved int) {
	panic("TODO, stubbed")
}

// Abc_NodeRestructure (static)
// C: Dec_Graph_t * Abc_NodeRestructure( Abc_ManRst_t * p, Abc_Obj_t * pNode, Cut_Cut_t * pCutList )
func (p *AbcManRst) abcNodeRestructure(pNode *ABCNetworkObject, pCutList *CutCut) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_NodeRestructureCut (static)
// C: Dec_Graph_t * Abc_NodeRestructureCut( Abc_ManRst_t * p, Abc_Obj_t * pRoot, Cut_Cut_t * pCut )
func (p *AbcManRst) abcNodeRestructureCut(pRoot *ABCNetworkObject, pCut *CutCut) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_NodeEdgeDsdPermute
// C: void Abc_NodeEdgeDsdPermute( Dec_Graph_t * pGraph, Abc_ManRst_t * pManRst, Vec_Int_t * vEdges, int fExor )
func (pGraph *DecGraph) AbcNodeEdgeDsdPermute(pManRst *AbcManRst, vEdges *IntVector, fExor int) {
	panic("TODO, stubbed")
}

// Abc_NodeEdgeDsdPushOrdered
// C: void Abc_NodeEdgeDsdPushOrdered( Dec_Graph_t * pGraph, Vec_Int_t * vEdges, int Edge )
func (pGraph *DecGraph) AbcNodeEdgeDsdPushOrdered(vEdges *IntVector, Edge int) {
	panic("TODO, stubbed")
}

// Abc_NodeEvaluateDsd_rec
// C: Dec_Edge_t Abc_NodeEvaluateDsd_rec( Dec_Graph_t * pGraph, Abc_ManRst_t * pManRst, Dsd_Node_t * pNodeDsd, int Required, int nNodesSaved, int * pnNodesAdded )
func (pGraph *DecGraph) AbcNodeEvaluateDsd_rec(pManRst *AbcManRst, pNodeDsd *DSDNode, Required int, nNodesSaved int, pnNodesAdded *int) DecEdge {
	panic("TODO, stubbed")
}

// Abc_NodeEvaluateDsd (static)
// C: Dec_Graph_t * Abc_NodeEvaluateDsd( Abc_ManRst_t * pManRst, Dsd_Node_t * pNodeDsd, Abc_Obj_t * pRoot, int Required, int nNodesSaved, int * pnNodesAdded )
func (pManRst *AbcManRst) abcNodeEvaluateDsd(pNodeDsd *DSDNode, pRoot *ABCNetworkObject, Required int, nNodesSaved int, pnNodesAdded *int) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_NtkStartCutManForRestruct (static)
// C: Cut_Man_t * Abc_NtkStartCutManForRestruct( Abc_Ntk_t * pNtk, int nCutMax, int fDag )
func (pNtk *ABCNetwork) abcNtkStartCutManForRestruct(nCutMax int, fDag int) *CutMan {
	panic("TODO, stubbed")
}

// Abc_NtkManRstStart (static)
// C: Abc_ManRst_t * Abc_NtkManRstStart( int nCutMax, int fUpdateLevel, int fUseZeros, int fVerbose )
func abcNtkManRstStart(nCutMax int, fUpdateLevel int, fUseZeros int, fVerbose int) *AbcManRst {
	panic("TODO, stubbed")
}

// Abc_NtkManRstStop (static)
// C: void Abc_NtkManRstStop( Abc_ManRst_t * p )
func (p *AbcManRst) abcNtkManRstStop() {
	panic("TODO, stubbed")
}

// Abc_NtkManRstPrintStats (static)
// C: void Abc_NtkManRstPrintStats( Abc_ManRst_t * p )
func (p *AbcManRst) abcNtkManRstPrintStats() {
	panic("TODO, stubbed")
}

// Abc_Abc_NodeResubCollectDivs
// C: int Abc_Abc_NodeResubCollectDivs( Abc_ManRst_t * p, Abc_Obj_t * pRoot, Cut_Cut_t * pCut )
func (p *AbcManRst) AbcAbcNodeResubCollectDivs(pRoot *ABCNetworkObject, pCut *CutCut) int {
	panic("TODO, stubbed")
}

// Abc_NodeResubMffc_rec
// C: int Abc_NodeResubMffc_rec( Abc_Obj_t * pNode )
func AbcNodeResubMffc_rec(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NodeResubMffc
// C: int Abc_NodeResubMffc( Abc_ManRst_t * p, Vec_Ptr_t * vDecs, int nLeaves, Abc_Obj_t * pRoot )
func (p *AbcManRst) AbcNodeResubMffc(vDecs *PointerVector[ABCNetworkObject], nLeaves int, pRoot *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NodeMffcSimulate
// C: void Abc_NodeMffcSimulate( Vec_Ptr_t * vDecs, int nLeaves, Vec_Int_t * vRands, Vec_Int_t * vSims )
func AbcNodeMffcSimulate(vDecs *PointerVector[ABCNetworkObject], nLeaves int, vRands *IntVector, vSims *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NodeCheckFull
// C: int Abc_NodeCheckFull( Abc_ManRst_t * p, Dec_Graph_t * pGraph )
func (p *AbcManRst) AbcNodeCheckFull(pGraph *DecGraph) int {
	panic("TODO, stubbed")
}

// Abc_NodeMffcConstants
// C: Dec_Graph_t * Abc_NodeMffcConstants( Abc_ManRst_t * p, Vec_Int_t * vSims )
func (p *AbcManRst) AbcNodeMffcConstants(vSims *IntVector) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_NodeMffcSingleVar
// C: Dec_Graph_t * Abc_NodeMffcSingleVar( Abc_ManRst_t * p, Vec_Int_t * vSims, int nNodes, Vec_Int_t * vOnes )
func (p *AbcManRst) AbcNodeMffcSingleVar(vSims *IntVector, nNodes int, vOnes *IntVector) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_NodeMffcSingleNode
// C: Dec_Graph_t * Abc_NodeMffcSingleNode( Abc_ManRst_t * p, Vec_Int_t * vSims, int nNodes, Vec_Int_t * vOnes )
func (p *AbcManRst) AbcNodeMffcSingleNode(vSims *IntVector, nNodes int, vOnes *IntVector) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_NodeMffcDoubleNode
// C: Dec_Graph_t * Abc_NodeMffcDoubleNode( Abc_ManRst_t * p, Vec_Int_t * vSims, int nNodes, Vec_Int_t * vOnes )
func (p *AbcManRst) AbcNodeMffcDoubleNode(vSims *IntVector, nNodes int, vOnes *IntVector) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_NodeResubEval
// C: Dec_Graph_t * Abc_NodeResubEval( Abc_ManRst_t * p, Abc_Obj_t * pRoot, Cut_Cut_t * pCut )
func (p *AbcManRst) AbcNodeResubEval(pRoot *ABCNetworkObject, pCut *CutCut) *DecGraph {
	panic("TODO, stubbed")
}

// Abc_NodeResubstitute (static)
// C: Dec_Graph_t * Abc_NodeResubstitute( Abc_ManRst_t * p, Abc_Obj_t * pNode, Cut_Cut_t * pCutList )
func (p *AbcManRst) abcNodeResubstitute(pNode *ABCNetworkObject, pCutList *CutCut) *DecGraph {
	panic("TODO, stubbed")
}
