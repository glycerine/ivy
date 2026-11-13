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

  Synopsis    [Performs one retiming step forward.]

  Description [Returns the pointer to the register output after retiming.]
               
  SideEffects [Remember to run Aig_ManSetCioIds() in advance.]

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs one retiming step backward.]

  Description [Returns the pointer to node after retiming.]
               
  SideEffects [Remember to run Aig_ManSetCioIds() in advance.]

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeNodeBwd(p *Aig_Man_t, pObjLo *Aig_Obj_t) *Aig_Obj_t {
    var pFanin0, pFanin1 *Aig_Obj_t;
    var pLo0New, pLo1New *Aig_Obj_t;
    var pLi0New, pLi1New *Aig_Obj_t;
    var pObj, pObjNew, pObjLi *Aig_Obj_t;
    var fCompl0, fCompl1 int32;

    assert(Saig_ManRegNum(p) > 0, "Saig_ManRegNum(p) > 0");
    assert(Aig_ObjCioId(pObjLo) > 0, "Aig_ObjCioId(pObjLo) > 0");
    assert(Saig_ObjIsLo(p, pObjLo), "Saig_ObjIsLo(p, pObjLo)");

    // get the corresponding latch input
    pObjLi = Saig_ManLi( p, Aig_ObjCioId(pObjLo) - Saig_ManPiNum(p) );

    // get the node
    pObj = Aig_ObjFanin0(pObjLi);
    if !Aig_ObjIsNode(pObj) {
        return nil; }

    // get the fanins
    pFanin0 = Aig_ObjFanin0(pObj);
    pFanin1 = Aig_ObjFanin1(pObj);

    // get the complemented attributes of the fanins
    fCompl0 = Aig_ObjFaninC0(pObj) ^ Aig_ObjFaninC0(pObjLi);
    fCompl1 = Aig_ObjFaninC1(pObj) ^ Aig_ObjFaninC0(pObjLi);

    // create latch inputs
    pLi0New = Aig_ObjCreateCo( p, Aig_NotCond(pFanin0, fCompl0) );
    pLi0New.CioId = Aig_ManCoNum(p) - 1;
    pLi1New = Aig_ObjCreateCo( p, Aig_NotCond(pFanin1, fCompl1) );
    pLi1New.CioId = Aig_ManCoNum(p) - 1;

    // create latch outputs
    pLo0New = Aig_ObjCreateCi(p);
    pLo0New.CioId = Aig_ManCiNum(p) - 1;
    pLo1New = Aig_ObjCreateCi(p);
    pLo1New.CioId = Aig_ManCiNum(p) - 1;
    pLo0New = Aig_NotCond( pLo0New, fCompl0 );
    pLo1New = Aig_NotCond( pLo1New, fCompl1 );
    p.nRegs += 2;

    // create node
    pObjNew = Aig_And( p, pLo0New, pLo1New );
//    assert( pObjNew->fPhase == 0 );
    return pObjNew;
}

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

