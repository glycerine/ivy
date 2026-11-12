package res

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/glycerine/abcgo/aig/aig"
)

// GraphParams bounds the exact cut-local resubstitution search.
type GraphParams struct {
	MaxSupport  int
	MaxDivisors int
	Context     context.Context
}

// GraphStats records exact local equivalence candidates and committed changes.
type GraphStats struct {
	Roots, SkippedCuts, Divisors, Candidates, Replacements, AndsBefore, AndsAfter int
}

// ResubstituteGraph performs exact cut-local AIG resubstitution. For each AND
// root, it considers existing signals and two-divisor AND/OR expressions over
// a bounded structural cut. A replacement is committed only when its
// MFFC is strictly smaller than the root's MFFC. No speculative simulation
// result is accepted without exact truth-table equality.
func ResubstituteGraph(g *aig.Graph, params GraphParams) (GraphStats, error) {
	var stats GraphStats
	if g == nil {
		return stats, errors.New("res: nil AIG graph")
	}
	if params.MaxSupport == 0 {
		params.MaxSupport = 6
	}
	if params.MaxDivisors == 0 {
		params.MaxDivisors = 32
	}
	if params.MaxSupport < 1 || params.MaxSupport > 6 {
		return stats, fmt.Errorf("res: support bound %d is outside 1..6", params.MaxSupport)
	}
	if params.MaxDivisors < 1 {
		return stats, errors.New("res: divisor bound must be positive")
	}
	stats.AndsBefore = countGraphAnds(g)

	// A snapshot keeps newly created candidates from becoming roots this pass.
	for _, root := range g.LiveEdges() {
		if params.Context != nil {
			if err := params.Context.Err(); err != nil {
				return stats, err
			}
		}
		if !g.ValidEdge(root) || g.Node(root).Kind != aig.And {
			continue
		}
		stats.Roots++
		leaves, err := graphCut(g, root, params.MaxSupport)
		if err != nil {
			return stats, err
		}
		if len(leaves) <= 1 || len(leaves) > params.MaxSupport {
			stats.SkippedCuts++
			continue
		}
		rootTruth, err := graphTruth(g, root, leaves)
		if err != nil {
			return stats, err
		}
		rootMFFC := int(aig.Aig_NodeMffcLabelCut(g, root, leaves))
		if rootMFFC <= 1 {
			continue
		}

		divisors := graphDivisors(g, root, leaves, params.MaxDivisors)
		stats.Divisors += len(divisors)
		best := graphCandidate{saving: 0}
		for _, divisor := range divisors {
			if divisor.edge.NodeID() == root.NodeID() && !divisor.edge.Complemented() {
				continue
			}
			if divisor.truth != rootTruth {
				continue
			}
			stats.Candidates++
			cost := graphMFFCCount(g, divisor.edge, leaves)
			if saving := rootMFFC - cost; saving > best.saving {
				best = graphCandidate{edge: divisor.edge, saving: saving, level: g.Node(divisor.edge).Level}
			}
		}

		// Match exact two-divisor conjunctions and disjunctions. These candidates
		// are materialized only after their truth tables match the target.
		for i := range divisors {
			for j := i; j < len(divisors); j++ {
				a, b := divisors[i], divisors[j]
				for _, isOr := range []bool{false, true} {
					truth := a.truth & b.truth
					if isOr {
						truth = a.truth | b.truth
					}
					if truth != rootTruth {
						continue
					}
					stats.Candidates++
					candidate := g.AND(a.edge, b.edge)
					if isOr {
						candidate = g.OR(a.edge, b.edge)
					}
					if !g.ValidEdge(candidate) || candidate.NodeID() == root.NodeID() || graphDependsOn(g, candidate, root.NodeID()) {
						continue
					}
					cost := graphMFFCCount(g, candidate, leaves)
					if saving := rootMFFC - cost; saving > best.saving {
						best = graphCandidate{edge: candidate, saving: saving, level: g.Node(candidate).Level}
					}
				}
			}
		}

		if best.saving > 0 {
			aig.Aig_ObjReplace(g, root, best.edge, 0)
			stats.Replacements++
		}
		// Candidate construction may leave unused hashed nodes; reclaim them
		// before the next root so they cannot become accidental divisors.
		aig.Aig_ManCleanup(g)
	}
	stats.AndsAfter = countGraphAnds(g)
	return stats, g.Check()
}

