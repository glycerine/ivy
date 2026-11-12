package bsat

import "fmt"

/**CFile****************************************************************

  FileName    [satTruth.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Truth table computation package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satTruth.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

//#include "satTruth.h"
//#include "misc/vec/vecSet.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Tru_Man_t_ struct {
    nVars              int32;        // the number of variables
    nWords              int32;       // the number of words in the truth table
    nEntrySize              int32;   // the size of one entry in 'int'
    nTableSize              int32;   // hash table size
    pTable             *int32;       // hash table
    pMem       *Vec_Set_t;         // memory for truth tables
    pZero            *word;        // temporary truth table 
    hIthVars [16]int32; // variable handles
    nTableLookups              int32;
}

    // 16 bytes minimum
type Tru_One_t struct {
    Handle              int32;       // support
    Next              int32;         // next one in the table
    word             pTruth[0];    // truth table
}

func Tru_ManReadOne(p *Tru_Man_t, h int32) *Tru_One_t { if h != 0 {
    return (Tru_One_t *)Vec_SetEntry(p.pMem, h);
} else {
    return nil;
} }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the hash key.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tru_ManHash(pTruth *word, nWords int32, nBins int32, pPrimes *int32) uint32 {
    var i int32;
    var uHash uint32;
    for  i = 0; i < nWords; i++  {
        uHash ^= pTruth[i] * pPrimes[i & 0x7];
    }
    return uHash % nBins;
}

/**Function*************************************************************

  Synopsis    [Returns the given record.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Tru_ManLookup_s_Primes [10]int32 = [10]int32{ 1291, 1699, 2357, 4177, 5147, 5647, 6343, 7103, 7873, 8147 };
func Tru_ManLookup(p *Tru_Man_t, pTruth *word) *int32 {
    
    var pEntry *Tru_One_t;
    var pSpot *int32;
    assert((pTruth[0] & 1) == 0, "(pTruth[0] & 1) == 0");
    pSpot = p.pTable + Tru_ManHash( pTruth, p.nWords, p.nTableSize, spStatic_Tru_ManLookup_s_Primes );
    for  pEntry = Tru_ManReadOne(p, *pSpot); (pEntry != nil); pSpot = &pEntry.Next, pEntry = Tru_ManReadOne(p, *pSpot)  {
        if Tru_ManEqual(pEntry.pTruth, pTruth, p.nWords) {
            return pSpot; }
    }
    return pSpot;
}

/**Function*************************************************************

  Synopsis    [Returns the given record.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tru_ManResize(p *Tru_Man_t) {
    var pThis *Tru_One_t;
    var pTableOld, pSpot *int32;
    var nTableSizeOld, iNext, Counter, i int32;
    assert(p.pTable != nil, "p.pTable != nil");
    // replace the table
    pTableOld = p.pTable;
    nTableSizeOld = p.nTableSize;
    p.nTableSize = 2 * p.nTableSize + 1; 
    // inlining of macro ABC_CALLOC
    p.pTable = ((int *) calloc((size_t)(p.nTableSize), sizeof(int)));
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < nTableSizeOld; i++  {
    for  pThis = Tru_ManReadOne(p, pTableOld[i]),
          iNext = (pThis? pThis.Next : 0);  
          (pThis != nil);  pThis = Tru_ManReadOne(p, iNext),   
          iNext = (pThis? pThis.Next : 0)   {
        assert(pThis.Handle, "pThis.Handle");
        pThis.Next = 0;
        pSpot = Tru_ManLookup( p, pThis.pTruth );
        assert(*pSpot == 0, "*pSpot == 0"); // should not be there
        *pSpot = pThis.Handle;
        Counter++;
    }
    }
    assert(Counter == Vec_SetEntryNum(p.pMem), "Counter == Vec_SetEntryNum(p.pMem)");
    // inlining of macro ABC_FREE
    ((pTableOld) ? (free((char *) (pTableOld)), (pTableOld) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Adds entry to the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tru_ManInsert(p *Tru_Man_t, pTruth *word) int32 {
    var fCompl int32;
    var pSpot *int32;
    if Tru_ManEqual0(pTruth, p.nWords) {
        return 0; }
    if Tru_ManEqual1(pTruth, p.nWords) {
        return 1; }
    p.nTableLookups++;
    if Vec_SetEntryNum(p.pMem) > 2 * p.nTableSize {
        Tru_ManResize( p ); }
    fCompl = pTruth[0] & 1;
    if fCompl != 0 {  
        Tru_ManNot( pTruth, p.nWords ); }
    pSpot = Tru_ManLookup( p, pTruth );
    if *pSpot == 0 {
        var pEntry *Tru_One_t;
        *pSpot = Vec_SetAppend( p.pMem, nil, p.nEntrySize );
        assert((*pSpot & 1) == 0, "(*pSpot & 1) == 0");
        pEntry = Tru_ManReadOne( p, *pSpot );
        Tru_ManCopy( pEntry.pTruth, pTruth, p.nWords );
        pEntry.Handle = *pSpot;
        pEntry.Next = 0;
    }
    if fCompl != 0 {  
        Tru_ManNot( pTruth, p.nWords ); }
    return *pSpot ^ fCompl;
}

/**Function*************************************************************

  Synopsis    [Start the truth table logging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tru_ManAlloc(nVars int32) *Tru_Man_t {
    var Masks [6]word = [6]word{ 
        ABC_CONST(0xAAAAAAAAAAAAAAAA), 
        ABC_CONST(0xCCCCCCCCCCCCCCCC), 
        ABC_CONST(0xF0F0F0F0F0F0F0F0), 
        ABC_CONST(0xFF00FF00FF00FF00), 
        ABC_CONST(0xFFFF0000FFFF0000), 
        ABC_CONST(0xFFFFFFFF00000000), 
    };
    
    var i, w int32;
    assert(nVars > 0 && nVars <= 16, "nVars > 0 && nVars <= 16");
    // inlining of macro ABC_CALLOC
    p := &Tru_Man_t{};
    p.nVars      = nVars;
    if nVars < 6 {
        p.nWords = 1;
    } else {
        p.nWords = (1 << (nVars-6));
    }
    p.nEntrySize = (sizeof(Tru_One_t) + p.nWords * sizeof(word))/sizeof(int);
    p.nTableSize = 8147;
    // inlining of macro ABC_CALLOC
    p.pTable     = ((int *) calloc((size_t)(p.nTableSize), sizeof(int)));
    p.pMem       = Vec_SetAlloc( 16 );
    // initialize truth tables
    // inlining of macro ABC_ALLOC
    p.pZero = ((word *) malloc(sizeof(word) * (size_t)(p.nWords)));
    for  i = 0; i < nVars; i++  {
        for  w = 0; w < p.nWords; w++  {
            if i < 6 {
                p.pZero[w] = Masks[i];
            } else if w & (1 << (i-6)) {
                p.pZero[w] = ~(word)0;
            } else {
                p.pZero[w] = 0; }
        }
        p.hIthVars[i] = Tru_ManInsert( p, p.pZero );
        assert(!(i != 0) || p.hIthVars[i] > p.hIthVars[i-1], "!i || p.hIthVars[i] > p.hIthVars[i-1]");
    }
    Tru_ManClear( p.pZero, p.nWords );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stop the truth table logging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tru_ManFree(p *Tru_Man_t) {
    fmt.Printf("Lookups = %d. Entries = %d.\n", p.nTableLookups, Vec_SetEntryNum(p.pMem));
    Vec_SetFree( p.pMem );
    // inlining of macro ABC_FREE
    ((p.pZero) ? (free((char *) (p.pZero)), (p.pZero) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pTable) ? (free((char *) (p.pTable)), (p.pTable) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Returns elementary variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tru_ManVar(p *Tru_Man_t, v int32) *word {
    assert(v >= 0 && v < p.nVars, "v >= 0 && v < p.nVars");
    return Tru_ManReadOne( p, p.hIthVars[v] ).pTruth;
}

/**Function*************************************************************

  Synopsis    [Returns stored truth table]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tru_ManFunc(p *Tru_Man_t, h int32) *word {
    assert((h & 1) == 0, "(h & 1) == 0");
    if h == 0 {
        return p.pZero; }
    return Tru_ManReadOne( p, h ).pTruth;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

