package amap

import (
	"fmt"
	"strings"
	"time"
)

import "os"

/**CFile****************************************************************

  FileHead    [amapLiberty.c]

  SystemHead  [ABC: Logic synthesis and verification system.]

  PackageHead [Technology mapper for standard cells.]

  Synopsis    [Liberty parser.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: amapLiberty.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "amapInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const ABC_MAX_LIB_STR_LEN = 5000

// entry types
 type Amap_LibertyType_t int32
const ( 
    AMAP_LIBERTY_NONE Amap_LibertyType_t = 0
        // 0:  unknown
    AMAP_LIBERTY_PROC Amap_LibertyType_t = AMAP_LIBERTY_NONE + 1
            // 1:  procedure :  key(head){body}
    AMAP_LIBERTY_EQUA Amap_LibertyType_t = AMAP_LIBERTY_PROC + 1
            // 2:  equation  :  key:head;
    AMAP_LIBERTY_LIST Amap_LibertyType_t = AMAP_LIBERTY_EQUA + 1             // 3:  list      :  key(head) 

)

 

   
type Amap_Pair_t struct {
    Beg             int32;          // item beginning
    End             int32;          // item end
}

   
type Amap_Item_t struct {
    Type             int32;         // Amap_LibertyType_t
    iLine             int32;        // file line where the item's spec begins
    Key     Amap_Pair_t;          // key part
    Head     Amap_Pair_t;         // head part 
    Body     Amap_Pair_t;         // body part
    Next             int32;         // next item in the list 
    Child             int32;        // first child item 
}

   
type Amap_Tree_t struct {
    pFileName           string;    // input Liberty file name
    pContents           string;    // file contents
    nContents             int32;    // file size
    nLines             int32;       // line counter
    nItems             int32;       // number of items
    nItermAlloc             int32;  // number of items allocated
    pItems    *Amap_Item_t;       // the items
    pError           string;       // the error string
}

func Amap_LibertyRoot(p *Amap_Tree_t) *Amap_Item_t { return p.pItems;                                                 }
func Amap_LibertyItem(p *Amap_Tree_t, v int32) *Amap_Item_t { assert(v < p.nItems, "v < p.nItems"); if v < 0 {
    return nil;
} else {
    return p.pItems + v;
}     }
func Amap_LibertyCompare(p *Amap_Tree_t, Pair Amap_Pair_t, pStr string) int32 { return strncmp( p.pContents+Pair.Beg, pStr, Pair.End-Pair.Beg ); }
func Amap_PrintWord(pFile *os.File, p *Amap_Tree_t, Pair Amap_Pair_t) { var pBeg string = p.pContents+Pair.Beg;
var pEnd string = p.pContents+Pair.End; for  pBeg < pEnd  { fputc( *pBeg++, pFile ); } }
func Amap_PrintSpace(pFile *os.File, nOffset int32) { var i int32; for  i = 0; i < nOffset; i++  { fputc(' ', pFile); }         }
func Amap_LibertyItemId(p *Amap_Tree_t, pItem *Amap_Item_t) int32 { return pItem - p.pItems;                                         }

// Original macro:
// #define Amap_ItemForEachChild( p, pItem, pChild ) \
//     for ( pChild = Amap_LibertyItem(p, pItem->Child); pChild; pChild = Amap_LibertyItem(p, pChild->Next) )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Prints parse tree in Liberty format.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyPrintLibertyItem(pFile *os.File, p *Amap_Tree_t, pItem *Amap_Item_t, nOffset int32) {
    if pItem.Type == AMAP_LIBERTY_PROC {
        Amap_PrintSpace( pFile, nOffset );
        Amap_PrintWord( pFile, p, pItem.Key );
        fmt.Fprintf(pFile, "(");
        Amap_PrintWord( pFile, p, pItem.Head );
        fmt.Fprintf(pFile, ") {\n");
        if Amap_LibertyItem(p, pItem.Child) != nil {
            Amap_LibertyPrintLibertyItem( pFile, p, Amap_LibertyItem(p, pItem.Child), nOffset + 1 ); }
        Amap_PrintSpace( pFile, nOffset );
        fmt.Fprintf(pFile, "}\n");
    
    } else if pItem.Type == AMAP_LIBERTY_EQUA {
        Amap_PrintSpace( pFile, nOffset );
        Amap_PrintWord( pFile, p, pItem.Key );
        fmt.Fprintf(pFile, " : ");
        Amap_PrintWord( pFile, p, pItem.Head );
        fmt.Fprintf(pFile, ";\n");
    
    } else if pItem.Type == AMAP_LIBERTY_LIST {
        Amap_PrintSpace( pFile, nOffset );
        Amap_PrintWord( pFile, p, pItem.Key );
        fmt.Fprintf(pFile, "(");
        Amap_PrintWord( pFile, p, pItem.Head );
        fmt.Fprintf(pFile, ");\n");
    
    } else { assert(false, "0"); }
    if Amap_LibertyItem(p, pItem.Next) != nil {
        Amap_LibertyPrintLibertyItem( pFile, p, Amap_LibertyItem(p, pItem.Next), nOffset ); }
}

/**Function*************************************************************

  Synopsis    [Prints parse tree in Liberty format.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyPrintLiberty(p *Amap_Tree_t, pFileName string) int32 {
    var pFile *os.File;
    if pFileName == nil {
        pFile = os.Stdout;
    } else {
        pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Printf("Amap_LibertyPrintLiberty(): The output file is unavailable (absent or open).\n");
            return 0;
        }
    }
    Amap_LibertyPrintLibertyItem( pFile, p, Amap_LibertyRoot(p), 0 );
    if pFile != os.Stdout {
        _ = (pFile).Close(); }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Returns the time stamp.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Amap_LibertyTimeStamp_Buffer [100]byte;
func Amap_LibertyTimeStamp() string {
    
    var TimeStamp string;
    var ltime int64;
    // get the current time
    time( &ltime );
    TimeStamp = asctime( localtime( &ltime ) );
    TimeStamp[ len(TimeStamp) - 1 ] = 0;
    assert(len(TimeStamp) < 100, "strlen(TimeStamp) < 100");
    strcpy( spStatic_Amap_LibertyTimeStamp_Buffer, TimeStamp );
    return spStatic_Amap_LibertyTimeStamp_Buffer;
}

/**Function*************************************************************

  Synopsis    [Returns cell's function.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyCellIsFlop(p *Amap_Tree_t, pCell *Amap_Item_t) int32 {
    var pAttr *Amap_Item_t;
    // inlining of macro Amap_ItemForEachChild
    for  pAttr = Amap_LibertyItem(p, pCell.Child); (pAttr != nil); pAttr = Amap_LibertyItem(p, pAttr.Next)  {
        if !(Amap_LibertyCompare(p, pAttr.Key, "ff") != 0) ||
             !(Amap_LibertyCompare(p, pAttr.Key, "latch") != 0) {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns cell's function.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyCellIsDontUse(p *Amap_Tree_t, pCell *Amap_Item_t) int32 {
    var pAttr *Amap_Item_t;
    // inlining of macro Amap_ItemForEachChild
    for  pAttr = Amap_LibertyItem(p, pCell.Child); (pAttr != nil); pAttr = Amap_LibertyItem(p, pAttr.Next)  {
        if !(Amap_LibertyCompare(p, pAttr.Key, "dont_use") != 0) {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns pin's function.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyPinFunction(p *Amap_Tree_t, pPin *Amap_Item_t) *Amap_Item_t {
    var pFunc *Amap_Item_t;
    // inlining of macro Amap_ItemForEachChild
    for  pFunc = Amap_LibertyItem(p, pPin.Child); (pFunc != nil); pFunc = Amap_LibertyItem(p, pFunc.Next)  {
        if !(Amap_LibertyCompare(p, pFunc.Key, "function") != 0) {
            return pFunc; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Returns output pin(s).]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyCellOutput(p *Amap_Tree_t, pCell *Amap_Item_t) *Amap_Item_t {
    var pPin *Amap_Item_t;
    // inlining of macro Amap_ItemForEachChild
    for  pPin = Amap_LibertyItem(p, pCell.Child); (pPin != nil); pPin = Amap_LibertyItem(p, pPin.Next)  {
        if Amap_LibertyCompare(p, pPin.Key, "pin") != 0 {
            continue; }
        if Amap_LibertyPinFunction(p, pPin) != nil {
            return pPin; }
    }
    return nil;
}
func Amap_LibertyCellOutputs(p *Amap_Tree_t, pCell *Amap_Item_t) *Vec_Ptr_t {
    var pPin *Amap_Item_t;
    var vOutPins *Vec_Ptr_t;
    vOutPins = Vec_PtrAlloc( 2 );
    // inlining of macro Amap_ItemForEachChild
    for  pPin = Amap_LibertyItem(p, pCell.Child); (pPin != nil); pPin = Amap_LibertyItem(p, pPin.Next)  {
        if Amap_LibertyCompare(p, pPin.Key, "pin") != 0 {
            continue; }
        if Amap_LibertyPinFunction(p, pPin) != nil {
            Vec_PtrPush( vOutPins, pPin ); }
    }
    return vOutPins;
}

/**Function*************************************************************

  Synopsis    [Returns cell's area.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyCellArea(p *Amap_Tree_t, pCell *Amap_Item_t) *Amap_Item_t {
    var pArea *Amap_Item_t;
    // inlining of macro Amap_ItemForEachChild
    for  pArea = Amap_LibertyItem(p, pCell.Child); (pArea != nil); pArea = Amap_LibertyItem(p, pArea.Next)  {
        if Amap_LibertyCompare(p, pArea.Key, "area") != 0 {
            continue; }
        return pArea;
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Count cell's output pins (pins with a logic function).]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyCellCountOutputs(p *Amap_Tree_t, pCell *Amap_Item_t) int32 {
    var pPin *Amap_Item_t;
    var Counter int32;
    // inlining of macro Amap_ItemForEachChild
    for  pPin = Amap_LibertyItem(p, pCell.Child); (pPin != nil); pPin = Amap_LibertyItem(p, pPin.Next)  {
        if Amap_LibertyCompare(p, pPin.Key, "pin") != 0 {
            continue; }
        if Amap_LibertyPinFunction(p, pPin) != nil {
            Counter++; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Gets the name to write.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Amap_LibertyGetString_Buffer [ABC_MAX_LIB_STR_LEN]byte;
func Amap_LibertyGetString(p *Amap_Tree_t, Pair Amap_Pair_t) string { 
     
    assert(Pair.End-Pair.Beg < ABC_MAX_LIB_STR_LEN, "Pair.End-Pair.Beg < ABC_MAX_LIB_STR_LEN");
    strncpy( spStatic_Amap_LibertyGetString_Buffer, p.pContents+Pair.Beg, Pair.End-Pair.Beg ); 
    spStatic_Amap_LibertyGetString_Buffer[Pair.End-Pair.Beg] = 0;
    return spStatic_Amap_LibertyGetString_Buffer;
}

/**Function*************************************************************

  Synopsis    [Gets the name to write.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Amap_LibertyGetStringFormula_Buffer [ABC_MAX_LIB_STR_LEN]byte;
func Amap_LibertyGetStringFormula(p *Amap_Tree_t, Pair Amap_Pair_t) string { 
     
    assert(Pair.End-Pair.Beg-2 < ABC_MAX_LIB_STR_LEN, "Pair.End-Pair.Beg-2 < ABC_MAX_LIB_STR_LEN");
    strncpy( spStatic_Amap_LibertyGetStringFormula_Buffer, p.pContents+Pair.Beg+1, Pair.End-Pair.Beg-2 ); 
    spStatic_Amap_LibertyGetStringFormula_Buffer[Pair.End-Pair.Beg-2] = 0;
    return spStatic_Amap_LibertyGetStringFormula_Buffer;
}

/**Function*************************************************************

  Synopsis    [Prints parse tree in Genlib format.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyPrintGenlib(p *Amap_Tree_t, pFileName string, fVerbose int32) int32 {
    var pFile *os.File;
    var vOutputs *Vec_Ptr_t;
    var pCell, pArea, pFunc, pPin, pOutput *Amap_Item_t;
    var pForm string;
    var i, Counter int32;
    if pFileName == nil {
        pFile = os.Stdout;
    } else {
        pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Printf("Amap_LibertyPrintGenlib(): The output file is unavailable (absent or open).\n");
            return 0;
        }
    }
    fmt.Fprintf(pFile, "# This Genlib file was generated by ABC on %s\n", Amap_LibertyTimeStamp());
    fmt.Fprintf(pFile, "# The standard cell library \"%s\" is from Liberty file \"%s\"\n", Amap_LibertyGetString(p, Amap_LibertyRoot(p).Head), p.pFileName);
    fmt.Fprintf(pFile, "# (To find out more about Genlib format, google for \"sis_paper.ps\")\n");

    fmt.Fprintf(pFile, "GATE  ");
    fmt.Fprintf(pFile, "%16s  ", "_const0_");
    fmt.Fprintf(pFile, "%f  ", 0.0);
    fmt.Fprintf(pFile, "%s=", "z");
    fmt.Fprintf(pFile, "%s;\n", "CONST0");

    fmt.Fprintf(pFile, "GATE  ");
    fmt.Fprintf(pFile, "%16s  ", "_const1_");
    fmt.Fprintf(pFile, "%f  ", 0.0);
    fmt.Fprintf(pFile, "%s=", "z");
    fmt.Fprintf(pFile, "%s;\n", "CONST1");

    // inlining of macro Amap_ItemForEachChild
    for  pCell = Amap_LibertyItem(p, Amap_LibertyRoot(p).Child); (pCell != nil); pCell = Amap_LibertyItem(p, pCell.Next)  {
/*
    if ( strcmp(Amap_LibertyGetString(p, pCell->Head), "HA1SVTX1") == 0 )
    {
        int s = 0;
    }
*/
        if Amap_LibertyCompare(p, pCell.Key, "cell") != 0 {
            continue; }
        if Amap_LibertyCellIsFlop(p, pCell) != 0 {
            if fVerbose != 0 {
                fmt.Printf("Amap_LibertyPrintGenlib() skipped sequential cell \"%s\".\n", Amap_LibertyGetString(p, pCell.Head)); }
            continue;
        }
        if Amap_LibertyCellIsDontUse(p, pCell) != 0 {
            if fVerbose != 0 {
                fmt.Printf("Amap_LibertyPrintGenlib() skipped cell \"%s\" due to dont_use attribute.\n", Amap_LibertyGetString(p, pCell.Head)); }
            continue;
        }
        Counter = Amap_LibertyCellCountOutputs( p, pCell );
        if Counter == 0 {
            if fVerbose != 0 {
                fmt.Printf("Amap_LibertyPrintGenlib() skipped cell \"%s\" without logic function.\n", Amap_LibertyGetString(p, pCell.Head)); }
            continue;
        }
