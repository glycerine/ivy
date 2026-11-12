package ivy

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [ivyCutTrav.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyCutTrav.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static unsigned * Ivy_NodeCutElementary( Vec_Int_t * vStore, int nWords, int NodeId );

// static void Ivy_NodeComputeVolume( Ivy_Obj_t * pObj, int nNodeLimit, Vec_Ptr_t * vNodes, Vec_Ptr_t * vFront );

// static void Ivy_NodeFindCutsMerge( Vec_Ptr_t * vCuts0, Vec_Ptr_t * vCuts1, Vec_Ptr_t * vCuts, int nLeaves, int nWords, Vec_Int_t * vStore );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes cuts for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ivy_NodeFindCutsTravAll_CutStore Ivy_Store_t;
var spStatic_Ivy_NodeFindCutsTravAll_pCutStore *Ivy_Store_t = &spStatic_Ivy_NodeFindCutsTravAll_CutStore;
func Ivy_NodeFindCutsTravAll(p *Ivy_Man_t, pObj *Ivy_Obj_t, nLeaves int32, nNodeLimit int32, vNodes *Vec_Ptr_t, vFront *Vec_Ptr_t, vStore *Vec_Int_t, vBitCuts *Vec_Vec_t) *Ivy_Store_t {
    
    var vCuts, vCuts0, vCuts1 *Vec_Ptr_t;
    var pBitCut *uint32;
    var pLeaf *Ivy_Obj_t;
    var pCut *Ivy_Cut_t;
    var i, k, nWords, nNodes int32;

    assert(nLeaves <= IVY_CUT_INPUT, "nLeaves <= IVY_CUT_INPUT");

    // find the given number of nodes in the TFI
    Ivy_NodeComputeVolume( pObj, nNodeLimit - 1, vNodes, vFront );
    nNodes = Vec_PtrSize(vNodes);
//    assert( nNodes <= nNodeLimit );

    // make sure vBitCuts has enough room
    Vec_VecExpand( vBitCuts, nNodes-1 );
    Vec_VecClear( vBitCuts );

    // prepare the memory manager
    Vec_IntClear( vStore );
    Vec_IntGrow( vStore, 64000 );

    // set elementary cuts for the leaves
    nWords = Extra_BitWordNum( nNodes );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pLeaf) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        assert(Ivy_ObjTravId(pLeaf) < nNodes, "Ivy_ObjTravId(pLeaf) < nNodes");
        // get the new bitcut
        pBitCut = Ivy_NodeCutElementary( vStore, nWords, Ivy_ObjTravId(pLeaf) );
        // set it as the cut of this leaf
        Vec_VecPush( vBitCuts, Ivy_ObjTravId(pLeaf), pBitCut );
    }

    // compute the cuts for each node
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pLeaf) = (Ivy_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // skip the leaves
        vCuts = Vec_VecEntry( vBitCuts, Ivy_ObjTravId(pLeaf) );
        if Vec_PtrSize(vCuts) > 0 {
            continue; }
        // add elementary cut
        pBitCut = Ivy_NodeCutElementary( vStore, nWords, Ivy_ObjTravId(pLeaf) );
        // set it as the cut of this leaf
        Vec_VecPush( vBitCuts, Ivy_ObjTravId(pLeaf), pBitCut );
        // get the fanin cuts
        vCuts0 = Vec_VecEntry( vBitCuts, Ivy_ObjTravId( Ivy_ObjFanin0(pLeaf) ) );
        vCuts1 = Vec_VecEntry( vBitCuts, Ivy_ObjTravId( Ivy_ObjFanin1(pLeaf) ) );
        assert(Vec_PtrSize(vCuts0) > 0, "Vec_PtrSize(vCuts0) > 0");
        assert(Vec_PtrSize(vCuts1) > 0, "Vec_PtrSize(vCuts1) > 0");
        // merge the cuts
        Ivy_NodeFindCutsMerge( vCuts0, vCuts1, vCuts, nLeaves, nWords, vStore );
    }

    // start the structure
    spStatic_Ivy_NodeFindCutsTravAll_pCutStore.nCuts = 0;
    spStatic_Ivy_NodeFindCutsTravAll_pCutStore.nCutsMax = IVY_CUT_LIMIT;
    // collect the cuts of the root node
    vCuts = Vec_VecEntry( vBitCuts, Ivy_ObjTravId(pObj) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCuts)) && (((pBitCut) = (unsigned *)Vec_PtrEntry(vCuts, i)), 1); i++  {
        pCut = spStatic_Ivy_NodeFindCutsTravAll_pCutStore.pCuts + spStatic_Ivy_NodeFindCutsTravAll_pCutStore.nCuts++;
        pCut.nSize = 0;
        pCut.nSizeMax = nLeaves;
        pCut.uHash = 0;
        for  k = 0; k < nNodes; k++  {
            if Extra_TruthHasBit(pBitCut, k) {
                pCut.pArray[ pCut.nSize++ ] = Ivy_ObjId( (Ivy_Obj_t *)Vec_PtrEntry(vNodes, k) ); }
        }
        assert(pCut.nSize <= nLeaves, "pCut.nSize <= nLeaves");
        if spStatic_Ivy_NodeFindCutsTravAll_pCutStore.nCuts == spStatic_Ivy_NodeFindCutsTravAll_pCutStore.nCutsMax {
            break; }
    }

    // clean the travIds
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pLeaf) = (Ivy_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pLeaf.TravId = 0;
    }
    return spStatic_Ivy_NodeFindCutsTravAll_pCutStore;
}

