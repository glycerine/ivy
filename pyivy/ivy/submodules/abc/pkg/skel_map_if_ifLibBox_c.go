package abc

import "os"

// Hand-checked skeleton of src/map/if/ifLibBox.c.

// If_BoxStart
// C: If_Box_t * If_BoxStart( char * pName, int Id, int nPis, int nPos, int fSeq, int fBlack, int fOuter )
func IfBoxStart(pName string, Id int, nPis int, nPos int, fSeq int, fBlack int, fOuter int) *IfBox {
	panic("TODO, stubbed")
}

// If_BoxDup
// C: If_Box_t * If_BoxDup( If_Box_t * p )
func IfBoxDup(p *IfBox) *IfBox {
	panic("TODO, stubbed")
}

// If_BoxFree
// C: void If_BoxFree( If_Box_t * p )
func IfBoxFree(p *IfBox) {
	panic("TODO, stubbed")
}

// If_LibBoxStart
// C: If_LibBox_t * If_LibBoxStart()
func IfLibBoxStart() *IfLibBox {
	panic("TODO, stubbed")
}

// If_LibBoxDup (the C parameter is an If_Box_t *)
// C: If_LibBox_t * If_LibBoxDup( If_Box_t * p )
func IfLibBoxDup(p *IfBox) *IfLibBox {
	panic("TODO, stubbed")
}

// If_LibBoxFree
// C: void If_LibBoxFree( If_LibBox_t * p )
func (p *IfLibBox) IfLibBoxFree() {
	panic("TODO, stubbed")
}

// If_LibBoxReadBox
// C: If_Box_t * If_LibBoxReadBox( If_LibBox_t * p, int Id )
func (p *IfLibBox) IfLibBoxReadBox(Id int) *IfBox {
	panic("TODO, stubbed")
}

// If_LibBoxFindBox
// C: If_Box_t * If_LibBoxFindBox( If_LibBox_t * p, char * pName )
func (p *IfLibBox) IfLibBoxFindBox(pName string) *IfBox {
	panic("TODO, stubbed")
}

// If_LibBoxAdd
// C: void If_LibBoxAdd( If_LibBox_t * p, If_Box_t * pBox )
func (p *IfLibBox) IfLibBoxAdd(pBox *IfBox) {
	panic("TODO, stubbed")
}

// If_LibBoxNum
// C: int If_LibBoxNum( If_LibBox_t * p )
func (p *IfLibBox) IfLibBoxNum() int {
	panic("TODO, stubbed")
}

// If_LibBoxRead2
// C: If_LibBox_t * If_LibBoxRead2( char * pFileName )
func IfLibBoxRead2(pFileName string) *IfLibBox {
	panic("TODO, stubbed")
}

// If_LibBoxGetToken (returns a pointer to a static buffer, or NULL)
// C: char * If_LibBoxGetToken( FILE * pFile )
func IfLibBoxGetToken(pFile *os.File) string {
	panic("TODO, stubbed")
}

// If_LibBoxRead
// C: If_LibBox_t * If_LibBoxRead( char * pFileName )
func IfLibBoxRead(pFileName string) *IfLibBox {
	panic("TODO, stubbed")
}

// If_LibBoxPrint
// C: void If_LibBoxPrint( FILE * pFile, If_LibBox_t * p )
func IfLibBoxPrint(pFile *os.File, p *IfLibBox) {
	panic("TODO, stubbed")
}

// If_LibBoxWrite
// C: void If_LibBoxWrite( char * pFileName, If_LibBox_t * p )
func IfLibBoxWrite(pFileName string, p *IfLibBox) {
	panic("TODO, stubbed")
}

// If_LibBoxLoad
// C: int If_LibBoxLoad( char * pFileName )
func IfLibBoxLoad(pFileName string) int {
	panic("TODO, stubbed")
}
