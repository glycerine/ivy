package dch

import (
	"fmt"

	aig "github.com/glycerine/abcgo/aig/aig"
)

func (m *Manager) sweep() error {
	// Source AIGs have topological IDs. An explicit order preserves unknown-
	// fanin blocking after Go graph rewrites that keep non-topological IDs.
	for _, node := range m.order {
		kind := m.graph.Node(node).Kind
		if kind != aig.And && kind != aig.ExclusiveOr {
			continue
		}
		n := m.graph.Node(node)
		if m.blocked[n.Fanins[0].NodeID()] || m.blocked[n.Fanins[1].NodeID()] {
			m.blocked[node.NodeID()] = true
			m.stats.Blocked++
			continue
		}
		for {
			representative := m.classes.representative(node)
			if representative == aig.InvalidEdge {
				break
			}
			// This can occur after non-topological graph rewrites. Never introduce a
			// representative link whose own logic depends on the candidate member.
			if aig.Aig_ObjCheckTfi(m.graph, representative, node) != 0 {
				m.classes.remove(node)
				break
			}
			status, assignment, err := m.provePair(representative, node)
			if err != nil {
				return err
			}
			if status == proofEquivalent {
				m.proven[node.NodeID()] = representative
				m.stats.Proofs++
				break
			}
			if status == proofUnknown {
				m.blocked[node.NodeID()] = true
				m.classes.remove(node)
				break
			}
			m.refineCounterexample(assignment)
			if m.classes.representative(node) == representative {
				return fmt.Errorf("dch: SAT counterexample did not refine node %d", node.NodeID())
			}
		}
	}
	return nil
}
