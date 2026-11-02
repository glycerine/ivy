package abc

// Hand-checked skeleton of src/aig/gia/giaStr.c.

const (
	STR_SUPER = 100
	MAX_TREE  = 10000
)

// Str_Type_t: object types (the enum of giaStr.c).
const (
	STR_NONE   = 0
	STR_CONST0 = 1
	STR_PI     = 2
	STR_AND    = 3
	STR_XOR    = 4
	STR_MUX    = 5
	STR_BUF    = 6
	STR_PO     = 7
	STR_UNUSED = 8
)

// StrObj is Str_Obj_t.
type StrObj struct {
	Type    uint32 // unsigned : 4 : object type
	nFanins uint32 // unsigned : 28 : fanin count
	iOffset int    // place where fanins are stored
	iTop    int    // top level MUX
	iCopy   int    // copy of this node
}

// StrNtk is Str_Ntk_t.
type StrNtk struct {
	nObjs      int       // object count
	nObjsAlloc int       // alloc objects
	pObjs      []StrObj  // Str_Obj_t * : objects
	vFanins    IntVector // Vec_Int_t (by value) : object fanins
	nObjCount  [STR_UNUSED]int
	nTrees     int
	nGroups    int
	DelayGain  int
}

// StrMan is Str_Man_t.
type StrMan struct {
	pOld     *GiaManager // manager
	nLutSize int         // LUT size
	fCutMin  int         // cut minimization
	pNtk     *StrNtk     // balanced network
	pNew     *GiaManager // newly constructed
	vDelays  *IntVector  // delays of each object
}

// StrEdg is Str_Edg_t.
type StrEdg struct {
	Fan    int // fanin ID
	fCompl int // fanin complement
	FanDel int // fanin delay
	Copy   int // fanin copy
}

// StrMux is Str_Mux_t.
type StrMux struct {
	Id       int       // node ID
	Delay    int       // node delay
	Copy     int       // node copy
	nLutSize int       // LUT size
	Edge     [3]StrEdg // fanins
}

// Str_NtkObj (static inline)
// C: Str_Obj_t * Str_NtkObj( Str_Ntk_t * p, int i )
func (p *StrNtk) strNtkObj(i int) *StrObj {
	panic("TODO, stubbed")
}

// Str_ObjFaninId (static inline)
// C: int Str_ObjFaninId( Str_Ntk_t * p, Str_Obj_t * pObj, int i )
func (p *StrNtk) strObjFaninId(pObj *StrObj, i int) int {
	panic("TODO, stubbed")
}

// Str_ObjFanin (static inline)
// C: Str_Obj_t * Str_ObjFanin( Str_Ntk_t * p, Str_Obj_t * pObj, int i )
func (p *StrNtk) strObjFanin(pObj *StrObj, i int) *StrObj {
	panic("TODO, stubbed")
}

// Str_ObjFaninC (static inline)
// C: int Str_ObjFaninC( Str_Ntk_t * p, Str_Obj_t * pObj, int i )
func (p *StrNtk) strObjFaninC(pObj *StrObj, i int) int {
	panic("TODO, stubbed")
}

// Str_ObjFaninCopy (static inline)
// C: int Str_ObjFaninCopy( Str_Ntk_t * p, Str_Obj_t * pObj, int i )
func (p *StrNtk) strObjFaninCopy(pObj *StrObj, i int) int {
	panic("TODO, stubbed")
}

// Str_ObjId (static inline)
// C: int Str_ObjId( Str_Ntk_t * p, Str_Obj_t * pObj )
func (p *StrNtk) strObjId(pObj *StrObj) int {
	panic("TODO, stubbed")
}

// Str_ObjCreate (static inline)
// C: int Str_ObjCreate( Str_Ntk_t * p, int Type, int nFanins, int * pFanins )
func (p *StrNtk) strObjCreate(Type int, nFanins int, pFanins []int) int {
	panic("TODO, stubbed")
}

// Str_NtkCreate (static inline)
// C: Str_Ntk_t * Str_NtkCreate( int nObjsAlloc, int nFaninsAlloc )
func strNtkCreate(nObjsAlloc int, nFaninsAlloc int) *StrNtk {
	panic("TODO, stubbed")
}

// Str_NtkDelete (static inline)
// C: void Str_NtkDelete( Str_Ntk_t * p )
func (p *StrNtk) strNtkDelete() {
	panic("TODO, stubbed")
}

