package nm

import (
	"fmt"
	"strings"
)

/**CFile****************************************************************

  FileName    [nmTable.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Name manager.]

  Synopsis    [Hash table for the name manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nmTable.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "nmInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// hashing for integers
func Nm_HashNumber(Num int32, TableSize int32) uint32 {
    var Key uint32;
    Key ^= ( Num        & 0xFF) * 7937;
    Key ^= ((Num >>  8) & 0xFF) * 2971;
    Key ^= ((Num >> 16) & 0xFF) * 911;
    Key ^= ((Num >> 24) & 0xFF) * 353;
    return Key % TableSize;
}

// hashing for strings
var spStatic_Nm_HashString_s_Primes [10]int32 = [10]int32{ 
        1291, 1699, 2357, 4177, 5147, 
        5647, 6343, 7103, 7873, 8147,
    };
func Nm_HashString(pName string, TableSize int32) uint32 {
    
    var i, Key uint32;
    for  i = 0; pName[i] != '\0'; i++  {
        Key ^= spStatic_Nm_HashString_s_Primes[i%10]*pName[i]*pName[i];
    }
    return Key % TableSize;
}

// static void Nm_ManResize( Nm_Man_t * p );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Adds an entry to two hash tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManTableAdd(p *Nm_Man_t, pEntry *Nm_Entry_t) int32 {
    var ppSpot **Nm_Entry_t;
    var pOther *Nm_Entry_t;
    // resize the tables if needed
    if p.nEntries > p.nBins * p.nSizeFactor {
        Nm_ManResize( p ); }
    // add the entry to the table Id->Name
    assert(Nm_ManTableLookupId(p, pEntry.ObjId) == nil, "Nm_ManTableLookupId(p, pEntry.ObjId) == nil");
    ppSpot = p.pBinsI2N + Nm_HashNumber(pEntry.ObjId, p.nBins);
    pEntry.pNextI2N = *ppSpot;
    *ppSpot = pEntry;
    // check if an entry with the same name already exists
    if pOther = Nm_ManTableLookupName(p, pEntry.Name, -1) {
        // entry with the same name already exists - add it to the ring
        if pOther.pNameSake {
            pEntry.pNameSake = pOther.pNameSake;
        } else {
            pEntry.pNameSake = pOther;
        }
        pOther.pNameSake = pEntry;
    
    } else {
        // entry with the same name does not exist - add it to the table
        ppSpot = p.pBinsN2I + Nm_HashString(pEntry.Name, p.nBins);
        pEntry.pNextN2I = *ppSpot;
        *ppSpot = pEntry;
    }
    // report successfully added entry
    p.nEntries++;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Deletes the entry from two hash tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManTableDelete(p *Nm_Man_t, ObjId int32) int32 {
    var ppSpot **Nm_Entry_t;
    var pEntry *Nm_Entry_t;
    var pPrev *Nm_Entry_t;
    var fRemoved int32;
    p.nEntries--;
    // remove the entry from the table Id->Name
    assert(Nm_ManTableLookupId(p, ObjId) != nil, "Nm_ManTableLookupId(p, ObjId) != nil");
    ppSpot = p.pBinsI2N + Nm_HashNumber(ObjId, p.nBins);
    for  (*ppSpot).ObjId != uint32(ObjId)  {
        ppSpot = &(*ppSpot).pNextI2N;
    }
    pEntry = *ppSpot; 
    *ppSpot = (*ppSpot).pNextI2N;
    // remove the entry from the table Name->Id
    ppSpot = p.pBinsN2I + Nm_HashString(pEntry.Name, p.nBins);
    for  *ppSpot && *ppSpot != pEntry  {
        ppSpot = &(*ppSpot).pNextN2I;
    }
    // remember if we found this one in the list
    fRemoved = (*ppSpot != nil);
    if *ppSpot {
        assert(*ppSpot == pEntry, "*ppSpot == pEntry");
        *ppSpot = (*ppSpot).pNextN2I;
    }
    // quit if this entry has no namesakes
    if pEntry.pNameSake == nil {
        assert((fRemoved != 0), "fRemoved");
        return 1;
    }
    // remove entry from the ring of namesakes
    assert(pEntry.pNameSake != pEntry, "pEntry.pNameSake != pEntry");
    for  pPrev = pEntry; pPrev.pNameSake != pEntry; pPrev = pPrev.pNameSake  {}
    assert(!(int32(strings.Compare(pPrev.Name, pEntry.Name)) != 0), "!strcmp(pPrev.Name, pEntry.Name)");
    assert(pPrev.pNameSake == pEntry, "pPrev.pNameSake == pEntry");
    if pEntry.pNameSake == pPrev { // two entries in the ring
        pPrev.pNameSake = nil;
    } else {
        pPrev.pNameSake = pEntry.pNameSake; }
    // reinsert the ring back if we removed its connection with the list in the table
    if fRemoved != 0 {
        assert(pPrev.pNextN2I == nil, "pPrev.pNextN2I == nil");
        pPrev.pNextN2I = *ppSpot;
        *ppSpot = pPrev;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Looks up the entry by ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManTableLookupId(p *Nm_Man_t, ObjId int32) *Nm_Entry_t {
    var pEntry *Nm_Entry_t;
    for  pEntry = p.pBinsI2N[ Nm_HashNumber(ObjId, p.nBins) ]; (pEntry != nil); pEntry = pEntry.pNextI2N  {
        if pEntry.ObjId == uint32(ObjId) {
            return pEntry; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Looks up the entry by name and type.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManTableLookupName(p *Nm_Man_t, pName string, Type int32) *Nm_Entry_t {
    var pEntry, pTemp *Nm_Entry_t;
    for  pEntry = p.pBinsN2I[ Nm_HashString(pName, p.nBins) ]; (pEntry != nil); pEntry = pEntry.pNextN2I  {
        // check the entry itself
        if !(int32(strings.Compare(pEntry.Name, pName)) != 0) && (Type == -1 || pEntry.Type == uint32(Type)) {
            return pEntry; }
        // if there is no namesakes, continue
        if pEntry.pNameSake == nil {
            continue; }
        // check the list of namesakes
        for  pTemp = pEntry.pNameSake; pTemp != pEntry; pTemp = pTemp.pNameSake  {
            if !(int32(strings.Compare(pTemp.Name, pName)) != 0) && (Type == -1 || pTemp.Type == uint32(Type)) {
                return pTemp; }
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Profiles hash tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManProfile(p *Nm_Man_t) {
    var pEntry *Nm_Entry_t;
    var Counter, e int32;
    fmt.Printf("I2N table: ");
    for  e = 0; e < p.nBins; e++  {
        Counter = 0;
        for  pEntry = p.pBinsI2N[e]; (pEntry != nil); pEntry = pEntry.pNextI2N  {
            Counter++;
        }
        fmt.Printf("%d ", Counter);
    }
    fmt.Printf("\n");
    fmt.Printf("N2I table: ");
    for  e = 0; e < p.nBins; e++  {
        Counter = 0;
        for  pEntry = p.pBinsN2I[e]; (pEntry != nil); pEntry = pEntry.pNextN2I  {
            Counter++;
        }
        fmt.Printf("%d ", Counter);
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Resizes the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManResize(p *Nm_Man_t) {
    var pBinsNewI2N **Nm_Entry_t;
    var pBinsNewN2I **Nm_Entry_t;
    var pEntry *Nm_Entry_t;
    var pEntry2 *Nm_Entry_t;
    var ppSpot **Nm_Entry_t;
    var nBinsNew, Counter, e int32;
    var clk abctime;

clk = Abc_Clock();
    // get the new table size
    nBinsNew = Abc_PrimeCudd( p.nGrowthFactor * p.nBins ); 
    // allocate a new array
    // inlining of macro ABC_ALLOC
    pBinsNewI2N = ((Nm_Entry_t * *) malloc(sizeof(Nm_Entry_t *) * (size_t)(nBinsNew)));
    // inlining of macro ABC_ALLOC
    pBinsNewN2I = ((Nm_Entry_t * *) malloc(sizeof(Nm_Entry_t *) * (size_t)(nBinsNew)));
    memset( pBinsNewI2N, 0, sizeof(Nm_Entry_t *) * nBinsNew );
    memset( pBinsNewN2I, 0, sizeof(Nm_Entry_t *) * nBinsNew );
    // rehash entries in Id->Name table
    Counter = 0;
    for  e = 0; e < p.nBins; e++  {
        for  pEntry = p.pBinsI2N[e], pEntry2 = pEntry? pEntry.pNextI2N : nil; 
              (pEntry != nil); pEntry = pEntry2, pEntry2 = pEntry? pEntry.pNextI2N : nil  {
                ppSpot = pBinsNewI2N + Nm_HashNumber(pEntry.ObjId, nBinsNew);
                pEntry.pNextI2N = *ppSpot;
                *ppSpot = pEntry;
                Counter++;
            }
    }
    // rehash entries in Name->Id table
    for  e = 0; e < p.nBins; e++  {
        for  pEntry = p.pBinsN2I[e], pEntry2 = pEntry? pEntry.pNextN2I : nil; 
              (pEntry != nil); pEntry = pEntry2, pEntry2 = pEntry? pEntry.pNextN2I : nil  {
                ppSpot = pBinsNewN2I + Nm_HashString(pEntry.Name, nBinsNew);
                pEntry.pNextN2I = *ppSpot;
                *ppSpot = pEntry;
            }
    }
    assert(Counter == p.nEntries, "Counter == p.nEntries");
//    printf( "Increasing the structural table size from %6d to %6d. ", p->nBins, nBinsNew );
//    ABC_PRT( "Time", Abc_Clock() - clk );
    // replace the table and the parameters
    // inlining of macro ABC_FREE
    ((p.pBinsI2N) ? (free((char *) (p.pBinsI2N)), (p.pBinsI2N) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pBinsN2I) ? (free((char *) (p.pBinsN2I)), (p.pBinsN2I) = 0) : 0);
    p.pBinsI2N = pBinsNewI2N;
    p.pBinsN2I = pBinsNewN2I;
    p.nBins = nBinsNew;
//    Nm_ManProfile( p );
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

