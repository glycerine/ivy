package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecTree.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Adder tree construction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecTree.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"

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
func Acec_TreeVerifyPhaseOne_rec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    var Truth0, Truth1 int32;
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return pObj.Value; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    assert(!Gia_ObjIsXor(pObj), "!Gia_ObjIsXor(pObj)");
    Truth0 = Acec_TreeVerifyPhaseOne_rec( p, Gia_ObjFanin0(pObj) );
    Truth1 = Acec_TreeVerifyPhaseOne_rec( p, Gia_ObjFanin1(pObj) );
    if Gia_ObjFaninC0(pObj) {
        Truth0 = 0xFF & ~Truth0;
    } else {
        Truth0 = Truth0;
    }
    if Gia_ObjFaninC1(pObj) {
        Truth1 = 0xFF & ~Truth1;
    } else {
        Truth1 = Truth1;
    }
    return (pObj.Value = Truth0 & Truth1);
}
func Acec_TreeVerifyPhaseOne(p *Gia_Man_t, vAdds *Vec_Int_t, iBox int32) {
    var pObj *Gia_Obj_t;
    var TruthXor uint32;
    var TruthMaj uint32;
    var Truths [3]uint32 = [3]uint32{ 0xAA, 0xCC, 0xF0 };
    var k int32;
    var iObj int32;
    var fFadd int32 = Vec_IntEntry(vAdds, 6*iBox+2) > 0;
    var fFlip int32  = !fFadd && Acec_SignBit2(vAdds, iBox, 2);

    Gia_ManIncrementTravId( p );
    for  k = 0; k < 3; k++  {
        iObj = Vec_IntEntry( vAdds, 6*iBox+k );
        if iObj == 0 {
            continue; }
        pObj = Gia_ManObj( p, iObj );
        if Acec_SignBit2(vAdds, iBox, k) ^ fFlip {
            pObj.Value = 0xFF & ~Truths[k];
        } else {
            pObj.Value = Truths[k];
        }
        Gia_ObjSetTravIdCurrent( p, pObj );
    }

    iObj = Vec_IntEntry( vAdds, 6*iBox+3 );
    TruthXor = Acec_TreeVerifyPhaseOne_rec( p, Gia_ManObj(p, iObj) );
    if Acec_SignBit2(vAdds, iBox, 3) ^ fFlip {
        TruthXor = 0xFF & ~TruthXor;
    } else {
        TruthXor = TruthXor;
    }

    iObj = Vec_IntEntry( vAdds, 6*iBox+4 );
    TruthMaj = Acec_TreeVerifyPhaseOne_rec( p, Gia_ManObj(p, iObj) );
    if Acec_SignBit2(vAdds, iBox, 4) ^ fFlip {
        TruthMaj = 0xFF & ~TruthMaj;
    } else {
        TruthMaj = TruthMaj;
    }

    if fFadd != 0 { // FADD
    
        if TruthXor != 0x96 {
            fmt.Printf("Fadd %d sum %d is wrong.\n", iBox, Vec_IntEntry( vAdds, 6*iBox+3 )); }
        if TruthMaj != 0xE8 {
            fmt.Printf("Fadd %d carry %d is wrong.\n", iBox, Vec_IntEntry( vAdds, 6*iBox+4 )); }
    
    } else {
        //printf( "Sign1 = %d%d%d %d\n", Acec_SignBit(vAdds, iBox, 0), Acec_SignBit(vAdds, iBox, 1), Acec_SignBit(vAdds, iBox, 2), Acec_SignBit(vAdds, iBox, 3) );
        //printf( "Sign2 = %d%d%d %d%d\n", Acec_SignBit2(vAdds, iBox, 0), Acec_SignBit2(vAdds, iBox, 1), Acec_SignBit2(vAdds, iBox, 2), Acec_SignBit2(vAdds, iBox, 3), Acec_SignBit2(vAdds, iBox, 4) );
        if TruthXor != 0x66 {
            fmt.Printf("Hadd %d sum %d is wrong.\n", iBox, Vec_IntEntry( vAdds, 6*iBox+3 )); }
        if TruthMaj != 0x88 {
            fmt.Printf("Hadd %d carry %d is wrong.\n", iBox, Vec_IntEntry( vAdds, 6*iBox+4 )); }
    }
}
func Acec_TreeVerifyPhases(p *Gia_Man_t, vAdds *Vec_Int_t, vBoxes *Vec_Wec_t) {
    var vLevel *Vec_Int_t;
    var i, k, Box int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vBoxes)) && (((vLevel) = Vec_WecEntry(vBoxes, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Box) = Vec_IntEntry(vLevel, k)), 1); k++  {
            Acec_TreeVerifyPhaseOne( p, vAdds, Box );
        }
    }
}
func Acec_TreeVerifyPhases2(p *Gia_Man_t, vAdds *Vec_Int_t, vBoxes *Vec_Wec_t) {
    var vPhase *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var vRoots *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var vLevel *Vec_Int_t;
    var i, k, n, Box int32;
    // mark all output points and their values
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vBoxes)) && (((vLevel) = Vec_WecEntry(vBoxes, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Box) = Vec_IntEntry(vLevel, k)), 1); k++  {
            Vec_BitWriteEntry( vRoots, Vec_IntEntry( vAdds, 6*Box+3 ), 1 );
            Vec_BitWriteEntry( vRoots, Vec_IntEntry( vAdds, 6*Box+4 ), 1 );
            Vec_BitWriteEntry( vPhase, Vec_IntEntry( vAdds, 6*Box+3 ), Acec_SignBit2(vAdds, Box, 3) );
            Vec_BitWriteEntry( vPhase, Vec_IntEntry( vAdds, 6*Box+4 ), Acec_SignBit2(vAdds, Box, 4) );
        }
    }
    // compare with input points
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vBoxes)) && (((vLevel) = Vec_WecEntry(vBoxes, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Box) = Vec_IntEntry(vLevel, k)), 1); k++  {
            for  n = 0; n < 3; n++  {
                if !Vec_BitEntry(vRoots, Vec_IntEntry(vAdds, 6*Box+n)) {
                    continue; }
                if Vec_BitEntry(vPhase, Vec_IntEntry(vAdds, 6*Box+n)) == Acec_SignBit2(vAdds, Box, n) {
                    continue; }
                fmt.Printf("Phase of input %d=%d is mismatched in box %d=(%d,%d).\n", n, Vec_IntEntry(vAdds, 6*Box+n), Box, Vec_IntEntry(vAdds, 6*Box+3), Vec_IntEntry(vAdds, 6*Box+4));
            }
        }
    }
    Vec_BitFree( vPhase );
    Vec_BitFree( vRoots );
}
func Acec_TreeVerifyConnections(p *Gia_Man_t, vAdds *Vec_Int_t, vBoxes *Vec_Wec_t) {
    var vCounts *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    var vLevel *Vec_Int_t;
    var i, k, n, Box int32;
    // mark outputs
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vBoxes)) && (((vLevel) = Vec_WecEntry(vBoxes, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Box) = Vec_IntEntry(vLevel, k)), 1); k++  {
            Vec_IntWriteEntry( vCounts, Vec_IntEntry( vAdds, 6*Box+3 ), 0 );
            Vec_IntWriteEntry( vCounts, Vec_IntEntry( vAdds, 6*Box+4 ), 0 );
        }
    }
    // count fanouts
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vBoxes)) && (((vLevel) = Vec_WecEntry(vBoxes, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Box) = Vec_IntEntry(vLevel, k)), 1); k++  {
            for  n = 0; n < 3; n++  {
                if Vec_IntEntry( vCounts, Vec_IntEntry(vAdds, 6*Box+n) ) != -1 {
                    Vec_IntAddToEntry( vCounts, Vec_IntEntry(vAdds, 6*Box+n), 1 ); }
            }
        }
    }
    // print out
    fmt.Printf("The adder tree has %d internal cut points. ", Vec_IntCountLarger(vCounts, -1));
    if Vec_IntCountLarger(vCounts, 1) == 0 {
        fmt.Printf("There is no internal fanouts.\n");
    } else {
        fmt.Printf("These %d points have more than one fanout:\n", Vec_IntCountLarger(vCounts, 1));
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vCounts)) && (((Box) = Vec_IntEntry(vCounts, i)), 1); i++  {
            if Box > 1 {
                fmt.Printf("Node %d(lev %d) has fanout %d.\n", i, Gia_ObjLevelId(p, i), Box); }
        }
    }
    Vec_IntFree( vCounts );
}

