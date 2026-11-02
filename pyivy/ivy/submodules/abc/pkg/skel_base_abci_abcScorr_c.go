package abc

import "os"

// Hand-checked skeleton of src/base/abci/abcScorr.c.

// TstDat is Tst_Dat_t.
type TstDat struct {
	pNetlist     *ABCNetwork
	pAig         *AigManager
	pGia         *GiaManager
	vId2Name     *IntVector
	pFileNameOut string // char *
	fFlopOnly    int
	fFfNdOnly    int
	fDumpBmc     int
}

// Abc_NtkMapGiaIntoNameId
// C: Vec_Int_t * Abc_NtkMapGiaIntoNameId( Abc_Ntk_t * pNetlist, Aig_Man_t * pAig, Gia_Man_t * pGia )
func (pNetlist *ABCNetwork) AbcNtkMapGiaIntoNameId(pAig *AigManager, pGia *GiaManager) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkTestScorrGetName
// C: char * Abc_NtkTestScorrGetName( Abc_Ntk_t * pNetlist, Vec_Int_t * vId2Name, int Id )
func (pNetlist *ABCNetwork) AbcNtkTestScorrGetName(vId2Name *IntVector, Id int) string {
	panic("TODO, stubbed")
}

// Abc_NtkTestScorrWriteEquivPair
// C: int Abc_NtkTestScorrWriteEquivPair( Abc_Ntk_t * pNetlist, Vec_Int_t * vId2Name, int Id1, int Id2, FILE * pFile, int fPol )
func (pNetlist *ABCNetwork) AbcNtkTestScorrWriteEquivPair(vId2Name *IntVector, Id1 int, Id2 int, pFile *os.File, fPol int) int {
	panic("TODO, stubbed")
}

// Abc_NtkTestScorrWriteEquivConst
// C: int Abc_NtkTestScorrWriteEquivConst( Abc_Ntk_t * pNetlist, Vec_Int_t * vId2Name, int Id1, FILE * pFile, int fPol )
func (pNetlist *ABCNetwork) AbcNtkTestScorrWriteEquivConst(vId2Name *IntVector, Id1 int, pFile *os.File, fPol int) int {
	panic("TODO, stubbed")
}

// Abc_NtkBmcFileName
// C: char * Abc_NtkBmcFileName( char * pName )
func AbcNtkBmcFileName(pName string) string {
	panic("TODO, stubbed")
}

// Abc_NtkTestScorrWriteEquivGia
// C: int Abc_NtkTestScorrWriteEquivGia( Tst_Dat_t * pData )
func AbcNtkTestScorrWriteEquivGia(pData *TstDat) int {
	panic("TODO, stubbed")
}

// Abc_NtkTestScorrWriteEquivAig
// C: int Abc_NtkTestScorrWriteEquivAig( Tst_Dat_t * pData )
func AbcNtkTestScorrWriteEquivAig(pData *TstDat) int {
	panic("TODO, stubbed")
}

// Abc_NtkTestScorr
// C: Abc_Ntk_t * Abc_NtkTestScorr( char * pFileNameIn, char * pFileNameOut, int nStepsMax, int nBTLimit, int fNewAlgo, int fFlopOnly, int fFfNdOnly, int fVerbose )
func AbcNtkTestScorr(pFileNameIn string, pFileNameOut string, nStepsMax int, nBTLimit int, fNewAlgo int, fFlopOnly int, fFfNdOnly int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}
