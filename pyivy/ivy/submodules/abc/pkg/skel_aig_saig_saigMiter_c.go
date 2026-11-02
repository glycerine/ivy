package abc

// Hand-checked skeleton of src/aig/saig/saigMiter.c.

// Sec_MiterStatus (returns a Sec_MtrStatus_t by value)
// C: Sec_MtrStatus_t Sec_MiterStatus( Aig_Man_t * p )
func (p *AigManager) SecMiterStatus() SaigMiterStatus {
	panic("TODO, stubbed")
}

// Saig_ManCreateMiter
// C: Aig_Man_t * Saig_ManCreateMiter( Aig_Man_t * p0, Aig_Man_t * p1, int Oper )
func (p0 *AigManager) SaigManCreateMiter(p1 *AigManager, Oper int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManCreateMiterComb
// C: Aig_Man_t * Saig_ManCreateMiterComb( Aig_Man_t * p0, Aig_Man_t * p1, int Oper )
func (p0 *AigManager) SaigManCreateMiterComb(p1 *AigManager, Oper int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_AndDualRail
// C: void Saig_AndDualRail( Aig_Man_t * pNew, Aig_Obj_t * pObj, Aig_Obj_t ** ppData, Aig_Obj_t ** ppNext )
func (pNew *AigManager) SaigAndDualRail(pObj *AigObject, ppData []*AigObject, ppNext []*AigObject) {
	panic("TODO, stubbed")
}

// Saig_ManDualRail
// C: Aig_Man_t * Saig_ManDualRail( Aig_Man_t * p, int fMiter )
func (p *AigManager) SaigManDualRail(fMiter int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManUnrollTwo
// C: Aig_Man_t * Saig_ManUnrollTwo( Aig_Man_t * pBot, Aig_Man_t * pTop, int nFrames )
func (pBot *AigManager) SaigManUnrollTwo(pTop *AigManager, nFrames int) *AigManager {
	panic("TODO, stubbed")
}

// Aig_ManDupNodesAll
// C: Aig_Man_t * Aig_ManDupNodesAll( Aig_Man_t * p, Vec_Ptr_t * vSet )
func (p *AigManager) AigManDupNodesAll(vSet *PointerVector[AigObject]) *AigManager {
	panic("TODO, stubbed")
}

// Aig_ManDupNodesHalf
// C: Aig_Man_t * Aig_ManDupNodesHalf( Aig_Man_t * p, Vec_Ptr_t * vSet, int iPart )
func (p *AigManager) AigManDupNodesHalf(vSet *PointerVector[AigObject], iPart int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManDemiterSimple
// C: int Saig_ManDemiterSimple( Aig_Man_t * p, Aig_Man_t ** ppAig0, Aig_Man_t ** ppAig1 )
func (p *AigManager) SaigManDemiterSimple(ppAig0 **AigManager, ppAig1 **AigManager) int {
	panic("TODO, stubbed")
}

// Saig_ManDemiterMarkPos
// C: void Saig_ManDemiterMarkPos( Aig_Man_t * p )
func (p *AigManager) SaigManDemiterMarkPos() {
	panic("TODO, stubbed")
}

// Saig_ManDemiterCheckPo
// C: int Saig_ManDemiterCheckPo( Aig_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t ** ppPo0, Aig_Obj_t ** ppPo1 )
func (p *AigManager) SaigManDemiterCheckPo(pObj *AigObject, ppPo0 **AigObject, ppPo1 **AigObject) int {
	panic("TODO, stubbed")
}

// Saig_ManDemiterSimpleDiff
// C: int Saig_ManDemiterSimpleDiff( Aig_Man_t * p, Aig_Man_t ** ppAig0, Aig_Man_t ** ppAig1 )
func (p *AigManager) SaigManDemiterSimpleDiff(ppAig0 **AigManager, ppAig1 **AigManager) int {
	panic("TODO, stubbed")
}

// Saig_ManDemiterDual
// C: int Saig_ManDemiterDual( Aig_Man_t * p, Aig_Man_t ** ppAig0, Aig_Man_t ** ppAig1 )
func (p *AigManager) SaigManDemiterDual(ppAig0 **AigManager, ppAig1 **AigManager) int {
	panic("TODO, stubbed")
}

// Saig_ManDemiterSimpleDiff_old
// C: int Saig_ManDemiterSimpleDiff_old( Aig_Man_t * p, Aig_Man_t ** ppAig0, Aig_Man_t ** ppAig1 )
func (p *AigManager) SaigManDemiterSimpleDiff_old(ppAig0 **AigManager, ppAig1 **AigManager) int {
	panic("TODO, stubbed")
}

// Saig_ManDemiterLabel_rec
// C: void Saig_ManDemiterLabel_rec( Aig_Man_t * p, Aig_Obj_t * pObj, int Value )
func (p *AigManager) SaigManDemiterLabel_rec(pObj *AigObject, Value int) {
	panic("TODO, stubbed")
}

// Saig_ManGetLabeledRegister_rec
// C: Aig_Obj_t * Saig_ManGetLabeledRegister_rec( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) SaigManGetLabeledRegister_rec(pObj *AigObject) *AigObject {
	panic("TODO, stubbed")
}

// Saig_ManDemiter
// C: int Saig_ManDemiter( Aig_Man_t * p, Aig_Man_t ** ppAig0, Aig_Man_t ** ppAig1 )
func (p *AigManager) SaigManDemiter(ppAig0 **AigManager, ppAig1 **AigManager) int {
	panic("TODO, stubbed")
}

// Saig_ManCreateMiterTwo
// C: Aig_Man_t * Saig_ManCreateMiterTwo( Aig_Man_t * pOld, Aig_Man_t * pNew, int nFrames )
func (pOld *AigManager) SaigManCreateMiterTwo(pNew *AigManager, nFrames int) *AigManager {
	panic("TODO, stubbed")
}

// Ssw_SecCexResimulate
// C: int Ssw_SecCexResimulate( Aig_Man_t * p, int * pModel, int * pnOutputs )
func (p *AigManager) SswSecCexResimulate(pModel []int, pnOutputs *int) int {
	panic("TODO, stubbed")
}

// Ssw_SecSpecial
// C: int Ssw_SecSpecial( Aig_Man_t * pPart0, Aig_Man_t * pPart1, int nFrames, int fVerbose )
func (pPart0 *AigManager) SswSecSpecial(pPart1 *AigManager, nFrames int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Ssw_SecSpecialMiter
// C: int Ssw_SecSpecialMiter( Aig_Man_t * p0, Aig_Man_t * p1, int nFrames, int fVerbose )
func (p0 *AigManager) SswSecSpecialMiter(p1 *AigManager, nFrames int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Saig_ManDemiterNew
// C: int Saig_ManDemiterNew( Aig_Man_t * pMan )
func (pMan *AigManager) SaigManDemiterNew() int {
	panic("TODO, stubbed")
}
