package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigScl.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Sequential cleanup.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigScl.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

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

  Synopsis    [Remaps the manager.]

  Description [Map in the array specifies for each CI node the node that
  should be used after remapping.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRemap(p *Aig_Man_t, vMap *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjMapped *Aig_Obj_t;
    var i, nTruePis int32;
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nAsserts = p.nAsserts;
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    assert(p.vFlopNums == nil || Vec_IntSize(p.vFlopNums) == p.nRegs, "p.vFlopNums == nil || Vec_IntSize(p.vFlopNums) == p.nRegs");
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    if p.vFlopReprs {
        pNew.vFlopReprs = Vec_IntDup( p.vFlopReprs ); }
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi(pNew);
    }
    // implement the mapping
    nTruePis = Aig_ManCiNum(p)-Aig_ManRegNum(p);
    if p.vFlopReprs {
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            pObj.pNext = (Aig_Obj_t *)int64(Vec_IntEntry( p.vFlopNums, i-nTruePis ));
        }
    }
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObjMapped = (Aig_Obj_t *)Vec_PtrEntry( vMap, i );
        pObj.pData = Aig_NotCond( (Aig_Obj_t *)Aig_Regular(pObjMapped).pData, Aig_IsComplement(pObjMapped) );
        if pNew.vFlopReprs && i >= nTruePis && pObj != pObjMapped {
            Vec_IntPush( pNew.vFlopReprs, Aig_ObjCioId(pObj) );
            if Aig_ObjIsConst1( Aig_Regular(pObjMapped) ) {
                Vec_IntPush( pNew.vFlopReprs, -1 );
            } else {
                assert(!Aig_IsComplement(pObjMapped), "!Aig_IsComplement(pObjMapped)");
                assert(Aig_ObjIsCi(pObjMapped), "Aig_ObjIsCi(pObjMapped)");
                assert(Aig_ObjCioId(pObj) != Aig_ObjCioId(pObjMapped), "Aig_ObjCioId(pObj) != Aig_ObjCioId(pObjMapped)");
                Vec_IntPush( pNew.vFlopReprs, Aig_ObjCioId(pObjMapped) );
            }
        }
    }
    if p.vFlopReprs {
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            pObj.pNext = nil;
        }
    }
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if Aig_ObjIsBuf(pObj) {
            pObj.pData = Aig_ObjChild0Copy(pObj);
        } else if Aig_ObjIsNode(pObj) {
            pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // add the POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    assert(Aig_ManNodeNum(p) >= Aig_ManNodeNum(pNew), "Aig_ManNodeNum(p) >= Aig_ManNodeNum(pNew)");
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManRemap(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Returns the number of dangling nodes removed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSeqCleanup_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vNodes *Vec_Ptr_t) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    // collect latch input corresponding to unmarked PI (latch output)
    if Aig_ObjIsCi(pObj) {
        Vec_PtrPush( vNodes, pObj.pNext );
        return;
    }
    if Aig_ObjIsCo(pObj) || Aig_ObjIsBuf(pObj) {
        Aig_ManSeqCleanup_rec( p, Aig_ObjFanin0(pObj), vNodes );
        return;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Aig_ManSeqCleanup_rec( p, Aig_ObjFanin0(pObj), vNodes );
    Aig_ManSeqCleanup_rec( p, Aig_ObjFanin1(pObj), vNodes );
}

/**Function*************************************************************

  Synopsis    [Returns the number of dangling nodes removed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSeqCleanup(p *Aig_Man_t) int32 {
    var vNodes, vCis, vCos *Vec_Ptr_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i, nTruePis, nTruePos int32;
//    assert( Aig_ManBufNum(p) == 0 );

    // mark the PIs
    Aig_ManIncrementTravId( p );
    Aig_ObjSetTravIdCurrent( p, Aig_ManConst1(p) );
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Aig_ObjSetTravIdCurrent( p, pObj );
    }

    // prepare to collect nodes reachable from POs
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Vec_PtrPush( vNodes, pObj );
    }

    // remember latch inputs in latch outputs
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        pObjLo.pNext = pObjLi;
    }
    // mark the nodes reachable from these nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Aig_ManSeqCleanup_rec( p, pObj, vNodes );
    }
    assert(Vec_PtrSize(vNodes) <= Aig_ManCoNum(p), "Vec_PtrSize(vNodes) <= Aig_ManCoNum(p)");
    // clean latch output pointers
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        pObjLo.pNext = nil;
    }

    // if some latches are removed, update PIs/POs
    if Vec_PtrSize(vNodes) < Aig_ManCoNum(p) {
        if p.vFlopNums {
            var nTruePos int32  = Aig_ManCoNum(p)-Aig_ManRegNum(p);
            var iNum, k int32;
            // inlining of macro Aig_ManForEachCo
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                if i >= nTruePos && Aig_ObjIsTravIdCurrent(p, pObj) {
                    iNum = Vec_IntEntry( p.vFlopNums, i - nTruePos );
                    Vec_IntWriteEntry( p.vFlopNums, k++, iNum );
                }
            }
            assert(k == Vec_PtrSize(vNodes) - nTruePos, "k == Vec_PtrSize(vNodes) - nTruePos");
            Vec_IntShrink( p.vFlopNums, k );
        }
        // collect new CIs/COs
        vCis = Vec_PtrAlloc( Aig_ManCiNum(p) );
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            if Aig_ObjIsTravIdCurrent(p, pObj) {
                Vec_PtrPush( vCis, pObj );
            } else {
                Vec_PtrWriteEntry( p.vObjs, pObj.Id, nil );
//                Aig_ManRecycleMemory( p, pObj );
            }
        }
        vCos = Vec_PtrAlloc( Aig_ManCoNum(p) );
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            if Aig_ObjIsTravIdCurrent(p, pObj) {
                Vec_PtrPush( vCos, pObj );
            } else {
                Aig_ObjDisconnect( p, pObj );
                Vec_PtrWriteEntry( p.vObjs, pObj.Id, nil );
//                Aig_ManRecycleMemory( p, pObj );
            }
        }
        // remember the number of true PIs/POs
        nTruePis = Aig_ManCiNum(p) - Aig_ManRegNum(p);
        nTruePos = Aig_ManCoNum(p) - Aig_ManRegNum(p);
        // set the new number of registers
        p.nRegs -= Aig_ManCoNum(p) - Vec_PtrSize(vNodes);
        // create new PIs/POs
        assert(Vec_PtrSize(vCis) == nTruePis + p.nRegs, "Vec_PtrSize(vCis) == nTruePis + p.nRegs");
        assert(Vec_PtrSize(vCos) == nTruePos + p.nRegs, "Vec_PtrSize(vCos) == nTruePos + p.nRegs");
        Vec_PtrFree( p.vCis );    p.vCis = vCis;
        Vec_PtrFree( p.vCos );    p.vCos = vCos;
        p.nObjs[AIG_OBJ_CI] = Vec_PtrSize( p.vCis );
        p.nObjs[AIG_OBJ_CO] = Vec_PtrSize( p.vCos );
                
    }
    Vec_PtrFree( vNodes );
    p.nTruePis = Aig_ManCiNum(p) - Aig_ManRegNum(p); 
    p.nTruePos = Aig_ManCoNum(p) - Aig_ManRegNum(p); 
    Aig_ManSetCioIds( p );
    // remove dangling nodes
    return Aig_ManCleanup( p );
}

/**Function*************************************************************

  Synopsis    [Returns the number of dangling nodes removed.]

  Description [This cleanup procedure is different in that 
  it removes logic but does not remove the dangling latches.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSeqCleanupBasic(p *Aig_Man_t) int32 {
    var vNodes *Vec_Ptr_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i int32;
//    assert( Aig_ManBufNum(p) == 0 );

    // mark the PIs
    Aig_ManIncrementTravId( p );
    Aig_ObjSetTravIdCurrent( p, Aig_ManConst1(p) );
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Aig_ObjSetTravIdCurrent( p, pObj );
    }

    // prepare to collect nodes reachable from POs
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Vec_PtrPush( vNodes, pObj );
    }

    // remember latch inputs in latch outputs
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        pObjLo.pNext = pObjLi;
    }
    // mark the nodes reachable from these nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Aig_ManSeqCleanup_rec( p, pObj, vNodes );
    }
    assert(Vec_PtrSize(vNodes) <= Aig_ManCoNum(p), "Vec_PtrSize(vNodes) <= Aig_ManCoNum(p)");
    // clean latch output pointers
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        pObjLo.pNext = nil;
    }

    // if some latches are removed, update PIs/POs
    if Vec_PtrSize(vNodes) < Aig_ManCoNum(p) {
        // add constant drivers to the dangling latches
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            if !Aig_ObjIsTravIdCurrent(p, pObj) {
                Aig_ObjPatchFanin0( p, pObj, Aig_ManConst0(p) ); }
        }
    }
    Vec_PtrFree( vNodes );
    // remove dangling nodes
    return Aig_ManCleanup( p );
}

/**Function*************************************************************

  Synopsis    [Returns the number of dangling nodes removed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCountMergeRegs(p *Aig_Man_t) int32 {
    var pObj, pFanin *Aig_Obj_t;
    var i, Counter, Const0, Const1 int32;
    Aig_ManIncrementTravId( p );
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pFanin = Aig_ObjFanin0(pObj);
        if Aig_ObjIsConst1(pFanin) {
            if Aig_ObjFaninC0(pObj) {
                Const0++;
            } else {
                Const1++; }
        }
        if Aig_ObjIsTravIdCurrent(p, pFanin) {
            continue; }
        Aig_ObjSetTravIdCurrent(p, pFanin);
        Counter++;
    }
    fmt.Printf("Regs = %d. Fanins = %d. Const0 = %d. Const1 = %d.\n", Aig_ManRegNum(p), Counter, Const0, Const1);
    return 0;
}


/**Function*************************************************************

  Synopsis    [Checks how many latches can be reduced.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManReduceLachesCount(p *Aig_Man_t) int32 {
    var pObj, pFanin *Aig_Obj_t;
    var i, Counter, Diffs int32;
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(!pObj.fMarkA && !pObj.fMarkB, "!pObj.fMarkA && !pObj.fMarkB"); }
    }
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pFanin = Aig_ObjFanin0(pObj);
        if Aig_ObjFaninC0(pObj) {
            if pFanin.fMarkB {
                Counter++;
            } else {
                pFanin.fMarkB = 1; }
        
        } else {
            if pFanin.fMarkA {
                Counter++;
            } else {
                pFanin.fMarkA = 1; }
        }
    }
    // count fanins that have both attributes
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pFanin = Aig_ObjFanin0(pObj);
        Diffs += pFanin.fMarkA && pFanin.fMarkB;
        pFanin.fMarkA = pFanin.fMarkB = 0;
    }
//    printf( "Diffs = %d.\n", Diffs );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Reduces the latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManReduceLachesOnce(p *Aig_Man_t) *Vec_Ptr_t {
    var vMap *Vec_Ptr_t;
    var pObj, pObjLi, pObjLo, pFanin *Aig_Obj_t;
    var pMapping *int32;
    var i int32;
    // start mapping by adding the true PIs
    vMap = Vec_PtrAlloc( Aig_ManCiNum(p) );
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Vec_PtrPush( vMap, pObj );
    }
    // create mapping of fanin nodes into the corresponding latch outputs
    // inlining of macro ABC_FALLOC
    pMapping = ((int *) memset(malloc(sizeof(int) * (size_t)(2 * Aig_ManObjNumMax(p))), 0xff, sizeof(int) * (size_t)(2 * Aig_ManObjNumMax(p))));
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        pFanin = Aig_ObjFanin0(pObjLi);
        if Aig_ObjFaninC0(pObjLi) {
            if pFanin.fMarkB {
                Vec_PtrPush( vMap, Aig_ManLo(p, pMapping[2*pFanin.Id + 1]) );
            
            } else {
                pFanin.fMarkB = 1;
                pMapping[2*pFanin.Id + 1] = i;
                Vec_PtrPush( vMap, pObjLo );
            }
        
        } else {
            if pFanin.fMarkA {
                Vec_PtrPush( vMap, Aig_ManLo(p, pMapping[2*pFanin.Id]) );
            
            } else {
                pFanin.fMarkA = 1;
                pMapping[2*pFanin.Id] = i;
                Vec_PtrPush( vMap, pObjLo );
            }
        }
    }
    // inlining of macro ABC_FREE
    ((pMapping) ? (free((char *) (pMapping)), (pMapping) = 0) : 0);
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pFanin = Aig_ObjFanin0(pObj);
        pFanin.fMarkA = pFanin.fMarkB = 0;
    }
    return vMap;
}

/**Function*************************************************************

  Synopsis    [Reduces the latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManReduceLaches(p *Aig_Man_t, fVerbose int32) *Aig_Man_t {
    var pTemp *Aig_Man_t;
    var vMap *Vec_Ptr_t;
    var nSaved, nCur int32;
    if fVerbose != 0 {
        fmt.Printf("Performing combinational register sweep:\n"); }
    for  nSaved = 0; (nCur = Aig_ManReduceLachesCount(p)); nSaved += nCur  {
//        printf( "Reducible = %d\n", nCur );
        vMap = Aig_ManReduceLachesOnce( p );
        p = Aig_ManRemap( pTemp = p, vMap );
        Vec_PtrFree( vMap );
        Aig_ManSeqCleanup( p );
        if fVerbose != 0 {
            Aig_ManReportImprovement( pTemp, p ); }
        Aig_ManStop( pTemp );
        if p.nRegs == 0 {
            break; }
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Computes strongly connected components of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManComputeSccs(p *Aig_Man_t) {
    var vSupports, vMatrix, vMatrix2 *Vec_Ptr_t;
    var vSupp, vSupp2, vComp *Vec_Int_t;
    var pVarsTot string;
    var i, k, m, iOut, iIn, nComps int32;
    if Aig_ManRegNum(p) == 0 {
        fmt.Printf("The network is combinational.\n");
        return;
    }
    // get structural supports for each output
    vSupports = Aig_ManSupports( p );
    // transforms the supports into the latch dependency matrix
    vMatrix = Vec_PtrStart( Aig_ManRegNum(p) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupports)) && (((vSupp) = (Vec_Int_t *)Vec_PtrEntry(vSupports, i)), 1); i++  {
        // skip true POs
        iOut = Vec_IntPop( vSupp );
        iOut -= Aig_ManCoNum(p) - Aig_ManRegNum(p);
        if iOut < 0 {
            continue; }
        // remove PIs
        m = 0;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vSupp)) && (((iIn) = Vec_IntEntry(vSupp, k)), 1); k++  {
            iIn -= Aig_ManCiNum(p) - Aig_ManRegNum(p);
            if iIn < 0 {
                continue; }
            assert(iIn < Aig_ManRegNum(p), "iIn < Aig_ManRegNum(p)");
            Vec_IntWriteEntry( vSupp, m++, iIn );
        }
        Vec_IntShrink( vSupp, m );
        // store support in the matrix
        assert(iOut < Aig_ManRegNum(p), "iOut < Aig_ManRegNum(p)");
        Vec_PtrWriteEntry( vMatrix, iOut, vSupp );
    }
    // create the reverse matrix
    vMatrix2 = Vec_PtrAlloc( Aig_ManRegNum(p) );
    for  i = 0; i < Aig_ManRegNum(p); i++  {
        Vec_PtrPush( vMatrix2, Vec_IntAlloc(8) );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMatrix)) && (((vSupp) = (Vec_Int_t *)Vec_PtrEntry(vMatrix, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vSupp)) && (((iIn) = Vec_IntEntry(vSupp, k)), 1); k++  {
            vSupp2 = (Vec_Int_t *)Vec_PtrEntry( vMatrix2, iIn );
            Vec_IntPush( vSupp2, i );
        }
    } 

    // detect strongly connected components
    vComp = Vec_IntAlloc( Aig_ManRegNum(p) );
    // inlining of macro ABC_ALLOC
    pVarsTot = ((char *) malloc(sizeof(char) * (size_t)(Aig_ManRegNum(p))));
    memset( pVarsTot, 0, Aig_ManRegNum(p) * sizeof(char) );
    for  nComps = 0; ; nComps++  {
        Vec_IntClear( vComp );
        // get the first support
        for  iOut = 0; iOut < Aig_ManRegNum(p); iOut++  {
            if pVarsTot[iOut] == 0 {
                break; }
        }
        if iOut == Aig_ManRegNum(p) {
            break; }
        pVarsTot[iOut] = 1;
        Vec_IntPush( vComp, iOut );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vComp)) && (((iOut) = Vec_IntEntry(vComp, i)), 1); i++  {
            vSupp = (Vec_Int_t *)Vec_PtrEntry( vMatrix, iOut );
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vSupp)) && (((iIn) = Vec_IntEntry(vSupp, k)), 1); k++  {
                if pVarsTot[iIn] {
                    continue; }
                pVarsTot[iIn] = 1;
                Vec_IntPush( vComp, iIn );
            }
            vSupp2 = (Vec_Int_t *)Vec_PtrEntry( vMatrix2, iOut );
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vSupp2)) && (((iIn) = Vec_IntEntry(vSupp2, k)), 1); k++  {
                if pVarsTot[iIn] {
                    continue; }
                pVarsTot[iIn] = 1;
                Vec_IntPush( vComp, iIn );
            }
        }
        if Vec_IntSize(vComp) == Aig_ManRegNum(p) {
            fmt.Printf("There is only one SCC of registers in this network.\n");
            break;
        }
        fmt.Printf("SCC #%d contains %5d registers.\n", nComps+1, Vec_IntSize(vComp));
    }
    // inlining of macro ABC_FREE
    ((pVarsTot) ? (free((char *) (pVarsTot)), (pVarsTot) = 0) : 0);
    Vec_IntFree( vComp );
    Vec_PtrFree( vMatrix );
    Vec_VecFree( (Vec_Vec_t *)vMatrix2 );
    Vec_VecFree( (Vec_Vec_t *)vSupports );
}

/**Function*************************************************************

  Synopsis    [Performs partitioned register sweep.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSclPart(pAig *Aig_Man_t, fLatchConst int32, fLatchEqual int32, fVerbose int32) *Aig_Man_t {
    var vResult *Vec_Ptr_t;
    var vPart *Vec_Int_t;
    var i, nCountPis, nCountRegs int32;
    var pMapBack *int32;
    var pTemp, pNew *Aig_Man_t;
    var nClasses int32;

    if pAig.vClockDoms {
        vResult = Vec_PtrAlloc( 100 );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize((Vec_Ptr_t *)pAig.vClockDoms)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry((Vec_Ptr_t *)pAig.vClockDoms, i)), 1); i++  {
            Vec_PtrPush( vResult, Vec_IntDup(vPart) );
        }
     
    } else {
        vResult = Aig_ManRegPartitionSimple( pAig, 0, 0 ); }

    Aig_ManReprStart( pAig, Aig_ManObjNumMax(pAig) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vResult)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vResult, i)), 1); i++  {
        pTemp = Aig_ManRegCreatePart( pAig, vPart, &nCountPis, &nCountRegs, &pMapBack );
        Aig_ManSetRegNum( pTemp, pTemp.nRegs );
        if nCountPis>0 {
            pNew = Aig_ManScl( pTemp, fLatchConst, fLatchEqual, 0, -1, -1, fVerbose, 0 );
            nClasses = Aig_TransferMappedClasses( pAig, pTemp, pMapBack );
            if fVerbose != 0 {
                fmt.Printf("%3d : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d. It = %3d. Cl = %5d\n", i, Vec_IntSize(vPart), Aig_ManCiNum(pTemp)-Vec_IntSize(vPart), nCountPis, nCountRegs, Aig_ManNodeNum(pTemp), 0, nClasses); }
            Aig_ManStop( pNew );
        }
        Aig_ManStop( pTemp );
        // inlining of macro ABC_FREE
        ((pMapBack) ? (free((char *) (pMapBack)), (pMapBack) = 0) : 0);
    }
    pNew = Aig_ManDupRepr( pAig, 0 );
    Aig_ManSeqCleanup( pNew );
    Vec_VecFree( (Vec_Vec_t*)vResult );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManScl(pAig *Aig_Man_t, fLatchConst int32, fLatchEqual int32, fUseMvSweep int32, nFramesSymb int32, nFramesSatur int32, fVerbose int32, fVeryVerbose int32) *Aig_Man_t {
    extern void Saig_ManReportUselessRegisters( Aig_Man_t * pAig );
    extern int Saig_ManReportComplements( Aig_Man_t * p );

    var pAigInit, pAigNew *Aig_Man_t;
    var pFlop1, pFlop2 *Aig_Obj_t;
    var i, Entry1, Entry2, nTruePis int32;//, nRegs;

    if pAig.vClockDoms && Vec_VecSize(pAig.vClockDoms) > 0 {
        return Aig_ManSclPart( pAig, fLatchConst, fLatchEqual, fVerbose); }

    // store the original AIG
    assert(pAig.vFlopNums == nil, "pAig.vFlopNums == nil");
    pAigInit = pAig;
    pAig = Aig_ManDupSimple( pAig );
    // create storage for latch numbers
    pAig.vFlopNums = Vec_IntStartNatural( pAig.nRegs );
    pAig.vFlopReprs = Vec_IntAlloc( 100 );
    Aig_ManSeqCleanup( pAig );
    if (fLatchConst != 0) && pAig.nRegs {
        pAig = Aig_ManConstReduce( pAig, fUseMvSweep, nFramesSymb, nFramesSatur, fVerbose, fVeryVerbose ); }
    if (fLatchEqual != 0) && pAig.nRegs {
        pAig = Aig_ManReduceLaches( pAig, fVerbose ); }
    // translate pairs into reprs
    nTruePis = Aig_ManCiNum(pAigInit)-Aig_ManRegNum(pAigInit);
    Aig_ManReprStart( pAigInit, Aig_ManObjNumMax(pAigInit) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(pAig.vFlopReprs)) && (((Entry1) = Vec_IntEntry(pAig.vFlopReprs, i)), 1); i++  {
        Entry2 = Vec_IntEntry( pAig.vFlopReprs, ++i ); 
        pFlop1 = Aig_ManCi( pAigInit, nTruePis + Entry1 );
        if Entry2 == -1 {
            pFlop2 = Aig_ManConst1(pAigInit);
        } else {
            pFlop2 = Aig_ManCi( pAigInit, nTruePis + Entry2 );
        }
        assert(pFlop1 != pFlop2, "pFlop1 != pFlop2");
        if pFlop1.Id > pFlop2.Id {
            pAigInit.pReprs[pFlop1.Id] = pFlop2;
        } else {
            pAigInit.pReprs[pFlop2.Id] = pFlop1; }
    }
    Aig_ManStop( pAig );
//    Aig_ManSeqCleanup( pAigInit );
    pAigNew = Aig_ManDupRepr( pAigInit, 0 );
    Aig_ManSeqCleanup( pAigNew );

//    Saig_ManReportUselessRegisters( pAigNew );
    if Aig_ManRegNum(pAigNew) == 0 {
        return pAigNew; }
//    nRegs = Saig_ManReportComplements( pAigNew );
//    if ( nRegs ) 
//    printf( "The number of complemented registers = %d.\n", nRegs );
    return pAigNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

