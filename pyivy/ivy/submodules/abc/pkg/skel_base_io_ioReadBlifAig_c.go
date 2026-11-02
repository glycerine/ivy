package abc

import "strings"

// Hand-checked skeleton of src/base/io/ioReadBlifAig.c.

// IoBlifObj is Io_BlifObj_t. Go has no bitfields; each is a uint32 whose
// width is in the comment.
type IoBlifObj struct {
	fPi    uint32            // unsigned : 1, the object is a primary input
	fPo    uint32            // unsigned : 1, the object is a primary output
	fLi    uint32            // unsigned : 1, the object is a latch input
	fLo    uint32            // unsigned : 1, the object is a latch output
	fDef   uint32            // unsigned : 1, the object is defined as a table (node, PO, LI)
	fLoop  uint32            // unsigned : 1, flag for loop detection
	Init   uint32            // unsigned : 2, the latch initial state
	Offset uint32            // unsigned : 24, temporary number
	pName  string            // char * : the name of this object
	pEquiv *ABCNetworkObject // void * : the AIG node representing this line
	pNext  *IoBlifObj        // the next obj in the hash table
}

// IoBlifMan is Io_BlifMan_t. pBuffer is a mutable []byte because the parser
// splits the buffer into lines in place; line and token vectors hold char *.
type IoBlifMan struct {
	// general info about file
	pFileName string                 // char * : the name of the file
	pBuffer   []byte                 // char * : the begining of the file buffer
	vLines    *PointerVector[string] // the line beginnings
	// temporary objects
	pObjects *IoBlifObj // the storage for objects (an array)
	nObjects int        // the number of objects allocated
	iObjNext int        // the next free object
	// file lines
	pModel   string                 // char * : .model line
	vInputs  *PointerVector[string] // .inputs lines
	vOutputs *PointerVector[string] // .outputs lines
	vLatches *PointerVector[string] // .latches lines
	vNames   *PointerVector[string] // .names lines
	// network objects
	vPis *PointerVector[IoBlifObj] // the PI structures
	vPos *PointerVector[IoBlifObj] // the PO structures
	vLis *PointerVector[IoBlifObj] // the LI structures
	vLos *PointerVector[IoBlifObj] // the LO structures
	// mapping of names into objects
	pTable     []*IoBlifObj // Io_BlifObj_t ** : the hash table
	nTableSize int          // the hash table size
	// current processing info
	pAig    *ABCNetwork            // the network under construction
	vTokens *PointerVector[string] // the current tokens
	sError  strings.Builder        // char sError[512] : the error string generated during parsing
	// statistics
	nTablesRead int // the number of processed tables
	nTablesLeft int // the number of dangling tables
}

// Io_BlifCharIsSpace (static)
// C: int Io_BlifCharIsSpace( char s )
func ioBlifCharIsSpace(s int8) int {
	panic("TODO, stubbed")
}

