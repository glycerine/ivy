package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaSplit.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Structural AIG splitting.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSplit.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
/**Function*************************************************************

  Synopsis    [Takes Nodes and Marks. Returns Leaves and Roots.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Computes LUTs that are fanouts of the node outside of the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Returns the number of fanins beloning to the set.]

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
func Spl_ManFindOne(p *Spl_Man_t) int32 {
    var vVec *Vec_Int_t;
    var nFanouts, iObj, iFan, i, k int32;
    var Res int32; 

    // deref
    // inlining of macro Gia_ManForEachLut2Vec
    for  i = 0; i < Vec_IntSize(p.vNodes) && (vVec = Vec_WecEntry(p.pGia.vMapping2, (iObj = Vec_IntEntry(p.vNodes, i)))); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((iFan) = Vec_IntEntry(vVec, k)), 1); k++  {
            Gia_ObjLutRefDecId( p.pGia, iFan );
        }
    }

    // collect external nodes
    if p.fReverse && (Vec_IntSize(p.vNodes) & 1) {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vNodes)) && (((iObj) = Vec_IntEntry(p.vNodes, i)), 1); i++  {
            if Gia_ObjLutRefNumId(p.pGia, iObj) == 0 {
                continue; }
            assert(Gia_ObjLutRefNumId(p.pGia, iObj) > 0, "Gia_ObjLutRefNumId(p.pGia, iObj) > 0");
            if Gia_ObjLutRefNumId(p.pGia, iObj) >= 5 {  // skip nodes with high fanout!
                continue; }
            nFanouts = Spl_ManLutFanouts( p.pGia, iObj, p.vFanouts, p.vMarksNo, p.vMarksCIO );
            if Gia_ObjLutRefNumId(p.pGia, iObj) == 1 && nFanouts == 1 {
                Res = Vec_IntEntry(p.vFanouts, 0);
                goto finish;
            }
            //Vec_IntAppend( p->vCands, p->vFanouts );
        }
    }

    // consider LUT inputs - get one that has no refs
    Vec_IntClear( p.vCands );
    Vec_IntClear( p.vInputs );
    // inlining of macro Gia_ManForEachLut2Vec
    for  i = 0; i < Vec_IntSize(p.vNodes) && (vVec = Vec_WecEntry(p.pGia.vMapping2, (iObj = Vec_IntEntry(p.vNodes, i)))); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((iFan) = Vec_IntEntry(vVec, k)), 1); k++  {
            if !Vec_BitEntry(p.vMarksNo, iFan) && !Vec_BitEntry(p.vMarksCIO, iFan) && !Gia_ObjLutRefNumId(p.pGia, iFan) {
                Vec_IntPush( p.vCands, iFan );
                Vec_IntPush( p.vInputs, iFan );
            }
        }
    }
    Res = Spl_ManFindGoodCand( p );
    if Res != 0 {
        goto finish; }

    // collect candidates
    Vec_IntClear( p.vCands );
    Vec_IntClear( p.vInputs );
    // inlining of macro Gia_ManForEachLut2Vec
    for  i = 0; i < Vec_IntSize(p.vNodes) && (vVec = Vec_WecEntry(p.pGia.vMapping2, (iObj = Vec_IntEntry(p.vNodes, i)))); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((iFan) = Vec_IntEntry(vVec, k)), 1); k++  {
            if !Vec_BitEntry(p.vMarksNo, iFan) && !Vec_BitEntry(p.vMarksCIO, iFan) {
                Vec_IntPush( p.vCands, iFan );
                Vec_IntPush( p.vInputs, iFan );
            }
        }
    }

    // all inputs have refs - collect external nodes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vNodes)) && (((iObj) = Vec_IntEntry(p.vNodes, i)), 1); i++  {
        if Gia_ObjLutRefNumId(p.pGia, iObj) == 0 {
            continue; }
        assert(Gia_ObjLutRefNumId(p.pGia, iObj) > 0, "Gia_ObjLutRefNumId(p.pGia, iObj) > 0");
        if Gia_ObjLutRefNumId(p.pGia, iObj) >= 5 {  // skip nodes with high fanout!
            continue; }
        nFanouts = Spl_ManLutFanouts( p.pGia, iObj, p.vFanouts, p.vMarksNo, p.vMarksCIO );
        if Gia_ObjLutRefNumId(p.pGia, iObj) == 1 && nFanouts == 1 {
            Res = Vec_IntEntry(p.vFanouts, 0);
            goto finish;
        }
        Vec_IntAppend( p.vCands, p.vFanouts );
    }

    // choose among not-so-good ones
    Res = Spl_ManFindGoodCand( p );
    if Res != 0 {
        goto finish; }

    // get the first candidate
    if Res == 0 && Vec_IntSize(p.vCands) > 0 {
        Res = Vec_IntEntry( p.vCands, 0 ); }

finish:
    // ref
    // inlining of macro Gia_ManForEachLut2Vec
    for  i = 0; i < Vec_IntSize(p.vNodes) && (vVec = Vec_WecEntry(p.pGia.vMapping2, (iObj = Vec_IntEntry(p.vNodes, i)))); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((iFan) = Vec_IntEntry(vVec, k)), 1); k++  {
            Gia_ObjLutRefIncId( p.pGia, iFan );
        }
    }
    return Res;
}



/**Function*************************************************************

  Synopsis    [Computing window for one pivot node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Spl_ManLutMffcSize(p *Gia_Man_t, iObj int32, vTemp *Vec_Int_t, vMarksAnd *Vec_Bit_t) int32 {
    var iTemp, i, k int32;
    assert(Gia_ObjIsLut2(p, iObj), "Gia_ObjIsLut2(p, iObj)");
//Vec_IntPrint( Gia_ObjLutFanins2(p, iObj) );
    Gia_ManIncrementTravId( p );
    Gia_ManCollectAnds( p, &iObj, 1, vTemp, Gia_ObjLutFanins2(p, iObj) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vTemp)) && (((iTemp) = Vec_IntEntry(vTemp, i)), 1); i++  {
        if !Vec_BitEntry(vMarksAnd, iTemp) {
            Vec_IntWriteEntry( vTemp, k++, iTemp ); }
    }
    Vec_IntShrink( vTemp, k );
    return k;
}
func Spl_ManAddNode(p *Spl_Man_t, iPivot int32, vCands *Vec_Int_t) {
    var i, iObj int32;
    Vec_IntPush( p.vNodes, iPivot );
    Vec_BitWriteEntry( p.vMarksNo, iPivot, 1 );
    Vec_IntAppend( p.vAnds, vCands );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCands)) && (((iObj) = Vec_IntEntry(vCands, i)), 1); i++  {
        Vec_BitWriteEntry( p.vMarksAnd, iObj, 1 );
    }
}
func Spl_ManComputeOne(p *Spl_Man_t, iPivot int32) int32 {
    var CountAdd, iObj, i int32;
    assert(Gia_ObjIsLut2(p.pGia, iPivot), "Gia_ObjIsLut2(p.pGia, iPivot)");
//Gia_ManPrintCone2( p->pGia, Gia_ManObj(p->pGia, iPivot) );
    // assume it was previously filled in
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vNodes)) && (((iObj) = Vec_IntEntry(p.vNodes, i)), 1); i++  {
        Vec_BitWriteEntry( p.vMarksNo, iObj, 0 );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vAnds)) && (((iObj) = Vec_IntEntry(p.vAnds, i)), 1); i++  {
        Vec_BitWriteEntry( p.vMarksAnd, iObj, 0 );
    }
    // double check that it is empty
    //Gia_ManForEachLut2( p->pGia, iObj )
    //    assert( !Vec_BitEntry(p->vMarksNo, iObj) );
    //Gia_ManForEachLut2( p->pGia, iObj )
    //    assert( !Vec_BitEntry(p->vMarksAnd, iObj) );
    // clean arrays
    Vec_IntClear( p.vNodes );
    Vec_IntClear( p.vAnds );
    // add root node
    Spl_ManLutMffcSize( p.pGia, iPivot, p.vCands, p.vMarksAnd );
    Spl_ManAddNode( p, iPivot, p.vCands );
    if Vec_IntSize(p.vAnds) > p.Limit {
        return 0; }
    //printf( "%d ", iPivot );
    // add one node at a time
    for  iObj = Spl_ManFindOne(p)  {
        assert(Gia_ObjIsLut2(p.pGia, iObj), "Gia_ObjIsLut2(p.pGia, iObj)");
        assert(!Vec_BitEntry(p.vMarksNo, iObj), "!Vec_BitEntry(p.vMarksNo, iObj)");
        CountAdd = Spl_ManLutMffcSize( p.pGia, iObj, p.vCands, p.vMarksAnd );
        if Vec_IntSize(p.vAnds) + CountAdd > p.Limit {
            break; }
        Spl_ManAddNode( p, iObj, p.vCands );
        //printf( "+%d ", iObj );
    }
    //printf( "\n" );
    Vec_IntSort( p.vNodes, 0 );
    Vec_IntSort( p.vAnds, 0 );
    Spl_ManWinFindLeavesRoots( p );
    Vec_IntSort( p.vLeaves, 0 );
    Vec_IntSort( p.vRoots, 0 );
    return 1;
}

/**Function*************************************************************

  Synopsis    [External procedures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManComputeOneWin(pGia *Gia_Man_t, iPivot int32, pvRoots **Vec_Int_t, pvNodes **Vec_Int_t, pvLeaves **Vec_Int_t, pvAnds **Vec_Int_t) int32 {
    var p *Spl_Man_t  = (Spl_Man_t *)pGia.pSatlutWinman;
    assert(p != nil, "p != nil");
    if iPivot == -1 {
        Spl_ManStop( p );
        pGia.pSatlutWinman = nil;
        return 0;
    }
    if !(Spl_ManComputeOne( p, iPivot ) != 0) {
        *pvRoots  = nil;
        *pvNodes  = nil;
        *pvLeaves = nil;
        *pvAnds   = nil;
        return 0;
    }
    *pvRoots  = p.vRoots;
    *pvNodes  = p.vNodes;
    *pvLeaves = p.vLeaves;
    *pvAnds   = p.vAnds;
    // Vec_IntPrint( p->vNodes );
    return Vec_IntSize(p.vAnds);
}
func Gia_ManComputeOneWinStart(pGia *Gia_Man_t, nAnds int32, fReverse int32) {
    assert(pGia.pSatlutWinman == nil, "pGia.pSatlutWinman == nil");
    pGia.pSatlutWinman = Spl_ManAlloc( pGia, nAnds, fReverse );
}

/**Function*************************************************************

  Synopsis    [Testing procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Spl_ManComputeOneTest(pGia *Gia_Man_t) {
    var iLut, Count int32;
    Gia_ManComputeOneWinStart( pGia, 64, 0 );
    // inlining of macro Gia_ManForEachLut2
    for  iLut = 1; iLut < Gia_ManObjNum(pGia); iLut++  { if !Gia_ObjIsLut2(pGia, iLut) {} else {
        var vRoots, vNodes, vLeaves, vAnds *Vec_Int_t;
        Count = Gia_ManComputeOneWin( pGia, iLut, &vRoots, &vNodes, &vLeaves, &vAnds );
        fmt.Printf("Obj = %6d : Leaf = %2d.  Node = %2d.  Root = %2d.    AND = %3d.\n", iLut, Vec_IntSize(vLeaves), Vec_IntSize(vNodes), Vec_IntSize(vRoots), Count); 
    }
    }
    Gia_ManComputeOneWin( pGia, -1, nil, nil, nil, nil );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

