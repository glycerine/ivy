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
		if _, ok := m.variables[id]; ok {
			return
		}
		variable := m.nextVar
		m.nextVar++
		if variable >= m.solver.NumVars() {
			m.solver.NewVariable()
		}
		m.variables[id] = variable
		if m.satGraph.Node(e).Kind != aig.Input {
			frontier = append(frontier, e)
		}
	}
	literal := func(e aig.Edge) int32 {
		v := m.variables[e.NodeID()] << 1
		if e.Complemented() {
			v ^= 1
		}
		if m.params.PolarFlip && aig.Aig_ObjPhase(m.satGraph, aig.Aig_Regular(e)) {
			v ^= 1
		}
		return v
	}
	add(edge)
	for i := 0; i < len(frontier); i++ {
		node := frontier[i]
		n := m.satGraph.Node(node)
		f := literal(node)
		var leaves []aig.Edge
		unique := func(e aig.Edge) {
			for _, old := range leaves {
				if old == e {
					return
				}
			}
			leaves = append(leaves, e)
		}
		var clauses [][]int32
		switch {
		case n.Kind == aig.And && aig.Aig_ObjIsMuxType(m.satGraph, node):
			c, t, e := aig.Aig_ObjRecognizeMux(m.satGraph, node)
			// Match source frontier allocation order: each child AND's fanins.
			left, right := m.satGraph.Node(n.Fanins[0]), m.satGraph.Node(n.Fanins[1])
			for _, fanin := range []aig.Edge{left.Fanins[0], right.Fanins[0], left.Fanins[1], right.Fanins[1]} {
				unique(aig.Aig_Regular(fanin))
			}
			for _, leaf := range leaves {
				add(leaf)
			}
			ci, ti, ei := literal(c), literal(t), literal(e)
			clauses = [][]int32{{ci ^ 1, ti ^ 1, f}, {ci ^ 1, ti, f ^ 1}, {ci, ei ^ 1, f}, {ci, ei, f ^ 1}}
			if t.NodeID() != e.NodeID() {
				clauses = append(clauses, []int32{ti, ei, f ^ 1}, []int32{ti ^ 1, ei ^ 1, f})
			}
		case n.Kind == aig.And:
			var collect func(aig.Edge, bool)
			collect = func(e aig.Edge, first bool) {
				object := m.satGraph.Node(e)
				if e.Complemented() || object.Kind != aig.And || !first && object.Refs > 1 || aig.Aig_ObjIsMuxType(m.satGraph, e) {
					unique(e)
					return
				}
				collect(object.Fanins[0], false)
				collect(object.Fanins[1], false)
			}
			collect(node, true)
			for _, leaf := range leaves {
				add(leaf)
			}
			reverse := make([]int32, 0, len(leaves)+1)
			for _, leaf := range leaves {
				value := literal(leaf)
				clauses = append(clauses, []int32{value, f ^ 1})
				reverse = append(reverse, value^1)
			}
			clauses = append(clauses, append(reverse, f))
		default:
			return 0, fmt.Errorf("dch: unsupported CNF node %d", node.NodeID())
		}
		for _, clause := range clauses {
			if !m.solver.AddClause(clause...) {
				return 0, fmt.Errorf("dch: inconsistent gate CNF for node %d", node.NodeID())
			}
		}
	}
	return literal(edge), nil
}
