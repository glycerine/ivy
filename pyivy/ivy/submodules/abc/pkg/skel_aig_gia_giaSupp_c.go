package abc

// Hand-checked skeleton of src/aig/gia/giaSupp.c.

// GiaManMin is Gia_ManMin_t.
type GiaManMin struct {
	pGia     *GiaManager
	iLits    [2]int
	vCis     [2]*IntVector
	vObjs    [2]*IntVector
	vCleared *IntVector
	dd       *CuddFoundation
	vFuncs   *PointerVector[DdNode]
	vSupp    *IntVector
}

// GiaMan2Min is Gia_Man2Min_t.
type GiaMan2Min struct {
	pGia      *GiaManager
	iLits     [2]int
	vCis      [2]*IntVector
	vObjs     [2]*IntVector
	pSat      *SatokoSolver             // satoko_t * : SAT solver
	vSims     *WordVector               // simulation
	vFrontier *PointerVector[GiaObject] // CNF construction
	vFanins   *PointerVector[GiaObject] // CNF construction
	vSatVars  *IntVector                // nodes
	nCisOld   int                       // previous number of CIs
	iPattern  int                       // the next pattern to write
	nSatSat   int
	nSatUnsat int
	nCalls    int
	nSims     int
	nSupps    int
}

// Gia_ManSuppStart
// C: Gia_ManMin_t * Gia_ManSuppStart( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaManSuppStart() *GiaManMin {
	panic("TODO, stubbed")
}

// Gia_ManSuppStop
// C: void Gia_ManSuppStop( Gia_ManMin_t * p )
func (p *GiaManMin) GiaManSuppStop() {
	panic("TODO, stubbed")
}

// Gia_ManFindRemoved
// C: int Gia_ManFindRemoved( Gia_ManMin_t * p )
func (p *GiaManMin) GiaManFindRemoved() int {
	panic("TODO, stubbed")
}

// Gia_ManRebuildOne
// C: int Gia_ManRebuildOne( Gia_ManMin_t * p, int n )
func (p *GiaManMin) GiaManRebuildOne(n int) int {
	panic("TODO, stubbed")
}

// Gia_ManGatherSupp_rec (static inline)
// C: int Gia_ManGatherSupp_rec( Gia_Man_t * p, int iObj, Vec_Int_t * vCis, Vec_Int_t * vObjs )
func (p *GiaManager) giaManGatherSupp_rec(iObj int, vCis *IntVector, vObjs *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManGatherSupp
// C: int Gia_ManGatherSupp( Gia_ManMin_t * p )
func (p *GiaManMin) GiaManGatherSupp() int {
	panic("TODO, stubbed")
}

// Gia_ManSupportAnd
// C: int Gia_ManSupportAnd( Gia_ManMin_t * p, int iLit0, int iLit1 )
func (p *GiaManMin) GiaManSupportAnd(iLit0 int, iLit1 int) int {
	panic("TODO, stubbed")
}

// Gia_ManSupportAndTest
// C: Gia_Man_t * Gia_ManSupportAndTest( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaManSupportAndTest() *GiaManager {
	panic("TODO, stubbed")
}

// Gia_Min2ObjSatId (static inline)
// C: int Gia_Min2ObjSatId( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) giaMin2ObjSatId(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Gia_Min2ObjSetSatId (static inline)
// C: int Gia_Min2ObjSetSatId( Gia_Man_t * p, Gia_Obj_t * pObj, int Num )
func (p *GiaManager) giaMin2ObjSetSatId(pObj *GiaObject, Num int) int {
	panic("TODO, stubbed")
}

// Gia_Min2ObjCleanSatId (static inline)
// C: void Gia_Min2ObjCleanSatId( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) giaMin2ObjCleanSatId(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_Man2SuppStart
// C: Gia_Man2Min_t * Gia_Man2SuppStart( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaMan2SuppStart() *GiaMan2Min {
	panic("TODO, stubbed")
}

// Gia_Man2SuppStop
// C: void Gia_Man2SuppStop( Gia_Man2Min_t * p )
func (p *GiaMan2Min) GiaMan2SuppStop() {
	panic("TODO, stubbed")
}

