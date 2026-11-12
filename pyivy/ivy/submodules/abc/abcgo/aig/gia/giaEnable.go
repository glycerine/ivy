package gia

import "fmt"

/**CFile****************************************************************

  FileName    [gia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Structural detection of enables, sets and resets.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_CollectSuper_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t) {
    // if the new node is complemented or a PI, another gate begins
    if Gia_IsComplement(pObj) || Gia_ObjIsCi(pObj) {
        Vec_IntPushUnique( vSuper, Gia_ObjId(p, Gia_Regular(pObj)) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    // go through the branches
    Gia_CollectSuper_rec( p, Gia_ObjChild0(pObj), vSuper );
    Gia_CollectSuper_rec( p, Gia_ObjChild1(pObj), vSuper );
}

/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_CollectSuper(p *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    Vec_IntClear( vSuper );
//    Gia_CollectSuper_rec( p, pObj, vSuper );
    if Gia_ObjIsAnd(pObj) {
        Vec_IntPushUnique( vSuper, Gia_ObjId(p, Gia_ObjFanin0(pObj)) );
        Vec_IntPushUnique( vSuper, Gia_ObjId(p, Gia_ObjFanin1(pObj)) );
    
    } else {
        Vec_IntPushUnique( vSuper, Gia_ObjId(p, Gia_Regular(pObj)) ); }

}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintSignals(p *Gia_Man_t, pFreq *int32, pStr string) {
    var vObjs *Vec_Int_t;
    var i, Counter, nTotal int32;
    vObjs = Vec_IntAlloc( 100 );
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        if pFreq[i] > 1 {
            nTotal += pFreq[i];
            Counter++;
        }
    }
    fmt.Printf("%s (total = %d  driven = %d)\n", pStr, Counter, nTotal);
    Counter = 0;
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        if pFreq[i] > 10 {
            fmt.Printf("%3d :   Obj = %6d   Refs = %6d   Freq = %6d\n", ++Counter, i, Gia_ObjRefNum(p, Gia_ManObj(p,i)), pFreq[i]);
            Vec_IntPush( vObjs, i );
        }
    }
    Vec_IntFree( vObjs );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDetectSeqSignals(p *Gia_Man_t, fSetReset int32, fVerbose int32) {
    var vSuper *Vec_Int_t;
    var pFlop, pObjC, pObj0, pObj1, pNode, pTemp *Gia_Obj_t;
    var i int32;
    var k int32;
    var Ent int32;
    var pSets *int32;
    var pResets *int32;
    var pEnables *int32;
    var nHaveSetReset, nHaveEnable int32;
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
    // inlining of macro ABC_CALLOC
    pSets    = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    // inlining of macro ABC_CALLOC
    pResets  = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    // inlining of macro ABC_CALLOC
    pEnables = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    vSuper   = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pFlop) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pNode = Gia_ObjFanin0(pFlop);
        if !Gia_ObjIsAnd(pNode) {
            continue; }
        // detect sets/resets
        Gia_CollectSuper( p, pNode, vSuper );
        if Gia_ObjFaninC0(pFlop) {
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vSuper)) && (((Ent) = Vec_IntEntry(vSuper, k)), 1); k++  {
                pSets[Ent]++;
            }
        } else {
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vSuper)) && (((Ent) = Vec_IntEntry(vSuper, k)), 1); k++  {
                pResets[Ent]++;
            } }
        // detect enables
        if !Gia_ObjIsMuxType(pNode) {
            continue; }
        pObjC = Gia_ObjRecognizeMux( pNode, &pObj0, &pObj1 );
        pTemp = Gia_ObjRiToRo( p, pFlop );
        if Gia_Regular(pObj0) != pTemp && Gia_Regular(pObj1) != pTemp {
            continue; }
        if !Gia_ObjFaninC0(pFlop) {
            pObj0 = Gia_Not(pObj0);
            pObj1 = Gia_Not(pObj1);
        }
        if Gia_IsComplement(pObjC) {
            pObjC = Gia_Not(pObjC);
            pTemp = pObj0;
            pObj0 = pObj1;
            pObj1 = pTemp;
        }
        // detect controls
//        Gia_CollectSuper( p, pObjC, vSuper );
//        Vec_IntForEachEntry( vSuper, Ent, k )
//            pEnables[Ent]++;
        pEnables[Gia_ObjId(p, pObjC)]++;
        nHaveEnable++;
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pFlop) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pNode = Gia_ObjFanin0(pFlop);
        if !Gia_ObjIsAnd(pNode) {
            continue; }
        // detect sets/resets
        Gia_CollectSuper( p, pNode, vSuper );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vSuper)) && (((Ent) = Vec_IntEntry(vSuper, k)), 1); k++  {
            if pSets[Ent] > 1 || pResets[Ent] > 1 {
                nHaveSetReset++;
                break;
            }
        }
    }
    Vec_IntFree( vSuper );
    // inlining of macro ABC_FREE
    ((p.pRefs) ? (free((char *) (p.pRefs)), (p.pRefs) = 0) : 0);
    Gia_ManCreateRefs( p );
    if fVerbose != 0 {
        fmt.Printf("Flops with set/reset = %6d. Flops with enable = %6d.\n", nHaveSetReset, nHaveEnable);
        if fSetReset != 0 {
            Gia_ManPrintSignals( p, pSets, "Set signals" );
            Gia_ManPrintSignals( p, pResets, "Reset signals" );
        }
        Gia_ManPrintSignals( p, pEnables, "Enable signals" );
    }
    // inlining of macro ABC_FREE
    ((p.pRefs) ? (free((char *) (p.pRefs)), (p.pRefs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pSets) ? (free((char *) (pSets)), (pSets) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pResets) ? (free((char *) (pResets)), (pResets) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pEnables) ? (free((char *) (pEnables)), (pEnables) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDetectSeqSignalsWithFanout(p *Gia_Man_t, nFanMax int32, fVerbose int32) *Vec_Int_t {
    var vResult *Vec_Int_t;
    var vSuper *Vec_Int_t;
    var pFlop, pObjC, pObj0, pObj1, pNode, pTemp *Gia_Obj_t;
    var i int32;
    var k int32;
    var Ent int32;
    var pSets *int32;
    var pResets *int32;
    var pEnables *int32;
    var nHaveSetReset, nHaveEnable int32;
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
    // inlining of macro ABC_CALLOC
    pSets    = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    // inlining of macro ABC_CALLOC
    pResets  = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    // inlining of macro ABC_CALLOC
    pEnables = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    vSuper   = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pFlop) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pNode = Gia_ObjFanin0(pFlop);
        if !Gia_ObjIsAnd(pNode) {
            continue; }
        // detect sets/resets
        Gia_CollectSuper( p, pNode, vSuper );
        if Gia_ObjFaninC0(pFlop) {
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vSuper)) && (((Ent) = Vec_IntEntry(vSuper, k)), 1); k++  {
                pSets[Ent]++;
            }
        } else {
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vSuper)) && (((Ent) = Vec_IntEntry(vSuper, k)), 1); k++  {
                pResets[Ent]++;
            } }
        // detect enables
        if !Gia_ObjIsMuxType(pNode) {
            continue; }
        pObjC = Gia_ObjRecognizeMux( pNode, &pObj0, &pObj1 );
        pTemp = Gia_ObjRiToRo( p, pFlop );
        if Gia_Regular(pObj0) != pTemp && Gia_Regular(pObj1) != pTemp {
            continue; }
        if !Gia_ObjFaninC0(pFlop) {
            pObj0 = Gia_Not(pObj0);
            pObj1 = Gia_Not(pObj1);
        }
        if Gia_IsComplement(pObjC) {
            pObjC = Gia_Not(pObjC);
            pTemp = pObj0;
            pObj0 = pObj1;
            pObj1 = pTemp;
        }
        // detect controls
//        Gia_CollectSuper( p, pObjC, vSuper );
//        Vec_IntForEachEntry( vSuper, Ent, k )
//            pEnables[Ent]++;
        pEnables[Gia_ObjId(p, pObjC)]++;
        nHaveEnable++;
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pFlop) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pNode = Gia_ObjFanin0(pFlop);
        if !Gia_ObjIsAnd(pNode) {
            continue; }
        // detect sets/resets
        Gia_CollectSuper( p, pNode, vSuper );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vSuper)) && (((Ent) = Vec_IntEntry(vSuper, k)), 1); k++  {
            if pSets[Ent] > 1 || pResets[Ent] > 1 {
                nHaveSetReset++;
                break;
            }
        }
    }
    Vec_IntFree( vSuper );
    vResult  = Vec_IntAlloc( 100 );
    for  i = 1; i < Gia_ManObjNum(p); i++  {
        if pSets[i] > nFanMax {
            if fVerbose != 0 {
                fmt.Printf("Adding set signal %d related to %d flops.\n", i, pSets[i]); }
            Vec_IntPushUnique( vResult, i );
        }
    }
    for  i = 1; i < Gia_ManObjNum(p); i++  {
        if pResets[i] > nFanMax {
            if fVerbose != 0 {
                fmt.Printf("Adding reset signal %d related to %d flops.\n", i, pResets[i]); }
            Vec_IntPushUnique( vResult, i );
        }
    }
    for  i = 1; i < Gia_ManObjNum(p); i++  {
        if pEnables[i] > nFanMax {
            if fVerbose != 0 {
                fmt.Printf("Adding enable signal %d related to %d flops.\n", i, pEnables[i]); }
            Vec_IntPushUnique( vResult, i );
        }
    }
    // inlining of macro ABC_FREE
    ((pSets) ? (free((char *) (pSets)), (pSets) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pResets) ? (free((char *) (pResets)), (pResets) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pEnables) ? (free((char *) (pEnables)), (pEnables) = 0) : 0);
    return vResult;
}


/**Function*************************************************************

  Synopsis    [Transfers attributes from the original one to the final one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTransferFrames(pAig *Gia_Man_t, pFrames *Gia_Man_t, nFrames int32, pNew *Gia_Man_t, vSigs *Vec_Int_t) *Vec_Int_t {
    var vSigsNew *Vec_Int_t;
    var pObj, pObjF *Gia_Obj_t;
    var k, f int32;
    vSigsNew = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachObjVec
    for  k = 0; (k < Vec_IntSize(vSigs)) && ((pObj) = Gia_ManObj(pAig, Vec_IntEntry(vSigs,k))); k++  {
        assert(Gia_ObjIsCand(pObj), "Gia_ObjIsCand(pObj)");
        for  f = 0; f < nFrames; f++  {
            pObjF = Gia_ManObj( pFrames, Abc_Lit2Var(Gia_ObjCopyF( pAig, f, pObj )) );
            if pObjF.Value && ~pObjF.Value {
                Vec_IntPushUnique( vSigsNew, Abc_Lit2Var(pObjF.Value) ); }
        }
    }
    return vSigsNew;    
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnrollInit(p *Gia_Man_t, nFrames int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var f, i int32;
    Vec_IntFill( &p.vCopies, nFrames * Gia_ManObjNum(p), -1 );
    pNew = Gia_ManStart( nFrames * Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ObjSetCopyF( p, 0, pObj, 0 );
    }
    for  f = 0; f < nFrames; f++  {
        Gia_ObjSetCopyF( p, f, Gia_ManConst0(p), 0 );
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            Gia_ObjSetCopyF( p, f, pObj, Gia_ManAppendCi(pNew) );
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            Gia_ObjSetCopyF( p, f, pObj, Gia_ManHashAnd(pNew, Gia_ObjFanin0CopyF(p, f, pObj), Gia_ObjFanin1CopyF(p, f, pObj)) ); }
        }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            Gia_ObjSetCopyF( p, f, pObj, Gia_ObjFanin0CopyF(p, f, pObj) );
        }
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            Gia_ManAppendCo( pNew, Gia_ObjCopyF(p, f, pObj) );
        }
        if f == nFrames - 1 {
            break; }
        // inlining of macro Gia_ManForEachRiRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            Gia_ObjSetCopyF( p, f+1, pObjRo, Gia_ObjCopyF(p, f, pObjRi) );
        }
    }
    Gia_ManHashStop( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Unrolls initialized timeframes while cofactoring some vars.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnrollAndCofactor(p *Gia_Man_t, nFrames int32, nFanMax int32, fVerbose int32) *Gia_Man_t {
    var vCofSigs, vTemp *Vec_Int_t;
    var pAig, pFrames, pNew *Gia_Man_t;
    // compute initialized timeframes
    pFrames  = Gia_ManUnrollInit( p, nFrames );
    pAig     = Gia_ManCleanup( pFrames );
    // compute and remap set/reset/enable signals
    vCofSigs = Gia_ManDetectSeqSignalsWithFanout( p, nFanMax, fVerbose );
    vCofSigs = Gia_ManTransferFrames( p, pFrames, nFrames, pAig, vTemp = vCofSigs );
    Vec_IntFree( vTemp );
    Gia_ManStop( pFrames );
    Vec_IntErase( &p.vCopies );
    // cofactor all these variables
    pNew = Gia_ManDupCofAllInt( pAig, vCofSigs, fVerbose );
    Vec_IntFree( vCofSigs );
    Gia_ManStop( pAig );
    return pNew;
}



/**Function*************************************************************

  Synopsis    [Transform seq circuits with enables by removing enables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRemoveEnables2(p *Gia_Man_t) *Gia_Man_t {
    var pNew, pAux *Gia_Man_t; 
    var pTemp, pObjC, pObj0, pObj1, pFlopIn, pFlopOut *Gia_Obj_t;
    var pThis, pNode *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pThis) = Gia_ManCi(p, i)); i++  {
        pThis.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pThis) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pThis) {} else {
        pThis.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pThis), Gia_ObjFanin1Copy(pThis) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pThis) = Gia_ManCo(p, i)); i++  {
        pThis.Value = Gia_ObjFanin0Copy(pThis);
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pFlopIn) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pNode = Gia_ObjFanin0(pFlopIn);
        if !Gia_ObjIsMuxType(pNode) {
            fmt.Printf("Cannot recognize enable of flop %d.\n", i);
            continue;
        }
        pObjC = Gia_ObjRecognizeMux( pNode, &pObj1, &pObj0 );
        pFlopOut = Gia_ObjRiToRo( p, pFlopIn );
        if Gia_Regular(pObj0) != pFlopOut && Gia_Regular(pObj1) != pFlopOut {
            fmt.Printf("Cannot recognize self-loop of enable flop %d.\n", i);
            continue;
        }
        if !Gia_ObjFaninC0(pFlopIn) {
            pObj0 = Gia_Not(pObj0);
            pObj1 = Gia_Not(pObj1);
        }
        if Gia_IsComplement(pObjC) {
            pObjC = Gia_Not(pObjC);
            pTemp = pObj0;
            pObj0 = pObj1;
            pObj1 = pTemp;
        }
        if Gia_Regular(pObj0) == pFlopOut {
//            printf( "FlopIn compl = %d. FlopOut is d0. Complement = %d.\n", 
//                Gia_ObjFaninC0(pFlopIn), Gia_IsComplement(pObj0) );
            pFlopIn.Value = Abc_LitNotCond(Gia_Regular(pObj1).Value, !Gia_IsComplement(pObj1));
        
        } else if Gia_Regular(pObj1) == pFlopOut {
//            printf( "FlopIn compl = %d. FlopOut is d1. Complement = %d.\n", 
//                Gia_ObjFaninC0(pFlopIn), Gia_IsComplement(pObj1) );
            pFlopIn.Value = Abc_LitNotCond(Gia_Regular(pObj0).Value, !Gia_IsComplement(pObj0));
        }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pThis) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, pThis.Value );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pAux = pNew );
    Gia_ManStop( pAux );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Transform seq circuits with enables by removing enables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRemoveEnables(p *Gia_Man_t) *Gia_Man_t {
    var vCtrls, vDatas *Vec_Ptr_t;
    var vFlopClasses *Vec_Int_t;
    var pNew, pAux *Gia_Man_t; 
    var pFlopIn *Gia_Obj_t;
    var pFlopOut *Gia_Obj_t;
    var pDriver *Gia_Obj_t;
    var pFan0 *Gia_Obj_t;
    var pFan1 *Gia_Obj_t;
    var pCtrl *Gia_Obj_t = nil;
    var pData *Gia_Obj_t;
    var pObj *Gia_Obj_t;
    var i, iClass, fCompl, Counter int32;
    vCtrls = Vec_PtrAlloc( 100 );
    Vec_PtrPush( vCtrls, nil );
    vDatas = Vec_PtrAlloc( Gia_ManRegNum(p) );
    vFlopClasses = Vec_IntAlloc( Gia_ManRegNum(p) );
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pFlopIn) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        fCompl = Gia_ObjFaninC0(pFlopIn);
        pDriver = Gia_ObjFanin0(pFlopIn);
        if !Gia_ObjIsAnd(pDriver) {
            fmt.Printf("The flop driver %d is not a node.\n", i);
            Vec_PtrPush( vDatas, nil );
            Vec_IntPush( vFlopClasses, 0 );
            Counter++;
            continue;
        }
        if !Gia_ObjFaninC0(pDriver) || !Gia_ObjFaninC1(pDriver) {
            fmt.Printf("The flop driver %d is not an OR gate.\n", i);
            Vec_PtrPush( vDatas, nil );
            Vec_IntPush( vFlopClasses, 0 );
            Counter++;
            continue;
        }
        pFan0 = Gia_ObjFanin0(pDriver);
        pFan1 = Gia_ObjFanin1(pDriver);
        if !Gia_ObjIsAnd(pFan0) || !Gia_ObjIsAnd(pFan1) {
            fmt.Printf("The flop driver fanin %d is not a node.\n", i);
            Vec_PtrPush( vDatas, nil );
            Vec_IntPush( vFlopClasses, 0 );
            Counter++;
            continue;
        }
        pFlopOut = Gia_ObjRiToRo( p, pFlopIn );
        pFlopOut = Gia_NotCond( pFlopOut, !fCompl );
        if Gia_ObjChild0(pFan0) != pFlopOut && Gia_ObjChild1(pFan0) != pFlopOut && 
             Gia_ObjChild0(pFan1) != pFlopOut && Gia_ObjChild1(pFan1) != pFlopOut {
            fmt.Printf("The flop %d does not have a self-loop.\n", i);
            Vec_PtrPush( vDatas, nil );
            Vec_IntPush( vFlopClasses, 0 );
            Counter++;
            continue;
        }
        pData = nil;
        if Gia_ObjChild0(pFan0) == pFlopOut {
            pCtrl = Gia_Not( Gia_ObjChild1(pFan0) );
            if Gia_ObjFanin0(pFan1) == Gia_Regular(pCtrl) {
                pData = Gia_ObjChild1(pFan1);
            } else {
                pData = Gia_ObjChild0(pFan1); }
        
        } else if Gia_ObjChild1(pFan0) == pFlopOut {
            pCtrl = Gia_Not( Gia_ObjChild0(pFan0) );
            if Gia_ObjFanin0(pFan1) == Gia_Regular(pCtrl) {
                pData = Gia_ObjChild1(pFan1);
            } else {
                pData = Gia_ObjChild0(pFan1); }
        
        } else if Gia_ObjChild0(pFan1) == pFlopOut {
            pCtrl = Gia_Not( Gia_ObjChild1(pFan1) );
            if Gia_ObjFanin0(pFan0) == Gia_Regular(pCtrl) {
                pData = Gia_ObjChild1(pFan0);
            } else {
                pData = Gia_ObjChild0(pFan0); }
        
        } else if Gia_ObjChild1(pFan1) == pFlopOut {
            pCtrl = Gia_Not( Gia_ObjChild0(pFan1) );
            if Gia_ObjFanin0(pFan0) == Gia_Regular(pCtrl) {
                pData = Gia_ObjChild1(pFan0);
            } else {
                pData = Gia_ObjChild0(pFan0); }
        
        } else { assert(false, "0"); }
        if Vec_PtrFind( vCtrls, pCtrl ) == -1 {
            Vec_PtrPush( vCtrls, pCtrl ); }
        iClass = Vec_PtrFind( vCtrls, pCtrl );
        pData = Gia_NotCond( pData, !fCompl );
        Vec_PtrPush( vDatas, pData );
        Vec_IntPush( vFlopClasses, iClass );
    }
    assert(Vec_PtrSize( vDatas ) == Gia_ManRegNum(p), "Vec_PtrSize( vDatas ) == Gia_ManRegNum(p)");
    assert(Vec_IntSize( vFlopClasses ) == Gia_ManRegNum(p), "Vec_IntSize( vFlopClasses ) == Gia_ManRegNum(p)");
    fmt.Printf("Detected %d classes.\n", Vec_PtrSize(vCtrls) - (Counter == 0));
    Vec_PtrFree( vCtrls );
    

    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsPo(p, pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pData = (Gia_Obj_t *)Vec_PtrEntry(vDatas, i);
        if pData == nil {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else {
            pObj.Value = Gia_ManAppendCo( pNew, Abc_LitNotCond(Gia_Regular(pData).Value, Gia_IsComplement(pData)) ); }
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    Vec_PtrFree( vDatas );


    pNew = Gia_ManCleanup( pAux = pNew );
    Gia_ManStop( pAux );
    pNew.vFlopClasses = vFlopClasses;
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

