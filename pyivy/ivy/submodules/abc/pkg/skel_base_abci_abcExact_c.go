package abc

import "os"

// Hand-checked skeleton of src/base/abci/abcExact.c.

const (
	ABC_EXACT_SOL_NVARS  = 0
	ABC_EXACT_SOL_NFUNC  = 1
	ABC_EXACT_SOL_NGATES = 2
	SES_STORE_TABLE_SIZE = 1024
)

// The ANSI_COLOR_* string macros are not represented.

// SesMan is Ses_Man_t.
type SesMan struct {
	pSat               *SatSolver  // SAT solver
	pSpec              []uint64    // word * : specification
	bSpecInv           int         // remembers whether spec was inverted for normalization
	nSpecVars          int         // number of variables in specification
	nSpecFunc          int         // number of functions to synthesize
	nSpecWords         int         // number of words for function
	nRows              int         // number of rows in the specification (without 0)
	nMaxDepth          int         // maximum depth (-1 if depth is not constrained)
	nMaxDepthTmp       int         // temporary copy to modify nMaxDepth temporarily
	pArrTimeProfile    []int       // int * : arrival times of inputs (NULL if arrival times are ignored)
	pArrTimeProfileTmp [8]int      // temporary copy to modify pArrTimeProfile temporarily
	nArrTimeDelta      int         // delta to the original arrival times
	nArrTimeMax        int         // maximum normalized arrival time
	nBTLimit           int         // conflict limit
	fMakeAIG           int         // create AIG instead of general network
	fVerbose           int         // be verbose
	fVeryVerbose       int         // be very verbose
	fExtractVerbose    int         // be verbose about solution extraction
	fSatVerbose        int         // be verbose about SAT solving
	fReasonVerbose     int         // be verbose about give-up reasons
	pTtValues          [4]uint64   // word : truth table values to assign
	vPolar             *IntVector  // variables with positive polarity
	vAssump            *IntVector  // assumptions
	nRandRowAssigns    int         // number of random row assignments to initialize CEGAR
	fKeepRowAssigns    int         // if 1, keep counter examples in CEGAR for next number of gates
	nGates             int         // number of gates
	nStartGates        int         // number of gates to start search
	nMaxGates          int         // maximum number of gates given max. delay and arrival times
	fDecStructure      int         // set to 1 or higher if nSpecFunc = 1 and f = x_i OP g(X \ {x_i})
	pDecVars           int         // mask of variables that can be decomposed at top-level
	vStairDecVars      *IntVector  // list of stair decomposable variables
	pStairDecFunc      [8]int      // list of stair decomposable functions
	pTtObjs            [100]uint64 // word : temporary truth tables
	nSimVars           int         // number of simulation vars x(i, t)
	nOutputVars        int         // number of output variables g(h, i)
	nGateVars          int         // number of gate variables f(i, p, q)
	nSelectVars        int         // number of select variables s(i, j, k)
	nDepthVars         int         // number of depth variables d(i, j)
	nSimOffset         int         // offset where gate variables start
	nOutputOffset      int         // offset where output variables start
	nGateOffset        int         // offset where gate variables start
	nSelectOffset      int         // offset where select variables start
	nDepthOffset       int         // offset where depth variables start
	fHitResLimit       int         // SAT solver gave up due to resource limit
	timeSat            int64       // abctime : SAT runtime
	timeSatSat         int64       // abctime : SAT runtime (sat instance)
	timeSatUnsat       int64       // abctime : SAT runtime (unsat instance)
	timeSatUndef       int64       // abctime : SAT runtime (undef instance)
	timeInstance       int64       // abctime : creating instance runtime
	timeTotal          int64       // abctime : all runtime
	nSatCalls          int         // number of SAT calls
	nUnsatCalls        int         // number of UNSAT calls
	nUndefCalls        int         // number of UNDEF calls
	nDebugOffset       int         // for debug printing
}

