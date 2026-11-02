package abc

import "os"

// Hand-checked skeleton of src/base/abci/abcDec.c.

// AbcTtStore is Abc_TtStore_t.
type AbcTtStore struct {
	nVars  int
	nWords int
	nFuncs int
	pFuncs [][]uint64 // word ** : the truth tables
}

// Abc_TruthGetBit (static)
// C: int Abc_TruthGetBit( word * p, int i )
func abcTruthGetBit(p []uint64, i int) int {
	panic("TODO, stubbed")
}

// Abc_TruthSetBit (static)
// C: void Abc_TruthSetBit( word * p, int i )
func abcTruthSetBit(p []uint64, i int) {
	panic("TODO, stubbed")
}

// Abc_TruthXorBit (static)
// C: void Abc_TruthXorBit( word * p, int i )
func abcTruthXorBit(p []uint64, i int) {
	panic("TODO, stubbed")
}

// Abc_TruthGetHex (static)
// C: int Abc_TruthGetHex( word * p, int k )
func abcTruthGetHex(p []uint64, k int) int {
	panic("TODO, stubbed")
}

// Abc_TruthSetHex (static)
// C: void Abc_TruthSetHex( word * p, int k, int d )
func abcTruthSetHex(p []uint64, k int, d int) {
	panic("TODO, stubbed")
}

// Abc_TruthXorHex (static)
// C: void Abc_TruthXorHex( word * p, int k, int d )
func abcTruthXorHex(p []uint64, k int, d int) {
	panic("TODO, stubbed")
}

// Abc_TruthReadHexDigit (static)
// C: int Abc_TruthReadHexDigit( char HexChar )
func abcTruthReadHexDigit(HexChar byte) int {
	panic("TODO, stubbed")
}

// Abc_TruthWriteHexDigit (static)
// C: void Abc_TruthWriteHexDigit( FILE * pFile, int HexDigit )
func abcTruthWriteHexDigit(pFile *os.File, HexDigit int) {
	panic("TODO, stubbed")
}

// Abc_TruthReadHex
// C: void Abc_TruthReadHex( word * pTruth, char * pString, int nVars )
func AbcTruthReadHex(pTruth []uint64, pString string, nVars int) {
	panic("TODO, stubbed")
}

// Abc_TruthWriteHex
// C: void Abc_TruthWriteHex( FILE * pFile, word * pTruth, int nVars )
func AbcTruthWriteHex(pFile *os.File, pTruth []uint64, nVars int) {
	panic("TODO, stubbed")
}

// Abc_TruthStoreAlloc
// C: Abc_TtStore_t * Abc_TruthStoreAlloc( int nVars, int nFuncs )
func AbcTruthStoreAlloc(nVars int, nFuncs int) *AbcTtStore {
	panic("TODO, stubbed")
}

// Abc_TruthStoreAlloc2
// C: Abc_TtStore_t * Abc_TruthStoreAlloc2( int nVars, int nFuncs, word * pBuffer )
func AbcTruthStoreAlloc2(nVars int, nFuncs int, pBuffer []uint64) *AbcTtStore {
	panic("TODO, stubbed")
}

// Abc_TtStoreFree
// C: void Abc_TtStoreFree( Abc_TtStore_t * p, int nVarNum )
func (p *AbcTtStore) AbcTtStoreFree(nVarNum int) {
	panic("TODO, stubbed")
}

// Abc_FileSize
// C: int Abc_FileSize( char * pFileName )
func AbcFileSize(pFileName string) int {
	panic("TODO, stubbed")
}

// Abc_FileRead
// C: char * Abc_FileRead( char * pFileName )
func AbcFileRead(pFileName string) string {
	panic("TODO, stubbed")
}

// Abc_TruthGetParams
// C: void Abc_TruthGetParams( char * pFileName, int * pnVars, int * pnTruths )
func AbcTruthGetParams(pFileName string, pnVars *int, pnTruths *int) {
	panic("TODO, stubbed")
}

// Abc_TruthStoreRead
// C: void Abc_TruthStoreRead( char * pFileName, Abc_TtStore_t * p )
func AbcTruthStoreRead(pFileName string, p *AbcTtStore) {
	panic("TODO, stubbed")
}

// Abc_TtStoreWrite
// C: void Abc_TtStoreWrite( char * pFileName, Abc_TtStore_t * p, int fBinary )
func AbcTtStoreWrite(pFileName string, p *AbcTtStore, fBinary int) {
	panic("TODO, stubbed")
}

// Abc_TtStoreLoad
// C: Abc_TtStore_t * Abc_TtStoreLoad( char * pFileName, int nVarNum )
func AbcTtStoreLoad(pFileName string, nVarNum int) *AbcTtStore {
	panic("TODO, stubbed")
}

// Abc_TtStoreLoadSave
// C: void Abc_TtStoreLoadSave( char * pFileName )
func AbcTtStoreLoadSave(pFileName string) {
	panic("TODO, stubbed")
}

// Abc_TtStoreLoadSaveBin
// C: void Abc_TtStoreLoadSaveBin( char * pFileName )
func AbcTtStoreLoadSaveBin(pFileName string) {
	panic("TODO, stubbed")
}

// Abc_TtStoreTest
// C: void Abc_TtStoreTest( char * pFileName )
func AbcTtStoreTest(pFileName string) {
	panic("TODO, stubbed")
}

// Abc_TruthDecPerform
// C: void Abc_TruthDecPerform( Abc_TtStore_t * p, int DecType, int fVerbose )
func (p *AbcTtStore) AbcTruthDecPerform(DecType int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_TruthDecTest
// C: void Abc_TruthDecTest( char * pFileName, int DecType, int nVarNum, int fVerbose )
func AbcTruthDecTest(pFileName string, DecType int, nVarNum int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_DecTest
// C: int Abc_DecTest( char * pFileName, int DecType, int nVarNum, int fVerbose )
func AbcDecTest(pFileName string, DecType int, nVarNum int, fVerbose int) int {
	panic("TODO, stubbed")
}
