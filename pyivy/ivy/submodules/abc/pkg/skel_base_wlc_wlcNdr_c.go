package abc

// Hand-checked skeleton of src/base/wlc/wlcNdr.c.

// Ndr_TypeNdr2Wlc
// C: int Ndr_TypeNdr2Wlc( int Type )
func NdrTypeNdr2Wlc(Type int) int {
	panic("TODO, stubbed")
}

// Ndr_TypeWlc2Ndr
// C: int Ndr_TypeWlc2Ndr( int Type )
func NdrTypeWlc2Ndr(Type int) int {
	panic("TODO, stubbed")
}

// Ndr_ObjWriteConstant
// C: char * Ndr_ObjWriteConstant( unsigned * pBits, int nBits )
func NdrObjWriteConstant(pBits []uint32, nBits int) string {
	panic("TODO, stubbed")
}

// Wlc_NtkToNdr (returns a void * that holds Ndr_Data_t)
// C: void * Wlc_NtkToNdr( Wlc_Ntk_t * pNtk )
func WlcNtkToNdr(pNtk *WLCNetwork) *NDRDesignData {
	panic("TODO, stubbed")
}

// Wlc_WriteNdr
// C: void Wlc_WriteNdr( Wlc_Ntk_t * pNtk, char * pFileName )
func WlcWriteNdr(pNtk *WLCNetwork, pFileName string) {
	panic("TODO, stubbed")
}

// Wlc_NtkToNdrTest
// C: void Wlc_NtkToNdrTest( Wlc_Ntk_t * pNtk )
func WlcNtkToNdrTest(pNtk *WLCNetwork) {
	panic("TODO, stubbed")
}

// Ndr_ObjReadRange (End and Beg are out parameters)
// C: int Ndr_ObjReadRange( Ndr_Data_t * p, int Obj, int * End, int * Beg )
func NdrObjReadRange(p *NDRDesignData, Obj int, End *int, Beg *int) int {
	panic("TODO, stubbed")
}

// Ndr_ObjReadConstant
// C: void Ndr_ObjReadConstant( Vec_Int_t * vFanins, char * pStr )
func NdrObjReadConstant(vFanins *IntVector, pStr string) {
	panic("TODO, stubbed")
}

// Ndr_NtkPrintNodes
// C: void Ndr_NtkPrintNodes( Wlc_Ntk_t * pNtk )
func NdrNtkPrintNodes(pNtk *WLCNetwork) {
	panic("TODO, stubbed")
}

// Wlc_NtkCheckIntegrity (pData is a void * that holds Ndr_Data_t)
// C: void Wlc_NtkCheckIntegrity( void * pData )
func WlcNtkCheckIntegrity(pData *NDRDesignData) {
	panic("TODO, stubbed")
}

// Wlc_NtkFromNdr (pData is a void * that holds Ndr_Data_t)
// C: Wlc_Ntk_t * Wlc_NtkFromNdr( void * pData )
func WlcNtkFromNdr(pData *NDRDesignData) *WLCNetwork {
	panic("TODO, stubbed")
}

// Wlc_ReadNdr
// C: Wlc_Ntk_t * Wlc_ReadNdr( char * pFileName )
func WlcReadNdr(pFileName string) *WLCNetwork {
	panic("TODO, stubbed")
}

// Wlc_ReadNdrTest
// C: void Wlc_ReadNdrTest()
func WlcReadNdrTest() {
	panic("TODO, stubbed")
}
