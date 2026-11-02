package abc

// Hand-checked skeleton of src/base/abci/abcRr.c.

// AbcRRMan is Abc_RRMan_t.
type AbcRRMan struct {
	pNtk          *ABCNetwork                      // the network
	nFaninLevels  int                              // the number of fanin levels
	nFanoutLevels int                              // the number of fanout levels
	pNode         *ABCNetworkObject                // the node
	pFanin        *ABCNetworkObject                // the fanin
	pFanout       *ABCNetworkObject                // the fanout
	vFaninLeaves  *PointerVector[ABCNetworkObject] // the leaves of the fanin cone
	vFanoutRoots  *PointerVector[ABCNetworkObject] // the roots of the fanout cone
	vLeaves       *PointerVector[ABCNetworkObject] // the leaves of the window
	vCone         *PointerVector[ABCNetworkObject] // the internal nodes of the window
	vRoots        *PointerVector[ABCNetworkObject] // the roots of the window
	pWnd          *ABCNetwork                      // the window derived for the edge
	pMiter        *ABCNetwork                      // the miter derived from the window
	pParams       *ProveParams                     // Prove_Params_t * : the miter proving parameters
	nNodesOld     int                              // the old number of nodes
	nLevelsOld    int                              // the old number of levels
	nEdgesTried   int                              // the number of nodes tried
	nEdgesRemoved int                              // the number of nodes proved
	timeWindow    int64                            // abctime : the time to construct the window
	timeMiter     int64                            // abctime : the time to construct the miter
	timeProve     int64                            // abctime : the time to prove the miter
	timeUpdate    int64                            // abctime : the network update time
	timeTotal     int64                            // abctime : the total runtime
}

// Abc_NtkRR
// C: int Abc_NtkRR( Abc_Ntk_t * pNtk, int nFaninLevels, int nFanoutLevels, int fUseFanouts, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkRR(nFaninLevels int, nFanoutLevels int, fUseFanouts int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_RRManStart (static)
// C: Abc_RRMan_t * Abc_RRManStart()
func abcRRManStart() *AbcRRMan {
	panic("TODO, stubbed")
}

// Abc_RRManStop (static)
// C: void Abc_RRManStop( Abc_RRMan_t * p )
func (p *AbcRRMan) abcRRManStop() {
	panic("TODO, stubbed")
}

// Abc_RRManPrintStats (static)
// C: void Abc_RRManPrintStats( Abc_RRMan_t * p )
func (p *AbcRRMan) abcRRManPrintStats() {
	panic("TODO, stubbed")
}

// Abc_RRManClean (static)
// C: void Abc_RRManClean( Abc_RRMan_t * p )
func (p *AbcRRMan) abcRRManClean() {
	panic("TODO, stubbed")
}

// Abc_NtkRRProve (static)
// C: int Abc_NtkRRProve( Abc_RRMan_t * p )
func (p *AbcRRMan) abcNtkRRProve() int {
	panic("TODO, stubbed")
}

// Abc_NtkRRUpdate (static)
// C: int Abc_NtkRRUpdate( Abc_Ntk_t * pNtk, Abc_Obj_t * pNode, Abc_Obj_t * pFanin, Abc_Obj_t * pFanout )
func abcNtkRRUpdate(pNtk *ABCNetwork, pNode *ABCNetworkObject, pFanin *ABCNetworkObject, pFanout *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkRRWindow (static)
// C: int Abc_NtkRRWindow( Abc_RRMan_t * p )
func (p *AbcRRMan) abcNtkRRWindow() int {
	panic("TODO, stubbed")
}

// Abc_NtkRRTfi_int (static)
// C: int Abc_NtkRRTfi_int( Vec_Ptr_t * vLeaves, int LevelLimit )
func abcNtkRRTfi_int(vLeaves *PointerVector[ABCNetworkObject], LevelLimit int) int {
	panic("TODO, stubbed")
}

// Abc_NtkRRTfo_int (static)
// C: int Abc_NtkRRTfo_int( Vec_Ptr_t * vLeaves, Vec_Ptr_t * vRoots, int LevelLimit, Abc_Obj_t * pEdgeFanin, Abc_Obj_t * pEdgeFanout )
func abcNtkRRTfo_int(vLeaves *PointerVector[ABCNetworkObject], vRoots *PointerVector[ABCNetworkObject], LevelLimit int, pEdgeFanin *ABCNetworkObject, pEdgeFanout *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkRRTfo_rec (static)
// C: int Abc_NtkRRTfo_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vRoots, int LevelLimit )
func abcNtkRRTfo_rec(pNode *ABCNetworkObject, vRoots *PointerVector[ABCNetworkObject], LevelLimit int) int {
	panic("TODO, stubbed")
}

// Abc_NtkRRTfi_rec (static)
// C: void Abc_NtkRRTfi_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vCone, int LevelLimit )
func abcNtkRRTfi_rec(pNode *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], vCone *PointerVector[ABCNetworkObject], LevelLimit int) {
	panic("TODO, stubbed")
}

// Abc_NtkWindow (static)
// C: Abc_Ntk_t * Abc_NtkWindow( Abc_Ntk_t * pNtk, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vCone, Vec_Ptr_t * vRoots )
func (pNtk *ABCNetwork) abcNtkWindow(vLeaves *PointerVector[ABCNetworkObject], vCone *PointerVector[ABCNetworkObject], vRoots *PointerVector[ABCNetworkObject]) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkRRSimulateStart (static)
// C: void Abc_NtkRRSimulateStart( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkRRSimulateStart() {
	panic("TODO, stubbed")
}

// Abc_NtkRRSimulateStop (static)
// C: void Abc_NtkRRSimulateStop( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkRRSimulateStop() {
	panic("TODO, stubbed")
}

// Abc_NtkRRSimulate
// C: Vec_Str_t * Abc_NtkRRSimulate( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkRRSimulate() *ByteVector {
	panic("TODO, stubbed")
}

// Sim_TraverseNodes_rec (static)
// C: void Sim_TraverseNodes_rec( Abc_Obj_t * pRoot, Vec_Str_t * vTargets, Vec_Ptr_t * vNodes )
func simTraverseNodes_rec(pRoot *ABCNetworkObject, vTargets *ByteVector, vNodes *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Sim_CollectNodes_rec (static)
// C: void Sim_CollectNodes_rec( Abc_Obj_t * pRoot, Vec_Ptr_t * vField )
func simCollectNodes_rec(pRoot *ABCNetworkObject, vField *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Sim_SimulateCollected (static)
// C: void Sim_SimulateCollected( Vec_Str_t * vTargets, Vec_Ptr_t * vNodes, Vec_Ptr_t * vField )
func simSimulateCollected(vTargets *ByteVector, vNodes *PointerVector[ABCNetworkObject], vField *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}
