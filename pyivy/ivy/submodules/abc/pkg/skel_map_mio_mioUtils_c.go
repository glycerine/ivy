package abc

import (
	"os"
	"strings"
)

// Hand-checked skeleton of src/map/mio/mioUtils.c.
//
// Mio_DeriveTruthTable_rec and an earlier definition of Mio_DeriveTruthTable
// sit inside an "#if 0" block and are not compiled; the live
// Mio_DeriveTruthTable follows it.

// Mio_LibraryDelete
// C: void Mio_LibraryDelete( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryDelete() {
	panic("TODO, stubbed")
}

// Mio_GateDelete
// C: void Mio_GateDelete( Mio_Gate_t * pGate )
func (pGate *MioGate) MioGateDelete() {
	panic("TODO, stubbed")
}

// Mio_PinDelete
// C: void Mio_PinDelete( Mio_Pin_t * pPin )
func (pPin *MioPin) MioPinDelete() {
	panic("TODO, stubbed")
}

// Mio_PinDup
// C: Mio_Pin_t * Mio_PinDup( Mio_Pin_t * pPin )
func (pPin *MioPin) MioPinDup() *MioPin {
	panic("TODO, stubbed")
}

// Mio_CheckPins
// C: int Mio_CheckPins( Mio_Pin_t * pPin1, Mio_Pin_t * pPin2 )
func (pPin1 *MioPin) MioCheckPins(pPin2 *MioPin) int {
	panic("TODO, stubbed")
}

// Mio_CheckGates
// C: int Mio_CheckGates( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioCheckGates() int {
	panic("TODO, stubbed")
}

// Mio_WritePin
// C: void Mio_WritePin( FILE * pFile, Mio_Pin_t * pPin, int NameLen, int fAllPins )
func MioWritePin(pFile *os.File, pPin *MioPin, NameLen int, fAllPins int) {
	panic("TODO, stubbed")
}

// Mio_WriteGate
// C: void Mio_WriteGate( FILE * pFile, Mio_Gate_t * pGate, int GateLen, int NameLen, int FormLen, int fPrintSops, int fAllPins )
func MioWriteGate(pFile *os.File, pGate *MioGate, GateLen int, NameLen int, FormLen int, fPrintSops int, fAllPins int) {
	panic("TODO, stubbed")
}

// Mio_WriteLibrary
// C: void Mio_WriteLibrary( FILE * pFile, Mio_Library_t * pLib, int fPrintSops, int fShort, int fSelected )
func MioWriteLibrary(pFile *os.File, pLib *MioLibrary, fPrintSops int, fShort int, fSelected int) {
	panic("TODO, stubbed")
}

// Exp_PrintNodeVerilog
// C: void Exp_PrintNodeVerilog( FILE * pFile, int nVars, Vec_Int_t * p, Vec_Ptr_t * vNames, int Node, int fCompl )
func ExpPrintNodeVerilog(pFile *os.File, nVars int, p *IntVector, vNames *PointerVector[string], Node int, fCompl int) {
	panic("TODO, stubbed")
}

// Exp_PrintLitVerilog
// C: void Exp_PrintLitVerilog( FILE * pFile, int nVars, Vec_Int_t * p, Vec_Ptr_t * vNames, int Lit )
func ExpPrintLitVerilog(pFile *os.File, nVars int, p *IntVector, vNames *PointerVector[string], Lit int) {
	panic("TODO, stubbed")
}

// Exp_PrintVerilog
// C: void Exp_PrintVerilog( FILE * pFile, int nVars, Vec_Int_t * p, Vec_Ptr_t * vNames )
func ExpPrintVerilog(pFile *os.File, nVars int, p *IntVector, vNames *PointerVector[string]) {
	panic("TODO, stubbed")
}

// Mio_WriteGateVerilog
// C: void Mio_WriteGateVerilog( FILE * pFile, Mio_Gate_t * pGate, Vec_Ptr_t * vNames )
func MioWriteGateVerilog(pFile *os.File, pGate *MioGate, vNames *PointerVector[string]) {
	panic("TODO, stubbed")
}

// Mio_WriteLibraryVerilog
// C: void Mio_WriteLibraryVerilog( FILE * pFile, Mio_Library_t * pLib, int fPrintSops, int fShort, int fSelected )
func MioWriteLibraryVerilog(pFile *os.File, pLib *MioLibrary, fPrintSops int, fShort int, fSelected int) {
	panic("TODO, stubbed")
}

// Mio_DelayCompare
// C: int Mio_DelayCompare( Mio_Gate_t ** ppG1, Mio_Gate_t ** ppG2 )
func MioDelayCompare(ppG1 **MioGate, ppG2 **MioGate) int {
	panic("TODO, stubbed")
}