/**Function*************************************************************

  Synopsis    [Creates elementary bit-cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCutElementary(vStore *Vec_Int_t, nWords int32, NodeId int32) *uint32 {
    var pBitCut *uint32;
    pBitCut = Vec_IntFetch( vStore, nWords );
    memset( pBitCut, 0, (size_t)(4 * nWords) );
    Extra_TruthSetBit( pBitCut, NodeId );
    return pBitCut;
}

/**Function*************************************************************

  Synopsis    [Compares the node by level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_CompareNodesByLevel(ppObj1 **Ivy_Obj_t, ppObj2 **Ivy_Obj_t) int32 {
    var pObj1 *Ivy_Obj_t  = *ppObj1;
    var pObj2 *Ivy_Obj_t  = *ppObj2;
    if pObj1.Level < pObj2.Level {
        return -1; }
    if pObj1.Level > pObj2.Level {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Mark all nodes up to the given depth.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeComputeVolumeTrav1_rec(pObj *Ivy_Obj_t, Depth int32) {
    if Ivy_ObjIsCi(pObj) || Depth == 0 {
        return; }
    Ivy_NodeComputeVolumeTrav1_rec( Ivy_ObjFanin0(pObj), Depth - 1 );
    Ivy_NodeComputeVolumeTrav1_rec( Ivy_ObjFanin1(pObj), Depth - 1 );
    pObj.fMarkA = 1;
}

/**Function*************************************************************

  Synopsis    [Collect the marked nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeComputeVolumeTrav2_rec(pObj *Ivy_Obj_t, vNodes *Vec_Ptr_t) {
    if !pObj.fMarkA {
        return; }
    Ivy_NodeComputeVolumeTrav2_rec( Ivy_ObjFanin0(pObj), vNodes );
    Ivy_NodeComputeVolumeTrav2_rec( Ivy_ObjFanin1(pObj), vNodes );
    Vec_PtrPush( vNodes, pObj );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeComputeVolume(pObj *Ivy_Obj_t, nNodeLimit int32, vNodes *Vec_Ptr_t, vFront *Vec_Ptr_t) {
    var pTemp, pFanin *Ivy_Obj_t;
    var i, nNodes int32;
    // mark nodes up to the given depth
    Ivy_NodeComputeVolumeTrav1_rec( pObj, 6 );
    // collect the marked nodes
    Vec_PtrClear( vFront );
    Ivy_NodeComputeVolumeTrav2_rec( pObj, vFront );
    // find the fanins that are not marked
    Vec_PtrClear( vNodes );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pTemp) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        pFanin = Ivy_ObjFanin0(pTemp);
        if !pFanin.fMarkA {
            pFanin.fMarkA = 1;
            Vec_PtrPush( vNodes, pFanin );
        }
        pFanin = Ivy_ObjFanin1(pTemp);
        if !pFanin.fMarkA {
            pFanin.fMarkA = 1;
            Vec_PtrPush( vNodes, pFanin );
        }
    }
    // remember the number of nodes in the frontier
    nNodes = Vec_PtrSize( vNodes );
    // add the remaining nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pTemp) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        Vec_PtrPush( vNodes, pTemp );
    }
    // unmark the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pTemp) = (Ivy_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pTemp.fMarkA = 0;
        pTemp.TravId = i;
    }
    // collect the frontier nodes
    Vec_PtrClear( vFront );
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < nNodes) && (((pTemp) = (Ivy_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Vec_PtrPush( vFront, pTemp );
    }
//    printf( "%d ", Vec_PtrSize(vNodes) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeComputeVolume2(pObj *Ivy_Obj_t, nNodeLimit int32, vNodes *Vec_Ptr_t, vFront *Vec_Ptr_t) {
    var pLeaf, pPivot, pFanin *Ivy_Obj_t;
    var LevelMax, i int32;
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    // clear arrays
    Vec_PtrClear( vNodes );
    Vec_PtrClear( vFront );
    // add the root
    pObj.fMarkA = 1;
    Vec_PtrPush( vNodes, pObj );
    Vec_PtrPush( vFront, pObj );
    // expand node with maximum level
    LevelMax = pObj.Level;
    do {
        // get the node to expand
        pPivot = nil;
        // inlining of macro Vec_PtrForEachEntryReverse
        for  i = Vec_PtrSize(vFront) - 1; (i >= 0) && (((pLeaf) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i--  {
            if int32(pLeaf.Level) == LevelMax {
                pPivot = pLeaf;
                break;
            }
        }
        // decrease level if we did not find the node
        if pPivot == nil {
            if LevelMax--; LevelMax == 0 {
                break; }
            continue;
        }
        // the node to expand is found
        // remove it from frontier
        Vec_PtrRemove( vFront, pPivot );
        // add fanins
        pFanin = Ivy_ObjFanin0(pPivot); 
        if !pFanin.fMarkA {
            pFanin.fMarkA = 1;
            Vec_PtrPush( vNodes, pFanin );
            Vec_PtrPush( vFront, pFanin );
        }
        pFanin = Ivy_ObjFanin1(pPivot); 
        if (pFanin != nil) && !pFanin.fMarkA {
            pFanin.fMarkA = 1;
            Vec_PtrPush( vNodes, pFanin );
            Vec_PtrPush( vFront, pFanin );
        }
        // quit if we collected enough nodes
    } while ( Vec_PtrSize(vNodes) < nNodeLimit );

    // sort nodes by level
    Vec_PtrSort( vNodes, (int (*)(void))Ivy_CompareNodesByLevel );
    // make sure the nodes are ordered in the increasing number of levels
    pFanin = (Ivy_Obj_t *)Vec_PtrEntry( vNodes, 0 );
    pPivot = (Ivy_Obj_t *)Vec_PtrEntryLast( vNodes );
    assert(pFanin.Level <= pPivot.Level, "pFanin.Level <= pPivot.Level");

    // clean the marks and remember node numbers in the TravId
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pFanin.fMarkA = 0;
        pFanin.TravId = i;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_TruthOrWords(pOut *uint32, pIn0 *uint32, pIn1 *uint32, nWords int32) {
    var w int32;
    for  w = nWords-1; w >= 0; w--  {
        pOut[w] = pIn0[w] | pIn1[w];
    }
}
func Extra_TruthIsImplyWords(pIn1 *uint32, pIn2 *uint32, nWords int32) int32 {
    var w int32;
    for  w = nWords-1; w >= 0; w--  {
        if pIn1[w] & ~pIn2[w] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Merges two sets of bit-cuts at a node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeFindCutsMerge(vCuts0 *Vec_Ptr_t, vCuts1 *Vec_Ptr_t, vCuts *Vec_Ptr_t, nLeaves int32, nWords int32, vStore *Vec_Int_t) {
    var pBitCut, pBitCut0, pBitCut1, pBitCutTest *uint32;
    var i, k, c, w, Counter int32;
    // iterate through the cut pairs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCuts0)) && (((pBitCut0) = (unsigned *)Vec_PtrEntry(vCuts0, i)), 1); i++  {
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vCuts1)) && (((pBitCut1) = (unsigned *)Vec_PtrEntry(vCuts1, k)), 1); k++  {
        // skip infeasible cuts
        Counter = 0;
        for  w = 0; w < nWords; w++  {
            Counter += Extra_WordCountOnes( pBitCut0[w] | pBitCut1[w] );
            if Counter > nLeaves {
                break; }
        }
        if Counter > nLeaves {
            continue; }
        // the new cut is feasible - create it
        pBitCutTest = Vec_IntFetch( vStore, nWords );
        Extra_TruthOrWords( pBitCutTest, pBitCut0, pBitCut1, nWords );
        // filter contained cuts; try to find containing cut
        w = 0;
        // inlining of macro Vec_PtrForEachEntry
        for  c = 0; (c < Vec_PtrSize(vCuts)) && (((pBitCut) = (unsigned *)Vec_PtrEntry(vCuts, c)), 1); c++  {
            if Extra_TruthIsImplyWords( pBitCut, pBitCutTest, nWords ) != 0 {
                break; }
            if Extra_TruthIsImplyWords( pBitCutTest, pBitCut, nWords ) != 0 {
                continue; }
            Vec_PtrWriteEntry( vCuts, w++, pBitCut );
        }
        if c != Vec_PtrSize(vCuts) {
            continue; }
        Vec_PtrShrink( vCuts, w );
        // add the cut
        Vec_PtrPush( vCuts, pBitCutTest );
    }
    }
}
 
/**Function*************************************************************

  Synopsis    [Compute the set of all cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManTestCutsTravAll(p *Ivy_Man_t) {
    var pStore *Ivy_Store_t;
    var pObj *Ivy_Obj_t;
    var vNodes, vFront *Vec_Ptr_t;
    var vStore *Vec_Int_t;
    var vBitCuts *Vec_Vec_t;
    var i, nCutsCut, nCutsTotal, nNodeTotal, nNodeOver int32;
    var clk abctime  = Abc_Clock();

    vNodes = Vec_PtrAlloc( 100 );
    vFront = Vec_PtrAlloc( 100 );
    vStore = Vec_IntAlloc( 100 );
    vBitCuts = Vec_VecAlloc( 100 );

    nNodeTotal = nNodeOver = 0;
    nCutsTotal = -Ivy_ManNodeNum(p);
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Ivy_ObjIsNode(pObj) {
            continue; }
        pStore = Ivy_NodeFindCutsTravAll( p, pObj, 4, 60, vNodes, vFront, vStore, vBitCuts );
        nCutsCut    = pStore.nCuts;
        nCutsTotal += nCutsCut;
        nNodeOver  += (nCutsCut == IVY_CUT_LIMIT);
        nNodeTotal++;
    }
    }
    fmt.Printf("Total cuts = %6d. Trivial = %6d.   Nodes = %6d. Satur = %6d.  ", nCutsTotal, Ivy_ManPiNum(p) + Ivy_ManNodeNum(p), nNodeTotal, nNodeOver);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;

    Vec_PtrFree( vNodes );
    Vec_PtrFree( vFront );
    Vec_IntFree( vStore );
    Vec_VecFree( vBitCuts );

}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

