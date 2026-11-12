package hop

/**CFile****************************************************************

  FileName    [hopMem.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Minimalistic And-Inverter Graph package.]

  Synopsis    [Memory management for the AIG nodes.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: hopMem.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "hop.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// memory management
const IVY_PAGE_SIZE = 12   // page size containing 2^IVY_PAGE_SIZE nodes
const IVY_PAGE_MASK = 4095 // page bitmask (2^IVY_PAGE_SIZE)-1

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the internal memory manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// Hop_ManStartMemory initializes the typed node-page pool. Go owns the page
// storage, so the C implementation's byte alignment and pointer casts vanish.
func Hop_ManStartMemory(p *Hop_Man_t) {
	p.vChunks = make([][]Hop_Obj_t, 0, 128)
	p.vPages = make([][]Hop_Obj_t, 0, 128)
	p.pListFree = nil
}

// Hop_ManStopMemory releases the manager's ownership of its node pages.
func Hop_ManStopMemory(p *Hop_Man_t) {
	p.vChunks = nil
	p.vPages = nil
	p.pListFree = nil
}

// Hop_ManAddMemory allocates one page and links its nodes into the free list.
// Free nodes use pNext; live nodes reuse that field for structural hashing.
func Hop_ManAddMemory(p *Hop_Man_t) {
	if p.pListFree != nil {
		panic("Hop_ManAddMemory: free list is not empty")
	}
	page := make([]Hop_Obj_t, 1<<IVY_PAGE_SIZE)
	p.vChunks = append(p.vChunks, page)
	p.vPages = append(p.vPages, page)
	for i := 0; i+1 < len(page); i++ {
		page[i].pNext = &page[i+1]
	}
	p.pListFree = &page[0]
}
