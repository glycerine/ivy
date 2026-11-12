package if_

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [ifTruth.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Computation of truth tables of the cuts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifTruth.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

//#define IF_TRY_NEW

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sorts the pins in the decreasing order of delays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutTruthPermute(pTruth *word, nLeaves int32, nVars int32, nWords int32, pDelays *float32, pVars *int32) {
    for {
        var i, fChange int32;
        for  i = 0; i < nLeaves - 1; i++  {
            if pDelays[i] >= pDelays[i+1] {
                continue; }
            ABC_SWAP( float, pDelays[i], pDelays[i+1] );
            ABC_SWAP( int, pVars[i], pVars[i+1] );
            if pTruth != nil {
                Abc_TtSwapAdjacent( pTruth, nWords, i ); }
            fChange = 1;
        }
        if !(fChange != 0) {
            return; }
    }
}
func If_CutRotatePins(p *If_Man_t, pCut *If_Cut_t) {
    var pLeaf *If_Obj_t;
    var PinDelays [IF_MAX_LUTSIZE]float32;
    var i, truthId int32;
    assert(!p.pPars.fUseTtPerm, "!p.pPars.fUseTtPerm");
    If_CutForEachLeaf( p, pCut, pLeaf, i )
        PinDelays[i] = If_ObjCutBest(pLeaf).Delay;
    if p.vTtMem[pCut.nLeaves] == nil {
        If_CutTruthPermute( nil, If_CutLeaveNum(pCut), pCut.nLeaves, p.nTruth6Words[pCut.nLeaves], PinDelays, If_CutLeaves(pCut) );
        return;
    }
    Abc_TtCopy( p.puTempW, If_CutTruthWR(p, pCut), p.nTruth6Words[pCut.nLeaves], 0 );
    If_CutTruthPermute( p.puTempW, If_CutLeaveNum(pCut), pCut.nLeaves, p.nTruth6Words[pCut.nLeaves], PinDelays, If_CutLeaves(pCut) );
    truthId        = Vec_MemHashInsert( p.vTtMem[pCut.nLeaves], p.puTempW );
    pCut.iCutFunc = Abc_Var2Lit( truthId, If_CutTruthIsCompl(pCut) );
    assert((p.puTempW[0] & 1) == 0, "(p.puTempW[0] & 1) == 0");
}

/**Function*************************************************************

  Synopsis    [Truth table computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutComputeTruth(p *If_Man_t, pCut *If_Cut_t, pCut0 *If_Cut_t, pCut1 *If_Cut_t, fCompl0 int32, fCompl1 int32) int32 {
    var fCompl, truthId, nLeavesNew, PrevSize, RetValue int32;
    var pTruth0s *word  = Vec_MemReadEntry( p.vTtMem[pCut0.nLeaves], Abc_Lit2Var(pCut0.iCutFunc) );
    var pTruth1s *word  = Vec_MemReadEntry( p.vTtMem[pCut1.nLeaves], Abc_Lit2Var(pCut1.iCutFunc) );
    var pTruth0 *word   = (word *)p.puTemp[0];
    var pTruth1 *word   = (word *)p.puTemp[1];
    var pTruth *word    = (word *)p.puTemp[2];
    Abc_TtCopy( pTruth0, pTruth0s, p.nTruth6Words[pCut0.nLeaves], fCompl0 ^ pCut0.fCompl ^ Abc_LitIsCompl(pCut0.iCutFunc) );
    Abc_TtCopy( pTruth1, pTruth1s, p.nTruth6Words[pCut1.nLeaves], fCompl1 ^ pCut1.fCompl ^ Abc_LitIsCompl(pCut1.iCutFunc) );
    Abc_TtStretch6( pTruth0, pCut0.nLeaves, pCut.nLeaves );
    Abc_TtStretch6( pTruth1, pCut1.nLeaves, pCut.nLeaves );
    Abc_TtExpand( pTruth0, pCut.nLeaves, pCut0.pLeaves, pCut0.nLeaves, pCut.pLeaves, pCut.nLeaves );
    Abc_TtExpand( pTruth1, pCut.nLeaves, pCut1.pLeaves, pCut1.nLeaves, pCut.pLeaves, pCut.nLeaves );
    fCompl         = (pTruth0[0] & pTruth1[0] & 1);
    Abc_TtAnd( pTruth, pTruth0, pTruth1, p.nTruth6Words[pCut.nLeaves], fCompl );
    if p.pPars.fCutMin && (pCut0.nLeaves + pCut1.nLeaves > pCut.nLeaves || pCut0.nLeaves == 0 || pCut1.nLeaves == 0) {
        nLeavesNew = Abc_TtMinBase( pTruth, pCut.pLeaves, pCut.nLeaves, pCut.nLeaves );
        if nLeavesNew < If_CutLeaveNum(pCut) {
            pCut.nLeaves = nLeavesNew;
            pCut.uSign   = If_ObjCutSignCompute( pCut );
            RetValue      = 1;
        }
    }
    PrevSize       = Vec_MemEntryNum( p.vTtMem[pCut.nLeaves] );   
    truthId        = Vec_MemHashInsert( p.vTtMem[pCut.nLeaves], pTruth );
    pCut.iCutFunc = Abc_Var2Lit( truthId, fCompl );
    assert((pTruth[0] & 1) == 0, "(pTruth[0] & 1) == 0");
#ifdef IF_TRY_NEW
    {
        word pCopy[1024];
        char pCanonPerm[16];
        memcpy( pCopy, If_CutTruthW(pCut), sizeof(word) * p->nTruth6Words[pCut->nLeaves] );
        Abc_TtCanonicize( pCopy, pCut->nLeaves, pCanonPerm );
    }
#endif
    if p.vTtIsops[pCut.nLeaves] && PrevSize != Vec_MemEntryNum(p.vTtMem[pCut.nLeaves]) {
        var vLevel *Vec_Int_t  = Vec_WecPushLevel( p.vTtIsops[pCut.nLeaves] );
        fCompl = Kit_TruthIsop( (unsigned *)pTruth, pCut.nLeaves, p.vCover, 1 );
        if fCompl >= 0 {
            Vec_IntGrow( vLevel, Vec_IntSize(p.vCover) );
            Vec_IntAppend( vLevel, p.vCover );
            if fCompl {
                vLevel.nCap ^= (1<<16); } // hack to remember complemented attribute
        }
        assert(Vec_WecSize(p.vTtIsops[pCut.nLeaves]) == Vec_MemEntryNum(p.vTtMem[pCut.nLeaves]), "Vec_WecSize(p.vTtIsops[pCut.nLeaves]) == Vec_MemEntryNum(p.vTtMem[pCut.nLeaves])");
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Truth table computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutComputeTruthPerm_int(p *If_Man_t, pCut *If_Cut_t, pCut0 *If_Cut_t, pCut1 *If_Cut_t, iCutFunc0 int32, iCutFunc1 int32) int32 {
    var fVerbose int32;
    var clk abctime;
    var pPerm [IF_MAX_LUTSIZE]int32;
    var v, Place, fCompl, truthId, nLeavesNew, RetValue int32;
    var pTruth0s *word  = Vec_MemReadEntry( p.vTtMem[pCut0.nLeaves], Abc_Lit2Var(iCutFunc0) );
    var pTruth1s *word  = Vec_MemReadEntry( p.vTtMem[pCut1.nLeaves], Abc_Lit2Var(iCutFunc1) );
    var pTruth0 *word   = (word *)p.puTemp[0];
    var pTruth1 *word   = (word *)p.puTemp[1];
    var pTruth *word    = (word *)p.puTemp[2];
    assert(pCut0.uMaskFunc >= 0, "pCut0.uMaskFunc >= 0");
    assert(pCut1.uMaskFunc >= 0, "pCut1.uMaskFunc >= 0");
    Abc_TtCopy( pTruth0, pTruth0s, p.nTruth6Words[pCut0.nLeaves], Abc_LitIsCompl(iCutFunc0) );
    Abc_TtCopy( pTruth1, pTruth1s, p.nTruth6Words[pCut1.nLeaves], Abc_LitIsCompl(iCutFunc1) );
    Abc_TtStretch6( pTruth0, pCut0.nLeaves, pCut.nLeaves );
    Abc_TtStretch6( pTruth1, pCut1.nLeaves, pCut.nLeaves );

if fVerbose != 0 {
//Kit_DsdPrintFromTruth( pTruth0, pCut0->nLeaves ); printf( "\n" );
//Kit_DsdPrintFromTruth( pTruth1, pCut1->nLeaves ); printf( "\n" );
}
    // create literals
    for  v = 0; v < int32(pCut0.nLeaves); v++  {
        pCut.pLeaves[v] = Abc_Var2Lit( pCut0.pLeaves[v], If_CutLeafBit(pCut0, v) );
    }
    for  v = 0; v < int32(pCut1.nLeaves); v++  {
        if p.pPerm[1][v] >= int32(pCut0.nLeaves) {
            pCut.pLeaves[p.pPerm[1][v]] = Abc_Var2Lit( pCut1.pLeaves[v], If_CutLeafBit(pCut1, v) );
        } else if If_CutLeafBit(pCut0, p.pPerm[1][v]) != If_CutLeafBit(pCut1, v) {
            Abc_TtFlip( pTruth1, p.nTruth6Words[pCut1.nLeaves], v ); }
    }  
    // permute variables
    for  v = int32(pCut1.nLeaves); v < int32(pCut.nLeaves); v++  {
        p.pPerm[1][v] = -1;
    }
    for  v = 0; v < int32(pCut1.nLeaves); v++  {
        Place = p.pPerm[1][v];
        if Place == v || Place == -1 {
            continue; }
        Abc_TtSwapVars( pTruth1, pCut.nLeaves, v, Place );
        p.pPerm[1][v] = p.pPerm[1][Place];
        p.pPerm[1][Place] = Place;
        v--;
    }

if fVerbose != 0 {
//Kit_DsdPrintFromTruth( pTruth0, pCut0->nLeaves ); printf( "\n" );
//Kit_DsdPrintFromTruth( pTruth1, pCut->nLeaves ); printf( "\n" );
}

    // perform operation
    Abc_TtAnd( pTruth, pTruth0, pTruth1, p.nTruth6Words[pCut.nLeaves], 0 );
    // minimize support
    if p.pPars.fCutMin && (pCut0.nLeaves + pCut1.nLeaves > pCut.nLeaves || pCut0.nLeaves == 0 || pCut1.nLeaves == 0) {
        nLeavesNew = Abc_TtMinBase( pTruth, pCut.pLeaves, pCut.nLeaves, pCut.nLeaves );
        if nLeavesNew < If_CutLeaveNum(pCut) {
            pCut.nLeaves = nLeavesNew;
            RetValue      = 1;
        }
    }
    // compute canonical form
if p.pPars.fVerbose {
clk = Abc_Clock(); }
    p.uCanonPhase = Abc_TtCanonicize( pTruth, pCut.nLeaves, p.pCanonPerm );
if p.pPars.fVerbose {
p.timeCache[3] += Abc_Clock() - clk; }
    for  v = 0; v < int32(pCut.nLeaves); v++  {
        pPerm[v] = Abc_LitNotCond( pCut.pLeaves[int32(p.pCanonPerm[v])], ((p.uCanonPhase>>v)&1) );
    }
    pCut.uMaskFunc = 0;
    for  v = 0; v < int32(pCut.nLeaves); v++  {
        pCut.pLeaves[v] = Abc_Lit2Var(pPerm[v]);
        if Abc_LitIsCompl(pPerm[v]) {
            pCut.uMaskFunc |= (1 << v); }
    }
    // create signature after lowering literals
    if RetValue != 0 {
        pCut.uSign = If_ObjCutSignCompute( pCut );
    } else {
        assert(pCut.uSign == If_ObjCutSignCompute( pCut ), "pCut.uSign == If_ObjCutSignCompute( pCut )"); }

    assert(Vec_IntSize(p.vTtOccurs[pCut.nLeaves]) == Vec_MemEntryNum(p.vTtMem[pCut.nLeaves]), "Vec_IntSize(p.vTtOccurs[pCut.nLeaves]) == Vec_MemEntryNum(p.vTtMem[pCut.nLeaves])");
    // hash function
    fCompl         = ((p.uCanonPhase >> pCut.nLeaves) & 1);
    truthId        = Vec_MemHashInsert( p.vTtMem[pCut.nLeaves], pTruth );
    pCut.iCutFunc = Abc_Var2Lit( truthId, fCompl );
    // count how many time this truth table is used
    if Vec_IntSize(p.vTtOccurs[pCut.nLeaves]) < Vec_MemEntryNum(p.vTtMem[pCut.nLeaves]) {
        Vec_IntPush( p.vTtOccurs[pCut.nLeaves], 0 ); }
    Vec_IntAddToEntry( p.vTtOccurs[pCut.nLeaves], truthId, 1 );

if fVerbose != 0 {
//Kit_DsdPrintFromTruth( pTruth, pCut->nLeaves ); printf( "\n" );
//If_CutPrint( pCut0 );
//If_CutPrint( pCut1 );
//If_CutPrint( pCut );
//printf( "%d\n\n", pCut->iCutFunc );
}

    return RetValue;
}
func If_CutComputeTruthPerm(p *If_Man_t, pCut *If_Cut_t, pCut0 *If_Cut_t, pCut1 *If_Cut_t, iCutFunc0 int32, iCutFunc1 int32) int32 {
    var clk abctime;
    var i, Num, nEntriesOld, RetValue int32;
    if pCut0.nLeaves + pCut1.nLeaves > pCut.nLeaves || iCutFunc0 < 2 || iCutFunc1 < 2 {
if p.pPars.fVerbose {
clk = Abc_Clock(); }
        RetValue = If_CutComputeTruthPerm_int( p, pCut, pCut0, pCut1, iCutFunc0, iCutFunc1 );
if p.pPars.fVerbose {
p.timeCache[0] += Abc_Clock() - clk; }
        return RetValue;
    }
    assert(pCut0.nLeaves + pCut1.nLeaves == pCut.nLeaves, "pCut0.nLeaves + pCut1.nLeaves == pCut.nLeaves");
    nEntriesOld = Hash_IntManEntryNum(p.vPairHash);
    Num = Hash_Int2ManInsert( p.vPairHash, (iCutFunc0 << 5)|pCut0.nLeaves, (iCutFunc1 << 5)|pCut1.nLeaves, -1 );
    assert(Num > 0, "Num > 0");
    if nEntriesOld == Hash_IntManEntryNum(p.vPairHash) {
        var pCanonPerm string;
        var v int32;
        var pPerm [IF_MAX_LUTSIZE]int32;
        pCut.iCutFunc = Vec_IntEntry( p.vPairRes, Num );
        // move complements from the fanin cuts
        for  v = 0; v < int32(pCut.nLeaves); v++  {
            if v < int32(pCut0.nLeaves) {
                pCut.pLeaves[v] = Abc_Var2Lit( pCut.pLeaves[v], If_CutLeafBit(pCut0, v) );
            } else {
                pCut.pLeaves[v] = Abc_Var2Lit( pCut.pLeaves[v], If_CutLeafBit(pCut1, v-int32(pCut0.nLeaves)) ); }
        }
        // reorder the cut
        pCanonPerm = Vec_StrEntryP( p.vPairPerms, Num * pCut.nLimit );
        for  v = 0; v < int32(pCut.nLeaves); v++  {
            pPerm[v] = Abc_LitNotCond( pCut.pLeaves[Abc_Lit2Var(int32(pCanonPerm[v]))], Abc_LitIsCompl(int32(pCanonPerm[v])) );
        }
        // generate the result
        pCut.uMaskFunc = 0;
        for  v = 0; v < int32(pCut.nLeaves); v++  {
            pCut.pLeaves[v] = Abc_Lit2Var(pPerm[v]);
            if Abc_LitIsCompl(pPerm[v]) {
                pCut.uMaskFunc |= (1 << v); }
        }
//        printf( "Found: %d(%d) %d(%d) -> %d(%d)\n", iCutFunc0, pCut0->nLeaves, iCutFunc1, pCut0->nLeaves, pCut->iCutFunc, pCut->nLeaves );
        p.nCacheHits++;
//p->timeCache[1] += Abc_Clock() - clk;
        return 0;
    }
if p.pPars.fVerbose {
clk = Abc_Clock(); }
    p.nCacheMisses++;
    RetValue = If_CutComputeTruthPerm_int( p, pCut, pCut0, pCut1, iCutFunc0, iCutFunc1 );
    assert(RetValue == 0, "RetValue == 0");
//    printf( "Added: %d(%d) %d(%d) -> %d(%d)\n", iCutFunc0, pCut0->nLeaves, iCutFunc1, pCut0->nLeaves, pCut->iCutFunc, pCut->nLeaves );
    // save the result
    assert(Num == Vec_IntSize(p.vPairRes), "Num == Vec_IntSize(p.vPairRes)");
    Vec_IntPush( p.vPairRes, pCut.iCutFunc );
    // save the permutation
    assert(Num * int32(pCut.nLimit) == Vec_StrSize(p.vPairPerms), "Num * int32(pCut.nLimit) == Vec_StrSize(p.vPairPerms)");
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Vec_StrPush( p.vPairPerms, byte(Abc_Var2Lit(int32(p.pCanonPerm[i]), ((p.uCanonPhase>>i)&1))) );
    }
    for  i = int32(pCut0.nLeaves) + int32(pCut1.nLeaves); i < int32(pCut.nLimit); i++  {
        Vec_StrPush( p.vPairPerms, byte(-1) );
    }
if p.pPars.fVerbose {
p.timeCache[2] += Abc_Clock() - clk; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Check the function of 6-input LUT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DeriveHashTable6(nVars int32, uTruth word) *Vec_Mem_t {
    var fVerbose int32;
    var nMints int32    = (1 << nVars);
    var nPerms int32    = Extra_Factorial( nVars );
    var pComp *int32   = Extra_GreyCodeSchedule( nVars );
    var pPerm *int32   = Extra_PermSchedule( nVars );
    var Canon word    = ABC_CONST(0xFFFFFFFFFFFFFFFF);
    var tCur, tTemp1, tTemp2 word;
    var vTtMem6 *Vec_Mem_t  = Vec_MemAllocForTTSimple( nVars );
    var i, p, c int32;
    for  i = 0; i < 2; i++  {
        if i {
            tCur = ~uTruth;
        } else {
            tCur = uTruth;
        }
        tTemp1 = tCur;
        for  p = 0; p < nPerms; p++  {
            tTemp2 = tCur;
            for  c = 0; c < nMints; c++  {
                if Canon > tCur {
                    Canon = tCur; }
                Vec_MemHashInsert( vTtMem6, &tCur );
                tCur = Abc_Tt6Flip( tCur, pComp[c] );
            }
            assert(tTemp2 == tCur, "tTemp2 == tCur");
            tCur = Abc_Tt6SwapAdjacent( tCur, pPerm[p] );
        }
        assert(tTemp1 == tCur, "tTemp1 == tCur");
    }
    ABC_FREE( pComp );
    ABC_FREE( pPerm );
    if fVerbose != 0 {
/*
        word * pEntry;
        Vec_MemForEachEntry( vTtMem6, pEntry, i )
        {
            Extra_PrintHex( stdout, (unsigned*)pEntry, nVars );  
            printf( ", " );
            if ( i % 4 == 3 )
                printf( "\n" );
        }
*/
        Extra_PrintHex( os.Stdout, (unsigned*)&uTruth, nVars );  fmt.Printf("\n");
        Extra_PrintHex( os.Stdout, (unsigned*)&Canon,  nVars );  fmt.Printf("\n");
        fmt.Printf("Members = %d.\n", Vec_MemEntryNum(vTtMem6));
    }
    return vTtMem6;
}

func If_CutCheckTruth6(p *If_Man_t, pCut *If_Cut_t) int32 {
    if pCut.nLeaves != 6 {
        return 0; }
    if p.vTtMem6 == nil {
        p.vTtMem6 = If_DeriveHashTable6( 6, ABC_CONST(0xfedcba9876543210) ); }
    if *Vec_MemHashLookup( p.vTtMem6, If_CutTruthWR(p, pCut) ) == -1 {
        return 0; }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

