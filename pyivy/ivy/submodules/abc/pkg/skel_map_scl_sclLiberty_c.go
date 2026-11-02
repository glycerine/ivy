package abc

import "os"

// Hand-checked skeleton of src/map/scl/sclLiberty.c.

// SclLibertyType is the Scl_LibertyType_t enum: entry types.
type SclLibertyType int

const (
	SCL_LIBERTY_NONE SclLibertyType = 0 // 0:  unknown
	SCL_LIBERTY_PROC                    // 1:  procedure :  key(head){body}
	SCL_LIBERTY_EQUA                    // 2:  equation  :  key:head;
	SCL_LIBERTY_LIST                    // 3:  list      :  key(head)
)

// SclPair is Scl_Pair_t.
type SclPair struct {
	Beg int // item beginning
	End int // item end
}

// SclItem is Scl_Item_t.
type SclItem struct {
	Type  int     // Scl_LibertyType_t
	iLine int     // file line where the item's spec begins
	Key   SclPair // key part
	Head  SclPair // head part
	Body  SclPair // body part
	Next  int     // next item in the list
	Child int     // first child item
}

// SclTree is Scl_Tree_t.
type SclTree struct {
	pFileName   string      // char * : input Liberty file name
	pContents   []byte      // char * : file contents
	nContents   int         // file size
	nLines      int         // line counter
	nItems      int         // number of items
	nItermAlloc int         // number of items allocated
	pItems      []SclItem   // Scl_Item_t * : the items
	pError      string      // char * : the error string
	clkStart    int64       // abctime : beginning time
	vBuffer     *ByteVector // temp string buffer
}

// Scl_LibertyRoot (static)
// C: Scl_Item_t * Scl_LibertyRoot( Scl_Tree_t * p )
func (p *SclTree) sclLibertyRoot() *SclItem {
	panic("TODO, stubbed")
}

// Scl_LibertyItem (static)
// C: Scl_Item_t * Scl_LibertyItem( Scl_Tree_t * p, int v )
func (p *SclTree) sclLibertyItem(v int) *SclItem {
	panic("TODO, stubbed")
}

// Scl_LibertyCompare (static)
// C: int Scl_LibertyCompare( Scl_Tree_t * p, Scl_Pair_t Pair, char * pStr )
func (p *SclTree) sclLibertyCompare(Pair SclPair, pStr string) int {
	panic("TODO, stubbed")
}

// Scl_PrintWord (static)
// C: void Scl_PrintWord( FILE * pFile, Scl_Tree_t * p, Scl_Pair_t Pair )
func sclPrintWord(pFile *os.File, p *SclTree, Pair SclPair) {
	panic("TODO, stubbed")
}

// Scl_PrintSpace (static)
// C: void Scl_PrintSpace( FILE * pFile, int nOffset )
func sclPrintSpace(pFile *os.File, nOffset int) {
	panic("TODO, stubbed")
}

// Scl_LibertyItemId (static)
// C: int Scl_LibertyItemId( Scl_Tree_t * p, Scl_Item_t * pItem )
func (p *SclTree) sclLibertyItemId(pItem *SclItem) int {
	panic("TODO, stubbed")
}

// Scl_LibertyParseDumpItem
// C: void Scl_LibertyParseDumpItem( FILE * pFile, Scl_Tree_t * p, Scl_Item_t * pItem, int nOffset )
func SclLibertyParseDumpItem(pFile *os.File, p *SclTree, pItem *SclItem, nOffset int) {
	panic("TODO, stubbed")
}

// Scl_LibertyParseDump
// C: int Scl_LibertyParseDump( Scl_Tree_t * p, char * pFileName )
func (p *SclTree) SclLibertyParseDump(pFileName string) int {
	panic("TODO, stubbed")
}

// Scl_LibertyCountItems (pBeg and pEnd bound a region of one buffer)
// C: int Scl_LibertyCountItems( char * pBeg, char * pEnd )
func SclLibertyCountItems(pBeg []byte, pEnd []byte) int {
	panic("TODO, stubbed")
}

