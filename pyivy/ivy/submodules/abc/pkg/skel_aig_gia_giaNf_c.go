package abc

// Hand-checked skeleton of src/aig/gia/giaNf.c.

const (
	NF_LEAF_MAX = 6
	NF_CUT_MAX  = 32
	NF_NO_LEAF  = 31
	NF_NO_FUNC  = 0x3FFFFFF
	NF_EPSILON  = 0.001
)

// NfCut is Nf_Cut_t.
type NfCut struct {
	Sign    uint64               // word : signature
	Delay   int                  // delay
	Flow    float32              // flow
	iFunc   uint32               // unsigned : 26 : function (NF_NO_FUNC)
	Useless uint32               // unsigned : 1 : function
	nLeaves uint32               // unsigned : 5 : leaf number (NF_NO_LEAF)
	pLeaves [NF_LEAF_MAX + 1]int // leaves
}

// NfCfg is Nf_Cfg_t.
type NfCfg struct {
	fCompl uint32 // unsigned : 1 : complemented
	Phase  uint32 // unsigned : 7 : match phase
	Perm   uint32 // unsigned : 24 : match permutation
}

// NfMat is Nf_Mat_t.
type NfMat struct {
	Gate   uint32  // unsigned : 20 : gate
	CutH   uint32  // unsigned : 10 : cut handle
	fCompl uint32  // unsigned : 1 : complemented
	fBest  uint32  // unsigned : 1 : best cut
	Cfg    NfCfg   // input literals
	D      int     // delay
	F      float32 // area
}

// NfObj is Nf_Obj_t.
type NfObj struct {
	M [2][2]NfMat // del/area (2x)
}

// NfMan is Nf_Man_t.
type NfMan struct {
	pGia       *GiaManager        // derived manager
	pPars      *JfPar             // parameters
	vTtMem     *WordMemory        // truth tables
	vTt2Match  *IntLevelVector    // matches for truth tables
	pCells     []MioCell2         // Mio_Cell2_t * : library gates
	nCells     int                // library gate count
	pNfObjs    []NfObj            // Nf_Obj_t * : best cuts
	vPages     PointerVector[int] // Vec_Ptr_t (by value) : cut memory
	vCutSets   IntVector          // Vec_Int_t (by value) : cut offsets
	vMapRefs   IntVector          // Vec_Int_t (by value) : mapping refs   (2x)
	vFlowRefs  FloatVector        // Vec_Flt_t (by value) : flow refs      (2x)
	vRequired  IntVector          // Vec_Int_t (by value) : required times (2x)
	vCutFlows  FloatVector        // Vec_Flt_t (by value) : temporary cut area
	vCutDelays IntVector          // Vec_Int_t (by value) : temporary cut delay
	vBackup    IntVector          // Vec_Int_t (by value) : backup literals
	iCur       int                // current position
	Iter       int                // mapping iterations
	fUseEla    int                // use exact area
	nInvs      int                // the inverter count
	InvDelayI  int                // inverter delay
	InvAreaW   uint64             // word : inverter delay
	InvAreaF   float32            // inverter area
	clkStart   int64              // abctime : starting time
	CutCount   [6]float64         // double : cut counts
	nCutUseAll int                // objects with useful cuts
}

// Nf_Cfg2Int (static inline)
// C: int Nf_Cfg2Int( Nf_Cfg_t Mat )
func nfCfg2Int(Mat NfCfg) int {
	panic("TODO, stubbed")
}

// Nf_Int2Cfg (static inline)
// C: Nf_Cfg_t Nf_Int2Cfg( int Int )
func nfInt2Cfg(Int int) NfCfg {
	panic("TODO, stubbed")
}

// Nf_ManObj (static inline)
// C: Nf_Obj_t * Nf_ManObj( Nf_Man_t * p, int i )
func (p *NfMan) nfManObj(i int) *NfObj {
	panic("TODO, stubbed")
}

// Nf_ManCell (static inline)
// C: Mio_Cell2_t* Nf_ManCell( Nf_Man_t * p, int i )
func (p *NfMan) nfManCell(i int) *MioCell2 {
	panic("TODO, stubbed")
}

