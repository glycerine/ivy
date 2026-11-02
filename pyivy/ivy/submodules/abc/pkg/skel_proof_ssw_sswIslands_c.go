package abc

// Hand-checked skeleton of src/proof/ssw/sswIslands.c.

// Ssw_CreatePair
// C: void Ssw_CreatePair( Vec_Int_t * vPairs, Aig_Obj_t * pObj0, Aig_Obj_t * pObj1 )
func SswCreatePair(vPairs *IntVector, pObj0 *AigObject, pObj1 *AigObject) {
	panic("TODO, stubbed")
}

// Ssw_MatchingStart
// C: void Ssw_MatchingStart( Aig_Man_t * p0, Aig_Man_t * p1, Vec_Int_t * vPairs )
func (p0 *AigManager) SswMatchingStart(p1 *AigManager, vPairs *IntVector) {
	panic("TODO, stubbed")
}

// Ssw_MatchingExtendOne
// C: void Ssw_MatchingExtendOne( Aig_Man_t * p, Vec_Ptr_t * vNodes )
func (p *AigManager) SswMatchingExtendOne(vNodes *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Ssw_MatchingCountUnmached
// C: int Ssw_MatchingCountUnmached( Aig_Man_t * p )
func (p *AigManager) SswMatchingCountUnmached() int {
	panic("TODO, stubbed")
}

// Ssw_MatchingExtend
// C: void Ssw_MatchingExtend( Aig_Man_t * p0, Aig_Man_t * p1, int nDist, int fVerbose )
func (p0 *AigManager) SswMatchingExtend(p1 *AigManager, nDist int, fVerbose int) {
	panic("TODO, stubbed")
}

// Ssw_MatchingComplete
// C: void Ssw_MatchingComplete( Aig_Man_t * p0, Aig_Man_t * p1 )
func (p0 *AigManager) SswMatchingComplete(p1 *AigManager) {
	panic("TODO, stubbed")
}

// Ssw_MatchingPairs
// C: Vec_Int_t * Ssw_MatchingPairs( Aig_Man_t * p0, Aig_Man_t * p1 )
func (p0 *AigManager) SswMatchingPairs(p1 *AigManager) *IntVector {
	panic("TODO, stubbed")
}

// Ssw_MatchingMiter
// C: Vec_Int_t * Ssw_MatchingMiter( Aig_Man_t * pMiter, Aig_Man_t * p0, Aig_Man_t * p1, Vec_Int_t * vPairsAll )
func (pMiter *AigManager) SswMatchingMiter(p0 *AigManager, p1 *AigManager, vPairsAll *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Ssw_SecWithSimilaritySweep
// C: Aig_Man_t * Ssw_SecWithSimilaritySweep( Aig_Man_t * p0, Aig_Man_t * p1, Vec_Int_t * vPairs, Ssw_Pars_t * pPars )
func (p0 *AigManager) SswSecWithSimilaritySweep(p1 *AigManager, vPairs *IntVector, pPars *SswPars) *AigManager {
	panic("TODO, stubbed")
}

// Ssw_SecWithSimilarityPairs
// C: int Ssw_SecWithSimilarityPairs( Aig_Man_t * p0, Aig_Man_t * p1, Vec_Int_t * vPairs, Ssw_Pars_t * pPars )
func (p0 *AigManager) SswSecWithSimilarityPairs(p1 *AigManager, vPairs *IntVector, pPars *SswPars) int {
	panic("TODO, stubbed")
}

// Saig_StrSimPerformMatching_hack
// C: Vec_Int_t * Saig_StrSimPerformMatching_hack( Aig_Man_t * p0, Aig_Man_t * p1 )
func (p0 *AigManager) SaigStrSimPerformMatching_hack(p1 *AigManager) *IntVector {
	panic("TODO, stubbed")
}

// Ssw_SecWithSimilarity
// C: int Ssw_SecWithSimilarity( Aig_Man_t * p0, Aig_Man_t * p1, Ssw_Pars_t * pPars )
func (p0 *AigManager) SswSecWithSimilarity(p1 *AigManager, pPars *SswPars) int {
	panic("TODO, stubbed")
}
