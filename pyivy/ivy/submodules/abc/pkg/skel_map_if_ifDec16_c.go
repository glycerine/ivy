package abc

// Hand-checked skeleton of src/map/if/ifDec16.c.

const (
	CLU_VAR_MAX = 16
	CLU_WRD_MAX = 1 << (CLU_VAR_MAX - 6)
	CLU_MEM_MAX = 1000 // 1 GB
	CLU_UNUSED  = 0xff
)

// IfGrp is If_Grp_t: a decomposition.
type IfGrp struct {
	nVars int8
	nMyu  int8
	pVars [CLU_VAR_MAX]int8
}

// IfHte is If_Hte_t: a hash table entry.
type IfHte struct {
	pNext   *IfHte
	Group   uint32
	Counter uint32
	pTruth  []uint64 // word pTruth[1] : flexible trailing array
}

// The static data tables PMasks, Truth6 and TruthAll[CLU_VAR_MAX][CLU_WRD_MAX]
// and the file-level truth table memories are not represented.

// If_CluGrp2Uns
// C: unsigned If_CluGrp2Uns( If_Grp_t * pG )
func ifCluGrp2Uns(pG *IfGrp) uint32 {
	panic("TODO, stubbed")
}

// If_CluUns2Grp
// C: void If_CluUns2Grp( unsigned Group, If_Grp_t * pG )
func ifCluUns2Grp(Group uint32, pG *IfGrp) {
	panic("TODO, stubbed")
}

// If_CluPrimeCudd
// C: unsigned int If_CluPrimeCudd( unsigned int p )
func IfCluPrimeCudd(p uint32) uint32 {
	panic("TODO, stubbed")
}

// If_CluWordNum
// C: int If_CluWordNum( int nVars )
func ifCluWordNum(nVars int) int {
	panic("TODO, stubbed")
}

// If_CluCountOnes
// C: int If_CluCountOnes( word t )
func ifCluCountOnes(t uint64) int {
	panic("TODO, stubbed")
}

// If_CluHashTableCheck
// C: void If_CluHashTableCheck( If_Man_t * p )
func (p *IfMan) IfCluHashTableCheck() {
	panic("TODO, stubbed")
}

// If_CluHashPrintStats
// C: void If_CluHashPrintStats( If_Man_t * p, int t )
func (p *IfMan) IfCluHashPrintStats(t int) {
	panic("TODO, stubbed")
}

// If_CluHashFindMedian
// C: int If_CluHashFindMedian( If_Man_t * p, int t )
func (p *IfMan) IfCluHashFindMedian(t int) int {
	panic("TODO, stubbed")
}

// If_CluHashKey
// C: int If_CluHashKey( word * pTruth, int nWords, int Size )
func IfCluHashKey(pTruth []uint64, nWords int, Size int) int {
	panic("TODO, stubbed")
}

// If_CluHashLookup (unsigned * into the hash entry)
// C: unsigned * If_CluHashLookup( If_Man_t * p, word * pTruth, int t )
func (p *IfMan) IfCluHashLookup(pTruth []uint64, t int) *uint32 {
	panic("TODO, stubbed")
}

// If_CluClear
// C: void If_CluClear( word * pIn, int nVars )
func ifCluClear(pIn []uint64, nVars int) {
	panic("TODO, stubbed")
}

// If_CluFill
// C: void If_CluFill( word * pIn, int nVars )
func ifCluFill(pIn []uint64, nVars int) {
	panic("TODO, stubbed")
}

// If_CluCopy
// C: void If_CluCopy( word * pOut, word * pIn, int nVars )
func ifCluCopy(pOut []uint64, pIn []uint64, nVars int) {
	panic("TODO, stubbed")
}

