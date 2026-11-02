package abc

// Hand-checked skeleton of src/bdd/llb/llb4Nonlin.c.

// LlbMnx is Llb_Mnx_t.
type LlbMnx struct {
	// user info
	pAig  *AigManager // AIG manager
	pPars *GiaParLlb  // parameters
	// intermediate BDDs
	dd       *CuddFoundation        // BDD manager
	bBad     *DdNode                // bad states in terms of CIs
	bReached *DdNode                // reached states
	bCurrent *DdNode                // from states
	bNext    *DdNode                // to states
	vRings   *PointerVector[DdNode] // onion rings in ddR
	vRoots   *PointerVector[DdNode] // BDDs for partitions
	// structural info
	vOrder    *IntVector // for each object ID, its BDD variable number or -1
	vVars2Q   *IntVector // 1 if variable is quantifiable; 0 othervise
	timeImage int64      // abctime
	timeRemap int64      // abctime
	timeReo   int64      // abctime
	timeOther int64      // abctime
	timeTotal int64      // abctime
}

// Llb_Nonlin4ComputeBad
// C: DdNode * Llb_Nonlin4ComputeBad( DdManager * dd, Aig_Man_t * pAig, Vec_Int_t * vOrder )
func LlbNonlin4ComputeBad(dd *CuddFoundation, pAig *AigManager, vOrder *IntVector) *DdNode {
	panic("TODO, stubbed")
}

// Llb_Nonlin4DerivePartitions
// C: Vec_Ptr_t * Llb_Nonlin4DerivePartitions( DdManager * dd, Aig_Man_t * pAig, Vec_Int_t * vOrder )
func LlbNonlin4DerivePartitions(dd *CuddFoundation, pAig *AigManager, vOrder *IntVector) *PointerVector[DdNode] {
	panic("TODO, stubbed")
}

// Llb_Nonlin4CreateOrderSimple
// C: Vec_Int_t * Llb_Nonlin4CreateOrderSimple( Aig_Man_t * pAig )
func (pAig *AigManager) LlbNonlin4CreateOrderSimple() *IntVector {
	panic("TODO, stubbed")
}

// Llb_Nonlin4CreateOrder_rec
// C: void Llb_Nonlin4CreateOrder_rec( Aig_Man_t * pAig, Aig_Obj_t * pObj, Vec_Int_t * vOrder, int * pCounter )
func (pAig *AigManager) LlbNonlin4CreateOrder_rec(pObj *AigObject, vOrder *IntVector, pCounter *int) {
	panic("TODO, stubbed")
}

// Llb_Nonlin4CollectHighRefNodes
// C: Vec_Int_t * Llb_Nonlin4CollectHighRefNodes( Aig_Man_t * pAig, int nFans )
func (pAig *AigManager) LlbNonlin4CollectHighRefNodes(nFans int) *IntVector {
	panic("TODO, stubbed")
}

// Llb_Nonlin4CreateOrder
// C: Vec_Int_t * Llb_Nonlin4CreateOrder( Aig_Man_t * pAig )
func (pAig *AigManager) LlbNonlin4CreateOrder() *IntVector {
	panic("TODO, stubbed")
}

// Llb_Nonlin4CreateVars2Q
// C: Vec_Int_t * Llb_Nonlin4CreateVars2Q( DdManager * dd, Aig_Man_t * pAig, Vec_Int_t * vOrder, int fBackward )
func LlbNonlin4CreateVars2Q(dd *CuddFoundation, pAig *AigManager, vOrder *IntVector, fBackward int) *IntVector {
	panic("TODO, stubbed")
}

// Llb_Nonlin4SetupVarMap
// C: void Llb_Nonlin4SetupVarMap( DdManager * dd, Aig_Man_t * pAig, Vec_Int_t * vOrder )
func LlbNonlin4SetupVarMap(dd *CuddFoundation, pAig *AigManager, vOrder *IntVector) {
	panic("TODO, stubbed")
}

