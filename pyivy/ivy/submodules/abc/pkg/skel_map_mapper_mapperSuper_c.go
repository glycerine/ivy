package abc

import "os"

// Hand-checked skeleton of src/map/mapper/mapperSuper.c.

// Map_LibraryRead
// C: int Map_LibraryRead( Map_SuperLib_t * pLib, char * pFileName )
func (pLib *MapSuperLib) MapLibraryRead(pFileName string) int {
	panic("TODO, stubbed")
}

// Map_LibraryReadFile (static)
// C: int Map_LibraryReadFile( Map_SuperLib_t * pLib, FILE * pFile )
func (pLib *MapSuperLib) mapLibraryReadFile(pFile *os.File) int {
	panic("TODO, stubbed")
}

// Map_LibraryReadGate (static)
// C: Map_Super_t * Map_LibraryReadGate( Map_SuperLib_t * pLib, char * pBuffer, int nVars )
func (pLib *MapSuperLib) mapLibraryReadGate(pBuffer string, nVars int) *MapSuper {
	panic("TODO, stubbed")
}

// Map_LibraryReadFormulaStep (returns the remainder of the formula; pStrings is filled in and pnStrings is an in/out counter)
// C: char * Map_LibraryReadFormulaStep( char * pFormula, char * pStrings[], int * pnStrings )
func MapLibraryReadFormulaStep(pFormula string, pStrings []string, pnStrings *int) string {
	panic("TODO, stubbed")
}

// Map_LibraryTruthVerify (static)
// C: int Map_LibraryTruthVerify( Map_SuperLib_t * pLib, Map_Super_t * pGate )
func (pLib *MapSuperLib) mapLibraryTruthVerify(pGate *MapSuper) int {
	panic("TODO, stubbed")
}

// Map_LibraryComputeTruth (static)
// C: void Map_LibraryComputeTruth( Map_SuperLib_t * pLib, char * pFormula, unsigned uTruthRes[] )
func (pLib *MapSuperLib) mapLibraryComputeTruth(pFormula string, uTruthRes []uint32) {
	panic("TODO, stubbed")
}

// Map_LibraryComputeTruth_rec (static)
// C: void Map_LibraryComputeTruth_rec( Map_SuperLib_t * pLib, char * pFormula, unsigned uTruthsIn[][2], unsigned uTruthRes[] )
func (pLib *MapSuperLib) mapLibraryComputeTruth_rec(pFormula string, uTruthsIn [][2]uint32, uTruthRes []uint32) {
	panic("TODO, stubbed")
}

// Map_LibraryPrintSupergate
// C: void Map_LibraryPrintSupergate( Map_Super_t * pGate )
func MapLibraryPrintSupergate(pGate *MapSuper) {
	panic("TODO, stubbed")
}

// Map_LibraryPrintClasses (static)
// C: void Map_LibraryPrintClasses( Map_SuperLib_t * p )
func (p *MapSuperLib) mapLibraryPrintClasses() {
	panic("TODO, stubbed")
}