// Io_ReadBlifAsAig
// C: Abc_Ntk_t * Io_ReadBlifAsAig( char * pFileName, int fCheck )
func IoReadBlifAsAig(pFileName string, fCheck int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Io_BlifAlloc (static)
// C: Io_BlifMan_t * Io_BlifAlloc()
func ioBlifAlloc() *IoBlifMan {
	panic("TODO, stubbed")
}

// Io_BlifFree (static)
// C: void Io_BlifFree( Io_BlifMan_t * p )
func (p *IoBlifMan) ioBlifFree() {
	panic("TODO, stubbed")
}

// Io_BlifHashString (static)
// C: unsigned Io_BlifHashString( char * pName, int TableSize )
func ioBlifHashString(pName string, TableSize int) uint32 {
	panic("TODO, stubbed")
}

// Io_BlifHashLookup (static; returns the address of the table slot)
// C: Io_BlifObj_t ** Io_BlifHashLookup( Io_BlifMan_t * p, char * pName )
func (p *IoBlifMan) ioBlifHashLookup(pName string) **IoBlifObj {
	panic("TODO, stubbed")
}

// Io_BlifHashFindOrAdd (static)
// C: Io_BlifObj_t * Io_BlifHashFindOrAdd( Io_BlifMan_t * p, char * pName )
func (p *IoBlifMan) ioBlifHashFindOrAdd(pName string) *IoBlifObj {
	panic("TODO, stubbed")
}

// Io_BlifCollectTokens (static)
// C: void Io_BlifCollectTokens( Vec_Ptr_t * vTokens, char * pInput, char * pOutput )
func ioBlifCollectTokens(vTokens *PointerVector[string], pInput string, pOutput string) {
	panic("TODO, stubbed")
}

// Io_BlifSplitIntoTokens (static)
// C: void Io_BlifSplitIntoTokens( Vec_Ptr_t * vTokens, char * pLine, char Stop )
func ioBlifSplitIntoTokens(vTokens *PointerVector[string], pLine string, Stop int8) {
	panic("TODO, stubbed")
}

// Io_BlifGetLine (static)
// C: int Io_BlifGetLine( Io_BlifMan_t * p, char * pToken )
func (p *IoBlifMan) ioBlifGetLine(pToken string) int {
	panic("TODO, stubbed")
}

// Io_BlifEstimatePiNum (static)
// C: int Io_BlifEstimatePiNum( Io_BlifMan_t * p )
func (p *IoBlifMan) ioBlifEstimatePiNum() int {
	panic("TODO, stubbed")
}

// Io_BlifEstimateAndNum (static)
// C: int Io_BlifEstimateAndNum( Io_BlifMan_t * p )
func (p *IoBlifMan) ioBlifEstimateAndNum() int {
	panic("TODO, stubbed")
}

// Io_BlifLoadFile (static; returns the file contents buffer)
// C: char * Io_BlifLoadFile( char * pFileName )
func ioBlifLoadFile(pFileName string) []byte {
	panic("TODO, stubbed")
}

// Io_BlifReadPreparse (static)
// C: void Io_BlifReadPreparse( Io_BlifMan_t * p )
func (p *IoBlifMan) ioBlifReadPreparse() {
	panic("TODO, stubbed")
}

// Io_BlifParse (static)
// C: Abc_Ntk_t * Io_BlifParse( Io_BlifMan_t * p )
func (p *IoBlifMan) ioBlifParse() *ABCNetwork {
	panic("TODO, stubbed")
}

// Io_BlifParseModel (static)
// C: int Io_BlifParseModel( Io_BlifMan_t * p, char * pLine )
func (p *IoBlifMan) ioBlifParseModel(pLine string) int {
	panic("TODO, stubbed")
}

// Io_BlifParseInputs (static)
// C: int Io_BlifParseInputs( Io_BlifMan_t * p, char * pLine )
func (p *IoBlifMan) ioBlifParseInputs(pLine string) int {
	panic("TODO, stubbed")
}

// Io_BlifParseOutputs (static)
// C: int Io_BlifParseOutputs( Io_BlifMan_t * p, char * pLine )
func (p *IoBlifMan) ioBlifParseOutputs(pLine string) int {
	panic("TODO, stubbed")
}

// Io_BlifParseLatch (static)
// C: int Io_BlifParseLatch( Io_BlifMan_t * p, char * pLine )
func (p *IoBlifMan) ioBlifParseLatch(pLine string) int {
	panic("TODO, stubbed")
}

// Io_BlifParseNames (static)
// C: int Io_BlifParseNames( Io_BlifMan_t * p, char * pLine )
func (p *IoBlifMan) ioBlifParseNames(pLine string) int {
	panic("TODO, stubbed")
}

// Io_BlifParseTable (static)
// C: Abc_Obj_t * Io_BlifParseTable( Io_BlifMan_t * p, char * pTable, Vec_Ptr_t * vFanins )
func (p *IoBlifMan) ioBlifParseTable(pTable string, vFanins *PointerVector[ABCNetworkObject]) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Io_BlifParseConstruct_rec (static)
// C: Abc_Obj_t * Io_BlifParseConstruct_rec( Io_BlifMan_t * p, char * pName )
func (p *IoBlifMan) ioBlifParseConstruct_rec(pName string) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Io_BlifParseConstruct (static)
// C: int Io_BlifParseConstruct( Io_BlifMan_t * p )
func (p *IoBlifMan) ioBlifParseConstruct() int {
	panic("TODO, stubbed")
}
