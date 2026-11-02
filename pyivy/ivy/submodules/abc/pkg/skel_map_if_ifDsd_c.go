package abc

import "os"

// Hand-checked skeleton of src/map/if/ifDsd.c.

const DSD_VERSION = "dsd1"

// If_DsdType_t: network types.
const (
	IF_DSD_NONE   = 0 // 0:  unknown
	IF_DSD_CONST0     // 1:  constant
	IF_DSD_VAR        // 2:  variable
	IF_DSD_AND        // 3:  AND
	IF_DSD_XOR        // 4:  XOR
	IF_DSD_MUX        // 5:  MUX
	IF_DSD_PRIME      // 6:  PRIME
)

// IfDsdObj is If_DsdObj_t.
type IfDsdObj struct {
	Id    uint32   // node ID
	Type  uint32   // unsigned : 3 : node type
	nSupp uint32   // unsigned : 5 : variable
	fMark uint32   // unsigned : 1 : user mark
	Count uint32   // unsigned : 18 : variable
	nFans uint32   // unsigned : 5 : fanin count
	pFans []uint32 // unsigned pFans[0] : fanins (flexible array member)
}

// IfDsdMan is If_DsdMan_t.
type IfDsdMan struct {
	pStore        string                                             // char * : input/output file
	nVars         int                                                // max var number
	LutSize       int                                                // LUT size
	nWords        int                                                // word number
	nBins         int                                                // table size
	pBins         []uint32                                           // unsigned * : hash table
	pMem          *ExtraFlexArena                                    // Mem_Flex_t * : memory for nodes
	vObjs         PointerVector[IfDsdObj]                            // Vec_Ptr_t (by value) : objects
	vNexts        IntVector                                          // Vec_Int_t (by value) : next pointers
	vTruths       IntVector                                          // Vec_Int_t (by value) : truth IDs of prime nodes
	vTemp1        *IntVector                                         // temp
	vTemp2        *IntVector                                         // temp
	pTtElems      [][]uint64                                         // word ** : elementary TTs
	vTtMem        [IF_MAX_FUNC_LUTSIZE + 1]*WordMemory               // truth table memory and hash table
	vTtDecs       [IF_MAX_FUNC_LUTSIZE + 1]*PointerVector[IntVector] // truth table decompositions
	vIsops        [IF_MAX_FUNC_LUTSIZE + 1]*IntLevelVector           // ISOP for each function
	pSched        [IF_MAX_FUNC_LUTSIZE][]int                         // int * : grey code schedules
	nTtBits       int                                                // the number of truth table bits
	nConfigWords  int                                                // the number of words for config data per node
	vConfigs      *WordVector                                        // permutations
	pTtGia        *GiaManager                                        // GIA to represent truth tables
	vCover        *IntVector                                         // temporary memory
	pSat          *SatSolver                                         // void * : SAT solver
	pCellStr      string                                             // char * : symbolic cell description
	nObjsPrev     int                                                // previous number of objects
	fNewAsUseless int                                                // set new as useless
	nUniqueHits   int                                                // statistics
	nUniqueMisses int                                                // statistics
	timeDsd       int64                                              // abctime : statistics
	timeCanon     int64                                              // abctime : statistics
	timeCheck     int64                                              // abctime : statistics
	timeCheck2    int64                                              // abctime : statistics
	timeVerify    int64                                              // abctime : statistics
}

const PAR_THR_MAX = 100

// Macros of src/opt/dau/dau.h used by Ifn_ThData_t.
const (
	DAU_MAX_VAR  = 12 // should be 6 or more
	DAU_MAX_WORD = 1 << (DAU_MAX_VAR - 6)
)

// IfnThData is Ifn_ThData_t.
type IfnThData struct {
	pNtk         *IfnNtk              // network
	pTruth       [DAU_MAX_WORD]uint64 // word pTruth[DAU_MAX_WORD]
	pConfig      [10]uint64           // configuration data
	nConfigWords int                  // configuration data word count
	nVars        int                  // support
	Id           int                  // object
	nConfls      int                  // conflicts
	Result       int                  // result
	Status       int                  // state
	clkUsed      int64                // abctime : total runtime
}

// If_DsdObjWordNum (static inline)
// C: int If_DsdObjWordNum( int nFans )
func ifDsdObjWordNum(nFans int) int {
	panic("TODO, stubbed")
}