// Scl_LibertyWipeOutComments (pBeg and pEnd bound a region of one buffer; comments are blanked in place)
// C: void Scl_LibertyWipeOutComments( char * pBeg, char * pEnd )
func SclLibertyWipeOutComments(pBeg []byte, pEnd []byte) {
	panic("TODO, stubbed")
}

// Scl_LibertyCharIsSpace (static)
// C: int Scl_LibertyCharIsSpace( char c )
func sclLibertyCharIsSpace(c int8) int {
	panic("TODO, stubbed")
}

// Scl_LibertySkipSpaces (static; ppPos is a cursor the function advances)
// C: int Scl_LibertySkipSpaces( Scl_Tree_t * p, char ** ppPos, char * pEnd, int fStopAtNewLine )
func (p *SclTree) sclLibertySkipSpaces(ppPos *[]byte, pEnd []byte, fStopAtNewLine int) int {
	panic("TODO, stubbed")
}

// Scl_LibertySkipEntry (static; ppPos is a cursor the function advances)
// C: int Scl_LibertySkipEntry( char ** ppPos, char * pEnd )
func sclLibertySkipEntry(ppPos *[]byte, pEnd []byte) int {
	panic("TODO, stubbed")
}

// Scl_LibertyFindMatch (static; returns the matching position)
// C: char * Scl_LibertyFindMatch( char * pPos, char * pEnd )
func sclLibertyFindMatch(pPos []byte, pEnd []byte) []byte {
	panic("TODO, stubbed")
}

// Scl_LibertyUpdateHead (static)
// C: Scl_Pair_t Scl_LibertyUpdateHead( Scl_Tree_t * p, Scl_Pair_t Head )
func (p *SclTree) sclLibertyUpdateHead(Head SclPair) SclPair {
	panic("TODO, stubbed")
}

// Scl_LibertyNewItem (static)
// C: Scl_Item_t * Scl_LibertyNewItem( Scl_Tree_t * p, int Type )
func (p *SclTree) sclLibertyNewItem(Type int) *SclItem {
	panic("TODO, stubbed")
}

// Scl_LibertyReadString
// C: char * Scl_LibertyReadString( Scl_Tree_t * p, Scl_Pair_t Pair )
func (p *SclTree) SclLibertyReadString(Pair SclPair) string {
	panic("TODO, stubbed")
}

// Scl_LibertyItemNum
// C: int Scl_LibertyItemNum( Scl_Tree_t * p, Scl_Item_t * pRoot, char * pName )
func (p *SclTree) SclLibertyItemNum(pRoot *SclItem, pName string) int {
	panic("TODO, stubbed")
}

// Scl_LibertyBuildItem (ppPos is a cursor the function advances)
// C: int Scl_LibertyBuildItem( Scl_Tree_t * p, char ** ppPos, char * pEnd )
func (p *SclTree) SclLibertyBuildItem(ppPos *[]byte, pEnd []byte) int {
	panic("TODO, stubbed")
}

// Scl_LibertyFixFileName (the file name is edited in place)
// C: void Scl_LibertyFixFileName( char * pFileName )
func SclLibertyFixFileName(pFileName []byte) {
	panic("TODO, stubbed")
}

// Scl_LibertyFileSize
// C: int Scl_LibertyFileSize( char * pFileName )
func SclLibertyFileSize(pFileName string) int {
	panic("TODO, stubbed")
}

// Scl_LibertyFileContents
// C: char * Scl_LibertyFileContents( char * pFileName, int nContents )
func SclLibertyFileContents(pFileName string, nContents int) []byte {
	panic("TODO, stubbed")
}

// Scl_LibertyStringDump
// C: void Scl_LibertyStringDump( char * pFileName, Vec_Str_t * vStr )
func SclLibertyStringDump(pFileName string, vStr *ByteVector) {
	panic("TODO, stubbed")
}