// Gia_Min2AddClausesMux
// C: void Gia_Min2AddClausesMux( Gia_Man_t * p, Gia_Obj_t * pNode, satoko_t * pSat )
func (p *GiaManager) GiaMin2AddClausesMux(pNode *GiaObject, pSat *SatokoSolver) {
	panic("TODO, stubbed")
}

// Gia_Min2AddClausesSuper
// C: void Gia_Min2AddClausesSuper( Gia_Man_t * p, Gia_Obj_t * pNode, Vec_Ptr_t * vSuper, satoko_t * pSat )
func (p *GiaManager) GiaMin2AddClausesSuper(pNode *GiaObject, vSuper *PointerVector[GiaObject], pSat *SatokoSolver) {
	panic("TODO, stubbed")
}

// Gia_Min2CollectSuper_rec
// C: void Gia_Min2CollectSuper_rec( Gia_Obj_t * pObj, Vec_Ptr_t * vSuper, int fFirst, int fUseMuxes )
func GiaMin2CollectSuper_rec(pObj *GiaObject, vSuper *PointerVector[GiaObject], fFirst int, fUseMuxes int) {
	panic("TODO, stubbed")
}

// Gia_Min2CollectSuper
// C: void Gia_Min2CollectSuper( Gia_Obj_t * pObj, int fUseMuxes, Vec_Ptr_t * vSuper )
func GiaMin2CollectSuper(pObj *GiaObject, fUseMuxes int, vSuper *PointerVector[GiaObject]) {
	panic("TODO, stubbed")
}

// Gia_Min2ObjAddToFrontier
// C: void Gia_Min2ObjAddToFrontier( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Ptr_t * vFrontier, satoko_t * pSat, Vec_Int_t * vSatVars )
func (p *GiaManager) GiaMin2ObjAddToFrontier(pObj *GiaObject, vFrontier *PointerVector[GiaObject], pSat *SatokoSolver, vSatVars *IntVector) {
	panic("TODO, stubbed")
}

// Gia_Min2ObjGetCnfVar
// C: int Gia_Min2ObjGetCnfVar( Gia_Man2Min_t * p, int iObj )
func (p *GiaMan2Min) GiaMin2ObjGetCnfVar(iObj int) int {
	panic("TODO, stubbed")
}

// Gia_Min2ManSimulate
// C: int Gia_Min2ManSimulate( Gia_Man2Min_t * p )
func (p *GiaMan2Min) GiaMin2ManSimulate() int {
	panic("TODO, stubbed")
}

// Gia_Min2SimSetInputBit (static inline)
// C: void Gia_Min2SimSetInputBit( Gia_Man2Min_t * p, int iObj, int Bit, int iPattern )
func (p *GiaMan2Min) giaMin2SimSetInputBit(iObj int, Bit int, iPattern int) {
	panic("TODO, stubbed")
}

// Gia_Min2ManSolve
// C: int Gia_Min2ManSolve( Gia_Man2Min_t * p )
func (p *GiaMan2Min) GiaMin2ManSolve() int {
	panic("TODO, stubbed")
}

// Gia_Min2ManGatherSupp_rec (static inline)
// C: int Gia_Min2ManGatherSupp_rec( Gia_Man_t * p, int iObj, Vec_Int_t * vCis, Vec_Int_t * vObjs )
func (p *GiaManager) giaMin2ManGatherSupp_rec(iObj int, vCis *IntVector, vObjs *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_Min2ManGatherSupp
// C: int Gia_Min2ManGatherSupp( Gia_Man2Min_t * p )
func (p *GiaMan2Min) GiaMin2ManGatherSupp() int {
	panic("TODO, stubbed")
}

// Gia_Man2SupportAnd
// C: int Gia_Man2SupportAnd( Gia_Man2Min_t * p, int iLit0, int iLit1 )
func (p *GiaMan2Min) GiaMan2SupportAnd(iLit0 int, iLit1 int) int {
	panic("TODO, stubbed")
}

// Gia_Man2SupportAndTest
// C: Gia_Man_t * Gia_Man2SupportAndTest( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaMan2SupportAndTest() *GiaManager {
	panic("TODO, stubbed")
}
