package abc

// Hand-checked skeleton of src/map/mpm/mpmPre.c.

// IfdObj is Ifd_Obj_t. Go has no bitfields; each is a uint32 whose C width is
// in the comment.
type IfdObj struct {
	nFreq uint32    // unsigned : 18, frequency
	nAnds uint32    // unsigned : 6, number of AND gates
	nSupp uint32    // unsigned : 5, support size
	Type  uint32    // unsigned : 2, type
	fWay  uint32    // unsigned : 1, transparent edge
	pFans [3]uint32 // unsigned pFans[3] : fanins
}

// IfdMan is Ifd_Man_t.
type IfdMan struct {
	pObjs      *IfdObj // an array of nObjs objects
	nObjs      int
	nObjsAlloc int
	// hashing operations
	vArgs    *IntVector              // iDsd1 op iDsdC
	vRes     *IntVector              // result of operation
	vHash    *FixedIntRecordInterner // Hsh_IntMan_t * : hash table
	vMarks   *IntVector              // marks where given N begins
	vTruths  *WordVector             // truth tables
	vClauses *IntVector              // truth tables
	// other data
	vSuper *IntVector
}

// Ifd_ObjIsVar (static inline)
// C: int Ifd_ObjIsVar( Ifd_Obj_t * p )
func (p *IfdObj) ifdObjIsVar() int {
	panic("TODO, stubbed")
}

// Ifd_ObjIsAnd (static inline)
// C: int Ifd_ObjIsAnd( Ifd_Obj_t * p )
func (p *IfdObj) ifdObjIsAnd() int {
	panic("TODO, stubbed")
}

// Ifd_ObjIsXor (static inline)
// C: int Ifd_ObjIsXor( Ifd_Obj_t * p )
func (p *IfdObj) ifdObjIsXor() int {
	panic("TODO, stubbed")
}

// Ifd_ObjIsMux (static inline)
// C: int Ifd_ObjIsMux( Ifd_Obj_t * p )
func (p *IfdObj) ifdObjIsMux() int {
	panic("TODO, stubbed")
}

// Ifd_ManObj (static inline)
// C: Ifd_Obj_t * Ifd_ManObj( Ifd_Man_t * p, int i )
func (p *IfdMan) ifdManObj(i int) *IfdObj {
	panic("TODO, stubbed")
}

// Ifd_ManObjFromLit (static inline)
// C: Ifd_Obj_t * Ifd_ManObjFromLit( Ifd_Man_t * p, int iLit )
func (p *IfdMan) ifdManObjFromLit(iLit int) *IfdObj {
	panic("TODO, stubbed")
}

// Ifd_ObjId (static inline)
// C: int Ifd_ObjId( Ifd_Man_t * p, Ifd_Obj_t * pObj )
func (p *IfdMan) ifdObjId(pObj *IfdObj) int {
	panic("TODO, stubbed")
}

// Ifd_LitSuppSize (static inline)
// C: int Ifd_LitSuppSize( Ifd_Man_t * p, int iLit )
func (p *IfdMan) ifdLitSuppSize(iLit int) int {
	panic("TODO, stubbed")
}

// Ifd_LitNumAnds (static inline)
// C: int Ifd_LitNumAnds( Ifd_Man_t * p, int iLit )
func (p *IfdMan) ifdLitNumAnds(iLit int) int {
	panic("TODO, stubbed")
}

// Ifd_ManStart
// C: Ifd_Man_t * Ifd_ManStart()
func IfdManStart() *IfdMan {
	panic("TODO, stubbed")
}

// Ifd_ManStop
// C: void Ifd_ManStop( Ifd_Man_t * p )
func (p *IfdMan) IfdManStop() {
	panic("TODO, stubbed")
}

