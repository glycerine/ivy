package wlc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
const WLV_PRS_MAX_LINE =   10000

    
type Wlc_Prs_t struct {
    nFileSize                    int32;
    pFileName                  string;
    pBuffer                  string;
    vLines             *Vec_Int_t;
    vStarts             *Vec_Int_t;
    vFanins             *Vec_Int_t;
    pNtk             *Wlc_Ntk_t;
    pMemTable            *Mem_Flex_t;
    vTables             *Vec_Ptr_t;
    vPoPairs             *Vec_Str_t;
    nConsts                    int32;
    nNonZero [4]int32;
    nNegative [4]int32;
    nReverse [4]int32;
    sError [WLV_PRS_MAX_LINE]byte;
}

func Wlc_PrsOffset(p *Wlc_Prs_t, pStr string) int32 { return pStr - p.pBuffer;                     }
func Wlc_PrsStr(p *Wlc_Prs_t, iOffset int32) string { return p.pBuffer + iOffset;                  }
func Wlc_PrsStrCmp(pStr string, pWhat string) int32 { return !strncmp( pStr, pWhat, len(pWhat)); }

// Original macro:
// #define Wlc_PrsForEachLine( p, pLine, i )             \
//     for ( i = 0; (i < Vec_IntSize((p)->vStarts)) && ((pLine) = Wlc_PrsStr(p, Vec_IntEntry((p)->vStarts, i))); i++ )

