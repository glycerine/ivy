package core

import (
	"github.com/glycerine/abcgo/aig/aig"
)

// Dar_ManRewrite performs source-library, four-input DAG-aware rewriting in place.
// Successful replacements preserve graph IDs through buffers; DFS duplication
// can subsequently compact those buffers and dangling logic.
func Dar_ManRewrite(g *aig.Graph, params *RewriteParams) error {
	m, err := Dar_ManStart(g, params)
	if err != nil {
		return err
	}
	defer Dar_ManStop(m)
	return m.Rewrite()
}
func (m *Manager) Rewrite() error { return m.rewriteNative() }

// Structural hashing may rediscover the old root inside a proposed network.
// Reject that network before replacement rather than introducing a cycle.
func graphDependsOn(g *aig.Graph, edge, target aig.Edge) bool {
	seen := make(map[int]bool)
	var visit func(aig.Edge) bool
	visit = func(e aig.Edge) bool {
		if e.NodeID() == target.NodeID() {
			return true
		}
		if seen[e.NodeID()] {
			return false
		}
		seen[e.NodeID()] = true
		n := g.Node(e)
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			return visit(n.Fanins[0]) || visit(n.Fanins[1])
		case aig.Buffer, aig.Output:
			return visit(n.Fanins[0])
		}
		return false
	}
	return visit(edge)
}
