package abc

import "strings"

// Hand-checked skeleton of src/map/scl/sclLibUtil.c.

// Abc_SclHashString (static)
// C: unsigned Abc_SclHashString( char * pName, int TableSize )
func abcSclHashString(pName string, TableSize int) uint32 {
	panic("TODO, stubbed")
}

// Abc_SclHashLookup (returns the address of the table slot)
// C: int * Abc_SclHashLookup( SC_Lib * p, char * pName )
func (p *SCLib) AbcSclHashLookup(pName string) *int {
	panic("TODO, stubbed")
}

// Abc_SclHashCells
// C: void Abc_SclHashCells( SC_Lib * p )
func (p *SCLib) AbcSclHashCells() {
	panic("TODO, stubbed")
}

// Abc_SclCellFind
// C: int Abc_SclCellFind( SC_Lib * p, char * pName )
func (p *SCLib) AbcSclCellFind(pName string) int {
	panic("TODO, stubbed")
}

// Abc_SclClassCellNum
// C: int Abc_SclClassCellNum( SC_Cell * pClass )
func (pClass *SCCell) AbcSclClassCellNum() int {
	panic("TODO, stubbed")
}

// Abc_SclLibClassNum
// C: int Abc_SclLibClassNum( SC_Lib * pLib )
func (pLib *SCLib) AbcSclLibClassNum() int {
	panic("TODO, stubbed")
}

// Abc_SclIsChar (static; mioUtils.c has a static of the same name, so this one carries a _scl suffix)
// C: int Abc_SclIsChar( char c )
func abcSclIsChar_scl(c int8) int {
	panic("TODO, stubbed")
}

// Abc_SclIsName (static; mioUtils.c has a static of the same name, so this one carries a _scl suffix)
// C: int Abc_SclIsName( char c )
func abcSclIsName_scl(c int8) int {
	panic("TODO, stubbed")
}

// Abc_SclFindLimit (static; mioUtils.c has a static of the same name, so this one carries a _scl suffix)
// C: char * Abc_SclFindLimit( char * pName )
func abcSclFindLimit_scl(pName string) string {
	panic("TODO, stubbed")
}

// Abc_SclAreEqual (static; mioUtils.c has a static of the same name, so this one carries a _scl suffix)
// C: int Abc_SclAreEqual( char * pBase, char * pName, char * pLimit )
func abcSclAreEqual_scl(pBase string, pName string, pLimit string) int {
	panic("TODO, stubbed")
}

// Abc_SclShortFormula (pBuffer is written)
// C: void Abc_SclShortFormula( SC_Cell * pCell, char * pForm, char * pBuffer )
func (pCell *SCCell) AbcSclShortFormula(pForm string, pBuffer *strings.Builder) {
	panic("TODO, stubbed")
}

// Abc_SclTimingUpdate (static; Buffer is a scratch character buffer)
// C: void Abc_SclTimingUpdate( SC_Cell * pCell, SC_Timing * p, char * Buffer )
func (pCell *SCCell) abcSclTimingUpdate(p *SCTiming, Buffer []byte) {
	panic("TODO, stubbed")
}

// Abc_SclTimingsUpdate (static; Buffer is a scratch character buffer)
// C: void Abc_SclTimingsUpdate( SC_Cell * pCell, SC_Timings * p, char * Buffer )
func (pCell *SCCell) abcSclTimingsUpdate(p *SCTimings, Buffer []byte) {
	panic("TODO, stubbed")
}

// Abc_SclPinUpdate (static; Buffer is a scratch character buffer)
// C: void Abc_SclPinUpdate( SC_Cell * pCell, SC_Pin * p, char * Buffer )
func (pCell *SCCell) abcSclPinUpdate(p *SCPin, Buffer []byte) {
	panic("TODO, stubbed")
}

// Abc_SclShortNames
// C: void Abc_SclShortNames( SC_Lib * p )
func (p *SCLib) AbcSclShortNames() {
	panic("TODO, stubbed")
}

// Abc_SclCompareCells (static)
// C: int Abc_SclCompareCells( SC_Cell ** pp1, SC_Cell ** pp2 )
func abcSclCompareCells(pp1 **SCCell, pp2 **SCCell) int {
	panic("TODO, stubbed")
}

