package abc

import "os"

// Hand-checked skeleton of src/aig/gia/giaAiger.c.

// Gia_FileFixName (edits the name in place)
// C: void Gia_FileFixName( char * pFileName )
func GiaFileFixName(pFileName *string) {
	panic("TODO, stubbed")
}

// Gia_FileNameGeneric
// C: char * Gia_FileNameGeneric( char * FileName )
func GiaFileNameGeneric(FileName string) string {
	panic("TODO, stubbed")
}

// Gia_FileSize
// C: int Gia_FileSize( char * pFileName )
func GiaFileSize(pFileName string) int {
	panic("TODO, stubbed")
}

// Gia_FileWriteBufferSize
// C: void Gia_FileWriteBufferSize( FILE * pFile, int nSize )
func GiaFileWriteBufferSize(pFile *os.File, nSize int) {
	panic("TODO, stubbed")
}

// Gia_AigerCollectLiterals
// C: Vec_Int_t * Gia_AigerCollectLiterals( Gia_Man_t * p )
func (p *GiaManager) GiaAigerCollectLiterals() *IntVector {
	panic("TODO, stubbed")
}

// Gia_AigerReadLiterals (ppPos is an unsigned char ** cursor; here a slice that is advanced)
// C: Vec_Int_t * Gia_AigerReadLiterals( unsigned char ** ppPos, int nEntries )
func GiaAigerReadLiterals(ppPos *[]uint8, nEntries int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_AigerWriteLiterals
// C: Vec_Str_t * Gia_AigerWriteLiterals( Vec_Int_t * vLits )
func GiaAigerWriteLiterals(vLits *IntVector) *ByteVector {
	panic("TODO, stubbed")
}

// Gia_AigerReadFromMemory (pContents is the binary file contents)
// C: Gia_Man_t * Gia_AigerReadFromMemory( char * pContents, int nFileSize, int fGiaSimple, int fSkipStrash, int fCheck )
func GiaAigerReadFromMemory(pContents []byte, nFileSize int, fGiaSimple int, fSkipStrash int, fCheck int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_AigerRead
// C: Gia_Man_t * Gia_AigerRead( char * pFileName, int fGiaSimple, int fSkipStrash, int fCheck )
func GiaAigerRead(pFileName string, fGiaSimple int, fSkipStrash int, fCheck int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_AigerWriteIntoMemoryStr
// C: Vec_Str_t * Gia_AigerWriteIntoMemoryStr( Gia_Man_t * p )
func (p *GiaManager) GiaAigerWriteIntoMemoryStr() *ByteVector {
	panic("TODO, stubbed")
}

// Gia_AigerWriteIntoMemoryStrPart
// C: Vec_Str_t * Gia_AigerWriteIntoMemoryStrPart( Gia_Man_t * p, Vec_Int_t * vCis, Vec_Int_t * vAnds, Vec_Int_t * vCos, int nRegs )
func (p *GiaManager) GiaAigerWriteIntoMemoryStrPart(vCis *IntVector, vAnds *IntVector, vCos *IntVector, nRegs int) *ByteVector {
	panic("TODO, stubbed")
}

// Gia_AigerWrite
// C: void Gia_AigerWrite( Gia_Man_t * pInit, char * pFileName, int fWriteSymbols, int fCompact, int fWriteNewLine )
func (pInit *GiaManager) GiaAigerWrite(pFileName string, fWriteSymbols int, fCompact int, fWriteNewLine int) {
	panic("TODO, stubbed")
}

// Gia_DumpAiger
// C: void Gia_DumpAiger( Gia_Man_t * p, char * pFilePrefix, int iFileNum, int nFileNumDigits )
func (p *GiaManager) GiaDumpAiger(pFilePrefix string, iFileNum int, nFileNumDigits int) {
	panic("TODO, stubbed")
}

// Gia_AigerWriteSimple
// C: void Gia_AigerWriteSimple( Gia_Man_t * pInit, char * pFileName )
func (pInit *GiaManager) GiaAigerWriteSimple(pFileName string) {
	panic("TODO, stubbed")
}
