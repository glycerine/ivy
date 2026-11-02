package abc

// Hand-checked skeleton of src/bdd/llb/llb3Image.c.

// LlbVar3 is Llb_Var_t of this file.
type LlbVar3 struct {
	iVar   int        // variable number
	nScore int        // variable score
	vParts *IntVector // partitions
}

// LlbPrt3 is Llb_Prt_t of this file.
type LlbPrt3 struct {
	iPart int        // partition number
	nSize int        // the number of BDD nodes
	bFunc *DdNode    // the partition
	vVars *IntVector // support
}

// LlbMgr3 is Llb_Mgr_t of this file.
type LlbMgr3 struct {
	pAig    *AigManager               // AIG manager
	vLeaves *PointerVector[AigObject] // leaves in the AIG manager
	vRoots  *PointerVector[AigObject] // roots in the AIG manager
	dd      *CuddFoundation           // working BDD manager
	pVars2Q []int                     // int * : variables to quantify
	// internal
	pParts    []*LlbPrt3 // Llb_Prt_t ** : partitions
	pVars     []*LlbVar3 // Llb_Var_t ** : variables
	iPartFree int        // next free partition
	nVars     int        // the number of BDD variables
	nSuppMax  int        // maximum support size
	// temporary
	pSupp []int // int * : temporary support storage
}

// The file-level globals abctime timeBuild, timeAndEx, timeOther and int nSuppMax are not represented.

// Llb_MgrVar (static inline)
// C: Llb_Var_t * Llb_MgrVar( Llb_Mgr_t * p, int i )
func (p *LlbMgr3) llbMgrVar(i int) *LlbVar3 {
	panic("TODO, stubbed")
}

// Llb_MgrPart (static inline)
// C: Llb_Prt_t * Llb_MgrPart( Llb_Mgr_t * p, int i )
func (p *LlbMgr3) llbMgrPart(i int) *LlbPrt3 {
	panic("TODO, stubbed")
}

// Llb_NonlinRemoveVar
// C: void Llb_NonlinRemoveVar( Llb_Mgr_t * p, Llb_Var_t * pVar )
func (p *LlbMgr3) LlbNonlinRemoveVar(pVar *LlbVar3) {
	panic("TODO, stubbed")
}

// Llb_NonlinRemovePart
// C: void Llb_NonlinRemovePart( Llb_Mgr_t * p, Llb_Prt_t * pPart )
func (p *LlbMgr3) LlbNonlinRemovePart(pPart *LlbPrt3) {
	panic("TODO, stubbed")
}

// Llb_NonlinCreateCube1
// C: DdNode * Llb_NonlinCreateCube1( Llb_Mgr_t * p, Llb_Prt_t * pPart )
func (p *LlbMgr3) LlbNonlinCreateCube1(pPart *LlbPrt3) *DdNode {
	panic("TODO, stubbed")
}

// Llb_NonlinCreateCube2
// C: DdNode * Llb_NonlinCreateCube2( Llb_Mgr_t * p, Llb_Prt_t * pPart1, Llb_Prt_t * pPart2 )
func (p *LlbMgr3) LlbNonlinCreateCube2(pPart1 *LlbPrt3, pPart2 *LlbPrt3) *DdNode {
	panic("TODO, stubbed")
}

// Llb_NonlinHasSingletonVars
// C: int Llb_NonlinHasSingletonVars( Llb_Mgr_t * p, Llb_Prt_t * pPart )
func (p *LlbMgr3) LlbNonlinHasSingletonVars(pPart *LlbPrt3) int {
	panic("TODO, stubbed")
}

// Llb_NonlinPrint
// C: void Llb_NonlinPrint( Llb_Mgr_t * p )
func (p *LlbMgr3) LlbNonlinPrint() {
	panic("TODO, stubbed")
}

// Llb_NonlinQuantify1
// C: int Llb_NonlinQuantify1( Llb_Mgr_t * p, Llb_Prt_t * pPart, int fSubset )
func (p *LlbMgr3) LlbNonlinQuantify1(pPart *LlbPrt3, fSubset int) int {
	panic("TODO, stubbed")
}

// Llb_NonlinQuantify2
// C: int Llb_NonlinQuantify2( Llb_Mgr_t * p, Llb_Prt_t * pPart1, Llb_Prt_t * pPart2 )
func (p *LlbMgr3) LlbNonlinQuantify2(pPart1 *LlbPrt3, pPart2 *LlbPrt3) int {
	panic("TODO, stubbed")
}

