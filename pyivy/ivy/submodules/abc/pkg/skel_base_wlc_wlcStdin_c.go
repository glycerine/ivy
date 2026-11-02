package abc

// Hand-checked skeleton of src/base/wlc/wlcStdin.c.

// Wlc_ComputeSum (pRes is updated in place)
// C: void Wlc_ComputeSum( char * pRes, char * pAdd, int nBits, int Radix )
func WlcComputeSum(pRes []int8, pAdd []int8, nBits int, Radix int) {
	panic("TODO, stubbed")
}

// Wlc_ConvertToRadix
// C: Vec_Str_t * Wlc_ConvertToRadix( unsigned * pBits, int Start, int nBits, int Radix )
func WlcConvertToRadix(pBits []uint32, Start int, nBits int, Radix int) *ByteVector {
	panic("TODO, stubbed")
}

// Wlc_NtkReport
// C: void Wlc_NtkReport( Wlc_Ntk_t * p, Abc_Cex_t * pCex, char * pName, int Radix )
func WlcNtkReport(p *WLCNetwork, pCex *Counterexample, pName string, Radix int) {
	panic("TODO, stubbed")
}

// Wlc_StdinCollectStop (static)
// C: int Wlc_StdinCollectStop( Vec_Str_t * vInput, char * pLine, int LineSize )
func wlcStdinCollectStop(vInput *ByteVector, pLine string, LineSize int) int {
	panic("TODO, stubbed")
}

// Wlc_StdinCollectProblem
// C: Vec_Str_t * Wlc_StdinCollectProblem( char * pDir )
func WlcStdinCollectProblem(pDir string) *ByteVector {
	panic("TODO, stubbed")
}

// Wlc_StdinCollectQuery
// C: Vec_Str_t * Wlc_StdinCollectQuery()
func WlcStdinCollectQuery() *ByteVector {
	panic("TODO, stubbed")
}

// Wlc_StdinProcessSmt
// C: int Wlc_StdinProcessSmt( Abc_Frame_t * pAbc, char * pCmd )
func (pAbc *MainFrame) WlcStdinProcessSmt(pCmd string) int {
	panic("TODO, stubbed")
}
