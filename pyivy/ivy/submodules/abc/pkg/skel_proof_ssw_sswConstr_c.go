package abc

// Hand-checked skeleton of src/proof/ssw/sswConstr.c.

// Ssw_FramesWithConstraints
// C: Aig_Man_t * Ssw_FramesWithConstraints( Aig_Man_t * p, int nFrames )
func (p *AigManager) SswFramesWithConstraints(nFrames int) *AigManager {
	panic("TODO, stubbed")
}

// Ssw_ManSetConstrPhases (pvInits is an optional (nullable) output parameter)
// C: int Ssw_ManSetConstrPhases( Aig_Man_t * p, int nFrames, Vec_Int_t ** pvInits )
func (p *AigManager) SswManSetConstrPhases(nFrames int, pvInits **IntVector) int {
	panic("TODO, stubbed")
}

// Ssw_ManSetConstrPhases_ (pvInits is an optional (nullable) output parameter)
// C: int Ssw_ManSetConstrPhases_( Aig_Man_t * p, int nFrames, Vec_Int_t ** pvInits )
func (p *AigManager) SswManSetConstrPhases_(nFrames int, pvInits **IntVector) int {
	panic("TODO, stubbed")
}

// Ssw_ManPrintPolarity
// C: void Ssw_ManPrintPolarity( Aig_Man_t * p )
func (p *AigManager) SswManPrintPolarity() {
	panic("TODO, stubbed")
}

// Ssw_ManRefineByConstrSim
// C: void Ssw_ManRefineByConstrSim( Ssw_Man_t * p )
func (p *SswMan) SswManRefineByConstrSim() {
	panic("TODO, stubbed")
}

// Ssw_ManSweepNodeConstr
// C: int Ssw_ManSweepNodeConstr( Ssw_Man_t * p, Aig_Obj_t * pObj, int f, int fBmc )
func (p *SswMan) SswManSweepNodeConstr(pObj *AigObject, f int, fBmc int) int {
	panic("TODO, stubbed")
}

// Ssw_ManSweepBmcConstr_rec
// C: Aig_Obj_t * Ssw_ManSweepBmcConstr_rec( Ssw_Man_t * p, Aig_Obj_t * pObj, int f )
func (p *SswMan) SswManSweepBmcConstr_rec(pObj *AigObject, f int) *AigObject {
	panic("TODO, stubbed")
}

// Ssw_ManSweepBmcConstr_old
// C: int Ssw_ManSweepBmcConstr_old( Ssw_Man_t * p )
func (p *SswMan) SswManSweepBmcConstr_old() int {
	panic("TODO, stubbed")
}

// Ssw_ManSweepBmcConstr
// C: int Ssw_ManSweepBmcConstr( Ssw_Man_t * p )
func (p *SswMan) SswManSweepBmcConstr() int {
	panic("TODO, stubbed")
}

// Ssw_FramesWithClasses_rec
// C: Aig_Obj_t * Ssw_FramesWithClasses_rec( Ssw_Man_t * p, Aig_Obj_t * pObj, int f )
func (p *SswMan) SswFramesWithClasses_rec(pObj *AigObject, f int) *AigObject {
	panic("TODO, stubbed")
}

// Ssw_ManSweepConstr
// C: int Ssw_ManSweepConstr( Ssw_Man_t * p )
func (p *SswMan) SswManSweepConstr() int {
	panic("TODO, stubbed")
}
