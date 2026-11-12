package cec

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [cecCorr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Latch/signal correspondence computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecCorr.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Gia_ManCorrSpecReduce_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, int f, int nPrefix );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes the real value of the literal w/o spec reduction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCorrSpecReal(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, f int32, nPrefix int32) int32 {
    if Gia_ObjIsAnd(pObj) {
        Gia_ManCorrSpecReduce_rec( pNew, p, Gia_ObjFanin0(pObj), f, nPrefix );
        Gia_ManCorrSpecReduce_rec( pNew, p, Gia_ObjFanin1(pObj), f, nPrefix );
        return Gia_ManHashAnd( pNew, Gia_ObjFanin0CopyF(p, f, pObj), Gia_ObjFanin1CopyF(p, f, pObj) );
    }
    if f == 0 {
        assert(Gia_ObjIsRo(p, pObj), "Gia_ObjIsRo(p, pObj)");
        return Gia_ObjCopyF(p, f, pObj);
    }
    assert((f != 0) && Gia_ObjIsRo(p, pObj), "f && Gia_ObjIsRo(p, pObj)");
    pObj = Gia_ObjRoToRi( p, pObj );
    Gia_ManCorrSpecReduce_rec( pNew, p, Gia_ObjFanin0(pObj), f-1, nPrefix );
    return Gia_ObjFanin0CopyF( p, f-1, pObj );
}