// Llb_NonlinCutNodes_rec
// C: void Llb_NonlinCutNodes_rec( Aig_Man_t * p, Aig_Obj_t * pObj, Vec_Ptr_t * vNodes )
func (p *AigManager) LlbNonlinCutNodes_rec(pObj *AigObject, vNodes *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_NonlinCutNodes
// C: Vec_Ptr_t * Llb_NonlinCutNodes( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper )
func (p *AigManager) LlbNonlinCutNodes(vLower *PointerVector[AigObject], vUpper *PointerVector[AigObject]) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_NonlinBuildBdds
// C: Vec_Ptr_t * Llb_NonlinBuildBdds( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper, DdManager * dd )
func (p *AigManager) LlbNonlinBuildBdds(vLower *PointerVector[AigObject], vUpper *PointerVector[AigObject], dd *CuddFoundation) *PointerVector[DdNode] {
	panic("TODO, stubbed")
}

// Llb_NonlinAddPair
// C: void Llb_NonlinAddPair( Llb_Mgr_t * p, DdNode * bFunc, int iPart, int iVar )
func (p *LlbMgr3) LlbNonlinAddPair(bFunc *DdNode, iPart int, iVar int) {
	panic("TODO, stubbed")
}

// Llb_NonlinAddPartition
// C: void Llb_NonlinAddPartition( Llb_Mgr_t * p, int i, DdNode * bFunc )
func (p *LlbMgr3) LlbNonlinAddPartition(i int, bFunc *DdNode) {
	panic("TODO, stubbed")
}

// Llb_NonlinStart
// C: int Llb_NonlinStart( Llb_Mgr_t * p )
func (p *LlbMgr3) LlbNonlinStart() int {
	panic("TODO, stubbed")
}

// Llb_NonlinCheckVars
// C: void Llb_NonlinCheckVars( Llb_Mgr_t * p )
func (p *LlbMgr3) LlbNonlinCheckVars() {
	panic("TODO, stubbed")
}

// Llb_NonlinNextPartitions
// C: int Llb_NonlinNextPartitions( Llb_Mgr_t * p, Llb_Prt_t ** ppPart1, Llb_Prt_t ** ppPart2 )
func (p *LlbMgr3) LlbNonlinNextPartitions(ppPart1 **LlbPrt3, ppPart2 **LlbPrt3) int {
	panic("TODO, stubbed")
}

// Llb_NonlinReorder
// C: void Llb_NonlinReorder( DdManager * dd, int fTwice, int fVerbose )
func LlbNonlinReorder(dd *CuddFoundation, fTwice int, fVerbose int) {
	panic("TODO, stubbed")
}

// Llb_NonlinRecomputeScores
// C: void Llb_NonlinRecomputeScores( Llb_Mgr_t * p )
func (p *LlbMgr3) LlbNonlinRecomputeScores() {
	panic("TODO, stubbed")
}

// Llb_NonlinVerifyScores
// C: void Llb_NonlinVerifyScores( Llb_Mgr_t * p )
func (p *LlbMgr3) LlbNonlinVerifyScores() {
	panic("TODO, stubbed")
}

// Llb_NonlinAlloc
// C: Llb_Mgr_t * Llb_NonlinAlloc( Aig_Man_t * pAig, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vRoots, int * pVars2Q, DdManager * dd )
func (pAig *AigManager) LlbNonlinAlloc(vLeaves *PointerVector[AigObject], vRoots *PointerVector[AigObject], pVars2Q []int, dd *CuddFoundation) *LlbMgr3 {
	panic("TODO, stubbed")
}

// Llb_NonlinFree
// C: void Llb_NonlinFree( Llb_Mgr_t * p )
func (p *LlbMgr3) LlbNonlinFree() {
	panic("TODO, stubbed")
}

// Llb_NonlinImage
// C: DdNode * Llb_NonlinImage( Aig_Man_t * pAig, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vRoots, int * pVars2Q, DdManager * dd, DdNode * bCurrent, int fReorder, int fVerbose, int * pOrder )
func (pAig *AigManager) LlbNonlinImage(vLeaves *PointerVector[AigObject], vRoots *PointerVector[AigObject], pVars2Q []int, dd *CuddFoundation, bCurrent *DdNode, fReorder int, fVerbose int, pOrder []int) *DdNode {
	panic("TODO, stubbed")
}

// Llb_NonlinImageStart
// C: DdManager * Llb_NonlinImageStart( Aig_Man_t * pAig, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vRoots, int * pVars2Q, int * pOrder, int fFirst, abctime TimeTarget )
func (pAig *AigManager) LlbNonlinImageStart(vLeaves *PointerVector[AigObject], vRoots *PointerVector[AigObject], pVars2Q []int, pOrder []int, fFirst int, TimeTarget int64) *CuddFoundation {
	panic("TODO, stubbed")
}

// Llb_NonlinImageCompute
// C: DdNode * Llb_NonlinImageCompute( DdNode * bCurrent, int fReorder, int fDrop, int fVerbose, int * pOrder )
func LlbNonlinImageCompute(bCurrent *DdNode, fReorder int, fDrop int, fVerbose int, pOrder []int) *DdNode {
	panic("TODO, stubbed")
}

// Llb_NonlinImageQuit
// C: void Llb_NonlinImageQuit()
func LlbNonlinImageQuit() {
	panic("TODO, stubbed")
}