// SesTimesEntry is Ses_TimesEntry_t.
type SesTimesEntry struct {
	pArrTimeProfile [8]int         // normalized arrival time profile
	fResLimit       int            // solution found after resource limit
	next            *SesTimesEntry // linked list pointer
	pNetwork        []byte         // char * : char array representation of optimum network
}

// SesTruthEntry is Ses_TruthEntry_t.
type SesTruthEntry struct {
	pTruth [4]uint64      // word : truth table for comparison
	nVars  int            // number of variables
	next   *SesTruthEntry // linked list pointer
	head   *SesTimesEntry // pointer to head of sub list with arrival times
}

// SesStore is Ses_Store_t.
type SesStore struct {
	fMakeAIG            int                                  // create AIG instead of general network
	fVerbose            int                                  // be verbose
	fVeryVerbose        int                                  // be very verbose
	nBTLimit            int                                  // conflict limit
	nEntriesCount       int                                  // number of entries
	nValidEntriesCount  int                                  // number of entries with network
	pEntries            [SES_STORE_TABLE_SIZE]*SesTruthEntry // hash table for truth table entries
	pSat                *SatSolver                           // own SAT solver instance
	pDebugEntries       *os.File                             // FILE * : debug unsynth. (rl) entries
	szDBName            string                               // char * : database file name
	nCutCount           uint64                               // unsigned long
	pCutCount           [9]uint64
	nUnsynthesizedImp   uint64
	pUnsynthesizedImp   [9]uint64
	nUnsynthesizedRL    uint64
	pUnsynthesizedRL    [9]uint64
	nSynthesizedTrivial uint64
	pSynthesizedTrivial [9]uint64
	nSynthesizedImp     uint64
	pSynthesizedImp     [9]uint64
	nSynthesizedRL      uint64
	pSynthesizedRL      [9]uint64
	nCacheHits          uint64
	pCacheHits          [9]uint64
	nSatCalls           uint64
	nUnsatCalls         uint64
	nUndefCalls         uint64
	timeExact           int64 // abctime
	timeSat             int64 // abctime
	timeSatSat          int64 // abctime
	timeSatUnsat        int64 // abctime
	timeSatUndef        int64 // abctime
	timeInstance        int64 // abctime
	timeTotal           int64 // abctime
}

// The file-level global s_pSesStore (Ses_Store_t *) is not represented.

// Abc_TtIsSubsetWithMask (static)
// C: int Abc_TtIsSubsetWithMask( word * pSmall, word * pLarge, word * pMask, int nWords )
func abcTtIsSubsetWithMask(pSmall []uint64, pLarge []uint64, pMask []uint64, nWords int) int {
	panic("TODO, stubbed")
}

// Abc_TtCofsOppositeWithMask (static)
// C: int Abc_TtCofsOppositeWithMask( word * pTruth, word * pMask, int nWords, int iVar )
func abcTtCofsOppositeWithMask(pTruth []uint64, pMask []uint64, nWords int, iVar int) int {
	panic("TODO, stubbed")
}

// Abc_TtIsTopDecomposable (static)
// C: int Abc_TtIsTopDecomposable( word * pTruth, word * pMask, int nWords, int iVar )
func abcTtIsTopDecomposable(pTruth []uint64, pMask []uint64, nWords int, iVar int) int {
	panic("TODO, stubbed")
}

// Abc_TtIsStairDecomposable (static)
// C: int Abc_TtIsStairDecomposable( word * pTruth, int nWords, int * pVars, int nSize, int * pStairFunc )
func abcTtIsStairDecomposable(pTruth []uint64, nWords int, pVars []int, nSize int, pStairFunc []int) int {
	panic("TODO, stubbed")
}

// Abc_DebugPrint (static)
// C: void Abc_DebugPrint( const char* str, int fCond )
func abcDebugPrint(str string, fCond int) {
	panic("TODO, stubbed")
}

