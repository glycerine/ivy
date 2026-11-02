package abc

// Hand-checked skeleton of src/aig/saig/saigRetFwd.c.

// Aig_ObjFanoutStatic (static inline)
// C: Aig_Obj_t * Aig_ObjFanoutStatic( Aig_Obj_t * pObj, int i )
func aigObjFanoutStatic(pObj *AigObject, i int) *AigObject {
	panic("TODO, stubbed")
}

// Aig_ObjSetFanoutStatic (static inline)
// C: void Aig_ObjSetFanoutStatic( Aig_Obj_t * pObj, Aig_Obj_t * pFan )
func aigObjSetFanoutStatic(pObj *AigObject, pFan *AigObject) {
	panic("TODO, stubbed")
}

// Aig_ManStaticFanoutStart
// C: Aig_Obj_t ** Aig_ManStaticFanoutStart( Aig_Man_t * p )
func (p *AigManager) AigManStaticFanoutStart() []*AigObject {
	panic("TODO, stubbed")
}

// Aig_ManMarkAutonomous_rec
// C: void Aig_ManMarkAutonomous_rec( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) AigManMarkAutonomous_rec(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Saig_ManMarkAutonomous
// C: void Saig_ManMarkAutonomous( Aig_Man_t * p )
func (p *AigManager) SaigManMarkAutonomous() {
	panic("TODO, stubbed")
}

// Saig_ManRetimeForwardOne
// C: Aig_Man_t * Saig_ManRetimeForwardOne( Aig_Man_t * p, int * pnRegFixed, int * pnRegMoves )
func (p *AigManager) SaigManRetimeForwardOne(pnRegFixed *int, pnRegMoves *int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManRetimeForward
// C: Aig_Man_t * Saig_ManRetimeForward( Aig_Man_t * p, int nMaxIters, int fVerbose )
func (p *AigManager) SaigManRetimeForward(nMaxIters int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}
