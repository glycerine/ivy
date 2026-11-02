package abc

// Hand-checked skeleton of src/opt/lpk/lpkAbcDsd.c.

// Lpk_FunComputeMinSuppSizeVar
// C: int Lpk_FunComputeMinSuppSizeVar( Lpk_Fun_t * p, unsigned ** ppTruths, int nTruths, unsigned ** ppCofs, unsigned uNonDecSupp )
func (p *LpkFun) LpkFunComputeMinSuppSizeVar(ppTruths [][]uint32, nTruths int, ppCofs [][]uint32, uNonDecSupp uint32) int {
	panic("TODO, stubbed")
}

// Lpk_ComputeBoundSets_rec
// C: unsigned Lpk_ComputeBoundSets_rec( Kit_DsdNtk_t * p, int iLit, Vec_Int_t * vSets, int nSizeMax )
func LpkComputeBoundSets_rec(p *KitDsdNtk, iLit int, vSets *IntVector, nSizeMax int) uint32 {
	panic("TODO, stubbed")
}

// Lpk_ComputeBoundSets
// C: Vec_Int_t * Lpk_ComputeBoundSets( Kit_DsdNtk_t * p, int nSizeMax )
func LpkComputeBoundSets(p *KitDsdNtk, nSizeMax int) *IntVector {
	panic("TODO, stubbed")
}

// Lpk_PrintSetOne (static; lpkSets.c has a static of the same name)
// C: void Lpk_PrintSetOne( int uSupport )
func lpkPrintSetOne(uSupport int) {
	panic("TODO, stubbed")
}

// Lpk_PrintSets (static; lpkSets.c has a static of the same name)
// C: void Lpk_PrintSets( Vec_Int_t * vSets )
func lpkPrintSets(vSets *IntVector) {
	panic("TODO, stubbed")
}

// Lpk_MergeBoundSets
// C: Vec_Int_t * Lpk_MergeBoundSets( Vec_Int_t * vSets0, Vec_Int_t * vSets1, int nSizeMax )
func LpkMergeBoundSets(vSets0 *IntVector, vSets1 *IntVector, nSizeMax int) *IntVector {
	panic("TODO, stubbed")
}

// Lpk_FunCompareBoundSets
// C: void Lpk_FunCompareBoundSets( Lpk_Fun_t * p, Vec_Int_t * vBSets, int nCofDepth, unsigned uNonDecSupp, unsigned uLateArrSupp, Lpk_Res_t * pRes )
func (p *LpkFun) LpkFunCompareBoundSets(vBSets *IntVector, nCofDepth int, uNonDecSupp uint32, uLateArrSupp uint32, pRes *LpkRes) {
	panic("TODO, stubbed")
}

// Lpk_DsdLateArriving
// C: unsigned Lpk_DsdLateArriving( Lpk_Fun_t * p )
func (p *LpkFun) LpkDsdLateArriving() uint32 {
	panic("TODO, stubbed")
}

// Lpk_DsdAnalizeOne
// C: int Lpk_DsdAnalizeOne( Lpk_Fun_t * p, unsigned * ppTruths[5][16], Kit_DsdNtk_t * pNtks[], char pCofVars[], int nCofDepth, Lpk_Res_t * pRes )
func (p *LpkFun) LpkDsdAnalizeOne(ppTruths *[5][16][]uint32, pNtks []*KitDsdNtk, pCofVars []int8, nCofDepth int, pRes *LpkRes) int {
	panic("TODO, stubbed")
}

// Lpk_DsdAnalize
// C: Lpk_Res_t * Lpk_DsdAnalize( Lpk_Man_t * pMan, Lpk_Fun_t * p, int nShared )
func (pMan *LpkMan) LpkDsdAnalize(p *LpkFun, nShared int) *LpkRes {
	panic("TODO, stubbed")
}

// Lpk_DsdSplit
// C: Lpk_Fun_t * Lpk_DsdSplit( Lpk_Man_t * pMan, Lpk_Fun_t * p, char * pCofVars, int nCofVars, unsigned uBoundSet )
func (pMan *LpkMan) LpkDsdSplit(p *LpkFun, pCofVars []int8, nCofVars int, uBoundSet uint32) *LpkFun {
	panic("TODO, stubbed")
}
