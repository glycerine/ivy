// Package dsd implements BDD-based disjoint-support decomposition. The recursive
// cofactor algorithm is adapted from ABC's Bertacco/Damiani/Matsunaga DSD code.
package dsd

import bdd "github.com/glycerine/abcgo/bdd/cudd/core"

type DdManager = bdd.DdManager
type DdNode = bdd.DdNode
type Dsd_Type_t int

const (
	DSD_NODE_NONE Dsd_Type_t = iota
	DSD_NODE_CONST1
	DSD_NODE_BUF
	DSD_NODE_OR
	DSD_NODE_EXOR
	DSD_NODE_PRIME
)

// Dsd_Node_t is an edge handle. Complementation uses a canonical inverse
// handle, never pointer tags. Inspect node payload through Dsd_Regular.
type Dsd_Node_t struct {
	Type             Dsd_Type_t
	G, S             *DdNode
	pDecs            []*Dsd_Node_t
	Mark             int
	visits           int
	regular, inverse *Dsd_Node_t
}
type Dsd_Manager_t struct {
	dd            *DdManager
	table         map[*DdNode]*Dsd_Node_t
	inputs, roots []*Dsd_Node_t
	constant      *Dsd_Node_t
	verbose       bool
	identity      map[identityKey]bool
}

func Dsd_IsComplement(p *Dsd_Node_t) bool { return p != nil && p.regular != nil }
func Dsd_Regular(p *Dsd_Node_t) *Dsd_Node_t {
	if Dsd_IsComplement(p) {
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
