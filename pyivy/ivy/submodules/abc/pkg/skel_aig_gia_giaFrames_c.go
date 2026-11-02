package abc

// Hand-checked skeleton of src/aig/gia/giaFrames.c.

// GiaManFra is Gia_ManFra_t.
type GiaManFra struct {
	pPars *GiaParFra                // parameters
	pAig  *GiaManager               // AIG to unroll
	vIns  *PointerVector[IntVector] // inputs of each timeframe
	vAnds *PointerVector[IntVector] // nodes of each timeframe
	vOuts *PointerVector[IntVector] // outputs of each timeframe
}

// GiaManUnr is Gia_ManUnr_t.
type GiaManUnr struct {
	pPars    *GiaParFra  // parameters
	pAig     *GiaManager // AIG to unroll (points to pOrder)
	pOrder   *GiaManager // AIG reordered (points to pAig)
	vLimit   *IntVector  // limits of each timeframe
	vRank    *IntVector  // rank of each node
	vDegree  *IntVector  // degree of each node
	vDegDiff *IntVector  // degree of each node
	vFirst   *IntVector  // first entry in the store
	vStore   *IntVector  // store for saved data
	pNew     *GiaManager // the resulting AIG
	LastLit  int         // the place to store the last literal
}

// Gia_ManUnrollDup_rec
// C: void Gia_ManUnrollDup_rec( Gia_Man_t * pNew, Gia_Obj_t * pObj, int Id )
func (pNew *GiaManager) GiaManUnrollDup_rec(pObj *GiaObject, Id int) {
	panic("TODO, stubbed")
}

// Gia_ManUnrollDup
// C: Gia_Man_t * Gia_ManUnrollDup( Gia_Man_t * p, Vec_Int_t * vLimit )
func (p *GiaManager) GiaManUnrollDup(vLimit *IntVector) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManUnrollAbs
// C: Vec_Ptr_t * Gia_ManUnrollAbs( Gia_Man_t * p, int nFrames )
func (p *GiaManager) GiaManUnrollAbs(nFrames int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Gia_ManUnrStart
// C: Gia_ManUnr_t * Gia_ManUnrStart( Gia_Man_t * pAig, Gia_ParFra_t * pPars )
func (pAig *GiaManager) GiaManUnrStart(pPars *GiaParFra) *GiaManUnr {
	panic("TODO, stubbed")
}

// Gia_ManUnrollStop (pMan is a void * that holds a Gia_ManUnr_t *)
// C: void Gia_ManUnrollStop( void * pMan )
func GiaManUnrollStop(pMan *GiaManUnr) {
	panic("TODO, stubbed")
}

// Gia_ObjUnrWrite (static inline)
// C: void Gia_ObjUnrWrite( Gia_ManUnr_t * p, int Id, int Entry )
func (p *GiaManUnr) giaObjUnrWrite(Id int, Entry int) {
	panic("TODO, stubbed")
}

// Gia_ObjUnrRead (static inline)
// C: int Gia_ObjUnrRead( Gia_ManUnr_t * p, int Id, int Degree )
func (p *GiaManUnr) giaObjUnrRead(Id int, Degree int) int {
	panic("TODO, stubbed")
}

// Gia_ObjUnrReadCopy0 (static inline)
// C: int Gia_ObjUnrReadCopy0( Gia_ManUnr_t * p, Gia_Obj_t * pObj, int Id )
func (p *GiaManUnr) giaObjUnrReadCopy0(pObj *GiaObject, Id int) int {
	panic("TODO, stubbed")
}

// Gia_ObjUnrReadCopy1 (static inline)
// C: int Gia_ObjUnrReadCopy1( Gia_ManUnr_t * p, Gia_Obj_t * pObj, int Id )
func (p *GiaManUnr) giaObjUnrReadCopy1(pObj *GiaObject, Id int) int {
	panic("TODO, stubbed")
}

// Gia_ObjUnrReadCi (static inline)
// C: int Gia_ObjUnrReadCi( Gia_ManUnr_t * p, int Id, int f, Gia_Man_t * pNew )
func (p *GiaManUnr) giaObjUnrReadCi(Id int, f int, pNew *GiaManager) int {
	panic("TODO, stubbed")
}

// Gia_ManUnrollStart (returns a void * that holds a Gia_ManUnr_t *)
// C: void * Gia_ManUnrollStart( Gia_Man_t * pAig, Gia_ParFra_t * pPars )
func (pAig *GiaManager) GiaManUnrollStart(pPars *GiaParFra) *GiaManUnr {
	panic("TODO, stubbed")
}

// Gia_ManUnrollAdd (pMan and the result are void * that hold a Gia_ManUnr_t * / Gia_Man_t *)
// C: void * Gia_ManUnrollAdd( void * pMan, int fMax )
func GiaManUnrollAdd(pMan *GiaManUnr, fMax int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManUnrollLastLit (pMan is a void * that holds a Gia_ManUnr_t *)
// C: int Gia_ManUnrollLastLit( void * pMan )
func GiaManUnrollLastLit(pMan *GiaManUnr) int {
	panic("TODO, stubbed")
}

// Gia_ManUnroll
// C: Gia_Man_t * Gia_ManUnroll( Gia_Man_t * pAig, Gia_ParFra_t * pPars )
func (pAig *GiaManager) GiaManUnroll(pPars *GiaParFra) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManFrames2
// C: Gia_Man_t * Gia_ManFrames2( Gia_Man_t * pAig, Gia_ParFra_t * pPars )
func (pAig *GiaManager) GiaManFrames2(pPars *GiaParFra) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManFraSetDefaultParams
// C: void Gia_ManFraSetDefaultParams( Gia_ParFra_t * p )
func GiaManFraSetDefaultParams(p *GiaParFra) {
	panic("TODO, stubbed")
}

// Gia_ManFraStart
// C: Gia_ManFra_t * Gia_ManFraStart( Gia_Man_t * pAig, Gia_ParFra_t * pPars )
func (pAig *GiaManager) GiaManFraStart(pPars *GiaParFra) *GiaManFra {
	panic("TODO, stubbed")
}

// Gia_ManFraStop
// C: void Gia_ManFraStop( Gia_ManFra_t * p )
func (p *GiaManFra) GiaManFraStop() {
	panic("TODO, stubbed")
}

// Gia_ManFraSupports
// C: void Gia_ManFraSupports( Gia_ManFra_t * p )
func (p *GiaManFra) GiaManFraSupports() {
	panic("TODO, stubbed")
}

// Gia_ManFramesInit
// C: Gia_Man_t * Gia_ManFramesInit( Gia_Man_t * pAig, Gia_ParFra_t * pPars )
func (pAig *GiaManager) GiaManFramesInit(pPars *GiaParFra) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManFrames
// C: Gia_Man_t * Gia_ManFrames( Gia_Man_t * pAig, Gia_ParFra_t * pPars )
func (pAig *GiaManager) GiaManFrames(pPars *GiaParFra) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManFramesInitSpecial
// C: Gia_Man_t * Gia_ManFramesInitSpecial( Gia_Man_t * pAig, int nFrames, int fVerbose )
func (pAig *GiaManager) GiaManFramesInitSpecial(nFrames int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
