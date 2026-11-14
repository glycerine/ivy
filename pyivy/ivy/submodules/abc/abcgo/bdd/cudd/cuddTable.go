package cudd

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/mtr"
)

/**CFile***********************************************************************

  FileName    [cuddTable.c]

  PackageName [cudd]

  Synopsis    [Unique table management functions.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_Prime()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddAllocNode()
                <li> cuddInitTable()
                <li> cuddFreeTable()
                <li> cuddGarbageCollect()
                <li> cuddZddGetNode()
                <li> cuddZddGetNodeIVO()
                <li> cuddUniqueInter()
                <li> cuddUniqueInterIVO()
                <li> cuddUniqueInterZdd()
                <li> cuddUniqueConst()
                <li> cuddRehash()
                <li> cuddShrinkSubtable()
                <li> cuddInsertSubtables()
                <li> cuddDestroySubtables()
                <li> cuddResizeTableZdd()
                <li> cuddSlowTableGrowth()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> ddRehashZdd()
                <li> ddResizeTable()
                <li> cuddFindParent()
                <li> cuddOrderedInsert()
                <li> cuddOrderedThread()
                <li> cuddRotateLeft()
                <li> cuddRotateRight()
                <li> cuddDoRebalance()
                <li> cuddCheckCollisionOrdering()
                </ul>]

  SeeAlso     []

  Author      [Fabio Somenzi]

  Copyright   [Copyright (c) 1995-2004, Regents of the University of Colorado

  All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions
  are met:

  Redistributions of source code must retain the above copyright
  notice, this list of conditions and the following disclaimer.

  Redistributions in binary form must reproduce the above copyright
  notice, this list of conditions and the following disclaimer in the
  documentation and/or other materials provided with the distribution.

  Neither the name of the University of Colorado nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
  "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
  LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
  FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
  COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
  INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
  BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
  LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
  CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
  LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN
  ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
  POSSIBILITY OF SUCH DAMAGE.]

******************************************************************************/

//#include "misc/util/util_hack.h"
//#include "cuddInt.h"

//ABC_NAMESPACE_IMPL_START

/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

// Go-native CUDD tables use manager-owned maps and Go heap nodes. C-only
// alignment, free-list, and red-black-tree declarations are not needed.

// Table helpers are implemented below and in the core manager.

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Returns the next prime &gt;= p.]

  Description []

  SideEffects [None]

******************************************************************************/
// Cudd_Prime returns the least prime greater than or equal to p.
func Cudd_Prime(p uint32) uint32 {
	if p <= 2 {
		return 2
	}
	for {
		if p&1 != 0 {
			prime := true
			for i := uint64(3); i*i <= uint64(p); i += 2 {
				if p%uint32(i) == 0 {
					prime = false
					break
				}
			}
			if prime {
				return p
			}
		}
		// There is no uint32 prime greater than MaxUint32. Return 0 to
		// avoid wrapping the candidate back to zero and looping forever.
		if p == ^uint32(0) {
			return 0
		}
		p++
	}
}

/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/*
*Function********************************************************************

	Synopsis    [Fast storage allocation for DdNodes in the table.]

	Description [Fast storage allocation for DdNodes in the table. The
	first 4 bytes of a chunk contain a pointer to the next block; the
	rest contains DD_MEM_CHUNK spaces for DdNodes.  Returns a pointer to
	a new node if successful; NULL is memory is full.]

	SideEffects [None]

	SeeAlso     [cuddDynamicAllocNode]

*****************************************************************************
*/
func cuddAllocNode(unique *DdManager) *DdNode {
	return core.Cudd_AllocNode(unique)
}

/*
*Function********************************************************************

	Synopsis    [Creates and initializes the unique table.]

	Description [Creates and initializes the unique table. Returns a pointer
	to the table if successful; NULL otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_Init cuddFreeTable]

*****************************************************************************
*/
func cuddInitTable(numVars, numVarsZ, numSlots, looseUpTo uint32) *DdManager {
	return core.Cudd_InitTable(numVars, numVarsZ, numSlots, looseUpTo)
}

/*
*Function********************************************************************

	Synopsis    [Frees the resources associated to a unique table.]

	Description []

	SideEffects [None]

	SeeAlso     [cuddInitTable]

*****************************************************************************
*/
func cuddFreeTable(unique *DdManager) {
	core.Cudd_Quit(unique)
}

/*
*Function********************************************************************

	Synopsis    [Performs garbage collection on the unique tables.]

	Description [Performs garbage collection on the BDD and ZDD unique tables.
	If clearCache is 0, the cache is not cleared. This should only be
	specified if the cache has been cleared right before calling
	cuddGarbageCollect. (As in the case of dynamic reordering.)
	Returns the total number of deleted nodes.]

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
func cuddGarbageCollect(unique *DdManager, clearCache int32) int32 {
	return core.Cudd_GarbageCollect(unique, clearCache != 0)
}

/*
*Function********************************************************************

	Synopsis [Wrapper for cuddUniqueInterZdd.]

	Description [Wrapper for cuddUniqueInterZdd, which applies the ZDD
	reduction rule. Returns a pointer to the result node under normal
	conditions; NULL if reordering occurred or memory was exhausted.]

	SideEffects [None]

	SeeAlso     [cuddUniqueInterZdd]

*****************************************************************************
*/
func cuddZddGetNode(zdd *DdManager, id int32, T, E *DdNode) *DdNode {
	return core.Cudd_ZddNode(zdd, int(id), T, E)
}

