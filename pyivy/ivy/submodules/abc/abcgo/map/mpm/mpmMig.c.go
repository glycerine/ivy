package mpm

import "fmt"

/**CFile****************************************************************

  FileName    [mpmMig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Configurable technology mapper.]

  Synopsis    [Subject graph data structure.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 1, 2013.]

  Revision    [$Id: mpmMig.c,v 1.00 2013/06/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mpmInt.h"

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
func Mig_ManStart() *Mig_Man_t {
    
    assert(sizeof(Mig_Obj_t) >= 16, "sizeof(Mig_Obj_t) >= 16");
    assert((1 << MIG_BASE) == MIG_MASK + 1, "(1 << MIG_BASE) == MIG_MASK + 1");
    // inlining of macro ABC_CALLOC
    p := &Mig_Man_t{};
    Vec_IntGrow( &p.vCis, 1024 );
    Vec_IntGrow( &p.vCos, 1024 );
    Mig_ManAppendObj( p ); // const0
    return p;
}
func Mig_ManStop(p *Mig_Man_t) {
    if false {
    fmt.Printf("Subject graph uses %d pages of %d objects with %d entries. Total memory = %.2f MB.\n", Vec_PtrSize(&p.vPages), MIG_MASK + 1, p.nObjs, 1.0 * Vec_PtrSize(&p.vPages) * (MIG_MASK + 1) * 16 / (1 << 20)); }
    // attributes
    // inlining of macro ABC_FREE
    ((p.vTravIds.pArray) ? (free((char *) (p.vTravIds.pArray)), (p.vTravIds.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vCopies.pArray) ? (free((char *) (p.vCopies.pArray)), (p.vCopies.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vLevels.pArray) ? (free((char *) (p.vLevels.pArray)), (p.vLevels.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vRefs.pArray) ? (free((char *) (p.vRefs.pArray)), (p.vRefs.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vSibls.pArray) ? (free((char *) (p.vSibls.pArray)), (p.vSibls.pArray) = 0) : 0);
    // pages
    // inlining of macro Vec_PtrForEachEntry
    for  p.iPage = 0; (p.iPage < Vec_PtrSize(&p.vPages)) && (((p.pPage) = (Mig_Obj_t *)Vec_PtrEntry(&p.vPages, p.iPage)), 1); p.iPage++  {
        // inlining of macro ABC_FREE
        --p.pPage, ((p.pPage) ? (free((char *) (p.pPage)), (p.pPage) = 0) : 0);
    }
    // objects
    // inlining of macro ABC_FREE
    ((p.vPages.pArray) ? (free((char *) (p.vPages.pArray)), (p.vPages.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vCis.pArray) ? (free((char *) (p.vCis.pArray)), (p.vCis.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vCos.pArray) ? (free((char *) (p.vCos.pArray)), (p.vCos.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pName) ? (free((char *) (p.pName)), (p.pName) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mig_ManTypeNum(p *Mig_Man_t, Type int32) int32 {
    var pObj *Mig_Obj_t;
    var Counter int32;
    // inlining of macro Mig_ManForEachNode
    // inlining of macro Mig_ManForEachObj
    for  p.iPage = 0; p.iPage < Vec_PtrSize(&p.vPages) &&                   ((p.pPage) = (Mig_Obj_t *)Vec_PtrEntry(&p.vPages, p.iPage)); p.iPage++  {         for  pObj = p.pPage; !Mig_ObjIsNone(pObj); pObj++  { if !Mig_ObjIsNode(pObj) {} else {
        Counter += (Mig_ObjNodeType(pObj) == Type); }
    }
    }
    return Counter;
}
func Mig_ManAndNum(p *Mig_Man_t) int32 {
    return Mig_ManTypeNum(p, 1);
}
func Mig_ManXorNum(p *Mig_Man_t) int32 {
    return Mig_ManTypeNum(p, 2);
}
func Mig_ManMuxNum(p *Mig_Man_t) int32 {
    return Mig_ManTypeNum(p, 3);
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mig_ManSetRefs(p *Mig_Man_t) {
    var pObj *Mig_Obj_t;
    var i, iFanin int32;
    // increment references
    Vec_IntFill( &p.vRefs, Mig_ManObjNum(p), 0 );
    // inlining of macro Mig_ManForEachObj
    for  p.iPage = 0; p.iPage < Vec_PtrSize(&p.vPages) &&                   ((p.pPage) = (Mig_Obj_t *)Vec_PtrEntry(&p.vPages, p.iPage)); p.iPage++  {         for  pObj = p.pPage; !Mig_ObjIsNone(pObj); pObj++  {
        // inlining of macro Mig_ObjForEachFaninId
        for  i = 0; Mig_ObjHasFanin(pObj, i) && ((iFanin) = Mig_ObjFaninId(pObj, i)); i++  {
            Vec_IntAddToEntry( &p.vRefs, iFanin, 1 );
        }
        if Mig_ObjSiblId(pObj) {
            Vec_IntAddToEntry( &p.vRefs, Mig_ObjSiblId(pObj), 1 ); }
    }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mig_ManSuppSize_rec(pObj *Mig_Obj_t) int32 {
    if pObj == nil {
        return 0; }
    if Mig_ObjIsTravIdCurrent(pObj) {
        return 0; }
    Mig_ObjSetTravIdCurrent(pObj);
    if Mig_ObjIsCi(pObj) {
        return 1; }
    assert(Mig_ObjIsNode(pObj), "Mig_ObjIsNode(pObj)");
    return Mig_ManSuppSize_rec( Mig_ObjFanin0(pObj) ) +
           Mig_ManSuppSize_rec( Mig_ObjFanin1(pObj) ) +
           Mig_ManSuppSize_rec( Mig_ObjFanin2(pObj) );
}
func Mig_ManSuppSize2_rec(p *Mig_Man_t, iObj int32) int32 {
    var pObj *Mig_Obj_t;
    if iObj == MIG_NONE {
        return 0; }
    if Mig_ObjIsTravIdCurrentId(p, iObj) {
        return 0; }
    Mig_ObjSetTravIdCurrentId(p, iObj);
    pObj = Mig_ManObj( p, iObj );
    if Mig_ObjIsCi(pObj) {
        return 1; }
    assert(Mig_ObjIsNode(pObj), "Mig_ObjIsNode(pObj)");
    return Mig_ManSuppSize2_rec( p, Mig_ObjFaninId0(pObj) ) +
           Mig_ManSuppSize2_rec( p, Mig_ObjFaninId1(pObj) ) +
           Mig_ManSuppSize2_rec( p, Mig_ObjFaninId2(pObj) );
}
func Mig_ManSuppSizeOne(pObj *Mig_Obj_t) int32 {
    Mig_ObjIncrementTravId( pObj );
//    return Mig_ManSuppSize_rec( pObj );
    return Mig_ManSuppSize2_rec( Mig_ObjMan(pObj), Mig_ObjId(pObj) );
}
func Mig_ManSuppSizeTest(p *Mig_Man_t) int32 {
    var pObj *Mig_Obj_t;
    var Counter int32;
    var clk abctime  = Abc_Clock();
    // inlining of macro Mig_ManForEachObj
    for  p.iPage = 0; p.iPage < Vec_PtrSize(&p.vPages) &&                   ((p.pPage) = (Mig_Obj_t *)Vec_PtrEntry(&p.vPages, p.iPage)); p.iPage++  {         for  pObj = p.pPage; !Mig_ObjIsNone(pObj); pObj++  {
        if Mig_ObjIsNode(pObj) {
            Counter += (Mig_ManSuppSizeOne(pObj) <= 16); }
    }
    }
    fmt.Printf("Nodes with small support %d (out of %d)\n", Counter, Mig_ManNodeNum(p));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return Counter;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