// If_DsdObjTruthId (static inline)
// C: int If_DsdObjTruthId( If_DsdMan_t * p, If_DsdObj_t * pObj )
func (p *IfDsdMan) ifDsdObjTruthId(pObj *IfDsdObj) int {
	panic("TODO, stubbed")
}

// If_DsdObjTruth (static inline)
// C: word * If_DsdObjTruth( If_DsdMan_t * p, If_DsdObj_t * pObj )
func (p *IfDsdMan) ifDsdObjTruth(pObj *IfDsdObj) []uint64 {
	panic("TODO, stubbed")
}

// If_DsdObjSetTruth (static inline)
// C: void If_DsdObjSetTruth( If_DsdMan_t * p, If_DsdObj_t * pObj, int Id )
func (p *IfDsdMan) ifDsdObjSetTruth(pObj *IfDsdObj, Id int) {
	panic("TODO, stubbed")
}

// If_DsdObjClean (static inline)
// C: void If_DsdObjClean( If_DsdObj_t * pObj )
func ifDsdObjClean(pObj *IfDsdObj) {
	panic("TODO, stubbed")
}

// If_DsdObjId (static inline)
// C: int If_DsdObjId( If_DsdObj_t * pObj )
func ifDsdObjId(pObj *IfDsdObj) int {
	panic("TODO, stubbed")
}

// If_DsdObjType (static inline)
// C: int If_DsdObjType( If_DsdObj_t * pObj )
func ifDsdObjType(pObj *IfDsdObj) int {
	panic("TODO, stubbed")
}

// If_DsdObjIsVar (static inline)
// C: int If_DsdObjIsVar( If_DsdObj_t * pObj )
func ifDsdObjIsVar(pObj *IfDsdObj) int {
	panic("TODO, stubbed")
}

// If_DsdObjSuppSize (static inline)
// C: int If_DsdObjSuppSize( If_DsdObj_t * pObj )
func ifDsdObjSuppSize(pObj *IfDsdObj) int {
	panic("TODO, stubbed")
}

// If_DsdObjFaninNum (static inline)
// C: int If_DsdObjFaninNum( If_DsdObj_t * pObj )
func ifDsdObjFaninNum(pObj *IfDsdObj) int {
	panic("TODO, stubbed")
}

// If_DsdObjFaninC (static inline)
// C: int If_DsdObjFaninC( If_DsdObj_t * pObj, int i )
func ifDsdObjFaninC(pObj *IfDsdObj, i int) int {
	panic("TODO, stubbed")
}

// If_DsdObjFaninLit (static inline)
// C: int If_DsdObjFaninLit( If_DsdObj_t * pObj, int i )
func ifDsdObjFaninLit(pObj *IfDsdObj, i int) int {
	panic("TODO, stubbed")
}

// If_DsdVecObj (static inline)
// C: If_DsdObj_t * If_DsdVecObj( Vec_Ptr_t * p, int Id )
func ifDsdVecObj(p *PointerVector[IfDsdObj], Id int) *IfDsdObj {
	panic("TODO, stubbed")
}

// If_DsdVecConst0 (static inline)
// C: If_DsdObj_t * If_DsdVecConst0( Vec_Ptr_t * p )
func ifDsdVecConst0(p *PointerVector[IfDsdObj]) *IfDsdObj {
	panic("TODO, stubbed")
}

// If_DsdVecVar (static inline)
// C: If_DsdObj_t * If_DsdVecVar( Vec_Ptr_t * p, int v )
func ifDsdVecVar(p *PointerVector[IfDsdObj], v int) *IfDsdObj {
	panic("TODO, stubbed")
}

// If_DsdVecObjSuppSize (static inline)
// C: int If_DsdVecObjSuppSize( Vec_Ptr_t * p, int iObj )
func ifDsdVecObjSuppSize(p *PointerVector[IfDsdObj], iObj int) int {
	panic("TODO, stubbed")
}

// If_DsdVecLitSuppSize (static inline)
// C: int If_DsdVecLitSuppSize( Vec_Ptr_t * p, int iLit )
func ifDsdVecLitSuppSize(p *PointerVector[IfDsdObj], iLit int) int {
	panic("TODO, stubbed")
}

// If_DsdVecObjRef (static inline)
// C: int If_DsdVecObjRef( Vec_Ptr_t * p, int iObj )
func ifDsdVecObjRef(p *PointerVector[IfDsdObj], iObj int) int {
	panic("TODO, stubbed")
}

