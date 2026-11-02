package abc

import "os"

// Hand-checked skeleton of src/aig/gia/giaGig.c.

// Gls_Type_t: gate types (the enum of giaGig.c).
const (
	GLS_NONE  = -1 // not used
	GLS_ZERO  = 0  // zero
	GLS_ONE   = 1  // one
	GLS_PI    = 2  // primary input
	GLS_PO    = 3  // primary output
	GLS_BAR   = 4  // barrier
	GLS_SEQ   = 5  // sequential
	GLS_SEL   = 6  // fan
	GLS_LUT4  = 7  // LUT4
	GLS_LUT6  = 8  // LUT6
	GLS_BOX   = 9  // sequential box
	GLS_DEL   = 10 // delay box
	GLS_FINAL = 11
)

// The static table s_Strs[GLS_FINAL] is not represented.

// GlsMan is Gls_Man_t.
type GlsMan struct {
	vLines       *ByteVector     // line types
	vTypes       *ByteVector     // gate types
	vIndexes     *IntVector      // gate indexes
	vLut4s       *IntVector      // 4-LUTs (4-tuples)
	vLut4TTs     *IntVector      // truth tables
	vLut6s       *IntVector      // 6-LUTs (6-tuples)
	vLut6TTs     *WordVector     // truth tables
	vBoxes       *IntVector      // boxes (5-tuples)
	vDelayIns    *IntLevelVector // delay fanins
	vDelayOuts   *IntLevelVector // delay fanouts
	vDelays      *IntVector      // delay values
	vOrderPis    *IntVector
	vOrderPos    *IntVector
	vOrderBoxes  *IntVector
	vOrderDelays *IntVector
	vOrderLuts   *IntVector
	vOrderSeqs   *IntVector
}

// Gls_ManAlloc
// C: Gls_Man_t * Gls_ManAlloc( Vec_Str_t * vLines, int * pCounts )
func GlsManAlloc(vLines *ByteVector, pCounts []int) *GlsMan {
	panic("TODO, stubbed")
}

// Gls_ManStop
// C: void Gls_ManStop( Gls_Man_t * p )
func (p *GlsMan) GlsManStop() {
	panic("TODO, stubbed")
}

// Gls_ManCount (pCounts has GLS_FINAL entries)
// C: Vec_Str_t * Gls_ManCount( FILE * pFile, int pCounts[GLS_FINAL] )
func GlsManCount(pFile *os.File, pCounts []int) *ByteVector {
	panic("TODO, stubbed")
}

// Gls_ManParseOne (ppLine is a char ** cursor; here the remaining string is re-sliced)
// C: int Gls_ManParseOne( char ** ppLine )
func GlsManParseOne(ppLine *string) int {
	panic("TODO, stubbed")
}

// Gls_ManParse
// C: int Gls_ManParse( FILE * pFile, Gls_Man_t * p )
func GlsManParse(pFile *os.File, p *GlsMan) int {
	panic("TODO, stubbed")
}

// Gls_ManConstruct
// C: Gia_Man_t * Gls_ManConstruct( Gls_Man_t * p, char * pFileName )
func (p *GlsMan) GlsManConstruct(pFileName string) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManReadGig
// C: Gia_Man_t * Gia_ManReadGig( char * pFileName )
func GiaManReadGig(pFileName string) *GiaManager {
	panic("TODO, stubbed")
}