/*
        if ( Counter > 1 )
        {
            if ( fVerbose )
                printf( "Amap_LibertyPrintGenlib() skipped multi-output cell \"%s\".\n", Amap_LibertyGetString(p, pCell->Head) );
            continue;
        }
*/
        pArea = Amap_LibertyCellArea( p, pCell );
        if pArea == nil {
            if fVerbose != 0 {
                fmt.Printf("Amap_LibertyPrintGenlib() skipped cell \"%s\" with unspecified area.\n", Amap_LibertyGetString(p, pCell.Head)); }
            continue;
        }
//        pOutput = Amap_LibertyCellOutput( p, pCell );
        vOutputs = Amap_LibertyCellOutputs( p, pCell );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vOutputs)) && (((pOutput) = (Amap_Item_t *)Vec_PtrEntry(vOutputs, i)), 1); i++  {
            pFunc   = Amap_LibertyPinFunction( p, pOutput );
            pForm   = Amap_LibertyGetStringFormula( p, pFunc.Head );
            if !(int32(strings.Compare(pForm, "0")) != 0) || !(int32(strings.Compare(pForm, "1")) != 0) {
                if fVerbose != 0 {
                    fmt.Printf("Amap_LibertyPrintGenlib() skipped cell \"%s\" with constant formula \"%s\".\n", Amap_LibertyGetString(p, pCell.Head), pForm); }
                continue;
            }
            fmt.Fprintf(pFile, "GATE  ");
            fmt.Fprintf(pFile, "%16s  ", Amap_LibertyGetString(p, pCell.Head));
            fmt.Fprintf(pFile, "%f  ", atof(Amap_LibertyGetString(p, pArea.Head)));
            fmt.Fprintf(pFile, "%s=", Amap_LibertyGetString(p, pOutput.Head));
            fmt.Fprintf(pFile, "%s;\n", Amap_LibertyGetStringFormula(p, pFunc.Head));
            // inlining of macro Amap_ItemForEachChild
            for  pPin = Amap_LibertyItem(p, pCell.Child); (pPin != nil); pPin = Amap_LibertyItem(p, pPin.Next)  {
                if Vec_PtrFind(vOutputs, pPin) == -1 && !(Amap_LibertyCompare(p, pPin.Key, "pin") != 0) {
                    fmt.Fprintf(pFile, "    PIN  %13s  UNKNOWN  1  999  1.00  0.00  1.00  0.00\n", Amap_LibertyGetString(p, pPin.Head)); }
            }
        }
        Vec_PtrFree( vOutputs );
    }
    if pFile != os.Stdout {
        _ = (pFile).Close(); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Prints parse tree in Genlib format.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyPrintGenlibStr(p *Amap_Tree_t, fVerbose int32) *Vec_Str_t {
    var vStr *Vec_Str_t;
    var Buffer string // was [100]byte in C

    var vOutputs *Vec_Ptr_t;
    var pCell, pArea, pFunc, pPin, pOutput *Amap_Item_t;
    var i, Counter int32;
    var pForm string;

    vStr = Vec_StrAlloc( 1000 );

    Vec_StrPrintStr( vStr, "GATE          _const0_  0.000000  z=CONST0;\n" );
    Vec_StrPrintStr( vStr, "GATE          _const1_  0.000000  z=CONST1;\n" );
    // inlining of macro Amap_ItemForEachChild
    for  pCell = Amap_LibertyItem(p, Amap_LibertyRoot(p).Child); (pCell != nil); pCell = Amap_LibertyItem(p, pCell.Next)  {
        if Amap_LibertyCompare(p, pCell.Key, "cell") != 0 {
            continue; }
        if Amap_LibertyCellIsFlop(p, pCell) != 0 {
            if fVerbose != 0 {
                fmt.Printf("Amap_LibertyPrintGenlib() skipped sequential cell \"%s\".\n", Amap_LibertyGetString(p, pCell.Head)); }
            continue;
        }
        if Amap_LibertyCellIsDontUse(p, pCell) != 0 {
            if fVerbose != 0 {
                fmt.Printf("Amap_LibertyPrintGenlib() skipped cell \"%s\" due to dont_use attribute.\n", Amap_LibertyGetString(p, pCell.Head)); }
            continue;
        }
        Counter = Amap_LibertyCellCountOutputs( p, pCell );
        if Counter == 0 {
            if fVerbose != 0 {
                fmt.Printf("Amap_LibertyPrintGenlib() skipped cell \"%s\" without logic function.\n", Amap_LibertyGetString(p, pCell.Head)); }
            continue;
        }
        pArea = Amap_LibertyCellArea( p, pCell );
        if pArea == nil {
            if fVerbose != 0 {
                fmt.Printf("Amap_LibertyPrintGenlib() skipped cell \"%s\" with unspecified area.\n", Amap_LibertyGetString(p, pCell.Head)); }
            continue;
        }
        vOutputs = Amap_LibertyCellOutputs( p, pCell );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vOutputs)) && (((pOutput) = (Amap_Item_t *)Vec_PtrEntry(vOutputs, i)), 1); i++  {
            pFunc   = Amap_LibertyPinFunction( p, pOutput );
            pForm   = Amap_LibertyGetStringFormula( p, pFunc.Head );
            if !(int32(strings.Compare(pForm, "0")) != 0) || !(int32(strings.Compare(pForm, "1")) != 0) {
                if fVerbose != 0 {
                    fmt.Printf("Amap_LibertyPrintGenlib() skipped cell \"%s\" with constant formula \"%s\".\n", Amap_LibertyGetString(p, pCell.Head), pForm); }
                continue;
            }
/*
            fprintf( pFile, "GATE  " );
            fprintf( pFile, "%16s  ", Amap_LibertyGetString(p, pCell->Head) );
            fprintf( pFile, "%f  ",   atof(Amap_LibertyGetString(p, pArea->Head)) );
            fprintf( pFile, "%s=",    Amap_LibertyGetString(p, pOutput->Head) );
            fprintf( pFile, "%s;\n",  Amap_LibertyGetStringFormula(p, pFunc->Head) );
            Amap_ItemForEachChild( p, pCell, pPin )
                if ( Vec_PtrFind(vOutputs, pPin) == -1 && !Amap_LibertyCompare(p, pPin->Key, "pin") )
                    fprintf( pFile, "    PIN  %13s  UNKNOWN  1  999  1.00  0.00  1.00  0.00\n", Amap_LibertyGetString(p, pPin->Head) );
*/
            Vec_StrPrintStr( vStr, "GATE " );
            Vec_StrPrintStr( vStr, Amap_LibertyGetString(p, pCell.Head) );
            Vec_StrPrintStr( vStr, " " );
            Buffer = fmt.Sprintf("%f", atof(Amap_LibertyGetString(p, pArea.Head)));
            Vec_StrPrintStr( vStr, Buffer );
            Vec_StrPrintStr( vStr, " " );
            Vec_StrPrintStr( vStr, Amap_LibertyGetString(p, pOutput.Head) );
            Vec_StrPrintStr( vStr, "=" );
            Vec_StrPrintStr( vStr, Amap_LibertyGetStringFormula(p, pFunc.Head) );
            Vec_StrPrintStr( vStr, ";\n" );
            // inlining of macro Amap_ItemForEachChild
            for  pPin = Amap_LibertyItem(p, pCell.Child); (pPin != nil); pPin = Amap_LibertyItem(p, pPin.Next)  {
                if Vec_PtrFind(vOutputs, pPin) == -1 && !(Amap_LibertyCompare(p, pPin.Key, "pin") != 0) {
                    Vec_StrPrintStr( vStr, "  PIN " );
                    Vec_StrPrintStr( vStr, Amap_LibertyGetString(p, pPin.Head) );
                    Vec_StrPrintStr( vStr, " UNKNOWN  1  999  1.00  0.00  1.00  0.00\n" );
                }
            }
        }
        Vec_PtrFree( vOutputs );
    }
    Vec_StrPrintStr( vStr, "\n.end\n" );
    Vec_StrPush( vStr, '\0' );