// Mio_AreaCompare
// C: int Mio_AreaCompare( Mio_Cell_t * pG1, Mio_Cell_t * pG2 )
func MioAreaCompare(pG1 *MioCell, pG2 *MioCell) int {
	panic("TODO, stubbed")
}

// Mio_AreaCompare2
// C: int Mio_AreaCompare2( Mio_Cell2_t * pG1, Mio_Cell2_t * pG2 )
func MioAreaCompare2(pG1 *MioCell2, pG2 *MioCell2) int {
	panic("TODO, stubbed")
}

// Mio_CellDelayAve (static)
// C: float Mio_CellDelayAve( Mio_Cell_t * pCell )
func mioCellDelayAve(pCell *MioCell) float32 {
	panic("TODO, stubbed")
}

// Mio_GateDelayAve (static)
// C: float Mio_GateDelayAve( Mio_Gate_t * pGate )
func mioGateDelayAve(pGate *MioGate) float32 {
	panic("TODO, stubbed")
}

// Mio_CompareTwoGates (static)
// C: int Mio_CompareTwoGates( Mio_Gate_t * pCell, Mio_Gate_t * pGate )
func mioCompareTwoGates(pCell *MioGate, pGate *MioGate) int {
	panic("TODO, stubbed")
}

// Mio_CollectRoots (pnGates is an out parameter; returns an array of gate pointers)
// C: Mio_Gate_t ** Mio_CollectRoots( Mio_Library_t * pLib, int nInputs, float tDelay, int fSkipInv, int * pnGates, int fVerbose )
func (pLib *MioLibrary) MioCollectRoots(nInputs int, tDelay float32, fSkipInv int, pnGates *int, fVerbose int) []*MioGate {
	panic("TODO, stubbed")
}

// Mio_CompareTwo (static)
// C: int Mio_CompareTwo( Mio_Cell_t * pCell, Mio_Gate_t * pGate )
func mioCompareTwo(pCell *MioCell, pGate *MioGate) int {
	panic("TODO, stubbed")
}

// Mio_CollectCopy (static)
// C: void Mio_CollectCopy( Mio_Cell_t * pCell, Mio_Gate_t * pGate )
func mioCollectCopy(pCell *MioCell, pGate *MioGate) {
	panic("TODO, stubbed")
}

// Mio_CollectRootsNew (pnGates is an out parameter; returns an array of cells)
// C: Mio_Cell_t * Mio_CollectRootsNew( Mio_Library_t * pLib, int nInputs, int * pnGates, int fVerbose )
func (pLib *MioLibrary) MioCollectRootsNew(nInputs int, pnGates *int, fVerbose int) []MioCell {
	panic("TODO, stubbed")
}

// Mio_CollectRootsNewDefault (pnGates is an out parameter; returns an array of cells)
// C: Mio_Cell_t * Mio_CollectRootsNewDefault( int nInputs, int * pnGates, int fVerbose )
func MioCollectRootsNewDefault(nInputs int, pnGates *int, fVerbose int) []MioCell {
	panic("TODO, stubbed")
}

// Mio_CompareTwo2 (static)
// C: int Mio_CompareTwo2( Mio_Cell2_t * pCell1, Mio_Cell2_t * pCell2 )
func mioCompareTwo2(pCell1 *MioCell2, pCell2 *MioCell2) int {
	panic("TODO, stubbed")
}

// Mio_CollectCopy2 (static)
// C: void Mio_CollectCopy2( Mio_Cell2_t * pCell, Mio_Gate_t * pGate )
func mioCollectCopy2(pCell *MioCell2, pGate *MioGate) {
	panic("TODO, stubbed")
}

// Mio_CollectRootsNew2 (pnGates is an out parameter; returns an array of cells)
// C: Mio_Cell2_t * Mio_CollectRootsNew2( Mio_Library_t * pLib, int nInputs, int * pnGates, int fVerbose )
func (pLib *MioLibrary) MioCollectRootsNew2(nInputs int, pnGates *int, fVerbose int) []MioCell2 {
	panic("TODO, stubbed")
}

// Mio_CollectRootsNewDefault2 (pnGates is an out parameter; returns an array of cells)
// C: Mio_Cell2_t * Mio_CollectRootsNewDefault2( int nInputs, int * pnGates, int fVerbose )
func MioCollectRootsNewDefault2(nInputs int, pnGates *int, fVerbose int) []MioCell2 {
	panic("TODO, stubbed")
}

// Mio_CollectRootsNewDefault3 (pvNames and pvTruths are out parameters)
// C: int Mio_CollectRootsNewDefault3( int nInputs, Vec_Ptr_t ** pvNames, Vec_Wrd_t ** pvTruths )
func MioCollectRootsNewDefault3(nInputs int, pvNames **PointerVector[string], pvTruths **WordVector) int {
	panic("TODO, stubbed")
}

