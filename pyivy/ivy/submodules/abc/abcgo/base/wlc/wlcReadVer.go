package wlc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [wlcReadVer.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Parses several flavors of word-level Verilog.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcReadVer.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Word-level Verilog file parser
const WLV_PRS_MAX_LINE = 10000

type Wlc_Prs_t struct {
	nFileSize int32
	pFileName string
	pBuffer   string
	vLines    *Vec_Int_t
	vStarts   *Vec_Int_t
	vFanins   *Vec_Int_t
	pNtk      *Wlc_Ntk_t
	pMemTable *Mem_Flex_t
	vTables   *Vec_Ptr_t
	vPoPairs  *Vec_Str_t
	nConsts   int32
	nNonZero  [4]int32
	nNegative [4]int32
	nReverse  [4]int32
	sError    [WLV_PRS_MAX_LINE]byte
}

// wlcPrsCursor is the Go equivalent of a C char* into Wlc_Prs_t.pBuffer.
// The offset is kept explicitly so advancing a parser cursor never copies or
// detaches it from the source buffer. An offset below zero represents NULL.
type wlcPrsCursor struct {
	parser *Wlc_Prs_t
	offset int
}

func wlcPrsCursorAt(p *Wlc_Prs_t, offset int) wlcPrsCursor {
	if p == nil || offset < 0 || offset > len(p.pBuffer) {
		return wlcPrsCursor{parser: p, offset: -1}
	}
	return wlcPrsCursor{parser: p, offset: offset}
}

func (c wlcPrsCursor) valid() bool {
	return c.parser != nil && c.offset >= 0 && c.offset <= len(c.parser.pBuffer)
}

// rest returns the NUL-terminated C string beginning at this cursor. Parser
// buffers are Go strings, so the first embedded NUL supplies C's terminator.
func (c wlcPrsCursor) rest() string {
	if !c.valid() {
		return ""
	}
	s := c.parser.pBuffer[c.offset:]
	if end := strings.IndexByte(s, 0); end >= 0 {
		s = s[:end]
	}
	return s
}

func (c wlcPrsCursor) advance(n int) wlcPrsCursor {
	if !c.valid() || n < 0 || c.offset+n > len(c.parser.pBuffer) {
		return wlcPrsCursor{parser: c.parser, offset: -1}
	}
	return wlcPrsCursorAt(c.parser, c.offset+n)
}

func (c wlcPrsCursor) retreat(n int) wlcPrsCursor {
	if !c.valid() || n < 0 || n > c.offset {
		return wlcPrsCursor{parser: c.parser, offset: -1}
	}
	return wlcPrsCursorAt(c.parser, c.offset-n)
}

func (c wlcPrsCursor) findSubstring(word string) (wlcPrsCursor, bool) {
	if !c.valid() {
		return wlcPrsCursor{parser: c.parser, offset: -1}, false
	}
	if i := strings.Index(c.rest(), word); i >= 0 {
		return c.advance(i), true
	}
	return wlcPrsCursor{parser: c.parser, offset: -1}, false
}

func (c wlcPrsCursor) withStringSuffix(s string) wlcPrsCursor {
	if !c.valid() {
		return wlcPrsCursor{parser: c.parser, offset: -1}
	}
	base := c.parser.pBuffer[c.offset:]
	if len(s) > len(base) || base[:len(s)] != s {
		return wlcPrsCursor{parser: c.parser, offset: -1}
	}
	return c.advance(len(base) - len(s))
}

func (c wlcPrsCursor) byte() byte {
	if s := c.rest(); len(s) != 0 {
		return s[0]
	}
	return 0
}

func (c wlcPrsCursor) skipSpaces() wlcPrsCursor {
	s := c.rest()
	i := 0
	for i < len(s) && s[i] == ' ' {
		i++
	}
	return c.advance(i)
}

func (c wlcPrsCursor) hasPrefix(word string) bool { return strings.HasPrefix(c.rest(), word) }

func prsCStringLen(s string) int {
	if i := strings.IndexByte(s, 0); i >= 0 {
		return i
	}
	return len(s)
}

// prsSuffix preserves a valid C pointer at the terminator. Empty Go strings
// remain the representation of C NULL / parse failure.
func prsSuffix(s string, offset int) string {
	if offset < 0 || offset > len(s) {
		return ""
	}
	if offset == len(s) {
		return "\x00"
	}
	return s[offset:]
}

// findSymbol follows Wlc_PrsFindSymbol, including its escaped-identifier
// behavior: a delimiter is recognized only outside a backslash name.
func (c wlcPrsCursor) findSymbol(sym byte) wlcPrsCursor {
	s := c.rest()
	notName := true
	for i := 0; i < len(s); i++ {
		if notName && s[i] == sym {
			return c.advance(i)
		}
		if s[i] == '\\' {
			notName = false
		} else if !notName && s[i] == ' ' {
			notName = true
		}
	}
	return wlcPrsCursor{parser: c.parser, offset: -1}
}

func (c wlcPrsCursor) findWord(word string) (wlcPrsCursor, bool) {
	c = c.skipSpaces()
	if c.hasPrefix(word) {
		return c, true
	}
	return c, false
}

// findName returns the next cursor and the token text. It mirrors the C
// helper's escaped-name rule, where whitespace terminates a backslash name.
func (c wlcPrsCursor) findName() (wlcPrsCursor, string) {
	c = c.skipSpaces()
	s := c.rest()
	if len(s) == 0 || !prsIsChar(s[0]) {
		return wlcPrsCursor{parser: c.parser, offset: -1}, ""
	}
	notName, count := true, 0
	for i := 0; i < len(s); i++ {
		if notName && !prsIsChar(s[i]) {
			return c.advance(i), s[:i]
		}
		if s[i] == '\\' {
			count++
			notName = false
		} else if !notName && s[i] == ' ' {
			count--
			if count == 0 {
				notName = true
			}
		}
	}
	return c.advance(len(s)), s
}

func prsIsChar(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_' || ch == '$' || ch == '\\'
}

func prsVecStrAlloc(n int32) *Vec_Str_t         { return vec.Vec_StrAlloc(int(n)) }
func prsVecStrEntryLast(v *Vec_Str_t) byte      { return vec.Vec_StrEntryLast(v) }
func prsVecStrReleaseArray(v *Vec_Str_t) []byte { return vec.Vec_StrReleaseArray(v) }

func Wlc_PrsOffset(p *Wlc_Prs_t, pStr string) int32 {
	if p == nil || len(pStr) > len(p.pBuffer) || !strings.HasSuffix(p.pBuffer, pStr) {
		return -1
	}
	return int32(len(p.pBuffer) - len(pStr))
}
func Wlc_PrsStr(p *Wlc_Prs_t, iOffset int32) string {
	if iOffset < 0 || int(iOffset) > len(p.pBuffer) {
		return ""
	}
	return p.pBuffer[iOffset:]
}
func Wlc_PrsStrCmp(pStr string, pWhat string) int32 {
	return boolToInt32(strings.HasPrefix(pStr, pWhat))
}

// Original macro:
// #define Wlc_PrsForEachLine( p, pLine, i )             \
//     for ( i = 0; (i < Vec_IntSize((p)->vStarts)) && ((pLine) = Wlc_PrsStr(p, Vec_IntEntry((p)->vStarts, i))); i++ )

// Original macro:
// #define Wlc_PrsForEachLineStart( p, pLine, i, Start ) \
//     for ( i = Start; (i < Vec_IntSize((p)->vStarts)) && ((pLine) = Wlc_PrsStr(p, Vec_IntEntry((p)->vStarts, i))); i++ )

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_PrsStart(pFileName string, pStr string) *Wlc_Prs_t {
	var p *Wlc_Prs_t
	if pFileName != "" && Extra_FileCheck(pFileName) == 0 {
		return nil
	}
	p = new(Wlc_Prs_t)
	p.pFileName = pFileName
	if pStr != "" {
		p.pBuffer = Abc_UtilStrsav(pStr)
	} else {
		p.pBuffer = Extra_FileReadContents(pFileName)
	}
	if nul := strings.IndexByte(p.pBuffer, 0); nul >= 0 {
		p.pBuffer = p.pBuffer[:nul]
	}
	p.nFileSize = int32(len(p.pBuffer))
	assert(p.nFileSize > 0, "p.nFileSize > 0")
	p.vLines = Vec_IntAlloc(p.nFileSize / 50)
	p.vStarts = Vec_IntAlloc(p.nFileSize / 50)
	p.vFanins = Vec_IntAlloc(100)
	p.vTables = Vec_PtrAlloc(1000)
	p.pMemTable = Mem_FlexStart()
	return p
}
func Wlc_PrsStop(p *Wlc_Prs_t) {
	if p.pNtk != nil {
		Wlc_NtkFree(p.pNtk)
	}
	if p.pMemTable != nil {
		Mem_FlexStop(p.pMemTable, 0)
	}
	Vec_StrFreeP(&p.vPoPairs)
	Vec_PtrFreeP(&p.vTables)
	Vec_IntFree(p.vLines)
	Vec_IntFree(p.vStarts)
	Vec_IntFree(p.vFanins)
	p.pBuffer = ""
	p = nil
}

/*
*Function*************************************************************

	Synopsis    [Prints the error message including the file name and line number.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_PrsFindLine(p *Wlc_Prs_t, pCur string) int32 {
	// C receives a pointer into pBuffer. Go represents that pointer as a
	// suffix string, so its byte offset is the difference in lengths.
	if p == nil || !strings.HasSuffix(p.pBuffer, pCur) {
		return -1
	}
	offset := len(p.pBuffer) - len(pCur)
	for i := int32(0); i < Vec_IntSize(p.vLines); i++ {
		if int(Vec_IntEntry(p.vLines, i)) > offset {
			return i + 1
		}
	}
	return -1
}
func Wlc_PrsWriteErrorMessage(p *Wlc_Prs_t, pCur string, format string, args ...any) int32 {
	message := fmt.Sprintf(format, args...)
	assert(len(message) < WLV_PRS_MAX_LINE-100, "strlen(pMessage) < WLV_PRS_MAX_LINE - 100")
	assert(p.sError[0] == 0, "p.sError[0] == 0")
	var rendered string
	if pCur == "" { // Empty string represents the C NULL cursor.
		rendered = fmt.Sprintf("%s: %s\n", p.pFileName, message)
	} else {
		rendered = fmt.Sprintf("%s (line %d): %s\n", p.pFileName, Wlc_PrsFindLine(p, pCur), message)
	}
	copy(p.sError[:], rendered)
	return 0
}
func Wlc_PrsPrintErrorMessage(p *Wlc_Prs_t) {
	if p.sError[0] == 0 {
		return
	}
	end := 0
	for end < len(p.sError) && p.sError[end] != 0 {
		end++
	}
	fmt.Fprint(os.Stdout, string(p.sError[:end]))
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_PrsIsDigit(pStr string) int32 {
	if len(pStr) == 0 {
		return 0
	}
	return boolToInt32(pStr[0] >= '0' && pStr[0] <= '9')
}
func Wlc_PrsIsChar(pStr string) int32 {
	if len(pStr) == 0 {
		return 0
	}
	return boolToInt32((pStr[0] >= 'a' && pStr[0] <= 'z') ||
		(pStr[0] >= 'A' && pStr[0] <= 'Z') ||
		(pStr[0] >= '0' && pStr[0] <= '9') ||
		pStr[0] == '_' || pStr[0] == '$' || pStr[0] == '\\')
}
func Wlc_PrsSkipSpaces(pStr string) string {
	originalLen := len(pStr)
	for len(pStr) != 0 && pStr[0] == ' ' {
		pStr = pStr[1:]
	}
	if len(pStr) == 0 && originalLen != 0 {
		return "\x00"
	}
	return pStr
}
func Wlc_PrsFindSymbol(pStr string, Symb byte) string {
	notName := true
	for i := 0; i < prsCStringLen(pStr); i++ {
		if notName && pStr[i] == Symb {
			return pStr[i:]
		}
		if pStr[i] == '\\' {
			notName = false
		} else if !notName && pStr[i] == ' ' {
			notName = true
		}
	}
	return ""
}
func Wlc_PrsFindSymbolTwo(pStr string, Symb byte, Symb2 byte) string {
	for i := 0; i+1 < prsCStringLen(pStr); i++ {
		if pStr[i] == Symb && pStr[i+1] == Symb2 {
			return pStr[i:]
		}
	}
	return ""
}
func Wlc_PrsFindClosingParenthesis(pStr string, Open byte, Close byte) string {
	var counter int32
	notName := true
	assert(len(pStr) != 0 && pStr[0] == Open, "*pStr == Open")
	for i := 0; i < prsCStringLen(pStr); i++ {
		if notName {
			if pStr[i] == Open {
				counter++
			}
			if pStr[i] == Close {
				counter--
			}
			if counter == 0 {
				return pStr[i:]
			}
		}
		if pStr[i] == '\\' {
			notName = false
		} else if !notName && pStr[i] == ' ' {
			notName = true
		}
	}
	return ""
}
func Wlc_PrsRemoveComments(p *Wlc_Prs_t) int32 {
	var fSpecifyFound int32
	buf := []byte(p.pBuffer)
	limit := int(p.nFileSize)
	if limit <= 0 || limit > len(buf) {
		limit = len(buf)
	}
	buf = buf[:limit]
	for i := 0; i < len(buf); i++ {
		// regular comment (//)
		if i+1 < len(buf) && buf[i] == '/' && buf[i+1] == '/' {
			if i+5 < len(buf) && string(buf[i+2:i+6]) == "abc2" {
				for k := 0; k < 6; k++ {
					buf[i+k] = ' '
				}
			} else if i+8 < len(buf) && string(buf[i+3:i+8]) == "Pair:" {
				if p.vPoPairs == nil {
					p.vPoPairs = prsVecStrAlloc(100)
				}
				for j := i + 9; j < len(buf) && buf[j] != '\n'; j++ {
					if buf[j] == ' ' {
						Vec_StrPush(p.vPoPairs, '\x00')
					} else if buf[j] != '\r' {
						Vec_StrPush(p.vPoPairs, buf[j])
					}
				}
				if prsVecStrEntryLast(p.vPoPairs) != 0 {
					Vec_StrPush(p.vPoPairs, 0)
				}

			} else {
				j := findSymbolBytes(buf, i, '\n')
				if j < 0 {
					p.pBuffer = string(buf)
					return Wlc_PrsWriteErrorMessage(p, p.pBuffer[i:], "Cannot find end-of-line after symbols \"//\".")
				}
				for ; i < j; i++ {
					buf[i] = ' '
				}
			}

			// skip preprocessor directive (`timescale, `celldefine, etc)
		} else if buf[i] == '`' {
			j := findSymbolBytes(buf, i, '\n')
			if j < 0 {
				p.pBuffer = string(buf)
				return Wlc_PrsWriteErrorMessage(p, p.pBuffer[i:], "Cannot find end-of-line after symbols \"`\".")
			}
			for ; i < j; i++ {
				buf[i] = ' '
			}

			// regular comment (/* ... */)
		} else if i+1 < len(buf) && buf[i] == '/' && buf[i+1] == '*' {
			j := findPairBytes(buf, i, '*', '/')
			if j < 0 {
				p.pBuffer = string(buf)
				return Wlc_PrsWriteErrorMessage(p, p.pBuffer[i:], "Cannot find symbols \"*/\" after symbols \"/*\".")
			}
			// overwrite comment
			for ; i < j+2; i++ {
				buf[i] = ' '
			}

			// 'specify' treated as comments
		} else if i+7 <= len(buf) && string(buf[i:i+7]) == "specify" {
			for j := i; j+10 < len(buf); j++ {
				if string(buf[j:j+10]) == "endspecify" {
					// overwrite comment
					for ; i < j+10; i++ {
						buf[i] = ' '
					}
					if fSpecifyFound == 0 {
						Abc_Print(0, "Ignoring specify/endspecify directives.\n")
					}
					fSpecifyFound = 1
					break
				}
			}

			// insert semicolons
		} else if i+9 <= len(buf) && string(buf[i:i+9]) == "endmodule" {
			if i+9 < len(buf) {
				buf[i+9] = ';'
			} else if i+9 == len(buf) {
				buf = append(buf, ';')
			}
		}
		// overwrite end-of-lines with spaces (less checking to do later on)
		if buf[i] == '\n' || buf[i] == '\r' || buf[i] == '\t' {
			buf[i] = ' '
		}
	}
	p.pBuffer = string(buf)
	return 1
}

