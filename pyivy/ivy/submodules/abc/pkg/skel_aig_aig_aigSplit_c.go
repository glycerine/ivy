package abc

// Hand-checked skeleton of src/aig/aig/aigSplit.c.

// Aig_NodeBddToMuxes_rec
// C: Aig_Obj_t * Aig_NodeBddToMuxes_rec( DdManager * dd, DdNode * bFunc, Aig_Man_t * pNew, st__table * tBdd2Node )
func AigNodeBddToMuxes_rec(dd *CuddFoundation, bFunc *DdNode, pNew *AigManager, tBdd2Node *StableTable[*DdNode, *AigObject]) *AigObject {
	panic("TODO, stubbed")
}

// Aig_ManConvertBddsToAigs
// C: Aig_Man_t * Aig_ManConvertBddsToAigs( Aig_Man_t * p, DdManager * dd, Vec_Ptr_t * vCofs )
func (p *AigManager) AigManConvertBddsToAigs(dd *CuddFoundation, vCofs *PointerVector[DdNode]) *AigManager {
	panic("TODO, stubbed")
}

// Aig_ManBuildPoBdd_rec
// C: DdNode * Aig_ManBuildPoBdd_rec( Aig_Man_t * p, Aig_Obj_t * pObj, DdManager * dd )
func (p *AigManager) AigManBuildPoBdd_rec(pObj *AigObject, dd *CuddFoundation) *DdNode {
	panic("TODO, stubbed")
}

// Aig_ManCofactorBdds (vSubset holds the BDD variables (pData of the support nodes))
// C: Vec_Ptr_t * Aig_ManCofactorBdds( Aig_Man_t * p, Vec_Ptr_t * vSubset, DdManager * dd, DdNode * bFunc )
func (p *AigManager) AigManCofactorBdds(vSubset *PointerVector[DdNode], dd *CuddFoundation, bFunc *DdNode) *PointerVector[DdNode] {
	panic("TODO, stubbed")
}

// Aig_ManBuildPoBdd (pbFunc is an out parameter)
// C: DdManager * Aig_ManBuildPoBdd( Aig_Man_t * p, DdNode ** pbFunc )
func (p *AigManager) AigManBuildPoBdd(pbFunc **DdNode) *CuddFoundation {
	panic("TODO, stubbed")
}

// Aig_ManVecRandSubset
// C: Vec_Ptr_t * Aig_ManVecRandSubset( Vec_Ptr_t * vVec, int nVars )
func AigManVecRandSubset(vVec *PointerVector[AigObject], nVars int) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Aig_ManSplit
// C: Aig_Man_t * Aig_ManSplit( Aig_Man_t * p, int nVars, int fVerbose )
func (p *AigManager) AigManSplit(nVars int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}
