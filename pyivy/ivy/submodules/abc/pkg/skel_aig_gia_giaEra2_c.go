package abc

// Hand-checked skeleton of src/aig/gia/giaEra2.c.

const (
	MAX_CALL_NUM = 1000000 // the max number of recursive calls
	MAX_ITEM_NUM = 1 << 20 // the number of items on a page
	MAX_PAGE_NUM = 1 << 11 // the max number of memory pages
	MAX_VARS_NUM = 1 << 14 // the max number of state vars allowed
	MAX_CUBE_NUM = 63      // the max number of cubes before rebalancing
)

// GiaPtrAre is Gia_PtrAre_t.
type GiaPtrAre struct {
	nItem uint32 // unsigned : 20 : item number (related to MAX_ITEM_NUM)
	nPage uint32 // unsigned : 11 : page number (related to MAX_PAGE_NUM)
	fMark uint32 // unsigned : 1 : user mark
}

// GiaPtrAreInt is Gia_PtrAreInt_t: a union of the two views.
type GiaPtrAreInt struct {
	iGia GiaPtrAre
	iInt uint32
}

// GiaObjAre is Gia_ObjAre_t.
type GiaObjAre struct {
	iVar   uint32       // unsigned : 14 : variable (related to MAX_VARS_NUM)
	nStas0 uint32       // unsigned : 6 : cube counter (related to MAX_CUBE_NUM)
	nStas1 uint32       // unsigned : 6 : cube counter (related to MAX_CUBE_NUM)
	nStas2 uint32       // unsigned : 6 : cube counter (related to MAX_CUBE_NUM)
	F      [3]GiaPtrAre // branches
}

// GiaStaAre is Gia_StaAre_t.
type GiaStaAre struct {
	iPrev GiaPtrAre // previous state
	iNext GiaPtrAre // next cube in the list
	pData []uint32  // unsigned pData[0] : state bits (flexible array member)
}

// GiaManAre is Gia_ManAre_t.
type GiaManAre struct {
	pAig       *GiaManager        // user's AIG manager
	pNew       *GiaManager        // temporary AIG manager
	ppObjs     [][]uint32         // unsigned ** : storage for objects (MAX_PAGE_NUM pages)
	ppStas     [][]uint32         // unsigned ** : storage for states  (MAX_PAGE_NUM pages)
	fMiter     int                // stops when a bug is discovered
	fStopped   int                // set high when reachability is stopped
	fTree      int                // working in the tree mode
	nWords     int                // the size of bit info in words
	nSize      int                // the size of state structure in words
	nObjPages  int                // the number of pages used for objects
	nStaPages  int                // the number of pages used for states
	nObjs      int                // the number of objects
	nStas      int                // the number of states
	iStaCur    int                // the next state to be explored
	Root       GiaPtrAre          // root of the tree
	vCiTfos    *NestedVector[int] // storage for nodes in the CI TFOs
	vCiLits    *NestedVector[int] // storage for literals of these nodes
	vCubesA    *IntVector         // checked cubes
	vCubesB    *IntVector         // unchecked cubes
	pSat       *SatSolver         // void * : SAT solver
	vSatNumCis *IntVector         // SAT variables for CIs
	vSatNumCos *IntVector         // SAT variables for COs
	vCofVars   *IntVector         // variables used to cofactor
	vAssumps   *IntVector         // temporary storage for assumptions
	pTarget    *GiaStaAre         // state that needs to be reached
	iOutFail   int                // the number of the failed output
	nChecks    int                // the number of timea cube was checked
	nEquals    int                // total number of equal
	nCompares  int                // the number of compares
	nRecCalls  int                // the number of rec calls
	nDisjs     int                // the number of disjoint cube pairs
	nDisjs2    int                // the number of disjoint cube pairs
	nDisjs3    int                // the number of disjoint cube pairs
	timeAig    int                // AIG cofactoring time
	timeCube   int                // cube checking time
}

// Gia_Int2Ptr (static inline)
// C: Gia_PtrAre_t Gia_Int2Ptr( unsigned n )
func giaInt2Ptr(n uint32) GiaPtrAre {
	panic("TODO, stubbed")
}

