package dch

import (
	"fmt"
	"time"

	aig "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

type proofResult int

const (
	proofUnknown    proofResult = -1
	proofDifferent  proofResult = 0
	proofEquivalent proofResult = 1
)

// provePair checks both directions of the miter using assumptions. Only two
// UNSAT results prove equivalence; SAT produces a counterexample and a conflict
// budget exhaustion remains unknown, never a guessed proof.
func (m *Manager) provePair(representative, node aig.Edge) (proofResult, []bool, error) {
	m.stats.SATCalls++
	m.callsSince++
	if m.solver == nil || (m.params.SatVarMax > 0 && int(m.nextVar) > m.params.SatVarMax && m.callsSince > m.params.CallsRecycle) {
		m.recycleSolver()
	}
	a, err := m.satLiteral(representative)
	if err != nil {
		return proofDifferent, nil, err
	}
	b, err := m.satLiteral(node)
	if err != nil {
		return proofDifferent, nil, err
	}
	if aig.Aig_ObjPhase(m.satGraph, representative) != aig.Aig_ObjPhase(m.satGraph, node) {
		b ^= 1
	}
	started := time.Now()
	defer func() { m.stats.Solving += time.Since(started) }()
	for direction, assumptions := range [][]int32{{a, b ^ 1}, {a ^ 1, b}} {
		status := m.solver.Solve(assumptions...)
		switch status {
		case bsat.Unsatisfiable:
			// The failed conjunction implies the opposite binary clause.
			// Retain it in the incremental solver as in Dch_NodesAreEquiv.
			if !m.solver.AddClause(assumptions[0]^1, assumptions[1]^1) {
				return proofUnknown, nil, fmt.Errorf("dch: inconsistent proved implication")
			}
			if direction == 0 && representative == m.satGraph.Const1() {
				return proofEquivalent, nil, nil
			}
		case bsat.Unknown:
			m.stats.Unknown++
			return proofUnknown, nil, nil
		case bsat.Satisfiable:
			assignment := make([]bool, m.graph.InputCount())
			for i, input := range m.satGraph.Inputs() {
				if variable, ok := m.variables[input.NodeID()]; ok {
					assignment[i], _ = m.solver.Value(variable)
				}
			}
			m.stats.Counterexamples++
			return proofDifferent, assignment, nil
		default:
			return proofUnknown, nil, fmt.Errorf("dch: invalid SAT status %d", status)
		}
	}
	return proofEquivalent, nil, nil
}
