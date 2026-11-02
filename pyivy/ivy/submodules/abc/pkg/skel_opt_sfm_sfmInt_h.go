package abc

// Hand-checked skeleton of src/opt/sfm/sfmInt.h.

// Hand-written skeleton of the types and static inline functions in
// src/opt/sfm/sfmInt.h. SFM_FANIN_MAX is 6, SFM_SUPP_MAX is 8, SFM_WORD_MAX is
// 4, SFM_WIN_MAX is 1000, SFM_DEC_MAX is 4 and SFM_SIM_WORDS is 8. The by-value
// Vec_* members are written as values.

// SfmNtk is Sfm_Ntk_t.
type SfmNtk struct {
	// parameters
	pPars *SfmPar // parameters
	// objects
	nPis      int // PI count (PIs should be first objects)
	nPos      int // PO count (POs should be last objects)
	nNodes    int // internal nodes
	nObjs     int // total objects
	nLevelMax int // maximum level
	// user data
	vFixed   *ByteVector    // persistent objects
	vEmpty   *ByteVector    // transparent objects
	vTruths  *WordVector    // truth tables
	vFanins  IntLevelVector // Vec_Wec_t (by value) : fanins
	vStarts  *IntVector     // offsets
	vTruths2 *WordVector    // truth tables (large)
	// attributes
	vFanouts IntLevelVector  // Vec_Wec_t (by value) : fanouts
	vLevels  IntVector       // Vec_Int_t (by value) : logic level
	vLevelsR IntVector       // Vec_Int_t (by value) : logic level
	vCounts  IntVector       // Vec_Int_t (by value) : fanin counters
	vId2Var  IntVector       // Vec_Int_t (by value) : ObjId -> SatVar
	vVar2Id  IntVector       // Vec_Int_t (by value) : SatVar -> ObjId
	vCnfs    *IntLevelVector // Vec_Wec_t * : CNFs
	vCover   *IntVector      // temporary
	// traversal IDs
	vTravIds  IntVector // Vec_Int_t (by value) : traversal IDs
	vTravIds2 IntVector // Vec_Int_t (by value) : traversal IDs
	nTravIds  int       // traversal IDs
	nTravIds2 int       // traversal IDs
	// window
	iPivotNode int        // window pivot
	vNodes     *IntVector // internal
	vDivs      *IntVector // divisors
	vRoots     *IntVector // roots
	vTfo       *IntVector // TFO (excluding iNode)
	// SAT solving
	pSat         *SatSolver // SAT solver
	nSatVars     int        // the number of variables
	nTryRemoves  int        // number of fanin removals
	nTryImproves int        // number of node improvements
	nTryResubs   int        // number of resubstitutions
	nRemoves     int        // number of fanin removals
	nImproves    int        // number of node improvements
	nResubs      int        // number of resubstitutions
	// counter-examples
	nCexes    int         // number of CEXes
	vDivCexes *WordVector // counter-examples
	// intermediate data
	vOrder    *IntVector      // object order
	vDivVars  *IntVector      // divisor SAT variables
	vDivIds   *IntVector      // divisors indexes
	vLits     *IntVector      // literals
	vValues   *IntVector      // SAT variable values
	vClauses  *IntLevelVector // CNF clauses for the node
	vFaninMap *IntVector      // mapping fanins into their SAT vars
	// nodes
	nTotalNodesBeg int
	nTotalEdgesBeg int
	nTotalNodesEnd int
	nTotalEdgesEnd int
	nNodesTried    int
	nTotalDivs     int
	nSatCalls      int
	nTimeOuts      int
	nMaxDivs       int
	// runtime
	timeWin   int64 // abctime
	timeDiv   int64
	timeCnf   int64
	timeSat   int64
	timeOther int64
	timeTotal int64
}

// Sfm_NtkPiNum (static inline)
// C: int Sfm_NtkPiNum( Sfm_Ntk_t * p )
func (p *SfmNtk) sfmNtkPiNum() int {
	panic("TODO, stubbed")
}

// Sfm_NtkPoNum (static inline)
// C: int Sfm_NtkPoNum( Sfm_Ntk_t * p )
func (p *SfmNtk) sfmNtkPoNum() int {
	panic("TODO, stubbed")
}

