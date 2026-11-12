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
func Gia_ManPrintStateEncoding(vCodes *Vec_Vec_t, nBits int32) {
    var pBuffer string;
    var vVec *Vec_Int_t;
    var i, k, Bit int32;
    // inlining of macro ABC_ALLOC
    pBuffer = ((char *) malloc(sizeof(char) * (size_t)(nBits + 1)));
    pBuffer[nBits] = 0;
    // inlining of macro Vec_VecForEachLevelInt
    for  i = 0; (i < Vec_VecSize(vCodes)) && (((vVec) = Vec_VecEntryInt(vCodes, i)), 1); i++  {
        fmt.Printf("%6d : ", i+1);
        memset( pBuffer, '-', (size_t)nBits );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((Bit) = Vec_IntEntry(vVec, k)), 1); k++  {
            assert(Bit < nBits, "Bit < nBits");
            pBuffer[Bit] = '1';
        }
        fmt.Printf("%s\n", pBuffer);
    }
    // inlining of macro ABC_FREE
    ((pBuffer) ? (free((char *) (pBuffer)), (pBuffer) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCreateOrGate(p *Gia_Man_t, vLits *Vec_Int_t) int32 {
    if Vec_IntSize(vLits) == 0 {
        return 0; }
    for  Vec_IntSize(vLits) > 1  {
        var i, k, Lit1, Lit2, LitRes int32;
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(vLits)) && (((Lit1) = Vec_IntEntry(vLits, i)), 1) && (((Lit2) = Vec_IntEntry(vLits, i+1)), 1); i += 2  {
            LitRes = Gia_ManHashOr( p, Lit1, Lit2 );
            Vec_IntWriteEntry( vLits, k++, LitRes );
        }
        if Vec_IntSize(vLits) & 1 {
            Vec_IntWriteEntry( vLits, k++, Vec_IntEntryLast(vLits) ); }
        Vec_IntShrink( vLits, k );
    }
    assert(Vec_IntSize(vLits) == 1, "Vec_IntSize(vLits) == 1");
    return Vec_IntEntry(vLits, 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAssignCodes(kHot int32, nStates int32, pnBits *int32) *Vec_Vec_t {
    var vCodes *Vec_Vec_t;
    var s, i1, i2, i3, i4, i5, nBits int32;
    assert(nStates > 0, "nStates > 0");
    assert(kHot >= 1 && kHot <= 5, "kHot >= 1 && kHot <= 5");
    vCodes = Vec_VecStart( nStates );
    *pnBits = -1;
    if kHot == 1 {
        for  i1 = 0; i1 < nStates; i1++  {
            Vec_VecPushInt( vCodes, i1, i1 );
        }
        *pnBits = nStates;
        return vCodes;
    }
    if kHot == 2 {
        for  nBits = kHot; nBits < ABC_INFINITY; nBits++  {
            if nBits * (nBits-1) / 2 >= nStates {
                break; }
        }
        *pnBits = nBits;
        s = 0;
        for  i1 = 0; i1 < nBits; i1++  {
        for  i2 = i1 + 1; i2 < nBits; i2++  {
            Vec_VecPushInt( vCodes, s, i1 );
            Vec_VecPushInt( vCodes, s, i2 );
            if s++; s == nStates {
                return vCodes; }
        }
        }
    }
    if kHot == 3 {
        for  nBits = kHot; nBits < ABC_INFINITY; nBits++  {
            if nBits * (nBits-1) * (nBits-2) / 6 >= nStates {
                break; }
        }       
        *pnBits = nBits;
        s = 0;
        for  i1 = 0; i1 < nBits; i1++  {
        for  i2 = i1 + 1; i2 < nBits; i2++  {
        for  i3 = i2 + 1; i3 < nBits; i3++  {
            Vec_VecPushInt( vCodes, s, i1 );
            Vec_VecPushInt( vCodes, s, i2 );
            Vec_VecPushInt( vCodes, s, i3 );
            if s++; s == nStates {
                return vCodes; }
        }
        }
        }
    }
    if kHot == 4 {
        for  nBits = kHot; nBits < ABC_INFINITY; nBits++  {
            if nBits * (nBits-1) * (nBits-2) * (nBits-3) / 24 >= nStates {
                break; }
        }       
        *pnBits = nBits;
        s = 0;
        for  i1 = 0; i1 < nBits; i1++  {
        for  i2 = i1 + 1; i2 < nBits; i2++  {
        for  i3 = i2 + 1; i3 < nBits; i3++  {
        for  i4 = i3 + 1; i4 < nBits; i4++  {
            Vec_VecPushInt( vCodes, s, i1 );
            Vec_VecPushInt( vCodes, s, i2 );
            Vec_VecPushInt( vCodes, s, i3 );
            Vec_VecPushInt( vCodes, s, i4 );
            if s++; s == nStates {
                return vCodes; }
        }
        }
        }
        }
    }
    if kHot == 5 {
        for  nBits = kHot; nBits < ABC_INFINITY; nBits++  {
            if nBits * (nBits-1) * (nBits-2) * (nBits-3) * (nBits-4) / 120 >= nStates {
                break; }
        }       
        *pnBits = nBits;
        s = 0;
        for  i1 = 0; i1 < nBits; i1++  {
        for  i2 = i1 + 1; i2 < nBits; i2++  {
        for  i3 = i2 + 1; i3 < nBits; i3++  {
        for  i4 = i3 + 1; i4 < nBits; i4++  {
        for  i5 = i4 + 1; i5 < nBits; i5++  {
            Vec_VecPushInt( vCodes, s, i1 );
            Vec_VecPushInt( vCodes, s, i2 );
            Vec_VecPushInt( vCodes, s, i3 );
            Vec_VecPushInt( vCodes, s, i4 );
            Vec_VecPushInt( vCodes, s, i5 );
            if s++; s == nStates {
                return vCodes; }
        }
        }
        }
        }
        }
    }
    assert(false, "0");
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStgKHot(vLines *Vec_Int_t, nIns int32, nOuts int32, nStates int32, kHot int32, fVerbose int32) *Gia_Man_t {
    var p, pTemp *Gia_Man_t;
    var vInMints, vCurs, vVec *Vec_Int_t;
    var vLitsNext, vLitsOuts, vCodes *Vec_Vec_t;
    var i, b, k, nBits, LitC, Lit int32;
    assert(Vec_IntSize(vLines) % 4 == 0, "Vec_IntSize(vLines) % 4 == 0");

    // produce state encoding
    vCodes = Gia_ManAssignCodes( kHot, nStates, &nBits );
    assert(Vec_VecSize(vCodes) == nStates, "Vec_VecSize(vCodes) == nStates");
    if fVerbose != 0 {
        Gia_ManPrintStateEncoding( vCodes, nBits ); }

    // start manager
    p = Gia_ManStart( 10000 );
    p.pName = Abc_UtilStrsav( "stg" );
    for  i = 0; i < nIns + nBits; i++  {
        Gia_ManAppendCi(p);
    }

    // create input minterms
    Gia_ManHashAlloc( p );
    vInMints = Vec_IntAlloc( 1 << nIns );
    for  i = 0; i < (1 << nIns); i++  {
        for  Lit, b = 1, 0; b < nIns; b++  {
            Lit = Gia_ManHashAnd( p, Lit, Abc_Var2Lit( b+1, !((i >> b) & 1) ) );
        }
        Vec_IntPush( vInMints, Lit );
    } 

    // create current states
    vCurs  = Vec_IntAlloc( nStates );
    // inlining of macro Vec_VecForEachLevelInt
    for  i = 0; (i < Vec_VecSize(vCodes)) && (((vVec) = Vec_VecEntryInt(vCodes, i)), 1); i++  {
        Lit = 1;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((b) = Vec_IntEntry(vVec, k)), 1); k++  {
            assert(b >= 0 && b < nBits, "b >= 0 && b < nBits");
            Lit = Gia_ManHashAnd( p, Lit, Abc_Var2Lit(1+nIns+b, (b<kHot)) );
        }
        Vec_IntPush( vCurs, Lit );
    }

    // go through the lines
    vLitsNext = Vec_VecStart( nBits );
    vLitsOuts = Vec_VecStart( nOuts );
    for  i = 0; i < Vec_IntSize(vLines);  {
        var iMint int32  = Vec_IntEntry(vLines, i++);
        var iCur int32   = Vec_IntEntry(vLines, i++);
        var iNext int32  = Vec_IntEntry(vLines, i++);
        var iOut int32   = Vec_IntEntry(vLines, i++);
        assert(iMint >= 0 && iMint < (1<<nIns), "iMint >= 0 && iMint < (1<<nIns)");
        assert(iCur  >= 0 && iCur  < nStates, "iCur  >= 0 && iCur  < nStates");
        assert(iNext >= 0 && iNext < nStates, "iNext >= 0 && iNext < nStates");
        assert(iOut  >= 0 && iOut  < (1<<nOuts), "iOut  >= 0 && iOut  < (1<<nOuts)");
        // create condition
        LitC = Gia_ManHashAnd( p, Vec_IntEntry(vInMints, iMint), Vec_IntEntry(vCurs, iCur) );
        // update next state
//        Vec_VecPushInt( vLitsNext, iNext, LitC );
        vVec = (Vec_Int_t *)Vec_VecEntryInt( vCodes, iNext );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((b) = Vec_IntEntry(vVec, k)), 1); k++  {
            Vec_VecPushInt( vLitsNext, b, LitC );
        }
        // update outputs
        for  b = 0; b < nOuts; b++  {
            if (iOut >> b) & 1 { 
                Vec_VecPushInt( vLitsOuts, b, LitC ); }
        }
    }
    Vec_IntFree( vInMints );
    Vec_IntFree( vCurs );
    Vec_VecFree( vCodes );

    // create POs
    // inlining of macro Vec_VecForEachLevelInt
    for  b = 0; (b < Vec_VecSize(vLitsOuts)) && (((vVec) = Vec_VecEntryInt(vLitsOuts, b)), 1); b++  {
        Gia_ManAppendCo( p, Gia_ManCreateOrGate(p, vVec) );
    }
    Vec_VecFree( vLitsOuts );

    // create next states
    // inlining of macro Vec_VecForEachLevelInt
    for  b = 0; (b < Vec_VecSize(vLitsNext)) && (((vVec) = Vec_VecEntryInt(vLitsNext, b)), 1); b++  {
        Gia_ManAppendCo( p, Abc_LitNotCond( Gia_ManCreateOrGate(p, vVec), (b<kHot) ) );
    }
    Vec_VecFree( vLitsNext );

    Gia_ManSetRegNum( p, nBits );
    Gia_ManHashStop( p );

    p = Gia_ManCleanup( pTemp = p );
    Gia_ManStop( pTemp );
    assert(!Gia_ManHasDangling(p), "!Gia_ManHasDangling(p)");
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStgOneHot(vLines *Vec_Int_t, nIns int32, nOuts int32, nStates int32) *Gia_Man_t {
    var p, pTemp *Gia_Man_t;
    var vInMints, vCurs, vVec *Vec_Int_t;
    var vLitsNext, vLitsOuts *Vec_Vec_t;
    var i, b, LitC, Lit int32;
    assert(Vec_IntSize(vLines) % 4 == 0, "Vec_IntSize(vLines) % 4 == 0");

    // start manager
    p = Gia_ManStart( 10000 );
    p.pName = Abc_UtilStrsav( "stg" );
    for  i = 0; i < nIns + nStates; i++  {
        Gia_ManAppendCi(p);
    }

    // create input minterms
    Gia_ManHashAlloc( p );
    vInMints = Vec_IntAlloc( 1 << nIns );
    for  i = 0; i < (1 << nIns); i++  {
        for  Lit, b = 1, 0; b < nIns; b++  {
            Lit = Gia_ManHashAnd( p, Lit, Abc_Var2Lit( b+1, !((i >> b) & 1) ) );
        }
        Vec_IntPush( vInMints, Lit );
    } 

    // create current states
    vCurs  = Vec_IntAlloc( nStates );
    for  i = 0; i < nStates; i++  {
        Vec_IntPush( vCurs, Abc_Var2Lit( 1+nIns+i, !i ) );
    }

    // go through the lines
    vLitsNext = Vec_VecStart( nStates );
    vLitsOuts = Vec_VecStart( nOuts );
    for  i = 0; i < Vec_IntSize(vLines);  {
        var iMint int32  = Vec_IntEntry(vLines, i++);
        var iCur int32   = Vec_IntEntry(vLines, i++) - 1;
        var iNext int32  = Vec_IntEntry(vLines, i++) - 1;
        var iOut int32   = Vec_IntEntry(vLines, i++);
        assert(iMint >= 0 && iMint < (1<<nIns), "iMint >= 0 && iMint < (1<<nIns)");
        assert(iCur  >= 0 && iCur  < nStates, "iCur  >= 0 && iCur  < nStates");
        assert(iNext >= 0 && iNext < nStates, "iNext >= 0 && iNext < nStates");
        assert(iOut  >= 0 && iOut  < (1<<nOuts), "iOut  >= 0 && iOut  < (1<<nOuts)");
        // create condition
        LitC = Gia_ManHashAnd( p, Vec_IntEntry(vInMints, iMint), Vec_IntEntry(vCurs, iCur) );
        // update next state
//        Lit = Gia_ManHashOr( p, LitC, Vec_IntEntry(vNexts, iNext) );
//        Vec_IntWriteEntry( vNexts, iNext, Lit );
        Vec_VecPushInt( vLitsNext, iNext, LitC );
        // update outputs
        for  b = 0; b < nOuts; b++  {
            if (iOut >> b) & 1 {
//                Lit = Gia_ManHashOr( p, LitC, Vec_IntEntry(vOuts, b) );
//                Vec_IntWriteEntry( vOuts, b, Lit );
                Vec_VecPushInt( vLitsOuts, b, LitC );
            }
        }
    }
    Vec_IntFree( vInMints );
    Vec_IntFree( vCurs );

    // create POs
    // inlining of macro Vec_VecForEachLevelInt
    for  i = 0; (i < Vec_VecSize(vLitsOuts)) && (((vVec) = Vec_VecEntryInt(vLitsOuts, i)), 1); i++  {
        Gia_ManAppendCo( p, Gia_ManCreateOrGate(p, vVec) );
    }
    Vec_VecFree( vLitsOuts );

    // create next states
    // inlining of macro Vec_VecForEachLevelInt
    for  i = 0; (i < Vec_VecSize(vLitsNext)) && (((vVec) = Vec_VecEntryInt(vLitsNext, i)), 1); i++  {
        Gia_ManAppendCo( p, Abc_LitNotCond( Gia_ManCreateOrGate(p, vVec), !i ) );
    }
    Vec_VecFree( vLitsNext );

    Gia_ManSetRegNum( p, nStates );
    Gia_ManHashStop( p );

    p = Gia_ManCleanup( pTemp = p );
    Gia_ManStop( pTemp );
    assert(!Gia_ManHasDangling(p), "!Gia_ManHasDangling(p)");
    return p;
}

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