// Nf_ManCutSet (static inline)
// C: int * Nf_ManCutSet( Nf_Man_t * p, int i )
func (p *NfMan) nfManCutSet(i int) []int {
	panic("TODO, stubbed")
}

// Nf_ObjCutSetId (static inline)
// C: int Nf_ObjCutSetId( Nf_Man_t * p, int i )
func (p *NfMan) nfObjCutSetId(i int) int {
	panic("TODO, stubbed")
}

// Nf_ObjCutSet (static inline)
// C: int * Nf_ObjCutSet( Nf_Man_t * p, int i )
func (p *NfMan) nfObjCutSet(i int) []int {
	panic("TODO, stubbed")
}

// Nf_ObjHasCuts (static inline)
// C: int Nf_ObjHasCuts( Nf_Man_t * p, int i )
func (p *NfMan) nfObjHasCuts(i int) int {
	panic("TODO, stubbed")
}

// Nf_ObjCutBest (static inline)
// C: int * Nf_ObjCutBest( Nf_Man_t * p, int i )
func (p *NfMan) nfObjCutBest(i int) []int {
	panic("TODO, stubbed")
}

// Nf_ObjCutUseless (static inline)
// C: int Nf_ObjCutUseless( Nf_Man_t * p, int TruthId )
func (p *NfMan) nfObjCutUseless(TruthId int) int {
	panic("TODO, stubbed")
}

// Nf_ObjCutFlow (static inline)
// C: float Nf_ObjCutFlow( Nf_Man_t * p, int i )
func (p *NfMan) nfObjCutFlow(i int) float32 {
	panic("TODO, stubbed")
}

// Nf_ObjCutDelay (static inline)
// C: int Nf_ObjCutDelay( Nf_Man_t * p, int i )
func (p *NfMan) nfObjCutDelay(i int) int {
	panic("TODO, stubbed")
}

// Nf_ObjSetCutFlow (static inline)
// C: void Nf_ObjSetCutFlow( Nf_Man_t * p, int i, float a )
func (p *NfMan) nfObjSetCutFlow(i int, a float32) {
	panic("TODO, stubbed")
}

// Nf_ObjSetCutDelay (static inline)
// C: void Nf_ObjSetCutDelay( Nf_Man_t * p, int i, int d )
func (p *NfMan) nfObjSetCutDelay(i int, d int) {
	panic("TODO, stubbed")
}

// Nf_ObjMapRefNum (static inline)
// C: int Nf_ObjMapRefNum( Nf_Man_t * p, int i, int c )
func (p *NfMan) nfObjMapRefNum(i int, c int) int {
	panic("TODO, stubbed")
}

// Nf_ObjMapRefInc (static inline)
// C: int Nf_ObjMapRefInc( Nf_Man_t * p, int i, int c )
func (p *NfMan) nfObjMapRefInc(i int, c int) int {
	panic("TODO, stubbed")
}

// Nf_ObjMapRefDec (static inline)
// C: int Nf_ObjMapRefDec( Nf_Man_t * p, int i, int c )
func (p *NfMan) nfObjMapRefDec(i int, c int) int {
	panic("TODO, stubbed")
}

// Nf_ObjFlowRefs (static inline)
// C: float Nf_ObjFlowRefs( Nf_Man_t * p, int i, int c )
func (p *NfMan) nfObjFlowRefs(i int, c int) float32 {
	panic("TODO, stubbed")
}

// Nf_ObjRequired (static inline)
// C: int Nf_ObjRequired( Nf_Man_t * p, int i, int c )
func (p *NfMan) nfObjRequired(i int, c int) int {
	panic("TODO, stubbed")
}

// Nf_ObjSetRequired (static inline)
// C: void Nf_ObjSetRequired( Nf_Man_t * p,int i, int c, int f )
func (p *NfMan) nfObjSetRequired(i int, c int, f int) {
	panic("TODO, stubbed")
}

// Nf_ObjUpdateRequired (static inline)
// C: void Nf_ObjUpdateRequired( Nf_Man_t * p,int i, int c, int f )
func (p *NfMan) nfObjUpdateRequired(i int, c int, f int) {
	panic("TODO, stubbed")
}

