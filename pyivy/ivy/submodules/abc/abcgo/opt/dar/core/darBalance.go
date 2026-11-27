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

// Dar_BalancePrintStats is darBalance.c's: for each EXOR that is not the
// single-fanout fanin of another EXOR, it prints the size of its EXOR
// supergate (Dar_BalanceCone_rec) and the levels of the leaves.
func Dar_BalancePrintStats(w io.Writer, p *aig.Graph) error {
	if aig.Aig_ManExorNum(p) == 0 {
		return writeProgress(w, "There is no EXOR gates.\n")
	}
	markA := make([]bool, p.NodeSlots())
	var exors []aig.Edge
	for _, e := range p.LiveEdges() {
		n := p.Node(e)
		if n.Kind != aig.ExclusiveOr {
			continue
		}
		exors = append(exors, e)
		markA[n.Fanins[0].NodeID()] = true
		markA[n.Fanins[1].NodeID()] = true
		assert(!n.Fanins[0].Complemented(), "!Aig_ObjFaninC0(pObj)")
		assert(!n.Fanins[1].Complemented(), "!Aig_ObjFaninC1(pObj)")
	}
	var text []byte
	for _, pObj := range exors {
		if markA[pObj.NodeID()] && p.Node(pObj).Refs == 1 {
			continue
		}
		vSuper := darBalanceCone_rec(p, pObj, pObj, nil)
		if len(vSuper) < 3 {
			continue
		}
		text = fmt.Appendf(text, "  %d(", len(vSuper))
		for _, pTemp := range vSuper {
			text = fmt.Appendf(text, " %d", p.Node(pTemp).Level)
		}
		text = append(text, " )"...)
	}
	text = append(text, '\n')
	return writeProgress(w, "%s", text)
}

// darBalanceCone_rec is darBalance.c's Dar_BalanceCone_rec: it collects
// the leaves of pRoot's supergate (nodes of the root's type with one
// fanout, looking through buffers), stopping after 10000 leaves.
func darBalanceCone_rec(p *aig.Graph, pRoot, pObj aig.Edge, vSuper []aig.Edge) []aig.Edge {
	if pObj != pRoot && (pObj.Complemented() || p.Node(pObj).Kind != p.Node(pRoot).Kind || p.Node(pObj).Refs > 1 || len(vSuper) > 10000) {
		return append(vSuper, pObj)
	}
	n := p.Node(pObj)
	vSuper = darBalanceCone_rec(p, pRoot, aig.Aig_ObjReal_rec(p, n.Fanins[0]), vSuper)
	return darBalanceCone_rec(p, pRoot, aig.Aig_ObjReal_rec(p, n.Fanins[1]), vSuper)
}
