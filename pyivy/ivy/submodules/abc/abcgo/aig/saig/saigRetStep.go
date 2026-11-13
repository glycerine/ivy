package saig

/**CFile****************************************************************

  FileName    [saigRetStep.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Implementation of retiming steps.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigRetStep.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Performs the given number of retiming steps.]

  Description [Returns the pointer to node after retiming.]
               
  SideEffects [Remember to run Aig_ManSetCioIds() in advance.]

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeSteps(p *Aig_Man_t, nSteps int32, fForward int32, fAddBugs int32) int32 {
    var pObj, pObjNew *Aig_Obj_t;
    var RetValue, s, i int32;
    Aig_ManSetCioIds( p );
    Aig_ManFanoutStart( p );
    p.fCreatePios = 1;
    if fForward != 0 {
        Saig_ManMarkAutonomous( p );
        for  s = 0; s < nSteps; s++  {
            // inlining of macro Aig_ManForEachNode
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
                pObjNew = Saig_ManRetimeNodeFwd( p, pObj, fAddBugs && (s == 10) );
//                pObjNew = Saig_ManRetimeNodeFwd( p, pObj, 0 );
                if pObjNew == nil {
                    continue; }
                Aig_ObjReplace( p, pObj, pObjNew, 0 );
                break;
            }
            }
            if i == Vec_PtrSize(p.vObjs) {
                break; }
        }
    
    } else {
        for  s = 0; s < nSteps; s++  {
            // inlining of macro Saig_ManForEachLo
            for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
                pObjNew = Saig_ManRetimeNodeBwd( p, pObj );
                if pObjNew == nil {
                    continue; }
                Aig_ObjReplace( p, pObj, pObjNew, 0 );
                break;
            }
            if i == Vec_PtrSize(p.vObjs) {
                break; }
        }
    }
    p.fCreatePios = 0;
    Aig_ManFanoutStop( p );
    RetValue = Aig_ManCleanup( p );
    assert(RetValue == 0, "RetValue == 0");
    Aig_ManSetRegNum( p, p.nRegs ); 
    return s;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

