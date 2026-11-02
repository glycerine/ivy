package abc

// Hand-checked skeleton of src/bdd/llb/llb4Image.c.

// LlbVar4 is Llb_Var_t of this file.
type LlbVar4 struct {
	iVar   int        // variable number
	nScore int        // variable score
	vParts *IntVector // partitions
}

// LlbPrt4 is Llb_Prt_t of this file.
type LlbPrt4 struct {
	iPart int        // partition number
	nSize int        // the number of BDD nodes
	bFunc *DdNode    // the partition
	vVars *IntVector // support
}

// LlbMgr4 is Llb_Mgr_t of this file.
type LlbMgr4 struct {
	dd       *CuddFoundation // working BDD manager
	vVars2Q  *IntVector      // variables to quantify
	nSizeMax int             // maximum size of the cluster
	// internal
	pParts    []*LlbPrt4 // Llb_Prt_t ** : partitions
	pVars     []*LlbVar4 // Llb_Var_t ** : variables
	iPartFree int        // next free partition
	nVars     int        // the number of BDD variables
	nSuppMax  int        // maximum support size
	// temporary
	pSupp []int // int * : temporary support storage
}

// Llb_MgrVar (static inline)
// C: Llb_Var_t * Llb_MgrVar( Llb_Mgr_t * p, int i )
func (p *LlbMgr4) llbMgrVar(i int) *LlbVar4 {
	panic("TODO, stubbed")
}

// Llb_MgrPart (static inline)
// C: Llb_Prt_t * Llb_MgrPart( Llb_Mgr_t * p, int i )
func (p *LlbMgr4) llbMgrPart(i int) *LlbPrt4 {
	panic("TODO, stubbed")
}

// Llb_Nonlin4RemoveVar
// C: void Llb_Nonlin4RemoveVar( Llb_Mgr_t * p, Llb_Var_t * pVar )
func (p *LlbMgr4) LlbNonlin4RemoveVar(pVar *LlbVar4) {
	panic("TODO, stubbed")
}

// Llb_Nonlin4RemovePart
// C: void Llb_Nonlin4RemovePart( Llb_Mgr_t * p, Llb_Prt_t * pPart )
func (p *LlbMgr4) LlbNonlin4RemovePart(pPart *LlbPrt4) {
	panic("TODO, stubbed")
}

// Llb_Nonlin4CreateCube1
// C: DdNode * Llb_Nonlin4CreateCube1( Llb_Mgr_t * p, Llb_Prt_t * pPart )
func (p *LlbMgr4) LlbNonlin4CreateCube1(pPart *LlbPrt4) *DdNode {
	panic("TODO, stubbed")
}

// Llb_Nonlin4CreateCube2
// C: DdNode * Llb_Nonlin4CreateCube2( Llb_Mgr_t * p, Llb_Prt_t * pPart1, Llb_Prt_t * pPart2 )
func (p *LlbMgr4) LlbNonlin4CreateCube2(pPart1 *LlbPrt4, pPart2 *LlbPrt4) *DdNode {
	panic("TODO, stubbed")
}

// Llb_Nonlin4HasSingletonVars
// C: int Llb_Nonlin4HasSingletonVars( Llb_Mgr_t * p, Llb_Prt_t * pPart )
func (p *LlbMgr4) LlbNonlin4HasSingletonVars(pPart *LlbPrt4) int {
	panic("TODO, stubbed")
}

// Llb_Nonlin4Print
// C: void Llb_Nonlin4Print( Llb_Mgr_t * p )
func (p *LlbMgr4) LlbNonlin4Print() {
	panic("TODO, stubbed")
}

// Llb_Nonlin4Quantify1
// C: int Llb_Nonlin4Quantify1( Llb_Mgr_t * p, Llb_Prt_t * pPart )
func (p *LlbMgr4) LlbNonlin4Quantify1(pPart *LlbPrt4) int {
	panic("TODO, stubbed")
}

