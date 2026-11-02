package abc

// Hand-checked skeleton of src/aig/gia/giaSimBase.c.

// GiaSimRsbMan is Gia_SimRsbMan_t.
type GiaSimRsbMan struct {
	pGia      *GiaManager
	vTfo      *IntVector
	vCands    *IntVector
	vFanins   *IntVector
	vFanins2  *IntVector
	vSimsObj  *WordVector
	vSimsObj2 *WordVector
	nWords    int
	pFunc     [3][]uint64 // word * pFunc[3]
}

// GiaSimAbsMan is Gia_SimAbsMan_t.
type GiaSimAbsMan struct {
	pGia        *GiaManager // AIG manager
	pSet        [2][]uint64 // word * pSet[2] : offset/onset truth tables
	nCands      int         // candidate count
	nWords      int         // word count
	vSims       *WordVector // candidate simulation info
	vResub      *IntVector  // the result
	fVerbose    int         // verbose
	vValues     *IntVector  // function values in each pattern
	vPatPairs   *IntVector  // used minterms
	nWordsTable int         // words of table data
	pTableTemp  []uint64    // word * : temporary table pattern
	vCoverTable *WordVector // columns = minterms; rows = classes
	vTtMints    *IntVector  // truth table minterms
}

// GiaRsbMan is Gia_RsbMan_t.
type GiaRsbMan struct {
	pGia    *GiaManager
	pOffSet []uint64 // word *
	pOnSet  []uint64 // word *
	nWords  int
	nWordsT int
	vSims   *WordVector
	vSimsT  *WordVector
	vCands  *IntVector
	vObjs   *IntVector
	vObjs2  *IntVector
	vSets   [2]*IntLevelVector
	pSet    [3][]uint64 // word * pSet[3]
	vActive *IntVector
}

