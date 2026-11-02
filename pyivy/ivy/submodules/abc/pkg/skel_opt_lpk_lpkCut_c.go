package abc

// Hand-checked skeleton of src/opt/lpk/lpkCut.c.

// Lpk_CutTruthBdd_rec (CloudNode * is represented by the validated port's CloudEdge handle)
// C: CloudNode * Lpk_CutTruthBdd_rec( CloudManager * dd, Hop_Man_t * pMan, Hop_Obj_t * pObj, int nVars )
func LpkCutTruthBdd_rec(dd *CloudManager, pMan *HopManager, pObj *HopObject, nVars int) CloudEdge {
	panic("TODO, stubbed")
}

// Lpk_CutTruthBdd (CloudNode * is represented by the validated port's CloudEdge handle)
// C: CloudNode * Lpk_CutTruthBdd( Lpk_Man_t * p, Lpk_Cut_t * pCut )
func (p *LpkMan) LpkCutTruthBdd(pCut *LpkCut) CloudEdge {
	panic("TODO, stubbed")
}

// Lpk_CutTruth_rec (piCount is an in/out counter; returns a pointer to a truth table)
// C: unsigned * Lpk_CutTruth_rec( Hop_Man_t * pMan, Hop_Obj_t * pObj, int nVars, Vec_Ptr_t * vTtNodes, int * piCount )
func LpkCutTruth_rec(pMan *HopManager, pObj *HopObject, nVars int, vTtNodes *PointerVector[uint32], piCount *int) []uint32 {
	panic("TODO, stubbed")
}

// Lpk_CutTruth (returns a pointer to a truth table)
// C: unsigned * Lpk_CutTruth( Lpk_Man_t * p, Lpk_Cut_t * pCut, int fInv )
func (p *LpkMan) LpkCutTruth(pCut *LpkCut, fInv int) []uint32 {
	panic("TODO, stubbed")
}

// Lpk_NodeRecordImpact
// C: void Lpk_NodeRecordImpact( Lpk_Man_t * p )
func (p *LpkMan) LpkNodeRecordImpact() {
	panic("TODO, stubbed")
}

// Lpk_NodeCutsCheckDsd
// C: int Lpk_NodeCutsCheckDsd( Lpk_Man_t * p, Lpk_Cut_t * pCut )
func (p *LpkMan) LpkNodeCutsCheckDsd(pCut *LpkCut) int {
	panic("TODO, stubbed")
}

// Lpk_NodeCutsOneDominance (static)
// C: int Lpk_NodeCutsOneDominance( Lpk_Cut_t * pDom, Lpk_Cut_t * pCut )
func lpkNodeCutsOneDominance(pDom *LpkCut, pCut *LpkCut) int {
	panic("TODO, stubbed")
}

// Lpk_NodeCutsOneFilter
// C: int Lpk_NodeCutsOneFilter( Lpk_Cut_t * pCuts, int nCuts, Lpk_Cut_t * pCutNew )
func LpkNodeCutsOneFilter(pCuts []LpkCut, nCuts int, pCutNew *LpkCut) int {
	panic("TODO, stubbed")
}

// Lpk_NodePrintCut
// C: void Lpk_NodePrintCut( Lpk_Man_t * p, Lpk_Cut_t * pCut, int fLeavesOnly )
func (p *LpkMan) LpkNodePrintCut(pCut *LpkCut, fLeavesOnly int) {
	panic("TODO, stubbed")
}

// Lpk_NodeCutSignature
// C: void Lpk_NodeCutSignature( Lpk_Cut_t * pCut )
func LpkNodeCutSignature(pCut *LpkCut) {
	panic("TODO, stubbed")
}

// Lpk_NodeCutsOne
// C: void Lpk_NodeCutsOne( Lpk_Man_t * p, Lpk_Cut_t * pCut, int Node )
func (p *LpkMan) LpkNodeCutsOne(pCut *LpkCut, Node int) {
	panic("TODO, stubbed")
}

// Lpk_CountSupp
// C: int Lpk_CountSupp( Abc_Ntk_t * p, Vec_Ptr_t * vNodes )
func LpkCountSupp(p *ABCNetwork, vNodes *PointerVector[ABCNetworkObject]) int {
	panic("TODO, stubbed")
}

// Lpk_NodeCuts
// C: int Lpk_NodeCuts( Lpk_Man_t * p )
func (p *LpkMan) LpkNodeCuts() int {
	panic("TODO, stubbed")
}