// Llb_Nonlin4Quantify2
// C: int Llb_Nonlin4Quantify2( Llb_Mgr_t * p, Llb_Prt_t * pPart1, Llb_Prt_t * pPart2 )
func (p *LlbMgr4) LlbNonlin4Quantify2(pPart1 *LlbPrt4, pPart2 *LlbPrt4) int {
	panic("TODO, stubbed")
}

// Llb_Nonlin4CutNodes_rec
// C: void Llb_Nonlin4CutNodes_rec( Aig_Man_t * p, Aig_Obj_t * pObj, Vec_Ptr_t * vNodes )
func (p *AigManager) LlbNonlin4CutNodes_rec(pObj *AigObject, vNodes *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_Nonlin4CutNodes
// C: Vec_Ptr_t * Llb_Nonlin4CutNodes( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper )
func (p *AigManager) LlbNonlin4CutNodes(vLower *PointerVector[AigObject], vUpper *PointerVector[AigObject]) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_Nonlin4AddPair
// C: void Llb_Nonlin4AddPair( Llb_Mgr_t * p, int iPart, int iVar )
func (p *LlbMgr4) LlbNonlin4AddPair(iPart int, iVar int) {
	panic("TODO, stubbed")
}

// Llb_Nonlin4AddPartition
// C: void Llb_Nonlin4AddPartition( Llb_Mgr_t * p, int i, DdNode * bFunc )
func (p *LlbMgr4) LlbNonlin4AddPartition(i int, bFunc *DdNode) {
	panic("TODO, stubbed")
}

// Llb_Nonlin4CheckVars
// C: void Llb_Nonlin4CheckVars( Llb_Mgr_t * p )
func (p *LlbMgr4) LlbNonlin4CheckVars() {
	panic("TODO, stubbed")
}

// Llb_Nonlin4NextPartitions
// C: int Llb_Nonlin4NextPartitions( Llb_Mgr_t * p, Llb_Prt_t ** ppPart1, Llb_Prt_t ** ppPart2 )
func (p *LlbMgr4) LlbNonlin4NextPartitions(ppPart1 **LlbPrt4, ppPart2 **LlbPrt4) int {
	panic("TODO, stubbed")
}

// Llb_Nonlin4RecomputeScores
// C: void Llb_Nonlin4RecomputeScores( Llb_Mgr_t * p )
func (p *LlbMgr4) LlbNonlin4RecomputeScores() {
	panic("TODO, stubbed")
}

// Llb_Nonlin4VerifyScores
// C: void Llb_Nonlin4VerifyScores( Llb_Mgr_t * p )
func (p *LlbMgr4) LlbNonlin4VerifyScores() {
	panic("TODO, stubbed")
}

// Llb_Nonlin4Alloc
// C: Llb_Mgr_t * Llb_Nonlin4Alloc( DdManager * dd, Vec_Ptr_t * vParts, DdNode * bCurrent, Vec_Int_t * vVars2Q, int nSizeMax )
func LlbNonlin4Alloc(dd *CuddFoundation, vParts *PointerVector[DdNode], bCurrent *DdNode, vVars2Q *IntVector, nSizeMax int) *LlbMgr4 {
	panic("TODO, stubbed")
}

// Llb_Nonlin4Free
// C: void Llb_Nonlin4Free( Llb_Mgr_t * p )
func (p *LlbMgr4) LlbNonlin4Free() {
	panic("TODO, stubbed")
}

// Llb_Nonlin4Image
// C: DdNode * Llb_Nonlin4Image( DdManager * dd, Vec_Ptr_t * vParts, DdNode * bCurrent, Vec_Int_t * vVars2Q )
func LlbNonlin4Image(dd *CuddFoundation, vParts *PointerVector[DdNode], bCurrent *DdNode, vVars2Q *IntVector) *DdNode {
	panic("TODO, stubbed")
}

// Llb_Nonlin4Group
// C: Vec_Ptr_t * Llb_Nonlin4Group( DdManager * dd, Vec_Ptr_t * vParts, Vec_Int_t * vVars2Q, int nSizeMax )
func LlbNonlin4Group(dd *CuddFoundation, vParts *PointerVector[DdNode], vVars2Q *IntVector, nSizeMax int) *PointerVector[DdNode] {
	panic("TODO, stubbed")
}
