package dar

import (
	"github.com/glycerine/abcgo/aig/aig"
)

type rewriteCandidate struct {
	class       *preparedClass
	root        preparedRoot
	inputs      [4]aig.Edge
	complement  bool
	gain, level int
}

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
func (m *Manager) Rewrite() error {
	g := m.Graph
	for _, root := range g.LiveEdges() {
		if m.Params.Context != nil {
			if err := m.Params.Context.Err(); err != nil {
				return err
			}
		}
		if !g.ValidEdge(root) || g.Node(root).Kind != aig.And {
			continue
		}
		m.Stats.Roots++
		var best *rewriteCandidate
		for _, cut := range NewCutEnumerator(g, m.Params.CutsMax).Cuts(root) {
			if len(cut.Leaves) != 4 {
				continue
			}
			m.Stats.Cuts++
			match := m.library.library.canonical.Lookup(cut.Truth)
			class := &m.library.classes[match.Class]
			var inputs [4]aig.Edge
			for i := range inputs {
				inputs[i] = cut.Leaves[match.Permutation[i]]
				if match.InputPhase&(1<<i) != 0 {
					inputs[i] = inputs[i].Not()
				}
			}
			oldArea := int(aig.Aig_NodeMffcLabelCut(g, root, cut.Leaves))
			existing := make([]aig.Edge, len(class.nodes)+4)
			levels := make([]int, len(existing))
			for i := range existing {
				existing[i] = aig.InvalidEdge
			}
			for i, e := range inputs {
				existing[i] = e
				levels[i] = g.Node(e).Level
			}
			for i, node := range class.nodes {
				a, b := node.fanins[0], node.fanins[1]
				x, y := existing[a>>1], existing[b>>1]
				if x != aig.InvalidEdge && y != aig.InvalidEdge {
					if a&1 != 0 {
						x = x.Not()
					}
					if b&1 != 0 {
						y = y.Not()
					}
					e := aig.Aig_TableLookupTwo(g, x, y)
					if e != aig.InvalidEdge && e.NodeID() != root.NodeID() {
						existing[i+4] = e
					}
				}
				levels[i+4] = 1 + max(levels[a>>1], levels[b>>1])
				if existing[i+4] != aig.InvalidEdge {
					levels[i+4] = g.Node(existing[i+4]).Level
				}
			}
			for _, candidate := range class.roots {
				if m.Params.UpdateLevel && levels[candidate.index] > g.Node(root).Level {
					continue
				}
				seen := make([]bool, len(existing))
				var cost func(int) int
				cost = func(index int) int {
					if index < 4 || seen[index] {
						return 0
					}
					seen[index] = true
					e := existing[index]
					if e != aig.InvalidEdge && !aig.Aig_ObjIsTravIdCurrent(g, e) {
						return 0
					}
					node := class.nodes[index-4]
					return 1 + cost(int(node.fanins[0]>>1)) + cost(int(node.fanins[1]>>1))
				}
				gain := oldArea - cost(candidate.index)
				if gain < 0 || gain == 0 && !m.Params.UseZeros {
					continue
				}
				if best == nil || gain > best.gain || gain == best.gain && levels[candidate.index] < best.level {
					best = &rewriteCandidate{class: class, root: candidate, inputs: inputs, complement: candidate.complement != match.OutputComplement, gain: gain, level: levels[candidate.index]}
				}
			}
		}
		if best == nil {
			continue
		}
		memo := make([]aig.Edge, len(best.class.nodes)+4)
		for i := range memo {
			memo[i] = aig.InvalidEdge
		}
		copy(memo, best.inputs[:])
		var build func(int) aig.Edge
		build = func(index int) aig.Edge {
			if memo[index] != aig.InvalidEdge {
				return memo[index]
			}
			n := best.class.nodes[index-4]
			a, b := build(int(n.fanins[0]>>1)), build(int(n.fanins[1]>>1))
			if n.fanins[0]&1 != 0 {
				a = a.Not()
			}
			if n.fanins[1]&1 != 0 {
				b = b.Not()
			}
			memo[index] = g.AND(a, b)
			return memo[index]
		}
		replacement := build(best.root.index)
		if best.complement {
			replacement = replacement.Not()
		}
		if graphDependsOn(g, replacement, root) {
			aig.Aig_ManCleanup(g)
			continue
		}
		aig.Aig_ObjReplace(g, root, replacement, 0)
		aig.Aig_ManCleanup(g)
		m.Stats.Replacements++
		m.Stats.NodesSaved += best.gain
	}
	return g.Check()
}

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
