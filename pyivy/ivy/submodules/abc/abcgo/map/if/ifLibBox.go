package if_

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [ifLibBox.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Box library.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifLibBox.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"
//#include "misc/extra/extra.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define If_LibBoxForEachBox( p, pBox, i )     \
//     Vec_PtrForEachEntry( If_Box_t *, p->vBoxes, pBox, i ) if ( pBox == NULL ) {} else


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_BoxStart(pName string, Id int32, nPis int32, nPos int32, fSeq int32, fBlack int32, fOuter int32) *If_Box_t {
    
    // inlining of macro ABC_CALLOC
    p := &If_Box_t{};
    p.pName   = pName; // consumes memory
    p.Id      = Id;
    p.fSeq    = byte(fSeq);
    p.fBlack  = byte(fBlack);
    p.fOuter  = byte(fOuter);
    p.nPis    = nPis;
    p.nPos    = nPos;
    // inlining of macro ABC_CALLOC
    p.pDelays = ((int *) calloc((size_t)(nPis * nPos), sizeof(int)));
    return p;
}
func If_BoxDup(p *If_Box_t) *If_Box_t {
    var pNew *If_Box_t  = nil;
    return pNew;
}
func If_BoxFree(p *If_Box_t) {
    // inlining of macro ABC_FREE
    ((p.pDelays) ? (free((char *) (p.pDelays)), (p.pDelays) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pName) ? (free((char *) (p.pName)), (p.pName) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_LibBoxStart() *If_LibBox_t {
    
    // inlining of macro ABC_CALLOC
    p := &If_LibBox_t{};
    p.vBoxes = Vec_PtrAlloc( 100 );
    return p;
}
func If_LibBoxDup(p *If_Box_t) *If_LibBox_t {
    var pNew *If_LibBox_t  = nil;
    return pNew;
}
func If_LibBoxFree(p *If_LibBox_t) {
    var pBox *If_Box_t;
    var i int32;
    if p == nil {
        return; }
    // inlining of macro If_LibBoxForEachBox
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (If_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  { if pBox == nil {} else {
        If_BoxFree( pBox ); }
    }
    Vec_PtrFree( p.vBoxes );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_LibBoxReadBox(p *If_LibBox_t, Id int32) *If_Box_t {
    return (If_Box_t *)Vec_PtrEntry( p.vBoxes, Id );
}
func If_LibBoxFindBox(p *If_LibBox_t, pName string) *If_Box_t {
    var pBox *If_Box_t;
    var i int32;
    if p == nil {
        return nil; }
    // inlining of macro If_LibBoxForEachBox
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (If_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  { if pBox == nil {} else
        if !(int32(strings.Compare(pBox.pName, pName)) != 0) {
            return pBox; }
    }
    return nil;
}
func If_LibBoxAdd(p *If_LibBox_t, pBox *If_Box_t) {
    if pBox.Id >= Vec_PtrSize(p.vBoxes) {
        Vec_PtrFillExtra( p.vBoxes, 2 * pBox.Id + 10, nil ); }
    assert(Vec_PtrEntry( p.vBoxes, pBox.Id ) == nil, "Vec_PtrEntry( p.vBoxes, pBox.Id ) == nil");
    Vec_PtrWriteEntry( p.vBoxes, pBox.Id, pBox );
    p.nBoxes++;
}
func If_LibBoxNum(p *If_LibBox_t) int32 {
    return p.nBoxes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_LibBoxRead2(pFileName string) *If_LibBox_t {
    var nSize int32  = 100000;
    var pBuffer string;
    var pFile *os.File;
    var p *If_LibBox_t  = nil;
    var pBox *If_Box_t  = nil;
    var pToken, pName string;
    var fSeq, fBlack, fOuter int32;
    var i, Id, nPis, nPos int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\".\n", pFileName);
        return nil;
    }
    // read lines
    nPis = nPos = 0;
    // inlining of macro ABC_ALLOC
    pBuffer = ((char *) malloc(sizeof(char) * (size_t)(nSize)));
    for  fgets( pBuffer, nSize, pFile )  {
        pToken = strtok( pBuffer, " \n\r\t" );
        if pToken == nil {
            continue; }
        if pToken[0] == '.' {
            if !(int32(strings.Compare(pToken, ".box")) != 0) {
                // save ID
                pToken = strtok( nil, " \n\r\t" );
                Id     = strconv.Atoi(pToken);
                // save name
                pToken = strtok( nil, " \n\r\t" );
                pName  = Abc_UtilStrsav(pToken);
                // save PIs
                pToken = strtok( nil, " \n\r\t" );
                nPis   = strconv.Atoi(pToken);
                // save POs
                pToken = strtok( nil, " \n\r\t" );
                nPos   = strconv.Atoi(pToken);
                // save attributes
                fSeq = fBlack = fOuter = 0;
                pToken = strtok( nil, " \n\r\t" );
                for  pToken  {
                    if !(int32(strings.Compare(pToken, "seq")) != 0) {
                        fSeq = 1;
                    } else if !(int32(strings.Compare(pToken, "black")) != 0) {
                        fBlack = 1;
                    } else if !(int32(strings.Compare(pToken, "outer")) != 0) {
                        fOuter = 1;
                    } else { assert(!(int32(strings.Compare(pToken, "comb")) != 0) || !(int32(strings.Compare(pToken, "white")) != 0) || !(int32(strings.Compare(pToken, "inner")) != 0), "!strcmp(pToken, \"comb\") || !strcmp(pToken, \"white\") || !strcmp(pToken, \"inner\")"); }
                    pToken = strtok( nil, " \n\r\t" );
                }
                // create library
                if p == nil {
                    p = If_LibBoxStart(); }
                // create box
                pBox = If_BoxStart( pName, Id, nPis, nPos, fSeq, fBlack, fOuter );
                If_LibBoxAdd( p, pBox );
            }
            continue;
        }
        // read the table
        assert(nPis > 0 && nPos > 0, "nPis > 0 && nPos > 0");
        for  i = 0; i < nPis * nPos; i++  {
            for  pToken == nil  {
                if fgets( pBuffer, nSize, pFile ) == nil { fmt.Printf("The table does not have enough entries.\n"); fflush(os.Stdout); assert(false, "0"); }
                pToken = strtok( pBuffer, " \n\r\t" );
            }
            pBox.pDelays[i] = (pToken[0] == '-') ? -1 : strconv.Atoi(pToken);
            pToken = strtok( nil, " \n\r\t" );
        }
        pBox = nil;
    }
    // inlining of macro ABC_FREE
    ((pBuffer) ? (free((char *) (pBuffer)), (pBuffer) = 0) : 0);
    _ = (pFile).Close();
    return p;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_If_LibBoxGetToken_pBuffer [1000]byte;
func If_LibBoxGetToken(pFile *os.File) string {
    
    var c int32; var pTemp string  = spStatic_If_LibBoxGetToken_pBuffer;
    for  (c = fgetc(pFile)) != -1  {
        if c == '#' {
            for  (c = fgetc(pFile)) != -1  {
                if c == '\n' {
                    break; }
            }
        }
        if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
            if pTemp > spStatic_If_LibBoxGetToken_pBuffer {
                break; }
            continue;
        }
        *pTemp++ = c;
    }
    *pTemp = 0;
    if pTemp > spStatic_If_LibBoxGetToken_pBuffer {
        return spStatic_If_LibBoxGetToken_pBuffer;
    } else {
        return nil;
    }
}
func If_LibBoxRead(pFileName string) *If_LibBox_t {
    var pFile *os.File;
    var p *If_LibBox_t;
    var pBox *If_Box_t;
    var pToken, pName string;
    var i, Id, fBlack, nPis, nPos int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\".\n", pFileName);
        return nil;
    }
    // get the library name
    pToken = If_LibBoxGetToken( pFile );
    if pToken == nil {
        _ = (pFile).Close();
        fmt.Printf("Cannot read library name from file \"%s\".\n", pFileName);
        return nil;
    }
    if pToken[0] == '.' {
        _ = (pFile).Close();
        fmt.Printf("Wrong box format. Please try \"read_box -e\".\n");
        return nil;
    }
       
    // create library
    p = If_LibBoxStart();
    for  pToken  {
        // save name
        pName  = Abc_UtilStrsav(pToken);
        // save ID
        pToken = If_LibBoxGetToken( pFile );
        Id     = strconv.Atoi(pToken);
        // save white/black
        pToken = If_LibBoxGetToken( pFile );
        fBlack = !strconv.Atoi(pToken);
        // save PIs
        pToken = If_LibBoxGetToken( pFile );
        nPis   = strconv.Atoi(pToken);
        // save POs
        pToken = If_LibBoxGetToken( pFile );
        nPos   = strconv.Atoi(pToken);
        // create box
        pBox   = If_BoxStart( pName, Id, nPis, nPos, 0, fBlack, 0 );
        If_LibBoxAdd( p, pBox );
        // read the table
        for  i = 0; i < nPis * nPos; i++  {
            pToken = If_LibBoxGetToken( pFile );
            pBox.pDelays[i] = (pToken[0] == '-') ? -ABC_INFINITY : strconv.Atoi(pToken);
        }
        // extract next name
        pToken = If_LibBoxGetToken( pFile );
    }
    _ = (pFile).Close();
    return p;
}
func If_LibBoxPrint(pFile *os.File, p *If_LibBox_t) {
    var pBox *If_Box_t;
    var i, j, k int32;
    fmt.Fprintf(pFile, "# Box library written by ABC on %s.\n", Extra_TimeStamp());
    fmt.Fprintf(pFile, "# <Name> <ID> <Type> <I> <O>\n");
    // inlining of macro If_LibBoxForEachBox
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (If_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  { if pBox == nil {} else {
        fmt.Fprintf(pFile, "%s %d %d %d %d\n", pBox.pName, pBox.Id, !pBox.fBlack, pBox.nPis, pBox.nPos);
        for  j = 0; j < pBox.nPos; j++, printf("\n")  {
            for  k = 0; k < pBox.nPis; k++  {
                if pBox.pDelays[j * pBox.nPis + k] == -ABC_INFINITY {
                    fmt.Fprintf(pFile, "    - ");
                } else {
                    fmt.Fprintf(pFile, "%5d ", pBox.pDelays[j * pBox.nPis + k]); }
            }
        }
    }
    }
}
func If_LibBoxWrite(pFileName string, p *If_LibBox_t) {
    var pFile *os.File;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\".\n", pFileName);
        return;
    }
    If_LibBoxPrint( pFile, p );
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_LibBoxLoad(pFileName string) int32 {
    var pFile *os.File;
    var pLib *If_LibBox_t;
    var pFileNameOther string;
    // check if library can be read
    pFileNameOther = Extra_FileNameGenericAppend( pFileName, ".cdl" );
    pFile = os.Open(pFileNameOther);
    if pFile == nil {
        return 0; }
    _ = (pFile).Close();
    // read library
    pLib = If_LibBoxRead2( pFileNameOther );
    // replace the current library
    If_LibBoxFree( (If_LibBox_t *)Abc_FrameReadLibBox() );
    Abc_FrameSetLibBox( pLib );
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