// If_DsdVecObjIncRef (static inline)
// C: void If_DsdVecObjIncRef( Vec_Ptr_t * p, int iObj )
func ifDsdVecObjIncRef(p *PointerVector[IfDsdObj], iObj int) {
	panic("TODO, stubbed")
}

// If_DsdObjFanin (static inline)
// C: If_DsdObj_t * If_DsdObjFanin( Vec_Ptr_t * p, If_DsdObj_t * pObj, int i )
func ifDsdObjFanin(p *PointerVector[IfDsdObj], pObj *IfDsdObj, i int) *IfDsdObj {
	panic("TODO, stubbed")
}

// If_DsdVecObjMark (static inline)
// C: int If_DsdVecObjMark( Vec_Ptr_t * p, int iObj )
func ifDsdVecObjMark(p *PointerVector[IfDsdObj], iObj int) int {
	panic("TODO, stubbed")
}

// If_DsdVecObjSetMark (static inline)
// C: void If_DsdVecObjSetMark( Vec_Ptr_t * p, int iObj )
func ifDsdVecObjSetMark(p *PointerVector[IfDsdObj], iObj int) {
	panic("TODO, stubbed")
}

// If_DsdVecObjClearMark (static inline)
// C: void If_DsdVecObjClearMark( Vec_Ptr_t * p, int iObj )
func ifDsdVecObjClearMark(p *PointerVector[IfDsdObj], iObj int) {
	panic("TODO, stubbed")
}

// If_DsdManFileName
// C: char * If_DsdManFileName( If_DsdMan_t * p )
func (p *IfDsdMan) IfDsdManFileName() string {
	panic("TODO, stubbed")
}

// If_DsdManVarNum
// C: int If_DsdManVarNum( If_DsdMan_t * p )
func (p *IfDsdMan) IfDsdManVarNum() int {
	panic("TODO, stubbed")
}

// If_DsdManObjNum
// C: int If_DsdManObjNum( If_DsdMan_t * p )
func (p *IfDsdMan) IfDsdManObjNum() int {
	panic("TODO, stubbed")
}

// If_DsdManLutSize
// C: int If_DsdManLutSize( If_DsdMan_t * p )
func (p *IfDsdMan) IfDsdManLutSize() int {
	panic("TODO, stubbed")
}

// If_DsdManTtBitNum
// C: int If_DsdManTtBitNum( If_DsdMan_t * p )
func (p *IfDsdMan) IfDsdManTtBitNum() int {
	panic("TODO, stubbed")
}

// If_DsdManPermBitNum
// C: int If_DsdManPermBitNum( If_DsdMan_t * p )
func (p *IfDsdMan) IfDsdManPermBitNum() int {
	panic("TODO, stubbed")
}

// If_DsdManSetLutSize
// C: void If_DsdManSetLutSize( If_DsdMan_t * p, int nLutSize )
func (p *IfDsdMan) IfDsdManSetLutSize(nLutSize int) {
	panic("TODO, stubbed")
}

// If_DsdManSuppSize
// C: int If_DsdManSuppSize( If_DsdMan_t * p, int iDsd )
func (p *IfDsdMan) IfDsdManSuppSize(iDsd int) int {
	panic("TODO, stubbed")
}

// If_DsdManCheckDec
// C: int If_DsdManCheckDec( If_DsdMan_t * p, int iDsd )
func (p *IfDsdMan) IfDsdManCheckDec(iDsd int) int {
	panic("TODO, stubbed")
}

// If_DsdManReadMark
// C: int If_DsdManReadMark( If_DsdMan_t * p, int iDsd )
func (p *IfDsdMan) IfDsdManReadMark(iDsd int) int {
	panic("TODO, stubbed")
}

// If_DsdManSetNewAsUseless
// C: void If_DsdManSetNewAsUseless( If_DsdMan_t * p )
func (p *IfDsdMan) IfDsdManSetNewAsUseless() {
	panic("TODO, stubbed")
}

// If_DsdManGetFuncConfig
// C: word * If_DsdManGetFuncConfig( If_DsdMan_t * p, int iDsd )
func (p *IfDsdMan) IfDsdManGetFuncConfig(iDsd int) []uint64 {
	panic("TODO, stubbed")
}

// If_DsdManGetCellStr
// C: char * If_DsdManGetCellStr( If_DsdMan_t * p )
func (p *IfDsdMan) IfDsdManGetCellStr() string {
	panic("TODO, stubbed")
}

