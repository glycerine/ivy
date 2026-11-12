package io

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [ioJson.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to read JSON.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioJson.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "misc/vec/vecWec.h"
//#include "misc/util/utilNam.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Json_EntryIsName(Fan int32) int32 { return Abc_LitIsCompl(Fan);                                                       }
func Json_EntryName(pStrs *Abc_Nam_t, Fan int32) string { assert((Json_EntryIsName(Fan) != 0), "Json_EntryIsName(Fan)");  return Abc_NamStr( pStrs, Abc_Lit2Var(Fan) );     }
func Json_EntryNode(vObjs *Vec_Wec_t, Fan int32) *Vec_Int_t { assert(!(Json_EntryIsName(Fan) != 0), "!Json_EntryIsName(Fan)"); return Vec_WecEntry( vObjs, Abc_Lit2Var(Fan) );   }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Writes JSON into a file.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nnc_LayerType2Str(pStr string) string {
    if !(int32(strings.Compare(pStr, "InputLayer")) != 0) {
        return "input  "; }
    if !(int32(strings.Compare(pStr, "Conv2D")) != 0) {
        return "convo  "; }
    if !(int32(strings.Compare(pStr, "BatchNormalization")) != 0) {
        return "batch  "; }
    if !(int32(strings.Compare(pStr, "Activation")) != 0) {
        return "relu   "; }
    if !(int32(strings.Compare(pStr, "Add")) != 0) {
        return "eltwise"; }
    if !(int32(strings.Compare(pStr, "MaxPooling2D")) != 0) {
        return "pool   "; }
    if !(int32(strings.Compare(pStr, "GlobalAveragePooling2D")) != 0) {
        return "pool   "; }
    if !(int32(strings.Compare(pStr, "Dense")) != 0) {
        return "fullcon"; }
    if !(int32(strings.Compare(pStr, "ZeroPadding2D")) != 0) {
        return "pad"; }
//    if ( !strcmp(pStr, "InputLayer") )
//        return "softmax";
    return nil;
}

func Json_Extract_rec(pFile *os.File, pStr *Abc_Nam_t, vObjs *Vec_Wec_t, vArray *Vec_Int_t, fWrite int32, pCount *int32) {
    var i, Entry1, Entry2 int32;
    if Vec_IntEntry(vArray, 0) { // array
    
        if Vec_IntSize(vArray) == 1 {
            return; }
        if Vec_IntSize(vArray) == 2 && (Json_EntryIsName(Vec_IntEntry(vArray,1)) != 0) {
            if fWrite != 0 {
                fmt.Fprintf(pFile, "%s", Json_EntryName(pStr, Vec_IntEntry(vArray,1))); }
            return;
        
        } else {
            // inlining of macro Vec_IntForEachEntryStart
            for  i = 1; (i < Vec_IntSize(vArray)) && (((Entry1) = Vec_IntEntry(vArray, i)), 1); i++  {
                if Json_EntryIsName(Entry1) != 0 {
                    var Digit int32  = Json_EntryName(pStr, Entry1)[0];
                    if (fWrite != 0) && Digit != '0' {
                        fmt.Fprintf(pFile, "%s%s", Json_EntryName(pStr, Entry1), Digit >= '0' && Digit <= '9' ? "" : " "); }
                
                } else {
                    Json_Extract_rec( pFile, pStr, vObjs, Json_EntryNode(vObjs, Entry1), fWrite, pCount ); }
            }
            return;
        }
    
    } else { // list of pairs
    
        var fHaveConfig int32;
        assert(Vec_IntSize(vArray) % 2 != 0, "Vec_IntSize(vArray) % 2 != 0");
        // inlining of macro Vec_IntForEachEntryDoubleStart
        for  i = 1; (i+1 < Vec_IntSize(vArray)) && (((Entry1) = Vec_IntEntry(vArray, i)), 1) && (((Entry2) = Vec_IntEntry(vArray, i+1)), 1); i += 2  {
            var pName1 string;
            if Json_EntryIsName(Entry1) != 0 {
                pName1 = Json_EntryName(pStr, Entry1);
            } else {
                pName1 = nil;
            }
            var pName2 string;
            if Json_EntryIsName(Entry2) != 0 {
                pName2 = Json_EntryName(pStr, Entry2);
            } else {
                pName2 = nil;
            }
            var pName3 string;
            if pName2 {
                pName3 = Nnc_LayerType2Str(pName2);
            } else {
                pName3 = nil;
            }
            if pName1 == nil {
                continue; }
            if !(int32(strings.Compare(pName1, "class_name")) != 0) {
                if pName3 {
                    fmt.Fprintf(pFile, "\n%3d : %-8s ", (*pCount)++, pName3); }
            
            } else if !(int32(strings.Compare(pName1, "name")) != 0) {
                if fHaveConfig != 0 {
                    fmt.Fprintf(pFile, " N=%s  ", pName2 ? pName2 : "???"); }
            
            } else if !(int32(strings.Compare(pName1, "kernel_size")) != 0) {
                fmt.Fprintf(pFile, " K=");
                Json_Extract_rec( pFile, pStr, vObjs, Json_EntryNode(vObjs, Entry2), 1, pCount );
            
            } else if !(int32(strings.Compare(pName1, "strides")) != 0) {
                fmt.Fprintf(pFile, " S=");
                Json_Extract_rec( pFile, pStr, vObjs, Json_EntryNode(vObjs, Entry2), 1, pCount );
            
            } else if !(int32(strings.Compare(pName1, "filters")) != 0) {
                fmt.Fprintf(pFile, " C=%s", pName2);
            } else if !(int32(strings.Compare(pName1, "inbound_nodes")) != 0) {
                Json_Extract_rec( pFile, pStr, vObjs, Json_EntryNode(vObjs, Entry2), 1, pCount );
            } else if !(int32(strings.Compare(pName1, "layers")) != 0) {
                Json_Extract_rec( pFile, pStr, vObjs, Json_EntryNode(vObjs, Entry2), 1, pCount );
            } else if !(int32(strings.Compare(pName1, "config")) != 0) {
                fHaveConfig = 1;
                Json_Extract_rec( pFile, pStr, vObjs, Json_EntryNode(vObjs, Entry2), 0, pCount );
            }
        }
    }
}
func Json_Extract(pFileName string, pStr *Abc_Nam_t, vObjs *Vec_Wec_t) {
    var Count int32;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return;
    }
    fmt.Fprintf(pFile, "# Data extracted from JSON file:\n");
    Json_Extract_rec( pFile, pStr, vObjs, Vec_WecEntry(vObjs, 0), 0, &Count );
    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Parsing.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Json_CharIsSpace(c byte) int32 { 
    return (c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == ':');   
}
func Json_SkipSpaces(pCur string) string {
    for  Json_CharIsSpace(*pCur) != 0  {
        pCur++;
    }    
    return pCur;
}
func Json_SkipNonSpaces(pCur string) string {
    for  !(Json_CharIsSpace(*pCur) != 0)  {
        pCur++;
    }    
    return pCur;
}
func Json_TokenCompare(pCur string, pNext string, ppCur2 *string, ppNext2 *string) int32 {
//    int i;
    if *pCur == '\"' {
        pCur++; }
    if *(pNext-1) == ',' {
        pNext--; }
    if *(pNext-1) == '\"' {
        pNext--; }
    *ppCur2 = pCur;
    *ppNext2 = pNext;
//    for ( i = 1; i < JSON_NUM_LINES; i++ )
//        if ( !strncmp( s_Types[i].pName, pCur, pNext - pCur ) )
//            return i;
    return 0;
}

/**Function*************************************************************

  Synopsis    [Writes JSON into a file.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Json_Write_rec(pFile *os.File, pStr *Abc_Nam_t, vObjs *Vec_Wec_t, vArray *Vec_Int_t, Level int32, fAddComma int32, fSpaces int32) {
    var i, Entry1, Entry2, fComma int32;
    if Vec_IntEntry(vArray, 0) { // array
    
        if Vec_IntSize(vArray) == 1 {
            fmt.Fprintf(pFile, "[]");
        } else if Vec_IntSize(vArray) == 2 && (Json_EntryIsName(Vec_IntEntry(vArray,1)) != 0) {
            fmt.Fprintf(pFile, "[ \"%s\" ]", Json_EntryName(pStr, Vec_IntEntry(vArray,1)));
        } else {
            if fSpaces != 0 {
                fmt.Fprintf(pFile, "%*s", 3*(Level-1), ""); }
            fmt.Fprintf(pFile, "[\n");
            // inlining of macro Vec_IntForEachEntryStart
            for  i = 1; (i < Vec_IntSize(vArray)) && (((Entry1) = Vec_IntEntry(vArray, i)), 1); i++  {
                fComma = (i < Vec_IntSize(vArray) - 1);
                if Json_EntryIsName(Entry1) != 0 {
                    fmt.Fprintf(pFile, "%*s\"%s\"%s\n", 3*Level, "", Json_EntryName(pStr, Entry1), fComma ? ",":"");
                } else {
                    Json_Write_rec( pFile, pStr, vObjs, Json_EntryNode(vObjs, Entry1), Level+1, fComma, 1 ); }
            }
            fmt.Fprintf(pFile, "%*s]", 3*(Level-1), "");
        }
        fmt.Fprintf(pFile, "%s\n", fAddComma ? ",":"");
    
    } else { // list of pairs
    
        if fSpaces != 0 {
            fmt.Fprintf(pFile, "%*s", 3*(Level-1), ""); }
        fmt.Fprintf(pFile, "{\n");
        assert(Vec_IntSize(vArray) % 2 != 0, "Vec_IntSize(vArray) % 2 != 0");
        // inlining of macro Vec_IntForEachEntryDoubleStart
        for  i = 1; (i+1 < Vec_IntSize(vArray)) && (((Entry1) = Vec_IntEntry(vArray, i)), 1) && (((Entry2) = Vec_IntEntry(vArray, i+1)), 1); i += 2  {
            fComma = (i < Vec_IntSize(vArray) - 3);
            if Json_EntryIsName(Entry1) != 0 {
                fmt.Fprintf(pFile, "%*s\"%s\"", 3*Level, "", Json_EntryName(pStr, Entry1));
            } else {
                Json_Write_rec( pFile, pStr, vObjs, Json_EntryNode(vObjs, Entry1), Level+1, 0, 1 ); }
            fmt.Fprintf(pFile, " : ");
            if Json_EntryIsName(Entry2) != 0 {
                fmt.Fprintf(pFile, "\"%s\"%s\n", Json_EntryName(pStr, Entry2), fComma ? ",":"");
            } else {
                Json_Write_rec( pFile, pStr, vObjs, Json_EntryNode(vObjs, Entry2), Level+1, fComma, 0 ); }
        }
        fmt.Fprintf(pFile, "%*s}%s\n", 3*(Level-1), "", fAddComma ? ",":"");
    }
}
func Json_Write(pFileName string, pStr *Abc_Nam_t, vObjs *Vec_Wec_t) {
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return;
    }
    Json_Write_rec( pFile, pStr, vObjs, Vec_WecEntry(vObjs, 0), 1, 0, 1 );
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Reads JSON from a file.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Json_ReadPreprocess(pIn string, nFileSize int32) string {
    // inlining of macro ABC_ALLOC
    var pOut string  = ((char *) malloc(sizeof(char) * (size_t)(3*nFileSize))); var i, k int32;
    for  i = 0; i < nFileSize; i++  {
        if pIn[i] == '{' || pIn[i] == '}' || pIn[i] == '[' || pIn[i] == ']' {
            pOut[k++] = ' ';
            pOut[k++] = pIn[i];
            pOut[k++] = ' ';
        
        } else {
            pOut[k++] = pIn[i]; }
    }
    pOut[k++] = '\0';
    return pOut;
}
func Json_Read(pFileName string, ppStrs **Abc_Nam_t) *Vec_Wec_t {
    var pStrs *Abc_Nam_t; 
    var vObjs *Vec_Wec_t; 
    var vStack, vTemp *Vec_Int_t;
    var pContents, pCur, pNext, pCur2, pNext2 string;
    var nFileSize, RetValue, iToken int32;
    var pFile *os.File;

    // read the file into the buffer
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for reading.\n", pFileName);
        return nil;
    }
    nFileSize = Extra_FileSize( pFileName );
    // inlining of macro ABC_ALLOC
    pContents = pCur = ((char *) malloc(sizeof(char) * (size_t)(nFileSize+1)));
    RetValue = fread( pContents, nFileSize, 1, pFile );
    pContents[nFileSize] = 0;
    _ = (pFile).Close();

    pContents = Json_ReadPreprocess( pCur = pContents, nFileSize );
    nFileSize = len(pContents);
    // inlining of macro ABC_FREE
    ((pCur) ? (free((char *) (pCur)), (pCur) = 0) : 0);
    pCur = pContents;

    // start data-structures
    vObjs  = Vec_WecAlloc( 1000 );
    vStack = Vec_IntAlloc( 100 );
    pStrs  = Abc_NamStart( 1000, 24 );
    //Json_AddTypes( pStrs );

    // read lines
    assert(Vec_WecSize(vObjs) == 0, "Vec_WecSize(vObjs) == 0");
    assert(Vec_IntSize(vStack) == 0, "Vec_IntSize(vStack) == 0");
    for  pCur < pContents + nFileSize  {
        pCur  = Json_SkipSpaces( pCur );
        if *pCur == '\0' {
            break; }
        pNext = Json_SkipNonSpaces( pCur );
        if *pCur == '{' || *pCur == '[' {
            // add fanin to node on the previous level
            if Vec_IntSize(vStack) > 0 {
                Vec_IntPush( Vec_WecEntry(vObjs, Vec_IntEntryLast(vStack)), Abc_Var2Lit(Vec_WecSize(vObjs), 0) ); }            
            // add node to the stack
            Vec_IntPush( vStack, Vec_WecSize(vObjs) );
            vTemp = Vec_WecPushLevel( vObjs );
            Vec_IntGrow( vTemp, 4 );
            // remember it as an array
            Vec_IntPush( vTemp, int32((*pCur == '[')) );
            pCur++;
            continue;
        }
        if *pCur == '}' || *pCur == ']' {
            Vec_IntPop( vStack );
            pCur++;
            continue;
        }
        if *pCur == ',' || *pCur == ':' {
            pCur++;
            continue;
        }
        iToken = Json_TokenCompare( pCur, pNext, &pCur2, &pNext2 );
        if iToken == 0 {
            iToken = Abc_NamStrFindOrAddLim( pStrs, pCur2, pNext2, nil ); }
        Vec_IntPush( Vec_WecEntry(vObjs, Vec_IntEntryLast(vStack)), Abc_Var2Lit(iToken, 1) );
        pCur = pNext;
    }
    Vec_IntFree( vStack );
    // inlining of macro ABC_FREE
    ((pContents) ? (free((char *) (pContents)), (pContents) = 0) : 0);
    *ppStrs = pStrs;
    return vObjs;
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Json_ReadTest(pFileName string) {
    var pStrs *Abc_Nam_t;
    var vObjs *Vec_Wec_t;
    vObjs = Json_Read( pFileName, &pStrs );
    if vObjs == nil {
        return; }
    Json_Write( "test.json", pStrs, vObjs );
    Abc_NamDeref( pStrs );
    Vec_WecFree( vObjs );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

