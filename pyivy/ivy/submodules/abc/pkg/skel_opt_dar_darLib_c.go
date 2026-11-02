package abc

// Hand-checked skeleton of src/opt/dar/darLib.c.

// DarLibObj is Dar_LibObj_t: library object (2 words). Go has no bitfields;
// each is a uint32 whose C width is in the comment.
type DarLibObj struct {
	Fan0    uint32 // unsigned : 16, the first fanin
	Fan1    uint32 // unsigned : 16, the second fanin
	fCompl0 uint32 // unsigned : 1, the first compl attribute
	fCompl1 uint32 // unsigned : 1, the second compl attribute
	fPhase  uint32 // unsigned : 1, the phase of the node
	fTerm   uint32 // unsigned : 1, indicates a PI
	Num     uint32 // unsigned : 28, internal use
}

// DarLibDat is Dar_LibDat_t: library object data. The anonymous union
// { Aig_Obj_t * pFunc; int iGunc; } is represented by both members.
type DarLibDat struct {
	pFunc  *AigObject // union member : the corresponding AIG node if it exists
	iGunc  int        // union member : the corresponding AIG node if it exists
	Level  int        // level of this node after it is constructured
	TravId int        // traversal ID of the library object data
	dProb  float32    // probability of the node being 1
	fMffc  uint8      // unsigned char : set to one if node is part of MFFC
	nLats  [3]uint8   // unsigned char nLats[3] : the number of latches on the input/output stem
}

// DarLib is Dar_Lib_t: library.
type DarLib struct {
	// objects
	pObjs []DarLibObj // Dar_LibObj_t * : the set of library objects
	nObjs int         // the number of objects used
	iObj  int         // the current object
	// structures by class
	nSubgr      [222]int   // the number of subgraphs by class
	pSubgr      [222][]int // int * pSubgr[222] : the subgraphs for each class
	pSubgrMem   []int      // int * : memory for subgraph pointers
	nSubgrTotal int        // the total number of subgraph
	// structure priorities
	pPriosMem []int      // int * : memory for priority of structures
	pPrios    [222][]int // int * pPrios[222] : pointers to the priority numbers
	// structure places in the priorities
	pPlaceMem []int      // int * : memory for places of structures in the priority lists
	pPlace    [222][]int // int * pPlace[222] : pointers to the places numbers
	// structure scores
	pScoreMem []int      // int * : memory for scores of structures
	pScore    [222][]int // int * pScore[222] : pointers to the scores numbers
	// nodes by class
	nNodes      [222]int   // the number of nodes by class
	pNodes      [222][]int // int * pNodes[222] : the nodes for each class
	pNodesMem   []int      // int * : memory for nodes pointers
	nNodesTotal int        // the total number of nodes
	// prepared library
	nSubgraphs int
	nNodes0Max int
	// nodes by class
	nNodes0      [222]int   // the number of nodes by class
	pNodes0      [222][]int // int * pNodes0[222] : the nodes for each class
	pNodes0Mem   []int      // int * : memory for nodes pointers
	nNodes0Total int        // the total number of nodes
	// structures by class
	nSubgr0      [222]int   // the number of subgraphs by class
	pSubgr0      [222][]int // int * pSubgr0[222] : the subgraphs for each class
	pSubgr0Mem   []int      // int * : memory for subgraph pointers
	nSubgr0Total int        // the total number of subgraph
	// object data
	pDatas []DarLibDat // Dar_LibDat_t * : an array of nDatas
	nDatas int
	// information about NPN classes
	pPerms4  [][]int8 // char ** pPerms4
	puCanons []uint16 // unsigned short *
	pPhases  []int8   // char *
	pPerms   []int8   // char *
	pMap     []uint8  // unsigned char *
}

// Dar_LibObj (static inline)
// C: Dar_LibObj_t * Dar_LibObj( Dar_Lib_t * p, int Id )
func (p *DarLib) darLibObj(Id int) *DarLibObj {
	panic("TODO, stubbed")
}

// Dar_LibObjTruth (static inline)
// C: int Dar_LibObjTruth( Dar_LibObj_t * pObj )
func darLibObjTruth(pObj *DarLibObj) int {
	panic("TODO, stubbed")
}

// Dar_LibAlloc
// C: Dar_Lib_t * Dar_LibAlloc( int nObjs )
func DarLibAlloc(nObjs int) *DarLib {
	panic("TODO, stubbed")
}

