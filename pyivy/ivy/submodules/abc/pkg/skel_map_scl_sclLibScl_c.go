package abc

import "os"

// Hand-checked skeleton of src/map/scl/sclLibScl.c.

// Abc_SclReadSurfaceGenlib (static)
// C: void Abc_SclReadSurfaceGenlib( SC_Surface * p )
func abcSclReadSurfaceGenlib(p *SCSurface) {
	panic("TODO, stubbed")
}

// Abc_SclReadLibraryGenlib (static)
// C: int Abc_SclReadLibraryGenlib( SC_Lib * p, Mio_Library_t * pLib )
func (p *SCLib) abcSclReadLibraryGenlib(pLib *MioLibrary) int {
	panic("TODO, stubbed")
}

// Abc_SclReadFromGenlib (pLib0 is a void * that holds a Mio_Library_t *)
// C: SC_Lib * Abc_SclReadFromGenlib( void * pLib0 )
func AbcSclReadFromGenlib(pLib0 *MioLibrary) *SCLib {
	panic("TODO, stubbed")
}

// Abc_SclReadSurface (static; pPos is an in/out position)
// C: void Abc_SclReadSurface( Vec_Str_t * vOut, int * pPos, SC_Surface * p )
func abcSclReadSurface(vOut *ByteVector, pPos *int, p *SCSurface) {
	panic("TODO, stubbed")
}

// Abc_SclReadLibrary (static; pPos is an in/out position)
// C: int Abc_SclReadLibrary( Vec_Str_t * vOut, int * pPos, SC_Lib * p )
func abcSclReadLibrary(vOut *ByteVector, pPos *int, p *SCLib) int {
	panic("TODO, stubbed")
}

// Abc_SclReadFromStr
// C: SC_Lib * Abc_SclReadFromStr( Vec_Str_t * vOut )
func AbcSclReadFromStr(vOut *ByteVector) *SCLib {
	panic("TODO, stubbed")
}

// Abc_SclReadFromFile
// C: SC_Lib * Abc_SclReadFromFile( char * pFileName )
func AbcSclReadFromFile(pFileName string) *SCLib {
	panic("TODO, stubbed")
}

// Abc_SclWriteSurface (static)
// C: void Abc_SclWriteSurface( Vec_Str_t * vOut, SC_Surface * p )
func abcSclWriteSurface(vOut *ByteVector, p *SCSurface) {
	panic("TODO, stubbed")
}

// Abc_SclWriteLibrary (static)
// C: void Abc_SclWriteLibrary( Vec_Str_t * vOut, SC_Lib * p )
func abcSclWriteLibrary(vOut *ByteVector, p *SCLib) {
	panic("TODO, stubbed")
}

// Abc_SclWriteScl
// C: void Abc_SclWriteScl( char * pFileName, SC_Lib * p )
func AbcSclWriteScl(pFileName string, p *SCLib) {
	panic("TODO, stubbed")
}

// Abc_SclWriteSurfaceText (static)
// C: void Abc_SclWriteSurfaceText( FILE * s, SC_Surface * p )
func abcSclWriteSurfaceText(s *os.File, p *SCSurface) {
	panic("TODO, stubbed")
}

// Abc_SclWriteLibraryText (static)
// C: void Abc_SclWriteLibraryText( FILE * s, SC_Lib * p )
func abcSclWriteLibraryText(s *os.File, p *SCLib) {
	panic("TODO, stubbed")
}

// Abc_SclWriteLiberty
// C: void Abc_SclWriteLiberty( char * pFileName, SC_Lib * p )
func AbcSclWriteLiberty(pFileName string, p *SCLib) {
	panic("TODO, stubbed")
}
