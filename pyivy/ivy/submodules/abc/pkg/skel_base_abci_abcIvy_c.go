package abc

// Hand-checked skeleton of src/base/abci/abcIvy.c.

// AbcEdge is Abc_Edge_t: typedef int Abc_Edge_t.
type AbcEdge = int

// Abc_EdgeCreate (static)
// C: Abc_Edge_t Abc_EdgeCreate( int Id, int fCompl )
func abcEdgeCreate(Id int, fCompl int) AbcEdge {
	panic("TODO, stubbed")
}

// Abc_EdgeId (static)
// C: int Abc_EdgeId( Abc_Edge_t Edge )
func abcEdgeId(Edge AbcEdge) int {
	panic("TODO, stubbed")
}

// Abc_EdgeIsComplement (static)
// C: int Abc_EdgeIsComplement( Abc_Edge_t Edge )
func abcEdgeIsComplement(Edge AbcEdge) int {
	panic("TODO, stubbed")
}

// Abc_EdgeRegular (static)
// C: Abc_Edge_t Abc_EdgeRegular( Abc_Edge_t Edge )
func abcEdgeRegular(Edge AbcEdge) AbcEdge {
	panic("TODO, stubbed")
}

// Abc_EdgeNot (static)
// C: Abc_Edge_t Abc_EdgeNot( Abc_Edge_t Edge )
func abcEdgeNot(Edge AbcEdge) AbcEdge {
	panic("TODO, stubbed")
}

// Abc_EdgeNotCond (static)
// C: Abc_Edge_t Abc_EdgeNotCond( Abc_Edge_t Edge, int fCond )
func abcEdgeNotCond(Edge AbcEdge, fCond int) AbcEdge {
	panic("TODO, stubbed")
}

// Abc_EdgeFromNode (static)
// C: Abc_Edge_t Abc_EdgeFromNode( Abc_Obj_t * pNode )
func abcEdgeFromNode(pNode *ABCNetworkObject) AbcEdge {
	panic("TODO, stubbed")
}