// Nf_ObjMatchD (static inline)
// C: Nf_Mat_t * Nf_ObjMatchD( Nf_Man_t * p, int i, int c )
func (p *NfMan) nfObjMatchD(i int, c int) *NfMat {
	panic("TODO, stubbed")
}

// Nf_ObjMatchA (static inline)
// C: Nf_Mat_t * Nf_ObjMatchA( Nf_Man_t * p, int i, int c )
func (p *NfMan) nfObjMatchA(i int, c int) *NfMat {
	panic("TODO, stubbed")
}

// Nf_CutSize (static inline)
// C: int Nf_CutSize( int * pCut )
func nfCutSize(pCut []int) int {
	panic("TODO, stubbed")
}

// Nf_CutFunc (static inline)
// C: int Nf_CutFunc( int * pCut )
func nfCutFunc(pCut []int) int {
	panic("TODO, stubbed")
}

// Nf_CutLeaves (static inline)
// C: int * Nf_CutLeaves( int * pCut )
func nfCutLeaves(pCut []int) []int {
	panic("TODO, stubbed")
}

// Nf_CutSetBoth (static inline)
// C: int Nf_CutSetBoth( int n, int f )
func nfCutSetBoth(n int, f int) int {
	panic("TODO, stubbed")
}

// Nf_CutIsTriv (static inline)
// C: int Nf_CutIsTriv( int * pCut, int i )
func nfCutIsTriv(pCut []int, i int) int {
	panic("TODO, stubbed")
}

// Nf_CutHandle (static inline)
// C: int Nf_CutHandle( int * pCutSet, int * pCut )
func nfCutHandle(pCutSet []int, pCut []int) int {
	panic("TODO, stubbed")
}

// Nf_CutFromHandle (static inline)
// C: int * Nf_CutFromHandle( int * pCutSet, int h )
func nfCutFromHandle(pCutSet []int, h int) []int {
	panic("TODO, stubbed")
}

// Nf_CfgVar (static inline)
// C: int Nf_CfgVar( Nf_Cfg_t Cfg, int i )
func nfCfgVar(Cfg NfCfg, i int) int {
	panic("TODO, stubbed")
}

// Nf_CfgCompl (static inline)
// C: int Nf_CfgCompl( Nf_Cfg_t Cfg, int i )
func nfCfgCompl(Cfg NfCfg, i int) int {
	panic("TODO, stubbed")
}

// Nf_StoCellIsDominated
// C: int Nf_StoCellIsDominated( Mio_Cell2_t * pCell, int * pFans, int * pProf )
func NfStoCellIsDominated(pCell *MioCell2, pFans []int, pProf []int) int {
	panic("TODO, stubbed")
}

// Nf_StoCreateGateAdd
// C: void Nf_StoCreateGateAdd( Vec_Mem_t * vTtMem, Vec_Wec_t * vTt2Match, Mio_Cell2_t * pCell, word uTruth, int * pFans, int nFans, Vec_Wec_t * vProfs, Vec_Int_t * vStore, int fPinFilter, int fPinPerm, int fPinQuick )
func NfStoCreateGateAdd(vTtMem *WordMemory, vTt2Match *IntLevelVector, pCell *MioCell2, uTruth uint64, pFans []int, nFans int, vProfs *IntLevelVector, vStore *IntVector, fPinFilter int, fPinPerm int, fPinQuick int) {
	panic("TODO, stubbed")
}

// Nf_StoCreateGateMaches
// C: void Nf_StoCreateGateMaches( Vec_Mem_t * vTtMem, Vec_Wec_t * vTt2Match, Mio_Cell2_t * pCell, int ** pComp, int ** pPerm, int * pnPerms, Vec_Wec_t * vProfs, Vec_Int_t * vStore, int fPinFilter, int fPinPerm, int fPinQuick )
func NfStoCreateGateMaches(vTtMem *WordMemory, vTt2Match *IntLevelVector, pCell *MioCell2, pComp [][]int, pPerm [][]int, pnPerms []int, vProfs *IntLevelVector, vStore *IntVector, fPinFilter int, fPinPerm int, fPinQuick int) {
	panic("TODO, stubbed")
}