type graphCandidate struct {
	edge   aig.Edge
	truth  uint64
	level  int
	saving int
}

func countGraphAnds(g *aig.Graph) int {
	n := 0
	for _, edge := range g.LiveEdges() {
		if g.Node(edge).Kind == aig.And {
			n++
		}
	}
	return n
}

func graphCut(g *aig.Graph, root aig.Edge, limit int) ([]aig.Edge, error) {
	if !g.ValidEdge(root) {
		return nil, fmt.Errorf("res: invalid root edge %d", root)
	}
	leaves := []aig.Edge{aig.Edge(root.NodeID() << 1)}
	for len(leaves) < limit {
		best := -1
		for i, edge := range leaves {
			node := g.Node(edge)
			if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
				continue
			}
			if best < 0 || node.Level > g.Node(leaves[best]).Level || node.Level == g.Node(leaves[best]).Level && edge.NodeID() > leaves[best].NodeID() {
				best = i
			}
		}
		if best < 0 {
			break
		}
		node := g.Node(leaves[best])
		expansion := []aig.Edge{aig.Edge(node.Fanins[0].NodeID() << 1)}
		if node.Kind == aig.And || node.Kind == aig.ExclusiveOr {
			expansion = append(expansion, aig.Edge(node.Fanins[1].NodeID()<<1))
		}
		unique := make(map[int]aig.Edge, len(leaves)+1)
		for i, edge := range leaves {
			if i != best {
				unique[edge.NodeID()] = edge
			}
		}
		for _, edge := range expansion {
			unique[edge.NodeID()] = edge
		}
		if len(unique) > limit {
			break
		}
		leaves = leaves[:0]
		for _, edge := range unique {
			leaves = append(leaves, edge)
		}
		sort.Slice(leaves, func(i, j int) bool { return leaves[i].NodeID() < leaves[j].NodeID() })
	}
	return leaves, nil
}

func graphTruth(g *aig.Graph, root aig.Edge, leaves []aig.Edge) (uint64, error) {
	leafPosition := make(map[int]int, len(leaves))
	for position, edge := range leaves {
		leafPosition[edge.NodeID()] = position
	}
	var truth uint64
	for assignment := 0; assignment < 1<<len(leaves); assignment++ {
		memo := make(map[int]bool)
		active := make(map[int]bool)
		var eval func(aig.Edge) (bool, error)
		eval = func(edge aig.Edge) (bool, error) {
			if !g.ValidEdge(edge) {
				return false, fmt.Errorf("res: invalid candidate edge %d", edge)
			}
			id := edge.NodeID()
			if position, boundary := leafPosition[id]; boundary {
				value := assignment&(1<<position) != 0
				if edge.Complemented() {
					value = !value
				}
				return value, nil
			}
			value, known := memo[id]
			if !known {
				if active[id] {
					return false, fmt.Errorf("res: cycle at AIG node %d", id)
				}
				active[id] = true
				node := g.Node(edge)
				switch node.Kind {
				case aig.Constant:
					value = true
				case aig.Input:
					return false, fmt.Errorf("res: candidate depends on an input outside the cut")
				case aig.And, aig.ExclusiveOr:
					a, err := eval(node.Fanins[0])
					if err != nil {
						return false, err
					}
					b, err := eval(node.Fanins[1])
					if err != nil {
						return false, err
					}
					if node.Kind == aig.And {
						value = a && b
					} else {
						value = a != b
					}
				case aig.Buffer, aig.Output:
					var err error
					value, err = eval(node.Fanins[0])
					if err != nil {
						return false, err
					}
				default:
					return false, fmt.Errorf("res: unsupported node kind %d", node.Kind)
				}
				active[id] = false
				memo[id] = value
			}
			if edge.Complemented() {
				value = !value
			}
			return value, nil
		}
		value, err := eval(root)
		if err != nil {
			return 0, err
		}
		if value {
			truth |= uint64(1) << assignment
		}
	}
	return truth, nil
}

