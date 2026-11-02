package abc

// Hand-checked skeleton of src/base/abci/abcDetect.c.

// Abc_FinType_t: fault-injection types.
const (
	ABC_FIN_NONE      = -100 // 0:  unknown
	ABC_FIN_SA0       = -99  // 1:
	ABC_FIN_SA1       = -98  // 2:
	ABC_FIN_NEG       = -97  // 3:
	ABC_FIN_RDOB_AND  = -96  // 4:
	ABC_FIN_RDOB_NAND = -95  // 5:
	ABC_FIN_RDOB_OR   = -94  // 6:
	ABC_FIN_RDOB_NOR  = -93  // 7:
	ABC_FIN_RDOB_XOR  = -92  // 8:
	ABC_FIN_RDOB_NXOR = -91  // 9:
	ABC_FIN_RDOB_NOT  = -90  // 10:
	ABC_FIN_RDOB_BUFF = -89  // 11:
	ABC_FIN_RDOB_LAST = -88  // 12:
)

// Abc_NtkGenFaultList
// C: void Abc_NtkGenFaultList( Abc_Ntk_t * pNtk, char * pFileName, int fStuckAt )
func (pNtk *ABCNetwork) AbcNtkGenFaultList(pFileName string, fStuckAt int) {
	panic("TODO, stubbed")
}

// Io_ReadFinTypeMapped
// C: int Io_ReadFinTypeMapped( Mio_Library_t * pLib, char * pThis )
func IoReadFinTypeMapped(pLib *MioLibrary, pThis string) int {
	panic("TODO, stubbed")
}

// Io_ReadFinType
// C: int Io_ReadFinType( char * pThis )
func IoReadFinType(pThis string) int {
	panic("TODO, stubbed")
}

// Io_WriteFinType
// C: char * Io_WriteFinType( int Type )
func IoWriteFinType(Type int) string {
	panic("TODO, stubbed")
}

// Io_ReadFins
// C: Vec_Int_t * Io_ReadFins( Abc_Ntk_t * pNtk, char * pFileName, int fVerbose )
func IoReadFins(pNtk *ABCNetwork, pFileName string, fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkFrameExtend
// C: void Abc_NtkFrameExtend( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkFrameExtend() {
	panic("TODO, stubbed")
}

