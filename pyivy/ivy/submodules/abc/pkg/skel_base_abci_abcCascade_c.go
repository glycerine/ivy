package abc

// Hand-checked skeleton of src/base/abci/abcCascade.c.

// Abc_ResBuildBdd
// C: DdNode * Abc_ResBuildBdd( Abc_Ntk_t * pNtk, DdManager * dd )
func AbcResBuildBdd(pNtk *ABCNetwork, dd *CuddFoundation) *DdNode {
	panic("TODO, stubbed")
}

// Abc_ResStartPart
// C: void Abc_ResStartPart( int nInputs, unsigned uParts[], int nParts )
func AbcResStartPart(nInputs int, uParts []uint32, nParts int) {
	panic("TODO, stubbed")
}

// Abc_ResStartPart2
// C: void Abc_ResStartPart2( int nInputs, unsigned uParts[], int nParts )
func AbcResStartPart2(nInputs int, uParts []uint32, nParts int) {
	panic("TODO, stubbed")
}

// Abc_ResCheckUnique
// C: int Abc_ResCheckUnique( char Pats[], int nPats, int pat )
func AbcResCheckUnique(Pats []int8, nPats int, pat int) int {
	panic("TODO, stubbed")
}

// Abc_ResCheckNonStrict
// C: int Abc_ResCheckNonStrict( char Pattern[], int nVars, int nBits )
func AbcResCheckNonStrict(Pattern []int8, nVars int, nBits int) int {
	panic("TODO, stubbed")
}

// Abc_ResCofCount
// C: int Abc_ResCofCount( DdManager * dd, DdNode * bFunc, unsigned uMask, int * pCheck )
func AbcResCofCount(dd *CuddFoundation, bFunc *DdNode, uMask uint32, pCheck []int) int {
	panic("TODO, stubbed")
}

// Abc_ResCost
// C: int Abc_ResCost( DdManager * dd, DdNode * bFunc, unsigned uMask, int * pnCofs, int * pCheck )
func AbcResCost(dd *CuddFoundation, bFunc *DdNode, uMask uint32, pnCofs *int, pCheck []int) int {
	panic("TODO, stubbed")
}

// Abc_ResMigrate
// C: int Abc_ResMigrate( DdManager * dd, DdNode * bFunc, int nInputs, unsigned uParts[], int iPart1, int iPart2 )
func AbcResMigrate(dd *CuddFoundation, bFunc *DdNode, nInputs int, uParts []uint32, iPart1 int, iPart2 int) int {
	panic("TODO, stubbed")
}

// Abc_ResPrint
// C: void Abc_ResPrint( DdManager * dd, DdNode * bFunc, int nInputs, unsigned uParts[], int nParts )
func AbcResPrint(dd *CuddFoundation, bFunc *DdNode, nInputs int, uParts []uint32, nParts int) {
	panic("TODO, stubbed")
}

// Abc_ResPrintAllCofs
// C: void Abc_ResPrintAllCofs( DdManager * dd, DdNode * bFunc, int nInputs, int nCofMax )
func AbcResPrintAllCofs(dd *CuddFoundation, bFunc *DdNode, nInputs int, nCofMax int) {
	panic("TODO, stubbed")
}

// Abc_ResSwapRandom
// C: void Abc_ResSwapRandom( DdManager * dd, DdNode * bFunc, int nInputs, unsigned uParts[], int nParts, int nTimes )
func AbcResSwapRandom(dd *CuddFoundation, bFunc *DdNode, nInputs int, uParts []uint32, nParts int, nTimes int) {
	panic("TODO, stubbed")
}

// Abc_ResPartition
// C: void Abc_ResPartition( DdManager * dd, DdNode * bFunc, int nInputs )
func AbcResPartition(dd *CuddFoundation, bFunc *DdNode, nInputs int) {
	panic("TODO, stubbed")
}

// Abc_ResPartitionTest
// C: void Abc_ResPartitionTest( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcResPartitionTest() {
	panic("TODO, stubbed")
}

