package abc

// Hand-checked skeleton of src/base/abci/abcMap.c.

// AbcTime is Abc_Time_t (src/base/abc/abc.h).
type AbcTime struct {
	Rise float32
	Fall float32
}

// Abc_NtkMap
// C: Abc_Ntk_t * Abc_NtkMap( Abc_Ntk_t * pNtk, double DelayTarget, double AreaMulti, double DelayMulti, float LogFan, float Slew, float Gain, int nGatesMin, int fRecovery, int fSwitching, int fSkipFanout, int fUseProfile, int fUseBuffs, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkMap(DelayTarget float64, AreaMulti float64, DelayMulti float64, LogFan float32, Slew float32, Gain float32, nGatesMin int, fRecovery int, fSwitching int, fSkipFanout int, fUseProfile int, fUseBuffs int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkMapCopyCiArrival
// C: Map_Time_t * Abc_NtkMapCopyCiArrival( Abc_Ntk_t * pNtk, Abc_Time_t * ppTimes )
func (pNtk *ABCNetwork) AbcNtkMapCopyCiArrival(ppTimes []AbcTime) []MapTime {
	panic("TODO, stubbed")
}

// Abc_NtkMapCopyCoRequired
// C: Map_Time_t * Abc_NtkMapCopyCoRequired( Abc_Ntk_t * pNtk, Abc_Time_t * ppTimes )
func (pNtk *ABCNetwork) AbcNtkMapCopyCoRequired(ppTimes []AbcTime) []MapTime {
	panic("TODO, stubbed")
}

// Abc_NtkMapCopyCiArrivalCon
// C: Map_Time_t * Abc_NtkMapCopyCiArrivalCon( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkMapCopyCiArrivalCon() []MapTime {
	panic("TODO, stubbed")
}

// Abc_NtkMapCopyCoRequiredCon
// C: Map_Time_t * Abc_NtkMapCopyCoRequiredCon( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkMapCopyCoRequiredCon() []MapTime {
	panic("TODO, stubbed")
}

// Abc_NtkToMap (static)
// C: Map_Man_t * Abc_NtkToMap( Abc_Ntk_t * pNtk, double DelayTarget, int fRecovery, float * pSwitching, int fVerbose )
func (pNtk *ABCNetwork) abcNtkToMap(DelayTarget float64, fRecovery int, pSwitching []float32, fVerbose int) *MapMan {
	panic("TODO, stubbed")
}

// Abc_NodeFromMapSuper_rec
// C: Abc_Obj_t * Abc_NodeFromMapSuper_rec( Abc_Ntk_t * pNtkNew, Map_Node_t * pNodeMap, Map_Super_t * pSuper, Abc_Obj_t * pNodePis[], int nNodePis )
func (pNtkNew *ABCNetwork) AbcNodeFromMapSuper_rec(pNodeMap *MapNode, pSuper *MapSuper, pNodePis []*ABCNetworkObject, nNodePis int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NodeFromMapPhase_rec (static)
// C: Abc_Obj_t * Abc_NodeFromMapPhase_rec( Abc_Ntk_t * pNtkNew, Map_Node_t * pNodeMap, int fPhase )
func (pNtkNew *ABCNetwork) abcNodeFromMapPhase_rec(pNodeMap *MapNode, fPhase int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NodeFromMap_rec (static)
// C: Abc_Obj_t * Abc_NodeFromMap_rec( Abc_Ntk_t * pNtkNew, Map_Node_t * pNodeMap, int fPhase )
func (pNtkNew *ABCNetwork) abcNodeFromMap_rec(pNodeMap *MapNode, fPhase int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkFromMap (static)
// C: Abc_Ntk_t * Abc_NtkFromMap( Map_Man_t * pMan, Abc_Ntk_t * pNtk, int fUseBuffs )
func (pMan *MapMan) abcNtkFromMap(pNtk *ABCNetwork, fUseBuffs int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkSuperChoice
// C: Abc_Ntk_t * Abc_NtkSuperChoice( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkSuperChoice() *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFromMapSuperChoice (static)
// C: Abc_Ntk_t * Abc_NtkFromMapSuperChoice( Map_Man_t * pMan, Abc_Ntk_t * pNtk )
func (pMan *MapMan) abcNtkFromMapSuperChoice(pNtk *ABCNetwork) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NodeSuperChoice (static)
// C: void Abc_NodeSuperChoice( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNode )
func (pNtkNew *ABCNetwork) abcNodeSuperChoice(pNode *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NodeFromMapCutPhase (static)
// C: void Abc_NodeFromMapCutPhase( Abc_Ntk_t * pNtkNew, Map_Cut_t * pCut, int fPhase )
func (pNtkNew *ABCNetwork) abcNodeFromMapCutPhase(pCut *MapCut, fPhase int) {
	panic("TODO, stubbed")
}

// Abc_NodeFromMapSuperChoice_rec (static)
// C: Abc_Obj_t * Abc_NodeFromMapSuperChoice_rec( Abc_Ntk_t * pNtkNew, Map_Super_t * pSuper, Abc_Obj_t * pNodePis[], int nNodePis )
func (pNtkNew *ABCNetwork) abcNodeFromMapSuperChoice_rec(pSuper *MapSuper, pNodePis []*ABCNetworkObject, nNodePis int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkFetchTwinNode
// C: Abc_Obj_t * Abc_NtkFetchTwinNode( Abc_Obj_t * pNode )
func AbcNtkFetchTwinNode(pNode *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkWriteMiniMapping
// C: Vec_Int_t * Abc_NtkWriteMiniMapping( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkWriteMiniMapping() *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkPrintMiniMapping
// C: void Abc_NtkPrintMiniMapping( int * pArray )
func AbcNtkPrintMiniMapping(pArray []int) {
	panic("TODO, stubbed")
}

// Abc_NtkOutputMiniMapping
// C: int * Abc_NtkOutputMiniMapping( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) AbcNtkOutputMiniMapping() []int {
	panic("TODO, stubbed")
}

// Abc_NtkTestMiniMapping
// C: void Abc_NtkTestMiniMapping( Abc_Ntk_t * p )
func (p *ABCNetwork) AbcNtkTestMiniMapping() {
	panic("TODO, stubbed")
}

// Abc_NtkSetCiArrivalTime
// C: void Abc_NtkSetCiArrivalTime( Abc_Frame_t * pAbc, int iCi, float Rise, float Fall )
func (pAbc *MainFrame) AbcNtkSetCiArrivalTime(iCi int, Rise float32, Fall float32) {
	panic("TODO, stubbed")
}

// Abc_NtkSetCoRequiredTime
// C: void Abc_NtkSetCoRequiredTime( Abc_Frame_t * pAbc, int iCo, float Rise, float Fall )
func (pAbc *MainFrame) AbcNtkSetCoRequiredTime(iCo int, Rise float32, Fall float32) {
	panic("TODO, stubbed")
}

// Abc_NtkSetAndGateDelay
// C: void Abc_NtkSetAndGateDelay( Abc_Frame_t * pAbc, float Delay )
func (pAbc *MainFrame) AbcNtkSetAndGateDelay(Delay float32) {
	panic("TODO, stubbed")
}