// Ifd_ObjPrint_rec (pCounter is an in/out counter)
// C: void Ifd_ObjPrint_rec( Ifd_Man_t * p, int iLit, int * pCounter, int DiffType )
func (p *IfdMan) IfdObjPrint_rec(iLit int, pCounter *int, DiffType int) {
	panic("TODO, stubbed")
}

// Ifd_ObjPrint
// C: void Ifd_ObjPrint( Ifd_Man_t * p, int iLit )
func (p *IfdMan) IfdObjPrint(iLit int) {
	panic("TODO, stubbed")
}

// Ifd_ManPrint2
// C: void Ifd_ManPrint2( Ifd_Man_t * p )
func (p *IfdMan) IfdManPrint2() {
	panic("TODO, stubbed")
}

// Ifd_ManPrint
// C: void Ifd_ManPrint( Ifd_Man_t * p )
func (p *IfdMan) IfdManPrint() {
	panic("TODO, stubbed")
}

// Ifd_ObjTruth_rec (pCounter is an in/out counter)
// C: word Ifd_ObjTruth_rec( Ifd_Man_t * p, int iLit, int * pCounter )
func (p *IfdMan) IfdObjTruth_rec(iLit int, pCounter *int) uint64 {
	panic("TODO, stubbed")
}

// Ifd_ObjTruth
// C: word Ifd_ObjTruth( Ifd_Man_t * p, int iLit )
func (p *IfdMan) IfdObjTruth(iLit int) uint64 {
	panic("TODO, stubbed")
}

// Ifd_ManTruthAll
// C: void Ifd_ManTruthAll( Ifd_Man_t * p )
func (p *IfdMan) IfdManTruthAll() {
	panic("TODO, stubbed")
}

// Mpm_ComputeCnfSizeOne
// C: int Mpm_ComputeCnfSizeOne( word Truth, int nVars, Vec_Int_t * vCover, Vec_Str_t * vCnf )
func MpmComputeCnfSizeOne(Truth uint64, nVars int, vCover *IntVector, vCnf *ByteVector) int {
	panic("TODO, stubbed")
}

// Mpm_ComputeCnfSizeAll
// C: void Mpm_ComputeCnfSizeAll( Ifd_Man_t * p )
func (p *IfdMan) MpmComputeCnfSizeAll() {
	panic("TODO, stubbed")
}

// Ifd_ManHashLookup
// C: int Ifd_ManHashLookup( Ifd_Man_t * p, int iDsd0, int iDsd1, int iDsdC, int Type )
func (p *IfdMan) IfdManHashLookup(iDsd0 int, iDsd1 int, iDsdC int, Type int) int {
	panic("TODO, stubbed")
}

// Ifd_ManHashInsert
// C: void Ifd_ManHashInsert( Ifd_Man_t * p, int iDsd0, int iDsd1, int iDsdC, int Type, int Res )
func (p *IfdMan) IfdManHashInsert(iDsd0 int, iDsd1 int, iDsdC int, Type int, Res int) {
	panic("TODO, stubbed")
}

// Ifd_ManHashFindOrAdd
// C: int Ifd_ManHashFindOrAdd( Ifd_Man_t * p, int iDsd0, int iDsd1, int iDsdC, int Type )
func (p *IfdMan) IfdManHashFindOrAdd(iDsd0 int, iDsd1 int, iDsdC int, Type int) int {
	panic("TODO, stubbed")
}

// Ifd_ManOperSuper_rec
// C: void Ifd_ManOperSuper_rec( Ifd_Man_t * p, int iLit, int Type, Vec_Int_t * vObjs )
func (p *IfdMan) IfdManOperSuper_rec(iLit int, Type int, vObjs *IntVector) {
	panic("TODO, stubbed")
}

// Ifd_ManOper
// C: int Ifd_ManOper( Ifd_Man_t * p, int iDsd0, int iDsd1, int iDsdC, int Type )
func (p *IfdMan) IfdManOper(iDsd0 int, iDsd1 int, iDsdC int, Type int) int {
	panic("TODO, stubbed")
}

