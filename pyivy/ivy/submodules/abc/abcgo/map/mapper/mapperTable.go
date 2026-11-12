package mapper

import "fmt"

/**CFile****************************************************************

  FileName    [mapperTable.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperTable.c,v 1.6 2005/01/23 06:59:44 alanmi Exp $]

***********************************************************************/

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the table function for the tables
// Original macro:
// #define  MAP_TABLE_HASH(u1,u2,nSize)  (((u1) + 2003 * (u2)) % nSize)


// static void Map_SuperTableResize( Map_HashTable_t * pLib );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates the hash table for supergates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperTableCreate(pLib *Map_SuperLib_t) *Map_HashTable_t {
    var p *Map_HashTable_t;
    // allocate the table
    // inlining of macro ABC_ALLOC
    p = ((Map_HashTable_t *) malloc(sizeof(Map_HashTable_t) * (size_t)(1)));
    memset( p, 0, sizeof(Map_HashTable_t) );
    p.mmMan = pLib.mmEntries;
    // allocate and clean the bins
    p.nBins = Abc_PrimeCudd(20000);
    // inlining of macro ABC_ALLOC
    p.pBins = ((Map_HashEntry_t * *) malloc(sizeof(Map_HashEntry_t *) * (size_t)(p.nBins)));
    memset( p.pBins, 0, sizeof(Map_HashEntry_t *) * p.nBins );
    return p;
}


/**Function*************************************************************

  Synopsis    [Deallocates the supergate hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperTableFree(p *Map_HashTable_t) {
    p.pBins = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Inserts a new entry into the hash table.]

  Description [This function inserts the new gate (pGate), which will be
  accessible through its canonical form (uTruthC).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperTableInsertC(p *Map_HashTable_t, uTruthC []uint32, pGate *Map_Super_t) int32 {
    var pEnt *Map_HashEntry_t;
    var Key uint32;
    // resize the table
    if p.nEntries >= 2 * p.nBins {
        Map_SuperTableResize( p ); }
    // check if another supergate with the same canonical form exists
    // inlining of macro MAP_TABLE_HASH
    Key = (((uTruthC[0]) + 2003 * (uTruthC[1])) % p.nBins);
    for  pEnt = p.pBins[Key]; (pEnt != nil); pEnt = pEnt.pNext  {
        if pEnt.uTruth[0] == uTruthC[0] && pEnt.uTruth[1] == uTruthC[1] {
            break; }
    }
    // create a new entry if it does not exist
    if pEnt == nil {
        // add the new entry to the table
        pEnt = (Map_HashEntry_t *)Extra_MmFixedEntryFetch( p.mmMan );
        memset( pEnt, 0, sizeof(Map_HashEntry_t) );
        pEnt.uTruth[0] = uTruthC[0];
        pEnt.uTruth[1] = uTruthC[1];
        // add the hash table entry to the corresponding linked list in the table
        pEnt.pNext   = p.pBins[Key];
        p.pBins[Key] = pEnt;
        p.nEntries++;
    }
    // add the supergate to the entry
    pGate.pNext = pEnt.pGates;
    pEnt.pGates = pGate;
    return 0;
}



/**Function*************************************************************

  Synopsis    [Inserts a new entry into the library.]

  Description [This function inserts the new gate (pGate), which will be
  accessible through its unfolded function (uTruth).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperTableInsert(p *Map_HashTable_t, uTruth []uint32, pGate *Map_Super_t, uPhase uint32) int32 {
    var pEnt *Map_HashEntry_t;
    var Key uint32;
    // resize the table
    if p.nEntries >= 2 * p.nBins {
        Map_SuperTableResize( p ); }
    // check if this entry already exists
    // inlining of macro MAP_TABLE_HASH
    Key = (((uTruth[0]) + 2003 * (uTruth[1])) % p.nBins);
    for  pEnt = p.pBins[Key]; (pEnt != nil); pEnt = pEnt.pNext  {
        if pEnt.uTruth[0] == uTruth[0] && pEnt.uTruth[1] == uTruth[1] {
            return 1; }
    }
    // add the new hash table entry to the table
    pEnt = (Map_HashEntry_t *)Extra_MmFixedEntryFetch( p.mmMan );
    memset( pEnt, 0, sizeof(Map_HashEntry_t) );
    pEnt.uTruth[0] = uTruth[0];
    pEnt.uTruth[1] = uTruth[1];
    pEnt.pGates    = pGate;
    pEnt.uPhase    = uPhase;
    // add the hash table to the corresponding linked list in the table
    pEnt.pNext   = p.pBins[Key];
    p.pBins[Key] = pEnt;
    p.nEntries++;
/*
printf( "Adding gate: %10u ", Key );
Map_LibraryPrintSupergate( pGate );
Extra_PrintBinary( stdout, uTruth, 32 );
printf( "\n" );
*/
    return 0;
}

