package abc

// Hand-checked skeleton of src/bdd/llb/llb3Nonlin.c.

// LlbMnn is Llb_Mnn_t.
type LlbMnn struct {
	pInit     *AigManager            // AIG manager
	pAig      *AigManager            // AIG manager
	pPars     *GiaParLlb             // parameters
	dd        *CuddFoundation        // BDD manager
	ddG       *CuddFoundation        // BDD manager
	ddR       *CuddFoundation        // BDD manager
	vRings    *PointerVector[DdNode] // onion rings in ddR
	vLeaves   *PointerVector[AigObject]
	vRoots    *PointerVector[AigObject]
	pVars2Q   []int      // int *
	pOrderL   []int      // int *
	pOrderL2  []int      // int *
	pOrderG   []int      // int *
	vCs2Glo   *IntVector // cur state variables into global variables
	vNs2Glo   *IntVector // next state variables into global variables
	vGlo2Cs   *IntVector // global variables into cur state variables
	vGlo2Ns   *IntVector // global variables into next state variables
	ddLocReos int
	ddLocGrbs int
	timeImage int64 // abctime
	timeTran1 int64 // abctime
	timeTran2 int64 // abctime
	timeGloba int64 // abctime
	timeOther int64 // abctime
	timeTotal int64 // abctime
	timeReo   int64 // abctime
	timeReoG  int64 // abctime
}

// The file-level declarations extern abctime timeBuild, timeAndEx, timeOther and extern int nSuppMax
// refer to the globals defined in llb3Image.c.

// Llb_NonlinFindBestVar
// C: int Llb_NonlinFindBestVar( DdManager * dd, DdNode * bFunc, Aig_Man_t * pAig )
func LlbNonlinFindBestVar(dd *CuddFoundation, bFunc *DdNode, pAig *AigManager) int {
	panic("TODO, stubbed")
}

// Llb_NonlinTrySubsetting
// C: void Llb_NonlinTrySubsetting( DdManager * dd, DdNode * bFunc )
func LlbNonlinTrySubsetting(dd *CuddFoundation, bFunc *DdNode) {
	panic("TODO, stubbed")
}

// Llb_NonlinPrepareVarMap
// C: void Llb_NonlinPrepareVarMap( Llb_Mnn_t * p )
func (p *LlbMnn) LlbNonlinPrepareVarMap() {
	panic("TODO, stubbed")
}

// Llb_NonlinComputeInitState
// C: DdNode * Llb_NonlinComputeInitState( Aig_Man_t * pAig, DdManager * dd )
func (pAig *AigManager) LlbNonlinComputeInitState(dd *CuddFoundation) *DdNode {
	panic("TODO, stubbed")
}

// Llb_NonlinDeriveCex
// C: Abc_Cex_t * Llb_NonlinDeriveCex( Llb_Mnn_t * p )
func (p *LlbMnn) LlbNonlinDeriveCex() *Counterexample {
	panic("TODO, stubbed")
}

// Llb_NonlinReoHook (Method is a void *)
// C: int Llb_NonlinReoHook( DdManager * dd, char * Type, void * Method )
func LlbNonlinReoHook(dd *CuddFoundation, Type string, Method any) int {
	panic("TODO, stubbed")
}

// Llb_NonlinCompPerms
// C: int Llb_NonlinCompPerms( DdManager * dd, int * pVar2Lev )
func LlbNonlinCompPerms(dd *CuddFoundation, pVar2Lev []int) int {
	panic("TODO, stubbed")
}

// Llb_NonlinReachability
// C: int Llb_NonlinReachability( Llb_Mnn_t * p )
func (p *LlbMnn) LlbNonlinReachability() int {
	panic("TODO, stubbed")
}

// Llb_MnnStart
// C: Llb_Mnn_t * Llb_MnnStart( Aig_Man_t * pInit, Aig_Man_t * pAig, Gia_ParLlb_t * pPars )
func (pInit *AigManager) LlbMnnStart(pAig *AigManager, pPars *GiaParLlb) *LlbMnn {
	panic("TODO, stubbed")
}

// Llb_MnnStop
// C: void Llb_MnnStop( Llb_Mnn_t * p )
func (p *LlbMnn) LlbMnnStop() {
	panic("TODO, stubbed")
}

// Llb_NonlinExperiment
// C: void Llb_NonlinExperiment( Aig_Man_t * pAig, int Num )
func (pAig *AigManager) LlbNonlinExperiment(Num int) {
	panic("TODO, stubbed")
}

// Llb_NonlinCoreReach
// C: int Llb_NonlinCoreReach( Aig_Man_t * pAig, Gia_ParLlb_t * pPars )
func (pAig *AigManager) LlbNonlinCoreReach(pPars *GiaParLlb) int {
	panic("TODO, stubbed")
}
