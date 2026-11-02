package abc

// Hand-checked skeleton of src/base/wlc/wlcReadSmt.c.

// SmtLineType is the Smt_LineType_t enum.
type SmtLineType int

const (
	SMT_PRS_NONE SmtLineType = iota
	SMT_PRS_SET_OPTION
	SMT_PRS_SET_LOGIC
	SMT_PRS_SET_INFO
	SMT_PRS_DEFINE_FUN
	SMT_PRS_DECLARE_FUN
	SMT_PRS_ASSERT
	SMT_PRS_LET
	SMT_PRS_CHECK_SAT
	SMT_PRS_GET_VALUE
	SMT_PRS_EXIT
	SMT_PRS_END
)

// SmtPrs is Smt_Prs_t. pBuffer is mutable ([]byte) because comments are
// removed in place; pLimit and pCur are positions in it.
type SmtPrs struct {
	// input data
	pName   string           // char * : file name
	pBuffer []byte           // char * : file contents
	pLimit  []byte           // char * : end of file
	pCur    []byte           // char * : current position
	pStrs   *StringNameTable // Abc_Nam_t * : string manager
	// network structure
	vStack IntVector // Vec_Int_t (by value) : current node on each level
	//Vec_Wec_t    vDepth;      // objects on each level   (commented out in the C source)
	vObjs     IntLevelVector // Vec_Wec_t (by value) : objects
	NameCount int
	nDigits   int
	vTempFans IntVector // Vec_Int_t (by value)
	// error handling
	ErrorStr [1000]int8 // char ErrorStr[1000]
}

// Smt_GetTypeName (static)
// C: char * Smt_GetTypeName( Smt_LineType_t Type )
func smtGetTypeName(Type SmtLineType) string {
	panic("TODO, stubbed")
}

// Smt_AddTypes (static)
// C: void Smt_AddTypes( Abc_Nam_t * p )
func smtAddTypes(p *StringNameTable) {
	panic("TODO, stubbed")
}

// Smt_EntryIsName (static)
// C: int Smt_EntryIsName( int Fan )
func smtEntryIsName(Fan int) int {
	panic("TODO, stubbed")
}

// Smt_EntryIsType (static)
// C: int Smt_EntryIsType( int Fan, Smt_LineType_t Type )
func smtEntryIsType(Fan int, Type SmtLineType) int {
	panic("TODO, stubbed")
}

// Smt_EntryName (static)
// C: char * Smt_EntryName( Smt_Prs_t * p, int Fan )
func (p *SmtPrs) smtEntryName(Fan int) string {
	panic("TODO, stubbed")
}

// Smt_EntryNode (static)
// C: Vec_Int_t * Smt_EntryNode( Smt_Prs_t * p, int Fan )
func (p *SmtPrs) smtEntryNode(Fan int) *IntVector {
	panic("TODO, stubbed")
}

// Smt_VecEntryIsType (static)
// C: int Smt_VecEntryIsType( Vec_Int_t * vFans, int i, Smt_LineType_t Type )
func smtVecEntryIsType(vFans *IntVector, i int, Type SmtLineType) int {
	panic("TODO, stubbed")
}

// Smt_VecEntryName (static)
// C: char * Smt_VecEntryName( Smt_Prs_t * p, Vec_Int_t * vFans, int i )
func (p *SmtPrs) smtVecEntryName(vFans *IntVector, i int) string {
	panic("TODO, stubbed")
}

// Smt_VecEntryNode (static)
// C: Vec_Int_t * Smt_VecEntryNode( Smt_Prs_t * p, Vec_Int_t * vFans, int i )
func (p *SmtPrs) smtVecEntryNode(vFans *IntVector, i int) *IntVector {
	panic("TODO, stubbed")
}

// Smt_StrToType (static; pfSigned is an out parameter)
// C: int Smt_StrToType( char * pName, int * pfSigned )
func smtStrToType(pName string, pfSigned *int) int {
	panic("TODO, stubbed")
}

// Smt_PrsReadType (static; pfSigned, Value1 and Value2 are out parameters)
// C: int Smt_PrsReadType( Smt_Prs_t * p, int iSig, int * pfSigned, int * Value1, int * Value2 )
func (p *SmtPrs) smtPrsReadType(iSig int, pfSigned *int, Value1 *int, Value2 *int) int {
	panic("TODO, stubbed")
}

// Smt_StrType (static)
// C: int Smt_StrType( char * str )
func smtStrType(str string) int {
	panic("TODO, stubbed")
}

// Smt_PrsCreateNodeOld (static)
// C: int Smt_PrsCreateNodeOld( Wlc_Ntk_t * pNtk, int Type, int fSigned, int Range, Vec_Int_t * vFanins, char * pName )
func smtPrsCreateNodeOld(pNtk *WLCNetwork, Type int, fSigned int, Range int, vFanins *IntVector, pName string) int {
	panic("TODO, stubbed")
}

// Smt_PrsCreateNode (static)
// C: int Smt_PrsCreateNode( Wlc_Ntk_t * pNtk, int Type, int fSigned, int Range, Vec_Int_t * vFanins, char * pName )
func smtPrsCreateNode(pNtk *WLCNetwork, Type int, fSigned int, Range int, vFanins *IntVector, pName string) int {
	panic("TODO, stubbed")
}

// Smt_GetHexFromDecimalString (static)
// C: char * Smt_GetHexFromDecimalString(char * pStr)
func smtGetHexFromDecimalString(pStr string) string {
	panic("TODO, stubbed")
}

