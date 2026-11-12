package mfs

/**CFile****************************************************************

  FileName    [mfsDiv.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The good old minimization with complete don't-cares.]

  Synopsis    [Procedures to compute candidate divisors.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mfsDiv.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mfsInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Marks and collects the TFI cone of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MfsWinMarkTfi_rec(pObj *Abc_Obj_t, vCone *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    if Abc_NodeIsTravIdCurrent(pObj) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    if Abc_ObjIsCi(pObj) {
        Vec_PtrPush( vCone, pObj );
        return;
    }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    // visit the fanins of the node
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_MfsWinMarkTfi_rec( pFanin, vCone );
    }
    Vec_PtrPush( vCone, pObj );
}

/**Function*************************************************************

  Synopsis    [Marks and collects the TFI cone of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MfsWinMarkTfi(pNode *Abc_Obj_t) *Vec_Ptr_t {
    var vCone *Vec_Ptr_t;
    vCone = Vec_PtrAlloc( 100 );
    Abc_MfsWinMarkTfi_rec( pNode, vCone );
    return vCone;
}

/**Function*************************************************************

  Synopsis    [Marks the TFO of the collected nodes up to the given level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MfsWinSweepLeafTfo_rec(pObj *Abc_Obj_t, nLevelLimit int32) {
    var pFanout *Abc_Obj_t;
    var i int32;
    if Abc_ObjIsCo(pObj) || int32(pObj.Level) > nLevelLimit {
        return; }
    if Abc_NodeIsTravIdCurrent(pObj) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        Abc_MfsWinSweepLeafTfo_rec( pFanout, nLevelLimit );
    }
}

/**Function*************************************************************

  Synopsis    [Dereferences the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MfsNodeDeref_rec(pNode *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    var i int32;
    var Counter int32 = 1;
    if Abc_ObjIsCi(pNode) {
        return 0; }
    Abc_NodeSetTravIdCurrent( pNode );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        assert(pFanin.vFanouts.nSize > 0, "pFanin.vFanouts.nSize > 0");
        if --pFanin.vFanouts.nSize == 0 {
            Counter += Abc_MfsNodeDeref_rec( pFanin ); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [References the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MfsNodeRef_rec(pNode *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    var i int32;
    var Counter int32 = 1;
    if Abc_ObjIsCi(pNode) {
        return 0; }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        if pFanin.vFanouts.nSize++ == 0 {
            Counter += Abc_MfsNodeRef_rec( pFanin ); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Labels MFFC of the node with the current trav ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MfsWinVisitMffc(pNode *Abc_Obj_t) int32 {
    var Count1, Count2 int32;
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    // dereference the node (mark with the current trav ID)
    Count1 = Abc_MfsNodeDeref_rec( pNode );
    // reference it back
    Count2 = Abc_MfsNodeRef_rec( pNode );
    assert(Count1 == Count2, "Count1 == Count2");
    return Count1;
}

/**Function*************************************************************

  Synopsis    [Computes divisors and add them to nodes in the window.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MfsComputeDivisors(p *Mfs_Man_t, pNode *Abc_Obj_t, nLevDivMax int32) *Vec_Ptr_t {
    var vCone, vDivs *Vec_Ptr_t;
    var pObj, pFanout, pFanin *Abc_Obj_t;
    var k, f, m int32;
    var nDivsPlus, nTrueSupp int32;
    assert(p.vDivs == nil, "p.vDivs == nil");

    // mark the TFI with the current trav ID
    Abc_NtkIncrementTravId( pNode.pNtk );
    vCone = Abc_MfsWinMarkTfi( pNode );

    // count the number of PIs
    nTrueSupp = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vCone)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vCone, k)), 1); k++  {
        nTrueSupp += Abc_ObjIsCi(pObj);
    }
//    printf( "%d(%d) ", Vec_PtrSize(p->vSupp), m );

    // mark with the current trav ID those nodes that should not be divisors:
    // (1) the node and its TFO
    // (2) the MFFC of the node
    // (3) the node's fanins (these are treated as a special case)
    Abc_NtkIncrementTravId( pNode.pNtk );
    Abc_MfsWinSweepLeafTfo_rec( pNode, nLevDivMax );
//    Abc_MfsWinVisitMffc( pNode );
    // inlining of macro Abc_ObjForEachFanin
    for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pObj) = Abc_ObjFanin(pNode, k)), 1); k++  {
        Abc_NodeSetTravIdCurrent( pObj );
    }

    // at this point the nodes are marked with two trav IDs:
    // nodes to be collected as divisors are marked with previous trav ID
    // nodes to be avoided as divisors are marked with current trav ID

    // start collecting the divisors
    vDivs = Vec_PtrAlloc( p.pPars.nWinMax );
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vCone)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vCone, k)), 1); k++  {
        if !Abc_NodeIsTravIdPrevious(pObj) {
            continue; }
        if int32(pObj.Level) > nLevDivMax {
            continue; }
        Vec_PtrPush( vDivs, pObj );
        if Vec_PtrSize(vDivs) >= p.pPars.nWinMax {
            break; }
    }
    Vec_PtrFree( vCone );

    // explore the fanouts of already collected divisors
    if Vec_PtrSize(vDivs) < p.pPars.nWinMax {
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vDivs)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vDivs, k)), 1); k++  {
        // consider fanouts of this node
        // inlining of macro Abc_ObjForEachFanout
        for  f = 0; (f < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, f)), 1); f++  {
            // stop if there are too many fanouts
            if p.pPars.nFanoutsMax && f > p.pPars.nFanoutsMax {
                break; }
            // skip nodes that are already added
            if Abc_NodeIsTravIdPrevious(pFanout) {
                continue; }
            // skip nodes in the TFO or in the MFFC of node
            if Abc_NodeIsTravIdCurrent(pFanout) {
                continue; }
            // skip COs
            if !Abc_ObjIsNode(pFanout) { 
                continue; }
            // skip nodes with large level
            if int32(pFanout.Level) > nLevDivMax {
                continue; }
            // skip nodes whose fanins are not divisors  -- here we skip more than we need to skip!!! (revise later)  August 7, 2009
            // inlining of macro Abc_ObjForEachFanin
            for  m = 0; (m < Abc_ObjFaninNum(pFanout)) && (((pFanin) = Abc_ObjFanin(pFanout, m)), 1); m++  {
                if !Abc_NodeIsTravIdPrevious(pFanin) {
                    break; }
            }
            if m < Abc_ObjFaninNum(pFanout) {
                continue; }
            // make sure this divisor in not among the nodes
//            Vec_PtrForEachEntry( Abc_Obj_t *, p->vNodes, pFanin, m )
//                assert( pFanout != pFanin );
            // add the node to the divisors
            Vec_PtrPush( vDivs, pFanout );
//            Vec_PtrPush( p->vNodes, pFanout );
            Vec_PtrPushUnique( p.vNodes, pFanout );
            Abc_NodeSetTravIdPrevious( pFanout );
            nDivsPlus++;
            if Vec_PtrSize(vDivs) >= p.pPars.nWinMax {
                break; }
        }
        if Vec_PtrSize(vDivs) >= p.pPars.nWinMax {
            break; }
    } }
    p.nMaxDivs += (Vec_PtrSize(vDivs) >= p.pPars.nWinMax);

    // sort the divisors by level in the increasing order
    Vec_PtrSort( vDivs, (int (*)(void))Abc_NodeCompareLevelsIncrease );

    // add the fanins of the node
    // inlining of macro Abc_ObjForEachFanin
    for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
        Vec_PtrPush( vDivs, pFanin );
    }

/*
    printf( "Node level = %d.  ", Abc_ObjLevel(p->pNode) );
    Vec_PtrForEachEntryStart( Abc_Obj_t *, vDivs, pObj, k, Vec_PtrSize(vDivs)-p->nDivsPlus )
        printf( "%d ", Abc_ObjLevel(pObj) );
    printf( "\n" );
*/
//printf( "%d ", p->nDivsPlus );
//    printf( "(%d+%d)(%d+%d+%d) ", Vec_PtrSize(p->vSupp), Vec_PtrSize(p->vNodes), 
//        nTrueSupp, Vec_PtrSize(vDivs)-nTrueSupp-nDivsPlus, nDivsPlus );
    return vDivs;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

