package wlc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"
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
	return c.advance(len(word)), false
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

func Wlc_PrsOffset(p *Wlc_Prs_t, pStr string) int32 { return pStr - p.pBuffer }
func Wlc_PrsStr(p *Wlc_Prs_t, iOffset int32) string { return p.pBuffer + iOffset }
func Wlc_PrsStrCmp(pStr string, pWhat string) int32 { return !strncmp(pStr, pWhat, len(pWhat)) }

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
	if pFileName && !Extra_FileCheck(pFileName) {
		return nil
	}
	p = new(Wlc_Prs_t)
	p.pFileName = pFileName
	if pStr {
		p.pBuffer = Abc_UtilStrsav(pStr)
	} else {
		p.pBuffer = Extra_FileReadContents(pFileName)
	}
	p.nFileSize = len(p.pBuffer)
	assert(p.nFileSize > 0, "p.nFileSize > 0")
	p.vLines = Vec_IntAlloc(p.nFileSize / 50)
	p.vStarts = Vec_IntAlloc(p.nFileSize / 50)
	p.vFanins = Vec_IntAlloc(100)
	p.vTables = Vec_PtrAlloc(1000)
	p.pMemTable = Mem_FlexStart()
	return p
}
func Wlc_PrsStop(p *Wlc_Prs_t) {
	if p.pNtk {
		Wlc_NtkFree(p.pNtk)
	}
	if p.pMemTable {
		Mem_FlexStop(p.pMemTable, 0)
	}
	Vec_StrFreeP(&p.vPoPairs)
	Vec_PtrFreeP(&p.vTables)
	Vec_IntFree(p.vLines)
	Vec_IntFree(p.vStarts)
	Vec_IntFree(p.vFanins)
	ABC_FREE(p.pBuffer)
	ABC_FREE(p)
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
	return (pStr[0] >= '0' && pStr[0] <= '9')
}
func Wlc_PrsIsChar(pStr string) int32 {
	return (pStr[0] >= 'a' && pStr[0] <= 'z') ||
		(pStr[0] >= 'A' && pStr[0] <= 'Z') ||
		(pStr[0] >= '0' && pStr[0] <= '9') ||
		pStr[0] == '_' || pStr[0] == '$' || pStr[0] == '\\'
}
func Wlc_PrsSkipSpaces(pStr string) string {
	for *pStr && *pStr == ' ' {
		pStr++
	}
	return pStr
}
func Wlc_PrsFindSymbol(pStr string, Symb byte) string {
	var fNotName int32 = 1
	for ; *pStr; pStr++ {
		if (fNotName != 0) && *pStr == Symb {
			return pStr
		}
		if pStr[0] == '\\' {
			fNotName = 0
		} else if !(fNotName != 0) && *pStr == ' ' {
			fNotName = 1
		}
	}
	return nil
}
func Wlc_PrsFindSymbolTwo(pStr string, Symb byte, Symb2 byte) string {
	for ; pStr[1]; pStr++ {
		if pStr[0] == Symb && pStr[1] == Symb2 {
			return pStr
		}
	}
	return nil
}
func Wlc_PrsFindClosingParenthesis(pStr string, Open byte, Close byte) string {
	var Counter int32
	var fNotName int32 = 1
	assert(*pStr == Open, "*pStr == Open")
	for ; *pStr; pStr++ {
		if fNotName != 0 {
			if *pStr == Open {
				Counter++
			}
			if *pStr == Close {
				Counter--
			}
			if Counter == 0 {
				return pStr
			}
		}
		if *pStr == '\\' {
			fNotName = 0
		} else if !(fNotName != 0) && *pStr == ' ' {
			fNotName = 1
		}
	}
	return nil
}
func Wlc_PrsRemoveComments(p *Wlc_Prs_t) int32 {
	var fSpecifyFound int32
	var pCur string
	var pNext string
	var pEnd string = p.pBuffer + p.nFileSize
	for pCur = p.pBuffer; pCur < pEnd; pCur++ {
		// regular comment (//)
		if *pCur == '/' && pCur[1] == '/' {
			if pCur+5 < pEnd && pCur[2] == 'a' && pCur[3] == 'b' && pCur[4] == 'c' && pCur[5] == '2' {
				for i := 0; i < 6; i++ {
					pCur[i] = ' '
				}
			} else if !strncmp(pCur+3, "Pair:", 5) {
				if p.vPoPairs == nil {
					p.vPoPairs = Vec_StrAlloc(100)
				}
				for pNext = pCur + 9; *pNext != '\n'; pNext++ {
					if *pNext == ' ' {
						Vec_StrPush(p.vPoPairs, '\x00')
					} else if *pNext != '\r' {
						Vec_StrPush(p.vPoPairs, *pNext)
					}
				}
				if Vec_StrEntryLast(p.vPoPairs) != 0 {
					Vec_StrPush(p.vPoPairs, 0)
				}

			} else {
				pNext = Wlc_PrsFindSymbol(pCur, '\n')
				if pNext == nil {
					return Wlc_PrsWriteErrorMessage(p, pCur, "Cannot find end-of-line after symbols \"//\".")
				}
				for ; pCur < pNext; pCur++ {
					*pCur = ' '
				}
			}

			// skip preprocessor directive (`timescale, `celldefine, etc)
		} else if *pCur == '`' {
			pNext = Wlc_PrsFindSymbol(pCur, '\n')
			if pNext == nil {
				return Wlc_PrsWriteErrorMessage(p, pCur, "Cannot find end-of-line after symbols \"`\".")
			}
			for ; pCur < pNext; pCur++ {
				*pCur = ' '
			}

			// regular comment (/* ... */)
		} else if *pCur == '/' && pCur[1] == '*' {
			pNext = Wlc_PrsFindSymbolTwo(pCur, '*', '/')
			if pNext == nil {
				return Wlc_PrsWriteErrorMessage(p, pCur, "Cannot find symbols \"*/\" after symbols \"/*\".")
			}
			// overwrite comment
			for ; pCur < pNext+2; pCur++ {
				*pCur = ' '
			}

			// 'specify' treated as comments
		} else if *pCur == 's' && pCur[1] == 'p' && pCur[2] == 'e' && !strncmp(pCur, "specify", 7) {
			for pNext = pCur; pNext < pEnd-10; pNext++ {
				if *pNext == 'e' && pNext[1] == 'n' && pNext[2] == 'd' && !strncmp(pNext, "endspecify", 10) {
					// overwrite comment
					for ; pCur < pNext+10; pCur++ {
						*pCur = ' '
					}
					if fSpecifyFound == 0 {
						Abc_Print(0, "Ignoring specify/endspecify directives.\n")
					}
					fSpecifyFound = 1
					break
				}
			}

			// insert semicolons
		} else if *pCur == 'e' && pCur[1] == 'n' && pCur[2] == 'd' && !strncmp(pCur, "endmodule", 9) {
			pCur[len("endmodule")] = ';'
		}
		// overwrite end-of-lines with spaces (less checking to do later on)
		if *pCur == '\n' || *pCur == '\r' || *pCur == '\t' {
			*pCur = ' '
		}
	}
	return 1
}
func Wlc_PrsPrepare(p *Wlc_Prs_t) int32 {
	var fPrettyPrint int32
	var fNotName int32 = 1
	var pTemp, pPrev, pThis string
	// collect info about lines
	assert(Vec_IntSize(p.vLines) == 0, "Vec_IntSize(p.vLines) == 0")
	for pTemp = p.pBuffer; *pTemp; pTemp++ {
		if *pTemp == '\n' {
			Vec_IntPush(p.vLines, pTemp-p.pBuffer)
		}
	}
	// delete comments and insert breaks
	if !(Wlc_PrsRemoveComments(p) != 0) {
		return 0
	}
	// collect info about breaks
	assert(Vec_IntSize(p.vStarts) == 0, "Vec_IntSize(p.vStarts) == 0")
	for pPrev, pThis = p.pBuffer, p.pBuffer; *pThis; pThis++ {
		if (fNotName != 0) && *pThis == ';' {
			*pThis = 0
			Vec_IntPush(p.vStarts, Wlc_PrsOffset(p, Wlc_PrsSkipSpaces(pPrev)))
			pPrev = pThis + 1
		}
		if *pThis == '\\' {
			fNotName = 0
		} else if !(fNotName != 0) && *pThis == ' ' {
			fNotName = 1
		}
	}

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
			for k = 1; pTemp[k]; k++ {
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
	s = strings.TrimLeft(s, delim)
	if s == "" {
		spStatic_Wlc_PrsStrtok_last = ""
		return ""
	}
	end := len(s)
	rest := ""
	if s[0] == '\\' {
		if space := strings.IndexByte(s, ' '); space >= 0 {
			end = space
			rest = s[space+1:]
		}
	} else if i := strings.IndexAny(s, delim); i >= 0 {
		end = i
		rest = s[i+1:]
	}
	token := s[:end]
	spStatic_Wlc_PrsStrtok_last = rest
	return token
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
	vStr := Vec_StrAlloc(1000)
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
		for pObj.Type == WLC_OBJ_BUF {
			pObj = Wlc_NtkObj(p, Wlc_ObjFaninId0(pObj))
		}

		var pInits *int32
		if pObj.Type == WLC_OBJ_CONST && pObj.fXConst == 0 {
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
	result := Vec_StrReleaseArray(vStr)
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
	for pObj.Type == WLC_OBJ_BUF {
		pObj = Wlc_NtkObj(p, Wlc_ObjFaninId0(pObj))
	}
	if pObj.Type != WLC_OBJ_CONST {
		return 0
	}
	pInits = Wlc_ObjConstValue(pObj)
	return Abc_InfoHasBit((*uint32)(unsafe.Pointer(pInits)), 0) == 0
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
	if pStr[0] != '[' {
		return pStr
	}
	pStr = Wlc_PrsSkipSpaces(pStr + 1)
	if !(Wlc_PrsIsDigit(pStr) != 0) && pStr[0] != '-' {
		return nil
	}
	n, _ := strconv.Atoi(pStr)
	*End = int32(n)
	*Beg = int32(n)
	if Wlc_PrsFindSymbol(pStr, ':') == nil {
		pStr = Wlc_PrsFindSymbol(pStr, ']')
		if pStr == nil {
			return nil
		}

	} else {
		pStr = Wlc_PrsFindSymbol(pStr, ':')
		pStr = Wlc_PrsSkipSpaces(pStr + 1)
		if !(Wlc_PrsIsDigit(pStr) != 0) && pStr[0] != '-' {
			return nil
		}
		*Beg = strconv.Atoi(pStr)
		pStr = Wlc_PrsFindSymbol(pStr, ']')
		if pStr == nil {
			return nil
		}
	}
	return pStr + 1
}
func Wlc_PrsFindWord(pStr string, pWord string, fFound *int32) string {
	*fFound = 0
	pStr = Wlc_PrsSkipSpaces(pStr)
	if !(Wlc_PrsStrCmp(pStr, pWord) != 0) {
		return pStr
	}
	*fFound = 1
	return pStr + len(pWord)
}

var spStatic_Wlc_PrsFindName_Buffer [WLV_PRS_MAX_LINE]byte

func Wlc_PrsFindName(pStr string, ppPlace *string) string {

	pStr = Wlc_PrsSkipSpaces(pStr)
	if pStr == "" || Wlc_PrsIsChar(pStr) == 0 {
		return ""
	}
	escaped := false
	for i := 0; i < len(pStr); i++ {
		if !escaped && Wlc_PrsIsChar(pStr[i:]) == 0 {
			break
		}
		if pStr[i] == '\\' {
			escaped = true
		} else if escaped && pStr[i] == ' ' {
			escaped = false
		}
	}
	consumed := len(pStr)
	// Escaped Verilog identifiers include their trailing space in the token.
	if escaped {
		if i := strings.IndexByte(pStr, ' '); i >= 0 {
			consumed = i + 1
		}
	} else {
		for i := 0; i < len(pStr); i++ {
			if Wlc_PrsIsChar(pStr[i:]) == 0 {
				consumed = i
				break
			}
		}
	}
	*ppPlace = pStr[:consumed]
	return pStr[consumed:]
}
func Wlc_PrsReadConstant(p *Wlc_Prs_t, pStr string, vFanins *Vec_Int_t, pRange *int32, pSigned *int32, pXValue *int32) string {
	var i int32
	var nDigits int32
	var nBits int32 = strconv.Atoi(pStr)
	*pRange = -1
	*pSigned = 0
	*pXValue = 0
	pStr = Wlc_PrsSkipSpaces(pStr)
	if Wlc_PrsFindSymbol(pStr, '\'') == nil {
		// handle decimal number
		var Number int32 = strconv.Atoi(pStr)
		*pRange = Abc_Base2Log(Number + 1)
		assert(*pRange < 32, "*pRange < 32")
		for Wlc_PrsIsDigit(pStr) != 0 {
			pStr++
		}
		Vec_IntFill(vFanins, 1, Number)
		return pStr
	}
	pStr = Wlc_PrsFindSymbol(pStr, '\'')
	if pStr[1] == 's' {
		*pSigned = 1
		pStr++
	}
	if pStr[1] == 'b' {
		Vec_IntFill(vFanins, Abc_BitWordNum(nBits), 0)
		for i = 0; i < nBits; i++ {
			if pStr[2+i] == '1' {
				Abc_InfoSetBit((*uint32)(unsafe.Pointer(&Vec_IntArray(vFanins)[0])), nBits-1-i)
			} else if pStr[2+i] != '0' {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Wrong digit in binary constant \"%c\".", pStr[2+i])
			}
		}
		*pRange = nBits
		pStr += 2 + nBits
		return pStr
	}
	if pStr[1] != 'h' {
		return Wlc_PrsWriteErrorMessage(p, pStr, "Expecting hexadecimal constant and not \"%c\".", pStr[1])
	}
	*pXValue = (pStr[2] == 'x' || pStr[2] == 'X')
	if *pXValue == 'X' {
		*pXValue = 'x'
	}
	Vec_IntFill(vFanins, Abc_BitWordNum(nBits), 0)
	nDigits = Abc_TtReadHexNumber((*uint32)(unsafe.Pointer(&Vec_IntArray(vFanins)[0])), pStr+2)
	if nDigits != (nBits+3)/4 {
		//        return Wlc_PrsWriteErrorMessage( p, pStr, "The length of a constant does not match." );
		//        printf( "Warning: The length of a constant (%d hex digits) does not match the number of bits (%d).\n", nDigits, nBits );
	}
	*pRange = nBits
	pStr += 2
	for Wlc_PrsIsChar(pStr) != 0 {
		pStr++
	}
	return pStr
}
func Wlc_PrsReadName(p *Wlc_Prs_t, pStr string, vFanins *Vec_Int_t) string {
	var NameId, fFound, iObj int32
	pStr = Wlc_PrsSkipSpaces(pStr)
	if Wlc_PrsIsDigit(pStr) != 0 {
		var Buffer string // was [100]byte in C

		var Range, Signed, XValue int32
		var vFanins *Vec_Int_t = Vec_IntAlloc(0)
		pStr = Wlc_PrsReadConstant(p, pStr, vFanins, &Range, &Signed, &XValue)
		if pStr == nil {
			Vec_IntFree(vFanins)
			return 0
		}
		// create new node
		iObj = Wlc_ObjAlloc(p.pNtk, WLC_OBJ_CONST, Signed, Range-1, 0)
		Wlc_ObjAddFanins(p.pNtk, Wlc_NtkObj(p.pNtk, iObj), vFanins)
		Wlc_NtkObj(p.pNtk, iObj).fXConst = XValue
		Vec_IntFree(vFanins)
		// add node's name
		Buffer = fmt.Sprintf("_c%d_", p.nConsts)
		p.nConsts++
		NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, Buffer, &fFound)
		if fFound {
			return Wlc_PrsWriteErrorMessage(p, pStr, "Name %s is already used.", Buffer)
		}
		assert(iObj == NameId, "iObj == NameId")

	} else {
		var pName string
		pStr = Wlc_PrsFindName(pStr, &pName)
		if pStr == nil {
			return Wlc_PrsWriteErrorMessage(p, pStr, "Cannot read name in assign-statement.")
		}
		NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
		if !fFound {
			return Wlc_PrsWriteErrorMessage(p, pStr, "Name %s is used but not declared.", pName)
		}
	}
	Vec_IntPush(vFanins, NameId)
	return Wlc_PrsSkipSpaces(pStr)
}
func Wlc_PrsFindDefinition(p *Wlc_Prs_t, pStr string, vFanins *Vec_Int_t, pXValue *int32) int32 {
	var pName string
	var Type int32 = WLC_OBJ_NONE
	var fRotating int32
	Vec_IntClear(vFanins)
	pStr = Wlc_PrsSkipSpaces(pStr)
	if pStr[0] != '=' {
		return 0
	}
	pStr = Wlc_PrsSkipSpaces(pStr + 1)
	if pStr[0] == '(' {
		// consider rotating shifter
		if Wlc_PrsFindSymbolTwo(pStr, '>', '>') && Wlc_PrsFindSymbolTwo(pStr, '<', '<') {
			// THIS IS A HACK TO DETECT rotating shifters
			var pClose string = Wlc_PrsFindClosingParenthesis(pStr, '(', ')')
			if pClose == nil {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Expecting closing parenthesis.")
			}
			*pStr = ' '
			*pClose = 0
			pStr = Wlc_PrsSkipSpaces(pStr)
			fRotating = 1

		} else {
			var pClose string = Wlc_PrsFindClosingParenthesis(pStr, '(', ')')
			if pClose == nil {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Expecting closing parenthesis.")
			}
			*pStr = ' '
			*pClose = ' '
			pStr = Wlc_PrsSkipSpaces(pStr)
		}
	}
	if Wlc_PrsIsDigit(pStr) != 0 {
		var Range, Signed int32
		pStr = Wlc_PrsReadConstant(p, pStr, vFanins, &Range, &Signed, pXValue)
		if pStr == nil {
			return 0
		}
		Type = WLC_OBJ_CONST

	} else if pStr[0] == '!' || (pStr[0] == '~' && pStr[1] != '&' && pStr[1] != '|' && pStr[1] != '^') || pStr[0] == '@' || pStr[0] == '#' {
		if pStr[0] == '!' {
			Type = WLC_OBJ_LOGIC_NOT
		} else if pStr[0] == '~' {
			Type = WLC_OBJ_BIT_NOT
		} else if pStr[0] == '@' {
			Type = WLC_OBJ_ARI_SQRT
		} else if pStr[0] == '#' {
			Type = WLC_OBJ_ARI_SQUARE
		} else {
			assert(false, "0")
		}
		// skip parentheses
		pStr = Wlc_PrsSkipSpaces(pStr + 1)
		if pStr[0] == '(' {
			var pClose string = Wlc_PrsFindClosingParenthesis(pStr, '(', ')')
			if pClose == nil {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Expecting closing parenthesis.")
			}
			*pStr = ' '
			*pClose = ' '
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
			Type = WLC_OBJ_ARI_MINUS
		} else if pStr[0] == '&' {
			Type = WLC_OBJ_REDUCT_AND
		} else if pStr[0] == '|' {
			Type = WLC_OBJ_REDUCT_OR
		} else if pStr[0] == '^' {
			Type = WLC_OBJ_REDUCT_XOR
		} else if pStr[0] == '~' && pStr[1] == '&' {
			Type = WLC_OBJ_REDUCT_NAND
			shift = 2
		} else if pStr[0] == '~' && pStr[1] == '|' {
			Type = WLC_OBJ_REDUCT_NOR
			shift = 2
		} else if pStr[0] == '~' && pStr[1] == '^' {
			Type = WLC_OBJ_REDUCT_NXOR
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
		if Wlc_PrsFindSymbol(pStr+1, '{') {
			if Wlc_PrsFindSymbol(pStr+1, '\'') {
				Type = WLC_OBJ_BIT_ZEROPAD
			} else {
				Type = WLC_OBJ_BIT_SIGNEXT
			}
			pStr = Wlc_PrsFindSymbol(pStr+1, ',')
			if pStr == nil {
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
				pStr = Wlc_PrsSkipSpaces(pStr + 1)
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
			Type = WLC_OBJ_BIT_CONCAT
		}
		assert(pStr[0] == '}', "pStr[0] == '}'")
		pStr++

	} else {
		pStr = Wlc_PrsReadName(p, pStr, vFanins)
		if pStr == "" {
			return 0
		}
		// get the next symbol
		if pStr[0] == 0 {
			Type = WLC_OBJ_BUF
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
			ABC_SWAP(int, Vec_IntArray(vFanins)[1], Vec_IntArray(vFanins)[2])
			Type = WLC_OBJ_MUX

		} else if pStr[0] == '[' {
			var End, Beg int32
			var pLine string = pStr
			pStr = Wlc_PrsFindRange(pStr, &End, &Beg)
			if pStr == nil {
				return Wlc_PrsWriteErrorMessage(p, pLine, "Non-standard range.")
			}
			Vec_IntPushTwo(vFanins, End, Beg)
			Type = WLC_OBJ_BIT_SELECT

		} else {
			if pStr[0] == '>' && pStr[1] == '>' && pStr[2] != '>' {
				pStr = pStr[2:]
				Type = WLC_OBJ_SHIFT_R
			} else if pStr[0] == '>' && pStr[1] == '>' && pStr[2] == '>' {
				pStr = pStr[3:]
				Type = WLC_OBJ_SHIFT_RA
			} else if pStr[0] == '<' && pStr[1] == '<' && pStr[2] != '<' {
				pStr = pStr[2:]
				Type = WLC_OBJ_SHIFT_L
			} else if pStr[0] == '<' && pStr[1] == '<' && pStr[2] == '<' {
				pStr = pStr[3:]
				Type = WLC_OBJ_SHIFT_LA
			} else if pStr[0] == '&' && pStr[1] != '&' {
				pStr = pStr[1:]
				Type = WLC_OBJ_BIT_AND
			} else if pStr[0] == '|' && pStr[1] != '|' {
				pStr = pStr[1:]
				Type = WLC_OBJ_BIT_OR
			} else if pStr[0] == '^' && pStr[1] != '^' {
				pStr = pStr[1:]
				Type = WLC_OBJ_BIT_XOR
			} else if pStr[0] == '~' && pStr[1] == '&' {
				pStr = pStr[2:]
				Type = WLC_OBJ_BIT_NAND
			} else if pStr[0] == '~' && pStr[1] == '|' {
				pStr = pStr[2:]
				Type = WLC_OBJ_BIT_NOR
			} else if pStr[0] == '~' && pStr[1] == '^' {
				pStr = pStr[2:]
				Type = WLC_OBJ_BIT_NXOR
			} else if pStr[0] == '=' && pStr[1] == '>' {
				pStr = pStr[2:]
				Type = WLC_OBJ_LOGIC_IMPL
			} else if pStr[0] == '&' && pStr[1] == '&' {
				pStr = pStr[2:]
				Type = WLC_OBJ_LOGIC_AND
			} else if pStr[0] == '|' && pStr[1] == '|' {
				pStr = pStr[2:]
				Type = WLC_OBJ_LOGIC_OR
			} else if pStr[0] == '^' && pStr[1] == '^' {
				pStr = pStr[2:]
				Type = WLC_OBJ_LOGIC_XOR
			} else if pStr[0] == '=' && pStr[1] == '=' {
				pStr = pStr[2:]
				Type = WLC_OBJ_COMP_EQU
			} else if pStr[0] == '!' && pStr[1] == '=' {
				pStr = pStr[2:]
				Type = WLC_OBJ_COMP_NOTEQU
			} else if pStr[0] == '<' && pStr[1] != '=' {
				pStr = pStr[1:]
				Type = WLC_OBJ_COMP_LESS
			} else if pStr[0] == '>' && pStr[1] != '=' {
				pStr = pStr[1:]
				Type = WLC_OBJ_COMP_MORE
			} else if pStr[0] == '<' && pStr[1] == '=' {
				pStr = pStr[2:]
				Type = WLC_OBJ_COMP_LESSEQU
			} else if pStr[0] == '>' && pStr[1] == '=' {
				pStr = pStr[2:]
				Type = WLC_OBJ_COMP_MOREEQU
			} else if pStr[0] == '+' {
				pStr = pStr[1:]
				Type = WLC_OBJ_ARI_ADD
			} else if pStr[0] == '-' {
				pStr = pStr[1:]
				Type = WLC_OBJ_ARI_SUB
			} else if pStr[0] == '*' && pStr[1] != '*' {
				pStr = pStr[1:]
				Type = WLC_OBJ_ARI_MULTI
			} else if pStr[0] == '/' {
				pStr = pStr[1:]
				Type = WLC_OBJ_ARI_DIVIDE
			} else if pStr[0] == '%' {
				pStr = pStr[1:]
				Type = WLC_OBJ_ARI_REM
			} else if pStr[0] == '*' && pStr[1] == '*' {
				pStr = pStr[2:]
				Type = WLC_OBJ_ARI_POWER
			} else {
				return Wlc_PrsWriteErrorMessage(p, pStr, "Unsupported operation (%c).", pStr[0])
			}
			pStr = Wlc_PrsReadName(p, pStr[1:], vFanins)
			if pStr == "" {
				return 0
			}
			pStr = Wlc_PrsSkipSpaces(pStr)
			if Type == WLC_OBJ_ARI_ADD && pStr[0] == '+' {
				pStr = Wlc_PrsReadName(p, pStr[1:], vFanins)
				if pStr == "" {
					return 0
				}
				pStr = Wlc_PrsSkipSpaces(pStr)
			}
			if pStr[0] {
				fmt.Printf("Warning: Trailing symbols \"%s\" in line %d.\n", pStr, Wlc_PrsFindLine(p, pStr))
			}
		}
	}
	// make sure there is nothing left there
	if pStr {
		pStr = Wlc_PrsFindName(pStr, &pName)
		if pStr != nil {
			return Wlc_PrsWriteErrorMessage(p, pStr, "Name %s is left at the end of the line.", pName)
		}
	}
	return Type
}
func Wlc_PrsReadDeclaration(p *Wlc_Prs_t, pStart string) int32 {
	var fFound int32
	var Type int32 = WLC_OBJ_NONE
	var iObj int32
	var pLine string
	var Signed, Beg, End, NameId, fIsPo int32
	if Wlc_PrsStrCmp(pStart, "input") != 0 {
		pStart = pStart[len("input"):]
		Type = WLC_OBJ_PI
	} else if Wlc_PrsStrCmp(pStart, "output") != 0 {
		pStart = pStart[len("output"):]
		fIsPo = 1
	}
	pStart = Wlc_PrsSkipSpaces(pStart)
	if Wlc_PrsStrCmp(pStart, "wire") != 0 {
		pStart += len("wire")
	} else if Wlc_PrsStrCmp(pStart, "reg") != 0 {
		pStart += len("reg")
	}
	// read 'signed'
	pStart = Wlc_PrsFindWord(pStart, "signed", &Signed)
	// read range
	pLine = pStart
	pStart = Wlc_PrsFindRange(pStart, &End, &Beg)
	if pStart == nil {
		return Wlc_PrsWriteErrorMessage(p, pLine, "Non-standard range.")
	}
	if End != 0 && Beg != 0 {
		if p.nNonZero[0] == 0 {
			p.nNonZero[0]++
			p.nNonZero[1] = End
			p.nNonZero[2] = Beg
			p.nNonZero[3] = Wlc_PrsFindLine(p, pStart)
		}
	}
	if End < 0 || Beg < 0 {
		if p.nNegative[0] == 0 {
			p.nNegative[0]++
			p.nNegative[1] = End
			p.nNegative[2] = Beg
			p.nNegative[3] = Wlc_PrsFindLine(p, pStart)
		}
	}
	if End < Beg {
		if p.nReverse[0] == 0 {
			p.nReverse[0]++
			p.nReverse[1] = End
			p.nReverse[2] = Beg
			p.nReverse[3] = Wlc_PrsFindLine(p, pStart)
		}
	}
	for {
		var pName string
		var XValue, TypeNew int32
		// read name
		pStart = Wlc_PrsFindName(pStart, &pName)
		if pStart == nil {
			return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name in declaration.")
		}
		NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
		if fFound {
			return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is declared more than once.", pName)
		}
		iObj = Wlc_ObjAlloc(p.pNtk, Type, Signed, End, Beg)
		if fIsPo != 0 {
			Wlc_ObjSetCo(p.pNtk, Wlc_NtkObj(p.pNtk, iObj), 0)
		}
		assert(iObj == NameId, "iObj == NameId")
		// check next definition
		pStart = Wlc_PrsSkipSpaces(pStart)
		if pStart[0] == ',' {
			pStart++
			continue
		}
		// check definition
		TypeNew = Wlc_PrsFindDefinition(p, pStart, p.vFanins, &XValue)
		if TypeNew != 0 {
			var pObj *Wlc_Obj_t = Wlc_NtkObj(p.pNtk, iObj)
			Wlc_ObjUpdateType(p.pNtk, pObj, TypeNew)
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)
			pObj.fXConst = XValue
		}
		break
	}
	return 1
}
func Wlc_PrsDerive(p *Wlc_Prs_t) int32 {
	var pObj *Wlc_Obj_t
	var pStart, pName string
	var i int32
	// go through the directives
	// inlining of macro Wlc_PrsForEachLine
	for i = int32(0); i < Vec_IntSize(p.vStarts); i++ {
		pStart = Wlc_PrsStr(p, Vec_IntEntry(p.vStarts, i))
		start := wlcPrsCursorAt(p, int(Vec_IntEntry(p.vStarts, i)))
	startword:
		module, isModule := start.findWord("module")
		if isModule {
			// get module name
			nameEnd, parsedName := module.advance(len("module")).findName()
			pName = parsedName
			spStatic_Wlc_PrsStrtok_last = nameEnd.rest()
			if pName == "" {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read model name.")
			}
			// THIS IS A HACK to skip definitions of modules beginning with "CPL_"
			if strings.HasPrefix(pName, "CPL_") {
				for i++; i < Vec_IntSize(p.vStarts); i++ {
					start = wlcPrsCursorAt(p, int(Vec_IntEntry(p.vStarts, i)))
					if end, found := start.findSubstring("endmodule"); found {
						start = end
						pStart = start.rest()
						break
					}
				}
				continue
			}
			if pName == "table" {
				// THIS IS A HACK to detect table module descriptions
				var Width1 int32 = -1
				var Width2 int32 = -1
				var v, b, Value, nBits, nInts int32
				var pTable *uint32
				var vValues *Vec_Int_t = Vec_IntAlloc(256)
				// inlining of macro Wlc_PrsForEachLineStart
				for i = i + 1; i < Vec_IntSize((p).vStarts); i++ {
					start = wlcPrsCursorAt(p, int(Vec_IntEntry(p.vStarts, i)))
					pStart = start.rest()
					if endcase, found := start.findWord("endcase"); found {
						pStart = endcase.rest()
						break
					}
					quote := start.findSymbol('\'')
					if !quote.valid() {
						continue
					}
					Width1 = atoi(quote.retreat(1).rest())
					quote = quote.advance(2).findSymbol('\'')
					if !quote.valid() {
						continue
					}
					Width2 = atoi(quote.retreat(1).rest())
					Value = 0
					var parsed [1]word
					Abc_TtReadHexNumber(parsed[:], quote.advance(2).rest())
					Value = int32(parsed[0])
					Vec_IntPush(vValues, Value)
				}
				//Vec_IntPrint( vValues );
				nBits = Abc_Base2Log(Vec_IntSize(vValues))
				if Vec_IntSize(vValues) != (1 << nBits) {
					Vec_IntFree(vValues)
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read module \"%s\".", pName)
				}
				assert(Width1 == nBits, "Width1 == nBits")
				// create bitmap
				nInts = Abc_BitWordNum(Width2 * Vec_IntSize(vValues))
				tableBytes := Mem_FlexEntryFetch(p.pMemTable, nInts*4)
				pTable = unsafe.Slice((*uint32)(unsafe.Pointer(&tableBytes[0])), int(nInts))
				clear(pTable)
				for v = 0; v < Vec_IntSize(vValues); v++ {
					Value = Vec_IntEntry(vValues, v)
					for b = 0; b < Width2; b++ {
						if (Value>>b)&1 != 0 {
							Abc_InfoSetBit(pTable, v*Width2+b)
						}
					}
				}
				Vec_PtrPush(p.vTables, pTable)
				Vec_IntFree(vValues)
				continue
			}
			if p.pNtk != nil {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Network is already defined.")
			}
			p.pNtk = Wlc_NtkAlloc(pName, Vec_IntSize(p.vStarts))
			p.pNtk.pManName = Abc_NamStart(Vec_IntSize(p.vStarts), 20)
			p.pNtk.pMemTable = p.pMemTable
			p.pMemTable = nil
			p.pNtk.vTables = p.vTables
			p.vTables = nil
			// read the argument definitions
			for pName = Wlc_PrsStrtok("", "(,)"); pName != ""; pName = Wlc_PrsStrtok("", "(,)") {
				pName = Wlc_PrsSkipSpaces(pName)
				if Wlc_PrsStrCmp(pName, "input") == 0 || Wlc_PrsStrCmp(pName, "output") == 0 || Wlc_PrsStrCmp(pName, "wire") == 0 {
					if Wlc_PrsReadDeclaration(p, pName) == 0 {
						return 0
					}
				}
			}

		} else if Wlc_PrsStrCmp(pStart, "endmodule") != 0 {
			var vTemp *Vec_Int_t = Vec_IntStartNatural(Wlc_NtkObjNumMax(p.pNtk))
			Vec_IntAppend(&p.pNtk.vNameIds, vTemp)
			Vec_IntFree(vTemp)
			if p.pNtk.vInits {
				// move FO/FI to be part of CI/CO
				assert((Vec_IntSize(&p.pNtk.vFfs)&1) == 0, "(Vec_IntSize(&p.pNtk.vFfs) & 1) == 0")
				assert(Vec_IntSize(&p.pNtk.vFfs) == 2*Vec_IntSize(p.pNtk.vInits), "Vec_IntSize(&p.pNtk.vFfs) == 2 * Vec_IntSize(p.pNtk.vInits)")
				// inlining of macro Wlc_NtkForEachFf
				for i = 0; i < Vec_IntSize(&p.pNtk.vFfs); i++ {
					pObj = Wlc_NtkFf(p.pNtk, i)
					if i & 1 {
						Wlc_ObjSetCo(p.pNtk, pObj, 1)
					} else {
						Wlc_ObjSetCi(p.pNtk, pObj)
					}
				}
				Vec_IntClear(&p.pNtk.vFfs)
				// convert init values into binary string
				//Vec_IntPrint( p->pNtk->vInits );
				p.pNtk.pInits = Wlc_PrsConvertInitValues(p.pNtk)
				//printf( "%s\n", p->pNtk->pInits );
			}
			if p.pNtk.vArsts && !p.pNtk.fAsyncRst {
				var i, NameIdArst int32
				for i = 0; i < Vec_IntSize(p.pNtk.vArsts); i++ {
					NameIdArst = Vec_IntEntry(p.pNtk.vArsts, i)
					if Wlc_PrsCheckBitConst0(p.pNtk, NameIdArst) != 0 {
						continue
					}
					p.pNtk.fAsyncRst = 1
					fmt.Printf("Detected async reset \"%s\".\n", Abc_NamStr(p.pNtk.pManName, NameIdArst))
					break
				}
			}
			if p.vPoPairs {
				assert(Vec_StrEntryLast(p.vPoPairs) == 0, "Vec_StrEntryLast(p.vPoPairs) == 0")
				Vec_StrPush(p.vPoPairs, 0)
				pName = Vec_StrArray(p.vPoPairs)
				for *pName {
					// inlining of macro Wlc_NtkForEachPo
					for i = 0; i < Wlc_NtkPoNum(p.pNtk); i++ {
						pObj = Wlc_NtkPo(p.pNtk, i)
						if !(int32(strings.Compare(Wlc_ObjName(p.pNtk, Wlc_ObjId(p.pNtk, pObj)), pName)) != 0) {
							Vec_IntPush(&p.pNtk.vPoPairs, i)
							break
						}
					}
					assert(i < Wlc_NtkPoNum(p.pNtk), "i < Wlc_NtkPoNum(p.pNtk)")
					pName += len(pName) + 1
				}
				assert(Vec_IntSize(&p.pNtk.vPoPairs)%2 == 0, "Vec_IntSize(&p.pNtk.vPoPairs) % 2 == 0")
				fmt.Printf("Finished parsing %d output pairs to be checked for equivalence.\n", Vec_IntSize(&p.pNtk.vPoPairs)/2)
			}
			break

			// these are read as part of the interface
		} else if (Wlc_PrsStrCmp(pStart, "input") != 0) || (Wlc_PrsStrCmp(pStart, "output") != 0) || (Wlc_PrsStrCmp(pStart, "wire") != 0) || (Wlc_PrsStrCmp(pStart, "reg") != 0) {
			if !(Wlc_PrsReadDeclaration(p, pStart) != 0) {
				return 0
			}

		} else if Wlc_PrsStrCmp(pStart, "assign") != 0 {
			var Type, NameId, fFound, XValue int32
			cStart := wlcPrsCursorAt(p, len(p.pBuffer)-len(pStart))
			cStart = cStart.advance(len("assign"))
			// read name
			cStart, pName = cStart.findName()
			if !cStart.valid() {
				return Wlc_PrsWriteErrorMessage(p, "", "Cannot read name after assign.")
			}
			NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
			if !fFound {
				return Wlc_PrsWriteErrorMessage(p, cStart.rest(), "Name %s is not declared.", pName)
			}
			// read definition
			Type = Wlc_PrsFindDefinition(p, cStart.rest(), p.vFanins, &XValue)
			if Type != 0 {
				pObj = Wlc_NtkObj(p.pNtk, NameId)
				Wlc_ObjUpdateType(p.pNtk, pObj, Type)
				Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)
				pObj.fXConst = XValue

			} else {
				return 0
			}

		} else if Wlc_PrsStrCmp(pStart, "table") != 0 {
			// THIS IS A HACK to detect tables
			var NameId int32
			var fFound int32
			cStart := wlcPrsCursorAt(p, len(p.pBuffer)-len(pStart))
			cIndex := cStart.advance(len("table"))
			indexText := cIndex.rest()
			indexEnd := 0
			for indexEnd < len(indexText) && Wlc_PrsIsDigit(indexText[indexEnd:]) != 0 {
				indexEnd++
			}
			var iTable int32 = strconv.Atoi(indexText[:indexEnd])
			// find opening
			cStart = cStart.findSymbol('(')
			if !cStart.valid() {
				return Wlc_PrsWriteErrorMessage(p, "", "Cannot read table.")
			}
			// read input
			cStart, pName = cStart.advance(1).findName()
			if !cStart.valid() {
				return Wlc_PrsWriteErrorMessage(p, "", "Cannot read name after assign.")
			}
			NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
			if !fFound {
				return Wlc_PrsWriteErrorMessage(p, cStart.rest(), "Name %s is not declared.", pName)
			}
			// save inputs
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameId)
			Vec_IntPush(p.vFanins, iTable)
			// find comma
			cStart = cStart.findSymbol(',')
			if !cStart.valid() {
				return Wlc_PrsWriteErrorMessage(p, "", "Cannot read table.")
			}
			// read output
			cStart, pName = cStart.advance(1).findName()
			if !cStart.valid() {
				return Wlc_PrsWriteErrorMessage(p, "", "Cannot read name after assign.")
			}
			NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
			if !fFound {
				return Wlc_PrsWriteErrorMessage(p, cStart.rest(), "Name %s is not declared.", pName)
			}
			pObj = Wlc_NtkObj(p.pNtk, NameId)
			Wlc_ObjUpdateType(p.pNtk, pObj, WLC_OBJ_TABLE)
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)

		} else if Wlc_PrsStrCmp(pStart, "always") != 0 {
			// THIS IS A HACK to detect always statement representing combinational MUX
			var NameId int32
			var NameIdOut int32 = -1
			var nValues int32
			var fDefaultFound bool
			cur := wlcPrsCursorAt(p, len(p.pBuffer)-len(pStart))
			// find control
			var found bool
			cur, found = cur.findWord("case")
			if !found {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read case statement.")
			}
			// read the name
			cur = cur.findSymbol('(')
			if !cur.valid() {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read table.")
			}
			cur = cur.advance(1).findSymbol('(')
			if !cur.valid() {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read table.")
			}
			cur, pName = cur.advance(1).findName()
			if !cur.valid() {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read name after case.")
			}
			fFound := int32(0)
			NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
			found = fFound != 0
			if !found {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Name %s is not declared.", pName)
			}
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameId)
			// read data inputs
			pObj = Wlc_NtkObj(p.pNtk, NameId)
			if pObj == nil {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot find the object in case statement.")
			}
			// remember the number of values
			nValues = (1 << Wlc_ObjRange(pObj))
			for {
				// find opening
				cur = cur.findSymbol(':')
				if !cur.valid() {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot find colon in the case statement.")
				}
				// find output name
				cur, pName = cur.advance(1).findName()
				if !cur.valid() {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read name after case.")
				}
				fFound = 0
				NameIdOut = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				found = fFound != 0
				if !found {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Name %s is not declared.", pName)
				}
				// find equality
				cur = cur.findSymbol('=')
				if !cur.valid() {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot find equality in the case statement.")
				}
				// find input name
				pStart = cur.advance(1).skipSpaces().rest()
				pStart = Wlc_PrsReadName(p, pStart, p.vFanins)
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside case statement.")
				}
				cur = cur.withStringSuffix(pStart)
				// consider default
				if fDefaultFound {
					EntryLast := Vec_IntEntryLast(p.vFanins)
					if nValues != Vec_IntSize(p.vFanins)-2 {
						Vec_IntFillExtra(p.vFanins, nValues+1, EntryLast)
					} else {
						Vec_IntPop(p.vFanins)
					}
					// get next line and check its opening character
					i++
					if i >= Vec_IntSize(p.vStarts) {
						return Wlc_PrsWriteErrorMessage(p, "", "Cannot read case statement.")
					}
					pStart = Wlc_PrsStr(p, Vec_IntEntry(p.vStarts, i))
					cur = wlcPrsCursorAt(p, len(p.pBuffer)-len(pStart)).skipSpaces()

				} else {
					// get next line and check its opening character
					i++
					if i >= Vec_IntSize(p.vStarts) {
						return Wlc_PrsWriteErrorMessage(p, "", "Cannot read case statement.")
					}
					pStart = Wlc_PrsStr(p, Vec_IntEntry(p.vStarts, i))
					cur = wlcPrsCursorAt(p, len(p.pBuffer)-len(pStart)).skipSpaces()
					if cur.byte() != 0 && Wlc_PrsIsDigit(cur.rest()) != 0 {
						continue
					}
					if !cur.hasPrefix("default") {
						fDefaultFound = true
						continue
					}
				}
				// find closing
				cur, found = cur.findWord("endcase")
				if !found {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read case statement.")
				}
				// find closing
				cur, found = cur.findWord("end")
				if !found {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read case statement.")
				}
				cur = cur.skipSpaces()
				break
			}
			// check range of the control
			if nValues < Vec_IntSize(p.vFanins)-1 { // may occur if default is not there

				//return Wlc_PrsWriteErrorMessage( p, pStart, "The number of values in the case statement is wrong.", pName );
				fmt.Printf("Warning:  The number of values in the case statement is wrong.\n")
				Vec_IntShrink(p.vFanins, nValues+1)

			} else if nValues > Vec_IntSize(p.vFanins)-1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "The number of values in the case statement is wrong.", pName)
			}
			if Wlc_ObjRange(pObj) == 1 {
				//                return Wlc_PrsWriteErrorMessage( p, pStart, "Always-statement with 1-bit control is not bit-blasted correctly.", pName );
				fmt.Printf("Warning:  Case-statement with 1-bit control is treated as a 2:1 MUX (correct for unsigned signals only).\n")
			}
			pStart = cur.rest()
			pObj = Wlc_NtkObj(p.pNtk, NameIdOut)
			Wlc_ObjUpdateType(p.pNtk, pObj, WLC_OBJ_MUX)
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)
			goto startword

		} else if Wlc_PrsStrCmp(pStart, "CPL_FF") != 0 {
			var NameId int32 = -1
			var NameIdIn int32 = -1
			var NameIdOut int32 = -1
			var fFound int32
			var nBits int32 = 1
			var fFlopIn int32
			var fFlopOut int32
			pStart += len("CPL_FF")
			if pStart[0] == '#' {
				nBits = strconv.Atoi(pStart + 1)
			}
			// read names
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == nil {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart + 1)
				if !p.pNtk.fAsyncRst && !strncmp(pStart, "arst", 4) && pStart[4] != 'v' {
					var NameIdArst int32
					pStart = Wlc_PrsFindSymbol(pStart, '(')
					if pStart == nil {
						return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
					}
					pStart = Wlc_PrsSkipSpaces(pStart + 1)
					if Wlc_PrsIsDigit(pStart) != 0 {
						var Range, Signed, XValue int32
						var vFanins *Vec_Int_t = Vec_IntAlloc(100)
						pStart = Wlc_PrsReadConstant(p, pStart, vFanins, &Range, &Signed, &XValue)
						if pStart && Vec_IntSize(vFanins) == 1 && Vec_IntEntry(vFanins, 0) == 0 {
							Vec_IntFree(vFanins)
							continue
						}
						fmt.Printf("Detected async reset.\n")
						p.pNtk.fAsyncRst = 1
						Vec_IntFree(vFanins)
						continue
					}
					pStart = Wlc_PrsFindName(pStart, &pName)
					if pStart == nil {
						return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
					}
					NameIdArst = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
					if p.pNtk.vArsts == nil {
						p.pNtk.vArsts = Vec_IntAlloc(100)
					}
					Vec_IntPushUnique(p.pNtk.vArsts, NameIdArst)
					continue
				}
				if pStart[0] != 'd' && (pStart[0] != 'q' || pStart[1] == 'b') && strncmp(pStart, "arstval", 7) {
					continue
				}
				fFlopIn = (pStart[0] == 'd')
				fFlopOut = (pStart[0] == 'q')
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart+1, &pName)
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				if fFlopIn != 0 {
					NameIdIn = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopOut != 0 {
					NameIdOut = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else {
					NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				}
				if !fFound {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameIdIn == -1 || NameIdOut == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of flop input or flop output is missing.")
			}
			// create flop output
			pObj = Wlc_NtkObj(p.pNtk, NameIdOut)
			Wlc_ObjUpdateType(p.pNtk, pObj, WLC_OBJ_FO)
			Vec_IntPush(&p.pNtk.vFfs, NameIdOut)
			//if ( nBits != Wlc_ObjRange(pObj) )
			//    printf( "Warning!  Flop output \"%s\" has bit-width (%d) that differs from the flop declaration (%d)\n", Abc_NamStr(p->pNtk->pManName, NameIdOut), Wlc_ObjRange(pObj), nBits );
			// create flop input
			pObj = Wlc_NtkObj(p.pNtk, NameIdIn)
			Vec_IntPush(&p.pNtk.vFfs, NameIdIn)
			//if ( nBits != Wlc_ObjRange(pObj) )
			//    printf( "Warning!  Flop input \"%s\" has bit-width (%d) that differs from the flop declaration (%d)\n", Abc_NamStr(p->pNtk->pManName, NameIdIn), Wlc_ObjRange(pObj), nBits );
			// save flop init value
			if NameId == -1 {
				fmt.Printf("Initial value of flop \"%s\" is not specified. Zero is assumed.\n", Abc_NamStr(p.pNtk.pManName, NameIdOut))
			} else {
				pObj = Wlc_NtkObj(p.pNtk, NameId)
				if nBits != Wlc_ObjRange(pObj) {
					fmt.Printf("Warning!  Flop init signal \"%s\" bit-width (%d) is different from the flop declaration (%d)\n", Abc_NamStr(p.pNtk.pManName, NameId), Wlc_ObjRange(pObj), nBits)
				}
			}
			if p.pNtk.vInits == nil {
				p.pNtk.vInits = Vec_IntAlloc(100)
			}
			Vec_IntPush(p.pNtk.vInits, nameIdOrNegativeWidth(NameId, nBits))
			// printf( "Created flop %s with range %d and init value %d (nameId = %d)\n",
			//     Abc_NamStr(p->pNtk->pManName, NameIdOut), Wlc_ObjRange(pObj), nBits, NameId );

		} else if Wlc_PrsStrCmp(pStart, "ABC_DFFRSE") != 0 {
			var NameId [10]int32
			var fFound int32
			var fFlopIn int32
			var fFlopClk int32
			var fFlopRst int32
			var fFlopSet int32
			var fFlopEna int32
			var fFlopAsync int32
			var fFlopSre int32
			var fFlopInit int32
			var fFlopOut int32
			pStart += len("ABC_DFFRSE")
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == nil {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart + 1)
				fFlopIn = (pStart[0] == 'd')
				fFlopClk = (pStart[0] == 'c')
				fFlopRst = (pStart[0] == 'r')
				fFlopSet = (pStart[0] == 's' && pStart[1] == 'e')
				fFlopEna = (pStart[0] == 'e')
				fFlopAsync = (pStart[0] == 'a')
				fFlopSre = (pStart[0] == 's' && pStart[1] == 'r')
				fFlopInit = (pStart[0] == 'i')
				fFlopOut = (pStart[0] == 'q')
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart+1, &pName)
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				if fFlopIn != 0 {
					NameId[0] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopClk != 0 {
					NameId[1] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopRst != 0 {
					NameId[2] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopSet != 0 {
					NameId[3] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopEna != 0 {
					NameId[4] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopAsync != 0 {
					NameId[5] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopSre != 0 {
					NameId[6] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopInit != 0 {
					NameId[7] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopOut != 0 {
					NameId[8] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else {
					assert(false, "0")
				}
				if !fFound {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameId[0] == -1 || NameId[7] == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of flop input or flop output is missing.")
			}
			// create output
			pObj = Wlc_NtkObj(p.pNtk, NameId[8])
			Wlc_ObjUpdateType(p.pNtk, pObj, WLC_OBJ_FF)
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameId[0])
			Vec_IntPush(p.vFanins, NameId[1])
			Vec_IntPush(p.vFanins, NameId[2])
			Vec_IntPush(p.vFanins, NameId[3])
			Vec_IntPush(p.vFanins, NameId[4])
			Vec_IntPush(p.vFanins, NameId[5])
			Vec_IntPush(p.vFanins, NameId[6])
			Vec_IntPush(p.vFanins, NameId[7])
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)

		} else if Wlc_PrsStrCmp(pStart, "ABC_DFF") != 0 {
			var NameId int32 = -1
			var NameIdIn int32 = -1
			var NameIdOut int32 = -1
			var fFound int32
			var nBits int32 = 1
			var fFlopIn int32
			var fFlopOut int32
			pStart += len("ABC_DFF")
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == nil {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart + 1)
				fFlopIn = (pStart[0] == 'd')
				fFlopOut = (pStart[0] == 'q')
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart+1, &pName)
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				if fFlopIn != 0 {
					NameIdIn = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopOut != 0 {
					NameIdOut = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else {
					NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				}
				if !fFound {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameIdIn == -1 || NameIdOut == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of flop input or flop output is missing.")
			}
			// create flop output
			pObj = Wlc_NtkObj(p.pNtk, NameIdOut)
			Wlc_ObjUpdateType(p.pNtk, pObj, WLC_OBJ_FO)
			Vec_IntPush(&p.pNtk.vFfs, NameIdOut)
			nBits = Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdOut))
			// create flop input
			pObj = Wlc_NtkObj(p.pNtk, NameIdIn)
			Vec_IntPush(&p.pNtk.vFfs, NameIdIn)
			// compare bit-width
			if Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdIn)) != nBits {
				fmt.Printf("Warning!  Flop input \"%s\" bit-width (%d) is different from that of flop output (%d)\n", Abc_NamStr(p.pNtk.pManName, NameId), Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdIn)), nBits)
			}
			// save flop init value
			if NameId == -1 {
				fmt.Printf("Initial value of flop \"%s\" is not specified. Zero is assumed.\n", Abc_NamStr(p.pNtk.pManName, NameIdOut))
			} else {
				if Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameId)) != nBits {
					fmt.Printf("Warning!  Flop init signal \"%s\" bit-width (%d) is different from that of flop output (%d)\n", Abc_NamStr(p.pNtk.pManName, NameId), Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameId)), nBits)
				}
			}
			if p.pNtk.vInits == nil {
				p.pNtk.vInits = Vec_IntAlloc(100)
			}
			Vec_IntPush(p.pNtk.vInits, nameIdOrNegativeWidth(NameId, Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdOut))))
			// printf( "Created flop %s with range %d and init value %d (nameId = %d)\n",
			//     Abc_NamStr(p->pNtk->pManName, NameIdOut), Wlc_ObjRange(pObj), nBits, NameId );
			p.pNtk.fEasyFfs = 1

		} else if Wlc_PrsStrCmp(pStart, "CPL_MEM_") != 0 {
			var pNameId *int32 = nil
			var NameOutput int32
			var NameMi int32 = -1
			var NameMo int32 = -1
			var NameAddr int32 = -1
			var NameDi int32 = -1
			var NameDo int32 = -1
			var fFound int32
			var fRead int32 = 1
			pStart += len("CPL_MEM_")
			if pStart[0] == 'W' {
				fRead = 0
			}
			// read names
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == nil {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart + 1)
				if !strncmp(pStart, "mem_data_in", 11) {
					pNameId = &NameMi
				} else if !strncmp(pStart, "data_in", 7) {
					pNameId = &NameDi
				} else if !strncmp(pStart, "data_out", 8) {
					if fRead != 0 {
						pNameId = &NameDo
					} else {
						pNameId = &NameMo
					}
				} else if !strncmp(pStart, "addr_in", 7) {
					pNameId = &NameAddr
				} else {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name of the input/output port.")
				}
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart+1, &pName)
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				*pNameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				if !fFound {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if (fRead != 0) && (NameMi == -1 || NameAddr == -1 || NameDo == -1) {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of one of signals of read port is missing.")
			}
			if !(fRead != 0) && (NameMi == -1 || NameAddr == -1 || NameDi == -1 || NameMo == -1) {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of one of signals of write port is missing.")
			}
			// create output
			if fRead != 0 {
				NameOutput = NameDo
			} else {
				NameOutput = NameMo
			}
			pObj = Wlc_NtkObj(p.pNtk, NameOutput)
			Wlc_ObjUpdateType(p.pNtk, pObj, readWriteObjType(fRead))
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameMi)
			Vec_IntPush(p.vFanins, NameAddr)
			if !(fRead != 0) {
				Vec_IntPush(p.vFanins, NameDi)
			}
			//printf( "Memory %s ", fRead ? "Read" : "Write" ); printf( "Fanins: " );  Vec_IntPrint( p->vFanins );
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)

		} else if Wlc_PrsStrCmp(pStart, "ABC_READ") != 0 {
			var pNameId *int32 = nil
			var NameMemIn int32 = -1
			var NameData int32 = -1
			var NameAddr int32 = -1
			var fFound int32
			pStart += len("ABC_READ")
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == nil {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart + 1)
				if !strncmp(pStart, "mem_in", 6) {
					pNameId = &NameMemIn
				} else if !strncmp(pStart, "addr", 4) {
					pNameId = &NameAddr
				} else if !strncmp(pStart, "data", 4) {
					pNameId = &NameData
				} else {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name of the input/output port.")
				}
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart+1, &pName)
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				*pNameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				if !fFound {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameMemIn == -1 || NameAddr == -1 || NameData == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of one of signals of read port is missing.")
			}
			// create output
			pObj = Wlc_NtkObj(p.pNtk, NameData)
			Wlc_ObjUpdateType(p.pNtk, pObj, WLC_OBJ_READ)
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameMemIn)
			Vec_IntPush(p.vFanins, NameAddr)
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)
			p.pNtk.fMemPorts = 1

		} else if Wlc_PrsStrCmp(pStart, "ABC_WRITE") != 0 {
			var pNameId *int32 = nil
			var NameMemIn int32 = -1
			var NameMemOut int32 = -1
			var NameData int32 = -1
			var NameAddr int32 = -1
			var fFound int32
			pStart += len("ABC_WRITE")
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == nil {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart + 1)
				if !strncmp(pStart, "mem_in", 6) {
					pNameId = &NameMemIn
				} else if !strncmp(pStart, "mem_out", 7) {
					pNameId = &NameMemOut
				} else if !strncmp(pStart, "data", 4) {
					pNameId = &NameData
				} else if !strncmp(pStart, "addr", 4) {
					pNameId = &NameAddr
				} else {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name of the input/output port.")
				}
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart+1, &pName)
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				*pNameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				if !fFound {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameMemIn == -1 || NameAddr == -1 || NameData == -1 || NameMemOut == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of one of signals of write port is missing.")
			}
			// create output
			pObj = Wlc_NtkObj(p.pNtk, NameMemOut)
			Wlc_ObjUpdateType(p.pNtk, pObj, WLC_OBJ_WRITE)
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameMemIn)
			Vec_IntPush(p.vFanins, NameAddr)
			Vec_IntPush(p.vFanins, NameData)
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)
			p.pNtk.fMemPorts = 1

		} else if (Wlc_PrsStrCmp(pStart, "CPL_RROT") != 0) || (Wlc_PrsStrCmp(pStart, "CPL_LROT") != 0) {
			// CPL_RROT #(128, 6) I_47479(.o ( E_46713 )  , .d ( E_46718 )  , .s ( E_46712 )  );
			var right_rotation int32 = Wlc_PrsStrCmp(pStart, "CPL_RROT")
			var NameId int32 = -1
			var NameIdOut int32 = -1
			var NameIdInD int32 = -1
			var NameIdInS int32 = -1
			var fFound int32
			var fRotInD int32
			var fRotInS int32
			var fRotOut int32
			pStart += len("CPL_RROT")

			// NOTE: no need to parse the parameter values
			//if ( pStart[0] == '#' )

			// read names
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == nil {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart + 1)
				if pStart[0] != 'o' && pStart[0] != 'd' && pStart[0] != 's' {
					continue
				}
				fRotInD = (pStart[0] == 'd')
				fRotInS = (pStart[0] == 's')
				fRotOut = (pStart[0] == 'o')
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the rotation description.")
				}
				pStart = Wlc_PrsFindName(pStart+1, &pName)
				if pStart == nil {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside rotation description.")
				}
				if fRotInD != 0 {
					NameIdInD = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fRotInS != 0 {
					NameIdInS = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fRotOut != 0 {
					NameIdOut = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else {
					NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				}
				if !fFound {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameIdOut == -1 || NameIdInD == -1 || NameIdInS == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Some fields of CPL_ROT are missing.")
			}
			// create rot output
			pObj = Wlc_NtkObj(p.pNtk, NameIdOut)
			Wlc_ObjUpdateType(p.pNtk, pObj, rotationObjType(right_rotation))
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameIdInD)
			Vec_IntPush(p.vFanins, NameIdInS)
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)

		} else if pStart[0] == '(' && pStart[1] == '*' { // skip comments

			for len(pStart) > 0 && !strings.HasPrefix(pStart, ")") {
				pStart = pStart[1:]
			}
			pStart = Wlc_PrsSkipSpaces(pStart)
			goto startword

		} else if pStart[0] != '`' {
			var iLine int32 = Wlc_PrsFindLine(p, pStart)
			pStart = Wlc_PrsFindName(pStart, &pName)
			return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read line %d beginning with %s.", iLine, verFallbackName(pName))
		}
	}
	if p.nNonZero[0] {
		fmt.Printf("Warning: Input file contains %d objects with non-zero-based ranges.\n", p.nNonZero[0])
		fmt.Printf("For example, signal with range [%d:%d] is declared in line %d.\n", p.nNonZero[1], p.nNonZero[2], p.nNonZero[3])
	}
	if p.nNegative[0] {
		fmt.Printf("Warning: Input file contains %d objects with negative ranges.\n", p.nNegative[0])
		fmt.Printf("For example, signal with range [%d:%d] is declared in line %d.\n", p.nNegative[1], p.nNegative[2], p.nNegative[3])
	}
	if p.nReverse[0] {
		fmt.Printf("Warning: Input file contains %d objects with reversed ranges.\n", p.nReverse[0])
		fmt.Printf("For example, signal with range [%d:%d] is declared in line %d.\n", p.nReverse[1], p.nReverse[2], p.nReverse[3])
	}
	return 1
}
func Wlc_ReadVer(pFileName string, pStr string) *Wlc_Ntk_t {
	var p *Wlc_Prs_t
	var pNtk *Wlc_Ntk_t = nil
	// either file name or string is given
	assert((pFileName == nil) != (pStr == nil), "(pFileName == nil) != (pStr == nil)")
	// start the parser
	p = Wlc_PrsStart(pFileName, pStr)
	if p == nil {
		return nil
	}
	// detect lines
	if !(Wlc_PrsPrepare(p) != 0) {
		goto finish
	}
	// parse models
	if !(Wlc_PrsDerive(p) != 0) {
		goto finish
	}
	// derive topological order
	if p.pNtk {
		var pObj *Wlc_Obj_t
		var i int32
		// inlining of macro Wlc_NtkForEachObj
		for i = 1; i < Wlc_NtkObjNumMax(p.pNtk); i++ {
			pObj = Wlc_NtkObj(p.pNtk, i)
			if pObj.Type == WLC_OBJ_FF {
				Vec_IntPush(&p.pNtk.vFfs2, Wlc_ObjId(p.pNtk, pObj))
			}
		}
		pNtk = Wlc_NtkDupDfs(p.pNtk, 0, 1)
		pNtk.pSpec = Abc_UtilStrsav(pFileName)
	}
finish:
	Wlc_PrsPrintErrorMessage(p)
	Wlc_PrsStop(p)
	return pNtk
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Io_ReadWordTest(pFileName string) {
	var pNew *Gia_Man_t
	var pNtk *Wlc_Ntk_t = Wlc_ReadVer(pFileName, nil)
	if pNtk == nil {
		return
	}
	Wlc_WriteVer(pNtk, "test.v", 0, 0)

	pNew = Wlc_NtkBitBlast(pNtk, nil)
	Gia_AigerWrite(pNew, "test.aig", 0, 0, 0)
	Gia_ManStop(pNew)

	Wlc_NtkFree(pNtk)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
