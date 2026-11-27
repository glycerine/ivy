// Package reo is a line-by-line port of src/bdd/reo (REO: a specialized DD
// reordering engine, Alan Mishchenko, UC Berkeley, 2002).
//
// The diagram is copied into units grouped by planes (one per level),
// adjacent planes are swapped in place, and the result is transferred back
// into the CUDD manager. C's unit pointers with the complement in the low
// bit become uref values: (unit index)<<1 | complement, 0 being NULL.
package reo

import (
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

type DdManager = cudd.DdManager
type DdNode = cudd.DdNode

// reordering parameters
const (
	REO_REORDER_LIMIT = 1.15 // determines the quality/runtime trade-off
	REO_QUAL_PAR      = 3    // the quality [1 = simple lower bound, 2 = strict, larger = heuristic]
)

// internal parameters
const (
	REO_CONST_LEVEL  = 30000     // the number of the constant level
	REO_TOPREF_UNDEF = 30000     // the undefined top reference
	REO_CHUNK_SIZE   = 5000      // the number of units allocated at one time (C)
	REO_COST_EPSILON = 0.0000001 // difference in cost large enough so that it counted as an error
	REO_HIGH_VALUE   = 10000000  // a large value used to initialize some variables
)

// interface parameters
const (
	REO_ENABLE  = 1 // the value of the enable flag
	REO_DISABLE = 0 // the value of the disable flag
)

// Minimize is reo_min_type: the types of minimization currently supported.
type Minimize int
type reo_min_type = Minimize

const (
	REO_MINIMIZE_NODES Minimize = iota
	REO_MINIMIZE_WIDTH          // may not work for BDDs with complemented edges
	REO_MINIMIZE_APL
)

// uref is a reo_unit pointer: unit index shifted left by one, with the
// complement in bit 0 (C's Unit_Not flips the low pointer bit).
type uref int32

func Unit_Regular(u uref) uref { return u &^ 1 }
func Unit_Not(u uref) uref     { return u ^ 1 }
func Unit_NotCond(u uref, c bool) uref {
	if c {
		return u ^ 1
	}
	return u
}
func Unit_IsComplement(u uref) bool { return u&1 != 0 }

// reo_unit represents one DD node during reordering.
type reo_unit struct {
	lev       int16 // the level of this node at the beginning
	TopRef    int16 // the top level from which this node is refed (used to update BDD width)
	TopRefNew int16 // the new top level from which this node is refed (used to update BDD width)
	n         int16 // the number of incoming edges (similar to ref count in the BDD)
	Sign      int32 // the signature
	pE        uref  // the pointer to the "else" branch (C keeps a constant's value here)
	pT        uref  // the pointer to the "then" branch
	Next      uref  // the link to the next one in the list
	value     int32 // the constant's value, (int)cuddV(F) as C stores it in pE
	Weight    float64
}

func Unit_IsConstant(u *reo_unit) bool { return u.lev == REO_CONST_LEVEL }

// reo_plane is the set of nodes on one level.
type reo_plane struct {
	fSifted        int     // to mark the sifted variables
	statsNodes     int     // the number of nodes in the current level
	statsWidth     int     // the width on the current level
	statsApl       float64 // the sum of node probabilities on this level
	statsCost      float64 // the current cost is stored here
	statsCostAbove float64 // the current cost is stored here
	statsCostBelow float64 // the current cost is stored here
	pHead          uref    // the pointer to the beginning of the unit list
}

// reo_hash is an entry in the hash table. During the transfers one side is
// a CUDD node, kept in Node (C casts it into Arg1 or Arg2).
type reo_hash struct {
	Sign int32 // signature of the current cache operation
	Arg1 uref  // the first argument
	Arg2 uref  // the second argument
	Arg3 uref  // the third argument
	Node *DdNode
}

// reo_man is the reordering manager.
type reo_man struct {
	// these paramaters can be set by the API functions
	fMinWidth int // the flag to enable reordering for minimum width
	fMinApl   int // the flag to enable reordering for minimum APL
	fVerbose  int // the verbosity level
	fVerify   int // the flag toggling verification
	fRemapUp  int // the flag to enable remapping
	nIters    int // the number of iterations of sifting to perform
	// parameters given by the user when reordering is called
	dd     *DdManager // the CUDD BDD manager
	pOrder []int      // the resulting variable order will be returned here
	// derived parameters
	fThisIsAdd  int   // this flag is one if the function is the ADD
	pSupp       []int // the support of the given function
	nSuppAlloc  int   // the max allowed number of support variables
	nSupp       int   // the number of support variables
	pOrderInt   []int // the array storing the internal variable permutation
	pVarCosts   []float64
	pLevelOrder []int
	pWidthCofs  []uref // temporary storage for cofactors used during reordering for width
	// parameters related to cost
	nNodesBeg int
	nNodesCur int
	nNodesEnd int
	nWidthCur int
	nWidthBeg int
	nWidthEnd int
	nAplCur   float64
	nAplBeg   float64
	nAplEnd   float64
	// mapping of the function into planes and back
	pMapToPlanes      []int // the mapping of var indexes into plane levels
	pMapToDdVarsOrig  []int // the mapping of plane levels into the original indexes
	pMapToDdVarsFinal []int // the mapping of plane levels into the final indexes
	// the planes table
	pPlanes    []reo_plane
	nPlanes    int
	pTops      []uref
	nTops      int
	nTopsAlloc int
	// the hash table
	HTable     []reo_hash
	nTableSize int
	tableMask  int  // nTableSize - 1
	tableShift uint // 64 - log2(nTableSize)
	Signature  int32
	// the referenced node list
	nNodesMaxAlloc int
	pRefNodes      []*DdNode
	nRefNodes      int
	nRefNodesAlloc int
	// unit memory management
	pUnitFreeList   uref
	units           []reo_unit // the chunks of C, contiguous
	nMemChunks      int
	nMemChunksAlloc int
	nUnitsUsed      int
	// statistic variables
	HashSuccess int
	HashFailure int
	nSwaps      int32 // the number of swaps
	nNISwaps    int   // the number of swaps without interaction

	// support/sharing traversal marks (C's ddSupportStep2/ddClearFlag2)
	visited nodeSet
}

// Manager is the Go name of reo_man used by callers.
type Manager = reo_man

// unit memory: one slice grown by unitChunk units at a time. Units are
// named by index, but a *reo_unit obtained from U is valid only until the
// next reoUnitsGetNextUnit, which may move the slice.
const unitChunk = 4096

// U dereferences a (possibly complemented) unit pointer.
func (p *reo_man) U(u uref) *reo_unit {
	return &p.units[u>>1]
}

// AddToLinkedList is reoSwap.c's macro.
func (p *reo_man) AddToLinkedList(ppList *uref, pLink uref) {
	p.U(pLink).Next = *ppList
	*ppList = pLink
}

// hash key macros from extraBdd.h (DD_P1, DD_P2 from cuddInt.h)
const (
	DD_P1 = 12582917
	DD_P2 = 4256249
)

// The table has a power-of-two size (at least C's 3*nNodesMax+1 entries,
// so it stays lossless) and takes the top bits of the mixed key; the probe
// order differs from C's, the entries found do not.
func hashKey2(a, b uint64, shift uint) int {
	return int(((a + b*DD_P1) * 0x9E3779B97F4A7C15) >> shift)
}
func hashKey3(a, b, c uint64, shift uint) int {
	return int(((((a+b)*DD_P1 + c) * DD_P2) * 0x9E3779B97F4A7C15) >> shift)
}

func ddMinF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
