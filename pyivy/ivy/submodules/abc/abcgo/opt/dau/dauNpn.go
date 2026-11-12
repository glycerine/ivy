package dau

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [dau.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dau.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"
//#include "bool/lucky/lucky.h"

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
func Dau_TruthEnum(nVars int32) {
    var fUseTable int32  = 1;
    var clk abctime  = Abc_Clock();
    var nSizeLog int32  = (1<<nVars) -2;
    var nSizeW int32  = 1 << nSizeLog;
    var nPerms int32   = Extra_Factorial( nVars );
    var nMints int32   = 1 << nVars;
    var pPerm *int32  = Extra_PermSchedule( nVars );
    var pComp *int32  = Extra_GreyCodeSchedule( nVars );
    var nFuncs word  = ((word)1 << (((word)1 << nVars)-1));
    // inlining of macro ABC_CALLOC
    var pPres *word  = ((word *) calloc((size_t)(1 << ((1<<nVars)-7)), sizeof(word)));
    var pTable *uint32;
    if fUseTable != 0 {
        // inlining of macro ABC_CALLOC
        pTable = (unsigned *)((word *) calloc((size_t)(nSizeW), sizeof(word)));
    } else {
        pTable = nil;
    }
    var vNpns *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var tMask word  = Abc_Tt6Mask( 1 << nVars );
    var tTemp, tCur word;
    var i, k int32;
    if pPres == nil {
        fmt.Printf("Cannot alloc memory for marks.\n");
        return;
    }
    if pTable == nil {
        fmt.Printf("Cannot alloc memory for table.\n"); }
    for  tCur = 0; tCur < nFuncs; tCur++  {
        if (tCur & 0x3FFFF) == 0 {
            fmt.Printf("Finished %08x.  Classes = %6d.  ", int32(tCur), Vec_IntSize(vNpns));
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            fflush(os.Stdout);
        }
        if Abc_TtGetBit(pPres, int32(tCur)) {
            continue; }
        //Extra_PrintBinary( stdout, (unsigned *)&tCur, 16 ); printf( " %04x\n", (int)tCur );
        //Dau_DsdPrintFromTruth( &tCur, 4 ); printf( "\n" );
        Vec_IntPush( vNpns, int32(tCur) );
        tTemp = tCur;
        for  i = 0; i < nPerms; i++  {
            for  k = 0; k < nMints; k++  {
                if tCur < nFuncs {
                    if pTable != nil { pTable[int32(tCur)] = tTemp; }
                    Abc_TtSetBit( pPres, int32(tCur) );
                }
                if (tMask & ~tCur) < nFuncs {
                    if pTable != nil { pTable[int32((tMask & ~tCur))] = tTemp; }
                    Abc_TtSetBit( pPres, int32((tMask & ~tCur)) );
                }
                tCur = Abc_Tt6Flip( tCur, pComp[k] );
            }
            tCur = Abc_Tt6SwapAdjacent( tCur, pPerm[i] );
        }
        assert(tTemp == tCur, "tTemp == tCur");
    }
    fmt.Printf("Computed %d NPN classes of %d variables.  ", Vec_IntSize(vNpns), nVars);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    fflush(os.Stdout);
    Vec_IntFree( vNpns );
    // inlining of macro ABC_FREE
    ((pPres) ? (free((char *) (pPres)), (pPres) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pPerm) ? (free((char *) (pPerm)), (pPerm) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pComp) ? (free((char *) (pComp)), (pComp) = 0) : 0);
    // write into file
    if pTable != nil {
        var pFile *os.File;
        var RetValue int32;
        var pFileName string // was [200]byte in C

        pFileName = fmt.Sprintf("tableW%d.data", nSizeLog);
        pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        RetValue = fwrite( pTable, 8, nSizeW, pFile );
        RetValue = 0;
        _ = (pFile).Close();
        // inlining of macro ABC_FREE
        ((pTable) ? (free((char *) (pTable)), (pTable) = 0) : 0);
    }
}

 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_ReadFile(pFileName string, nSizeW int32) *uint32 {
    var clk abctime  = Abc_Clock();
    var pFile *os.File  = os.Open(pFileName);
    // inlining of macro ABC_CALLOC
    var p *uint32  = (unsigned *)((word *) calloc((size_t)(nSizeW), sizeof(word)));
    var RetValue int32;
    if pFile != nil {
        RetValue = fread( p, sizeof(word), nSizeW, pFile );
    } else {
        RetValue = 0;
    }
    RetValue = 0;
    if pFile != nil {
        fmt.Printf("Finished reading file \"%s\".\n", pFileName);
        _ = (pFile).Close();
    
    } else {
        fmt.Printf("Cannot open input file \"%s\".\n", pFileName); }
    Abc_PrintTime( 1, "File reading", Abc_Clock() - clk );
    return p;
}
func Dau_AddFunction(tCur word, nVars int32, pTable *uint32, vNpns *Vec_Int_t, vNpns_ *Vec_Int_t) int32 {
    var Digit int32   = (1 << nVars)-1;
    var tMask word  = Abc_Tt6Mask( 1 << nVars );
    var tNorm word;
    if (tCur >> Digit) & 1 {
        tNorm = ~tCur;
    } else {
        tNorm = tCur;
    }
    var t uint32  = uint32((tNorm & tMask));
    var tRep uint32  = pTable[t] & 0x7FFFFFFF;
    var tRep2 uint32  = pTable[tRep];
    assert(((tNorm >> Digit) & 1) == 0, "((tNorm >> Digit) & 1) == 0");
    if (tRep2 >> 31) == 0 { // first time
    
        Vec_IntPush( vNpns, tRep2 );
        if Abc_TtSupportSize(&tCur, nVars) < nVars {
            Vec_IntPush( vNpns_, tRep2 ); }
        pTable[tRep] = tRep2 | (1 << 31);
        return tRep2;
    }
    return 0;
}
func Dau_NetworkEnum(nVars int32) {
    var clk abctime  = Abc_Clock();
    var Limit int32  = 2;
    var UseTwo int32;
    var nSizeLog int32  = (1<<nVars) -2;
    var nSizeW int32  = 1 << nSizeLog;
    var pFileName string // was [200]byte in C

    var pTable *uint32;
    var vNpns *Vec_Wec_t   = Vec_WecStart( 32 );
    var vNpns_ *Vec_Wec_t  = Vec_WecStart( 32 );
    var i, v, u, g, k, m, n, Res, Entry int32;
    var Inv uint32  = uint32(Abc_Tt6Mask(1 << (nVars-1)));
    pFileName = fmt.Sprintf("tableW%d.data", nSizeLog);
    pTable  = Dau_ReadFile( pFileName, nSizeW );
    // create constant function and buffer/inverter function
    pTable[0]   |= (1 << 31);
    pTable[Inv] |= (1 << 31);
    Vec_IntPushTwo( Vec_WecEntry(vNpns,  0), 0, Inv );
    Vec_IntPushTwo( Vec_WecEntry(vNpns_, 0), 0, Inv );
    fmt.Printf("Nodes = %2d.   New = %6d. Total = %6d.   New = %6d. Total = %6d.  ", 0, Vec_IntSize(Vec_WecEntry(vNpns,  0)), Vec_WecSizeSize(vNpns), Vec_IntSize(Vec_WecEntry(vNpns_, 0)), Vec_WecSizeSize(vNpns_));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    // numerate other functions based on how many nodes they have
    for  n = 1; n < 32; n++  {
        var vFuncsN2 *Vec_Int_t;
        if n > 1 {
            vFuncsN2 = Vec_WecEntry( vNpns, n-2 );
        } else {
            vFuncsN2 = nil;
        }
        var vFuncsN1 *Vec_Int_t  = Vec_WecEntry( vNpns, n-1 );
        var vFuncsN *Vec_Int_t   = Vec_WecEntry( vNpns, n   );
        var vFuncsN_ *Vec_Int_t  = Vec_WecEntry( vNpns_,n   );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vFuncsN1)) && (((Entry) = Vec_IntEntry(vFuncsN1, i)), 1); i++  {
            var uTruth word  = (((word)Entry) << 32) | (word)Entry;
            var nSupp int32  = Abc_TtSupportSize( &uTruth, nVars );
            assert(nSupp == 6 || !Abc_Tt6HasVar(uTruth, nVars-1-nSupp), "nSupp == 6 || !Abc_Tt6HasVar(uTruth, nVars-1-nSupp)");
            //printf( "Exploring function %4d with %d vars: ", i, nSupp );
            //printf( " %04x\n", (int)uTruth );
            //Dau_DsdPrintFromTruth( &uTruth, 4 );
            for  v = 0; v < nSupp; v++  {
                var tGate, tCur word;
                var Cof0 word  = Abc_Tt6Cofactor0( uTruth, nVars-1-v );
                var Cof1 word  = Abc_Tt6Cofactor1( uTruth, nVars-1-v );
                for  g = 0; g < Limit; g++  {
                    if nSupp < nVars {
                        if g == 0 {
                            tGate = s_Truths6[nVars-1-v] & s_Truths6[nVars-1-nSupp];
                            tCur  = (tGate & Cof1) | (~tGate & Cof0);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );

                            tCur  = (tGate & Cof0) | (~tGate & Cof1);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );
                        
                        } else {
                            tGate = s_Truths6[nVars-1-v] ^ s_Truths6[nVars-1-nSupp];
                            tCur  = (tGate & Cof1) | (~tGate & Cof0);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );
                        }
                    }
                }
                for  g = 0; g < Limit; g++  {
                    // add one cross bar
                    for  k = 0; k < nSupp; k++  { if k != v {
                        if g == 0 {
                            tGate = s_Truths6[nVars-1-v] & s_Truths6[nVars-1-k];
                            tCur  = (tGate & Cof1) | (~tGate & Cof0);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );

                            tCur  = (tGate & Cof0) | (~tGate & Cof1);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );

                            tGate = s_Truths6[nVars-1-v] & ~s_Truths6[nVars-1-k];
                            tCur  = (tGate & Cof1) | (~tGate & Cof0);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );

                            tCur  = (tGate & Cof0) | (~tGate & Cof1);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );
                        
                        } else {
                            tGate = s_Truths6[nVars-1-v] ^ s_Truths6[nVars-1-k];
                            tCur  = (tGate & Cof1) | (~tGate & Cof0);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );
                        }
                    }
                    }
                }
                for  g = 0; g < Limit; g++  {
                    // add two cross bars
                    for  k = 0;   k < nSupp; k++  { if k != v {
                    for  m = k+1; m < nSupp; m++  { if m != v {
                        if g == 0 {
                            tGate = s_Truths6[nVars-1-m] & s_Truths6[nVars-1-k];
                            tCur  = (tGate & Cof1) | (~tGate & Cof0);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );

                            tCur  = (tGate & Cof0) | (~tGate & Cof1);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );

                            tGate = s_Truths6[nVars-1-m] & ~s_Truths6[nVars-1-k];
                            tCur  = (tGate & Cof1) | (~tGate & Cof0);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );

                            tCur  = (tGate & Cof0) | (~tGate & Cof1);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );

                            tGate = ~s_Truths6[nVars-1-m] & s_Truths6[nVars-1-k];
                            tCur  = (tGate & Cof1) | (~tGate & Cof0);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );

                            tCur  = (tGate & Cof0) | (~tGate & Cof1);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );

                            tGate = ~s_Truths6[nVars-1-m] & ~s_Truths6[nVars-1-k];
                            tCur  = (tGate & Cof1) | (~tGate & Cof0);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );

                            tCur  = (tGate & Cof0) | (~tGate & Cof1);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );
                        
                        } else {
                            tGate = s_Truths6[nVars-1-m] ^ s_Truths6[nVars-1-k];
                            tCur  = (tGate & Cof1) | (~tGate & Cof0);
                            Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );
                        }
                    }
                    } }
                    }
                }
            }
        }
        if (UseTwo != 0) && (vFuncsN2 != nil) {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vFuncsN2)) && (((Entry) = Vec_IntEntry(vFuncsN2, i)), 1); i++  {
            var uTruth word  = (((word)Entry) << 32) | (word)Entry;
            var nSupp int32  = Abc_TtSupportSize( &uTruth, nVars );
            assert(nSupp == 6 || !Abc_Tt6HasVar(uTruth, nVars-1-nSupp), "nSupp == 6 || !Abc_Tt6HasVar(uTruth, nVars-1-nSupp)");
            //printf( "Exploring function %4d with %d vars: ", i, nSupp );
            //printf( " %04x\n", (int)uTruth );
            //Dau_DsdPrintFromTruth( &uTruth, 4 );
            for  v = 0; v < nSupp; v++  {
//            for ( u = v+1; u < nSupp; u++ ) if ( u != v )
            for  u = 0; u < nSupp; u++  { if u != v {
                var tGate1, tGate2, tCur word;
                var Cof0 word  = Abc_Tt6Cofactor0( uTruth, nVars-1-v );
                var Cof1 word  = Abc_Tt6Cofactor1( uTruth, nVars-1-v );

                var Cof00 word  = Abc_Tt6Cofactor0( Cof0, nVars-1-u );
                var Cof01 word  = Abc_Tt6Cofactor1( Cof0, nVars-1-u );
                var Cof10 word  = Abc_Tt6Cofactor0( Cof1, nVars-1-u );
                var Cof11 word  = Abc_Tt6Cofactor1( Cof1, nVars-1-u );

                tGate1 = s_Truths6[nVars-1-v] & s_Truths6[nVars-1-u];
                tGate2 = s_Truths6[nVars-1-v] ^ s_Truths6[nVars-1-u];

                Cof0  = (tGate2 & Cof00) | (~tGate2 & Cof01);
                Cof1  = (tGate2 & Cof10) | (~tGate2 & Cof11);

                tCur  = (tGate1 & Cof1)  | (~tGate1 & Cof0);
                Res = Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );
                if Res {
                    fmt.Printf("Found function %d\n", Res); }

                tCur  = (tGate1 & Cof0)  | (~tGate1 & Cof1);
                Res = Dau_AddFunction( tCur, nVars, pTable, vFuncsN, vFuncsN_ );
                if Res {
                    fmt.Printf("Found function %d\n", Res); }
            }
            }
            }
        } }
        fmt.Printf("Nodes = %2d.   New = %6d. Total = %6d.   New = %6d. Total = %6d.  ", n, Vec_IntSize(vFuncsN), Vec_WecSizeSize(vNpns), Vec_IntSize(vFuncsN_), Vec_WecSizeSize(vNpns_));
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        fflush(os.Stdout);
        if Vec_IntSize(vFuncsN) == 0 {
            break; }
    }
