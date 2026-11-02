package abc

// Hand-checked skeleton of src/opt/dar/darRefact.c.

// RefMan is Ref_Man_t.
type RefMan struct {
	// input data
	pPars *DarRefPar  // rewriting parameters
	pAig  *AigManager // AIG manager
	// computed cuts
	vCuts *NestedVector[*AigObject] // Vec_Vec_t * : the storage for cuts
	// truth table and ISOP
	vTruthElem  *PointerVector[uint32]    // elementary truth tables
	vTruthStore *PointerVector[uint32]    // storage for truth tables
	vMemory     *IntVector                // storage for ISOP
	vCutNodes   *PointerVector[AigObject] // storage for internal nodes of the cut
	// various data members
	vLeavesBest *PointerVector[AigObject] // the best set of leaves
	pGraphBest  *KitGraph                 // the best factored form
	GainBest    int                       // the best gain
	LevelBest   int                       // the level of node with the best gain
	// bi-decomposition
	DecPars BdcPar    // Bdc_Par_t (by value) : decomposition parameters
	pManDec *BdcTable // Bdc_Man_t * : decomposition manager
	// node statistics
	nNodesInit  int // the initial number of nodes
	nNodesTried int // the number of nodes tried
	nNodesBelow int // the number of nodes below the level limit
	nNodesExten int // the number of nodes with extended cut
	nCutsUsed   int // the number of rewriting steps
	nCutsTried  int // the number of cuts tries
	// timing statistics
	timeCuts  int64 // abctime
	timeEval  int64
	timeOther int64
	timeTotal int64
}

// Dar_ManDefaultRefParams
// C: void Dar_ManDefaultRefParams( Dar_RefPar_t * pPars )
func (pPars *DarRefPar) DarManDefaultRefParams() {
	panic("TODO, stubbed")
}

// Dar_ManRefStart
// C: Ref_Man_t * Dar_ManRefStart( Aig_Man_t * pAig, Dar_RefPar_t * pPars )
func (pAig *AigManager) DarManRefStart(pPars *DarRefPar) *RefMan {
	panic("TODO, stubbed")
}

// Dar_ManRefPrintStats
// C: void Dar_ManRefPrintStats( Ref_Man_t * p )
func (p *RefMan) DarManRefPrintStats() {
	panic("TODO, stubbed")
}

// Dar_ManRefStop
// C: void Dar_ManRefStop( Ref_Man_t * p )
func (p *RefMan) DarManRefStop() {
	panic("TODO, stubbed")
}

// Ref_ObjComputeCuts
// C: void Ref_ObjComputeCuts( Aig_Man_t * pAig, Aig_Obj_t * pRoot, Vec_Vec_t * vCuts )
func (pAig *AigManager) RefObjComputeCuts(pRoot *AigObject, vCuts *NestedVector[*AigObject]) {
	panic("TODO, stubbed")
}

// Ref_ObjPrint
// C: void Ref_ObjPrint( Aig_Obj_t * pObj )
func RefObjPrint(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Dar_RefactTryGraph
// C: int Dar_RefactTryGraph( Aig_Man_t * pAig, Aig_Obj_t * pRoot, Vec_Ptr_t * vCut, Kit_Graph_t * pGraph, int NodeMax, int LevelMax )
func (pAig *AigManager) DarRefactTryGraph(pRoot *AigObject, vCut *PointerVector[AigObject], pGraph *KitGraph, NodeMax int, LevelMax int) int {
	panic("TODO, stubbed")
}

// Dar_RefactBuildGraph
// C: Aig_Obj_t * Dar_RefactBuildGraph( Aig_Man_t * pAig, Vec_Ptr_t * vCut, Kit_Graph_t * pGraph )
func (pAig *AigManager) DarRefactBuildGraph(vCut *PointerVector[AigObject], pGraph *KitGraph) *AigObject {
	panic("TODO, stubbed")
}

// Dar_ManRefactorTryCuts
// C: int Dar_ManRefactorTryCuts( Ref_Man_t * p, Aig_Obj_t * pObj, int nNodesSaved, int Required )
func (p *RefMan) DarManRefactorTryCuts(pObj *AigObject, nNodesSaved int, Required int) int {
	panic("TODO, stubbed")
}

// Dar_ObjCutLevelAchieved
// C: int Dar_ObjCutLevelAchieved( Vec_Ptr_t * vCut, int nLevelMin )
func DarObjCutLevelAchieved(vCut *PointerVector[AigObject], nLevelMin int) int {
	panic("TODO, stubbed")
}

// Dar_ManRefactor
// C: int Dar_ManRefactor( Aig_Man_t * pAig, Dar_RefPar_t * pPars )
func (pAig *AigManager) DarManRefactor(pPars *DarRefPar) int {
	panic("TODO, stubbed")
}