// Scl_LibertyStart
// C: Scl_Tree_t * Scl_LibertyStart( char * pFileName )
func SclLibertyStart(pFileName string) *SclTree {
	panic("TODO, stubbed")
}

// Scl_LibertyStop
// C: void Scl_LibertyStop( Scl_Tree_t * p, int fVerbose )
func (p *SclTree) SclLibertyStop(fVerbose int) {
	panic("TODO, stubbed")
}

// Scl_LibertyParse
// C: Scl_Tree_t * Scl_LibertyParse( char * pFileName, int fVerbose )
func SclLibertyParse(pFileName string, fVerbose int) *SclTree {
	panic("TODO, stubbed")
}

// Scl_LibertyReadCellIsFlop
// C: int Scl_LibertyReadCellIsFlop( Scl_Tree_t * p, Scl_Item_t * pCell )
func (p *SclTree) SclLibertyReadCellIsFlop(pCell *SclItem) int {
	panic("TODO, stubbed")
}

// Scl_LibertyReadCellIsDontUse
// C: int Scl_LibertyReadCellIsDontUse( Scl_Tree_t * p, Scl_Item_t * pCell )
func (p *SclTree) SclLibertyReadCellIsDontUse(pCell *SclItem) int {
	panic("TODO, stubbed")
}

// Scl_LibertyReadCellArea
// C: char * Scl_LibertyReadCellArea( Scl_Tree_t * p, Scl_Item_t * pCell )
func (p *SclTree) SclLibertyReadCellArea(pCell *SclItem) string {
	panic("TODO, stubbed")
}

// Scl_LibertyReadCellLeakage
// C: char * Scl_LibertyReadCellLeakage( Scl_Tree_t * p, Scl_Item_t * pCell )
func (p *SclTree) SclLibertyReadCellLeakage(pCell *SclItem) string {
	panic("TODO, stubbed")
}

// Scl_LibertyReadPinFormula
// C: char * Scl_LibertyReadPinFormula( Scl_Tree_t * p, Scl_Item_t * pPin )
func (p *SclTree) SclLibertyReadPinFormula(pPin *SclItem) string {
	panic("TODO, stubbed")
}

// Scl_LibertyReadCellIsThreeState
// C: int Scl_LibertyReadCellIsThreeState( Scl_Tree_t * p, Scl_Item_t * pCell )
func (p *SclTree) SclLibertyReadCellIsThreeState(pCell *SclItem) int {
	panic("TODO, stubbed")
}

// Scl_LibertyReadCellOutputNum
// C: int Scl_LibertyReadCellOutputNum( Scl_Tree_t * p, Scl_Item_t * pCell )
func (p *SclTree) SclLibertyReadCellOutputNum(pCell *SclItem) int {
	panic("TODO, stubbed")
}

// Scl_LibertyReadGenlibStr
// C: Vec_Str_t * Scl_LibertyReadGenlibStr( Scl_Tree_t * p, int fVerbose )
func (p *SclTree) SclLibertyReadGenlibStr(fVerbose int) *ByteVector {
	panic("TODO, stubbed")
}

// Scl_LibertyParseGenlibStr
// C: Vec_Str_t * Scl_LibertyParseGenlibStr( char * pFileName, int fVerbose )
func SclLibertyParseGenlibStr(pFileName string, fVerbose int) *ByteVector {
	panic("TODO, stubbed")
}

// Vec_StrPutI_ (static)
// C: void Vec_StrPutI_( Vec_Str_t * vOut, int Val )
func vecStrPutI_(vOut *ByteVector, Val int) {
	panic("TODO, stubbed")
}

// Vec_StrPutW_ (static)
// C: void Vec_StrPutW_( Vec_Str_t * vOut, word Val )
func vecStrPutW_(vOut *ByteVector, Val uint64) {
	panic("TODO, stubbed")
}

// Vec_StrPutF_ (static)
// C: void Vec_StrPutF_( Vec_Str_t * vOut, float Val )
func vecStrPutF_(vOut *ByteVector, Val float32) {
	panic("TODO, stubbed")
}

