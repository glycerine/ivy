package abc

// Hand-checked skeleton of src/map/mio/mioRead.c.

// Mio_LibraryRead
// C: Mio_Library_t * Mio_LibraryRead( char * FileName, char * pBuffer, char * ExcludeFile, int fVerbose )
func MioLibraryRead(FileName string, pBuffer string, ExcludeFile string, fVerbose int) *MioLibrary {
	panic("TODO, stubbed")
}

// Mio_ReadFile (returns the file contents buffer)
// C: char * Mio_ReadFile( char * FileName, int fAddEnd )
func MioReadFile(FileName string, fAddEnd int) []byte {
	panic("TODO, stubbed")
}

// Mio_LibraryReadBuffer (pBuffer is parsed (and modified) in place)
// C: Mio_Library_t * Mio_LibraryReadBuffer( char * pBuffer, int fExtendedFormat, st__table * tExcludeGate, int fVerbose )
func MioLibraryReadBuffer(pBuffer []byte, fExtendedFormat int, tExcludeGate *StableTable[string, int], fVerbose int) *MioLibrary {
	panic("TODO, stubbed")
}

// Mio_LibraryReadOne (static)
// C: Mio_Library_t * Mio_LibraryReadOne( char * FileName, int fExtendedFormat, st__table * tExcludeGate, int fVerbose )
func mioLibraryReadOne(FileName string, fExtendedFormat int, tExcludeGate *StableTable[string, int], fVerbose int) *MioLibrary {
	panic("TODO, stubbed")
}

// Mio_LibraryReadInternal (static; pBuffer is parsed (and modified) in place)
// C: int Mio_LibraryReadInternal( Mio_Library_t * pLib, char * pBuffer, int fExtendedFormat, st__table * tExcludeGate, int fVerbose )
func (pLib *MioLibrary) mioLibraryReadInternal(pBuffer []byte, fExtendedFormat int, tExcludeGate *StableTable[string, int], fVerbose int) int {
	panic("TODO, stubbed")
}

// Mio_LibraryCleanStr (cleans the string in place and returns it)
// C: char * Mio_LibraryCleanStr( char * p )
func MioLibraryCleanStr(p []byte) []byte {
	panic("TODO, stubbed")
}

// Mio_LibraryReadGate (static; ppToken is a cursor the function advances)
// C: Mio_Gate_t * Mio_LibraryReadGate( char ** ppToken, int fExtendedFormat )
func mioLibraryReadGate(ppToken *string, fExtendedFormat int) *MioGate {
	panic("TODO, stubbed")
}

// Mio_LibraryReadPin (static; ppToken is a cursor the function advances)
// C: Mio_Pin_t * Mio_LibraryReadPin( char ** ppToken, int fExtendedFormat )
func mioLibraryReadPin(ppToken *string, fExtendedFormat int) *MioPin {
	panic("TODO, stubbed")
}

// chomp (static; strips trailing newline characters in place)
// C: char * chomp( char *s )
func chomp(s []byte) []byte {
	panic("TODO, stubbed")
}

// Mio_LibraryCompareGatesByArea
// C: int Mio_LibraryCompareGatesByArea( Mio_Gate_t ** pp1, Mio_Gate_t ** pp2 )
func MioLibraryCompareGatesByArea(pp1 **MioGate, pp2 **MioGate) int {
	panic("TODO, stubbed")
}

// Mio_LibraryCompareGatesByName
// C: int Mio_LibraryCompareGatesByName( Mio_Gate_t ** pp1, Mio_Gate_t ** pp2 )
func MioLibraryCompareGatesByName(pp1 **MioGate, pp2 **MioGate) int {
	panic("TODO, stubbed")
}

// Mio_LibrarySortGates
// C: void Mio_LibrarySortGates( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibrarySortGates() {
	panic("TODO, stubbed")
}

// Mio_GateCompare (static)
// C: Mio_Gate_t * Mio_GateCompare( Mio_Gate_t * pThis, Mio_Gate_t * pNew, word uTruth )
func mioGateCompare(pThis *MioGate, pNew *MioGate, uTruth uint64) *MioGate {
	panic("TODO, stubbed")
}

// Mio_LibraryDetectSpecialGates (static)
// C: void Mio_LibraryDetectSpecialGates( Mio_Library_t * pLib )
func (pLib *MioLibrary) mioLibraryDetectSpecialGates() {
	panic("TODO, stubbed")
}

// Mio_LibraryReadExclude
// C: int Mio_LibraryReadExclude( char * ExcludeFile, st__table * tExcludeGate )
func MioLibraryReadExclude(ExcludeFile string, tExcludeGate *StableTable[string, int]) int {
	panic("TODO, stubbed")
}

// Io_ReadFileRemoveComments (static; pnDots and pnLines are out parameters (nullable))
// C: void Io_ReadFileRemoveComments( char * pBuffer, int * pnDots, int * pnLines )
func ioReadFileRemoveComments(pBuffer []byte, pnDots *int, pnLines *int) {
	panic("TODO, stubbed")
}
