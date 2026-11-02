package abc

// Hand-checked skeleton of src/base/wlc/wlcPth.c.

// With ABC_USE_PTHREADS undefined, the #ifndef branch (stubs) is the live code;
// the signatures are the same in both branches.

// Wla_CallBackToStop
// C: int Wla_CallBackToStop( int RunId )
func WlaCallBackToStop(RunId int) int {
	panic("TODO, stubbed")
}

// Wla_GetGlobalRunId
// C: int Wla_GetGlobalRunId()
func WlaGetGlobalRunId() int {
	panic("TODO, stubbed")
}

// Wla_ManJoinThread
// C: void Wla_ManJoinThread( Wla_Man_t * pWla, int RunId )
func (pWla *WlaMan) WlaManJoinThread(RunId int) {
	panic("TODO, stubbed")
}

// Wla_ManConcurrentBmc3 (ppCex is an out parameter)
// C: void Wla_ManConcurrentBmc3( Wla_Man_t * pWla, Aig_Man_t * pAig, Abc_Cex_t ** ppCex )
func (pWla *WlaMan) WlaManConcurrentBmc3(pAig *AigManager, ppCex **Counterexample) {
	panic("TODO, stubbed")
}
