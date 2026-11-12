package gia

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [gia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStgPrint(pFile *os.File, vLines *Vec_Int_t, nIns int32, nOuts int32, nStates int32) {
    var i int32;
    var nDigits int32 = Abc_Base10Log( nStates );
    assert(Vec_IntSize(vLines) % 4 == 0, "Vec_IntSize(vLines) % 4 == 0");
    for  i = 0; i < Vec_IntSize(vLines); i += 4  {
        var iMint int32  = Vec_IntEntry(vLines, i  );
        var iCur int32   = Vec_IntEntry(vLines, i+1) - 1;
        var iNext int32  = Vec_IntEntry(vLines, i+2) - 1;
        var iOut int32   = Vec_IntEntry(vLines, i+3);
        assert(iMint >= 0 && iMint < (1<<nIns), "iMint >= 0 && iMint < (1<<nIns)");
        assert(iCur  >= 0 && iCur  < nStates, "iCur  >= 0 && iCur  < nStates");
        assert(iNext >= 0 && iNext < nStates, "iNext >= 0 && iNext < nStates");
        assert(iOut  >= 0 && iOut  < (1<<nOuts), "iOut  >= 0 && iOut  < (1<<nOuts)");
        Extra_PrintBinary( pFile, (unsigned *)Vec_IntEntryP(vLines, i),  nIns );
        fmt.Fprintf(pFile, " %*d", nDigits, Vec_IntEntry(vLines,  i+1));
        fmt.Fprintf(pFile, " %*d ", nDigits, Vec_IntEntry(vLines,  i+2));
        Extra_PrintBinary( pFile, (unsigned *)Vec_IntEntryP(vLines, i+3), nOuts );
        fmt.Fprintf(pFile, "\n");
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStgReadLines(pFileName string, pnIns *int32, pnOuts *int32, pnStates *int32) *Vec_Int_t {
    var vLines *Vec_Int_t;
    var pBuffer [1000]byte;
    var pToken string;
    var Number int32;
    var nInputs int32 = -1;
    var nOutputs int32 = -1;
    var nStates int32 = 1;
    var pFile *os.File;
    if !(int32(strings.Compare(pFileName + len(pFileName) - 3, "aig")) != 0) {
        fmt.Printf("Input file \"%s\" has extension \"%s\".\n", pFileName, "aig");
        return nil;
    }
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\".\n", pFileName);
        return nil;
    }
    vLines = Vec_IntAlloc( 1000 );
    for  fgets( pBuffer, 1000, pFile ) != nil  {
        if pBuffer[0] == '.' || pBuffer[0] == '#' {
            continue; }
        // read condition
        pToken = strtok( pBuffer, " \r\n" );
        if nInputs == -1 {
            nInputs = len(pToken);
        } else {
            assert(nInputs == int32(len(pToken)), "nInputs == int32(strlen(pToken))"); }
        Number = Extra_ReadBinary( pToken );
        Vec_IntPush( vLines, Number );
        // read current state
        pToken = strtok( nil, " \r\n" );
        Vec_IntPush( vLines, strconv.Atoi(pToken) );
        nStates = Abc_MaxInt( nStates, Vec_IntEntryLast(vLines)+1 );
        // read next state
        pToken = strtok( nil, " \r\n" );
        Vec_IntPush( vLines, strconv.Atoi(pToken) );
        // read output
        pToken = strtok( nil, " \r\n" );
        if nOutputs == -1 {
            nOutputs = len(pToken);
        } else {
            assert(nOutputs == int32(len(pToken)), "nOutputs == int32(strlen(pToken))"); }
        Number = Extra_ReadBinary( pToken );
        Vec_IntPush( vLines, Number );
    }
    _ = (pFile).Close();
    if pnIns != nil {
        *pnIns = nInputs; }
    if pnOuts != nil {
        *pnOuts = nOutputs; }
    if pnStates != nil {
        *pnStates = nStates; }
    return vLines;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStgRead(pFileName string, kHot int32, fVerbose int32) *Gia_Man_t {
    var p *Gia_Man_t;
    var vLines *Vec_Int_t;
    var nIns, nOuts, nStates int32;
    vLines = Gia_ManStgReadLines( pFileName, &nIns, &nOuts, &nStates );
    if vLines == nil {
        return nil; }
//    p = Gia_ManStgOneHot( vLines, nIns, nOuts, nStates );
    p = Gia_ManStgKHot( vLines, nIns, nOuts, nStates, kHot, fVerbose );
    Vec_IntFree( vLines );
    return p;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

