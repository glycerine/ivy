package tim

import "fmt"

/**CFile****************************************************************

  FileName    [timTime.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchy/timing manager.]

  Synopsis    [Setting and resetting timing information of the boxes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: timTime.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "timInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Initializes arrival time of the PI.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManInitPiArrival(p *Tim_Man_t, iPi int32, Delay float32) {
    assert(iPi < p.nCis, "iPi < p.nCis");
    p.pCis[iPi].timeArr = Delay;
}

/**Function*************************************************************

  Synopsis    [Initializes required time of the PO.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManInitPoRequired(p *Tim_Man_t, iPo int32, Delay float32) {
    assert(iPo < p.nCos, "iPo < p.nCos");
    p.pCos[iPo].timeReq = Delay;
}

/**Function*************************************************************

  Synopsis    [Sets arrival times of all PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManInitPiArrivalAll(p *Tim_Man_t, Delay float32) {
    var pObj *Tim_Obj_t;
    var i int32;
    // inlining of macro Tim_ManForEachPi
    // inlining of macro Tim_ManForEachCi
    for  i = 0; (i < (p).nCis) && ((pObj) = (p).pCis + i); i++  { if pObj.iObj2Box >= 0 {} else { 
        Tim_ManInitPiArrival( p, i, Delay ); }
    }
}

/**Function*************************************************************

  Synopsis    [Sets required times of all POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManInitPoRequiredAll(p *Tim_Man_t, Delay float32) {
    var pObj *Tim_Obj_t;
    var i int32;
    // inlining of macro Tim_ManForEachPo
    // inlining of macro Tim_ManForEachCo
    for  i = 0; (i < (p).nCos) && ((pObj) = (p).pCos + i); i++  { if pObj.iObj2Box >= 0 {} else { 
        Tim_ManSetCoRequired( p, i, Delay ); }
    }
}

/**Function*************************************************************

  Synopsis    [Updates arrival time of the CO.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManSetCoArrival(p *Tim_Man_t, iCo int32, Delay float32) {
    assert(iCo < p.nCos, "iCo < p.nCos");
    assert(!p.fUseTravId || p.pCos[iCo].TravId != p.nTravIds, "!p.fUseTravId || p.pCos[iCo].TravId != p.nTravIds");
    p.pCos[iCo].timeArr = Delay;
    p.pCos[iCo].TravId = p.nTravIds;
}

/**Function*************************************************************

  Synopsis    [Updates required time of the CI.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManSetCiRequired(p *Tim_Man_t, iCi int32, Delay float32) {
    assert(iCi < p.nCis, "iCi < p.nCis");
    assert(!p.fUseTravId || p.pCis[iCi].TravId != p.nTravIds, "!p.fUseTravId || p.pCis[iCi].TravId != p.nTravIds");
    p.pCis[iCi].timeReq = Delay;
    p.pCis[iCi].TravId = p.nTravIds;
}

/**Function*************************************************************

  Synopsis    [Updates required time of the CO.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManSetCoRequired(p *Tim_Man_t, iCo int32, Delay float32) {
    assert(iCo < p.nCos, "iCo < p.nCos");
    assert(!p.fUseTravId || !p.nTravIds || p.pCos[iCo].TravId != p.nTravIds, "!p.fUseTravId || !p.nTravIds || p.pCos[iCo].TravId != p.nTravIds");
    p.pCos[iCo].timeReq = Delay;
    p.pCos[iCo].TravId = p.nTravIds;
}


/**Function*************************************************************

  Synopsis    [Returns CO arrival time.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManGetCiArrival(p *Tim_Man_t, iCi int32) float32 {
    var pBox *Tim_Box_t;
    var pObjThis, pObj, pObjRes *Tim_Obj_t;
    var pTable *float32;
    var pDelays *float32;
    var DelayBest float32;
    var i, k int32;
    // consider the already processed PI
    pObjThis = Tim_ManCi( p, iCi );
    if p.fUseTravId && pObjThis.TravId == p.nTravIds {
        return pObjThis.timeArr; }
    pObjThis.TravId = p.nTravIds;
    // consider the main PI
    pBox = Tim_ManCiBox( p, iCi );
    if pBox == nil {
        return pObjThis.timeArr; }
    // update box timing
    pBox.TravId = p.nTravIds;
    // get the arrival times of the inputs of the box (POs)
    if p.fUseTravId {
    // inlining of macro Tim_ManBoxForEachInput
    for  i = 0; (i < (pBox).nInputs) && ((pObj) = Tim_ManBoxInput(p, pBox, i)); i++  {
        if pObj.TravId != p.nTravIds {
            fmt.Printf("Tim_ManGetCiArrival(): Input arrival times of the box are not up to date!\n"); }
    } }
    // compute the arrival times for each output of the box (PIs)
    pTable = Tim_ManBoxDelayTable( p, pBox.iBox );
    // inlining of macro Tim_ManBoxForEachOutput
    for  i = 0; (i < (pBox).nOutputs) && ((pObjRes) = Tim_ManBoxOutput(p, pBox, i)); i++  {
        pDelays = pTable + 3 + i * pBox.nInputs;
        DelayBest = -TIM_ETERNITY;
        // inlining of macro Tim_ManBoxForEachInput
        for  k = 0; (k < (pBox).nInputs) && ((pObj) = Tim_ManBoxInput(p, pBox, k)); k++  {
            if pDelays[k] != -ABC_INFINITY {
                DelayBest = Abc_MaxInt( DelayBest, pObj.timeArr + pDelays[k] ); }
        }
        pObjRes.timeArr = DelayBest;
        pObjRes.TravId = p.nTravIds;
    }
    return pObjThis.timeArr;
}

/**Function*************************************************************

  Synopsis    [Returns CO required time.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManGetCoRequired(p *Tim_Man_t, iCo int32) float32 {
    var pBox *Tim_Box_t;
    var pObjThis, pObj, pObjRes *Tim_Obj_t;
    var pTable *float32;
    var pDelays *float32;
    var DelayBest float32;
    var i, k int32;
    // consider the already processed PO
    pObjThis = Tim_ManCo( p, iCo );
    if p.fUseTravId && pObjThis.TravId == p.nTravIds {
        return pObjThis.timeReq; }
    pObjThis.TravId = p.nTravIds;
    // consider the main PO
    pBox = Tim_ManCoBox( p, iCo );
    if pBox == nil {
        return pObjThis.timeReq; }
    // update box timing
    pBox.TravId = p.nTravIds;
    // get the required times of the outputs of the box (PIs)
    if p.fUseTravId {
    // inlining of macro Tim_ManBoxForEachOutput
    for  i = 0; (i < (pBox).nOutputs) && ((pObj) = Tim_ManBoxOutput(p, pBox, i)); i++  {
        if pObj.TravId != p.nTravIds {
            fmt.Printf("Tim_ManGetCoRequired(): Output required times of output %d the box %d are not up to date!\n", i, pBox.iBox); }
    } }
    // compute the required times for each input of the box (POs)
    pTable = Tim_ManBoxDelayTable( p, pBox.iBox );
    // inlining of macro Tim_ManBoxForEachInput
    for  i = 0; (i < (pBox).nInputs) && ((pObjRes) = Tim_ManBoxInput(p, pBox, i)); i++  {
        DelayBest = TIM_ETERNITY;
        // inlining of macro Tim_ManBoxForEachOutput
        for  k = 0; (k < (pBox).nOutputs) && ((pObj) = Tim_ManBoxOutput(p, pBox, k)); k++  {
            pDelays = pTable + 3 + k * pBox.nInputs;
            if pDelays[k] != -ABC_INFINITY {
                DelayBest = Abc_MinFloat( DelayBest, pObj.timeReq - pDelays[i] ); }
        }
        pObjRes.timeReq = DelayBest;
        pObjRes.TravId = p.nTravIds;
    }
    return pObjThis.timeReq;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

