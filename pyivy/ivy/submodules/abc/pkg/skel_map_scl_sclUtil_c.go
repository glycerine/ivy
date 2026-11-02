package abc

// Hand-checked skeleton of src/map/scl/sclUtil.c.

// Abc_SclMioGates2SclGates
// C: void Abc_SclMioGates2SclGates( SC_Lib * pLib, Abc_Ntk_t * p )
func (pLib *SCLib) AbcSclMioGates2SclGates(p *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_SclSclGates2MioGates
// C: void Abc_SclSclGates2MioGates( SC_Lib * pLib, Abc_Ntk_t * p )
func (pLib *SCLib) AbcSclSclGates2MioGates(p *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_SclTransferGates
// C: void Abc_SclTransferGates( Abc_Ntk_t * pOld, Abc_Ntk_t * pNew )
func AbcSclTransferGates(pOld *ABCNetwork, pNew *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_SclManPrintGateSizes
// C: void Abc_SclManPrintGateSizes( SC_Lib * pLib, Abc_Ntk_t * p, Vec_Int_t * vGates )
func (pLib *SCLib) AbcSclManPrintGateSizes(p *ABCNetwork, vGates *IntVector) {
	panic("TODO, stubbed")
}

// Abc_SclPrintGateSizes
// C: void Abc_SclPrintGateSizes( SC_Lib * pLib, Abc_Ntk_t * p )
func (pLib *SCLib) AbcSclPrintGateSizes(p *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_SclFindMaxAreaCell
// C: SC_Cell * Abc_SclFindMaxAreaCell( SC_Cell * pRepr )
func AbcSclFindMaxAreaCell(pRepr *SCCell) *SCCell {
	panic("TODO, stubbed")
}

// Abc_SclFindMinAreas
// C: Vec_Int_t * Abc_SclFindMinAreas( SC_Lib * pLib, int fUseMax )
func (pLib *SCLib) AbcSclFindMinAreas(fUseMax int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_SclMinsizePerform
// C: void Abc_SclMinsizePerform( SC_Lib * pLib, Abc_Ntk_t * p, int fUseMax, int fVerbose )
func (pLib *SCLib) AbcSclMinsizePerform(p *ABCNetwork, fUseMax int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_SclCountMinSize
// C: int Abc_SclCountMinSize( SC_Lib * pLib, Abc_Ntk_t * p, int fUseMax )
func (pLib *SCLib) AbcSclCountMinSize(p *ABCNetwork, fUseMax int) int {
	panic("TODO, stubbed")
}

// Abc_SclReadTimingConstr
// C: void Abc_SclReadTimingConstr( Abc_Frame_t * pAbc, char * pFileName, int fVerbose )
func (pAbc *MainFrame) AbcSclReadTimingConstr(pFileName string, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_SclExtractBarBufs
// C: Vec_Int_t * Abc_SclExtractBarBufs( Abc_Ntk_t * pNtk )
func AbcSclExtractBarBufs(pNtk *ABCNetwork) *IntVector {
	panic("TODO, stubbed")
}

// Abc_SclInsertBarBufs
// C: void Abc_SclInsertBarBufs( Abc_Ntk_t * pNtk, Vec_Int_t * vBufs )
func AbcSclInsertBarBufs(pNtk *ABCNetwork, vBufs *IntVector) {
	panic("TODO, stubbed")
}