// Str_NtkPs (static inline)
// C: void Str_NtkPs( Str_Ntk_t * p, abctime clk )
func (p *StrNtk) strNtkPs(clk int64) {
	panic("TODO, stubbed")
}

// Str_ObjReadGroup (static inline)
// C: void Str_ObjReadGroup( Str_Ntk_t * p, Str_Obj_t * pObj, int * pnGroups, int * pnMuxes )
func (p *StrNtk) strObjReadGroup(pObj *StrObj, pnGroups *int, pnMuxes *int) {
	panic("TODO, stubbed")
}

// Str_NtkPrintGroups (static inline)
// C: void Str_NtkPrintGroups( Str_Ntk_t * p )
func (p *StrNtk) strNtkPrintGroups() {
	panic("TODO, stubbed")
}

// Str_NtkToGia
// C: Gia_Man_t * Str_NtkToGia( Gia_Man_t * pGia, Str_Ntk_t * p )
func (pGia *GiaManager) StrNtkToGia(p *StrNtk) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManDupMuxesNoHash
// C: Gia_Man_t * Gia_ManDupMuxesNoHash( Gia_Man_t * p )
func (p *GiaManager) GiaManDupMuxesNoHash() *GiaManager {
	panic("TODO, stubbed")
}

// Str_MuxInputsCollect_rec
// C: void Str_MuxInputsCollect_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vNodes )
func (p *GiaManager) StrMuxInputsCollect_rec(pObj *GiaObject, vNodes *IntVector) {
	panic("TODO, stubbed")
}

// Str_MuxInputsCollect
// C: void Str_MuxInputsCollect( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vNodes )
func (p *GiaManager) StrMuxInputsCollect(pObj *GiaObject, vNodes *IntVector) {
	panic("TODO, stubbed")
}

// Str_MuxStructCollect_rec
// C: void Str_MuxStructCollect_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vNodes )
func (p *GiaManager) StrMuxStructCollect_rec(pObj *GiaObject, vNodes *IntVector) {
	panic("TODO, stubbed")
}

// Str_MuxStructCollect
// C: void Str_MuxStructCollect( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vNodes )
func (p *GiaManager) StrMuxStructCollect(pObj *GiaObject, vNodes *IntVector) {
	panic("TODO, stubbed")
}

// Str_MuxStructDump_rec
// C: void Str_MuxStructDump_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Str_t * vStr )
func (p *GiaManager) StrMuxStructDump_rec(pObj *GiaObject, vStr *ByteVector) {
	panic("TODO, stubbed")
}

// Str_MuxStructDump
// C: void Str_MuxStructDump( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Str_t * vStr )
func (p *GiaManager) StrMuxStructDump(pObj *GiaObject, vStr *ByteVector) {
	panic("TODO, stubbed")
}

// Str_ManMuxCountOne
// C: int Str_ManMuxCountOne( char * p )
func StrManMuxCountOne(p string) int {
	panic("TODO, stubbed")
}

// Str_ManDeriveTrees
// C: Vec_Wec_t * Str_ManDeriveTrees( Gia_Man_t * p )
func (p *GiaManager) StrManDeriveTrees() *IntLevelVector {
	panic("TODO, stubbed")
}

// Str_ManCreateRoots
// C: Vec_Int_t * Str_ManCreateRoots( Vec_Wec_t * vGroups, int nObjs )
func StrManCreateRoots(vGroups *IntLevelVector, nObjs int) *IntVector {
	panic("TODO, stubbed")
}

// Str_MuxTraverse_rec
// C: void Str_MuxTraverse_rec( Gia_Man_t * p, int i )
func (p *GiaManager) StrMuxTraverse_rec(i int) {
	panic("TODO, stubbed")
}