//    printf( "%s", Vec_StrArray(vStr) );
    return vStr;
}


/**Function*************************************************************

  Synopsis    [Returns the file size.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyFileSize(pFileName string) int32 {
    var pFile *os.File;
    var nFileSize int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Amap_LibertyFileSize(): The input file is unavailable (absent or open).\n");
        return 0;
    }
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile ); 
    _ = (pFile).Close();
    return nFileSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyFixFileHead(pFileName string) {
    var pHead string;
    for  pHead = pFileName; *pHead; pHead++  {
        if *pHead == '>' {
            *pHead = '\\'; }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyCountItems(pBeg string, pEnd string) int32 {
    var Counter int32;
    for  ; pBeg < pEnd; pBeg++  {
        Counter += (*pBeg == '(' || *pBeg == ':');
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Removes C-style comments.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyWipeOutComments(pBeg string, pEnd string) {
    var pCur, pStart string;
    for  pCur = pBeg; pCur < pEnd; pCur++  {
    if pCur[0] == '/' && pCur[1] == '*' {
        for  pStart = pCur; pCur < pEnd; pCur++  {
        if pCur[0] == '*' && pCur[1] == '/' {
            for  ; pStart < pCur + 2; pStart++  {
            if *pStart != '\n' { *pStart = ' '; }
            }
            break;
        }
        } }
    }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the character is space.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyCharIsSpace(c byte) int32 {
    return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\\';
}

/**Function*************************************************************

  Synopsis    [Skips spaces.]

  Description [Returns 1 if reached the end.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertySkipSpaces(p *Amap_Tree_t, ppPos *string, pEnd string, fStopAtNewLine int32) int32 {
    var pPos string  = *ppPos;
    for  ; pPos < pEnd; pPos++  {
        if *pPos == '\n' {
            p.nLines++;
            if fStopAtNewLine != 0 {
                break; }
        }        
        if !(Amap_LibertyCharIsSpace(*pPos) != 0) {
            break; }
    }
    *ppPos = pPos;
    return pPos == pEnd;
}

/**Function*************************************************************

  Synopsis    [Skips entry delimited by " :;(){}" ]

  Description [Returns 1 if reached the end.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertySkipEntry(ppPos *string, pEnd string) int32 {
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

/**Function*************************************************************

  Synopsis    [Find the matching closing symbol.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyFindMatch(pPos string, pEnd string) string {
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

/**Function*************************************************************

  Synopsis    [Find the matching closing symbol.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyUpdateHead(p *Amap_Tree_t, Head Amap_Pair_t) Amap_Pair_t {
    var Res Amap_Pair_t;
    var pBeg string  = p.pContents + Head.Beg;
    var pEnd string  = p.pContents + Head.End;
    var pFirstNonSpace string  = nil;
    var pLastNonSpace string  = nil;
    var pChar string;
    for  pChar = pBeg; pChar < pEnd; pChar++  {
        if *pChar == '\n' {
            p.nLines++; }
        if Amap_LibertyCharIsSpace(*pChar) != 0 {
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

/**Function*************************************************************

  Synopsis    [Returns free item.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyNewItem(p *Amap_Tree_t, Type int32) *Amap_Item_t {
    p.pItems[p.nItems].iLine = p.nLines;
    p.pItems[p.nItems].Type  = Type;
    p.pItems[p.nItems].Child = -1;
    p.pItems[p.nItems].Next  = -1;
    return p.pItems + p.nItems++;
}

/**Function*************************************************************

  Synopsis    [Returns free item.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyBuildItem(p *Amap_Tree_t, ppPos *string, pEnd string) int32 {
    var pItem *Amap_Item_t;
    var Key, Head, Body Amap_Pair_t;
    var pNext, pStop string;
    Key.End = 0;
    if Amap_LibertySkipSpaces( p, ppPos, pEnd, 0 ) != 0 {
        return -2; }
    Key.Beg = *ppPos - p.pContents;
    if Amap_LibertySkipEntry( ppPos, pEnd ) != 0 {
        goto exit; }
    Key.End = *ppPos - p.pContents;
    if Amap_LibertySkipSpaces( p, ppPos, pEnd, 0 ) != 0 {
        goto exit; }
    pNext = *ppPos;
    if *pNext == ':' {
        *ppPos = pNext + 1;
        if Amap_LibertySkipSpaces( p, ppPos, pEnd, 0 ) != 0 {
            goto exit; }
        Head.Beg = *ppPos - p.pContents;
        if Amap_LibertySkipEntry( ppPos, pEnd ) != 0 {
            goto exit; }
        Head.End = *ppPos - p.pContents;
        if Amap_LibertySkipSpaces( p, ppPos, pEnd, 1 ) != 0 {
            goto exit; }
        pNext = *ppPos;
        if *pNext != ';' && *pNext != '\n' {
            goto exit; }
        *ppPos = pNext + 1;
        // end of equation
        pItem = Amap_LibertyNewItem( p, AMAP_LIBERTY_EQUA );
        pItem.Key  = Key;
        pItem.Head = Amap_LibertyUpdateHead( p, Head );
        pItem.Next = Amap_LibertyBuildItem( p, ppPos, pEnd );
        if pItem.Next == -1 {
            goto exit; }
        return Amap_LibertyItemId( p, pItem );
    }
    if *pNext == '(' {
        pStop = Amap_LibertyFindMatch( pNext, pEnd );
        Head.Beg = pNext - p.pContents + 1;
        Head.End = pStop - p.pContents;
        *ppPos = pStop + 1;
        if Amap_LibertySkipSpaces( p, ppPos, pEnd, 0 ) != 0 {
            // end of list
            pItem = Amap_LibertyNewItem( p, AMAP_LIBERTY_LIST );
            pItem.Key  = Key;
            pItem.Head = Amap_LibertyUpdateHead( p, Head );
            return Amap_LibertyItemId( p, pItem );
        }
        pNext = *ppPos;
        if *pNext == '{' { // beginning of body
        
            pStop = Amap_LibertyFindMatch( pNext, pEnd );
            Body.Beg = pNext - p.pContents + 1;
            Body.End = pStop - p.pContents;
            // end of body
            pItem = Amap_LibertyNewItem( p, AMAP_LIBERTY_PROC );
            pItem.Key  = Key;
            pItem.Head = Amap_LibertyUpdateHead( p, Head );
            pItem.Body = Body;
            *ppPos = pNext + 1;
            pItem.Child = Amap_LibertyBuildItem( p, ppPos, pStop );
            if pItem.Child == -1 {
                goto exit; }
            *ppPos = pStop + 1;
            pItem.Next = Amap_LibertyBuildItem( p, ppPos, pEnd );
            if pItem.Next == -1 {
                goto exit; }
            return Amap_LibertyItemId( p, pItem );
        }
        // end of list
        if *pNext == ';' {
            *ppPos = pNext + 1; }
        pItem = Amap_LibertyNewItem( p, AMAP_LIBERTY_LIST );
        pItem.Key  = Key;
        pItem.Head = Amap_LibertyUpdateHead( p, Head );
        pItem.Next = Amap_LibertyBuildItem( p, ppPos, pEnd );
        if pItem.Next == -1 {
            goto exit; }
        return Amap_LibertyItemId( p, pItem );
    }
exit:
    if p.pError == nil {
        // inlining of macro ABC_ALLOC
        p.pError = ((char *) malloc(sizeof(char) * (size_t)(1000)));
        sprintf( p.pError, "File \"%s\". Line %6d. Failed to parse entry \"%s\".\n", 
                    p.pFileName, p.nLines, Amap_LibertyGetString(p, Key) );
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Starts the parsing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyStart(pFileName string) *Amap_Tree_t {
    var pFile *os.File;
    var p *Amap_Tree_t;
    var RetValue int32;
    // start the manager
    // inlining of macro ABC_ALLOC
    p = ((Amap_Tree_t *) malloc(sizeof(Amap_Tree_t) * (size_t)(1)));
    *p = Amap_Tree_t{};
    // read the file into the buffer
    Amap_LibertyFixFileHead( pFileName );
    p.nContents = Amap_LibertyFileSize( pFileName );
    if p.nContents == 0 {
        p = nil;
        return nil;
    }
    pFile = os.Open(pFileName);
    // inlining of macro ABC_ALLOC
    p.pContents = ((char *) malloc(sizeof(char) * (size_t)(p.nContents+1)));
    RetValue = fread( p.pContents, p.nContents, 1, pFile );
    _ = (pFile).Close();
    p.pContents[p.nContents] = 0;
    // other 
    p.pFileName = Abc_UtilStrsav( pFileName );
    p.nItermAlloc = 10 + Amap_LibertyCountItems( p.pContents, p.pContents+p.nContents );
    // inlining of macro ABC_CALLOC
    p.pItems = ((Amap_Item_t *) calloc((size_t)(p.nItermAlloc), sizeof(Amap_Item_t)));
    p.nItems = 0;
    p.nLines = 1;
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the parsing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyStop(p *Amap_Tree_t) {
    p.pFileName = nil;
    p.pContents = nil;
    p.pItems = nil;
    p.pError = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Parses the standard cell library in Liberty format.]

  Description [Writes the resulting file in Genlib format.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyParse(pFileName string, fVerbose int32) int32 {
    var p *Amap_Tree_t;
    var pPos string;
    var clk abctime  = Abc_Clock();
    var RetValue int32;
    p = Amap_LibertyStart( pFileName );
    if p == nil {
        return 0; }
    pPos = p.pContents;
    Amap_LibertyWipeOutComments( p.pContents, p.pContents+p.nContents );
    if Amap_LibertyBuildItem( p, &pPos, p.pContents + p.nContents ) == 0 {
        if fVerbose != 0 {
        fmt.Printf("Parsing finished successfully.\n"); }
//        Amap_LibertyPrintLiberty( p, "temp_.lib" );
        Amap_LibertyPrintGenlib( p, Extra_FileNameGenericAppend(pFileName, ".genlib"), fVerbose );
        RetValue = 1;
    
    } else {
        if p.pError {
            fmt.Printf("%s", p.pError); }
        if fVerbose != 0 {
        fmt.Printf("Parsing failed.\n"); }
        RetValue = 0;
    }
    if fVerbose != 0 {
    fmt.Printf("Memory = %7.2f MB. ", 1.0*(p.nContents+p.nItermAlloc*sizeof(Amap_Item_t))/(1<<20));
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    Amap_LibertyStop( p );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Parses the standard cell library in Liberty format.]

  Description [Writes the resulting file in Genlib format.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibertyParseStr(pFileName string, fVerbose int32) *Vec_Str_t {
    var p *Amap_Tree_t;
    var vStr *Vec_Str_t  = nil;
    var pPos string;
    var clk abctime  = Abc_Clock();
    var RetValue int32;
    p = Amap_LibertyStart( pFileName );
    if p == nil {
        return 0; }
    pPos = p.pContents;
    Amap_LibertyWipeOutComments( p.pContents, p.pContents+p.nContents );
    if Amap_LibertyBuildItem( p, &pPos, p.pContents + p.nContents ) == 0 {
        if fVerbose != 0 {
        fmt.Printf("Parsing finished successfully.\n"); }
//        Amap_LibertyPrintLiberty( p, "temp_.lib" );
        vStr = Amap_LibertyPrintGenlibStr( p, fVerbose );
        RetValue = 1;
    
    } else {
        if p.pError {
            fmt.Printf("%s", p.pError); }
        if fVerbose != 0 {
        fmt.Printf("Parsing failed.\n"); }
        RetValue = 0;
    }
    if fVerbose != 0 {
    fmt.Printf("Memory = %7.2f MB. ", 1.0*(p.nContents+p.nItermAlloc*sizeof(Amap_Item_t))/(1<<20));
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    Amap_LibertyStop( p );
    return vStr;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