// If_ManDsdTtElems (static inline)
// C: word ** If_ManDsdTtElems()
func ifManDsdTtElems() [][]uint64 {
	panic("TODO, stubbed")
}

// If_DsdObjAlloc
// C: If_DsdObj_t * If_DsdObjAlloc( If_DsdMan_t * p, int Type, int nFans )
func (p *IfDsdMan) IfDsdObjAlloc(Type int, nFans int) *IfDsdObj {
	panic("TODO, stubbed")
}

// If_DsdManAlloc
// C: If_DsdMan_t * If_DsdManAlloc( int nVars, int LutSize )
func IfDsdManAlloc(nVars int, LutSize int) *IfDsdMan {
	panic("TODO, stubbed")
}

// If_DsdManAllocIsops
// C: void If_DsdManAllocIsops( If_DsdMan_t * p, int nLutSize )
func (p *IfDsdMan) IfDsdManAllocIsops(nLutSize int) {
	panic("TODO, stubbed")
}

// If_DsdManFree
// C: void If_DsdManFree( If_DsdMan_t * p, int fVerbose )
func (p *IfDsdMan) IfDsdManFree(fVerbose int) {
	panic("TODO, stubbed")
}

// If_DsdManDumpDsd
// C: void If_DsdManDumpDsd( If_DsdMan_t * p, int Support )
func (p *IfDsdMan) IfDsdManDumpDsd(Support int) {
	panic("TODO, stubbed")
}

// If_DsdManDumpAll
// C: void If_DsdManDumpAll( If_DsdMan_t * p, int Support )
func (p *IfDsdMan) IfDsdManDumpAll(Support int) {
	panic("TODO, stubbed")
}

// If_DsdManHasMarks
// C: int If_DsdManHasMarks( If_DsdMan_t * p )
func (p *IfDsdMan) IfDsdManHasMarks() int {
	panic("TODO, stubbed")
}

// If_DsdManHashProfile
// C: void If_DsdManHashProfile( If_DsdMan_t * p )
func (p *IfDsdMan) IfDsdManHashProfile() {
	panic("TODO, stubbed")
}

// If_DsdManCheckNonDec_rec
// C: int If_DsdManCheckNonDec_rec( If_DsdMan_t * p, int Id )
func (p *IfDsdMan) IfDsdManCheckNonDec_rec(Id int) int {
	panic("TODO, stubbed")
}

// If_DsdManPrint_rec
// C: void If_DsdManPrint_rec( FILE * pFile, If_DsdMan_t * p, int iDsdLit, unsigned char * pPermLits, int * pnSupp )
func IfDsdManPrint_rec(pFile *os.File, p *IfDsdMan, iDsdLit int, pPermLits []uint8, pnSupp *int) {
	panic("TODO, stubbed")
}

// If_DsdManPrintOne
// C: void If_DsdManPrintOne( FILE * pFile, If_DsdMan_t * p, int iObjId, unsigned char * pPermLits, int fNewLine )
func IfDsdManPrintOne(pFile *os.File, p *IfDsdMan, iObjId int, pPermLits []uint8, fNewLine int) {
	panic("TODO, stubbed")
}

// If_DsdManPrintDecs
// C: void If_DsdManPrintDecs( FILE * pFile, If_DsdMan_t * p )
func IfDsdManPrintDecs(pFile *os.File, p *IfDsdMan) {
	panic("TODO, stubbed")
}

// If_DsdManPrintOccurs
// C: void If_DsdManPrintOccurs( FILE * pFile, If_DsdMan_t * p )
func IfDsdManPrintOccurs(pFile *os.File, p *IfDsdMan) {
	panic("TODO, stubbed")
}

// If_DsdManPrintDistrib
// C: void If_DsdManPrintDistrib( If_DsdMan_t * p )
func (p *IfDsdMan) IfDsdManPrintDistrib() {
	panic("TODO, stubbed")
}

// If_DsdManPrint
// C: void If_DsdManPrint( If_DsdMan_t * p, char * pFileName, int Number, int Support, int fOccurs, int fTtDump, int fVerbose )
func (p *IfDsdMan) IfDsdManPrint(pFileName string, Number int, Support int, fOccurs int, fTtDump int, fVerbose int) {
	panic("TODO, stubbed")
}

// If_DsdManCheckNonTriv
// C: int If_DsdManCheckNonTriv( If_DsdMan_t * p, int Id, int nVars, int iVarMax )
func (p *IfDsdMan) IfDsdManCheckNonTriv(Id int, nVars int, iVarMax int) int {
	panic("TODO, stubbed")
}