// Original macro:
// #define Wlc_PrsForEachLineStart( p, pLine, i, Start ) \
//     for ( i = Start; (i < Vec_IntSize((p)->vStarts)) && ((pLine) = Wlc_PrsStr(p, Vec_IntEntry((p)->vStarts, i))); i++ )



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_PrsStart(pFileName string, pStr string) *Wlc_Prs_t {
    var p *Wlc_Prs_t;
    if pFileName && !Extra_FileCheck( pFileName ) {
        return nil; }
    p = new(Wlc_Prs_t);
    p.pFileName = pFileName;
    if pStr {
        p.pBuffer = Abc_UtilStrsav(pStr);
    } else {
        p.pBuffer = Extra_FileReadContents( pFileName );
    }
    p.nFileSize = len(p.pBuffer);  assert(p.nFileSize > 0, "p.nFileSize > 0");
    p.vLines    = Vec_IntAlloc( p.nFileSize / 50 );
    p.vStarts   = Vec_IntAlloc( p.nFileSize / 50 );
    p.vFanins   = Vec_IntAlloc( 100 );
    p.vTables   = Vec_PtrAlloc( 1000 );
    p.pMemTable = Mem_FlexStart();
    return p;
}
func Wlc_PrsStop(p *Wlc_Prs_t) {
    if p.pNtk {
        Wlc_NtkFree( p.pNtk ); }
    if p.pMemTable {
        Mem_FlexStop( p.pMemTable, 0 ); }
    Vec_StrFreeP( &p.vPoPairs );
    Vec_PtrFreeP( &p.vTables );
    Vec_IntFree( p.vLines );
    Vec_IntFree( p.vStarts );
    Vec_IntFree( p.vFanins );
    ABC_FREE( p.pBuffer );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Prints the error message including the file name and line number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_PrsFindLine(p *Wlc_Prs_t, pCur string) int32 {
    var Entry, iLine int32;
    Vec_IntForEachEntry( p.vLines, Entry, iLine )
        if ( Entry > pCur - p.pBuffer )
            return iLine + 1;
    return -1;
}
int Wlc_PrsWriteErrorMessage( Wlc_Prs_t * p, char * pCur, char * format, ... )
{
    var pMessage string;
    // derive message
    va_list args;
    va_start( args, format );
    pMessage = vnsprintf( format, args );
    va_end( args );
    // print messsage
    assert(len(pMessage) < WLV_PRS_MAX_LINE - 100, "strlen(pMessage) < WLV_PRS_MAX_LINE - 100");
    assert(p.sError[0] == 0, "p.sError[0] == 0");
    if pCur == nil { // the line number is not given
        sprintf( p.sError, "%s: %s\n", p.pFileName, pMessage );
    } else {                // print the error message with the line number
    
        var iLine int32  = Wlc_PrsFindLine( p, pCur );
        sprintf( p.sError, "%s (line %d): %s\n", p.pFileName, iLine, pMessage );
    }
    ABC_FREE( pMessage );
    return 0;
}
func Wlc_PrsPrintErrorMessage(p *Wlc_Prs_t) {
    if p.sError[0] == 0 {
        return; }
    fmt.Fprintf(os.Stdout, "%s", p.sError);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_PrsIsDigit(pStr string) int32 {
    return (pStr[0] >= '0' && pStr[0] <= '9');
}
func Wlc_PrsIsChar(pStr string) int32 {
    return (pStr[0] >= 'a' && pStr[0] <= 'z') || 
           (pStr[0] >= 'A' && pStr[0] <= 'Z') || 
           (pStr[0] >= '0' && pStr[0] <= '9') || 
            pStr[0] == '_' || pStr[0] == '$' || pStr[0] == '\\';
}
func Wlc_PrsSkipSpaces(pStr string) string {
    for  *pStr && *pStr == ' '  {
        pStr++;
    }
    return pStr;
}
func Wlc_PrsFindSymbol(pStr string, Symb byte) string {
    var fNotName int32  = 1;
    for  ; *pStr; pStr++  {
        if (fNotName != 0) && *pStr == Symb {
            return pStr; }
        if pStr[0] == '\\' {
            fNotName = 0;
        } else if !(fNotName != 0) && *pStr == ' ' {
            fNotName = 1; }
    }
    return nil;
}
func Wlc_PrsFindSymbolTwo(pStr string, Symb byte, Symb2 byte) string {
    for  ; pStr[1]; pStr++  {
        if pStr[0] == Symb  && pStr[1] == Symb2 {
            return pStr; }
    }
    return nil;
}
func Wlc_PrsFindClosingParenthesis(pStr string, Open byte, Close byte) string {
    var Counter int32;
    var fNotName int32  = 1;
    assert(*pStr == Open, "*pStr == Open");
    for  ; *pStr; pStr++  {
        if fNotName != 0 {
            if *pStr == Open {
                Counter++; }
            if *pStr == Close {
                Counter--; }
            if Counter == 0 {
                return pStr; }
        }
        if *pStr == '\\' {
            fNotName = 0;
        } else if !(fNotName != 0) && *pStr == ' ' {
            fNotName = 1; }
    }
    return nil;
}
func Wlc_PrsRemoveComments(p *Wlc_Prs_t) int32 {
    var fSpecifyFound int32;
    var pCur string;
    var pNext string;
    var pEnd string = p.pBuffer + p.nFileSize;
    for  pCur = p.pBuffer; pCur < pEnd; pCur++  {
        // regular comment (//)
        if *pCur == '/' && pCur[1] == '/' {
            if pCur + 5 < pEnd && pCur[2] == 'a' && pCur[3] == 'b' && pCur[4] == 'c' && pCur[5] == '2' {
                pCur[0] = pCur[1] = pCur[2] = pCur[3] = pCur[4] = pCur[5] = ' ';
            } else if !strncmp(pCur + 3, "Pair:", 5) {
                if p.vPoPairs == nil {
                    p.vPoPairs = Vec_StrAlloc( 100 ); }
                for  pNext = pCur + 9; *pNext != '\n'; pNext++  {
                    if *pNext == ' ' {
                        Vec_StrPush( p.vPoPairs, '\0' );
                    } else if *pNext != '\r' {
                        Vec_StrPush( p.vPoPairs, *pNext ); }
                }
                if Vec_StrEntryLast(p.vPoPairs) != 0 {
                    Vec_StrPush(p.vPoPairs, 0); }
            
            } else {
                pNext = Wlc_PrsFindSymbol( pCur, '\n' );
                if pNext == nil {
                    return Wlc_PrsWriteErrorMessage( p, pCur, "Cannot find end-of-line after symbols \"//\"." ); }
                for  ; pCur < pNext; pCur++  {
                    *pCur = ' ';
                }
            }
        
        // skip preprocessor directive (`timescale, `celldefine, etc)
        } else if *pCur == '`' {
            pNext = Wlc_PrsFindSymbol( pCur, '\n' );
            if pNext == nil {
                return Wlc_PrsWriteErrorMessage( p, pCur, "Cannot find end-of-line after symbols \"`\"." ); }
            for  ; pCur < pNext; pCur++  {
                *pCur = ' ';
            }
        
        // regular comment (/* ... */)
        } else if *pCur == '/' && pCur[1] == '*' {
            pNext = Wlc_PrsFindSymbolTwo( pCur, '*', '/' );
            if pNext == nil {
                return Wlc_PrsWriteErrorMessage( p, pCur, "Cannot find symbols \"*/\" after symbols \"/*\"." ); }
            // overwrite comment
            for  ; pCur < pNext + 2; pCur++  {
                *pCur = ' ';
            }
        
        // 'specify' treated as comments
        } else if *pCur == 's' && pCur[1] == 'p' && pCur[2] == 'e' && !strncmp(pCur, "specify", 7) {
            for  pNext = pCur; pNext < pEnd - 10; pNext++  {
                if *pNext == 'e' && pNext[1] == 'n' && pNext[2] == 'd' && !strncmp(pNext, "endspecify", 10) {
                    // overwrite comment
                    for  ; pCur < pNext + 10; pCur++  {
                        *pCur = ' ';
                    }
                    if fSpecifyFound == 0 {
                        Abc_Print( 0, "Ignoring specify/endspecify directives.\n" ); }
                    fSpecifyFound = 1;
                    break;
                }
            }
        
        // insert semicolons
        } else if *pCur == 'e' && pCur[1] == 'n' && pCur[2] == 'd' && !strncmp(pCur, "endmodule", 9) {
            pCur[len("endmodule")] = ';'; }
        // overwrite end-of-lines with spaces (less checking to do later on)
        if *pCur == '\n' || *pCur == '\r'  || *pCur == '\t' {
            *pCur = ' '; }
    }
    return 1;
}
func Wlc_PrsPrepare(p *Wlc_Prs_t) int32 {
    var fPrettyPrint int32;
    var fNotName int32  = 1;
    var pTemp, pPrev, pThis string;
    // collect info about lines
    assert(Vec_IntSize(p.vLines) == 0, "Vec_IntSize(p.vLines) == 0");
    for  pTemp = p.pBuffer; *pTemp; pTemp++  {
        if *pTemp == '\n' {
            Vec_IntPush( p.vLines, pTemp - p.pBuffer ); }
    }
    // delete comments and insert breaks
    if !(Wlc_PrsRemoveComments( p ) != 0) {
        return 0; }
    // collect info about breaks
    assert(Vec_IntSize(p.vStarts) == 0, "Vec_IntSize(p.vStarts) == 0");
    for  pPrev = pThis = p.pBuffer; *pThis; pThis++  {
        if (fNotName != 0) && *pThis == ';' {
            *pThis = 0;
            Vec_IntPush( p.vStarts, Wlc_PrsOffset(p, Wlc_PrsSkipSpaces(pPrev)) );
            pPrev = pThis + 1;
        }
        if *pThis == '\\' {
            fNotName = 0;
        } else if !(fNotName != 0) && *pThis == ' ' {
            fNotName = 1; }
    }

    if fPrettyPrint != 0 {
        var i, k int32;
        // print the line types
        // inlining of macro Wlc_PrsForEachLine
        for  i = 0; (i < Vec_IntSize((p).vStarts)) && ((pTemp) = Wlc_PrsStr(p, Vec_IntEntry((p).vStarts, i))); i++  {
            if Wlc_PrsStrCmp( pTemp, "module" ) != 0 {
                fmt.Printf("\n"); }
            if !(Wlc_PrsStrCmp( pTemp, "module" ) != 0) && !(Wlc_PrsStrCmp( pTemp, "endmodule" ) != 0) {
                fmt.Printf("    "); }
            fmt.Printf("%c", pTemp[0]);
            for  k = 1; pTemp[k]; k++  {
                if pTemp[k] != ' ' || pTemp[k-1] != ' ' {
                    fmt.Printf("%c", pTemp[k]); }
            }
            fmt.Printf(";\n");
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
    return 1;
}

/**Function*************************************************************

  Synopsis    [Modified version of strtok().]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Wlc_PrsStrtok_last string;
func Wlc_PrsStrtok(s string, delim string) string {
  char *spanp;
  var c, sc int32;
  var tok string;
  
  if s == nil && (s = spStatic_Wlc_PrsStrtok_last) == nil {
      return nil; }
  // skip leading delimiters
cont:
  c = *s++;
  for spanp = delim; (sc = *spanp++) != 0; { 
      if c == sc {
          goto cont; }
  }
  if c == 0 {    // no non-delimiter characters 
      return (spStatic_Wlc_PrsStrtok_last = nil); }
//  tok = s - 1;
  if c != '\\' {
      tok = s - 1;
  } else {
      tok = s - 1; }
  // go back to the first non-delimiter character
  s--;
  // find the token
  for {
      c = *s++;
      if c == '\\' {  // skip blind characters
      
          for  c != ' '  {
              c = *s++;
          }
          c = *s++;
      }
      spanp = delim;
      for {
          if (sc = *spanp++) == c {
              if c == 0 {
                  s = nil;
              } else {
                  s[-1] = 0; }
              spStatic_Wlc_PrsStrtok_last = s;
              return (tok);
          }
      
if !(sc != 0) { break }
} 
  }
  // not reached
  return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_PrsConvertInitValues(p *Wlc_Ntk_t) string {
    var pObj *Wlc_Obj_t; 
    var i int32;
    var k int32;
    var Value int32;
    var pInits *int32;
    var pResult string;
    var vStr *Vec_Str_t  = Vec_StrAlloc( 1000 );
    Vec_IntForEachEntry( p.vInits, Value, i )
    {
        if Value < 0 {
            for  k = 0; k < -Value; k++  {
                Vec_StrPush( vStr, '0' );
            }
            continue;
        }
        pObj = Wlc_NtkObj( p, Value );
        Value = Wlc_ObjRange(pObj);
        for  pObj.Type == WLC_OBJ_BUF  {
            pObj = Wlc_NtkObj( p, Wlc_ObjFaninId0(pObj) );
        }
        if pObj.Type == WLC_OBJ_CONST && !pObj.fXConst {
            pInits = Wlc_ObjConstValue(pObj);
        } else {
            pInits = nil;
        }
        for  k = 0; k < Abc_MinInt(Value, Wlc_ObjRange(pObj)); k++  {
            Vec_StrPush( vStr, byte((pInits ? '0' + Abc_InfoHasBit((unsigned *)pInits, k) : 'x')) );
        }
        // extend values with zero, in case the init value signal has different range compared to constant used
        for  ; k < Value; k++  {
            Vec_StrPush( vStr, '0' );
        }
        // update vInits to contain either number of values or PI index
        Vec_IntWriteEntry( p.vInits, i, (pInits || pObj.fXConst) ? -Value : Wlc_ObjCiId(pObj) );
    }
    Vec_StrPush( vStr, '\0' );
    pResult = Vec_StrReleaseArray( vStr );
    Vec_StrFree( vStr );
    return pResult;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_PrsCheckBitConst0(p *Wlc_Ntk_t, NameId int32) int32 {
    var pObj *Wlc_Obj_t  = Wlc_NtkObj( p, NameId );
    var pInits *int32;
    if Wlc_ObjRange(pObj) != 1 {
        return 0; }
    for  pObj.Type == WLC_OBJ_BUF  {
        pObj = Wlc_NtkObj( p, Wlc_ObjFaninId0(pObj) );
    }
    if pObj.Type != WLC_OBJ_CONST {
        return 0; }
    pInits = Wlc_ObjConstValue(pObj);
    return Abc_InfoHasBit((unsigned *)pInits, 0) == 0;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_PrsFindRange(pStr string, End *int32, Beg *int32) string {
    *End = *Beg = 0;
    pStr = Wlc_PrsSkipSpaces( pStr );
    if pStr[0] != '[' {
        return pStr; }
    pStr = Wlc_PrsSkipSpaces( pStr+1 );
    if !(Wlc_PrsIsDigit(pStr) != 0) && pStr[0] != '-' {
        return nil; }
    *End = *Beg = strconv.Atoi(pStr);
    if Wlc_PrsFindSymbol( pStr, ':' ) == nil {
        pStr = Wlc_PrsFindSymbol( pStr, ']' );
        if pStr == nil {
            return nil; }
    
    } else {
        pStr = Wlc_PrsFindSymbol( pStr, ':' );
        pStr = Wlc_PrsSkipSpaces( pStr+1 );
        if !(Wlc_PrsIsDigit(pStr) != 0) && pStr[0] != '-' {
            return nil; }
        *Beg = strconv.Atoi(pStr);
        pStr = Wlc_PrsFindSymbol( pStr, ']' );
        if pStr == nil {
            return nil; }
    }
    return pStr + 1;
}
func Wlc_PrsFindWord(pStr string, pWord string, fFound *int32) string {
    *fFound = 0;
    pStr = Wlc_PrsSkipSpaces( pStr );
    if !(Wlc_PrsStrCmp(pStr, pWord) != 0) {
        return pStr; }
    *fFound = 1;
    return pStr + len(pWord);
}
var spStatic_Wlc_PrsFindName_Buffer [WLV_PRS_MAX_LINE]byte;
func Wlc_PrsFindName(pStr string, ppPlace *string) string {
    
    var pThis string  = *ppPlace = spStatic_Wlc_PrsFindName_Buffer;
    var fNotName int32 = 1;
    var Count int32;
    pStr = Wlc_PrsSkipSpaces( pStr );
    if !(Wlc_PrsIsChar(pStr) != 0) {
        return nil; }
//    while ( Wlc_PrsIsChar(pStr) )
//        *pThis++ = *pStr++;
    for  *pStr  {
        if (fNotName != 0) && !(Wlc_PrsIsChar(pStr) != 0) {
            break; }
        if *pStr == '\\' {
            Count++;
            fNotName = 0;
        
        } else if !(fNotName != 0) && *pStr == ' ' {
            Count--;
            if !(Count != 0) {
                fNotName = 1; }
        }
        *pThis++ = *pStr++;
    }
    *pThis = 0;
    return pStr;
}
func Wlc_PrsReadConstant(p *Wlc_Prs_t, pStr string, vFanins *Vec_Int_t, pRange *int32, pSigned *int32, pXValue *int32) string {
    var i int32;
    var nDigits int32;
    var nBits int32 = strconv.Atoi(pStr);
    *pRange = -1;
    *pSigned = 0;
    *pXValue = 0;
    pStr = Wlc_PrsSkipSpaces( pStr );
    if Wlc_PrsFindSymbol( pStr, '\'' ) == nil {
        // handle decimal number
        var Number int32  = strconv.Atoi(pStr);
        *pRange = Abc_Base2Log( Number+1 );
        assert(*pRange < 32, "*pRange < 32");
        for  Wlc_PrsIsDigit(pStr) != 0  {
            pStr++;
        }
        Vec_IntFill( vFanins, 1, Number );
        return pStr;
    }
    pStr = Wlc_PrsFindSymbol( pStr, '\'' );
    if pStr[1] == 's' {
        *pSigned = 1;
        pStr++;
    }
    if pStr[1] == 'b' {
        Vec_IntFill( vFanins, Abc_BitWordNum(nBits), 0 );
        for  i = 0; i < nBits; i++  {
            if pStr[2+i] == '1' {
                Abc_InfoSetBit( (unsigned *)Vec_IntArray(vFanins), nBits-1-i );
            } else if pStr[2+i] != '0' {
                return (char *)(ABC_PTRINT_T)Wlc_PrsWriteErrorMessage( p, pStr, "Wrong digit in binary constant \"%c\".", pStr[2+i] ); }
        }
        *pRange = nBits;
        pStr += 2 + nBits;
        return pStr;
    }
    if pStr[1] != 'h' {
        return (char *)(ABC_PTRINT_T)Wlc_PrsWriteErrorMessage( p, pStr, "Expecting hexadecimal constant and not \"%c\".", pStr[1] ); }
    *pXValue = (pStr[2] == 'x' || pStr[2] == 'X');
    if *pXValue == 'X' {
        *pXValue = 'x'; }
    Vec_IntFill( vFanins, Abc_BitWordNum(nBits), 0 );
    nDigits = Abc_TtReadHexNumber( (word *)Vec_IntArray(vFanins), pStr+2 );
    if nDigits != (nBits + 3)/4 {
//        return (char *)(ABC_PTRINT_T)Wlc_PrsWriteErrorMessage( p, pStr, "The length of a constant does not match." );
//        printf( "Warning: The length of a constant (%d hex digits) does not match the number of bits (%d).\n", nDigits, nBits );
    }
    *pRange = nBits;
    pStr += 2;
    for  Wlc_PrsIsChar(pStr) != 0  {
        pStr++;
    }
    return pStr;
}
func Wlc_PrsReadName(p *Wlc_Prs_t, pStr string, vFanins *Vec_Int_t) string {
    var NameId, fFound, iObj int32;
    pStr = Wlc_PrsSkipSpaces( pStr );
    if Wlc_PrsIsDigit(pStr) != 0 {
        var Buffer string // was [100]byte in C

        var Range, Signed, XValue int32;
        var vFanins *Vec_Int_t  = Vec_IntAlloc(0);
        pStr = Wlc_PrsReadConstant( p, pStr, vFanins, &Range, &Signed, &XValue );
        if pStr == nil {
            Vec_IntFree( vFanins );
            return 0;
        }
        // create new node
        iObj = Wlc_ObjAlloc( p.pNtk, WLC_OBJ_CONST, Signed, Range-1, 0 );
        Wlc_ObjAddFanins( p.pNtk, Wlc_NtkObj(p.pNtk, iObj), vFanins );
        Wlc_NtkObj(p.pNtk, iObj).fXConst = XValue;
        Vec_IntFree( vFanins );
        // add node's name
        Buffer = fmt.Sprintf("_c%d_", p.nConsts++);
        NameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, Buffer, &fFound );
        if fFound {
            return (char *)(ABC_PTRINT_T)Wlc_PrsWriteErrorMessage( p, pStr, "Name %s is already used.", Buffer ); }
        assert(iObj == NameId, "iObj == NameId");
    
    } else {
        var pName string;
        pStr = Wlc_PrsFindName( pStr, &pName );
        if pStr == nil {
            return (char *)(ABC_PTRINT_T)Wlc_PrsWriteErrorMessage( p, pStr, "Cannot read name in assign-statement." ); }
        NameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
        if !fFound {
            return (char *)(ABC_PTRINT_T)Wlc_PrsWriteErrorMessage( p, pStr, "Name %s is used but not declared.", pName ); }
    }
    Vec_IntPush( vFanins, NameId );
    return Wlc_PrsSkipSpaces( pStr );
}
func Wlc_PrsFindDefinition(p *Wlc_Prs_t, pStr string, vFanins *Vec_Int_t, pXValue *int32) int32 {
    var pName string;
    var Type int32  = WLC_OBJ_NONE;
    var fRotating int32;
    Vec_IntClear( vFanins );
    pStr = Wlc_PrsSkipSpaces( pStr );
    if pStr[0] != '=' {
        return 0; }
    pStr = Wlc_PrsSkipSpaces( pStr+1 );
    if pStr[0] == '(' {
        // consider rotating shifter
        if Wlc_PrsFindSymbolTwo(pStr, '>', '>') && Wlc_PrsFindSymbolTwo(pStr, '<', '<') {
            // THIS IS A HACK TO DETECT rotating shifters
            var pClose string  = Wlc_PrsFindClosingParenthesis( pStr, '(', ')' );
            if pClose == nil {
                return Wlc_PrsWriteErrorMessage( p, pStr, "Expecting closing parenthesis." ); }
            *pStr = ' '; *pClose = 0;
            pStr = Wlc_PrsSkipSpaces( pStr );
            fRotating = 1;
        
        } else {
            var pClose string  = Wlc_PrsFindClosingParenthesis( pStr, '(', ')' );
            if pClose == nil {
                return Wlc_PrsWriteErrorMessage( p, pStr, "Expecting closing parenthesis." ); }
            *pStr = *pClose = ' ';
            pStr = Wlc_PrsSkipSpaces( pStr );
        }
    }
    if Wlc_PrsIsDigit(pStr) != 0 {
        var Range, Signed int32;
        pStr = Wlc_PrsReadConstant( p, pStr, vFanins, &Range, &Signed, pXValue );
        if pStr == nil {
            return 0; }
        Type = WLC_OBJ_CONST;
    
    } else if pStr[0] == '!' || (pStr[0] == '~' && pStr[1] != '&' && pStr[1] != '|' && pStr[1] != '^') || pStr[0] == '@' || pStr[0] == '#' {
        if pStr[0] == '!' {
            Type = WLC_OBJ_LOGIC_NOT;
        } else if pStr[0] == '~' {
            Type = WLC_OBJ_BIT_NOT;
        } else if pStr[0] == '@' {
            Type = WLC_OBJ_ARI_SQRT;
        } else if pStr[0] == '#' {
            Type = WLC_OBJ_ARI_SQUARE;
        } else { assert(false, "0"); }
        // skip parentheses
        pStr = Wlc_PrsSkipSpaces( pStr+1 );
        if pStr[0] == '(' {
            var pClose string  = Wlc_PrsFindClosingParenthesis( pStr, '(', ')' );
            if pClose == nil {
                return Wlc_PrsWriteErrorMessage( p, pStr, "Expecting closing parenthesis." ); }
            *pStr = *pClose = ' ';
        }
        if !(pStr = Wlc_PrsReadName(p, pStr, vFanins)) {
            return Wlc_PrsWriteErrorMessage( p, pStr, "Cannot read name after !." ); }
    
    } else if pStr[0] == '-' || 
              pStr[0] == '&' || pStr[0] == '|' || pStr[0] == '^' || 
              (pStr[0] == '~' && pStr[1] == '&') || 
              (pStr[0] == '~' && pStr[1] == '|') || 
              (pStr[0] == '~' && pStr[1] == '^') {
        var shift int32  = 1;
        if pStr[0] == '-' {
            Type = WLC_OBJ_ARI_MINUS;
        } else if pStr[0] == '&' {
            Type = WLC_OBJ_REDUCT_AND;
        } else if pStr[0] == '|' {
            Type = WLC_OBJ_REDUCT_OR;
        } else if pStr[0] == '^' {
            Type = WLC_OBJ_REDUCT_XOR;
        } else if pStr[0] == '~' && pStr[1] == '&' {Type = WLC_OBJ_REDUCT_NAND; shift = 2;
        } else if pStr[0] == '~' && pStr[1] == '|' {Type = WLC_OBJ_REDUCT_NOR; shift = 2;
        } else if pStr[0] == '~' && pStr[1] == '^' {Type = WLC_OBJ_REDUCT_NXOR; shift = 2;
        } else { assert(false, "0"); }
        if !(pStr = Wlc_PrsReadName(p, pStr+shift, vFanins)) {
            return Wlc_PrsWriteErrorMessage( p, pStr, "Cannot read name after a unary operator." ); }
    
    } else if pStr[0] == '{' {
        // THIS IS A HACK TO DETECT zero padding AND sign extension
        if Wlc_PrsFindSymbol(pStr+1, '{') {
            if Wlc_PrsFindSymbol(pStr+1, '\'') {
                Type = WLC_OBJ_BIT_ZEROPAD;
            } else {
                Type = WLC_OBJ_BIT_SIGNEXT; }
            pStr = Wlc_PrsFindSymbol(pStr+1, ',');
            if pStr == nil {
                return Wlc_PrsWriteErrorMessage( p, pStr, "Expecting one comma in this line." ); }
            if !(pStr = Wlc_PrsReadName(p, pStr+1, vFanins)) {
                return Wlc_PrsWriteErrorMessage( p, pStr, "Cannot read name in sign-extension." ); }
            pStr = Wlc_PrsSkipSpaces( pStr );
            if pStr[0] != '}' {
                return Wlc_PrsWriteErrorMessage( p, pStr, "There is no closing brace (})." ); }
        
        } else { // concatenation
        
            for {
                pStr = Wlc_PrsSkipSpaces( pStr+1 );
                if !(pStr = Wlc_PrsReadName(p, pStr, vFanins)) {
                    return Wlc_PrsWriteErrorMessage( p, pStr, "Cannot read name in concatenation." ); }
                if pStr[0] == '}' {
                    break; }
                if pStr[0] != ',' {
                    return Wlc_PrsWriteErrorMessage( p, pStr, "Expected comma (,) in this place." ); }
            }
            Type = WLC_OBJ_BIT_CONCAT;
        }
        assert(pStr[0] == '}', "pStr[0] == '}'");
        pStr++;
    
    } else {
        if !(pStr = Wlc_PrsReadName(p, pStr, vFanins)) {
            return 0; }
        // get the next symbol
        if pStr[0] == 0 {
            Type = WLC_OBJ_BUF;
        } else if pStr[0] == '?' {
            if !(pStr = Wlc_PrsReadName(p, pStr+1, vFanins)) {
                return Wlc_PrsWriteErrorMessage( p, pStr, "Cannot read name in MUX." ); }
            if pStr[0] != ':' {
                return Wlc_PrsWriteErrorMessage( p, pStr, "MUX lacks the colon symbol (:)." ); }
            if !(pStr = Wlc_PrsReadName(p, pStr+1, vFanins)) {
                return Wlc_PrsWriteErrorMessage( p, pStr, "Cannot read name in MUX." ); }
            assert(Vec_IntSize(vFanins) == 3, "Vec_IntSize(vFanins) == 3");
            ABC_SWAP( int, Vec_IntArray(vFanins)[1], Vec_IntArray(vFanins)[2] );
            Type = WLC_OBJ_MUX;
        
        } else if pStr[0] == '[' {
            var End, Beg int32; var pLine string  = pStr;
            pStr = Wlc_PrsFindRange( pStr, &End, &Beg );
            if pStr == nil {
                return Wlc_PrsWriteErrorMessage( p, pLine, "Non-standard range." ); }
            Vec_IntPushTwo( vFanins, End, Beg );
            Type = WLC_OBJ_BIT_SELECT;
        
        } else {
                 if pStr[0] == '>' && pStr[1] == '>' && pStr[2] != '>' { pStr += 2, Type = fRotating ? WLC_OBJ_ROTATE_R : WLC_OBJ_SHIFT_R;
            } else if pStr[0] == '>' && pStr[1] == '>' && pStr[2] == '>' { pStr += 3, Type = WLC_OBJ_SHIFT_RA;      
            } else if pStr[0] == '<' && pStr[1] == '<' && pStr[2] != '<' { pStr += 2, Type = fRotating ? WLC_OBJ_ROTATE_L : WLC_OBJ_SHIFT_L;
            } else if pStr[0] == '<' && pStr[1] == '<' && pStr[2] == '<' { pStr += 3, Type = WLC_OBJ_SHIFT_LA;      
            } else if pStr[0] == '&' && pStr[1] != '&' { pStr += 1, Type = WLC_OBJ_BIT_AND;       
            } else if pStr[0] == '|' && pStr[1] != '|' { pStr += 1, Type = WLC_OBJ_BIT_OR;        
            } else if pStr[0] == '^' && pStr[1] != '^' { pStr += 1, Type = WLC_OBJ_BIT_XOR;       
            } else if pStr[0] == '~' && pStr[1] == '&' { pStr += 2, Type = WLC_OBJ_BIT_NAND;       
            } else if pStr[0] == '~' && pStr[1] == '|' { pStr += 2, Type = WLC_OBJ_BIT_NOR;       
            } else if pStr[0] == '~' && pStr[1] == '^' { pStr += 2, Type = WLC_OBJ_BIT_NXOR;       
            } else if pStr[0] == '=' && pStr[1] == '>' { pStr += 2, Type = WLC_OBJ_LOGIC_IMPL;     
            } else if pStr[0] == '&' && pStr[1] == '&' { pStr += 2, Type = WLC_OBJ_LOGIC_AND;     
            } else if pStr[0] == '|' && pStr[1] == '|' { pStr += 2, Type = WLC_OBJ_LOGIC_OR;      
            } else if pStr[0] == '^' && pStr[1] == '^' { pStr += 2, Type = WLC_OBJ_LOGIC_XOR;      
            } else if pStr[0] == '=' && pStr[1] == '=' { pStr += 2, Type = WLC_OBJ_COMP_EQU;      
            } else if pStr[0] == '!' && pStr[1] == '=' { pStr += 2, Type = WLC_OBJ_COMP_NOTEQU;      
            } else if pStr[0] == '<' && pStr[1] != '=' { pStr += 1, Type = WLC_OBJ_COMP_LESS;     
            } else if pStr[0] == '>' && pStr[1] != '=' { pStr += 1, Type = WLC_OBJ_COMP_MORE;     
            } else if pStr[0] == '<' && pStr[1] == '=' { pStr += 2, Type = WLC_OBJ_COMP_LESSEQU;
            } else if pStr[0] == '>' && pStr[1] == '=' { pStr += 2, Type = WLC_OBJ_COMP_MOREEQU;  
            } else if pStr[0] == '+' { pStr += 1, Type = WLC_OBJ_ARI_ADD;       
            } else if pStr[0] == '-' { pStr += 1, Type = WLC_OBJ_ARI_SUB;       
            } else if pStr[0] == '*' && pStr[1] != '*' { pStr += 1, Type = WLC_OBJ_ARI_MULTI;     
            } else if pStr[0] == '/' { pStr += 1, Type = WLC_OBJ_ARI_DIVIDE;        
            } else if pStr[0] == '%' { pStr += 1, Type = WLC_OBJ_ARI_REM;   
            } else if pStr[0] == '*' && pStr[1] == '*' { pStr += 2, Type = WLC_OBJ_ARI_POWER;
            } else { return Wlc_PrsWriteErrorMessage( p, pStr, "Unsupported operation (%c).", pStr[0] ); }
            if !(pStr = Wlc_PrsReadName(p, pStr+1, vFanins)) {
                return 0; }
            pStr = Wlc_PrsSkipSpaces( pStr );
            if Type == WLC_OBJ_ARI_ADD && pStr[0] == '+' {
                if !(pStr = Wlc_PrsReadName(p, pStr+1, vFanins)) {
                    return 0; }
                pStr = Wlc_PrsSkipSpaces( pStr );
            }
            if pStr[0] {
                fmt.Printf("Warning: Trailing symbols \"%s\" in line %d.\n", pStr, Wlc_PrsFindLine(p, pStr)); }
        }
    }
    // make sure there is nothing left there
    if pStr {
        pStr = Wlc_PrsFindName( pStr, &pName );
        if pStr != nil {
            return Wlc_PrsWriteErrorMessage( p, pStr, "Name %s is left at the end of the line.", pName ); }   
    }
    return Type;
}
func Wlc_PrsReadDeclaration(p *Wlc_Prs_t, pStart string) int32 {
    var fFound int32;
    var Type int32 = WLC_OBJ_NONE;
    var iObj int32; var pLine string;
    var Signed, Beg, End, NameId, fIsPo int32;
    if Wlc_PrsStrCmp( pStart, "input" ) != 0 {
        pStart += len("input"), Type = WLC_OBJ_PI;
    } else if Wlc_PrsStrCmp( pStart, "output" ) != 0 {
        pStart += len("output"), fIsPo = 1; }
    pStart = Wlc_PrsSkipSpaces( pStart );
    if Wlc_PrsStrCmp( pStart, "wire" ) != 0 {
        pStart += len("wire");
    } else if Wlc_PrsStrCmp( pStart, "reg" ) != 0 {
        pStart += len("reg"); }
    // read 'signed'
    pStart = Wlc_PrsFindWord( pStart, "signed", &Signed );
    // read range
    pLine = pStart;
    pStart = Wlc_PrsFindRange( pStart, &End, &Beg );
    if pStart == nil {
        return Wlc_PrsWriteErrorMessage( p, pLine, "Non-standard range." ); }
    if End != 0 && Beg != 0 {
        if p.nNonZero[0]++ == 0 {
            p.nNonZero[1] = End;
            p.nNonZero[2] = Beg;
            p.nNonZero[3] = Wlc_PrsFindLine(p, pStart);
        }
    }
    if End < 0 || Beg < 0 {
        if p.nNegative[0]++ == 0 {
            p.nNegative[1] = End;
            p.nNegative[2] = Beg;
            p.nNegative[3] = Wlc_PrsFindLine(p, pStart);
        }
    }
    if End < Beg {
        if p.nReverse[0]++ == 0 {
            p.nReverse[1] = End;
            p.nReverse[2] = Beg;
            p.nReverse[3] = Wlc_PrsFindLine(p, pStart);
        }
    }
    for {
        var pName string; var XValue, TypeNew int32;
        // read name
        pStart = Wlc_PrsFindName( pStart, &pName );
        if pStart == nil {
            return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name in declaration." ); }
        NameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
        if fFound {
            return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is declared more than once.", pName ); }
        iObj = Wlc_ObjAlloc( p.pNtk, Type, Signed, End, Beg );
        if fIsPo != 0 { Wlc_ObjSetCo( p.pNtk, Wlc_NtkObj(p.pNtk, iObj), 0 ); }
        assert(iObj == NameId, "iObj == NameId");
        // check next definition
        pStart = Wlc_PrsSkipSpaces( pStart );
        if pStart[0] == ',' {
            pStart++;
            continue;
        }
        // check definition
        TypeNew = Wlc_PrsFindDefinition( p, pStart, p.vFanins, &XValue );
        if TypeNew != 0 {
            var pObj *Wlc_Obj_t  = Wlc_NtkObj( p.pNtk, iObj );
            Wlc_ObjUpdateType( p.pNtk, pObj, TypeNew );
            Wlc_ObjAddFanins( p.pNtk, pObj, p.vFanins );
            pObj.fXConst = XValue;
        }
        break;
    }
    return 1;
}
func Wlc_PrsDerive(p *Wlc_Prs_t) int32 {
    var pObj *Wlc_Obj_t;
    var pStart, pName string;
    var i int32;
    // go through the directives
    // inlining of macro Wlc_PrsForEachLine
    for  i = 0; (i < Vec_IntSize((p).vStarts)) && ((pStart) = Wlc_PrsStr(p, Vec_IntEntry((p).vStarts, i))); i++  {
startword:
        if Wlc_PrsStrCmp( pStart, "module" ) != 0 {
            // get module name
            pName = Wlc_PrsStrtok( pStart + len("module"), " \r\n\t(,)" );
            if pName == nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read model name." ); }
            // THIS IS A HACK to skip definitions of modules beginning with "CPL_"
            if Wlc_PrsStrCmp( pName, "CPL_" ) != 0 {
                for  ++i < Vec_IntSize(p.vStarts)  {
                    pStart = Wlc_PrsStr(p, Vec_IntEntry(p.vStarts, i));
                    pStart = strstr( pStart, "endmodule" );
                    if pStart != nil {
                        break; }
                }
                continue;
            }
            if Wlc_PrsStrCmp( pName, "table" ) != 0 {
                // THIS IS A HACK to detect table module descriptions
                var Width1 int32 = -1;
                var Width2 int32 = -1;
                var v, b, Value, nBits, nInts int32;
                var pTable *uint32;
                var vValues *Vec_Int_t  = Vec_IntAlloc( 256 );
                // inlining of macro Wlc_PrsForEachLineStart
                for  i = i+1; (i < Vec_IntSize((p).vStarts)) && ((pStart) = Wlc_PrsStr(p, Vec_IntEntry((p).vStarts, i))); i++  {
                    if Wlc_PrsStrCmp( pStart, "endcase" ) != 0 {
                        break; }
                    pStart = Wlc_PrsFindSymbol( pStart, '\'' );
                    if pStart == nil {
                        continue; }
                    Width1 = strconv.Atoi(pStart-1);
                    pStart = Wlc_PrsFindSymbol( pStart+2, '\'' );
                    if pStart == nil {
                        continue; }
                    Width2 = strconv.Atoi(pStart-1);
                    Value = 0;
                    Abc_TtReadHexNumber( (word *)&Value, pStart+2 );
                    Vec_IntPush( vValues, Value );
                }
                //Vec_IntPrint( vValues );
                nBits = Abc_Base2Log( Vec_IntSize(vValues) );
                if Vec_IntSize(vValues) != (1 << nBits) {
                    Vec_IntFree( vValues );
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read module \"%s\".", pName );
                }
                assert(Width1 == nBits, "Width1 == nBits");
                // create bitmap
                nInts = Abc_BitWordNum( Width2 * Vec_IntSize(vValues) );
                pTable = (unsigned *)Mem_FlexEntryFetch( p.pMemTable, nInts * sizeof(unsigned) );
                memset( pTable, 0, nInts * sizeof(unsigned) );
                Vec_IntForEachEntry( vValues, Value, v )
                    for ( b = 0; b < Width2; b++ )
                        if ( (Value >> b) & 1 )
                            Abc_InfoSetBit( pTable, v * Width2 + b );
                Vec_PtrPush( p.vTables, pTable );
                Vec_IntFree( vValues );
                continue;
            }
            if p.pNtk != nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Network is already defined." ); }
            p.pNtk = Wlc_NtkAlloc( pName, Vec_IntSize(p.vStarts) );
            p.pNtk.pManName = Abc_NamStart( Vec_IntSize(p.vStarts), 20 );
            p.pNtk.pMemTable = p.pMemTable; p.pMemTable = nil;
            p.pNtk.vTables = p.vTables; p.vTables = nil;
            // read the argument definitions
            for  pName = Wlc_PrsStrtok( nil, "(,)" )  {
                pName = Wlc_PrsSkipSpaces( pName );
                if (Wlc_PrsStrCmp( pName, "input" ) != 0) || (Wlc_PrsStrCmp( pName, "output" ) != 0) || (Wlc_PrsStrCmp( pName, "wire" ) != 0) {
                    if !(Wlc_PrsReadDeclaration( p, pName ) != 0) {
                        return 0; }
                }
            }
        
        } else if Wlc_PrsStrCmp( pStart, "endmodule" ) != 0 {
            var vTemp *Vec_Int_t  = Vec_IntStartNatural( Wlc_NtkObjNumMax(p.pNtk) );
            Vec_IntAppend( &p.pNtk.vNameIds, vTemp );
            Vec_IntFree( vTemp );
            if p.pNtk.vInits {
                // move FO/FI to be part of CI/CO
                assert((Vec_IntSize(&p.pNtk.vFfs) & 1) == 0, "(Vec_IntSize(&p.pNtk.vFfs) & 1) == 0");
                assert(Vec_IntSize(&p.pNtk.vFfs) == 2 * Vec_IntSize(p.pNtk.vInits), "Vec_IntSize(&p.pNtk.vFfs) == 2 * Vec_IntSize(p.pNtk.vInits)");
                // inlining of macro Wlc_NtkForEachFf
                for  i = 0; (i < Vec_IntSize(&p.pNtk.vFfs)) && (((pObj) = Wlc_NtkFf(p.pNtk, i)), 1); i++  {
                    if i & 1 {
                        Wlc_ObjSetCo( p.pNtk, pObj, 1 );
                    } else {
                        Wlc_ObjSetCi( p.pNtk, pObj ); }
                }
                Vec_IntClear( &p.pNtk.vFfs );
                // convert init values into binary string
                //Vec_IntPrint( p->pNtk->vInits );
                p.pNtk.pInits = Wlc_PrsConvertInitValues( p.pNtk );
                //printf( "%s\n", p->pNtk->pInits );
            }
            if p.pNtk.vArsts && !p.pNtk.fAsyncRst {
                var i, NameIdArst int32;
                Vec_IntForEachEntry( p.pNtk.vArsts, NameIdArst, i )
                {
                    if Wlc_PrsCheckBitConst0(p.pNtk, NameIdArst) != 0 {
                        continue; }
                    p.pNtk.fAsyncRst = 1;
                    fmt.Printf("Detected async reset \"%s\".\n", Abc_NamStr(p.pNtk.pManName, NameIdArst));
                    break;
                }
            }
            if p.vPoPairs {
                assert(Vec_StrEntryLast(p.vPoPairs) == 0, "Vec_StrEntryLast(p.vPoPairs) == 0");
                Vec_StrPush( p.vPoPairs, 0 );
                pName = Vec_StrArray(p.vPoPairs);
                for  *pName  {
                    // inlining of macro Wlc_NtkForEachPo
                    for  i = 0; (i < Wlc_NtkPoNum(p.pNtk)) && (((pObj) = Wlc_NtkPo(p.pNtk, i)), 1); i++  {
                        if !(int32(strings.Compare(Wlc_ObjName(p.pNtk, Wlc_ObjId(p.pNtk, pObj)), pName)) != 0) {
                            Vec_IntPush( &p.pNtk.vPoPairs, i );
                            break;
                        }
                    }
                    assert(i < Wlc_NtkPoNum(p.pNtk), "i < Wlc_NtkPoNum(p.pNtk)");
                    pName += len(pName) + 1;
                }
                assert(Vec_IntSize(&p.pNtk.vPoPairs) % 2 == 0, "Vec_IntSize(&p.pNtk.vPoPairs) % 2 == 0");
                fmt.Printf("Finished parsing %d output pairs to be checked for equivalence.\n", Vec_IntSize(&p.pNtk.vPoPairs)/2);
            }
            break;
        
        // these are read as part of the interface
        } else if (Wlc_PrsStrCmp( pStart, "input" ) != 0) || (Wlc_PrsStrCmp( pStart, "output" ) != 0) || (Wlc_PrsStrCmp( pStart, "wire" ) != 0) || (Wlc_PrsStrCmp( pStart, "reg" ) != 0) {
            if !(Wlc_PrsReadDeclaration( p, pStart ) != 0) {
                return 0; }
        
        } else if Wlc_PrsStrCmp( pStart, "assign" ) != 0 {
            var Type, NameId, fFound, XValue int32;
            pStart += len("assign");
            // read name
            pStart = Wlc_PrsFindName( pStart, &pName );
            if pStart == nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name after assign." ); }
            NameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
            if !fFound {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }
            // read definition
            Type = Wlc_PrsFindDefinition( p, pStart, p.vFanins, &XValue );
            if Type != 0 {
                pObj = Wlc_NtkObj( p.pNtk, NameId );
                Wlc_ObjUpdateType( p.pNtk, pObj, Type );
                Wlc_ObjAddFanins( p.pNtk, pObj, p.vFanins );
                pObj.fXConst = XValue;
            
            } else {
                return 0; }
        
        } else if Wlc_PrsStrCmp( pStart, "table" ) != 0 {
            // THIS IS A HACK to detect tables
            var NameId int32;
            var fFound int32;
            var iTable int32 = strconv.Atoi(pStart + len("table"));
            // find opening
            pStart = Wlc_PrsFindSymbol( pStart, '(' );
            if pStart == nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read table." ); }
            // read input
            pStart = Wlc_PrsFindName( pStart+1, &pName );
            if pStart == nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name after assign." ); }
            NameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
            if !fFound {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }
            // save inputs
            Vec_IntClear( p.vFanins );
            Vec_IntPush( p.vFanins, NameId );
            Vec_IntPush( p.vFanins, iTable );
            // find comma
            pStart = Wlc_PrsFindSymbol( pStart, ',' );
            if pStart == nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read table." ); }
            // read output
            pStart = Wlc_PrsFindName( pStart+1, &pName );
            if pStart == nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name after assign." ); }
            NameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
            if !fFound {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }
            pObj = Wlc_NtkObj( p.pNtk, NameId );
            Wlc_ObjUpdateType( p.pNtk, pObj, WLC_OBJ_TABLE );
            Wlc_ObjAddFanins( p.pNtk, pObj, p.vFanins );
        
        } else if Wlc_PrsStrCmp( pStart, "always" ) != 0 {
            // THIS IS A HACK to detect always statement representing combinational MUX
            var NameId int32;
            var NameIdOut int32 = -1;
            var fFound int32;
            var nValues int32;
            var fDefaultFound int32;
            // find control
            pStart = Wlc_PrsFindWord( pStart, "case", &fFound );
            if pStart == nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read case statement." ); }
            // read the name
            pStart = Wlc_PrsFindSymbol( pStart, '(' );
            if pStart == nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read table." ); }
            pStart = Wlc_PrsFindSymbol( pStart+1, '(' );
            if pStart == nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read table." ); }
            pStart = Wlc_PrsFindName( pStart+1, &pName );
            if pStart == nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name after case." ); }
            NameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
            if !fFound {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }
            Vec_IntClear( p.vFanins );
            Vec_IntPush( p.vFanins, NameId );
            // read data inputs
            pObj = Wlc_NtkObj( p.pNtk, NameId );
            if pObj == nil {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot find the object in case statement." ); }
            // remember the number of values
            nValues = (1 << Wlc_ObjRange(pObj));
            for {                
                // find opening
                pStart = Wlc_PrsFindSymbol( pStart, ':' );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot find colon in the case statement." ); }
                // find output name
                pStart = Wlc_PrsFindName( pStart+1, &pName );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name after case." ); }
                NameIdOut = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                if !fFound {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }
                // find equality
                pStart = Wlc_PrsFindSymbol( pStart, '=' );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot find equality in the case statement." ); }
                // find input name
                pStart = Wlc_PrsSkipSpaces( pStart+1 );
                pStart = Wlc_PrsReadName( p, pStart, p.vFanins );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name inside case statement." ); }
                // consider default
                if fDefaultFound != 0 {
                    var EntryLast int32  = Vec_IntEntryLast( p.vFanins );
                    if nValues != Vec_IntSize(p.vFanins)-2 {
                        Vec_IntFillExtra( p.vFanins, nValues + 1, EntryLast );
                    } else {
                        Vec_IntPop(p.vFanins); }
                    // get next line and check its opening character
                    pStart = Wlc_PrsStr(p, Vec_IntEntry(p.vStarts, ++i));
                    pStart = Wlc_PrsSkipSpaces( pStart );
                
                } else {
                    // get next line and check its opening character
                    pStart = Wlc_PrsStr(p, Vec_IntEntry(p.vStarts, ++i));
                    pStart = Wlc_PrsSkipSpaces( pStart );
                    if Wlc_PrsIsDigit(pStart) != 0 {
                        continue; }
                    if Wlc_PrsStrCmp( pStart, "default" ) != 0 {
                        fDefaultFound = 1;
                        continue;
                    }
                }
                // find closing
                pStart = Wlc_PrsFindWord( pStart, "endcase", &fFound );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read case statement." ); }
                // find closing
                pStart = Wlc_PrsFindWord( pStart, "end", &fFound );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read case statement." ); }
                pStart = Wlc_PrsSkipSpaces( pStart );
                break;
            }
            // check range of the control
            if nValues < Vec_IntSize(p.vFanins) - 1 { // may occur if default is not there
            
                //return Wlc_PrsWriteErrorMessage( p, pStart, "The number of values in the case statement is wrong.", pName );
                fmt.Printf("Warning:  The number of values in the case statement is wrong.\n");
                Vec_IntShrink(p.vFanins,nValues+1);
            
            } else if nValues > Vec_IntSize(p.vFanins) - 1 {
                return Wlc_PrsWriteErrorMessage( p, pStart, "The number of values in the case statement is wrong.", pName ); }
            if Wlc_ObjRange(pObj) == 1 {
//                return Wlc_PrsWriteErrorMessage( p, pStart, "Always-statement with 1-bit control is not bit-blasted correctly.", pName );
                fmt.Printf("Warning:  Case-statement with 1-bit control is treated as a 2:1 MUX (correct for unsigned signals only).\n");
            }
            pObj = Wlc_NtkObj( p.pNtk, NameIdOut );
            Wlc_ObjUpdateType( p.pNtk, pObj, WLC_OBJ_MUX );
            Wlc_ObjAddFanins( p.pNtk, pObj, p.vFanins );
            goto startword;
        
        } else if Wlc_PrsStrCmp( pStart, "CPL_FF" ) != 0 {
            var NameId int32 = -1;
            var NameIdIn int32 = -1;
            var NameIdOut int32 = -1;
            var fFound int32;
            var nBits int32 = 1;
            var fFlopIn int32;
            var fFlopOut int32;
            pStart += len("CPL_FF");
            if pStart[0] == '#' {
                nBits = strconv.Atoi(pStart+1); }
            // read names
            for {
                pStart = Wlc_PrsFindSymbol( pStart, '.' );
                if pStart == nil {
                    break; }
                pStart = Wlc_PrsSkipSpaces( pStart+1 );
                if !p.pNtk.fAsyncRst && !strncmp(pStart, "arst", 4) && pStart[4] != 'v' {
                    var NameIdArst int32;
                    pStart = Wlc_PrsFindSymbol( pStart, '(' );
                    if pStart == nil {
                        return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read opening parenthesis in the flop description." ); }
                    pStart = Wlc_PrsSkipSpaces( pStart+1 );
                    if Wlc_PrsIsDigit(pStart) != 0 {
                        var Range, Signed, XValue int32;
                        var vFanins *Vec_Int_t  = Vec_IntAlloc( 100 );
                        pStart = Wlc_PrsReadConstant( p, pStart, vFanins, &Range, &Signed, &XValue );
                        if pStart && Vec_IntSize(vFanins) == 1 && Vec_IntEntry(vFanins, 0) == 0 {
                            Vec_IntFree( vFanins );
                            continue;
                        }
                        fmt.Printf("Detected async reset.\n");
                        p.pNtk.fAsyncRst = 1;
                        Vec_IntFree( vFanins );
                        continue;
                    }
                    pStart = Wlc_PrsFindName( pStart, &pName );
                    if pStart == nil {
                        return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name inside flop description." ); }
                    NameIdArst = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                    if p.pNtk.vArsts == nil {
                        p.pNtk.vArsts = Vec_IntAlloc( 100 ); }
                    Vec_IntPushUnique( p.pNtk.vArsts, NameIdArst );
                    continue;
                }
                if pStart[0] != 'd' && (pStart[0] != 'q' || pStart[1] == 'b') && strncmp(pStart, "arstval", 7) {
                    continue; }
                fFlopIn = (pStart[0] == 'd');
                fFlopOut = (pStart[0] == 'q');
                pStart = Wlc_PrsFindSymbol( pStart, '(' );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read opening parenthesis in the flop description." ); }
                pStart = Wlc_PrsFindName( pStart+1, &pName );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name inside flop description." ); }
                if fFlopIn != 0 {
                    NameIdIn = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else if fFlopOut != 0 { 
                    NameIdOut = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else {
                    NameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound ); }
                if !fFound {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }
            }
            if NameIdIn == -1 || NameIdOut == -1 {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Name of flop input or flop output is missing." ); }
            // create flop output
            pObj = Wlc_NtkObj( p.pNtk, NameIdOut );
            Wlc_ObjUpdateType( p.pNtk, pObj, WLC_OBJ_FO );
            Vec_IntPush( &p.pNtk.vFfs, NameIdOut );
            //if ( nBits != Wlc_ObjRange(pObj) )
            //    printf( "Warning!  Flop output \"%s\" has bit-width (%d) that differs from the flop declaration (%d)\n", Abc_NamStr(p->pNtk->pManName, NameIdOut), Wlc_ObjRange(pObj), nBits );
            // create flop input
            pObj = Wlc_NtkObj( p.pNtk, NameIdIn );
            Vec_IntPush( &p.pNtk.vFfs, NameIdIn );
            //if ( nBits != Wlc_ObjRange(pObj) )
            //    printf( "Warning!  Flop input \"%s\" has bit-width (%d) that differs from the flop declaration (%d)\n", Abc_NamStr(p->pNtk->pManName, NameIdIn), Wlc_ObjRange(pObj), nBits );
            // save flop init value
            if NameId == -1 {
                fmt.Printf("Initial value of flop \"%s\" is not specified. Zero is assumed.\n", Abc_NamStr(p.pNtk.pManName, NameIdOut));
            } else {
                pObj = Wlc_NtkObj( p.pNtk, NameId );
                if nBits != Wlc_ObjRange(pObj) {
                    fmt.Printf("Warning!  Flop init signal \"%s\" bit-width (%d) is different from the flop declaration (%d)\n", Abc_NamStr(p.pNtk.pManName, NameId), Wlc_ObjRange(pObj), nBits); }
            }
            if p.pNtk.vInits == nil {
                p.pNtk.vInits = Vec_IntAlloc( 100 ); }
            Vec_IntPush( p.pNtk.vInits, NameId > 0 ? NameId : -nBits );
            // printf( "Created flop %s with range %d and init value %d (nameId = %d)\n", 
            //     Abc_NamStr(p->pNtk->pManName, NameIdOut), Wlc_ObjRange(pObj), nBits, NameId );
        
        } else if Wlc_PrsStrCmp( pStart, "ABC_DFFRSE" ) != 0 {
            var NameId [10]int32 ;
            var fFound int32;
            var fFlopIn int32;
            var fFlopClk int32;
            var fFlopRst int32;
            var fFlopSet int32;
            var fFlopEna int32;
            var fFlopAsync int32;
            var fFlopSre int32;
            var fFlopInit int32;
            var fFlopOut int32;
            pStart += len("ABC_DFFRSE");
            for {
                pStart = Wlc_PrsFindSymbol( pStart, '.' );
                if pStart == nil {
                    break; }
                pStart = Wlc_PrsSkipSpaces( pStart+1 );
                fFlopIn    = (pStart[0] == 'd');
                fFlopClk   = (pStart[0] == 'c');
                fFlopRst   = (pStart[0] == 'r');
                fFlopSet   = (pStart[0] == 's' && pStart[1] == 'e');
                fFlopEna   = (pStart[0] == 'e');
                fFlopAsync = (pStart[0] == 'a');
                fFlopSre   = (pStart[0] == 's' && pStart[1] == 'r');
                fFlopInit  = (pStart[0] == 'i');
                fFlopOut   = (pStart[0] == 'q');
                pStart = Wlc_PrsFindSymbol( pStart, '(' );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read opening parenthesis in the flop description." ); }
                pStart = Wlc_PrsFindName( pStart+1, &pName );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name inside flop description." ); }
                if fFlopIn != 0 {
                    NameId[0] = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else if fFlopClk != 0 { 
                    NameId[1] = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else if fFlopRst != 0 { 
                    NameId[2] = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else if fFlopSet != 0 { 
                    NameId[3] = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else if fFlopEna != 0 { 
                    NameId[4] = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else if fFlopAsync != 0 { 
                    NameId[5] = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else if fFlopSre != 0 { 
                    NameId[6] = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else if fFlopInit != 0 { 
                    NameId[7] = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else if fFlopOut != 0 { 
                    NameId[8] = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else {
                    assert(false, "0"); }
                if !fFound {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }
            }
            if NameId[0] == -1 || NameId[7] == -1 {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Name of flop input or flop output is missing." ); }
            // create output
            pObj = Wlc_NtkObj( p.pNtk, NameId[8] );
            Wlc_ObjUpdateType( p.pNtk, pObj, WLC_OBJ_FF );
            Vec_IntClear( p.vFanins );
            Vec_IntPush( p.vFanins, NameId[0] );
            Vec_IntPush( p.vFanins, NameId[1] );
            Vec_IntPush( p.vFanins, NameId[2] );
            Vec_IntPush( p.vFanins, NameId[3] );
            Vec_IntPush( p.vFanins, NameId[4] );
            Vec_IntPush( p.vFanins, NameId[5] );
            Vec_IntPush( p.vFanins, NameId[6] );
            Vec_IntPush( p.vFanins, NameId[7] );
            Wlc_ObjAddFanins( p.pNtk, pObj, p.vFanins );
        
        } else if Wlc_PrsStrCmp( pStart, "ABC_DFF" ) != 0 {
            var NameId int32 = -1;
            var NameIdIn int32 = -1;
            var NameIdOut int32 = -1;
            var fFound int32;
            var nBits int32 = 1;
            var fFlopIn int32;
            var fFlopOut int32;
            pStart += len("ABC_DFF");
            for {
                pStart = Wlc_PrsFindSymbol( pStart, '.' );
                if pStart == nil {
                    break; }
                pStart = Wlc_PrsSkipSpaces( pStart+1 );
                fFlopIn = (pStart[0] == 'd');
                fFlopOut = (pStart[0] == 'q');
                pStart = Wlc_PrsFindSymbol( pStart, '(' );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read opening parenthesis in the flop description." ); }
                pStart = Wlc_PrsFindName( pStart+1, &pName );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name inside flop description." ); }
                if fFlopIn != 0 {
                    NameIdIn = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else if fFlopOut != 0 { 
                    NameIdOut = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                } else {
                    NameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound ); }
                if !fFound {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }
            }
            if NameIdIn == -1 || NameIdOut == -1 {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Name of flop input or flop output is missing." ); }
            // create flop output
            pObj = Wlc_NtkObj( p.pNtk, NameIdOut );
            Wlc_ObjUpdateType( p.pNtk, pObj, WLC_OBJ_FO );
            Vec_IntPush( &p.pNtk.vFfs, NameIdOut );
            nBits = Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdOut));
            // create flop input
            pObj = Wlc_NtkObj( p.pNtk, NameIdIn );
            Vec_IntPush( &p.pNtk.vFfs, NameIdIn );
            // compare bit-width
            if Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdIn)) != nBits {
                fmt.Printf("Warning!  Flop input \"%s\" bit-width (%d) is different from that of flop output (%d)\n", Abc_NamStr(p.pNtk.pManName, NameId), Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdIn)), nBits); }
            // save flop init value
            if NameId == -1 {
                fmt.Printf("Initial value of flop \"%s\" is not specified. Zero is assumed.\n", Abc_NamStr(p.pNtk.pManName, NameIdOut));
            } else {
                if Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameId)) != nBits {
                    fmt.Printf("Warning!  Flop init signal \"%s\" bit-width (%d) is different from that of flop output (%d)\n", Abc_NamStr(p.pNtk.pManName, NameId), Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameId)), nBits); }
            }
            if p.pNtk.vInits == nil {
                p.pNtk.vInits = Vec_IntAlloc( 100 ); }
            Vec_IntPush( p.pNtk.vInits, NameId > 0 ? NameId : -Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdOut)) );
            // printf( "Created flop %s with range %d and init value %d (nameId = %d)\n", 
            //     Abc_NamStr(p->pNtk->pManName, NameIdOut), Wlc_ObjRange(pObj), nBits, NameId );
            p.pNtk.fEasyFfs = 1;
        
        } else if Wlc_PrsStrCmp( pStart, "CPL_MEM_" ) != 0 {
            var pNameId *int32 = nil;
            var NameOutput int32;
            var NameMi int32 = -1;
            var NameMo int32 = -1;
            var NameAddr int32 = -1;
            var NameDi int32 = -1;
            var NameDo int32 = -1;
            var fFound int32;
            var fRead int32 = 1;
            pStart += len("CPL_MEM_");
            if pStart[0] == 'W' {
                fRead = 0; }
            // read names
            for {
                pStart = Wlc_PrsFindSymbol( pStart, '.' );
                if pStart == nil {
                    break; }
                pStart = Wlc_PrsSkipSpaces( pStart+1 );
                if !strncmp(pStart, "mem_data_in", 11) {
                    pNameId = &NameMi;
                } else if !strncmp(pStart, "data_in", 7) {
                    pNameId = &NameDi;
                } else if !strncmp(pStart, "data_out", 8) {
                    if fRead != 0 {
                        pNameId = &NameDo;
                    } else {
                        pNameId = &NameMo;
                    }
                } else if !strncmp(pStart, "addr_in", 7) {
                    pNameId = &NameAddr;
                } else {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name of the input/output port." ); }
                pStart = Wlc_PrsFindSymbol( pStart, '(' );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read opening parenthesis in the flop description." ); }
                pStart = Wlc_PrsFindName( pStart+1, &pName );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name inside flop description." ); }
                *pNameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                if !fFound {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }
            }
            if (fRead != 0) && (NameMi == -1 || NameAddr == -1 || NameDo == -1) {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Name of one of signals of read port is missing." ); }
            if !(fRead != 0) && (NameMi == -1 || NameAddr == -1 || NameDi == -1 || NameMo == -1) {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Name of one of signals of write port is missing." ); }
            // create output
            if fRead != 0 {
                NameOutput = NameDo;
            } else {
                NameOutput = NameMo;
            }
            pObj = Wlc_NtkObj( p.pNtk, NameOutput );
            Wlc_ObjUpdateType( p.pNtk, pObj, fRead ? WLC_OBJ_READ : WLC_OBJ_WRITE );
            Vec_IntClear( p.vFanins );
            Vec_IntPush( p.vFanins, NameMi );
            Vec_IntPush( p.vFanins, NameAddr );
            if !(fRead != 0) {
                Vec_IntPush( p.vFanins, NameDi ); }
            //printf( "Memory %s ", fRead ? "Read" : "Write" ); printf( "Fanins: " );  Vec_IntPrint( p->vFanins );
            Wlc_ObjAddFanins( p.pNtk, pObj, p.vFanins );
        
        } else if Wlc_PrsStrCmp( pStart, "ABC_READ" ) != 0 {
            var pNameId *int32 = nil;
            var NameMemIn int32 = -1;
            var NameData int32 = -1;
            var NameAddr int32 = -1;
            var fFound int32;
            pStart += len("ABC_READ");
            for {
                pStart = Wlc_PrsFindSymbol( pStart, '.' );
                if pStart == nil {
                    break; }
                pStart = Wlc_PrsSkipSpaces( pStart+1 );
                if !strncmp(pStart, "mem_in", 6) {
                    pNameId = &NameMemIn;
                } else if !strncmp(pStart, "addr", 4) {
                    pNameId = &NameAddr;
                } else if !strncmp(pStart, "data", 4) {
                    pNameId = &NameData;
                } else {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name of the input/output port." ); }
                pStart = Wlc_PrsFindSymbol( pStart, '(' );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read opening parenthesis in the flop description." ); }
                pStart = Wlc_PrsFindName( pStart+1, &pName );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name inside flop description." ); }
                *pNameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                if !fFound {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }
            }
            if NameMemIn == -1 || NameAddr == -1 || NameData == -1 {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Name of one of signals of read port is missing." ); }
            // create output
            pObj = Wlc_NtkObj( p.pNtk, NameData );
            Wlc_ObjUpdateType( p.pNtk, pObj, WLC_OBJ_READ );
            Vec_IntClear( p.vFanins );
            Vec_IntPush( p.vFanins, NameMemIn );
            Vec_IntPush( p.vFanins, NameAddr );
            Wlc_ObjAddFanins( p.pNtk, pObj, p.vFanins );
            p.pNtk.fMemPorts = 1;
        
        } else if Wlc_PrsStrCmp( pStart, "ABC_WRITE" ) != 0 {
            var pNameId *int32 = nil;
            var NameMemIn int32 = -1;
            var NameMemOut int32 = -1;
            var NameData int32 = -1;
            var NameAddr int32 = -1;
            var fFound int32;
            pStart += len("ABC_WRITE");
            for {
                pStart = Wlc_PrsFindSymbol( pStart, '.' );
                if pStart == nil {
                    break; }
                pStart = Wlc_PrsSkipSpaces( pStart+1 );
                if !strncmp(pStart, "mem_in", 6) {
                    pNameId = &NameMemIn;
                } else if !strncmp(pStart, "mem_out", 7) {
                    pNameId = &NameMemOut;
                } else if !strncmp(pStart, "data", 4) {
                    pNameId = &NameData;
                } else if !strncmp(pStart, "addr", 4) {
                    pNameId = &NameAddr;
                } else {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name of the input/output port." ); }
                pStart = Wlc_PrsFindSymbol( pStart, '(' );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read opening parenthesis in the flop description." ); }
                pStart = Wlc_PrsFindName( pStart+1, &pName );
                if pStart == nil {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name inside flop description." ); }
                *pNameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );
                if !fFound {
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }
            }
            if NameMemIn == -1 || NameAddr == -1 || NameData == -1 || NameMemOut == -1 {
                return Wlc_PrsWriteErrorMessage( p, pStart, "Name of one of signals of write port is missing." ); }
            // create output
            pObj = Wlc_NtkObj( p.pNtk, NameMemOut );
            Wlc_ObjUpdateType( p.pNtk, pObj, WLC_OBJ_WRITE );
            Vec_IntClear( p.vFanins );
            Vec_IntPush( p.vFanins, NameMemIn );
            Vec_IntPush( p.vFanins, NameAddr );
            Vec_IntPush( p.vFanins, NameData );
            Wlc_ObjAddFanins( p.pNtk, pObj, p.vFanins );
            p.pNtk.fMemPorts = 1;
        
        } else if (Wlc_PrsStrCmp( pStart, "CPL_RROT" ) != 0) || (Wlc_PrsStrCmp( pStart, "CPL_LROT" ) != 0) {                                                                                                                         
            // CPL_RROT #(128, 6) I_47479(.o ( E_46713 )  , .d ( E_46718 )  , .s ( E_46712 )  );                                  
            var right_rotation int32  = Wlc_PrsStrCmp( pStart, "CPL_RROT" );                                                             
            var NameId int32 = -1;
            var NameIdOut int32 = -1;
            var NameIdInD int32 = -1;
            var NameIdInS int32 = -1;
            var fFound int32;
            var fRotInD int32;
            var fRotInS int32;
            var fRotOut int32;                   
            pStart += len("CPL_RROT");                                                                                         
                                                                                                                                  
            // NOTE: no need to parse the parameter values                                                                        
            //if ( pStart[0] == '#' )                                                                                             
                                                                                                                                  
            // read names                                                                                                         
            for {                                                                                                                     
                pStart = Wlc_PrsFindSymbol( pStart, '.' );                                                                        
                if pStart == nil {                                                                                             
                    break; }                                                                                                        
                pStart = Wlc_PrsSkipSpaces( pStart+1 );                                                                           
                if pStart[0] != 'o' && pStart[0] != 'd' && pStart[0] != 's' {                                                    
                    continue; }                                                                                                     
                fRotInD = (pStart[0] == 'd');                                                                                     
                fRotInS = (pStart[0] == 's');                                                                                     
                fRotOut = (pStart[0] == 'o');                                                                                     
                pStart = Wlc_PrsFindSymbol( pStart, '(' );                                                                        
                if pStart == nil {                                                                                             
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read opening parenthesis in the rotation description." ); } 
                pStart = Wlc_PrsFindName( pStart+1, &pName );                                                                     
                if pStart == nil {                                                                                             
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read name inside rotation description." ); }                
                if fRotInD != 0 {                                                                                                    
                    NameIdInD = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );                                         
                } else if fRotInS != 0 {                                                                                               
                    NameIdInS = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );                                         
                } else if fRotOut != 0 {                                                                                               
                    NameIdOut = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound );                                         
                } else {                                                                                                              
                    NameId = Abc_NamStrFindOrAdd( p.pNtk.pManName, pName, &fFound ); }                                            
                if !fFound {                                                                                                    
                    return Wlc_PrsWriteErrorMessage( p, pStart, "Name %s is not declared.", pName ); }                              
            }                                                                                                                     
            if NameIdOut == -1 || NameIdInD == -1 || NameIdInS == -1 {                                                          
                return Wlc_PrsWriteErrorMessage( p, pStart, "Some fields of CPL_ROT are missing." ); }                              
            // create rot output                                                                                                  
            pObj = Wlc_NtkObj( p.pNtk, NameIdOut );                                                                              
            Wlc_ObjUpdateType( p.pNtk, pObj, right_rotation ? WLC_OBJ_ROTATE_R : WLC_OBJ_ROTATE_L );                             
            Vec_IntClear( p.vFanins );                                                                                           
            Vec_IntPush( p.vFanins, NameIdInD );                                                                                 
            Vec_IntPush( p.vFanins, NameIdInS );                                                                                 
            Wlc_ObjAddFanins( p.pNtk, pObj, p.vFanins );                                                                        
                                                                                                                                 
        } else if pStart[0] == '(' && pStart[1] == '*' { // skip comments
        
            for  *pStart++ != ')'  {; }
            pStart = Wlc_PrsSkipSpaces( pStart );
            goto startword;
        
        } else if pStart[0] != '`' {
            var iLine int32  = Wlc_PrsFindLine(p, pStart);
            pStart = Wlc_PrsFindName( pStart, &pName );
            return Wlc_PrsWriteErrorMessage( p, pStart, "Cannot read line %d beginning with %s.", iLine, (!pName || !pName[0]) ? "\"?\"" : pName );
        }
    }
    if p.nNonZero[0] {
        fmt.Printf("Warning: Input file contains %d objects with non-zero-based ranges.\n", p.nNonZero[0]);
        fmt.Printf("For example, signal with range [%d:%d] is declared in line %d.\n", p.nNonZero[1], p.nNonZero[2], p.nNonZero[3]);
    }
    if p.nNegative[0] {
        fmt.Printf("Warning: Input file contains %d objects with negative ranges.\n", p.nNegative[0]);
        fmt.Printf("For example, signal with range [%d:%d] is declared in line %d.\n", p.nNegative[1], p.nNegative[2], p.nNegative[3]);
    }
    if p.nReverse[0] {
        fmt.Printf("Warning: Input file contains %d objects with reversed ranges.\n", p.nReverse[0]);
        fmt.Printf("For example, signal with range [%d:%d] is declared in line %d.\n", p.nReverse[1], p.nReverse[2], p.nReverse[3]);
    }
    return 1;
}
func Wlc_ReadVer(pFileName string, pStr string) *Wlc_Ntk_t {
    var p *Wlc_Prs_t;
    var pNtk *Wlc_Ntk_t  = nil;
    // either file name or string is given
    assert((pFileName == nil) != (pStr == nil), "(pFileName == nil) != (pStr == nil)");
    // start the parser 
    p = Wlc_PrsStart( pFileName, pStr );
    if p == nil {
        return nil; }
    // detect lines
    if !(Wlc_PrsPrepare( p ) != 0) {
        goto finish; }
    // parse models
    if !(Wlc_PrsDerive( p ) != 0) {
        goto finish; }
    // derive topological order
    if p.pNtk {
        var pObj *Wlc_Obj_t; var i int32;
        // inlining of macro Wlc_NtkForEachObj
        for  i = 1; (i < Wlc_NtkObjNumMax(p.pNtk)) && (((pObj) = Wlc_NtkObj(p.pNtk, i)), 1); i++  {
            if pObj.Type == WLC_OBJ_FF {
                Vec_IntPush( &p.pNtk.vFfs2, Wlc_ObjId(p.pNtk, pObj) ); }
        }
        pNtk = Wlc_NtkDupDfs( p.pNtk, 0, 1 );
        pNtk.pSpec = Abc_UtilStrsav( pFileName );
    }
finish:
    Wlc_PrsPrintErrorMessage( p );
    Wlc_PrsStop( p );
    return pNtk;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadWordTest(pFileName string) {
    var pNew *Gia_Man_t;
    var pNtk *Wlc_Ntk_t  = Wlc_ReadVer( pFileName, nil );
    if pNtk == nil {
        return; }
    Wlc_WriteVer( pNtk, "test.v", 0, 0 );

    pNew = Wlc_NtkBitBlast( pNtk, nil );
    Gia_AigerWrite( pNew, "test.aig", 0, 0, 0 );
    Gia_ManStop( pNew );

    Wlc_NtkFree( pNtk );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

