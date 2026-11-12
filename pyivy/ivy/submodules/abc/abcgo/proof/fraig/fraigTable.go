package fraig

import "fmt"

/**CFile****************************************************************

  FileName    [fraigTable.c]

  PackageName [FRAIG: Functionally reduced AND-INV graphs.]

  Synopsis    [Structural and functional hash tables.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - October 1, 2004]

  Revision    [$Id: fraigTable.c,v 1.7 2005/07/08 01:01:34 alanmi Exp $]

***********************************************************************/

//#include "fraigInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Fraig_TableResizeS( Fraig_HashTable_t * p );

// static void Fraig_TableResizeF( Fraig_HashTable_t * p, int fUseSimR );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_HashTableCreate(nSize int32) *Fraig_HashTable_t {
    var p *Fraig_HashTable_t;
    // allocate the table
    // inlining of macro ABC_ALLOC
    p = ((Fraig_HashTable_t *) malloc(sizeof(Fraig_HashTable_t) * (size_t)(1)));
    memset( p, 0, sizeof(Fraig_HashTable_t) );
    // allocate and clean the bins
    p.nBins = Abc_PrimeCudd(nSize);
    // inlining of macro ABC_ALLOC
    p.pBins = ((Fraig_Node_t * *) malloc(sizeof(Fraig_Node_t *) * (size_t)(p.nBins)));
    memset( p.pBins, 0, sizeof(Fraig_Node_t *) * p.nBins );
    return p;
}

