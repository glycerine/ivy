package int

/**CFile****************************************************************

  FileName    [intFrames.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Sequential AIG unrolling for interpolation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 24, 2008.]

  Revision    [$Id: intFrames.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "intInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create timeframes of the manager for interpolation.]

  Description [The resulting manager is combinational. The primary inputs
  corresponding to register outputs are ordered first. The only POs of the 
  manager is the property output of the last timeframe.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManFramesInter(pAig *Aig_Man_t, nFrames int32, fAddRegOuts int32, fUseTwoFrames int32) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var pLastPo *Aig_Obj_t  = nil;
    var i, f int32;
    assert(Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0");
    assert(Saig_ManPoNum(pAig)-Saig_ManConstrNum(pAig) == 1, "Saig_ManPoNum(pAig)-Saig_ManConstrNum(pAig) == 1");
    pFrames = Aig_ManStart( Aig_ManNodeNum(pAig) * nFrames );
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pFrames );
    // create variables for register outputs
    if fAddRegOuts != 0 {
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
            pObj.pData = Aig_ManConst0( pFrames );
        }
    
    } else {
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
            pObj.pData = Aig_ObjCreateCi( pFrames );
        }
    }
    // add timeframes
    for  f = 0; f < nFrames; f++  {
        // create PI nodes for this frame
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            pObj.pData = Aig_ObjCreateCi( pFrames );
        }
        // add internal nodes of this frame
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.pData = Aig_And( pFrames, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
        }
        // add outputs for constraints
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            if i < Saig_ManPoNum(pAig)-Saig_ManConstrNum(pAig) {
                continue; }
            Aig_ObjCreateCo( pFrames, Aig_Not( Aig_ObjChild0Copy(pObj) ) );
        }
        if f == nFrames - 1 {
            break; }
        // remember the last PO
        pObj = Aig_ManCo( pAig, 0 );
        pLastPo = Aig_ObjChild0Copy(pObj);
        // save register inputs
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
            pObj.pData = Aig_ObjChild0Copy(pObj);
        }
        // transfer to register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
            pObjLo.pData = pObjLi.pData;
        }
    }
    // create POs for each register output
    if fAddRegOuts != 0 {
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
            Aig_ObjCreateCo( pFrames, Aig_ObjChild0Copy(pObj) );
        }
    
    // create the only PO of the manager
    } else {
        pObj = Aig_ManCo( pAig, 0 );
        // add the last PO
        if pLastPo == nil || !(fUseTwoFrames != 0) {
            pLastPo = Aig_ObjChild0Copy(pObj);
        } else {
            pLastPo = Aig_Or( pFrames, pLastPo, Aig_ObjChild0Copy(pObj) ); }
        Aig_ObjCreateCo( pFrames, pLastPo );
    }
    Aig_ManCleanup( pFrames );
    return pFrames;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

