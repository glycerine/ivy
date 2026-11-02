package abc

// Hand-checked skeleton of src/opt/lpk/lpkSets.c.

// LpkSet is Lpk_Set_t.
type LpkSet struct {
	iVar     int8   // char : the cofactoring variable
	Over     int8   // char : the overlap in supports
	SRed     int8   // char : the support reduction
	Size     int8   // char : the size of the boundset
	uSubset0 uint32 // the first subset (with removed)
	uSubset1 uint32 // the second subset (with removed)
}

// Lpk_ComputeSets_rec
// C: unsigned Lpk_ComputeSets_rec( Kit_DsdNtk_t * p, int iLit, Vec_Int_t * vSets )
func LpkComputeSets_rec(p *KitDsdNtk, iLit int, vSets *IntVector) uint32 {
	panic("TODO, stubbed")
}

// Lpk_ComputeSets
// C: unsigned Lpk_ComputeSets( Kit_DsdNtk_t * p, Vec_Int_t * vSets )
func LpkComputeSets(p *KitDsdNtk, vSets *IntVector) uint32 {
	panic("TODO, stubbed")
}

// Lpk_PrintSetOne (static; lpkAbcDsd.c has a static of the same name, so this one carries a _sets suffix)
// C: void Lpk_PrintSetOne( int uSupport )
func lpkPrintSetOne_sets(uSupport int) {
	panic("TODO, stubbed")
}

// Lpk_PrintSets (static; lpkAbcDsd.c has a static of the same name, so this one carries a _sets suffix)
// C: void Lpk_PrintSets( Vec_Int_t * vSets )
func lpkPrintSets_sets(vSets *IntVector) {
	panic("TODO, stubbed")
}

// Lpk_ComposeSets (pSize is an in/out counter)
// C: void Lpk_ComposeSets( Vec_Int_t * vSets0, Vec_Int_t * vSets1, int nVars, int iCofVar, Lpk_Set_t * pStore, int * pSize, int nSizeLimit )
func LpkComposeSets(vSets0 *IntVector, vSets1 *IntVector, nVars int, iCofVar int, pStore []LpkSet, pSize *int, nSizeLimit int) {
	panic("TODO, stubbed")
}

// Lpk_MapSuppPrintSet
// C: void Lpk_MapSuppPrintSet( Lpk_Set_t * pSet, int i )
func LpkMapSuppPrintSet(pSet *LpkSet, i int) {
	panic("TODO, stubbed")
}

// Lpk_MapSuppRedDecSelect (piVar and piVarReused are out parameters)
// C: unsigned Lpk_MapSuppRedDecSelect( Lpk_Man_t * p, unsigned * pTruth, int nVars, int * piVar, int * piVarReused )
func (p *LpkMan) LpkMapSuppRedDecSelect(pTruth []uint32, nVars int, piVar *int, piVarReused *int) uint32 {
	panic("TODO, stubbed")
}