func graphDivisors(g *aig.Graph, root aig.Edge, leaves []aig.Edge, limit int) []graphCandidate {
	leafSet := make(map[int]bool, len(leaves))
	for _, edge := range leaves {
		leafSet[edge.NodeID()] = true
	}
	var candidates []graphCandidate
	for _, regular := range g.LiveEdges() {
		kind := g.Node(regular).Kind
		if kind == aig.Output || kind == aig.Buffer && regular.NodeID() == root.NodeID() {
			continue
		}
		if regular.NodeID() == root.NodeID() || graphDependsOn(g, regular, root.NodeID()) {
			continue
		}
		fitsCut, err := graphFitsCut(g, regular, leafSet)
		if err != nil || !fitsCut {
			continue
		}
		truth, err := graphTruth(g, regular, leaves)
		if err != nil {
			continue
		}
		candidates = append(candidates,
			graphCandidate{edge: regular, truth: truth, level: g.Node(regular).Level},
			graphCandidate{edge: regular.Not(), truth: fullTruthMask(len(leaves)) ^ truth, level: g.Node(regular).Level},
		)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].level != candidates[j].level {
			return candidates[i].level < candidates[j].level
		}
		if candidates[i].edge.NodeID() != candidates[j].edge.NodeID() {
			return candidates[i].edge.NodeID() < candidates[j].edge.NodeID()
		}
		return !candidates[i].edge.Complemented() && candidates[j].edge.Complemented()
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates
}

func graphDependsOn(g *aig.Graph, edge aig.Edge, target int) bool {
	seen := make(map[int]bool)
	var visit func(aig.Edge) bool
	visit = func(current aig.Edge) bool {
		id := current.NodeID()
		if id == target {
			return true
		}
		if seen[id] || !g.ValidEdge(current) {
			return false
		}
		seen[id] = true
		node := g.Node(current)
		if node.Kind == aig.And || node.Kind == aig.ExclusiveOr {
			return visit(node.Fanins[0]) || visit(node.Fanins[1])
		}
		if node.Kind == aig.Buffer || node.Kind == aig.Output {
			return visit(node.Fanins[0])
		}
		return false
	}
	return visit(edge)
}

func graphFitsCut(g *aig.Graph, edge aig.Edge, leaves map[int]bool) (bool, error) {
	seen := make(map[int]bool)
	var visit func(aig.Edge) (bool, error)
	visit = func(current aig.Edge) (bool, error) {
		if !g.ValidEdge(current) {
			return false, fmt.Errorf("res: invalid divisor edge %d", current)
		}
		id := current.NodeID()
		if leaves[id] {
			return true, nil
		}
		if seen[id] {
			return true, nil
		}
		seen[id] = true
		node := g.Node(current)
		switch node.Kind {
		case aig.Constant:
			return true, nil
		case aig.Input:
			return false, nil
		case aig.And, aig.ExclusiveOr:
			a, err := visit(node.Fanins[0])
			if err != nil || !a {
				return a, err
			}
			return visit(node.Fanins[1])
		case aig.Buffer, aig.Output:
			return visit(node.Fanins[0])
		default:
			return false, fmt.Errorf("res: unsupported divisor node kind %d", node.Kind)
		}
	}
	return visit(edge)
}

func graphMFFCCount(g *aig.Graph, edge aig.Edge, leaves []aig.Edge) int {
	seen := make(map[int]bool)
	cutLeaves := make(map[int]bool, len(leaves))
	for _, leaf := range leaves {
		cutLeaves[leaf.NodeID()] = true
	}
	for {
		if !g.ValidEdge(edge) || seen[edge.NodeID()] {
			return 0
		}
		if cutLeaves[edge.NodeID()] {
			return 0
		}
		seen[edge.NodeID()] = true
		node := g.Node(edge)
		switch node.Kind {
		case aig.Constant, aig.Input:
			return 0
		case aig.Buffer, aig.Output:
			fanin := node.Fanins[0]
			if edge.Complemented() {
				fanin = fanin.Not()
			}
			edge = fanin
		default:
			regular := aig.Edge(edge.NodeID() << 1)
			return int(aig.Aig_NodeMffcLabelCut(g, regular, leaves))
		}
	}
}

func fullTruthMask(nvars int) uint64 {
	if nvars == 6 {
		return ^uint64(0)
	}
	return uint64(1)<<(1<<nvars) - 1
}
