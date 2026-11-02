package abc

// Hand-checked skeleton of src/bool/kit/kitBdd.c.

// Kit_SopToBdd
// C: DdNode * Kit_SopToBdd( DdManager * dd, Kit_Sop_t * cSop, int nVars )
func KitSopToBdd(dd *CuddFoundation, cSop *KitSop, nVars int) *DdNode {
	panic("TODO, stubbed")
}

// Kit_GraphToBdd
// C: DdNode * Kit_GraphToBdd( DdManager * dd, Kit_Graph_t * pGraph )
func KitGraphToBdd(dd *CuddFoundation, pGraph *KitGraph) *DdNode {
	panic("TODO, stubbed")
}

// Kit_TruthToBdd_rec
// C: DdNode * Kit_TruthToBdd_rec( DdManager * dd, unsigned * pTruth, int iBit, int nVars, int nVarsTotal, int fMSBonTop )
func KitTruthToBdd_rec(dd *CuddFoundation, pTruth []uint32, iBit int, nVars int, nVarsTotal int, fMSBonTop int) *DdNode {
	panic("TODO, stubbed")
}

// Kit_TruthToBdd
// C: DdNode * Kit_TruthToBdd( DdManager * dd, unsigned * pTruth, int nVars, int fMSBonTop )
func KitTruthToBdd(dd *CuddFoundation, pTruth []uint32, nVars int, fMSBonTop int) *DdNode {
	panic("TODO, stubbed")
}

// Kit_SopFactorVerify
// C: int Kit_SopFactorVerify( Vec_Int_t * vCover, Kit_Graph_t * pFForm, int nVars )
func KitSopFactorVerify(vCover *IntVector, pFForm *KitGraph, nVars int) int {
	panic("TODO, stubbed")
}
