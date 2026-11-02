package abc

// Hand-checked skeleton of src/proof/pdr/pdrCnf.c.

// The USE_PG variants (disabled by "//#define USE_PG" in the C source) are not
// represented; the live variants are.

// Pdr_ObjSatVar1 (static inline)
// C: int Pdr_ObjSatVar1( Pdr_Man_t * p, int k, Aig_Obj_t * pObj )
func (p *PdrMan) pdrObjSatVar1(k int, pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Pdr_ObjSatVar2FindOrAdd (static inline; live (non-USE_PG) variant)
// C: int Pdr_ObjSatVar2FindOrAdd( Pdr_Man_t * p, int k, Aig_Obj_t * pObj, int * pfNewVar )
func (p *PdrMan) pdrObjSatVar2FindOrAdd(k int, pObj *AigObject, pfNewVar *int) int {
	panic("TODO, stubbed")
}

// Pdr_ObjSatVar2 (live (non-USE_PG) variant)
// C: int Pdr_ObjSatVar2( Pdr_Man_t * p, int k, Aig_Obj_t * pObj, int Level, int Pol )
func (p *PdrMan) PdrObjSatVar2(k int, pObj *AigObject, Level int, Pol int) int {
	panic("TODO, stubbed")
}

// Pdr_ObjSatVar
// C: int Pdr_ObjSatVar( Pdr_Man_t * p, int k, int Pol, Aig_Obj_t * pObj )
func (p *PdrMan) PdrObjSatVar(k int, Pol int, pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Pdr_ObjRegNum1 (static inline)
// C: int Pdr_ObjRegNum1( Pdr_Man_t * p, int k, int iSatVar )
func (p *PdrMan) pdrObjRegNum1(k int, iSatVar int) int {
	panic("TODO, stubbed")
}

// Pdr_ObjRegNum2 (static inline)
// C: int Pdr_ObjRegNum2( Pdr_Man_t * p, int k, int iSatVar )
func (p *PdrMan) pdrObjRegNum2(k int, iSatVar int) int {
	panic("TODO, stubbed")
}

// Pdr_ObjRegNum
// C: int Pdr_ObjRegNum( Pdr_Man_t * p, int k, int iSatVar )
func (p *PdrMan) PdrObjRegNum(k int, iSatVar int) int {
	panic("TODO, stubbed")
}

// Pdr_ManFreeVar
// C: int Pdr_ManFreeVar( Pdr_Man_t * p, int k )
func (p *PdrMan) PdrManFreeVar(k int) int {
	panic("TODO, stubbed")
}

// Pdr_ManNewSolver1 (static inline)
// C: sat_solver * Pdr_ManNewSolver1( sat_solver * pSat, Pdr_Man_t * p, int k, int fInit )
func pdrManNewSolver1(pSat *SatSolver, p *PdrMan, k int, fInit int) *SatSolver {
	panic("TODO, stubbed")
}

// Pdr_ManNewSolver2 (static inline)
// C: sat_solver * Pdr_ManNewSolver2( sat_solver * pSat, Pdr_Man_t * p, int k, int fInit )
func pdrManNewSolver2(pSat *SatSolver, p *PdrMan, k int, fInit int) *SatSolver {
	panic("TODO, stubbed")
}

// Pdr_ManNewSolver
// C: sat_solver * Pdr_ManNewSolver( sat_solver * pSat, Pdr_Man_t * p, int k, int fInit )
func PdrManNewSolver(pSat *SatSolver, p *PdrMan, k int, fInit int) *SatSolver {
	panic("TODO, stubbed")
}
