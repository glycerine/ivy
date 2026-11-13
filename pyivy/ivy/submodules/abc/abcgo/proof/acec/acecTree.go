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