/**Function*************************************************************

  Synopsis    [Recursively performs speculative reduction for the object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCorrSpecReduce_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, f int32, nPrefix int32) {
    var pRepr *Gia_Obj_t;
    var iLitNew int32;
    if ~Gia_ObjCopyF(p, f, pObj) {
        return; }
    if f >= nPrefix && (pRepr = Gia_ObjReprObj(p, Gia_ObjId(p, pObj))) {
        Gia_ManCorrSpecReduce_rec( pNew, p, pRepr, f, nPrefix );
        iLitNew = Abc_LitNotCond( Gia_ObjCopyF(p, f, pRepr), Gia_ObjPhase(pRepr) ^ Gia_ObjPhase(pObj) );
        Gia_ObjSetCopyF( p, f, pObj, iLitNew );
        return;
    }
    assert(Gia_ObjIsCand(pObj), "Gia_ObjIsCand(pObj)");
    iLitNew = Gia_ManCorrSpecReal( pNew, p, pObj, f, nPrefix );
    Gia_ObjSetCopyF( p, f, pObj, iLitNew );
}

/**Function*************************************************************

  Synopsis    [Derives SRM for signal correspondence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCorrSpecReduce(p *Gia_Man_t, nFrames int32, fScorr int32, pvOutputs **Vec_Int_t, fRings int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pRepr *Gia_Obj_t;
    var vXorLits *Vec_Int_t;
    var f, i, iPrev, iObj, iPrevNew, iObjNew int32;
    assert(nFrames > 0, "nFrames > 0");
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
    assert(p.pReprs != nil, "p.pReprs != nil");
    Vec_IntFill( &p.vCopies, (nFrames+fScorr)*Gia_ManObjNum(p), -1 );
    Gia_ManSetPhase( p );
    pNew = Gia_ManStart( nFrames * Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ObjSetCopyF( p, 0, Gia_ManConst0(p), 0 );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ObjSetCopyF( p, 0, pObj, Gia_ManAppendCi(pNew) );
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if pRepr = Gia_ObjReprObj(p, Gia_ObjId(p, pObj)) {
            Gia_ObjSetCopyF( p, 0, pObj, Gia_ObjCopyF(p, 0, pRepr) ); }
    }
    for  f = 0; f < nFrames+fScorr; f++  { 
        Gia_ObjSetCopyF( p, f, Gia_ManConst0(p), 0 );
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            Gia_ObjSetCopyF( p, f, pObj, Gia_ManAppendCi(pNew) );
        }
    }
    *pvOutputs = Vec_IntAlloc( 1000 );
    vXorLits = Vec_IntAlloc( 1000 );
    if fRings != 0 {
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if Gia_ObjIsConst( p, i ) {
                iObjNew = Gia_ManCorrSpecReal( pNew, p, pObj, nFrames, 0 );
                iObjNew = Abc_LitNotCond( iObjNew, Gia_ObjPhase(pObj) );
                if iObjNew != 0 {
                    Vec_IntPush( *pvOutputs, 0 );
                    Vec_IntPush( *pvOutputs, i );
                    Vec_IntPush( vXorLits, iObjNew );
                }
            
            } else if Gia_ObjIsHead( p, i ) {
                iPrev = i;
                // inlining of macro Gia_ClassForEachObj1
                for  assert(Gia_ObjIsHead(p, i), "Gia_ObjIsHead(p, i)"), iObj = Gia_ObjNext(p, i); iObj > 0; iObj = Gia_ObjNext(p, iObj)  {
                    iPrevNew = Gia_ManCorrSpecReal( pNew, p, Gia_ManObj(p, iPrev), nFrames, 0 );
                    iObjNew  = Gia_ManCorrSpecReal( pNew, p, Gia_ManObj(p, iObj), nFrames, 0 );
                    iPrevNew = Abc_LitNotCond( iPrevNew, Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iPrev)) );
                    iObjNew  = Abc_LitNotCond( iObjNew,  Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iObj)) );
                    if iPrevNew != iObjNew && iPrevNew != 0 && iObjNew != 1 {
                        Vec_IntPush( *pvOutputs, iPrev );
                        Vec_IntPush( *pvOutputs, iObj );
                        Vec_IntPush( vXorLits, Gia_ManHashAnd(pNew, iPrevNew, Abc_LitNot(iObjNew)) );
                    }
                    iPrev = iObj;
                }
                iObj = i;
                iPrevNew = Gia_ManCorrSpecReal( pNew, p, Gia_ManObj(p, iPrev), nFrames, 0 );
                iObjNew  = Gia_ManCorrSpecReal( pNew, p, Gia_ManObj(p, iObj), nFrames, 0 );
                iPrevNew = Abc_LitNotCond( iPrevNew, Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iPrev)) );
                iObjNew  = Abc_LitNotCond( iObjNew,  Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iObj)) );
                if iPrevNew != iObjNew && iPrevNew != 0 && iObjNew != 1 {
                    Vec_IntPush( *pvOutputs, iPrev );
                    Vec_IntPush( *pvOutputs, iObj );
                    Vec_IntPush( vXorLits, Gia_ManHashAnd(pNew, iPrevNew, Abc_LitNot(iObjNew)) );
                }
            }
        }
    
    } else {
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            pRepr = Gia_ObjReprObj( p, Gia_ObjId(p,pObj) );
            if pRepr == nil {
                continue; }
            if Gia_ObjIsConst(p, i) {
                iPrevNew = 0;
            } else {
                iPrevNew = Gia_ManCorrSpecReal( pNew, p, pRepr, nFrames, 0 );
            }
            iObjNew  = Gia_ManCorrSpecReal( pNew, p, pObj, nFrames, 0 );
            iObjNew  = Abc_LitNotCond( iObjNew, Gia_ObjPhase(pRepr) ^ Gia_ObjPhase(pObj) );
            if iPrevNew != iObjNew {
                Vec_IntPush( *pvOutputs, Gia_ObjId(p, pRepr) );
                Vec_IntPush( *pvOutputs, Gia_ObjId(p, pObj) );
                Vec_IntPush( vXorLits, Gia_ManHashXor(pNew, iPrevNew, iObjNew) );
            }
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXorLits)) && (((iObjNew) = Vec_IntEntry(vXorLits, i)), 1); i++  {
        Gia_ManAppendCo( pNew, iObjNew );
    }
    Vec_IntFree( vXorLits );
    Gia_ManHashStop( pNew );
    Vec_IntErase( &p.vCopies );
//Abc_Print( 1, "Before sweeping = %d\n", Gia_ManAndNum(pNew) );
    pNew = Gia_ManCleanup( pTemp = pNew );
//Abc_Print( 1, "After sweeping = %d\n", Gia_ManAndNum(pNew) );
    Gia_ManStop( pTemp );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Derives SRM for signal correspondence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCorrSpecReduceInit(p *Gia_Man_t, nFrames int32, nPrefix int32, fScorr int32, pvOutputs **Vec_Int_t, fRings int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pRepr *Gia_Obj_t;
    var vXorLits *Vec_Int_t;
    var f, i, iPrevNew, iObjNew int32;
    assert((!(fScorr != 0) && nFrames > 1) || ((fScorr != 0) && nFrames > 0) || (nPrefix != 0), "(!fScorr && nFrames > 1) || (fScorr && nFrames > 0) || nPrefix");
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
    assert(p.pReprs != nil, "p.pReprs != nil");
    Vec_IntFill( &p.vCopies, (nFrames+nPrefix+fScorr)*Gia_ManObjNum(p), -1 );
    Gia_ManSetPhase( p );
    pNew = Gia_ManStart( (nFrames+nPrefix) * Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ManAppendCi(pNew);
        Gia_ObjSetCopyF( p, 0, pObj, 0 );
    }
    for  f = 0; f < nFrames+nPrefix+fScorr; f++  { 
        Gia_ObjSetCopyF( p, f, Gia_ManConst0(p), 0 );
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            Gia_ObjSetCopyF( p, f, pObj, Gia_ManAppendCi(pNew) );
        }
    }
    *pvOutputs = Vec_IntAlloc( 1000 );
    vXorLits = Vec_IntAlloc( 1000 );
    for  f = nPrefix; f < nFrames+nPrefix; f++  {
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            pRepr = Gia_ObjReprObj( p, Gia_ObjId(p,pObj) );
            if pRepr == nil {
                continue; }
            if Gia_ObjIsConst(p, i) {
                iPrevNew = 0;
            } else {
                iPrevNew = Gia_ManCorrSpecReal( pNew, p, pRepr, f, nPrefix );
            }
            iObjNew  = Gia_ManCorrSpecReal( pNew, p, pObj, f, nPrefix );
            iObjNew  = Abc_LitNotCond( iObjNew, Gia_ObjPhase(pRepr) ^ Gia_ObjPhase(pObj) );
            if iPrevNew != iObjNew {
                Vec_IntPush( *pvOutputs, Gia_ObjId(p, pRepr) );
                Vec_IntPush( *pvOutputs, Gia_ObjId(p, pObj) );
                Vec_IntPush( vXorLits, Gia_ManHashXor(pNew, iPrevNew, iObjNew) );
            }
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXorLits)) && (((iObjNew) = Vec_IntEntry(vXorLits, i)), 1); i++  {
        Gia_ManAppendCo( pNew, iObjNew );
    }
    Vec_IntFree( vXorLits );
    Gia_ManHashStop( pNew );
    Vec_IntErase( &p.vCopies );
//Abc_Print( 1, "Before sweeping = %d\n", Gia_ManAndNum(pNew) );
    pNew = Gia_ManCleanup( pTemp = pNew );
//Abc_Print( 1, "After sweeping = %d\n", Gia_ManAndNum(pNew) );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Initializes simulation info for lcorr/scorr counter-examples.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManStartSimInfo(vInfo *Vec_Ptr_t, nFlops int32) {
    var pInfo *uint32;
    var k, w, nWords int32;
    nWords = Vec_PtrReadWordsSimInfo( vInfo );
    assert(nFlops <= Vec_PtrSize(vInfo), "nFlops <= Vec_PtrSize(vInfo)");
    for  k = 0; k < nFlops; k++  {
        pInfo = (unsigned *)Vec_PtrEntry( vInfo, k );
        for  w = 0; w < nWords; w++  {
            pInfo[w] = 0;
        }
    }
    for  k = nFlops; k < Vec_PtrSize(vInfo); k++  {
        pInfo = (unsigned *)Vec_PtrEntry( vInfo, k );
        for  w = 0; w < nWords; w++  {
            pInfo[w] = Gia_ManRandom( 0 );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Remaps simulation info from SRM to the original AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCorrRemapSimInfo(p *Gia_Man_t, vInfo *Vec_Ptr_t) {
    var pObj, pRepr *Gia_Obj_t;
    var pInfoObj, pInfoRepr *uint32;
    var i, w, nWords int32;
    nWords = Vec_PtrReadWordsSimInfo( vInfo );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        // skip ROs without representatives
        pRepr = Gia_ObjReprObj( p, Gia_ObjId(p,pObj) );
        if pRepr == nil || Gia_ObjFailed(p, Gia_ObjId(p,pObj)) {
            continue; }
        pInfoObj = (unsigned *)Vec_PtrEntry( vInfo, i );
        for  w = 0; w < nWords; w++  {
            assert(pInfoObj[w] == 0, "pInfoObj[w] == 0");
        }
        // skip ROs with constant representatives
        if Gia_ObjIsConst0(pRepr) {
            continue; }
        assert(Gia_ObjIsRo(p, pRepr), "Gia_ObjIsRo(p, pRepr)");
//        Abc_Print( 1, "%d -> %d    ", i, Gia_ObjId(p, pRepr) );
        // transfer info from the representative
        pInfoRepr = (unsigned *)Vec_PtrEntry( vInfo, Gia_ObjCioId(pRepr) - Gia_ManPiNum(p) );
        for  w = 0; w < nWords; w++  {
            pInfoObj[w] = pInfoRepr[w];
        }
    }
//    Abc_Print( 1, "\n" );
}

/**Function*************************************************************

  Synopsis    [Collects information about remapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCorrCreateRemapping(p *Gia_Man_t) *Vec_Int_t {
    var vPairs *Vec_Int_t;
    var pObj, pRepr *Gia_Obj_t;
    var i int32;
    vPairs = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        // skip ROs without representatives
        pRepr = Gia_ObjReprObj( p, Gia_ObjId(p,pObj) );
        if pRepr == nil || Gia_ObjIsConst0(pRepr) || Gia_ObjFailed(p, Gia_ObjId(p,pObj)) {
//        if ( pRepr == NULL || Gia_ObjIsConst0(pRepr) || Gia_ObjIsFailedPair(p, Gia_ObjId(p, pRepr), Gia_ObjId(p, pObj)) )
            continue; }
        assert(Gia_ObjIsRo(p, pRepr), "Gia_ObjIsRo(p, pRepr)");
//        Abc_Print( 1, "%d -> %d    ", Gia_ObjId(p,pObj), Gia_ObjId(p, pRepr) );
        // remember the pair
        Vec_IntPush( vPairs, Gia_ObjCioId(pRepr) - Gia_ManPiNum(p) );
        Vec_IntPush( vPairs, i );
    }
    return vPairs;
}

/**Function*************************************************************

  Synopsis    [Remaps simulation info from SRM to the original AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCorrPerformRemapping(vPairs *Vec_Int_t, vInfo *Vec_Ptr_t) {
    var pInfoObj, pInfoRepr *uint32;
    var w, i, iObj, iRepr, nWords int32;
    nWords = Vec_PtrReadWordsSimInfo( vInfo );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPairs)) && (((iRepr) = Vec_IntEntry(vPairs, i)), 1); i++  {
        iObj = Vec_IntEntry( vPairs, ++i );
        pInfoObj = (unsigned *)Vec_PtrEntry( vInfo, iObj );
        pInfoRepr = (unsigned *)Vec_PtrEntry( vInfo, iRepr );
        for  w = 0; w < nWords; w++  {
            assert(pInfoObj[w] == 0, "pInfoObj[w] == 0");
            pInfoObj[w] = pInfoRepr[w];
        }
    }
}

/**Function*************************************************************

  Synopsis    [Packs one counter-examples into the array of simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

*************************************`**********************************/
func Cec_ManLoadCounterExamplesTry(vInfo *Vec_Ptr_t, vPres *Vec_Ptr_t, iBit int32, pLits *int32, nLits int32) int32 {
    var pInfo, pPres *uint32;
    var i int32;
    for  i = 0; i < nLits; i++  {
        pInfo = (unsigned *)Vec_PtrEntry(vInfo, Abc_Lit2Var(pLits[i]));
        pPres = (unsigned *)Vec_PtrEntry(vPres, Abc_Lit2Var(pLits[i]));
        if Abc_InfoHasBit( pPres, iBit ) && 
             Abc_InfoHasBit( pInfo, iBit ) == Abc_LitIsCompl(pLits[i]) {
             return 0; }
    }
    for  i = 0; i < nLits; i++  {
        pInfo = (unsigned *)Vec_PtrEntry(vInfo, Abc_Lit2Var(pLits[i]));
        pPres = (unsigned *)Vec_PtrEntry(vPres, Abc_Lit2Var(pLits[i]));
        Abc_InfoSetBit( pPres, iBit );
        if Abc_InfoHasBit( pInfo, iBit ) == Abc_LitIsCompl(pLits[i]) {
            Abc_InfoXorBit( pInfo, iBit ); }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Performs bitpacking of counter-examples.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManLoadCounterExamples(vInfo *Vec_Ptr_t, vCexStore *Vec_Int_t, iStart int32) int32 { 
    var vPat *Vec_Int_t;
    var vPres *Vec_Ptr_t;
    var nWords int32  = Vec_PtrReadWordsSimInfo(vInfo);
    var nBits int32  = 32 * nWords; 
    var k, nSize, kMax int32;//, iBit = 1;
    vPat  = Vec_IntAlloc( 100 );
    vPres = Vec_PtrAllocSimInfo( Vec_PtrSize(vInfo), nWords );
    Vec_PtrCleanSimInfo( vPres, 0, nWords );
    for  iStart < Vec_IntSize(vCexStore)  {
        // skip the output number
        iStart++;
        // get the number of items
        nSize = Vec_IntEntry( vCexStore, iStart++ );
        if nSize <= 0 {
            continue; }
        // extract pattern
        Vec_IntClear( vPat );
        for  k = 0; k < nSize; k++  {
            Vec_IntPush( vPat, Vec_IntEntry( vCexStore, iStart++ ) );
        }
        // add pattern to storage
        for  k = 1; k < nBits; k++  {
            if Cec_ManLoadCounterExamplesTry( vInfo, vPres, k, (int *)Vec_IntArray(vPat), Vec_IntSize(vPat) ) {
                break; }
        }
        kMax = Abc_MaxInt( kMax, k );
        if k == nBits-1 {
            break; }
    }
    Vec_PtrFree( vPres );
    Vec_IntFree( vPat );
    return iStart;
}

/**Function*************************************************************

  Synopsis    [Performs bitpacking of counter-examples.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManLoadCounterExamples2(vInfo *Vec_Ptr_t, vCexStore *Vec_Int_t, iStart int32) int32 { 
    var pInfo *uint32;
    var nBits int32  = 32 * Vec_PtrReadWordsSimInfo(vInfo); 
    var k int32;
    var iLit int32;
    var nLits int32;
    var Out int32;
    var iBit int32 = 1;
    for  iStart < Vec_IntSize(vCexStore)  {
        // skip the output number
//        iStart++;
        Out = Vec_IntEntry( vCexStore, iStart++ );
//        Abc_Print( 1, "iBit = %d. Out = %d.\n", iBit, Out );
        // get the number of items
        nLits = Vec_IntEntry( vCexStore, iStart++ );
        if nLits <= 0 {
            continue; }
        // add pattern to storage
        for  k = 0; k < nLits; k++  {
            iLit = Vec_IntEntry( vCexStore, iStart++ );
            pInfo = (unsigned *)Vec_PtrEntry( vInfo, Abc_Lit2Var(iLit) );
            if Abc_InfoHasBit( pInfo, iBit ) == Abc_LitIsCompl(iLit) {
                Abc_InfoXorBit( pInfo, iBit ); }
        }
        if iBit++; iBit == nBits {
            break; }
    }
//    Abc_Print( 1, "added %d bits\n", iBit-1 );
    return iStart;
}

/**Function*************************************************************

  Synopsis    [Resimulates counter-examples derived by the SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManResimulateCounterExamples(pSim *Cec_ManSim_t, vCexStore *Vec_Int_t, nFrames int32) int32 { 
    var vPairs *Vec_Int_t;
    var vSimInfo *Vec_Ptr_t; 
    var RetValue, iStart int32;
    vPairs = Gia_ManCorrCreateRemapping( pSim.pAig );
    Gia_ManCreateValueRefs( pSim.pAig );
//    pSim->pPars->nWords  = 63;
    pSim.pPars.nFrames = nFrames;
    vSimInfo = Vec_PtrAllocSimInfo( Gia_ManRegNum(pSim.pAig) + Gia_ManPiNum(pSim.pAig) * nFrames, pSim.pPars.nWords );
    for  iStart < Vec_IntSize(vCexStore)  {
        Cec_ManStartSimInfo( vSimInfo, Gia_ManRegNum(pSim.pAig) );
        iStart = Cec_ManLoadCounterExamples( vSimInfo, vCexStore, iStart );
//        iStart = Cec_ManLoadCounterExamples2( vSimInfo, vCexStore, iStart );
//        Gia_ManCorrRemapSimInfo( pSim->pAig, vSimInfo );
        Gia_ManCorrPerformRemapping( vPairs, vSimInfo );
        RetValue |= Cec_ManSeqResimulate( pSim, vSimInfo );
//        Cec_ManSeqResimulateInfo( pSim->pAig, vSimInfo, NULL );
    }
//Gia_ManEquivPrintOne( pSim->pAig, 85, 0 );
    assert(iStart == Vec_IntSize(vCexStore), "iStart == Vec_IntSize(vCexStore)");
    Vec_PtrFree( vSimInfo );
    Vec_IntFree( vPairs );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Resimulates counter-examples derived by the SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManResimulateCounterExamplesComb(pSim *Cec_ManSim_t, vCexStore *Vec_Int_t) int32 { 
    var vSimInfo *Vec_Ptr_t; 
    var RetValue, iStart int32;
    Gia_ManCreateValueRefs( pSim.pAig );
    pSim.pPars.nFrames = 1;
    vSimInfo = Vec_PtrAllocSimInfo( Gia_ManCiNum(pSim.pAig), pSim.pPars.nWords );
    for  iStart < Vec_IntSize(vCexStore)  {
        Cec_ManStartSimInfo( vSimInfo, 0 );
        iStart = Cec_ManLoadCounterExamples( vSimInfo, vCexStore, iStart );
        RetValue |= Cec_ManSeqResimulate( pSim, vSimInfo );
    }
    assert(iStart == Vec_IntSize(vCexStore), "iStart == Vec_IntSize(vCexStore)");
    Vec_PtrFree( vSimInfo );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Updates equivalence classes by marking those that timed out.]

  Description [Returns 1 if all ndoes are proved.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckRefinements(p *Gia_Man_t, vStatus *Vec_Str_t, vOutputs *Vec_Int_t, pSim *Cec_ManSim_t, fRings int32) int32 {
    var i, status, iRepr, iObj int32;
    var Counter int32;
    assert(2 * Vec_StrSize(vStatus) == Vec_IntSize(vOutputs), "2 * Vec_StrSize(vStatus) == Vec_IntSize(vOutputs)");
    // inlining of macro Vec_StrForEachEntry
    for  i = 0; (i < Vec_StrSize(vStatus)) && (((status) = Vec_StrEntry(vStatus, i)), 1); i++  {
        iRepr = Vec_IntEntry( vOutputs, 2*i );
        iObj  = Vec_IntEntry( vOutputs, 2*i+1 );
        if status == 1 {
            continue; }
        if status == 0 {
            if Gia_ObjHasSameRepr(p, iRepr, iObj) {
                Counter++; }
//            if ( Gia_ObjHasSameRepr(p, iRepr, iObj) )
//                Abc_Print( 1, "Gia_ManCheckRefinements(): Disproved equivalence (%d,%d) is not refined!\n", iRepr, iObj );
//            if ( Gia_ObjHasSameRepr(p, iRepr, iObj) )
//                Cec_ManSimClassRemoveOne( pSim, iObj );
            continue;
        }
        if status == -1 {
//            if ( !Gia_ObjFailed( p, iObj ) )
//                Abc_Print( 1, "Gia_ManCheckRefinements(): Failed equivalence is not marked as failed!\n" );
//            Gia_ObjSetFailed( p, iRepr );
//            Gia_ObjSetFailed( p, iObj );        
//            if ( fRings )
//            Cec_ManSimClassRemoveOne( pSim, iRepr );
            Cec_ManSimClassRemoveOne( pSim, iObj );
            continue;
        }
    }
//    if ( Counter )
//    Abc_Print( 1, "Gia_ManCheckRefinements(): Could not refine %d nodes.\n", Counter );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCorrReduce_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    var pRepr *Gia_Obj_t;
    if pRepr = Gia_ObjReprObj(p, Gia_ObjId(p, pObj)) {
        Gia_ManCorrReduce_rec( pNew, p, pRepr );
        pObj.Value = Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) );
        return;
    }
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManCorrReduce_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Gia_ManCorrReduce_rec( pNew, p, Gia_ObjFanin1(pObj) );
    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Reduces AIG using equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCorrReduce(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManSetPhase( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManCorrReduce_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Prints statistics during solving.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManRefinedClassPrintStats(p *Gia_Man_t, vStatus *Vec_Str_t, iIter int32, Time abctime) { 
    var nLits, CounterX, Counter0, Counter int32;
    var i, Entry, nProve, nDispr, nFail int32;
    for  i = 1; i < Gia_ManObjNum(p); i++  {
        if Gia_ObjIsNone(p, i) {
            CounterX++;
        } else if Gia_ObjIsConst(p, i) {
            Counter0++;
        } else if Gia_ObjIsHead(p, i) {
            Counter++; }
    }
    CounterX -= Gia_ManCoNum(p);
    nLits = Gia_ManCiNum(p) + Gia_ManAndNum(p) - Counter - CounterX;
    if iIter == -1 {
        Abc_Print( 1, "BMC : " );
    } else {
        Abc_Print( 1, "%3d : ", iIter ); }
    Abc_Print( 1, "c =%8d  cl =%7d  lit =%8d  ", Counter0, Counter, nLits );
    if vStatus != nil {
    // inlining of macro Vec_StrForEachEntry
    for  i = 0; (i < Vec_StrSize(vStatus)) && (((Entry) = Vec_StrEntry(vStatus, i)), 1); i++  {
        if Entry == 1 {
            nProve++;
        } else if Entry == 0 {
            nDispr++;
        } else if Entry == -1 {
            nFail++; }
    } }
    Abc_Print( 1, "p =%6d  d =%6d  f =%6d  ", nProve, nDispr, nFail );
    Abc_Print( 1, "%c  ", Gia_ObjIsConst( p, Gia_ObjFaninId0p(p, Gia_ManPo(p, 0)) ) ? '+' : '-' );
    Abc_PrintTime( 1, "T", Time );
}

/**Function*************************************************************

  Synopsis    [Runs BMC for the equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManLSCorrespondenceBmc(pAig *Gia_Man_t, pPars *Cec_ParCor_t, nPrefs int32) {  
    var ParsSim Cec_ParSim_t;
    var pParsSim *Cec_ParSim_t = &ParsSim;
    var ParsSat Cec_ParSat_t;
    var pParsSat *Cec_ParSat_t = &ParsSat;
    var vStatus *Vec_Str_t;
    var vOutputs *Vec_Int_t;
    var vCexStore *Vec_Int_t;
    var pSim *Cec_ManSim_t;
    var pSrm *Gia_Man_t;
    var fChanges, RetValue int32;
    // prepare simulation manager
    Cec_ManSimSetDefaultParams( pParsSim );
    pParsSim.nWords     = pPars.nWords;
    pParsSim.nFrames    = pPars.nRounds;
    pParsSim.fVerbose   = pPars.fVerbose;
    pParsSim.fLatchCorr = pPars.fLatchCorr;
    pParsSim.fSeqSimulate = 1;
    pSim = Cec_ManSimStart( pAig, pParsSim );
    // prepare SAT solving
    Cec_ManSatSetDefaultParams( pParsSat );
    pParsSat.nBTLimit = pPars.nBTLimit;
    pParsSat.fVerbose = pPars.fVerbose;
    fChanges = 1;
    for  fChanges != 0  {
        var clkBmc abctime  = Abc_Clock();
        fChanges = 0;
        pSrm = Gia_ManCorrSpecReduceInit( pAig, pPars.nFrames, nPrefs, !pPars.fLatchCorr, &vOutputs, pPars.fUseRings );
        if Gia_ManPoNum(pSrm) == 0 {
            Gia_ManStop( pSrm );
            Vec_IntFree( vOutputs );
            break;
        } 
        pParsSat.nBTLimit *= 10;
        if pPars.fUseCSat {
            vCexStore = Tas_ManSolveMiterNc( pSrm, pPars.nBTLimit, &vStatus, 0 );
        } else {
            vCexStore = Cec_ManSatSolveMiter( pSrm, pParsSat, &vStatus ); }
        // refine classes with these counter-examples
        if Vec_IntSize(vCexStore) {
            RetValue = Cec_ManResimulateCounterExamples( pSim, vCexStore, pPars.nFrames + 1 + nPrefs );
            Gia_ManCheckRefinements( pAig, vStatus, vOutputs, pSim, pPars.fUseRings );
            fChanges = 1;
        }
        if pPars.fVerbose {
            Cec_ManRefinedClassPrintStats( pAig, vStatus, -1, Abc_Clock() - clkBmc ); }
        // recycle
        Vec_IntFree( vCexStore );
        Vec_StrFree( vStatus );
        Gia_ManStop( pSrm );
        Vec_IntFree( vOutputs );
    }
    Cec_ManSimStop( pSim );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManLSCorrAnalyzeDependence(p *Gia_Man_t, vEquivs *Vec_Int_t, vStatus *Vec_Str_t) int32 {
    var pObj, pObjRo *Gia_Obj_t;
    var i, Iter, iObj, iRepr, fPrev, Total, Count0, Count1 int32;
    assert(Vec_StrSize(vStatus) * 2 == Vec_IntSize(vEquivs), "Vec_StrSize(vStatus) * 2 == Vec_IntSize(vEquivs)");
    Total = 0;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        assert(pObj.fMark1 == 0, "pObj.fMark1 == 0");
        if Gia_ObjHasRepr(p, i) {
            Total++; }
    }
    Count0 = 0;
    for  i = 0; i < Vec_StrSize(vStatus); i++  {
        iRepr = Vec_IntEntry(vEquivs, 2*i);
        iObj = Vec_IntEntry(vEquivs, 2*i+1);
        assert(iRepr == Gia_ObjRepr(p, iObj), "iRepr == Gia_ObjRepr(p, iObj)");
        if Vec_StrEntry(vStatus, i) != 1 { // disproved or undecided
        
            Gia_ManObj(p, iObj).fMark1 = 1;
            Count0++;
        }
    }
    for  Iter = 0; Iter < 100; Iter++  {
        var fChanges int32;
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if Gia_ObjIsCi(pObj) {
                continue; }
            assert(Gia_ObjIsAnd(pObj) || Gia_ObjIsCo(pObj), "Gia_ObjIsAnd(pObj) || Gia_ObjIsCo(pObj)");
//            fPrev = pObj->fMark1;
            if Gia_ObjIsAnd(pObj) {
                pObj.fMark1 |= Gia_ObjFanin0(pObj).fMark1 | Gia_ObjFanin1(pObj).fMark1;
            } else {
                pObj.fMark1 |= Gia_ObjFanin0(pObj).fMark1; }
//            fChanges += fPrev ^ pObj->fMark1;
        }
        // inlining of macro Gia_ManForEachRiRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            fPrev = pObjRo.fMark1;
            pObjRo.fMark1 = pObj.fMark1;
            fChanges += fPrev ^ pObjRo.fMark1;
        }
        if fChanges == 0 {
            break; }
    }
    Count1 = 0;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if pObj.fMark1 && Gia_ObjHasRepr(p, i) {
            Count1++; }
        pObj.fMark1 = 0;
    }
    fmt.Printf("%5d -> %5d (%3d)  ", Count0, Count1, Iter);
    return 0;
}

/**Function*************************************************************

  Synopsis    [Internal procedure for register correspondence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManLSCorrespondenceClasses(pAig *Gia_Man_t, pPars *Cec_ParCor_t) int32 {  
    var nIterMax int32      = 100000;
    var nAddFrames int32    = 1; // additional timeframes to simulate
    var fRunBmcFirst int32  = 1;
    var vStatus *Vec_Str_t;
    var vOutputs *Vec_Int_t;
    var vCexStore *Vec_Int_t;
    var ParsSim Cec_ParSim_t;
    var pParsSim *Cec_ParSim_t = &ParsSim;
    var ParsSat Cec_ParSat_t;
    var pParsSat *Cec_ParSat_t = &ParsSat;
    var pSim *Cec_ManSim_t;
    var pSrm *Gia_Man_t;
    var r, RetValue int32;
    var clkTotal abctime  = Abc_Clock();
    var clkSat, clkSim, clkSrm abctime;
    var clk2 abctime;
    var clk abctime = Abc_Clock();
    if Gia_ManRegNum(pAig) == 0 {
        Abc_Print( 1, "Cec_ManLatchCorrespondence(): Not a sequential AIG.\n" );
        return 0;
    }
    Gia_ManRandom( 1 );
    // prepare simulation manager
    Cec_ManSimSetDefaultParams( pParsSim );
    pParsSim.nWords     = pPars.nWords;
    pParsSim.nFrames    = pPars.nFrames;
    pParsSim.fVerbose   = pPars.fVerbose;
    pParsSim.fLatchCorr = pPars.fLatchCorr;
    pParsSim.fConstCorr = pPars.fConstCorr;
    pParsSim.fSeqSimulate = 1;
    // create equivalence classes of registers
    pSim = Cec_ManSimStart( pAig, pParsSim );
    if pAig.pReprs == nil {
        Cec_ManSimClassesPrepare( pSim, pPars.nLevelMax );
        Cec_ManSimClassesRefine( pSim );
    }
    // prepare SAT solving
    Cec_ManSatSetDefaultParams( pParsSat );
    pParsSat.nBTLimit = pPars.nBTLimit;
    pParsSat.fVerbose = pPars.fVerbose;
    // limit the number of conflicts in the circuit-based solver
    if pPars.fUseCSat {
        pParsSat.nBTLimit = Abc_MinInt( pParsSat.nBTLimit, 1000 ); }
    if pPars.fVerbose {
        Abc_Print( 1, "Obj = %7d. And = %7d. Conf = %5d. Fr = %d. Lcorr = %d. Ring = %d. CSat = %d.\n",
            Gia_ManObjNum(pAig), Gia_ManAndNum(pAig), 
            pPars.nBTLimit, pPars.nFrames, pPars.fLatchCorr, pPars.fUseRings, pPars.fUseCSat );
        Cec_ManRefinedClassPrintStats( pAig, nil, 0, Abc_Clock() - clk );
    }
    // check the base case
    if (fRunBmcFirst != 0) && (!pPars.fLatchCorr || pPars.nFrames > 1) {
        Cec_ManLSCorrespondenceBmc( pAig, pPars, 0 ); }
    if pPars.pFunc {
        ((int (*)(void *))pPars.pFunc)( pPars.pData );
        ((int (*)(void *))pPars.pFunc)( pPars.pData );
    }
    if pPars.nStepsMax == 0 {
        Abc_Print( 1, "Stopped signal correspondence after BMC.\n" );
        Cec_ManSimStop( pSim );
        return 1;
    }
    // perform refinement of equivalence classes
    for  r = 0; r < nIterMax; r++  { 
        if pPars.nStepsMax == r {
            Cec_ManSimStop( pSim );
            Abc_Print( 1, "Stopped signal correspondence after %d refiment iterations.\n", r );
            return 1;
        }
        clk = Abc_Clock();
        // perform speculative reduction
        clk2 = Abc_Clock();
        pSrm = Gia_ManCorrSpecReduce( pAig, pPars.nFrames, !pPars.fLatchCorr, &vOutputs, pPars.fUseRings );
        assert(Gia_ManRegNum(pSrm) == 0 && Gia_ManPiNum(pSrm) == Gia_ManRegNum(pAig)+(pPars.nFrames+!pPars.fLatchCorr)*Gia_ManPiNum(pAig), "Gia_ManRegNum(pSrm) == 0 && Gia_ManPiNum(pSrm) == Gia_ManRegNum(pAig)+(pPars.nFrames+!pPars.fLatchCorr)*Gia_ManPiNum(pAig)");
        clkSrm += Abc_Clock() - clk2;
        if Gia_ManCoNum(pSrm) == 0 {
            Vec_IntFree( vOutputs );
            Gia_ManStop( pSrm );            
            break;
        }
//Gia_DumpAiger( pSrm, "corrsrm", r, 2 );
        // found counter-examples to speculation
        clk2 = Abc_Clock();
        if pPars.fUseCSat {
            vCexStore = Cbs_ManSolveMiterNc( pSrm, pPars.nBTLimit, &vStatus, 0, 0 );
        } else {
            vCexStore = Cec_ManSatSolveMiter( pSrm, pParsSat, &vStatus ); }
        Gia_ManStop( pSrm );
        clkSat += Abc_Clock() - clk2;
        if Vec_IntSize(vCexStore) == 0 {
            Vec_IntFree( vCexStore );
            Vec_StrFree( vStatus );
            Vec_IntFree( vOutputs );
            break;
        }
//        Cec_ManLSCorrAnalyzeDependence( pAig, vOutputs, vStatus );        

        // refine classes with these counter-examples
        clk2 = Abc_Clock();
        RetValue = Cec_ManResimulateCounterExamples( pSim, vCexStore, pPars.nFrames + 1 + nAddFrames );
        Vec_IntFree( vCexStore );
        clkSim += Abc_Clock() - clk2;
        Gia_ManCheckRefinements( pAig, vStatus, vOutputs, pSim, pPars.fUseRings );
        if pPars.fVerbose {
            Cec_ManRefinedClassPrintStats( pAig, vStatus, r+1, Abc_Clock() - clk ); }
        Vec_StrFree( vStatus );
        Vec_IntFree( vOutputs );
//Gia_ManEquivPrintClasses( pAig, 1, 0 );
        if pPars.pFunc {
            ((int (*)(void *))pPars.pFunc)( pPars.pData ); }
        // quit if const is no longer there
        if pPars.fStopWhenGone && Gia_ManPoNum(pAig) == 1 && !Gia_ObjIsConst( pAig, Gia_ObjFaninId0p(pAig, Gia_ManPo(pAig, 0)) ) {
            fmt.Printf("Iterative refinement is stopped after iteration %d\n", r);
            fmt.Printf("because the property output is no longer a candidate constant.\n");
            Cec_ManSimStop( pSim );
            return 0;
        }
    }
    if pPars.fVerbose {
        Cec_ManRefinedClassPrintStats( pAig, nil, r+1, Abc_Clock() - clk ); }
    // check the overflow
    if r == nIterMax {
        Abc_Print( 1, "The refinement was not finished. The result may be incorrect.\n" ); }
    Cec_ManSimStop( pSim );
    // check the base case
    if !(fRunBmcFirst != 0) && (!pPars.fLatchCorr || pPars.nFrames > 1) {
        Cec_ManLSCorrespondenceBmc( pAig, pPars, 0 ); }
    clkTotal = Abc_Clock() - clkTotal;
    // report the results
    if pPars.fVerbose {
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Srm  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((clkSrm)))/(float64(int64(time.Second))), (float64((clkTotal)))? 100.0*(float64((clkSrm)))/(float64((clkTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Sat  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((clkSat)))/(float64(int64(time.Second))), (float64((clkTotal)))? 100.0*(float64((clkSat)))/(float64((clkTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Sim  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((clkSim)))/(float64(int64(time.Second))), (float64((clkTotal)))? 100.0*(float64((clkSim)))/(float64((clkTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Other"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((clkTotal-clkSat-clkSrm-clkSim)))/(float64(int64(time.Second))), (float64((clkTotal)))? 100.0*(float64((clkTotal-clkSat-clkSrm-clkSim)))/(float64((clkTotal))) : 0.0);;
        Abc_PrintTime( 1, "TOTAL",  clkTotal );
    }
    return 1;
}    

/**Function*************************************************************

  Synopsis    [Computes new initial state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManComputeInitState(pAig *Gia_Man_t, nFrames int32) *uint32 {  
    var pObj, pObjRo, pObjRi *Gia_Obj_t;
    var pInitState *uint32;
    var i, f int32; 
    Gia_ManRandom( 1 );
//    Abc_Print( 1, "Simulating %d timeframes.\n", nFrames );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+i)); i++  {
        pObj.fMark1 = 0;
    }
    for  f = 0; f < nFrames; f++  {
        Gia_ManConst0(pAig).fMark1 = 0;
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(pAig)) && ((pObj) = Gia_ManCi(pAig, i)); i++  {
            pObj.fMark1 = Gia_ManRandom(0) & 1;
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark1 = (Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj)) & 
                (Gia_ObjFanin1(pObj).fMark1 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCo(pAig, Gia_ManPoNum(pAig)+i)); i++  {
            pObj.fMark1 = (Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj));
        }
        // inlining of macro Gia_ManForEachRiRo
        for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObjRi) = Gia_ManCo(pAig, Gia_ManPoNum(pAig)+i)) && ((pObjRo) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+i)); i++  {
            pObjRo.fMark1 = pObjRi.fMark1;
        }
    }
    // inlining of macro ABC_CALLOC
    pInitState = ((unsigned *) calloc((size_t)(Abc_BitWordNum(Gia_ManRegNum(pAig))), sizeof(unsigned)));
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+i)); i++  {
        if pObj.fMark1 {
            Abc_InfoSetBit( pInitState, i ); }
//        Abc_Print( 1, "%d", pObj->fMark1 );
    }
//    Abc_Print( 1, "\n" );
    Gia_ManCleanMark1( pAig );
    return pInitState;
}

/**Function*************************************************************

  Synopsis    [Prints flop equivalences.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPrintFlopEquivs(p *Gia_Man_t) {
    var pObj, pRepr *Gia_Obj_t;
    var i int32;
    assert(p.vNamesIn != nil, "p.vNamesIn != nil");
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if Gia_ObjIsConst(p, Gia_ObjId(p, pObj)) {
            Abc_Print( 1, "Original flop %s is proved equivalent to constant.\n", Vec_PtrEntry(p.vNamesIn, Gia_ObjCioId(pObj)) );
        } else if pRepr = Gia_ObjReprObj(p, Gia_ObjId(p, pObj)) {
            if Gia_ObjIsCi(pRepr) {
                Abc_Print( 1, "Original flop %s is proved equivalent to flop %s.\n",
                    Vec_PtrEntry( p.vNamesIn, Gia_ObjCioId(pObj)  ),
                    Vec_PtrEntry( p.vNamesIn, Gia_ObjCioId(pRepr) ) );
            } else {
                Abc_Print( 1, "Original flop %s is proved equivalent to internal node %d.\n",
                    Vec_PtrEntry( p.vNamesIn, Gia_ObjCioId(pObj) ), Gia_ObjId(p, pRepr) ); }
        }
    }
}


/**Function*************************************************************

  Synopsis    [Top-level procedure for register correspondence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManLSCorrespondence(pAig *Gia_Man_t, pPars *Cec_ParCor_t) *Gia_Man_t {  
    var pNew, pTemp *Gia_Man_t;
    var pInitState *uint32;
    var RetValue int32;
    // inlining of macro ABC_FREE
    ((pAig.pReprs) ? (free((char *) (pAig.pReprs)), (pAig.pReprs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pAig.pNexts) ? (free((char *) (pAig.pNexts)), (pAig.pNexts) = 0) : 0);
    if pPars.nPrefix == 0 {
        RetValue = Cec_ManLSCorrespondenceClasses( pAig, pPars );
        if RetValue == 0 {
            return Gia_ManDup( pAig ); }
    
    } else {
        // compute the cycles AIG
        pInitState = Cec_ManComputeInitState( pAig, pPars.nPrefix );
        pTemp = Gia_ManDupFlip( pAig, (int *)pInitState );
        // inlining of macro ABC_FREE
        ((pInitState) ? (free((char *) (pInitState)), (pInitState) = 0) : 0);
        // compute classes of this AIG
        RetValue = Cec_ManLSCorrespondenceClasses( pTemp, pPars );
        // transfer the class info
        pAig.pReprs = pTemp.pReprs; pTemp.pReprs = nil;
        pAig.pNexts = pTemp.pNexts; pTemp.pNexts = nil;
        // perform additional BMC
        pPars.fUseCSat = 0;
        pPars.nBTLimit = Abc_MaxInt( pPars.nBTLimit, 1000 );
        Cec_ManLSCorrespondenceBmc( pAig, pPars, pPars.nPrefix );
/*
        // transfer the class info back
        pTemp->pReprs = pAig->pReprs; pAig->pReprs = NULL;
        pTemp->pNexts = pAig->pNexts; pAig->pNexts = NULL;
        // continue refining
        RetValue = Cec_ManLSCorrespondenceClasses( pTemp, pPars );
        // transfer the class info
        pAig->pReprs = pTemp->pReprs; pTemp->pReprs = NULL;
        pAig->pNexts = pTemp->pNexts; pTemp->pNexts = NULL;
*/
        Gia_ManStop( pTemp );
    }
    // derive reduced AIG
    if pPars.fMakeChoices {
        pNew = Gia_ManEquivToChoices( pAig, 1 );
//        Gia_ManHasChoices_very_old( pNew );
    
    } else {
//        Gia_ManEquivImprove( pAig );
        pNew = Gia_ManCorrReduce( pAig );
        pNew = Gia_ManSeqCleanup( pTemp = pNew );
        Gia_ManStop( pTemp );
        //Gia_AigerWrite( pNew, "reduced.aig", 0, 0, 0 );
    }
    // report the results
    if pPars.fVerbose {
        Abc_Print( 1, "NBeg = %d. NEnd = %d. (Gain = %6.2f %%).  RBeg = %d. REnd = %d. (Gain = %6.2f %%).\n", 
            Gia_ManAndNum(pAig), Gia_ManAndNum(pNew), 
            100.0*(Gia_ManAndNum(pAig)-Gia_ManAndNum(pNew))/(Gia_ManAndNum(pAig)?Gia_ManAndNum(pAig):1), 
            Gia_ManRegNum(pAig), Gia_ManRegNum(pNew), 
            100.0*(Gia_ManRegNum(pAig)-Gia_ManRegNum(pNew))/(Gia_ManRegNum(pAig)?Gia_ManRegNum(pAig):1) );
    }
    if pPars.nPrefix && (Gia_ManAndNum(pNew) < Gia_ManAndNum(pAig) || Gia_ManRegNum(pNew) < Gia_ManRegNum(pAig)) {
        Abc_Print( 1, "The reduced AIG was produced using %d-th invariants and will not verify.\n", pPars.nPrefix ); }
    // print verbose info about equivalences
    if pPars.fVerboseFlops {
        if pAig.vNamesIn == nil {
            Abc_Print( 1, "Flop output names are not available. Use command \"&get -n\".\n" );
        } else {
            Cec_ManPrintFlopEquivs( pAig ); }
    }
    // copy names if present
    if pAig.vNamesIn {
        var pName string; var i int32;
        pNew.vNamesIn = Vec_PtrDupStr( pAig.vNamesIn );
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Gia_ManCiNum(pNew); (i < Vec_PtrSize(pNew.vNamesIn)) && (((pName) = (char *)Vec_PtrEntry(pNew.vNamesIn, i)), 1); i++  {
            // inlining of macro ABC_FREE
            ((pName) ? (free((char *) (pName)), (pName) = 0) : 0);
        }
        Vec_PtrShrink( pNew.vNamesIn, Gia_ManCiNum(pNew) );
    }
    if pAig.vNamesOut {
        var pName string; var i int32;
        pNew.vNamesOut = Vec_PtrDupStr( pAig.vNamesOut );
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Gia_ManCoNum(pNew); (i < Vec_PtrSize(pNew.vNamesOut)) && (((pName) = (char *)Vec_PtrEntry(pNew.vNamesOut, i)), 1); i++  {
            // inlining of macro ABC_FREE
            ((pName) ? (free((char *) (pName)), (pName) = 0) : 0);
        }
        Vec_PtrShrink( pNew.vNamesOut, Gia_ManCoNum(pNew) );
    }
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

