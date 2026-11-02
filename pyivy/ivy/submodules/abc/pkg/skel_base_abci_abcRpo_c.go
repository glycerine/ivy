package abc

import "os"

// Hand-checked skeleton of src/base/abci/abcRpo.c.

// RpoTtStore is Rpo_TtStore_t.
type RpoTtStore struct {
	nVars  int
	nWords int
	nFuncs int
	pFuncs [][]uint64 // word ** : the truth tables
}

// Abc_TruthGetBit (static)
// C: int Abc_TruthGetBit(word * p, int i)
func abcTruthGetBit_rpo(p []uint64, i int) int {
	panic("TODO, stubbed")
}

// Abc_TruthSetBit (static)
// C: void Abc_TruthSetBit(word * p, int i)
func abcTruthSetBit_rpo(p []uint64, i int) {
	panic("TODO, stubbed")
}

// Abc_TruthXorBit (static)
// C: void Abc_TruthXorBit(word * p, int i)
func abcTruthXorBit_rpo(p []uint64, i int) {
	panic("TODO, stubbed")
}

// Abc_TruthGetHex (static)
// C: int Abc_TruthGetHex(word * p, int k)
func abcTruthGetHex_rpo(p []uint64, k int) int {
	panic("TODO, stubbed")
}

// Abc_TruthSetHex (static)
// C: void Abc_TruthSetHex(word * p, int k, int d)
func abcTruthSetHex_rpo(p []uint64, k int, d int) {
	panic("TODO, stubbed")
}

// Abc_TruthXorHex (static)
// C: void Abc_TruthXorHex(word * p, int k, int d)
func abcTruthXorHex_rpo(p []uint64, k int, d int) {
	panic("TODO, stubbed")
}

// Abc_TruthReadHexDigit (static)
// C: int Abc_TruthReadHexDigit(char HexChar)
func abcTruthReadHexDigit_rpo(HexChar byte) int {
	panic("TODO, stubbed")
}

// Abc_TruthWriteHexDigit (static)
// C: void Abc_TruthWriteHexDigit(FILE * pFile, int HexDigit)
func abcTruthWriteHexDigit_rpo(pFile *os.File, HexDigit int) {
	panic("TODO, stubbed")
}

// Abc_TruthReadHex (static)
// C: void Abc_TruthReadHex(word * pTruth, char * pString, int nVars)
func abcTruthReadHex_rpo(pTruth []uint64, pString string, nVars int) {
	panic("TODO, stubbed")
}

// Abc_TruthWriteHex (static)
// C: void Abc_TruthWriteHex(FILE * pFile, word * pTruth, int nVars)
func abcTruthWriteHex_rpo(pFile *os.File, pTruth []uint64, nVars int) {
	panic("TODO, stubbed")
}

// Abc_TruthStoreAlloc (static)
// C: Rpo_TtStore_t * Abc_TruthStoreAlloc(int nVars, int nFuncs)
func abcTruthStoreAlloc_rpo(nVars int, nFuncs int) *RpoTtStore {
	panic("TODO, stubbed")
}

// Abc_TruthStoreAlloc2 (static)
// C: Rpo_TtStore_t * Abc_TruthStoreAlloc2(int nVars, int nFuncs, word * pBuffer)
func abcTruthStoreAlloc2_rpo(nVars int, nFuncs int, pBuffer []uint64) *RpoTtStore {
	panic("TODO, stubbed")
}

// Abc_TtStoreFree (static)
// C: void Abc_TtStoreFree(Rpo_TtStore_t * p, int nVarNum)
func (p *RpoTtStore) abcTtStoreFree_rpo(nVarNum int) {
	panic("TODO, stubbed")
}

// Abc_TruthStoreRead (static)
// C: void Abc_TruthStoreRead(char * pFileName, Rpo_TtStore_t * p)
func abcTruthStoreRead_rpo(pFileName string, p *RpoTtStore) {
	panic("TODO, stubbed")
}

// Abc_TtStoreWrite (static)
// C: void Abc_TtStoreWrite(char * pFileName, Rpo_TtStore_t * p, int fBinary)
func abcTtStoreWrite_rpo(pFileName string, p *RpoTtStore, fBinary int) {
	panic("TODO, stubbed")
}

// Abc_TtStoreLoad (static)
// C: Rpo_TtStore_t * Abc_TtStoreLoad(char * pFileName, int nVarNum)
func abcTtStoreLoad_rpo(pFileName string, nVarNum int) *RpoTtStore {
	panic("TODO, stubbed")
}

// Abc_TruthRpoPerform
// C: void Abc_TruthRpoPerform(Rpo_TtStore_t * p, int nThreshold, int fVerbose)
func (p *RpoTtStore) AbcTruthRpoPerform(nThreshold int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_TruthRpoTest
// C: void Abc_TruthRpoTest(char * pFileName, int nVarNum, int nThreshold, int fVerbose)
func AbcTruthRpoTest(pFileName string, nVarNum int, nThreshold int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_RpoTest
// C: int Abc_RpoTest(char * pFileName, int nVarNum,int nThreshold, int fVerbose)
func AbcRpoTest(pFileName string, nVarNum int, nThreshold int, fVerbose int) int {
	panic("TODO, stubbed")
}
