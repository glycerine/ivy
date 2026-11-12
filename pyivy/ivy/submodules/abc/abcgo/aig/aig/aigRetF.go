package aig

/**CFile****************************************************************

  FileName    [aigRetF.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Retiming frontier.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigRetF.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Mark the nodes reachable from the PIs in the reverse order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRetimeMark_rec(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    if pObj.fMarkB {
        return 1; }
    if Aig_ObjIsCi(pObj) || Aig_ObjIsConst1(pObj) {
        return 0; }
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return pObj.fMarkB; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Aig_ManRetimeMark_rec( p, Aig_ObjFanin0(pObj) ) != 0 {
        return pObj.fMarkB = 1; }
    if Aig_ObjIsNode(pObj) && (Aig_ManRetimeMark_rec( p, Aig_ObjFanin1(pObj) ) != 0) {
        return pObj.fMarkB = 1; }
    assert(pObj.fMarkB == 0, "pObj.fMarkB == 0");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Mark the nodes reachable from the true PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRetimeMark(p *Aig_Man_t) {
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var fChange, i int32;
    // mark the PIs
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(pObj.fMarkB == 0, "pObj.fMarkB == 0"); }
    }
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.fMarkB = 1;
    }
    // map registers into each other
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        assert(pObjLo.pNext == nil, "pObjLo.pNext == nil");
        assert(pObjLi.pNext == nil, "pObjLi.pNext == nil");
        pObjLo.pNext = pObjLi;
        pObjLi.pNext = pObjLo;
    }
    // iterativively mark the logic reachable from PIs
    fChange = 1;
    for  fChange != 0  {
        fChange = 0;
        Aig_ManIncrementTravId( p );
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            if pObj.fMarkB {
                continue; }
            if Aig_ManRetimeMark_rec( p, pObj ) != 0 {
                if pObj.pNext {
                    pObj.pNext.fMarkB = 1; }
                fChange = 1;
            }
        }
    }
    // clean register mapping
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        pObjLo.pNext = pObjLi.pNext = nil;
    }
}


/**Function*************************************************************

  Synopsis    [Performs forward retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRetimeFrontier(p *Aig_Man_t, nStepsMax int32) *Aig_Man_t {
    var pObj, pObjNew, pObjLo, pObjLo0, pObjLo1, pObjLi, pObjLi0, pObjLi1 *Aig_Obj_t;//, * pObjLi0_, * pObjLi1_, * pObjLi0__, * pObjLi1__;
    var i, Counter, fCompl, fChange int32;
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    // remove structural hashing table
    Aig_TableClear( p );
    // mark the retimable nodes
    Aig_ManRetimeMark( p );
    // mark the register outputs
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        pObjLo.fMarkA = 1;
        pObjLo.pNext = pObjLi;
        pObjLi.pNext = pObjLo;
    }
    // go through the nodes and find retimable ones
    Counter = 0;
    fChange = 1;
    for  fChange != 0  {
        fChange = 0;
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            if !pObj.fMarkB {
                continue; }
            if Aig_ObjIsBuf(pObj) {
                continue; }
            // get the real inputs of the node (skipping the buffers)
            pObjLo0 = Aig_ObjReal_rec( Aig_ObjChild0(pObj) );
            pObjLo1 = Aig_ObjReal_rec( Aig_ObjChild1(pObj) );
            if !Aig_Regular(pObjLo0).fMarkA || !Aig_Regular(pObjLo1).fMarkA {
                continue; }
            // remember complemented attribute
            fCompl = Aig_IsComplement(pObjLo0) & Aig_IsComplement(pObjLo1);
            // get the register inputs
//            pObjLi0_ = Aig_Regular(pObjLo0)->pNext;
//            pObjLi1_ = Aig_Regular(pObjLo1)->pNext;
//            pObjLi0__ = Aig_ObjChild0(Aig_Regular(pObjLo0)->pNext);
//            pObjLi1__ = Aig_ObjChild0(Aig_Regular(pObjLo1)->pNext);
            pObjLi0 = Aig_NotCond( Aig_ObjChild0(Aig_Regular(pObjLo0).pNext), Aig_IsComplement(pObjLo0) );
            pObjLi1 = Aig_NotCond( Aig_ObjChild0(Aig_Regular(pObjLo1).pNext), Aig_IsComplement(pObjLo1) );
            // create new node
            pObjNew = Aig_And( p, pObjLi0, pObjLi1 );
            pObjNew.fMarkB = 1;
            // create new register
            pObjLo = Aig_ObjCreateCi(p);
            pObjLo.fMarkA = 1;
            pObjLi = Aig_ObjCreateCo( p, Aig_NotCond(pObjNew, fCompl) );
            p.nRegs++;
            pObjLo.pNext = pObjLi;
            pObjLi.pNext = pObjLo;
            // add the buffer
            Aig_ObjDisconnect( p, pObj );
            pObj.Type = AIG_OBJ_BUF;
            p.nObjs[AIG_OBJ_AND]--;
            p.nObjs[AIG_OBJ_BUF]++;
            Aig_ObjConnect( p, pObj, Aig_NotCond(pObjLo, fCompl), nil );
            // create HAIG if defined
            // mark the change
            fChange = 1;
            // check the limit
            if Counter++; Counter >= nStepsMax {
                fChange = 0;
                break;
            }
        }
        }
    }
    // clean the markings
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        pObjLo.fMarkA = 0;
        pObjLo.pNext = pObjLi.pNext = nil;
    }
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.fMarkB = 0; }
    }
    // remove useless registers
    Aig_ManSeqCleanup( p );
    // rehash the nodes
    return Aig_ManDupOrdered( p ); 
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

