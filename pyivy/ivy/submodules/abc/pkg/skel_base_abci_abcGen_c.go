package abc

import "os"

// Hand-checked skeleton of src/base/abci/abcGen.c.

// Abc_WriteFullAdder
// C: void Abc_WriteFullAdder( FILE * pFile )
func AbcWriteFullAdder(pFile *os.File) {
	panic("TODO, stubbed")
}

// Abc_WriteAdder
// C: void Abc_WriteAdder( FILE * pFile, int nVars )
func AbcWriteAdder(pFile *os.File, nVars int) {
	panic("TODO, stubbed")
}

// Abc_GenAdder
// C: void Abc_GenAdder( char * pFileName, int nVars )
func AbcGenAdder(pFileName string, nVars int) {
	panic("TODO, stubbed")
}

// Abc_WriteMulti
// C: void Abc_WriteMulti( FILE * pFile, int nVars )
func AbcWriteMulti(pFile *os.File, nVars int) {
	panic("TODO, stubbed")
}

// Abc_GenMulti
// C: void Abc_GenMulti( char * pFileName, int nVars )
func AbcGenMulti(pFileName string, nVars int) {
	panic("TODO, stubbed")
}

// Abc_WriteComp
// C: void Abc_WriteComp( FILE * pFile )
func AbcWriteComp(pFile *os.File) {
	panic("TODO, stubbed")
}

// Abc_WriteLayer
// C: void Abc_WriteLayer( FILE * pFile, int nVars, int fSkip1 )
func AbcWriteLayer(pFile *os.File, nVars int, fSkip1 int) {
	panic("TODO, stubbed")
}

// Abc_GenSorter
// C: void Abc_GenSorter( char * pFileName, int nVars )
func AbcGenSorter(pFileName string, nVars int) {
	panic("TODO, stubbed")
}

// Abc_WriteCell
// C: void Abc_WriteCell( FILE * pFile )
func AbcWriteCell(pFile *os.File) {
	panic("TODO, stubbed")
}

// Abc_GenMesh
// C: void Abc_GenMesh( char * pFileName, int nVars )
func AbcGenMesh(pFileName string, nVars int) {
	panic("TODO, stubbed")
}

// Abc_WriteKLut
// C: void Abc_WriteKLut( FILE * pFile, int nLutSize )
func AbcWriteKLut(pFile *os.File, nLutSize int) {
	panic("TODO, stubbed")
}

// Abc_GenFpga
// C: void Abc_GenFpga( char * pFileName, int nLutSize, int nLuts, int nVars )
func AbcGenFpga(pFileName string, nLutSize int, nLuts int, nVars int) {
	panic("TODO, stubbed")
}

// Abc_GenOneHot
// C: void Abc_GenOneHot( char * pFileName, int nVars )
func AbcGenOneHot(pFileName string, nVars int) {
	panic("TODO, stubbed")
}

// Abc_GenOneHotIntervals
// C: void Abc_GenOneHotIntervals( char * pFileName, int nPis, int nRegs, Vec_Ptr_t * vOnehots )
func AbcGenOneHotIntervals(pFileName string, nPis int, nRegs int, vOnehots *PointerVector[IntVector]) {
	panic("TODO, stubbed")
}

// Abc_GenRandom
// C: void Abc_GenRandom( char * pFileName, int nPis )
func AbcGenRandom(pFileName string, nPis int) {
	panic("TODO, stubbed")
}

// Abc_GenFsmCond
// C: void Abc_GenFsmCond( Vec_Str_t * vCond, int nPis, int Prob )
func AbcGenFsmCond(vCond *ByteVector, nPis int, Prob int) {
	panic("TODO, stubbed")
}

// Abc_GenFsm
// C: void Abc_GenFsm( char * pFileName, int nPis, int nPos, int nStates, int nLines, int ProbI, int ProbO )
func AbcGenFsm(pFileName string, nPis int, nPos int, nStates int, nLines int, ProbI int, ProbO int) {
	panic("TODO, stubbed")
}

// Abc_AdderTree
// C: void Abc_AdderTree( FILE * pFile, int nArgs, int nBits )
func AbcAdderTree(pFile *os.File, nArgs int, nBits int) {
	panic("TODO, stubbed")
}

// Abc_GenAdderTree
// C: void Abc_GenAdderTree( char * pFileName, int nArgs, int nBits )
func AbcGenAdderTree(pFileName string, nArgs int, nBits int) {
	panic("TODO, stubbed")
}

// Abc_GenSignedBoothPP
// C: int Abc_GenSignedBoothPP( Gia_Man_t * p, int a, int b, int c, int d, int e )
func (p *GiaManager) AbcGenSignedBoothPP(a int, b int, c int, d int, e int) int {
	panic("TODO, stubbed")
}

// Abc_GenSignedBoothPPTest
// C: Gia_Man_t * Abc_GenSignedBoothPPTest( int nArgA, int nArgB )
func AbcGenSignedBoothPPTest(nArgA int, nArgB int) *GiaManager {
	panic("TODO, stubbed")
}

// Abc_GenSignedBooth
// C: Gia_Man_t * Abc_GenSignedBooth( int nArgN, int nArgM )
func AbcGenSignedBooth(nArgN int, nArgM int) *GiaManager {
	panic("TODO, stubbed")
}

// Abc_GenSignedBoothMini
// C: Mini_Aig_t * Abc_GenSignedBoothMini( int nArgN, int nArgM )
func AbcGenSignedBoothMini(nArgN int, nArgM int) *MiniAIG {
	panic("TODO, stubbed")
}

// Abc_WriteBoothPartialProducts
// C: void Abc_WriteBoothPartialProducts( FILE * pFile, int nVars )
func AbcWriteBoothPartialProducts(pFile *os.File, nVars int) {
	panic("TODO, stubbed")
}

// Abc_WriteBooth
// C: void Abc_WriteBooth( FILE * pFile, int nVars )
func AbcWriteBooth(pFile *os.File, nVars int) {
	panic("TODO, stubbed")
}

// Abc_GenBooth
// C: void Abc_GenBooth( char * pFileName, int nVars )
func AbcGenBooth(pFileName string, nVars int) {
	panic("TODO, stubbed")
}
