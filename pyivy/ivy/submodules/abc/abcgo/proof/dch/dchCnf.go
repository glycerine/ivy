package dch

import (
	"fmt"

	aig "github.com/glycerine/abcgo/aig/aig"
)

func (m *Manager) satLiteral(edge aig.Edge) (int, error) {
	id := edge.NodeID()
	variable, exists := m.variables[id]
	if !exists {
		variable = m.solver.AddVariable(false)
		m.variables[id] = variable
		n := m.satGraph.Node(aig.Aig_Regular(edge))
		var clauses [][]int
		raw := variable << 1
		if m.params.PolarFlip && n.Phase {
			raw ^= 1
		}
		switch n.Kind {
		case aig.Constant:
			clauses = [][]int{{raw}}
		case aig.Input:
		default:
			a, err := m.satLiteral(n.Fanins[0])
			if err != nil {
				return 0, err
			}
			switch n.Kind {
			case aig.Buffer, aig.Output:
				clauses = [][]int{{raw ^ 1, a}, {raw, a ^ 1}}
			case aig.And, aig.ExclusiveOr:
				b, err := m.satLiteral(n.Fanins[1])
				if err != nil {
					return 0, err
				}
				if n.Kind == aig.And {
					clauses = [][]int{{raw ^ 1, a}, {raw ^ 1, b}, {raw, a ^ 1, b ^ 1}}
				} else {
					clauses = [][]int{{a, b, raw ^ 1}, {a ^ 1, b ^ 1, raw ^ 1}, {a, b ^ 1, raw}, {a ^ 1, b, raw}}
				}
			}
		}
		for _, clause := range clauses {
			if !m.solver.AddClause(clause) {
				return 0, fmt.Errorf("dch: inconsistent gate CNF for node %d", id)
			}
		}
	}
	literal := variable << 1
	if edge.Complemented() {
		literal ^= 1
	}
	if m.params.PolarFlip && aig.Aig_ObjPhase(m.satGraph, aig.Aig_Regular(edge)) {
		literal ^= 1
	}
	return literal, nil
}
