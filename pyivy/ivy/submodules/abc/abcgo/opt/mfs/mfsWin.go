package mfs

/**CFile****************************************************************

  FileName    [mfsWin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The good old minimization with complete don't-cares.]

  Synopsis    [Procedures to compute windows stretching to the PIs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mfsWin.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Returns 1 if the node should be a root.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MfsComputeRootsCheck(pNode *Abc_Obj_t, nLevelMax int32, nFanoutLimit int32) int32 {
    var pFanout *Abc_Obj_t;
    var i int32;
    // the node is the root if one of the following is true:
    // (1) the node has more than fanouts than the limit
    if Abc_ObjFanoutNum(pNode) > nFanoutLimit {
        return 1; }
    // (2) the node has CO fanouts
    // (3) the node has fanouts above the cutoff level
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, i)), 1); i++  {
        if Abc_ObjIsCo(pFanout) || int32(pFanout.Level) > nLevelMax {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Recursively collects the root candidates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MfsComputeRoots_rec(pNode *Abc_Obj_t, nLevelMax int32, nFanoutLimit int32, vRoots *Vec_Ptr_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    if Abc_NodeIsTravIdCurrent(pNode) {
        return; }
    Abc_NodeSetTravIdCurrent( pNode );
    // check if the node should be the root
    if Abc_MfsComputeRootsCheck( pNode, nLevelMax, nFanoutLimit ) != 0 {
        Vec_PtrPush( vRoots, pNode );
    } else { // if not, explore its fanouts
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, i)), 1); i++  {
            Abc_MfsComputeRoots_rec( pFanout, nLevelMax, nFanoutLimit, vRoots );
        } }
}

/**Function*************************************************************

  Synopsis    [Recursively collects the root candidates.]

  Description [Returns 1 if the only root is this node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MfsComputeRoots(pNode *Abc_Obj_t, nWinTfoMax int32, nFanoutLimit int32) *Vec_Ptr_t {
    var vRoots *Vec_Ptr_t;
    vRoots = Vec_PtrAlloc( 10 );
    Abc_NtkIncrementTravId( pNode.pNtk );
    Abc_MfsComputeRoots_rec( pNode, pNode.Level + nWinTfoMax, nFanoutLimit, vRoots );
    assert(Vec_PtrSize(vRoots) > 0, "Vec_PtrSize(vRoots) > 0");
//    if ( Vec_PtrSize(vRoots) == 1 && Vec_PtrEntry(vRoots, 0) == pNode )
//        return 0;
    return vRoots;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