// Mio_DeriveTruthTable6
// C: word Mio_DeriveTruthTable6( Mio_Gate_t * pGate )
func (pGate *MioGate) MioDeriveTruthTable6() uint64 {
	panic("TODO, stubbed")
}

// Mio_DeriveTruthTable
// C: void Mio_DeriveTruthTable( Mio_Gate_t * pGate, unsigned uTruthsIn[][2], int nSigns, int nInputs, unsigned uTruthRes[] )
func (pGate *MioGate) MioDeriveTruthTable(uTruthsIn [][2]uint32, nSigns int, nInputs int, uTruthRes []uint32) {
	panic("TODO, stubbed")
}

// Mio_SopGetVarNum
// C: int Mio_SopGetVarNum( char * pSop )
func MioSopGetVarNum(pSop string) int {
	panic("TODO, stubbed")
}

// Mio_DeriveTruthTable2
// C: void Mio_DeriveTruthTable2( Mio_Gate_t * pGate, unsigned uTruthsIn[][2], int nTruths, int nInputs, unsigned uTruthRes[] )
func (pGate *MioGate) MioDeriveTruthTable2(uTruthsIn [][2]uint32, nTruths int, nInputs int, uTruthRes []uint32) {
	panic("TODO, stubbed")
}

// Mio_DeriveGateDelays (ptPinDelayMax is an out parameter)
// C: void Mio_DeriveGateDelays( Mio_Gate_t * pGate, float ** ptPinDelays, int nPins, int nInputs, float tDelayZero, float * ptDelaysRes, float * ptPinDelayMax )
func (pGate *MioGate) MioDeriveGateDelays(ptPinDelays [][]float32, nPins int, nInputs int, tDelayZero float32, ptDelaysRes []float32, ptPinDelayMax *float32) {
	panic("TODO, stubbed")
}

// Mio_GateCreatePseudo
// C: Mio_Gate_t * Mio_GateCreatePseudo( int nInputs )
func MioGateCreatePseudo(nInputs int) *MioGate {
	panic("TODO, stubbed")
}

// Mio_LibraryShiftDelay
// C: void Mio_LibraryShiftDelay( Mio_Library_t * pLib, double Shift )
func (pLib *MioLibrary) MioLibraryShiftDelay(Shift float64) {
	panic("TODO, stubbed")
}

// Mio_LibraryMultiArea
// C: void Mio_LibraryMultiArea( Mio_Library_t * pLib, double Multi )
func (pLib *MioLibrary) MioLibraryMultiArea(Multi float64) {
	panic("TODO, stubbed")
}

// Mio_LibraryMultiDelay
// C: void Mio_LibraryMultiDelay( Mio_Library_t * pLib, double Multi )
func (pLib *MioLibrary) MioLibraryMultiDelay(Multi float64) {
	panic("TODO, stubbed")
}

// Mio_LibraryTransferDelays
// C: void Mio_LibraryTransferDelays( Mio_Library_t * pLibD, Mio_Library_t * pLibS )
func (pLibD *MioLibrary) MioLibraryTransferDelays(pLibS *MioLibrary) {
	panic("TODO, stubbed")
}

// Nf_ManPrepareGate
// C: void Nf_ManPrepareGate( int nVars, word uTruth, int * pComp, int * pPerm, Vec_Wrd_t * vResult )
func NfManPrepareGate(nVars int, uTruth uint64, pComp []int, pPerm []int, vResult *WordVector) {
	panic("TODO, stubbed")
}

// Nf_ManPreparePrint (Line is char Line[2*720*64][8], filled in)
// C: void Nf_ManPreparePrint( int nVars, int * pComp, int * pPerm, char Line[2*720*64][8] )
func NfManPreparePrint(nVars int, pComp []int, pPerm []int, Line [][8]int8) {
	panic("TODO, stubbed")
}

// Nf_ManPrepareLibrary
// C: void Nf_ManPrepareLibrary( Mio_Library_t * pLib )
func (pLib *MioLibrary) NfManPrepareLibrary() {
	panic("TODO, stubbed")
}

// Nf_ManPrepareLibraryTest2
// C: void Nf_ManPrepareLibraryTest2()
func NfManPrepareLibraryTest2() {
	panic("TODO, stubbed")
}

// Mio_LibraryTransferCellIds
// C: void Mio_LibraryTransferCellIds()
func MioLibraryTransferCellIds() {
	panic("TODO, stubbed")
}

// Mio_LibraryReadProfile
// C: void Mio_LibraryReadProfile( FILE * pFile, Mio_Library_t * pLib )
func MioLibraryReadProfile(pFile *os.File, pLib *MioLibrary) {
	panic("TODO, stubbed")
}

// Mio_LibraryWriteProfile
// C: void Mio_LibraryWriteProfile( FILE * pFile, Mio_Library_t * pLib )
func MioLibraryWriteProfile(pFile *os.File, pLib *MioLibrary) {
	panic("TODO, stubbed")
}

