package abc

// Hand-checked skeleton of src/map/mapper/mapperLib.c.

// Map_SuperLibCreate
// C: Map_SuperLib_t * Map_SuperLibCreate( Mio_Library_t * pGenlib, Vec_Str_t * vStr, char * pFileName, char * pExcludeFile, int fAlgorithm, int fVerbose )
func (pGenlib *MioLibrary) MapSuperLibCreate(vStr *ByteVector, pFileName string, pExcludeFile string, fAlgorithm int, fVerbose int) *MapSuperLib {
	panic("TODO, stubbed")
}

// Map_SuperLibFree
// C: void Map_SuperLibFree( Map_SuperLib_t * p )
func (p *MapSuperLib) MapSuperLibFree() {
	panic("TODO, stubbed")
}

// Map_SuperLibDeriveFromGenlib
// C: int Map_SuperLibDeriveFromGenlib( Mio_Library_t * pLib, int fVerbose )
func (pLib *MioLibrary) MapSuperLibDeriveFromGenlib(fVerbose int) int {
	panic("TODO, stubbed")
}

// Map_SuperLibDeriveFromGenlib2
// C: int Map_SuperLibDeriveFromGenlib2( Mio_Library_t * pLib, int fVerbose )
func (pLib *MioLibrary) MapSuperLibDeriveFromGenlib2(fVerbose int) int {
	panic("TODO, stubbed")
}
