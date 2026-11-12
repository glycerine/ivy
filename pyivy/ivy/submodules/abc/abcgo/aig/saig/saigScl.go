package saig

import "fmt"

/**CFile****************************************************************

  FileName    [saigScl.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Sequential cleanup.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigScl.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Report registers useless for SEC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManReportUselessRegisters(pAig *Aig_Man_t) {
    var pObj, pDriver *Aig_Obj_t;
    var i, Counter1, Counter2 int32;
    // set PIO numbers
    Aig_ManSetCioIds( pAig );
    // check how many POs are driven by a register whose ref count is 1
    Counter1 = 0;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        pDriver = Aig_ObjFanin0(pObj);
        if Saig_ObjIsLo(pAig, pDriver) && Aig_ObjRefs(pDriver) == 1 {
            Counter1++; }
    }
    // check how many PIs have ref count 1 and drive a register
    Counter2 = 0;
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        pDriver = Aig_ObjFanin0(pObj);
        if Saig_ObjIsPi(pAig, pDriver) && Aig_ObjRefs(pDriver) == 1 {
            Counter2++; }
    }
    if Counter1 {
        fmt.Printf("Network has %d (out of %d) registers driving POs.\n", Counter1, Saig_ManRegNum(pAig)); }
    if Counter2 != 0 {
        fmt.Printf("Network has %d (out of %d) registers driven by PIs.\n", Counter2, Saig_ManRegNum(pAig)); }
}


/**Function*************************************************************

  Synopsis    [Report the number of pairs of complemented registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManReportComplements(p *Aig_Man_t) int32 {
    var pObj, pFanin *Aig_Obj_t;
    var i, Counter int32;
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(!pObj.fMarkA, "!pObj.fMarkA"); }
    }
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pFanin = Aig_ObjFanin0(pObj);
        if pFanin.fMarkA {
            Counter++;
        } else {
            pFanin.fMarkA = 1; }
    }
    // count fanins that have both attributes
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pFanin = Aig_ObjFanin0(pObj);
        pFanin.fMarkA = 0;
    }
    return Counter;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

