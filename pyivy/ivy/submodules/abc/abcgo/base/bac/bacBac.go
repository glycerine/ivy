package bac

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [bacBac.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Verilog parser.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacBac.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bac.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Read CBA.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func BacManReadBacLine(vOut *Vec_Str_t, pPos *int32, pBuffer string, pLimit string) int32 {
    var c byte;
    for  (c = Vec_StrEntry(vOut, (*pPos)++)) != '\n' && pBuffer < pLimit  {
        *pBuffer++ = c;
    }
    *pBuffer = 0;
    return pBuffer < pLimit;
}
func BacManReadBacNameAndNums(pBuffer string, Num1 *int32, Num2 *int32, Num3 *int32, Num4 *int32) int32 {
    *Num1 = *Num2 = *Num3 = *Num4 = -1;
    // read name
    for  *pBuffer && *pBuffer != ' '  {
        pBuffer++;
    }
    if !*pBuffer {
        return 0; }
    assert(*pBuffer == ' ', "*pBuffer == ' '");
    *pBuffer = 0;
    // read Num1
    *Num1 = strconv.Atoi(++pBuffer);
    for  *pBuffer && *pBuffer != ' '  {
        pBuffer++;
    }
    if !*pBuffer {
        return 0; }
    // read Num2
    assert(*pBuffer == ' ', "*pBuffer == ' '");
    *Num2 = strconv.Atoi(++pBuffer);
    for  *pBuffer && *pBuffer != ' '  {
        pBuffer++;
    }
    if !*pBuffer {
        return 1; }
    // read Num3
    assert(*pBuffer == ' ', "*pBuffer == ' '");
    *Num3 = strconv.Atoi(++pBuffer);
    for  *pBuffer && *pBuffer != ' '  {
        pBuffer++;
    }
    if !*pBuffer {
        return 1; }
    // read Num4
    assert(*pBuffer == ' ', "*pBuffer == ' '");
    *Num4 = strconv.Atoi(++pBuffer);
    return 1;
}
func Bac_ManReadBacVecStr(vOut *Vec_Str_t, pPos *int32, p *Vec_Str_t, nSize int32) {
    memcpy( Vec_StrArray(p), Vec_StrArray(vOut) + *pPos, (size_t)nSize );
    *pPos += nSize;
    p.nSize = nSize;
    assert(Vec_StrSize(p) == Vec_StrCap(p), "Vec_StrSize(p) == Vec_StrCap(p)");
}
func Bac_ManReadBacVecInt(vOut *Vec_Str_t, pPos *int32, p *Vec_Int_t, nSize int32) {
    memcpy( Vec_IntArray(p), Vec_StrArray(vOut) + *pPos, (size_t)nSize );
    *pPos += nSize;
    p.nSize = nSize / 4;
    assert(Vec_IntSize(p) == Vec_IntCap(p), "Vec_IntSize(p) == Vec_IntCap(p)");
}
func Bac_ManReadBacNtk(vOut *Vec_Str_t, pPos *int32, pNtk *Bac_Ntk_t) {
    var i, Type int32;
    //char * pName; int iObj, NameId;
    Bac_ManReadBacVecStr( vOut, pPos, &pNtk.vType,      Bac_NtkObjNumAlloc(pNtk) );
    Bac_ManReadBacVecInt( vOut, pPos, &pNtk.vFanin, 4 * Bac_NtkObjNumAlloc(pNtk) );
    Bac_ManReadBacVecInt( vOut, pPos, &pNtk.vInfo, 12 * Bac_NtkInfoNumAlloc(pNtk) );
    // inlining of macro Bac_NtkForEachObjType
    for  i = 0; (i < Bac_NtkObjNum(pNtk))  && (((Type) = Bac_ObjType(pNtk, i)), 1); i++  { if !(Type != 0) {} else {
        if Type == BAC_OBJ_PI {
            Vec_IntPush( &pNtk.vInputs, i ); }
        if Type == BAC_OBJ_PO {
            Vec_IntPush( &pNtk.vOutputs, i ); }
    }
    }
    assert(Bac_NtkPiNum(pNtk)  == Bac_NtkPiNumAlloc(pNtk), "Bac_NtkPiNum(pNtk)  == Bac_NtkPiNumAlloc(pNtk)");
    assert(Bac_NtkPoNum(pNtk)  == Bac_NtkPoNumAlloc(pNtk), "Bac_NtkPoNum(pNtk)  == Bac_NtkPoNumAlloc(pNtk)");
    assert(Bac_NtkObjNum(pNtk) == Bac_NtkObjNumAlloc(pNtk), "Bac_NtkObjNum(pNtk) == Bac_NtkObjNumAlloc(pNtk)");
    assert(Bac_NtkInfoNum(pNtk) == Bac_NtkInfoNumAlloc(pNtk), "Bac_NtkInfoNum(pNtk) == Bac_NtkInfoNumAlloc(pNtk)");
/*
    // read input/output/box names
    Bac_NtkForEachPiMain( pNtk, iObj, i )
    {
        pName = Vec_StrEntryP( vOut, Pos );     
        NameId = Abc_NamStrFindOrAdd( p->pStrs, pName, NULL );
        Pos += strlen(pName) + 1;
    }
    Bac_NtkForEachPoMain( pNtk, iObj, i )
    {
        pName = Vec_StrEntryP( vOut, Pos );     
        NameId = Abc_NamStrFindOrAdd( p->pStrs, pName, NULL );
        Pos += strlen(pName) + 1;
    }
    Bac_NtkForEachBox( pNtk, iObj )
    {
        pName = Vec_StrEntryP( vOut, Pos );     
        NameId = Abc_NamStrFindOrAdd( p->pStrs, pName, NULL );
        Pos += strlen(pName) + 1;
    }
*/
}
func Bac_ManReadBacInt(vOut *Vec_Str_t) *Bac_Man_t {
    var p *Bac_Man_t;
    var pNtk *Bac_Ntk_t;
    var Buffer [1000]byte = "#"; 
    var i, NameId, Pos, nNtks, Num1, Num2, Num3, Num4 int32;
    for  Buffer[0] == '#'  {
        if !(BacManReadBacLine(vOut, &Pos, Buffer, Buffer+1000) != 0) {
            return nil; }
    }
    if !(BacManReadBacNameAndNums(Buffer, &nNtks, &Num2, &Num3, &Num4) != 0) {
        return nil; }
    // start manager
    assert(nNtks > 0, "nNtks > 0");
    p = Bac_ManAlloc( Buffer, nNtks );
    // start networks
    // inlining of macro Bac_ManForEachNtk
    for  i = 1; (i <= Bac_ManNtkNum(p)) && (((pNtk) = Bac_ManNtk(p, i)), 1); i++  {
        if !(BacManReadBacLine(vOut, &Pos, Buffer, Buffer+1000) != 0) {
            Bac_ManFree( p );
            return nil;
        }
        if !(BacManReadBacNameAndNums(Buffer, &Num1, &Num2, &Num3, &Num4) != 0) {
            Bac_ManFree( p );
            return nil;
        }
        assert(Num1 >= 0 && Num2 >= 0 && Num3 >= 0, "Num1 >= 0 && Num2 >= 0 && Num3 >= 0");
        NameId = Abc_NamStrFindOrAdd( p.pStrs, Buffer, nil );
        Bac_NtkAlloc( pNtk, NameId, Num1, Num2, Num3 );
        Vec_IntFill( &pNtk.vInfo, 3 * Num4, -1 );
    }
    // read networks
    // inlining of macro Bac_ManForEachNtk
    for  i = 1; (i <= Bac_ManNtkNum(p)) && (((pNtk) = Bac_ManNtk(p, i)), 1); i++  { 
        Bac_ManReadBacNtk( vOut, &Pos, pNtk );
    }
    assert(Bac_ManNtkNum(p) == nNtks, "Bac_ManNtkNum(p) == nNtks");
    assert(Pos == Vec_StrSize(vOut), "Pos == Vec_StrSize(vOut)");
    return p;
}
func Bac_ManReadBac(pFileName string) *Bac_Man_t {
    var p *Bac_Man_t;
    var pFile *os.File;
    var vOut *Vec_Str_t;
    var nFileSize int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for reading.\n", pFileName);
        return nil;
    }
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile );  
    rewind( pFile ); 
    // load the contents
    vOut = Vec_StrAlloc( nFileSize );
    vOut.nSize = vOut.nCap;
    assert(nFileSize == Vec_StrSize(vOut), "nFileSize == Vec_StrSize(vOut)");
    nFileSize = fread( Vec_StrArray(vOut), 1, Vec_StrSize(vOut), pFile );
    assert(nFileSize == Vec_StrSize(vOut), "nFileSize == Vec_StrSize(vOut)");
    _ = (pFile).Close();
    // read the networks
    p = Bac_ManReadBacInt( vOut );
    if p != nil {
        // inlining of macro ABC_FREE
        ((p.pSpec) ? (free((char *) (p.pSpec)), (p.pSpec) = 0) : 0);
        p.pSpec = Abc_UtilStrsav( pFileName );
    }
    Vec_StrFree( vOut );
    return p;
}

