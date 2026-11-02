package abc

// Hand-checked skeleton of src/aig/gia/giaSatEdge.c.

// SegMan is Seg_Man_t.
type SegMan struct {
	pSat      *SatSolver // SAT solver
	nVars     int        // max vars (edge num)
	LogN      int        // base-2 log of max vars
	Power2    int        // power-2 of LogN
	FirstVar  int        // first variable to be used
	nBTLimit  int        // conflicts
	DelayMax  int        // external delay
	nEdges    int        // the number of edges
	fDelay    int        // delay mode
	fReverse  int        // reverse windowing
	fVerbose  int        // verbose
	pGia      *GiaManager
	vPolars   *IntVector // polarity
	vToSkip   *IntVector // edges to skip
	vEdges    *IntVector // edges as fanin/fanout pairs
	vFirsts   *IntVector // first SAT variable
	vNvars    *IntVector // number of SAT variables
	vLits     *IntVector // literals
	pLevels   []int      // int * : levels
	timeStart int64      // abctime
}

// Seg_ManCountIntEdges
// C: Vec_Int_t * Seg_ManCountIntEdges( Gia_Man_t * p, Vec_Int_t * vPolars, Vec_Int_t * vToSkip, int nFanouts )
func (p *GiaManager) SegManCountIntEdges(vPolars *IntVector, vToSkip *IntVector, nFanouts int) *IntVector {
	panic("TODO, stubbed")
}

// Seg_ManCollectObjEdges
// C: Vec_Wec_t * Seg_ManCollectObjEdges( Vec_Int_t * vEdges, int nObjs )
func SegManCollectObjEdges(vEdges *IntVector, nObjs int) *IntLevelVector {
	panic("TODO, stubbed")
}

// Seg_ManCountIntLevels
// C: int Seg_ManCountIntLevels( Seg_Man_t * p, int iStartVar )
func (p *SegMan) SegManCountIntLevels(iStartVar int) int {
	panic("TODO, stubbed")
}

// Seg_ManAlloc
// C: Seg_Man_t * Seg_ManAlloc( Gia_Man_t * pGia, int nFanouts )
func (pGia *GiaManager) SegManAlloc(nFanouts int) *SegMan {
	panic("TODO, stubbed")
}

// Seg_ManClean
// C: void Seg_ManClean( Seg_Man_t * p )
func (p *SegMan) SegManClean() {
	panic("TODO, stubbed")
}

// Seg_ManStop
// C: void Seg_ManStop( Seg_Man_t * p )
func (p *SegMan) SegManStop() {
	panic("TODO, stubbed")
}

// Seg_ManCreateCnf
// C: void Seg_ManCreateCnf( Seg_Man_t * p, int fTwo, int fVerbose )
func (p *SegMan) SegManCreateCnf(fTwo int, fVerbose int) {
	panic("TODO, stubbed")
}

// Seg_ManConvertResult
// C: Vec_Int_t * Seg_ManConvertResult( Seg_Man_t * p )
func (p *SegMan) SegManConvertResult() *IntVector {
	panic("TODO, stubbed")
}

// Seg_ManComputeDelay
// C: void Seg_ManComputeDelay( Gia_Man_t * pGia, int DelayInit, int nFanouts, int fTwo, int fVerbose )
func (pGia *GiaManager) SegManComputeDelay(DelayInit int, nFanouts int, fTwo int, fVerbose int) {
	panic("TODO, stubbed")
}
