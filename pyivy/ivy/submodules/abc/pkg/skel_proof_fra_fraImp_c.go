package abc

// Hand-checked skeleton of src/proof/fra/fraImp.c.

// Fra_SmlCountOnesOne (static inline)
// C: int Fra_SmlCountOnesOne( Fra_Sml_t * p, int Node )
func (p *FraSml) fraSmlCountOnesOne(Node int) int {
	panic("TODO, stubbed")
}

// Fra_SmlCountOnes (static inline)
// C: int * Fra_SmlCountOnes( Fra_Sml_t * p )
func (p *FraSml) fraSmlCountOnes() []int {
	panic("TODO, stubbed")
}

// Sml_NodeCheckImp (static inline)
// C: int Sml_NodeCheckImp( Fra_Sml_t * p, int Left, int Right )
func (p *FraSml) smlNodeCheckImp(Left int, Right int) int {
	panic("TODO, stubbed")
}

// Sml_NodeNotImpWeight (static inline)
// C: int Sml_NodeNotImpWeight( Fra_Sml_t * p, int Left, int Right )
func (p *FraSml) smlNodeNotImpWeight(Left int, Right int) int {
	panic("TODO, stubbed")
}

// Sml_NodeSaveNotImpPatterns (static inline)
// C: void Sml_NodeSaveNotImpPatterns( Fra_Sml_t * p, int Left, int Right, unsigned * pResult )
func (p *FraSml) smlNodeSaveNotImpPatterns(Left int, Right int, pResult []uint32) {
	panic("TODO, stubbed")
}

// Fra_SmlSortUsingOnes
// C: Vec_Ptr_t * Fra_SmlSortUsingOnes( Fra_Sml_t * p, int fLatchCorr )
func (p *FraSml) FraSmlSortUsingOnes(fLatchCorr int) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Fra_SmlSelectMaxCost
// C: Vec_Int_t * Fra_SmlSelectMaxCost( Vec_Int_t * vImps, int * pCosts, int nCostMax, int nImpLimit, int * pCostRange )
// pCostRange is an optional (nullable) output parameter.
func FraSmlSelectMaxCost(vImps *IntVector, pCosts []int, nCostMax int, nImpLimit int, pCostRange *int) *IntVector {
	panic("TODO, stubbed")
}

// Sml_CompareMaxId
// C: int Sml_CompareMaxId( unsigned short * pImp1, unsigned short * pImp2 )
// Each argument points at an implication pair (entries [0] and [1]).
func SmlCompareMaxId(pImp1 *[2]uint16, pImp2 *[2]uint16) int {
	panic("TODO, stubbed")
}

// Fra_ImpDerive
// C: Vec_Int_t * Fra_ImpDerive( Fra_Man_t * p, int nImpMaxLimit, int nImpUseLimit, int fLatchCorr )
func (p *FraMan) FraImpDerive(nImpMaxLimit int, nImpUseLimit int, fLatchCorr int) *IntVector {
	panic("TODO, stubbed")
}

// Fra_ImpAddToSolver
// C: void Fra_ImpAddToSolver( Fra_Man_t * p, Vec_Int_t * vImps, int * pSatVarNums )
func (p *FraMan) FraImpAddToSolver(vImps *IntVector, pSatVarNums []int) {
	panic("TODO, stubbed")
}

// Fra_ImpCheckForNode
// C: int Fra_ImpCheckForNode( Fra_Man_t * p, Vec_Int_t * vImps, Aig_Obj_t * pNode, int Pos )
func (p *FraMan) FraImpCheckForNode(vImps *IntVector, pNode *AigObject, Pos int) int {
	panic("TODO, stubbed")
}

// Fra_ImpRefineUsingCex
// C: int Fra_ImpRefineUsingCex( Fra_Man_t * p, Vec_Int_t * vImps )
func (p *FraMan) FraImpRefineUsingCex(vImps *IntVector) int {
	panic("TODO, stubbed")
}

// Fra_ImpCompactArray
// C: void Fra_ImpCompactArray( Vec_Int_t * vImps )
func FraImpCompactArray(vImps *IntVector) {
	panic("TODO, stubbed")
}

// Fra_ImpComputeStateSpaceRatio
// C: double Fra_ImpComputeStateSpaceRatio( Fra_Man_t * p )
func (p *FraMan) FraImpComputeStateSpaceRatio() float64 {
	panic("TODO, stubbed")
}

// Fra_ImpVerifyUsingSimulation
// C: int Fra_ImpVerifyUsingSimulation( Fra_Man_t * p )
func (p *FraMan) FraImpVerifyUsingSimulation() int {
	panic("TODO, stubbed")
}

// Fra_ImpRecordInManager
// C: void Fra_ImpRecordInManager( Fra_Man_t * p, Aig_Man_t * pNew )
func (p *FraMan) FraImpRecordInManager(pNew *AigManager) {
	panic("TODO, stubbed")
}