// Vec_StrPutS_ (static)
// C: void Vec_StrPutS_( Vec_Str_t * vOut, char * Val )
func vecStrPutS_(vOut *ByteVector, Val string) {
	panic("TODO, stubbed")
}

// Vec_StrPut_ (static)
// C: void Vec_StrPut_( Vec_Str_t * vOut )
func vecStrPut_(vOut *ByteVector) {
	panic("TODO, stubbed")
}

// Scl_LibertyReadDefaultWireLoad
// C: char * Scl_LibertyReadDefaultWireLoad( Scl_Tree_t * p )
func (p *SclTree) SclLibertyReadDefaultWireLoad() string {
	panic("TODO, stubbed")
}

// Scl_LibertyReadDefaultWireLoadSel
// C: char * Scl_LibertyReadDefaultWireLoadSel( Scl_Tree_t * p )
func (p *SclTree) SclLibertyReadDefaultWireLoadSel() string {
	panic("TODO, stubbed")
}

// Scl_LibertyReadDefaultMaxTrans
// C: float Scl_LibertyReadDefaultMaxTrans( Scl_Tree_t * p )
func (p *SclTree) SclLibertyReadDefaultMaxTrans() float32 {
	panic("TODO, stubbed")
}

// Scl_LibertyReadTimeUnit
// C: int Scl_LibertyReadTimeUnit( Scl_Tree_t * p )
func (p *SclTree) SclLibertyReadTimeUnit() int {
	panic("TODO, stubbed")
}

// Scl_LibertyReadLoadUnit
// C: void Scl_LibertyReadLoadUnit( Scl_Tree_t * p, Vec_Str_t * vOut )
func (p *SclTree) SclLibertyReadLoadUnit(vOut *ByteVector) {
	panic("TODO, stubbed")
}

// Scl_LibertyReadWireLoad
// C: void Scl_LibertyReadWireLoad( Scl_Tree_t * p, Vec_Str_t * vOut )
func (p *SclTree) SclLibertyReadWireLoad(vOut *ByteVector) {
	panic("TODO, stubbed")
}

// Scl_LibertyReadWireLoadSelect
// C: void Scl_LibertyReadWireLoadSelect( Scl_Tree_t * p, Vec_Str_t * vOut )
func (p *SclTree) SclLibertyReadWireLoadSelect(vOut *ByteVector) {
	panic("TODO, stubbed")
}

// Scl_LibertyReadDeriveStrength
// C: int Scl_LibertyReadDeriveStrength( Scl_Tree_t * p, Scl_Item_t * pCell )
func (p *SclTree) SclLibertyReadDeriveStrength(pCell *SclItem) int {
	panic("TODO, stubbed")
}

// Scl_LibertyReadPinDirection
// C: int Scl_LibertyReadPinDirection( Scl_Tree_t * p, Scl_Item_t * pPin )
func (p *SclTree) SclLibertyReadPinDirection(pPin *SclItem) int {
	panic("TODO, stubbed")
}

// Scl_LibertyReadPinCap
// C: float Scl_LibertyReadPinCap( Scl_Tree_t * p, Scl_Item_t * pPin, char * pName )
func (p *SclTree) SclLibertyReadPinCap(pPin *SclItem, pName string) float32 {
	panic("TODO, stubbed")
}

// Scl_LibertyReadPinTiming
// C: Scl_Item_t * Scl_LibertyReadPinTiming( Scl_Tree_t * p, Scl_Item_t * pPinOut, char * pNameIn )
func (p *SclTree) SclLibertyReadPinTiming(pPinOut *SclItem, pNameIn string) *SclItem {
	panic("TODO, stubbed")
}

// Scl_LibertyReadPinTimingAll
// C: Vec_Ptr_t * Scl_LibertyReadPinTimingAll( Scl_Tree_t * p, Scl_Item_t * pPinOut, char * pNameIn )
func (p *SclTree) SclLibertyReadPinTimingAll(pPinOut *SclItem, pNameIn string) *PointerVector[SclItem] {
	panic("TODO, stubbed")
}

