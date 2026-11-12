package dar

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"sort"
)

// Dar_ManRefactor refactors bounded fanout-free cones through exact cofactors.
// Typed truth tables replace the source Kit temporary-memory graph. Each proposed
// network must improve MFFC area (or tie when requested) and obey the level bound.
func Dar_ManRefactor(g *aig.Graph, params *RefactorParams) error {
	p := DefaultRefactorParams()
	if params != nil {
		p = *params
	}
	if g == nil {
		return fmt.Errorf("dar: nil graph")
	}
	if p.LeafMax < 2 || p.LeafMax > 16 || p.CutsMax < 1 || p.MffcMin < 0 {
		return fmt.Errorf("dar: invalid refactor limits")
	}
	if err := g.Check(); err != nil {
		return err
	}
	for _, root := range g.LiveEdges() {
		if p.Context != nil {
			if err := p.Context.Err(); err != nil {
				return err
			}
		}
		if !g.ValidEdge(root) || g.Node(root).Kind != aig.And {
			continue
		}
		leaves, ok := refactorBoundary(g, root, p.LeafMax, p.Extend)
		if !ok {
			continue
		}
		oldArea := int(aig.Aig_NodeMffcLabelCut(g, root, leaves))
		if oldArea < p.MffcMin {
			continue
		}
		removable := make(map[int]bool)
		oldIDs := make(map[int]bool)
		for _, e := range g.LiveEdges() {
			oldIDs[e.NodeID()] = true
			if aig.Aig_ObjIsTravIdCurrent(g, e) {
				removable[e.NodeID()] = true
			}
		}
		truth := refactorTruth(g, root, leaves)
		replacement := refactorBuild(g, leaves, truth)
		if graphDependsOn(g, replacement, root) {
			aig.Aig_ManCleanup(g)
			continue
		}
		seen := make(map[int]bool)
		var cost func(aig.Edge) int
		cost = func(e aig.Edge) int {
			id := e.NodeID()
			if seen[id] {
				return 0
			}
			seen[id] = true
			if oldIDs[id] && !removable[id] {
				return 0
			}
			n := g.Node(e)
			switch n.Kind {
			case aig.And, aig.ExclusiveOr:
				return 1 + cost(n.Fanins[0]) + cost(n.Fanins[1])
			case aig.Buffer, aig.Output:
				return cost(n.Fanins[0])
			}
			return 0
		}
		gain := oldArea - cost(replacement)
		if gain < 0 || gain == 0 && !p.UseZeros || p.UpdateLevel && g.Node(replacement).Level > g.Node(root).Level {
			aig.Aig_ManCleanup(g)
			continue
		}
		aig.Aig_ObjReplace(g, root, replacement, 0)
		aig.Aig_ManCleanup(g)
	}
	return g.Check()
}
func refactorBoundary(g *aig.Graph, root aig.Edge, limit int, extend bool) ([]aig.Edge, bool) {
	boundary := make(map[aig.Edge]bool)
	visited := make(map[int]bool)
	var collect func(aig.Edge, bool)
	collect = func(e aig.Edge, top bool) {
		e = aig.Edge(e.NodeID() << 1)
		if visited[e.NodeID()] {
			return
		}
		visited[e.NodeID()] = true
		n := g.Node(e)
		if n.Kind == aig.Constant {
			return
		}
		if n.Kind == aig.Input || !extend && !top && n.Refs > 1 {
			boundary[e] = true
			return
		}
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			collect(n.Fanins[0], false)
			collect(n.Fanins[1], false)
		case aig.Buffer, aig.Output:
			collect(n.Fanins[0], false)
		}
	}
	collect(root, true)
	if len(boundary) > limit {
		return nil, false
	}
	leaves := make([]aig.Edge, 0, len(boundary))
	for e := range boundary {
		leaves = append(leaves, e)
	}
	sort.Slice(leaves, func(i, j int) bool { return leaves[i] < leaves[j] })
	return leaves, true
}
func refactorTruth(g *aig.Graph, root aig.Edge, leaves []aig.Edge) []bool {
	truth := make([]bool, 1<<len(leaves))
	for assignment := range truth {
		values := map[int]bool{0: true}
		for i, e := range leaves {
			values[e.NodeID()] = assignment&(1<<i) != 0
		}
		var evaluate func(aig.Edge) bool
		evaluate = func(e aig.Edge) bool {
			v, ok := values[e.NodeID()]
			if !ok {
				n := g.Node(e)
				a := evaluate(n.Fanins[0])
				switch n.Kind {
				case aig.And:
					v = a && evaluate(n.Fanins[1])
				case aig.ExclusiveOr:
					v = a != evaluate(n.Fanins[1])
				case aig.Buffer, aig.Output:
					v = a
				default:
					panic("dar: incomplete refactor boundary")
				}
				values[e.NodeID()] = v
			}
			return v != e.Complemented()
		}
		truth[assignment] = evaluate(root)
	}
	return truth
}
func refactorBuild(g *aig.Graph, leaves []aig.Edge, truth []bool) aig.Edge {
	all := truth[0]
	constant := true
	for _, v := range truth {
		if v != all {
			constant = false
			break
		}
	}
	if constant {
		if all {
			return 0
		}
		return 1
	}
	variable := len(leaves) - 1
	half := len(truth) / 2
	low, high := refactorBuild(g, leaves[:variable], truth[:half]), refactorBuild(g, leaves[:variable], truth[half:])
	if low == high {
		return low
	}
	return g.Mux(leaves[variable], high, low)
}
