package saig

/**CFile****************************************************************

  FileName    [saigDual.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Various duplication procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigDual.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Saig_ObjSetDual(vCopies *Vec_Ptr_t, Id int32, fPos int32, pItem *Aig_Obj_t) { Vec_PtrWriteEntry( vCopies, 2*Id+fPos, pItem );         }
func Saig_ObjDual(vCopies *Vec_Ptr_t, Id int32, fPos int32) *Aig_Obj_t { return (Aig_Obj_t *)Vec_PtrEntry( vCopies, 2*Id+fPos ); }

func Saig_ObjDualFanin(pAigNew *Aig_Man_t, vCopies *Vec_Ptr_t, pObj *Aig_Obj_t, iFanin int32, ppRes0 **Aig_Obj_t, ppRes1 **Aig_Obj_t) {

    var pTemp0, pTemp1, pCare *Aig_Obj_t;
    var fCompl int32;
    assert(iFanin == 0 || iFanin == 1, "iFanin == 0 || iFanin == 1");
    if iFanin == 0 {
        pTemp0 = Saig_ObjDual( vCopies, Aig_ObjFaninId0(pObj), 0 );
        pTemp1 = Saig_ObjDual( vCopies, Aig_ObjFaninId0(pObj), 1 );
        fCompl = Aig_ObjFaninC0( pObj );
    
    } else {
        pTemp0 = Saig_ObjDual( vCopies, Aig_ObjFaninId1(pObj), 0 );
        pTemp1 = Saig_ObjDual( vCopies, Aig_ObjFaninId1(pObj), 1 );
        fCompl = Aig_ObjFaninC1( pObj );
    }
    if fCompl != 0 {
        pCare   = Aig_Or( pAigNew, pTemp0, pTemp1 );
        *ppRes0 = Aig_And( pAigNew, pTemp1, pCare );
        *ppRes1 = Aig_And( pAigNew, pTemp0, pCare );
    
    } else {
        *ppRes0 = pTemp0;
        *ppRes1 = pTemp1;
    }
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Transforms sequential AIG into dual-rail miter.]

  Description [Transforms sequential AIG into a miter encoding ternary
  problem formulated as follows "none of the POs has a ternary value".
  Interprets the first nDualPis as having ternary value.  Sets flops
  to have ternary intial value when fDualFfs is set to 1.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManDupDual(pAig *Aig_Man_t, vDcFlops *Vec_Int_t, nDualPis int32, fDualFfs int32, fMiterFfs int32, fComplPo int32, fCheckZero int32, fCheckOne int32) *Aig_Man_t {
    var vCopies *Vec_Ptr_t;
    var pAigNew *Aig_Man_t;
    var pObj, pTemp0, pTemp1, pTemp2, pTemp3, pCare, pMiter *Aig_Obj_t;
    var i int32;
    assert(Saig_ManPoNum(pAig) > 0, "Saig_ManPoNum(pAig) > 0");
    assert(nDualPis >= 0 && nDualPis <= Saig_ManPiNum(pAig), "nDualPis >= 0 && nDualPis <= Saig_ManPiNum(pAig)");
    assert(vDcFlops == nil || Vec_IntSize(vDcFlops) == Aig_ManRegNum(pAig), "vDcFlops == nil || Vec_IntSize(vDcFlops) == Aig_ManRegNum(pAig)");
    vCopies = Vec_PtrStart( 2*Aig_ManObjNum(pAig) );
    // start the new manager
    pAigNew = Aig_ManStart( Aig_ManNodeNum(pAig) );
    pAigNew.pName = Abc_UtilStrsav( pAig.pName );
    // map the constant node
    Saig_ObjSetDual( vCopies, 0, 0, Aig_ManConst0(pAigNew) );
    Saig_ObjSetDual( vCopies, 0, 1, Aig_ManConst1(pAigNew) );
    // create variables for PIs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        if i < nDualPis {
            pTemp0 = Aig_ObjCreateCi( pAigNew );
            pTemp1 = Aig_ObjCreateCi( pAigNew );
        
        } else if i < Saig_ManPiNum(pAig) {
            pTemp1 = Aig_ObjCreateCi( pAigNew );
            pTemp0 = Aig_Not( pTemp1 );
        
        } else {
            pTemp0 = Aig_ObjCreateCi( pAigNew );
            pTemp1 = Aig_ObjCreateCi( pAigNew );
            if vDcFlops != nil {
                pTemp0 = Aig_NotCond( pTemp0, !Vec_IntEntry(vDcFlops, i-Saig_ManPiNum(pAig)) );
            } else {
                pTemp0 = Aig_NotCond( pTemp0, !fDualFfs ); }
        }
        Saig_ObjSetDual( vCopies, Aig_ObjId(pObj), 0, Aig_And(pAigNew, pTemp0, Aig_Not(pTemp1)) );
        Saig_ObjSetDual( vCopies, Aig_ObjId(pObj), 1, Aig_And(pAigNew, pTemp1, Aig_Not(pTemp0)) );
    }
    // create internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Saig_ObjDualFanin( pAigNew, vCopies, pObj, 0, &pTemp0, &pTemp1 );
        Saig_ObjDualFanin( pAigNew, vCopies, pObj, 1, &pTemp2, &pTemp3 );
        Saig_ObjSetDual( vCopies, Aig_ObjId(pObj), 0, Aig_Or (pAigNew, pTemp0, pTemp2) );
        Saig_ObjSetDual( vCopies, Aig_ObjId(pObj), 1, Aig_And(pAigNew, pTemp1, pTemp3) );
    }
    }
    // create miter and flops
    pMiter = Aig_ManConst0(pAigNew);
    if fMiterFfs != 0 {
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
            Saig_ObjDualFanin( pAigNew, vCopies, pObj, 0, &pTemp0, &pTemp1 );
            if fCheckZero != 0 {
                pCare  = Aig_And( pAigNew, pTemp0, Aig_Not(pTemp1) );
                pMiter = Aig_Or( pAigNew, pMiter, pCare );
            
            } else if fCheckOne != 0 {
                pCare  = Aig_And( pAigNew, Aig_Not(pTemp0), pTemp1 );
                pMiter = Aig_Or( pAigNew, pMiter, pCare );
            
            } else { // check X
            
                pCare  = Aig_And( pAigNew, Aig_Not(pTemp0), Aig_Not(pTemp1) );
                pMiter = Aig_Or( pAigNew, pMiter, pCare );
            }
        }
    
    } else {
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            Saig_ObjDualFanin( pAigNew, vCopies, pObj, 0, &pTemp0, &pTemp1 );
            if fCheckZero != 0 {
                pCare  = Aig_And( pAigNew, pTemp0, Aig_Not(pTemp1) );
                pMiter = Aig_Or( pAigNew, pMiter, pCare );
            
            } else if fCheckOne != 0 {
                pCare  = Aig_And( pAigNew, Aig_Not(pTemp0), pTemp1 );
                pMiter = Aig_Or( pAigNew, pMiter, pCare );
            
            } else { // check X
            
                pCare  = Aig_And( pAigNew, Aig_Not(pTemp0), Aig_Not(pTemp1) );
                pMiter = Aig_Or( pAigNew, pMiter, pCare );
            }
        }
    }
    // create PO
    pMiter = Aig_NotCond( pMiter, fComplPo );
    Aig_ObjCreateCo( pAigNew, pMiter );
    // create flops
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        Saig_ObjDualFanin( pAigNew, vCopies, pObj, 0, &pTemp0, &pTemp1 );
        if vDcFlops != nil {
            pTemp0 = Aig_NotCond( pTemp0, !Vec_IntEntry(vDcFlops, i) );
        } else {
            pTemp0 = Aig_NotCond( pTemp0, !fDualFfs ); }
        Aig_ObjCreateCo( pAigNew, pTemp0 );
        Aig_ObjCreateCo( pAigNew, pTemp1 );
    }
    // set the flops
    Aig_ManSetRegNum( pAigNew, 2 * Aig_ManRegNum(pAig) );
    Aig_ManCleanup( pAigNew );
    Vec_PtrFree( vCopies );
    return pAigNew;
}

/**Function*************************************************************

  Synopsis    [Transforms sequential AIG to block the PO for N cycles.]

  Description [This procedure should be applied to a safety property 
  miter to make the propetry 'true' (const 0) during the first N cycles.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBlockPo(pAig *Aig_Man_t, nCycles int32) {
    var pObj, pCond, pPrev, pTemp *Aig_Obj_t;
    var i int32;
    assert(nCycles > 0, "nCycles > 0");
    // add N flops (assuming 1-hot encoding of cycles)
    pPrev = Aig_ManConst1(pAig);
    pCond = Aig_ManConst1(pAig);
    for  i = 0; i < nCycles; i++  {
        Aig_ObjCreateCo( pAig, pPrev );
        pPrev = Aig_ObjCreateCi( pAig );
        pCond = Aig_And( pAig, pCond, pPrev );
    }
    // update the POs
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        pTemp = Aig_And( pAig, Aig_ObjChild0(pObj), pCond );
        Aig_ObjPatchFanin0( pAig, pObj, pTemp );
    }
    // set the flops
    Aig_ManSetRegNum( pAig, Aig_ManRegNum(pAig) + nCycles );
    Aig_ManCleanup( pAig );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

