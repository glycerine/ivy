package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecXor.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Detection of XOR trees.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecXor.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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
func Acec_FindXorLeaves(p *Gia_Man_t, vXors *Vec_Int_t, vAdds *Vec_Int_t, vXorRoots *Vec_Int_t, vRanks *Vec_Int_t, pvAddBoxes **Vec_Wec_t) *Vec_Wec_t {
    var vMapXors *Vec_Bit_t  = Acec_MapXorOuts2( p, vXors, vRanks );
    var vMapMajs *Vec_Int_t  = Acec_MapMajOuts2( p, vAdds, vRanks );
    var vXorLeaves *Vec_Wec_t  = Vec_WecStart( Vec_IntSize(vXorRoots) );
    var vAddBoxes *Vec_Wec_t   = Vec_WecStart( Vec_IntSize(vXorRoots) );
    var i, k int32;
    for  i = 0; 4*i < Vec_IntSize(vXors); i++  {
        var Xor int32  = Vec_IntEntry(vXors, 4*i);
        var Rank int32  = Vec_IntEntry(vRanks, Xor);
        if Rank == -1 {
            continue; }
        for  k = 1; k < 4; k++  {
            var Fanin int32  = Vec_IntEntry(vXors, 4*i+k);
            //int RankFanin = Vec_IntEntry(vRanks, Fanin);
            if Fanin == 0 {
                continue; }
            if Vec_BitEntry(vMapXors, Fanin) {
                assert(Rank == Vec_IntEntry(vRanks, Fanin), "Rank == Vec_IntEntry(vRanks, Fanin)");
                continue;
            }
//            if ( Vec_BitEntry(vMapXors, Fanin) && Rank == RankFanin )
//                continue;
            if Vec_IntEntry(vMapMajs, Fanin) == -1 { // no adder driving this input
                Vec_WecPush( vXorLeaves, Rank, Fanin );
            } else if Vec_IntEntry(vRanks, Xor) > 0 { // save adder box
                Vec_WecPush( vAddBoxes, Rank-1, Vec_IntEntry(vMapMajs, Fanin) ); }
        }
    }
    Vec_BitFree( vMapXors );
    Vec_IntFree( vMapMajs );
    if pvAddBoxes {
        *pvAddBoxes = vAddBoxes; }
    return vXorLeaves;
}
func Acec_CheckBoothPPs(p *Gia_Man_t, vLitLeaves *Vec_Wec_t) {
    var vMarked *Vec_Bit_t  = Acec_MultMarkPPs( p );
    var vLevel *Vec_Int_t;
    var i, k, iLit int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vLitLeaves)) && (((vLevel) = Vec_WecEntry(vLitLeaves, i)), 1); i++  {
        var CountPI, CountB, CountNB int32;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((iLit) = Vec_IntEntry(vLevel, k)), 1); k++  {
            if !Gia_ObjIsAnd(Gia_ManObj(p, Abc_Lit2Var(iLit))) {
                CountPI++;
            } else if Vec_BitEntry( vMarked, Abc_Lit2Var(iLit) ) {
                CountB++;
            } else {
                CountNB++; }
        }

        fmt.Printf("Rank %2d : Lits = %5d    PI = %d  Booth = %5d  Non-Booth = %5d\n", i, Vec_IntSize(vLevel), CountPI, CountB, CountNB);
    }
    Vec_BitFree( vMarked );
}
func Acec_FindBox(p *Gia_Man_t, vAdds *Vec_Int_t, vAddBoxes *Vec_Wec_t, vXorLeaves *Vec_Wec_t, vXorRoots *Vec_Int_t) *Acec_Box_t {
    extern Vec_Int_t * Acec_TreeCarryMap( Gia_Man_t * p, Vec_Int_t * vAdds, Vec_Wec_t * vBoxes );
    extern void Acec_TreePhases_rec( Gia_Man_t * p, Vec_Int_t * vAdds, Vec_Int_t * vMap, int Node, int fPhase, Vec_Bit_t * vVisit );
    extern void Acec_TreeVerifyPhases( Gia_Man_t * p, Vec_Int_t * vAdds, Vec_Wec_t * vBoxes );
    extern void Acec_TreeVerifyPhases2( Gia_Man_t * p, Vec_Int_t * vAdds, Vec_Wec_t * vBoxes );

    var MaxRank int32  = Vec_WecSize( vAddBoxes );
    var vVisit *Vec_Bit_t   = Vec_BitStart( Vec_IntSize(vAdds)/6 );
    var vIsLeaf *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var vIsRoot *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var vLevel, vLevel2, vMap *Vec_Int_t;
    var i, j, k, Box, Node int32;

    // inlining of macro ABC_CALLOC
    var pBox *Acec_Box_t  = &Acec_Box_t{};
    pBox.pGia        = p;
    pBox.vAdds       = vAddBoxes; // Vec_WecDup( vAddBoxes );
    pBox.vLeafLits   = Vec_WecStart( MaxRank + 0 );
    pBox.vRootLits   = Vec_WecStart( MaxRank + 0 );

    assert(Vec_WecSize(vAddBoxes) == Vec_WecSize(vXorLeaves), "Vec_WecSize(vAddBoxes) == Vec_WecSize(vXorLeaves)");
    assert(Vec_WecSize(vAddBoxes) == Vec_IntSize(vXorRoots), "Vec_WecSize(vAddBoxes) == Vec_IntSize(vXorRoots)");

    // collect boxes; mark inputs/outputs
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pBox.vAdds)) && (((vLevel) = Vec_WecEntry(pBox.vAdds, i)), 1); i++  {
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vLevel)) && (((Box) = Vec_IntEntry(vLevel, k)), 1); k++  {
        Vec_BitWriteEntry( vIsLeaf, Vec_IntEntry(vAdds, 6*Box+0), 1 );
        Vec_BitWriteEntry( vIsLeaf, Vec_IntEntry(vAdds, 6*Box+1), 1 );
        Vec_BitWriteEntry( vIsLeaf, Vec_IntEntry(vAdds, 6*Box+2), 1 );
        Vec_BitWriteEntry( vIsRoot, Vec_IntEntry(vAdds, 6*Box+3), 1 );
        Vec_BitWriteEntry( vIsRoot, Vec_IntEntry(vAdds, 6*Box+4), 1 );
    }
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
                    Vec_WecPush( pBox.vRootLits, k == 4 ? i + 1 : i, Abc_Var2Lit(Vec_IntEntry(vAdds, 6*Box+k), Acec_SignBit2(vAdds, Box, k)) ); }
            }
            if Vec_IntEntry(vAdds, 6*Box+2) == 0 && Acec_SignBit2(vAdds, Box, 2) {
                Vec_WecPush( pBox.vLeafLits, i, 1 ); }
        }
    }
    Vec_BitFree( vIsLeaf );
    Vec_BitFree( vIsRoot );

    // collect last bit
    vLevel  = Vec_WecEntry( pBox.vLeafLits, Vec_WecSize(pBox.vLeafLits)-1 );
    vLevel2 = Vec_WecEntry( vXorLeaves, Vec_WecSize(vXorLeaves)-1 );
    if Vec_IntSize(vLevel) == 0 && Vec_IntSize(vLevel2) > 0 {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel2)) && (((Node) = Vec_IntEntry(vLevel2, k)), 1); k++  {
            Vec_IntPush( vLevel, Abc_Var2Lit(Node, 0) );
        }
    }
    vLevel  = Vec_WecEntry( pBox.vRootLits, Vec_WecSize(pBox.vRootLits)-1 );
    Vec_IntFill( vLevel, 1, Abc_Var2Lit(Vec_IntEntryLast(vXorRoots), 0) );

    // sort each level
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pBox.vLeafLits)) && (((vLevel) = Vec_WecEntry(pBox.vLeafLits, i)), 1); i++  {
        Vec_IntSort( vLevel, 0 );
    }
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pBox.vRootLits)) && (((vLevel) = Vec_WecEntry(pBox.vRootLits, i)), 1); i++  {
        Vec_IntSort( vLevel, 1 );
    }

    //Acec_CheckBoothPPs( p, pBox->vLeafLits );
    return pBox;
}