// Abc_DebugPrintInt (static)
// C: void Abc_DebugPrintInt( const char* fmt, int n, int fCond )
func abcDebugPrintInt(fmt string, n int, fCond int) {
	panic("TODO, stubbed")
}

// Abc_DebugPrintIntInt (static)
// C: void Abc_DebugPrintIntInt( const char* fmt, int n1, int n2, int fCond )
func abcDebugPrintIntInt(fmt string, n1 int, n2 int, fCond int) {
	panic("TODO, stubbed")
}

// Abc_DebugErase (static)
// C: void Abc_DebugErase( int n, int fCond )
func abcDebugErase(n int, fCond int) {
	panic("TODO, stubbed")
}

// Abc_NormalizeArrivalTimes (static)
// C: int Abc_NormalizeArrivalTimes( int * pArrTimeProfile, int nVars, int * maxNormalized )
func abcNormalizeArrivalTimes(pArrTimeProfile []int, nVars int, maxNormalized *int) int {
	panic("TODO, stubbed")
}

// Ses_StoreAlloc (static)
// C: Ses_Store_t * Ses_StoreAlloc( int nBTLimit, int fMakeAIG, int fVerbose )
func sesStoreAlloc(nBTLimit int, fMakeAIG int, fVerbose int) *SesStore {
	panic("TODO, stubbed")
}

// Ses_StoreClean (static)
// C: void Ses_StoreClean( Ses_Store_t * pStore )
func (pStore *SesStore) sesStoreClean() {
	panic("TODO, stubbed")
}

