package hop

import "fmt"

/**CFile****************************************************************

  FileName    [hopTable.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Minimalistic And-Inverter Graph package.]

  Synopsis    [Structural hashing table.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006. ]

  Revision    [$Id: hopTable.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "hop.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// hashing the node
func Hop_Hash(pObj *Hop_Obj_t, TableSize int32) unsigned long {
    unsigned long Key = Hop_ObjIsExor(pObj) * 1699;
    Key ^= Hop_ObjFanin0(pObj).Id * 7937;
    Key ^= Hop_ObjFanin1(pObj).Id * 2971;
    Key ^= Hop_ObjFaninC0(pObj) * 911;
    Key ^= Hop_ObjFaninC1(pObj) * 353;
    return Key % TableSize;
}

// returns the place where this node is stored (or should be stored)
func Hop_TableFind(p *Hop_Man_t, pObj *Hop_Obj_t) **Hop_Obj_t {
    var ppEntry **Hop_Obj_t;
    assert(Hop_ObjChild0(pObj) && Hop_ObjChild1(pObj), "Hop_ObjChild0(pObj) && Hop_ObjChild1(pObj)");
    assert(Hop_ObjFanin0(pObj).Id < Hop_ObjFanin1(pObj).Id, "Hop_ObjFanin0(pObj).Id < Hop_ObjFanin1(pObj).Id");
    for  ppEntry = p.pTable + Hop_Hash(pObj, p.nTableSize); *ppEntry; ppEntry = &(*ppEntry).pNext  {
        if *ppEntry == pObj {
            return ppEntry; }
    }
    assert(*ppEntry == nil, "*ppEntry == nil");
    return ppEntry;
}

// static void         Hop_TableResize( Hop_Man_t * p );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Checks if a node with the given attributes is in the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_TableLookup(p *Hop_Man_t, pGhost *Hop_Obj_t) *Hop_Obj_t {
    var pEntry *Hop_Obj_t;
    assert(!Hop_IsComplement(pGhost), "!Hop_IsComplement(pGhost)");
    assert(Hop_ObjChild0(pGhost) && Hop_ObjChild1(pGhost), "Hop_ObjChild0(pGhost) && Hop_ObjChild1(pGhost)");
    assert(Hop_ObjFanin0(pGhost).Id < Hop_ObjFanin1(pGhost).Id, "Hop_ObjFanin0(pGhost).Id < Hop_ObjFanin1(pGhost).Id");
    if p.fRefCount && (!Hop_ObjRefs(Hop_ObjFanin0(pGhost)) || !Hop_ObjRefs(Hop_ObjFanin1(pGhost))) {
        return nil; }
    for  pEntry = p.pTable[Hop_Hash(pGhost, p.nTableSize)]; (pEntry != nil); pEntry = pEntry.pNext  {
        if Hop_ObjChild0(pEntry) == Hop_ObjChild0(pGhost) && 
             Hop_ObjChild1(pEntry) == Hop_ObjChild1(pGhost) && 
             Hop_ObjType(pEntry) == Hop_ObjType(pGhost) {
            return pEntry; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Adds the new node to the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_TableInsert(p *Hop_Man_t, pObj *Hop_Obj_t) {
    var ppPlace **Hop_Obj_t;
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    assert(Hop_TableLookup(p, pObj) == nil, "Hop_TableLookup(p, pObj) == nil");
    if (pObj.Id & 0xFF) == 0 && 2 * p.nTableSize < Hop_ManNodeNum(p) {
        Hop_TableResize( p ); }
    ppPlace = Hop_TableFind( p, pObj );
    assert(*ppPlace == nil, "*ppPlace == nil");
    *ppPlace = pObj;
}

/**Function*************************************************************

  Synopsis    [Deletes the node from the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_TableDelete(p *Hop_Man_t, pObj *Hop_Obj_t) {
    var ppPlace **Hop_Obj_t;
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    ppPlace = Hop_TableFind( p, pObj );
    assert(*ppPlace == pObj, "*ppPlace == pObj"); // node should be in the table
    // remove the node
    *ppPlace = pObj.pNext;
    pObj.pNext = nil;
}

/**Function*************************************************************

  Synopsis    [Count the number of nodes in the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_TableCountEntries(p *Hop_Man_t) int32 {
    var pEntry *Hop_Obj_t;
    var i, Counter int32;
    for  i = 0; i < p.nTableSize; i++  {
        for  pEntry = p.pTable[i]; (pEntry != nil); pEntry = pEntry.pNext  {
            Counter++;
        }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Resizes the table.]

  Description [Typically this procedure should not be called.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_TableResize(p *Hop_Man_t) {
    var pEntry, pNext *Hop_Obj_t;
    var pTableOld, ppPlace **Hop_Obj_t;
    var nTableSizeOld, Counter, nEntries, i int32;
    var clk abctime;
clk = Abc_Clock();
    // save the old table
    pTableOld = p.pTable;
    nTableSizeOld = p.nTableSize;
    // get the new table
    p.nTableSize = Abc_PrimeCudd( 2 * Hop_ManNodeNum(p) ); 
    // inlining of macro ABC_ALLOC
    p.pTable = ((Hop_Obj_t * *) malloc(sizeof(Hop_Obj_t *) * (size_t)(p.nTableSize)));
    memset( p.pTable, 0, sizeof(Hop_Obj_t *) * p.nTableSize );
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < nTableSizeOld; i++  {
    for  pEntry = pTableOld[i], pNext = pEntry? pEntry.pNext : nil; (pEntry != nil); pEntry = pNext, pNext = pEntry? pEntry.pNext : nil  {
        // get the place where this entry goes in the table 
        ppPlace = Hop_TableFind( p, pEntry );
        assert(*ppPlace == nil, "*ppPlace == nil"); // should not be there
        // add the entry to the list
        *ppPlace = pEntry;
        pEntry.pNext = nil;
        Counter++;
    }
    }
    nEntries = Hop_ManNodeNum(p);
    assert(Counter == nEntries, "Counter == nEntries");
//    printf( "Increasing the structural table size from %6d to %6d. ", nTableSizeOld, p->nTableSize );
//    ABC_PRT( "Time", Abc_Clock() - clk );
    // replace the table and the parameters
    // inlining of macro ABC_FREE
    ((pTableOld) ? (free((char *) (pTableOld)), (pTableOld) = 0) : 0);
}

/**Function********************************************************************

  Synopsis    [Profiles the hash table.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Hop_TableProfile(p *Hop_Man_t) {
    var pEntry *Hop_Obj_t;
    var i, Counter int32;
    for  i = 0; i < p.nTableSize; i++  {
        Counter = 0;
        for  pEntry = p.pTable[i]; (pEntry != nil); pEntry = pEntry.pNext  {
            Counter++;
        }
        if Counter != 0 { 
            fmt.Printf("%d ", Counter); }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