// Dar_LibFree
// C: void Dar_LibFree( Dar_Lib_t * p )
func (p *DarLib) DarLibFree() {
	panic("TODO, stubbed")
}

// Dar_LibReturnClass
// C: int Dar_LibReturnClass( unsigned uTruth )
func DarLibReturnClass(uTruth uint32) int {
	panic("TODO, stubbed")
}

// Dar_LibReturnCanonicals
// C: void Dar_LibReturnCanonicals( unsigned * pCanons )
func DarLibReturnCanonicals(pCanons []uint32) {
	panic("TODO, stubbed")
}

// Dar_LibAddNode
// C: void Dar_LibAddNode( Dar_Lib_t * p, int Id0, int Id1, int fCompl0, int fCompl1 )
func (p *DarLib) DarLibAddNode(Id0 int, Id1 int, fCompl0 int, fCompl1 int) {
	panic("TODO, stubbed")
}

// Dar_LibSetup_rec
// C: void Dar_LibSetup_rec( Dar_Lib_t * p, Dar_LibObj_t * pObj, int Class, int fCollect )
func (p *DarLib) DarLibSetup_rec(pObj *DarLibObj, Class int, fCollect int) {
	panic("TODO, stubbed")
}

// Dar_LibSetup
// C: void Dar_LibSetup( Dar_Lib_t * p, Vec_Int_t * vOuts, Vec_Int_t * vPrios )
func (p *DarLib) DarLibSetup(vOuts *IntVector, vPrios *IntVector) {
	panic("TODO, stubbed")
}

// Dar_LibCreateData
// C: void Dar_LibCreateData( Dar_Lib_t * p, int nDatas )
func (p *DarLib) DarLibCreateData(nDatas int) {
	panic("TODO, stubbed")
}

// Dar_LibSetup0_rec
// C: void Dar_LibSetup0_rec( Dar_Lib_t * p, Dar_LibObj_t * pObj, int Class, int fCollect )
func (p *DarLib) DarLibSetup0_rec(pObj *DarLibObj, Class int, fCollect int) {
	panic("TODO, stubbed")
}

// Dar_LibPrepare
// C: void Dar_LibPrepare( int nSubgraphs )
func DarLibPrepare(nSubgraphs int) {
	panic("TODO, stubbed")
}

// Dar_LibRead
// C: Dar_Lib_t * Dar_LibRead()
func DarLibRead() *DarLib {
	panic("TODO, stubbed")
}

// Dar_LibStart
// C: void Dar_LibStart()
func DarLibStart() {
	panic("TODO, stubbed")
}

// Dar_LibStop
// C: void Dar_LibStop()
func DarLibStop() {
	panic("TODO, stubbed")
}

// Dar_LibIncrementScore
// C: void Dar_LibIncrementScore( int Class, int Out, int Gain )
func DarLibIncrementScore(Class int, Out int, Gain int) {
	panic("TODO, stubbed")
}

// Dar_LibDumpPriorities
// C: void Dar_LibDumpPriorities()
func DarLibDumpPriorities() {
	panic("TODO, stubbed")
}

// Dar_LibCutMatch
// C: int Dar_LibCutMatch( Dar_Man_t * p, Dar_Cut_t * pCut )
func (p *DarMan) DarLibCutMatch(pCut *DarCut) int {
	panic("TODO, stubbed")
}

// Dar_LibCutMarkMffc (pPower is an in/out scalar)
// C: int Dar_LibCutMarkMffc( Aig_Man_t * p, Aig_Obj_t * pRoot, int nLeaves, float * pPower )
func (p *AigManager) DarLibCutMarkMffc(pRoot *AigObject, nLeaves int, pPower *float32) int {
	panic("TODO, stubbed")
}

// Dar_LibObjPrint_rec
// C: void Dar_LibObjPrint_rec( Dar_LibObj_t * pObj )
func DarLibObjPrint_rec(pObj *DarLibObj) {
	panic("TODO, stubbed")
}

// Dar_LibEvalAssignNums
// C: void Dar_LibEvalAssignNums( Dar_Man_t * p, int Class, Aig_Obj_t * pRoot )
func (p *DarMan) DarLibEvalAssignNums(Class int, pRoot *AigObject) {
	panic("TODO, stubbed")
}