func findSymbolBytes(buf []byte, start int, sym byte) int {
	notName := true
	for i := start; i < len(buf); i++ {
		if buf[i] == 0 {
			break
		}
		if notName && buf[i] == sym {
			return i
		}
		if buf[i] == '\\' {
			notName = false
		} else if !notName && buf[i] == ' ' {
			notName = true
		}
	}
	return -1
}

func findPairBytes(buf []byte, start int, first, second byte) int {
	for i := start; i+1 < len(buf) && buf[i] != 0 && buf[i+1] != 0; i++ {
		if buf[i] == first && buf[i+1] == second {
			return i
		}
	}
	return -1
}
func Wlc_PrsPrepare(p *Wlc_Prs_t) int32 {
	var fPrettyPrint int32
	var fNotName int32 = 1
	var pTemp string
	// collect info about lines
	assert(Vec_IntSize(p.vLines) == 0, "Vec_IntSize(p.vLines) == 0")
	for i := 0; i < prsCStringLen(p.pBuffer); i++ {
		if p.pBuffer[i] == '\n' {
			Vec_IntPush(p.vLines, int32(i))
		}
	}
	// delete comments and insert breaks
	if !(Wlc_PrsRemoveComments(p) != 0) {
		return 0
	}
	// collect info about breaks
	assert(Vec_IntSize(p.vStarts) == 0, "Vec_IntSize(p.vStarts) == 0")
	buf := []byte(p.pBuffer)
	prev := 0
	for i := 0; i < len(buf) && buf[i] != 0; i++ {
		if fNotName != 0 && buf[i] == ';' {
			buf[i] = 0
			lineStart := prev
			for lineStart < i && buf[lineStart] == ' ' {
				lineStart++
			}
			Vec_IntPush(p.vStarts, int32(lineStart))
			prev = i + 1
		}
		if buf[i] == '\\' {
			fNotName = 0
		} else if !(fNotName != 0) && buf[i] == ' ' {
			fNotName = 1
		}
	}
	p.pBuffer = string(buf)

	if fPrettyPrint != 0 {
		var i, k int32
		// print the line types
		// inlining of macro Wlc_PrsForEachLine
		for i = int32(0); i < Vec_IntSize(p.vStarts); i++ {
			pTemp = Wlc_PrsStr(p, Vec_IntEntry(p.vStarts, i))
			if Wlc_PrsStrCmp(pTemp, "module") != 0 {
				fmt.Printf("\n")
			}
			if !(Wlc_PrsStrCmp(pTemp, "module") != 0) && !(Wlc_PrsStrCmp(pTemp, "endmodule") != 0) {
				fmt.Printf("    ")
			}
			fmt.Printf("%c", pTemp[0])
			for k = 1; k < int32(len(pTemp)) && pTemp[k] != 0; k++ {
				if pTemp[k] != ' ' || pTemp[k-1] != ' ' {
					fmt.Printf("%c", pTemp[k])
				}
			}
			fmt.Printf(";\n")
		}
		/*
		   // print the line types
		   Wlc_PrsForEachLine( p, pTemp, i )
		   {
		       int k;
		       if ( !Wlc_PrsStrCmp( pTemp, "module" ) )
		           continue;
		       printf( "%3d : ", i );
		       for ( k = 0; k < 40; k++ )
		           printf( "%c", pTemp[k] ? pTemp[k] : ' ' );
		       printf( "\n" );
		   }
		*/
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Modified version of strtok().]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Wlc_PrsStrtok_last string

func Wlc_PrsStrtok(s string, delim string) string {
	if s == "" {
		s = spStatic_Wlc_PrsStrtok_last
	}
	if s == "" {
		return ""
	}
	length := prsCStringLen(s)
	i := 0
	for i < length && strings.IndexByte(delim, s[i]) >= 0 {
		i++
	}
	if i == length {
		spStatic_Wlc_PrsStrtok_last = ""
		return ""
	}
	start := i
	for i < length {
		ch := s[i]
		i++
		if ch == '\\' {
			for ch != ' ' && i < length {
				ch = s[i]
				i++
			}
			if ch == ' ' {
				if i < length {
					ch = s[i]
					i++
				} else {
					ch = 0
				}
			}
		}
		if ch == 0 {
			spStatic_Wlc_PrsStrtok_last = ""
			return s[start:length] + "\x00"
		}
		if strings.IndexByte(delim, ch) >= 0 {
			spStatic_Wlc_PrsStrtok_last = s[i:length]
			return s[start:i-1] + "\x00"
		}
	}
	spStatic_Wlc_PrsStrtok_last = ""
	return s[start:length] + "\x00"
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_PrsConvertInitValues(p *Wlc_Ntk_t) string {
	vStr := prsVecStrAlloc(1000)
	for i := int32(0); i < Vec_IntSize(p.vInits); i++ {
		value := Vec_IntEntry(p.vInits, i)
		if value < 0 {
			for k := int32(0); k < -value; k++ {
				Vec_StrPush(vStr, '0')
			}
			continue
		}

		pObj := Wlc_NtkObj(p, value)
		value = Wlc_ObjRange(pObj)
		for pObj.Type == uint32(WLC_OBJ_BUF) {
			pObj = Wlc_NtkObj(p, Wlc_ObjFaninId0(pObj))
		}

		var pInits *int32
		if pObj.Type == uint32(WLC_OBJ_CONST) && pObj.fXConst == 0 {
			pInits = Wlc_ObjConstValue(pObj)
		}
		constRange := Wlc_ObjRange(pObj)
		for k := int32(0); k < Abc_MinInt(value, constRange); k++ {
			bit := byte('x')
			if pInits != nil {
				words := unsafe.Slice((*uint32)(unsafe.Pointer(pInits)), int((constRange+31)/32))
				if words[k>>5]&(uint32(1)<<uint(k&31)) != 0 {
					bit = '1'
				} else {
					bit = '0'
				}
			}
			Vec_StrPush(vStr, bit)
		}
		// Extend with zero when the initialized object is narrower than its signal.
		for k := Abc_MinInt(value, constRange); k < value; k++ {
			Vec_StrPush(vStr, '0')
		}
		if pInits != nil || pObj.fXConst != 0 {
			Vec_IntWriteEntry(p.vInits, i, -value)
		} else {
			Vec_IntWriteEntry(p.vInits, i, Wlc_ObjCiId(pObj))
		}
	}
	Vec_StrPush(vStr, '\x00')
	result := prsVecStrReleaseArray(vStr)
	Vec_StrFree(vStr)
	return string(result)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_PrsCheckBitConst0(p *Wlc_Ntk_t, NameId int32) int32 {
	var pObj *Wlc_Obj_t = Wlc_NtkObj(p, NameId)
	var pInits *int32
	if Wlc_ObjRange(pObj) != 1 {
		return 0
	}
	for pObj.Type == uint32(WLC_OBJ_BUF) {
		pObj = Wlc_NtkObj(p, Wlc_ObjFaninId0(pObj))
	}
	if pObj.Type != uint32(WLC_OBJ_CONST) {
		return 0
	}
	pInits = Wlc_ObjConstValue(pObj)
	return boolToInt32(!Abc_InfoHasBit(unsafe.Slice((*uint32)(unsafe.Pointer(pInits)), 1), 0))
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_PrsFindRange(pStr string, End *int32, Beg *int32) string {
	*End = 0
	*Beg = 0
	pStr = Wlc_PrsSkipSpaces(pStr)
	if len(pStr) == 0 || pStr[0] != '[' {
		return pStr
	}
	pStr = Wlc_PrsSkipSpaces(pStr[1:])
	if Wlc_PrsIsDigit(pStr) == 0 && (len(pStr) == 0 || pStr[0] != '-') {
		return ""
	}
	*End = prsAtoi(pStr)
	*Beg = *End
	colon := Wlc_PrsFindSymbol(pStr, ':')
	if colon == "" {
		pStr = Wlc_PrsFindSymbol(pStr, ']')
		if pStr == "" {
			return ""
		}
		return prsSuffix(pStr, 1)
	}
	pStr = Wlc_PrsSkipSpaces(colon[1:])
	if Wlc_PrsIsDigit(pStr) == 0 && (len(pStr) == 0 || pStr[0] != '-') {
		return ""
	}
	*Beg = prsAtoi(pStr)
	pStr = Wlc_PrsFindSymbol(pStr, ']')
	if pStr == "" {
		return ""
	}
	return prsSuffix(pStr, 1)
}

func prsParseSignedInt(s string) (int32, int) {
	i := 0
	if len(s) > 0 && (s[0] == '-' || s[0] == '+') {
		i++
	}
	startDigits := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == startDigits {
		return 0, 0
	}
	n, err := strconv.ParseInt(s[:i], 10, 32)
	if err != nil {
		return 0, 0
	}
	return int32(n), i
}

func prsAtoi(s string) int32 {
	n, _ := prsParseSignedInt(s)
	return n
}
func Wlc_PrsFindWord(pStr string, pWord string, fFound *int32) string {
	*fFound = 0
	pStr = Wlc_PrsSkipSpaces(pStr)
	if Wlc_PrsStrCmp(pStr, pWord) == 0 {
		return pStr
	}
	*fFound = 1
	return prsSuffix(pStr, len(pWord))
}

var spStatic_Wlc_PrsFindName_Buffer [WLV_PRS_MAX_LINE]byte

func Wlc_PrsFindName(pStr string, ppPlace *string) string {
	pStr = Wlc_PrsSkipSpaces(pStr)
	if pStr == "" || Wlc_PrsIsChar(pStr) == 0 {
		return ""
	}
	notName, count, consumed := true, 0, 0
	for consumed < len(pStr) && pStr[consumed] != 0 {
		ch := pStr[consumed]
		if notName && Wlc_PrsIsChar(pStr[consumed:]) == 0 {
			break
		}
		if ch == '\\' {
			count++
			notName = false
		} else if !notName && ch == ' ' {
			count--
			if count == 0 {
				notName = true
			}
		}
		consumed++
	}
	*ppPlace = pStr[:consumed]
	return prsSuffix(pStr, consumed)
}
func Wlc_PrsReadConstant(p *Wlc_Prs_t, pStr string, vFanins *Vec_Int_t, pRange *int32, pSigned *int32, pXValue *int32) string {
	pStr = Wlc_PrsSkipSpaces(pStr)
	nBits, width := prsParseSignedInt(pStr)
	*pRange = -1
	*pSigned = 0
	*pXValue = 0
	if Wlc_PrsIsDigit(pStr) == 0 || width == 0 {
		return ""
	}
	quote := strings.IndexByte(pStr, '\'')
	if quote < 0 {
		// handle decimal number
		Number := nBits
		*pRange = Abc_Base2Log(Number + 1)
		assert(*pRange < 32, "*pRange < 32")
		pStr = prsSuffix(pStr, width)
		Vec_IntFill(vFanins, 1, Number)
		return pStr
	}
	pStr = pStr[quote:]
	if pStr[1] == 's' {
		*pSigned = 1
		pStr = pStr[1:]
	}
	if pStr[1] == 'b' {
		nWords := util.Abc_BitWordNum(nBits)
		Vec_IntFill(vFanins, nWords, 0)
		for i := int32(0); i < nBits; i++ {
			if int(i+2) >= len(pStr) {
				return ""
			}
			if pStr[2+i] == '1' {
				words := unsafe.Slice((*uint32)(unsafe.Pointer(&Vec_IntArray(vFanins)[0])), int(nWords))
				Abc_InfoSetBit(words, nBits-1-i)
			} else if pStr[2+i] != '0' {
				Wlc_PrsWriteErrorMessage(p, pStr, "Wrong digit in binary constant \"%c\".", pStr[2+i])
				return ""
			}
		}
		*pRange = nBits
		return prsSuffix(pStr, 2+int(nBits))
	}
	if pStr[1] != 'h' {
		Wlc_PrsWriteErrorMessage(p, pStr, "Expecting hexadecimal constant and not \"%c\".", pStr[1])
		return ""
	}
	if len(pStr) > 2 && (pStr[2] == 'x' || pStr[2] == 'X') {
		*pXValue = 1
	}
	nWords := util.Abc_BitWordNum(nBits)
	Vec_IntFill(vFanins, nWords, 0)
	words := unsafe.Slice((*uint32)(unsafe.Pointer(&Vec_IntArray(vFanins)[0])), int(nWords))
	nDigits := int32(0)
	for nDigits < int32(len(pStr)-2) && prsHexDigit(pStr[2+int(nDigits)]) >= 0 {
		nDigits++
	}
	for k := int32(0); k < nDigits; k++ {
		digit := uint32(prsHexDigit(pStr[2+int(nDigits-1-k)]))
		bit := 4 * k
		if int(bit>>5) < len(words) {
			words[bit>>5] |= digit << uint(bit&31)
		}
	}
	if nDigits != (nBits+3)/4 {
		//        return Wlc_PrsWriteErrorMessage( p, pStr, "The length of a constant does not match." );
		//        printf( "Warning: The length of a constant (%d hex digits) does not match the number of bits (%d).\n", nDigits, nBits );
	}
	*pRange = nBits
	pStr = prsSuffix(pStr, 2)
	for Wlc_PrsIsChar(pStr) != 0 {
		pStr = prsSuffix(pStr, 1)
	}
	return pStr
}

func prsHexDigit(c byte) int32 {
	switch {
	case c >= '0' && c <= '9':
		return int32(c - '0')
	case c >= 'a' && c <= 'f':
		return int32(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int32(c - 'A' + 10)
	default:
		return -1
	}
}

func prsReplaceChar(p *Wlc_Prs_t, suffix string, ch byte) string {
	off := int(Wlc_PrsOffset(p, suffix))
	if off < 0 || off >= len(p.pBuffer) {
		return ""
	}
	b := []byte(p.pBuffer)
	b[off] = ch
	p.pBuffer = string(b)
	return p.pBuffer[off:]
}
func Wlc_PrsReadName(p *Wlc_Prs_t, pStr string, vFanins *Vec_Int_t) string {
	var NameId, fFound, iObj int32
	pStr = Wlc_PrsSkipSpaces(pStr)
	if Wlc_PrsIsDigit(pStr) != 0 {
		var Buffer string // was [100]byte in C

		var Range, Signed, XValue int32
		var vFanins *Vec_Int_t = Vec_IntAlloc(0)
		pStr = Wlc_PrsReadConstant(p, pStr, vFanins, &Range, &Signed, &XValue)
		if pStr == "" {
			Vec_IntFree(vFanins)
			return ""
		}
		// create new node
		iObj = Wlc_ObjAlloc(p.pNtk, int32(WLC_OBJ_CONST), Signed, Range-1, 0)
		Wlc_ObjAddFanins(p.pNtk, Wlc_NtkObj(p.pNtk, iObj), vFanins)
		Wlc_NtkObj(p.pNtk, iObj).fXConst = uint32(XValue)
		Vec_IntFree(vFanins)
		// add node's name
		Buffer = fmt.Sprintf("_c%d_", p.nConsts)
		p.nConsts++
		NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, Buffer, &fFound)
		if fFound != 0 {
			Wlc_PrsWriteErrorMessage(p, pStr, "Name %s is already used.", Buffer)
			return ""
		}
		assert(iObj == NameId, "iObj == NameId")

	} else {
		var pName string
		pStr = Wlc_PrsFindName(pStr, &pName)
		if pStr == "" {
			Wlc_PrsWriteErrorMessage(p, pStr, "Cannot read name in assign-statement.")
			return ""
		}
		NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
		if fFound == 0 {
			Wlc_PrsWriteErrorMessage(p, pStr, "Name %s is used but not declared.", pName)
			return ""
		}
	}
	Vec_IntPush(vFanins, NameId)
	return Wlc_PrsSkipSpaces(pStr)
}
func Wlc_PrsFindDefinition(p *Wlc_Prs_t, pStr string, vFanins *Vec_Int_t, pXValue *int32) int32 {
	var pName string
	var Type int32 = int32(WLC_OBJ_NONE)
	var fRotating int32
	Vec_IntClear(vFanins)
	pStr = Wlc_PrsSkipSpaces(pStr)
	if pStr[0] != '=' {
		return 0
	}
	pStr = Wlc_PrsSkipSpaces(pStr[1:])
	if pStr[0] == '(' {
		// consider rotating shifter
		if Wlc_PrsFindSymbolTwo(pStr, '>', '>') != "" && Wlc_PrsFindSymbolTwo(pStr, '<', '<') != "" {
			// THIS IS A HACK TO DETECT rotating shifters
			var pClose string = Wlc_PrsFindClosingParenthesis(pStr, '(', ')')
			if pClose == "" {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Expecting closing parenthesis.")
			}
			pStr = prsReplaceChar(p, pStr, ' ')
			pClose = prsReplaceChar(p, pClose, 0)
			pStr = Wlc_PrsSkipSpaces(pStr)
			fRotating = 1

		} else {
			var pClose string = Wlc_PrsFindClosingParenthesis(pStr, '(', ')')
			if pClose == "" {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Expecting closing parenthesis.")
			}
			pStr = prsReplaceChar(p, pStr, ' ')
			pClose = prsReplaceChar(p, pClose, ' ')
			pStr = Wlc_PrsSkipSpaces(pStr)
		}
	}
	if Wlc_PrsIsDigit(pStr) != 0 {
		var Range, Signed int32
		pStr = Wlc_PrsReadConstant(p, pStr, vFanins, &Range, &Signed, pXValue)
		if pStr == "" {
			return 0
		}
		Type = int32(WLC_OBJ_CONST)

	} else if pStr[0] == '!' || (pStr[0] == '~' && pStr[1] != '&' && pStr[1] != '|' && pStr[1] != '^') || pStr[0] == '@' || pStr[0] == '#' {
		if pStr[0] == '!' {
			Type = int32(WLC_OBJ_LOGIC_NOT)
		} else if pStr[0] == '~' {
			Type = int32(WLC_OBJ_BIT_NOT)
		} else if pStr[0] == '@' {
			Type = int32(WLC_OBJ_ARI_SQRT)
		} else if pStr[0] == '#' {
			Type = int32(WLC_OBJ_ARI_SQUARE)
		} else {
			assert(false, "0")
		}
		// skip parentheses
		pStr = Wlc_PrsSkipSpaces(pStr[1:])
		if pStr[0] == '(' {
			var pClose string = Wlc_PrsFindClosingParenthesis(pStr, '(', ')')
			if pClose == "" {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Expecting closing parenthesis.")
			}
			pStr = prsReplaceChar(p, pStr, ' ')
			pClose = prsReplaceChar(p, pClose, ' ')
		}
		pStr = Wlc_PrsReadName(p, pStr, vFanins)
		if pStr == "" {
			return Wlc_PrsWriteErrorMessage(p, pStr, "Cannot read name after !.")
		}

	} else if pStr[0] == '-' ||
		pStr[0] == '&' || pStr[0] == '|' || pStr[0] == '^' ||
		(pStr[0] == '~' && pStr[1] == '&') ||
		(pStr[0] == '~' && pStr[1] == '|') ||
		(pStr[0] == '~' && pStr[1] == '^') {
		var shift int32 = 1
		if pStr[0] == '-' {
			Type = int32(WLC_OBJ_ARI_MINUS)
		} else if pStr[0] == '&' {
			Type = int32(WLC_OBJ_REDUCT_AND)
		} else if pStr[0] == '|' {
			Type = int32(WLC_OBJ_REDUCT_OR)
		} else if pStr[0] == '^' {
			Type = int32(WLC_OBJ_REDUCT_XOR)
		} else if pStr[0] == '~' && pStr[1] == '&' {
			Type = int32(WLC_OBJ_REDUCT_NAND)
			shift = 2
		} else if pStr[0] == '~' && pStr[1] == '|' {
			Type = int32(WLC_OBJ_REDUCT_NOR)
			shift = 2
		} else if pStr[0] == '~' && pStr[1] == '^' {
			Type = int32(WLC_OBJ_REDUCT_NXOR)
			shift = 2
		} else {
			assert(false, "0")
		}
		pStr = Wlc_PrsReadName(p, pStr[shift:], vFanins)
		if pStr == "" {
			return Wlc_PrsWriteErrorMessage(p, pStr, "Cannot read name after a unary operator.")
		}

	} else if pStr[0] == '{' {
		// THIS IS A HACK TO DETECT zero padding AND sign extension
		if Wlc_PrsFindSymbol(pStr[1:], '{') != "" {
			if Wlc_PrsFindSymbol(pStr[1:], '\'') != "" {
				Type = int32(WLC_OBJ_BIT_ZEROPAD)
			} else {
				Type = int32(WLC_OBJ_BIT_SIGNEXT)
			}
			pStr = Wlc_PrsFindSymbol(pStr[1:], ',')
			if pStr == "" {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Expecting one comma in this line.")
			}
			pStr = Wlc_PrsReadName(p, pStr[1:], vFanins)
			if pStr == "" {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Cannot read name in sign-extension.")
			}
			pStr = Wlc_PrsSkipSpaces(pStr)
			if pStr[0] != '}' {
				return Wlc_PrsWriteErrorMessage(p, pStr, "There is no closing brace (}).")
			}

		} else { // concatenation

			for {
				pStr = Wlc_PrsSkipSpaces(pStr[1:])
				pStr = Wlc_PrsReadName(p, pStr, vFanins)
				if pStr == "" {
					return Wlc_PrsWriteErrorMessage(p, pStr, "Cannot read name in concatenation.")
				}
				if pStr[0] == '}' {
					break
				}
				if pStr[0] != ',' {
					return Wlc_PrsWriteErrorMessage(p, pStr, "Expected comma (,) in this place.")
				}
			}
			Type = int32(WLC_OBJ_BIT_CONCAT)
		}
		assert(pStr[0] == '}', "pStr[0] == '}'")
		pStr = pStr[1:]

	} else {
		pStr = Wlc_PrsReadName(p, pStr, vFanins)
		if pStr == "" {
			return 0
		}
		// get the next symbol
		if pStr[0] == 0 {
			Type = int32(WLC_OBJ_BUF)
		} else if pStr[0] == '?' {
			pStr = Wlc_PrsReadName(p, pStr[1:], vFanins)
			if pStr == "" {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Cannot read name in MUX.")
			}
			if pStr[0] != ':' {
				return Wlc_PrsWriteErrorMessage(p, pStr, "MUX lacks the colon symbol (:).")
			}
			pStr = Wlc_PrsReadName(p, pStr[1:], vFanins)
			if pStr == "" {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Cannot read name in MUX.")
			}
			assert(Vec_IntSize(vFanins) == 3, "Vec_IntSize(vFanins) == 3")
			entries := Vec_IntArray(vFanins)
			entries[1], entries[2] = entries[2], entries[1]
			Type = int32(WLC_OBJ_MUX)

		} else if pStr[0] == '[' {
			var End, Beg int32
			var pLine string = pStr
			pStr = Wlc_PrsFindRange(pStr, &End, &Beg)
			if pStr == "" {
				return Wlc_PrsWriteErrorMessage(p, pLine, "Non-standard range.")
			}
			Vec_IntPushTwo(vFanins, End, Beg)
			Type = int32(WLC_OBJ_BIT_SELECT)

		} else {
			if pStr[0] == '>' && pStr[1] == '>' && pStr[2] != '>' {
				pStr = pStr[2:]
				if fRotating != 0 {
					Type = int32(WLC_OBJ_ROTATE_R)
				} else {
					Type = int32(WLC_OBJ_SHIFT_R)
				}
			} else if pStr[0] == '>' && pStr[1] == '>' && pStr[2] == '>' {
				pStr = pStr[3:]
				Type = int32(WLC_OBJ_SHIFT_RA)
			} else if pStr[0] == '<' && pStr[1] == '<' && pStr[2] != '<' {
				pStr = pStr[2:]
				if fRotating != 0 {
					Type = int32(WLC_OBJ_ROTATE_L)
				} else {
					Type = int32(WLC_OBJ_SHIFT_L)
				}
			} else if pStr[0] == '<' && pStr[1] == '<' && pStr[2] == '<' {
				pStr = pStr[3:]
				Type = int32(WLC_OBJ_SHIFT_LA)
			} else if pStr[0] == '&' && pStr[1] != '&' {
				pStr = pStr[1:]
				Type = int32(WLC_OBJ_BIT_AND)
			} else if pStr[0] == '|' && pStr[1] != '|' {
				pStr = pStr[1:]
				Type = int32(WLC_OBJ_BIT_OR)
			} else if pStr[0] == '^' && pStr[1] != '^' {
				pStr = pStr[1:]
				Type = int32(WLC_OBJ_BIT_XOR)
			} else if pStr[0] == '~' && pStr[1] == '&' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_BIT_NAND)
			} else if pStr[0] == '~' && pStr[1] == '|' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_BIT_NOR)
			} else if pStr[0] == '~' && pStr[1] == '^' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_BIT_NXOR)
			} else if pStr[0] == '=' && pStr[1] == '>' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_LOGIC_IMPL)
			} else if pStr[0] == '&' && pStr[1] == '&' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_LOGIC_AND)
			} else if pStr[0] == '|' && pStr[1] == '|' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_LOGIC_OR)
			} else if pStr[0] == '^' && pStr[1] == '^' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_LOGIC_XOR)
			} else if pStr[0] == '=' && pStr[1] == '=' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_COMP_EQU)
			} else if pStr[0] == '!' && pStr[1] == '=' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_COMP_NOTEQU)
			} else if pStr[0] == '<' && pStr[1] != '=' {
				pStr = pStr[1:]
				Type = int32(WLC_OBJ_COMP_LESS)
			} else if pStr[0] == '>' && pStr[1] != '=' {
				pStr = pStr[1:]
				Type = int32(WLC_OBJ_COMP_MORE)
			} else if pStr[0] == '<' && pStr[1] == '=' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_COMP_LESSEQU)
			} else if pStr[0] == '>' && pStr[1] == '=' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_COMP_MOREEQU)
			} else if pStr[0] == '+' {
				pStr = pStr[1:]
				Type = int32(WLC_OBJ_ARI_ADD)
			} else if pStr[0] == '-' {
				pStr = pStr[1:]
				Type = int32(WLC_OBJ_ARI_SUB)
			} else if pStr[0] == '*' && pStr[1] != '*' {
				pStr = pStr[1:]
				Type = int32(WLC_OBJ_ARI_MULTI)
			} else if pStr[0] == '/' {
				pStr = pStr[1:]
				Type = int32(WLC_OBJ_ARI_DIVIDE)
			} else if pStr[0] == '%' {
				pStr = pStr[1:]
				Type = int32(WLC_OBJ_ARI_REM)
			} else if pStr[0] == '*' && pStr[1] == '*' {
				pStr = pStr[2:]
				Type = int32(WLC_OBJ_ARI_POWER)
			} else {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Unsupported operation (%c).", pStr[0])
			}
			pStr = Wlc_PrsReadName(p, pStr[1:], vFanins)
			if pStr == "" {
				return 0
			}
			pStr = Wlc_PrsSkipSpaces(pStr)
			if Type == int32(WLC_OBJ_ARI_ADD) && len(pStr) > 0 && pStr[0] == '+' {
				pStr = Wlc_PrsReadName(p, pStr[1:], vFanins)
				if pStr == "" {
					return 0
				}
				pStr = Wlc_PrsSkipSpaces(pStr)
			}
			if len(pStr) != 0 && pStr[0] != 0 {
				fmt.Printf("Warning: Trailing symbols \"%s\" in line %d.\n", pStr, Wlc_PrsFindLine(p, pStr))
			}
		}
	}
	// make sure there is nothing left there
	if pStr != "" {
		pStr = Wlc_PrsFindName(pStr, &pName)
		if pStr != "" {
			return Wlc_PrsWriteErrorMessage(p, pStr, "Name %s is left at the end of the line.", pName)
		}
	}
	return Type
}

//ABC_NAMESPACE_IMPL_END
