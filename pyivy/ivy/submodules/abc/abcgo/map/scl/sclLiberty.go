package scl

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [sclLiberty.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Liberty parser.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclLiberty.c,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sclLib.h"
//#include "misc/st/st.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// #define ABC_MAX_LIB_STR_LEN 5000

// entry types
 type Scl_LibertyType_t int32
const ( 
    SCL_LIBERTY_NONE Scl_LibertyType_t = 0
        // 0:  unknown
    SCL_LIBERTY_PROC Scl_LibertyType_t = SCL_LIBERTY_NONE + 1
            // 1:  procedure :  key(head){body}
    SCL_LIBERTY_EQUA Scl_LibertyType_t = SCL_LIBERTY_PROC + 1
            // 2:  equation  :  key:head;
    SCL_LIBERTY_LIST Scl_LibertyType_t = SCL_LIBERTY_EQUA + 1             // 3:  list      :  key(head) 

)

 

   
type Scl_Pair_t struct {
    Beg             int32;          // item beginning
    End             int32;          // item end
}

   
type Scl_Item_t struct {
    Type             int32;         // Scl_LibertyType_t
    iLine             int32;        // file line where the item's spec begins
    Key      Scl_Pair_t;          // key part
    Head      Scl_Pair_t;         // head part 
    Body      Scl_Pair_t;         // body part
    Next             int32;         // next item in the list 
    Child             int32;        // first child item 
}

   
type Scl_Tree_t struct {
    pFileName           string;    // input Liberty file name
    pContents           string;    // file contents
    nContents             int32;    // file size
    nLines             int32;       // line counter
    nItems             int32;       // number of items
    nItermAlloc             int32;  // number of items allocated
    pItems     *Scl_Item_t;       // the items
    pError           string;       // the error string
    clkStart         abctime;     // beginning time
    vBuffer      *Vec_Str_t;      // temp string buffer
}

func Scl_LibertyRoot(p *Scl_Tree_t) *Scl_Item_t { return p.pItems;                                                 }
func Scl_LibertyItem(p *Scl_Tree_t, v int32) *Scl_Item_t { assert(v < p.nItems, "v < p.nItems"); if v < 0 {
    return nil;
} else {
    return p.pItems + v;
}     }
func Scl_LibertyCompare(p *Scl_Tree_t, Pair Scl_Pair_t, pStr string) int32 { return strncmp( p.pContents+Pair.Beg, pStr, Pair.End-Pair.Beg ) || (int32(len(pStr)) != Pair.End-Pair.Beg); }
func Scl_PrintWord(pFile *os.File, p *Scl_Tree_t, Pair Scl_Pair_t) { var pBeg string = p.pContents+Pair.Beg;
var pEnd string = p.pContents+Pair.End; for  pBeg < pEnd  { fputc( *pBeg++, pFile ); } }
func Scl_PrintSpace(pFile *os.File, nOffset int32) { var i int32; for  i = 0; i < nOffset; i++  { fputc(' ', pFile); }         }
func Scl_LibertyItemId(p *Scl_Tree_t, pItem *Scl_Item_t) int32 { return pItem - p.pItems;                                         }

// Original macro:
// #define Scl_ItemForEachChild( p, pItem, pChild ) \
//     for ( pChild = Scl_LibertyItem(p, pItem->Child); pChild; pChild = Scl_LibertyItem(p, pChild->Next) )

// Original macro:
// #define Scl_ItemForEachChildName( p, pItem, pChild, pName ) \
//     for ( pChild = Scl_LibertyItem(p, pItem->Child); pChild; pChild = Scl_LibertyItem(p, pChild->Next) ) if ( Scl_LibertyCompare(p, pChild->Key, pName) ) {} else


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Prints parse tree in Liberty format.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibertyParseDumpItem(pFile *os.File, p *Scl_Tree_t, pItem *Scl_Item_t, nOffset int32) {
    if pItem.Type == SCL_LIBERTY_PROC {
        Scl_PrintSpace( pFile, nOffset );
        Scl_PrintWord( pFile, p, pItem.Key );
        fmt.Fprintf(pFile, "(");
        Scl_PrintWord( pFile, p, pItem.Head );
        fmt.Fprintf(pFile, ") {\n");
        if Scl_LibertyItem(p, pItem.Child) != nil {
            Scl_LibertyParseDumpItem( pFile, p, Scl_LibertyItem(p, pItem.Child), nOffset + 2 ); }
        Scl_PrintSpace( pFile, nOffset );
        fmt.Fprintf(pFile, "}\n");
    
    } else if pItem.Type == SCL_LIBERTY_EQUA {
        Scl_PrintSpace( pFile, nOffset );
        Scl_PrintWord( pFile, p, pItem.Key );
        fmt.Fprintf(pFile, " : ");
        Scl_PrintWord( pFile, p, pItem.Head );
        fmt.Fprintf(pFile, ";\n");
    
    } else if pItem.Type == SCL_LIBERTY_LIST {
        Scl_PrintSpace( pFile, nOffset );
        Scl_PrintWord( pFile, p, pItem.Key );
        fmt.Fprintf(pFile, "(");
        Scl_PrintWord( pFile, p, pItem.Head );
        fmt.Fprintf(pFile, ");\n");
    
    } else { assert(false, "0"); }
    if Scl_LibertyItem(p, pItem.Next) != nil {
        Scl_LibertyParseDumpItem( pFile, p, Scl_LibertyItem(p, pItem.Next), nOffset ); }
}
func Scl_LibertyParseDump(p *Scl_Tree_t, pFileName string) int32 {
    var pFile *os.File;
    if pFileName == nil {
        pFile = os.Stdout;
    } else {
        pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Printf("Scl_LibertyParseDump(): The output file is unavailable (absent or open).\n");
            return 0;
        }
    }
    Scl_LibertyParseDumpItem( pFile, p, Scl_LibertyRoot(p), 0 );
    if pFile != os.Stdout {
        _ = (pFile).Close(); }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Gets the name to write.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibertyCountItems(pBeg string, pEnd string) int32 {
    var Counter int32;
    for  ; pBeg < pEnd; pBeg++  {
        Counter += (*pBeg == '(' || *pBeg == ':');
    }
    return Counter;
}
// removes C-style comments
/*
void Scl_LibertyWipeOutComments( char * pBeg, char * pEnd )
{
    char * pCur, * pStart;
    for ( pCur = pBeg; pCur < pEnd; pCur++ )
    if ( pCur[0] == '/' && pCur[1] == '*' )
        for ( pStart = pCur; pCur < pEnd; pCur++ )
        if ( pCur[0] == '*' && pCur[1] == '/' )
        {
            for ( ; pStart < pCur + 2; pStart++ )
            if ( *pStart != '\n' ) *pStart = ' ';
            break;
        }
}
*/
func Scl_LibertyWipeOutComments(pBeg string, pEnd string) {
    var pCur, pStart string;
    for  pCur = pBeg; pCur < pEnd-1; pCur++  {
        if pCur[0] == '/' && pCur[1] == '*' {
            for  pStart = pCur; pCur < pEnd-1; pCur++  {
                if pCur[0] == '*' && pCur[1] == '/' {
                    for  ; pStart < pCur + 2; pStart++  {
                    if *pStart != '\n' { *pStart = ' '; }
                    }
                    break;
                }
            }
        
        } else if pCur[0] == '/' && pCur[1] == '/' {
            for  pStart = pCur; pCur < pEnd; pCur++  {
                if pCur[0] == '\n' || pCur == pEnd-1 {
                    for  ; pStart < pCur; pStart++  { *pStart = ' '; }
                    break;
                }
            }
        }
    }
}
func Scl_LibertyCharIsSpace(c byte) int32 {
    return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\\';
}
func Scl_LibertySkipSpaces(p *Scl_Tree_t, ppPos *string, pEnd string, fStopAtNewLine int32) int32 {
    var pPos string  = *ppPos;
    for  ; pPos < pEnd; pPos++  {
        if *pPos == '\n' {
            p.nLines++;
            if fStopAtNewLine != 0 {
                break; }
        }        
        if !(Scl_LibertyCharIsSpace(*pPos) != 0) {
            break; }
    }
    *ppPos = pPos;
    return pPos == pEnd;
}
// skips entry delimited by " :;(){}" and returns 1 if reached the end
func Scl_LibertySkipEntry(ppPos *string, pEnd string) int32 {
    var pPos string  = *ppPos;
    if *pPos == '\"' {
        for  pPos++; pPos < pEnd; pPos++  {
            if *pPos == '\"' {
                pPos++;
                break;
            }
        }
    
    } else {
        for  ; pPos < pEnd; pPos++  {
            if *pPos == ' ' || *pPos == '\r' || *pPos == '\n' || *pPos == '\t' ||
                 *pPos == ':' || *pPos == ';'  || 
                 *pPos == '(' || *pPos == ')'  || 
                 *pPos == '{' || *pPos == '}' {
                break; }
        }
    }
    *ppPos = pPos;
    return pPos == pEnd;
}
// finds the matching closing symbol
func Scl_LibertyFindMatch(pPos string, pEnd string) string {
    var Counter int32;
    assert(*pPos == '(' || *pPos == '{', "*pPos == '(' || *pPos == '{'");
    if *pPos == '(' {
        for  ; pPos < pEnd; pPos++  {
            if *pPos == '(' {
                Counter++; }
            if *pPos == ')' {
                Counter--; }
            if Counter == 0 {
                break; }
        }
    
    } else {
        for  ; pPos < pEnd; pPos++  {
            if *pPos == '{' {
                Counter++; }
            if *pPos == '}' {
                Counter--; }
            if Counter == 0 {
                break; }
        }
    }
    assert(*pPos == ')' || *pPos == '}', "*pPos == ')' || *pPos == '}'");
    return pPos;
}
// trims spaces around the head
func Scl_LibertyUpdateHead(p *Scl_Tree_t, Head Scl_Pair_t) Scl_Pair_t {
    var Res Scl_Pair_t;
    var pBeg string  = p.pContents + Head.Beg;
    var pEnd string  = p.pContents + Head.End;
    var pFirstNonSpace string  = nil;
    var pLastNonSpace string  = nil;
    var pChar string;
    for  pChar = pBeg; pChar < pEnd; pChar++  {
        if *pChar == '\n' {
            p.nLines++; }
        if Scl_LibertyCharIsSpace(*pChar) != 0 {
            continue; }
        pLastNonSpace = pChar;
        if pFirstNonSpace == nil {
            pFirstNonSpace = pChar; }
    }
    if pFirstNonSpace == nil || pLastNonSpace == nil {
        return Head; }
    assert(pFirstNonSpace && pLastNonSpace, "pFirstNonSpace && pLastNonSpace");
    Res.Beg = pFirstNonSpace - p.pContents;
    Res.End = pLastNonSpace  - p.pContents + 1;
    return Res;
}
// returns new item
func Scl_LibertyNewItem(p *Scl_Tree_t, Type int32) *Scl_Item_t {
    p.pItems[p.nItems].iLine = p.nLines;
    p.pItems[p.nItems].Type  = Type;
    p.pItems[p.nItems].Child = -1;
    p.pItems[p.nItems].Next  = -1;
    return p.pItems + p.nItems++;
}


/**Function*************************************************************

  Synopsis    [Gets the name to write.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibertyReadString(p *Scl_Tree_t, Pair Scl_Pair_t) string { 
    // static char Buffer[ABC_MAX_LIB_STR_LEN]; 
    var Buffer string;
    if Pair.End - Pair.Beg + 2 > Vec_StrSize(p.vBuffer) {
        Vec_StrFill( p.vBuffer, Pair.End - Pair.Beg + 100, '\0' ); }
    Buffer = Vec_StrArray( p.vBuffer );
    strncpy( Buffer, p.pContents+Pair.Beg, Pair.End-Pair.Beg ); 
    if Pair.Beg < Pair.End && Buffer[0] == '\"' {
        assert(Buffer[Pair.End-Pair.Beg-1] == '\"', "Buffer[Pair.End-Pair.Beg-1] == '\\\"'");
        Buffer[Pair.End-Pair.Beg-1] = 0;
        return Buffer + 1;
    }
    Buffer[Pair.End-Pair.Beg] = 0;
    return Buffer;
}
func Scl_LibertyItemNum(p *Scl_Tree_t, pRoot *Scl_Item_t, pName string) int32 {
    var pItem *Scl_Item_t;
    var Counter int32;
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, pRoot.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, pName) != 0 {} else {
        Counter++; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns free item.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibertyBuildItem(p *Scl_Tree_t, ppPos *string, pEnd string) int32 {
    var pItem *Scl_Item_t;
    var Key, Head, Body Scl_Pair_t;
    var pNext, pStop string;
    Key.End = 0;
    if Scl_LibertySkipSpaces( p, ppPos, pEnd, 0 ) != 0 {
        return -2; }
    Key.Beg = *ppPos - p.pContents;
    if Scl_LibertySkipEntry( ppPos, pEnd ) != 0 {
        goto exit; }
    Key.End = *ppPos - p.pContents;
    if Scl_LibertySkipSpaces( p, ppPos, pEnd, 0 ) != 0 {
        goto exit; }
    pNext = *ppPos;
    if *pNext == ':' {
        *ppPos = pNext + 1;
        if Scl_LibertySkipSpaces( p, ppPos, pEnd, 0 ) != 0 {
            goto exit; }
        Head.Beg = *ppPos - p.pContents;
        if Scl_LibertySkipEntry( ppPos, pEnd ) != 0 {
            goto exit; }
        Head.End = *ppPos - p.pContents;
        if Scl_LibertySkipSpaces( p, ppPos, pEnd, 1 ) != 0 {
            goto exit; }
        pNext = *ppPos;
        for  *pNext == '+' || *pNext == '-' || *pNext == '*' || *pNext == '/'  {
        (*ppPos) += 1;
            if Scl_LibertySkipSpaces( p, ppPos, pEnd, 0 ) != 0 {
                goto exit; }
            if Scl_LibertySkipEntry( ppPos, pEnd ) != 0 {
                goto exit; }
            Head.End = *ppPos - p.pContents;
            if Scl_LibertySkipSpaces( p, ppPos, pEnd, 1 ) != 0 {
                goto exit; }
        pNext = *ppPos;
        }
        if *pNext != ';' && *pNext != '\n' {
            goto exit; }
        *ppPos = pNext + 1;
        // end of equation
        pItem = Scl_LibertyNewItem( p, SCL_LIBERTY_EQUA );
        pItem.Key  = Key;
        pItem.Head = Scl_LibertyUpdateHead( p, Head );
        pItem.Next = Scl_LibertyBuildItem( p, ppPos, pEnd );
        if pItem.Next == -1 {
            goto exit; }
        return Scl_LibertyItemId( p, pItem );
    }
    if *pNext == '(' {
        pStop = Scl_LibertyFindMatch( pNext, pEnd );
        Head.Beg = pNext - p.pContents + 1;
        Head.End = pStop - p.pContents;
        *ppPos = pStop + 1;
        if Scl_LibertySkipSpaces( p, ppPos, pEnd, 0 ) != 0 {
            // end of list
            pItem = Scl_LibertyNewItem( p, SCL_LIBERTY_LIST );
            pItem.Key  = Key;
            pItem.Head = Scl_LibertyUpdateHead( p, Head );
            return Scl_LibertyItemId( p, pItem );
        }
        pNext = *ppPos;
        if *pNext == '{' { // beginning of body
        
            pStop = Scl_LibertyFindMatch( pNext, pEnd );
            Body.Beg = pNext - p.pContents + 1;
            Body.End = pStop - p.pContents;
            // end of body
            pItem = Scl_LibertyNewItem( p, SCL_LIBERTY_PROC );
            pItem.Key  = Key;
            pItem.Head = Scl_LibertyUpdateHead( p, Head );
            pItem.Body = Body;
            *ppPos = pNext + 1;
            pItem.Child = Scl_LibertyBuildItem( p, ppPos, pStop );
            if pItem.Child == -1 {
                goto exit; }
            *ppPos = pStop + 1;
            pItem.Next = Scl_LibertyBuildItem( p, ppPos, pEnd );
            if pItem.Next == -1 {
                goto exit; }
            return Scl_LibertyItemId( p, pItem );
        }
        // end of list
        if *pNext == ';' {
            *ppPos = pNext + 1; }
        pItem = Scl_LibertyNewItem( p, SCL_LIBERTY_LIST );
        pItem.Key  = Key;
        pItem.Head = Scl_LibertyUpdateHead( p, Head );
        pItem.Next = Scl_LibertyBuildItem( p, ppPos, pEnd );
        if pItem.Next == -1 {
            goto exit; }
        return Scl_LibertyItemId( p, pItem );
    }
    if *pNext == ';' {
        *ppPos = pNext + 1;
        return Scl_LibertyBuildItem(p, ppPos, pEnd);
    }
exit:
    if p.pError == nil {
        // inlining of macro ABC_ALLOC
        p.pError = ((char *) malloc(sizeof(char) * (size_t)(1000)));
        sprintf( p.pError, "File \"%s\". Line %6d. Failed to parse entry \"%s\".\n", 
                    p.pFileName, p.nLines, Scl_LibertyReadString(p, Key) );
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [File management.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibertyFixFileName(pFileName string) {
    var pHead string;
    for  pHead = pFileName; *pHead; pHead++  {
        if *pHead == '>' {
            *pHead = '\\'; }
    }
}
func Scl_LibertyFileSize(pFileName string) int32 {
    var pFile *os.File;
    var nFileSize int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Scl_LibertyFileSize(): The input file is unavailable (absent or open).\n");
        return 0;
    }
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile ); 
    _ = (pFile).Close();
    return nFileSize;
}
func Scl_LibertyFileContents(pFileName string, nContents int32) string {
    var pFile *os.File  = os.Open(pFileName);
    // inlining of macro ABC_ALLOC
    var pContents string  = ((char *) malloc(sizeof(char) * (size_t)(nContents+1)));
    var RetValue int32;
    RetValue = fread( pContents, nContents, 1, pFile );
    _ = (pFile).Close();
    pContents[nContents] = 0;
    return pContents;
}
func Scl_LibertyStringDump(pFileName string, vStr *Vec_Str_t) {
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    var RetValue int32;
    if pFile == nil {
        fmt.Printf("Scl_LibertyStringDump(): The output file is unavailable.\n");
        return;
    }
    RetValue = fwrite( Vec_StrArray(vStr), 1, Vec_StrSize(vStr), pFile );
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Starts the parsing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibertyStart(pFileName string) *Scl_Tree_t {
    var p *Scl_Tree_t;
    var RetValue int32;
    // read the file into the buffer
    Scl_LibertyFixFileName( pFileName );
    RetValue = Scl_LibertyFileSize( pFileName );
    if RetValue == 0 {
        return nil; }
    // start the manager
    // inlining of macro ABC_ALLOC
    p = ((Scl_Tree_t *) malloc(sizeof(Scl_Tree_t) * (size_t)(1)));
    *p = Scl_Tree_t{};
    p.clkStart  = Abc_Clock();
    p.nContents = RetValue;
    p.pContents = Scl_LibertyFileContents( pFileName, p.nContents );
    // other 
    p.pFileName = Abc_UtilStrsav( pFileName );
    p.nItermAlloc = 10 + Scl_LibertyCountItems( p.pContents, p.pContents+p.nContents );
    // inlining of macro ABC_CALLOC
    p.pItems = ((Scl_Item_t *) calloc((size_t)(p.nItermAlloc), sizeof(Scl_Item_t)));
    p.nItems = 0;
    p.nLines = 1;
    p.vBuffer = Vec_StrStart( 10 );
    return p;
}
func Scl_LibertyStop(p *Scl_Tree_t, fVerbose int32) {
    if fVerbose != 0 {
        fmt.Printf("Memory = %7.2f MB. ", 1.0 * (p.nContents + p.nItermAlloc * sizeof(Scl_Item_t))/(1<<20));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - p.clkStart)))/(float64(int64(time.Second))));;
    }
    Vec_StrFree( p.vBuffer );
    // inlining of macro ABC_FREE
    ((p.pFileName) ? (free((char *) (p.pFileName)), (p.pFileName) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pContents) ? (free((char *) (p.pContents)), (p.pContents) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pItems) ? (free((char *) (p.pItems)), (p.pItems) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pError) ? (free((char *) (p.pError)), (p.pError) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Scl_LibertyParse(pFileName string, fVerbose int32) *Scl_Tree_t {
    var p *Scl_Tree_t;
    var pPos string;
    if (p = Scl_LibertyStart(pFileName)) == nil {
        return nil; }
    pPos = p.pContents;
    Scl_LibertyWipeOutComments( p.pContents, p.pContents+p.nContents );
    if (!Scl_LibertyBuildItem( p, &pPos, p.pContents + p.nContents )) == 0 {
        if p.pError { fmt.Printf("%s", p.pError); }
        fmt.Printf("Parsing failed.  ");
        Abc_PrintTime( 1, "Parsing time", Abc_Clock() - p.clkStart );
    
    } else if fVerbose != 0 {
        fmt.Printf("Parsing finished successfully.  ");
        Abc_PrintTime( 1, "Parsing time", Abc_Clock() - p.clkStart );
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Fetching attributes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibertyReadCellIsFlop(p *Scl_Tree_t, pCell *Scl_Item_t) int32 {
    var pAttr *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChild
    for  pAttr = Scl_LibertyItem(p, pCell.Child); (pAttr != nil); pAttr = Scl_LibertyItem(p, pAttr.Next)  {
        if !(Scl_LibertyCompare(p, pAttr.Key, "ff") != 0) ||
             !(Scl_LibertyCompare(p, pAttr.Key, "latch") != 0) {
            return 1; }
    }
    return 0;
}
func Scl_LibertyReadCellIsDontUse(p *Scl_Tree_t, pCell *Scl_Item_t) int32 {
    var pAttr *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChild
    for  pAttr = Scl_LibertyItem(p, pCell.Child); (pAttr != nil); pAttr = Scl_LibertyItem(p, pAttr.Next)  {
        if !(Scl_LibertyCompare(p, pAttr.Key, "dont_use") != 0) {
            return 1; }
    }
    return 0;
}
func Scl_LibertyReadCellArea(p *Scl_Tree_t, pCell *Scl_Item_t) string {
    var pArea *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pArea = Scl_LibertyItem(p, pCell.Child); (pArea != nil); pArea = Scl_LibertyItem(p, pArea.Next)  { if Scl_LibertyCompare(p, pArea.Key, "area") != 0 {} else {
        return Scl_LibertyReadString(p, pArea.Head); }
    }
    return 0;
}
func Scl_LibertyReadCellLeakage(p *Scl_Tree_t, pCell *Scl_Item_t) string {
    var pItem, pChild *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, pCell.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "cell_leakage_power") != 0 {} else {
        return Scl_LibertyReadString(p, pItem.Head); }
    }
    // look for another type
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, pCell.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "leakage_power") != 0 {} else {
        // inlining of macro Scl_ItemForEachChildName
        for  pChild = Scl_LibertyItem(p, pItem.Child); (pChild != nil); pChild = Scl_LibertyItem(p, pChild.Next)  { if Scl_LibertyCompare(p, pChild.Key, "when") != 0 {} else {
            break; }
        }
        if (pChild != nil) && !(Scl_LibertyCompare(p, pChild.Key, "when") != 0) {
            continue; }
        // inlining of macro Scl_ItemForEachChildName
        for  pChild = Scl_LibertyItem(p, pItem.Child); (pChild != nil); pChild = Scl_LibertyItem(p, pChild.Next)  { if Scl_LibertyCompare(p, pChild.Key, "value") != 0 {} else {
            return Scl_LibertyReadString(p, pChild.Head); }
        }
    }
    }
    return 0;
}
func Scl_LibertyReadPinFormula(p *Scl_Tree_t, pPin *Scl_Item_t) string {
    var pFunc *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pFunc = Scl_LibertyItem(p, pPin.Child); (pFunc != nil); pFunc = Scl_LibertyItem(p, pFunc.Next)  { if Scl_LibertyCompare(p, pFunc.Key, "function") != 0 {} else {
        return Scl_LibertyReadString(p, pFunc.Head); }
    }
    return nil;
}
func Scl_LibertyReadCellIsThreeState(p *Scl_Tree_t, pCell *Scl_Item_t) int32 {
    var pPin, pItem *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pPin = Scl_LibertyItem(p, pCell.Child); (pPin != nil); pPin = Scl_LibertyItem(p, pPin.Next)  { if Scl_LibertyCompare(p, pPin.Key, "pin") != 0 {} else {
        // inlining of macro Scl_ItemForEachChildName
        for  pItem = Scl_LibertyItem(p, pPin.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "three_state") != 0 {} else {
            return 1; }
        } }
    }
    return 0;
}
func Scl_LibertyReadCellOutputNum(p *Scl_Tree_t, pCell *Scl_Item_t) int32 {
    var pPin *Scl_Item_t;
    var Counter int32;
    // inlining of macro Scl_ItemForEachChildName
    for  pPin = Scl_LibertyItem(p, pCell.Child); (pPin != nil); pPin = Scl_LibertyItem(p, pPin.Next)  { if Scl_LibertyCompare(p, pPin.Key, "pin") != 0 {} else
        if Scl_LibertyReadPinFormula(p, pPin) {
            Counter++; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Parses the standard cell library in Liberty format.]

  Description [Writes the resulting file in Genlib format.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibertyReadGenlibStr(p *Scl_Tree_t, fVerbose int32) *Vec_Str_t {
    var vStr *Vec_Str_t;
    var pCell, pOutput, pInput *Scl_Item_t;
    var pFormula string;
    vStr = Vec_StrAlloc( 1000 );
    Vec_StrPrintStr( vStr, "GATE          _const0_  0.000000  z=CONST0;\n" );
    Vec_StrPrintStr( vStr, "GATE          _const1_  0.000000  z=CONST1;\n" );
    // inlining of macro Scl_ItemForEachChildName
    for  pCell = Scl_LibertyItem(p, Scl_LibertyRoot(p).Child); (pCell != nil); pCell = Scl_LibertyItem(p, pCell.Next)  { if Scl_LibertyCompare(p, pCell.Key, "cell") != 0 {} else {
        if Scl_LibertyReadCellIsFlop(p, pCell) != 0 {
            if fVerbose != 0 {  fmt.Printf("Scl_LibertyReadGenlib() skipped sequential cell \"%s\".\n", Scl_LibertyReadString(p, pCell.Head)); }
            continue;
        }
        if Scl_LibertyReadCellIsDontUse(p, pCell) != 0 {
            if fVerbose != 0 {  fmt.Printf("Scl_LibertyReadGenlib() skipped cell \"%s\" due to dont_use attribute.\n", Scl_LibertyReadString(p, pCell.Head)); }
            continue;
        }
        if Scl_LibertyReadCellIsThreeState(p, pCell) != 0 {
            if fVerbose != 0 {  fmt.Printf("Scl_LibertyReadGenlib() skipped three-state cell \"%s\".\n", Scl_LibertyReadString(p, pCell.Head)); }
            continue;
        }
        if Scl_LibertyReadCellOutputNum(p, pCell) == 0 {
            if fVerbose != 0 {  fmt.Printf("Scl_LibertyReadGenlib() skipped cell \"%s\" without logic function.\n", Scl_LibertyReadString(p, pCell.Head)); }
            continue;
        }
        // iterate through output pins
        // inlining of macro Scl_ItemForEachChildName
        for  pOutput = Scl_LibertyItem(p, pCell.Child); pOutput; pOutput = Scl_LibertyItem(p, pOutput.Next)  { if Scl_LibertyCompare(p, pOutput.Key, "pin") != 0 {} else {
            if !(pFormula = Scl_LibertyReadPinFormula(p, pOutput)) { 
                continue; }
            if !(int32(strings.Compare(pFormula, "0")) != 0) || !(int32(strings.Compare(pFormula, "1")) != 0) {
                if fVerbose != 0 { fmt.Printf("Scl_LibertyReadGenlib() skipped cell \"%s\" with constant formula \"%s\".\n", Scl_LibertyReadString(p, pCell.Head), pFormula); }
                break;
            }
            Vec_StrPrintStr( vStr, "GATE " );
            Vec_StrPrintStr( vStr, Scl_LibertyReadString(p, pCell.Head) );
            Vec_StrPrintStr( vStr, " " );
            Vec_StrPrintStr( vStr, Scl_LibertyReadCellArea(p, pCell) );
            Vec_StrPrintStr( vStr, " " );
            Vec_StrPrintStr( vStr, Scl_LibertyReadString(p, pOutput.Head) );
            Vec_StrPrintStr( vStr, "=" );
            Vec_StrPrintStr( vStr, pFormula );
            Vec_StrPrintStr( vStr, ";\n" );
            // iterate through input pins
            // inlining of macro Scl_ItemForEachChildName
            for  pInput = Scl_LibertyItem(p, pCell.Child); (pInput != nil); pInput = Scl_LibertyItem(p, pInput.Next)  { if Scl_LibertyCompare(p, pInput.Key, "pin") != 0 {} else {
                if Scl_LibertyReadPinFormula(p, pInput) == nil {
                    continue; }
                Vec_StrPrintStr( vStr, "  PIN " );
                Vec_StrPrintStr( vStr, Scl_LibertyReadString(p, pInput.Head) );
                Vec_StrPrintStr( vStr, " UNKNOWN  1  999  1.00  0.00  1.00  0.00\n" );
            }
            }
        }
        }
    }
    }
    Vec_StrPrintStr( vStr, "\n.end\n" );
    Vec_StrPush( vStr, '\0' );
//    printf( "%s", Vec_StrArray(vStr) );
    return vStr;
}
func Scl_LibertyParseGenlibStr(pFileName string, fVerbose int32) *Vec_Str_t {
    var p *Scl_Tree_t;
    var vStr *Vec_Str_t;
    p = Scl_LibertyParse( pFileName, fVerbose );
    if p == nil {
        return nil; }
//    Scl_LibertyRead( p, "temp_.lib" );
    vStr = Scl_LibertyReadGenlibStr( p, fVerbose );
    Scl_LibertyStop( p, fVerbose );
//    Scl_LibertyStringDump( "test_genlib.lib", vStr );
    return vStr;
}


/**Function*************************************************************

  Synopsis    [Enabling debug output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
//#define SCL_DEBUG
#ifdef SCL_DEBUG
static inline void Vec_StrPutI_( Vec_Str_t * vOut, int Val )     {  printf( "%d ",  Val );        Vec_StrPutI( vOut, Val );  }
static inline void Vec_StrPutW_( Vec_Str_t * vOut, word Val )    {  printf( "%lu ", (long)Val );  Vec_StrPutW( vOut, Val );  }
static inline void Vec_StrPutF_( Vec_Str_t * vOut, float Val )   {  printf( "%f ",  Val );        Vec_StrPutF( vOut, Val );  }
static inline void Vec_StrPutS_( Vec_Str_t * vOut, char * Val )  {  printf( "%s ",  Val );        Vec_StrPutS( vOut, Val );  }
static inline void Vec_StrPut_( Vec_Str_t * vOut )               {  printf( "\n" ); }
#else
static inline void Vec_StrPutI_( Vec_Str_t * vOut, int Val )     { Vec_StrPutI( vOut, Val );  }
static inline void Vec_StrPutW_( Vec_Str_t * vOut, word Val )    { Vec_StrPutW( vOut, Val );  }
static inline void Vec_StrPutF_( Vec_Str_t * vOut, float Val )   { Vec_StrPutF( vOut, Val );  }
static inline void Vec_StrPutS_( Vec_Str_t * vOut, char * Val )  { Vec_StrPutS( vOut, Val );  }
static inline void Vec_StrPut_( Vec_Str_t * vOut )               { }
#endif

/**Function*************************************************************

  Synopsis    [Parsing Liberty into internal data representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibertyReadDefaultWireLoad(p *Scl_Tree_t) string {
    var pItem *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, Scl_LibertyRoot(p).Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "default_wire_load") != 0 {} else {
        return Scl_LibertyReadString(p, pItem.Head); }
    }
    return "";
}
func Scl_LibertyReadDefaultWireLoadSel(p *Scl_Tree_t) string {
    var pItem *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, Scl_LibertyRoot(p).Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "default_wire_load_selection") != 0 {} else {
        return Scl_LibertyReadString(p, pItem.Head); }
    }
    return "";
}
func Scl_LibertyReadDefaultMaxTrans(p *Scl_Tree_t) float32 {
    var pItem *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, Scl_LibertyRoot(p).Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "default_max_transition") != 0 {} else {
        return atof(Scl_LibertyReadString(p, pItem.Head)); }
    }
    return 0;
}
func Scl_LibertyReadTimeUnit(p *Scl_Tree_t) int32 {
    var pItem *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, Scl_LibertyRoot(p).Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "time_unit") != 0 {} else {
        var pUnit string  = Scl_LibertyReadString(p, pItem.Head);
        // 9=1ns, 10=100ps, 11=10ps, 12=1ps
        if !(int32(strings.Compare(pUnit, "1ns")) != 0) {
            return 9; }
        if !(int32(strings.Compare(pUnit, "100ps")) != 0) {
            return 10; }
        if !(int32(strings.Compare(pUnit, "10ps")) != 0) {
            return 11; }
        if !(int32(strings.Compare(pUnit, "1ps")) != 0) {
            return 12; }
        break;
    }
    }
    fmt.Printf("Libery parser cannot read \"time_unit\".  Assuming   time_unit : \"1ns\".\n");
    return 9;
}
func Scl_LibertyReadLoadUnit(p *Scl_Tree_t, vOut *Vec_Str_t) {
    var pItem *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, Scl_LibertyRoot(p).Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "capacitive_load_unit") != 0 {} else {
        // expecting (1.00,ff) or (1, pf) ... 12 or 15 for 'pf' or 'ff'
        var pHead string    = Scl_LibertyReadString(p, pItem.Head);
        var First float32     = atof(strtok(pHead, " \t\n\r\\\","));
        var pSecond string  = strtok(nil, " \t\n\r\\\",");
        Vec_StrPutF_( vOut, First );
        if pSecond && !(int32(strings.Compare(pSecond, "pf")) != 0) {
            Vec_StrPutI_( vOut, 12 );
        } else if pSecond && !(int32(strings.Compare(pSecond, "ff")) != 0) {
            Vec_StrPutI_( vOut, 15 );
        } else { break; }
        return;
    }
    }
    fmt.Printf("Libery parser cannot read \"capacitive_load_unit\". Assuming   capacitive_load_unit(1, pf).\n");
    Vec_StrPutF_( vOut, 1.0 );
    Vec_StrPutI_( vOut, 12 );
}
func Scl_LibertyReadWireLoad(p *Scl_Tree_t, vOut *Vec_Str_t) {
    var pItem, pChild *Scl_Item_t;
    Vec_StrPutI_( vOut, Scl_LibertyItemNum(p, Scl_LibertyRoot(p), "wire_load") );
    Vec_StrPut_( vOut );
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, Scl_LibertyRoot(p).Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "wire_load") != 0 {} else {
        Vec_StrPutS_( vOut, Scl_LibertyReadString(p, pItem.Head) );
        // inlining of macro Scl_ItemForEachChildName
        for  pChild = Scl_LibertyItem(p, pItem.Child); (pChild != nil); pChild = Scl_LibertyItem(p, pChild.Next)  { if Scl_LibertyCompare(p, pChild.Key, "capacitance") != 0 {} else {
            Vec_StrPutF_( vOut, atof(Scl_LibertyReadString(p, pChild.Head)) ); }
        }
        // inlining of macro Scl_ItemForEachChildName
        for  pChild = Scl_LibertyItem(p, pItem.Child); (pChild != nil); pChild = Scl_LibertyItem(p, pChild.Next)  { if Scl_LibertyCompare(p, pChild.Key, "slope") != 0 {} else {
            Vec_StrPutF_( vOut, atof(Scl_LibertyReadString(p, pChild.Head)) ); }
        }
        Vec_StrPut_( vOut );
        Vec_StrPutI_( vOut, Scl_LibertyItemNum(p, pItem, "fanout_length") );
        Vec_StrPut_( vOut );
        // inlining of macro Scl_ItemForEachChildName
        for  pChild = Scl_LibertyItem(p, pItem.Child); (pChild != nil); pChild = Scl_LibertyItem(p, pChild.Next)  { if Scl_LibertyCompare(p, pChild.Key, "fanout_length") != 0 {} else {
            var pHead string   = Scl_LibertyReadString(p, pChild.Head);
            var First int32   = strconv.Atoi(strtok(pHead, " ,"));
            var Second float32  = atof( strtok(nil, " ") );
            Vec_StrPutI_( vOut, First );
            Vec_StrPutF_( vOut, Second );
            Vec_StrPut_( vOut );
        }
        }
        Vec_StrPut_( vOut );
    }
    }
    Vec_StrPut_( vOut );
}
func Scl_LibertyReadWireLoadSelect(p *Scl_Tree_t, vOut *Vec_Str_t) {
    var pItem, pChild *Scl_Item_t;
    Vec_StrPutI_( vOut, Scl_LibertyItemNum(p, Scl_LibertyRoot(p), "wire_load_selection") );
    Vec_StrPut_( vOut );
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, Scl_LibertyRoot(p).Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "wire_load_selection") != 0 {} else {
        Vec_StrPutS_( vOut, Scl_LibertyReadString(p, pItem.Head) );
        Vec_StrPut_( vOut );
        Vec_StrPutI_( vOut, Scl_LibertyItemNum(p, pItem, "wire_load_from_area") );
        Vec_StrPut_( vOut );
        // inlining of macro Scl_ItemForEachChildName
        for  pChild = Scl_LibertyItem(p, pItem.Child); (pChild != nil); pChild = Scl_LibertyItem(p, pChild.Next)  { if Scl_LibertyCompare(p, pChild.Key, "wire_load_from_area") != 0 {} else {
            var pHead string   = Scl_LibertyReadString(p, pChild.Head);
            var First float32   = atof( strtok(pHead, " ,") );
            var Second float32  = atof( strtok(nil, " ,") );
            var pThird string  = strtok(nil, " ");
            if pThird[0] == '\"' {
                assert(pThird[len(pThird)-1] == '\"', "pThird[strlen(pThird)-1] == '\\\"'");
pThird[len(pThird)-1] = 0;
pThird++; }
            Vec_StrPutF_( vOut, First );
            Vec_StrPutF_( vOut, Second );
            Vec_StrPutS_( vOut, pThird );
            Vec_StrPut_( vOut );
        }
        }
        Vec_StrPut_( vOut );
    }
    }
    Vec_StrPut_( vOut );
}
func Scl_LibertyReadDeriveStrength(p *Scl_Tree_t, pCell *Scl_Item_t) int32 {
    var pItem *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, pCell.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "drive_strength") != 0 {} else {
        return strconv.Atoi(Scl_LibertyReadString(p, pItem.Head)); }
    }
    return 0;
}
func Scl_LibertyReadPinDirection(p *Scl_Tree_t, pPin *Scl_Item_t) int32 {
    var pItem *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, pPin.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "direction") != 0 {} else {
        var pToken string  = Scl_LibertyReadString(p, pItem.Head);
        if !(int32(strings.Compare(pToken, "input")) != 0) {
            return 0; }
        if !(int32(strings.Compare(pToken, "output")) != 0) {
            return 1; }
        break;
    }
    }
    return -1;
}
func Scl_LibertyReadPinCap(p *Scl_Tree_t, pPin *Scl_Item_t, pName string) float32 {
    var pItem *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, pPin.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, pName) != 0 {} else {
        return atof(Scl_LibertyReadString(p, pItem.Head)); }
    }
    return 0;
}
func Scl_LibertyReadPinTiming(p *Scl_Tree_t, pPinOut *Scl_Item_t, pNameIn string) *Scl_Item_t {
    var pTiming, pPinIn *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pTiming = Scl_LibertyItem(p, pPinOut.Child); (pTiming != nil); pTiming = Scl_LibertyItem(p, pTiming.Next)  { if Scl_LibertyCompare(p, pTiming.Key, "timing") != 0 {} else {
        // inlining of macro Scl_ItemForEachChildName
        for  pPinIn = Scl_LibertyItem(p, pTiming.Child); (pPinIn != nil); pPinIn = Scl_LibertyItem(p, pPinIn.Next)  { if Scl_LibertyCompare(p, pPinIn.Key, "related_pin") != 0 {} else
            if !(int32(strings.Compare(Scl_LibertyReadString(p, pPinIn.Head), pNameIn)) != 0) {
                return pTiming; }
        } }
    }
    return nil;
}
func Scl_LibertyReadPinTimingAll(p *Scl_Tree_t, pPinOut *Scl_Item_t, pNameIn string) *Vec_Ptr_t {
    var vTimings *Vec_Ptr_t;
    var pTiming, pPinIn *Scl_Item_t;
    vTimings = Vec_PtrAlloc( 16 );
    // inlining of macro Scl_ItemForEachChildName
    for  pTiming = Scl_LibertyItem(p, pPinOut.Child); (pTiming != nil); pTiming = Scl_LibertyItem(p, pTiming.Next)  { if Scl_LibertyCompare(p, pTiming.Key, "timing") != 0 {} else {
        // inlining of macro Scl_ItemForEachChildName
        for  pPinIn = Scl_LibertyItem(p, pTiming.Child); (pPinIn != nil); pPinIn = Scl_LibertyItem(p, pPinIn.Next)  { if Scl_LibertyCompare(p, pPinIn.Key, "related_pin") != 0 {} else
            if !(int32(strings.Compare(Scl_LibertyReadString(p, pPinIn.Head), pNameIn)) != 0) {
                Vec_PtrPush( vTimings, pTiming ); }
        } }
    }
    return vTimings;
}
func Scl_LibertyReadTimingSense(p *Scl_Tree_t, pPin *Scl_Item_t) int32 {
    var pItem *Scl_Item_t;
    // inlining of macro Scl_ItemForEachChildName
    for  pItem = Scl_LibertyItem(p, pPin.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  { if Scl_LibertyCompare(p, pItem.Key, "timing_sense") != 0 {} else {
        var pToken string  = Scl_LibertyReadString(p, pItem.Head);
        if !(int32(strings.Compare(pToken, "positive_unate")) != 0) {
            return sc_ts_Pos; }
        if !(int32(strings.Compare(pToken, "negative_unate")) != 0) {
            return sc_ts_Neg; }
        if !(int32(strings.Compare(pToken, "non_unate")) != 0) {
            return sc_ts_Non; }
        break;
    }
    }
    return sc_ts_Non;
}
func Scl_LibertyReadFloatVec(pName string) *Vec_Flt_t {
    var pToken string;
    var vValues *Vec_Flt_t  = Vec_FltAlloc( 100 );
    for  pToken = strtok(pName, " \t\n\r\\\","); pToken; pToken = strtok(nil, " \t\n\r\\\",")  {
        Vec_FltPush( vValues, atof(pToken) );
    }
    return vValues;
}

func Scl_LibertyDumpTables(vOut *Vec_Str_t, vInd1 *Vec_Flt_t, vInd2 *Vec_Flt_t, vValues *Vec_Flt_t) {
    var i int32; var Entry float32;
    // write entries
    Vec_StrPutI_( vOut, Vec_FltSize(vInd1) );
    // inlining of macro Vec_FltForEachEntry
    for  i = 0; (i < Vec_FltSize(vInd1)) && (((Entry) = Vec_FltEntry(vInd1, i)), 1); i++  {
        Vec_StrPutF_( vOut, Entry );
    }
    Vec_StrPut_( vOut );
    // write entries
    Vec_StrPutI_( vOut, Vec_FltSize(vInd2) );
    // inlining of macro Vec_FltForEachEntry
    for  i = 0; (i < Vec_FltSize(vInd2)) && (((Entry) = Vec_FltEntry(vInd2, i)), 1); i++  {
        Vec_StrPutF_( vOut, Entry );
    }
    Vec_StrPut_( vOut );
    Vec_StrPut_( vOut );
    // write entries
    assert(Vec_FltSize(vInd1) * Vec_FltSize(vInd2) == Vec_FltSize(vValues), "Vec_FltSize(vInd1) * Vec_FltSize(vInd2) == Vec_FltSize(vValues)");
    // inlining of macro Vec_FltForEachEntry
    for  i = 0; (i < Vec_FltSize(vValues)) && (((Entry) = Vec_FltEntry(vValues, i)), 1); i++  {
        Vec_StrPutF_( vOut, Entry );
        if i % Vec_FltSize(vInd2) == Vec_FltSize(vInd2)-1 {
            Vec_StrPut_( vOut ); }
    }
    // dump approximations
    Vec_StrPut_( vOut );
    for  i = 0; i < 3; i++  { 
        Vec_StrPutF_( vOut, 0 );
    }
    for  i = 0; i < 4; i++  { 
        Vec_StrPutF_( vOut, 0 );
    }
    for  i = 0; i < 6; i++  { 
        Vec_StrPutF_( vOut, 0 );
    }
    Vec_StrPut_( vOut );
    Vec_StrPut_( vOut );
}
func Scl_LibertyScanTable(p *Scl_Tree_t, vOut *Vec_Ptr_t, pTiming *Scl_Item_t, pName string, vTemples *Vec_Ptr_t) int32 {
    var vIndex1 *Vec_Flt_t  = nil;
    var vIndex2 *Vec_Flt_t  = nil;
    var vValues *Vec_Flt_t  = nil;
    var vInd1, vInd2 *Vec_Flt_t;
    var pItem *Scl_Item_t;
    var pTable *Scl_Item_t = nil;
    var pThis string;
    var pTempl string = nil;
    var iPlace, i int32;
    var Entry float32;
    // find the table
    // inlining of macro Scl_ItemForEachChildName
    for  pTable = Scl_LibertyItem(p, pTiming.Child); (pTable != nil); pTable = Scl_LibertyItem(p, pTable.Next)  { if Scl_LibertyCompare(p, pTable.Key, pName) != 0 {} else {
        break; }
    }
    if pTable == nil {
        return 0; }
    // find the template
    pTempl = Scl_LibertyReadString(p, pTable.Head);
    if pTempl == nil || pTempl[0] == 0 {
        // read the numbers
        // inlining of macro Scl_ItemForEachChild
        for  pItem = Scl_LibertyItem(p, pTable.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  {
            if !(Scl_LibertyCompare(p, pItem.Key, "index_1") != 0) {
                assert(vIndex1 == nil, "vIndex1 == nil");
vIndex1 = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) );
            } else if !(Scl_LibertyCompare(p, pItem.Key, "index_2") != 0) {
                assert(vIndex2 == nil, "vIndex2 == nil");
vIndex2 = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) );
            } else if !(Scl_LibertyCompare(p, pItem.Key, "values") != 0) {
                assert(vValues == nil, "vValues == nil");
vValues = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) ); }
        }
        if vIndex1 == nil || vIndex2 == nil || vValues == nil { fmt.Printf("Incomplete table specification\n"); return 0; }
        // dump the table
        vInd1 = vIndex1;
        vInd2 = vIndex2;
        // write entries
        Vec_PtrPush( vOut, vInd1 );
        Vec_PtrPush( vOut, vInd2 );
        Vec_PtrPush( vOut, vValues );
    
    } else if !(int32(strings.Compare(pTempl, "scalar")) != 0) {
        // inlining of macro Scl_ItemForEachChild
        for  pItem = Scl_LibertyItem(p, pTable.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  {
            if !(Scl_LibertyCompare(p, pItem.Key, "values") != 0) {
                assert(vValues == nil, "vValues == nil");
                vValues = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) );
                assert(Vec_FltSize(vValues) == 1, "Vec_FltSize(vValues) == 1");
                // write entries
                Vec_PtrPush( vOut, Vec_IntStart(1) );
                Vec_PtrPush( vOut, Vec_IntStart(1) );
                Vec_PtrPush( vOut, vValues );
                break;
            
            } else { fmt.Printf("Cannot read \"scalar\" template\n"); return 0; }
        }
    
    } else {
        // fetch the template
        iPlace = -1;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vTemples)) && (((pThis) = (char *)Vec_PtrEntry(vTemples, i)), 1); i++  {
            if i % 4 == 0 && !(int32(strings.Compare(pTempl, pThis)) != 0) {  
                iPlace = i;
                break;
            }
        }
        if iPlace == -1 { fmt.Printf("Template cannot be found in the template library\n"); return 0; }
        // read the numbers
        // inlining of macro Scl_ItemForEachChild
        for  pItem = Scl_LibertyItem(p, pTable.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  {
            if !(Scl_LibertyCompare(p, pItem.Key, "index_1") != 0) {
                assert(vIndex1 == nil, "vIndex1 == nil");
vIndex1 = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) );
            } else if !(Scl_LibertyCompare(p, pItem.Key, "index_2") != 0) {
                assert(vIndex2 == nil, "vIndex2 == nil");
vIndex2 = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) );
            } else if !(Scl_LibertyCompare(p, pItem.Key, "values") != 0) {
                assert(vValues == nil, "vValues == nil");
vValues = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) ); }
        }
        // check the template style
        vInd1 = (Vec_Flt_t *)Vec_PtrEntry( vTemples, iPlace + 2 ); // slew
        vInd2 = (Vec_Flt_t *)Vec_PtrEntry( vTemples, iPlace + 3 ); // load
        if Vec_PtrEntry(vTemples, iPlace + 1) == nil { // normal order (vIndex1 is slew; vIndex2 is load)
        
            assert(!(vIndex1 != nil) || Vec_FltSize(vIndex1) == Vec_FltSize(vInd1), "!vIndex1 || Vec_FltSize(vIndex1) == Vec_FltSize(vInd1)");
            assert(!(vIndex2 != nil) || Vec_FltSize(vIndex2) == Vec_FltSize(vInd2), "!vIndex2 || Vec_FltSize(vIndex2) == Vec_FltSize(vInd2)");
            if vIndex1 != nil {
                vInd1 = vIndex1;
            } else {
                vInd1 = vInd1;
            }
            if vIndex2 != nil {
                vInd2 = vIndex2;
            } else {
                vInd2 = vInd2;
            }
            // write entries
            Vec_PtrPush( vOut, Vec_FltDup(vInd1) );
            Vec_PtrPush( vOut, Vec_FltDup(vInd2) );
            Vec_PtrPush( vOut, Vec_FltDup(vValues) );
        
        } else {  // reverse order (vIndex2 is slew; vIndex1 is load)
        
            var vValues2 *Vec_Flt_t  = Vec_FltAlloc( Vec_FltSize(vValues) );
            assert(!(vIndex2 != nil) || Vec_FltSize(vIndex2) == Vec_FltSize(vInd1), "!vIndex2 || Vec_FltSize(vIndex2) == Vec_FltSize(vInd1)");
            assert(!(vIndex1 != nil) || Vec_FltSize(vIndex1) == Vec_FltSize(vInd2), "!vIndex1 || Vec_FltSize(vIndex1) == Vec_FltSize(vInd2)");
            if vIndex2 != nil {
                vInd1 = vIndex2;
            } else {
                vInd1 = vInd1;
            }
            if vIndex1 != nil {
                vInd2 = vIndex1;
            } else {
                vInd2 = vInd2;
            }
            // write entries -- transpose
            assert(Vec_FltSize(vInd1) * Vec_FltSize(vInd2) == Vec_FltSize(vValues), "Vec_FltSize(vInd1) * Vec_FltSize(vInd2) == Vec_FltSize(vValues)");
            // inlining of macro Vec_FltForEachEntry
            for  i = 0; (i < Vec_FltSize(vValues)) && (((Entry) = Vec_FltEntry(vValues, i)), 1); i++  {
                var x int32  = i % Vec_FltSize(vInd2);
                var y int32  = i / Vec_FltSize(vInd2);
                Entry = Vec_FltEntry( vValues, x * Vec_FltSize(vInd1) + y );
                Vec_FltPush( vValues2, Entry );
            }
            assert(Vec_FltSize(vValues) == Vec_FltSize(vValues2), "Vec_FltSize(vValues) == Vec_FltSize(vValues2)");
            // write entries
            Vec_PtrPush( vOut, Vec_FltDup(vInd1) );
            Vec_PtrPush( vOut, Vec_FltDup(vInd2) );
            Vec_PtrPush( vOut, vValues2 );
        }
        Vec_FltFreeP( &vIndex1 );
        Vec_FltFreeP( &vIndex2 );
        Vec_FltFreeP( &vValues );
    }
    return 1;
}
func Scl_LibertyComputeWorstCase(vTables *Vec_Ptr_t, pvInd0 **Vec_Flt_t, pvInd1 **Vec_Flt_t, pvValues **Vec_Flt_t) int32 {
    var vInd0, vInd1, vValues *Vec_Flt_t;
    var vind0, vind1, vvalues *Vec_Flt_t;
    var i int32;
    var k int32;
    var nTriples int32 = Vec_PtrSize(vTables) / 3;
    var Entry float32;
    assert(Vec_PtrSize(vTables) > 0 && Vec_PtrSize(vTables) % 3 == 0, "Vec_PtrSize(vTables) > 0 && Vec_PtrSize(vTables) % 3 == 0");
    if nTriples == 1 {
        *pvInd0   = (Vec_Flt_t *)Vec_PtrEntry(vTables, 0);
        *pvInd1   = (Vec_Flt_t *)Vec_PtrEntry(vTables, 1);
        *pvValues = (Vec_Flt_t *)Vec_PtrEntry(vTables, 2);
        Vec_PtrShrink( vTables, 0 );
        return 1;
    }
    vInd0   = Vec_FltDup( (Vec_Flt_t *)Vec_PtrEntry(vTables, 0) );
    vInd1   = Vec_FltDup( (Vec_Flt_t *)Vec_PtrEntry(vTables, 1) );
    vValues = Vec_FltDup( (Vec_Flt_t *)Vec_PtrEntry(vTables, 2) );
    for  i = 1; i < nTriples; i++  {
        vind0   = (Vec_Flt_t *)Vec_PtrEntry(vTables, i*3+0);
        vind1   = (Vec_Flt_t *)Vec_PtrEntry(vTables, i*3+1);
        vvalues = (Vec_Flt_t *)Vec_PtrEntry(vTables, i*3+2);
        // check equality of indexes
        if !Vec_FltEqual(vind0, vInd0) {
            continue; }//return 0;
        if !Vec_FltEqual(vind1, vInd1) {
            continue; }//return 0;
//        Vec_FltForEachEntry( vvalues, Entry, k )
//            Vec_FltAddToEntry( vValues, k, Entry );
        // inlining of macro Vec_FltForEachEntry
        for  k = 0; (k < Vec_FltSize(vvalues)) && (((Entry) = Vec_FltEntry(vvalues, k)), 1); k++  {
            if Vec_FltEntry(vValues, k) < Entry {
                Vec_FltWriteEntry( vValues, k, Entry ); }
        }
    }
//    Vec_FltForEachEntry( vValues, Entry, k )
//        Vec_FltWriteEntry( vValues, k, Entry/nTriples );
    // return the result
    *pvInd0 = vInd0;
    *pvInd1 = vInd1;
    *pvValues = vValues;
    return 1;
}

func Scl_LibertyReadTable(p *Scl_Tree_t, vOut *Vec_Str_t, pTiming *Scl_Item_t, pName string, vTemples *Vec_Ptr_t) int32 {
    var vIndex1 *Vec_Flt_t  = nil;
    var vIndex2 *Vec_Flt_t  = nil;
    var vValues *Vec_Flt_t  = nil;
    var vInd1, vInd2 *Vec_Flt_t;
    var pItem *Scl_Item_t;
    var pTable *Scl_Item_t = nil;
    var pThis string;
    var pTempl string = nil;
    var iPlace, i int32;
    var Entry float32;
    // find the table
    // inlining of macro Scl_ItemForEachChildName
    for  pTable = Scl_LibertyItem(p, pTiming.Child); (pTable != nil); pTable = Scl_LibertyItem(p, pTable.Next)  { if Scl_LibertyCompare(p, pTable.Key, pName) != 0 {} else {
        break; }
    }
    if pTable == nil {
        return 0; }
    // find the template
    pTempl = Scl_LibertyReadString(p, pTable.Head);
    if pTempl == nil || pTempl[0] == 0 {
        // read the numbers
        // inlining of macro Scl_ItemForEachChild
        for  pItem = Scl_LibertyItem(p, pTable.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  {
            if !(Scl_LibertyCompare(p, pItem.Key, "index_1") != 0) {
                assert(vIndex1 == nil, "vIndex1 == nil");
vIndex1 = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) );
            } else if !(Scl_LibertyCompare(p, pItem.Key, "index_2") != 0) {
                assert(vIndex2 == nil, "vIndex2 == nil");
vIndex2 = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) );
            } else if !(Scl_LibertyCompare(p, pItem.Key, "values") != 0) {
                assert(vValues == nil, "vValues == nil");
vValues = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) ); }
        }
        if vIndex1 == nil || vIndex2 == nil || vValues == nil { fmt.Printf("Incomplete table specification\n"); return 0; }
        // dump the table
        vInd1 = vIndex1;
        vInd2 = vIndex2;
        // write entries
        Vec_StrPutI_( vOut, Vec_FltSize(vInd1) );
        // inlining of macro Vec_FltForEachEntry
        for  i = 0; (i < Vec_FltSize(vInd1)) && (((Entry) = Vec_FltEntry(vInd1, i)), 1); i++  {
            Vec_StrPutF_( vOut, Entry );
        }
        Vec_StrPut_( vOut );
        // write entries
        Vec_StrPutI_( vOut, Vec_FltSize(vInd2) );
        // inlining of macro Vec_FltForEachEntry
        for  i = 0; (i < Vec_FltSize(vInd2)) && (((Entry) = Vec_FltEntry(vInd2, i)), 1); i++  {
            Vec_StrPutF_( vOut, Entry );
        }
        Vec_StrPut_( vOut );
        Vec_StrPut_( vOut );
        // write entries
        assert(Vec_FltSize(vInd1) * Vec_FltSize(vInd2) == Vec_FltSize(vValues), "Vec_FltSize(vInd1) * Vec_FltSize(vInd2) == Vec_FltSize(vValues)");
        // inlining of macro Vec_FltForEachEntry
        for  i = 0; (i < Vec_FltSize(vValues)) && (((Entry) = Vec_FltEntry(vValues, i)), 1); i++  {
            Vec_StrPutF_( vOut, Entry );
            if i % Vec_FltSize(vInd2) == Vec_FltSize(vInd2)-1 {
                Vec_StrPut_( vOut ); }
        }
    
    } else {
        // fetch the template
        iPlace = -1;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vTemples)) && (((pThis) = (char *)Vec_PtrEntry(vTemples, i)), 1); i++  {
            if i % 4 == 0 && !(int32(strings.Compare(pTempl, pThis)) != 0) {  
                iPlace = i;
                break;
            }
        }
        if iPlace == -1 { fmt.Printf("Template cannot be found in the template library\n"); return 0; }
        // read the numbers
        // inlining of macro Scl_ItemForEachChild
        for  pItem = Scl_LibertyItem(p, pTable.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  {
            if !(Scl_LibertyCompare(p, pItem.Key, "index_1") != 0) {
                assert(vIndex1 == nil, "vIndex1 == nil");
vIndex1 = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) );
            } else if !(Scl_LibertyCompare(p, pItem.Key, "index_2") != 0) {
                assert(vIndex2 == nil, "vIndex2 == nil");
vIndex2 = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) );
            } else if !(Scl_LibertyCompare(p, pItem.Key, "values") != 0) {
                assert(vValues == nil, "vValues == nil");
vValues = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) ); }
        }
        // check the template style
        vInd1 = (Vec_Flt_t *)Vec_PtrEntry( vTemples, iPlace + 2 ); // slew
        vInd2 = (Vec_Flt_t *)Vec_PtrEntry( vTemples, iPlace + 3 ); // load
        if Vec_PtrEntry(vTemples, iPlace + 1) == nil { // normal order (vIndex1 is slew; vIndex2 is load)
        
            assert(!(vIndex1 != nil) || Vec_FltSize(vIndex1) == Vec_FltSize(vInd1), "!vIndex1 || Vec_FltSize(vIndex1) == Vec_FltSize(vInd1)");
            assert(!(vIndex2 != nil) || Vec_FltSize(vIndex2) == Vec_FltSize(vInd2), "!vIndex2 || Vec_FltSize(vIndex2) == Vec_FltSize(vInd2)");
            if vIndex1 != nil {
                vInd1 = vIndex1;
            } else {
                vInd1 = vInd1;
            }
            if vIndex2 != nil {
                vInd2 = vIndex2;
            } else {
                vInd2 = vInd2;
            }
            // write entries
            Vec_StrPutI_( vOut, Vec_FltSize(vInd1) );
            // inlining of macro Vec_FltForEachEntry
            for  i = 0; (i < Vec_FltSize(vInd1)) && (((Entry) = Vec_FltEntry(vInd1, i)), 1); i++  {
                Vec_StrPutF_( vOut, Entry );
            }
            Vec_StrPut_( vOut );
            // write entries
            Vec_StrPutI_( vOut, Vec_FltSize(vInd2) );
            // inlining of macro Vec_FltForEachEntry
            for  i = 0; (i < Vec_FltSize(vInd2)) && (((Entry) = Vec_FltEntry(vInd2, i)), 1); i++  {
                Vec_StrPutF_( vOut, Entry );
            }
            Vec_StrPut_( vOut );
            Vec_StrPut_( vOut );
            // write entries
            assert(Vec_FltSize(vInd1) * Vec_FltSize(vInd2) == Vec_FltSize(vValues), "Vec_FltSize(vInd1) * Vec_FltSize(vInd2) == Vec_FltSize(vValues)");
            // inlining of macro Vec_FltForEachEntry
            for  i = 0; (i < Vec_FltSize(vValues)) && (((Entry) = Vec_FltEntry(vValues, i)), 1); i++  {
                Vec_StrPutF_( vOut, Entry );
                if i % Vec_FltSize(vInd2) == Vec_FltSize(vInd2)-1 {
                    Vec_StrPut_( vOut ); }
            }
        
        } else {  // reverse order (vIndex2 is slew; vIndex1 is load)
        
            assert(!(vIndex2 != nil) || Vec_FltSize(vIndex2) == Vec_FltSize(vInd1), "!vIndex2 || Vec_FltSize(vIndex2) == Vec_FltSize(vInd1)");
            assert(!(vIndex1 != nil) || Vec_FltSize(vIndex1) == Vec_FltSize(vInd2), "!vIndex1 || Vec_FltSize(vIndex1) == Vec_FltSize(vInd2)");
            if vIndex2 != nil {
                vInd1 = vIndex2;
            } else {
                vInd1 = vInd1;
            }
            if vIndex1 != nil {
                vInd2 = vIndex1;
            } else {
                vInd2 = vInd2;
            }
            // write entries
            Vec_StrPutI_( vOut, Vec_FltSize(vInd1) );
            // inlining of macro Vec_FltForEachEntry
            for  i = 0; (i < Vec_FltSize(vInd1)) && (((Entry) = Vec_FltEntry(vInd1, i)), 1); i++  {
                Vec_StrPutF_( vOut, Entry );
            }
            Vec_StrPut_( vOut );
            // write entries
            Vec_StrPutI_( vOut, Vec_FltSize(vInd2) );
            // inlining of macro Vec_FltForEachEntry
            for  i = 0; (i < Vec_FltSize(vInd2)) && (((Entry) = Vec_FltEntry(vInd2, i)), 1); i++  {
                Vec_StrPutF_( vOut, Entry );
            }
            Vec_StrPut_( vOut );
            Vec_StrPut_( vOut );
            // write entries -- transpose
            assert(Vec_FltSize(vInd1) * Vec_FltSize(vInd2) == Vec_FltSize(vValues), "Vec_FltSize(vInd1) * Vec_FltSize(vInd2) == Vec_FltSize(vValues)");
            // inlining of macro Vec_FltForEachEntry
            for  i = 0; (i < Vec_FltSize(vValues)) && (((Entry) = Vec_FltEntry(vValues, i)), 1); i++  {
                var x int32  = i % Vec_FltSize(vInd2);
                var y int32  = i / Vec_FltSize(vInd2);
                Entry = Vec_FltEntry( vValues, x * Vec_FltSize(vInd1) + y );
                Vec_StrPutF_( vOut, Entry );
                if i % Vec_FltSize(vInd2) == Vec_FltSize(vInd2)-1 {
                    Vec_StrPut_( vOut ); }
            }
        }
    }
    Vec_StrPut_( vOut );
    for  i = 0; i < 3; i++  { 
        Vec_StrPutF_( vOut, 0 );
    }
    for  i = 0; i < 4; i++  { 
        Vec_StrPutF_( vOut, 0 );
    }
    for  i = 0; i < 6; i++  { 
        Vec_StrPutF_( vOut, 0 );
    }
    Vec_FltFreeP( &vIndex1 );
    Vec_FltFreeP( &vIndex2 );
    Vec_FltFreeP( &vValues );
    Vec_StrPut_( vOut );
    Vec_StrPut_( vOut );
    return 1;
}
func Scl_LibertyPrintTemplates(vRes *Vec_Ptr_t) {
    var vArray *Vec_Flt_t; var i int32;
    assert(Vec_PtrSize(vRes) % 4 == 0, "Vec_PtrSize(vRes) % 4 == 0");
    fmt.Printf("There are %d slew/load templates\n", Vec_PtrSize(vRes) % 4);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRes)) && (((vArray) = (Vec_Flt_t *)Vec_PtrEntry(vRes, i)), 1); i++  {
        if i % 4 == 0 {
            fmt.Printf("%s\n", (char *)vArray);
        } else if i % 4 == 1 {
            fmt.Printf("%d\n", int32((vArray != nil)));
        } else if i % 4 == 2 || i % 4 == 3 {
            Vec_FltPrint( vArray ); }
        if i % 4 == 3 {
            fmt.Printf("\n"); }
    }
}
func Scl_LibertyReadTemplates(p *Scl_Tree_t) *Vec_Ptr_t {
    var vRes *Vec_Ptr_t  = nil;
    var vIndex1, vIndex2 *Vec_Flt_t;
    var pTempl, pItem *Scl_Item_t;
    var pVar1, pVar2 string;
    var fFlag0, fFlag1 int32;
    vRes = Vec_PtrAlloc( 100 );
    // inlining of macro Scl_ItemForEachChildName
    for  pTempl = Scl_LibertyItem(p, Scl_LibertyRoot(p).Child); pTempl; pTempl = Scl_LibertyItem(p, pTempl.Next)  { if Scl_LibertyCompare(p, pTempl.Key, "lu_table_template") != 0 {} else {
        pVar1 = pVar2 = nil;
        vIndex1 = vIndex2 = nil;
        // inlining of macro Scl_ItemForEachChild
        for  pItem = Scl_LibertyItem(p, pTempl.Child); (pItem != nil); pItem = Scl_LibertyItem(p, pItem.Next)  {
            if !(Scl_LibertyCompare(p, pItem.Key, "index_1") != 0) {
                assert(vIndex1 == nil, "vIndex1 == nil");
vIndex1 = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) );
            } else if !(Scl_LibertyCompare(p, pItem.Key, "index_2") != 0) {
                assert(vIndex2 == nil, "vIndex2 == nil");
vIndex2 = Scl_LibertyReadFloatVec( Scl_LibertyReadString(p, pItem.Head) );
            } else if !(Scl_LibertyCompare(p, pItem.Key, "variable_1") != 0) {
                assert(pVar1 == nil, "pVar1 == nil");
pVar1 = Abc_UtilStrsav( Scl_LibertyReadString(p, pItem.Head) );
            } else if !(Scl_LibertyCompare(p, pItem.Key, "variable_2") != 0) {
                assert(pVar2 == nil, "pVar2 == nil");
pVar2 = Abc_UtilStrsav( Scl_LibertyReadString(p, pItem.Head) ); }
        }
        if pVar1 == nil || pVar2 == nil {
            // inlining of macro ABC_FREE
            ((pVar1) ? (free((char *) (pVar1)), (pVar1) = 0) : 0);  
            // inlining of macro ABC_FREE
            ((pVar2) ? (free((char *) (pVar2)), (pVar2) = 0) : 0);
            Vec_FltFreeP( &vIndex1 );
            Vec_FltFreeP( &vIndex2 );
            continue;
        }
        assert(pVar1 != nil && pVar2 != nil, "pVar1 != nil && pVar2 != nil");
        fFlag0 = (!int32(strings.Compare(pVar1, "input_net_transition")) && !int32(strings.Compare(pVar2, "total_output_net_capacitance")));
        fFlag1 = (!int32(strings.Compare(pVar2, "input_net_transition")) && !int32(strings.Compare(pVar1, "total_output_net_capacitance")));
        // inlining of macro ABC_FREE
        ((pVar1) ? (free((char *) (pVar1)), (pVar1) = 0) : 0);  
        // inlining of macro ABC_FREE
        ((pVar2) ? (free((char *) (pVar2)), (pVar2) = 0) : 0);
        if !fFlag0 && !(fFlag1 != 0) {
            Vec_FltFreeP( &vIndex1 );
            Vec_FltFreeP( &vIndex2 );
            continue;
        }
        Vec_PtrPush( vRes, Abc_UtilStrsav( Scl_LibertyReadString(p, pTempl.Head) ) );
        Vec_PtrPush( vRes, fFlag0 ? nil : (void *)(ABC_PTRINT_T)1 );
        Vec_PtrPush( vRes, fFlag0 ? vIndex1 : vIndex2 );
        Vec_PtrPush( vRes, fFlag0 ? vIndex2 : vIndex1 );
    }
    }
    if Vec_PtrSize(vRes) == 0 {
        Abc_Print( 0, "Templates are not defined.\n" ); }
    // print templates
//    printf( "Found %d templates\n", Vec_PtrSize(vRes)/4 );
//    Scl_LibertyPrintTemplates( vRes );
    return vRes;
}
func Scl_LibertyReadSclStr(p *Scl_Tree_t, fVerbose int32, fVeryVerbose int32) *Vec_Str_t {
    var fUseFirstTable int32;
    var vOut *Vec_Str_t;
    var vNameIns *Vec_Ptr_t;
    var vTemples *Vec_Ptr_t = nil;
    var pCell, pPin, pTiming *Scl_Item_t;
    var vTruth *Vec_Wrd_t;
    var pFormula, pName string;
    var i, k, Counter, nOutputs, nCells int32;
    var nSkipped [4]int32 ;

    // read delay-table templates
    vTemples = Scl_LibertyReadTemplates( p );

    // start the library
    vOut = Vec_StrAlloc( 10000 );
    Vec_StrPutI_( vOut, ABC_SCL_CUR_VERSION );

    // top level information
    Vec_StrPut_( vOut );
    Vec_StrPutS_( vOut, Scl_LibertyReadString(p, Scl_LibertyRoot(p).Head) );
    Vec_StrPutS_( vOut, Scl_LibertyReadDefaultWireLoad(p) );
    Vec_StrPutS_( vOut, Scl_LibertyReadDefaultWireLoadSel(p) );
    Vec_StrPutF_( vOut, Scl_LibertyReadDefaultMaxTrans(p) );
    Vec_StrPutI_( vOut, Scl_LibertyReadTimeUnit(p) );
    Scl_LibertyReadLoadUnit( p, vOut );
    Vec_StrPut_( vOut );
    Vec_StrPut_( vOut );

    // read wire loads
    Scl_LibertyReadWireLoad( p, vOut );
    Scl_LibertyReadWireLoadSelect( p, vOut );

    // count cells
    nCells = 0;
    // inlining of macro Scl_ItemForEachChildName
    for  pCell = Scl_LibertyItem(p, Scl_LibertyRoot(p).Child); (pCell != nil); pCell = Scl_LibertyItem(p, pCell.Next)  { if Scl_LibertyCompare(p, pCell.Key, "cell") != 0 {} else {
        if Scl_LibertyReadCellIsFlop(p, pCell) != 0 {
            if fVeryVerbose != 0 {  fmt.Printf("Scl_LibertyReadGenlib() skipped sequential cell \"%s\".\n", Scl_LibertyReadString(p, pCell.Head)); }
            nSkipped[0]++;
            continue;
        }
        if Scl_LibertyReadCellIsDontUse(p, pCell) != 0 {
            if fVeryVerbose != 0 {  fmt.Printf("Scl_LibertyReadGenlib() skipped cell \"%s\" due to dont_use attribute.\n", Scl_LibertyReadString(p, pCell.Head)); }
            nSkipped[3]++;
            continue;
        }
        if Scl_LibertyReadCellIsThreeState(p, pCell) != 0 {
            if fVeryVerbose != 0 {  fmt.Printf("Scl_LibertyReadGenlib() skipped three-state cell \"%s\".\n", Scl_LibertyReadString(p, pCell.Head)); }
            nSkipped[1]++;
            continue;
        }
        if (Counter = Scl_LibertyReadCellOutputNum(p, pCell)) == 0 {
            if fVeryVerbose != 0 {  fmt.Printf("Scl_LibertyReadGenlib() skipped cell \"%s\" without logic function.\n", Scl_LibertyReadString(p, pCell.Head)); }
            nSkipped[2]++;
            continue;
        }
        nCells++;
    }
    }
    // read cells
    Vec_StrPutI_( vOut, nCells );
    Vec_StrPut_( vOut );
    Vec_StrPut_( vOut );
    // inlining of macro Scl_ItemForEachChildName
    for  pCell = Scl_LibertyItem(p, Scl_LibertyRoot(p).Child); (pCell != nil); pCell = Scl_LibertyItem(p, pCell.Next)  { if Scl_LibertyCompare(p, pCell.Key, "cell") != 0 {} else {
        if Scl_LibertyReadCellIsFlop(p, pCell) != 0 {
            continue; }
        if Scl_LibertyReadCellIsDontUse(p, pCell) != 0 {
            continue; }
        if Scl_LibertyReadCellIsThreeState(p, pCell) != 0 {
            continue; }
        if (Counter = Scl_LibertyReadCellOutputNum(p, pCell)) == 0 {
            continue; }
        // top level information
        Vec_StrPutS_( vOut, Scl_LibertyReadString(p, pCell.Head) );
        pName = Scl_LibertyReadCellArea(p, pCell);
        Vec_StrPutF_( vOut, pName ? atof(pName) : 1 );
        pName = Scl_LibertyReadCellLeakage(p, pCell);
        Vec_StrPutF_( vOut, pName ? atof(pName) : 0 );
        Vec_StrPutI_( vOut, Scl_LibertyReadDeriveStrength(p, pCell) );
        // pin count
        nOutputs = Scl_LibertyReadCellOutputNum( p, pCell );
        Vec_StrPutI_( vOut, Scl_LibertyItemNum(p, pCell, "pin") - nOutputs );
        Vec_StrPutI_( vOut, nOutputs );
        Vec_StrPut_( vOut );
        Vec_StrPut_( vOut );

        // input pins
        vNameIns = Vec_PtrAlloc( 16 );
        // inlining of macro Scl_ItemForEachChildName
        for  pPin = Scl_LibertyItem(p, pCell.Child); (pPin != nil); pPin = Scl_LibertyItem(p, pPin.Next)  { if Scl_LibertyCompare(p, pPin.Key, "pin") != 0 {} else {
            var CapOne, CapRise, CapFall float32;
            if Scl_LibertyReadPinFormula(p, pPin) != nil { // skip output pin
                continue; }
            assert(Scl_LibertyReadPinDirection(p, pPin) == 0, "Scl_LibertyReadPinDirection(p, pPin) == 0");
            pName = Scl_LibertyReadString(p, pPin.Head);
            Vec_PtrPush( vNameIns, Abc_UtilStrsav(pName) );
            Vec_StrPutS_( vOut, pName );
            CapOne  = Scl_LibertyReadPinCap( p, pPin, "capacitance" );
            CapRise = Scl_LibertyReadPinCap( p, pPin, "rise_capacitance" );
            CapFall = Scl_LibertyReadPinCap( p, pPin, "fall_capacitance" );
            if CapRise == 0 {
                CapRise = CapOne; }
            if CapFall == 0 {
                CapFall = CapOne; }
            Vec_StrPutF_( vOut, CapRise );
            Vec_StrPutF_( vOut, CapFall );
            Vec_StrPut_( vOut );
        }
        }
        Vec_StrPut_( vOut );
        // output pins
        // inlining of macro Scl_ItemForEachChildName
        for  pPin = Scl_LibertyItem(p, pCell.Child); (pPin != nil); pPin = Scl_LibertyItem(p, pPin.Next)  { if Scl_LibertyCompare(p, pPin.Key, "pin") != 0 {} else {
            if !Scl_LibertyReadPinFormula(p, pPin) { // skip input pin
                continue; }
            assert(Scl_LibertyReadPinDirection(p, pPin) == 1, "Scl_LibertyReadPinDirection(p, pPin) == 1");
            pName = Scl_LibertyReadString(p, pPin.Head);
            Vec_StrPutS_( vOut, pName );
            Vec_StrPutF_( vOut, Scl_LibertyReadPinCap( p, pPin, "max_capacitance" ) );
            Vec_StrPutF_( vOut, Scl_LibertyReadPinCap( p, pPin, "max_transition" ) );
            Vec_StrPutI_( vOut, Vec_PtrSize(vNameIns) );
            pFormula = Scl_LibertyReadPinFormula(p, pPin);
            Vec_StrPutS_( vOut, pFormula );
            // write truth table
            vTruth = Mio_ParseFormulaTruth( pFormula, (char **)Vec_PtrArray(vNameIns), Vec_PtrSize(vNameIns) );
            if vTruth == nil {
                return nil; }
            for  i = 0; i < Abc_Truth6WordNum(Vec_PtrSize(vNameIns)); i++  {
                Vec_StrPutW_( vOut, Vec_WrdEntry(vTruth, i) );
            }
            Vec_WrdFree( vTruth );
            Vec_StrPut_( vOut );
            Vec_StrPut_( vOut );

            // write the delay tables
            if fUseFirstTable != 0 {
                // inlining of macro Vec_PtrForEachEntry
                for  i = 0; (i < Vec_PtrSize(vNameIns)) && (((pName) = (char *)Vec_PtrEntry(vNameIns, i)), 1); i++  {
                    pTiming = Scl_LibertyReadPinTiming( p, pPin, pName );
                    Vec_StrPutS_( vOut, pName );
                    Vec_StrPutI_( vOut, int32((pTiming != nil)) );
                    if pTiming == nil { // output does not depend on input
                        continue; }
                    Vec_StrPutI_( vOut, Scl_LibertyReadTimingSense(p, pTiming) );
                    Vec_StrPut_( vOut );
                    Vec_StrPut_( vOut );
                    // some cells only have 'rise' or 'fall' but not both - here we work around this
                    if !(Scl_LibertyReadTable( p, vOut, pTiming, "cell_rise",           vTemples ) != 0) {
                        if !(Scl_LibertyReadTable( p, vOut, pTiming, "cell_fall",       vTemples ) != 0) { fmt.Printf("Table cannot be found\n"); return nil; } }                              
                    if !(Scl_LibertyReadTable( p, vOut, pTiming, "cell_fall",           vTemples ) != 0) {
                        if !(Scl_LibertyReadTable( p, vOut, pTiming, "cell_rise",       vTemples ) != 0) { fmt.Printf("Table cannot be found\n"); return nil; } }                              
                    if !(Scl_LibertyReadTable( p, vOut, pTiming, "rise_transition",     vTemples ) != 0) {
                        if !(Scl_LibertyReadTable( p, vOut, pTiming, "fall_transition", vTemples ) != 0) { fmt.Printf("Table cannot be found\n"); return nil; } }                              
                    if !(Scl_LibertyReadTable( p, vOut, pTiming, "fall_transition",     vTemples ) != 0) {
                        if !(Scl_LibertyReadTable( p, vOut, pTiming, "rise_transition", vTemples ) != 0) { fmt.Printf("Table cannot be found\n"); return nil; } }  
                }
                continue;
            }

            // write the timing tables
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(vNameIns)) && (((pName) = (char *)Vec_PtrEntry(vNameIns, i)), 1); i++  {
                var vTables [4]*Vec_Ptr_t;
                var vTimings *Vec_Ptr_t;
                vTimings = Scl_LibertyReadPinTimingAll( p, pPin, pName );
                Vec_StrPutS_( vOut, pName );
                Vec_StrPutI_( vOut, int32((Vec_PtrSize(vTimings) != 0)) );
                if Vec_PtrSize(vTimings) == 0 { // output does not depend on input
                
                    Vec_PtrFree( vTimings );
                    continue;
                }
                Vec_StrPutI_( vOut, Scl_LibertyReadTimingSense(p, (Scl_Item_t *)Vec_PtrEntry(vTimings, 0)) );
                Vec_StrPut_( vOut );
                Vec_StrPut_( vOut );
                // collect the timing tables
                for  k = 0; k < 4; k++  {
                    vTables[k] = Vec_PtrAlloc( 16 );
                }
                // inlining of macro Vec_PtrForEachEntry
                for  k = 0; (k < Vec_PtrSize(vTimings)) && (((pTiming) = (Scl_Item_t *)Vec_PtrEntry(vTimings, k)), 1); k++  {
                    // some cells only have 'rise' or 'fall' but not both - here we work around this
                    if !(Scl_LibertyScanTable( p, vTables[0], pTiming, "cell_rise",           vTemples ) != 0) {
                        if !(Scl_LibertyScanTable( p, vTables[0], pTiming, "cell_fall",       vTemples ) != 0) { fmt.Printf("Table cannot be found\n"); return nil; } }                              
                    if !(Scl_LibertyScanTable( p, vTables[1], pTiming, "cell_fall",           vTemples ) != 0) {
                        if !(Scl_LibertyScanTable( p, vTables[1], pTiming, "cell_rise",       vTemples ) != 0) { fmt.Printf("Table cannot be found\n"); return nil; } }                              
                    if !(Scl_LibertyScanTable( p, vTables[2], pTiming, "rise_transition",     vTemples ) != 0) {
                        if !(Scl_LibertyScanTable( p, vTables[2], pTiming, "fall_transition", vTemples ) != 0) { fmt.Printf("Table cannot be found\n"); return nil; } }                              
                    if !(Scl_LibertyScanTable( p, vTables[3], pTiming, "fall_transition",     vTemples ) != 0) {
                        if !(Scl_LibertyScanTable( p, vTables[3], pTiming, "rise_transition", vTemples ) != 0) { fmt.Printf("Table cannot be found\n"); return nil; } }  
                }
                Vec_PtrFree( vTimings );
                // compute worse case of the tables
                for  k = 0; k < 4; k++  {
                    var vInd0, vInd1, vValues *Vec_Flt_t;
                    if !(Scl_LibertyComputeWorstCase( vTables[k], &vInd0, &vInd1, &vValues ) != 0) { fmt.Printf("Table indexes have different values\n"); return nil; }  
                    Vec_VecFree( (Vec_Vec_t *)vTables[k] );
                    Scl_LibertyDumpTables( vOut, vInd0, vInd1, vValues );
                    Vec_FltFree( vInd0 );
                    Vec_FltFree( vInd1 );
                    Vec_FltFree( vValues );
                }
            }
        }
        }
        Vec_StrPut_( vOut );
        Vec_PtrFreeFree( vNameIns );
    }
    }
    // free templates
    if vTemples != nil {
        var vArray *Vec_Flt_t;
        assert(Vec_PtrSize(vTemples) % 4 == 0, "Vec_PtrSize(vTemples) % 4 == 0");
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vTemples)) && (((vArray) = (Vec_Flt_t *)Vec_PtrEntry(vTemples, i)), 1); i++  {
            if vArray == nil {
                continue; }
            if i % 4 == 0 {
                // inlining of macro ABC_FREE
                ((vArray) ? (free((char *) (vArray)), (vArray) = 0) : 0);
            } else if i % 4 == 2 || i % 4 == 3 {
                Vec_FltFree( vArray ); }
        }
        Vec_PtrFree( vTemples );
    }
    if fVerbose != 0 {
        fmt.Printf("Library \"%s\" from \"%s\" has %d cells ", Scl_LibertyReadString(p, Scl_LibertyRoot(p).Head), p.pFileName, nCells);
        fmt.Printf("(%d skipped: %d seq; %d tri-state; %d no func; %d dont_use).  ", nSkipped[0]+nSkipped[1]+nSkipped[2], nSkipped[0], nSkipped[1], nSkipped[2], nSkipped[3]);
        Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    }
    return vOut;
}
func Abc_SclReadLiberty(pFileName string, fVerbose int32, fVeryVerbose int32) *SC_Lib {
    var pLib *SC_Lib;
    var p *Scl_Tree_t;
    var vStr *Vec_Str_t;
    p = Scl_LibertyParse( pFileName, fVeryVerbose );
    if p == nil {
        return nil; }
//    Scl_LibertyParseDump( p, "temp_.lib" );
    // collect relevant data
    vStr = Scl_LibertyReadSclStr( p, fVerbose, fVeryVerbose );
    Scl_LibertyStop( p, fVeryVerbose );
    if vStr == nil {
        return nil; }
    // construct SCL data-structure
    pLib = Abc_SclReadFromStr( vStr );
    if pLib == nil {
        return nil; }
    pLib.pFileName = Abc_UtilStrsav( pFileName );
    Abc_SclLibNormalize( pLib );
    Vec_StrFree( vStr );
//    printf( "Average slew = %.2f ps\n", Abc_SclComputeAverageSlew(pLib) );
    return pLib;
}

/**Function*************************************************************

  Synopsis    [Experiments with Liberty parsing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibertyTest() {
    var pFileName string  = "bwrc.lib";
    var fVerbose int32  = 1;
    var fVeryVerbose int32;
    var p *Scl_Tree_t;
    var vStr *Vec_Str_t;
//    return;
    p = Scl_LibertyParse( pFileName, fVeryVerbose );
    if p == nil {
        return; }
//    Scl_LibertyParseDump( p, "temp_.lib" );
    vStr = Scl_LibertyReadSclStr( p, fVerbose, fVeryVerbose );
    Scl_LibertyStringDump( "test_scl.lib", vStr );
    Vec_StrFree( vStr );
    Scl_LibertyStop( p, fVerbose );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

