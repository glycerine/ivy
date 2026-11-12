package dch

import (
	"fmt"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// CheckChoices validates choice-chain ownership and the combined logic/choice
// DAG, not just the ordinary output function. Alternate implementations must
// remain dangling in ordinary logic so a mapper can select them independently.
func CheckChoices(g *aig.Graph, skipRedundantSupport bool) error {
	if g == nil {
		return fmt.Errorf("dch: nil choice graph")
	}
	if err := g.Check(); err != nil {
		return err
	}
	incoming := map[int]int{}
	for _, node := range g.LiveEdges() {
		next := aig.Aig_ObjEquiv(g, node)
		if next == aig.InvalidEdge {
			continue
		}
		headKind := g.Node(node).Kind
		if headKind != aig.And && headKind != aig.ExclusiveOr {
			return fmt.Errorf("dch: non-gate choice head %d", node.NodeID())
		}
		if !g.ValidEdge(next) || next.Complemented() {
			return fmt.Errorf("dch: invalid choice edge at %d", node.NodeID())
		}
		kind := g.Node(next).Kind
		if kind != aig.And && kind != aig.ExclusiveOr {
			return fmt.Errorf("dch: non-gate choice alternate at %d", next.NodeID())
		}
		incoming[next.NodeID()]++
		if incoming[next.NodeID()] > 1 {
			return fmt.Errorf("dch: alternate %d belongs to multiple chains", next.NodeID())
		}
		if g.Node(next).Refs != 0 {
			return fmt.Errorf("dch: referenced choice alternate %d", next.NodeID())
		}
	}
	for _, head := range g.LiveEdges() {
		if incoming[head.NodeID()] != 0 || aig.Aig_ObjEquiv(g, head) == aig.InvalidEdge {
			continue
		}
		for next := aig.Aig_ObjEquiv(g, head); next != aig.InvalidEdge; next = aig.Aig_ObjEquiv(g, next) {
			if skipRedundantSupport && Dch_ObjCheckSuppRed(g, head, next) {
				return fmt.Errorf("dch: redundant support in choice %d", head.NodeID())
			}
		}
	}
	colors := map[int]uint8{}
	var visit func(aig.Edge) error
	visit = func(edge aig.Edge) error {
		regular := aig.Aig_Regular(edge)
		id := regular.NodeID()
		if colors[id] == 2 {
			return nil
		}
		if colors[id] == 1 {
			return fmt.Errorf("dch: logic/choice cycle at %d", id)
		}
		colors[id] = 1
		n := g.Node(regular)
		switch n.Kind {
		case aig.Buffer, aig.Output:
			if err := visit(n.Fanins[0]); err != nil {
				return err
			}
		case aig.And, aig.ExclusiveOr:
			for _, f := range n.Fanins {
				if err := visit(f); err != nil {
					return err
				}
			}
		}
		if next := aig.Aig_ObjEquiv(g, regular); next != aig.InvalidEdge {
			if err := visit(next); err != nil {
				return err
			}
		}
		colors[id] = 2
		return nil
	}
	for _, node := range g.LiveEdges() {
		if err := visit(node); err != nil {
			return err
		}
	}
	return nil
}
func Dch_CheckChoices(g *aig.Graph, skip int) error { return CheckChoices(g, skip != 0) }
