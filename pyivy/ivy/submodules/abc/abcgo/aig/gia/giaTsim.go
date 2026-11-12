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

func Gia_ManTerSimInfoGet(pInfo *uint32, i int32) int32 {
    return 3 & (pInfo[i >> 4] >> ((i & 15) << 1));
}
func Gia_ManTerSimInfoSet(pInfo *uint32, i int32, Value int32) {
    assert(Value >= GIA_ZER && Value <= GIA_UND, "Value >= GIA_ZER && Value <= GIA_UND");
    Value ^= Gia_ManTerSimInfoGet( pInfo, i );
    pInfo[i >> 4] ^= (Value << ((i & 15) << 1));
}

func Gia_ManTerStateNext(pState *uint32, nWords int32) *uint32 { return *((unsigned **)(pState + nWords));  }
func Gia_ManTerStateSetNext(pState *uint32, nWords int32, pNext *uint32) { *((unsigned **)(pState + nWords)) = pNext; }

// ternary simulation manager
   
type Gia_ManTer_t struct {
    pAig     *Gia_Man_t;
    nIters            int32;
    nStateWords            int32;
    vStates     *Vec_Ptr_t;
    vFlops     *Vec_Ptr_t;
    vRetired     *Vec_Int_t;     // retired registers
    pRetired          string;     // retired registers
    pCount0           *int32;
    pCountX           *int32;
    // hash table for states
    nBins            int32;
    pBins     **uint32;
    // simulation information
    pDataSim      *uint32;     // simulation data
    pDataSimCis      *uint32;  // simulation data for CIs
    pDataSimCos      *uint32;  // simulation data for COs
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerCreate(pAig *Gia_Man_t) *Gia_ManTer_t {
    
    // inlining of macro ABC_CALLOC
    p := &Gia_ManTer_t{};
    p.pAig   = Gia_ManFront( pAig );
    p.nIters = 300;
    // inlining of macro ABC_ALLOC
    p.pDataSim    = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Abc_BitWordNum(2*p.pAig.nFront))));
    // inlining of macro ABC_ALLOC
    p.pDataSimCis = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Abc_BitWordNum(2*Gia_ManCiNum(p.pAig)))));
    // inlining of macro ABC_ALLOC
    p.pDataSimCos = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Abc_BitWordNum(2*Gia_ManCoNum(p.pAig)))));
    // allocate storage for terminary states
    p.nStateWords = Abc_BitWordNum( 2*Gia_ManRegNum(pAig) );
    p.vStates  = Vec_PtrAlloc( 1000 );
    // inlining of macro ABC_CALLOC
    p.pCount0  = ((int *) calloc((size_t)(Gia_ManRegNum(pAig)), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pCountX  = ((int *) calloc((size_t)(Gia_ManRegNum(pAig)), sizeof(int)));
    p.nBins    = Abc_PrimeCudd( 500 );
    // inlining of macro ABC_CALLOC
    p.pBins    = ((unsigned * *) calloc((size_t)(p.nBins), sizeof(unsigned *)));
    p.vRetired = Vec_IntAlloc( 100 );
    // inlining of macro ABC_CALLOC
    p.pRetired = ((char *) calloc((size_t)(Gia_ManRegNum(pAig)), sizeof(char)));
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerStatesFree(vStates *Vec_Ptr_t) {
    var pTemp *uint32;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vStates)) && (((pTemp) = (unsigned *)Vec_PtrEntry(vStates, i)), 1); i++  {
        // inlining of macro ABC_FREE
        ((pTemp) ? (free((char *) (pTemp)), (pTemp) = 0) : 0);
    }
    Vec_PtrFree( vStates );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerDelete(p *Gia_ManTer_t) {
    if p.vStates { 
        Gia_ManTerStatesFree( p.vStates ); }
    if p.vFlops { 
        Gia_ManTerStatesFree( p.vFlops ); }
    Gia_ManStop( p.pAig );
    Vec_IntFree( p.vRetired );
    // inlining of macro ABC_FREE
    ((p.pRetired) ? (free((char *) (p.pRetired)), (p.pRetired) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pCount0) ? (free((char *) (p.pCount0)), (p.pCount0) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pCountX) ? (free((char *) (p.pCountX)), (p.pCountX) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pBins) ? (free((char *) (p.pBins)), (p.pBins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pDataSim) ? (free((char *) (p.pDataSim)), (p.pDataSim) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pDataSimCis) ? (free((char *) (p.pDataSimCis)), (p.pDataSimCis) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pDataSimCos) ? (free((char *) (p.pDataSimCos)), (p.pDataSimCos) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerSimulateCi(p *Gia_ManTer_t, pObj *Gia_Obj_t, iCi int32) {
    Gia_ManTerSimInfoSet( p.pDataSim, Gia_ObjValue(pObj), Gia_ManTerSimInfoGet(p.pDataSimCis, iCi) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerSimulateCo(p *Gia_ManTer_t, iCo int32, pObj *Gia_Obj_t) {
    var Value int32  = Gia_ManTerSimInfoGet( p.pDataSim, Gia_ObjDiff0(pObj) );
    Gia_ManTerSimInfoSet( p.pDataSimCos, iCo, Gia_XsimNotCond( Value, Gia_ObjFaninC0(pObj) ) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerSimulateNode(p *Gia_ManTer_t, pObj *Gia_Obj_t) {
    var Value0 int32  = Gia_ManTerSimInfoGet( p.pDataSim, Gia_ObjDiff0(pObj) );
    var Value1 int32  = Gia_ManTerSimInfoGet( p.pDataSim, Gia_ObjDiff1(pObj) );
    Gia_ManTerSimInfoSet( p.pDataSim, Gia_ObjValue(pObj), Gia_XsimAndCond( Value0, Gia_ObjFaninC0(pObj), Value1, Gia_ObjFaninC1(pObj) ) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerSimInfoInit(p *Gia_ManTer_t) {
    var i int32;
    for  ; i < Gia_ManPiNum(p.pAig); i++  {
        Gia_ManTerSimInfoSet( p.pDataSimCis, i, GIA_UND );
    }
    for  ; i < Gia_ManCiNum(p.pAig); i++  {
        Gia_ManTerSimInfoSet( p.pDataSimCis, i, GIA_ZER );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTerSimInfoTransfer(p *Gia_ManTer_t) {
    var i int32;
    for  ; i < Gia_ManPiNum(p.pAig); i++  {
        Gia_ManTerSimInfoSet( p.pDataSimCis, i, GIA_UND );
    }
    for  ; i < Gia_ManCiNum(p.pAig); i++  {
        Gia_ManTerSimInfoSet( p.pDataSimCis, i, Gia_ManTerSimInfoGet( p.pDataSimCos, Gia_ManCoNum(p.pAig)-Gia_ManCiNum(p.pAig)+i ) );
    }
}


/**Function*************************************************************

  Synopsis    [Computes hash value of the node using its simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Gia_ManTerStateHash_s_FPrimes [128]int32 = [128]int32{ 
        1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459, 
        1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997, 
        2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543, 
        2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089, 
        3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671, 
        3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243, 
        4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871, 
        4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471, 
        5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073, 
        6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689, 
        6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309, 
        7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933, 
        8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
    };
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
func Gia_ManTerSimulateRound(p *Gia_ManTer_t) {
    var pObj *Gia_Obj_t;
    var i, iCis, iCos int32;
    assert(p.pAig.nFront > 0, "p.pAig.nFront > 0");
    assert(Gia_ManConst0(p.pAig).Value == 0, "Gia_ManConst0(p.pAig).Value == 0");
    Gia_ManTerSimInfoSet( p.pDataSim, 0, GIA_ZER );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.pAig.nObjs) && ((pObj) = Gia_ManObj(p.pAig, i)); i++  {
        if Gia_ObjIsAndOrConst0(pObj) {
            assert(Gia_ObjValue(pObj) < p.pAig.nFront, "Gia_ObjValue(pObj) < p.pAig.nFront");
            Gia_ManTerSimulateNode( p, pObj );
        
        } else if Gia_ObjIsCi(pObj) {
            assert(Gia_ObjValue(pObj) < p.pAig.nFront, "Gia_ObjValue(pObj) < p.pAig.nFront");
            Gia_ManTerSimulateCi( p, pObj, iCis++ );
        
        } else { // if ( Gia_ObjIsCo(pObj) )
        
            assert(Gia_ObjValue(pObj) == GIA_NONE, "Gia_ObjValue(pObj) == GIA_NONE");
            Gia_ManTerSimulateCo( p, iCos++, pObj );
        }
    }
    assert(Gia_ManCiNum(p.pAig) == iCis, "Gia_ManCiNum(p.pAig) == iCis");
    assert(Gia_ManCoNum(p.pAig) == iCos, "Gia_ManCoNum(p.pAig) == iCos");
}

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
func Gia_ManReduceConst(pAig *Gia_Man_t, fVerbose int32) *Gia_Man_t {
    var p *Gia_ManTer_t;
    var pNew *Gia_Man_t  = nil;
    var pCi2Lit *int32;
    p = Gia_ManTerSimulate( pAig, fVerbose );
    Gia_ManTerAnalyze( p );
    pCi2Lit = Gia_ManTerCreateMap( p, fVerbose );
    Gia_ManTerDelete( p );
    pNew = Gia_ManDupDfsCiMap( pAig, pCi2Lit, nil );
    // inlining of macro ABC_FREE
    ((pCi2Lit) ? (free((char *) (pCi2Lit)), (pCi2Lit) = 0) : 0);
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