// If_DsdObjCompare
// C: int If_DsdObjCompare( If_DsdMan_t * pMan, Vec_Ptr_t * p, int iLit0, int iLit1 )
func (pMan *IfDsdMan) IfDsdObjCompare(p *PointerVector[IfDsdObj], iLit0 int, iLit1 int) int {
	panic("TODO, stubbed")
}

// If_DsdObjSort
// C: void If_DsdObjSort( If_DsdMan_t * pMan, Vec_Ptr_t * p, int * pLits, int nLits, int * pPerm )
func (pMan *IfDsdMan) IfDsdObjSort(p *PointerVector[IfDsdObj], pLits []int, nLits int, pPerm []int) {
	panic("TODO, stubbed")
}

// If_DsdObjHashKey (static inline)
// C: unsigned If_DsdObjHashKey( If_DsdMan_t * p, int Type, int * pLits, int nLits, int truthId )
func (p *IfDsdMan) ifDsdObjHashKey(Type int, pLits []int, nLits int, truthId int) uint32 {
	panic("TODO, stubbed")
}

// If_DsdObjHashLookup
// C: unsigned * If_DsdObjHashLookup( If_DsdMan_t * p, int Type, int * pLits, int nLits, int truthId )
func (p *IfDsdMan) IfDsdObjHashLookup(Type int, pLits []int, nLits int, truthId int) *uint32 {
	panic("TODO, stubbed")
}

// If_DsdObjHashResize
// C: void If_DsdObjHashResize( If_DsdMan_t * p )
func (p *IfDsdMan) ifDsdObjHashResize() {
	panic("TODO, stubbed")
}

// If_DsdObjCreate
// C: int If_DsdObjCreate( If_DsdMan_t * p, int Type, int * pLits, int nLits, int truthId )
func (p *IfDsdMan) IfDsdObjCreate(Type int, pLits []int, nLits int, truthId int) int {
	panic("TODO, stubbed")
}

// If_DsdObjFindOrAdd
// C: int If_DsdObjFindOrAdd( If_DsdMan_t * p, int Type, int * pLits, int nLits, word * pTruth )
func (p *IfDsdMan) IfDsdObjFindOrAdd(Type int, pLits []int, nLits int, pTruth []uint64) int {
	panic("TODO, stubbed")
}

// If_DsdManSave
// C: void If_DsdManSave( If_DsdMan_t * p, char * pFileName )
func (p *IfDsdMan) IfDsdManSave(pFileName string) {
	panic("TODO, stubbed")
}

// If_DsdManLoad
// C: If_DsdMan_t * If_DsdManLoad( char * pFileName )
func IfDsdManLoad(pFileName string) *IfDsdMan {
	panic("TODO, stubbed")
}

// If_DsdManMerge
// C: void If_DsdManMerge( If_DsdMan_t * p, If_DsdMan_t * pNew )
func (p *IfDsdMan) IfDsdManMerge(pNew *IfDsdMan) {
	panic("TODO, stubbed")
}

// If_DsdManCleanOccur
// C: void If_DsdManCleanOccur( If_DsdMan_t * p, int fVerbose )
func (p *IfDsdMan) IfDsdManCleanOccur(fVerbose int) {
	panic("TODO, stubbed")
}

// If_DsdManCleanMarks
// C: void If_DsdManCleanMarks( If_DsdMan_t * p, int fVerbose )
func (p *IfDsdMan) IfDsdManCleanMarks(fVerbose int) {
	panic("TODO, stubbed")
}

// If_DsdManInvertMarks
// C: void If_DsdManInvertMarks( If_DsdMan_t * p, int fVerbose )
func (p *IfDsdMan) IfDsdManInvertMarks(fVerbose int) {
	panic("TODO, stubbed")
}

// If_DsdManFilter_rec
// C: void If_DsdManFilter_rec( If_DsdMan_t * pNew, If_DsdMan_t * p, int i, Vec_Int_t * vMap )
func (pNew *IfDsdMan) IfDsdManFilter_rec(p *IfDsdMan, i int, vMap *IntVector) {
	panic("TODO, stubbed")
}

// If_DsdManFilter
// C: If_DsdMan_t * If_DsdManFilter( If_DsdMan_t * p, int Limit )
func (p *IfDsdMan) IfDsdManFilter(Limit int) *IfDsdMan {
	panic("TODO, stubbed")
}

