package util

import "fmt"

/**CFile****************************************************************

  FileName    [utilCex.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Handling counter-examples.]

  Synopsis    [Handling counter-examples.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Feburary 13, 2011.]

  Revision    [$Id: utilCex.c,v 1.00 2011/02/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <string.h>
//#include <stdlib.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "utilCex.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Allocates a counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexAlloc(nRegs int32, nRealPis int32, nFrames int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var nWords int32  = Abc_BitWordNum( nRegs + nRealPis * nFrames );
    // inlining of macro ABC_ALLOC
    pCex = (Abc_Cex_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Abc_Cex_t) + sizeof(unsigned) * nWords)));
    memset( pCex, 0, sizeof(Abc_Cex_t) + sizeof(unsigned) * nWords );
    pCex.nRegs  = nRegs;
    pCex.nPis   = nRealPis;
    pCex.nBits  = nRegs + nRealPis * nFrames;
    return pCex;
}
func Abc_CexAllocFull(nRegs int32, nRealPis int32, nFrames int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var nWords int32  = Abc_BitWordNum( nRegs + nRealPis * nFrames );
    // inlining of macro ABC_ALLOC
    pCex = (Abc_Cex_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Abc_Cex_t) + sizeof(unsigned) * nWords)));
    memset( pCex, 0xFF, sizeof(Abc_Cex_t) + sizeof(unsigned) * nWords );
    pCex.nRegs  = nRegs;
    pCex.nPis   = nRealPis;
    pCex.nBits  = nRegs + nRealPis * nFrames;
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Make the trivial counter-example for the trivially asserted output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexMakeTriv(nRegs int32, nTruePis int32, nTruePos int32, iFrameOut int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var iPo, iFrame int32;
    assert(nRegs > 0, "nRegs > 0");
    iPo    = iFrameOut % nTruePos;
    iFrame = iFrameOut / nTruePos;
    // allocate the counter example
    pCex = Abc_CexAlloc( nRegs, nTruePis, iFrame + 1 );
    pCex.iPo    = iPo;
    pCex.iFrame = iFrame;
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Derives the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexCreate(nRegs int32, nPis int32, pArray *int32, iFrame int32, iPo int32, fSkipRegs int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var i int32;
    pCex = Abc_CexAlloc( nRegs, nPis, iFrame+1 );
    pCex.iPo    = iPo;
    pCex.iFrame = iFrame;
    if pArray == nil {
        return pCex; }
    if fSkipRegs != 0 {
        for  i = nRegs; i < pCex.nBits; i++  {
            if pArray[i-nRegs] {
                Abc_InfoSetBit( pCex.pData, i ); }
        }
    
    } else {
        for  i = 0; i < pCex.nBits; i++  {
            if pArray[i] {
                Abc_InfoSetBit( pCex.pData, i ); }
        }
    }
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Make the trivial counter-example for the trivially asserted output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexDup(p *Abc_Cex_t, nRegsNew int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var i int32;
    if p == (Abc_Cex_t *)(ABC_PTRINT_T)1 {
        return p; }
    if nRegsNew == -1 {
        nRegsNew = p.nRegs; }
    pCex = Abc_CexAlloc( nRegsNew, p.nPis, p.iFrame+1 );
    pCex.iPo    = p.iPo;
    pCex.iFrame = p.iFrame;
    for  i = p.nRegs; i < p.nBits; i++  {
        if Abc_InfoHasBit(p.pData, i) {
            Abc_InfoSetBit( pCex.pData, pCex.nRegs + i - p.nRegs ); }
    }
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Derives CEX from comb model.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexDeriveFromCombModel(pModel *int32, nPis int32, nRegs int32, iPo int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var i int32;
    pCex = Abc_CexAlloc( nRegs, nPis, 1 );
    pCex.iPo = iPo;
    pCex.iFrame = 0;
    for  i = 0; i < nPis; i++  {
        if pModel[i] {
            pCex.pData[i>>5] |= (1<<(i & 31)); }
    }     
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Derives CEX from comb model.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexMerge(pCex *Abc_Cex_t, pPart *Abc_Cex_t, iFrBeg int32, iFrEnd int32) *Abc_Cex_t {
    var pNew *Abc_Cex_t;
    var nFramesGain int32;
    var i, f, iBit int32;

    if iFrBeg < 0 { fmt.Printf("Starting frame is less than 0.\n"); return nil; }
    if iFrEnd < 0 { fmt.Printf("Stopping frame is less than 0.\n"); return nil; }
    if iFrBeg > pCex.iFrame { fmt.Printf("Starting frame is more than the last frame of CEX (%d).\n", pCex.iFrame); return nil; }
    if iFrEnd > pCex.iFrame { fmt.Printf("Stopping frame is more than the last frame of CEX (%d).\n", pCex.iFrame); return nil; }
    if iFrBeg > iFrEnd { fmt.Printf("Starting frame (%d) should be less than stopping frame (%d).\n", iFrBeg, iFrEnd); return nil; }
    assert(iFrBeg >= 0 && iFrBeg <= pCex.iFrame, "iFrBeg >= 0 && iFrBeg <= pCex.iFrame");
    assert(iFrEnd >= 0 && iFrEnd <= pCex.iFrame, "iFrEnd >= 0 && iFrEnd <= pCex.iFrame");
    assert(iFrBeg <= iFrEnd, "iFrBeg <= iFrEnd");

    assert(pCex.nPis == pPart.nPis, "pCex.nPis == pPart.nPis");
    assert(iFrEnd - iFrBeg + pPart.iPo >= pPart.iFrame, "iFrEnd - iFrBeg + pPart.iPo >= pPart.iFrame");

    nFramesGain = iFrEnd - iFrBeg + pPart.iPo - pPart.iFrame;
    pNew = Abc_CexAlloc( pCex.nRegs, pCex.nPis, pCex.iFrame + 1 - nFramesGain );
    pNew.iPo    = pCex.iPo;
    pNew.iFrame = pCex.iFrame - nFramesGain;

    for  iBit = 0; iBit < pCex.nRegs; iBit++  {
        if Abc_InfoHasBit(pCex.pData, iBit) {
            Abc_InfoSetBit( pNew.pData, iBit ); }
    }
    for  f = 0; f < iFrBeg; f++  {
        for  i = 0; i < pCex.nPis; i, iBit = i+1, iBit+1  {
            if Abc_InfoHasBit(pCex.pData, pCex.nRegs + pCex.nPis * f + i) {
                Abc_InfoSetBit( pNew.pData, iBit ); }
        }
    }
    for  f = 0; f < pPart.iFrame; f++  {
        for  i = 0; i < pCex.nPis; i, iBit = i+1, iBit+1  {
            if Abc_InfoHasBit(pPart.pData, pPart.nRegs + pCex.nPis * f + i) {
                Abc_InfoSetBit( pNew.pData, iBit ); }
        }
    }
    for  f = iFrEnd; f <= pCex.iFrame; f++  {
        for  i = 0; i < pCex.nPis; i, iBit = i+1, iBit+1  {
            if Abc_InfoHasBit(pCex.pData, pCex.nRegs + pCex.nPis * f + i) {
                Abc_InfoSetBit( pNew.pData, iBit ); }
        }
    }
    assert(iBit == pNew.nBits, "iBit == pNew.nBits");

    return pNew;
}

/**Function*************************************************************

  Synopsis    [Prints out the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexPrintStats(p *Abc_Cex_t) {
    var k, Counter int32;
    if p == nil {
        fmt.Printf("The counter example is NULL.\n");
        return;
    }
    if p == (Abc_Cex_t *)(ABC_PTRINT_T)1 {
        fmt.Printf("The counter example is present but not available (pointer has value \"1\").\n");
        return;
    }
    for  k = 0; k < p.nBits; k++  {
        Counter += Abc_InfoHasBit(p.pData, k);
    }
    fmt.Printf("CEX: Po =%4d  Frame =%4d  FF = %d  PI = %d  Bit =%8d  1s =%8d (%5.2f %%)\n", p.iPo, p.iFrame, p.nRegs, p.nPis, p.nBits, Counter, 100.0 * Counter / (p.nBits - p.nRegs));
}
func Abc_CexPrintStatsInputs(p *Abc_Cex_t, nRealPis int32) {
    var k, Counter, CounterPi, CounterPpi int32;
    if p == nil {
        fmt.Printf("The counter example is NULL.\n");
        return;
    }
    if p == (Abc_Cex_t *)(ABC_PTRINT_T)1 {
        fmt.Printf("The counter example is present but not available (pointer has value \"1\").\n");
        return;
    }
    assert(nRealPis <= p.nPis, "nRealPis <= p.nPis");
    for  k = 0; k < p.nBits; k++  {
        Counter += Abc_InfoHasBit(p.pData, k);
        if nRealPis == p.nPis {
            continue; }
        if (k - p.nRegs) % p.nPis < nRealPis {
            CounterPi += Abc_InfoHasBit(p.pData, k);
        } else {
            CounterPpi += Abc_InfoHasBit(p.pData, k); }
    }
    fmt.Printf("CEX: Po =%4d  Fr =%4d  FF = %d  PI = %d  Bit =%7d  1 =%8d (%5.2f %%)", p.iPo, p.iFrame, p.nRegs, p.nPis, p.nBits, Counter, 100.0 * Counter    / ((p.iFrame + 1) * p.nPis ));
    if nRealPis < p.nPis {
        fmt.Printf(" 1pi =%8d (%5.2f %%) 1ppi =%8d (%5.2f %%)", CounterPi, 100.0 * CounterPi  / ((p.iFrame + 1) * nRealPis ), CounterPpi, 100.0 * CounterPpi / ((p.iFrame + 1) * (p.nPis - nRealPis)));
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Prints out the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexPrint(p *Abc_Cex_t) {
    var i, f, k int32;
    if p == nil {
        fmt.Printf("The counter example is NULL.\n");
        return;
    }
    if p == (Abc_Cex_t *)(ABC_PTRINT_T)1 {
        fmt.Printf("The counter example is present but not available (pointer has value \"1\").\n");
        return;
    }
    Abc_CexPrintStats( p );
    fmt.Printf("State    : ");
    for  k = 0; k < p.nRegs; k++  {
        fmt.Printf("%d", Abc_InfoHasBit(p.pData, k));
    }
    fmt.Printf("\n");
    for  f = 0; f <= p.iFrame; f++  {
        fmt.Printf("Frame %3d : ", f);
        for  i = 0; i < p.nPis; i++  {
            fmt.Printf("%d", Abc_InfoHasBit(p.pData, k++));
        }
        fmt.Printf("\n");
    }
    assert(k == p.nBits, "k == p.nBits");
}

/**Function*************************************************************

  Synopsis    [Frees the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexFreeP(p **Abc_Cex_t) {
    if *p == nil {
        return; }
    if *p == (Abc_Cex_t *)(ABC_PTRINT_T)1 {
        *p = nil;
    } else {
        // inlining of macro ABC_FREE
        ((*p) ? (free((char *) (*p)), (*p) = 0) : 0); }
}

/**Function*************************************************************

  Synopsis    [Frees the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexFree(p *Abc_Cex_t) {
    if p == (Abc_Cex_t *)(ABC_PTRINT_T)1 {
        return; }
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Transform CEX after phase abstraction with nFrames frames.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexTransformPhase(p *Abc_Cex_t, nPisOld int32, nPosOld int32, nRegsOld int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var nFrames int32  = p.nPis / nPisOld;
    var nPosNew int32  = nPosOld * nFrames;
    assert(p.nPis % nPisOld == 0, "p.nPis % nPisOld == 0");
    assert(p.iPo < nPosNew, "p.iPo < nPosNew");
    pCex = Abc_CexDup( p, nRegsOld );
    pCex.nPis   = nPisOld;
    pCex.iPo    = -1;
    pCex.iFrame = (p.iFrame + 1) * nFrames - 1;
    pCex.nBits  = p.nBits;
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Transform CEX after phase temporal decomposition with nFrames frames.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexTransformTempor(p *Abc_Cex_t, nPisOld int32, nPosOld int32, nRegsOld int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var i int32;
    var k int32;
    var iBit int32 = nRegsOld;
    var nFrames int32 = p.nPis / nPisOld - 1;
    assert(p.iFrame > 0, "p.iFrame > 0"); // otherwise tempor did not properly check for failures in the prefix
    assert(p.nPis % nPisOld == 0, "p.nPis % nPisOld == 0");
    pCex = Abc_CexAlloc( nRegsOld, nPisOld, nFrames + p.iFrame + 1 );
    pCex.iPo    = p.iPo;
    pCex.iFrame = nFrames + p.iFrame;
    for  i = 0; i < nFrames; i++  {
        for  k = 0; k < nPisOld; k, iBit = k+1, iBit+1  {
            if Abc_InfoHasBit(p.pData, p.nRegs + (i+1)*nPisOld + k) {
                Abc_InfoSetBit( pCex.pData, iBit ); }
        }
    }
    for  i = 0; i <= p.iFrame; i++  {
        for  k = 0; k < nPisOld; k, iBit = k+1, iBit+1  {
            if Abc_InfoHasBit(p.pData, p.nRegs + i*p.nPis + k) {
                Abc_InfoSetBit( pCex.pData, iBit ); }
        }
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Transform CEX after "logic; undc; st; zero".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexTransformUndc(p *Abc_Cex_t, pInit string) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var nFlops int32  = len(pInit);
    var i, f, iBit, iAddPi, nAddPis int32;
    // count how many flops got a new PI
    for  i = 0; i < nFlops; i++  {
        nAddPis += int32((pInit[i] == 'x' || pInit[i] == 'X'));
    }
    // create new CEX
    pCex = Abc_CexAlloc( nFlops, p.nPis - nAddPis, p.iFrame + 1 );
    pCex.iPo    = p.iPo;
    pCex.iFrame = p.iFrame;
    for  iBit = 0; iBit < nFlops; iBit++  {
        if pInit[iBit] == '1' || ((pInit[iBit] == 'x' || pInit[iBit] == 'X') && Abc_InfoHasBit(p.pData, p.nRegs + p.nPis - nAddPis + iAddPi)) {
            Abc_InfoSetBit( pCex.pData, iBit ); }
        iAddPi += int32((pInit[iBit] == 'x' || pInit[iBit] == 'X'));
    }
    assert(iAddPi == nAddPis, "iAddPi == nAddPis");
    // add timeframes
    for  f = 0; f <= p.iFrame; f++  {
        for  i = 0; i < pCex.nPis; i, iBit = i+1, iBit+1  {
            if Abc_InfoHasBit(p.pData, p.nRegs + p.nPis * f + i) {
                Abc_InfoSetBit( pCex.pData, iBit ); }
        }
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Derives permuted CEX using permutation of its inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexPermute(p *Abc_Cex_t, vMapOld2New *Vec_Int_t) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var i, iNew int32;
    assert(Vec_IntSize(vMapOld2New) == p.nPis, "Vec_IntSize(vMapOld2New) == p.nPis");
    pCex = Abc_CexAlloc( p.nRegs, p.nPis, p.iFrame+1 );
    pCex.iPo    = p.iPo;
    pCex.iFrame = p.iFrame;
    for  i = p.nRegs; i < p.nBits; i++  {
        if Abc_InfoHasBit(p.pData, i) {
            iNew = p.nRegs + p.nPis * ((i - p.nRegs) / p.nPis) + Vec_IntEntry( vMapOld2New, (i - p.nRegs) % p.nPis );
            Abc_InfoSetBit( pCex.pData, iNew );
        }
    }
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Derives permuted CEX using two canonical permutations.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexPermuteTwo(p *Abc_Cex_t, vPermOld *Vec_Int_t, vPermNew *Vec_Int_t) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var vPerm *Vec_Int_t;
    var i, eOld, eNew int32;
    assert(Vec_IntSize(vPermOld) == p.nPis, "Vec_IntSize(vPermOld) == p.nPis");
    assert(Vec_IntSize(vPermNew) == p.nPis, "Vec_IntSize(vPermNew) == p.nPis");
    vPerm = Vec_IntStartFull( p.nPis );
    // inlining of macro Vec_IntForEachEntryTwo
    for  i = 0; (i < Vec_IntSize(vPermOld)) && (((eOld) = Vec_IntEntry(vPermOld, i)), 1) && (((eNew) = Vec_IntEntry(vPermNew, i)), 1); i++  {
        Vec_IntWriteEntry( vPerm, eOld, eNew );
    }
    pCex = Abc_CexPermute( p, vPerm );
    Vec_IntFree( vPerm );
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Count the number of 1s in the CEX.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CexOnes32(i uint32) int32 {
    i = i - ((i >> 1) & 0x55555555);
    i = (i & 0x33333333) + ((i >> 2) & 0x33333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F);
    return (i*(0x01010101))>>24;
}
func Abc_CexCountOnes(p *Abc_Cex_t) int32 {
    var nWords int32  = Abc_BitWordNum( p.nBits );
    var i, Count int32;
    for  i = 0; i < nWords; i++  {
        Count += Abc_CexOnes32( p.pData[i] );
    }
    return Count;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

