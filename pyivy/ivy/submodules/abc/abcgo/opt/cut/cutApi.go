package cut

/**CFile****************************************************************

  FileName    [cutNode.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [K-feasible cut computation package.]

  Synopsis    [Procedures to compute cuts for a node.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cutNode.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cutInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the pointer to the linked list of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeReadCutsNew(p *Cut_Man_t, Node int32) *Cut_Cut_t {
    if Node >= p.vCutsNew.nSize {
        return nil; }
    return (Cut_Cut_t *)Vec_PtrEntry( p.vCutsNew, Node );
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the linked list of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeReadCutsOld(p *Cut_Man_t, Node int32) *Cut_Cut_t {
    assert(Node < p.vCutsOld.nSize, "Node < p.vCutsOld.nSize");
    return (Cut_Cut_t *)Vec_PtrEntry( p.vCutsOld, Node );
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the linked list of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeReadCutsTemp(p *Cut_Man_t, Node int32) *Cut_Cut_t {
    assert(Node < p.vCutsTemp.nSize, "Node < p.vCutsTemp.nSize");
    return (Cut_Cut_t *)Vec_PtrEntry( p.vCutsTemp, Node );
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the linked list of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeWriteCutsNew(p *Cut_Man_t, Node int32, pList *Cut_Cut_t) {
    Vec_PtrWriteEntry( p.vCutsNew, Node, pList );
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the linked list of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeWriteCutsOld(p *Cut_Man_t, Node int32, pList *Cut_Cut_t) {
    Vec_PtrWriteEntry( p.vCutsOld, Node, pList );
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the linked list of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeWriteCutsTemp(p *Cut_Man_t, Node int32, pList *Cut_Cut_t) {
    Vec_PtrWriteEntry( p.vCutsTemp, Node, pList );
}

/**Function*************************************************************

  Synopsis    [Sets the trivial cut for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeSetTriv(p *Cut_Man_t, Node int32) {
    assert(Cut_NodeReadCutsNew(p, Node) == nil, "Cut_NodeReadCutsNew(p, Node) == nil");
    Cut_NodeWriteCutsNew( p, Node, Cut_CutCreateTriv(p, Node) );
}

/**Function*************************************************************

  Synopsis    [Consider dropping cuts if they are useless by now.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeTryDroppingCuts(p *Cut_Man_t, Node int32) {
    var nFanouts int32;
    assert(p.vFanCounts, "p.vFanCounts");
    nFanouts = Vec_IntEntry( p.vFanCounts, Node );
    assert(nFanouts > 0, "nFanouts > 0");
    if nFanouts--; nFanouts == 0 {
        Cut_NodeFreeCuts( p, Node ); }
    Vec_IntWriteEntry( p.vFanCounts, Node, nFanouts );
}

/**Function*************************************************************

  Synopsis    [Deallocates the cuts at the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeFreeCuts(p *Cut_Man_t, Node int32) {
    var pList, pCut, pCut2 *Cut_Cut_t;
    pList = Cut_NodeReadCutsNew( p, Node );
    if pList == nil {
        return; }
    // inlining of macro Cut_ListForEachCutSafe
    for  pCut = pList,                                             pCut2 = pCut? pCut.pNext: nil;                          pCut;                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
        Cut_CutRecycle( p, pCut );
    }
    Cut_NodeWriteCutsNew( p, Node, nil );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