// If_DsdManCollect_rec
// C: void If_DsdManCollect_rec( If_DsdMan_t * p, int Id, Vec_Int_t * vNodes, Vec_Int_t * vFirsts, int * pnSupp )
func (p *IfDsdMan) IfDsdManCollect_rec(Id int, vNodes *IntVector, vFirsts *IntVector, pnSupp *int) {
	panic("TODO, stubbed")
}

// If_DsdManCollect
// C: void If_DsdManCollect( If_DsdMan_t * p, int Id, Vec_Int_t * vNodes, Vec_Int_t * vFirsts )
func (p *IfDsdMan) IfDsdManCollect(Id int, vNodes *IntVector, vFirsts *IntVector) {
	panic("TODO, stubbed")
}

// If_DsdManComputeTruth_rec
// C: void If_DsdManComputeTruth_rec( If_DsdMan_t * p, int iDsd, word * pRes, unsigned char * pPermLits, int * pnSupp )
func (p *IfDsdMan) IfDsdManComputeTruth_rec(iDsd int, pRes []uint64, pPermLits []uint8, pnSupp *int) {
	panic("TODO, stubbed")
}

// If_DsdManComputeTruthPtr
// C: void If_DsdManComputeTruthPtr( If_DsdMan_t * p, int iDsd, unsigned char * pPermLits, word * pRes )
func (p *IfDsdMan) IfDsdManComputeTruthPtr(iDsd int, pPermLits []uint8, pRes []uint64) {
	panic("TODO, stubbed")
}

// If_DsdManComputeTruth
// C: word * If_DsdManComputeTruth( If_DsdMan_t * p, int iDsd, unsigned char * pPermLits )
func (p *IfDsdMan) IfDsdManComputeTruth(iDsd int, pPermLits []uint8) []uint64 {
	panic("TODO, stubbed")
}

// If_DsdManCheckInv_rec
// C: int If_DsdManCheckInv_rec( If_DsdMan_t * p, int iLit )
func (p *IfDsdMan) IfDsdManCheckInv_rec(iLit int) int {
	panic("TODO, stubbed")
}

// If_DsdManPushInv_rec
// C: int If_DsdManPushInv_rec( If_DsdMan_t * p, int iLit, unsigned char * pPerm )
func (p *IfDsdMan) IfDsdManPushInv_rec(iLit int, pPerm []uint8) int {
	panic("TODO, stubbed")
}

// If_DsdManPushInv
// C: int If_DsdManPushInv( If_DsdMan_t * p, int iLit, unsigned char * pPerm )
func (p *IfDsdMan) IfDsdManPushInv(iLit int, pPerm []uint8) int {
	panic("TODO, stubbed")
}

// If_DsdManComputeFirstArray
// C: int If_DsdManComputeFirstArray( If_DsdMan_t * p, int * pLits, int nLits, int * pFirsts )
func (p *IfDsdMan) IfDsdManComputeFirstArray(pLits []int, nLits int, pFirsts []int) int {
	panic("TODO, stubbed")
}

// If_DsdManComputeFirst
// C: int If_DsdManComputeFirst( If_DsdMan_t * p, If_DsdObj_t * pObj, int * pFirsts )
func (p *IfDsdMan) IfDsdManComputeFirst(pObj *IfDsdObj, pFirsts []int) int {
	panic("TODO, stubbed")
}

// If_DsdManOperation
// C: int If_DsdManOperation( If_DsdMan_t * p, int Type, int * pLits, int nLits, unsigned char * pPerm, word * pTruth )
func (p *IfDsdMan) IfDsdManOperation(Type int, pLits []int, nLits int, pPerm []uint8, pTruth []uint64) int {
	panic("TODO, stubbed")
}

// If_DsdMergeMatches
// C: void If_DsdMergeMatches( char * pDsd, int * pMatches )
func ifDsdMergeMatches(pDsd string, pMatches []int) {
	panic("TODO, stubbed")
}

// If_DsdManAddDsd_rec (p is a char ** cursor into pStr; here the offset into pStr)
// C: int If_DsdManAddDsd_rec( char * pStr, char ** p, int * pMatches, If_DsdMan_t * pMan, word * pTruth, unsigned char * pPerm, int * pnSupp )
func IfDsdManAddDsd_rec(pStr string, p *int, pMatches []int, pMan *IfDsdMan, pTruth []uint64, pPerm []uint8, pnSupp *int) int {
	panic("TODO, stubbed")
}

