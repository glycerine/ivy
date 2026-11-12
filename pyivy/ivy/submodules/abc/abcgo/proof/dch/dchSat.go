package dch

import (
	"fmt"
	"time"

	aig "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/satoko"
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
	m.callsSince++
	if m.solver == nil || (m.params.SatVarMax > 0 && m.solver.VariableCount() > m.params.SatVarMax && m.callsSince > m.params.CallsRecycle) {
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
	if aig.Aig_ObjPhase(m.graph, representative) != aig.Aig_ObjPhase(m.graph, node) {
		b ^= 1
	}
	started := time.Now()
	defer func() { m.stats.Solving += time.Since(started) }()
	for _, assumptions := range [][]int{{a, b ^ 1}, {a ^ 1, b}} {
		m.stats.SATCalls++
		status := m.solver.SolveAssumptionsLimit(assumptions, int64(m.params.ConflictLimit))
		switch status {
		case satoko.SATOKO_UNSAT:
		case satoko.SATOKO_UNDEC:
			m.stats.Unknown++
			return proofUnknown, nil, nil
		case satoko.SATOKO_SAT:
			assignment := make([]bool, m.graph.InputCount())
			for i, input := range m.graph.Inputs() {
				if variable, ok := m.variables[input.NodeID()]; ok {
					assignment[i] = m.solver.ModelValue(variable) == 1
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
