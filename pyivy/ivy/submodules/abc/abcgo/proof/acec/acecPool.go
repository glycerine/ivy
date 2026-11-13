package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecPool.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecPool.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecWec.h"
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
func Acec_ManCountXorTreeInputs_rec(p *Gia_Man_t, Node int32, vXors *Vec_Int_t, vXorMap *Vec_Int_t, vIsCarryRoot *Vec_Bit_t, vCarryRootSet *Vec_Int_t, vXorSet *Vec_Int_t) {
    var k, iXorBox int32;
    if Node == 0 || Gia_ObjIsTravIdCurrentId(p, Node) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, Node);
    iXorBox = Vec_IntEntry( vXorMap, Node );
    if iXorBox == -1 {
        if Vec_BitEntry(vIsCarryRoot, Node) {
            Vec_IntPush( vCarryRootSet, Node ); }
        return;
    }
    for  k = 1; k < 4; k++  {
        Acec_ManCountXorTreeInputs_rec( p, Vec_IntEntry(vXors, 4*iXorBox+k), vXors, vXorMap, vIsCarryRoot, vCarryRootSet, vXorSet );
    }
    Vec_IntPush( vXorSet, Vec_IntEntry(vXors, 4*iXorBox) );
}
func Acec_ManCollectCarryRootSets(p *Gia_Man_t, vAdds *Vec_Int_t, vCarryMap *Vec_Int_t, vXors *Vec_Int_t, vXorRoots *Vec_Int_t, vCarryRoots *Vec_Int_t) *Vec_Wec_t {
    var vCarryRootSets *Vec_Wec_t  = Vec_WecAlloc( 100 ); // XorBoxes, CarryRoots, AdderBoxes, Ins/Ranks, Outs/Ranks
    var vCarryRootSet *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vIsCarryRoot *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var vXorSet *Vec_Int_t = Vec_IntAlloc( 100 );
    var vLevel *Vec_Int_t;
    var i, k, XorRoot, CarryRoot int32;
    // map XORs into their cuts
    var vXorMap *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );  
    for  i = 0; 4*i < Vec_IntSize(vXors); i++  {
        Vec_IntWriteEntry( vXorMap, Vec_IntEntry(vXors, 4*i), i );
    }
    // create carry root marks
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCarryRoots)) && (((CarryRoot) = Vec_IntEntry(vCarryRoots, i)), 1); i++  {
        Vec_BitWriteEntry( vIsCarryRoot, CarryRoot, 1 );
    }
    // collect carry roots attached to each XOR root
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXorRoots)) && (((XorRoot) = Vec_IntEntry(vXorRoots, i)), 1); i++  {
        Vec_IntClear( vXorSet );
        Vec_IntClear( vCarryRootSet );
        Gia_ManIncrementTravId( p );
        Acec_ManCountXorTreeInputs_rec( p, XorRoot, vXors, vXorMap, vIsCarryRoot, vCarryRootSet, vXorSet );
        // skip trivial
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vCarryRootSet)) && (((CarryRoot) = Vec_IntEntry(vCarryRootSet, k)), 1); k++  {
            if Acec_ManCheckCarryMap(p, CarryRoot, vAdds, vCarryMap) != 0 {
                break; }
        }
        if k == Vec_IntSize(vCarryRootSet) {
            continue; }
        Vec_IntSort( vCarryRootSet, 0 );
        vLevel = Vec_WecPushLevel( vCarryRootSets );
        Vec_IntAppend( vLevel, vXorSet );
        vLevel = Vec_WecPushLevel( vCarryRootSets );
        Vec_IntAppend( vLevel, vCarryRootSet );
        vLevel = Vec_WecPushLevel( vCarryRootSets );
        vLevel = Vec_WecPushLevel( vCarryRootSets );
        vLevel = Vec_WecPushLevel( vCarryRootSets );
        // unmark carry root set
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vCarryRootSet)) && (((CarryRoot) = Vec_IntEntry(vCarryRootSet, k)), 1); k++  {
            assert(Vec_BitEntry(vIsCarryRoot, CarryRoot), "Vec_BitEntry(vIsCarryRoot, CarryRoot)");
            Vec_BitWriteEntry( vIsCarryRoot, CarryRoot, 0 );
        }
    }
    Vec_IntFree( vCarryRootSet );
    Vec_IntFree( vXorSet );
    Vec_IntFree( vXorMap );
    // collect unmarked carry roots
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vCarryRoots)) && (((CarryRoot) = Vec_IntEntry(vCarryRoots, k)), 1); k++  {
        if !Vec_BitEntry(vIsCarryRoot, CarryRoot) {
            continue; }
        if !(Acec_ManCheckCarryMap(p, CarryRoot, vAdds, vCarryMap) != 0) {
            continue; }
        vLevel = Vec_WecPushLevel( vCarryRootSets );
        vLevel = Vec_WecPushLevel( vCarryRootSets );
        Vec_IntFill( vLevel, 1, CarryRoot );
        vLevel = Vec_WecPushLevel( vCarryRootSets );
        vLevel = Vec_WecPushLevel( vCarryRootSets );
        vLevel = Vec_WecPushLevel( vCarryRootSets );
    }
    Vec_BitFree( vIsCarryRoot );
    return vCarryRootSets;
}
func Acec_ManCompareTwo(pPair0 *int32, pPair1 *int32) int32 {
    if pPair0[1] < pPair1[1] { return -1; }
    if pPair0[1] > pPair1[1] { return  1; }
    return 0;
}
func Acec_ManCollectInsOuts(p *Gia_Man_t, vAdds *Vec_Int_t, vBoxes *Vec_Int_t, vBoxRanks *Vec_Int_t, vBoxIns *Vec_Bit_t, vBoxOuts *Vec_Bit_t, vResIns *Vec_Int_t, vResOuts *Vec_Int_t) {
    var i, k, iBox, iObj, Rank, RankMax int32;
    // mark box inputs/outputs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vBoxes)) && (((iBox) = Vec_IntEntry(vBoxes, i)), 1); i++  {
        Vec_BitWriteEntry( vBoxIns, Vec_IntEntry(vAdds, 6*iBox+0), 1 );
        Vec_BitWriteEntry( vBoxIns, Vec_IntEntry(vAdds, 6*iBox+1), 1 );
        Vec_BitWriteEntry( vBoxIns, Vec_IntEntry(vAdds, 6*iBox+2), 1 );
        Vec_BitWriteEntry( vBoxOuts, Vec_IntEntry(vAdds, 6*iBox+3), 1 );
        Vec_BitWriteEntry( vBoxOuts, Vec_IntEntry(vAdds, 6*iBox+4), 1 );
    }
    // collect unmarked inputs/output and their ranks
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vBoxes)) && (((iBox) = Vec_IntEntry(vBoxes, i)), 1); i++  {
        for  k = 0; k < 3; k++  {
            if !Vec_BitEntry(vBoxOuts, Vec_IntEntry(vAdds, 6*iBox+k)) {
                Vec_IntPushTwo( vResIns, Vec_IntEntry(vAdds, 6*iBox+k), Vec_IntEntry(vBoxRanks, iBox) ); }
        }
        for  k = 3; k < 5; k++  {
            if Vec_IntEntry(vAdds, 6*iBox+k) && !Vec_BitEntry(vBoxIns, Vec_IntEntry(vAdds, 6*iBox+k)) {
                Vec_IntPushTwo( vResOuts, Vec_IntEntry(vAdds, 6*iBox+k), Vec_IntEntry(vBoxRanks, iBox)-int32((k==4)) ); }
        }
    }
    // unmark box inputs/outputs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vBoxes)) && (((iBox) = Vec_IntEntry(vBoxes, i)), 1); i++  {
        Vec_BitWriteEntry( vBoxIns, Vec_IntEntry(vAdds, 6*iBox+0), 0 );
        Vec_BitWriteEntry( vBoxIns, Vec_IntEntry(vAdds, 6*iBox+1), 0 );
        Vec_BitWriteEntry( vBoxIns, Vec_IntEntry(vAdds, 6*iBox+2), 0 );
        Vec_BitWriteEntry( vBoxOuts, Vec_IntEntry(vAdds, 6*iBox+3), 0 );
        Vec_BitWriteEntry( vBoxOuts, Vec_IntEntry(vAdds, 6*iBox+4), 0 );
    }
    // normalize ranks
    // inlining of macro Vec_IntForEachEntryDouble
    for  k = 0; (k+1 < Vec_IntSize(vResIns)) && (((iObj) = Vec_IntEntry(vResIns, k)), 1) && (((Rank) = Vec_IntEntry(vResIns, k+1)), 1); k += 2  {
        RankMax = Abc_MaxInt( RankMax, Rank );
    }
    // inlining of macro Vec_IntForEachEntryDouble
    for  k = 0; (k+1 < Vec_IntSize(vResOuts)) && (((iObj) = Vec_IntEntry(vResOuts, k)), 1) && (((Rank) = Vec_IntEntry(vResOuts, k+1)), 1); k += 2  {
        RankMax = Abc_MaxInt( RankMax, Rank );
    }
    // inlining of macro Vec_IntForEachEntryDouble
    for  k = 0; (k+1 < Vec_IntSize(vResIns)) && (((iObj) = Vec_IntEntry(vResIns, k)), 1) && (((Rank) = Vec_IntEntry(vResIns, k+1)), 1); k += 2  {
        Vec_IntWriteEntry( vResIns, k+1, 1 + RankMax - Rank );
    }
    // inlining of macro Vec_IntForEachEntryDouble
    for  k = 0; (k+1 < Vec_IntSize(vResOuts)) && (((iObj) = Vec_IntEntry(vResOuts, k)), 1) && (((Rank) = Vec_IntEntry(vResOuts, k+1)), 1); k += 2  {
        Vec_IntWriteEntry( vResOuts, k+1, 1 + RankMax - Rank );
    }
    // sort by rank
    qsort( Vec_IntArray(vResIns),  (size_t)(Vec_IntSize(vResIns)/2),  8, (int (*)(void *, void *))Acec_ManCompareTwo );
    qsort( Vec_IntArray(vResOuts), (size_t)(Vec_IntSize(vResOuts)/2), 8, (int (*)(void *, void *))Acec_ManCompareTwo );
}
func Acec_ManCollectBoxSets_rec(p *Gia_Man_t, Carry int32, iRank int32, vAdds *Vec_Int_t, vCarryMap *Vec_Int_t, vBoxes *Vec_Int_t, vBoxRanks *Vec_Int_t) {
    var iBox int32  = Vec_IntEntry( vCarryMap, Carry );
    if iBox == -1 {
        return; }
    Acec_ManCollectBoxSets_rec( p, Vec_IntEntry(vAdds, 6*iBox+0), iRank+1, vAdds, vCarryMap, vBoxes, vBoxRanks );
    Acec_ManCollectBoxSets_rec( p, Vec_IntEntry(vAdds, 6*iBox+1), iRank+1, vAdds, vCarryMap, vBoxes, vBoxRanks );
    if Vec_IntEntry(vAdds, 6*iBox+2) {
        Acec_ManCollectBoxSets_rec( p, Vec_IntEntry(vAdds, 6*iBox+2), iRank+1, vAdds, vCarryMap, vBoxes, vBoxRanks ); }
    Vec_IntPush( vBoxes, iBox );
    Vec_IntWriteEntry( vBoxRanks, iBox, iRank );
}
func Acec_ManCollectBoxSets(p *Gia_Man_t, vAdds *Vec_Int_t, vXors *Vec_Int_t) *Vec_Wec_t {
    var vCarryMap *Vec_Int_t    = Acec_ManCreateCarryMap( p, vAdds );
    var vCarryRoots *Vec_Int_t  = Acec_ManCollectCarryRoots( p, vAdds );
    var vXorRoots *Vec_Int_t    = Acec_ManCollectXorRoots( p, vXors );
    var vBoxSets *Vec_Wec_t     = Acec_ManCollectCarryRootSets( p, vAdds, vCarryMap, vXors, vXorRoots, vCarryRoots );
    var vBoxRanks *Vec_Int_t    = Vec_IntStart( Vec_IntSize(vAdds)/6 );
    var vBoxIns *Vec_Bit_t      = Vec_BitStart( Gia_ManObjNum(p) );
    var vBoxOuts *Vec_Bit_t     = Vec_BitStart( Gia_ManObjNum(p) ); var i, k, Root int32;
    Vec_IntFree( vCarryRoots );
    Vec_IntFree( vXorRoots );
    // collect boxes for each carry set
    assert(Vec_WecSize(vBoxSets) % 5 == 0, "Vec_WecSize(vBoxSets) % 5 == 0");
    for  i = 0; 5*i < Vec_WecSize(vBoxSets); i++  {
        var vRoots *Vec_Int_t  = Vec_WecEntry( vBoxSets, 5*i+1 );
        var vBoxes *Vec_Int_t  = Vec_WecEntry( vBoxSets, 5*i+2 );
        var vIns *Vec_Int_t    = Vec_WecEntry( vBoxSets, 5*i+3 );
        var vOuts *Vec_Int_t   = Vec_WecEntry( vBoxSets, 5*i+4 );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vRoots)) && (((Root) = Vec_IntEntry(vRoots, k)), 1); k++  {
            Acec_ManCollectBoxSets_rec( p, Root, 1, vAdds, vCarryMap, vBoxes, vBoxRanks );
        }
        Acec_ManCollectInsOuts( p, vAdds, vBoxes, vBoxRanks, vBoxIns, vBoxOuts, vIns, vOuts );
    }
    Vec_IntFree( vBoxRanks );
    Vec_BitFree( vBoxIns );
    Vec_BitFree( vBoxOuts );
    Vec_IntFree( vCarryMap );
    return vBoxSets;
}
func Acec_ManPrintRanks2(vPairs *Vec_Int_t) {
    var k, iObj, Rank int32;
    // inlining of macro Vec_IntForEachEntryDouble
    for  k = 0; (k+1 < Vec_IntSize(vPairs)) && (((iObj) = Vec_IntEntry(vPairs, k)), 1) && (((Rank) = Vec_IntEntry(vPairs, k+1)), 1); k += 2  {
        fmt.Printf("%d ", Rank);
    }
    fmt.Printf("\n");
}
func Acec_ManPrintRanks(vPairs *Vec_Int_t) {
    var k, iObj, Count, Rank, RankMax int32;
    var vCounts *Vec_Int_t  = Vec_IntStart( 100 );
    // inlining of macro Vec_IntForEachEntryDouble
    for  k = 0; (k+1 < Vec_IntSize(vPairs)) && (((iObj) = Vec_IntEntry(vPairs, k)), 1) && (((Rank) = Vec_IntEntry(vPairs, k+1)), 1); k += 2  {
        Vec_IntFillExtra( vCounts, Rank+1, 0 );
        Vec_IntAddToEntry( vCounts, Rank, 1 );
        RankMax = Abc_MaxInt( RankMax, Rank );
    }
    // inlining of macro Vec_IntForEachEntryStartStop
    for  Rank = 1; (Rank < RankMax+1) && (((Count) = Vec_IntEntry(vCounts, Rank)), 1); Rank++  {
        fmt.Printf("%2d=%2d ", Rank, Count);
    }
    fmt.Printf("\n");
    Vec_IntFree( vCounts );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_ManProfile(p *Gia_Man_t, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var vBoxes *Vec_Wec_t; var i int32;
    var vXors *Vec_Int_t;
    var vAdds *Vec_Int_t = Ree_ManComputeCuts( p, &vXors, fVerbose );

    //Ree_ManPrintAdders( vAdds, 1 );
    fmt.Printf("Detected %d full-adders and %d half-adders.  Found %d XOR-cuts.  ", Ree_ManCountFadds(vAdds), Vec_IntSize(vAdds)/6-Ree_ManCountFadds(vAdds), Vec_IntSize(vXors)/4);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    clk = Abc_Clock();
    vBoxes = Acec_ManCollectBoxSets( p, vAdds, vXors );
    fmt.Printf("Detected %d adder-tree%s.  ", Vec_WecSize(vBoxes)/5, Vec_WecSize(vBoxes)/5 > 1 ? "s":"");
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    if fVerbose != 0 {
    for  i = 0; 5*i < Vec_WecSize(vBoxes); i++  {
        fmt.Printf("Tree %3d : ", i);
        fmt.Printf("Xor = %4d  ", Vec_IntSize(Vec_WecEntry(vBoxes,5*i+0)));
        fmt.Printf("Root = %4d  ", Vec_IntSize(Vec_WecEntry(vBoxes,5*i+1)));
        //printf( "(Top = %5d)  ",   Vec_IntEntryLast(Vec_WecEntry(vBoxes,5*i+1)) );
        fmt.Printf("Adder = %4d  ", Vec_IntSize(Vec_WecEntry(vBoxes,5*i+2)));
        fmt.Printf("In = %4d  ", Vec_IntSize(Vec_WecEntry(vBoxes,5*i+3))/2);
        fmt.Printf("Out = %4d  ", Vec_IntSize(Vec_WecEntry(vBoxes,5*i+4))/2);
        fmt.Printf("\n");
        fmt.Printf("           Ins:  ");
        Acec_ManPrintRanks( Vec_WecEntry(vBoxes,5*i+3) );
        fmt.Printf("           Outs: ");
        Acec_ManPrintRanks( Vec_WecEntry(vBoxes,5*i+4) );
    } }

    Vec_IntFree( vXors );
    Vec_IntFree( vAdds );
    Vec_WecFree( vBoxes );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_ManPool(p *Gia_Man_t) {
    var vTops, vTree *Vec_Int_t;
    var vTrees *Vec_Wec_t;

    var clk abctime  = Abc_Clock();
    var vAdds *Vec_Int_t  = Ree_ManComputeCuts( p, nil, 1 );
    var i int32;
    var nFadds int32 = Ree_ManCountFadds( vAdds );
    fmt.Printf("Detected %d FAs and %d HAs.  ", nFadds, Vec_IntSize(vAdds)/6-nFadds);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    clk = Abc_Clock();
    nFadds = Ree_ManCountFadds( vAdds );
    fmt.Printf("Detected %d FAs and %d HAs.  ", nFadds, Vec_IntSize(vAdds)/6-nFadds);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    //Ree_ManPrintAdders( vAdds, 1 );

    // detect topmost nodes
    vTops = Acec_ManPoolTopMost( p, vAdds );
    fmt.Printf("Detected %d topmost adder%s.\n", Vec_IntSize(vTops), Vec_IntSize(vTops) > 1 ? "s":"");

    // collect adder trees
    vTrees = Gia_PolynCoreOrderArray( p, vAdds, vTops );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vTrees)) && (((vTree) = Vec_WecEntry(vTrees, i)), 1); i++  {
        fmt.Printf("Adder %5d : Tree with %5d nodes.\n", Vec_IntEntry(vTops, i), Vec_IntSize(vTree));
    }

    Vec_WecFree( vTrees );
    Vec_IntFree( vAdds );
    Vec_IntFree( vTops );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

