package abc

// Hand-checked skeleton of src/opt/res/resSim_old.c.

// Res_SimAlloc (named with _old: resSim.c defines the same function)
// C: Res_Sim_t * Res_SimAlloc( int nWords )
func ResSimAlloc_old(nWords int) *ResSim {
	panic("TODO, stubbed")
}

// Res_SimAdjust (named with _old: resSim.c defines the same function)
// C: void Res_SimAdjust( Res_Sim_t * p, Abc_Ntk_t * pAig )
func (p *ResSim) ResSimAdjust_old(pAig *ABCNetwork) {
	panic("TODO, stubbed")
}

// Res_SimFree (named with _old: resSim.c defines the same function)
// C: void Res_SimFree( Res_Sim_t * p )
func (p *ResSim) ResSimFree_old() {
	panic("TODO, stubbed")
}

// Res_SimSetRandom (named with _old: resSim.c defines the same function)
// C: void Res_SimSetRandom( Res_Sim_t * p )
func (p *ResSim) ResSimSetRandom_old() {
	panic("TODO, stubbed")
}

// Res_SimSetGiven (named with _old: resSim.c defines the same function)
// C: void Res_SimSetGiven( Res_Sim_t * p, Vec_Ptr_t * vInfo )
func (p *ResSim) ResSimSetGiven_old(vInfo *PointerVector[uint32]) {
	panic("TODO, stubbed")
}

// Res_SimPerformOne (named with _old: resSim.c defines the same function)
// C: void Res_SimPerformOne( Abc_Obj_t * pNode, Vec_Ptr_t * vSimInfo, int nSimWords )
func ResSimPerformOne_old(pNode *ABCNetworkObject, vSimInfo *PointerVector[uint32], nSimWords int) {
	panic("TODO, stubbed")
}

// Res_SimTransferOne (named with _old: resSim.c defines the same function)
// C: void Res_SimTransferOne( Abc_Obj_t * pNode, Vec_Ptr_t * vSimInfo, int nSimWords )
func ResSimTransferOne_old(pNode *ABCNetworkObject, vSimInfo *PointerVector[uint32], nSimWords int) {
	panic("TODO, stubbed")
}

// Res_SimPerformRound (named with _old: resSim.c defines the same function)
// C: void Res_SimPerformRound( Res_Sim_t * p )
func (p *ResSim) ResSimPerformRound_old() {
	panic("TODO, stubbed")
}

// Res_SimProcessPats (named with _old: resSim.c defines the same function)
// C: void Res_SimProcessPats( Res_Sim_t * p )
func (p *ResSim) ResSimProcessPats_old() {
	panic("TODO, stubbed")
}

// Res_SimPadSimInfo (named with _old: resSim.c defines the same function)
// C: void Res_SimPadSimInfo( Vec_Ptr_t * vPats, int nPats, int nWords )
func ResSimPadSimInfo_old(vPats *PointerVector[uint32], nPats int, nWords int) {
	panic("TODO, stubbed")
}

// Res_SimDeriveInfoReplicate (named with _old: resSim.c defines the same function)
// C: void Res_SimDeriveInfoReplicate( Res_Sim_t * p )
func (p *ResSim) ResSimDeriveInfoReplicate_old() {
	panic("TODO, stubbed")
}

// Res_SimDeriveInfoComplement (named with _old: resSim.c defines the same function)
// C: void Res_SimDeriveInfoComplement( Res_Sim_t * p )
func (p *ResSim) ResSimDeriveInfoComplement_old() {
	panic("TODO, stubbed")
}

// Res_SimReportOne (named with _old: resSim.c defines the same function)
// C: void Res_SimReportOne( Res_Sim_t * p )
func (p *ResSim) ResSimReportOne_old() {
	panic("TODO, stubbed")
}

// Res_SimPrintOutPatterns (named with _old: resSim.c defines the same function)
// C: void Res_SimPrintOutPatterns( Res_Sim_t * p, Abc_Ntk_t * pAig )
func (p *ResSim) ResSimPrintOutPatterns_old(pAig *ABCNetwork) {
	panic("TODO, stubbed")
}

// Res_SimPrepare (named with _old: resSim.c defines the same function)
// C: int Res_SimPrepare( Res_Sim_t * p, Abc_Ntk_t * pAig, int nTruePis, int fVerbose )
func (p *ResSim) ResSimPrepare_old(pAig *ABCNetwork, nTruePis int, fVerbose int) int {
	panic("TODO, stubbed")
}
