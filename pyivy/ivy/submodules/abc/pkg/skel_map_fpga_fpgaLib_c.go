package abc

// Hand-checked skeleton of src/map/fpga/fpgaLib.c.

// Fpga_LutLibReadVarMax
// C: int Fpga_LutLibReadVarMax( Fpga_LutLib_t * p )
func (p *FpgaLutLib) FpgaLutLibReadVarMax() int {
	panic("TODO, stubbed")
}

// Fpga_LutLibReadLutAreas
// C: float * Fpga_LutLibReadLutAreas( Fpga_LutLib_t * p )
func (p *FpgaLutLib) FpgaLutLibReadLutAreas() []float32 {
	panic("TODO, stubbed")
}

// Fpga_LutLibReadLutArea
// C: float Fpga_LutLibReadLutArea( Fpga_LutLib_t * p, int Size )
func (p *FpgaLutLib) FpgaLutLibReadLutArea(Size int) float32 {
	panic("TODO, stubbed")
}

// Fpga_LutLibRead
// C: Fpga_LutLib_t * Fpga_LutLibRead( char * FileName, int fVerbose )
func FpgaLutLibRead(FileName string, fVerbose int) *FpgaLutLib {
	panic("TODO, stubbed")
}

// Fpga_LutLibDup
// C: Fpga_LutLib_t * Fpga_LutLibDup( Fpga_LutLib_t * p )
func (p *FpgaLutLib) FpgaLutLibDup() *FpgaLutLib {
	panic("TODO, stubbed")
}

// Fpga_LutLibFree
// C: void Fpga_LutLibFree( Fpga_LutLib_t * pLutLib )
func (pLutLib *FpgaLutLib) FpgaLutLibFree() {
	panic("TODO, stubbed")
}

// Fpga_LutLibPrint
// C: void Fpga_LutLibPrint( Fpga_LutLib_t * pLutLib )
func (pLutLib *FpgaLutLib) FpgaLutLibPrint() {
	panic("TODO, stubbed")
}

// Fpga_LutLibDelaysAreDiscrete
// C: int Fpga_LutLibDelaysAreDiscrete( Fpga_LutLib_t * pLutLib )
func (pLutLib *FpgaLutLib) FpgaLutLibDelaysAreDiscrete() int {
	panic("TODO, stubbed")
}
