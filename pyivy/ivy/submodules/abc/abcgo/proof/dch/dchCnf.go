package dch

import (
	"fmt"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// satLiteral implements Dch_CnfNodeAddToSolver: assign variables to a
// breadth-first frontier and encode AND supergates and recognized MUXes.
func (m *Manager) satLiteral(edge aig.Edge) (int32, error) {
	var frontier []aig.Edge
	add := func(e aig.Edge) {
		e = aig.Aig_Regular(e)
		id := e.NodeID()
		if _, ok := m.variables.get(id); ok {
			return
		}
		variable := m.nextVar
		m.nextVar++
		if variable >= m.solver.NumVars() {
			m.solver.NewVariable()
		}
		m.variables.set(id, variable)
		if m.satGraph.Node(e).Kind != aig.Input {
			frontier = append(frontier, e)
		}
	}
	literal := func(e aig.Edge) int32 {
		v := m.variables.at(e.NodeID()) << 1
		if e.Complemented() {
			v ^= 1
		}
		if m.params.PolarFlip && aig.Aig_ObjPhase(m.satGraph, aig.Aig_Regular(e)) {
			v ^= 1
		}
		return v
	}
	add(edge)
	// scratch reused across the frontier: the cone's leaves and its clauses
	// (flat, ends[k] closing clause k), added to the solver in the same order
	var leaves []aig.Edge
	unique := func(e aig.Edge) {
		for _, old := range leaves {
			if old == e {
				return
			}
		}
		leaves = append(leaves, e)
	}
	var flat []int32
	var ends []int
	clause := func(lits ...int32) {
		flat = append(flat, lits...)
		ends = append(ends, len(flat))
	}
	var collect func(aig.Edge, bool)
	collect = func(e aig.Edge, first bool) {
		kind, f0, f1 := m.satGraph.KindFanins(e)
		if e.Complemented() || kind != aig.And || !first && m.satGraph.Node(e).Refs > 1 || aig.Aig_ObjIsMuxType(m.satGraph, e) {
			unique(e)
			return
		}
		collect(f0, false)
		collect(f1, false)
	}
	for i := 0; i < len(frontier); i++ {
		node := frontier[i]
		kind, fanin0, fanin1 := m.satGraph.KindFanins(node)
		f := literal(node)
		leaves, flat, ends = leaves[:0], flat[:0], ends[:0]
		switch {
		case kind == aig.And && aig.Aig_ObjIsMuxType(m.satGraph, node):
			c, t, e := aig.Aig_ObjRecognizeMux(m.satGraph, node)
			// Match source frontier allocation order: each child AND's fanins.
			_, l0, l1 := m.satGraph.KindFanins(fanin0)
			_, r0, r1 := m.satGraph.KindFanins(fanin1)
			for _, fanin := range [4]aig.Edge{l0, r0, l1, r1} {
				unique(aig.Aig_Regular(fanin))
			}
			for _, leaf := range leaves {
				add(leaf)
			}
			ci, ti, ei := literal(c), literal(t), literal(e)
			clause(ci^1, ti^1, f)
			clause(ci^1, ti, f^1)
			clause(ci, ei^1, f)
			clause(ci, ei, f^1)
			if t.NodeID() != e.NodeID() {
				clause(ti, ei, f^1)
				clause(ti^1, ei^1, f)
			}
		case kind == aig.And:
			collect(node, true)
			for _, leaf := range leaves {
				add(leaf)
			}
			for _, leaf := range leaves {
				clause(literal(leaf), f^1)
			}
			for _, leaf := range leaves {
				flat = append(flat, literal(leaf)^1)
			}
			flat = append(flat, f)
			ends = append(ends, len(flat))
		default:
			return 0, fmt.Errorf("dch: unsupported CNF node %d", node.NodeID())
		}
		start := 0
		for _, end := range ends {
			if !m.solver.AddClause(flat[start:end]...) {
				return 0, fmt.Errorf("dch: inconsistent gate CNF for node %d", node.NodeID())
			}
			start = end
		}
	}
	return literal(edge), nil
}