/*
*Function********************************************************************

	Synopsis [Wrapper for cuddUniqueInterZdd that is independent of variable
	ordering.]

	Description [Wrapper for cuddUniqueInterZdd that is independent of
	variable ordering (IVO). This function does not require parameter
	index to precede the indices of the top nodes of g and h in the
	variable order.  Returns a pointer to the result node under normal
	conditions; NULL if reordering occurred or memory was exhausted.]

	SideEffects [None]

	SeeAlso     [cuddZddGetNode cuddZddIsop]

*****************************************************************************
*/
func cuddZddGetNodeIVO(dd *DdManager, index int32, g, h *DdNode) *DdNode {
	variable := core.Cudd_zddIthVar(dd, int(index))
	if variable == nil {
		return nil
	}
	product := core.Cudd_zddProduct(dd, variable, g)
	if product == nil {
		return nil
	}
	return core.Cudd_zddUnion(dd, product, h)
}

/*
*Function********************************************************************

	Synopsis    [Checks the unique table for the existence of an internal node.]

	Description [Checks the unique table for the existence of an internal
	node. If it does not exist, it creates a new one.  Does not
	modify the reference count of whatever is returned.  A newly created
	internal node comes back with a reference count 0.  For a newly
	created node, increments the reference counts of what T and E point
	to.  Returns a pointer to the new node if successful; NULL if memory
	is exhausted or if reordering took place.]

	SideEffects [None]

	SeeAlso     [cuddUniqueInterZdd]

*****************************************************************************
*/
func cuddUniqueInter(unique *DdManager, index int32, T, E *DdNode) *DdNode {
	return core.Cudd_BddNode(unique, int(index), T, E)
}

/*
*Function********************************************************************

	Synopsis [Wrapper for cuddUniqueInter that is independent of variable
	ordering.]

	Description [Wrapper for cuddUniqueInter that is independent of
	variable ordering (IVO). This function does not require parameter
	index to precede the indices of the top nodes of T and E in the
	variable order.  Returns a pointer to the result node under normal
	conditions; NULL if reordering occurred or memory was exhausted.]

	SideEffects [None]

	SeeAlso     [cuddUniqueInter Cudd_MakeBddFromZddCover]

*****************************************************************************
*/
func cuddUniqueInterIVO(unique *DdManager, index int32, T, E *DdNode) *DdNode {
	variable := core.Cudd_bddIthVar(unique, int(index))
	if variable == nil {
		return nil
	}
	return core.Cudd_bddIte(unique, variable, T, E)
}

/*
*Function********************************************************************

	Synopsis    [Checks the unique table for the existence of an internal
	ZDD node.]

	Description [Checks the unique table for the existence of an internal
	ZDD node. If it does not exist, it creates a new one.  Does not
	modify the reference count of whatever is returned.  A newly created
	internal node comes back with a reference count 0.  For a newly
	created node, increments the reference counts of what T and E point
	to.  Returns a pointer to the new node if successful; NULL if memory
	is exhausted or if reordering took place.]

	SideEffects [None]

	SeeAlso     [cuddUniqueInter]

*****************************************************************************
*/
func cuddUniqueInterZdd(unique *DdManager, index int32, T, E *DdNode) *DdNode {
	return core.Cudd_ZddNode(unique, int(index), T, E)
}

/*
*Function********************************************************************

	Synopsis    [Checks the unique table for the existence of a constant node.]

	Description [Checks the unique table for the existence of a constant node.
	If it does not exist, it creates a new one.  Does not
	modify the reference count of whatever is returned.  A newly created
	internal node comes back with a reference count 0.  Returns a
	pointer to the new node.]

	SideEffects [None]

*****************************************************************************
*/
func cuddUniqueConst(unique *DdManager, value CUDD_VALUE_TYPE) *DdNode {
	return core.Cudd_addConst(unique, float64(value))
}

/*
*Function********************************************************************

	Synopsis    [Rehashes a unique subtable.]

	Description [Doubles the size of a unique subtable and rehashes its
	contents.]

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
func cuddRehash(unique *DdManager, i int32) {
	core.Cudd_RehashBDDSubtable(unique, int(i))
}

/*
*Function********************************************************************

	Synopsis    [Shrinks a subtable.]

	Description [Shrinks a subtable.]

	SideEffects [None]

	SeeAlso     [cuddRehash]

*****************************************************************************
*/
func cuddShrinkSubtable(unique *DdManager, i int32) {
	core.Cudd_ShrinkBDDSubtable(unique, int(i))
}