// Gia_Ptr2Int (static inline)
// C: unsigned Gia_Ptr2Int( Gia_PtrAre_t n )
func giaPtr2Int(n GiaPtrAre) uint32 {
	panic("TODO, stubbed")
}

// Gia_ObjHasBranch0 (static inline)
// C: int Gia_ObjHasBranch0( Gia_ObjAre_t * q )
func giaObjHasBranch0(q *GiaObjAre) int {
	panic("TODO, stubbed")
}

// Gia_ObjHasBranch1 (static inline)
// C: int Gia_ObjHasBranch1( Gia_ObjAre_t * q )
func giaObjHasBranch1(q *GiaObjAre) int {
	panic("TODO, stubbed")
}

// Gia_ObjHasBranch2 (static inline)
// C: int Gia_ObjHasBranch2( Gia_ObjAre_t * q )
func giaObjHasBranch2(q *GiaObjAre) int {
	panic("TODO, stubbed")
}

// Gia_ManAreObj (static inline)
// C: Gia_ObjAre_t * Gia_ManAreObj( Gia_ManAre_t * p, Gia_PtrAre_t n )
func (p *GiaManAre) giaManAreObj(n GiaPtrAre) *GiaObjAre {
	panic("TODO, stubbed")
}

// Gia_ManAreSta (static inline)
// C: Gia_StaAre_t * Gia_ManAreSta( Gia_ManAre_t * p, Gia_PtrAre_t n )
func (p *GiaManAre) giaManAreSta(n GiaPtrAre) *GiaStaAre {
	panic("TODO, stubbed")
}

// Gia_ManAreObjInt (static inline)
// C: Gia_ObjAre_t * Gia_ManAreObjInt( Gia_ManAre_t * p, int n )
func (p *GiaManAre) giaManAreObjInt(n int) *GiaObjAre {
	panic("TODO, stubbed")
}

// Gia_ManAreStaInt (static inline)
// C: Gia_StaAre_t * Gia_ManAreStaInt( Gia_ManAre_t * p, int n )
func (p *GiaManAre) giaManAreStaInt(n int) *GiaStaAre {
	panic("TODO, stubbed")
}

// Gia_ManAreObjLast (static inline)
// C: Gia_ObjAre_t * Gia_ManAreObjLast( Gia_ManAre_t * p )
func (p *GiaManAre) giaManAreObjLast() *GiaObjAre {
	panic("TODO, stubbed")
}

// Gia_ManAreStaLast (static inline)
// C: Gia_StaAre_t * Gia_ManAreStaLast( Gia_ManAre_t * p )
func (p *GiaManAre) giaManAreStaLast() *GiaStaAre {
	panic("TODO, stubbed")
}

// Gia_ObjNextObj0 (static inline)
// C: Gia_ObjAre_t * Gia_ObjNextObj0( Gia_ManAre_t * p, Gia_ObjAre_t * q )
func (p *GiaManAre) giaObjNextObj0(q *GiaObjAre) *GiaObjAre {
	panic("TODO, stubbed")
}

// Gia_ObjNextObj1 (static inline)
// C: Gia_ObjAre_t * Gia_ObjNextObj1( Gia_ManAre_t * p, Gia_ObjAre_t * q )
func (p *GiaManAre) giaObjNextObj1(q *GiaObjAre) *GiaObjAre {
	panic("TODO, stubbed")
}

// Gia_ObjNextObj2 (static inline)
// C: Gia_ObjAre_t * Gia_ObjNextObj2( Gia_ManAre_t * p, Gia_ObjAre_t * q )
func (p *GiaManAre) giaObjNextObj2(q *GiaObjAre) *GiaObjAre {
	panic("TODO, stubbed")
}

// Gia_StaHasValue0 (static inline)
// C: int Gia_StaHasValue0( Gia_StaAre_t * p, int iReg )
func giaStaHasValue0(p *GiaStaAre, iReg int) int {
	panic("TODO, stubbed")
}

// Gia_StaHasValue1 (static inline)
// C: int Gia_StaHasValue1( Gia_StaAre_t * p, int iReg )
func giaStaHasValue1(p *GiaStaAre, iReg int) int {
	panic("TODO, stubbed")
}

