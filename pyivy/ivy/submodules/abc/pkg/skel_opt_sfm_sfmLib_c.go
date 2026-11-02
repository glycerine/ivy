package abc

// Hand-checked skeleton of src/opt/sfm/sfmLib.c.

// SfmFun is Sfm_Fun_t. SFM_SUPP_MAX is 8.
type SfmFun struct {
	Next   int         // next function in the list
	Area   int         // area of this function
	pFansT [8 + 1]int8 // char pFansT[SFM_SUPP_MAX+1] : top gate ID, followed by fanin perm
	pFansB [8 + 1]int8 // char pFansB[SFM_SUPP_MAX+1] : bottom gate ID, followed by fanin perm
}

// SfmLib is Sfm_Lib_t.
type SfmLib struct {
	nVars       int         // variable count
	nWords      int         // truth table words
	fVerbose    int         // verbose statistics
	pCells      []MioCell2  // Mio_Cell2_t * : library gates (an array of nCells)
	nCells      int         // library gate count
	fDelay      int         // uses delay profile
	nObjs       int         // object count
	nObjsAlloc  int         // object count
	pObjs       []SfmFun    // Sfm_Fun_t * : objects
	vTtMem      *WordMemory // truth tables
	vLists      IntVector   // Vec_Int_t (by value) : lists of funcs for each truth table
	vCounts     IntVector   // Vec_Int_t (by value) : counters of functions for each truth table
	vHits       IntVector   // Vec_Int_t (by value) : the number of times this function was used
	vProfs      IntVector   // Vec_Int_t (by value) : area/delay profiles
	vStore      IntVector   // Vec_Int_t (by value) : storage for area/delay profiles
	vTemp       IntVector   // Vec_Int_t (by value) : temporary storage for candidates
	nObjSkipped int
	nObjRemoved int
}

// Sfm_LibFun (static inline)
// C: Sfm_Fun_t * Sfm_LibFun( Sfm_Lib_t * p, int i )
func (p *SfmLib) sfmLibFun(i int) *SfmFun {
	panic("TODO, stubbed")
}

// Sfm_LibFunId (static inline)
// C: int Sfm_LibFunId( Sfm_Lib_t * p, Sfm_Fun_t * pFun )
func (p *SfmLib) sfmLibFunId(pFun *SfmFun) int {
	panic("TODO, stubbed")
}

// Sfm_DecCreateCnf
// C: void Sfm_DecCreateCnf( Vec_Int_t * vGateSizes, Vec_Wrd_t * vGateFuncs, Vec_Wec_t * vGateCnfs )
func SfmDecCreateCnf(vGateSizes *IntVector, vGateFuncs *WordVector, vGateCnfs *IntLevelVector) {
	panic("TODO, stubbed")
}

// Sfm_LibPreprocess
// C: void Sfm_LibPreprocess( Mio_Library_t * pLib, Vec_Int_t * vGateSizes, Vec_Wrd_t * vGateFuncs, Vec_Wec_t * vGateCnfs, Vec_Ptr_t * vGateHands )
func (pLib *MioLibrary) SfmLibPreprocess(vGateSizes *IntVector, vGateFuncs *WordVector, vGateCnfs *IntLevelVector, vGateHands *PointerVector[MioGate]) {
	panic("TODO, stubbed")
}

// Sfm_LibFindComplInputGate (piFaninNew is an out parameter)
// C: int Sfm_LibFindComplInputGate( Vec_Wrd_t * vFuncs, int iGate, int nFanins, int iFanin, int * piFaninNew )
func SfmLibFindComplInputGate(vFuncs *WordVector, iGate int, nFanins int, iFanin int, piFaninNew *int) int {
	panic("TODO, stubbed")
}

// Sfm_LibStart
// C: Sfm_Lib_t * Sfm_LibStart( int nVars, int fDelay, int fVerbose )
func SfmLibStart(nVars int, fDelay int, fVerbose int) *SfmLib {
	panic("TODO, stubbed")
}

// Sfm_LibStop
// C: void Sfm_LibStop( Sfm_Lib_t * p )
func (p *SfmLib) SfmLibStop() {
	panic("TODO, stubbed")
}

// Sfm_LibTruth6Two
// C: word Sfm_LibTruth6Two( Mio_Cell2_t * pCellBot, Mio_Cell2_t * pCellTop, int InTop )
func SfmLibTruth6Two(pCellBot *MioCell2, pCellTop *MioCell2, InTop int) uint64 {
	panic("TODO, stubbed")
}

// Sfm_LibTruth8Two
// C: void Sfm_LibTruth8Two( Mio_Cell2_t * pCellBot, Mio_Cell2_t * pCellTop, int InTop, word * pRes )
func SfmLibTruth8Two(pCellBot *MioCell2, pCellTop *MioCell2, InTop int, pRes []uint64) {
	panic("TODO, stubbed")
}

// Sfm_LibCellProfile
// C: void Sfm_LibCellProfile( Mio_Cell2_t * pCellBot, Mio_Cell2_t * pCellTop, int InTop, int nFanins, int * Perm, int * pProf )
func SfmLibCellProfile(pCellBot *MioCell2, pCellTop *MioCell2, InTop int, nFanins int, Perm []int, pProf []int) {
	panic("TODO, stubbed")
}