//    printf( "Functions with 7 nodes:\n" );
//    Vec_IntForEachEntry( Vec_WecEntry(vNpns_,7), Entry, i )
//        printf( "%04x ", Entry );
//    printf( "\n" );

    Vec_WecFree( vNpns );
    Vec_WecFree( vNpns_ );
    // inlining of macro ABC_FREE
    ((pTable) ? (free((char *) (pTable)), (pTable) = 0) : 0);
    Abc_PrintTime( 1, "Total time", Abc_Clock() - clk );
    fflush(os.Stdout);
}
func Dau_NetworkEnumTest() {
    //Dau_TruthEnum(3);
    Dau_NetworkEnum(4);
}



/**Function*************************************************************

  Synopsis    [Count the number of symmetric pairs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_CountSymms(t word, nVars int32) int32 {
    var Cof0, Cof1 word;
    var i, j, nPairs int32;
    for  i = 0; i < nVars; i++  {
    for  j = i+1; j < nVars; j++  {
        nPairs += Abc_TtVarsAreSymmetric(&t, nVars, i, j, &Cof0, &Cof1);
    }
    }
    return nPairs;
}
func Dau_CountSymms2(t word, nVars int32) int32 {
    var Cof0, Cof1 word;
    var i, j, SymVars int32;
    for  i = 0; i < nVars; i++  {
    for  j = i+1; j < nVars; j++  {
        if Abc_TtVarsAreSymmetric(&t, nVars, i, j, &Cof0, &Cof1) {
            SymVars |= (1 << j); }
    }
    }
    return SymVars;
}
func Dau_CountCompl1(t word, v int32, nVars int32) int32 {
    var tNew word  = Abc_Tt6Flip(t, v);
    var k int32;
    if tNew == ~t {
        return 1; }
    for  k = 0; k < nVars; k++  { if k != v {
        if tNew == Abc_Tt6Flip(t, k) {
            return 1; } }
    }
    return 0;
}
func Dau_CountCompl(t word, nVars int32) int32 {
    var i, nPairs int32;
    for  i = 0; i < nVars; i++  {
        nPairs += Dau_CountCompl1(t, i, nVars);
    }
    return nPairs;
}

/**Function*************************************************************

  Synopsis    [Performs exact canonicization of semi-canonical classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_ExactNpnForClasses(vTtMem *Vec_Mem_t, vNodSup *Vec_Int_t, nVars int32, nInputs int32) *Vec_Wrd_t {
    var vCanons *Vec_Wrd_t  = Vec_WrdStart( Vec_IntSize(vNodSup) );
    var pAuxWord, pAuxWord1 [1024]word;
    var uTruth word; var i, Entry int32;
    var pi *permInfo  = setPermInfoPtr(nVars);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodSup)) && (((Entry) = Vec_IntEntry(vNodSup, i)), 1); i++  {
        if (Entry & 0xF) > nVars {
            continue; }
        uTruth = *Vec_MemReadEntry( vTtMem, i );
        simpleMinimal(&uTruth, pAuxWord, pAuxWord1, pi, nVars);
        Vec_WrdWriteEntry( vCanons, i, uTruth );
    }
    freePermInfoPtr(pi);
    return vCanons;
}
func Dau_ExactNpnPrint(vTtMem *Vec_Mem_t, vNodSup *Vec_Int_t, nVars int32, nInputs int32, nNodesMax int32) {
    var clk abctime  = Abc_Clock(); var n, nTotal int32;
    var vCanons *Vec_Wrd_t  = Dau_ExactNpnForClasses( vTtMem, vNodSup, nVars, nInputs );
    var vTtMem2 *Vec_Mem_t   = Vec_MemAlloc( Vec_MemEntrySize(vTtMem), 10 );
    Vec_MemHashAlloc( vTtMem2, 1<<10 );
    Abc_PrintTime( 1, "Exact NPN computation time", Abc_Clock() - clk );
    fmt.Printf("Final results:\n");
    for  n = 0; n <= nNodesMax; n++  {
        var i, Entry, Entry2, nEntries2, Counter, Counter2 int32;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vNodSup)) && (((Entry) = Vec_IntEntry(vNodSup, i)), 1); i++  {
            if (Entry & 0xF) > nVars || (Entry >> 16) != n {
                continue; }
            Counter++;
            nEntries2 = Vec_MemEntryNum(vTtMem2);
            Entry2 = Vec_MemHashInsert( vTtMem2, Vec_WrdEntryP(vCanons, i) );
            if nEntries2 == Vec_MemEntryNum(vTtMem2) { // found in the table - not new
                continue; }
            Counter2++;
        }
        nTotal += Counter2;
        fmt.Printf("Nodes = %2d.  ", n);
        fmt.Printf("Semi-canonical = %8d.  ", Counter);
        fmt.Printf("Canonical = %8d.  ", Counter2);
        fmt.Printf("Total = %8d.", nTotal);
        fmt.Printf("\n");
    }
    Vec_MemHashFree( vTtMem2 );
    Vec_MemFreeP( &vTtMem2 );
    Vec_WrdFree( vCanons );
    fflush(os.Stdout);
}

/**Function*************************************************************

  Synopsis    [Saving hash tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_TablesSave(nInputs int32, nVars int32, vTtMem *Vec_Mem_t, vNodSup *Vec_Int_t, nFronts int32, clk abctime) {
    var pFile *os.File; 
    var FileName string // was [100]byte in C

    var i int32;
    var nWords int32 = Abc_TtWordNum(nInputs);
    // NPN classes
    FileName = fmt.Sprintf("npn%d%d.ttd", nInputs, nVars);
    pFile = fopen( FileName, "wb" );
    for  i = 0; i < Vec_MemEntryNum(vTtMem); i++  {
        fwrite( Vec_MemReadEntry(vTtMem, i), 8, nWords, pFile );
    }
    fwrite( Vec_IntArray(vNodSup), 4, Vec_IntSize(vNodSup), pFile );
    _ = (pFile).Close();
//    printf( "Dumped files with %10d classes after exploring %10d frontiers.\n", 
//        Vec_IntSize(vNodSup), nFronts );
    printf( "Dumped file \"%s\" with %10d classes after exploring %10d frontiers.  ", 
            FileName, Vec_IntSize(vNodSup), nFronts );
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    fflush(os.Stdout);
}

/**Function*************************************************************

  Synopsis    [Dump functions by the number of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DumpFuncs(vTtMem *Vec_Mem_t, vNodSup *Vec_Int_t, nVars int32, nMax int32) {
    var pFile [20]*os.File;
    var Counters [20]int32 ;
    var n, i int32;
    assert(nVars == 4 || nVars == 5, "nVars == 4 || nVars == 5");
    for  n = 0; n <= nMax; n++  {
        var FileName string // was [100]byte in C

        FileName = fmt.Sprintf("func%d_min%d.tt", nVars, n);
        pFile[n] = fopen( FileName, "wb" );
    }
    for  i = 0; i < Vec_MemEntryNum(vTtMem); i++  {
        var pTruth *word  = Vec_MemReadEntry( vTtMem, i );
        var NodSup int32  = Vec_IntEntry( vNodSup, i );
        if (NodSup & 0xF) != nVars {
            continue; }
        Counters[NodSup >> 16]++;
        if nVars == 4 {
            fmt.Fprintf(pFile[NodSup >> 16], "%04x\n", int32((0xFFFF & pTruth[0])));
        } else if nVars == 5 {
            fmt.Fprintf(pFile[NodSup >> 16], "%08x\n", int32((0xFFFFFFFF & pTruth[0]))); }
    }
    for  n = 0; n <= nMax; n++  {
        fmt.Printf("Dumped %8d  %d-node %d-input functions into file.\n", Counters[n], n, nVars);
        _ = (pFile[n]).Close();
    }
}

/**Function*************************************************************

  Synopsis    [Function enumeration.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_CountFuncs(vNodSup *Vec_Int_t, iStart int32, iStop int32, nVars int32) int32 {
    var i, Entry, Count int32;
    // inlining of macro Vec_IntForEachEntryStartStop
    for  i = iStart; (i < iStop) && (((Entry) = Vec_IntEntry(vNodSup, i)), 1); i++  {
        Count += ((Entry & 0xF) <= nVars);
    }
    return Count;
}
func Dau_PrintStats(nNodes int32, nInputs int32, nVars int32, vNodSup *Vec_Int_t, iStart int32, iStop int32, nSteps word, Count2 int32, clk abctime) int32 {
    var nNew int32;
    fmt.Printf("N =%2d | ", nNodes);
    fmt.Printf("C =%12.0f  ", float64((iword))nSteps);
    fmt.Printf("New%d =%10d  ", nInputs, iStop-iStart);
    fmt.Printf("All%d =%10d | ", nInputs, iStop);
    fmt.Printf("New%d =%8d  ", nVars, nNew = Dau_CountFuncs(vNodSup, iStart, iStop, nVars));
    fmt.Printf("All%d =%8d  ", nVars, Dau_CountFuncs(vNodSup,      0, iStop, nVars));
    fmt.Printf("Two =%6d ", Count2);
    //Abc_PrintTime( 1, "T",   Abc_Clock() - clk );
    Abc_Print(1, "%9.2f sec\n", 1.0*(Abc_Clock() - clk)/(int64(time.Second)));
    fflush(os.Stdout);
    return nNew;
}


func Dau_InsertFunction(pMan *Abc_TtHieMan_t, pCur *word, nNodes int32, nInputs int32, nVars0 int32, nVars int32, vTtMem *Vec_Mem_t, vNodSup *Vec_Int_t, nFronts int32, clk abctime) int32 {
    var DumpDelta int32  = 1000000;
    var Perm [16]byte ;
    var nVarsNew int32  = Abc_TtMinBase( pCur, nil, nVars, nInputs );
    //unsigned Phase = Abc_TtCanonicizeHie( pMan, pCur, nVarsNew, Perm, 1 );
    var Phase uint32  = Abc_TtCanonicizeWrap( Abc_TtCanonicizeAda, pMan, pCur, nVarsNew, Perm, 99 );
    var nEntries int32  = Vec_MemEntryNum(vTtMem);
    var Entry int32  = Vec_MemHashInsert( vTtMem, pCur );
    if nEntries == Vec_MemEntryNum(vTtMem) { // found in the table - not new
        return 0; }
    Entry = 0;
    Phase = 0;
    // this is a new class
    Vec_IntPush( vNodSup, (nNodes << 16) | nVarsNew );
    assert(Vec_MemEntryNum(vTtMem) == Vec_IntSize(vNodSup), "Vec_MemEntryNum(vTtMem) == Vec_IntSize(vNodSup)");
    if Vec_IntSize(vNodSup) % DumpDelta == 0 {
        Dau_TablesSave( nInputs, nVars0, vTtMem, vNodSup, nFronts, clk ); }
    return 1;
}
func Dau_FunctionEnum(nInputs int32, nVars int32, nNodeMax int32, fUseTwo int32, fReduce int32, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var nWords int32  = Abc_TtWordNum(nInputs); var nSteps word;
    var pMan *Abc_TtHieMan_t  = Abc_TtHieManStart( nInputs, 5 );
    var vTtMem *Vec_Mem_t   = Vec_MemAlloc( nWords, 16 );
    var vNodSup *Vec_Int_t  = Vec_IntAlloc( 1 << 16 );
    var v int32;
    var u int32;
    var k int32;
    var m int32;
    var n int32;
    var Entry int32;
    var nNew int32;
    var Limit [32]int32 = [32]int32{1, 2};
    var Truth [4]word ;
    assert(nVars >= 3 && nVars <= nInputs && nInputs <= 6, "nVars >= 3 && nVars <= nInputs && nInputs <= 6");
    Vec_MemHashAlloc( vTtMem,  1<<16 );
    // add constant 0
    Vec_MemHashInsert( vTtMem,  Truth );
    Vec_IntPush( vNodSup, 0 ); // nodes=0, supp=0
    // add buffer/inverter
    Abc_TtIthVar( Truth, 0, nInputs );
    Abc_TtNot( Truth, nWords );
    Vec_MemHashInsert( vTtMem,  Truth );
    Vec_IntPush( vNodSup, 1 ); // nodes=0, supp=1
    Dau_PrintStats( 0, nInputs, nVars, vNodSup, 0, 2, nSteps, 0, clk );
    // numerate other functions based on how many nodes they have
    for  n = 1; n <= nNodeMax; n++  {
        var Count2 int32;
        var fExpand int32  = !(fReduce && n == nNodeMax);
        for  Entry = Limit[n-1]; Entry < Limit[n]; Entry++  {
            var pTruth *word  = Vec_MemReadEntry( vTtMem, Entry );
            var NodSup int32  = Vec_IntEntry(vNodSup, Entry);
            var nSupp int32  = 0xF & NodSup;
            var SymVars int32  = Dau_CountSymms2( pTruth[0], nSupp );
            assert(n-1 == (NodSup >> 16), "n-1 == (NodSup >> 16)");
            assert(!Abc_Tt6HasVar(*pTruth, nSupp), "!Abc_Tt6HasVar(*pTruth, nSupp)");
            //printf( "Exploring function %4d with %d vars: ", i, nSupp );
            //printf( " %04x\n", (int)uTruth );
            //Dau_DsdPrintFromTruth( &uTruth, 4 );
            for  v = 0; v < nSupp; v++  { if (SymVars & (1 << v)) == 0 {
                var tGate, tCur word;
                var Cof0 word  = Abc_Tt6Cofactor0( *pTruth, v );
                var Cof1 word  = Abc_Tt6Cofactor1( *pTruth, v );
                // add one extra variable to support
                if nSupp < nInputs && (fExpand != 0) {
                    tGate = s_Truths6[v] & s_Truths6[nSupp];
                    tCur  = (tGate & Cof1) | (~tGate & Cof0);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp+1, vTtMem, vNodSup, Entry, clk );

                    tCur  = (tGate & Cof0) | (~tGate & Cof1);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp+1, vTtMem, vNodSup, Entry, clk );


                    tGate = s_Truths6[v] ^ s_Truths6[nSupp];
                    tCur  = (tGate & Cof1) | (~tGate & Cof0);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp+1, vTtMem, vNodSup, Entry, clk );

                    nSteps += 3;
                }
                // add one cross bar
                if fExpand != 0 {
                for  k = 0; k < nSupp; k++  { if k != v && ((SymVars & (1 << k)) == 0 || k == v+1) {
                    tGate = s_Truths6[v] & s_Truths6[k];
                    tCur  = (tGate & Cof1) | (~tGate & Cof0);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    tCur  = (tGate & Cof0) | (~tGate & Cof1);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    tGate = s_Truths6[v] & ~s_Truths6[k];
                    tCur  = (tGate & Cof1) | (~tGate & Cof0);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    tCur  = (tGate & Cof0) | (~tGate & Cof1);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );


                    tGate = s_Truths6[v] ^ s_Truths6[k];
                    tCur  = (tGate & Cof1) | (~tGate & Cof0);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    nSteps += 5;
                }
                } }
                // add two cross bars
                for  k = 0;   k < nSupp; k++  { if k != v {//&& ((SymVars & (1 << k)) == 0) )
                for  m = k+1; m < nSupp; m++  { if m != v {//&& ((SymVars & (1 << m)) == 0 || m == k+1) )
                
                    tGate = s_Truths6[m] & s_Truths6[k];
                    tCur  = (tGate & Cof1) | (~tGate & Cof0);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    tCur  = (tGate & Cof0) | (~tGate & Cof1);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    tGate = s_Truths6[m] & ~s_Truths6[k];
                    tCur  = (tGate & Cof1) | (~tGate & Cof0);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    tCur  = (tGate & Cof0) | (~tGate & Cof1);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    tGate = ~s_Truths6[m] & s_Truths6[k];
                    tCur  = (tGate & Cof1) | (~tGate & Cof0);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    tCur  = (tGate & Cof0) | (~tGate & Cof1);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    tGate = ~s_Truths6[m] & ~s_Truths6[k];
                    tCur  = (tGate & Cof1) | (~tGate & Cof0);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    tCur  = (tGate & Cof0) | (~tGate & Cof1);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    
                    tGate = s_Truths6[m] ^ s_Truths6[k];
                    tCur  = (tGate & Cof1) | (~tGate & Cof0);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    tGate = s_Truths6[m] ^ s_Truths6[k];
                    tCur  = (tGate & Cof0) | (~tGate & Cof1);
                    Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );

                    nSteps += 10;
                }
                } }
                }
            }
            }
        } 
        if (fUseTwo != 0) && n > 2 && (fExpand != 0) {
        for  Entry = Limit[n-2]; Entry < Limit[n-1]; Entry++  {
            var pTruth *word  = Vec_MemReadEntry( vTtMem, Entry );
            var NodSup int32  = Vec_IntEntry(vNodSup, Entry);
            var nSupp int32  = 0xF & NodSup; var g1, g2 int32;
            assert(n-2 == (NodSup >> 16), "n-2 == (NodSup >> 16)");
            assert(!Abc_Tt6HasVar(*pTruth, nSupp), "!Abc_Tt6HasVar(*pTruth, nSupp)");
            for  v = 0; v < nSupp; v++  {
            for  u = 0; u < nSupp; u++  { if u != v {
                var Cof0 word  = Abc_Tt6Cofactor0( *pTruth, v );
                var Cof1 word  = Abc_Tt6Cofactor1( *pTruth, v );

                var Cof00 word  = Abc_Tt6Cofactor0( Cof0, u );
                var Cof01 word  = Abc_Tt6Cofactor1( Cof0, u );
                var Cof10 word  = Abc_Tt6Cofactor0( Cof1, u );
                var Cof11 word  = Abc_Tt6Cofactor1( Cof1, u );

                var tGates [5]word;
                var tCur word;
                tGates[0] =  s_Truths6[v] &  s_Truths6[u];
                tGates[1] =  s_Truths6[v] & ~s_Truths6[u];
                tGates[2] = ~s_Truths6[v] &  s_Truths6[u];
                tGates[3] =  s_Truths6[v] |  s_Truths6[u];
                tGates[4] =  s_Truths6[v] ^  s_Truths6[u];

                for  g1 = 0;    g1 < 5; g1++  {
                for  g2 = g1+1; g2 < 5; g2++  {
                    Cof0  = (tGates[g1] & Cof01) | (~tGates[g1] & Cof00);
                    Cof1  = (tGates[g1] & Cof11) | (~tGates[g1] & Cof10);

                    tCur  = (tGates[g2] & Cof1)  | (~tGates[g2] & Cof0);
                    Count2 += Dau_InsertFunction( pMan, &tCur, n, nInputs, nVars, nSupp, vTtMem, vNodSup, Entry, clk );
                }
                }
            }
            }
            }
        } }
        Limit[n+1] = Vec_IntSize(vNodSup);
        nNew = Dau_PrintStats( n, nInputs, nVars, vNodSup, Limit[n], Limit[n+1], nSteps, Count2, clk );
        if nNew == 0 {
            break; }
    }
    Dau_TablesSave( nInputs, nVars, vTtMem, vNodSup, Vec_IntSize(vNodSup), clk );
    Abc_PrintTime( 1, "Total time", Abc_Clock() - clk );
    //Dau_DumpFuncs( vTtMem, vNodSup, nVars, nNodeMax );
    //Dau_ExactNpnPrint( vTtMem, vNodSup, nVars, nInputs, n );
    Abc_TtHieManStop( pMan );
    Vec_MemHashFree( vTtMem );
    Vec_MemFreeP( &vTtMem );
    Vec_IntFree( vNodSup );
    fflush(os.Stdout);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

