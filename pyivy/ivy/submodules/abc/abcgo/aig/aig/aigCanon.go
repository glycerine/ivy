package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigCanon.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Processing the library of semi-canonical AIGs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigCanon.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"
//#include "bool/kit/kit.h"
//#include "bool/bdc/bdc.h"
//#include "aig/ioa/ioa.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const RMAN_MAXVARS =  12
#define RMAX_MAXWORD  (RMAN_MAXVARS <= 5 ? 1 : (1 << (RMAN_MAXVARS - 5)))

   
type Aig_VSig_t struct {
    nOnes           int32;
    nCofOnes [RMAN_MAXVARS]int32;
}

   
type Aig_Tru_t struct {
    pNext    *Aig_Tru_t;
    Id           int32;    
    nVisits uint32; /* C bitfield: 27 bits; review packing and truncation */
    nVars uint32; /* C bitfield: 5 bits; review packing and truncation */
    pTruth [0]uint32;
}

   
type Aig_RMan_t struct {
    nVars           int32;       // the largest variable number
    pAig    *Aig_Man_t;        // recorded subgraphs
    // hash table
    nBins           int32;
    pBins   **Aig_Tru_t;
    nEntries           int32;
    pMemTrus *Aig_MmFlex_t;
    // bidecomposion
    pBidec    *Bdc_Man_t;
    // temporaries
    pTruthInit [RMAX_MAXWORD]uint32; // canonical truth table
    pTruth [RMAX_MAXWORD]uint32;     // current truth table
    pTruthC [RMAX_MAXWORD]uint32;    // canonical truth table
    pTruthTemp [RMAX_MAXWORD]uint32; // temporary truth table
    pMints [2*RMAN_MAXVARS]Aig_VSig_t;   // minterm count
    pPerm [RMAN_MAXVARS]byte;      // permutation
    pPermR [RMAN_MAXVARS]byte;     // reverse permutation
    // statistics
    nVarFuncs [RMAN_MAXVARS+1]int32;
    nTotal           int32;
    nTtDsd           int32;
    nTtDsdPart           int32;
    nTtDsdNot           int32;
    nUniqueVars           int32;
}

var s_pRMan *Aig_RMan_t = nil;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates recording manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Aig_RManStart_Pars Bdc_Par_t ;
var spStatic_Aig_RManStart_pPars *Bdc_Par_t = &spStatic_Aig_RManStart_Pars;
func Aig_RManStart() *Aig_RMan_t {
    
    var p *Aig_RMan_t;
    p = new(Aig_RMan_t);
    *p = Aig_RMan_t{};
    p.nVars = RMAN_MAXVARS;
    p.pAig  = Aig_ManStart( 1000000 );
    Aig_IthVar( p.pAig, p.nVars-1 );
    // create hash table
    p.nBins = Abc_PrimeCudd(5000);
    p.pBins = ABC_CALLOC( Aig_Tru_t *, p.nBins );
    p.pMemTrus = Aig_MmFlexStart();
    // bi-decomposition manager
    spStatic_Aig_RManStart_pPars.nVarsMax = p.nVars;
    spStatic_Aig_RManStart_pPars.fVerbose = 0;
    p.pBidec = Bdc_ManAlloc( spStatic_Aig_RManStart_pPars );
    return p;
}

