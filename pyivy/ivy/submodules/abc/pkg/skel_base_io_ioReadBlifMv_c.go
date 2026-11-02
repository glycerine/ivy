package abc

import "strings"

// Hand-checked skeleton of src/base/io/ioReadBlifMv.c.

// IoMvVar is Io_MvVar_t.
type IoMvVar struct {
	nValues int      // the number of values
	pNames  []string // char ** : the value names
}

// IoMvMod is Io_MvMod_t. The line vectors hold char * lines.
type IoMvMod struct {
	// file lines
	pName          string                 // char * : .model line
	vInputs        *PointerVector[string] // .inputs lines
	vOutputs       *PointerVector[string] // .outputs lines
	vLatches       *PointerVector[string] // .latch lines
	vFlops         *PointerVector[string] // .flop lines
	vResets        *PointerVector[string] // .reset lines
	vNames         *PointerVector[string] // .names lines
	vSubckts       *PointerVector[string] // .subckt lines
	vShorts        *PointerVector[string] // .short lines
	vOnehots       *PointerVector[string] // .onehot lines
	vMvs           *PointerVector[string] // .mv lines
	vConstrs       *PointerVector[string] // .constraint lines
	vLtlProperties *PointerVector[string]
	fBlackBox      int // indicates blackbox model
	// the resulting network
	pNtk        *ABCNetwork
	pResetLatch *ABCNetworkObject
	// the parent manager
	pMan *IoMvMan
}

// IoMvMan is Io_MvMan_t. pBuffer is a mutable []byte because the parser
// splits it into lines in place.
type IoMvMan struct {
	// general info about file
	fBlifMv   int                    // the file is BLIF-MV
	fUseReset int                    // the reset circuitry is added
	pFileName string                 // char * : the name of the file
	pBuffer   []byte                 // char * : the contents of the file
	vLines    *PointerVector[string] // the line beginnings
	// the results of reading
	pDesign  *AbcDes // the design under construction
	nNDnodes int     // the counter of ND nodes
	// intermediate storage for models
	vModels *PointerVector[IoMvMod] // vector of models
	pLatest *IoMvMod                // the current model
	// current processing info
	vTokens  *PointerVector[string] // the current tokens
	vTokens2 *PointerVector[string] // the current tokens
	vFunc    *ByteVector            // Vec_Str_t * : the local function
	// error reporting
	sError strings.Builder // char sError[512] : the error string generated during parsing
	// statistics
	nTablesRead int // the number of processed tables
	nTablesLeft int // the number of dangling tables
}

// Buflist is the `buflist` struct typedef (a chain of 1 MiB read buffers
// used by Io_MvLoadFileBz2).
type Buflist struct {
	buf  [1 << 20]byte // char buf[1<<20]
	nBuf int
	next *Buflist
}

// Io_MvCharIsSpace (static)
// C: int Io_MvCharIsSpace( char s )
func ioMvCharIsSpace(s int8) int {
	panic("TODO, stubbed")
}

// Io_MvCharIsMvSymb (static)
// C: int Io_MvCharIsMvSymb( char s )
func ioMvCharIsMvSymb(s int8) int {
	panic("TODO, stubbed")
}