// Ifd_ManFindDsd_rec (p is a cursor into pStr that the function advances (char **))
// C: int Ifd_ManFindDsd_rec( Ifd_Man_t * pMan, char * pStr, char ** p, int * pMatches )
func (pMan *IfdMan) IfdManFindDsd_rec(pStr string, p *string, pMatches []int) int {
	panic("TODO, stubbed")
}

// Ifd_ManComputeMatches
// C: int * Ifd_ManComputeMatches( char * p )
func IfdManComputeMatches(p string) []int {
	panic("TODO, stubbed")
}

// Ifd_ManFindDsd
// C: int Ifd_ManFindDsd( Ifd_Man_t * pMan, char * p )
func (pMan *IfdMan) IfdManFindDsd(p string) int {
	panic("TODO, stubbed")
}

// Ifd_ManDsdTest2
// C: void Ifd_ManDsdTest2()
func IfdManDsdTest2() {
	panic("TODO, stubbed")
}

// Ifd_ManDsdTruths
// C: Vec_Wrd_t * Ifd_ManDsdTruths( int nVars )
func IfdManDsdTruths(nVars int) *WordVector {
	panic("TODO, stubbed")
}

// Ifd_ManDsdPermPrint
// C: void Ifd_ManDsdPermPrint( int * perm, int size )
func IfdManDsdPermPrint(perm []int, size int) {
	panic("TODO, stubbed")
}

// Ifd_ManDsdPermJT
// C: Vec_Int_t * Ifd_ManDsdPermJT( int n )
func IfdManDsdPermJT(n int) *IntVector {
	panic("TODO, stubbed")
}

// Ifd_ManDsdTest4
// C: int Ifd_ManDsdTest4()
func IfdManDsdTest4() int {
	panic("TODO, stubbed")
}

// Extra_Truth6SwapAdjacent (static inline; suffixed because misc_extra_dsd.go already defines extraTruth6SwapAdjacent)
// C: word Extra_Truth6SwapAdjacent( word t, int iVar )
func extraTruth6SwapAdjacent_mpmPre(t uint64, iVar int) uint64 {
	panic("TODO, stubbed")
}

// Extra_Truth6ChangePhase (static inline; suffixed because misc_extra_dsd.go may define the same name)
// C: word Extra_Truth6ChangePhase( word t, int iVar)
func extraTruth6ChangePhase_mpmPre(t uint64, iVar int) uint64 {
	panic("TODO, stubbed")
}

// Extra_Truth6AllConfigs2
// C: Vec_Wrd_t * Extra_Truth6AllConfigs2( word t, int * pComp, int * pPerm, int nVars )
func ExtraTruth6AllConfigs2(t uint64, pComp []int, pPerm []int, nVars int) *WordVector {
	panic("TODO, stubbed")
}

// Extra_Truth6AllConfigs
// C: Vec_Wrd_t * Extra_Truth6AllConfigs( word t, int * pComp, int * pPerm, int nVars )
func ExtraTruth6AllConfigs(t uint64, pComp []int, pPerm []int, nVars int) *WordVector {
	panic("TODO, stubbed")
}

// Ifd_ComputeSignature (the array parameter decays to a pointer and is filled in)
// C: void Ifd_ComputeSignature( word uTruth, int pCounts[6] )
func IfdComputeSignature(uTruth uint64, pCounts *[6]int) {
	panic("TODO, stubbed")
}

// Ifd_ManDsdTest33
// C: int Ifd_ManDsdTest33()
func IfdManDsdTest33() int {
	panic("TODO, stubbed")
}

// Ifd_ManDsdTest
// C: int Ifd_ManDsdTest()
func IfdManDsdTest() int {
	panic("TODO, stubbed")
}

// Ifd_ManDsdTest55
// C: int Ifd_ManDsdTest55()
func IfdManDsdTest55() int {
	panic("TODO, stubbed")
}
