package ivy

import "fmt"

/**CFile****************************************************************

  FileName    [ivyCheck.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [AIG checking procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyCheck.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

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
func Ivy_ManCheck(p *Ivy_Man_t) int32 {
    var pObj, pObj2 *Ivy_Obj_t;
    var i int32;
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        // skip deleted nodes
        if Ivy_ObjId(pObj) != i {
            fmt.Printf("Ivy_ManCheck: Node with ID %d is listed as number %d in the array of objects.\n", pObj.Id, i);
            return 0;
        }
        // consider the constant node and PIs
        if i == 0 || Ivy_ObjIsPi(pObj) {
            if Ivy_ObjFaninId0(pObj) || Ivy_ObjFaninId1(pObj) || Ivy_ObjLevel(pObj) {
                fmt.Printf("Ivy_ManCheck: The AIG has non-standard constant or PI node with ID \"%d\".\n", pObj.Id);
                return 0;
            }
            continue;
        }
        if Ivy_ObjIsPo(pObj) {
            if Ivy_ObjFaninId1(pObj) {
                fmt.Printf("Ivy_ManCheck: The AIG has non-standard PO node with ID \"%d\".\n", pObj.Id);
                return 0;
            }
            continue;
        }
        if Ivy_ObjIsBuf(pObj) {
            if Ivy_ObjFanin1(pObj) {
                fmt.Printf("Ivy_ManCheck: The buffer with ID \"%d\" contains second fanin.\n", pObj.Id);
                return 0;
            }
            continue;
        }
        if Ivy_ObjIsLatch(pObj) {
            if Ivy_ObjFanin1(pObj) {
                fmt.Printf("Ivy_ManCheck: The latch with ID \"%d\" contains second fanin.\n", pObj.Id);
                return 0;
            }
            if Ivy_ObjInit(pObj) == IVY_INIT_NONE {
                fmt.Printf("Ivy_ManCheck: The latch with ID \"%d\" does not have initial state.\n", pObj.Id);
                return 0;
            }
            pObj2 = Ivy_TableLookup( p, pObj );
            if pObj2 != pObj {
                fmt.Printf("Ivy_ManCheck: Latch with ID \"%d\" is not in the structural hashing table.\n", pObj.Id); }
            continue;
        }
        // consider the AND node
        if !Ivy_ObjFanin0(pObj) || !Ivy_ObjFanin1(pObj) {
            fmt.Printf("Ivy_ManCheck: The AIG has internal node \"%d\" with a NULL fanin.\n", pObj.Id);
            return 0;
        }
        if Ivy_ObjFaninId0(pObj) >= Ivy_ObjFaninId1(pObj) {
            fmt.Printf("Ivy_ManCheck: The AIG has node \"%d\" with a wrong ordering of fanins.\n", pObj.Id);
            return 0;
        }
        if Ivy_ObjLevel(pObj) != Ivy_ObjLevelNew(pObj) {
            fmt.Printf("Ivy_ManCheck: Node with ID \"%d\" has level %d but should have level %d.\n", pObj.Id, Ivy_ObjLevel(pObj), Ivy_ObjLevelNew(pObj)); }
        pObj2 = Ivy_TableLookup( p, pObj );
        if pObj2 != pObj {
            fmt.Printf("Ivy_ManCheck: Node with ID \"%d\" is not in the structural hashing table.\n", pObj.Id); }
        if Ivy_ObjRefs(pObj) == 0 {
            fmt.Printf("Ivy_ManCheck: Node with ID \"%d\" has no fanouts.\n", pObj.Id); }
        // check fanouts
        if p.fFanout && Ivy_ObjRefs(pObj) != Ivy_ObjFanoutNum(p, pObj) {
            fmt.Printf("Ivy_ManCheck: Node with ID \"%d\" has mismatch between the number of fanouts and refs.\n", pObj.Id); }
    }
    }
    // count the number of nodes in the table
    if Ivy_TableCountEntries(p) != Ivy_ManAndNum(p) + Ivy_ManExorNum(p) + Ivy_ManLatchNum(p) {
        fmt.Printf("Ivy_ManCheck: The number of nodes in the structural hashing table is wrong.\n");
        return 0;
    }
//    if ( !Ivy_ManCheckFanouts(p) )
//        return 0;
    if !Ivy_ManIsAcyclic(p) {
        return 0; }
    return 1; 
}

/**Function*************************************************************

  Synopsis    [Verifies the fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManCheckFanoutNums(p *Ivy_Man_t) int32 {
    var pObj *Ivy_Obj_t;
    var i, Counter int32;
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if Ivy_ObjIsNode(pObj) {
            Counter += (Ivy_ObjRefs(pObj) == 0); }
    }
    if Counter != 0 {
        fmt.Printf("Sequential AIG has %d dangling nodes.\n", Counter); }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Verifies the fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManCheckFanouts(p *Ivy_Man_t) int32 {
    var vFanouts *Vec_Ptr_t;
    var pObj, pFanout, pFanin *Ivy_Obj_t;
    var i int32;
    var k int32;
    var RetValue int32 = 1;
    if !p.fFanout {
        return 1; }
    vFanouts = Vec_PtrAlloc( 100 );
    // make sure every fanin is a fanout
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pFanin = Ivy_ObjFanin0(pObj);
        if pFanin == nil {
            continue; }
        // inlining of macro Ivy_ObjForEachFanout
        for  k = 0, Ivy_ObjCollectFanouts(p, pFanin, vFanouts);                                  k < Vec_PtrSize(vFanouts) && ((pFanout) = (Ivy_Obj_t *)Vec_PtrEntry(vFanouts,k)); k++  {
            if pFanout == pObj {
                break; }
        }
        if k == Vec_PtrSize(vFanouts) {
            fmt.Printf("Node %d is a fanin of node %d but the fanout is not there.\n", pFanin.Id, pObj.Id);
            RetValue = 0;
        }

        pFanin = Ivy_ObjFanin1(pObj);
        if pFanin == nil {
            continue; }
        // inlining of macro Ivy_ObjForEachFanout
        for  k = 0, Ivy_ObjCollectFanouts(p, pFanin, vFanouts);                                  k < Vec_PtrSize(vFanouts) && ((pFanout) = (Ivy_Obj_t *)Vec_PtrEntry(vFanouts,k)); k++  {
            if pFanout == pObj {
                break; }
        }
        if k == Vec_PtrSize(vFanouts) {
            fmt.Printf("Node %d is a fanin of node %d but the fanout is not there.\n", pFanin.Id, pObj.Id);
            RetValue = 0;
        }
        // check that the previous fanout has the same fanin
        if pObj.pPrevFan0 {
            if Ivy_ObjFanin0(pObj.pPrevFan0) != Ivy_ObjFanin0(pObj) && 
                 Ivy_ObjFanin0(pObj.pPrevFan0) != Ivy_ObjFanin1(pObj) && 
                 Ivy_ObjFanin1(pObj.pPrevFan0) != Ivy_ObjFanin0(pObj) && 
                 Ivy_ObjFanin1(pObj.pPrevFan0) != Ivy_ObjFanin1(pObj) {
                fmt.Printf("Node %d has prev %d without common fanin.\n", pObj.Id, pObj.pPrevFan0.Id);
                RetValue = 0;
            }
        }
        // check that the previous fanout has the same fanin
        if pObj.pPrevFan1 {
            if Ivy_ObjFanin0(pObj.pPrevFan1) != Ivy_ObjFanin0(pObj) && 
                 Ivy_ObjFanin0(pObj.pPrevFan1) != Ivy_ObjFanin1(pObj) && 
                 Ivy_ObjFanin1(pObj.pPrevFan1) != Ivy_ObjFanin0(pObj) && 
                 Ivy_ObjFanin1(pObj.pPrevFan1) != Ivy_ObjFanin1(pObj) {
                fmt.Printf("Node %d has prev %d without common fanin.\n", pObj.Id, pObj.pPrevFan1.Id);
                RetValue = 0;
            }
        }
    }
    }
    // make sure every fanout is a fanin
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        // inlining of macro Ivy_ObjForEachFanout
        for  k = 0, Ivy_ObjCollectFanouts(p, pObj, vFanouts);                                  k < Vec_PtrSize(vFanouts) && ((pFanout) = (Ivy_Obj_t *)Vec_PtrEntry(vFanouts,k)); k++  {
            if Ivy_ObjFanin0(pFanout) != pObj && Ivy_ObjFanin1(pFanout) != pObj {
                fmt.Printf("Node %d is a fanout of node %d but the fanin is not there.\n", pFanout.Id, pObj.Id);
                RetValue = 0;
            }
        }
    }
    }
    Vec_PtrFree( vFanouts );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Checks that each choice node has exactly one node with fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManCheckChoices(p *Ivy_Man_t) int32 {
    var pObj, pTemp *Ivy_Obj_t;
    var i int32;
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pHaig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pHaig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Ivy_ObjRefs(pObj) == 0 {
            continue; }
        // count the number of nodes in the loop
        assert(!Ivy_IsComplement(pObj.pEquiv), "!Ivy_IsComplement(pObj.pEquiv)");
        for  pTemp = pObj.pEquiv; (pTemp != nil) && pTemp != pObj; pTemp = Ivy_Regular(pTemp.pEquiv)  {
            if Ivy_ObjRefs(pTemp) > 1 {
                fmt.Printf("Node %d has member %d in its equiv class with %d fanouts.\n", pObj.Id, pTemp.Id, Ivy_ObjRefs(pTemp)); }
        }
    }
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

