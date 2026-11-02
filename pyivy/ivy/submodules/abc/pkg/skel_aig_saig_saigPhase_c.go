package abc

import "os"

// Hand-checked skeleton of src/aig/saig/saigPhase.c.

const (
	TSIM_MAX_ROUNDS = 10000
	TSIM_ONE_SERIES = 3000
	SAIG_XVSX       = 3
)

// SaigTsim is Saig_Tsim_t: the ternary simulation manager.
type SaigTsim struct {
	pAig   *AigManager // the original AIG manager
	nWords int         // the number of words in the states
	// ternary state representation
	vStates   *PointerVector[uint32] // the collection of ternary states
	pMem      *AigFixedArena         // Aig_MmFixed_t * : memory for ternary states
	nPrefix   int                    // prefix of the ternary state space
	nCycle    int                    // cycle of the ternary state space
	nNonXRegs int                    // the number of candidate registers
	vNonXRegs *IntVector             // the candidate registers
	// hash table for terminary states
	pBins [][]uint32 // unsigned ** : hash table bins
	nBins int
}

// Saig_XsimConvertValue (static inline)
// C: int Saig_XsimConvertValue( int v )
func saigXsimConvertValue(v int) int {
	panic("TODO, stubbed")
}

// Saig_ObjSetXsim (static inline)
// C: void Saig_ObjSetXsim( Aig_Obj_t * pObj, int Value )
func saigObjSetXsim(pObj *AigObject, Value int) {
	panic("TODO, stubbed")
}