// Str_ManCheckOverlap
// C: void Str_ManCheckOverlap( Gia_Man_t * p, Vec_Wec_t * vGroups )
func (p *GiaManager) StrManCheckOverlap(vGroups *IntLevelVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimplifyXor (static inline)
// C: void Gia_ManSimplifyXor( Vec_Int_t * vSuper )
func giaManSimplifyXor_str(vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimplifyAnd (static inline)
// C: void Gia_ManSimplifyAnd( Vec_Int_t * vSuper )
func giaManSimplifyAnd_str(vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManSuperCollectXor_rec (static inline)
// C: void Gia_ManSuperCollectXor_rec( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) giaManSuperCollectXor_rec_str(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManSuperCollectAnd_rec (static inline)
// C: void Gia_ManSuperCollectAnd_rec( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) giaManSuperCollectAnd_rec_str(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManSuperCollect (static inline)
// C: void Gia_ManSuperCollect( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) giaManSuperCollect_str(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Str_ManNormalize_rec
// C: void Str_ManNormalize_rec( Str_Ntk_t * pNtk, Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Wec_t * vGroups, Vec_Int_t * vRoots )
func (pNtk *StrNtk) StrManNormalize_rec(p *GiaManager, pObj *GiaObject, vGroups *IntLevelVector, vRoots *IntVector) {
	panic("TODO, stubbed")
}

// Str_ManNormalizeInt
// C: Str_Ntk_t * Str_ManNormalizeInt( Gia_Man_t * p, Vec_Wec_t * vGroups, Vec_Int_t * vRoots )
func (p *GiaManager) StrManNormalizeInt(vGroups *IntLevelVector, vRoots *IntVector) *StrNtk {
	panic("TODO, stubbed")
}

// Str_ManNormalize
// C: Str_Ntk_t * Str_ManNormalize( Gia_Man_t * p )
func (p *GiaManager) StrManNormalize() *StrNtk {
	panic("TODO, stubbed")
}

// Str_Delay2 (static inline)
// C: int Str_Delay2( int d0, int d1, int nLutSize )
func strDelay2(d0 int, d1 int, nLutSize int) int {
	panic("TODO, stubbed")
}

// Str_Delay3 (static inline)
// C: int Str_Delay3( int d0, int d1, int d2, int nLutSize )
func strDelay3(d0 int, d1 int, d2 int, nLutSize int) int {
	panic("TODO, stubbed")
}

// Str_ObjDelay (static inline)
// C: int Str_ObjDelay( Gia_Man_t * pNew, int iObj, int nLutSize, Vec_Int_t * vDelay )
func (pNew *GiaManager) strObjDelay(iObj int, nLutSize int, vDelay *IntVector) int {
	panic("TODO, stubbed")
}

// transpose64 (static inline)
// C: void transpose64( word A[64] )
func transpose64_giaStr(A []uint64) {
	panic("TODO, stubbed")
}

// Str_ManNum (static inline)
// C: int Str_ManNum( Gia_Man_t * p, int iObj )
func (p *GiaManager) strManNum(iObj int) int {
	panic("TODO, stubbed")
}

// Str_ManSetNum (static inline)
// C: void Str_ManSetNum( Gia_Man_t * p, int iObj, int Num )
func (p *GiaManager) strManSetNum(iObj int, Num int) {
	panic("TODO, stubbed")
}

// Str_ManVectorAffinity
// C: int Str_ManVectorAffinity( Gia_Man_t * p, Vec_Int_t * vSuper, Vec_Int_t * vDelay, word Matrix[256], int nLimit )
func (p *GiaManager) StrManVectorAffinity(vSuper *IntVector, vDelay *IntVector, Matrix []uint64, nLimit int) int {
	panic("TODO, stubbed")
}

// Str_CountBits (static inline)
// C: int Str_CountBits( word i )
func strCountBits(i uint64) int {
	panic("TODO, stubbed")
}

// Str_PrintState (static inline)
// C: void Str_PrintState( int * pCost, int * pSuper, word * pMatrix, int nSize )
func strPrintState(pCost []int, pSuper []int, pMatrix []uint64, nSize int) {
	panic("TODO, stubbed")
}

// Str_NtkBalanceMulti2
// C: void Str_NtkBalanceMulti2( Gia_Man_t * pNew, Str_Ntk_t * p, Str_Obj_t * pObj, Vec_Int_t * vDelay, int nLutSize )
func (pNew *GiaManager) StrNtkBalanceMulti2(p *StrNtk, pObj *StrObj, vDelay *IntVector, nLutSize int) {
	panic("TODO, stubbed")
}

// Str_NtkBalanceTwo
// C: int Str_NtkBalanceTwo( Gia_Man_t * pNew, Str_Ntk_t * p, Str_Obj_t * pObj, int i, int j, Vec_Int_t * vDelay, int * pCost, int * pSuper, word * pMatrix, int nSize, int nLutSize, int CostBest )
func (pNew *GiaManager) StrNtkBalanceTwo(p *StrNtk, pObj *StrObj, i int, j int, vDelay *IntVector, pCost []int, pSuper []int, pMatrix []uint64, nSize int, nLutSize int, CostBest int) int {
	panic("TODO, stubbed")
}

// Str_NtkBalanceMulti
// C: void Str_NtkBalanceMulti( Gia_Man_t * pNew, Str_Ntk_t * p, Str_Obj_t * pObj, Vec_Int_t * vDelay, int nLutSize )
func (pNew *GiaManager) StrNtkBalanceMulti(p *StrNtk, pObj *StrObj, vDelay *IntVector, nLutSize int) {
	panic("TODO, stubbed")
}

// Str_NtkBalanceMux
// C: void Str_NtkBalanceMux( Gia_Man_t * pNew, Str_Ntk_t * p, Str_Obj_t * pObj, Vec_Int_t * vDelay, int nLutSize, int nGroups, int nMuxes, int fRecursive, int fOptArea, int fVerbose )
func (pNew *GiaManager) StrNtkBalanceMux(p *StrNtk, pObj *StrObj, vDelay *IntVector, nLutSize int, nGroups int, nMuxes int, fRecursive int, fOptArea int, fVerbose int) {
	panic("TODO, stubbed")
}

// Str_NtkBalance
// C: Gia_Man_t * Str_NtkBalance( Gia_Man_t * pGia, Str_Ntk_t * p, int nLutSize, int fUseMuxes, int fRecursive, int fOptArea, int fVerbose )
func (pGia *GiaManager) StrNtkBalance(p *StrNtk, nLutSize int, fUseMuxes int, fRecursive int, fOptArea int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManLutBalance
// C: Gia_Man_t * Gia_ManLutBalance( Gia_Man_t * p, int nLutSize, int fUseMuxes, int fRecursive, int fOptArea, int fVerbose )
func (p *GiaManager) GiaManLutBalance(nLutSize int, fUseMuxes int, fRecursive int, fOptArea int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Str_MuxFanin (static inline)
// C: Str_Mux_t * Str_MuxFanin( Str_Mux_t * pMux, int i )
func strMuxFanin(pMux *StrMux, i int) *StrMux {
	panic("TODO, stubbed")
}

// Str_MuxHasFanin (static inline)
// C: int Str_MuxHasFanin( Str_Mux_t * pMux, int i )
func strMuxHasFanin(pMux *StrMux, i int) int {
	panic("TODO, stubbed")
}

// Str_MuxDelayPrint_rec
// C: void Str_MuxDelayPrint_rec( Str_Mux_t * pMux, int i )
func StrMuxDelayPrint_rec(pMux *StrMux, i int) {
	panic("TODO, stubbed")
}

// Str_MuxDelayEdge_rec
// C: int Str_MuxDelayEdge_rec( Str_Mux_t * pMux, int i )
func StrMuxDelayEdge_rec(pMux *StrMux, i int) int {
	panic("TODO, stubbed")
}

// Str_MuxCreate
// C: void Str_MuxCreate( Str_Mux_t * pTree, Str_Ntk_t * pNtk, int iMux, int nMuxes, Vec_Int_t * vDelay, int nLutSize )
func StrMuxCreate(pTree *StrMux, pNtk *StrNtk, iMux int, nMuxes int, vDelay *IntVector, nLutSize int) {
	panic("TODO, stubbed")
}

// Str_MuxToGia_rec
// C: int Str_MuxToGia_rec( Gia_Man_t * pNew, Str_Mux_t * pMux, int i, Vec_Int_t * vDelay )
func (pNew *GiaManager) StrMuxToGia_rec(pMux *StrMux, i int, vDelay *IntVector) int {
	panic("TODO, stubbed")
}

// Str_MuxChangeOnce
// C: void Str_MuxChangeOnce( Str_Mux_t * pTree, int * pPath, int i, int k, Str_Mux_t * pBackup, Gia_Man_t * pNew, Vec_Int_t * vDelay )
func StrMuxChangeOnce(pTree *StrMux, pPath []int, i int, k int, pBackup *StrMux, pNew *GiaManager, vDelay *IntVector) {
	panic("TODO, stubbed")
}

// Str_MuxChangeUndo
// C: void Str_MuxChangeUndo( Str_Mux_t * pTree, int * pPath, int i, int k, Str_Mux_t * pBackup )
func StrMuxChangeUndo(pTree *StrMux, pPath []int, i int, k int, pBackup *StrMux) {
	panic("TODO, stubbed")
}

// Str_MuxFindPathEdge_rec
// C: int Str_MuxFindPathEdge_rec( Str_Mux_t * pMux, int i, int * pPath, int * pnLength )
func StrMuxFindPathEdge_rec(pMux *StrMux, i int, pPath []int, pnLength *int) int {
	panic("TODO, stubbed")
}

// Str_MuxFindPath_rec
// C: int Str_MuxFindPath_rec( Str_Mux_t * pMux, int * pPath, int * pnLength )
func StrMuxFindPath_rec(pMux *StrMux, pPath []int, pnLength *int) int {
	panic("TODO, stubbed")
}

// Str_MuxFindBranching
// C: Str_Mux_t * Str_MuxFindBranching( Str_Mux_t * pRoot, int i )
func StrMuxFindBranching(pRoot *StrMux, i int) *StrMux {
	panic("TODO, stubbed")
}

// Str_MuxTryOnce
// C: int Str_MuxTryOnce( Gia_Man_t * pNew, Str_Ntk_t * pNtk, Str_Mux_t * pTree, Str_Mux_t * pRoot, int Edge, Vec_Int_t * vDelay, int fVerbose )
func (pNew *GiaManager) StrMuxTryOnce(pNtk *StrNtk, pTree *StrMux, pRoot *StrMux, Edge int, vDelay *IntVector, fVerbose int) int {
	panic("TODO, stubbed")
}

// Str_MuxRestruct_rec
// C: int Str_MuxRestruct_rec( Gia_Man_t * pNew, Str_Ntk_t * pNtk, Str_Mux_t * pTree, Str_Mux_t * pRoot, int Edge, Vec_Int_t * vDelay, int fVerbose )
func (pNew *GiaManager) StrMuxRestruct_rec(pNtk *StrNtk, pTree *StrMux, pRoot *StrMux, Edge int, vDelay *IntVector, fVerbose int) int {
	panic("TODO, stubbed")
}

// Str_MuxRestructure2
// C: int Str_MuxRestructure2( Gia_Man_t * pNew, Str_Ntk_t * pNtk, int iMux, int nMuxes, Vec_Int_t * vDelay, int nLutSize, int fVerbose )
func (pNew *GiaManager) StrMuxRestructure2(pNtk *StrNtk, iMux int, nMuxes int, vDelay *IntVector, nLutSize int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Str_MuxRestructure1
// C: int Str_MuxRestructure1( Gia_Man_t * pNew, Str_Ntk_t * pNtk, int iMux, int nMuxes, Vec_Int_t * vDelay, int nLutSize, int fVerbose )
func (pNew *GiaManager) StrMuxRestructure1(pNtk *StrNtk, iMux int, nMuxes int, vDelay *IntVector, nLutSize int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Str_MuxRestructure
// C: int Str_MuxRestructure( Gia_Man_t * pNew, Str_Ntk_t * pNtk, int iMux, int nMuxes, Vec_Int_t * vDelay, int nLutSize, int fRecursive, int fOptArea, int fVerbose )
func (pNew *GiaManager) StrMuxRestructure(pNtk *StrNtk, iMux int, nMuxes int, vDelay *IntVector, nLutSize int, fRecursive int, fOptArea int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Str_MuxRestructAreaThree
// C: int Str_MuxRestructAreaThree( Gia_Man_t * pNew, Str_Mux_t * pMux, Vec_Int_t * vDelay, int fVerbose )
func (pNew *GiaManager) StrMuxRestructAreaThree(pMux *StrMux, vDelay *IntVector, fVerbose int) int {
	panic("TODO, stubbed")
}

// Str_MuxRestructArea_rec
// C: int Str_MuxRestructArea_rec( Gia_Man_t * pNew, Str_Mux_t * pTree, Str_Mux_t * pRoot, int i, Vec_Int_t * vDelay, int fVerbose )
func (pNew *GiaManager) StrMuxRestructArea_rec(pTree *StrMux, pRoot *StrMux, i int, vDelay *IntVector, fVerbose int) int {
	panic("TODO, stubbed")
}

// Str_MuxRestructureArea
// C: int Str_MuxRestructureArea( Gia_Man_t * pNew, Str_Ntk_t * pNtk, int iMux, int nMuxes, Vec_Int_t * vDelay, int nLutSize, int fVerbose )
func (pNew *GiaManager) StrMuxRestructureArea(pNtk *StrNtk, iMux int, nMuxes int, vDelay *IntVector, nLutSize int, fVerbose int) int {
	panic("TODO, stubbed")
}