// Gia_StaSetValue0 (static inline)
// C: void Gia_StaSetValue0( Gia_StaAre_t * p, int iReg )
func giaStaSetValue0(p *GiaStaAre, iReg int) {
	panic("TODO, stubbed")
}

// Gia_StaSetValue1 (static inline)
// C: void Gia_StaSetValue1( Gia_StaAre_t * p, int iReg )
func giaStaSetValue1(p *GiaStaAre, iReg int) {
	panic("TODO, stubbed")
}

// Gia_StaPrev (static inline)
// C: Gia_StaAre_t * Gia_StaPrev( Gia_ManAre_t * p, Gia_StaAre_t * pS )
func (p *GiaManAre) giaStaPrev(pS *GiaStaAre) *GiaStaAre {
	panic("TODO, stubbed")
}

// Gia_StaNext (static inline)
// C: Gia_StaAre_t * Gia_StaNext( Gia_ManAre_t * p, Gia_StaAre_t * pS )
func (p *GiaManAre) giaStaNext(pS *GiaStaAre) *GiaStaAre {
	panic("TODO, stubbed")
}

// Gia_StaIsGood (static inline)
// C: int Gia_StaIsGood( Gia_ManAre_t * p, Gia_StaAre_t * pS )
func (p *GiaManAre) giaStaIsGood(pS *GiaStaAre) int {
	panic("TODO, stubbed")
}

// Gia_StaSetUnused (static inline)
// C: void Gia_StaSetUnused( Gia_StaAre_t * pS )
func giaStaSetUnused(pS *GiaStaAre) {
	panic("TODO, stubbed")
}

// Gia_StaIsUnused (static inline)
// C: int Gia_StaIsUnused( Gia_StaAre_t * pS )
func giaStaIsUnused(pS *GiaStaAre) int {
	panic("TODO, stubbed")
}

// Gia_StaIsUsed (static inline)
// C: int Gia_StaIsUsed( Gia_StaAre_t * pS )
func giaStaIsUsed(pS *GiaStaAre) int {
	panic("TODO, stubbed")
}

// Gia_ManCountMintermsInCube
// C: void Gia_ManCountMintermsInCube( Gia_StaAre_t * pCube, int nVars, unsigned * pStore )
func GiaManCountMintermsInCube(pCube *GiaStaAre, nVars int, pStore []uint32) {
	panic("TODO, stubbed")
}

// Gia_ManCountMinterms
// C: int Gia_ManCountMinterms( Gia_ManAre_t * p )
func (p *GiaManAre) GiaManCountMinterms() int {
	panic("TODO, stubbed")
}