// Saig_ObjGetXsim (static inline)
// C: int Saig_ObjGetXsim( Aig_Obj_t * pObj )
func saigObjGetXsim(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Saig_XsimInv (static inline)
// C: int Saig_XsimInv( int Value )
func saigXsimInv(Value int) int {
	panic("TODO, stubbed")
}

// Saig_XsimAnd (static inline)
// C: int Saig_XsimAnd( int Value0, int Value1 )
func saigXsimAnd(Value0 int, Value1 int) int {
	panic("TODO, stubbed")
}

// Saig_XsimRand2 (static inline)
// C: int Saig_XsimRand2()
func saigXsimRand2() int {
	panic("TODO, stubbed")
}

// Saig_XsimRand3 (static inline)
// C: int Saig_XsimRand3()
func saigXsimRand3() int {
	panic("TODO, stubbed")
}

// Saig_ObjGetXsimFanin0 (static inline)
// C: int Saig_ObjGetXsimFanin0( Aig_Obj_t * pObj )
func saigObjGetXsimFanin0(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Saig_ObjGetXsimFanin1 (static inline)
// C: int Saig_ObjGetXsimFanin1( Aig_Obj_t * pObj )
func saigObjGetXsimFanin1(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Saig_XsimPrint (static inline)
// C: void Saig_XsimPrint( FILE * pFile, int Value )
func saigXsimPrint(pFile *os.File, Value int) {
	panic("TODO, stubbed")
}

// Saig_TsiNext (static inline)
// C: unsigned * Saig_TsiNext( unsigned * pState, int nWords )
func saigTsiNext(pState []uint32, nWords int) []uint32 {
	panic("TODO, stubbed")
}

// Saig_TsiSetNext (static inline)
// C: void Saig_TsiSetNext( unsigned * pState, int nWords, unsigned * pNext )
func saigTsiSetNext(pState []uint32, nWords int, pNext []uint32) {
	panic("TODO, stubbed")
}

// Saig_TsiStart
// C: Saig_Tsim_t * Saig_TsiStart( Aig_Man_t * pAig )
func (pAig *AigManager) SaigTsiStart() *SaigTsim {
	panic("TODO, stubbed")
}

// Saig_TsiStop
// C: void Saig_TsiStop( Saig_Tsim_t * p )
func (p *SaigTsim) SaigTsiStop() {
	panic("TODO, stubbed")
}

// Saig_TsiStateHash
// C: int Saig_TsiStateHash( unsigned * pState, int nWords, int nTableSize )
func SaigTsiStateHash(pState []uint32, nWords int, nTableSize int) int {
	panic("TODO, stubbed")
}

// Saig_TsiCountNonXValuedRegisters
// C: int Saig_TsiCountNonXValuedRegisters( Saig_Tsim_t * p, int nPref )
func (p *SaigTsim) SaigTsiCountNonXValuedRegisters(nPref int) int {
	panic("TODO, stubbed")
}

// Saig_TsiComputeTransient
// C: Vec_Int_t * Saig_TsiComputeTransient( Saig_Tsim_t * p, int nPref )
func (p *SaigTsim) SaigTsiComputeTransient(nPref int) *IntVector {
	panic("TODO, stubbed")
}

// Saig_TsiPrintTraces
// C: void Saig_TsiPrintTraces( Saig_Tsim_t * p, int nWords, int nPrefix, int nLoop )
func (p *SaigTsim) SaigTsiPrintTraces(nWords int, nPrefix int, nLoop int) {
	panic("TODO, stubbed")
}

// Saig_TsiComputePrefix
// C: int Saig_TsiComputePrefix( Saig_Tsim_t * p, unsigned * pState, int nWords )
func (p *SaigTsim) SaigTsiComputePrefix(pState []uint32, nWords int) int {
	panic("TODO, stubbed")
}

// Saig_TsiStateLookup
// C: int Saig_TsiStateLookup( Saig_Tsim_t * p, unsigned * pState, int nWords )
func (p *SaigTsim) SaigTsiStateLookup(pState []uint32, nWords int) int {
	panic("TODO, stubbed")
}

// Saig_TsiStateInsert
// C: void Saig_TsiStateInsert( Saig_Tsim_t * p, unsigned * pState, int nWords )
func (p *SaigTsim) SaigTsiStateInsert(pState []uint32, nWords int) {
	panic("TODO, stubbed")
}

// Saig_TsiStateNew
// C: unsigned * Saig_TsiStateNew( Saig_Tsim_t * p )
func (p *SaigTsim) SaigTsiStateNew() []uint32 {
	panic("TODO, stubbed")
}

// Saig_TsiStatePrint
// C: void Saig_TsiStatePrint( Saig_Tsim_t * p, unsigned * pState )
func (p *SaigTsim) SaigTsiStatePrint(pState []uint32) {
	panic("TODO, stubbed")
}

// Saig_TsiStateCount
// C: int Saig_TsiStateCount( Saig_Tsim_t * p, unsigned * pState )
func (p *SaigTsim) SaigTsiStateCount(pState []uint32) int {
	panic("TODO, stubbed")
}

// Saig_TsiStateOrAll
// C: void Saig_TsiStateOrAll( Saig_Tsim_t * pTsi, unsigned * pState )
func (pTsi *SaigTsim) SaigTsiStateOrAll(pState []uint32) {
	panic("TODO, stubbed")
}

// Saig_ManReachableTernary
// C: Saig_Tsim_t * Saig_ManReachableTernary( Aig_Man_t * p, Vec_Int_t * vInits, int fVerbose )
func (p *AigManager) SaigManReachableTernary(vInits *IntVector, fVerbose int) *SaigTsim {
	panic("TODO, stubbed")
}

// Saig_ManAnalizeControl
// C: void Saig_ManAnalizeControl( Aig_Man_t * p, int Reg )
func (p *AigManager) SaigManAnalizeControl(Reg int) {
	panic("TODO, stubbed")
}

// Saig_ManFindRegisters
// C: int Saig_ManFindRegisters( Saig_Tsim_t * pTsi, int nFrames, int fIgnore, int fVerbose )
func (pTsi *SaigTsim) SaigManFindRegisters(nFrames int, fIgnore int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Saig_ObjFrames (static inline)
// C: Aig_Obj_t * Saig_ObjFrames( Aig_Obj_t ** pObjMap, int nFs, Aig_Obj_t * pObj, int i )
func saigObjFrames(pObjMap []*AigObject, nFs int, pObj *AigObject, i int) *AigObject {
	panic("TODO, stubbed")
}

// Saig_ObjSetFrames (static inline)
// C: void Saig_ObjSetFrames( Aig_Obj_t ** pObjMap, int nFs, Aig_Obj_t * pObj, int i, Aig_Obj_t * pNode )
func saigObjSetFrames(pObjMap []*AigObject, nFs int, pObj *AigObject, i int, pNode *AigObject) {
	panic("TODO, stubbed")
}

// Saig_ObjChild0Frames (static inline)
// C: Aig_Obj_t * Saig_ObjChild0Frames( Aig_Obj_t ** pObjMap, int nFs, Aig_Obj_t * pObj, int i )
func saigObjChild0Frames(pObjMap []*AigObject, nFs int, pObj *AigObject, i int) *AigObject {
	panic("TODO, stubbed")
}

// Saig_ObjChild1Frames (static inline)
// C: Aig_Obj_t * Saig_ObjChild1Frames( Aig_Obj_t ** pObjMap, int nFs, Aig_Obj_t * pObj, int i )
func saigObjChild1Frames(pObjMap []*AigObject, nFs int, pObj *AigObject, i int) *AigObject {
	panic("TODO, stubbed")
}

// Saig_ManPerformAbstraction
// C: Aig_Man_t * Saig_ManPerformAbstraction( Saig_Tsim_t * pTsi, int nFrames, int fVerbose )
func (pTsi *SaigTsim) SaigManPerformAbstraction(nFrames int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManPhaseFrameNum
// C: int Saig_ManPhaseFrameNum( Aig_Man_t * p, Vec_Int_t * vInits )
func (p *AigManager) SaigManPhaseFrameNum(vInits *IntVector) int {
	panic("TODO, stubbed")
}

// Saig_ManPhasePrefixLength
// C: int Saig_ManPhasePrefixLength( Aig_Man_t * p, int fVerbose, int fVeryVerbose, Vec_Int_t ** pvTrans )
func (p *AigManager) SaigManPhasePrefixLength(fVerbose int, fVeryVerbose int, pvTrans **IntVector) int {
	panic("TODO, stubbed")
}

// Saig_ManPhaseAbstract
// C: Aig_Man_t * Saig_ManPhaseAbstract( Aig_Man_t * p, Vec_Int_t * vInits, int nFrames, int nPref, int fIgnore, int fPrint, int fVerbose )
func (p *AigManager) SaigManPhaseAbstract(vInits *IntVector, nFrames int, nPref int, fIgnore int, fPrint int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManPhaseAbstractAuto
// C: Aig_Man_t * Saig_ManPhaseAbstractAuto( Aig_Man_t * p, int fVerbose )
func (p *AigManager) SaigManPhaseAbstractAuto(fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_PhaseTranslateCex
// C: Abc_Cex_t * Saig_PhaseTranslateCex( Aig_Man_t * p, Abc_Cex_t * pCex )
func (p *AigManager) SaigPhaseTranslateCex(pCex *Counterexample) *Counterexample {
	panic("TODO, stubbed")
}