// Io_ReadBlifMv
// C: Abc_Ntk_t * Io_ReadBlifMv( char * pFileName, int fBlifMv, int fCheck )
func IoReadBlifMv(pFileName string, fBlifMv int, fCheck int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Io_MvAlloc (static)
// C: Io_MvMan_t * Io_MvAlloc()
func ioMvAlloc() *IoMvMan {
	panic("TODO, stubbed")
}

// Io_MvFree (static)
// C: void Io_MvFree( Io_MvMan_t * p )
func (p *IoMvMan) ioMvFree() {
	panic("TODO, stubbed")
}

// Io_MvModAlloc (static)
// C: Io_MvMod_t * Io_MvModAlloc()
func ioMvModAlloc() *IoMvMod {
	panic("TODO, stubbed")
}

// Io_MvModFree (static)
// C: void Io_MvModFree( Io_MvMod_t * p )
func (p *IoMvMod) ioMvModFree() {
	panic("TODO, stubbed")
}

// Io_MvCountChars (static)
// C: int Io_MvCountChars( char * pLine, char Char )
func ioMvCountChars(pLine string, Char int8) int {
	panic("TODO, stubbed")
}

// Io_MvFindArrow (static; returns the suffix of pLine starting at the arrow)
// C: char * Io_MvFindArrow( char * pLine )
func ioMvFindArrow(pLine string) string {
	panic("TODO, stubbed")
}

// Io_MvCollectTokens (static)
// C: void Io_MvCollectTokens( Vec_Ptr_t * vTokens, char * pInput, char * pOutput )
func ioMvCollectTokens(vTokens *PointerVector[string], pInput string, pOutput string) {
	panic("TODO, stubbed")
}

// Io_MvSplitIntoTokens (static)
// C: void Io_MvSplitIntoTokens( Vec_Ptr_t * vTokens, char * pLine, char Stop )
func ioMvSplitIntoTokens(vTokens *PointerVector[string], pLine string, Stop int8) {
	panic("TODO, stubbed")
}

// Io_MvSplitIntoTokensMv (static)
// C: void Io_MvSplitIntoTokensMv( Vec_Ptr_t * vTokens, char * pLine )
func ioMvSplitIntoTokensMv(vTokens *PointerVector[string], pLine string) {
	panic("TODO, stubbed")
}

// Io_MvSplitIntoTokensAndClear (static)
// C: void Io_MvSplitIntoTokensAndClear( Vec_Ptr_t * vTokens, char * pLine, char Stop, char Char )
func ioMvSplitIntoTokensAndClear(vTokens *PointerVector[string], pLine string, Stop int8, Char int8) {
	panic("TODO, stubbed")
}

// Io_MvGetLine (static)
// C: int Io_MvGetLine( Io_MvMan_t * p, char * pToken )
func (p *IoMvMan) ioMvGetLine(pToken string) int {
	panic("TODO, stubbed")
}

// Io_MvLoadFileBz2 (returns the file contents buffer; pnFileSize is an out parameter)
// C: char * Io_MvLoadFileBz2( char * pFileName, int * pnFileSize )
func IoMvLoadFileBz2(pFileName string, pnFileSize *int) []byte {
	panic("TODO, stubbed")
}

// Io_MvLoadFileGz (static; returns the file contents buffer; pnFileSize is an out parameter)
// C: char * Io_MvLoadFileGz( char * pFileName, int * pnFileSize )
func ioMvLoadFileGz(pFileName string, pnFileSize *int) []byte {
	panic("TODO, stubbed")
}

// Io_MvLoadFile (static; returns the file contents buffer)
// C: char * Io_MvLoadFile( char * pFileName )
func ioMvLoadFile(pFileName string) []byte {
	panic("TODO, stubbed")
}

// Io_MvReadPreparse (static)
// C: void Io_MvReadPreparse( Io_MvMan_t * p )
func (p *IoMvMan) ioMvReadPreparse() {
	panic("TODO, stubbed")
}

// Io_MvReadInterfaces (static)
// C: int Io_MvReadInterfaces( Io_MvMan_t * p )
func (p *IoMvMan) ioMvReadInterfaces() int {
	panic("TODO, stubbed")
}

// Io_MvParse (static)
// C: Abc_Des_t * Io_MvParse( Io_MvMan_t * p )
func (p *IoMvMan) ioMvParse() *AbcDes {
	panic("TODO, stubbed")
}

// Io_MvParseLineModel (static)
// C: int Io_MvParseLineModel( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineModel(pLine string) int {
	panic("TODO, stubbed")
}

// Io_MvParseLineInputs (static)
// C: int Io_MvParseLineInputs( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineInputs(pLine string) int {
	panic("TODO, stubbed")
}

// Io_MvParseLineOutputs (static)
// C: int Io_MvParseLineOutputs( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineOutputs(pLine string) int {
	panic("TODO, stubbed")
}

// Io_MvParseLineConstrs (static)
// C: int Io_MvParseLineConstrs( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineConstrs(pLine string) int {
	panic("TODO, stubbed")
}

// Io_MvParseLineLtlProperty (static)
// C: int Io_MvParseLineLtlProperty( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineLtlProperty(pLine string) int {
	panic("TODO, stubbed")
}

// Io_MvParseLineLatch (static)
// C: int Io_MvParseLineLatch( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineLatch(pLine string) int {
	panic("TODO, stubbed")
}

// Io_MvParseLineFlop (static)
// C: int Io_MvParseLineFlop( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineFlop(pLine string) int {
	panic("TODO, stubbed")
}

// Io_MvParseLineSubckt (static)
// C: int Io_MvParseLineSubckt( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineSubckt(pLine string) int {
	panic("TODO, stubbed")
}

// Io_MvParseLineOnehot (static)
// C: Vec_Int_t * Io_MvParseLineOnehot( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineOnehot(pLine string) *IntVector {
	panic("TODO, stubbed")
}

// Io_MvParseLineMv (static)
// C: int Io_MvParseLineMv( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineMv(pLine string) int {
	panic("TODO, stubbed")
}

// Io_MvWriteValues (static)
// C: int Io_MvWriteValues( Abc_Obj_t * pNode, Vec_Str_t * vFunc )
func ioMvWriteValues(pNode *ABCNetworkObject, vFunc *ByteVector) int {
	panic("TODO, stubbed")
}

