package abc

// Hand-checked skeleton of src/map/fpga/fpga.c.

// Fpga_Init
// C: void Fpga_Init( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) FpgaInit() {
	panic("TODO, stubbed")
}

// Fpga_End
// C: void Fpga_End( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) FpgaEnd() {
	panic("TODO, stubbed")
}

// Fpga_CommandReadLibrary (static)
// C: int Fpga_CommandReadLibrary( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) fpgaCommandReadLibrary(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Fpga_CommandPrintLibrary (static)
// C: int Fpga_CommandPrintLibrary( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) fpgaCommandPrintLibrary(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Fpga_SetSimpleLutLib
// C: void Fpga_SetSimpleLutLib( int nLutSize )
func FpgaSetSimpleLutLib(nLutSize int) {
	panic("TODO, stubbed")
}
