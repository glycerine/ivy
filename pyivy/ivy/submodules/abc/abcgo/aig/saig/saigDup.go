package saig

import "fmt"

/**CFile****************************************************************

  FileName    [saigDup.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Various duplication procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigDup.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Duplicates while ORing the POs of sequential circuit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManDupOrpos(pAig *Aig_Man_t) *Aig_Man_t {
    var pAigNew *Aig_Man_t;
    var pObj, pMiter *Aig_Obj_t;
    var i int32;
    if pAig.nConstrs > 0 {
        fmt.Printf("The AIG manager should have no constraints.\n");
        return nil;
    }
    // start the new manager
    pAigNew = Aig_ManStart( Aig_ManNodeNum(pAig) );
    pAigNew.pName = Abc_UtilStrsav( pAig.pName );
    pAigNew.nConstrs = pAig.nConstrs;
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pAigNew );
    // create variables for PIs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pAigNew );
    }
    // add internal nodes of this frame
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pAigNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // create PO of the circuit
    pMiter = Aig_ManConst0( pAigNew );
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        pMiter = Aig_Or( pAigNew, pMiter, Aig_ObjChild0Copy(pObj) );
    }
    Aig_ObjCreateCo( pAigNew, pMiter );
    // transfer to register outputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        Aig_ObjCreateCo( pAigNew, Aig_ObjChild0Copy(pObj) );
    }
    Aig_ManCleanup( pAigNew );
    Aig_ManSetRegNum( pAigNew, Aig_ManRegNum(pAig) );
    return pAigNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates while ORing the POs of sequential circuit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCreateEquivMiter(pAig *Aig_Man_t, vPairs *Vec_Int_t, fAddOuts int32) *Aig_Man_t {
    var pAigNew *Aig_Man_t;
    var pObj, pObj2, pMiter *Aig_Obj_t;
    var i int32;
    if pAig.nConstrs > 0 {
        fmt.Printf("The AIG manager should have no constraints.\n");
        return nil;
    }
    // start the new manager
    pAigNew = Aig_ManStart( Aig_ManNodeNum(pAig) );
    pAigNew.pName = Abc_UtilStrsav( pAig.pName );
    pAigNew.nConstrs = pAig.nConstrs;
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pAigNew );
    // create variables for PIs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pAigNew );
    }
    // add internal nodes of this frame
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pAigNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // create POs
    assert(Vec_IntSize(vPairs) % 2 == 0, "Vec_IntSize(vPairs) % 2 == 0");
    // inlining of macro Aig_ManForEachObjVec
    for  i = 0; i < Vec_IntSize(vPairs) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vPairs,i))), 1); i++  {
        pObj2  = Aig_ManObj( pAig, Vec_IntEntry(vPairs, ++i) );
        pMiter = Aig_Exor( pAigNew, (Aig_Obj_t *)pObj.pData, (Aig_Obj_t *)pObj2.pData );
        pMiter = Aig_NotCond( pMiter, pObj.fPhase ^ pObj2.fPhase );
        Aig_ObjCreateCo( pAigNew, pMiter );
    }
    // transfer to register outputs
    if fAddOuts != 0 {
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        Aig_ObjCreateCo( pAigNew, Aig_ObjChild0Copy(pObj) );
    } }
    Aig_ManCleanup( pAigNew );
    if fAddOuts != 0 {
    Aig_ManSetRegNum( pAigNew, Aig_ManRegNum(pAig) ); }
    return pAigNew;
}

/**Function*************************************************************

  Synopsis    [Trims the model by removing PIs without fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManTrimPis(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i, fAllPisHaveNoRefs int32;
    // check the refs of PIs    
    fAllPisHaveNoRefs = 1;
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        if pObj.nRefs {
            fAllPisHaveNoRefs = 0; }
    }
    // start the new manager
    pNew = Aig_ManStart( Aig_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.nConstrs = p.nConstrs;
    // start mapping of the CI numbers
    pNew.vCiNumsOrig = Vec_IntAlloc( Aig_ManCiNum(p) );
    // map const and primary inputs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        if (fAllPisHaveNoRefs != 0) || pObj.nRefs || Saig_ObjIsLo(p, pObj) {
            pObj.pData = Aig_ObjCreateCi( pNew );
            Vec_IntPush( pNew.vCiNumsOrig, Vec_IntEntry(p.vCiNumsOrig, i) );
        }
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManAbstractionDfs_rec(pNew *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    if pObj.pData {
        return (Aig_Obj_t *)pObj.pData; }
    Saig_ManAbstractionDfs_rec( pNew, Aig_ObjFanin0(pObj) );
    Saig_ManAbstractionDfs_rec( pNew, Aig_ObjFanin1(pObj) );
    return (Aig_Obj_t *)(pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ));
}

/**Function*************************************************************

  Synopsis    [Performs abstraction of the AIG to preserve the included flops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManDupAbstraction(p *Aig_Man_t, vFlops *Vec_Int_t) *Aig_Man_t { 
    var pNew *Aig_Man_t;//, * pTemp;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i, Entry int32;
    Aig_ManCleanData( p );
    // start the new manager
    pNew = Aig_ManStart( 5000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    // map the constant node
    Aig_ManConst1(p).pData = Aig_ManConst1( pNew );
    // label included flops
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFlops)) && (((Entry) = Vec_IntEntry(vFlops, i)), 1); i++  {
        pObjLi = Saig_ManLi( p, Entry );
        assert(pObjLi.fMarkA == 0, "pObjLi.fMarkA == 0");
        pObjLi.fMarkA = 1;
        pObjLo = Saig_ManLo( p, Entry );
        assert(pObjLo.fMarkA == 0, "pObjLo.fMarkA == 0");
        pObjLo.fMarkA = 1;
    }
    // create variables for PIs
    assert(p.vCiNumsOrig == nil, "p.vCiNumsOrig == nil");
    pNew.vCiNumsOrig = Vec_IntAlloc( Aig_ManCiNum(p) );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        if !pObj.fMarkA {
            pObj.pData = Aig_ObjCreateCi( pNew );
            Vec_IntPush( pNew.vCiNumsOrig, i );
        }
    }
    // create variables for LOs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        if pObj.fMarkA {
            pObj.fMarkA = 0;
            pObj.pData = Aig_ObjCreateCi( pNew );
            Vec_IntPush( pNew.vCiNumsOrig, i );
        }
    }
    // add internal nodes 
//    Aig_ManForEachNode( p, pObj, i )
//        pObj->pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    // create POs
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Saig_ManAbstractionDfs_rec( pNew, Aig_ObjFanin0(pObj) );
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    // create LIs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if pObj.fMarkA {
            pObj.fMarkA = 0;
            Saig_ManAbstractionDfs_rec( pNew, Aig_ObjFanin0(pObj) );
            Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        }
    }
    Aig_ManSetRegNum( pNew, Vec_IntSize(vFlops) );
    Aig_ManSeqCleanup( pNew );
    // remove PIs without fanout
//    pNew = Saig_ManTrimPis( pTemp = pNew );
//    Aig_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Resimulates the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManVerifyCex(pAig *Aig_Man_t, p *Abc_Cex_t) int32 {
    var pObj, pObjRi, pObjRo *Aig_Obj_t;
    var RetValue, i, k, iBit int32;
    Aig_ManCleanMarkB(pAig);
    Aig_ManConst1(pAig).fMarkB = 1;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.fMarkB = Abc_InfoHasBit(p.pData, iBit++);
    }
    for  i = 0; i <= p.iFrame; i++  {
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, k)), 1); k++  {
            pObj.fMarkB = Abc_InfoHasBit(p.pData, iBit++);
        }
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, k)), 1); k++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.fMarkB = (Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj)) & 
                           (Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj)); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, k)), 1); k++  {
            pObj.fMarkB = Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj);
        }
        if i == p.iFrame {
            break; }
        // inlining of macro Saig_ManForEachLiLo
        for  k = 0; (k < Saig_ManRegNum(pAig)) && (((pObjRi) = Saig_ManLi(pAig, k)), 1)            && (((pObjRo)=Saig_ManLo(pAig, k)), 1); k++  {
            pObjRo.fMarkB = pObjRi.fMarkB;
        }
    }
    assert(iBit == p.nBits, "iBit == p.nBits");
    RetValue = Aig_ManCo(pAig, p.iPo).fMarkB;
    Aig_ManCleanMarkB(pAig);
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Resimulates the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManVerifyCexNoClear(pAig *Aig_Man_t, p *Abc_Cex_t) int32 {
    var pObj, pObjRi, pObjRo *Aig_Obj_t;
    var RetValue, i, k, iBit int32;
    Aig_ManCleanMarkB(pAig);
    Aig_ManConst1(pAig).fMarkB = 1;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.fMarkB = Abc_InfoHasBit(p.pData, iBit++);
    }
    for  i = 0; i <= p.iFrame; i++  {
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, k)), 1); k++  {
            pObj.fMarkB = Abc_InfoHasBit(p.pData, iBit++);
        }
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, k)), 1); k++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.fMarkB = (Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj)) & 
                           (Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj)); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, k)), 1); k++  {
            pObj.fMarkB = Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj);
        }
        if i == p.iFrame {
            break; }
        // inlining of macro Saig_ManForEachLiLo
        for  k = 0; (k < Saig_ManRegNum(pAig)) && (((pObjRi) = Saig_ManLi(pAig, k)), 1)            && (((pObjRo)=Saig_ManLo(pAig, k)), 1); k++  {
            pObjRo.fMarkB = pObjRi.fMarkB;
        }
    }
    assert(iBit == p.nBits, "iBit == p.nBits");
    RetValue = Aig_ManCo(pAig, p.iPo).fMarkB;
    //Aig_ManCleanMarkB(pAig);
    return RetValue;
}
func Saig_ManReturnFailingState(pAig *Aig_Man_t, p *Abc_Cex_t, fNextOne int32) *Vec_Int_t {
    var pObj *Aig_Obj_t;
    var vState *Vec_Int_t;
    var i int32;
    var RetValue int32 = Saig_ManVerifyCexNoClear( pAig, p );
    if RetValue == 0 {
        Aig_ManCleanMarkB(pAig);
        fmt.Printf("CEX does fail the given sequential miter.\n");
        return nil;
    }
    vState = Vec_IntAlloc( Aig_ManRegNum(pAig) );
    if fNextOne != 0 {
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
            Vec_IntPush( vState, pObj.fMarkB );
        }
    
    } else {
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
            Vec_IntPush( vState, pObj.fMarkB );
        }
    }
    Aig_ManCleanMarkB(pAig);
    return vState;
}

/**Function*************************************************************

  Synopsis    [Resimulates the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManExtendCex(pAig *Aig_Man_t, p *Abc_Cex_t) *Abc_Cex_t {
    var pNew *Abc_Cex_t;
    var pObj, pObjRi, pObjRo *Aig_Obj_t;
    var RetValue, i, k, iBit int32;
    // create new counter-example
    pNew = Abc_CexAlloc( 0, Aig_ManCiNum(pAig), p.iFrame + 1 );
    pNew.iPo = p.iPo;
    pNew.iFrame = p.iFrame;
    // simulate the AIG
    Aig_ManCleanMarkB(pAig);
    Aig_ManConst1(pAig).fMarkB = 1;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.fMarkB = Abc_InfoHasBit(p.pData, iBit++);
    }
    for  i = 0; i <= p.iFrame; i++  {
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, k)), 1); k++  {
            pObj.fMarkB = Abc_InfoHasBit(p.pData, iBit++);
        }
        ///////// write PI+LO values ////////////
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, k)), 1); k++  {
            if pObj.fMarkB {
                Abc_InfoSetBit(pNew.pData, Aig_ManCiNum(pAig)*i + k); }
        }
        /////////////////////////////////////////
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, k)), 1); k++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.fMarkB = (Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj)) & 
                           (Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj)); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, k)), 1); k++  {
            pObj.fMarkB = Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj);
        }
        if i == p.iFrame {
            break; }
        // inlining of macro Saig_ManForEachLiLo
        for  k = 0; (k < Saig_ManRegNum(pAig)) && (((pObjRi) = Saig_ManLi(pAig, k)), 1)            && (((pObjRo)=Saig_ManLo(pAig, k)), 1); k++  {
            pObjRo.fMarkB = pObjRi.fMarkB;
        }
    }
    assert(iBit == p.nBits, "iBit == p.nBits");
    RetValue = Aig_ManCo(pAig, p.iPo).fMarkB;
    Aig_ManCleanMarkB(pAig);
    if RetValue == 0 {
        fmt.Printf("Saig_ManExtendCex(): The counter-example is invalid!!!\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Resimulates the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManFindFailedPoCex(pAig *Aig_Man_t, p *Abc_Cex_t) int32 {
    var pObj, pObjRi, pObjRo *Aig_Obj_t;
    var RetValue, i, k, iBit int32;
    Aig_ManCleanMarkB(pAig);
    Aig_ManConst1(pAig).fMarkB = 1;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.fMarkB = Abc_InfoHasBit(p.pData, iBit++);
    }
    for  i = 0; i <= p.iFrame; i++  {
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, k)), 1); k++  {
            pObj.fMarkB = Abc_InfoHasBit(p.pData, iBit++);
        }
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, k)), 1); k++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.fMarkB = (Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj)) & 
                           (Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj)); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, k)), 1); k++  {
            pObj.fMarkB = Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj);
        }
        if i == p.iFrame {
            break; }
        // inlining of macro Saig_ManForEachLiLo
        for  k = 0; (k < Saig_ManRegNum(pAig)) && (((pObjRi) = Saig_ManLi(pAig, k)), 1)            && (((pObjRo)=Saig_ManLo(pAig, k)), 1); k++  {
            pObjRo.fMarkB = pObjRi.fMarkB;
        }
    }
    assert(iBit == p.nBits, "iBit == p.nBits");
    // remember the number of failed output
    RetValue = -1;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        if pObj.fMarkB {
            RetValue = i;
            break;
        }
    }
    Aig_ManCleanMarkB(pAig);
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Duplicates while ORing the POs of sequential circuit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManDupWithPhase(pAig *Aig_Man_t, vInit *Vec_Int_t) *Aig_Man_t {
    var pAigNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    assert(Aig_ManRegNum(pAig) <= Vec_IntSize(vInit), "Aig_ManRegNum(pAig) <= Vec_IntSize(vInit)");
    // start the new manager
    pAigNew = Aig_ManStart( Aig_ManNodeNum(pAig) );
    pAigNew.pName = Abc_UtilStrsav( pAig.pName );
    pAigNew.nConstrs = pAig.nConstrs;
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pAigNew );
    // create variables for PIs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pAigNew );
    }
    // update the flop variables
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.pData = Aig_NotCond( (Aig_Obj_t *)pObj.pData, Vec_IntEntry(vInit, i) );
    }
    // add internal nodes of this frame
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pAigNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // transfer to register outputs
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        Aig_ObjCreateCo( pAigNew, Aig_ObjChild0Copy(pObj) );
    }
    // update the flop variables
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        Aig_ObjCreateCo( pAigNew, Aig_NotCond(Aig_ObjChild0Copy(pObj), Vec_IntEntry(vInit, i)) );
    }
    // finalize
    Aig_ManCleanup( pAigNew );
    Aig_ManSetRegNum( pAigNew, Aig_ManRegNum(pAig) );
    return pAigNew;
}

/**Function*************************************************************

  Synopsis    [Copy an AIG structure related to the selected POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManDupCones_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vLeaves *Vec_Ptr_t, vNodes *Vec_Ptr_t, vRoots *Vec_Ptr_t) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Aig_ObjIsNode(pObj) {
        Saig_ManDupCones_rec( p, Aig_ObjFanin0(pObj), vLeaves, vNodes, vRoots );
        Saig_ManDupCones_rec( p, Aig_ObjFanin1(pObj), vLeaves, vNodes, vRoots );
        Vec_PtrPush( vNodes, pObj );
    
    } else if Aig_ObjIsCo(pObj) {
        Saig_ManDupCones_rec( p, Aig_ObjFanin0(pObj), vLeaves, vNodes, vRoots );
    } else if Saig_ObjIsLo(p, pObj) {
        Vec_PtrPush( vRoots, Saig_ObjLoToLi(p, pObj) );
    } else if Saig_ObjIsPi(p, pObj) {
        Vec_PtrPush( vLeaves, pObj );
    } else { assert(false, "0"); }
}
func Saig_ManDupCones(pAig *Aig_Man_t, pPos *int32, nPos int32) *Aig_Man_t {
    var pAigNew *Aig_Man_t;
    var vLeaves, vNodes, vRoots *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;

    // collect initial POs
    vLeaves = Vec_PtrAlloc( 100 );
    vNodes = Vec_PtrAlloc( 100 );
    vRoots = Vec_PtrAlloc( 100 );
    for  i = 0; i < nPos; i++  {
        Vec_PtrPush( vRoots, Aig_ManCo(pAig, pPos[i]) );
    }

    // mark internal nodes
    Aig_ManIncrementTravId( pAig );
    Aig_ObjSetTravIdCurrent( pAig, Aig_ManConst1(pAig) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Saig_ManDupCones_rec( pAig, pObj, vLeaves, vNodes, vRoots );
    }

    // start the new manager
    pAigNew = Aig_ManStart( Vec_PtrSize(vNodes) );
    pAigNew.pName = Abc_UtilStrsav( pAig.pName );
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pAigNew );
    // create PIs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pAigNew );
    }
    // create LOs
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = nPos; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Saig_ObjLiToLo(pAig, pObj).pData = Aig_ObjCreateCi( pAigNew );
    }
    // create internal nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.pData = Aig_And( pAigNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }
    // create COs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Aig_ObjCreateCo( pAigNew, Aig_ObjChild0Copy(pObj) );
    }
    // finalize
    Aig_ManSetRegNum( pAigNew, Vec_PtrSize(vRoots)-nPos );
    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vNodes );
    Vec_PtrFree( vRoots );
    return pAigNew;

}

#ifndef ABC_USE_CUDD
int Aig_ManVerifyUsingBdds( Aig_Man_t * pInit, Saig_ParBbr_t * pPars ) { return 0; }
void Bbr_ManSetDefaultParams( Saig_ParBbr_t * p ) {}
#endif

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

