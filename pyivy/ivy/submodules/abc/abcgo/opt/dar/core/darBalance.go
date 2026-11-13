package core

import (
	"context"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"io"
	"os"
	"sort"
)

// Balance rebuilds positive-polarity associative AND cones using lowest-level
// pairs first. Shared nodes and complemented edges remain cone boundaries.
// The input graph is unchanged; register and interface metadata survive cloning.
func Balance(ctx context.Context, g *aig.Graph, updateLevel bool) (*aig.Graph, error) {
	if g == nil {
		return nil, fmt.Errorf("dar: nil graph")
	}
	if err := g.Check(); err != nil {
		return nil, err
	}
	return balanceRecursive(ctx, g, updateLevel)
}
func Dar_ManBalance(g *aig.Graph, updateLevel bool) *aig.Graph {
	balanced, err := Balance(context.Background(), g, updateLevel)
	if err != nil {
		panic(err)
	}
	return balanced
}

// Dar_BalanceCone collects an AND supergate without crossing shared fanouts.
func Dar_BalanceCone(g *aig.Graph, root aig.Edge) []aig.Edge {
	return balanceConeKind(g, root, g.Node(root).Kind)
}
func balanceConeKind(g *aig.Graph, root aig.Edge, kind aig.NodeKind) []aig.Edge {
	var leaves []aig.Edge
	var collect func(aig.Edge, bool)
	collect = func(e aig.Edge, isRoot bool) {
		n := g.Node(e)
		if !e.Complemented() && n.Kind == aig.Buffer && n.Refs == 1 {
			collect(n.Fanins[0], false)
			return
		}
		if !e.Complemented() && n.Kind == kind && (isRoot || n.Refs == 1) {
			collect(n.Fanins[0], false)
			collect(n.Fanins[1], false)
		} else {
			leaves = append(leaves, e)
		}
	}
	collect(root, true)
	return leaves
}

// Dar_BalanceUniqify removes duplicates and detects contradictory literals.
func Dar_BalanceUniqify(leaves []aig.Edge) ([]aig.Edge, bool) {
	result := append([]aig.Edge(nil), leaves...)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	unique := result[:0]
	for _, e := range result {
		if e == 1 {
			return nil, true
		}
		if e == 0 {
			continue
		}
		if len(unique) > 0 {
			last := unique[len(unique)-1]
			if last == e {
				continue
			}
			if last.NodeID() == e.NodeID() {
				return nil, true
			}
		}
		unique = append(unique, e)
	}
	return unique, false
}
func Dar_BalanceBuildSuper(g *aig.Graph, leaves []aig.Edge) aig.Edge {
	return balanceBuildKind(g, leaves, aig.And)
}
func balanceBuildKind(g *aig.Graph, leaves []aig.Edge, kind aig.NodeKind) aig.Edge {
	work, contradiction := Dar_BalanceUniqify(leaves)
	phase := false
	if kind == aig.ExclusiveOr {
		parity := make(map[aig.Edge]bool)
		for _, e := range leaves {
			phase = phase != e.Complemented()
			regular := aig.Edge(e.NodeID() << 1)
			if regular == 0 {
				phase = !phase
				continue
			}
			parity[regular] = !parity[regular]
		}
		work = nil
		for e, odd := range parity {
			if odd {
				work = append(work, e)
			}
		}
		contradiction = false
	}
	if contradiction {
		return 1
	}
	if len(work) == 0 {
		if kind == aig.ExclusiveOr {
			if phase {
				return 0
			}
			return 1
		}
		return 0
	}
	for len(work) > 1 {
		sort.Slice(work, func(i, j int) bool {
			a, b := g.Node(work[i]).Level, g.Node(work[j]).Level
			if a != b {
				return a < b
			}
			return work[i] < work[j]
		})
		edge := aig.InvalidEdge
		if kind == aig.ExclusiveOr {
			edge = g.NewXor(work[0], work[1])
		} else {
			edge = g.AND(work[0], work[1])
		}
		work = append(work[2:], edge)
	}
	if kind == aig.ExclusiveOr && phase {
		return work[0].Not()
	}
	return work[0]
}

func Dar_ManBalanceXor(g *aig.Graph, exor, updateLevel, verbose bool) *aig.Graph {
	if exor {
		g = aig.Aig_ManDupExor(g)
	}
	if verbose {
		if err := Dar_BalancePrintStats(os.Stdout, g); err != nil {
			panic(err)
		}
	}
	return Dar_ManBalance(g, updateLevel)
}
func Dar_BalancePrintStats(w io.Writer, g *aig.Graph) error {
	count := 0
	for _, e := range g.LiveEdges() {
		if g.Node(e).Kind == aig.ExclusiveOr {
			count++
		}
	}
	return writeProgress(w, "XOR gates: %d\n", count)
}