/**Function*************************************************************

  Synopsis    [Write CBA.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_ManWriteBacNtk(vOut *Vec_Str_t, pNtk *Bac_Ntk_t) {
    //char * pName; int iObj, NameId;
    Vec_StrPushBuffer( vOut, (char *)Vec_StrArray(&pNtk.vType),       Bac_NtkObjNum(pNtk) );
    Vec_StrPushBuffer( vOut, (char *)Vec_IntArray(&pNtk.vFanin),  4 * Bac_NtkObjNum(pNtk) );
    Vec_StrPushBuffer( vOut, (char *)Vec_IntArray(&pNtk.vInfo),  12 * Bac_NtkInfoNum(pNtk) );
/*
    // write input/output/box names
    Bac_NtkForEachPiMain( pNtk, iObj, i )
    {
        pName = Bac_ObjNameStr( pNtk, iObj );
        Vec_StrPrintStr( vOut, pName );
        Vec_StrPush( vOut, '\0' );
    }
    Bac_NtkForEachPoMain( pNtk, iObj, i )
    {
        pName = Bac_ObjNameStr( pNtk, iObj );
        Vec_StrPrintStr( vOut, pName );
        Vec_StrPush( vOut, '\0' );
    }
    Bac_NtkForEachBox( pNtk, iObj )
    {
        pName = Bac_ObjNameStr( pNtk, iObj );  
        Vec_StrPrintStr( vOut, pName );
        Vec_StrPush( vOut, '\0' );
    }
*/
}
func Bac_ManWriteBacInt(vOut *Vec_Str_t, p *Bac_Man_t) {
    var Buffer string // was [1000]byte in C

    var pNtk *Bac_Ntk_t; var i int32;
    Buffer = fmt.Sprintf("# Design \"%s\" written by ABC on %s\n", Bac_ManName(p), Extra_TimeStamp());
    Vec_StrPrintStr( vOut, Buffer );
    // write short info
    Buffer = fmt.Sprintf("%s %d \n", Bac_ManName(p), Bac_ManNtkNum(p));
    Vec_StrPrintStr( vOut, Buffer );
    // inlining of macro Bac_ManForEachNtk
    for  i = 1; (i <= Bac_ManNtkNum(p)) && (((pNtk) = Bac_ManNtk(p, i)), 1); i++  {
        Buffer = fmt.Sprintf("%s %d %d %d %d \n", Bac_NtkName(pNtk), Bac_NtkPiNum(pNtk), Bac_NtkPoNum(pNtk), Bac_NtkObjNum(pNtk), Bac_NtkInfoNum(pNtk));
        Vec_StrPrintStr( vOut, Buffer );
    }
    // inlining of macro Bac_ManForEachNtk
    for  i = 1; (i <= Bac_ManNtkNum(p)) && (((pNtk) = Bac_ManNtk(p, i)), 1); i++  { 
        Bac_ManWriteBacNtk( vOut, pNtk );
    }
}
func Bac_ManWriteBac(pFileName string, p *Bac_Man_t) {
    var vOut *Vec_Str_t;
    assert(p.pMioLib == nil, "p.pMioLib == nil");
    vOut = Vec_StrAlloc( 10000 );
    Bac_ManWriteBacInt( vOut, p );
    if Vec_StrSize(vOut) > 0 {
        var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        } else {
            fwrite( Vec_StrArray(vOut), 1, Vec_StrSize(vOut), pFile );
            _ = (pFile).Close();
        }
    }
    Vec_StrFree( vOut );    
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