// If_DsdManAddDsd
// C: int If_DsdManAddDsd( If_DsdMan_t * p, char * pDsd, word * pTruth, unsigned char * pPerm, int * pnSupp )
func (p *IfDsdMan) IfDsdManAddDsd(pDsd string, pTruth []uint64, pPerm []uint8, pnSupp *int) int {
	panic("TODO, stubbed")
}

// If_DsdSign_rec
// C: unsigned If_DsdSign_rec( If_DsdMan_t * p, If_DsdObj_t * pObj, int * pnSupp )
func (p *IfDsdMan) IfDsdSign_rec(pObj *IfDsdObj, pnSupp *int) uint32 {
	panic("TODO, stubbed")
}

// If_DsdSign
// C: unsigned If_DsdSign( If_DsdMan_t * p, If_DsdObj_t * pObj, int iFan, int iFirst, int fShared )
func (p *IfDsdMan) IfDsdSign(pObj *IfDsdObj, iFan int, iFirst int, fShared int) uint32 {
	panic("TODO, stubbed")
}

// If_DsdManGetSuppSizes
// C: void If_DsdManGetSuppSizes( If_DsdMan_t * p, If_DsdObj_t * pObj, int * pSSizes )
func (p *IfDsdMan) IfDsdManGetSuppSizes(pObj *IfDsdObj, pSSizes []int) {
	panic("TODO, stubbed")
}

// If_DsdManCheckAndXor
// C: unsigned If_DsdManCheckAndXor( If_DsdMan_t * p, int iFirst, unsigned uMaskNot, If_DsdObj_t * pObj, int nSuppAll, int LutSize, int fDerive, int fVerbose )
func (p *IfDsdMan) IfDsdManCheckAndXor(iFirst int, uMaskNot uint32, pObj *IfDsdObj, nSuppAll int, LutSize int, fDerive int, fVerbose int) uint32 {
	panic("TODO, stubbed")
}

// If_DsdManCheckMux
// C: unsigned If_DsdManCheckMux( If_DsdMan_t * p, int iFirst, unsigned uMaskNot, If_DsdObj_t * pObj, int nSuppAll, int LutSize, int fDerive, int fVerbose )
func (p *IfDsdMan) IfDsdManCheckMux(iFirst int, uMaskNot uint32, pObj *IfDsdObj, nSuppAll int, LutSize int, fDerive int, fVerbose int) uint32 {
	panic("TODO, stubbed")
}

// If_DsdManCheckPrime
// C: unsigned If_DsdManCheckPrime( If_DsdMan_t * p, int iFirst, unsigned uMaskNot, If_DsdObj_t * pObj, int nSuppAll, int LutSize, int fDerive, int fVerbose )
func (p *IfDsdMan) IfDsdManCheckPrime(iFirst int, uMaskNot uint32, pObj *IfDsdObj, nSuppAll int, LutSize int, fDerive int, fVerbose int) uint32 {
	panic("TODO, stubbed")
}

// If_DsdManCheckXY_int
// C: unsigned If_DsdManCheckXY_int( If_DsdMan_t * p, int iDsd, int LutSize, int fDerive, unsigned uMaskNot, int fVerbose )
func (p *IfDsdMan) IfDsdManCheckXY_int(iDsd int, LutSize int, fDerive int, uMaskNot uint32, fVerbose int) uint32 {
	panic("TODO, stubbed")
}

// If_DsdManCheckXY
// C: unsigned If_DsdManCheckXY( If_DsdMan_t * p, int iDsd, int LutSize, int fDerive, unsigned uMaskNot, int fHighEffort, int fVerbose )
func (p *IfDsdMan) IfDsdManCheckXY(iDsd int, LutSize int, fDerive int, uMaskNot uint32, fHighEffort int, fVerbose int) uint32 {
	panic("TODO, stubbed")
}

// If_DsdManCheckXYZ
// C: unsigned If_DsdManCheckXYZ( If_DsdMan_t * p, int iDsd, int LutSize, int fDerive, int fVerbose )
func (p *IfDsdMan) IfDsdManCheckXYZ(iDsd int, LutSize int, fDerive int, fVerbose int) uint32 {
	panic("TODO, stubbed")
}

