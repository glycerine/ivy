package saig

import "fmt"

/**CFile****************************************************************

  FileName    [saigCone.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Cone of influence computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigCone.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Counts the support size of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManSupport_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vSupp *Vec_Ptr_t) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Aig_ObjIsConst1(pObj) {
        return; }
    if Aig_ObjIsCi(pObj) {
        if Saig_ObjIsLo(p,pObj) {
            pObj = Saig_ManLi( p, Aig_ObjCioId(pObj)-Saig_ManPiNum(p) );
            Vec_PtrPush( vSupp, pObj );
        }
        return;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Saig_ManSupport_rec( p, Aig_ObjFanin0(pObj), vSupp );
    Saig_ManSupport_rec( p, Aig_ObjFanin1(pObj), vSupp );
}

/**Function*************************************************************

  Synopsis    [Counts the support size of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManSupport(p *Aig_Man_t, vNodes *Vec_Ptr_t) *Vec_Ptr_t {
    var vSupp *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vSupp = Vec_PtrAlloc( 100 );
    Aig_ManIncrementTravId( p );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        assert(Aig_ObjIsCo(pObj), "Aig_ObjIsCo(pObj)");
        Saig_ManSupport_rec( p, Aig_ObjFanin0(pObj), vSupp );
    }
    return vSupp;
}

/**Function*************************************************************

  Synopsis    [Prints information about cones of influence of the POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManPrintConeOne(p *Aig_Man_t, pObj *Aig_Obj_t) {
    var vPrev, vCur, vTotal *Vec_Ptr_t;
    var s, i, nCurNew, nCurPrev, nCurOld int32;
    assert(Saig_ObjIsPo(p, pObj), "Saig_ObjIsPo(p, pObj)");
    // start the array
    vPrev = Vec_PtrAlloc( 100 );
    Vec_PtrPush( vPrev, pObj );
    // get the current support
    vCur = Saig_ManSupport( p, vPrev );  
    Vec_PtrClear( vPrev );
    fmt.Printf("    PO %3d  ", Aig_ObjCioId(pObj));
    // continue computing supports as long as there are now nodes
    vTotal = Vec_PtrAlloc( 100 );
    for  s = 0; ; s++  {
        // classify current into those new, prev, and older 
        nCurNew = nCurPrev = nCurOld = 0;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vCur)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCur, i)), 1); i++  {
            if Vec_PtrFind(vTotal, pObj) == -1 {
                Vec_PtrPush( vTotal, pObj );
                nCurNew++;
            
            } else if Vec_PtrFind(vPrev, pObj) >= 0 {
                nCurPrev++;
            } else {
                nCurOld++; }
        }
        assert(nCurNew + nCurPrev + nCurOld == Vec_PtrSize(vCur), "nCurNew + nCurPrev + nCurOld == Vec_PtrSize(vCur)");
        // print the result
        fmt.Printf("%d:%d %d=%d+%d+%d  ", s, Vec_PtrSize(vTotal), Vec_PtrSize(vCur), nCurNew, nCurPrev, nCurOld); 
        if nCurNew == 0 {
            break; }
        // compute one more step
        Vec_PtrFree( vPrev );
        vCur = Saig_ManSupport( p, vPrev = vCur );
    }
    fmt.Printf("\n");
    Vec_PtrFree( vPrev );
    Vec_PtrFree( vCur );
    Vec_PtrFree( vTotal );
}

/**Function*************************************************************

  Synopsis    [Prints information about cones of influence of the POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManPrintCones(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    fmt.Printf("The format of this print-out: For each PO, x:a b=c+d+e, where \n");
    fmt.Printf("- x is the time-frame counting back from the PO\n");
    fmt.Printf("- a is the total number of registers in the COI of the PO so far\n");
    fmt.Printf("- b is the number of registers in the COI of the PO in this time-frame\n");
    fmt.Printf("- c is the number of registers in b that are new (appear for the first time)\n");
    fmt.Printf("- d is the number of registers in b in common with the previous time-frame\n");
    fmt.Printf("- e is the number of registers in b in common with other time-frames\n");
    Aig_ManSetCioIds( p );
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Saig_ManPrintConeOne( p, pObj );
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

