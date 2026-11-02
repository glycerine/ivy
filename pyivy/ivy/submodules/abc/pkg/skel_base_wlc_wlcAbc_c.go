package abc

// Hand-checked skeleton of src/base/wlc/wlcAbc.c.

// Wlc_NtkPrintInputInfo
// C: void Wlc_NtkPrintInputInfo( Wlc_Ntk_t * pNtk )
func WlcNtkPrintInputInfo(pNtk *WLCNetwork) {
	panic("TODO, stubbed")
}

// Wlc_NtkPrintInvStats
// C: void Wlc_NtkPrintInvStats( Wlc_Ntk_t * pNtk, Vec_Int_t * vCounts, int fVerbose )
func WlcNtkPrintInvStats(pNtk *WLCNetwork, vCounts *IntVector, fVerbose int) {
	panic("TODO, stubbed")
}

// Wlc_NtkGetInv
// C: Abc_Ntk_t * Wlc_NtkGetInv( Wlc_Ntk_t * pNtk, Vec_Int_t * vInv )
func WlcNtkGetInv(pNtk *WLCNetwork, vInv *IntVector) *ABCNetwork {
	panic("TODO, stubbed")
}

// Wlc_NtkGetPut
// C: Vec_Int_t * Wlc_NtkGetPut( Abc_Ntk_t * pNtk, Gia_Man_t * pGia )
func WlcNtkGetPut(pNtk *ABCNetwork, pGia *GiaManager) *IntVector {
	panic("TODO, stubbed")
}