// Smt_PrsBuildConstant (static)
// C: int Smt_PrsBuildConstant( Wlc_Ntk_t * pNtk, char * pStr, int nBits, char * pName )
func smtPrsBuildConstant(pNtk *WLCNetwork, pStr string, nBits int, pName string) int {
	panic("TODO, stubbed")
}

// Smt_PrsBuildNode
// C: int Smt_PrsBuildNode( Wlc_Ntk_t * pNtk, Smt_Prs_t * p, int iNode, int RangeOut, char * pName )
func (pNtk *WLCNetwork) SmtPrsBuildNode(p *SmtPrs, iNode int, RangeOut int, pName string) int {
	panic("TODO, stubbed")
}

// Smt_PrsBuild
// C: Wlc_Ntk_t * Smt_PrsBuild( Smt_Prs_t * p )
func (p *SmtPrs) SmtPrsBuild() *WLCNetwork {
	panic("TODO, stubbed")
}

// Smt_PrsGenName
// C: char * Smt_PrsGenName( Smt_Prs_t * p )
func (p *SmtPrs) SmtPrsGenName() string {
	panic("TODO, stubbed")
}

// Smt_PrsBuild2_rec
// C: int Smt_PrsBuild2_rec( Wlc_Ntk_t * pNtk, Smt_Prs_t * p, int iNode, int iObjPrev, char * pName )
func (pNtk *WLCNetwork) SmtPrsBuild2_rec(p *SmtPrs, iNode int, iObjPrev int, pName string) int {
	panic("TODO, stubbed")
}

// Smt_PrsBuild2
// C: Wlc_Ntk_t * Smt_PrsBuild2( Smt_Prs_t * p )
func (p *SmtPrs) SmtPrsBuild2() *WLCNetwork {
	panic("TODO, stubbed")
}

// Smt_PrsErrorSet (static)
// C: int Smt_PrsErrorSet( Smt_Prs_t * p, char * pError, int Value )
func (p *SmtPrs) smtPrsErrorSet(pError string, Value int) int {
	panic("TODO, stubbed")
}

// Smt_PrsErrorPrint (static)
// C: int Smt_PrsErrorPrint( Smt_Prs_t * p )
func (p *SmtPrs) smtPrsErrorPrint() int {
	panic("TODO, stubbed")
}

// Smt_PrsLoadFile (static; ppLimit is an out parameter (the end of the returned buffer))
// C: char * Smt_PrsLoadFile( char * pFileName, char ** ppLimit )
func smtPrsLoadFile(pFileName string, ppLimit *[]byte) []byte {
	panic("TODO, stubbed")
}

// Smt_PrsRemoveComments (static; pLimit is the end of pBuffer)
// C: int Smt_PrsRemoveComments( char * pBuffer, char * pLimit )
func smtPrsRemoveComments(pBuffer []byte, pLimit []byte) int {
	panic("TODO, stubbed")
}

// Smt_PrsAlloc (static; pLimit is the end of pBuffer)
// C: Smt_Prs_t * Smt_PrsAlloc( char * pFileName, char * pBuffer, char * pLimit, int nObjs )
func smtPrsAlloc(pFileName string, pBuffer []byte, pLimit []byte, nObjs int) *SmtPrs {
	panic("TODO, stubbed")
}

// Smt_PrsFree (static)
// C: void Smt_PrsFree( Smt_Prs_t * p )
func (p *SmtPrs) smtPrsFree() {
	panic("TODO, stubbed")
}

// Smt_PrsIsSpace (static)
// C: int Smt_PrsIsSpace( char c )
func smtPrsIsSpace(c int8) int {
	panic("TODO, stubbed")
}

// Smt_PrsSkipSpaces (static)
// C: void Smt_PrsSkipSpaces( Smt_Prs_t * p )
func (p *SmtPrs) smtPrsSkipSpaces() {
	panic("TODO, stubbed")
}

// Smt_PrsSkipNonSpaces (static)
// C: void Smt_PrsSkipNonSpaces( Smt_Prs_t * p )
func (p *SmtPrs) smtPrsSkipNonSpaces() {
	panic("TODO, stubbed")
}

// Smt_PrsReadLines
// C: void Smt_PrsReadLines( Smt_Prs_t * p )
func (p *SmtPrs) SmtPrsReadLines() {
	panic("TODO, stubbed")
}

// Smt_PrsPrintParser_rec
// C: void Smt_PrsPrintParser_rec( Smt_Prs_t * p, int iObj, int Depth )
func (p *SmtPrs) SmtPrsPrintParser_rec(iObj int, Depth int) {
	panic("TODO, stubbed")
}

// Smt_PrsPrintParser
// C: void Smt_PrsPrintParser( Smt_Prs_t * p )
func (p *SmtPrs) SmtPrsPrintParser() {
	panic("TODO, stubbed")
}

// Wlc_ReadSmtBuffer (pLimit is the end of pBuffer)
// C: Wlc_Ntk_t * Wlc_ReadSmtBuffer( char * pFileName, char * pBuffer, char * pLimit, int fOldParser, int fPrintTree )
func WlcReadSmtBuffer(pFileName string, pBuffer []byte, pLimit []byte, fOldParser int, fPrintTree int) *WLCNetwork {
	panic("TODO, stubbed")
}

// Wlc_ReadSmt
// C: Wlc_Ntk_t * Wlc_ReadSmt( char * pFileName, int fOldParser, int fPrintTree )
func WlcReadSmt(pFileName string, fOldParser int, fPrintTree int) *WLCNetwork {
	panic("TODO, stubbed")
}