/**Function*************************************************************

  Synopsis    [Returns the hash key.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManTableHash(pTruth *uint32, nVars int32, nBins int32, pPrimes *int32) int32 {
    var i int32;
    var nWords int32 = Kit_TruthWordNum( nVars );
    var uHash uint32;
    for  i = 0; i < nWords; i++  {
        uHash ^= pTruth[i] * pPrimes[i & 0xf];
    }
    return int32((uHash % nBins));
}

/**Function*************************************************************

  Synopsis    [Returns the given record.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Aig_RManTableLookup_s_Primes [16]int32 = [16]int32{ 
        1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 
        4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147 };
func Aig_RManTableLookup(p *Aig_RMan_t, pTruth *uint32, nVars int32) **Aig_Tru_t {
    
    var ppSpot **Aig_Tru_t;
    var pEntry *Aig_Tru_t;
    ppSpot = p.pBins + Aig_RManTableHash( pTruth, nVars, p.nBins, spStatic_Aig_RManTableLookup_s_Primes );
    for  pEntry = *ppSpot; (pEntry != nil); ppSpot = &pEntry.pNext, pEntry = pEntry.pNext  {
        if Kit_TruthIsEqual( pEntry.pTruth, pTruth, nVars ) {
            return ppSpot; }
    }
    return ppSpot;
}

/**Function*************************************************************

  Synopsis    [Find or add new entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManTableResize(p *Aig_RMan_t) {
    var pEntry, pNext *Aig_Tru_t;
    var pBinsOld, ppPlace **Aig_Tru_t;
    var nBinsOld, Counter, i int32;
    var clk abctime;
    assert(p.pBins != nil, "p.pBins != nil");
clk = Abc_Clock();
    // save the old Bins
    pBinsOld = p.pBins;
    nBinsOld = p.nBins;
    // get the new Bins
    p.nBins = Abc_PrimeCudd( 3 * nBinsOld ); 
    p.pBins = ABC_CALLOC( Aig_Tru_t *, p.nBins );
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < nBinsOld; i++  {
    for  pEntry = pBinsOld[i], pNext = pEntry? pEntry.pNext : nil; 
          (pEntry != nil); pEntry = pNext, pNext = pEntry? pEntry.pNext : nil  {
        // get the place where this entry goes in the Bins 
        ppPlace = Aig_RManTableLookup( p, pEntry.pTruth, pEntry.nVars );
        assert(*ppPlace == nil, "*ppPlace == nil"); // should not be there
        // add the entry to the list
        *ppPlace = pEntry;
        pEntry.pNext = nil;
        Counter++;
    }
    }
    assert(Counter == p.nEntries, "Counter == p.nEntries");
//    ABC_PRT( "Time", Abc_Clock() - clk );
    ABC_FREE( pBinsOld );
}

/**Function*************************************************************

  Synopsis    [Find or add new entry.]

  Description [Returns 1 if this is a new entry.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManTableFindOrAdd(p *Aig_RMan_t, pTruth *uint32, nVars int32) int32 {
    var ppSpot **Aig_Tru_t;
    var pEntry *Aig_Tru_t;
    var nBytes int32;
    ppSpot = Aig_RManTableLookup( p, pTruth, nVars );
    if *ppSpot {
        (*ppSpot).nVisits++;
        return 0;
    }
    nBytes = sizeof(Aig_Tru_t) + sizeof(unsigned) * Kit_TruthWordNum(nVars);
    if p.nEntries == 3*p.nBins {
        Aig_RManTableResize( p ); }
    pEntry = (Aig_Tru_t *)Aig_MmFlexEntryFetch( p.pMemTrus, nBytes );
    pEntry.Id = p.nEntries++;
    pEntry.nVars = nVars;
    pEntry.nVisits = 1;
    pEntry.pNext = nil;
    memcpy( pEntry.pTruth, pTruth, sizeof(unsigned) * Kit_TruthWordNum(nVars) );
    *ppSpot = pEntry;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Deallocates recording manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManStop(p *Aig_RMan_t) {
    var i int32;
    fmt.Printf("Total funcs    = %10d\n", p.nTotal);
    fmt.Printf("Full DSD funcs = %10d\n", p.nTtDsd);
    fmt.Printf("Part DSD funcs = %10d\n", p.nTtDsdPart);
    fmt.Printf("Non- DSD funcs = %10d\n", p.nTtDsdNot);
    fmt.Printf("Uniq-var funcs = %10d\n", p.nUniqueVars);
    fmt.Printf("Unique   funcs = %10d\n", p.nEntries);
    fmt.Printf("Distribution of functions:\n");
    for  i = 5; i <= p.nVars; i++  {
        fmt.Printf("%2d = %8d\n", i, p.nVarFuncs[i]);
    }
    Aig_MmFlexStop( p.pMemTrus, 0 );
    Aig_ManStop( p.pAig );
    Bdc_ManFree( p.pBidec );
    ABC_FREE( p.pBins );
    ABC_FREE( p );
}


/**Function*************************************************************

  Synopsis    [Stops recording.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManQuit() {
//    extern void Ioa_WriteAiger( Aig_Man_t * pMan, char * pFileName, int fWriteSymbols, int fCompact );
    var Buffer string // was [20]byte in C

    if s_pRMan == nil {
        return; }
    // dump the library file
    Buffer = fmt.Sprintf("aiglib%02d.aig", s_pRMan.nVars);
    Ioa_WriteAiger( s_pRMan.pAig, Buffer, 0, 1 );
    // quit the manager
    Aig_RManStop( s_pRMan );
    s_pRMan = nil;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if all variables are unique.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManPrintVarProfile(pTruth *uint32, nVars int32, pTruthAux *uint32) {
    var pStore2 [32]int32;
    var i int32;
    Kit_TruthCountOnesInCofsSlow( pTruth, nVars, pStore2, pTruthAux );
    for  i = 0; i < nVars; i++  {
        fmt.Printf("%2d/%2d ", pStore2[2*i], pStore2[2*i+1]);
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Sorts numbers in the increasing order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManSortNums(pArray *int32, nVars int32) {
    var i, j, best_i, tmp int32;
    for  i = 0; i < nVars-1; i++  {
        best_i = i;
        for  j = i+1; j < nVars; j++  {
            if pArray[j] > pArray[best_i] {
                best_i = j; }
        }
        tmp = pArray[i]; pArray[i] = pArray[best_i]; pArray[best_i] = tmp;
    }
}

/**Function*************************************************************

  Synopsis    [Prints signatures for all variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManPrintSigs(pSigs *Aig_VSig_t, nVars int32) {
    var v, i, k int32;
    for  v = 0; v < nVars; v++  {
        fmt.Printf("%2d : ", v);
        for  k = 0; k < 2; k++  {
            fmt.Printf("%5d  ", pSigs[2*v+k].nOnes);
            fmt.Printf("(");
            for  i = 0; i < nVars; i++  {
                fmt.Printf("%4d ", pSigs[2*v+k].nCofOnes[i]);
            }
            fmt.Printf(")  ");
        }
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Computes signatures for all variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManComputeVSigs(pTruth *uint32, nVars int32, pSigs *Aig_VSig_t, pAux *uint32) {
    var v int32;
    for  v = 0; v < nVars; v++  {
        Kit_TruthCofactor0New( pAux, pTruth, nVars, v );
        pSigs[2*v+0].nOnes = Kit_TruthCountOnes( pAux, nVars );
        Kit_TruthCountOnesInCofs0( pAux, nVars, pSigs[2*v+0].nCofOnes );
        Aig_RManSortNums( pSigs[2*v+0].nCofOnes, nVars );

        Kit_TruthCofactor1New( pAux, pTruth, nVars, v );
        pSigs[2*v+1].nOnes = Kit_TruthCountOnes( pAux, nVars );
        Kit_TruthCountOnesInCofs0( pAux, nVars, pSigs[2*v+1].nCofOnes );
        Aig_RManSortNums( pSigs[2*v+1].nCofOnes, nVars );
    }
}

/**Function*************************************************************

  Synopsis    [Computs signatures for all variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManCompareSigs(p0 *Aig_VSig_t, p1 *Aig_VSig_t, nVars int32) int32 {
//    return memcmp( p0, p1, sizeof(int) + sizeof(int) * nVars );
    return memcmp( p0, p1, sizeof(int) );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if all variables are unique.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManVarsAreUnique(pMints *Aig_VSig_t, nVars int32) int32 {
    var i int32;
    for  i = 0; i < nVars - 1; i++  {
        if Aig_RManCompareSigs( &pMints[2*i], &pMints[2*(i+1)], nVars ) == 0 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if all variables are unique.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManPrintUniqueVars(pMints *Aig_VSig_t, nVars int32) {
    var i int32;
    for  i = 0; i < nVars; i++  {
        if Aig_RManCompareSigs( &pMints[2*i], &pMints[2*i+1], nVars ) == 0 {
            fmt.Printf("=");
        } else { 
            fmt.Printf("x"); }
    }
    fmt.Printf("\n");

    fmt.Printf("0");
    for  i = 1; i < nVars; i++  {
        if Aig_RManCompareSigs( &pMints[2*(i-1)], &pMints[2*i], nVars ) == 0 {
            fmt.Printf("-");
        } else if i < 10 {
            fmt.Printf("%c", '0' + i);
        } else { 
            fmt.Printf("%c", 'A' + i-10); }
    }
    fmt.Printf("\n");
}
 
/**Function*************************************************************

  Synopsis    [Canonicize the truth table.]

  Description [Returns the phase. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManSemiCanonicize(pOut *uint32, pIn *uint32, nVars int32, pCanonPerm string, pSigs *Aig_VSig_t, fReturnIn int32) uint32 {
    var TempSig Aig_VSig_t;
    var i, Temp, fChange, Counter int32;
    var pTemp *uint32;
    var uCanonPhase uint32;
    // collect signatures 
    Aig_RManComputeVSigs( pIn, nVars, pSigs, pOut );
    // canonicize phase
    for  i = 0; i < nVars; i++  {
//        if ( pStore[2*i+0] <= pStore[2*i+1] )
        if Aig_RManCompareSigs( &pSigs[2*i+0], &pSigs[2*i+1], nVars ) <= 0 {
            continue; }
        uCanonPhase |= (1 << i);
        TempSig = pSigs[2*i+0];
        pSigs[2*i+0] = pSigs[2*i+1];
        pSigs[2*i+1] = TempSig;
        Kit_TruthChangePhase( pIn, nVars, i );
    }
    // permute
    Counter = 0;
    do {
        fChange = 0;
        for  i = 0; i < nVars-1; i++  {
//            if ( pStore[2*i] <= pStore[2*(i+1)] )
            if Aig_RManCompareSigs( &pSigs[2*i], &pSigs[2*(i+1)], nVars ) <= 0 {
                continue; }
            Counter++;
            fChange = 1;

            Temp = pCanonPerm[i];
            pCanonPerm[i] = pCanonPerm[i+1];
            pCanonPerm[i+1] = Temp;

            TempSig = pSigs[2*i];
            pSigs[2*i] = pSigs[2*(i+1)];
            pSigs[2*(i+1)] = TempSig;

            TempSig = pSigs[2*i+1];
            pSigs[2*i+1] = pSigs[2*(i+1)+1];
            pSigs[2*(i+1)+1] = TempSig;

            Kit_TruthSwapAdjacentVars( pOut, pIn, nVars, i );
            pTemp = pIn; pIn = pOut; pOut = pTemp;
        }
    } while ( fChange );

    // swap if it was moved an even number of times
    if fReturnIn ^ !(Counter & 1) {
        Kit_TruthCopy( pOut, pIn, nVars ); }
    return uCanonPhase;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_FunCopyHop(pObj *Bdc_Fun_t) *Aig_Obj_t { return Aig_NotCond( (Aig_Obj_t *)Bdc_FuncCopy(Bdc_Regular(pObj)), Bdc_IsComplement(pObj) );  }

/**Function*************************************************************

  Synopsis    [Records one function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManSaveOne(p *Aig_RMan_t, pTruth *uint32, nVars int32) {
    var i, nNodes, RetValue int32;
    var pFunc *Bdc_Fun_t;
    var pTerm *Aig_Obj_t;
    // perform decomposition
    RetValue = Bdc_ManDecompose( p.pBidec, pTruth, nil, nVars, nil, 1000 );
    if RetValue < 0 {
        fmt.Printf("Decomposition failed.\n");
        return;
    }
    // convert back into HOP
    Bdc_FuncSetCopy( Bdc_ManFunc( p.pBidec, 0 ), Aig_ManConst1(p.pAig) );
    for  i = 0; i < nVars; i++  {
        Bdc_FuncSetCopy( Bdc_ManFunc( p.pBidec, i+1 ), Aig_IthVar(p.pAig, i) );
    }
    nNodes = Bdc_ManNodeNum(p.pBidec);
    for  i = nVars + 1; i < nNodes; i++  {
        pFunc = Bdc_ManFunc( p.pBidec, i );
        Bdc_FuncSetCopy( pFunc, Aig_And( p.pAig, 
            Bdc_FunCopyHop(Bdc_FuncFanin0(pFunc)), 
            Bdc_FunCopyHop(Bdc_FuncFanin1(pFunc)) ) );
    }
    pTerm = Bdc_FunCopyHop( Bdc_ManRoot(p.pBidec) );
    pTerm = Aig_ObjCreateCo( p.pAig, pTerm );
//    assert( pTerm->fPhase == 0 );
}

/**Function*************************************************************

  Synopsis    [Records one function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_RManRecord(pTruth *uint32, nVarsInit int32) {
    var fVerify int32  = 1;
    var pNtk *Kit_DsdNtk_t;
    var pObj *Kit_DsdObj_t;
    var uPhaseC uint32;
    var i, nVars, nWords int32;
    var fUniqueVars int32;

    if nVarsInit > RMAN_MAXVARS {
        fmt.Printf("The number of variables in too large.\n");
        return;
    }

    if s_pRMan == nil {
        s_pRMan = Aig_RManStart(); }
    s_pRMan.nTotal++;
    // canonicize the function
    pNtk = Kit_DsdDecompose( pTruth, nVarsInit );
    pObj = Kit_DsdNonDsdPrimeMax( pNtk );
    if pObj == nil || pObj.nFans == 3 {
        s_pRMan.nTtDsd++;
        Kit_DsdNtkFree( pNtk );
        return;
    }
    nVars = pObj.nFans;
    s_pRMan.nVarFuncs[nVars]++;
    if nVars < nVarsInit {
        s_pRMan.nTtDsdPart++;
    } else {
        s_pRMan.nTtDsdNot++; }
    // compute the number of words
    nWords = Abc_TruthWordNum( nVars );
    // copy the function
    memcpy( s_pRMan.pTruthInit, Kit_DsdObjTruth(pObj), (size_t)(4*nWords) );
    Kit_DsdNtkFree( pNtk );
    // canonicize the output
    if s_pRMan.pTruthInit[0] & 1 {
        Kit_TruthNot( s_pRMan.pTruthInit, s_pRMan.pTruthInit, nVars ); }
    memcpy( s_pRMan.pTruth, s_pRMan.pTruthInit, 4*nWords );

    // canonize the function
    for  i = 0; i < nVars; i++  {
        s_pRMan.pPerm[i] = i;
    }
    uPhaseC = Aig_RManSemiCanonicize( s_pRMan.pTruthTemp, s_pRMan.pTruth, nVars, s_pRMan.pPerm, s_pRMan.pMints, 1 );
    // check unique variables
    fUniqueVars = Aig_RManVarsAreUnique( s_pRMan.pMints, nVars );
    s_pRMan.nUniqueVars += fUniqueVars;

/*
    printf( "%4d : ", s_pRMan->nTotal );
    printf( "%2d %2d  ", nVarsInit, nVars );
    Extra_PrintBinary( stdout, &uPhaseC, nVars );
    printf( "  " );
    for ( i = 0; i < nVars; i++ )
        printf( "%2d/%2d ", s_pRMan->pMints[2*i], s_pRMan->pMints[2*i+1] );
    printf( "\n" );
    Aig_RManPrintUniqueVars( s_pRMan->pMints, nVars );
Extra_PrintBinary( stdout, s_pRMan->pTruth, 1<<nVars ); printf( "\n\n" );
*/
/*
    printf( "\n" );
    printf( "%4d : ", s_pRMan->nTotal );
    printf( "%2d %2d  ", nVarsInit, nVars );
    printf( "   " );
    printf( "\n" );
    Aig_RManPrintUniqueVars( s_pRMan->pMints, nVars );
//    Aig_RManPrintSigs( s_pRMan->pMints, nVars );
*/