/*
*Function********************************************************************

	Synopsis [Inserts n new subtables in a unique table at level.]

	Description [Inserts n new subtables in a unique table at level.
	The number n should be positive, and level should be an existing level.
	Returns 1 if successful; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [cuddDestroySubtables]

*****************************************************************************
*/
func cuddInsertSubtables(unique *DdManager, n int32, level int32) int32 {
	if core.Cudd_InsertBDDSubtables(unique, int(n), int(level)) {
		return 1
	}
	return 0
}

/*
*Function********************************************************************

	Synopsis [Destroys the n most recently created subtables in a unique table.]

	Description [Destroys the n most recently created subtables in a unique
	table.  n should be positive. The subtables should not contain any live
	nodes, except the (isolated) projection function. The projection
	functions are freed.  Returns 1 if successful; 0 otherwise.]

	SideEffects [The variable map used for fast variable substitution is
	destroyed if it exists. In this case the cache is also cleared.]

	SeeAlso     [cuddInsertSubtables Cudd_SetVarMap]

*****************************************************************************
*/
func cuddDestroySubtables(unique *DdManager, n int32) int32 {
	if core.Cudd_DestroyBDDSubtables(unique, int(n)) {
		return 1
	}
	return 0
}

/*
*Function********************************************************************

	Synopsis [Increases the number of ZDD subtables in a unique table so
	that it meets or exceeds index.]

	Description [Increases the number of ZDD subtables in a unique table so
	that it meets or exceeds index.  When new ZDD variables are created, it
	is possible to preserve the functions unchanged, or it is possible to
	preserve the covers unchanged, but not both. cuddResizeTableZdd preserves
	the covers.  Returns 1 if successful; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [ddResizeTable]

*****************************************************************************
*/
func cuddResizeTableZdd(unique *DdManager, index int32) int32 {
	if core.Cudd_ResizeZddTable(unique, int(index)) {
		return 1
	}
	return 0
}

/*
*Function********************************************************************

	Synopsis    [Adjusts parameters of a table to slow down its growth.]

	Description []

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
func cuddSlowTableGrowth(unique *DdManager) {
	core.Cudd_SlowTableGrowth(unique)
}

/*
*Function********************************************************************

	Synopsis    [Rehashes a ZDD unique subtable.]

	Description []

	SideEffects [None]

	SeeAlso     [cuddRehash]

*****************************************************************************
*/
func ddRehashZdd(unique *DdManager, i int32) {
	core.Cudd_RehashZddSubtable(unique, int(i))
}

/*
*Function********************************************************************

	Synopsis [Increases the number of subtables in a unique table so
	that it meets or exceeds index.]

	Description [Increases the number of subtables in a unique table so
	that it meets or exceeds index. Returns 1 if successful; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [cuddResizeTableZdd]

*****************************************************************************
*/
func ddResizeTable(unique *DdManager, index int32) int32 {
	if core.Cudd_ResizeBDDTable(unique, int(index)) {
		return 1
	}
	return 0
}

/*
*Function********************************************************************

	Synopsis    [Searches the subtables above node for a parent.]

	Description [Searches the subtables above node for a parent. Returns 1
	as soon as one parent is found. Returns 0 is the search is fruitless.]

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
func cuddFindParent(table *DdManager, node *DdNode) int32 {
	if core.Cudd_FindParent(table, node) {
		return 1
	}
	return 0
}

/**Function********************************************************************

  Synopsis    [Adjusts the values of table limits.]

  Description [Adjusts the values of table fields controlling the.
  sizes of subtables and computed table. If the computed table is too small
  according to the new values, it is resized.]

  SideEffects [Modifies manager fields. May resize computed table.]

  SeeAlso     []

******************************************************************************/
//DD_INLINE
func ddFixLimits(unique *DdManager) {
	core.Cudd_FixTableLimits(unique)
}

// C-only free-list red-black tree routines are subsumed by Go heap
// allocation and manager-owned unique maps.

// CUDD's optional red-black free-list backend consists of ordered insert,
// thread, rotate, and rebalance helpers. The Go manager stores unique nodes
// in canonical maps and allocates from the runtime heap, so those C-only
// helpers have no corresponding mutable table state.

func ddPatchTree(dd *DdManager, treenode *mtr.MtrNode) {
	for node := treenode; node != nil; node = node.Younger {
		node.Low = uint32(core.Cudd_ReadPerm(dd, int(node.Index)))
		if node.Child != nil {
			ddPatchTree(dd, node.Child)
		}
	}
}

// C-only debug collision ordering checks and fatal reference-count diagnostics
// have no table-list or process-abort equivalent in the Go manager.

// Debug-only collision-list validation and fatal refcount reporting relied on
// C collision lists and process aborts; Go maps and error returns replace them.

//ABC_NAMESPACE_IMPL_END
