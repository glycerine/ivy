package util

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [utilBridge.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName []

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: utilBridge.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <string.h>
//#include <stdlib.h>
//#include <assert.h>

//#include <misc/util/abc_global.h>

#if defined(LIN) || defined(LIN64)
#include <unistd.h>
#endif

//#include "aig/gia/gia.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const BRIDGE_TEXT_MESSAGE = 999996

const BRIDGE_ABORT =        5
const BRIDGE_PROGRESS =     3
const BRIDGE_RESULTS =      101
const BRIDGE_BAD_ABS =      105
//#define BRIDGE_NETLIST      106
//#define BRIDGE_ABS_NETLIST  107

const BRIDGE_VALUE_X = 0
const BRIDGE_VALUE_0 = 2
const BRIDGE_VALUE_1 = 3


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManToBridgeVec(p *Gia_Man_t) *Vec_Str_t {
    var vStr *Vec_Str_t;
    var pObj *Gia_Obj_t;
    var i, uLit0, uLit1, nNodes int32;
    assert(Gia_ManPoNum(p) > 0, "Gia_ManPoNum(p) > 0");

    // start with const1 node (number 1)
    nNodes = 1;
    // assign literals(!!!) to the value field
    Gia_ManConst0(p).Value = Abc_Var2Lit( nNodes++, 1 );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Abc_Var2Lit( nNodes++, 0 );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Abc_Var2Lit( nNodes++, 0 ); }
    }

    // write header
    vStr = Vec_StrAlloc( 1000 );
    Gia_AigerWriteUnsigned( vStr, Gia_ManPiNum(p) );
    Gia_AigerWriteUnsigned( vStr, Gia_ManRegNum(p) );
    Gia_AigerWriteUnsigned( vStr, Gia_ManAndNum(p) );

    // write the nodes 
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        uLit0 = Gia_ObjFanin0Copy( pObj );
        uLit1 = Gia_ObjFanin1Copy( pObj );
        assert(uLit0 != uLit1, "uLit0 != uLit1");
        Gia_AigerWriteUnsigned( vStr, uLit0 << 1 );
        Gia_AigerWriteUnsigned( vStr, uLit1 );
    }
    }

    // write latch drivers
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        uLit0 = Gia_ObjFanin0Copy( pObj );
        Gia_AigerWriteUnsigned( vStr, (uLit0 << 2) | BRIDGE_VALUE_0 );
    }

    // write PO drivers
    Gia_AigerWriteUnsigned( vStr, Gia_ManPoNum(p) );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        uLit0 = Gia_ObjFanin0Copy( pObj );
        // complement property output!!!
        Gia_AigerWriteUnsigned( vStr, Abc_LitNot(uLit0) );
    }
    return vStr;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