// Mio_LibraryHasProfile
// C: int Mio_LibraryHasProfile( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryHasProfile() int {
	panic("TODO, stubbed")
}

// Mio_LibraryTransferProfile
// C: void Mio_LibraryTransferProfile( Mio_Library_t * pLibDst, Mio_Library_t * pLibSrc )
func (pLibDst *MioLibrary) MioLibraryTransferProfile(pLibSrc *MioLibrary) {
	panic("TODO, stubbed")
}

// Mio_LibraryTransferProfile2
// C: void Mio_LibraryTransferProfile2( Mio_Library_t * pLibDst, Mio_Library_t * pLibSrc )
func (pLibDst *MioLibrary) MioLibraryTransferProfile2(pLibSrc *MioLibrary) {
	panic("TODO, stubbed")
}

// Mio_LibraryCleanProfile2
// C: void Mio_LibraryCleanProfile2( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryCleanProfile2() {
	panic("TODO, stubbed")
}

// Mio_LibraryHashGates
// C: void Mio_LibraryHashGates( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryHashGates() {
	panic("TODO, stubbed")
}

// Abc_SclIsChar (static)
// C: int Abc_SclIsChar( char c )
func abcSclIsChar(c int8) int {
	panic("TODO, stubbed")
}

// Abc_SclIsName (static)
// C: int Abc_SclIsName( char c )
func abcSclIsName(c int8) int {
	panic("TODO, stubbed")
}

// Abc_SclFindLimit (static; returns the suffix of pName at the limit)
// C: char * Abc_SclFindLimit( char * pName )
func abcSclFindLimit(pName string) string {
	panic("TODO, stubbed")
}

// Abc_SclAreEqual (static)
// C: int Abc_SclAreEqual( char * pBase, char * pName, char * pLimit )
func abcSclAreEqual(pBase string, pName string, pLimit string) int {
	panic("TODO, stubbed")
}

// Mio_LibraryShortFormula (pBuffer is written)
// C: void Mio_LibraryShortFormula( Mio_Gate_t * pCell, char * pForm, char * pBuffer )
func (pCell *MioGate) MioLibraryShortFormula(pForm string, pBuffer *strings.Builder) {
	panic("TODO, stubbed")
}

// Mio_LibraryShortNames
// C: void Mio_LibraryShortNames( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryShortNames() {
	panic("TODO, stubbed")
}

// Mio_LibraryMatchesStop
// C: void Mio_LibraryMatchesStop( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryMatchesStop() {
	panic("TODO, stubbed")
}

// Mio_LibraryMatchesStart
// C: void Mio_LibraryMatchesStart( Mio_Library_t * pLib, int fPinFilter, int fPinPerm, int fPinQuick )
func (pLib *MioLibrary) MioLibraryMatchesStart(fPinFilter int, fPinPerm int, fPinQuick int) {
	panic("TODO, stubbed")
}

// Mio_LibraryMatchesFetch (pvTtMem, pvTt2Match, ppCells and pnCells are out parameters)
// C: void Mio_LibraryMatchesFetch( Mio_Library_t * pLib, Vec_Mem_t ** pvTtMem, Vec_Wec_t ** pvTt2Match, Mio_Cell2_t ** ppCells, int * pnCells, int fPinFilter, int fPinPerm, int fPinQuick )
func (pLib *MioLibrary) MioLibraryMatchesFetch(pvTtMem **WordMemory, pvTt2Match **IntLevelVector, ppCells *[]MioCell2, pnCells *int, fPinFilter int, fPinPerm int, fPinQuick int) {
	panic("TODO, stubbed")
}

// Mio_LibraryMatches2Stop
// C: void Mio_LibraryMatches2Stop( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryMatches2Stop() {
	panic("TODO, stubbed")
}

// Mio_LibraryMatches2Start
// C: void Mio_LibraryMatches2Start( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryMatches2Start() {
	panic("TODO, stubbed")
}

// Mio_LibraryMatches2Fetch (the first four are out parameters; the array parameters are filled in)
// C: void Mio_LibraryMatches2Fetch( Mio_Library_t * pLib, Vec_Ptr_t ** pvNames, Vec_Wrd_t ** pvTruths, Vec_Int_t ** pvTt2Match4, Vec_Int_t ** pvConfigs, Vec_Mem_t * pvTtMem2[3], Vec_Int_t * pvTt2Match2[3] )
func (pLib *MioLibrary) MioLibraryMatches2Fetch(pvNames **PointerVector[string], pvTruths **WordVector, pvTt2Match4 **IntVector, pvConfigs **IntVector, pvTtMem2 [3]*WordMemory, pvTt2Match2 [3]*IntVector) {
	panic("TODO, stubbed")
}
