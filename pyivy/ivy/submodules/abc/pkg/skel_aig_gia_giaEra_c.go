package abc

// Hand-checked skeleton of src/aig/gia/giaEra.c.

// GiaObjEra is Gia_ObjEra_t.
type GiaObjEra struct {
	Num   int      // ID of this state
	Cond  int      // input condition
	iPrev int      // previous state
	iNext int      // next state in the hash table
	pData []uint32 // unsigned pData[0] : state bits (flexible array member)
}

// GiaManEra is Gia_ManEra_t.
type GiaManEra struct {
	pAig      *GiaManager               // user's AIG manager
	nWordsSim int                       // 2^(PInum)
	nWordsDat int                       // Abc_BitWordNum
	pDataSim  []uint32                  // unsigned * : simulation data
	pMemory   *MmrFixedArena            // Mem_Fixed_t * : memory manager
	vStates   *PointerVector[GiaObjEra] // reached states
	pStateNew *GiaObjEra                // temporary state
	iCurState int                       // the current state
	vBugTrace *IntVector                // the sequence of transitions
	vStgDump  *IntVector                // STG written into a file
	nBins     int
	pBins     []uint32 // unsigned *
}

// Gia_ManEraData (static inline)
// C: unsigned * Gia_ManEraData( Gia_ManEra_t * p, int i )
func (p *GiaManEra) giaManEraData(i int) []uint32 {
	panic("TODO, stubbed")
}

// Gia_ManEraState (static inline)
// C: Gia_ObjEra_t * Gia_ManEraState( Gia_ManEra_t * p, int i )
func (p *GiaManEra) giaManEraState(i int) *GiaObjEra {
	panic("TODO, stubbed")
}

// Gia_ManEraCreate
// C: Gia_ManEra_t * Gia_ManEraCreate( Gia_Man_t * pAig )
func (pAig *GiaManager) GiaManEraCreate() *GiaManEra {
	panic("TODO, stubbed")
}

// Gia_ManEraFree
// C: void Gia_ManEraFree( Gia_ManEra_t * p )
func (p *GiaManEra) GiaManEraFree() {
	panic("TODO, stubbed")
}

// Gia_ManEraCreateState
// C: Gia_ObjEra_t * Gia_ManEraCreateState( Gia_ManEra_t * p )
func (p *GiaManEra) GiaManEraCreateState() *GiaObjEra {
	panic("TODO, stubbed")
}

// Gia_ManEraStateHash
// C: int Gia_ManEraStateHash( unsigned * pState, int nWordsSim, int nTableSize )
func GiaManEraStateHash(pState []uint32, nWordsSim int, nTableSize int) int {
	panic("TODO, stubbed")
}

// Gia_ManEraHashFind (static inline)
// C: unsigned * Gia_ManEraHashFind( Gia_ManEra_t * p, Gia_ObjEra_t * pState, int * pStateNum )
func (p *GiaManEra) giaManEraHashFind(pState *GiaObjEra, pStateNum *int) *uint32 {
	panic("TODO, stubbed")
}

// Gia_ManEraHashResize
// C: void Gia_ManEraHashResize( Gia_ManEra_t * p )
func (p *GiaManEra) GiaManEraHashResize() {
	panic("TODO, stubbed")
}

// Gia_ManInsertState
// C: void Gia_ManInsertState( Gia_ManEra_t * p, Gia_ObjEra_t * pState )
func (p *GiaManEra) GiaManInsertState(pState *GiaObjEra) {
	panic("TODO, stubbed")
}

// Gia_ManOutputAsserted (static inline)
// C: int Gia_ManOutputAsserted( Gia_ManEra_t * p, Gia_Obj_t * pObj )
func (p *GiaManEra) giaManOutputAsserted(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Gia_ManSimulateCo (static inline)
// C: void Gia_ManSimulateCo( Gia_ManEra_t * p, Gia_Obj_t * pObj )
func (p *GiaManEra) giaManSimulateCo(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManSimulateNode (static inline)
// C: void Gia_ManSimulateNode( Gia_ManEra_t * p, Gia_Obj_t * pObj )
func (p *GiaManEra) giaManSimulateNode(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManPerformOneIter
// C: void Gia_ManPerformOneIter( Gia_ManEra_t * p )
func (p *GiaManEra) GiaManPerformOneIter() {
	panic("TODO, stubbed")
}

// Gia_ManCollectBugTrace
// C: Vec_Int_t * Gia_ManCollectBugTrace( Gia_ManEra_t * p, Gia_ObjEra_t * pState, int iCond )
func (p *GiaManEra) GiaManCollectBugTrace(pState *GiaObjEra, iCond int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManCountDepth
// C: int Gia_ManCountDepth( Gia_ManEra_t * p )
func (p *GiaManEra) GiaManCountDepth() int {
	panic("TODO, stubbed")
}

// Gia_ManAnalyzeResult
// C: int Gia_ManAnalyzeResult( Gia_ManEra_t * p, Gia_ObjEra_t * pState, int fMiter, int fStgDump )
func (p *GiaManEra) GiaManAnalyzeResult(pState *GiaObjEra, fMiter int, fStgDump int) int {
	panic("TODO, stubbed")
}

// Gia_ManCollectReachable
// C: int Gia_ManCollectReachable( Gia_Man_t * pAig, int nStatesMax, int fMiter, int fDumpFile, int fVerbose )
func (pAig *GiaManager) GiaManCollectReachable(nStatesMax int, fMiter int, fDumpFile int, fVerbose int) int {
	panic("TODO, stubbed")
}