// Sfm_LibNewIsContained (static)
// C: int Sfm_LibNewIsContained( Sfm_Fun_t * pObj, int * pProf, int Area, int * pProfNew, int nFanins )
func sfmLibNewIsContained(pObj *SfmFun, pProf []int, Area int, pProfNew []int, nFanins int) int {
	panic("TODO, stubbed")
}

// Sfm_LibNewContains (static)
// C: int Sfm_LibNewContains( Sfm_Fun_t * pObj, int * pProf, int Area, int * pProfNew, int nFanins )
func sfmLibNewContains(pObj *SfmFun, pProf []int, Area int, pProfNew []int, nFanins int) int {
	panic("TODO, stubbed")
}

// Sfm_LibPrepareAdd
// C: void Sfm_LibPrepareAdd( Sfm_Lib_t * p, word * pTruth, int * Perm, int nFanins, Mio_Cell2_t * pCellBot, Mio_Cell2_t * pCellTop, int InTop )
func (p *SfmLib) SfmLibPrepareAdd(pTruth []uint64, Perm []int, nFanins int, pCellBot *MioCell2, pCellTop *MioCell2, InTop int) {
	panic("TODO, stubbed")
}

// Sfm_LibPrepare
// C: Sfm_Lib_t * Sfm_LibPrepare( int nVars, int fTwo, int fDelay, int fVerbose, int fLibVerbose )
func SfmLibPrepare(nVars int, fTwo int, fDelay int, fVerbose int, fLibVerbose int) *SfmLib {
	panic("TODO, stubbed")
}

// Sfm_LibPrintGate (pFanins and pFanins2 are fanin-permutation byte arrays, not text)
// C: void Sfm_LibPrintGate( Mio_Cell2_t * pCell, char * pFanins, Mio_Cell2_t * pCell2, char * pFanins2 )
func SfmLibPrintGate(pCell *MioCell2, pFanins []int8, pCell2 *MioCell2, pFanins2 []int8) {
	panic("TODO, stubbed")
}

// Sfm_LibPrintObj
// C: void Sfm_LibPrintObj( Sfm_Lib_t * p, Sfm_Fun_t * pObj )
func (p *SfmLib) SfmLibPrintObj(pObj *SfmFun) {
	panic("TODO, stubbed")
}

// Sfm_LibPrint
// C: void Sfm_LibPrint( Sfm_Lib_t * p )
func (p *SfmLib) SfmLibPrint() {
	panic("TODO, stubbed")
}

// Sfm_LibTest
// C: void Sfm_LibTest()
func SfmLibTest() {
	panic("TODO, stubbed")
}

// Sfm_LibFindAreaMatch (piObj is an out parameter)
// C: int Sfm_LibFindAreaMatch( Sfm_Lib_t * p, word * pTruth, int nFanins, int * piObj )
func (p *SfmLib) SfmLibFindAreaMatch(pTruth []uint64, nFanins int, piObj *int) int {
	panic("TODO, stubbed")
}

// Sfm_LibFindDelayMatches
// C: int Sfm_LibFindDelayMatches( Sfm_Lib_t * p, word * pTruth, int * pFanins, int nFanins, Vec_Ptr_t * vGates, Vec_Ptr_t * vFans )
func (p *SfmLib) SfmLibFindDelayMatches(pTruth []uint64, pFanins []int, nFanins int, vGates *PointerVector[MioGate], vFans *PointerVector[int8]) int {
	panic("TODO, stubbed")
}

// Sfm_LibImplementSimple
// C: int Sfm_LibImplementSimple( Sfm_Lib_t * p, word * pTruth, int * pFanins, int nFanins, Vec_Int_t * vGates, Vec_Wec_t * vFanins )
func (p *SfmLib) SfmLibImplementSimple(pTruth []uint64, pFanins []int, nFanins int, vGates *IntVector, vFanins *IntLevelVector) int {
	panic("TODO, stubbed")
}

// Sfm_LibImplementGatesArea
// C: int Sfm_LibImplementGatesArea( Sfm_Lib_t * p, int * pFanins, int nFanins, int iObj, Vec_Int_t * vGates, Vec_Wec_t * vFanins )
func (p *SfmLib) SfmLibImplementGatesArea(pFanins []int, nFanins int, iObj int, vGates *IntVector, vFanins *IntLevelVector) int {
	panic("TODO, stubbed")
}

// Sfm_LibImplementGatesDelay
// C: int Sfm_LibImplementGatesDelay( Sfm_Lib_t * p, int * pFanins, Mio_Gate_t * pGateB, Mio_Gate_t * pGateT, char * pFansB, char * pFansT, Vec_Int_t * vGates, Vec_Wec_t * vFanins )
func (p *SfmLib) SfmLibImplementGatesDelay(pFanins []int, pGateB *MioGate, pGateT *MioGate, pFansB []int8, pFansT []int8, vGates *IntVector, vFanins *IntLevelVector) int {
	panic("TODO, stubbed")
}