// Abc_SclLinkCells
// C: void Abc_SclLinkCells( SC_Lib * p )
func (p *SCLib) AbcSclLinkCells() {
	panic("TODO, stubbed")
}

// Abc_SclFindInvertor
// C: SC_Cell * Abc_SclFindInvertor( SC_Lib * p, int fFindBuff )
func (p *SCLib) AbcSclFindInvertor(fFindBuff int) *SCCell {
	panic("TODO, stubbed")
}

// Abc_SclFindSmallestGate
// C: SC_Cell * Abc_SclFindSmallestGate( SC_Cell * p, float CinMin )
func (p *SCCell) AbcSclFindSmallestGate(CinMin float32) *SCCell {
	panic("TODO, stubbed")
}

// Abc_SclFetchWireLoadModel
// C: SC_WireLoad * Abc_SclFetchWireLoadModel( SC_Lib * p, char * pWLoadUsed )
func (p *SCLib) AbcSclFetchWireLoadModel(pWLoadUsed string) *SCWireLoad {
	panic("TODO, stubbed")
}

// Abc_SclFindWireLoadModel
// C: SC_WireLoad * Abc_SclFindWireLoadModel( SC_Lib * p, float Area )
func (p *SCLib) AbcSclFindWireLoadModel(Area float32) *SCWireLoad {
	panic("TODO, stubbed")
}

// Abc_SclHasDelayInfo (pScl is a void * that holds an SC_Lib *)
// C: int Abc_SclHasDelayInfo( void * pScl )
func AbcSclHasDelayInfo(pScl *SCLib) int {
	panic("TODO, stubbed")
}

// Abc_SclComputeAverageSlew
// C: float Abc_SclComputeAverageSlew( SC_Lib * p )
func (p *SCLib) AbcSclComputeAverageSlew() float32 {
	panic("TODO, stubbed")
}

// Abc_SclComputeParametersPin (pLD and pPD are out parameters)
// C: int Abc_SclComputeParametersPin( SC_Lib * p, SC_Cell * pCell, int iPin, float Slew, float * pLD, float * pPD )
func (p *SCLib) AbcSclComputeParametersPin(pCell *SCCell, iPin int, Slew float32, pLD *float32, pPD *float32) int {
	panic("TODO, stubbed")
}

// Abc_SclComputeParametersCell (pLD and pPD are out parameters)
// C: int Abc_SclComputeParametersCell( SC_Lib * p, SC_Cell * pCell, float Slew, float * pLD, float * pPD )
func (p *SCLib) AbcSclComputeParametersCell(pCell *SCCell, Slew float32, pLD *float32, pPD *float32) int {
	panic("TODO, stubbed")
}

// Abc_SclComputeParametersClass (pLD and pPD are out parameters)
// C: void Abc_SclComputeParametersClass( SC_Lib * p, SC_Cell * pRepr, float Slew, float * pLD, float * pPD )
func (p *SCLib) AbcSclComputeParametersClass(pRepr *SCCell, Slew float32, pLD *float32, pPD *float32) {
	panic("TODO, stubbed")
}

// Abc_SclComputeParametersClassPin (pLD and pPD are out parameters)
// C: void Abc_SclComputeParametersClassPin( SC_Lib * p, SC_Cell * pRepr, int iPin, float Slew, float * pLD, float * pPD )
func (p *SCLib) AbcSclComputeParametersClassPin(pRepr *SCCell, iPin int, Slew float32, pLD *float32, pPD *float32) {
	panic("TODO, stubbed")
}

// Abc_SclComputeDelayCellPin
// C: float Abc_SclComputeDelayCellPin( SC_Lib * p, SC_Cell * pCell, int iPin, float Slew, float Gain )
func (p *SCLib) AbcSclComputeDelayCellPin(pCell *SCCell, iPin int, Slew float32, Gain float32) float32 {
	panic("TODO, stubbed")
}

// Abc_SclComputeDelayClassPin
// C: float Abc_SclComputeDelayClassPin( SC_Lib * p, SC_Cell * pRepr, int iPin, float Slew, float Gain )
func (p *SCLib) AbcSclComputeDelayClassPin(pRepr *SCCell, iPin int, Slew float32, Gain float32) float32 {
	panic("TODO, stubbed")
}

