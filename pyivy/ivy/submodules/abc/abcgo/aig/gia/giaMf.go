package gia

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [giaMf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Cut computation.]

  Author      [Alan Mishchenko]`
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaMf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecMem.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"
//#include "sat/cnf/cnf.h"
//#include "opt/dau/dau.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Mf_SetForEachCut( pList, pCut, i )      for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += Mf_CutSize(pCut) + 1 )

// Original macro:
// #define Mf_ObjForEachCut( pCuts, i, nCuts )     for ( i = 0, i < nCuts; i++ )


// extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computing truth tables of useful DSD classes of 6-functions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var s_nCalls int32;
var s_vTtMem *Vec_Mem_t = nil;
func Mf_ManTruthCanonicize(t *word, nVars int32) int32 {
    var Temp word;
    var Best word = *t;
    var r, i, Config int32;
    for  r = 0; r < 1; r++  {
        if Best > (Temp = ~Best) {
            Best = Temp;
Config ^= (1 << nVars); }
        for  i = 0; i < nVars; i++  {
            if Best > (Temp = Abc_Tt6Flip(Best, i)) {
                Best = Temp;
Config ^= (1 << i); }
        }
    }
    *t = Best;
    if s_vTtMem == nil {
        s_vTtMem = Vec_MemAllocForTT( 6, 0 ); }
    Vec_MemHashInsert( s_vTtMem, t );
    s_nCalls++;
    return Config;
}
func Mf_ManTruthQuit() {
    if s_vTtMem == nil {
        return; }
    fmt.Printf("TT = %d (%.2f %%)\n", Vec_MemEntryNum(s_vTtMem), 100.0 * Vec_MemEntryNum(s_vTtMem) / s_nCalls);
    Vec_MemHashFree( s_vTtMem );
    Vec_MemFree( s_vTtMem );
    s_vTtMem = nil;
    s_nCalls = 0;
}

func Mf_ManTruthCollect(Limit int32) *Vec_Wrd_t {
    extern Vec_Wrd_t * Mpm_ManGetTruthWithCnf( int Limit );
    var pPerm *int32  = Extra_PermSchedule( 6 );
    var pComp *int32  = Extra_GreyCodeSchedule( 6 );
    var vTruths *Vec_Wrd_t  = Mpm_ManGetTruthWithCnf( Limit );
    var vResult *Vec_Wrd_t  = Vec_WrdAlloc( 1 << 20 );
    var uTruth, tCur, tTemp1, tTemp2 word;
    var i, p, c, k int32;
    Vec_WrdForEachEntry( vTruths, uTruth, k )
    {
        for  i = 0; i < 2; i++  {
            if i != 0 {
                tCur = ~uTruth;
            } else {
                tCur = uTruth;
            }
            tTemp1 = tCur;
            for  p = 0; p < 720; p++  {
                tTemp2 = tCur;
                for  c = 0; c < 64; c++  {
                    tCur = Abc_Tt6Flip( tCur, pComp[c] );
                    Vec_WrdPush( vResult, tCur );
                }
                assert(tTemp2 == tCur, "tTemp2 == tCur");
                tCur = Abc_Tt6SwapAdjacent( tCur, pPerm[p] );
            }
            assert(tTemp1 == tCur, "tTemp1 == tCur");
        }
    }
    pPerm = nil;
    pComp = nil;
    fmt.Printf("Original = %d.  ", Vec_WrdSize(vTruths));
    Vec_WrdFree( vTruths );
    fmt.Printf("Total = %d.  ", Vec_WrdSize(vResult));
    vTruths = Vec_WrdUniqifyHash( vResult, 1 );
    Vec_WrdFree( vResult );
    fmt.Printf("Unique = %d.  ", Vec_WrdSize(vTruths));
    Vec_WrdForEachEntry( vTruths, uTruth, k )
    {
        Mf_ManTruthCanonicize( &uTruth, 6 );
        Vec_WrdWriteEntry( vTruths, k, uTruth );
    }
    vResult = Vec_WrdUniqifyHash( vTruths, 1 );
    Vec_WrdFree( vTruths );
    fmt.Printf("Unique = %d.  \n", Vec_WrdSize(vResult));
    return vResult;
}
func Mf_ManTruthCount() int32 {
    var vTruths *Vec_Wrd_t  = Mf_ManTruthCollect( 10 );
    var RetValue int32  = Vec_WrdSize( vTruths );
    Vec_WrdFree( vTruths );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Collect truth tables used by the mapper.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mf_ManProfileTruths(p *Mf_Man_t) {
    var vCounts *Vec_Int_t;
    var i int32;
    var Entry int32;
    var pCut *int32;
    var Counter int32;
    vCounts = Vec_IntStart( Vec_IntSize(&p.vCnfSizes) );
    Gia_ManForEachAndId( p.pGia, i )
    {
        if !(Mf_ObjMapRefNum(p, i) != 0) {
            continue; }
        pCut = Mf_ObjCutBest( p, i );
        Vec_IntAddToEntry( vCounts, Abc_Lit2Var(Mf_CutFunc(pCut)), 1 );
    }
    Vec_IntForEachEntry( vCounts, Entry, i )
    {
        if Entry == 0 {
            continue; }
        fmt.Printf("%6d : ", Counter++);
        fmt.Printf("%6d : ", i);
        fmt.Printf("Occur = %4d  ", Entry); 
        fmt.Printf("CNF size = %2d  ", Vec_IntEntry(&p.vCnfSizes, i));
        Dau_DsdPrintFromTruth( Vec_MemReadEntry(p.vTtMem, i), p.pPars.nLutSize );
    }
    Vec_IntFree( vCounts );
}

/**Function*************************************************************

  Synopsis    [Derives CNFs for each function used in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    [Derives CNF for the AIG using the mapping.]

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



func Mf_CutPrint(p *Mf_Man_t, pCut *Mf_Cut_t) {
    var i int32;
    var nDigits int32 = Abc_Base10Log(Gia_ManObjNum(p.pGia)); 
    fmt.Printf("%d  {", pCut.nLeaves);
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        fmt.Printf(" %*d", nDigits, pCut.pLeaves[i]);
    }
    for  ; i < int32(p.pPars.nLutSize); i++  {
        fmt.Printf(" %*s", nDigits, " ");
    }
    fmt.Printf("  }   D = %4d  A = %9.4f  F = %6d  ", pCut.Delay, pCut.Flow, pCut.iFunc);
    if p.vTtMem {
        if p.pPars.fGenCnf {
            fmt.Printf("CNF = %2d  ", Vec_IntEntry(&p.vCnfSizes, Abc_Lit2Var(pCut.iFunc))); }
        if p.pPars.fGenLit {
            fmt.Printf("Lit = %2d  ", Vec_IntEntry(&p.vCnfSizes, Abc_Lit2Var(pCut.iFunc))); }
        Dau_DsdPrintFromTruth( Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut.iFunc)), pCut.nLeaves );
    
    } else {
        fmt.Printf("\n"); }
}




/**Function*************************************************************

  Synopsis    [Check correctness of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func Mf_SetCheckArray(ppCuts **Mf_Cut_t, nCuts int32) int32 {
    var pCut0, pCut1 *Mf_Cut_t; 
    var i, k, m, n, Value int32;
    assert(nCuts > 0, "nCuts > 0");
    for  i = 0; i < nCuts; i++  {
        pCut0 = ppCuts[i];
        assert(pCut0.nLeaves <= MF_LEAF_MAX, "pCut0.nLeaves <= MF_LEAF_MAX");
        assert(pCut0.Sign == Mf_CutGetSign(pCut0.pLeaves, pCut0.nLeaves), "pCut0.Sign == Mf_CutGetSign(pCut0.pLeaves, pCut0.nLeaves)");
        // check duplicates
        for  m = 0; m < int32(pCut0.nLeaves); m++  {
        for  n = m + 1; n < int32(pCut0.nLeaves); n++  {
            assert(pCut0.pLeaves[m] < pCut0.pLeaves[n], "pCut0.pLeaves[m] < pCut0.pLeaves[n]");
        }
        }
        // check pairs
        for  k = 0; k < nCuts; k++  {
            pCut1 = ppCuts[k];
            if pCut0 == pCut1 {
                continue; }
            // check containments
            Value = Mf_CutCheck( pCut0, pCut1 );
            assert(Value == 0, "Value == 0");
        }
    }
    return 1;
}


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

  Synopsis    [Flow and area.]

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
func Mf_ManPrintFanoutProfile(p *Mf_Man_t, vFanCounts *Vec_Int_t) {
    var pGia *Gia_Man_t  = p.pGia0;
    var i int32;
    var Count int32;
    var nMax int32 = Vec_IntFindMax( vFanCounts );
    var vCounts *Vec_Int_t  = Vec_IntStart( nMax + 1 );
    Vec_IntForEachEntry( vFanCounts, Count, i )
        if ( Count && Gia_ObjIsAnd(Gia_ManObj(pGia, i)) ) 
            Vec_IntAddToEntry( vCounts, Count, 1 );
    fmt.Printf("\nFanout distribution for internal nodes:\n");
    Vec_IntForEachEntry( vCounts, Count, i )
        if ( Count ) fmt.Printf("Fanout = %5d : Nodes = %5d.\n", i, Count);
    fmt.Printf("Total nodes with fanout = %d. Max fanout = %d.\n\n", Vec_IntCountPositive(vCounts), nMax);
    Vec_IntFree( vCounts );
}
func Mf_ManPrintMfccStats(p *Mf_Man_t, iObj int32) int32 {
    var pGia *Gia_Man_t  = p.pGia0;
    var Area int32;
    fmt.Printf("%5d : Level = %5d  Refs = %5d  Mffc = %5d\n", iObj, Gia_ObjLevelId(pGia, iObj), Mf_ObjMapRefNum(p, iObj), (Area = Mf_CutAreaMffc(p, iObj)));
    return Area;
}
func Mf_ManOptimizationOne(p *Mf_Man_t, iObj int32) {
    var pGia *Gia_Man_t  = p.pGia0;
    var ppCuts [32]*int32;
    var nCuts int32;
    var iFanout, i, nAreaSum, nAreaBest int32;
    // skip pivots whose MFFC fanouts are pointed to by COs
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        if ( Gia_ObjIsCo(Gia_ManObj(pGia, iFanout)) )
            return;
    // the pivot is used in the mapping as well as all of its fanouts
    assert(Mf_ObjMapRefNum(p, iObj) > 1, "Mf_ObjMapRefNum(p, iObj) > 1");
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        assert(Mf_ObjMapRefNum(p, iFanout) > 0, "Mf_ObjMapRefNum(p, iFanout) > 0");
    // print this pivot and its fanouts
    fmt.Printf("\nPivot node = %d\n", iObj);
    fmt.Printf("Pivot ");
Mf_ManPrintMfccStats( p, iObj );
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        fmt.Printf("Node  ");
nAreaSum += Mf_ManPrintMfccStats( p, iFanout );
    // calculate the shared MFFC
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        Mf_ObjMapRefInc( p, iFanout );
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        ppCuts[nCuts++] = Mf_ObjCutBest( p, iFanout );
    nAreaBest = Mf_CutAreaRefed2Multi( p, iObj, ppCuts, nCuts );
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        Mf_ObjMapRefDec( p, iFanout );
    fmt.Printf("Sum of MFFC sizes = %d\n", nAreaSum);
    fmt.Printf("Shared MFFC size  = %d\n", nAreaBest);
}
func Mf_ManOptimization(p *Mf_Man_t) {
    var nOutMax int32  = 3;
    var pGia *Gia_Man_t  = p.pGia0;
    var i int32;
    var Count int32;
    var nNodes int32 = Mf_ManMappingFromMapping( p );
    Gia_ManLevelNum( pGia );
    Gia_ManStaticMappingFanoutStart( pGia );
    Mf_ManPrintFanoutProfile( p, pGia.vFanoutNums );
    fmt.Printf("\nIndividual logic cones for mapping with %d nodes:\n", nNodes);
    Vec_IntForEachEntry( pGia.vFanoutNums, Count, i )
        if ( Count >= 2 && Count <= nOutMax && Gia_ObjIsAnd(Gia_ManObj(pGia, i)) )
            Mf_ManOptimizationOne( p, i );
    fmt.Printf("\nFinished printing individual logic cones.\n");
    Gia_ManStaticFanoutStop( pGia );
    Vec_IntFreeP( &pGia.vMapping );
}

/**Function*************************************************************

  Synopsis    [Technology mappping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [CNF generation]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

