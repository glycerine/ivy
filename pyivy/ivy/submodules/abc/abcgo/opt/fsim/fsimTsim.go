package fsim

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [fsimTsim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast sequential AIG simulator.]

  Synopsis    [Varius utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: fsimTsim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fsimInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const FSIM_ZER = 1
const FSIM_ONE = 2
const FSIM_UND = 3

func Aig_XsimNotCond(Value int32, fCompl int32) int32 { 
    if Value == FSIM_UND {
        return FSIM_UND; }
    if Value == FSIM_ZER + fCompl {
        return FSIM_ZER; }
    return FSIM_ONE;
}
func Aig_XsimAndCond(Value0 int32, fCompl0 int32, Value1 int32, fCompl1 int32) int32 { 
    if Value0 == FSIM_UND || Value1 == FSIM_UND {
        return FSIM_UND; }
    if Value0 == FSIM_ZER + fCompl0 || Value1 == FSIM_ZER + fCompl1 {
        return FSIM_ZER; }
    return FSIM_ONE;
}

func Fsim_ManTerSimInfoGet(pInfo *uint32, i int32) int32 {
    return 3 & (pInfo[i >> 4] >> ((i & 15) << 1));
}
func Fsim_ManTerSimInfoSet(pInfo *uint32, i int32, Value int32) {
    assert(Value >= FSIM_ZER && Value <= FSIM_UND, "Value >= FSIM_ZER && Value <= FSIM_UND");
    Value ^= Fsim_ManTerSimInfoGet( pInfo, i );
    pInfo[i >> 4] ^= (Value << ((i & 15) << 1));
}

func Fsim_ManTerStateNext(pState *uint32, nWords int32) *uint32 { return *((unsigned **)(pState + nWords));  }
func Fsim_ManTerStateSetNext(pState *uint32, nWords int32, pNext *uint32) { *((unsigned **)(pState + nWords)) = pNext; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTerSimulateCi(p *Fsim_Man_t, iNode int32, iCi int32) {
    Fsim_ManTerSimInfoSet( p.pDataSim, iNode, Fsim_ManTerSimInfoGet(p.pDataSimCis, iCi) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTerSimulateCo(p *Fsim_Man_t, iCo int32, iFan0 int32) {
    var Value int32  = Fsim_ManTerSimInfoGet( p.pDataSim, Fsim_Lit2Var(iFan0) );
    Fsim_ManTerSimInfoSet( p.pDataSimCos, iCo, Aig_XsimNotCond( Value, Fsim_LitIsCompl(iFan0) ) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTerSimulateNode(p *Fsim_Man_t, iNode int32, iFan0 int32, iFan1 int32) {
    var Value0 int32  = Fsim_ManTerSimInfoGet( p.pDataSim, Fsim_Lit2Var(iFan0) );
    var Value1 int32  = Fsim_ManTerSimInfoGet( p.pDataSim, Fsim_Lit2Var(iFan1) );
    Fsim_ManTerSimInfoSet( p.pDataSim, iNode, Aig_XsimAndCond( Value0, Fsim_LitIsCompl(iFan0), Value1, Fsim_LitIsCompl(iFan1) ) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTerSimInfoInit(p *Fsim_Man_t) {
    var iPioNum, i int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCis2Ids)) && (((iPioNum) = Vec_IntEntry(p.vCis2Ids, i)), 1); i++  {
        if iPioNum < p.nPis {
            Fsim_ManTerSimInfoSet( p.pDataSimCis, i, FSIM_UND );
        } else {
            Fsim_ManTerSimInfoSet( p.pDataSimCis, i, FSIM_ZER ); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTerSimInfoTransfer(p *Fsim_Man_t) {
    var iPioNum, i int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCis2Ids)) && (((iPioNum) = Vec_IntEntry(p.vCis2Ids, i)), 1); i++  {
        if iPioNum < p.nPis {
            Fsim_ManTerSimInfoSet( p.pDataSimCis, i, FSIM_UND );
        } else {
            Fsim_ManTerSimInfoSet( p.pDataSimCis, i, Fsim_ManTerSimInfoGet( p.pDataSimCos, p.nPos+iPioNum-p.nPis ) ); }
    }
}


/**Function*************************************************************

  Synopsis    [Computes hash value of the node using its simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Fsim_ManTerStateHash_s_FPrimes [128]int32 = [128]int32{ 
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
func Fsim_ManTerStateHash(pState *uint32, nWords int32, nTableSize int32) int32 {
    
    var uHash uint32;
    var i int32;
    uHash = 0;
    for  i = 0; i < nWords; i++  {
        uHash ^= pState[i] * spStatic_Fsim_ManTerStateHash_s_FPrimes[i & 0x7F];
    }
    return uHash % nTableSize;
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTerStateLookup(pState *uint32, nWords int32, pBins **uint32, nBins int32) int32 {
    var pEntry *uint32;
    var Hash int32;
    Hash = Fsim_ManTerStateHash( pState, nWords, nBins );
    for  pEntry = pBins[Hash]; (pEntry != nil); pEntry = Fsim_ManTerStateNext(pEntry, nWords)  {
        if !memcmp( pEntry, pState, sizeof(unsigned) * nWords ) {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTerStateInsert(pState *uint32, nWords int32, pBins **uint32, nBins int32) {
    var Hash int32  = Fsim_ManTerStateHash( pState, nWords, nBins );
    assert(!(Fsim_ManTerStateLookup( pState, nWords, pBins, nBins ) != 0), "!Fsim_ManTerStateLookup( pState, nWords, pBins, nBins )");
    Fsim_ManTerStateSetNext( pState, nWords, pBins[Hash] );
    pBins[Hash] = pState;    
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTerStateCreate(pInfo *uint32, nPis int32, nCis int32, nWords int32) *uint32 {
    var pRes *uint32;
    var i int32;
    // inlining of macro ABC_CALLOC
    pRes = (unsigned *)((char *) calloc((size_t)(sizeof(unsigned) * nWords + sizeof(unsigned *)), sizeof(char)));
    for  i = nPis; i < nCis; i++  {
        Fsim_ManTerSimInfoSet( pRes, i-nPis, Fsim_ManTerSimInfoGet(pInfo, i) );
    }
    return pRes;    
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTerStatePrint(pState *uint32, nRegs int32) {
    var i, Value, nZeros, nOnes, nDcs int32;
    for  i = 0; i < nRegs; i++  {
        Value = (Aig_InfoHasBit( pState, 2 * i + 1 ) << 1) | Aig_InfoHasBit( pState, 2 * i );
        if Value == 1 {
            fmt.Printf("0");
nZeros++;
        } else if Value == 2 {
            fmt.Printf("1");
nOnes++;
        } else if Value == 3 {
            fmt.Printf("x");
nDcs++;
        } else {
            assert(false, "0"); }
    }
    fmt.Printf(" (0=%5d, 1=%5d, x=%5d)\n", nZeros, nOnes, nDcs);
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTerSimulateRound(p *Fsim_Man_t) {
    var pCur, pEnd *int32;
    var iCis, iCos int32;
    if Aig_ObjRefs(Aig_ManConst1(p.pAig)) {
        Fsim_ManTerSimInfoSet( p.pDataSimCis, 1, FSIM_ONE ); }
    pCur = p.pDataAig2 + 6;
    pEnd = p.pDataAig2 + 3 * p.nObjs;
    for  pCur < pEnd  {
        if pCur[1] == 0 {
            Fsim_ManTerSimulateCi( p, pCur[0], iCis++ );
        } else if pCur[2] == 0 {
            Fsim_ManTerSimulateCo( p, iCos++, pCur[1] );
        } else {
            Fsim_ManTerSimulateNode( p, pCur[0], pCur[1], pCur[2] ); }
        pCur += 3;
    }
    assert(iCis == p.nCis, "iCis == p.nCis");
    assert(iCos == p.nCos, "iCos == p.nCos");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTerSimulate(pAig *Aig_Man_t, fVerbose int32) *Vec_Ptr_t {
    var p *Fsim_Man_t;
    var vStates *Vec_Ptr_t;
    var pBins **uint32;
    var pState *uint32;
    var i, nWords, nBins int32;
    var clk int64;
    var clkTotal int64 = time.Now().UnixNano();
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    // create manager
    clk = time.Now().UnixNano();
    p = Fsim_ManCreate( pAig );
    if fVerbose != 0 {
        fmt.Printf("Obj = %8d (%8d). Cut = %6d. Front = %6d.  FrtMem = %7.2f MB. ", p.nObjs, p.nCis + p.nNodes, p.nCrossCutMax, p.nFront, 4.0*Aig_BitWordNum(2 * p.nFront)/(1<<20));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
    }
    // create simulation frontier
    clk = time.Now().UnixNano();
    Fsim_ManFront( p, 0 );
    if fVerbose != 0 {
        fmt.Printf("Max ID = %8d. Log max ID = %2d.  AigMem = %7.2f MB (%5.2f byte/obj).  ", p.iNumber, Aig_Base2Log(p.iNumber), 1.0*(p.pDataCur-p.pDataAig)/(1<<20), 1.0*(p.pDataCur-p.pDataAig)/p.nObjs); 
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
    }
    // allocate storage for terminary states
    nWords  = Abc_BitWordNum( 2*Aig_ManRegNum(pAig) );
    vStates = Vec_PtrAlloc( 1000 );
    nBins   = Abc_PrimeCudd( 500 );
    // inlining of macro ABC_ALLOC
    pBins   = ((unsigned * *) malloc(sizeof(unsigned *) * (size_t)(nBins)));
    memset( pBins, 0, sizeof(unsigned *) * nBins );
    // perform simulation
    assert(p.pDataSim == nil, "p.pDataSim == nil");
    // inlining of macro ABC_ALLOC
    p.pDataSim = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Aig_BitWordNum(2 * p.nFront) * sizeof(unsigned))));
    // inlining of macro ABC_ALLOC
    p.pDataSimCis = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Aig_BitWordNum(2 * p.nCis) * sizeof(unsigned))));
    // inlining of macro ABC_ALLOC
    p.pDataSimCos = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Aig_BitWordNum(2 * p.nCos) * sizeof(unsigned))));
    Fsim_ManTerSimInfoInit( p );
    // hash the first state
    pState = Fsim_ManTerStateCreate( p.pDataSimCis, p.nPis, p.nCis, nWords );
    Vec_PtrPush( vStates, pState );
    Fsim_ManTerStateInsert( pState, nWords, pBins, nBins );
    // perform simuluation till convergence
    for  i = 0; ; i++  {
        Fsim_ManTerSimulateRound( p );
        Fsim_ManTerSimInfoTransfer( p );
        // hash the first state
        pState = Fsim_ManTerStateCreate( p.pDataSimCis, p.nPis, p.nCis, nWords );
        Vec_PtrPush( vStates, pState );
        if Fsim_ManTerStateLookup(pState, nWords, pBins, nBins) != 0 {
            break; }
        Fsim_ManTerStateInsert( pState, nWords, pBins, nBins );
    }
    if fVerbose != 0 {
        fmt.Printf("Maxcut = %8d.  AigMem = %7.2f MB.  SimMem = %7.2f MB.  ", p.nCrossCutMax, p.pDataAig2? 12.0*p.nObjs/(1<<20) : 1.0*(p.pDataCur-p.pDataAig)/(1<<20), 4.0*(Aig_BitWordNum(2 * p.nFront)+Aig_BitWordNum(2 * p.nCis)+Aig_BitWordNum(2 * p.nCos))/(1<<20));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Sim time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clkTotal)))/(float64(int64(time.Second))));;
    }
    // inlining of macro ABC_FREE
    ((pBins) ? (free((char *) (pBins)), (pBins) = 0) : 0);
    Fsim_ManDelete( p );
    return vStates;

}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

