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

