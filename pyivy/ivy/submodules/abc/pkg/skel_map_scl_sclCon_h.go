package abc

// Hand-checked skeleton of src/map/scl/sclCon.h.

// SclCon is Scl_Con_t.
type SclCon struct {
	pFileName   string                // char * : constraint file name
	pModelName  string                // char * : current design name
	pInCellDef  string                // char * : default input driving gate
	tInArrDef   int                   // default input arrival time
	tInSlewDef  int                   // default input slew
	tInLoadDef  int                   // default input load
	tOutReqDef  int                   // default output required time
	tOutLoadDef int                   // default output load
	vInCells    PointerVector[string] // Vec_Ptr_t (by value) : input driving gate names
	vInCellsPtr PointerVector[SCCell] // Vec_Ptr_t (by value) : input driving gates
	vInArrs     IntVector             // Vec_Int_t (by value) : input arrival times
	vInSlews    IntVector             // Vec_Int_t (by value) : input slews
	vInLoads    IntVector             // Vec_Int_t (by value) : input loads
	vOutReqs    IntVector             // Vec_Int_t (by value) : output required times
	vOutLoads   IntVector             // Vec_Int_t (by value) : output loads
	pNamI       *StringNameTable      // Abc_Nam_t * : input names
	pNamO       *StringNameTable      // Abc_Nam_t * : output names
}

// Scl_Flt2Int (static inline)
// C: int Scl_Flt2Int( float w )
func sclFlt2Int(w float32) int {
	panic("TODO, stubbed")
}

// Scl_Int2Flt (static inline)
// C: float Scl_Int2Flt( int i )
func sclInt2Flt(i int) float32 {
	panic("TODO, stubbed")
}

// Scl_ConAlloc (static inline)
// C: Scl_Con_t * Scl_ConAlloc( char * pFileName, Abc_Nam_t * pNamI, Abc_Nam_t * pNamO )
func sclConAlloc(pFileName string, pNamI *StringNameTable, pNamO *StringNameTable) *SclCon {
	panic("TODO, stubbed")
}

// Scl_ConFree (static inline)
// C: void Scl_ConFree( Scl_Con_t * p )
func (p *SclCon) sclConFree() {
	panic("TODO, stubbed")
}

// Scl_ConParse (static inline)
// C: int Scl_ConParse( Scl_Con_t * p, Abc_Nam_t * pNamI, Abc_Nam_t * pNamO )
func (p *SclCon) sclConParse(pNamI *StringNameTable, pNamO *StringNameTable) int {
	panic("TODO, stubbed")
}

// Scl_ConRead (static inline)
// C: Scl_Con_t * Scl_ConRead( char * pFileName, Abc_Nam_t * pNamI, Abc_Nam_t * pNamO )
func sclConRead(pFileName string, pNamI *StringNameTable, pNamO *StringNameTable) *SclCon {
	panic("TODO, stubbed")
}

// Scl_ConWrite (static inline)
// C: void Scl_ConWrite( Scl_Con_t * p, char * pFileName )
func (p *SclCon) sclConWrite(pFileName string) {
	panic("TODO, stubbed")
}

// Scl_ConHasInCells_ (static inline)
// C: int Scl_ConHasInCells_( Scl_Con_t * p )
func (p *SclCon) sclConHasInCells_() int {
	panic("TODO, stubbed")
}

// Scl_ConHasInArrs_ (static inline)
// C: int Scl_ConHasInArrs_( Scl_Con_t * p )
func (p *SclCon) sclConHasInArrs_() int {
	panic("TODO, stubbed")
}

// Scl_ConHasInSlews_ (static inline)
// C: int Scl_ConHasInSlews_( Scl_Con_t * p )
func (p *SclCon) sclConHasInSlews_() int {
	panic("TODO, stubbed")
}

// Scl_ConHasInLoads_ (static inline)
// C: int Scl_ConHasInLoads_( Scl_Con_t * p )
func (p *SclCon) sclConHasInLoads_() int {
	panic("TODO, stubbed")
}

// Scl_ConHasOutReqs_ (static inline)
// C: int Scl_ConHasOutReqs_( Scl_Con_t * p )
func (p *SclCon) sclConHasOutReqs_() int {
	panic("TODO, stubbed")
}

// Scl_ConHasOutLoads_ (static inline)
// C: int Scl_ConHasOutLoads_( Scl_Con_t * p )
func (p *SclCon) sclConHasOutLoads_() int {
	panic("TODO, stubbed")
}

// Scl_ConGetInCell_ (static inline)
// C: char * Scl_ConGetInCell_( Scl_Con_t * p, int i )
func (p *SclCon) sclConGetInCell_(i int) string {
	panic("TODO, stubbed")
}