// Io_MvParseLiteralMv (static)
// C: int Io_MvParseLiteralMv( Io_MvMod_t * p, Abc_Obj_t * pNode, char * pToken, Vec_Str_t * vFunc, int iLit )
func (p *IoMvMod) ioMvParseLiteralMv(pNode *ABCNetworkObject, pToken string, vFunc *ByteVector, iLit int) int {
	panic("TODO, stubbed")
}

// Io_MvParseTableMv (static)
// C: char * Io_MvParseTableMv( Io_MvMod_t * p, Abc_Obj_t * pNode, Vec_Ptr_t * vTokens2, int nInputs, int nOutputs, int iOut )
func (p *IoMvMod) ioMvParseTableMv(pNode *ABCNetworkObject, vTokens2 *PointerVector[string], nInputs int, nOutputs int, iOut int) string {
	panic("TODO, stubbed")
}

// Io_MvParseAddResetCircuit (static)
// C: Abc_Obj_t * Io_MvParseAddResetCircuit( Io_MvMod_t * p, char * pName )
func (p *IoMvMod) ioMvParseAddResetCircuit(pName string) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Io_MvParseLineNamesMvOne (static)
// C: int Io_MvParseLineNamesMvOne( Io_MvMod_t * p, Vec_Ptr_t * vTokens, Vec_Ptr_t * vTokens2, int nInputs, int nOutputs, int iOut, int fReset )
func (p *IoMvMod) ioMvParseLineNamesMvOne(vTokens *PointerVector[string], vTokens2 *PointerVector[string], nInputs int, nOutputs int, iOut int, fReset int) int {
	panic("TODO, stubbed")
}

// Io_MvParseLineNamesMv (static)
// C: int Io_MvParseLineNamesMv( Io_MvMod_t * p, char * pLine, int fReset )
func (p *IoMvMod) ioMvParseLineNamesMv(pLine string, fReset int) int {
	panic("TODO, stubbed")
}

// Io_MvParseTableBlif (static)
// C: char * Io_MvParseTableBlif( Io_MvMod_t * p, char * pTable, int nFanins )
func (p *IoMvMod) ioMvParseTableBlif(pTable string, nFanins int) string {
	panic("TODO, stubbed")
}

// Io_MvParseLineNamesBlif (static)
// C: int Io_MvParseLineNamesBlif( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineNamesBlif(pLine string) int {
	panic("TODO, stubbed")
}

// Io_MvParseLineShortBlif (static)
// C: int Io_MvParseLineShortBlif( Io_MvMod_t * p, char * pLine )
func (p *IoMvMod) ioMvParseLineShortBlif(pLine string) int {
	panic("TODO, stubbed")
}

// Abc_NtkMvVarDup (static)
// C: Io_MvVar_t * Abc_NtkMvVarDup( Abc_Ntk_t * pNtk, Io_MvVar_t * pVar )
func abcNtkMvVarDup(pNtk *ABCNetwork, pVar *IoMvVar) *IoMvVar {
	panic("TODO, stubbed")
}

// Io_ReadBlifCleanName (static)
// C: char * Io_ReadBlifCleanName( char * pName )
func ioReadBlifCleanName(pName string) string {
	panic("TODO, stubbed")
}

// Io_MvParseLineGateBlif (static)
// C: int Io_MvParseLineGateBlif( Io_MvMod_t * p, Vec_Ptr_t * vTokens )
func (p *IoMvMod) ioMvParseLineGateBlif(vTokens *PointerVector[string]) int {
	panic("TODO, stubbed")
}

// Abc_MapBoxSetPrevNext (static inline)
// C: void Abc_MapBoxSetPrevNext( Vec_Ptr_t * vDrivers, Vec_Int_t * vMapIn, Vec_Int_t * vMapOut, int Id )
func abcMapBoxSetPrevNext(vDrivers *PointerVector[ABCNetworkObject], vMapIn *IntVector, vMapOut *IntVector, Id int) {
	panic("TODO, stubbed")
}

// Abc_MapBox2Next (static inline)
// C: int Abc_MapBox2Next( Vec_Ptr_t * vDrivers, Vec_Int_t * vMapIn, Vec_Int_t * vMapOut, int Id )
func abcMapBox2Next(vDrivers *PointerVector[ABCNetworkObject], vMapIn *IntVector, vMapOut *IntVector, Id int) int {
	panic("TODO, stubbed")
}

// Abc_MapBox2Prev (static inline)
// C: int Abc_MapBox2Prev( Vec_Ptr_t * vDrivers, Vec_Int_t * vMapIn, Vec_Int_t * vMapOut, int Id )
func abcMapBox2Prev(vDrivers *PointerVector[ABCNetworkObject], vMapIn *IntVector, vMapOut *IntVector, Id int) int {
	panic("TODO, stubbed")
}