// If_CluEqual
// C: int If_CluEqual( word * pOut, word * pIn, int nVars )
func ifCluEqual(pOut []uint64, pIn []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_CluAnd
// C: void If_CluAnd( word * pRes, word * pIn1, word * pIn2, int nVars )
func ifCluAnd(pRes []uint64, pIn1 []uint64, pIn2 []uint64, nVars int) {
	panic("TODO, stubbed")
}

// If_CluSharp
// C: void If_CluSharp( word * pRes, word * pIn1, word * pIn2, int nVars )
func ifCluSharp(pRes []uint64, pIn1 []uint64, pIn2 []uint64, nVars int) {
	panic("TODO, stubbed")
}

// If_CluOr
// C: void If_CluOr( word * pRes, word * pIn1, word * pIn2, int nVars )
func ifCluOr(pRes []uint64, pIn1 []uint64, pIn2 []uint64, nVars int) {
	panic("TODO, stubbed")
}

// If_CluAdjust
// C: word If_CluAdjust( word t, int nVars )
func ifCluAdjust(t uint64, nVars int) uint64 {
	panic("TODO, stubbed")
}

// If_CluAdjustBig
// C: void If_CluAdjustBig( word * pF, int nVarsCur, int nVarsMax )
func ifCluAdjustBig(pF []uint64, nVarsCur int, nVarsMax int) {
	panic("TODO, stubbed")
}

// If_CluSwapAdjacent
// C: void If_CluSwapAdjacent( word * pOut, word * pIn, int iVar, int nVars )
func ifCluSwapAdjacent(pOut []uint64, pIn []uint64, iVar int, nVars int) {
	panic("TODO, stubbed")
}

// If_CluChangePhase
// C: void If_CluChangePhase( word * pF, int nVars, int iVar )
func IfCluChangePhase(pF []uint64, nVars int, iVar int) {
	panic("TODO, stubbed")
}

// If_CluCountOnesInCofs
// C: void If_CluCountOnesInCofs( word * pTruth, int nVars, int * pStore )
func IfCluCountOnesInCofs(pTruth []uint64, nVars int, pStore []int) {
	panic("TODO, stubbed")
}

// If_CluSemiCanonicize
// C: unsigned If_CluSemiCanonicize( word * pTruth, int nVars, int * pCanonPerm )
func IfCluSemiCanonicize(pTruth []uint64, nVars int, pCanonPerm []int) uint32 {
	panic("TODO, stubbed")
}

// If_CluSemiCanonicizeVerify
// C: void If_CluSemiCanonicizeVerify( word * pTruth, word * pTruth0, int nVars, int * pCanonPerm, unsigned uCanonPhase )
func IfCluSemiCanonicizeVerify(pTruth []uint64, pTruth0 []uint64, nVars int, pCanonPerm []int, uCanonPhase uint32) {
	panic("TODO, stubbed")
}

// If_CluPrintGroup
// C: void If_CluPrintGroup( If_Grp_t * g )
func IfCluPrintGroup(g *IfGrp) {
	panic("TODO, stubbed")
}

// If_CluPrintConfig
// C: void If_CluPrintConfig( int nVars, If_Grp_t * g, If_Grp_t * r, word BStruth, word * pFStruth )
func IfCluPrintConfig(nVars int, g *IfGrp, r *IfGrp, BStruth uint64, pFStruth []uint64) {
	panic("TODO, stubbed")
}

// If_CluInitTruthTables
// C: void If_CluInitTruthTables()
func IfCluInitTruthTables() {
	panic("TODO, stubbed")
}

// If_CluComposeLut
// C: void If_CluComposeLut( int nVars, If_Grp_t * g, word * t, word f[6][CLU_WRD_MAX], word * r )
func ifCluComposeLut(nVars int, g *IfGrp, t []uint64, f *[6][CLU_WRD_MAX]uint64, r []uint64) {
	panic("TODO, stubbed")
}

// If_CluVerify
// C: void If_CluVerify( word * pF, int nVars, If_Grp_t * g, If_Grp_t * r, word BStruth, word * pFStruth )
func IfCluVerify(pF []uint64, nVars int, g *IfGrp, r *IfGrp, BStruth uint64, pFStruth []uint64) {
	panic("TODO, stubbed")
}

// If_CluVerify3
// C: void If_CluVerify3( word * pF, int nVars, If_Grp_t * g, If_Grp_t * g2, If_Grp_t * r, word BStruth, word BStruth2, word FStruth )
func IfCluVerify3(pF []uint64, nVars int, g *IfGrp, g2 *IfGrp, r *IfGrp, BStruth uint64, BStruth2 uint64, FStruth uint64) {
	panic("TODO, stubbed")
}

// If_CluSwapVars
// C: void If_CluSwapVars( word * pTruth, int nVars, int * V2P, int * P2V, int iVar, int jVar )
func IfCluSwapVars(pTruth []uint64, nVars int, V2P []int, P2V []int, iVar int, jVar int) {
	panic("TODO, stubbed")
}

// If_CluReverseOrder
// C: void If_CluReverseOrder( word * pTruth, int nVars, int * V2P, int * P2V, int iVarStart )
func IfCluReverseOrder(pTruth []uint64, nVars int, V2P []int, P2V []int, iVarStart int) {
	panic("TODO, stubbed")
}

// If_CluMoveVar2
// C: void If_CluMoveVar2( word * pF, int nVars, int * Var2Pla, int * Pla2Var, int v, int p )
func IfCluMoveVar2(pF []uint64, nVars int, Var2Pla []int, Pla2Var []int, v int, p int) {
	panic("TODO, stubbed")
}

// If_CluMoveVar
// C: void If_CluMoveVar( word * pF, int nVars, int * Var2Pla, int * Pla2Var, int v, int p )
func IfCluMoveVar(pF []uint64, nVars int, Var2Pla []int, Pla2Var []int, v int, p int) {
	panic("TODO, stubbed")
}

// If_CluMoveGroupToMsb
// C: void If_CluMoveGroupToMsb( word * pF, int nVars, int * V2P, int * P2V, If_Grp_t * g )
func IfCluMoveGroupToMsb(pF []uint64, nVars int, V2P []int, P2V []int, g *IfGrp) {
	panic("TODO, stubbed")
}

// If_CluReverseOrder_old
// C: void If_CluReverseOrder_old( word * pF, int nVars, int * V2P, int * P2V, int iVarStart )
func IfCluReverseOrder_old(pF []uint64, nVars int, V2P []int, P2V []int, iVarStart int) {
	panic("TODO, stubbed")
}

// If_CluCountCofs
// C: int If_CluCountCofs( word * pF, int nVars, int nBSsize, int iShift, word pCofs[3][CLU_WRD_MAX/4] )
func IfCluCountCofs(pF []uint64, nVars int, nBSsize int, iShift int, pCofs *[3][CLU_WRD_MAX / 4]uint64) int {
	panic("TODO, stubbed")
}

// If_CluCountCofs4
// C: int If_CluCountCofs4( word * pF, int nVars, int nBSsize, word pCofs[6][CLU_WRD_MAX/4] )
func IfCluCountCofs4(pF []uint64, nVars int, nBSsize int, pCofs *[6][CLU_WRD_MAX / 4]uint64) int {
	panic("TODO, stubbed")
}

// If_CluCofactors
// C: void If_CluCofactors( word * pF, int nVars, int iVar, word * pCof0, word * pCof1 )
func IfCluCofactors(pF []uint64, nVars int, iVar int, pCof0 []uint64, pCof1 []uint64) {
	panic("TODO, stubbed")
}

// If_CluDetectSpecialCaseCofs
// C: int If_CluDetectSpecialCaseCofs( word * pF, int nVars, int iVar )
func IfCluDetectSpecialCaseCofs(pF []uint64, nVars int, iVar int) int {
	panic("TODO, stubbed")
}

// If_CluDecUsingCofs
// C: If_Grp_t If_CluDecUsingCofs( word * pTruth, int nVars, int nLutLeaf )
func IfCluDecUsingCofs(pTruth []uint64, nVars int, nLutLeaf int) IfGrp {
	panic("TODO, stubbed")
}

// If_CluDeriveDisjoint
// C: word If_CluDeriveDisjoint( word * pF, int nVars, int * V2P, int * P2V, If_Grp_t * g, If_Grp_t * r )
func IfCluDeriveDisjoint(pF []uint64, nVars int, V2P []int, P2V []int, g *IfGrp, r *IfGrp) uint64 {
	panic("TODO, stubbed")
}

// If_CluDeriveDisjoint4
// C: void If_CluDeriveDisjoint4( word * pF, int nVars, int * V2P, int * P2V, If_Grp_t * g, If_Grp_t * r, word * pTruth0, word * pTruth1 )
func IfCluDeriveDisjoint4(pF []uint64, nVars int, V2P []int, P2V []int, g *IfGrp, r *IfGrp, pTruth0 []uint64, pTruth1 []uint64) {
	panic("TODO, stubbed")
}

// If_CluDeriveNonDisjoint
// C: word If_CluDeriveNonDisjoint( word * pF, int nVars, int * V2P, int * P2V, If_Grp_t * g, If_Grp_t * r )
func IfCluDeriveNonDisjoint(pF []uint64, nVars int, V2P []int, P2V []int, g *IfGrp, r *IfGrp) uint64 {
	panic("TODO, stubbed")
}

// If_CluCheckNonDisjointGroup
// C: int If_CluCheckNonDisjointGroup( word * pF, int nVars, int * V2P, int * P2V, If_Grp_t * g )
func IfCluCheckNonDisjointGroup(pF []uint64, nVars int, V2P []int, P2V []int, g *IfGrp) int {
	panic("TODO, stubbed")
}

// If_CluFindGroup
// C: If_Grp_t If_CluFindGroup( word * pF, int nVars, int iVarStart, int iVarStop, int * V2P, int * P2V, int nBSsize, int fDisjoint )
func IfCluFindGroup(pF []uint64, nVars int, iVarStart int, iVarStop int, V2P []int, P2V []int, nBSsize int, fDisjoint int) IfGrp {
	panic("TODO, stubbed")
}

// If_CluCheckGroup
// C: void If_CluCheckGroup( word * pTruth, int nVars, If_Grp_t * g )
func IfCluCheckGroup(pTruth []uint64, nVars int, g *IfGrp) {
	panic("TODO, stubbed")
}

// If_CluCheckPerm
// C: void If_CluCheckPerm( word * pTruth, word * pF, int nVars, int * V2P, int * P2V )
func IfCluCheckPerm(pTruth []uint64, pF []uint64, nVars int, V2P []int, P2V []int) {
	panic("TODO, stubbed")
}

// If_CluSuppIsMinBase
// C: int If_CluSuppIsMinBase( int Supp )
func ifCluSuppIsMinBase(Supp int) int {
	panic("TODO, stubbed")
}

// If_CluHasVar
// C: int If_CluHasVar( word * t, int nVars, int iVar )
func ifCluHasVar(t []uint64, nVars int, iVar int) int {
	panic("TODO, stubbed")
}

// If_CluSupport
// C: int If_CluSupport( word * t, int nVars )
func ifCluSupport(t []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_CluSupportSize
// C: int If_CluSupportSize( word * t, int nVars )
func IfCluSupportSize(t []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_CluTruthShrink
// C: void If_CluTruthShrink( word * pF, int nVars, int nVarsAll, unsigned Phase )
func ifCluTruthShrink(pF []uint64, nVars int, nVarsAll int, Phase uint32) {
	panic("TODO, stubbed")
}

// If_CluMinimumBase
// C: int If_CluMinimumBase( word * t, int * pSupp, int nVarsAll, int * pnVars )
func IfCluMinimumBase(t []uint64, pSupp []int, nVarsAll int, pnVars *int) int {
	panic("TODO, stubbed")
}

// If_CluCheck
// C: If_Grp_t If_CluCheck( If_Man_t * p, word * pTruth0, int nVars, int iVarStart, int iVarStop, int nLutLeaf, int nLutRoot, If_Grp_t * pR, word * pFunc0, word * pFunc1, word * pLeftOver, int fHashing )
func (p *IfMan) IfCluCheck(pTruth0 []uint64, nVars int, iVarStart int, iVarStop int, nLutLeaf int, nLutRoot int, pR *IfGrp, pFunc0 []uint64, pFunc1 []uint64, pLeftOver []uint64, fHashing int) IfGrp {
	panic("TODO, stubbed")
}

// If_CluCheckDecInAny
// C: int If_CluCheckDecInAny( word t, int nVars )
func IfCluCheckDecInAny(t uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_CluCheckDecIn
// C: int If_CluCheckDecIn( word t, int nVars )
func IfCluCheckDecIn(t uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_CluCheckDecInU
// C: int If_CluCheckDecInU( word t, int nVars )
func IfCluCheckDecInU(t uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_CluCheckDecOut
// C: int If_CluCheckDecOut( word t, int nVars )
func IfCluCheckDecOut(t uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_CluCheckDecOutU
// C: int If_CluCheckDecOutU( word t, int nVars )
func IfCluCheckDecOutU(t uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_CutPerformCheck45
// C: int If_CutPerformCheck45( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr )
func (p *IfMan) IfCutPerformCheck45(pTruth []uint32, nVars int, nLeaves int, pStr string) int {
	panic("TODO, stubbed")
}

// If_CutPerformCheck54
// C: int If_CutPerformCheck54( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr )
func (p *IfMan) IfCutPerformCheck54(pTruth []uint32, nVars int, nLeaves int, pStr string) int {
	panic("TODO, stubbed")
}

// If_CluCheck3
// C: If_Grp_t If_CluCheck3( If_Man_t * p, word * pTruth0, int nVars, int nLutLeaf, int nLutLeaf2, int nLutRoot, If_Grp_t * pR, If_Grp_t * pG2, word * pFunc0, word * pFunc1, word * pFunc2 )
func (p *IfMan) IfCluCheck3(pTruth0 []uint64, nVars int, nLutLeaf int, nLutLeaf2 int, nLutRoot int, pR *IfGrp, pG2 *IfGrp, pFunc0 []uint64, pFunc1 []uint64, pFunc2 []uint64) IfGrp {
	panic("TODO, stubbed")
}

// If_CluCheckExt (pLut0/pLut1 are char * receiving a memcpy of an If_Grp_t; pMan is a void *)
// C: int If_CluCheckExt( void * pMan, word * pTruth, int nVars, int nLutLeaf, int nLutRoot, char * pLut0, char * pLut1, word * pFunc0, word * pFunc1 )
func IfCluCheckExt(pMan *IfMan, pTruth []uint64, nVars int, nLutLeaf int, nLutRoot int, pLut0 []byte, pLut1 []byte, pFunc0 []uint64, pFunc1 []uint64) int {
	panic("TODO, stubbed")
}

// If_CluCheckExt3 (pLut0/pLut1/pLut2 are char * receiving a memcpy of an If_Grp_t; pMan is a void *)
// C: int If_CluCheckExt3( void * pMan, word * pTruth, int nVars, int nLutLeaf, int nLutLeaf2, int nLutRoot, char * pLut0, char * pLut1, char * pLut2, word * pFunc0, word * pFunc1, word * pFunc2 )
func IfCluCheckExt3(pMan *IfMan, pTruth []uint64, nVars int, nLutLeaf int, nLutLeaf2 int, nLutRoot int, pLut0 []byte, pLut1 []byte, pLut2 []byte, pFunc0 []uint64, pFunc1 []uint64, pFunc2 []uint64) int {
	panic("TODO, stubbed")
}

// If_CluDelayMax
// C: float If_CluDelayMax( If_Grp_t * g, float * pDelays )
func IfCluDelayMax(g *IfGrp, pDelays []float32) float32 {
	panic("TODO, stubbed")
}

// If_CutDelayLutStruct
// C: float If_CutDelayLutStruct( If_Man_t * p, If_Cut_t * pCut, char * pStr, float WireDelay )
func (p *IfMan) IfCutDelayLutStruct(pCut *IfCut, pStr string, WireDelay float32) float32 {
	panic("TODO, stubbed")
}

// If_TtMemCutNum
// C: int If_TtMemCutNum()
func IfTtMemCutNum() int {
	panic("TODO, stubbed")
}

// If_TtMemCutNum2
// C: int If_TtMemCutNum2()
func IfTtMemCutNum2() int {
	panic("TODO, stubbed")
}

// If_CutPerformCheck16
// C: int If_CutPerformCheck16( If_Man_t * p, unsigned * pTruth0, int nVars, int nLeaves, char * pStr )
func (p *IfMan) IfCutPerformCheck16(pTruth0 []uint32, nVars int, nLeaves int, pStr string) int {
	panic("TODO, stubbed")
}

// If_CluTest
// C: void If_CluTest()
func IfCluTest() {
	panic("TODO, stubbed")
}
