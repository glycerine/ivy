package abc

// Hand-checked skeleton of src/map/mio/mio.c.

// Mio_Init
// C: void Mio_Init( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) MioInit() {
	panic("TODO, stubbed")
}

// Mio_End
// C: void Mio_End( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) MioEnd() {
	panic("TODO, stubbed")
}

// Mio_UpdateGenlib
// C: void Mio_UpdateGenlib( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioUpdateGenlib() {
	panic("TODO, stubbed")
}

// Mio_UpdateGenlib2
// C: int Mio_UpdateGenlib2( Vec_Str_t * vStr, Vec_Str_t * vStr2, char * pFileName, int fVerbose )
func MioUpdateGenlib2(vStr *ByteVector, vStr2 *ByteVector, pFileName string, fVerbose int) int {
	panic("TODO, stubbed")
}

// Mio_CommandReadLiberty (static)
// C: int Mio_CommandReadLiberty( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) mioCommandReadLiberty(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Mio_CommandReadGenlib (static)
// C: int Mio_CommandReadGenlib( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) mioCommandReadGenlib(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Mio_CommandWriteGenlib (static)
// C: int Mio_CommandWriteGenlib( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) mioCommandWriteGenlib(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Mio_CommandPrintGenlib (static)
// C: int Mio_CommandPrintGenlib( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) mioCommandPrintGenlib(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Mio_CommandReadProfile (static)
// C: int Mio_CommandReadProfile( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) mioCommandReadProfile(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Mio_CommandWriteProfile (static)
// C: int Mio_CommandWriteProfile( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) mioCommandWriteProfile(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Mio_CommandPrintProfile (static)
// C: int Mio_CommandPrintProfile( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) mioCommandPrintProfile(argc int, argv []string) int {
	panic("TODO, stubbed")
}
