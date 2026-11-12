package aig

import "fmt"

/**CFile****************************************************************
 
  FileName    [aigCheck.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [AIG checking procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigCheck.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/
 
//#include "aig.h"

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
func Aig_ManCheck(p *Aig_Man_t) int32 {
    var pObj, pObj2 *Aig_Obj_t;
    var i int32;
    // check primary inputs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        if Aig_ObjFanin0(pObj) || Aig_ObjFanin1(pObj) {
            fmt.Printf("Aig_ManCheck: The PI node \"%p\" has fanins.\n", pObj);
            return 0;
        }
    }
    // check primary outputs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if !Aig_ObjFanin0(pObj) {
            fmt.Printf("Aig_ManCheck: The PO node \"%p\" has NULL fanin.\n", pObj);
            return 0;
        }
        if Aig_ObjFanin1(pObj) {
            fmt.Printf("Aig_ManCheck: The PO node \"%p\" has second fanin.\n", pObj);
            return 0;
        }
    }
    // check internal nodes
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsNode(pObj) {
            continue; }
        if !Aig_ObjFanin0(pObj) || !Aig_ObjFanin1(pObj) {
            fmt.Printf("Aig_ManCheck: The AIG has internal node \"%p\" with a NULL fanin.\n", pObj);
            return 0;
        }
        if Aig_ObjFanin0(pObj).Id >= Aig_ObjFanin1(pObj).Id {
            fmt.Printf("Aig_ManCheck: The AIG has node \"%p\" with a wrong ordering of fanins.\n", pObj);
            return 0;
        }
        pObj2 = Aig_TableLookup( p, pObj );
        if pObj2 != pObj {
            fmt.Printf("Aig_ManCheck: Node \"%p\" is not in the structural hashing table.\n", pObj);
            return 0;
        }
    }
    }
    // count the total number of nodes
    if Aig_ManObjNum(p) != 1 + Aig_ManCiNum(p) + Aig_ManCoNum(p) + 
        Aig_ManBufNum(p) + Aig_ManAndNum(p) + Aig_ManExorNum(p) {
        fmt.Printf("Aig_ManCheck: The number of created nodes is wrong.\n");
        fmt.Printf("C1 = %d. Pi = %d. Po = %d. Buf = %d. And = %d. Xor = %d. Total = %d.\n", 1, Aig_ManCiNum(p), Aig_ManCoNum(p), Aig_ManBufNum(p), Aig_ManAndNum(p), Aig_ManExorNum(p), 1 + Aig_ManCiNum(p) + Aig_ManCoNum(p) + Aig_ManBufNum(p) + Aig_ManAndNum(p) + Aig_ManExorNum(p));
        fmt.Printf("Created = %d. Deleted = %d. Existing = %d.\n", Aig_ManObjNumMax(p), p.nDeleted, Aig_ManObjNum(p));
        return 0;
    }
    // count the number of nodes in the table
    if Aig_TableCountEntries(p) != Aig_ManAndNum(p) + Aig_ManExorNum(p) {
        fmt.Printf("Aig_ManCheck: The number of nodes in the structural hashing table is wrong.\n");
        fmt.Printf("Entries = %d. And = %d. Xor = %d. Total = %d.\n", Aig_TableCountEntries(p), Aig_ManAndNum(p), Aig_ManExorNum(p), Aig_ManAndNum(p) + Aig_ManExorNum(p));

        return 0;
    }
//    if ( !Aig_ManIsAcyclic(p) )
//        return 0;
    return 1; 
}

/**Function*************************************************************

  Synopsis    [Checks if the markA is reset.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCheckMarkA(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(pObj.fMarkA == 0, "pObj.fMarkA == 0"); }
    }
}

/**Function*************************************************************

  Synopsis    [Checks the consistency of phase assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCheckPhase(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if Aig_ObjIsCi(pObj) {
            assert(int32(pObj.fPhase) == 0, "int32(pObj.fPhase) == 0");
        } else {
            assert(int32(pObj.fPhase) == (Aig_ObjPhaseReal(Aig_ObjChild0(pObj)) & Aig_ObjPhaseReal(Aig_ObjChild1(pObj))), "int32(pObj.fPhase) == (Aig_ObjPhaseReal(Aig_ObjChild0(pObj)) & Aig_ObjPhaseReal(Aig_ObjChild1(pObj)))"); }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

