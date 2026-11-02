package abc

// Hand-checked skeleton of src/aig/aig/aigPartSat.c.

// Aig_ManPartitionMonolithic
// C: Vec_Int_t * Aig_ManPartitionMonolithic( Aig_Man_t * p )
func (p *AigManager) AigManPartitionMonolithic() *IntVector {
	panic("TODO, stubbed")
}

// Aig_ManPartitionLevelized
// C: Vec_Int_t * Aig_ManPartitionLevelized( Aig_Man_t * p, int nPartSize )
func (p *AigManager) AigManPartitionLevelized(nPartSize int) *IntVector {
	panic("TODO, stubbed")
}

// Aig_ManPartitionDfs
// C: Vec_Int_t * Aig_ManPartitionDfs( Aig_Man_t * p, int nPartSize, int fPreorder )
func (p *AigManager) AigManPartitionDfs(nPartSize int, fPreorder int) *IntVector {
	panic("TODO, stubbed")
}

// Aig_ManPartSplitOne_rec
// C: void Aig_ManPartSplitOne_rec( Aig_Man_t * pNew, Aig_Man_t * p, Aig_Obj_t * pObj, Vec_Int_t * vPio2Id )
func (pNew *AigManager) AigManPartSplitOne_rec(p *AigManager, pObj *AigObject, vPio2Id *IntVector) {
	panic("TODO, stubbed")
}

// Aig_ManPartSplitOne (pvPio2Id is an out parameter)
// C: Aig_Man_t * Aig_ManPartSplitOne( Aig_Man_t * p, Vec_Ptr_t * vNodes, Vec_Int_t ** pvPio2Id )
func (p *AigManager) AigManPartSplitOne(vNodes *PointerVector[AigObject], pvPio2Id **IntVector) *AigManager {
	panic("TODO, stubbed")
}

// Aig_ManPartSplit (pvPio2Id and pvPart2Pos are out parameters; the C declares pvPart2Pos as Vec_Ptr_t ** but stores a Vec_Vec_t of Aig_Obj_t * levels)
// C: Vec_Ptr_t * Aig_ManPartSplit( Aig_Man_t * p, Vec_Int_t * vNode2Part, Vec_Ptr_t ** pvPio2Id, Vec_Ptr_t ** pvPart2Pos )
func (p *AigManager) AigManPartSplit(vNode2Part *IntVector, pvPio2Id **PointerVector[IntVector], pvPart2Pos **PointerVector[PointerVector[AigObject]]) *PointerVector[AigManager] {
	panic("TODO, stubbed")
}

// Aig_ManPartResetNodePolarity
// C: void Aig_ManPartResetNodePolarity( Aig_Man_t * pPart )
func (pPart *AigManager) AigManPartResetNodePolarity() {
	panic("TODO, stubbed")
}

// Aig_ManPartSetNodePolarity
// C: void Aig_ManPartSetNodePolarity( Aig_Man_t * p, Aig_Man_t * pPart, Vec_Int_t * vPio2Id )
func (p *AigManager) AigManPartSetNodePolarity(pPart *AigManager, vPio2Id *IntVector) {
	panic("TODO, stubbed")
}

// Aig_ManDeriveCounterExample
// C: void Aig_ManDeriveCounterExample( Aig_Man_t * p, Vec_Int_t * vNode2Var, sat_solver * pSat )
func (p *AigManager) AigManDeriveCounterExample(vNode2Var *IntVector, pSat *SatSolver) {
	panic("TODO, stubbed")
}

// Aig_ManAddNewCnfToSolver
// C: int Aig_ManAddNewCnfToSolver( sat_solver * pSat, Aig_Man_t * pAig, Vec_Int_t * vNode2Var, Vec_Int_t * vPioIds, Vec_Ptr_t * vPartPos, int fAlignPol )
func AigManAddNewCnfToSolver(pSat *SatSolver, pAig *AigManager, vNode2Var *IntVector, vPioIds *IntVector, vPartPos *PointerVector[AigObject], fAlignPol int) int {
	panic("TODO, stubbed")
}

// Aig_ManPartitionedSat
// C: int Aig_ManPartitionedSat( Aig_Man_t * p, int nAlgo, int nPartSize, int nConfPart, int nConfTotal, int fAlignPol, int fSynthesize, int fVerbose )
func (p *AigManager) AigManPartitionedSat(nAlgo int, nPartSize int, nConfPart int, nConfTotal int, fAlignPol int, fSynthesize int, fVerbose int) int {
	panic("TODO, stubbed")
}