/**Function*************************************************************

  Synopsis    [Deallocates the supergate hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_HashTableFree(p *Fraig_HashTable_t) {
    // inlining of macro ABC_FREE
    ((p.pBins) ? (free((char *) (p.pBins)), (p.pBins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Looks up an entry in the structural hash table.]

  Description [If the entry with the same children does not exists, 
  creates it, inserts it into the table, and returns 0. If the entry 
  with the same children exists, finds it, and return 1. In both cases, 
  the new/old entry is returned in ppNodeRes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_HashTableLookupS(pMan *Fraig_Man_t, p1 *Fraig_Node_t, p2 *Fraig_Node_t, ppNodeRes **Fraig_Node_t) int32 {
    var p *Fraig_HashTable_t  = pMan.pTableS;
    var pEnt *Fraig_Node_t;
    var Key uint32;

    // order the arguments
    // inlining of macro Fraig_Regular
    if  ((Fraig_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .Num > ((Fraig_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) .Num {
        pEnt = p1;
p1 = p2;
p2 = pEnt; }

    // inlining of macro Fraig_HashKey2
    Key = (((ABC_PTRUINT_T)(p1) + (ABC_PTRUINT_T)(p2) * 12582917) % p.nBins);
    // inlining of macro Fraig_TableBinForEachEntryS
    for  pEnt = p.pBins[Key];                                                      (pEnt != nil);                                                             pEnt = pEnt.pNextS  {
        if pEnt.p1 == p1 && pEnt.p2 == p2 {
            *ppNodeRes = pEnt;
            return 1;
        }
    }
    // check if it is a good time for table resizing
    if p.nEntries >= 2 * p.nBins {
        Fraig_TableResizeS( p );
        // inlining of macro Fraig_HashKey2
        Key = (((ABC_PTRUINT_T)(p1) + (ABC_PTRUINT_T)(p2) * 12582917) % p.nBins);
    }
    // create the new node
    pEnt = Fraig_NodeCreate( pMan, p1, p2 );
    // add the node to the corresponding linked list in the table
    pEnt.pNextS = p.pBins[Key];
    p.pBins[Key] = pEnt;
    *ppNodeRes = pEnt;
    p.nEntries++;
    return 0;
}


/**Function*************************************************************

  Synopsis    [Insert the entry in the functional hash table.]

  Description [If the entry with the same key exists, return it right away.  
  If the entry with the same key does not exists, inserts it and returns NULL. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_HashTableLookupF(pMan *Fraig_Man_t, pNode *Fraig_Node_t) *Fraig_Node_t {
    var p *Fraig_HashTable_t  = pMan.pTableF;
    var pEnt, pEntD *Fraig_Node_t;
    var Key uint32;

    // go through the hash table entries
    Key = pNode.uHashR % p.nBins;
    // inlining of macro Fraig_TableBinForEachEntryF
    for  pEnt = p.pBins[Key];                                                      (pEnt != nil);                                                             pEnt = pEnt.pNextF  {
        // if their simulation info differs, skip
        if !(Fraig_CompareSimInfo( pNode, pEnt, pMan.nWordsRand, 1 ) != 0) {
            continue; }
        // equivalent up to the complement
        // inlining of macro Fraig_TableBinForEachEntryD
        for  pEntD = pEnt;                                                      (pEntD != nil);                                                             pEntD = pEntD.pNextD  {
            // if their simulation info differs, skip
            if !(Fraig_CompareSimInfo( pNode, pEntD, pMan.iWordStart, 0 ) != 0) {
                continue; }
            // found a simulation-equivalent node
            return pEntD; 
        }
        // did not find a simulation equivalent node
        // add the node to the corresponding linked list
        pNode.pNextD = pEnt.pNextD;
        pEnt.pNextD  = pNode;
        // return NULL, because there is no functional equivalence in this case
        return nil;
    }

    // check if it is a good time for table resizing
    if p.nEntries >= 2 * p.nBins {
        Fraig_TableResizeF( p, 1 );
        Key = pNode.uHashR % p.nBins;
    }

    // add the node to the corresponding linked list in the table
    pNode.pNextF = p.pBins[Key];
    p.pBins[Key] = pNode;
    p.nEntries++;
    // return NULL, because there is no functional equivalence in this case
    return nil;
}

/**Function*************************************************************

  Synopsis    [Insert the entry in the functional hash table.]

  Description [If the entry with the same key exists, return it right away.  
  If the entry with the same key does not exists, inserts it and returns NULL. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_HashTableLookupF0(pMan *Fraig_Man_t, pNode *Fraig_Node_t) *Fraig_Node_t {
    var p *Fraig_HashTable_t  = pMan.pTableF0;
    var pEnt *Fraig_Node_t;
    var Key uint32;

    // go through the hash table entries
    Key = pNode.uHashD % p.nBins;
    // inlining of macro Fraig_TableBinForEachEntryF
    for  pEnt = p.pBins[Key];                                                      (pEnt != nil);                                                             pEnt = pEnt.pNextF  {
        // if their simulation info differs, skip
        if !(Fraig_CompareSimInfo( pNode, pEnt, pMan.iWordStart, 0 ) != 0) {
            continue; }
        // found a simulation-equivalent node
        return pEnt; 
    }

    // check if it is a good time for table resizing
    if p.nEntries >= 2 * p.nBins {
        Fraig_TableResizeF( p, 0 );
        Key = pNode.uHashD % p.nBins;
    }

    // add the node to the corresponding linked list in the table
    pNode.pNextF = p.pBins[Key];
    p.pBins[Key] = pNode;
    p.nEntries++;
    // return NULL, because there is no functional equivalence in this case
    return nil;
}

/**Function*************************************************************

  Synopsis    [Insert the entry in the functional hash table.]

  Description [Unconditionally add the node to the corresponding 
  linked list in the table.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_HashTableInsertF0(pMan *Fraig_Man_t, pNode *Fraig_Node_t) {
    var p *Fraig_HashTable_t  = pMan.pTableF0;
    var Key uint32  = pNode.uHashD % p.nBins;

    pNode.pNextF = p.pBins[Key];
    p.pBins[Key] = pNode;
    p.nEntries++;
}


/**Function*************************************************************

  Synopsis    [Resizes the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_TableResizeS(p *Fraig_HashTable_t) {
    var pBinsNew **Fraig_Node_t;
    var pEnt, pEnt2 *Fraig_Node_t;
    var nBinsNew, Counter, i int32;
    var clk abctime;
    var Key uint32;

clk = Abc_Clock();
    // get the new table size
    nBinsNew = Abc_PrimeCudd(2 * p.nBins); 
    // allocate a new array
    // inlining of macro ABC_ALLOC
    pBinsNew = ((Fraig_Node_t * *) malloc(sizeof(Fraig_Node_t *) * (size_t)(nBinsNew)));
    memset( pBinsNew, 0, sizeof(Fraig_Node_t *) * nBinsNew );
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < p.nBins; i++  {
        // inlining of macro Fraig_TableBinForEachEntrySafeS
        for  pEnt = p.pBins[i],                                                      pEnt2 = pEnt? pEnt.pNextS: nil;                                 (pEnt != nil);                                                             pEnt = pEnt2,                                                     pEnt2 = pEnt? pEnt.pNextS: nil  {
            // inlining of macro Fraig_HashKey2
            Key = (((ABC_PTRUINT_T)(pEnt.p1) + (ABC_PTRUINT_T)(pEnt.p2) * 12582917) % nBinsNew);
            pEnt.pNextS = pBinsNew[Key];
            pBinsNew[Key] = pEnt;
            Counter++;
        }
    }
    assert(Counter == p.nEntries, "Counter == p.nEntries");
//    printf( "Increasing the structural table size from %6d to %6d. ", p->nBins, nBinsNew );
//    ABC_PRT( "Time", Abc_Clock() - clk );
    // replace the table and the parameters
    // inlining of macro ABC_FREE
    ((p.pBins) ? (free((char *) (p.pBins)), (p.pBins) = 0) : 0);
    p.pBins = pBinsNew;
    p.nBins = nBinsNew;
}

/**Function*************************************************************

  Synopsis    [Resizes the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_TableResizeF(p *Fraig_HashTable_t, fUseSimR int32) {
    var pBinsNew **Fraig_Node_t;
    var pEnt, pEnt2 *Fraig_Node_t;
    var nBinsNew, Counter, i int32;
    var clk abctime;
    var Key uint32;

clk = Abc_Clock();
    // get the new table size
    nBinsNew = Abc_PrimeCudd(2 * p.nBins); 
    // allocate a new array
    // inlining of macro ABC_ALLOC
    pBinsNew = ((Fraig_Node_t * *) malloc(sizeof(Fraig_Node_t *) * (size_t)(nBinsNew)));
    memset( pBinsNew, 0, sizeof(Fraig_Node_t *) * nBinsNew );
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < p.nBins; i++  {
        // inlining of macro Fraig_TableBinForEachEntrySafeF
        for  pEnt = p.pBins[i],                                                      pEnt2 = pEnt? pEnt.pNextF: nil;                                 (pEnt != nil);                                                             pEnt = pEnt2,                                                     pEnt2 = pEnt? pEnt.pNextF: nil  {
            if fUseSimR != 0 {
                Key = pEnt.uHashR % nBinsNew;
            } else {
                Key = pEnt.uHashD % nBinsNew; }
            pEnt.pNextF = pBinsNew[Key];
            pBinsNew[Key] = pEnt;
            Counter++;
        }
    }
    assert(Counter == p.nEntries, "Counter == p.nEntries");
//    printf( "Increasing the functional table size from %6d to %6d. ", p->nBins, nBinsNew );
//    ABC_PRT( "Time", Abc_Clock() - clk );
    // replace the table and the parameters
    // inlining of macro ABC_FREE
    ((p.pBins) ? (free((char *) (p.pBins)), (p.pBins) = 0) : 0);
    p.pBins = pBinsNew;
    p.nBins = nBinsNew;
}


/**Function*************************************************************

  Synopsis    [Compares two pieces of simulation info.]

  Description [Returns 1 if they are equal.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_CompareSimInfo(pNode1 *Fraig_Node_t, pNode2 *Fraig_Node_t, iWordLast int32, fUseRand int32) int32 {
    var i int32;
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode1) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode1) & 01))))");
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode2) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode2) & 01))))");
    if fUseRand != 0 {
        // if their signatures differ, skip
        if pNode1.uHashR != pNode2.uHashR {
            return 0; }
        // check the simulation info
        for  i = 0; i < iWordLast; i++  {
            if pNode1.puSimR[i] != pNode2.puSimR[i] {
                return 0; }
        }
    
    } else {
        // if their signatures differ, skip
        if pNode1.uHashD != pNode2.uHashD {
            return 0; }
        // check the simulation info
        for  i = 0; i < iWordLast; i++  {
            if pNode1.puSimD[i] != pNode2.puSimD[i] {
                return 0; }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Find the number of the different pattern.]

  Description [Returns -1 if there is no such pattern]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_FindFirstDiff(pNode1 *Fraig_Node_t, pNode2 *Fraig_Node_t, fCompl int32, iWordLast int32, fUseRand int32) int32 {
    var i, v int32;
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode1) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode1) & 01))))");
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode2) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode2) & 01))))");
    // take into account possible internal complementation
    fCompl ^= pNode1.fInv;
    fCompl ^= pNode2.fInv;
    // find the pattern
    if fCompl != 0 {
        if fUseRand != 0 {
            for  i = 0; i < iWordLast; i++  {
                if pNode1.puSimR[i] != ~pNode2.puSimR[i] {
                    for  v = 0; v < 32; v++  {
                        if (pNode1.puSimR[i] ^ ~pNode2.puSimR[i]) & (1 << v) {
                            return i * 32 + v; }
                    } }
            }
        
        } else {
            for  i = 0; i < iWordLast; i++  {
                if pNode1.puSimD[i] !=  ~pNode2.puSimD[i] {
                    for  v = 0; v < 32; v++  {
                        if (pNode1.puSimD[i] ^ ~pNode2.puSimD[i]) & (1 << v) {
                            return i * 32 + v; }
                    } }
            }
        }
    
    } else {
        if fUseRand != 0 {
            for  i = 0; i < iWordLast; i++  {
                if pNode1.puSimR[i] != pNode2.puSimR[i] {
                    for  v = 0; v < 32; v++  {
                        if (pNode1.puSimR[i] ^ pNode2.puSimR[i]) & (1 << v) {
                            return i * 32 + v; }
                    } }
            }
        
        } else {
            for  i = 0; i < iWordLast; i++  {
                if pNode1.puSimD[i] !=  pNode2.puSimD[i] {
                    for  v = 0; v < 32; v++  {
                        if (pNode1.puSimD[i] ^ pNode2.puSimD[i]) & (1 << v) {
                            return i * 32 + v; }
                    } }
            }
        }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Compares two pieces of simulation info.]

  Description [Returns 1 if they are equal.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_CompareSimInfoUnderMask(pNode1 *Fraig_Node_t, pNode2 *Fraig_Node_t, iWordLast int32, fUseRand int32, puMask *uint32) int32 {
    var pSims1, pSims2 *uint32;
    var i int32;
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode1) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode1) & 01))))");
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode2) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode2) & 01))))");
    // get hold of simulation info
    if fUseRand != 0 {
        pSims1 = pNode1.puSimR;
    } else {
        pSims1 = pNode1.puSimD;
    }
    if fUseRand != 0 {
        pSims2 = pNode2.puSimR;
    } else {
        pSims2 = pNode2.puSimD;
    }    
    // check the simulation info
    for  i = 0; i < iWordLast; i++  {
        if (pSims1[i] & puMask[i]) != (pSims2[i] & puMask[i]) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Compares two pieces of simulation info.]

  Description [Returns 1 if they are equal.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_CollectXors(pNode1 *Fraig_Node_t, pNode2 *Fraig_Node_t, iWordLast int32, fUseRand int32, puMask *uint32) {
    var pSims1, pSims2 *uint32;
    var i int32;
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode1) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode1) & 01))))");
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode2) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode2) & 01))))");
    // get hold of simulation info
    if fUseRand != 0 {
        pSims1 = pNode1.puSimR;
    } else {
        pSims1 = pNode1.puSimD;
    }
    if fUseRand != 0 {
        pSims2 = pNode2.puSimR;
    } else {
        pSims2 = pNode2.puSimD;
    }    
    // check the simulation info
    for  i = 0; i < iWordLast; i++  {
        puMask[i] = ( pSims1[i] ^ pSims2[i] );
    }
}


/**Function*************************************************************

  Synopsis    [Prints stats of the structural table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_TablePrintStatsS(pMan *Fraig_Man_t) {
    var pT *Fraig_HashTable_t  = pMan.pTableS;
    var pNode *Fraig_Node_t;
    var i, Counter int32;

    fmt.Printf("Structural table. Table size = %d. Number of entries = %d.\n", pT.nBins, pT.nEntries);
    for  i = 0; i < pT.nBins; i++  {
        Counter = 0;
        // inlining of macro Fraig_TableBinForEachEntryS
        for  pNode = pT.pBins[i];                                                      (pNode != nil);                                                             pNode = pNode.pNextS  {
            Counter++;
        }
        if Counter > 1 {
            fmt.Printf("%d ", Counter);
            if Counter > 50 {
                fmt.Printf("{%d} ", i); }
        }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Prints stats of the structural table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_TablePrintStatsF(pMan *Fraig_Man_t) {
    var pT *Fraig_HashTable_t  = pMan.pTableF;
    var pNode *Fraig_Node_t;
    var i, Counter int32;

    fmt.Printf("Functional table. Table size = %d. Number of entries = %d.\n", pT.nBins, pT.nEntries);
    for  i = 0; i < pT.nBins; i++  {
        Counter = 0;
        // inlining of macro Fraig_TableBinForEachEntryF
        for  pNode = pT.pBins[i];                                                      (pNode != nil);                                                             pNode = pNode.pNextF  {
            Counter++;
        }
        if Counter > 1 {
            fmt.Printf("{%d} ", Counter); }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Prints stats of the structural table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_TablePrintStatsF0(pMan *Fraig_Man_t) {
    var pT *Fraig_HashTable_t  = pMan.pTableF0;
    var pNode *Fraig_Node_t;
    var i, Counter int32;

    fmt.Printf("Zero-node table. Table size = %d. Number of entries = %d.\n", pT.nBins, pT.nEntries);
    for  i = 0; i < pT.nBins; i++  {
        Counter = 0;
        // inlining of macro Fraig_TableBinForEachEntryF
        for  pNode = pT.pBins[i];                                                      (pNode != nil);                                                             pNode = pNode.pNextF  {
            Counter++;
        }
        if Counter == 0 {
            continue; }
/*
        printf( "\nBin = %4d :  Number of entries = %4d\n", i, Counter );
        Fraig_TableBinForEachEntryF( pT->pBins[i], pNode )
            printf( "Node %5d. Hash = %10d.\n", pNode->Num, pNode->uHashD );
*/
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Rehashes the table after the simulation info has changed.]

  Description [Assumes that the hash values have been updated after performing
  additional simulation. Rehashes the table using the new hash values. 
  Uses pNextF to link the entries in the bins. Uses pNextD to link the entries 
  with identical hash values. Returns 1 if the identical entries have been found.
  Note that identical hash values may mean that the simulation data is different.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_TableRehashF0(pMan *Fraig_Man_t, fLinkEquiv int32) int32 {
    var pT *Fraig_HashTable_t  = pMan.pTableF0;
    var pBinsNew **Fraig_Node_t;
    var pEntF, pEntF2, pEnt, pEntD2, pEntN *Fraig_Node_t;
    var ReturnValue, Counter, i int32;
    var Key uint32;

    // allocate a new array of bins
    // inlining of macro ABC_ALLOC
    pBinsNew = ((Fraig_Node_t * *) malloc(sizeof(Fraig_Node_t *) * (size_t)(pT.nBins)));
    memset( pBinsNew, 0, sizeof(Fraig_Node_t *) * pT.nBins );

    // rehash the entries in the table
    // go through all the nodes in the F-lists (and possible in D-lists, if used)
    Counter = 0;
    ReturnValue = 0;
    for  i = 0; i < pT.nBins; i++  {
        // inlining of macro Fraig_TableBinForEachEntrySafeF
        for  pEntF = pT.pBins[i],                                                      pEntF2 = pEntF? pEntF.pNextF: nil;                                 pEntF;                                                             pEntF = pEntF2,                                                     pEntF2 = pEntF? pEntF.pNextF: nil  {
        // inlining of macro Fraig_TableBinForEachEntrySafeD
        for  pEnt = pEntF,                                                      pEntD2 = pEnt? pEnt.pNextD: nil;                                 (pEnt != nil);                                                             pEnt = pEntD2,                                                     pEntD2 = pEnt? pEnt.pNextD: nil  {
            // decide where to put entry pEnt
            Key = pEnt.uHashD % pT.nBins;
            if fLinkEquiv != 0 {
                // go through the entries in the new bin
                // inlining of macro Fraig_TableBinForEachEntryF
                for  pEntN = pBinsNew[Key];                                                      (pEntN != nil);                                                             pEntN = pEntN.pNextF  {
                    // if they have different values skip
                    if pEnt.uHashD != pEntN.uHashD {
                        continue; }
                    // they have the same hash value, add pEnt to the D-list pEnt3
                    pEnt.pNextD  = pEntN.pNextD;
                    pEntN.pNextD = pEnt;
                    ReturnValue = 1;
                    Counter++;
                    break;
                }
                if pEntN != nil { // already linked
                    continue; }
                // we did not find equal entry
            }
            // link the new entry
            pEnt.pNextF  = pBinsNew[Key];
            pBinsNew[Key] = pEnt;
            pEnt.pNextD  = nil;
            Counter++;
        }
        }
    }
    assert(Counter == pT.nEntries, "Counter == pT.nEntries");
    // replace the table and the parameters
    // inlining of macro ABC_FREE
    ((pT.pBins) ? (free((char *) (pT.pBins)), (pT.pBins) = 0) : 0);
    pT.pBins = pBinsNew;
    return ReturnValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

