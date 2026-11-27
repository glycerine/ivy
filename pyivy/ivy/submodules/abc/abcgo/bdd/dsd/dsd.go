// Package dsd is a line-by-line port of ABC's src/bdd/dsd (dsd.h,
// dsdInt.h, dsdApi.c, dsdMan.c, dsdProc.c, dsdTree.c, dsdLocal.c,
// dsdCheck.c): BDD-based disjoint-support decomposition.
//
// This fast BDD-based recursive algorithm for simple (single-output) DSD
// is based on the following papers:
// (1) V. Bertacco and M. Damiani, "Disjunctive decomposition of logic
// functions," Proc. ICCAD '97, pp. 78-82.
// (2) Y. Matsunaga, "An exact and efficient algorithm for disjunctive
// decomposition", Proc. SASIMI '98, pp. 44-50.
package dsd

import bdd "github.com/glycerine/abcgo/bdd/cudd/core"

type DdManager = bdd.DdManager
type DdNode = bdd.DdNode

// Dsd_Type_t: types of DSD nodes
type Dsd_Type_t int

const (
	DSD_NODE_NONE   Dsd_Type_t = 0
	DSD_NODE_CONST1 Dsd_Type_t = 1
	DSD_NODE_BUF    Dsd_Type_t = 2
	DSD_NODE_OR     Dsd_Type_t = 3
	DSD_NODE_EXOR   Dsd_Type_t = 4
	DSD_NODE_PRIME  Dsd_Type_t = 5
)

// Dsd_Manager_t is the DSD manager (dsdInt.h).
type Dsd_Manager_t struct {
	dd      *DdManager              // the BDD manager
	table   map[*DdNode]*Dsd_Node_t // the mapping of BDDs into their DEs
	order   []*Dsd_Node_t           // the table's entries in insertion order (st__table walk)
	nInputs int                     // the number of primary inputs
	inputs  []*Dsd_Node_t           // the primary input nodes
	roots   []*Dsd_Node_t           // the primary output nodes (nRoots = len)
	pConst1 *Dsd_Node_t             // the constant node
	verbose bool                    // the verbosity level

	// dsdProc.c's static counters
	sMark          int
	sNDecBlocks    int
	sNLiterals     int
	sNExorGates    int
	sNReusedBlocks int
	sNCascades     int
	sNPrimeBlocks  int

	// dsdCheck.c's cache (pCache), allocated by Dsd_ManagerStart
	cache *dsdCache
}

// Dsd_Node_t is the DSD node (dsdInt.h). C tags the low pointer bit for a
// complemented node; Go uses a canonical inverse handle instead: a
// complemented handle has regular != nil and carries no payload.
type Dsd_Node_t struct {
	Type    Dsd_Type_t    // decomposition type
	G       *DdNode       // function of the node
	S       *DdNode       // support of this function
	pDecs   []*Dsd_Node_t // pointer to structures for formal inputs (nDecs = len)
	Mark    int           // the user mark (Dsd_NodeSetMark)
	nVisits int           // the counter of visits

	// kernelMark is the mark used by CASE 4 of disjoint decomposition.
	// C shares Mark with the user, whose marks are Abc_Obj_t pointers
	// (abcDsd.c) and so never equal the small s_Mark counter; Go user
	// marks are object ids, so the kernel keeps its own field.
	kernelMark int

	regular, inverse *Dsd_Node_t
}

// complementation and testing for pointers for decomposition entries
func Dsd_IsComplement(p *Dsd_Node_t) bool { return p != nil && p.regular != nil }
func Dsd_Regular(p *Dsd_Node_t) *Dsd_Node_t {
	if p != nil && p.regular != nil {
		return p.regular
	}
	return p
}
func Dsd_Not(p *Dsd_Node_t) *Dsd_Node_t {
	if p == nil {
		return nil
	}
	return p.inverse
}
func Dsd_NotCond(p *Dsd_Node_t, c bool) *Dsd_Node_t {
	if c {
		return Dsd_Not(p)
	}
	return p
}

const MAXINPUTS = 1000
