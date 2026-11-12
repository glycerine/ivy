package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigTable.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Structural hashing table.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigTable.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// hashing the node
func Aig_Hash(pObj *Aig_Obj_t, TableSize int32) unsigned long {
    unsigned long Key = Aig_ObjIsExor(pObj) * 1699;
    Key ^= Aig_ObjFanin0(pObj).Id * 7937;
    Key ^= Aig_ObjFanin1(pObj).Id * 2971;
    Key ^= Aig_ObjFaninC0(pObj) * 911;
    Key ^= Aig_ObjFaninC1(pObj) * 353;
    return Key % TableSize;
}

// returns the place where this node is stored (or should be stored)
func Aig_TableFind(p *Aig_Man_t, pObj *Aig_Obj_t) **Aig_Obj_t {
    var ppEntry **Aig_Obj_t;
    assert(Aig_ObjChild0(pObj) && Aig_ObjChild1(pObj), "Aig_ObjChild0(pObj) && Aig_ObjChild1(pObj)");
    assert(Aig_ObjFanin0(pObj).Id < Aig_ObjFanin1(pObj).Id, "Aig_ObjFanin0(pObj).Id < Aig_ObjFanin1(pObj).Id");
    for  ppEntry = p.pTable + Aig_Hash(pObj, p.nTableSize); *ppEntry; ppEntry = &(*ppEntry).pNext  {
        if *ppEntry == pObj {
            return ppEntry; }
    }
    assert(*ppEntry == nil, "*ppEntry == nil");
    return ppEntry;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Resizes the table.]

  Description [Typically this procedure should not be called.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TableResize(p *Aig_Man_t) {
    var pEntry, pNext *Aig_Obj_t;
    var pTableOld, ppPlace **Aig_Obj_t;
    var nTableSizeOld, Counter, i int32;
    var clk abctime;
    assert(p.pTable != nil, "p.pTable != nil");
clk = Abc_Clock();
    // save the old table
    pTableOld = p.pTable;
    nTableSizeOld = p.nTableSize;
    // get the new table
    p.nTableSize = Abc_PrimeCudd( 2 * Aig_ManNodeNum(p) ); 
    // inlining of macro ABC_ALLOC
    p.pTable = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(p.nTableSize)));
    memset( p.pTable, 0, sizeof(Aig_Obj_t *) * p.nTableSize );
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < nTableSizeOld; i++  {
    for  pEntry = pTableOld[i], pNext = pEntry? pEntry.pNext : nil; 
          (pEntry != nil); pEntry = pNext, pNext = pEntry? pEntry.pNext : nil  {
        // get the place where this entry goes in the table 
        ppPlace = Aig_TableFind( p, pEntry );
        assert(*ppPlace == nil, "*ppPlace == nil"); // should not be there
        // add the entry to the list
        *ppPlace = pEntry;
        pEntry.pNext = nil;
        Counter++;
    }
    }
    assert(Counter == Aig_ManNodeNum(p), "Counter == Aig_ManNodeNum(p)");