void Gia_CreateHeader( FILE * pFile, int Type, int Size, unsigned char * pBuffer )
{
    fmt.Fprintf(pFile, "%.6d", Type);
    fmt.Fprintf(pFile, " ");
    fmt.Fprintf(pFile, "%.16d", Size);
    fmt.Fprintf(pFile, " ");
#if !defined(LIN) && !defined(LIN64)
    {
    int RetValue;
    RetValue = fwrite( pBuffer, Size, 1, pFile );
    assert( RetValue == 1 || Size == 0);
    fflush( pFile );
    }
#else
    fflush(pFile);
    int fd = fileno(pFile);

    ssize_t bytes_written = 0;
    while (bytes_written < Size){
        ssize_t n = write(fd, &pBuffer[bytes_written], Size - bytes_written);
        if (n < 0){
            fprintf(stderr, "BridgeMode: failed to send package; aborting\n"); fflush(stderr);
            _exit(255);
        }
        bytes_written += n;
    }
#endif
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Gia_ManToBridgeText( FILE * pFile, int Size, unsigned char * pBuffer )
{
    Gia_CreateHeader( pFile, BRIDGE_TEXT_MESSAGE, Size, pBuffer );
    return 1;
}


int Gia_ManToBridgeAbort( FILE * pFile, int Size, unsigned char * pBuffer )
{
    Gia_CreateHeader( pFile, BRIDGE_ABORT, Size, pBuffer );
    return 1;
}


int Gia_ManToBridgeProgress( FILE * pFile, int Size, unsigned char * pBuffer )
{
    Gia_CreateHeader( pFile, BRIDGE_PROGRESS, Size, pBuffer );
    return 1;
}


func Gia_ManToBridgeAbsNetlist(pFile *os.File, p any, pkg_type int32) int32 {
    var vBuffer *Vec_Str_t;
    vBuffer = Gia_ManToBridgeVec( (Gia_Man_t *)p );
    Gia_CreateHeader( pFile, pkg_type, Vec_StrSize(vBuffer), (unsigned char *)Vec_StrArray(vBuffer) );
    Vec_StrFree( vBuffer );
    return 1;
}


func Gia_ManToBridgeBadAbs(pFile *os.File) int32 {
    Gia_CreateHeader( pFile, BRIDGE_BAD_ABS, 0, nil );
    return 1;
}


func aigerNumSize(x uint32) int32 {
    var sz int32  = 1;
    for x & ~0x7f {
        sz++;
        x >>= 7;
    }
    return sz;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFromBridgeHolds(pFile *os.File, iPoProved int32) {
    fmt.Fprintf(pFile, "%.6d", 101 /*message type = Result*/);
    fmt.Fprintf(pFile, " ");
    fmt.Fprintf(pFile, "%.16d", 3 + aigerNumSize(iPoProved) /*size in bytes*/);
    fmt.Fprintf(pFile, " ");

    fputc( byte(BRIDGE_VALUE_1), pFile ); // true
    fputc( byte(1), pFile ); // size of vector (Armin's encoding)
    Gia_AigerWriteUnsignedFile( pFile, iPoProved ); // number of the property (Armin's encoding)
    fputc( byte(0), pFile ); // no invariant
    fflush(pFile);
}
func Gia_ManFromBridgeUnknown(pFile *os.File, iPoUnknown int32) {
    fmt.Fprintf(pFile, "%.6d", 101 /*message type = Result*/);
    fmt.Fprintf(pFile, " ");
    fmt.Fprintf(pFile, "%.16d", 2 + aigerNumSize(iPoUnknown) /*size in bytes*/);
    fmt.Fprintf(pFile, " ");

    fputc( byte(BRIDGE_VALUE_X), pFile ); // undef
    fputc( byte(1), pFile ); // size of vector (Armin's encoding)
    Gia_AigerWriteUnsignedFile( pFile, iPoUnknown ); // number of the property (Armin's encoding)
    fflush(pFile);
}
func Gia_ManFromBridgeCex(pFile *os.File, pCex *Abc_Cex_t) {
    var i, f, iBit int32;//, RetValue;
    var vStr *Vec_Str_t  = Vec_StrAlloc( 1000 );
    Vec_StrPush( vStr, byte(BRIDGE_VALUE_0) ); // false
    Vec_StrPush( vStr, byte(1) ); // size of vector (Armin's encoding)
    Gia_AigerWriteUnsigned( vStr, pCex.iPo ); // number of the property (Armin's encoding)
    Vec_StrPush( vStr, byte(1) ); // size of vector (Armin's encoding)
    Gia_AigerWriteUnsigned( vStr, pCex.iFrame ); // depth

    Gia_AigerWriteUnsigned( vStr, 1 ); // concrete
    Gia_AigerWriteUnsigned( vStr, pCex.iFrame + 1 ); // number of frames (1 more than depth)
    iBit = pCex.nRegs;
    for  f = 0; f <= pCex.iFrame; f++  {
        Gia_AigerWriteUnsigned( vStr, pCex.nPis ); // num of inputs
        for  i = 0; i < pCex.nPis; i, iBit = i+1, iBit+1  {
            Vec_StrPush( vStr, byte((Abc_InfoHasBit(pCex.pData, iBit) ? BRIDGE_VALUE_1:BRIDGE_VALUE_0)) );
        } // value
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    Vec_StrPush( vStr, byte(1) ); // the number of frames (for a concrete counter-example)
    Gia_AigerWriteUnsigned( vStr, pCex.nRegs ); // num of flops
    for  i = 0; i < pCex.nRegs; i++  {
        Vec_StrPush( vStr, byte(BRIDGE_VALUE_0) );
    } // always zero!!!
//    RetValue = fwrite( Vec_StrArray(vStr), Vec_StrSize(vStr), 1, pFile );
    Gia_CreateHeader(pFile, 101/*type=Result*/, Vec_StrSize(vStr), (unsigned char*)Vec_StrArray(vStr));

    Vec_StrFree( vStr );
    fflush(pFile);
}
func Gia_ManToBridgeResult(pFile *os.File, Result int32, pCex *Abc_Cex_t, iPoProved int32) int32 {
    if Result == 0 { // sat
        Gia_ManFromBridgeCex( pFile, pCex );
    } else if Result == 1 { // unsat
        Gia_ManFromBridgeHolds( pFile, iPoProved );
    } else if Result == -1 { // undef
        Gia_ManFromBridgeUnknown( pFile, iPoProved );
    } else { assert(false, "0"); }
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
Gia_Man_t *  Gia_ManFromBridgeReadBody( int Size, unsigned char * pBuffer, Vec_Int_t ** pvInits )
{
    var fHash int32;
    var vLits, vInits *Vec_Int_t;
    var p *Gia_Man_t  = nil;
    unsigned char * pBufferPivot, * pBufferEnd = pBuffer + Size;
    var i, nInputs, nFlops, nGates, nProps int32;
    var verFairness, nFairness, nConstraints int32;
    var iFan0, iFan1 uint32;

    nInputs = Gia_AigerReadUnsigned( &pBuffer );
    nFlops  = Gia_AigerReadUnsigned( &pBuffer );
    nGates  = Gia_AigerReadUnsigned( &pBuffer );

    vLits = Vec_IntAlloc( 1000 );
    Vec_IntPush( vLits, -999 );
    Vec_IntPush( vLits,  1 );

    // start the AIG package
    p = Gia_ManStart( nInputs + nFlops * 2 + nGates + 1 + 1 ); // PI+FO+FI+AND+PO+1
    p.pName = Abc_UtilStrsav( "temp" );

    // create PIs
    for  i = 0; i < nInputs; i++  {
        Vec_IntPush( vLits, Gia_ManAppendCi( p ) );
    }

    // create flop outputs
    for  i = 0; i < nFlops; i++  {
        Vec_IntPush( vLits, Gia_ManAppendCi( p ) );
    }

    // create nodes
    if fHash != 0 {
        Gia_ManHashAlloc( p ); }
    for  i = 0; i < nGates; i++  {
        iFan0 = Gia_AigerReadUnsigned( &pBuffer );
        iFan1 = Gia_AigerReadUnsigned( &pBuffer );
        assert(!(iFan0 & 1), "!(iFan0 & 1)");
        iFan0 >>= 1;
        iFan0 = Abc_LitNotCond( Vec_IntEntry(vLits, iFan0 >> 1), iFan0 & 1 );
        iFan1 = Abc_LitNotCond( Vec_IntEntry(vLits, iFan1 >> 1), iFan1 & 1 );
        if fHash != 0 {
            Vec_IntPush( vLits, Gia_ManHashAnd(p, iFan0, iFan1) );
        } else {
            Vec_IntPush( vLits, Gia_ManAppendAnd(p, iFan0, iFan1) ); }

    }
    if fHash != 0 {
        Gia_ManHashStop( p ); }

    // remember where flops begin
    pBufferPivot = pBuffer;
    // scroll through flops
    for  i = 0; i < nFlops; i++  {
        Gia_AigerReadUnsigned( &pBuffer );
    }

    // create POs
    nProps = Gia_AigerReadUnsigned( &pBuffer );
//    assert( nProps == 1 );
    for  i = 0; i < nProps; i++  {
        iFan0 = Gia_AigerReadUnsigned( &pBuffer );
        iFan0 = Abc_LitNotCond( Vec_IntEntry(vLits, iFan0 >> 1), iFan0 & 1 );
        // complement property output!!!
        Gia_ManAppendCo( p, Abc_LitNot(iFan0) );
    }

    verFairness = Gia_AigerReadUnsigned( &pBuffer );
    assert(verFairness == 1, "verFairness == 1");

    nFairness = Gia_AigerReadUnsigned( &pBuffer );
    assert(nFairness == 0, "nFairness == 0");

    nConstraints = Gia_AigerReadUnsigned( &pBuffer );
    assert(nConstraints == 0, "nConstraints == 0");

    // make sure the end of buffer is reached
    assert(pBufferEnd == pBuffer, "pBufferEnd == pBuffer");

    // resetting to flops
    pBuffer = pBufferPivot;
    vInits = Vec_IntAlloc( nFlops );
    for  i = 0; i < nFlops; i++  {
        iFan0 = Gia_AigerReadUnsigned( &pBuffer );
        assert((iFan0 & 3) == BRIDGE_VALUE_0, "(iFan0 & 3) == BRIDGE_VALUE_0");
        Vec_IntPush( vInits, iFan0 & 3 ); // 0 = X value; 1 = not used; 2 = false; 3 = true
        iFan0 >>= 2;
        iFan0 = Abc_LitNotCond( Vec_IntEntry(vLits, iFan0 >> 1), iFan0 & 1 );
        Gia_ManAppendCo( p, iFan0 );
    }
    Gia_ManSetRegNum( p, nFlops );
    Vec_IntFree( vLits );

    // remove wholes in the node list
    if fHash != 0 {
        var pTemp *Gia_Man_t;
        p = Gia_ManCleanup( pTemp = p );
        Gia_ManStop( pTemp );
    }

    // return
    if pvInits {
        *pvInits = vInits;
    } else {
        Vec_IntFree( vInits ); }
    return p;
}


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
int Gia_ManFromBridgeReadPackage( FILE * pFile, int * pType, int * pSize, unsigned char ** ppBuffer )
{
    var Temp [24]byte;
    var RetValue int32;
    RetValue = fread( Temp, 24, 1, pFile );
    if RetValue != 1 {
        fmt.Printf("Gia_ManFromBridgeReadPackage();  Error 1: Something is wrong!\n");
        return 0;
    }
    Temp[6] = 0;
    Temp[23]= 0;

    *pType = strconv.Atoi(Temp);
    *pSize = strconv.Atoi(Temp + 7);

    *ppBuffer = ABC_ALLOC( unsigned char, *pSize );
    RetValue = fread( *ppBuffer, *pSize, 1, pFile );
    if RetValue != 1 && *pSize != 0 {
        ABC_FREE( *ppBuffer );
        fmt.Printf("Gia_ManFromBridgeReadPackage();  Error 2: Something is wrong!\n");
        return 0;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFromBridge(pFile *os.File, pvInit **Vec_Int_t) *Gia_Man_t {
    unsigned char * pBuffer;
    var Type, Size, RetValue int32;
    var p *Gia_Man_t  = nil;

    RetValue = Gia_ManFromBridgeReadPackage( pFile, &Type, &Size, &pBuffer );
    ABC_FREE( pBuffer );
    if !(RetValue != 0) {
        return nil; }

    RetValue = Gia_ManFromBridgeReadPackage( pFile, &Type, &Size, &pBuffer );
    if !(RetValue != 0) {
        return nil; }

    p = Gia_ManFromBridgeReadBody( Size, pBuffer, pvInit );
    ABC_FREE( pBuffer );
    if p == nil {
        return nil; }

    RetValue = Gia_ManFromBridgeReadPackage( pFile, &Type, &Size, &pBuffer );
    ABC_FREE( pBuffer );
    if !(RetValue != 0) {
        return nil; }

    return p;
}

/*
    {
        extern void Gia_ManFromBridgeTest( char * pFileName );
        Gia_ManFromBridgeTest( "C:\\_projects\\abc\\_TEST\\bug\\65\\par.dump" );
    }
*/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManToBridgeAbsNetlistTest(pFileName string, p *Gia_Man_t, msg_type int32) {
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open output file \"%s\".\n", pFileName);
        return;
    }
    Gia_ManToBridgeAbsNetlist( pFile, p, msg_type );
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFromBridgeTest(pFileName string) {
    var p *Gia_Man_t;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open input file \"%s\".\n", pFileName);
        return;
    }

    p = Gia_ManFromBridge( pFile, nil );
    _ = (pFile).Close();

    Gia_ManPrintStats( p, nil );
    Gia_AigerWrite( p, "temp.aig", 0, 0, 0 );

    Gia_ManToBridgeAbsNetlistTest( "par_.dump", p, BRIDGE_ABS_NETLIST );
    Gia_ManStop( p );
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