// Scl_ConGetInArr_ (static inline)
// C: int Scl_ConGetInArr_( Scl_Con_t * p, int i )
func (p *SclCon) sclConGetInArr_(i int) int {
	panic("TODO, stubbed")
}

// Scl_ConGetInSlew_ (static inline)
// C: int Scl_ConGetInSlew_( Scl_Con_t * p, int i )
func (p *SclCon) sclConGetInSlew_(i int) int {
	panic("TODO, stubbed")
}

// Scl_ConGetInLoad_ (static inline)
// C: int Scl_ConGetInLoad_( Scl_Con_t * p, int i )
func (p *SclCon) sclConGetInLoad_(i int) int {
	panic("TODO, stubbed")
}

// Scl_ConGetOutReq_ (static inline)
// C: int Scl_ConGetOutReq_( Scl_Con_t * p, int i )
func (p *SclCon) sclConGetOutReq_(i int) int {
	panic("TODO, stubbed")
}

// Scl_ConGetOutLoad_ (static inline)
// C: int Scl_ConGetOutLoad_( Scl_Con_t * p, int i )
func (p *SclCon) sclConGetOutLoad_(i int) int {
	panic("TODO, stubbed")
}

// Scl_ConIsRunning (static inline)
// C: int Scl_ConIsRunning()
func sclConIsRunning() int {
	panic("TODO, stubbed")
}

// Scl_ConHasInCells (static inline)
// C: int Scl_ConHasInCells()
func sclConHasInCells() int {
	panic("TODO, stubbed")
}

// Scl_ConHasInArrs (static inline)
// C: int Scl_ConHasInArrs()
func sclConHasInArrs() int {
	panic("TODO, stubbed")
}

// Scl_ConHasInSlews (static inline)
// C: int Scl_ConHasInSlews()
func sclConHasInSlews() int {
	panic("TODO, stubbed")
}

// Scl_ConHasInLoads (static inline)
// C: int Scl_ConHasInLoads()
func sclConHasInLoads() int {
	panic("TODO, stubbed")
}

// Scl_ConHasOutReqs (static inline)
// C: int Scl_ConHasOutReqs()
func sclConHasOutReqs() int {
	panic("TODO, stubbed")
}

// Scl_ConHasOutLoads (static inline)
// C: int Scl_ConHasOutLoads()
func sclConHasOutLoads() int {
	panic("TODO, stubbed")
}

// Scl_ConGetInCell (static inline)
// C: char * Scl_ConGetInCell( int i )
func sclConGetInCell(i int) string {
	panic("TODO, stubbed")
}

// Scl_ConGetInArr (static inline)
// C: int Scl_ConGetInArr( int i )
func sclConGetInArr(i int) int {
	panic("TODO, stubbed")
}

// Scl_ConGetInSlew (static inline)
// C: int Scl_ConGetInSlew( int i )
func sclConGetInSlew(i int) int {
	panic("TODO, stubbed")
}

// Scl_ConGetInLoad (static inline)
// C: int Scl_ConGetInLoad( int i )
func sclConGetInLoad(i int) int {
	panic("TODO, stubbed")
}

// Scl_ConGetOutReq (static inline)
// C: int Scl_ConGetOutReq( int i )
func sclConGetOutReq(i int) int {
	panic("TODO, stubbed")
}

// Scl_ConGetOutLoad (static inline)
// C: int Scl_ConGetOutLoad( int i )
func sclConGetOutLoad(i int) int {
	panic("TODO, stubbed")
}

// Scl_ConGetInArrFloat (static inline)
// C: float Scl_ConGetInArrFloat( int i )
func sclConGetInArrFloat(i int) float32 {
	panic("TODO, stubbed")
}

// Scl_ConGetInSlewFloat (static inline)
// C: float Scl_ConGetInSlewFloat( int i )
func sclConGetInSlewFloat(i int) float32 {
	panic("TODO, stubbed")
}

// Scl_ConGetInLoadFloat (static inline)
// C: float Scl_ConGetInLoadFloat( int i )
func sclConGetInLoadFloat(i int) float32 {
	panic("TODO, stubbed")
}

// Scl_ConGetOutReqFloat (static inline)
// C: float Scl_ConGetOutReqFloat( int i )
func sclConGetOutReqFloat(i int) float32 {
	panic("TODO, stubbed")
}

// Scl_ConGetOutLoadFloat (static inline)
// C: float Scl_ConGetOutLoadFloat( int i )
func sclConGetOutLoadFloat(i int) float32 {
	panic("TODO, stubbed")
}