// Abc_NtkDetectObjClasses_rec
// C: int Abc_NtkDetectObjClasses_rec( Abc_Obj_t * pObj, Vec_Int_t * vMap, Hsh_VecMan_t * pHash, Vec_Int_t * vTemp )
func AbcNtkDetectObjClasses_rec(pObj *ABCNetworkObject, vMap *IntVector, pHash *IntVectorInterner, vTemp *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkDetectObjClasses
// C: Vec_Wec_t * Abc_NtkDetectObjClasses( Abc_Ntk_t * pNtk, Vec_Int_t * vObjs, Vec_Wec_t ** pvCos )
func (pNtk *ABCNetwork) AbcNtkDetectObjClasses(vObjs *IntVector, pvCos **IntLevelVector) *IntLevelVector {
	panic("TODO, stubbed")
}

// Abc_NtkDetectClassesTest2
// C: void Abc_NtkDetectClassesTest2( Abc_Ntk_t * pNtk, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) AbcNtkDetectClassesTest2(fVerbose int, fVeryVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkFinMiterCollect_rec
// C: void Abc_NtkFinMiterCollect_rec( Abc_Obj_t * pObj, Vec_Int_t * vCis, Vec_Int_t * vNodes )
func AbcNtkFinMiterCollect_rec(pObj *ABCNetworkObject, vCis *IntVector, vNodes *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkFinMiterCollect
// C: void Abc_NtkFinMiterCollect( Abc_Ntk_t * pNtk, Vec_Int_t * vCos, Vec_Int_t * vCis, Vec_Int_t * vNodes )
func (pNtk *ABCNetwork) AbcNtkFinMiterCollect(vCos *IntVector, vCis *IntVector, vNodes *IntVector) {
	panic("TODO, stubbed")
}

// Mio_LibGateSimulate
// C: void Mio_LibGateSimulate( Mio_Gate_t * pGate, word * ppFaninSims[6], int nWords, word * pObjSim )
func MioLibGateSimulate(pGate *MioGate, ppFaninSims [][]uint64, nWords int, pObjSim []uint64) {
	panic("TODO, stubbed")
}

// Mio_LibGateSimulateOne
// C: int Mio_LibGateSimulateOne( Mio_Gate_t * pGate, int iBits[6] )
func MioLibGateSimulateOne(pGate *MioGate, iBits []int) int {
	panic("TODO, stubbed")
}

// Mio_LibGateSimulateGia
// C: int Mio_LibGateSimulateGia( Gia_Man_t * pGia, Mio_Gate_t * pGate, int iLits[6], Vec_Int_t * vLits )
func MioLibGateSimulateGia(pGia *GiaManager, pGate *MioGate, iLits []int, vLits *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkFinSimOneLit (static)
// C: int Abc_NtkFinSimOneLit( Gia_Man_t * pNew, Abc_Obj_t * pObj, int Type, Vec_Int_t * vMap, int n, Vec_Int_t * vTemp )
func abcNtkFinSimOneLit(pNew *GiaManager, pObj *ABCNetworkObject, Type int, vMap *IntVector, n int, vTemp *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkFinSimOneBit (static)
// C: int Abc_NtkFinSimOneBit( Abc_Obj_t * pObj, int Type, Vec_Wrd_t * vSims, int nWords, int iBit )
func abcNtkFinSimOneBit(pObj *ABCNetworkObject, Type int, vSims *WordVector, nWords int, iBit int) int {
	panic("TODO, stubbed")
}

// Abc_NtkFinSimOneWord (static)
// C: void Abc_NtkFinSimOneWord( Abc_Obj_t * pObj, int Type, Vec_Wrd_t * vSims, int nWords )
func abcNtkFinSimOneWord(pObj *ABCNetworkObject, Type int, vSims *WordVector, nWords int) {
	panic("TODO, stubbed")
}

// Abc_NtkFinCompareSimTwo (static)
// C: int Abc_NtkFinCompareSimTwo( Abc_Ntk_t * pNtk, Vec_Int_t * vCos, Vec_Wrd_t * vSims, int nWords, int i1, int i2 )
func (pNtk *ABCNetwork) abcNtkFinCompareSimTwo(vCos *IntVector, vSims *WordVector, nWords int, i1 int, i2 int) int {
	panic("TODO, stubbed")
}

// Abc_NtkFinMiterToGia
// C: Gia_Man_t * Abc_NtkFinMiterToGia( Abc_Ntk_t * pNtk, Vec_Int_t * vTypes, Vec_Int_t * vCos, Vec_Int_t * vCis, Vec_Int_t * vNodes, int iObjs[2], int Types[2], Vec_Int_t * vLits )
func (pNtk *ABCNetwork) AbcNtkFinMiterToGia(vTypes *IntVector, vCos *IntVector, vCis *IntVector, vNodes *IntVector, iObjs []int, Types []int, vLits *IntVector) *GiaManager {
	panic("TODO, stubbed")
}

// Abc_NtkFinSimulateOne
// C: void Abc_NtkFinSimulateOne( Abc_Ntk_t * pNtk, Vec_Int_t * vTypes, Vec_Int_t * vCos, Vec_Int_t * vCis, Vec_Int_t * vNodes, Vec_Wec_t * vMap2, Vec_Int_t * vPat, Vec_Wrd_t * vSims, int nWords, Vec_Int_t * vPairs, Vec_Wec_t * vRes, int iLevel, int iItem )
func (pNtk *ABCNetwork) AbcNtkFinSimulateOne(vTypes *IntVector, vCos *IntVector, vCis *IntVector, vNodes *IntVector, vMap2 *IntLevelVector, vPat *IntVector, vSims *WordVector, nWords int, vPairs *IntVector, vRes *IntLevelVector, iLevel int, iItem int) {
	panic("TODO, stubbed")
}

// Abc_NtkFinCheckPair
// C: Vec_Int_t * Abc_NtkFinCheckPair( Abc_Ntk_t * pNtk, Vec_Int_t * vTypes, Vec_Int_t * vCos, Vec_Int_t * vCis, Vec_Int_t * vNodes, int iObjs[2], int Types[2], Vec_Int_t * vLits )
func (pNtk *ABCNetwork) AbcNtkFinCheckPair(vTypes *IntVector, vCos *IntVector, vCis *IntVector, vNodes *IntVector, iObjs []int, Types []int, vLits *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkFinLocalSetup
// C: void Abc_NtkFinLocalSetup( Vec_Int_t * vPairs, Vec_Int_t * vList, Vec_Wec_t * vMap2, Vec_Int_t * vResArray )
func AbcNtkFinLocalSetup(vPairs *IntVector, vList *IntVector, vMap2 *IntLevelVector, vResArray *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkFinLocalSetdown
// C: void Abc_NtkFinLocalSetdown( Vec_Int_t * vPairs, Vec_Int_t * vList, Vec_Wec_t * vMap2 )
func AbcNtkFinLocalSetdown(vPairs *IntVector, vList *IntVector, vMap2 *IntLevelVector) {
	panic("TODO, stubbed")
}

// Abc_NtkFinRefinement
// C: int Abc_NtkFinRefinement( Abc_Ntk_t * pNtk, Vec_Int_t * vTypes, Vec_Int_t * vCos, Vec_Int_t * vCis, Vec_Int_t * vNodes, Vec_Int_t * vPairs, Vec_Int_t * vList, Vec_Wec_t * vMap2, Vec_Wec_t * vResult )
func (pNtk *ABCNetwork) AbcNtkFinRefinement(vTypes *IntVector, vCos *IntVector, vCis *IntVector, vNodes *IntVector, vPairs *IntVector, vList *IntVector, vMap2 *IntLevelVector, vResult *IntLevelVector) int {
	panic("TODO, stubbed")
}

// Abc_ObjFinGateType (static)
// C: int Abc_ObjFinGateType( Abc_Obj_t * pNode )
func abcObjFinGateType(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkFinCheckTypesOk
// C: int Abc_NtkFinCheckTypesOk( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkFinCheckTypesOk() int {
	panic("TODO, stubbed")
}

// Abc_NtkFinCheckTypesOk2
// C: int Abc_NtkFinCheckTypesOk2( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkFinCheckTypesOk2() int {
	panic("TODO, stubbed")
}

// Abc_NtkFinComputeTypes
// C: Vec_Int_t * Abc_NtkFinComputeTypes( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkFinComputeTypes() *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkFinComputeObjects
// C: Vec_Int_t * Abc_NtkFinComputeObjects( Vec_Int_t * vPairs, Vec_Wec_t ** pvMap, int nObjs )
func AbcNtkFinComputeObjects(vPairs *IntVector, pvMap **IntLevelVector, nObjs int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkFinCreateList
// C: Vec_Int_t * Abc_NtkFinCreateList( Vec_Wec_t * vMap, Vec_Int_t * vClass )
func AbcNtkFinCreateList(vMap *IntLevelVector, vClass *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkFinCountPairs
// C: int Abc_NtkFinCountPairs( Vec_Wec_t * vClasses )
func AbcNtkFinCountPairs(vClasses *IntLevelVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkDetectFinClasses
// C: Vec_Wec_t * Abc_NtkDetectFinClasses( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDetectFinClasses(fVerbose int) *IntLevelVector {
	panic("TODO, stubbed")
}

// Abc_NtkPrintFinResults
// C: void Abc_NtkPrintFinResults( Vec_Wec_t * vClasses )
func AbcNtkPrintFinResults(vClasses *IntLevelVector) {
	panic("TODO, stubbed")
}

// Abc_NtkDetectClassesTest
// C: void Abc_NtkDetectClassesTest( Abc_Ntk_t * pNtk, int fSeq, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) AbcNtkDetectClassesTest(fSeq int, fVerbose int, fVeryVerbose int) {
	panic("TODO, stubbed")
}
