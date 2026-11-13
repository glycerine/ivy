package ssw

/**CFile****************************************************************

  FileName    [sswSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Sequential simulator used by the inductive prover.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswSim.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// simulation manager





////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns 1 if the node appears to be constant 1 candidate.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if the nodes appear equal.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Counts the number of 1s in the XOR of simulation data.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Checks implication.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Counts the number of 1s in the implication.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Counts the number of 1s in the implication.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Counts the number of one's in the pattern of the object.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Counts the number of one's in the pattern of the objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Generated const 0 pattern.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [[Generated const 1 pattern.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Creates the counter-example from the successful pattern.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if the one of the output is already non-constant 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Assigns random patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns random patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns constant patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns constant patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns constant patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assings distance-1 simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assings distance-1 simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Assings random simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assings random simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assings random simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Check if any of the POs becomes non-constant.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates AIG manager.]

  Description [Assumes that the PI simulation info is attached.]

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Converts simulation information to be not normallized.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates AIG manager.]

  Description [Assumes that the PI simulation info is attached.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSimulateOneDyn_rec(p *Ssw_Sml_t, pObj *Aig_Obj_t, f int32, pVisited *int32, nVisCounter int32) {
//    if ( Aig_ObjIsTravIdCurrent(p->pAig, pObj) )
//        return;
//    Aig_ObjSetTravIdCurrent(p->pAig, pObj);
    if pVisited[p.nFrames*pObj.Id+f] == nVisCounter {
        return; }
    pVisited[p.nFrames*pObj.Id+f] = nVisCounter;
    if Saig_ObjIsPi( p.pAig, pObj ) || Aig_ObjIsConst1(pObj) {
        return; }
    if Saig_ObjIsLo( p.pAig, pObj ) {
        if f == 0 {
            return; }
        Ssw_SmlSimulateOneDyn_rec( p, Saig_ObjLoToLi(p.pAig, pObj), f-1, pVisited, nVisCounter );
        Ssw_SmlNodeTransferNext( p, Saig_ObjLoToLi(p.pAig, pObj), pObj, f-1 );
        return;
    }
    if Saig_ObjIsLi( p.pAig, pObj ) {
        Ssw_SmlSimulateOneDyn_rec( p, Aig_ObjFanin0(pObj), f, pVisited, nVisCounter );
        Ssw_SmlNodeCopyFanin( p, pObj, f );
        return;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Ssw_SmlSimulateOneDyn_rec( p, Aig_ObjFanin0(pObj), f, pVisited, nVisCounter );
    Ssw_SmlSimulateOneDyn_rec( p, Aig_ObjFanin1(pObj), f, pVisited, nVisCounter );
    Ssw_SmlNodeSimulate( p, pObj, f );
}

/**Function*************************************************************

  Synopsis    [Simulates AIG manager.]

  Description [Assumes that the PI simulation info is attached.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSimulateOneFrame(p *Ssw_Sml_t) {
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i int32;
    var clk abctime;
clk = Abc_Clock();
    // simulate the nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Ssw_SmlNodeSimulate( p, pObj, 0 ); }
    }
    // copy simulation info into outputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i+Saig_ManPoNum(p.pAig))), 1); i++  {
        Ssw_SmlNodeCopyFanin( p, pObj, 0 );
    }
    // copy simulation info into the inputs
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(p.pAig, i)), 1); i++  {
        Ssw_SmlNodeTransferNext( p, pObjLi, pObjLo, 0 );
    }
p.timeSim += Abc_Clock() - clk;
p.nSimRounds++;
}


/**Function*************************************************************

  Synopsis    [Performs simulation of the uninitialized circuit.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Performs simulation of the initialized circuit.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Performs next round of sequential simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Creates sequential counter-example from the simulation info.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
