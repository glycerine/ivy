package abc

// Hand-checked skeleton of src/opt/lpk/lpkAbcUtil.c.

// Lpk_FunAlloc
// C: Lpk_Fun_t * Lpk_FunAlloc( int nVars )
func LpkFunAlloc(nVars int) *LpkFun {
	panic("TODO, stubbed")
}

// Lpk_FunFree
// C: void Lpk_FunFree( Lpk_Fun_t * p )
func (p *LpkFun) LpkFunFree() {
	panic("TODO, stubbed")
}

// Lpk_FunCreate
// C: Lpk_Fun_t * Lpk_FunCreate( Abc_Ntk_t * pNtk, Vec_Ptr_t * vLeaves, unsigned * pTruth, int nLutK, int AreaLim, int DelayLim )
func LpkFunCreate(pNtk *ABCNetwork, vLeaves *PointerVector[ABCNetworkObject], pTruth []uint32, nLutK int, AreaLim int, DelayLim int) *LpkFun {
	panic("TODO, stubbed")
}

// Lpk_FunDup
// C: Lpk_Fun_t * Lpk_FunDup( Lpk_Fun_t * p, unsigned * pTruth )
func (p *LpkFun) LpkFunDup(pTruth []uint32) *LpkFun {
	panic("TODO, stubbed")
}

// Lpk_FunSuppMinimize
// C: int Lpk_FunSuppMinimize( Lpk_Fun_t * p )
func (p *LpkFun) LpkFunSuppMinimize() int {
	panic("TODO, stubbed")
}

// Lpk_FunComputeCofSupps
// C: void Lpk_FunComputeCofSupps( Lpk_Fun_t * p )
func (p *LpkFun) LpkFunComputeCofSupps() {
	panic("TODO, stubbed")
}

// Lpk_SuppDelay
// C: int Lpk_SuppDelay( unsigned uSupp, int * pDelays )
func LpkSuppDelay(uSupp uint32, pDelays []int) int {
	panic("TODO, stubbed")
}

// Lpk_SuppToVars (pVars is filled in)
// C: int Lpk_SuppToVars( unsigned uBoundSet, char * pVars )
func LpkSuppToVars(uBoundSet uint32, pVars []int8) int {
	panic("TODO, stubbed")
}