/**Function*************************************************************

  Synopsis    [Looks up an entry in the library.]

  Description [This function looks up the function, given by its truth table,
  and return two things: (1) the linked list of supergates, which can implement
  the functions of this N-class; (2) the phase, which should be applied to the
  given function, in order to derive the canonical form of this N-class.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperTableLookupC(p *Map_SuperLib_t, uTruth []uint32) *Map_Super_t {
    var pEnt *Map_HashEntry_t;
    var Key uint32;
    // inlining of macro MAP_TABLE_HASH
    Key = (((uTruth[0]) + 2003 * (uTruth[1])) % p.tTableC.nBins);
    for  pEnt = p.tTableC.pBins[Key]; (pEnt != nil); pEnt = pEnt.pNext  {
        if pEnt.uTruth[0] == uTruth[0] && pEnt.uTruth[1] == uTruth[1] {
            return pEnt.pGates; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Looks up an entry in the library.]

  Description [This function looks up the function, given by its truth table,
  and return two things: (1) the linked list of supergates, which can implement
  the functions of this N-class; (2) the phase, which should be applied to the
  given function, in order to derive the canonical form of this N-class.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperTableLookup(p *Map_HashTable_t, uTruth []uint32, puPhase *uint32) *Map_Super_t {
    var pEnt *Map_HashEntry_t;
    var Key uint32;
    // inlining of macro MAP_TABLE_HASH
    Key = (((uTruth[0]) + 2003 * (uTruth[1])) % p.nBins);
    for  pEnt = p.pBins[Key]; (pEnt != nil); pEnt = pEnt.pNext  {
        if pEnt.uTruth[0] == uTruth[0] && pEnt.uTruth[1] == uTruth[1] {
            *puPhase = pEnt.uPhase;
            return pEnt.pGates;
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Resizes the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperTableResize(p *Map_HashTable_t) {
    var pBinsNew **Map_HashEntry_t;
    var pEnt, pEnt2 *Map_HashEntry_t;
    var nBinsNew, Counter, i int32;
    var Key uint32;
    // get the new table size
    nBinsNew = Abc_PrimeCudd(2 * p.nBins); 
    // allocate a new array
    // inlining of macro ABC_ALLOC
    pBinsNew = ((Map_HashEntry_t * *) malloc(sizeof(Map_HashEntry_t *) * (size_t)(nBinsNew)));
    memset( pBinsNew, 0, sizeof(Map_HashEntry_t *) * nBinsNew );
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < p.nBins; i++  {
        for  pEnt = p.pBins[i], pEnt2 = pEnt? pEnt.pNext: nil; (pEnt != nil); 
              pEnt = pEnt2, pEnt2 = pEnt? pEnt.pNext: nil  {
            // inlining of macro MAP_TABLE_HASH
            Key = (((pEnt.uTruth[0]) + 2003 * (pEnt.uTruth[1])) % nBinsNew);
            pEnt.pNext   = pBinsNew[Key];
            pBinsNew[Key] = pEnt;
            Counter++;
        }
    }
    assert(Counter == p.nEntries, "Counter == p.nEntries");
    // replace the table and the parameters
    p.pBins = nil;
    p.pBins = pBinsNew;
    p.nBins = nBinsNew;
}

/**Function*************************************************************

  Synopsis    [Compares the supergates by the number of times they are used.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperTableCompareSupergates(ppS1 **Map_Super_t, ppS2 **Map_Super_t) int32 {
    if (*ppS1).nUsed > (*ppS2).nUsed {
        return -1; }
    if (*ppS1).nUsed < (*ppS2).nUsed {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Compares the supergates by the number of times they are used.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperTableCompareGatesInList(ppS1 **Map_Super_t, ppS2 **Map_Super_t) int32 {
//   if ( (*ppS1)->tDelayMax.Rise > (*ppS2)->tDelayMax.Rise )
    if (*ppS1).Area > (*ppS2).Area {
        return -1; }
//   if ( (*ppS1)->tDelayMax.Rise < (*ppS2)->tDelayMax.Rise )
    if (*ppS1).Area < (*ppS2).Area {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Sorts supergates by usefulness and prints out most useful.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperTableSortSupergates(p *Map_HashTable_t, nSupersMax int32) {
    var pEnt *Map_HashEntry_t;
    var ppSupers **Map_Super_t;
    var pSuper *Map_Super_t;
    var nSupers, i int32;

    // copy all the supergates into one array
    // inlining of macro ABC_ALLOC
    ppSupers = ((Map_Super_t * *) malloc(sizeof(Map_Super_t *) * (size_t)(nSupersMax)));
    nSupers = 0;
    for  i = 0; i < p.nBins; i++  {
        for  pEnt = p.pBins[i]; (pEnt != nil); pEnt = pEnt.pNext  {
            for  pSuper = pEnt.pGates; (pSuper != nil); pSuper = pSuper.pNext  {
                ppSupers[nSupers++] = pSuper;
            }
        }
    }

    // sort by usage
    qsort( (void *)ppSupers, (size_t)nSupers, sizeof(Map_Super_t *), 
                (int (*)(void *, void *)) Map_SuperTableCompareSupergates );
    assert(Map_SuperTableCompareSupergates( ppSupers, ppSupers + nSupers - 1 ) <= 0, "Map_SuperTableCompareSupergates( ppSupers, ppSupers + nSupers - 1 ) <= 0");

    // print out the "top ten"
//    for ( i = 0; i < nSupers; i++ )
    for  i = 0; i < 10; i++  {
        if ppSupers[i].nUsed == 0 {
            break; }
        fmt.Printf("%5d : ", ppSupers[i].nUsed);
        fmt.Printf("%5d   ", ppSupers[i].Num);
        fmt.Printf("A = %5.2f   ", ppSupers[i].Area);
        fmt.Printf("D = %5.2f   ", ppSupers[i].tDelayMax.Rise);
        fmt.Printf("%s", ppSupers[i].pFormula);
        fmt.Printf("\n");
    }
    ppSupers = nil;
}

/**Function*************************************************************

  Synopsis    [Sorts supergates by max delay for each truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperTableSortSupergatesByDelay(p *Map_HashTable_t, nSupersMax int32) {
    var pEnt *Map_HashEntry_t;
    var ppSupers **Map_Super_t;
    var pSuper *Map_Super_t;
    var nSupers, i, k int32;

    // inlining of macro ABC_ALLOC
    ppSupers = ((Map_Super_t * *) malloc(sizeof(Map_Super_t *) * (size_t)(nSupersMax)));
    for  i = 0; i < p.nBins; i++  {
        for  pEnt = p.pBins[i]; (pEnt != nil); pEnt = pEnt.pNext  {
            // collect the gates in this entry
            nSupers = 0;
            for  pSuper = pEnt.pGates; (pSuper != nil); pSuper = pSuper.pNext  {
                // skip supergates, whose root is the AND gate
//                if ( strcmp( Mio_GateReadName(pSuper->pRoot), "and" ) == 0 )
//                    continue;
                ppSupers[nSupers++] = pSuper;
            }
            pEnt.pGates = nil;
            if nSupers == 0 {
                continue; }
            // sort the gates by delay
            qsort( (void *)ppSupers, (size_t)nSupers, sizeof(Map_Super_t *), 
                                (int (*)(void *, void *)) Map_SuperTableCompareGatesInList );
            assert(Map_SuperTableCompareGatesInList( ppSupers, ppSupers + nSupers - 1 ) <= 0, "Map_SuperTableCompareGatesInList( ppSupers, ppSupers + nSupers - 1 ) <= 0");
            // link them in the reverse order
            for  k = 0; k < nSupers; k++  {
                ppSupers[k].pNext = pEnt.pGates;
                pEnt.pGates = ppSupers[k];
            }
            // save the number of supergates in the list
            pEnt.pGates.nSupers = nSupers;
        }
    }
    ppSupers = nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

