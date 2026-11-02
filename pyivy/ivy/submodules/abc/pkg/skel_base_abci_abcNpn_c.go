package abc

// Hand-checked skeleton of src/base/abci/abcNpn.c.

// Abc_TruthHashKey (static)
// C: int Abc_TruthHashKey( word * pFunc, int nWords, int nTableSize )
func abcTruthHashKey(pFunc []uint64, nWords int, nTableSize int) int {
	panic("TODO, stubbed")
}

// Abc_TruthHashLookup (static)
// C: int Abc_TruthHashLookup( word ** pFuncs, int iThis, int nWords, int * pTable, int * pNexts, int Key )
func abcTruthHashLookup(pFuncs [][]uint64, iThis int, nWords int, pTable []int, pNexts []int, Key int) int {
	panic("TODO, stubbed")
}

// Abc_TruthNpnCountUnique
// C: int Abc_TruthNpnCountUnique( Abc_TtStore_t * p )
func (p *AbcTtStore) AbcTruthNpnCountUnique() int {
	panic("TODO, stubbed")
}

// Abc_TruthCompare
// C: int Abc_TruthCompare( word ** p1, word ** p2 )
func AbcTruthCompare(p1 *[]uint64, p2 *[]uint64) int {
	panic("TODO, stubbed")
}

// Abc_TruthNpnCountUniqueSort
// C: int Abc_TruthNpnCountUniqueSort( Abc_TtStore_t * p )
func (p *AbcTtStore) AbcTruthNpnCountUniqueSort() int {
	panic("TODO, stubbed")
}

// Abc_TruthNpnPrint
// C: void Abc_TruthNpnPrint( char * pCanonPermInit, unsigned uCanonPhase, int nVars )
func AbcTruthNpnPrint(pCanonPermInit []int8, uCanonPhase uint32, nVars int) {
	panic("TODO, stubbed")
}

// Abc_TruthNpnPerform
// C: void Abc_TruthNpnPerform( Abc_TtStore_t * p, int NpnType, int fVerbose )
func (p *AbcTtStore) AbcTruthNpnPerform(NpnType int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_TruthNpnTest
// C: void Abc_TruthNpnTest( char * pFileName, int NpnType, int nVarNum, int fDumpRes, int fBinary, int fVerbose )
func AbcTruthNpnTest(pFileName string, NpnType int, nVarNum int, fDumpRes int, fBinary int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NpnTest
// C: int Abc_NpnTest( char * pFileName, int NpnType, int nVarNum, int fDumpRes, int fBinary, int fVerbose )
func AbcNpnTest(pFileName string, NpnType int, nVarNum int, fDumpRes int, fBinary int, fVerbose int) int {
	panic("TODO, stubbed")
}
