package abc

// Hand-checked skeleton of src/bdd/llb/llb2Core.c.

// LlbImg is Llb_Img_t.
type LlbImg struct {
	pInit    *AigManager                    // AIG manager
	pAig     *AigManager                    // AIG manager
	pPars    *GiaParLlb                     // parameters
	dd       *CuddFoundation                // BDD manager
	ddG      *CuddFoundation                // BDD manager
	ddR      *CuddFoundation                // BDD manager
	vDdMans  *PointerVector[CuddFoundation] // BDD managers for each partition
	vRings   *PointerVector[DdNode]         // onion rings in ddR
	vDriRefs *IntVector                     // driver references
	vVarsCs  *IntVector                     // cur state variables
	vVarsNs  *IntVector                     // next state variables
	vCs2Glo  *IntVector                     // cur state variables into global variables
	vNs2Glo  *IntVector                     // next state variables into global variables
	vGlo2Cs  *IntVector                     // global variables into cur state variables
	vGlo2Ns  *IntVector                     // global variables into next state variables
}

// Llb_CoreComputeCube
// C: DdNode * Llb_CoreComputeCube( DdManager * dd, Vec_Int_t * vVars, int fUseVarIndex, char * pValues )
func LlbCoreComputeCube(dd *CuddFoundation, vVars *IntVector, fUseVarIndex int, pValues []int8) *DdNode {
	panic("TODO, stubbed")
}

// Llb_CoreDeriveCex
// C: Abc_Cex_t * Llb_CoreDeriveCex( Llb_Img_t * p )
func (p *LlbImg) LlbCoreDeriveCex() *Counterexample {
	panic("TODO, stubbed")
}

// Llb_CoreReachability_int
// C: int Llb_CoreReachability_int( Llb_Img_t * p, Vec_Ptr_t * vQuant0, Vec_Ptr_t * vQuant1 )
func (p *LlbImg) LlbCoreReachability_int(vQuant0 *PointerVector[IntVector], vQuant1 *PointerVector[IntVector]) int {
	panic("TODO, stubbed")
}

// Llb_CoreReachability
// C: int Llb_CoreReachability( Llb_Img_t * p )
func (p *LlbImg) LlbCoreReachability() int {
	panic("TODO, stubbed")
}

// Llb_CoreConstructAll
// C: Vec_Ptr_t * Llb_CoreConstructAll( Aig_Man_t * p, Vec_Ptr_t * vResult, Vec_Int_t * vVarsNs, abctime TimeTarget )
func (p *AigManager) LlbCoreConstructAll(vResult *PointerVector[PointerVector[AigObject]], vVarsNs *IntVector, TimeTarget int64) *PointerVector[CuddFoundation] {
	panic("TODO, stubbed")
}

// Llb_CoreSetVarMaps
// C: void Llb_CoreSetVarMaps( Llb_Img_t * p )
func (p *LlbImg) LlbCoreSetVarMaps() {
	panic("TODO, stubbed")
}

// Llb_CoreStart
// C: Llb_Img_t * Llb_CoreStart( Aig_Man_t * pInit, Aig_Man_t * pAig, Gia_ParLlb_t * pPars )
func (pInit *AigManager) LlbCoreStart(pAig *AigManager, pPars *GiaParLlb) *LlbImg {
	panic("TODO, stubbed")
}

// Llb_CoreStop
// C: void Llb_CoreStop( Llb_Img_t * p )
func (p *LlbImg) LlbCoreStop() {
	panic("TODO, stubbed")
}

// Llb_CoreExperiment
// C: int Llb_CoreExperiment( Aig_Man_t * pInit, Aig_Man_t * pAig, Gia_ParLlb_t * pPars, Vec_Ptr_t * vResult, abctime TimeTarget )
func (pInit *AigManager) LlbCoreExperiment(pAig *AigManager, pPars *GiaParLlb, vResult *PointerVector[PointerVector[AigObject]], TimeTarget int64) int {
	panic("TODO, stubbed")
}

// Llb_ManReachMinCut
// C: int Llb_ManReachMinCut( Aig_Man_t * pAig, Gia_ParLlb_t * pPars )
func (pAig *AigManager) LlbManReachMinCut(pPars *GiaParLlb) int {
	panic("TODO, stubbed")
}