// Abc_NtkBddCofCount
// C: int Abc_NtkBddCofCount( DdManager * dd, DdNode * bFunc, DdNode ** pbVars, int nVars )
func AbcNtkBddCofCount(dd *CuddFoundation, bFunc *DdNode, pbVars []*DdNode, nVars int) int {
	panic("TODO, stubbed")
}

// Abc_NtkExploreCofs2
// C: void Abc_NtkExploreCofs2( DdManager * dd, DdNode * bFunc, DdNode ** pbVars, int nIns, int nLutSize )
func AbcNtkExploreCofs2(dd *CuddFoundation, bFunc *DdNode, pbVars []*DdNode, nIns int, nLutSize int) {
	panic("TODO, stubbed")
}

// Abc_NtkExploreCofs
// C: void Abc_NtkExploreCofs( DdManager * dd, DdNode * bFunc, DdNode ** pbVars, int nIns, int nLutSize )
func AbcNtkExploreCofs(dd *CuddFoundation, bFunc *DdNode, pbVars []*DdNode, nIns int, nLutSize int) {
	panic("TODO, stubbed")
}

// Abc_NtkBddFindAddConst
// C: DdNode * Abc_NtkBddFindAddConst( DdManager * dd, DdNode * bFunc, int nOuts )
func AbcNtkBddFindAddConst(dd *CuddFoundation, bFunc *DdNode, nOuts int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkBddToAdd_rec
// C: DdNode * Abc_NtkBddToAdd_rec( DdManager * dd, DdNode * bFunc, int nOuts, stmm_table * tTable, int fCompl )
func AbcNtkBddToAdd_rec(dd *CuddFoundation, bFunc *DdNode, nOuts int, tTable *StableTable[*DdNode, *DdNode], fCompl int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkBddToAdd
// C: DdNode * Abc_NtkBddToAdd( DdManager * dd, DdNode * bFunc, int nOuts )
func AbcNtkBddToAdd(dd *CuddFoundation, bFunc *DdNode, nOuts int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkAddToBdd_rec
// C: DdNode * Abc_NtkAddToBdd_rec( DdManager * dd, DdNode * aFunc, int nIns, int nOuts, stmm_table * tTable )
func AbcNtkAddToBdd_rec(dd *CuddFoundation, aFunc *DdNode, nIns int, nOuts int, tTable *StableTable[*DdNode, *DdNode]) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkAddToBdd
// C: DdNode * Abc_NtkAddToBdd( DdManager * dd, DdNode * aFunc, int nIns, int nOuts )
func AbcNtkAddToBdd(dd *CuddFoundation, aFunc *DdNode, nIns int, nOuts int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkBddDecCharFunc
// C: DdNode * Abc_NtkBddDecCharFunc( DdManager * dd, DdNode ** pFuncs, int nOuts, int Mask, int nBits )
func AbcNtkBddDecCharFunc(dd *CuddFoundation, pFuncs []*DdNode, nOuts int, Mask int, nBits int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkBddDecTry (pReo is a reo_man *)
// C: DdNode * Abc_NtkBddDecTry( reo_man * pReo, DdManager * dd, DdNode ** pFuncs, int nIns, int nOuts, int Mask, int nBits )
func AbcNtkBddDecTry(pReo *ReoMan, dd *CuddFoundation, pFuncs []*DdNode, nIns int, nOuts int, Mask int, nBits int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkBddDecInt (pReo is a reo_man *)
// C: DdNode * Abc_NtkBddDecInt( reo_man * pReo, DdManager * dd, DdNode ** pFuncs, int nIns, int nOuts )
func AbcNtkBddDecInt(pReo *ReoMan, dd *CuddFoundation, pFuncs []*DdNode, nIns int, nOuts int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkCreateFromCharFunc
// C: Abc_Ntk_t * Abc_NtkCreateFromCharFunc( Abc_Ntk_t * pNtk, DdManager * dd, DdNode * bFunc )
func (pNtk *ABCNetwork) AbcNtkCreateFromCharFunc(dd *CuddFoundation, bFunc *DdNode) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkBddDec
// C: Abc_Ntk_t * Abc_NtkBddDec( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkBddDec(fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}