// Dar_LibEval_rec (pPower is an in/out scalar)
// C: int Dar_LibEval_rec( Dar_LibObj_t * pObj, int Out, int nNodesSaved, int Required, float * pPower )
func DarLibEval_rec(pObj *DarLibObj, Out int, nNodesSaved int, Required int, pPower *float32) int {
	panic("TODO, stubbed")
}

// Dar_LibEval (pnMffcSize is an out parameter)
// C: void Dar_LibEval( Dar_Man_t * p, Aig_Obj_t * pRoot, Dar_Cut_t * pCut, int Required, int * pnMffcSize )
func (p *DarMan) DarLibEval(pRoot *AigObject, pCut *DarCut, Required int, pnMffcSize *int) {
	panic("TODO, stubbed")
}

// Dar_LibBuildClear_rec (pCounter is an in/out counter)
// C: void Dar_LibBuildClear_rec( Dar_LibObj_t * pObj, int * pCounter )
func DarLibBuildClear_rec(pObj *DarLibObj, pCounter *int) {
	panic("TODO, stubbed")
}

// Dar_LibBuildBest_rec
// C: Aig_Obj_t * Dar_LibBuildBest_rec( Dar_Man_t * p, Dar_LibObj_t * pObj )
func (p *DarMan) DarLibBuildBest_rec(pObj *DarLibObj) *AigObject {
	panic("TODO, stubbed")
}

// Dar_LibBuildBest
// C: Aig_Obj_t * Dar_LibBuildBest( Dar_Man_t * p )
func (p *DarMan) DarLibBuildBest() *AigObject {
	panic("TODO, stubbed")
}

// Dar2_LibCutMatch
// C: int Dar2_LibCutMatch( Gia_Man_t * p, Vec_Int_t * vCutLits, unsigned uTruth )
func (p *GiaManager) Dar2LibCutMatch(vCutLits *IntVector, uTruth uint32) int {
	panic("TODO, stubbed")
}

// Dar2_LibEvalAssignNums
// C: void Dar2_LibEvalAssignNums( Gia_Man_t * p, int Class )
func (p *GiaManager) Dar2LibEvalAssignNums(Class int) {
	panic("TODO, stubbed")
}

// Dar2_LibEval_rec
// C: int Dar2_LibEval_rec( Dar_LibObj_t * pObj, int Out )
func Dar2LibEval_rec(pObj *DarLibObj, Out int) int {
	panic("TODO, stubbed")
}

// Dar2_LibEval
// C: int Dar2_LibEval( Gia_Man_t * p, Vec_Int_t * vCutLits, unsigned uTruth, int fKeepLevel, Vec_Int_t * vLeavesBest2 )
func (p *GiaManager) Dar2LibEval(vCutLits *IntVector, uTruth uint32, fKeepLevel int, vLeavesBest2 *IntVector) int {
	panic("TODO, stubbed")
}

// Dar2_LibBuildClear_rec (pCounter is an in/out counter)
// C: void Dar2_LibBuildClear_rec( Dar_LibObj_t * pObj, int * pCounter )
func Dar2LibBuildClear_rec(pObj *DarLibObj, pCounter *int) {
	panic("TODO, stubbed")
}

// Dar2_LibBuildBest_rec
// C: int Dar2_LibBuildBest_rec( Gia_Man_t * p, Dar_LibObj_t * pObj )
func (p *GiaManager) Dar2LibBuildBest_rec(pObj *DarLibObj) int {
	panic("TODO, stubbed")
}

// Dar2_LibBuildBest
// C: int Dar2_LibBuildBest( Gia_Man_t * p, Vec_Int_t * vLeavesBest2, int OutBest )
func (p *GiaManager) Dar2LibBuildBest(vLeavesBest2 *IntVector, OutBest int) int {
	panic("TODO, stubbed")
}

// Dar_LibEvalBuild
// C: int Dar_LibEvalBuild( Gia_Man_t * p, Vec_Int_t * vCutLits, unsigned uTruth, int fKeepLevel, Vec_Int_t * vLeavesBest2 )
func (p *GiaManager) DarLibEvalBuild(vCutLits *IntVector, uTruth uint32, fKeepLevel int, vLeavesBest2 *IntVector) int {
	panic("TODO, stubbed")
}