// Gia_ManDeriveCiTfo_rec
// C: int Gia_ManDeriveCiTfo_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vRes )
func (p *GiaManager) GiaManDeriveCiTfo_rec(pObj *GiaObject, vRes *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManDeriveCiTfoOne
// C: Vec_Int_t * Gia_ManDeriveCiTfoOne( Gia_Man_t * p, Gia_Obj_t * pPivot )
func (p *GiaManager) GiaManDeriveCiTfoOne(pPivot *GiaObject) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManDeriveCiTfo
// C: Vec_Vec_t * Gia_ManDeriveCiTfo( Gia_Man_t * p )
func (p *GiaManager) GiaManDeriveCiTfo() *NestedVector[int] {
	panic("TODO, stubbed")
}

// Gia_StaAreEqual (static inline)
// C: int Gia_StaAreEqual( Gia_StaAre_t * p1, Gia_StaAre_t * p2, int nWords )
func giaStaAreEqual(p1 *GiaStaAre, p2 *GiaStaAre, nWords int) int {
	panic("TODO, stubbed")
}

// Gia_StaAreDisjoint (static inline)
// C: int Gia_StaAreDisjoint( Gia_StaAre_t * p1, Gia_StaAre_t * p2, int nWords )
func giaStaAreDisjoint(p1 *GiaStaAre, p2 *GiaStaAre, nWords int) int {
	panic("TODO, stubbed")
}

// Gia_StaAreContain (static inline)
// C: int Gia_StaAreContain( Gia_StaAre_t * p1, Gia_StaAre_t * p2, int nWords )
func giaStaAreContain(p1 *GiaStaAre, p2 *GiaStaAre, nWords int) int {
	panic("TODO, stubbed")
}

// Gia_StaAreDashNum (static inline)
// C: int Gia_StaAreDashNum( Gia_StaAre_t * p1, Gia_StaAre_t * p2, int nWords )
func giaStaAreDashNum(p1 *GiaStaAre, p2 *GiaStaAre, nWords int) int {
	panic("TODO, stubbed")
}

// Gia_StaAreSharpVar (static inline)
// C: int Gia_StaAreSharpVar( Gia_StaAre_t * p1, Gia_StaAre_t * p2, int nWords )
func giaStaAreSharpVar(p1 *GiaStaAre, p2 *GiaStaAre, nWords int) int {
	panic("TODO, stubbed")
}

// Gia_StaAreDisjointVar (static inline)
// C: int Gia_StaAreDisjointVar( Gia_StaAre_t * p1, Gia_StaAre_t * p2, int nWords )
func giaStaAreDisjointVar(p1 *GiaStaAre, p2 *GiaStaAre, nWords int) int {
	panic("TODO, stubbed")
}

// Gia_ManAreCreate
// C: Gia_ManAre_t * Gia_ManAreCreate( Gia_Man_t * pAig )
func (pAig *GiaManager) GiaManAreCreate() *GiaManAre {
	panic("TODO, stubbed")
}

// Gia_ManAreFree
// C: void Gia_ManAreFree( Gia_ManAre_t * p )
func (p *GiaManAre) GiaManAreFree() {
	panic("TODO, stubbed")
}

// Gia_ManAreCreateObj (static inline)
// C: Gia_ObjAre_t * Gia_ManAreCreateObj( Gia_ManAre_t * p )
func (p *GiaManAre) giaManAreCreateObj() *GiaObjAre {
	panic("TODO, stubbed")
}

// Gia_ManAreCreateSta (static inline)
// C: Gia_StaAre_t * Gia_ManAreCreateSta( Gia_ManAre_t * p )
func (p *GiaManAre) giaManAreCreateSta() *GiaStaAre {
	panic("TODO, stubbed")
}

// Gia_ManAreRycycleSta (static inline)
// C: void Gia_ManAreRycycleSta( Gia_ManAre_t * p, Gia_StaAre_t * pSta )
func (p *GiaManAre) giaManAreRycycleSta(pSta *GiaStaAre) {
	panic("TODO, stubbed")
}

// Gia_ManAreCreateStaNew (static inline)
// C: Gia_StaAre_t * Gia_ManAreCreateStaNew( Gia_ManAre_t * p )
func (p *GiaManAre) giaManAreCreateStaNew() *GiaStaAre {
	panic("TODO, stubbed")
}

// Gia_ManAreCreateStaInit (static inline)
// C: Gia_StaAre_t * Gia_ManAreCreateStaInit( Gia_ManAre_t * p )
func (p *GiaManAre) giaManAreCreateStaInit() *GiaStaAre {
	panic("TODO, stubbed")
}

// Gia_ManArePrintCube
// C: void Gia_ManArePrintCube( Gia_ManAre_t * p, Gia_StaAre_t * pSta )
func (p *GiaManAre) GiaManArePrintCube(pSta *GiaStaAre) {
	panic("TODO, stubbed")
}

// Gia_ManAreDepth
// C: int Gia_ManAreDepth( Gia_ManAre_t * p, int iState )
func (p *GiaManAre) GiaManAreDepth(iState int) int {
	panic("TODO, stubbed")
}

// Gia_ManAreListCountListUsed (static inline)
// C: int Gia_ManAreListCountListUsed( Gia_ManAre_t * p, Gia_PtrAre_t Root )
func (p *GiaManAre) giaManAreListCountListUsed(Root GiaPtrAre) int {
	panic("TODO, stubbed")
}

// Gia_ManAreListCountUsed_rec
// C: int Gia_ManAreListCountUsed_rec( Gia_ManAre_t * p, Gia_PtrAre_t Root, int fTree )
func (p *GiaManAre) GiaManAreListCountUsed_rec(Root GiaPtrAre, fTree int) int {
	panic("TODO, stubbed")
}

// Gia_ManAreListCountUsed (static inline)
// C: int Gia_ManAreListCountUsed( Gia_ManAre_t * p )
func (p *GiaManAre) giaManAreListCountUsed() int {
	panic("TODO, stubbed")
}

// Gia_ManArePrintListUsed (static inline)
// C: int Gia_ManArePrintListUsed( Gia_ManAre_t * p, Gia_PtrAre_t Root )
func (p *GiaManAre) giaManArePrintListUsed(Root GiaPtrAre) int {
	panic("TODO, stubbed")
}

// Gia_ManArePrintUsed_rec
// C: int Gia_ManArePrintUsed_rec( Gia_ManAre_t * p, Gia_PtrAre_t Root, int fTree )
func (p *GiaManAre) GiaManArePrintUsed_rec(Root GiaPtrAre, fTree int) int {
	panic("TODO, stubbed")
}

// Gia_ManArePrintUsed (static inline)
// C: int Gia_ManArePrintUsed( Gia_ManAre_t * p )
func (p *GiaManAre) giaManArePrintUsed() int {
	panic("TODO, stubbed")
}

// Gia_ManAreFindBestVar
// C: int Gia_ManAreFindBestVar( Gia_ManAre_t * p, Gia_PtrAre_t List )
func (p *GiaManAre) GiaManAreFindBestVar(List GiaPtrAre) int {
	panic("TODO, stubbed")
}

// Gia_ManAreRebalance (static inline)
// C: void Gia_ManAreRebalance( Gia_ManAre_t * p, Gia_PtrAre_t * pRoot )
func (p *GiaManAre) giaManAreRebalance(pRoot *GiaPtrAre) {
	panic("TODO, stubbed")
}

// Gia_ManAreCompress (static inline)
// C: void Gia_ManAreCompress( Gia_ManAre_t * p, Gia_PtrAre_t * pRoot )
func (p *GiaManAre) giaManAreCompress(pRoot *GiaPtrAre) {
	panic("TODO, stubbed")
}

// Gia_ManAreCubeCheckList (static inline)
// C: int Gia_ManAreCubeCheckList( Gia_ManAre_t * p, Gia_PtrAre_t * pRoot, Gia_StaAre_t * pSta )
func (p *GiaManAre) giaManAreCubeCheckList(pRoot *GiaPtrAre, pSta *GiaStaAre) int {
	panic("TODO, stubbed")
}

// Gia_ManAreCubeAddToList (static inline)
// C: void Gia_ManAreCubeAddToList( Gia_ManAre_t * p, Gia_PtrAre_t * pRoot, Gia_StaAre_t * pSta )
func (p *GiaManAre) giaManAreCubeAddToList(pRoot *GiaPtrAre, pSta *GiaStaAre) {
	panic("TODO, stubbed")
}

// Gia_ManAreCubeCheckTree_rec
// C: int Gia_ManAreCubeCheckTree_rec( Gia_ManAre_t * p, Gia_ObjAre_t * pObj, Gia_StaAre_t * pSta )
func (p *GiaManAre) GiaManAreCubeCheckTree_rec(pObj *GiaObjAre, pSta *GiaStaAre) int {
	panic("TODO, stubbed")
}

// Gia_ManAreCubeAddToTree_rec
// C: void Gia_ManAreCubeAddToTree_rec( Gia_ManAre_t * p, Gia_ObjAre_t * pObj, Gia_StaAre_t * pSta )
func (p *GiaManAre) GiaManAreCubeAddToTree_rec(pObj *GiaObjAre, pSta *GiaStaAre) {
	panic("TODO, stubbed")
}

// Gia_ManAreCubeCollectList (static inline)
// C: int Gia_ManAreCubeCollectList( Gia_ManAre_t * p, Gia_PtrAre_t * pRoot, Gia_StaAre_t * pSta )
func (p *GiaManAre) giaManAreCubeCollectList(pRoot *GiaPtrAre, pSta *GiaStaAre) int {
	panic("TODO, stubbed")
}

// Gia_ManAreCubeCollectTree_rec
// C: int Gia_ManAreCubeCollectTree_rec( Gia_ManAre_t * p, Gia_ObjAre_t * pObj, Gia_StaAre_t * pSta )
func (p *GiaManAre) GiaManAreCubeCollectTree_rec(pObj *GiaObjAre, pSta *GiaStaAre) int {
	panic("TODO, stubbed")
}

// Gia_ManAreCubeCheckTree
// C: int Gia_ManAreCubeCheckTree( Gia_ManAre_t * p, Gia_StaAre_t * pSta )
func (p *GiaManAre) GiaManAreCubeCheckTree(pSta *GiaStaAre) int {
	panic("TODO, stubbed")
}

// Gia_ManAreCubeProcess (static inline)
// C: int Gia_ManAreCubeProcess( Gia_ManAre_t * p, Gia_StaAre_t * pSta )
func (p *GiaManAre) giaManAreCubeProcess(pSta *GiaStaAre) int {
	panic("TODO, stubbed")
}

// Gia_ManAreMostUsedPi_rec
// C: void Gia_ManAreMostUsedPi_rec( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) GiaManAreMostUsedPi_rec(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManAreMostUsedPi (static inline)
// C: Gia_Obj_t * Gia_ManAreMostUsedPi( Gia_ManAre_t * p )
func (p *GiaManAre) giaManAreMostUsedPi() *GiaObject {
	panic("TODO, stubbed")
}

// Gia_ManCheckPOs_rec
// C: int Gia_ManCheckPOs_rec( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) GiaManCheckPOs_rec(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Gia_ManCheckPOs (static inline)
// C: int Gia_ManCheckPOs( Gia_ManAre_t * p )
func (p *GiaManAre) giaManCheckPOs() int {
	panic("TODO, stubbed")
}

// Gia_ManCheckPOstatus (static inline)
// C: int Gia_ManCheckPOstatus( Gia_ManAre_t * p )
func (p *GiaManAre) giaManCheckPOstatus() int {
	panic("TODO, stubbed")
}

// Gia_ManAreDeriveNexts_rec
// C: int Gia_ManAreDeriveNexts_rec( Gia_ManAre_t * p, Gia_PtrAre_t Sta )
func (p *GiaManAre) GiaManAreDeriveNexts_rec(Sta GiaPtrAre) int {
	panic("TODO, stubbed")
}

// Gia_ManAreDeriveNexts
// C: int Gia_ManAreDeriveNexts( Gia_ManAre_t * p, Gia_PtrAre_t Sta )
func (p *GiaManAre) GiaManAreDeriveNexts(Sta GiaPtrAre) int {
	panic("TODO, stubbed")
}

// Gia_ManArePrintReport
// C: void Gia_ManArePrintReport( Gia_ManAre_t * p, abctime Time, int fFinal )
func (p *GiaManAre) GiaManArePrintReport(Time int64, fFinal int) {
	panic("TODO, stubbed")
}

// Gia_ManArePerform
// C: int Gia_ManArePerform( Gia_Man_t * pAig, int nStatesMax, int fMiter, int fVerbose )
func (pAig *GiaManager) GiaManArePerform(nStatesMax int, fMiter int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_ManAreDeriveCexSatStart
// C: void Gia_ManAreDeriveCexSatStart( Gia_ManAre_t * p )
func (p *GiaManAre) GiaManAreDeriveCexSatStart() {
	panic("TODO, stubbed")
}

// Gia_ManAreDeriveCexSatStop
// C: void Gia_ManAreDeriveCexSatStop( Gia_ManAre_t * p )
func (p *GiaManAre) GiaManAreDeriveCexSatStop() {
	panic("TODO, stubbed")
}

// Gia_ManAreDeriveCexSat
// C: void Gia_ManAreDeriveCexSat( Gia_ManAre_t * p, Gia_StaAre_t * pCur, Gia_StaAre_t * pNext, int iOutFailed )
func (p *GiaManAre) GiaManAreDeriveCexSat(pCur *GiaStaAre, pNext *GiaStaAre, iOutFailed int) {
	panic("TODO, stubbed")
}

// Gia_ManAreDeriveCex
// C: Abc_Cex_t * Gia_ManAreDeriveCex( Gia_ManAre_t * p, Gia_StaAre_t * pLast )
func (p *GiaManAre) GiaManAreDeriveCex(pLast *GiaStaAre) *Counterexample {
	panic("TODO, stubbed")
}
