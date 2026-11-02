package abc

// Hand-checked skeleton of src/base/abci/abcFx.c.

// FxMan is Fx_Man_t.
type FxMan struct {
	vCubes      *IntLevelVector     // cube -> lit
	LitCountMax int                 // max size of divisor to extract
	fCanonDivs  int                 // use only AND/XOR/MUX
	vLits       *IntLevelVector     // lit -> cube
	vCounts     *IntVector          // literal counts (currently not used)
	pHash       *IntVectorInterner  // Hsh_VecMan_t * : hash table for normalized divisors
	vWeights    *FloatVector        // divisor weights
	vPrio       *FloatPriorityQueue // priority queue for divisors by weight
	vVarCube    *IntVector          // mapping ObjId into its first cube
	vLevels     *IntVector          // variable levels
	vCubesS     *IntVector          // single cubes for the given divisor
	vCubesD     *IntVector          // cube pairs for the given divisor
	vCompls     *IntVector          // complemented attribute of each cube pair
	vCubeFree   *IntVector          // cube-free divisor
	vDiv        *IntVector          // selected divisor
	vSCC        *IntVector          // single cubes containment cubes
	timeStart   int64               // abctime : starting time
	nVars       int                 // original problem variables
	nLits       int                 // the number of SOP literals
	nDivs       int                 // the number of extracted divisors
	nCompls     int                 // the number of complements
	nPairsS     int                 // number of lit pairs
	nPairsD     int                 // number of cube pairs
	nDivsS      int                 // single cube divisors
	nDivMux     [3]int              // 0 = mux, 1 = compl mux, 2 = no mux
}

// Fx_ManGetFirstVarCube (static)
// C: int Fx_ManGetFirstVarCube( Fx_Man_t * p, Vec_Int_t * vCube )
func (p *FxMan) fxManGetFirstVarCube(vCube *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkFxRetrieve
// C: Vec_Wec_t * Abc_NtkFxRetrieve( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkFxRetrieve() *IntLevelVector {
	panic("TODO, stubbed")
}

// Abc_NtkFxInsert
// C: void Abc_NtkFxInsert( Abc_Ntk_t * pNtk, Vec_Wec_t * vCubes )
func (pNtk *ABCNetwork) AbcNtkFxInsert(vCubes *IntLevelVector) {
	panic("TODO, stubbed")
}

// Abc_NtkFxCheck
// C: int Abc_NtkFxCheck( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkFxCheck() int {
	panic("TODO, stubbed")
}

// Abc_NtkFxPerform
// C: int Abc_NtkFxPerform( Abc_Ntk_t * pNtk, int nNewNodesMax, int LitCountMax, int fCanonDivs, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) AbcNtkFxPerform(nNewNodesMax int, LitCountMax int, fCanonDivs int, fVerbose int, fVeryVerbose int) int {
	panic("TODO, stubbed")
}

// Fx_ManStart
// C: Fx_Man_t * Fx_ManStart( Vec_Wec_t * vCubes )
func FxManStart(vCubes *IntLevelVector) *FxMan {
	panic("TODO, stubbed")
}

// Fx_ManStop
// C: void Fx_ManStop( Fx_Man_t * p )
func (p *FxMan) FxManStop() {
	panic("TODO, stubbed")
}

// Fx_ManComputeLevelDiv (static)
// C: int Fx_ManComputeLevelDiv( Fx_Man_t * p, Vec_Int_t * vCubeFree )
func (p *FxMan) fxManComputeLevelDiv(vCubeFree *IntVector) int {
	panic("TODO, stubbed")
}

// Fx_ManComputeLevelCube (static)
// C: int Fx_ManComputeLevelCube( Fx_Man_t * p, Vec_Int_t * vCube )
func (p *FxMan) fxManComputeLevelCube(vCube *IntVector) int {
	panic("TODO, stubbed")
}

// Fx_ManComputeLevel
// C: void Fx_ManComputeLevel( Fx_Man_t * p )
func (p *FxMan) FxManComputeLevel() {
	panic("TODO, stubbed")
}

// Fx_PrintDivArray (static)
// C: void Fx_PrintDivArray( Vec_Int_t * vDiv )
func fxPrintDivArray(vDiv *IntVector) {
	panic("TODO, stubbed")
}

// Fx_PrintDiv (static)
// C: void Fx_PrintDiv( Fx_Man_t * p, int iDiv )
func (p *FxMan) fxPrintDiv(iDiv int) {
	panic("TODO, stubbed")
}

// Fx_PrintDivisors (static)
// C: void Fx_PrintDivisors( Fx_Man_t * p )
func (p *FxMan) fxPrintDivisors() {
	panic("TODO, stubbed")
}

// Fx_PrintStats (static)
// C: void Fx_PrintStats( Fx_Man_t * p, abctime clk )
func (p *FxMan) fxPrintStats(clk int64) {
	panic("TODO, stubbed")
}

