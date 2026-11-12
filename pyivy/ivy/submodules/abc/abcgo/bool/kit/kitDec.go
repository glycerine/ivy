package kit

/**CFile****************************************************************

  FileName    [kitDec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Computation kit.]

  Synopsis    [Decomposition manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 18, 2009.]

  Revision    [$Id: kitDec.c,v 1.00 2006/12/06 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// decomposition manager
   
type Kit_ManDec_t struct {
    nVarsMax             int32;     // the max number of variables
    nWordsMax             int32;    // the max number of words
    vTruthVars      *Vec_Ptr_t;   // elementary truth tables
    vTruthNodes      *Vec_Ptr_t;  // internal truth tables
    // current problem
    nVarsIn             int32;      // the current number of variables
    vLutsIn      *Vec_Int_t;      // LUT truth tables
    vSuppIn      *Vec_Int_t;      // LUT supports
    ATimeIn [64]byte;  // variable arrival times
    // extracted information
    pTruthIn       *uint32;     // computed truth table
    pTruthOut       *uint32;    // computed truth table
    nVarsOut             int32;     // the current number of variables
    nWordsOut             int32;    // the current number of words
    Order [32]byte;    // new vars into old vars after supp minimization
    // computed information
    vLutsOut      *Vec_Int_t;     // problem decomposition
    vSuppOut      *Vec_Int_t;     // problem decomposition
    ATimeOut [64]byte; // variable arrival times
}

func Kit_DecOuputArrival(nVars int32, vLuts *Vec_Int_t, ATimes []byte) int32 { return ATimes[nVars + Vec_IntSize(vLuts) - 1]; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts Decmetry manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_ManDecStart(nVarsMax int32) *Kit_ManDec_t {
    var p *Kit_ManDec_t;
    assert(nVarsMax <= 20, "nVarsMax <= 20");
    p = new(Kit_ManDec_t);
    p.nVarsMax   = nVarsMax;
    p.nWordsMax  = Kit_TruthWordNum( p.nVarsMax );
    p.vTruthVars = Vec_PtrAllocTruthTables( p.nVarsMax );
    p.vTruthNodes = Vec_PtrAllocSimInfo( 64, p.nWordsMax );
    p.vLutsIn    = Vec_IntAlloc( 50 );
    p.vSuppIn    = Vec_IntAlloc( 50 );
    p.vLutsOut   = Vec_IntAlloc( 50 );
    p.vSuppOut   = Vec_IntAlloc( 50 );
    p.pTruthIn   = ABC_ALLOC( unsigned, p.nWordsMax );
    p.pTruthOut  = ABC_ALLOC( unsigned, p.nWordsMax );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops Decmetry manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_ManDecStop(p *Kit_ManDec_t) {
    ABC_FREE( p.pTruthIn );
    ABC_FREE( p.pTruthOut );
    Vec_IntFreeP( &p.vLutsIn );
    Vec_IntFreeP( &p.vSuppIn );
    Vec_IntFreeP( &p.vLutsOut );
    Vec_IntFreeP( &p.vSuppOut );
    Vec_PtrFreeP( &p.vTruthVars );
    Vec_PtrFreeP( &p.vTruthNodes );
    ABC_FREE( p );
}


/**Function*************************************************************

  Synopsis    [Deriving timing information for the decomposed structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_DecComputeOuputArrival(nVars int32, vSupps *Vec_Int_t, LutSize int32, ATimesIn []byte, ATimesOut []byte) int32 {
    var i, v, iVar, nLuts, Delay int32;
    nLuts = Vec_IntSize(vSupps) / LutSize;
    assert(nLuts > 0, "nLuts > 0");
    assert(Vec_IntSize(vSupps) % LutSize == 0, "Vec_IntSize(vSupps) % LutSize == 0");
    for  v = 0; v < nVars; v++  {
        ATimesOut[v] = ATimesIn[v];
    }
    for  v = 0; v < nLuts; v++  {
        Delay = 0;
        for  i = 0; i < LutSize; i++  {
            iVar = Vec_IntEntry( vSupps, v * LutSize + i );
            assert(iVar < nVars + v, "iVar < nVars + v");
            Delay = Abc_MaxInt( Delay, ATimesOut[iVar] );
        }
        ATimesOut[nVars + v] = Delay + 1;
    }
    return ATimesOut[nVars + nLuts - 1];
}

/**Function*************************************************************

  Synopsis    [Derives the truth table]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_DecComputeTruthOne(LutSize int32, pTruthLut *uint32, nVars int32, pTruths []*uint32, pTemp *uint32, pRes *uint32) {
    var i, v int32;
    Kit_TruthClear( pRes, nVars );
    for  i = 0; i < (1<<LutSize); i++  {
        if !Kit_TruthHasBit( pTruthLut, i ) {
            continue; }
        Kit_TruthFill( pTemp, nVars );
        for  v = 0; v < LutSize; v++  {
            if i & (1<<v) {
                Kit_TruthAnd( pTemp, pTemp, pTruths[v], nVars );
            } else {
                Kit_TruthSharp( pTemp, pTemp, pTruths[v], nVars ); }
        }
        Kit_TruthOr( pRes, pRes, pTemp, nVars );
    }
}

/**Function*************************************************************

  Synopsis    [Derives the truth table]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_DecComputeTruth(p *Kit_ManDec_t, nVars int32, vSupps *Vec_Int_t, LutSize int32, vLuts *Vec_Int_t, pRes *uint32) {
    var pResult *uint32;
    var pTruthLuts *uint32;
    var pTruths [17]*uint32;
    var nTruthLutWords, i, v, iVar, nLuts int32;
    nLuts = Vec_IntSize(vSupps) / LutSize;
    pTruthLuts = (unsigned *)Vec_IntArray( vLuts );
    nTruthLutWords = Kit_TruthWordNum( LutSize );
    assert(nLuts > 0, "nLuts > 0");
    assert(Vec_IntSize(vSupps) % LutSize == 0, "Vec_IntSize(vSupps) % LutSize == 0");
    assert(nLuts * nTruthLutWords == Vec_IntSize(vLuts), "nLuts * nTruthLutWords == Vec_IntSize(vLuts)");
    for  v = 0; v < nLuts; v++  {
        for  i = 0; i < LutSize; i++  {
            iVar = Vec_IntEntry( vSupps, v * LutSize + i );
            assert(iVar < nVars + v, "iVar < nVars + v");
            pTruths[i] = (iVar < nVars)? (unsigned *)Vec_PtrEntry(p.vTruthVars, iVar) : (unsigned *)Vec_PtrEntry(p.vTruthNodes, iVar-nVars);
        }
        if v == nLuts - 1 {
            pResult = pRes;
        } else {
            pResult = (unsigned *)Vec_PtrEntry(p.vTruthNodes, v);
        }
        Kit_DecComputeTruthOne( LutSize, pTruthLuts, nVars, pTruths, (unsigned *)Vec_PtrEntry(p.vTruthNodes, v+1), pResult );
        pTruthLuts += nTruthLutWords;
    }
}

/**Function*************************************************************

  Synopsis    [Derives the truth table]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_DecComputePattern(nVars int32, pTruth *uint32, LutSize int32, Pattern []int32) int32 {
    var nCofs int32  = (1 << LutSize);
    var i, k, nMyu int32;
    assert(LutSize <= 6, "LutSize <= 6");
    assert(LutSize < nVars, "LutSize < nVars");
    if nVars - LutSize <= 5 {
        var uCofs [64]uint32;
        var nBits int32  = (1 << (nVars - LutSize));
        for  i = 0; i < nCofs; i++  {
            uCofs[i] = (pTruth[(i*nBits)/32] >> ((i*nBits)%32)) & ((1<<nBits)-1);
        }
        for  i = 0; i < nCofs; i++  {
            for  k = 0; k < nMyu; k++  {
                if uCofs[i] == uCofs[k] {
                    Pattern[i] = k;
                    break;
                }
            }
            if k == i {
                Pattern[nMyu++] = i; }
        }
    
    } else {
        var puCofs [64]*uint32;
        var nWords int32  = (1 << (nVars - LutSize - 5));
        for  i = 0; i < nCofs; i++  {
            puCofs[i] = pTruth + nWords;
        }
        for  i = 0; i < nCofs; i++  {
            for  k = 0; k < nMyu; k++  {
                if Kit_TruthIsEqual( puCofs[i], puCofs[k], nVars - LutSize - 5 ) {
                    Pattern[i] = k;
                    break;
                }
            }
            if k == i {
                Pattern[nMyu++] = i; }
        }
    }
    return nMyu;
}

/**Function*************************************************************

  Synopsis    [Returns the number of shared variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_DecComputeShared_rec(Pattern []int32, Vars []int32, nVars int32, Shared []int32, iVarTry int32) int32 {
    var Pat0 [32]int32;
    var Pat1 [32]int32;
    var Shared0 [5]int32;
    var Shared1 [5]int32;
    var VarsNext [5]int32;
    var v int32;
    var u int32;
    var iVarsNext int32;
    var iPat0 int32;
    var iPat1 int32;
    var m int32;
    var nMints int32 = (1 << nVars);
    var nShared0, nShared1, nShared int32;
    for  v = iVarTry; v < nVars; v++  {
        iVarsNext = 0;
        for  u = 0; u < nVars; u++  {
            if u == v {
                VarsNext[iVarsNext++] = Vars[u]; }
        }
        iPat0 = iPat1 = 0;
        for  m = 0; m < nMints; m++  {
            if m & (1 << v) {
                Pat1[iPat1++] = m;
            } else {
                Pat0[iPat0++] = m; }
        }
        assert(iPat0 == nMints / 2, "iPat0 == nMints / 2");
        assert(iPat1 == nMints / 2, "iPat1 == nMints / 2");
        nShared0 = Kit_DecComputeShared_rec( Pat0, VarsNext, nVars-1, Shared0, v + 1 );
        if nShared0 == 0 {
            continue; }
        nShared1 = Kit_DecComputeShared_rec( Pat1, VarsNext, nVars-1, Shared1, v + 1 );
        if nShared1 == 0 {
            continue; }
        Shared[nShared++] = v;
        for  u = 0; u < nShared0; u++  {
        for  m = 0; m < nShared1; m++  {
            if Shared0[u] >= 0 && Shared1[m] >= 0 && Shared0[u] == Shared1[m] {
                Shared[nShared++] = Shared0[u];
                Shared0[u] = Shared1[m] = -1;
            }
        }
        }
        return nShared;
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns the number of shared variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_DecComputeShared(Pattern []int32, LutSize int32, Shared []int32) int32 {
    var i int32;
    var Vars [6]int32;
    assert(LutSize <= 6, "LutSize <= 6");
    for  i = 0; i < LutSize; i++  {
        Vars[i] = i;
    }
    return Kit_DecComputeShared_rec( Pattern, Vars, LutSize, Shared, 0 );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