// Llb_Nonlin4ComputeInitState
// C: DdNode * Llb_Nonlin4ComputeInitState( DdManager * dd, Aig_Man_t * pAig, Vec_Int_t * vOrder, int fBackward )
func LlbNonlin4ComputeInitState(dd *CuddFoundation, pAig *AigManager, vOrder *IntVector, fBackward int) *DdNode {
	panic("TODO, stubbed")
}

// Llb_Nonlin4ComputeCube
// C: DdNode * Llb_Nonlin4ComputeCube( DdManager * dd, Aig_Man_t * pAig, Vec_Int_t * vOrder, char * pValues, int Flag )
func LlbNonlin4ComputeCube(dd *CuddFoundation, pAig *AigManager, vOrder *IntVector, pValues []int8, Flag int) *DdNode {
	panic("TODO, stubbed")
}

// Llb_Nonlin4RecordState
// C: void Llb_Nonlin4RecordState( Aig_Man_t * pAig, Vec_Int_t * vOrder, unsigned * pState, char * pValues, int fBackward )
func (pAig *AigManager) LlbNonlin4RecordState(vOrder *IntVector, pState []uint32, pValues []int8, fBackward int) {
	panic("TODO, stubbed")
}

// Llb_Nonlin4Multiply
// C: Vec_Ptr_t * Llb_Nonlin4Multiply( DdManager * dd, DdNode * bCube, Vec_Ptr_t * vParts )
func LlbNonlin4Multiply(dd *CuddFoundation, bCube *DdNode, vParts *PointerVector[DdNode]) *PointerVector[DdNode] {
	panic("TODO, stubbed")
}

// Llb_Nonlin4Deref
// C: void Llb_Nonlin4Deref( DdManager * dd, Vec_Ptr_t * vParts )
func LlbNonlin4Deref(dd *CuddFoundation, vParts *PointerVector[DdNode]) {
	panic("TODO, stubbed")
}

// Llb_Nonlin4DeriveCex
// C: Vec_Ptr_t * Llb_Nonlin4DeriveCex( Llb_Mnx_t * p, int fBackward, int fVerbose )
func (p *LlbMnx) LlbNonlin4DeriveCex(fBackward int, fVerbose int) *PointerVector[uint32] {
	panic("TODO, stubbed")
}

// Llb_Nonlin4Reachability
// C: int Llb_Nonlin4Reachability( Llb_Mnx_t * p )
func (p *LlbMnx) LlbNonlin4Reachability() int {
	panic("TODO, stubbed")
}

// Llb_Nonlin4Reorder
// C: void Llb_Nonlin4Reorder( DdManager * dd, int fTwice, int fVerbose )
func LlbNonlin4Reorder(dd *CuddFoundation, fTwice int, fVerbose int) {
	panic("TODO, stubbed")
}

// Llb_MnxStart
// C: Llb_Mnx_t * Llb_MnxStart( Aig_Man_t * pAig, Gia_ParLlb_t * pPars )
func (pAig *AigManager) LlbMnxStart(pPars *GiaParLlb) *LlbMnx {
	panic("TODO, stubbed")
}

// Llb_MnxStop
// C: void Llb_MnxStop( Llb_Mnx_t * p )
func (p *LlbMnx) LlbMnxStop() {
	panic("TODO, stubbed")
}

// Llb_MnxCheckNextStateVars
// C: void Llb_MnxCheckNextStateVars( Llb_Mnx_t * p )
func (p *LlbMnx) LlbMnxCheckNextStateVars() {
	panic("TODO, stubbed")
}

// Llb_Nonlin4CoreReach
// C: int Llb_Nonlin4CoreReach( Aig_Man_t * pAig, Gia_ParLlb_t * pPars )
func (pAig *AigManager) LlbNonlin4CoreReach(pPars *GiaParLlb) int {
	panic("TODO, stubbed")
}

// Llb_ReachableStates
// C: Aig_Man_t * Llb_ReachableStates( Aig_Man_t * pAig )
func (pAig *AigManager) LlbReachableStates() *AigManager {
	panic("TODO, stubbed")
}

// Llb_ReachableStatesGia
// C: Gia_Man_t * Llb_ReachableStatesGia( Gia_Man_t * p )
func (p *GiaManager) LlbReachableStatesGia() *GiaManager {
	panic("TODO, stubbed")
}