// Sfm_NtkNodeNum (static inline)
// C: int Sfm_NtkNodeNum( Sfm_Ntk_t * p )
func (p *SfmNtk) sfmNtkNodeNum() int {
	panic("TODO, stubbed")
}

// Sfm_ObjIsPi (static inline)
// C: int Sfm_ObjIsPi( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) sfmObjIsPi(i int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjIsPo (static inline)
// C: int Sfm_ObjIsPo( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) sfmObjIsPo(i int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjIsNode (static inline)
// C: int Sfm_ObjIsNode( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) sfmObjIsNode(i int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjIsFixed (static inline)
// C: int Sfm_ObjIsFixed( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) sfmObjIsFixed(i int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjAddsLevelArray (static inline)
// C: int Sfm_ObjAddsLevelArray( Vec_Str_t * p, int i )
func sfmObjAddsLevelArray(p *ByteVector, i int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjAddsLevel (static inline)
// C: int Sfm_ObjAddsLevel( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) sfmObjAddsLevel(i int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjFiArray (static inline)
// C: Vec_Int_t * Sfm_ObjFiArray( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) sfmObjFiArray(i int) *IntVector {
	panic("TODO, stubbed")
}

// Sfm_ObjFoArray (static inline)
// C: Vec_Int_t * Sfm_ObjFoArray( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) sfmObjFoArray(i int) *IntVector {
	panic("TODO, stubbed")
}

// Sfm_ObjFaninNum (static inline)
// C: int Sfm_ObjFaninNum( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) sfmObjFaninNum(i int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjFanoutNum (static inline)
// C: int Sfm_ObjFanoutNum( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) sfmObjFanoutNum(i int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjRefIncrement (static inline)
// C: int Sfm_ObjRefIncrement( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) sfmObjRefIncrement(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjRefDecrement (static inline)
// C: int Sfm_ObjRefDecrement( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) sfmObjRefDecrement(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjFanin (static inline)
// C: int Sfm_ObjFanin( Sfm_Ntk_t * p, int i, int k )
func (p *SfmNtk) sfmObjFanin(i int, k int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjFanout (static inline)
// C: int Sfm_ObjFanout( Sfm_Ntk_t * p, int i, int k )
func (p *SfmNtk) sfmObjFanout(i int, k int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjSatVar (static inline)
// C: int Sfm_ObjSatVar( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) sfmObjSatVar(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjSetSatVar (static inline)
// C: void Sfm_ObjSetSatVar( Sfm_Ntk_t * p, int iObj, int Num )
func (p *SfmNtk) sfmObjSetSatVar(iObj int, Num int) {
	panic("TODO, stubbed")
}

// Sfm_ObjCleanSatVar (static inline)
// C: void Sfm_ObjCleanSatVar( Sfm_Ntk_t * p, int Num )
func (p *SfmNtk) sfmObjCleanSatVar(Num int) {
	panic("TODO, stubbed")
}

// Sfm_NtkCleanVars (static inline)
// C: void Sfm_NtkCleanVars( Sfm_Ntk_t * p )
func (p *SfmNtk) sfmNtkCleanVars() {
	panic("TODO, stubbed")
}

// Sfm_ObjLevel (static inline)
// C: int Sfm_ObjLevel( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) sfmObjLevel(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjSetLevel (static inline)
// C: void Sfm_ObjSetLevel( Sfm_Ntk_t * p, int iObj, int Lev )
func (p *SfmNtk) sfmObjSetLevel(iObj int, Lev int) {
	panic("TODO, stubbed")
}

// Sfm_ObjLevelR (static inline)
// C: int Sfm_ObjLevelR( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) sfmObjLevelR(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjSetLevelR (static inline)
// C: void Sfm_ObjSetLevelR( Sfm_Ntk_t * p, int iObj, int Lev )
func (p *SfmNtk) sfmObjSetLevelR(iObj int, Lev int) {
	panic("TODO, stubbed")
}

// Sfm_ObjUpdateFaninCount (static inline)
// C: int Sfm_ObjUpdateFaninCount( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) sfmObjUpdateFaninCount(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjResetFaninCount (static inline)
// C: void Sfm_ObjResetFaninCount( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) sfmObjResetFaninCount(iObj int) {
	panic("TODO, stubbed")
}