// Gia_ManSimPatAssignInputs
// C: void Gia_ManSimPatAssignInputs( Gia_Man_t * p, int nWords, Vec_Wrd_t * vSims, Vec_Wrd_t * vSimsIn )
func (p *GiaManager) GiaManSimPatAssignInputs(nWords int, vSims *WordVector, vSimsIn *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimPatSimAnd (static inline)
// C: void Gia_ManSimPatSimAnd( Gia_Man_t * p, int i, Gia_Obj_t * pObj, int nWords, Vec_Wrd_t * vSims )
func (p *GiaManager) giaManSimPatSimAnd(i int, pObj *GiaObject, nWords int, vSims *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimPatSimPo (static inline)
// C: void Gia_ManSimPatSimPo( Gia_Man_t * p, int i, Gia_Obj_t * pObj, int nWords, Vec_Wrd_t * vSims )
func (p *GiaManager) giaManSimPatSimPo(i int, pObj *GiaObject, nWords int, vSims *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimPatSimNot (static inline)
// C: void Gia_ManSimPatSimNot( Gia_Man_t * p, int i, Gia_Obj_t * pObj, int nWords, Vec_Wrd_t * vSims )
func (p *GiaManager) giaManSimPatSimNot(i int, pObj *GiaObject, nWords int, vSims *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimPatSim
// C: Vec_Wrd_t * Gia_ManSimPatSim( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaManSimPatSim() *WordVector {
	panic("TODO, stubbed")
}

// Gia_ManSimPatResim
// C: void Gia_ManSimPatResim( Gia_Man_t * pGia, Vec_Int_t * vObjs, int nWords, Vec_Wrd_t * vSims )
func (pGia *GiaManager) GiaManSimPatResim(vObjs *IntVector, nWords int, vSims *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimPatWrite
// C: void Gia_ManSimPatWrite( char * pFileName, Vec_Wrd_t * vSimsIn, int nWords )
func GiaManSimPatWrite(pFileName string, vSimsIn *WordVector, nWords int) {
	panic("TODO, stubbed")
}

// Gia_ManSimPatAssignInputs2
// C: void Gia_ManSimPatAssignInputs2( Gia_Man_t * p, int nWords, Vec_Wrd_t * vSims, Vec_Wrd_t * vSimsIn )
func (p *GiaManager) GiaManSimPatAssignInputs2(nWords int, vSims *WordVector, vSimsIn *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimPatSimAnd2 (static inline)
// C: void Gia_ManSimPatSimAnd2( Gia_Man_t * p, int i, Gia_Obj_t * pObj, int nWords, Vec_Wrd_t * vSims )
func (p *GiaManager) giaManSimPatSimAnd2(i int, pObj *GiaObject, nWords int, vSims *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimPatSimPo2 (static inline)
// C: void Gia_ManSimPatSimPo2( Gia_Man_t * p, int i, Gia_Obj_t * pObj, int nWords, Vec_Wrd_t * vSims )
func (p *GiaManager) giaManSimPatSimPo2(i int, pObj *GiaObject, nWords int, vSims *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimPatSim2
// C: Vec_Wrd_t * Gia_ManSimPatSim2( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaManSimPatSim2() *WordVector {
	panic("TODO, stubbed")
}

// Gia_ManSimPatValuesDerive
// C: void Gia_ManSimPatValuesDerive( Gia_Man_t * p, int nWords, Vec_Wrd_t * vSims, Vec_Wrd_t * vValues )
func (p *GiaManager) GiaManSimPatValuesDerive(nWords int, vSims *WordVector, vValues *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimPatValues
// C: Vec_Wrd_t * Gia_ManSimPatValues( Gia_Man_t * p )
func (p *GiaManager) GiaManSimPatValues() *WordVector {
	panic("TODO, stubbed")
}

// Gia_ManSimCombine
// C: Vec_Wrd_t * Gia_ManSimCombine( int nInputs, Vec_Wrd_t * vBase, Vec_Wrd_t * vAddOn, int nWordsUse )
func GiaManSimCombine(nInputs int, vBase *WordVector, vAddOn *WordVector, nWordsUse int) *WordVector {
	panic("TODO, stubbed")
}

// Gia_ManSimBitPackOne
// C: int Gia_ManSimBitPackOne( int nWords, Vec_Wrd_t * vSimsIn, Vec_Wrd_t * vSimsCare, int iPat, int * pLits, int nLits )
func GiaManSimBitPackOne(nWords int, vSimsIn *WordVector, vSimsCare *WordVector, iPat int, pLits []int, nLits int) int {
	panic("TODO, stubbed")
}

// Gia_ManSimBitPacking
// C: Vec_Wrd_t * Gia_ManSimBitPacking( Gia_Man_t * p, Vec_Int_t * vCexStore, int nCexes, int nUnDecs )
func (p *GiaManager) GiaManSimBitPacking(vCexStore *IntVector, nCexes int, nUnDecs int) *WordVector {
	panic("TODO, stubbed")
}

// Gia_ManSimPatHashPatterns
// C: int Gia_ManSimPatHashPatterns( Gia_Man_t * p, int nWords, Vec_Wrd_t * vSims, int * pnC0, int * pnC1 )
func (p *GiaManager) GiaManSimPatHashPatterns(nWords int, vSims *WordVector, pnC0 *int, pnC1 *int) int {
	panic("TODO, stubbed")
}

// Gia_ManSimPatGenMiter
// C: Gia_Man_t * Gia_ManSimPatGenMiter( Gia_Man_t * p, Vec_Wrd_t * vSims )
func (p *GiaManager) GiaManSimPatGenMiter(vSims *WordVector) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManSimProfile
// C: void Gia_ManSimProfile( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaManSimProfile() {
	panic("TODO, stubbed")
}

// Gia_ManPatSatImprove
// C: void Gia_ManPatSatImprove( Gia_Man_t * p, int nWords0, int fVerbose )
func (p *GiaManager) GiaManPatSatImprove(nWords0 int, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_SimRsbAlloc
// C: Gia_SimRsbMan_t * Gia_SimRsbAlloc( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaSimRsbAlloc() *GiaSimRsbMan {
	panic("TODO, stubbed")
}

// Gia_SimRsbFree
// C: void Gia_SimRsbFree( Gia_SimRsbMan_t * p )
func (p *GiaSimRsbMan) GiaSimRsbFree() {
	panic("TODO, stubbed")
}

// Gia_SimRsbTfo_rec
// C: void Gia_SimRsbTfo_rec( Gia_Man_t * p, int iObj, int iFanout, Vec_Int_t * vTfo )
func (p *GiaManager) GiaSimRsbTfo_rec(iObj int, iFanout int, vTfo *IntVector) {
	panic("TODO, stubbed")
}

// Gia_SimRsbTfo
// C: Vec_Int_t * Gia_SimRsbTfo( Gia_SimRsbMan_t * p, int iObj, int iFanout )
func (p *GiaSimRsbMan) GiaSimRsbTfo(iObj int, iFanout int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_SimRsbFunc
// C: word * Gia_SimRsbFunc( Gia_SimRsbMan_t * p, int iObj, Vec_Int_t * vFanins, int fOnSet )
func (p *GiaSimRsbMan) GiaSimRsbFunc(iObj int, vFanins *IntVector, fOnSet int) []uint64 {
	panic("TODO, stubbed")
}

// Gia_SimRsbResubVerify
// C: int Gia_SimRsbResubVerify( Gia_SimRsbMan_t * p, int iObj, Vec_Int_t * vFanins )
func (p *GiaSimRsbMan) GiaSimRsbResubVerify(iObj int, vFanins *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_SimRsbSimAndCareSet (static inline)
// C: void Gia_SimRsbSimAndCareSet( Gia_Man_t * p, int i, Gia_Obj_t * pObj, int nWords, Vec_Wrd_t * vSims, Vec_Wrd_t * vSims2 )
func (p *GiaManager) giaSimRsbSimAndCareSet(i int, pObj *GiaObject, nWords int, vSims *WordVector, vSims2 *WordVector) {
	panic("TODO, stubbed")
}

// Gia_SimRsbCareSet
// C: word * Gia_SimRsbCareSet( Gia_SimRsbMan_t * p, int iObj, Vec_Int_t * vTfo )
func (p *GiaSimRsbMan) GiaSimRsbCareSet(iObj int, vTfo *IntVector) []uint64 {
	panic("TODO, stubbed")
}

// Gia_ObjSimCollect
// C: void Gia_ObjSimCollect( Gia_SimRsbMan_t * p )
func (p *GiaSimRsbMan) GiaObjSimCollect() {
	panic("TODO, stubbed")
}

// Gia_ObjSimCands
// C: Vec_Int_t * Gia_ObjSimCands( Gia_SimRsbMan_t * p, int iObj, int nCands )
func (p *GiaSimRsbMan) GiaObjSimCands(iObj int, nCands int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ObjSimRsb
// C: int Gia_ObjSimRsb( Gia_SimRsbMan_t * p, int iObj, int nCands, int fVerbose, int * pnBufs, int * pnInvs )
func (p *GiaSimRsbMan) GiaObjSimRsb(iObj int, nCands int, fVerbose int, pnBufs *int, pnInvs *int) int {
	panic("TODO, stubbed")
}

// Gia_ManSimRsb
// C: int Gia_ManSimRsb( Gia_Man_t * pGia, int nCands, int fVerbose )
func (pGia *GiaManager) GiaManSimRsb(nCands int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_ManSimRelAssignInputs
// C: void Gia_ManSimRelAssignInputs( Gia_Man_t * p, int nWords, Vec_Wrd_t * vSims, int nWordsIn, Vec_Wrd_t * vSimsIn )
func (p *GiaManager) GiaManSimRelAssignInputs(nWords int, vSims *WordVector, nWordsIn int, vSimsIn *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimRelCompare
// C: int Gia_ManSimRelCompare( Gia_Man_t * p, int nWords, Vec_Wrd_t * vSims, int nWordsOut, Vec_Wrd_t * vSimsOut, int iPat, int iMint )
func (p *GiaManager) GiaManSimRelCompare(nWords int, vSims *WordVector, nWordsOut int, vSimsOut *WordVector, iPat int, iMint int) int {
	panic("TODO, stubbed")
}

// Gia_ManSimRelCollectOutputs
// C: int Gia_ManSimRelCollectOutputs( Gia_Man_t * p, int nWords, Vec_Wrd_t * vSims, int nWordsOut, Vec_Wrd_t * vSimsOut, Vec_Wrd_t * vRel )
func (p *GiaManager) GiaManSimRelCollectOutputs(nWords int, vSims *WordVector, nWordsOut int, vSimsOut *WordVector, vRel *WordVector) int {
	panic("TODO, stubbed")
}

// Gia_ManSimRel
// C: Vec_Wrd_t * Gia_ManSimRel( Gia_Man_t * p, Vec_Int_t * vObjs, Vec_Wrd_t * vVals )
func (p *GiaManager) GiaManSimRel(vObjs *IntVector, vVals *WordVector) *WordVector {
	panic("TODO, stubbed")
}

// Gia_ManSimRelCheckFuncs
// C: void Gia_ManSimRelCheckFuncs( Gia_Man_t * p, Vec_Wrd_t * vRel, int nOuts, Vec_Wrd_t * vFuncs )
func (p *GiaManager) GiaManSimRelCheckFuncs(vRel *WordVector, nOuts int, vFuncs *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimRelDeriveFuncs
// C: Vec_Wrd_t * Gia_ManSimRelDeriveFuncs( Gia_Man_t * p, Vec_Wrd_t * vRel, int nOuts )
func (p *GiaManager) GiaManSimRelDeriveFuncs(vRel *WordVector, nOuts int) *WordVector {
	panic("TODO, stubbed")
}

// Gia_ManSimRelDeriveFuncs2
// C: Vec_Wrd_t * Gia_ManSimRelDeriveFuncs2( Gia_Man_t * p, Vec_Wrd_t * vRel, int nOuts )
func (p *GiaManager) GiaManSimRelDeriveFuncs2(vRel *WordVector, nOuts int) *WordVector {
	panic("TODO, stubbed")
}

// Gia_ManSimRelPrint
// C: void Gia_ManSimRelPrint( Gia_Man_t * p, Vec_Wrd_t * vRel, Vec_Int_t * vOutMints )
func (p *GiaManager) GiaManSimRelPrint(vRel *WordVector, vOutMints *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimPatStart
// C: Vec_Int_t * Gia_ManSimPatStart( int nItems )
func GiaManSimPatStart(nItems int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManSimRelTest
// C: void Gia_ManSimRelTest( Gia_Man_t * p )
func (p *GiaManager) GiaManSimRelTest() {
	panic("TODO, stubbed")
}

// Gia_Sim5CollectValues
// C: Vec_Int_t * Gia_Sim5CollectValues( word * pOffSet, word * pOnSet, int nWords )
func GiaSim5CollectValues(pOffSet []uint64, pOnSet []uint64, nWords int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_SimAbsAlloc
// C: Gia_SimAbsMan_t * Gia_SimAbsAlloc( Gia_Man_t * pGia, word * pOffSet, word * pOnSet, Vec_Wrd_t * vSims, int nWords, Vec_Int_t * vResub, int fVerbose )
func (pGia *GiaManager) GiaSimAbsAlloc(pOffSet []uint64, pOnSet []uint64, vSims *WordVector, nWords int, vResub *IntVector, fVerbose int) *GiaSimAbsMan {
	panic("TODO, stubbed")
}

// Gia_SimAbsFree
// C: void Gia_SimAbsFree( Gia_SimAbsMan_t * p )
func (p *GiaSimAbsMan) GiaSimAbsFree() {
	panic("TODO, stubbed")
}

// Gia_SimAbsCheckSolution
// C: void Gia_SimAbsCheckSolution( Gia_SimAbsMan_t * p )
func (p *GiaSimAbsMan) GiaSimAbsCheckSolution() {
	panic("TODO, stubbed")
}

// Gia_SimAbsSolve
// C: void Gia_SimAbsSolve( Gia_SimAbsMan_t * p )
func (p *GiaSimAbsMan) GiaSimAbsSolve() {
	panic("TODO, stubbed")
}

// Gia_SimAbsRefine
// C: int Gia_SimAbsRefine( Gia_SimAbsMan_t * p )
func (p *GiaSimAbsMan) GiaSimAbsRefine() int {
	panic("TODO, stubbed")
}

// Gia_SimAbsFind
// C: Vec_Int_t * Gia_SimAbsFind( Vec_Int_t * vValues, int Value )
func GiaSimAbsFind(vValues *IntVector, Value int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_SimAbsInit
// C: void Gia_SimAbsInit( Gia_SimAbsMan_t * p )
func (p *GiaSimAbsMan) GiaSimAbsInit() {
	panic("TODO, stubbed")
}

// Gia_SimAbsPerformOne
// C: Vec_Int_t * Gia_SimAbsPerformOne( Gia_Man_t * pGia, word * pOffSet, word * pOnSet, Vec_Wrd_t * vSimsCands, int nWords, int fVerbose )
func (pGia *GiaManager) GiaSimAbsPerformOne(pOffSet []uint64, pOnSet []uint64, vSimsCands *WordVector, nWords int, fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_RsbAlloc
// C: Gia_RsbMan_t * Gia_RsbAlloc( Gia_Man_t * pGia, word * pOffSet, word * pOnSet, Vec_Wrd_t * vSims, int nWords, Vec_Wrd_t * vSimsT, int nWordsT, Vec_Int_t * vCands )
func (pGia *GiaManager) GiaRsbAlloc(pOffSet []uint64, pOnSet []uint64, vSims *WordVector, nWords int, vSimsT *WordVector, nWordsT int, vCands *IntVector) *GiaRsbMan {
	panic("TODO, stubbed")
}

// Gia_RsbFree
// C: void Gia_RsbFree( Gia_RsbMan_t * p )
func (p *GiaRsbMan) GiaRsbFree() {
	panic("TODO, stubbed")
}

// Gia_RsbCost
// C: int Gia_RsbCost( Gia_RsbMan_t * p )
func (p *GiaRsbMan) GiaRsbCost() int {
	panic("TODO, stubbed")
}

// Gia_RsbPrint
// C: void Gia_RsbPrint( Gia_RsbMan_t * p )
func (p *GiaRsbMan) GiaRsbPrint() {
	panic("TODO, stubbed")
}

// Gia_RsbUpdateAdd
// C: void Gia_RsbUpdateAdd( Gia_RsbMan_t * p, int iObj )
func (p *GiaRsbMan) GiaRsbUpdateAdd(iObj int) {
	panic("TODO, stubbed")
}

// Gia_RsbUpdateRemove
// C: void Gia_RsbUpdateRemove( Gia_RsbMan_t * p, int Index )
func (p *GiaRsbMan) GiaRsbUpdateRemove(Index int) {
	panic("TODO, stubbed")
}

// Gia_RsbRemovalCost
// C: int Gia_RsbRemovalCost( Gia_RsbMan_t * p, int Index )
func (p *GiaRsbMan) GiaRsbRemovalCost(Index int) int {
	panic("TODO, stubbed")
}

// Gia_RsbFindNodeToRemove
// C: int Gia_RsbFindNodeToRemove( Gia_RsbMan_t * p, int * pMinCost )
func (p *GiaRsbMan) GiaRsbFindNodeToRemove(pMinCost *int) int {
	panic("TODO, stubbed")
}

// Gia_RsbFindMints
// C: void Gia_RsbFindMints( Gia_RsbMan_t * p, int * pMint0, int * pMint1 )
func (p *GiaRsbMan) GiaRsbFindMints(pMint0 *int, pMint1 *int) {
	panic("TODO, stubbed")
}

// Gia_RsbFindNode
// C: int Gia_RsbFindNode( Gia_RsbMan_t * p )
func (p *GiaRsbMan) GiaRsbFindNode() int {
	panic("TODO, stubbed")
}

// Gia_RsbCollectValid
// C: int Gia_RsbCollectValid( Gia_RsbMan_t * p )
func (p *GiaRsbMan) GiaRsbCollectValid() int {
	panic("TODO, stubbed")
}

// Gia_RsbSolve
// C: Vec_Int_t * Gia_RsbSolve( Gia_RsbMan_t * p )
func (p *GiaRsbMan) GiaRsbSolve() *IntVector {
	panic("TODO, stubbed")
}

// Gia_RsbSetFind
// C: Vec_Int_t * Gia_RsbSetFind( word * pOffSet, word * pOnSet, Vec_Wrd_t * vSims, int nWords, Vec_Wrd_t * vSimsT, int nWordsT, Vec_Int_t * vCands )
func GiaRsbSetFind(pOffSet []uint64, pOnSet []uint64, vSims *WordVector, nWords int, vSimsT *WordVector, nWordsT int, vCands *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_SimQualityOne
// C: Vec_Int_t * Gia_SimQualityOne( Gia_Man_t * p, Vec_Int_t * vPat, int fPoOnly )
func (p *GiaManager) GiaSimQualityOne(vPat *IntVector, fPoOnly int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_SimQualityTest
// C: void Gia_SimQualityTest( Gia_Man_t * p )
func (p *GiaManager) GiaSimQualityTest() {
	panic("TODO, stubbed")
}

// Gia_SimGenerateStats
// C: Vec_Int_t * Gia_SimGenerateStats( Gia_Man_t * p )
func (p *GiaManager) GiaSimGenerateStats() *IntVector {
	panic("TODO, stubbed")
}

// Gia_SimComputeScore
// C: double Gia_SimComputeScore( Gia_Man_t * p, Vec_Int_t * vTotal, Vec_Int_t * vThis )
func (p *GiaManager) GiaSimComputeScore(vTotal *IntVector, vThis *IntVector) float64 {
	panic("TODO, stubbed")
}

// Gia_SimQualityPatternsMax
// C: int Gia_SimQualityPatternsMax( Gia_Man_t * p, Vec_Int_t * vPat, int Iter, int fVerbose, Vec_Int_t * vStats )
func (p *GiaManager) GiaSimQualityPatternsMax(vPat *IntVector, Iter int, fVerbose int, vStats *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManPatCollectOne
// C: Vec_Int_t * Gia_ManPatCollectOne( Gia_Man_t * p, Vec_Wrd_t * vPatterns, int n, int nWords )
func (p *GiaManager) GiaManPatCollectOne(vPatterns *WordVector, n int, nWords int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManPatUpdateOne
// C: void Gia_ManPatUpdateOne( Gia_Man_t * p, Vec_Wrd_t * vPatterns, int n, int nWords, Vec_Int_t * vPat )
func (p *GiaManager) GiaManPatUpdateOne(vPatterns *WordVector, n int, nWords int, vPat *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManPatDistImprove
// C: void Gia_ManPatDistImprove( Gia_Man_t * p, int fVerbose )
func (p *GiaManager) GiaManPatDistImprove(fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_SimCollectRare
// C: Vec_Int_t * Gia_SimCollectRare( Gia_Man_t * p, Vec_Wrd_t * vPatterns, int RareLimit )
func (p *GiaManager) GiaSimCollectRare(vPatterns *WordVector, RareLimit int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_SimQualityImpact
// C: Vec_Flt_t * Gia_SimQualityImpact( Gia_Man_t * p, Vec_Int_t * vPat, Vec_Int_t * vRareCounts )
func (p *GiaManager) GiaSimQualityImpact(vPat *IntVector, vRareCounts *IntVector) *FloatVector {
	panic("TODO, stubbed")
}

// Gia_SimCollectBest
// C: Vec_Int_t * Gia_SimCollectBest( Vec_Flt_t * vQuo )
func GiaSimCollectBest(vQuo *FloatVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManPatGetQuo
// C: float Gia_ManPatGetQuo( Gia_Man_t * p, Vec_Int_t * vRareCounts, Vec_Wrd_t * vSims, int n, int nWords )
func (p *GiaManager) GiaManPatGetQuo(vRareCounts *IntVector, vSims *WordVector, n int, nWords int) float32 {
	panic("TODO, stubbed")
}

// Gia_ManPatGetTotalQuo
// C: float Gia_ManPatGetTotalQuo( Gia_Man_t * p, int RareLimit, Vec_Wrd_t * vPatterns, int nWords )
func (p *GiaManager) GiaManPatGetTotalQuo(RareLimit int, vPatterns *WordVector, nWords int) float32 {
	panic("TODO, stubbed")
}

// Gia_ManPatGetOneQuo
// C: float Gia_ManPatGetOneQuo( Gia_Man_t * p, int RareLimit, Vec_Wrd_t * vPatterns, int nWords, int n )
func (p *GiaManager) GiaManPatGetOneQuo(RareLimit int, vPatterns *WordVector, nWords int, n int) float32 {
	panic("TODO, stubbed")
}

// Gia_ManPatRareImprove
// C: void Gia_ManPatRareImprove( Gia_Man_t * p, int RareLimit, int fVerbose )
func (p *GiaManager) GiaManPatRareImprove(RareLimit int, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManSimTest
// C: void Gia_ManSimTest( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaManSimTest() {
	panic("TODO, stubbed")
}

// Gia_ManSimGen
// C: void Gia_ManSimGen( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaManSimGen() {
	panic("TODO, stubbed")
}