//    printf( "Increasing the structural table size from %6d to %6d. ", nTableSizeOld, p->nTableSize );
//    ABC_PRT( "Time", Abc_Clock() - clk );
    // replace the table and the parameters
    // inlining of macro ABC_FREE
    ((pTableOld) ? (free((char *) (pTableOld)), (pTableOld) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Checks if node with the given attributes is in the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TableLookup(p *Aig_Man_t, pGhost *Aig_Obj_t) *Aig_Obj_t {
    var pEntry *Aig_Obj_t;
    assert(!Aig_IsComplement(pGhost), "!Aig_IsComplement(pGhost)");
    assert(Aig_ObjIsNode(pGhost), "Aig_ObjIsNode(pGhost)");
    assert(Aig_ObjChild0(pGhost) && Aig_ObjChild1(pGhost), "Aig_ObjChild0(pGhost) && Aig_ObjChild1(pGhost)");
    assert(Aig_ObjFanin0(pGhost).Id < Aig_ObjFanin1(pGhost).Id, "Aig_ObjFanin0(pGhost).Id < Aig_ObjFanin1(pGhost).Id");
    if p.pTable == nil || !Aig_ObjRefs(Aig_ObjFanin0(pGhost)) || !Aig_ObjRefs(Aig_ObjFanin1(pGhost)) {
        return nil; }
    for  pEntry = p.pTable[Aig_Hash(pGhost, p.nTableSize)]; (pEntry != nil); pEntry = pEntry.pNext  {
        if Aig_ObjChild0(pEntry) == Aig_ObjChild0(pGhost) && 
             Aig_ObjChild1(pEntry) == Aig_ObjChild1(pGhost) && 
             Aig_ObjType(pEntry) == Aig_ObjType(pGhost) {
            return pEntry; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Checks if node with the given attributes is in the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TableLookupTwo(p *Aig_Man_t, pFanin0 *Aig_Obj_t, pFanin1 *Aig_Obj_t) *Aig_Obj_t {
    var pGhost *Aig_Obj_t;
    // consider simple cases
    if pFanin0 == pFanin1 {
        return pFanin0; }
    if pFanin0 == Aig_Not(pFanin1) {
        return Aig_ManConst0(p); }
    if Aig_Regular(pFanin0) == Aig_ManConst1(p) {
        if pFanin0 == Aig_ManConst1(p) {
            return pFanin1;
        } else {
            return Aig_ManConst0(p);
        } }
    if Aig_Regular(pFanin1) == Aig_ManConst1(p) {
        if pFanin1 == Aig_ManConst1(p) {
            return pFanin0;
        } else {
            return Aig_ManConst0(p);
        } }
    pGhost = Aig_ObjCreateGhost( p, pFanin0, pFanin1, AIG_OBJ_AND );
    return Aig_TableLookup( p, pGhost );
}

/**Function*************************************************************

  Synopsis    [Adds the new node to the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TableInsert(p *Aig_Man_t, pObj *Aig_Obj_t) {
    var ppPlace **Aig_Obj_t;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(Aig_TableLookup(p, pObj) == nil, "Aig_TableLookup(p, pObj) == nil");
    if (pObj.Id & 0xFF) == 0 && 2 * p.nTableSize < Aig_ManNodeNum(p) {
        Aig_TableResize( p ); }
    ppPlace = Aig_TableFind( p, pObj );
    assert(*ppPlace == nil, "*ppPlace == nil");
    *ppPlace = pObj;
}

/**Function*************************************************************

  Synopsis    [Deletes the node from the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TableDelete(p *Aig_Man_t, pObj *Aig_Obj_t) {
    var ppPlace **Aig_Obj_t;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    ppPlace = Aig_TableFind( p, pObj );
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
func Aig_TableCountEntries(p *Aig_Man_t) int32 {
    var pEntry *Aig_Obj_t;
    var i, Counter int32;
    for  i = 0; i < p.nTableSize; i++  {
        for  pEntry = p.pTable[i]; (pEntry != nil); pEntry = pEntry.pNext  {
            Counter++;
        }
    }
    return Counter;
}

/**Function********************************************************************

  Synopsis    [Profiles the hash table.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Aig_TableProfile(p *Aig_Man_t) {
    var pEntry *Aig_Obj_t;
    var i, Counter int32;
    fmt.Printf("Table size = %d. Entries = %d.\n", p.nTableSize, Aig_ManNodeNum(p));
    for  i = 0; i < p.nTableSize; i++  {
        Counter = 0;
        for  pEntry = p.pTable[i]; (pEntry != nil); pEntry = pEntry.pNext  {
            Counter++;
        }
        if Counter != 0 { 
            fmt.Printf("%d ", Counter); }
    }
}

/**Function********************************************************************

  Synopsis    [Profiles the hash table.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Aig_TableClear(p *Aig_Man_t) {
    // inlining of macro ABC_FREE
    ((p.pTable) ? (free((char *) (p.pTable)), (p.pTable) = 0) : 0);
    p.nTableSize = 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

