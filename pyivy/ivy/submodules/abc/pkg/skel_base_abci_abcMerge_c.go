package abc

// Hand-checked skeleton of src/base/abci/abcMerge.c.

// NwkLMPars is Nwk_LMPars_t (src/opt/nwk/nwkMerge.h).
type NwkLMPars struct {
	nMaxLutSize   int // the max LUT size for merging (N=5)
	nMaxSuppSize  int // the max total support size after merging (S=5)
	nMaxDistance  int // the max number of nodes separating LUTs
	nMaxLevelDiff int // the max difference in levels
	nMaxFanout    int // the max number of fanouts to traverse
	fUseDiffSupp  int // enables the use of nodes with different support
	fUseTfiTfo    int // enables the use of TFO/TFO nodes as candidates
	fVeryVerbose  int // enables additional verbose output
	fVerbose      int // enables verbose output
}

// Abc_NtkMarkFanins_rec
// C: void Abc_NtkMarkFanins_rec( Abc_Obj_t * pLut, int nLevMin )
func AbcNtkMarkFanins_rec(pLut *ABCNetworkObject, nLevMin int) {
	panic("TODO, stubbed")
}

// Abc_NtkMarkFanouts_rec
// C: void Abc_NtkMarkFanouts_rec( Abc_Obj_t * pLut, int nLevMax, int nFanMax )
func AbcNtkMarkFanouts_rec(pLut *ABCNetworkObject, nLevMax int, nFanMax int) {
	panic("TODO, stubbed")
}

// Abc_NtkCollectCircle
// C: void Abc_NtkCollectCircle( Vec_Ptr_t * vStart, Vec_Ptr_t * vNext, int nFanMax )
func AbcNtkCollectCircle(vStart *PointerVector[ABCNetworkObject], vNext *PointerVector[ABCNetworkObject], nFanMax int) {
	panic("TODO, stubbed")
}

// Abc_NtkCollectNonOverlapCands
// C: void Abc_NtkCollectNonOverlapCands( Abc_Obj_t * pLut, Vec_Ptr_t * vStart, Vec_Ptr_t * vNext, Vec_Ptr_t * vCands, Nwk_LMPars_t * pPars )
func AbcNtkCollectNonOverlapCands(pLut *ABCNetworkObject, vStart *PointerVector[ABCNetworkObject], vNext *PointerVector[ABCNetworkObject], vCands *PointerVector[ABCNetworkObject], pPars *NwkLMPars) {
	panic("TODO, stubbed")
}

// Abc_NtkCountTotalFanins
// C: int Abc_NtkCountTotalFanins( Abc_Obj_t * pLut, Abc_Obj_t * pCand )
func AbcNtkCountTotalFanins(pLut *ABCNetworkObject, pCand *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkCollectOverlapCands
// C: void Abc_NtkCollectOverlapCands( Abc_Obj_t * pLut, Vec_Ptr_t * vCands, Nwk_LMPars_t * pPars )
func AbcNtkCollectOverlapCands(pLut *ABCNetworkObject, vCands *PointerVector[ABCNetworkObject], pPars *NwkLMPars) {
	panic("TODO, stubbed")
}

// Abc_NtkLutMerge
// C: Vec_Int_t * Abc_NtkLutMerge( Abc_Ntk_t * pNtk, Nwk_LMPars_t * pPars )
func (pNtk *ABCNetwork) AbcNtkLutMerge(pPars *NwkLMPars) *IntVector {
	panic("TODO, stubbed")
}