// Nf_StoDeriveMatches
// C: Mio_Cell2_t * Nf_StoDeriveMatches( Vec_Mem_t * vTtMem, Vec_Wec_t * vTt2Match, int * pnCells, int fPinFilter, int fPinPerm, int fPinQuick )
func NfStoDeriveMatches(vTtMem *WordMemory, vTt2Match *IntLevelVector, pnCells *int, fPinFilter int, fPinPerm int, fPinQuick int) []MioCell2 {
	panic("TODO, stubbed")
}

// Nf_StoPrintOne
// C: void Nf_StoPrintOne( Nf_Man_t * p, int Count, int t, int i, int GateId, Nf_Cfg_t Mat )
func (p *NfMan) NfStoPrintOne(Count int, t int, i int, GateId int, Mat NfCfg) {
	panic("TODO, stubbed")
}

// Nf_StoPrint
// C: void Nf_StoPrint( Nf_Man_t * p, int fVerbose )
func (p *NfMan) NfStoPrint(fVerbose int) {
	panic("TODO, stubbed")
}

// Nf_StoCreate
// C: Nf_Man_t * Nf_StoCreate( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) NfStoCreate(pPars *JfPar) *NfMan {
	panic("TODO, stubbed")
}

// Nf_StoDelete
// C: void Nf_StoDelete( Nf_Man_t * p )
func (p *NfMan) NfStoDelete() {
	panic("TODO, stubbed")
}

// Nf_CutComputeTruth6 (static inline)
// C: int Nf_CutComputeTruth6( Nf_Man_t * p, Nf_Cut_t * pCut0, Nf_Cut_t * pCut1, int fCompl0, int fCompl1, Nf_Cut_t * pCutR, int fIsXor )
func (p *NfMan) nfCutComputeTruth6(pCut0 *NfCut, pCut1 *NfCut, fCompl0 int, fCompl1 int, pCutR *NfCut, fIsXor int) int {
	panic("TODO, stubbed")
}

// Nf_CutComputeTruthMux6 (static inline)
// C: int Nf_CutComputeTruthMux6( Nf_Man_t * p, Nf_Cut_t * pCut0, Nf_Cut_t * pCut1, Nf_Cut_t * pCutC, int fCompl0, int fCompl1, int fComplC, Nf_Cut_t * pCutR )
func (p *NfMan) nfCutComputeTruthMux6(pCut0 *NfCut, pCut1 *NfCut, pCutC *NfCut, fCompl0 int, fCompl1 int, fComplC int, pCutR *NfCut) int {
	panic("TODO, stubbed")
}

// Nf_CutCountBits (static inline)
// C: int Nf_CutCountBits( word i )
func nfCutCountBits(i uint64) int {
	panic("TODO, stubbed")
}

// Nf_CutGetSign (static inline)
// C: word Nf_CutGetSign( int * pLeaves, int nLeaves )
func nfCutGetSign(pLeaves []int, nLeaves int) uint64 {
	panic("TODO, stubbed")
}

// Nf_CutCreateUnit (static inline)
// C: int Nf_CutCreateUnit( Nf_Cut_t * p, int i )
func nfCutCreateUnit(p *NfCut, i int) int {
	panic("TODO, stubbed")
}

// Nf_CutPrint (static inline)
// C: void Nf_CutPrint( Nf_Man_t * p, Nf_Cut_t * pCut )
func (p *NfMan) nfCutPrint(pCut *NfCut) {
	panic("TODO, stubbed")
}

// Nf_ManPrepareCuts (static inline)
// C: int Nf_ManPrepareCuts( Nf_Cut_t * pCuts, Nf_Man_t * p, int iObj, int fAddUnit )
func nfManPrepareCuts(pCuts []NfCut, p *NfMan, iObj int, fAddUnit int) int {
	panic("TODO, stubbed")
}

// Nf_ManSaveCuts (static inline)
// C: int Nf_ManSaveCuts( Nf_Man_t * p, Nf_Cut_t ** pCuts, int nCuts, int fUseful )
func (p *NfMan) nfManSaveCuts(pCuts []*NfCut, nCuts int, fUseful int) int {
	panic("TODO, stubbed")
}