func Acec_ProduceBox(p *Gia_Man_t, fVerbose int32) *Acec_Box_t {
    extern void Acec_TreeVerifyConnections( Gia_Man_t * p, Vec_Int_t * vAdds, Vec_Wec_t * vBoxes );

    var clk abctime  = Abc_Clock();
    var pBox *Acec_Box_t  = nil;
    var vXors *Vec_Int_t;
    var vAdds *Vec_Int_t = Ree_ManComputeCuts( p, &vXors, 0 );
    var vTemp *Vec_Int_t;
    var vXorRoots *Vec_Int_t = Acec_FindXorRoots( p, vXors ); 
    var vRanks *Vec_Int_t  = Acec_RankTrees( p, vXors, vXorRoots ); 
    var vXorLeaves *Vec_Wec_t;
    var vAddBoxes *Vec_Wec_t = nil; 

    Gia_ManLevelNum(p);

    //Acec_CheckXors( p, vXors );

    //Ree_ManPrintAdders( vAdds, 1 );
    if fVerbose != 0 {
        fmt.Printf("Detected %d full-adders and %d half-adders.  Found %d XOR-cuts.  ", Ree_ManCountFadds(vAdds), Vec_IntSize(vAdds)/6-Ree_ManCountFadds(vAdds), Vec_IntSize(vXors)/4); }
    if fVerbose != 0 {
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk ); }

    vXorRoots = Acec_OrderTreeRoots( p, vAdds, vTemp = vXorRoots, vRanks );
    Vec_IntFree( vTemp );
    Vec_IntFree( vRanks );

    vRanks = Acec_RankTrees( p, vXors, vXorRoots ); 
    vXorLeaves = Acec_FindXorLeaves( p, vXors, vAdds, vXorRoots, vRanks, &vAddBoxes ); 
    Vec_IntFree( vRanks );

    //printf( "XOR roots after reordering: \n" );
    //Vec_IntPrint( vXorRoots );
    //printf( "XOR leaves: \n" );
    //Vec_WecPrint( vXorLeaves, 0 );
    //printf( "Adder boxes: \n" );
    //Vec_WecPrint( vAddBoxes, 0 );

    Acec_TreeVerifyConnections( p, vAdds, vAddBoxes );

    pBox = Acec_FindBox( p, vAdds, vAddBoxes, vXorLeaves, vXorRoots );
    //Vec_WecFree( vAddBoxes );

    if fVerbose != 0 {
        Acec_TreePrintBox( pBox, vAdds ); }

    Vec_IntFree( vXorRoots );
    Vec_WecFree( vXorLeaves );

    Vec_IntFree( vXors );
    Vec_IntFree( vAdds );

    return pBox;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