// Ses_StoreTableHash (static)
// C: int Ses_StoreTableHash( word * pTruth, int nVars )
func sesStoreTableHash(pTruth []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// Ses_StoreTruthEqual (static)
// C: int Ses_StoreTruthEqual( Ses_TruthEntry_t * pEntry, word * pTruth, int nVars )
func sesStoreTruthEqual(pEntry *SesTruthEntry, pTruth []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// Ses_StoreTruthCopy (static)
// C: void Ses_StoreTruthCopy( Ses_TruthEntry_t * pEntry, word * pTruthSrc, int nVars )
func sesStoreTruthCopy(pEntry *SesTruthEntry, pTruthSrc []uint64, nVars int) {
	panic("TODO, stubbed")
}

// Ses_StoreTimesEqual (static)
// C: int Ses_StoreTimesEqual( int * pTimes1, int * pTimes2, int nVars )
func sesStoreTimesEqual(pTimes1 []int, pTimes2 []int, nVars int) int {
	panic("TODO, stubbed")
}

// Ses_StoreTimesCopy (static)
// C: void Ses_StoreTimesCopy( int * pTimesDest, int * pTimesSrc, int nVars )
func sesStoreTimesCopy(pTimesDest []int, pTimesSrc []int, nVars int) {
	panic("TODO, stubbed")
}

// Ses_StorePrintEntry (static)
// C: void Ses_StorePrintEntry( Ses_TruthEntry_t * pEntry, Ses_TimesEntry_t * pTiEntry )
func sesStorePrintEntry(pEntry *SesTruthEntry, pTiEntry *SesTimesEntry) {
	panic("TODO, stubbed")
}

// Ses_StorePrintDebugEntry (static)
// C: void Ses_StorePrintDebugEntry( Ses_Store_t * pStore, word * pTruth, int nVars, int * pNormalArrTime, int nMaxDepth, char * pSol, int nStartGates )
func (pStore *SesStore) sesStorePrintDebugEntry(pTruth []uint64, nVars int, pNormalArrTime []int, nMaxDepth int, pSol []byte, nStartGates int) {
	panic("TODO, stubbed")
}

// Abc_ExactNormalizeArrivalTimesForNetwork (static)
// C: void Abc_ExactNormalizeArrivalTimesForNetwork( int nVars, int * pArrTimeProfile, char * pSol )
func abcExactNormalizeArrivalTimesForNetwork(nVars int, pArrTimeProfile []int, pSol []byte) {
	panic("TODO, stubbed")
}

// Ses_StoreWrite (static)
// C: void Ses_StoreWrite( Ses_Store_t * pStore, const char * pFilename, int fSynthImp, int fSynthRL, int fUnsynthImp, int fUnsynthRL )
func (pStore *SesStore) sesStoreWrite(pFilename string, fSynthImp int, fSynthRL int, fUnsynthImp int, fUnsynthRL int) {
	panic("TODO, stubbed")
}

// Ses_StoreAddEntry
// C: int Ses_StoreAddEntry( Ses_Store_t * pStore, word * pTruth, int nVars, int * pArrTimeProfile, char * pSol, int fResLimit )
func (pStore *SesStore) SesStoreAddEntry(pTruth []uint64, nVars int, pArrTimeProfile []int, pSol []byte, fResLimit int) int {
	panic("TODO, stubbed")
}

// Ses_StoreGetEntrySimple
// C: int Ses_StoreGetEntrySimple( Ses_Store_t * pStore, word * pTruth, int nVars, int * pArrTimeProfile, char ** pSol )
func (pStore *SesStore) SesStoreGetEntrySimple(pTruth []uint64, nVars int, pArrTimeProfile []int, pSol *[]byte) int {
	panic("TODO, stubbed")
}

// Ses_StoreGetEntry
// C: int Ses_StoreGetEntry( Ses_Store_t * pStore, word * pTruth, int nVars, int * pArrTimeProfile, char ** pSol )
func (pStore *SesStore) SesStoreGetEntry(pTruth []uint64, nVars int, pArrTimeProfile []int, pSol *[]byte) int {
	panic("TODO, stubbed")
}

// Ses_StoreRead (static)
// C: void Ses_StoreRead( Ses_Store_t * pStore, const char * pFilename, int fSynthImp, int fSynthRL, int fUnsynthImp, int fUnsynthRL )
func (pStore *SesStore) sesStoreRead(pFilename string, fSynthImp int, fSynthRL int, fUnsynthImp int, fUnsynthRL int) {
	panic("TODO, stubbed")
}

// Ses_ManComputeTopDec (static)
// C: void Ses_ManComputeTopDec( Ses_Man_t * pSes )
func (pSes *SesMan) sesManComputeTopDec() {
	panic("TODO, stubbed")
}

// Ses_ManAlloc (static)
// C: Ses_Man_t * Ses_ManAlloc( word * pTruth, int nVars, int nFunc, int nMaxDepth, int * pArrTimeProfile, int fMakeAIG, int nBTLimit, int fVerbose )
func sesManAlloc(pTruth []uint64, nVars int, nFunc int, nMaxDepth int, pArrTimeProfile []int, fMakeAIG int, nBTLimit int, fVerbose int) *SesMan {
	panic("TODO, stubbed")
}

// Ses_ManCleanLight (static)
// C: void Ses_ManCleanLight( Ses_Man_t * pSes )
func (pSes *SesMan) sesManCleanLight() {
	panic("TODO, stubbed")
}

// Ses_ManClean (static)
// C: void Ses_ManClean( Ses_Man_t * pSes )
func (pSes *SesMan) sesManClean() {
	panic("TODO, stubbed")
}

// Ses_ManSimVar (static)
// C: int Ses_ManSimVar( Ses_Man_t * pSes, int i, int t )
func (pSes *SesMan) sesManSimVar(i int, t int) int {
	panic("TODO, stubbed")
}

// Ses_ManOutputVar (static)
// C: int Ses_ManOutputVar( Ses_Man_t * pSes, int h, int i )
func (pSes *SesMan) sesManOutputVar(h int, i int) int {
	panic("TODO, stubbed")
}

// Ses_ManGateVar (static)
// C: int Ses_ManGateVar( Ses_Man_t * pSes, int i, int p, int q )
func (pSes *SesMan) sesManGateVar(i int, p int, q int) int {
	panic("TODO, stubbed")
}

// Ses_ManSelectVar (static)
// C: int Ses_ManSelectVar( Ses_Man_t * pSes, int i, int j, int k )
func (pSes *SesMan) sesManSelectVar(i int, j int, k int) int {
	panic("TODO, stubbed")
}

// Ses_ManDepthVar (static)
// C: int Ses_ManDepthVar( Ses_Man_t * pSes, int i, int j )
func (pSes *SesMan) sesManDepthVar(i int, j int) int {
	panic("TODO, stubbed")
}

// Ses_ManDeriveTruth (static)
// C: word * Ses_ManDeriveTruth( Ses_Man_t * pSes, char * pSol, int fInvert )
func (pSes *SesMan) sesManDeriveTruth(pSol []byte, fInvert int) []uint64 {
	panic("TODO, stubbed")
}

// Ses_ManCreateVars (static)
// C: void Ses_ManCreateVars( Ses_Man_t * pSes, int nGates )
func (pSes *SesMan) sesManCreateVars(nGates int) {
	panic("TODO, stubbed")
}

// Ses_ManGateCannotHaveChild (static)
// C: void Ses_ManGateCannotHaveChild( Ses_Man_t * pSes, int i, int j )
func (pSes *SesMan) sesManGateCannotHaveChild(i int, j int) {
	panic("TODO, stubbed")
}

// Ses_ManCreateMainClause (static)
// C: void Ses_ManCreateMainClause( Ses_Man_t * pSes, int t, int i, int j, int k, int a, int b, int c )
func (pSes *SesMan) sesManCreateMainClause(t int, i int, j int, k int, a int, b int, c int) {
	panic("TODO, stubbed")
}

// Ses_ManCreateTruthTableClause (static inline)
// C: int Ses_ManCreateTruthTableClause( Ses_Man_t * pSes, int t )
func (pSes *SesMan) sesManCreateTruthTableClause(t int) int {
	panic("TODO, stubbed")
}

// Ses_ManCreateClauses (static)
// C: int Ses_ManCreateClauses( Ses_Man_t * pSes )
func (pSes *SesMan) sesManCreateClauses() int {
	panic("TODO, stubbed")
}

// Ses_ManCreateDepthClauses (static)
// C: int Ses_ManCreateDepthClauses( Ses_Man_t * pSes )
func (pSes *SesMan) sesManCreateDepthClauses() int {
	panic("TODO, stubbed")
}

// Sat_Wrd2Dbl (static)
// C: double Sat_Wrd2Dbl( word w )
func satWrd2Dbl(w uint64) float64 {
	panic("TODO, stubbed")
}

// Ses_ManSolve (static)
// C: int Ses_ManSolve( Ses_Man_t * pSes )
func (pSes *SesMan) sesManSolve() int {
	panic("TODO, stubbed")
}

// Ses_ManExtractSolution (static)
// C: char * Ses_ManExtractSolution( Ses_Man_t * pSes )
func (pSes *SesMan) sesManExtractSolution() []byte {
	panic("TODO, stubbed")
}

// Ses_ManExtractNtk (static)
// C: Abc_Ntk_t * Ses_ManExtractNtk( char const * pSol )
func sesManExtractNtk(pSol []byte) *ABCNetwork {
	panic("TODO, stubbed")
}

// Ses_ManExtractGia (static)
// C: Gia_Man_t * Ses_ManExtractGia( char const * pSol )
func sesManExtractGia(pSol []byte) *GiaManager {
	panic("TODO, stubbed")
}

// Ses_ManPrintRuntime (static)
// C: void Ses_ManPrintRuntime( Ses_Man_t * pSes )
func (pSes *SesMan) sesManPrintRuntime() {
	panic("TODO, stubbed")
}

// Ses_ManPrintFuncs (static)
// C: void Ses_ManPrintFuncs( Ses_Man_t * pSes )
func (pSes *SesMan) sesManPrintFuncs() {
	panic("TODO, stubbed")
}

// Ses_ManPrintVars (static)
// C: void Ses_ManPrintVars( Ses_Man_t * pSes )
func (pSes *SesMan) sesManPrintVars() {
	panic("TODO, stubbed")
}

// Ses_ManComputeMaxGates (static)
// C: void Ses_ManComputeMaxGates( Ses_Man_t * pSes )
func (pSes *SesMan) sesManComputeMaxGates() {
	panic("TODO, stubbed")
}

// Ses_CheckDepthConsistency (static)
// C: int Ses_CheckDepthConsistency( Ses_Man_t * pSes )
func (pSes *SesMan) sesCheckDepthConsistency() int {
	panic("TODO, stubbed")
}

// Ses_CheckGatesConsistency (static)
// C: int Ses_CheckGatesConsistency( Ses_Man_t * pSes, int nGates )
func (pSes *SesMan) sesCheckGatesConsistency(nGates int) int {
	panic("TODO, stubbed")
}

// Abc_ExactCopyDepthAndArrivalTimes (static)
// C: void Abc_ExactCopyDepthAndArrivalTimes( int nVars, int nDepthFrom, int * nDepthTo, int * pTimesFrom, int * pTimesTo )
func abcExactCopyDepthAndArrivalTimes(nVars int, nDepthFrom int, nDepthTo *int, pTimesFrom []int, pTimesTo []int) {
	panic("TODO, stubbed")
}

// Ses_ManStoreDepthAndArrivalTimes (static)
// C: void Ses_ManStoreDepthAndArrivalTimes( Ses_Man_t * pSes )
func (pSes *SesMan) sesManStoreDepthAndArrivalTimes() {
	panic("TODO, stubbed")
}

// Ses_ManRestoreDepthAndArrivalTimes (static)
// C: void Ses_ManRestoreDepthAndArrivalTimes( Ses_Man_t * pSes )
func (pSes *SesMan) sesManRestoreDepthAndArrivalTimes() {
	panic("TODO, stubbed")
}

// Abc_ExactAdjustDepthAndArrivalTimes (static)
// C: void Abc_ExactAdjustDepthAndArrivalTimes( int nVars, int nGates, int nDepthFrom, int * nDepthTo, int * pTimesFrom, int * pTimesTo, int nTimesMax )
func abcExactAdjustDepthAndArrivalTimes(nVars int, nGates int, nDepthFrom int, nDepthTo *int, pTimesFrom []int, pTimesTo []int, nTimesMax int) {
	panic("TODO, stubbed")
}

// Ses_AdjustDepthAndArrivalTimes (static)
// C: void Ses_AdjustDepthAndArrivalTimes( Ses_Man_t * pSes, int nGates )
func (pSes *SesMan) sesAdjustDepthAndArrivalTimes(nGates int) {
	panic("TODO, stubbed")
}

// Ses_ManFindNetworkExact (static)
// C: int Ses_ManFindNetworkExact( Ses_Man_t * pSes, int nGates )
func (pSes *SesMan) sesManFindNetworkExact(nGates int) int {
	panic("TODO, stubbed")
}

// Ses_ManFindNetworkExactCEGAR (static)
// C: int Ses_ManFindNetworkExactCEGAR( Ses_Man_t * pSes, int nGates, char ** pSol )
func (pSes *SesMan) sesManFindNetworkExactCEGAR(nGates int, pSol *[]byte) int {
	panic("TODO, stubbed")
}

// Ses_ManFindMinimumSizeBottomUp (static)
// C: char * Ses_ManFindMinimumSizeBottomUp( Ses_Man_t * pSes )
func (pSes *SesMan) sesManFindMinimumSizeBottomUp() []byte {
	panic("TODO, stubbed")
}

// Ses_ManFindMinimumSizeTopDown (static)
// C: char * Ses_ManFindMinimumSizeTopDown( Ses_Man_t * pSes, int nMinGates )
func (pSes *SesMan) sesManFindMinimumSizeTopDown(nMinGates int) []byte {
	panic("TODO, stubbed")
}

// Ses_ManFindMinimumSize (static)
// C: char * Ses_ManFindMinimumSize( Ses_Man_t * pSes )
func (pSes *SesMan) sesManFindMinimumSize() []byte {
	panic("TODO, stubbed")
}

// Abc_NtkFindExact
// C: Abc_Ntk_t * Abc_NtkFindExact( word * pTruth, int nVars, int nFunc, int nMaxDepth, int * pArrTimeProfile, int nBTLimit, int nStartGates, int fVerbose )
func AbcNtkFindExact(pTruth []uint64, nVars int, nFunc int, nMaxDepth int, pArrTimeProfile []int, nBTLimit int, nStartGates int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Gia_ManFindExact
// C: Gia_Man_t * Gia_ManFindExact( word * pTruth, int nVars, int nFunc, int nMaxDepth, int * pArrTimeProfile, int nBTLimit, int nStartGates, int fVerbose )
func GiaManFindExact(pTruth []uint64, nVars int, nFunc int, nMaxDepth int, pArrTimeProfile []int, nBTLimit int, nStartGates int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Abc_NtkFromTruthTable
// C: Abc_Ntk_t * Abc_NtkFromTruthTable( word * pTruth, int nVars )
func AbcNtkFromTruthTable(pTruth []uint64, nVars int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_ExactTestSingleOutput
// C: void Abc_ExactTestSingleOutput( int fVerbose )
func AbcExactTestSingleOutput(fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_ExactTestSingleOutputAIG
// C: void Abc_ExactTestSingleOutputAIG( int fVerbose )
func AbcExactTestSingleOutputAIG(fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_ExactTest
// C: void Abc_ExactTest( int fVerbose )
func AbcExactTest(fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_ExactIsRunning
// C: int Abc_ExactIsRunning()
func AbcExactIsRunning() int {
	panic("TODO, stubbed")
}

// Abc_ExactInputNum
// C: int Abc_ExactInputNum()
func AbcExactInputNum() int {
	panic("TODO, stubbed")
}

// Abc_ExactStart
// C: void Abc_ExactStart( int nBTLimit, int fMakeAIG, int fVerbose, int fVeryVerbose, const char * pFilename )
func AbcExactStart(nBTLimit int, fMakeAIG int, fVerbose int, fVeryVerbose int, pFilename string) {
	panic("TODO, stubbed")
}

// Abc_ExactStop
// C: void Abc_ExactStop( const char * pFilename )
func AbcExactStop(pFilename string) {
	panic("TODO, stubbed")
}

// Abc_ExactStats
// C: void Abc_ExactStats()
func AbcExactStats() {
	panic("TODO, stubbed")
}

// Abc_ExactDelayCost
// C: int Abc_ExactDelayCost( word * pTruth, int nVars, int * pArrTimeProfile, char * pPerm, int * Cost, int AigLevel )
func AbcExactDelayCost(pTruth []uint64, nVars int, pArrTimeProfile []int, pPerm []int8, Cost *int, AigLevel int) int {
	panic("TODO, stubbed")
}

// Abc_ExactBuildNode
// C: Abc_Obj_t * Abc_ExactBuildNode( word * pTruth, int nVars, int * pArrTimeProfile, Abc_Obj_t ** pFanins, Abc_Ntk_t * pNtk )
func AbcExactBuildNode(pTruth []uint64, nVars int, pArrTimeProfile []int, pFanins []*ABCNetworkObject, pNtk *ABCNetwork) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_ExactStoreTest
// C: void Abc_ExactStoreTest( int fVerbose )
func AbcExactStoreTest(fVerbose int) {
	panic("TODO, stubbed")
}