// Scl_LibertyReadTimingSense
// C: int Scl_LibertyReadTimingSense( Scl_Tree_t * p, Scl_Item_t * pPin )
func (p *SclTree) SclLibertyReadTimingSense(pPin *SclItem) int {
	panic("TODO, stubbed")
}

// Scl_LibertyReadFloatVec
// C: Vec_Flt_t * Scl_LibertyReadFloatVec( char * pName )
func SclLibertyReadFloatVec(pName string) *FloatVector {
	panic("TODO, stubbed")
}

// Scl_LibertyDumpTables
// C: void Scl_LibertyDumpTables( Vec_Str_t * vOut, Vec_Flt_t * vInd1, Vec_Flt_t * vInd2, Vec_Flt_t * vValues )
func SclLibertyDumpTables(vOut *ByteVector, vInd1 *FloatVector, vInd2 *FloatVector, vValues *FloatVector) {
	panic("TODO, stubbed")
}

// Scl_LibertyScanTable (vOut receives Vec_Flt_t * and Vec_Int_t * entries and vTemples holds names and tables (untyped void * entries))
// C: int Scl_LibertyScanTable( Scl_Tree_t * p, Vec_Ptr_t * vOut, Scl_Item_t * pTiming, char * pName, Vec_Ptr_t * vTemples )
func (p *SclTree) SclLibertyScanTable(vOut *PointerVector[any], pTiming *SclItem, pName string, vTemples *PointerVector[any]) int {
	panic("TODO, stubbed")
}

// Scl_LibertyComputeWorstCase (pvInd0, pvInd1 and pvValues are out parameters)
// C: int Scl_LibertyComputeWorstCase( Vec_Ptr_t * vTables, Vec_Flt_t ** pvInd0, Vec_Flt_t ** pvInd1, Vec_Flt_t ** pvValues )
func SclLibertyComputeWorstCase(vTables *PointerVector[FloatVector], pvInd0 **FloatVector, pvInd1 **FloatVector, pvValues **FloatVector) int {
	panic("TODO, stubbed")
}

// Scl_LibertyReadTable (vTemples holds names and tables (untyped void * entries))
// C: int Scl_LibertyReadTable( Scl_Tree_t * p, Vec_Str_t * vOut, Scl_Item_t * pTiming, char * pName, Vec_Ptr_t * vTemples )
func (p *SclTree) SclLibertyReadTable(vOut *ByteVector, pTiming *SclItem, pName string, vTemples *PointerVector[any]) int {
	panic("TODO, stubbed")
}

// Scl_LibertyPrintTemplates (vRes holds names and tables (untyped void * entries))
// C: void Scl_LibertyPrintTemplates( Vec_Ptr_t * vRes )
func SclLibertyPrintTemplates(vRes *PointerVector[any]) {
	panic("TODO, stubbed")
}

// Scl_LibertyReadTemplates (the result holds names and tables (untyped void * entries))
// C: Vec_Ptr_t * Scl_LibertyReadTemplates( Scl_Tree_t * p )
func (p *SclTree) SclLibertyReadTemplates() *PointerVector[any] {
	panic("TODO, stubbed")
}

// Scl_LibertyReadSclStr
// C: Vec_Str_t * Scl_LibertyReadSclStr( Scl_Tree_t * p, int fVerbose, int fVeryVerbose )
func (p *SclTree) SclLibertyReadSclStr(fVerbose int, fVeryVerbose int) *ByteVector {
	panic("TODO, stubbed")
}

// Abc_SclReadLiberty
// C: SC_Lib * Abc_SclReadLiberty( char * pFileName, int fVerbose, int fVeryVerbose )
func AbcSclReadLiberty(pFileName string, fVerbose int, fVeryVerbose int) *SCLib {
	panic("TODO, stubbed")
}

// Scl_LibertyTest
// C: void Scl_LibertyTest()
func SclLibertyTest() {
	panic("TODO, stubbed")
}