// Abc_SclComputeAreaClass
// C: float Abc_SclComputeAreaClass( SC_Cell * pRepr )
func AbcSclComputeAreaClass(pRepr *SCCell) float32 {
	panic("TODO, stubbed")
}

// Abc_SclMarkSkippedCells
// C: void Abc_SclMarkSkippedCells( SC_Lib * p )
func (p *SCLib) AbcSclMarkSkippedCells() {
	panic("TODO, stubbed")
}

// Abc_SclPrintCells
// C: void Abc_SclPrintCells( SC_Lib * p, float SlewInit, float Gain, int fInvOnly, int fShort )
func (p *SCLib) AbcSclPrintCells(SlewInit float32, Gain float32, fInvOnly int, fShort int) {
	panic("TODO, stubbed")
}

// Abc_SclConvertLeakageIntoArea
// C: void Abc_SclConvertLeakageIntoArea( SC_Lib * p, float A, float B )
func (p *SCLib) AbcSclConvertLeakageIntoArea(A float32, B float32) {
	panic("TODO, stubbed")
}

// Abc_SclLibNormalizeSurface
// C: void Abc_SclLibNormalizeSurface( SC_Surface * p, float Time, float Load )
func AbcSclLibNormalizeSurface(p *SCSurface, Time float32, Load float32) {
	panic("TODO, stubbed")
}

// Abc_SclLibNormalize
// C: void Abc_SclLibNormalize( SC_Lib * p )
func (p *SCLib) AbcSclLibNormalize() {
	panic("TODO, stubbed")
}

// Abc_SclProduceGenlibStrSimple
// C: Vec_Str_t * Abc_SclProduceGenlibStrSimple( SC_Lib * p )
func (p *SCLib) AbcSclProduceGenlibStrSimple() *ByteVector {
	panic("TODO, stubbed")
}

// Abc_SclDeriveGenlibSimple (pScl is a void * that holds an SC_Lib *)
// C: Mio_Library_t * Abc_SclDeriveGenlibSimple( void * pScl )
func AbcSclDeriveGenlibSimple(pScl *SCLib) *MioLibrary {
	panic("TODO, stubbed")
}

// Abc_SclProduceGenlibStr (pnCellCount is an out parameter)
// C: Vec_Str_t * Abc_SclProduceGenlibStr( SC_Lib * p, float Slew, float Gain, int nGatesMin, int * pnCellCount )
func (p *SCLib) AbcSclProduceGenlibStr(Slew float32, Gain float32, nGatesMin int, pnCellCount *int) *ByteVector {
	panic("TODO, stubbed")
}

// Abc_SclProduceGenlibStrProfile (pnCellCount is an out parameter)
// C: Vec_Str_t * Abc_SclProduceGenlibStrProfile( SC_Lib * p, Mio_Library_t * pLib, float Slew, float Gain, int nGatesMin, int * pnCellCount )
func (p *SCLib) AbcSclProduceGenlibStrProfile(pLib *MioLibrary, Slew float32, Gain float32, nGatesMin int, pnCellCount *int) *ByteVector {
	panic("TODO, stubbed")
}

// Abc_SclDumpGenlib
// C: void Abc_SclDumpGenlib( char * pFileName, SC_Lib * p, float SlewInit, float Gain, int nGatesMin )
func AbcSclDumpGenlib(pFileName string, p *SCLib, SlewInit float32, Gain float32, nGatesMin int) {
	panic("TODO, stubbed")
}

// Abc_SclDeriveGenlib (pScl holds an SC_Lib * and pMio a Mio_Library_t * (both void *))
// C: Mio_Library_t * Abc_SclDeriveGenlib( void * pScl, void * pMio, float SlewInit, float Gain, int nGatesMin, int fVerbose )
func AbcSclDeriveGenlib(pScl *SCLib, pMio *MioLibrary, SlewInit float32, Gain float32, nGatesMin int, fVerbose int) *MioLibrary {
	panic("TODO, stubbed")
}

// Abc_SclInstallGenlib (pScl is a void * that holds an SC_Lib *)
// C: void Abc_SclInstallGenlib( void * pScl, float SlewInit, float Gain, int nGatesMin )
func AbcSclInstallGenlib(pScl *SCLib, SlewInit float32, Gain float32, nGatesMin int) {
	panic("TODO, stubbed")
}
