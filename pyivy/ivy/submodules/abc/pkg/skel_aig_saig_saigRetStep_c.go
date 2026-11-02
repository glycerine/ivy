package abc

// Hand-checked skeleton of src/aig/saig/saigRetStep.c.

// Saig_ManRetimeNodeFwd
// C: Aig_Obj_t * Saig_ManRetimeNodeFwd( Aig_Man_t * p, Aig_Obj_t * pObj, int fMakeBug )
func (p *AigManager) SaigManRetimeNodeFwd(pObj *AigObject, fMakeBug int) *AigObject {
	panic("TODO, stubbed")
}

// Saig_ManRetimeNodeBwd
// C: Aig_Obj_t * Saig_ManRetimeNodeBwd( Aig_Man_t * p, Aig_Obj_t * pObjLo )
func (p *AigManager) SaigManRetimeNodeBwd(pObjLo *AigObject) *AigObject {
	panic("TODO, stubbed")
}

// Saig_ManRetimeSteps
// C: int Saig_ManRetimeSteps( Aig_Man_t * p, int nSteps, int fForward, int fAddBugs )
func (p *AigManager) SaigManRetimeSteps(nSteps int, fForward int, fAddBugs int) int {
	panic("TODO, stubbed")
}