// Fx_ManDivNormalize (static)
// C: int Fx_ManDivNormalize( Vec_Int_t * vCubeFree )
func fxManDivNormalize(vCubeFree *IntVector) int {
	panic("TODO, stubbed")
}

// Fx_ManDivFindCubeFree
// C: int Fx_ManDivFindCubeFree( Vec_Int_t * vArr1, Vec_Int_t * vArr2, Vec_Int_t * vCubeFree, int * fWarning )
func FxManDivFindCubeFree(vArr1 *IntVector, vArr2 *IntVector, vCubeFree *IntVector, fWarning *int) int {
	panic("TODO, stubbed")
}

// Fx_ManDivFindPivots (static)
// C: void Fx_ManDivFindPivots( Vec_Int_t * vDiv, int * pLit0, int * pLit1 )
func fxManDivFindPivots(vDiv *IntVector, pLit0 *int, pLit1 *int) {
	panic("TODO, stubbed")
}

// Fx_ManDivRemoveLits (static)
// C: int Fx_ManDivRemoveLits( Vec_Int_t * vCube, Vec_Int_t * vDiv, int fCompl )
func fxManDivRemoveLits(vCube *IntVector, vDiv *IntVector, fCompl int) int {
	panic("TODO, stubbed")
}

// Fx_ManDivAddLits (static)
// C: void Fx_ManDivAddLits( Vec_Int_t * vCube, Vec_Int_t * vCube2, Vec_Int_t * vDiv )
func fxManDivAddLits(vCube *IntVector, vCube2 *IntVector, vDiv *IntVector) {
	panic("TODO, stubbed")
}

// Fx_ManCreateLiterals
// C: void Fx_ManCreateLiterals( Fx_Man_t * p, int nVars )
func (p *FxMan) FxManCreateLiterals(nVars int) {
	panic("TODO, stubbed")
}

// Fx_ManCubeSingleCubeDivisors
// C: int Fx_ManCubeSingleCubeDivisors( Fx_Man_t * p, Vec_Int_t * vPivot, int fRemove, int fUpdate )
func (p *FxMan) FxManCubeSingleCubeDivisors(vPivot *IntVector, fRemove int, fUpdate int) int {
	panic("TODO, stubbed")
}

// Fx_ManCubeDoubleCubeDivisors
// C: void Fx_ManCubeDoubleCubeDivisors( Fx_Man_t * p, int iFirst, Vec_Int_t * vPivot, int fRemove, int fUpdate, int * fWarning )
func (p *FxMan) FxManCubeDoubleCubeDivisors(iFirst int, vPivot *IntVector, fRemove int, fUpdate int, fWarning *int) {
	panic("TODO, stubbed")
}

// Fx_ManCreateDivisors
// C: void Fx_ManCreateDivisors( Fx_Man_t * p )
func (p *FxMan) FxManCreateDivisors() {
	panic("TODO, stubbed")
}

// Fx_ManCompressCubes (static)
// C: void Fx_ManCompressCubes( Vec_Wec_t * vCubes, Vec_Int_t * vLit2Cube )
func fxManCompressCubes(vCubes *IntLevelVector, vLit2Cube *IntVector) {
	panic("TODO, stubbed")
}

// Fx_ManGetCubeVar (static)
// C: int Fx_ManGetCubeVar( Vec_Wec_t * vCubes, int iCube )
func fxManGetCubeVar(vCubes *IntLevelVector, iCube int) int {
	panic("TODO, stubbed")
}

// Fx_ManFindCommonPairs
// C: void Fx_ManFindCommonPairs( Vec_Wec_t * vCubes, Vec_Int_t * vPart0, Vec_Int_t * vPart1, Vec_Int_t * vPairs, Vec_Int_t * vCompls, Vec_Int_t * vDiv, Vec_Int_t * vCubeFree, int * fWarning )
func FxManFindCommonPairs(vCubes *IntLevelVector, vPart0 *IntVector, vPart1 *IntVector, vPairs *IntVector, vCompls *IntVector, vDiv *IntVector, vCubeFree *IntVector, fWarning *int) {
	panic("TODO, stubbed")
}

// Fx_ManUpdate
// C: void Fx_ManUpdate( Fx_Man_t * p, int iDiv, int * fWarning )
func (p *FxMan) FxManUpdate(iDiv int, fWarning *int) {
	panic("TODO, stubbed")
}

// Fx_FastExtract
// C: int Fx_FastExtract( Vec_Wec_t * vCubes, int ObjIdMax, int nNewNodesMax, int LitCountMax, int fCanonDivs, int fVerbose, int fVeryVerbose )
func FxFastExtract(vCubes *IntLevelVector, ObjIdMax int, nNewNodesMax int, LitCountMax int, fCanonDivs int, fVerbose int, fVeryVerbose int) int {
	panic("TODO, stubbed")
}
