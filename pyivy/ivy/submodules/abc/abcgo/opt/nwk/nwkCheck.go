package nwk

import "fmt"

/**CFile****************************************************************

  FileName    [nwkCheck.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Logic network representation.]

  Synopsis    [Consistency checking procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkCheck.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "nwk.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checking the logic network for consistency.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManCheck(p *Nwk_Man_t) int32 {
    var pObj, pNext *Nwk_Obj_t;
    var i, k, m int32;
    // check if the nodes have duplicated fanins
    // inlining of macro Nwk_ManForEachNode
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {         if (pObj) == nil || !Nwk_ObjIsNode(pObj) {} else {
        for  k = 0; k < pObj.nFanins; k++  {
            for  m = k + 1; m < pObj.nFanins; m++  {
                if pObj.pFanio[k] == pObj.pFanio[m] {
                    fmt.Printf("Node %d has duplicated fanin %d.\n", pObj.Id, pObj.pFanio[k].Id); }
            }
        }
    }
    }
    // check if all nodes are in the correct fanin/fanout relationship
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[k]); k++  {
            if Nwk_ObjFanoutNum(pNext) < 100 && Nwk_ObjFindFanout( pNext, pObj ) == -1 {
                fmt.Printf("Nwk_ManCheck(): Object %d has fanin %d which does not have a corresponding fanout.\n", pObj.Id, pNext.Id); }
        }
        // inlining of macro Nwk_ObjForEachFanout
        for  k = 0; (k < int32((pObj).nFanouts)) && ((pNext) = (pObj).pFanio[(pObj).nFanins+k]); k++  {
            if Nwk_ObjFindFanin( pNext, pObj ) == -1 {
                fmt.Printf("Nwk_ManCheck(): Object %d has fanout %d which does not have a corresponding fanin.\n", pObj.Id, pNext.Id); }
        }
    }
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