//Extra_PrintBinary( stdout, s_pRMan->pTruth, 1<<nVars ); printf( "\n\n" );

    if Aig_RManTableFindOrAdd( s_pRMan, s_pRMan.pTruth, nVars ) != 0 {
        Aig_RManSaveOne( s_pRMan, s_pRMan.pTruth, nVars ); }

    if fVerify != 0 {
        // derive reverse permutation
        for  i = 0; i < nVars; i++  {
            s_pRMan.pPermR[i] = s_pRMan.pPerm[i];
        }
        // implement permutation
        Kit_TruthPermute( s_pRMan.pTruthTemp, s_pRMan.pTruth, nVars, s_pRMan.pPermR, 1 );
        // implement polarity
        for  i = 0; i < nVars; i++  {
            if uPhaseC & (1 << i) {
                Kit_TruthChangePhase( s_pRMan.pTruth, nVars, i ); }
        }

        // perform verification
        if (fUniqueVars != 0) && !Kit_TruthIsEqual( s_pRMan.pTruth, s_pRMan.pTruthInit, nVars ) {
            fmt.Printf("Verification failed.\n"); }
    }
//Aig_RManPrintVarProfile( s_pRMan->pTruth, nVars, s_pRMan->pTruthTemp );
//Extra_PrintBinary( stdout, s_pRMan->pTruth, 1<<nVars ); printf( "\n" );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

