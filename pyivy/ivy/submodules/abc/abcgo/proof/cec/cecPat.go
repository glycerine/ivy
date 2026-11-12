package cec

import "time"

/**CFile****************************************************************

  FileName    [cecPat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Simulation pattern manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecPat.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatStoreNum(p *Cec_ManPat_t, Num int32) {
    var x uint32  = uint32(Num);
    assert(Num >= 0, "Num >= 0");
    for  x & ~0x7f  {
        Vec_StrPush( p.vStorage, byte(((x & 0x7f) | 0x80)) );
        x >>= 7;
    }
    Vec_StrPush( p.vStorage, byte(x) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatRestoreNum(p *Cec_ManPat_t) int32 {
    var ch, i, x int32;
    for  i = 0; (ch = Vec_StrEntry(p.vStorage, p.iStart++)) & 0x80; i++  {
        x |= (ch & 0x7f) << (7 * i);
    }
    return x | (ch << (7 * i));
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatStore(p *Cec_ManPat_t, vPat *Vec_Int_t) {
    var i, Number, NumberPrev int32;
    assert(Vec_IntSize(vPat) > 0, "Vec_IntSize(vPat) > 0");
    Cec_ManPatStoreNum( p, Vec_IntSize(vPat) );
    NumberPrev = Vec_IntEntry( vPat, 0 );
    Cec_ManPatStoreNum( p, NumberPrev );
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(vPat)) && (((Number) = Vec_IntEntry(vPat, i)), 1); i++  {
        assert(NumberPrev < Number, "NumberPrev < Number");
        Cec_ManPatStoreNum( p, Number - NumberPrev );
        NumberPrev = Number;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatRestore(p *Cec_ManPat_t, vPat *Vec_Int_t) {
    var i, Size, Number int32;
    Vec_IntClear( vPat );
    Size = Cec_ManPatRestoreNum( p );
    Number = Cec_ManPatRestoreNum( p );
    Vec_IntPush( vPat, Number );
    for  i = 1; i < Size; i++  {
        Number += Cec_ManPatRestoreNum( p );
        Vec_IntPush( vPat, Number );
    }
    assert(Vec_IntSize(vPat) == Size, "Vec_IntSize(vPat) == Size");
}


/**Function*************************************************************

  Synopsis    [Derives satisfying assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatComputePattern_rec(pSat *Cec_ManSat_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    var Counter int32;
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return 0; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        pObj.fMark1 = Cec_ObjSatVarValue( pSat, pObj );
        return 1;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Counter += Cec_ManPatComputePattern_rec( pSat, p, Gia_ObjFanin0(pObj) );
    Counter += Cec_ManPatComputePattern_rec( pSat, p, Gia_ObjFanin1(pObj) );
    pObj.fMark1 = (Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj)) & 
                   (Gia_ObjFanin1(pObj).fMark1 ^ Gia_ObjFaninC1(pObj));
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Derives satisfying assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatComputePattern1_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vPat *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        Vec_IntPush( vPat, Abc_Var2Lit( Gia_ObjCioId(pObj), pObj.fMark1==0 ) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if pObj.fMark1 == 1 {
        Cec_ManPatComputePattern1_rec( p, Gia_ObjFanin0(pObj), vPat );
        Cec_ManPatComputePattern1_rec( p, Gia_ObjFanin1(pObj), vPat );
    
    } else {
        assert((Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj)) == 0 ||
                        (Gia_ObjFanin1(pObj).fMark1 ^ Gia_ObjFaninC1(pObj)) == 0, "(Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj)) == 0 ||\n                (Gia_ObjFanin1(pObj).fMark1 ^ Gia_ObjFaninC1(pObj)) == 0");
        if (Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj)) == 0 {
            Cec_ManPatComputePattern1_rec( p, Gia_ObjFanin0(pObj), vPat );
        } else {
            Cec_ManPatComputePattern1_rec( p, Gia_ObjFanin1(pObj), vPat ); }
    }
}

/**Function*************************************************************

  Synopsis    [Derives satisfying assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatComputePattern2_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vPat *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        Vec_IntPush( vPat, Abc_Var2Lit( Gia_ObjCioId(pObj), pObj.fMark1==0 ) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if pObj.fMark1 == 1 {
        Cec_ManPatComputePattern2_rec( p, Gia_ObjFanin0(pObj), vPat );
        Cec_ManPatComputePattern2_rec( p, Gia_ObjFanin1(pObj), vPat );
    
    } else {
        assert((Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj)) == 0 ||
                        (Gia_ObjFanin1(pObj).fMark1 ^ Gia_ObjFaninC1(pObj)) == 0, "(Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj)) == 0 ||\n                (Gia_ObjFanin1(pObj).fMark1 ^ Gia_ObjFaninC1(pObj)) == 0");
        if (Gia_ObjFanin1(pObj).fMark1 ^ Gia_ObjFaninC1(pObj)) == 0 {
            Cec_ManPatComputePattern2_rec( p, Gia_ObjFanin1(pObj), vPat );
        } else {
            Cec_ManPatComputePattern2_rec( p, Gia_ObjFanin0(pObj), vPat ); }
    }
}

/**Function*************************************************************

  Synopsis    [Derives satisfying assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatComputePattern3_rec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    var Value0, Value1, Value int32;
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return (pObj.fMark1 << 1) | pObj.fMark0; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        pObj.fMark0 = 1;
        pObj.fMark1 = 1;
        return GIA_UND;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Value0 = Cec_ManPatComputePattern3_rec( p, Gia_ObjFanin0(pObj) );
    Value1 = Cec_ManPatComputePattern3_rec( p, Gia_ObjFanin1(pObj) );
    Value = Gia_XsimAndCond( Value0, Gia_ObjFaninC0(pObj), Value1, Gia_ObjFaninC1(pObj) );
    pObj.fMark0 =  (Value & 1);
    pObj.fMark1 = ((Value >> 1) & 1);
    return Value;
}

/**Function*************************************************************

  Synopsis    [Derives satisfying assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatVerifyPattern(p *Gia_Man_t, pObj *Gia_Obj_t, vPat *Vec_Int_t) {
    var pTemp *Gia_Obj_t;
    var i, Value int32;
    Gia_ManIncrementTravId( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPat)) && (((Value) = Vec_IntEntry(vPat, i)), 1); i++  {
        pTemp = Gia_ManCi( p, Abc_Lit2Var(Value) );
//        assert( Abc_LitIsCompl(Value) != (int)pTemp->fMark1 );
        if pTemp.fMark1 {
            pTemp.fMark0 = 0;
            pTemp.fMark1 = 1;
        
        } else {
            pTemp.fMark0 = 1;
            pTemp.fMark1 = 0;
        }
        Gia_ObjSetTravIdCurrent( p, pTemp );
    }
    Value = Cec_ManPatComputePattern3_rec( p, Gia_ObjFanin0(pObj) );
    Value = Gia_XsimNotCond( Value, Gia_ObjFaninC0(pObj) );
    if Value != GIA_ONE {
        Abc_Print( 1, "Cec_ManPatVerifyPattern(): Verification failed.\n" ); }
    assert(Value == GIA_ONE, "Value == GIA_ONE");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatComputePattern4_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    pObj.fMark0 = 0;
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Cec_ManPatComputePattern4_rec( p, Gia_ObjFanin0(pObj) );
    Cec_ManPatComputePattern4_rec( p, Gia_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatCleanMark0(p *Gia_Man_t, pObj *Gia_Obj_t) {
    assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
    Gia_ManIncrementTravId( p );
    Cec_ManPatComputePattern4_rec( p, Gia_ObjFanin0(pObj) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatSavePattern(pMan *Cec_ManPat_t, p *Cec_ManSat_t, pObj *Gia_Obj_t) {
    var vPat *Vec_Int_t;
    var nPatLits int32;
    var clkTotal abctime  = Abc_Clock();
//    abctime clk;
    assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
    pMan.nPats++;
    pMan.nPatsAll++;
    // compute values in the cone of influence
//clk = Abc_Clock();
    Gia_ManIncrementTravId( p.pAig );
    nPatLits = Cec_ManPatComputePattern_rec( p, p.pAig, Gia_ObjFanin0(pObj) );
    assert((Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj)) == 1, "(Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj)) == 1");
    pMan.nPatLits += nPatLits;
    pMan.nPatLitsAll += nPatLits;
//pMan->timeFind += Abc_Clock() - clk;
    // compute sensitizing path
//clk = Abc_Clock();
    Vec_IntClear( pMan.vPattern1 );
    Gia_ManIncrementTravId( p.pAig );
    Cec_ManPatComputePattern1_rec( p.pAig, Gia_ObjFanin0(pObj), pMan.vPattern1 );
    // compute sensitizing path
    Vec_IntClear( pMan.vPattern2 );
    Gia_ManIncrementTravId( p.pAig );
    Cec_ManPatComputePattern2_rec( p.pAig, Gia_ObjFanin0(pObj), pMan.vPattern2 );
    // compare patterns
    if Vec_IntSize(pMan.vPattern1) < Vec_IntSize(pMan.vPattern2) {
        vPat = pMan.vPattern1;
    } else {
        vPat = pMan.vPattern2;
    }
    pMan.nPatLitsMin += Vec_IntSize(vPat);
    pMan.nPatLitsMinAll += Vec_IntSize(vPat);
//pMan->timeShrink += Abc_Clock() - clk;
    // verify pattern using ternary simulation
//clk = Abc_Clock();
//    Cec_ManPatVerifyPattern( p->pAig, pObj, vPat );
//pMan->timeVerify += Abc_Clock() - clk;
    // sort pattern
//clk = Abc_Clock();
    Vec_IntSort( vPat, 0 );
//pMan->timeSort += Abc_Clock() - clk;
    // save pattern
    Cec_ManPatStore( pMan, vPat );
    pMan.timeTotal += Abc_Clock() - clkTotal;
}
func Cec_ManPatSavePatternCSat(pMan *Cec_ManPat_t, vPat *Vec_Int_t) {
    // sort pattern
    Vec_IntSort( vPat, 0 );
    // save pattern
    Cec_ManPatStore( pMan, vPat );
}

/**Function*************************************************************

  Synopsis    [Packs patterns into array of simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

*************************************`**********************************/
func Cec_ManPatCollectTry(vInfo *Vec_Ptr_t, vPres *Vec_Ptr_t, iBit int32, pLits *int32, nLits int32) int32 {
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

  Synopsis    [Packs patterns into array of simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatCollectPatterns(pMan *Cec_ManPat_t, nInputs int32, nWordsInit int32) *Vec_Ptr_t {
    var vPat *Vec_Int_t  = pMan.vPattern1;
    var vInfo, vPres *Vec_Ptr_t;
    var k int32;
    var kMax int32 = -1;
    var nPatterns int32;
    var iStartOld int32  = pMan.iStart;
    var nWords int32  = nWordsInit;
    var nBits int32  = 32 * nWords;
    var clk abctime  = Abc_Clock();
    vInfo = Vec_PtrAllocSimInfo( nInputs, nWords );
    Gia_ManRandomInfo( vInfo, 0, 0, nWords );
    vPres = Vec_PtrAllocSimInfo( nInputs, nWords );
    Vec_PtrCleanSimInfo( vPres, 0, nWords );
    for  pMan.iStart < Vec_StrSize(pMan.vStorage)  {
        nPatterns++;
        Cec_ManPatRestore( pMan, vPat );
        for  k = 1; k < nBits; k++, k += ((k % (32 * nWordsInit)) == 0)  {
            if Cec_ManPatCollectTry( vInfo, vPres, k, (int *)Vec_IntArray(vPat), Vec_IntSize(vPat) ) {
                break; }
        }
        kMax = Abc_MaxInt( kMax, k );
        if k == nBits-1 {
            Vec_PtrReallocSimInfo( vInfo );
            Gia_ManRandomInfo( vInfo, 0, nWords, 2*nWords );
            Vec_PtrReallocSimInfo( vPres );
            Vec_PtrCleanSimInfo( vPres, nWords, 2*nWords );
            nWords *= 2;
            nBits *= 2;
        }
    }
    Vec_PtrFree( vPres );
    pMan.nSeries = Vec_PtrReadWordsSimInfo(vInfo) / nWordsInit;
    pMan.timePack += Abc_Clock() - clk;
    pMan.timeTotal += Abc_Clock() - clk;
    pMan.iStart = iStartOld;
    if pMan.fVerbose {
        Abc_Print( 1, "Total = %5d. Max used = %5d. Full = %5d. Series = %d. ", 
            nPatterns, kMax, nWordsInit*32, pMan.nSeries );
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        Cec_ManPatPrintStats( pMan );
    }
    return vInfo;
}


/**Function*************************************************************

  Synopsis    [Packs patterns into array of simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatPackPatterns(vCexStore *Vec_Int_t, nInputs int32, nRegs int32, nWordsInit int32) *Vec_Ptr_t {
    var vPat *Vec_Int_t;
    var vInfo, vPres *Vec_Ptr_t;
    var k, nSize, iStart, kMax, nPatterns int32;
    var nWords int32  = nWordsInit;
    var nBits int32  = 32 * nWords;
//    int RetValue;
    assert(nRegs <= nInputs, "nRegs <= nInputs");
    vPat  = Vec_IntAlloc( 100 );

    vInfo = Vec_PtrAllocSimInfo( nInputs, nWords );
    Vec_PtrCleanSimInfo( vInfo, 0, nWords );
    Gia_ManRandomInfo( vInfo, nRegs, 0, nWords );

    vPres = Vec_PtrAllocSimInfo( nInputs, nWords );
    Vec_PtrCleanSimInfo( vPres, 0, nWords );
    iStart = 0;
    for  iStart < Vec_IntSize(vCexStore)  {
        nPatterns++;
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
        for  k = 1; k < nBits; k++, k += ((k % (32 * nWordsInit)) == 0)  {
            if Cec_ManPatCollectTry( vInfo, vPres, k, (int *)Vec_IntArray(vPat), Vec_IntSize(vPat) ) {
                break; }
        }

//        k = kMax + 1;
//        RetValue = Cec_ManPatCollectTry( vInfo, vPres, k, (int *)Vec_IntArray(vPat), Vec_IntSize(vPat) );
//        assert( RetValue == 1 );

        kMax = Abc_MaxInt( kMax, k );
        if k == nBits-1 {
            Vec_PtrReallocSimInfo( vInfo );
            Vec_PtrCleanSimInfo( vInfo, nWords, 2*nWords );
            Gia_ManRandomInfo( vInfo, nRegs, nWords, 2*nWords );

            Vec_PtrReallocSimInfo( vPres );
            Vec_PtrCleanSimInfo( vPres, nWords, 2*nWords );
            nWords *= 2;
            nBits *= 2;
        }
    }
//    Abc_Print( 1, "packed %d patterns into %d vectors (out of %d)\n", nPatterns, kMax, nBits );
    Vec_PtrFree( vPres );
    Vec_IntFree( vPat );
    return vInfo;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