// Abc_EdgeToNode (static)
// C: Abc_Obj_t * Abc_EdgeToNode( Abc_Ntk_t * p, Abc_Edge_t Edge )
func (p *ABCNetwork) abcEdgeToNode(Edge AbcEdge) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_ObjFanin0Ivy (static)
// C: Abc_Obj_t * Abc_ObjFanin0Ivy( Abc_Ntk_t * p, Ivy_Obj_t * pObj )
func (p *ABCNetwork) abcObjFanin0Ivy(pObj *IvyObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_ObjFanin1Ivy (static)
// C: Abc_Obj_t * Abc_ObjFanin1Ivy( Abc_Ntk_t * p, Ivy_Obj_t * pObj )
func (p *ABCNetwork) abcObjFanin1Ivy(pObj *IvyObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkIvyBefore
// C: Ivy_Man_t * Abc_NtkIvyBefore( Abc_Ntk_t * pNtk, int fSeq, int fUseDc )
func (pNtk *ABCNetwork) AbcNtkIvyBefore(fSeq int, fUseDc int) *IvyManager {
	panic("TODO, stubbed")
}

// Abc_NtkIvyAfter
// C: Abc_Ntk_t * Abc_NtkIvyAfter( Abc_Ntk_t * pNtk, Ivy_Man_t * pMan, int fSeq, int fHaig )
func (pNtk *ABCNetwork) AbcNtkIvyAfter(pMan *IvyManager, fSeq int, fHaig int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkIvyStrash
// C: Abc_Ntk_t * Abc_NtkIvyStrash( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkIvyStrash() *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkIvyHaig
// C: Abc_Ntk_t * Abc_NtkIvyHaig( Abc_Ntk_t * pNtk, int nIters, int fUseZeroCost, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkIvyHaig(nIters int, fUseZeroCost int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkIvyCuts
// C: void Abc_NtkIvyCuts( Abc_Ntk_t * pNtk, int nInputs )
func (pNtk *ABCNetwork) AbcNtkIvyCuts(nInputs int) {
	panic("TODO, stubbed")
}

// Abc_NtkIvyRewrite
// C: Abc_Ntk_t * Abc_NtkIvyRewrite( Abc_Ntk_t * pNtk, int fUpdateLevel, int fUseZeroCost, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkIvyRewrite(fUpdateLevel int, fUseZeroCost int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkIvyRewriteSeq
// C: Abc_Ntk_t * Abc_NtkIvyRewriteSeq( Abc_Ntk_t * pNtk, int fUseZeroCost, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkIvyRewriteSeq(fUseZeroCost int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkIvyResyn0
// C: Abc_Ntk_t * Abc_NtkIvyResyn0( Abc_Ntk_t * pNtk, int fUpdateLevel, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkIvyResyn0(fUpdateLevel int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkIvyResyn
// C: Abc_Ntk_t * Abc_NtkIvyResyn( Abc_Ntk_t * pNtk, int fUpdateLevel, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkIvyResyn(fUpdateLevel int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkIvySat
// C: Abc_Ntk_t * Abc_NtkIvySat( Abc_Ntk_t * pNtk, int nConfLimit, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkIvySat(nConfLimit int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkTransferPointers
// C: void Abc_NtkTransferPointers( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkAig )
func (pNtk *ABCNetwork) AbcNtkTransferPointers(pNtkAig *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NtkIvyFraig
// C: Abc_Ntk_t * Abc_NtkIvyFraig( Abc_Ntk_t * pNtk, int nConfLimit, int fDoSparse, int fProve, int fTransfer, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkIvyFraig(nConfLimit int, fDoSparse int, fProve int, fTransfer int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkIvyProve (pPars is a void * that holds an Prove_Params_t *)
// C: int Abc_NtkIvyProve( Abc_Ntk_t ** ppNtk, void * pPars )
func AbcNtkIvyProve(ppNtk **ABCNetwork, pPars *ProveParams) int {
	panic("TODO, stubbed")
}

// Abc_NtkIvy
// C: Abc_Ntk_t * Abc_NtkIvy( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkIvy() *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFromIvy (static)
// C: Abc_Ntk_t * Abc_NtkFromIvy( Abc_Ntk_t * pNtkOld, Ivy_Man_t * pMan )
func (pNtkOld *ABCNetwork) abcNtkFromIvy(pMan *IvyManager) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFromIvySeq (static)
// C: Abc_Ntk_t * Abc_NtkFromIvySeq( Abc_Ntk_t * pNtkOld, Ivy_Man_t * pMan, int fHaig )
func (pNtkOld *ABCNetwork) abcNtkFromIvySeq(pMan *IvyManager, fHaig int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkToIvy (static)
// C: Ivy_Man_t * Abc_NtkToIvy( Abc_Ntk_t * pNtkOld )
func (pNtkOld *ABCNetwork) abcNtkToIvy() *IvyManager {
	panic("TODO, stubbed")
}

// Abc_NtkStrashPerformAig (static)
// C: void Abc_NtkStrashPerformAig( Abc_Ntk_t * pNtk, Ivy_Man_t * pMan )
func (pNtk *ABCNetwork) abcNtkStrashPerformAig(pMan *IvyManager) {
	panic("TODO, stubbed")
}

// Abc_NodeStrashAig (static)
// C: Ivy_Obj_t * Abc_NodeStrashAig( Ivy_Man_t * pMan, Abc_Obj_t * pNode )
func (pMan *IvyManager) abcNodeStrashAig(pNode *ABCNetworkObject) *IvyObject {
	panic("TODO, stubbed")
}

// Abc_NodeStrashAigSopAig (static)
// C: Ivy_Obj_t * Abc_NodeStrashAigSopAig( Ivy_Man_t * pMan, Abc_Obj_t * pNode, char * pSop )
func (pMan *IvyManager) abcNodeStrashAigSopAig(pNode *ABCNetworkObject, pSop string) *IvyObject {
	panic("TODO, stubbed")
}

// Abc_NodeStrashAigExorAig (static)
// C: Ivy_Obj_t * Abc_NodeStrashAigExorAig( Ivy_Man_t * pMan, Abc_Obj_t * pNode, char * pSop )
func (pMan *IvyManager) abcNodeStrashAigExorAig(pNode *ABCNetworkObject, pSop string) *IvyObject {
	panic("TODO, stubbed")
}

// Abc_NodeStrashAigFactorAig (static)
// C: Ivy_Obj_t * Abc_NodeStrashAigFactorAig( Ivy_Man_t * pMan, Abc_Obj_t * pRoot, char * pSop )
func (pMan *IvyManager) abcNodeStrashAigFactorAig(pRoot *ABCNetworkObject, pSop string) *IvyObject {
	panic("TODO, stubbed")
}

// Abc_NtkCollectLatchValuesIvy (static)
// C: Vec_Int_t * Abc_NtkCollectLatchValuesIvy( Abc_Ntk_t * pNtk, int fUseDcs )
func (pNtk *ABCNetwork) abcNtkCollectLatchValuesIvy(fUseDcs int) *IntVector {
	panic("TODO, stubbed")
}
