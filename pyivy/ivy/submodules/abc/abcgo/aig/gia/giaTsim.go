package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaTsim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Ternary simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaTsim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Gia_ManTerStateHash(pState *uint32, nWords int32, nTableSize int32) int32 {
    
    var uHash uint32;
    var i int32;
    uHash = 0;
    for  i = 0; i < nWords; i++  {
        uHash ^= pState[i] * spStatic_Gia_ManTerStateHash_s_FPrimes[i & 0x7F];
    }
    return uHash % nTableSize;
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerStateLookup(pState *uint32, nWords int32, pBins **uint32, nBins int32) *uint32 {
    var pEntry *uint32;
    var Hash int32  = Gia_ManTerStateHash( pState, nWords, nBins );
    for  pEntry = pBins[Hash]; (pEntry != nil); pEntry = Gia_ManTerStateNext(pEntry, nWords)  {
        if !memcmp( pEntry, pState, sizeof(unsigned) * nWords ) {
            return pEntry; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerStateInsert(pState *uint32, nWords int32, pBins **uint32, nBins int32) {
    var Hash int32  = Gia_ManTerStateHash( pState, nWords, nBins );
    assert(!(Gia_ManTerStateLookup( pState, nWords, pBins, nBins ) != nil), "!Gia_ManTerStateLookup( pState, nWords, pBins, nBins )");
    Gia_ManTerStateSetNext( pState, nWords, pBins[Hash] );
    pBins[Hash] = pState;    
}

/**Function*************************************************************

  Synopsis    [Allocs new ternary state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerStateAlloc(nWords int32) *uint32 {
    // inlining of macro ABC_CALLOC
    return (unsigned *)((char *) calloc((size_t)(sizeof(unsigned) * nWords + sizeof(unsigned *)), sizeof(char)));
}

/**Function*************************************************************

  Synopsis    [Creates new ternary state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerStateCreate(p *Gia_ManTer_t) *uint32 {
    var i int32;
    var Value int32;
    var nPis int32 = Gia_ManPiNum(p.pAig);
    var pRes *uint32  = Gia_ManTerStateAlloc( p.nStateWords );
    for  i = nPis; i < Gia_ManCiNum(p.pAig); i++  {
        Value = Gia_ManTerSimInfoGet( p.pDataSimCis, i );
        Gia_ManTerSimInfoSet( pRes, i-nPis, Value );
        if Value == GIA_ZER {
            p.pCount0[i-nPis]++; }
        if Value == GIA_UND {
            p.pCountX[i-nPis]++; }
    }
    Vec_PtrPush( p.vStates, pRes );
    return pRes;    
}

/**Function*************************************************************

  Synopsis    [Performs one round of ternary simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Retires a set of registers to speed up convergence.]

  Description [Retire all non-ternary registers which has max number 
  of ternary values so far.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerRetire2(p *Gia_ManTer_t, pState *uint32) int32 {
    var i int32;
    var Entry int32;
    var iMaxTerValue int32 = -1;
    // find non-retired register with this value
    for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
        if Gia_ManTerSimInfoGet( pState, i ) != GIA_UND && !p.pRetired[i] && iMaxTerValue < p.pCountX[i] {
            iMaxTerValue = p.pCountX[i]; }
    }
    assert(iMaxTerValue >= 0, "iMaxTerValue >= 0");
    // retire the first registers with this value
    for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
        if Gia_ManTerSimInfoGet( pState, i ) != GIA_UND && !p.pRetired[i] && iMaxTerValue == p.pCountX[i] {
            assert(p.pRetired[i] == 0, "p.pRetired[i] == 0");
            p.pRetired[i] = 1;
            Vec_IntPush( p.vRetired, i );
            if iMaxTerValue == 0 {
                break; }
        }
    }
    // update all the retired registers
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vRetired)) && (((Entry) = Vec_IntEntry(p.vRetired, i)), 1); i++  {
        Gia_ManTerSimInfoSet( p.pDataSimCis, Gia_ManPiNum(p.pAig)+Entry, GIA_UND );
    }
    return Vec_IntSize(p.vRetired);
}

/**Function*************************************************************

  Synopsis    [Retires a set of registers to speed up convergence.]

  Description [Retire all non-ternary registers which has max number 
  of ternary values so far.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerRetire(p *Gia_ManTer_t, pThis *uint32, pPrev *uint32) int32 {
    var i, Entry int32;
    // find registers whose value has changed
    Vec_IntClear( p.vRetired );
    for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
        if Gia_ManTerSimInfoGet( pThis, i ) != Gia_ManTerSimInfoGet( pPrev, i ) {
            Vec_IntPush( p.vRetired, i ); }
    }
    // set all of them to zero
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vRetired)) && (((Entry) = Vec_IntEntry(p.vRetired, i)), 1); i++  {
        Gia_ManTerSimInfoSet( p.pDataSimCis, Gia_ManPiNum(p.pAig)+Entry, GIA_UND );
    }
    return Vec_IntSize(p.vRetired);
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerStatePrint(pState *uint32, nRegs int32, iNum int32) {
    var i, nZeros, nOnes, nDcs int32;
    fmt.Printf(" %4d : ", iNum);
    for  i = 0; i < nRegs; i++  {
        if Gia_ManTerSimInfoGet(pState, i) == GIA_ZER {
            fmt.Printf("0");
nZeros++;
        } else if Gia_ManTerSimInfoGet(pState, i) == GIA_ONE {
            fmt.Printf("1");
nOnes++;
        } else if Gia_ManTerSimInfoGet(pState, i) == GIA_UND {
            fmt.Printf("x");
nDcs++;
        } else {
            assert(false, "0"); }
    }
    fmt.Printf(" (0=%4d, 1=%4d, x=%4d)\n", nZeros, nOnes, nDcs);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerAnalyze2(vStates *Vec_Ptr_t, nRegs int32) {
    var pTemp *uint32;
    var pStates *uint32 = (unsigned *)Vec_PtrPop( vStates );
    var i, w, nZeros, nConsts, nStateWords int32;
    // detect constant zero registers
    nStateWords = Abc_BitWordNum( 2*nRegs );
    memset( pStates, 0, sizeof(int) * nStateWords );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vStates)) && (((pTemp) = (unsigned *)Vec_PtrEntry(vStates, i)), 1); i++  {
        for  w = 0; w < nStateWords; w++  {
            pStates[w] |= pTemp[w];
        }
    }
    // count the number of zeros
    nZeros = 0;
    for  i = 0; i < nRegs; i++  {
        if Gia_ManTerSimInfoGet(pStates, i) == GIA_ZER {
            nZeros++; }
    }
    fmt.Printf("Found %d constant registers.\n", nZeros);
    // detect non-ternary registers
    memset( pStates, 0, sizeof(int) * nStateWords );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vStates)) && (((pTemp) = (unsigned *)Vec_PtrEntry(vStates, i)), 1); i++  {
        for  w = 0; w < nStateWords; w++  {
            pStates[w] |= (~(pTemp[w] ^ (pTemp[w] >> 1)) & 0x55555555);
        }
    }
    // count the nonternary registers
    nConsts = 0;
    for  i = 0; i < nRegs; i++  {
        if Gia_ManTerSimInfoGet(pStates, i) == 0 {
            nConsts++; }
    }
    fmt.Printf("Found %d non-ternary registers.\n", nConsts);
    // return the state back
    Vec_PtrPush( vStates, pStates );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerAnalyze(p *Gia_ManTer_t) {
    var i, nZeros, nConsts int32;
    for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
        if p.pCount0[i] == Vec_PtrSize(p.vStates) {
            nZeros++;
        } else if p.pCountX[i] == 0 {
            nConsts++; }
    }
//    printf( "Found %d constant registers.\n", nZeros );
//    printf( "Found %d non-ternary registers.\n", nConsts );
}



/**Function*************************************************************

  Synopsis    [Transposes state vector for non-ternary registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerTranspose(p *Gia_ManTer_t) *Vec_Ptr_t {
    var vFlops *Vec_Ptr_t;
    var pState, pFlop *uint32;
    var i, k, nFlopWords int32;
    vFlops = Vec_PtrAlloc( 100 );
    nFlopWords = Abc_BitWordNum( 2*Vec_PtrSize(p.vStates) );
    for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
        if p.pCount0[i] == Vec_PtrSize(p.vStates) { 
            continue; }
        if p.pCountX[i] > 0 {
            continue; }
        pFlop = Gia_ManTerStateAlloc( nFlopWords );
        Vec_PtrPush( vFlops, pFlop );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(p.vStates)) && (((pState) = (unsigned *)Vec_PtrEntry(p.vStates, k)), 1); k++  {
            Gia_ManTerSimInfoSet( pFlop, k, Gia_ManTerSimInfoGet(pState, i) );
        }
//Gia_ManTerStatePrint( pFlop, Vec_PtrSize(p->vStates), i );
    }
    return vFlops;
}

/**Function*************************************************************

  Synopsis    [Transposes state vector for non-ternary registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFindEqualFlop(vFlops *Vec_Ptr_t, iFlop int32, nFlopWords int32) int32 {
    var pFlop, pTemp *uint32;
    var i int32;
    pFlop = (unsigned *)Vec_PtrEntry( vFlops, iFlop );
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < iFlop) && (((pTemp) = (unsigned *)Vec_PtrEntry(vFlops, i)), 1); i++  {
        if !memcmp( pTemp, pFlop, sizeof(unsigned) * nFlopWords ) {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Creates map of registers to replace.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerCreateMap(p *Gia_ManTer_t, fVerbose int32) *int32 {
    var pCi2Lit *int32;
    var pObj *Gia_Obj_t;
    var vMapKtoI *Vec_Int_t;
    var i, iRepr, nFlopWords, Counter0, CounterE int32;
    nFlopWords = Abc_BitWordNum( 2*Vec_PtrSize(p.vStates) );
    p.vFlops = Gia_ManTerTranspose( p );
    // inlining of macro ABC_FALLOC
    pCi2Lit = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManCiNum(p.pAig))), 0xff, sizeof(int) * (size_t)(Gia_ManCiNum(p.pAig))));
    vMapKtoI = Vec_IntAlloc( 100 );
    for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
        if p.pCount0[i] == Vec_PtrSize(p.vStates) {
            pCi2Lit[Gia_ManPiNum(p.pAig)+i] = 0, Counter0++;
        } else if p.pCountX[i] == 0 {
            iRepr = Gia_ManFindEqualFlop( p.vFlops, Vec_IntSize(vMapKtoI), nFlopWords );
            Vec_IntPush( vMapKtoI, i );
            if iRepr < 0 {
                continue; }
            pObj = Gia_ManCi( p.pAig, Gia_ManPiNum(p.pAig)+Vec_IntEntry(vMapKtoI, iRepr) );
            pCi2Lit[Gia_ManPiNum(p.pAig)+i] = Abc_Var2Lit( Gia_ObjId( p.pAig, pObj ), 0 );
            CounterE++;
        }
    }
    Vec_IntFree( vMapKtoI );
    if fVerbose != 0 {
        fmt.Printf("Transforming %d const and %d equiv registers.\n", Counter0, CounterE); }
    return pCi2Lit;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerSimulate(pAig *Gia_Man_t, fVerbose int32) *Gia_ManTer_t {
    var p *Gia_ManTer_t;
    var pState, pPrev, pLoop *uint32;
    var i, Counter int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    assert(Gia_ManRegNum(pAig) > 0, "Gia_ManRegNum(pAig) > 0");
    // create manager
    clk = Abc_Clock();
    p = Gia_ManTerCreate( pAig );
    if false {
        fmt.Printf("Obj = %8d (%8d). F = %6d. ", pAig.nObjs, Gia_ManCiNum(pAig) + Gia_ManAndNum(pAig), p.pAig.nFront);
        fmt.Printf("AIG = %7.2f MB. F-mem = %7.2f MB. Other = %7.2f MB.  ", 12.0*Gia_ManObjNum(p.pAig)/(1<<20), 4.0*Abc_BitWordNum(2 * p.pAig.nFront)/(1<<20), 4.0*Abc_BitWordNum(2 * (Gia_ManCiNum(pAig) + Gia_ManCoNum(pAig)))/(1<<20));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    // perform simulation
    Gia_ManTerSimInfoInit( p );
    // hash the first state
    pState = Gia_ManTerStateCreate( p );
    Gia_ManTerStateInsert( pState, p.nStateWords, p.pBins, p.nBins );
//Gia_ManTerStatePrint( pState, Gia_ManRegNum(pAig), 0 );
    // perform simuluation till convergence
    pPrev = nil;
    for  i = 0; ; i++  {
        Gia_ManTerSimulateRound( p );
        Gia_ManTerSimInfoTransfer( p );
        pState = Gia_ManTerStateCreate( p );
//Gia_ManTerStatePrint( pState, Gia_ManRegNum(pAig), i+1 );
        if pLoop = Gia_ManTerStateLookup(pState, p.nStateWords, p.pBins, p.nBins) {
            pAig.nTerStates = Vec_PtrSize( p.vStates );
            pAig.nTerLoop = Vec_PtrFind( p.vStates, pLoop );
            break;
        }
        Gia_ManTerStateInsert( pState, p.nStateWords, p.pBins, p.nBins );
        if i >= p.nIters && i % 10 == 0 {
            Counter = Gia_ManTerRetire( p, pState, pPrev );
//            Counter = Gia_ManTerRetire2( p, pState );
//            if ( fVerbose )
//                printf( "Retired %d registers.\n", Counter );
        }
        pPrev = pState;
    }
    if fVerbose != 0 {
        fmt.Printf("Ternary simulation saturated after %d iterations. ", i+1);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

