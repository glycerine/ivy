package hop

import "fmt"

/**CFile****************************************************************

  FileName    [hopCheck.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Minimalistic And-Inverter Graph package.]

  Synopsis    [AIG checking procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: hopCheck.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "hop.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checks the consistency of the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManCheck(p *Hop_Man_t) int32 {
    var pObj, pObj2 *Hop_Obj_t;
    var i int32;
    // check primary inputs
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        if Hop_ObjFanin0(pObj) || Hop_ObjFanin1(pObj) {
            fmt.Printf("Hop_ManCheck: The PI node \"%p\" has fanins.\n", pObj);
            return 0;
        }
    }
    // check primary outputs
    // inlining of macro Hop_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        if !Hop_ObjFanin0(pObj) {
            fmt.Printf("Hop_ManCheck: The PO node \"%p\" has NULL fanin.\n", pObj);
            return 0;
        }
        if Hop_ObjFanin1(pObj) {
            fmt.Printf("Hop_ManCheck: The PO node \"%p\" has second fanin.\n", pObj);
            return 0;
        }
    }
    // check internal nodes
    // inlining of macro Hop_ManForEachNode
    for  i = 0; i < p.nTableSize; i++  {                                               if ((pObj) = p.pTable[i]) == nil {} else {
        if !Hop_ObjFanin0(pObj) || !Hop_ObjFanin1(pObj) {
            fmt.Printf("Hop_ManCheck: The AIG has internal node \"%p\" with a NULL fanin.\n", pObj);
            return 0;
        }
        if Hop_ObjFanin0(pObj).Id >= Hop_ObjFanin1(pObj).Id {
            fmt.Printf("Hop_ManCheck: The AIG has node \"%p\" with a wrong ordering of fanins.\n", pObj);
            return 0;
        }
        pObj2 = Hop_TableLookup( p, pObj );
        if pObj2 != pObj {
            fmt.Printf("Hop_ManCheck: Node \"%p\" is not in the structural hashing table.\n", pObj);
            return 0;
        }
    }
    }
    // count the total number of nodes
    if Hop_ManObjNum(p) != 1 + Hop_ManPiNum(p) + Hop_ManPoNum(p) + Hop_ManAndNum(p) + Hop_ManExorNum(p) {
        fmt.Printf("Hop_ManCheck: The number of created nodes is wrong.\n");
        return 0;
    }
    // count the number of nodes in the table
    if Hop_TableCountEntries(p) != Hop_ManAndNum(p) + Hop_ManExorNum(p) {
        fmt.Printf("Hop_ManCheck: The number of nodes in the structural hashing table is wrong.\n");
        return 0;
    }
//    if ( !Hop_ManIsAcyclic(p) )
//        return 0;
    return 1; 
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