// If_DsdManCompute
// C: int If_DsdManCompute( If_DsdMan_t * p, word * pTruth, int nLeaves, unsigned char * pPerm, char * pLutStruct )
func (p *IfDsdMan) IfDsdManCompute(pTruth []uint64, nLeaves int, pPerm []uint8, pLutStruct string) int {
	panic("TODO, stubbed")
}

// If_DsdManTest
// C: void If_DsdManTest()
func IfDsdManTest() {
	panic("TODO, stubbed")
}

// If_CutDsdBalancePinDelays_rec
// C: int If_CutDsdBalancePinDelays_rec( If_DsdMan_t * p, int Id, int * pTimes, word * pRes, int * pnSupp, int nSuppAll, char * pPermLits )
func (p *IfDsdMan) IfCutDsdBalancePinDelays_rec(Id int, pTimes []int, pRes *uint64, pnSupp *int, nSuppAll int, pPermLits []int8) int {
	panic("TODO, stubbed")
}

// If_CutDsdBalancePinDelays
// C: int If_CutDsdBalancePinDelays( If_Man_t * p, If_Cut_t * pCut, char * pPerm )
func (p *IfMan) IfCutDsdBalancePinDelays(pCut *IfCut, pPerm []int8) int {
	panic("TODO, stubbed")
}

// If_CutDsdPermLitMax
// C: int If_CutDsdPermLitMax( char * pPermLits, int nVars, int iVar )
func IfCutDsdPermLitMax(pPermLits []int8, nVars int, iVar int) int {
	panic("TODO, stubbed")
}

// If_CutDsdBalanceEval_rec
// C: int If_CutDsdBalanceEval_rec( If_DsdMan_t * p, int Id, int * pTimes, int * pnSupp, Vec_Int_t * vAig, int * piLit, int nSuppAll, int * pArea, char * pPermLits )
func (p *IfDsdMan) IfCutDsdBalanceEval_rec(Id int, pTimes []int, pnSupp *int, vAig *IntVector, piLit *int, nSuppAll int, pArea *int, pPermLits []int8) int {
	panic("TODO, stubbed")
}

// If_CutDsdBalanceEvalInt
// C: int If_CutDsdBalanceEvalInt( If_DsdMan_t * p, int iDsd, int * pTimes, Vec_Int_t * vAig, int * pArea, char * pPermLits )
func (p *IfDsdMan) IfCutDsdBalanceEvalInt(iDsd int, pTimes []int, vAig *IntVector, pArea *int, pPermLits []int8) int {
	panic("TODO, stubbed")
}

// If_CutDsdBalanceEval
// C: int If_CutDsdBalanceEval( If_Man_t * p, If_Cut_t * pCut, Vec_Int_t * vAig )
func (p *IfMan) IfCutDsdBalanceEval(pCut *IfCut, vAig *IntVector) int {
	panic("TODO, stubbed")
}

// If_DsdManTune
// C: void If_DsdManTune( If_DsdMan_t * p, int LutSize, int fFast, int fAdd, int fSpec, int fVerbose )
func (p *IfDsdMan) IfDsdManTune(LutSize int, fFast int, fAdd int, fSpec int, fVerbose int) {
	panic("TODO, stubbed")
}

// Id_DsdManTuneStr1
// C: void Id_DsdManTuneStr1( If_DsdMan_t * p, char * pStruct, int nConfls, int fVerbose )
func (p *IfDsdMan) IdDsdManTuneStr1(pStruct string, nConfls int, fVerbose int) {
	panic("TODO, stubbed")
}

// Id_DsdManTuneStr (pthreads branch (ABC_USE_PTHREADS))
// C: void Id_DsdManTuneStr( If_DsdMan_t * p, char * pStruct, int nConfls, int nProcs, int fVerbose )
func (p *IfDsdMan) IdDsdManTuneStr(pStruct string, nConfls int, nProcs int, fVerbose int) {
	panic("TODO, stubbed")
}

// Ifn_WorkerThread (pArg is a void * that holds an Ifn_ThData_t *)
// C: void * Ifn_WorkerThread( void * pArg )
func IfnWorkerThread(pArg *IfnThData) any {
	panic("TODO, stubbed")
}

// Id_DsdManTuneThresh
// C: void Id_DsdManTuneThresh( If_DsdMan_t * p, int fUnate, int fThresh, int fThreshHeuristic, int fVerbose )
func (p *IfDsdMan) IdDsdManTuneThresh(fUnate int, fThresh int, fThreshHeuristic int, fVerbose int) {
	panic("TODO, stubbed")
}
