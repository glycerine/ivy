package dch

import (
	"fmt"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// sweep incrementally builds the FRAIG used for SAT, as in Dch_ManSweep.
// The source graph stays intact for simulation and publishing representatives.
func (m *Manager) sweep() error {
	m.satGraph = aig.NewGraph()
	m.mapped = make([]aig.Edge, len(m.graph.Nodes()))
	for i := range m.mapped {
		m.mapped[i] = aig.InvalidEdge
	}
	m.mapped[0] = m.satGraph.Const1()
	for _, input := range m.graph.Inputs() {
		m.mapped[input.NodeID()] = m.satGraph.NewInput()
	}
	mappedEdge := func(edge aig.Edge) aig.Edge {
		if edge == aig.InvalidEdge {
			return edge
		}
		result := m.mapped[edge.NodeID()]
		if result != aig.InvalidEdge && edge.Complemented() {
			result = result.Not()
		}
		return result
	}
	for _, node := range m.order {
		n := m.graph.Node(node)
		if n.Kind != aig.And {
			continue
		}
		left, right := mappedEdge(n.Fanins[0]), mappedEdge(n.Fanins[1])
		if left == aig.InvalidEdge || right == aig.InvalidEdge {
			m.blocked[node.NodeID()] = true
			m.stats.Blocked++
			continue
		}
		m.mapped[node.NodeID()] = m.satGraph.AND(left, right)
		representative := m.classes.representative(node)
		if representative == aig.InvalidEdge {
			continue
		}
		rep := m.mapped[representative.NodeID()]
		if rep == aig.InvalidEdge {
			continue
		}
		candidate := m.mapped[node.NodeID()]
		status := proofEquivalent
		var assignment []bool
		var err error
		if aig.Aig_Regular(rep) != aig.Aig_Regular(candidate) {
			status, assignment, err = m.provePair(aig.Aig_Regular(rep), aig.Aig_Regular(candidate))
		}
		if err != nil {
			return err
		}
		if status == proofEquivalent {
			m.proven[node.NodeID()] = representative
			m.stats.Proofs++
			if aig.Aig_ObjPhase(m.graph, representative) != aig.Aig_ObjPhase(m.graph, node) {
				rep = rep.Not()
			}
			m.mapped[node.NodeID()] = rep
			continue
		}
		if status == proofUnknown {
			m.blocked[node.NodeID()] = true
			m.mapped[node.NodeID()] = aig.InvalidEdge
			continue
		}
		m.refineCounterexample(assignment)
		if m.classes.representative(node) == representative {
			return fmt.Errorf("dch: SAT counterexample did not refine node %d", node.NodeID())
		}
	}
	return nil
}