// Nf_ManCountUseful (static inline)
// C: int Nf_ManCountUseful( Nf_Cut_t ** pCuts, int nCuts )
func nfManCountUseful(pCuts []*NfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Nf_ManCountMatches (static inline)
// C: int Nf_ManCountMatches( Nf_Man_t * p, Nf_Cut_t ** pCuts, int nCuts )
func (p *NfMan) nfManCountMatches(pCuts []*NfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Nf_CutCheck (static inline)
// C: int Nf_CutCheck( Nf_Cut_t * pBase, Nf_Cut_t * pCut )
func nfCutCheck(pBase *NfCut, pCut *NfCut) int {
	panic("TODO, stubbed")
}

// Nf_SetCheckArray (static inline)
// C: int Nf_SetCheckArray( Nf_Cut_t ** ppCuts, int nCuts )
func nfSetCheckArray(ppCuts []*NfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Nf_CutMergeOrder (static inline)
// C: int Nf_CutMergeOrder( Nf_Cut_t * pCut0, Nf_Cut_t * pCut1, Nf_Cut_t * pCut, int nLutSize )
func nfCutMergeOrder(pCut0 *NfCut, pCut1 *NfCut, pCut *NfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Nf_CutMergeOrderMux (static inline)
// C: int Nf_CutMergeOrderMux( Nf_Cut_t * pCut0, Nf_Cut_t * pCut1, Nf_Cut_t * pCut2, Nf_Cut_t * pCut, int nLutSize )
func nfCutMergeOrderMux(pCut0 *NfCut, pCut1 *NfCut, pCut2 *NfCut, pCut *NfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Nf_SetCutIsContainedOrder (static inline)
// C: int Nf_SetCutIsContainedOrder( Nf_Cut_t * pBase, Nf_Cut_t * pCut )
func nfSetCutIsContainedOrder(pBase *NfCut, pCut *NfCut) int {
	panic("TODO, stubbed")
}

// Nf_SetLastCutIsContained (static inline)
// C: int Nf_SetLastCutIsContained( Nf_Cut_t ** pCuts, int nCuts )
func nfSetLastCutIsContained(pCuts []*NfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Nf_SetLastCutContainsArea (static inline)
// C: int Nf_SetLastCutContainsArea( Nf_Cut_t ** pCuts, int nCuts )
func nfSetLastCutContainsArea(pCuts []*NfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Nf_CutCompareArea (static inline)
// C: int Nf_CutCompareArea( Nf_Cut_t * pCut0, Nf_Cut_t * pCut1 )
func nfCutCompareArea(pCut0 *NfCut, pCut1 *NfCut) int {
	panic("TODO, stubbed")
}

// Nf_SetSortByArea (static inline)
// C: void Nf_SetSortByArea( Nf_Cut_t ** pCuts, int nCuts )
func nfSetSortByArea(pCuts []*NfCut, nCuts int) {
	panic("TODO, stubbed")
}

// Nf_SetAddCut (static inline)
// C: int Nf_SetAddCut( Nf_Cut_t ** pCuts, int nCuts, int nCutNum )
func nfSetAddCut(pCuts []*NfCut, nCuts int, nCutNum int) int {
	panic("TODO, stubbed")
}

// Nf_CutArea (static inline)
// C: int Nf_CutArea( Nf_Man_t * p, int nLeaves )
func (p *NfMan) nfCutArea(nLeaves int) int {
	panic("TODO, stubbed")
}

// Nf_CutParams (static inline)
// C: void Nf_CutParams( Nf_Man_t * p, Nf_Cut_t * pCut, float FlowRefs )
func (p *NfMan) nfCutParams(pCut *NfCut, FlowRefs float32) {
	panic("TODO, stubbed")
}

// Nf_ObjMergeOrder
// C: void Nf_ObjMergeOrder( Nf_Man_t * p, int iObj )
func (p *NfMan) NfObjMergeOrder(iObj int) {
	panic("TODO, stubbed")
}

// Nf_ManComputeCuts
// C: void Nf_ManComputeCuts( Nf_Man_t * p )
func (p *NfMan) NfManComputeCuts() {
	panic("TODO, stubbed")
}

// Nf_ManPrintStats
// C: void Nf_ManPrintStats( Nf_Man_t * p, char * pTitle )
func (p *NfMan) NfManPrintStats(pTitle string) {
	panic("TODO, stubbed")
}

// Nf_ManPrintInit
// C: void Nf_ManPrintInit( Nf_Man_t * p )
func (p *NfMan) NfManPrintInit() {
	panic("TODO, stubbed")
}

// Nf_ManPrintQuit
// C: void Nf_ManPrintQuit( Nf_Man_t * p )
func (p *NfMan) NfManPrintQuit() {
	panic("TODO, stubbed")
}

// Nf_ManCutMatchPrint
// C: void Nf_ManCutMatchPrint( Nf_Man_t * p, int iObj, char * pStr, Nf_Mat_t * pM )
func (p *NfMan) NfManCutMatchPrint(iObj int, pStr string, pM *NfMat) {
	panic("TODO, stubbed")
}

// Nf_ManCutMatchOne
// C: void Nf_ManCutMatchOne( Nf_Man_t * p, int iObj, int * pCut, int * pCutSet )
func (p *NfMan) NfManCutMatchOne(iObj int, pCut []int, pCutSet []int) {
	panic("TODO, stubbed")
}

// Nf_ObjPrepareCi (static inline)
// C: void Nf_ObjPrepareCi( Nf_Man_t * p, int iObj, int Time )
func (p *NfMan) nfObjPrepareCi(iObj int, Time int) {
	panic("TODO, stubbed")
}

// Nf_ObjPrepareBuf (static inline)
// C: void Nf_ObjPrepareBuf( Nf_Man_t * p, Gia_Obj_t * pObj )
func (p *NfMan) nfObjPrepareBuf(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Nf_CutRequired (static inline)
// C: int Nf_CutRequired( Nf_Man_t * p, Nf_Mat_t * pM, int * pCutSet )
func (p *NfMan) nfCutRequired(pM *NfMat, pCutSet []int) int {
	panic("TODO, stubbed")
}

// Nf_ObjComputeRequired (static inline)
// C: void Nf_ObjComputeRequired( Nf_Man_t * p, int iObj )
func (p *NfMan) nfObjComputeRequired(iObj int) {
	panic("TODO, stubbed")
}

// Nf_ManCutMatch
// C: void Nf_ManCutMatch( Nf_Man_t * p, int iObj )
func (p *NfMan) NfManCutMatch(iObj int) {
	panic("TODO, stubbed")
}

// Nf_ManComputeMapping
// C: void Nf_ManComputeMapping( Nf_Man_t * p )
func (p *NfMan) NfManComputeMapping() {
	panic("TODO, stubbed")
}

// Nf_ObjMatchBest (static inline)
// C: Nf_Mat_t * Nf_ObjMatchBest( Nf_Man_t * p, int i, int c )
func (p *NfMan) nfObjMatchBest(i int, c int) *NfMat {
	panic("TODO, stubbed")
}

// Nf_ManSetOutputRequireds
// C: void Nf_ManSetOutputRequireds( Nf_Man_t * p, int fPropCompl )
func (p *NfMan) NfManSetOutputRequireds(fPropCompl int) {
	panic("TODO, stubbed")
}

// Nf_ManSetMapRefsGate
// C: void Nf_ManSetMapRefsGate( Nf_Man_t * p, int iObj, int Required, Nf_Mat_t * pM )
func (p *NfMan) NfManSetMapRefsGate(iObj int, Required int, pM *NfMat) {
	panic("TODO, stubbed")
}

// Nf_ManPrintMatches
// C: void Nf_ManPrintMatches( Nf_Man_t * p )
func (p *NfMan) NfManPrintMatches() {
	panic("TODO, stubbed")
}

// Nf_ManSetMapRefs
// C: int Nf_ManSetMapRefs( Nf_Man_t * p )
func (p *NfMan) NfManSetMapRefs() int {
	panic("TODO, stubbed")
}

// Nf_MatchDeref_rec
// C: word Nf_MatchDeref_rec( Nf_Man_t * p, int i, int c, Nf_Mat_t * pM )
func (p *NfMan) NfMatchDeref_rec(i int, c int, pM *NfMat) uint64 {
	panic("TODO, stubbed")
}

// Nf_MatchRef_rec
// C: word Nf_MatchRef_rec( Nf_Man_t * p, int i, int c, Nf_Mat_t * pM, int Required, Vec_Int_t * vBackup )
func (p *NfMan) NfMatchRef_rec(i int, c int, pM *NfMat, Required int, vBackup *IntVector) uint64 {
	panic("TODO, stubbed")
}

// Nf_MatchRefArea
// C: word Nf_MatchRefArea( Nf_Man_t * p, int i, int c, Nf_Mat_t * pM, int Required )
func (p *NfMan) NfMatchRefArea(i int, c int, pM *NfMat, Required int) uint64 {
	panic("TODO, stubbed")
}

// Nf_ManElaBestMatchOne
// C: void Nf_ManElaBestMatchOne( Nf_Man_t * p, int iObj, int c, int * pCut, int * pCutSet, Nf_Mat_t * pRes, int Required )
func (p *NfMan) NfManElaBestMatchOne(iObj int, c int, pCut []int, pCutSet []int, pRes *NfMat, Required int) {
	panic("TODO, stubbed")
}

// Nf_ManElaBestMatch
// C: void Nf_ManElaBestMatch( Nf_Man_t * p, int iObj, int c, Nf_Mat_t * pRes, int Required )
func (p *NfMan) NfManElaBestMatch(iObj int, c int, pRes *NfMat, Required int) {
	panic("TODO, stubbed")
}

// Nf_ManComputeArrival
// C: int Nf_ManComputeArrival( Nf_Man_t * p, Nf_Mat_t * pM, int * pCutSet )
func (p *NfMan) NfManComputeArrival(pM *NfMat, pCutSet []int) int {
	panic("TODO, stubbed")
}

// Nf_ManResetMatches
// C: void Nf_ManResetMatches( Nf_Man_t * p, int Round )
func (p *NfMan) NfManResetMatches(Round int) {
	panic("TODO, stubbed")
}

// Nf_ManComputeMappingEla
// C: void Nf_ManComputeMappingEla( Nf_Man_t * p )
func (p *NfMan) NfManComputeMappingEla() {
	panic("TODO, stubbed")
}

// Nf_ManFixPoDrivers
// C: void Nf_ManFixPoDrivers( Nf_Man_t * p )
func (p *NfMan) NfManFixPoDrivers() {
	panic("TODO, stubbed")
}

// Nf_ManDeriveMapping
// C: Gia_Man_t * Nf_ManDeriveMapping( Nf_Man_t * p )
func (p *NfMan) NfManDeriveMapping() *GiaManager {
	panic("TODO, stubbed")
}

// Nf_ManUpdateStats
// C: void Nf_ManUpdateStats( Nf_Man_t * p )
func (p *NfMan) NfManUpdateStats() {
	panic("TODO, stubbed")
}

// Nf_ManExtractWindow (pMan is a void *)
// C: int Nf_ManExtractWindow( void * pMan, Vec_Int_t * vRoots, Vec_Wec_t * vCuts, Vec_Wec_t * vObjCuts, Vec_Int_t * vSolCuts, Vec_Int_t * vCutGates, Vec_Wrd_t * vCutAreas, word * pInvArea, int StartVar, int nVars )
func NfManExtractWindow(pMan *NfMan, vRoots *IntVector, vCuts *IntLevelVector, vObjCuts *IntLevelVector, vSolCuts *IntVector, vCutGates *IntVector, vCutAreas *WordVector, pInvArea *uint64, StartVar int, nVars int) int {
	panic("TODO, stubbed")
}

// Nf_ManSetDefaultPars
// C: void Nf_ManSetDefaultPars( Jf_Par_t * pPars )
func NfManSetDefaultPars(pPars *JfPar) {
	panic("TODO, stubbed")
}

// Nf_ManPerformMapping
// C: Gia_Man_t * Nf_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) NfManPerformMapping(pPars *JfPar) *GiaManager {
	panic("TODO, stubbed")
}
