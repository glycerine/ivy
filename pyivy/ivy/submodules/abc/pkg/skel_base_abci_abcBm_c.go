package abc

// Hand-checked skeleton of src/base/abci/abcBm.c.

// getDependencies
// C: void getDependencies(Abc_Ntk_t *pNtk, Vec_Int_t** iDep, Vec_Int_t** oDep)
func (pNtk *ABCNetwork) GetDependencies(iDep *[]*IntVector, oDep *[]*IntVector) {
	panic("TODO, stubbed")
}

// initMatchList
// C: void initMatchList(Abc_Ntk_t *pNtk, Vec_Int_t** iDep, Vec_Int_t** oDep, Vec_Int_t** iMatch, int* iLastItem, Vec_Int_t** oMatch, int* oLastItem, int* iGroup, int* oGroup, int p_equivalence)
func (pNtk *ABCNetwork) InitMatchList(iDep []*IntVector, oDep []*IntVector, iMatch []*IntVector, iLastItem *int, oMatch []*IntVector, oLastItem *int, iGroup []int, oGroup []int, p_equivalence int) {
	panic("TODO, stubbed")
}

// iSortDependencies
// C: void iSortDependencies(Abc_Ntk_t *pNtk, Vec_Int_t** iDep, int* oGroup)
func (pNtk *ABCNetwork) ISortDependencies(iDep []*IntVector, oGroup []int) {
	panic("TODO, stubbed")
}

// oSortDependencies
// C: void oSortDependencies(Abc_Ntk_t *pNtk, Vec_Int_t** oDep, int* iGroup)
func (pNtk *ABCNetwork) OSortDependencies(oDep []*IntVector, iGroup []int) {
	panic("TODO, stubbed")
}

// oSplitByDep
// C: int oSplitByDep(Abc_Ntk_t *pNtk, Vec_Int_t** oDep, Vec_Int_t** oMatch, int* oGroup, int* oLastItem, int* iGroup)
func (pNtk *ABCNetwork) OSplitByDep(oDep []*IntVector, oMatch []*IntVector, oGroup []int, oLastItem *int, iGroup []int) int {
	panic("TODO, stubbed")
}

// iSplitByDep
// C: int iSplitByDep(Abc_Ntk_t *pNtk, Vec_Int_t** iDep, Vec_Int_t** iMatch, int* iGroup, int* iLastItem, int* oGroup)
func (pNtk *ABCNetwork) ISplitByDep(iDep []*IntVector, iMatch []*IntVector, iGroup []int, iLastItem *int, oGroup []int) int {
	panic("TODO, stubbed")
}

// findTopologicalOrder
// C: Vec_Ptr_t ** findTopologicalOrder( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) FindTopologicalOrder() []*PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkSimulateOneNode
// C: int * Abc_NtkSimulateOneNode( Abc_Ntk_t * pNtk, int * pModel, int input, Vec_Ptr_t ** topOrder )
func (pNtk *ABCNetwork) AbcNtkSimulateOneNode(pModel []int, input int, topOrder []*PointerVector[ABCNetworkObject]) []int {
	panic("TODO, stubbed")
}

// refineIOBySimulation
// C: int refineIOBySimulation(Abc_Ntk_t *pNtk, Vec_Int_t** iMatch, int* iLastItem, int * iGroup, Vec_Int_t** iDep, Vec_Int_t** oMatch, int* oLastItem, int * oGroup, Vec_Int_t** oDep, char * vPiValues, int * observability, Vec_Ptr_t ** topOrder)
func (pNtk *ABCNetwork) RefineIOBySimulation(iMatch []*IntVector, iLastItem *int, iGroup []int, iDep []*IntVector, oMatch []*IntVector, oLastItem *int, oGroup []int, oDep []*IntVector, vPiValues []int8, observability []int, topOrder []*PointerVector[ABCNetworkObject]) int {
	panic("TODO, stubbed")
}