/**Function*************************************************************

  Synopsis    [Creates polarity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_TreeCarryMap(p *Gia_Man_t, vAdds *Vec_Int_t, vBoxes *Vec_Wec_t) *Vec_Int_t {
    var vMap *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    var vLevel *Vec_Int_t;
    var i, k, Box int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vBoxes)) && (((vLevel) = Vec_WecEntry(vBoxes, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Box) = Vec_IntEntry(vLevel, k)), 1); k++  {
            Vec_IntWriteEntry( vMap, Vec_IntEntry(vAdds, 6*Box+4), Box );
        }
    }
    return vMap;
}
func Acec_TreePhases_rec(p *Gia_Man_t, vAdds *Vec_Int_t, vMap *Vec_Int_t, Node int32, fPhase int32, vVisit *Vec_Bit_t) {
    var k, iBox, iXor, fXorPhase, fPhaseThis int32;
    assert(Node != 0, "Node != 0");
    iBox = Vec_IntEntry( vMap, Node );
    if iBox == -1 {
        return; }
    assert(Node == Vec_IntEntry( vAdds, 6*iBox+4 ), "Node == Vec_IntEntry( vAdds, 6*iBox+4 )");
    if Vec_BitEntry(vVisit, iBox) {
        return; }
    Vec_BitWriteEntry( vVisit, iBox, 1 );
    iXor = Vec_IntEntry( vAdds, 6*iBox+3 );
    fXorPhase = Acec_SignBit(vAdds, iBox, 3);
    if Vec_IntEntry(vAdds, 6*iBox+2) == 0 {
        //fPhaseThis = Acec_SignBit( vAdds, iBox, 2 ) ^ fPhase;
        //fXorPhase ^= fPhaseThis;
        //Acec_SignSetBit2( vAdds, iBox, 2, fPhaseThis ); // complemented HADD -- create const1 input
        fPhase ^= Acec_SignBit( vAdds, iBox, 2 );
        fXorPhase ^= fPhase;
        Acec_SignSetBit2( vAdds, iBox, 2, fPhase ); // complemented HADD -- create const1 input
    }
    for  k = 0; k < 3; k++  {
        var iObj int32  = Vec_IntEntry( vAdds, 6*iBox+k );
        if iObj == 0 {
            continue; }
        fPhaseThis = Acec_SignBit(vAdds, iBox, k) ^ fPhase;
        fXorPhase ^= fPhaseThis;
        Acec_TreePhases_rec( p, vAdds, vMap, iObj, fPhaseThis, vVisit );
        Acec_SignSetBit2( vAdds, iBox, k, fPhaseThis );
    }
    Acec_SignSetBit2( vAdds, iBox, 3, fXorPhase );
    Acec_SignSetBit2( vAdds, iBox, 4, fPhase );
}

/**Function*************************************************************

  Synopsis    [Find internal cut points with exactly one adder fanin/fanout.]

  Description [Returns a map of point into its input/output adder.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_PrintAdders(vBoxes *Vec_Wec_t, vAdds *Vec_Int_t) {
    var vLevel *Vec_Int_t;
    var i, k, iBox int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vBoxes)) && (((vLevel) = Vec_WecEntry(vBoxes, i)), 1); i++  {
        fmt.Printf(" %4d : %2d  {", i, Vec_IntSize(vLevel));
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((iBox) = Vec_IntEntry(vLevel, k)), 1); k++  {
            fmt.Printf(" %s%d=(%d,%d)", Vec_IntEntry(vAdds, 6*iBox+2) == 0 ? "*":"", iBox, Vec_IntEntry(vAdds, 6*iBox+3), Vec_IntEntry(vAdds, 6*iBox+4));
            //printf( "(%d,%d,%d)", Vec_IntEntry(vAdds, 6*iBox+0), Vec_IntEntry(vAdds, 6*iBox+1), Vec_IntEntry(vAdds, 6*iBox+2) );
        }
        fmt.Printf(" }\n");
    }
}
func Acec_TreePrintBox(pBox *Acec_Box_t, vAdds *Vec_Int_t) {
    fmt.Printf("Adders:\n");
    Acec_PrintAdders( pBox.vAdds, vAdds );
    fmt.Printf("Inputs:\n");
    Vec_WecPrintLits( pBox.vLeafLits );
    fmt.Printf("Outputs:\n");
    Vec_WecPrintLits( pBox.vRootLits );
//    printf( "Node %d has level %d.\n", 3715, Gia_ObjLevelId(pBox->pGia, 3715) );
//    printf( "Node %d has level %d.\n", 167, Gia_ObjLevelId(pBox->pGia, 167) );
//    printf( "Node %d has level %d.\n", 278, Gia_ObjLevelId(pBox->pGia, 278) );
//    printf( "Node %d has level %d.\n", 597, Gia_ObjLevelId(pBox->pGia, 597) );
}

func Acec_CreateBoxMaxRank(vTree *Vec_Int_t) int32 {
    var k, Box, Rank, MaxRank int32;
    // inlining of macro Vec_IntForEachEntryDouble
    for  k = 0; (k+1 < Vec_IntSize(vTree)) && (((Box) = Vec_IntEntry(vTree, k)), 1) && (((Rank) = Vec_IntEntry(vTree, k+1)), 1); k += 2  {
        MaxRank = Abc_MaxInt( MaxRank, Rank );
    }
    return MaxRank;
}
func Acec_CreateBox(p *Gia_Man_t, vAdds *Vec_Int_t, vTree *Vec_Int_t) *Acec_Box_t {
    var MaxRank int32  = Acec_CreateBoxMaxRank(vTree);
    var vVisit *Vec_Bit_t   = Vec_BitStart( Vec_IntSize(vAdds)/6 );
    var vIsLeaf *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var vIsRoot *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var vLevel, vMap *Vec_Int_t;
    var i, j, k, Box, Rank int32;//, Count = 0;

    // inlining of macro ABC_CALLOC
    var pBox *Acec_Box_t  = &Acec_Box_t{};
    pBox.pGia        = p;
    pBox.vAdds       = Vec_WecStart( MaxRank + 1 );
    pBox.vLeafLits   = Vec_WecStart( MaxRank + 1 );
    pBox.vRootLits   = Vec_WecStart( MaxRank + 2 );

    // collect boxes; mark inputs/outputs
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vTree)) && (((Box) = Vec_IntEntry(vTree, i)), 1) && (((Rank) = Vec_IntEntry(vTree, i+1)), 1); i += 2  {
//        if ( 37 == Box && 6 == Rank )
//        {
//            printf( "Skipping one adder...\n" );
//            continue;
//        }
        Vec_BitWriteEntry( vIsLeaf, Vec_IntEntry(vAdds, 6*Box+0), 1 );
        Vec_BitWriteEntry( vIsLeaf, Vec_IntEntry(vAdds, 6*Box+1), 1 );
        Vec_BitWriteEntry( vIsLeaf, Vec_IntEntry(vAdds, 6*Box+2), 1 );
        Vec_BitWriteEntry( vIsRoot, Vec_IntEntry(vAdds, 6*Box+3), 1 );
        Vec_BitWriteEntry( vIsRoot, Vec_IntEntry(vAdds, 6*Box+4), 1 );
        Vec_WecPush( pBox.vAdds, Rank, Box );
    }
    // sort each level
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pBox.vAdds)) && (((vLevel) = Vec_WecEntry(pBox.vAdds, i)), 1); i++  {
        Vec_IntSort( vLevel, 0 );
    }

    // set phases starting from roots
    vMap = Acec_TreeCarryMap( p, vAdds, pBox.vAdds );
    // inlining of macro Vec_WecForEachLevelReverse
    for  i = Vec_WecSize(pBox.vAdds)-1; (i >= 0) && (((vLevel) = Vec_WecEntry(pBox.vAdds, i)), 1); i--  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Box) = Vec_IntEntry(vLevel, k)), 1); k++  {
            if !Vec_BitEntry( vIsLeaf, Vec_IntEntry(vAdds, 6*Box+4) ) {
                //printf( "Pushing phase of output %d of box %d\n", Vec_IntEntry(vAdds, 6*Box+4), Box );
                Acec_TreePhases_rec( p, vAdds, vMap, Vec_IntEntry(vAdds, 6*Box+4), Vec_IntEntry(vAdds, 6*Box+2) != 0, vVisit );
            }
        }
    }
    Acec_TreeVerifyPhases( p, vAdds, pBox.vAdds );
    Acec_TreeVerifyPhases2( p, vAdds, pBox.vAdds );
    Vec_BitFree( vVisit );
    Vec_IntFree( vMap );

    // collect inputs/outputs
    Vec_BitWriteEntry( vIsRoot, 0, 1 );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pBox.vAdds)) && (((vLevel) = Vec_WecEntry(pBox.vAdds, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  j = 0; (j < Vec_IntSize(vLevel)) && (((Box) = Vec_IntEntry(vLevel, j)), 1); j++  {
            for  k = 0; k < 3; k++  {
                if !Vec_BitEntry( vIsRoot, Vec_IntEntry(vAdds, 6*Box+k) ) {
                    Vec_WecPush( pBox.vLeafLits, i, Abc_Var2Lit(Vec_IntEntry(vAdds, 6*Box+k), Acec_SignBit2(vAdds, Box, k)) ); }
            }
            for  k = 3; k < 5; k++  {
                if !Vec_BitEntry( vIsLeaf, Vec_IntEntry(vAdds, 6*Box+k) ) {
                    //if ( Vec_IntEntry(vAdds, 6*Box+k) == 10942 )
                    //{
                    //    printf( "++++++++++++ Skipping special\n" );
                    //    continue;
                    //}
                    Vec_WecPush( pBox.vRootLits, k == 4 ? i + 1 : i, Abc_Var2Lit(Vec_IntEntry(vAdds, 6*Box+k), Acec_SignBit2(vAdds, Box, k)) );
                }
            }
            if Vec_IntEntry(vAdds, 6*Box+2) == 0 && Acec_SignBit2(vAdds, Box, 2) {
                Vec_WecPush( pBox.vLeafLits, i, 1 ); }
        }
    }
    Vec_BitFree( vIsLeaf );
    Vec_BitFree( vIsRoot );
    // sort each level
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pBox.vLeafLits)) && (((vLevel) = Vec_WecEntry(pBox.vLeafLits, i)), 1); i++  {
        Vec_IntSort( vLevel, 0 );
    }
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pBox.vRootLits)) && (((vLevel) = Vec_WecEntry(pBox.vRootLits, i)), 1); i++  {
        Vec_IntSort( vLevel, 1 );
    }
    //return pBox;
/*
    // push literals forward
    //Vec_WecPrint( pBox->vLeafLits, 0 );
    Vec_WecForEachLevel( pBox->vLeafLits, vLevel, i )
    {
        int This, Prev = Vec_IntEntry(vLevel, 0);
        Vec_IntForEachEntryStart( vLevel, This, j, 1 )
        {
            if ( Prev != This )
            {
                Prev = This;
                continue;
            }
            if ( i+1 >= Vec_WecSize(pBox->vLeafLits) )
                continue;
            Vec_IntPushOrder( Vec_WecEntry(pBox->vLeafLits, i+1), This );
            Vec_IntDrop( vLevel, j-- );
            Vec_IntDrop( vLevel, j-- );
            Prev = -1;
            Count++;
        }
    }
    printf( "Pushed forward %d input literals.\n", Count );
*/
    //Vec_WecPrint( pBox->vLeafLits, 0 );
    return pBox;
}
func Acec_CreateBoxTest(p *Gia_Man_t) {
    var pBox *Acec_Box_t;
    var vTrees *Vec_Wec_t;
    var vTree *Vec_Int_t;

    var clk abctime  = Abc_Clock();
    var vAdds *Vec_Int_t  = Ree_ManComputeCuts( p, nil, 1 );
    var i int32;
    var nFadds int32 = Ree_ManCountFadds( vAdds );
    fmt.Printf("Detected %d adders (%d FAs and %d HAs).  ", Vec_IntSize(vAdds)/6, nFadds, Vec_IntSize(vAdds)/6-nFadds);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    clk = Abc_Clock();
    vTrees = Acec_TreeFindTrees( p, vAdds, nil, 0, 0 );
    fmt.Printf("Collected %d trees with %d adders in them.  ", Vec_WecSize(vTrees), Vec_WecSizeSize(vTrees)/2);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    //Vec_WecPrint( vTrees, 0 );

    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vTrees)) && (((vTree) = Vec_WecEntry(vTrees, i)), 1); i++  {
        pBox = Acec_CreateBox( p, vAdds, Vec_WecEntry(vTrees, i) );
        fmt.Printf("Processing tree %d:  Ranks = %d.  Adders = %d.  Leaves = %d.  Roots = %d.\n", i, Vec_WecSize(pBox.vAdds), Vec_WecSizeSize(pBox.vAdds), Vec_WecSizeSize(pBox.vLeafLits), Vec_WecSizeSize(pBox.vRootLits));
        Acec_TreePrintBox( pBox, vAdds );
        Acec_BoxFreeP( &pBox );
    }

    Vec_WecFree( vTrees );
    Vec_IntFree( vAdds );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_DeriveBox(p *Gia_Man_t, vIgnore *Vec_Bit_t, fFilterIn int32, fFilterOut int32, fVerbose int32) *Acec_Box_t {
    var pBox *Acec_Box_t  = nil;
    var vAdds *Vec_Int_t  = Ree_ManComputeCuts( p, nil, fVerbose );
    var vTrees *Vec_Wec_t  = Acec_TreeFindTrees( p, vAdds, vIgnore, fFilterIn, fFilterOut );
    if (vTrees != nil) && Vec_WecSize(vTrees) > 0 {
        pBox = Acec_CreateBox( p, vAdds, Vec_WecEntry(vTrees, 0) );
        Acec_VerifyBoxLeaves( pBox, vIgnore );
    }
    if pBox != nil {//&& fVerbose )
        fmt.Printf("Processing tree %d:  Ranks = %d.  Adders = %d.  Leaves = %d.  Roots = %d.\n", 0, Vec_WecSize(pBox.vAdds), Vec_WecSizeSize(pBox.vAdds), Vec_WecSizeSize(pBox.vLeafLits), Vec_WecSizeSize(pBox.vRootLits)); }
    if (pBox != nil) && (fVerbose != 0) {
        Acec_TreePrintBox( pBox, vAdds ); }
    //Acec_PrintAdders( pBox0->vAdds, vAdds );
    //Acec_MultDetectInputs( p, pBox->vLeafLits, pBox->vRootLits );
    Vec_WecFreeP( &vTrees );
    Vec_IntFree( vAdds );
    return pBox;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

