package tim

/**CFile****************************************************************

  FileName    [timTrav.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchy/timing manager.]

  Synopsis    [Manipulation of traversal IDs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: timTrav.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

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

  Synopsis    [Increments the trav ID of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManIncrementTravId(p *Tim_Man_t) {
    var i int32;
    if p.nTravIds >= (1<<30)-1 {
        p.nTravIds = 0;
        for  i = 0; i < p.nCis; i++  {
            p.pCis[i].TravId = 0;
        }
        for  i = 0; i < p.nCos; i++  {
            p.pCos[i].TravId = 0;
        }
    }
    assert(p.nTravIds < (1<<30)-1, "p.nTravIds < (1<<30)-1");
    p.nTravIds++;
}

/**Function*************************************************************

  Synopsis    [Label box inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManSetCurrentTravIdBoxInputs(p *Tim_Man_t, iBox int32) {
    var pBox *Tim_Box_t;
    var pObj *Tim_Obj_t;
    var i int32;
    pBox = Tim_ManBox( p, iBox );
    // inlining of macro Tim_ManBoxForEachInput
    for  i = 0; (i < (pBox).nInputs) && ((pObj) = Tim_ManBoxInput(p, pBox, i)); i++  {
        pObj.TravId = p.nTravIds;
    }
}

/**Function*************************************************************

  Synopsis    [Label box outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManSetCurrentTravIdBoxOutputs(p *Tim_Man_t, iBox int32) {
    var pBox *Tim_Box_t;
    var pObj *Tim_Obj_t;
    var i int32;
    pBox = Tim_ManBox( p, iBox );
    // inlining of macro Tim_ManBoxForEachOutput
    for  i = 0; (i < (pBox).nOutputs) && ((pObj) = Tim_ManBoxOutput(p, pBox, i)); i++  {
        pObj.TravId = p.nTravIds;
    }
}

/**Function*************************************************************

  Synopsis    [Label box inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManSetPreviousTravIdBoxInputs(p *Tim_Man_t, iBox int32) {
    var pBox *Tim_Box_t;
    var pObj *Tim_Obj_t;
    var i int32;
    pBox = Tim_ManBox( p, iBox );
    // inlining of macro Tim_ManBoxForEachInput
    for  i = 0; (i < (pBox).nInputs) && ((pObj) = Tim_ManBoxInput(p, pBox, i)); i++  {
        pObj.TravId = p.nTravIds - 1;
    }
}

/**Function*************************************************************

  Synopsis    [Label box outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManSetPreviousTravIdBoxOutputs(p *Tim_Man_t, iBox int32) {
    var pBox *Tim_Box_t;
    var pObj *Tim_Obj_t;
    var i int32;
    pBox = Tim_ManBox( p, iBox );
    // inlining of macro Tim_ManBoxForEachOutput
    for  i = 0; (i < (pBox).nOutputs) && ((pObj) = Tim_ManBoxOutput(p, pBox, i)); i++  {
        pObj.TravId = p.nTravIds - 1;
    }
}

/**Function*************************************************************

  Synopsis    [Updates required time of the CI.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManIsCiTravIdCurrent(p *Tim_Man_t, iCi int32) int32 {
    assert(iCi < p.nCis, "iCi < p.nCis");
    assert(p.fUseTravId, "p.fUseTravId");
    return p.pCis[iCi].TravId == p.nTravIds;
}

/**Function*************************************************************

  Synopsis    [Updates required time of the CO.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManIsCoTravIdCurrent(p *Tim_Man_t, iCo int32) int32 {
    assert(iCo < p.nCos, "iCo < p.nCos");
    assert(p.fUseTravId, "p.fUseTravId");
    return p.pCos[iCo].TravId == p.nTravIds;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