// Abc_NtkMiterBm
// C: Abc_Ntk_t * Abc_NtkMiterBm( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, Vec_Ptr_t * iCurrMatch, Vec_Ptr_t * oCurrMatch )
func (pNtk1 *ABCNetwork) AbcNtkMiterBm(pNtk2 *ABCNetwork, iCurrMatch *PointerVector[ABCNetworkObject], oCurrMatch *PointerVector[ABCNetworkObject]) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkVerifyReportError
// C: void Abc_NtkVerifyReportError( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel, Vec_Int_t * mismatch )
func (pNtk1 *ABCNetwork) AbcNtkVerifyReportError(pNtk2 *ABCNetwork, pModel []int, mismatch *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkMiterSatBm
// C: int Abc_NtkMiterSatBm( Abc_Ntk_t * pNtk, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, int fVerbose, ABC_INT64_T * pNumConfs, ABC_INT64_T * pNumInspects)
func (pNtk *ABCNetwork) AbcNtkMiterSatBm(nConfLimit int64, nInsLimit int64, fVerbose int, pNumConfs *int64, pNumInspects *int64) int {
	panic("TODO, stubbed")
}

// Abc_NtkBmSat
// C: int Abc_NtkBmSat( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, Vec_Ptr_t * iMatchPairs, Vec_Ptr_t * oMatchPairs, Vec_Int_t * mismatch, int mode)
func (pNtk1 *ABCNetwork) AbcNtkBmSat(pNtk2 *ABCNetwork, iMatchPairs *PointerVector[ABCNetworkObject], oMatchPairs *PointerVector[ABCNetworkObject], mismatch *IntVector, mode int) int {
	panic("TODO, stubbed")
}

// checkEquivalence
// C: int checkEquivalence( Abc_Ntk_t * pNtk1, Vec_Int_t* matchedInputs1, Vec_Int_t * matchedOutputs1, Abc_Ntk_t * pNtk2, Vec_Int_t* matchedInputs2, Vec_Int_t * matchedOutputs2)
func (pNtk1 *ABCNetwork) CheckEquivalence(matchedInputs1 *IntVector, matchedOutputs1 *IntVector, pNtk2 *ABCNetwork, matchedInputs2 *IntVector, matchedOutputs2 *IntVector) int {
	panic("TODO, stubbed")
}

// computeCofactor
// C: Abc_Ntk_t * computeCofactor(Abc_Ntk_t * pNtk, Vec_Ptr_t ** nodesInLevel, int * bitVector, Vec_Int_t * currInputs)
func (pNtk *ABCNetwork) ComputeCofactor(nodesInLevel []*PointerVector[ABCNetworkObject], bitVector []int, currInputs *IntVector) *ABCNetwork {
	panic("TODO, stubbed")
}

// matchNonSingletonOutputs
// C: int matchNonSingletonOutputs(Abc_Ntk_t * pNtk1, Vec_Ptr_t ** nodesInLevel1, Vec_Int_t ** iMatch1, Vec_Int_t ** iDep1, Vec_Int_t * matchedInputs1, int * iGroup1, Vec_Int_t ** oMatch1, int * oGroup1, Abc_Ntk_t * pNtk2, Vec_Ptr_t ** nodesInLevel2, Vec_Int_t ** iMatch2, Vec_Int_t ** iDep2, Vec_Int_t * matchedInputs2, int * iGroup2, Vec_Int_t ** oMatch2, int * oGroup2, Vec_Int_t * matchedOutputs1, Vec_Int_t * matchedOutputs2, Vec_Int_t * oMatchedGroups, Vec_Int_t * iNonSingleton, Abc_Ntk_t * subNtk1, Abc_Ntk_t * subNtk2, Vec_Ptr_t * oMatchPairs, Vec_Int_t * oNonSingleton, int oI, int idx, int ii, int iidx)
func (pNtk1 *ABCNetwork) MatchNonSingletonOutputs(nodesInLevel1 []*PointerVector[ABCNetworkObject], iMatch1 []*IntVector, iDep1 []*IntVector, matchedInputs1 *IntVector, iGroup1 []int, oMatch1 []*IntVector, oGroup1 []int, pNtk2 *ABCNetwork, nodesInLevel2 []*PointerVector[ABCNetworkObject], iMatch2 []*IntVector, iDep2 []*IntVector, matchedInputs2 *IntVector, iGroup2 []int, oMatch2 []*IntVector, oGroup2 []int, matchedOutputs1 *IntVector, matchedOutputs2 *IntVector, oMatchedGroups *IntVector, iNonSingleton *IntVector, subNtk1 *ABCNetwork, subNtk2 *ABCNetwork, oMatchPairs *PointerVector[ABCNetworkObject], oNonSingleton *IntVector, oI int, idx int, ii int, iidx int) int {
	panic("TODO, stubbed")
}

// match1by1
// C: int match1by1(Abc_Ntk_t * pNtk1, Vec_Ptr_t ** nodesInLevel1, Vec_Int_t ** iMatch1, Vec_Int_t ** iDep1, Vec_Int_t * matchedInputs1, int * iGroup1, Vec_Int_t ** oMatch1, int * oGroup1, Abc_Ntk_t * pNtk2, Vec_Ptr_t ** nodesInLevel2, Vec_Int_t ** iMatch2, Vec_Int_t ** iDep2, Vec_Int_t * matchedInputs2, int * iGroup2, Vec_Int_t ** oMatch2, int * oGroup2, Vec_Int_t * matchedOutputs1, Vec_Int_t * matchedOutputs2, Vec_Int_t * oMatchedGroups, Vec_Int_t * iNonSingleton, int ii, int idx)
func (pNtk1 *ABCNetwork) Match1by1(nodesInLevel1 []*PointerVector[ABCNetworkObject], iMatch1 []*IntVector, iDep1 []*IntVector, matchedInputs1 *IntVector, iGroup1 []int, oMatch1 []*IntVector, oGroup1 []int, pNtk2 *ABCNetwork, nodesInLevel2 []*PointerVector[ABCNetworkObject], iMatch2 []*IntVector, iDep2 []*IntVector, matchedInputs2 *IntVector, iGroup2 []int, oMatch2 []*IntVector, oGroup2 []int, matchedOutputs1 *IntVector, matchedOutputs2 *IntVector, oMatchedGroups *IntVector, iNonSingleton *IntVector, ii int, idx int) int {
	panic("TODO, stubbed")
}

// refineBySAT
// C: float refineBySAT(Abc_Ntk_t * pNtk1, Vec_Int_t ** iMatch1, int * iGroup1, Vec_Int_t ** iDep1, int* iLastItem1, Vec_Int_t ** oMatch1, int * oGroup1, Vec_Int_t ** oDep1, int* oLastItem1, int * observability1, Abc_Ntk_t * pNtk2, Vec_Int_t ** iMatch2, int * iGroup2, Vec_Int_t ** iDep2, int* iLastItem2, Vec_Int_t ** oMatch2, int * oGroup2, Vec_Int_t ** oDep2, int* oLastItem2, int * observability2)
func (pNtk1 *ABCNetwork) RefineBySAT(iMatch1 []*IntVector, iGroup1 []int, iDep1 []*IntVector, iLastItem1 *int, oMatch1 []*IntVector, oGroup1 []int, oDep1 []*IntVector, oLastItem1 *int, observability1 []int, pNtk2 *ABCNetwork, iMatch2 []*IntVector, iGroup2 []int, iDep2 []*IntVector, iLastItem2 *int, oMatch2 []*IntVector, oGroup2 []int, oDep2 []*IntVector, oLastItem2 *int, observability2 []int) float32 {
	panic("TODO, stubbed")
}

// checkListConsistency
// C: int checkListConsistency(Vec_Int_t ** iMatch1, Vec_Int_t ** oMatch1, Vec_Int_t ** iMatch2, Vec_Int_t ** oMatch2, int iLastItem1, int oLastItem1, int iLastItem2, int oLastItem2)
func CheckListConsistency(iMatch1 []*IntVector, oMatch1 []*IntVector, iMatch2 []*IntVector, oMatch2 []*IntVector, iLastItem1 int, oLastItem1 int, iLastItem2 int, oLastItem2 int) int {
	panic("TODO, stubbed")
}

// bmGateWay
// C: void bmGateWay( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int p_equivalence )
func (pNtk1 *ABCNetwork) BmGateWay(pNtk2 *ABCNetwork, p_equivalence int) {
	panic("TODO, stubbed")
}
