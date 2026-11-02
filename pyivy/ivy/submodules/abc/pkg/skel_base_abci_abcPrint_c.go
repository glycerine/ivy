package abc

import "os"

// Hand-checked skeleton of src/base/abci/abcPrint.c.

// Abc_NtkCompareAndSaveBest
// C: int Abc_NtkCompareAndSaveBest( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkCompareAndSaveBest() int {
	panic("TODO, stubbed")
}

// Abc_NtkMemory
// C: double Abc_NtkMemory( Abc_Ntk_t * p )
func (p *ABCNetwork) AbcNtkMemory() float64 {
	panic("TODO, stubbed")
}

// Abc_NtkMfsTotalSwitching
// C: float Abc_NtkMfsTotalSwitching( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkMfsTotalSwitching() float32 {
	panic("TODO, stubbed")
}

// Abc_NtkGetArea
// C: float Abc_NtkGetArea( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkGetArea() float32 {
	panic("TODO, stubbed")
}

// Abc_NtkPrintStats
// C: void Abc_NtkPrintStats( Abc_Ntk_t * pNtk, int fFactored, int fSaveBest, int fDumpResult, int fUseLutLib, int fPrintMuxes, int fPower, int fGlitch, int fSkipBuf, int fSkipSmall, int fPrintMem )
func (pNtk *ABCNetwork) AbcNtkPrintStats(fFactored int, fSaveBest int, fDumpResult int, fUseLutLib int, fPrintMuxes int, fPower int, fGlitch int, fSkipBuf int, fSkipSmall int, fPrintMem int) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintIo
// C: void Abc_NtkPrintIo( FILE * pFile, Abc_Ntk_t * pNtk, int fPrintFlops )
func AbcNtkPrintIo(pFile *os.File, pNtk *ABCNetwork, fPrintFlops int) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintLatch
// C: void Abc_NtkPrintLatch( FILE * pFile, Abc_Ntk_t * pNtk )
func AbcNtkPrintLatch(pFile *os.File, pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NtkFaninFanoutCounters
// C: void Abc_NtkFaninFanoutCounters( Abc_Ntk_t * pNtk, Vec_Int_t * vFan, Vec_Int_t * vFon, Vec_Int_t * vFanR, Vec_Int_t * vFonR )
func (pNtk *ABCNetwork) AbcNtkFaninFanoutCounters(vFan *IntVector, vFon *IntVector, vFanR *IntVector, vFonR *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkInputOutputCounters
// C: void Abc_NtkInputOutputCounters( Abc_Ntk_t * pNtk, Vec_Int_t * vFan, Vec_Int_t * vFon, Vec_Int_t * vFanR, Vec_Int_t * vFonR )
func (pNtk *ABCNetwork) AbcNtkInputOutputCounters(vFan *IntVector, vFon *IntVector, vFanR *IntVector, vFonR *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkCollectCoSupps
// C: Vec_Int_t * Abc_NtkCollectCoSupps( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkCollectCoSupps(fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkCollectCiSupps
// C: Vec_Int_t * Abc_NtkCollectCiSupps( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkCollectCiSupps(fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkInOutSupportCounters
// C: void Abc_NtkInOutSupportCounters( Abc_Ntk_t * pNtk, Vec_Int_t * vFan, Vec_Int_t * vFon, Vec_Int_t * vFanR, Vec_Int_t * vFonR )
func (pNtk *ABCNetwork) AbcNtkInOutSupportCounters(vFan *IntVector, vFon *IntVector, vFanR *IntVector, vFonR *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkCollectCoCones
// C: Vec_Int_t * Abc_NtkCollectCoCones( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkCollectCoCones(fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkCollectCiCones
// C: Vec_Int_t * Abc_NtkCollectCiCones( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkCollectCiCones(fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkInOutConeCounters
// C: void Abc_NtkInOutConeCounters( Abc_Ntk_t * pNtk, Vec_Int_t * vFan, Vec_Int_t * vFon, Vec_Int_t * vFanR, Vec_Int_t * vFonR )
func (pNtk *ABCNetwork) AbcNtkInOutConeCounters(vFan *IntVector, vFon *IntVector, vFanR *IntVector, vFonR *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintDistribInternal
// C: void Abc_NtkPrintDistribInternal( FILE * pFile, Abc_Ntk_t * pNtk, char * pFanins, char * pFanouts, char * pNode, char * pFanin, char * pFanout, Vec_Int_t * vFan, Vec_Int_t * vFon, Vec_Int_t * vFanR, Vec_Int_t * vFonR )
func AbcNtkPrintDistribInternal(pFile *os.File, pNtk *ABCNetwork, pFanins string, pFanouts string, pNode string, pFanin string, pFanout string, vFan *IntVector, vFon *IntVector, vFanR *IntVector, vFonR *IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintFanio
// C: void Abc_NtkPrintFanio( FILE * pFile, Abc_Ntk_t * pNtk, int fUseFanio, int fUsePio, int fUseSupp, int fUseCone )
func AbcNtkPrintFanio(pFile *os.File, pNtk *ABCNetwork, fUseFanio int, fUsePio int, fUseSupp int, fUseCone int) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintFanioNew
// C: void Abc_NtkPrintFanioNew( FILE * pFile, Abc_Ntk_t * pNtk, int fMffc )
func AbcNtkPrintFanioNew(pFile *os.File, pNtk *ABCNetwork, fMffc int) {
	panic("TODO, stubbed")
}

// Abc_NodePrintFanio
// C: void Abc_NodePrintFanio( FILE * pFile, Abc_Obj_t * pNode )
func AbcNodePrintFanio(pFile *os.File, pNode *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintMffc
// C: void Abc_NtkPrintMffc( FILE * pFile, Abc_Ntk_t * pNtk )
func AbcNtkPrintMffc(pFile *os.File, pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintFactor
// C: void Abc_NtkPrintFactor( FILE * pFile, Abc_Ntk_t * pNtk, int fUseRealNames )
func AbcNtkPrintFactor(pFile *os.File, pNtk *ABCNetwork, fUseRealNames int) {
	panic("TODO, stubbed")
}

// Abc_NodePrintFactor
// C: void Abc_NodePrintFactor( FILE * pFile, Abc_Obj_t * pNode, int fUseRealNames )
func AbcNodePrintFactor(pFile *os.File, pNode *ABCNetworkObject, fUseRealNames int) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintLevel
// C: void Abc_NtkPrintLevel( FILE * pFile, Abc_Ntk_t * pNtk, int fProfile, int fListNodes, int fVerbose )
func AbcNtkPrintLevel(pFile *os.File, pNtk *ABCNetwork, fProfile int, fListNodes int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NodePrintLevel
// C: void Abc_NodePrintLevel( FILE * pFile, Abc_Obj_t * pNode )
func AbcNodePrintLevel(pFile *os.File, pNode *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NodePrintKMap
// C: void Abc_NodePrintKMap( Abc_Obj_t * pNode, int fUseRealNames )
func AbcNodePrintKMap(pNode *ABCNetworkObject, fUseRealNames int) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintGates
// C: void Abc_NtkPrintGates( Abc_Ntk_t * pNtk, int fUseLibrary, int fUpdateProfile )
func (pNtk *ABCNetwork) AbcNtkPrintGates(fUseLibrary int, fUpdateProfile int) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintSharing
// C: void Abc_NtkPrintSharing( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkPrintSharing() {
	panic("TODO, stubbed")
}

// Abc_NtkCountPis
// C: int Abc_NtkCountPis( Vec_Ptr_t * vSupp )
func AbcNtkCountPis(vSupp *PointerVector[ABCNetworkObject]) int {
	panic("TODO, stubbed")
}

// Abc_NtkPrintStrSupports
// C: void Abc_NtkPrintStrSupports( Abc_Ntk_t * pNtk, int fMatrix )
func (pNtk *ABCNetwork) AbcNtkPrintStrSupports(fMatrix int) {
	panic("TODO, stubbed")
}

// Abc_ObjPrint
// C: void Abc_ObjPrint( FILE * pFile, Abc_Obj_t * pObj )
func AbcObjPrint(pFile *os.File, pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintMiter
// C: void Abc_NtkPrintMiter( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkPrintMiter() {
	panic("TODO, stubbed")
}

// Abc_NtkPrintPoEquivs
// C: void Abc_NtkPrintPoEquivs( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkPrintPoEquivs() {
	panic("TODO, stubbed")
}

// Abc_NtkMfsTotalGlitchingLut
// C: float Abc_NtkMfsTotalGlitchingLut( Abc_Ntk_t * pNtk, int nPats, int Prob, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkMfsTotalGlitchingLut(nPats int, Prob int, fVerbose int) float32 {
	panic("TODO, stubbed")
}

// Abc_NtkMfsTotalGlitching
// C: float Abc_NtkMfsTotalGlitching( Abc_Ntk_t * pNtk, int nPats, int Prob, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkMfsTotalGlitching(nPats int, Prob int, fVerbose int) float32 {
	panic("TODO, stubbed")
}

// Abc_Show6VarFunc
// C: void Abc_Show6VarFunc( word F0, word F1 )
func AbcShow6VarFunc(F0 uint64, F1 uint64) {
	panic("TODO, stubbed")
}

// Abc_NtkShow6VarFunc
// C: void Abc_NtkShow6VarFunc( char * pF0, char * pF1 )
func AbcNtkShow6VarFunc(pF0 string, pF1 string) {
	panic("TODO, stubbed")
}
