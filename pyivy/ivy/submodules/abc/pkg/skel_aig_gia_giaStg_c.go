package abc

import "os"

// Hand-checked skeleton of src/aig/gia/giaStg.c.

// Gia_ManPrintStateEncoding
// C: void Gia_ManPrintStateEncoding( Vec_Vec_t * vCodes, int nBits )
func GiaManPrintStateEncoding(vCodes *NestedVector[int], nBits int) {
	panic("TODO, stubbed")
}

// Gia_ManCreateOrGate
// C: int Gia_ManCreateOrGate( Gia_Man_t * p, Vec_Int_t * vLits )
func (p *GiaManager) GiaManCreateOrGate(vLits *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManAssignCodes
// C: Vec_Vec_t * Gia_ManAssignCodes( int kHot, int nStates, int * pnBits )
func GiaManAssignCodes(kHot int, nStates int, pnBits *int) *NestedVector[int] {
	panic("TODO, stubbed")
}

// Gia_ManStgKHot
// C: Gia_Man_t * Gia_ManStgKHot( Vec_Int_t * vLines, int nIns, int nOuts, int nStates, int kHot, int fVerbose )
func GiaManStgKHot(vLines *IntVector, nIns int, nOuts int, nStates int, kHot int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManStgOneHot
// C: Gia_Man_t * Gia_ManStgOneHot( Vec_Int_t * vLines, int nIns, int nOuts, int nStates )
func GiaManStgOneHot(vLines *IntVector, nIns int, nOuts int, nStates int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManStgPrint
// C: void Gia_ManStgPrint( FILE * pFile, Vec_Int_t * vLines, int nIns, int nOuts, int nStates )
func GiaManStgPrint(pFile *os.File, vLines *IntVector, nIns int, nOuts int, nStates int) {
	panic("TODO, stubbed")
}

// Gia_ManStgReadLines
// C: Vec_Int_t * Gia_ManStgReadLines( char * pFileName, int * pnIns, int * pnOuts, int * pnStates )
func GiaManStgReadLines(pFileName string, pnIns *int, pnOuts *int, pnStates *int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManStgRead
// C: Gia_Man_t * Gia_ManStgRead( char * pFileName, int kHot, int fVerbose )
func GiaManStgRead(pFileName string, kHot int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
