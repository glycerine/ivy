package dch

import (
	"fmt"

	aig "github.com/glycerine/abcgo/aig/aig"
)

func phaseEdge(edge aig.Edge, flip bool) aig.Edge {
	if flip {
		return edge.Not()
	}
	return edge
}
func resolveRepresentative(g *aig.Graph, edge aig.Edge) aig.Edge {
	seen := map[int]bool{}
	for {
		regular := aig.Aig_Regular(edge)
		representative := aig.Aig_ObjRepr(g, regular)
		if representative == aig.InvalidEdge {
			return edge
		}
		if seen[regular.NodeID()] {
			panic("dch: representative cycle")
		}
		seen[regular.NodeID()] = true
		edge = phaseEdge(representative, edge.Complemented() != (aig.Aig_ObjPhase(g, regular) != aig.Aig_ObjPhase(g, representative)))
	}
}
func structuralSupport(g *aig.Graph, root aig.Edge) map[int]bool {
	support, seen := map[int]bool{}, map[int]bool{}
	var visit func(aig.Edge)
	visit = func(edge aig.Edge) {
		id := edge.NodeID()
		if seen[id] {
			return
		}
		seen[id] = true
		n := g.Node(aig.Aig_Regular(edge))
		switch n.Kind {
		case aig.Input:
			support[id] = true
		case aig.Buffer, aig.Output:
			visit(n.Fanins[0])
		case aig.And, aig.ExclusiveOr:
			visit(n.Fanins[0])
			visit(n.Fanins[1])
		}
	}
	visit(root)
	return support
}
func Dch_ObjCheckSuppRed(g *aig.Graph, a, b aig.Edge) bool {
	x, y := structuralSupport(g, a), structuralSupport(g, b)
	if len(x) != len(y) {
		return true
	}
	for id := range x {
		if !y[id] {
			return true
		}
	}
	return false
}
func Dch_ObjCheckTfi(g *aig.Graph, node, representative aig.Edge) bool {
	var scratch tfiScratch
	return scratch.check(g, node, representative)
}

// tfiScratch holds Dch_ObjCheckTfi's marks (ABC's fMarkA on the choice
// chain and the traversal ID, as generation stamps reused across calls)
// next to a compact mirror of the graph's fanins and choice links: the
// check walks whole fanin cones once per candidate choice, and ABC keeps
// all of this in the object, one cache line per visit. The mirror follows
// nodes as they are added (sync) and choices as they are linked (link).
type tfiScratch struct {
	nodes []tfiNode
	gen   uint32
}

type tfiNode struct {
	fanin0, fanin1, equiv int32 // node IDs, -1 for none
	target, seen          uint32
}

func tfiNodeID(e aig.Edge) int32 {
	if e == aig.InvalidEdge {
		return -1
	}
	return int32(e.NodeID())
}

// sync mirrors the nodes added to g since the last call.
func (s *tfiScratch) sync(g *aig.Graph) {
	for id := len(s.nodes); id < g.NodeSlots(); id++ {
		e := aig.Aig_ManObj(g, int32(id))
		n := tfiNode{fanin0: -1, fanin1: -1, equiv: tfiNodeID(aig.Aig_ObjEquiv(g, e))}
		switch kind, f0, f1 := g.KindFanins(e); kind {
		case aig.Buffer, aig.Output:
			n.fanin0 = tfiNodeID(f0)
		case aig.And, aig.ExclusiveOr:
			n.fanin0, n.fanin1 = tfiNodeID(f0), tfiNodeID(f1)
		}
		s.nodes = append(s.nodes, n)
	}
}

// link records that tail's next choice is member (Aig_ObjSetEquiv).
func (s *tfiScratch) link(g *aig.Graph, tail, member aig.Edge) {
	s.sync(g)
	s.nodes[tail.NodeID()].equiv = int32(member.NodeID())
}

func (s *tfiScratch) check(g *aig.Graph, node, representative aig.Edge) bool {
	s.sync(g)
	s.gen++
	if s.gen == 0 { // wrapped: restart the stamps
		for i := range s.nodes {
			s.nodes[i].target, s.nodes[i].seen = 0, 0
		}
		s.gen = 1
	}
	for id := int32(representative.NodeID()); id >= 0; id = s.nodes[id].equiv {
		if s.nodes[id].target == s.gen {
			panic("dch: cyclic choice chain")
		}
		s.nodes[id].target = s.gen
	}
	return s.reaches(int32(node.NodeID()))
}

// reaches is Dch_ObjCheckTfi_rec: whether the fanin cone of node id,
// through choice alternatives, reaches a marked representative.
func (s *tfiScratch) reaches(id int32) bool {
	n := &s.nodes[id]
	if n.target == s.gen {
		return true
	}
	if n.seen == s.gen {
		return false
	}
	n.seen = s.gen
	if n.fanin0 >= 0 && s.reaches(n.fanin0) || n.fanin1 >= 0 && s.reaches(n.fanin1) {
		return true
	}
	n = &s.nodes[id] // (the slice does not move during a check)
	return n.equiv >= 0 && s.reaches(n.equiv)
}

// DeriveChoiceAig replaces proved constants/CI aliases and redirects fanins
// through representatives. Dangling alternate implementations are linked only
// when they cannot create a combinational/choice cycle.
func DeriveChoiceAig(source *aig.Graph, skipRedundantSupport bool) (*aig.Graph, error) {
	if source == nil {
		return nil, fmt.Errorf("dch: nil source AIG")
	}
	if err := source.Check(); err != nil {
		return nil, fmt.Errorf("dch: %w", err)
	}
	if !aig.Aig_ManHasReprs(source) {
		return nil, fmt.Errorf("dch: representative storage is required")
	}
	var tfi tfiScratch
	result := aig.NewGraph()
	aig.Aig_ManReprStart(result, int32(source.NodeCount()))
	aig.Aig_ManStartEquivs(result) // Dch_DeriveChoiceAig always allocates pEquivs
	copies := map[int]aig.Edge{0: result.Const1()}
	for _, input := range source.Inputs() {
		copies[input.NodeID()] = result.NewInput()
	}
	active := map[int]bool{}
	var copyEdge func(aig.Edge) (aig.Edge, error)
	copyEdge = func(edge aig.Edge) (aig.Edge, error) {
		regular := aig.Aig_Regular(edge)
		id := regular.NodeID()
		if old, ok := copies[id]; ok {
			return phaseEdge(old, edge.Complemented()), nil
		}
		if active[id] {
			return aig.InvalidEdge, fmt.Errorf("dch: representative recursion at %d", id)
		}
		active[id] = true
		defer delete(active, id)
		representative := aig.Aig_ObjRepr(source, regular)
		var repCopy aig.Edge = aig.InvalidEdge
		if representative != aig.InvalidEdge {
			if !source.ValidEdge(representative) || representative.Complemented() || representative.NodeID() >= id {
				return aig.InvalidEdge, fmt.Errorf("dch: invalid representative at %d", id)
			}
			copied, err := copyEdge(representative)
			if err != nil {
				return aig.InvalidEdge, err
			}
			repCopy = copied
			kind := source.Node(representative).Kind
			if kind == aig.Constant || kind == aig.Input {
				value := phaseEdge(repCopy, aig.Aig_ObjPhase(source, regular) != aig.Aig_ObjPhase(source, representative))
				copies[id] = value
				return phaseEdge(value, edge.Complemented()), nil
			}
		}
		n := source.Node(regular)
		a, err := copyEdge(n.Fanins[0])
		if err != nil {
			return aig.InvalidEdge, err
		}
		a = resolveRepresentative(result, a)
		var value aig.Edge
		switch n.Kind {
		case aig.Buffer, aig.Output:
			value = a
		case aig.And, aig.ExclusiveOr:
			b, err := copyEdge(n.Fanins[1])
			if err != nil {
				return aig.InvalidEdge, err
			}
			b = resolveRepresentative(result, b)
			if n.Kind == aig.And {
				value = result.AND(a, b)
			} else {
				value = result.NewXor(a, b)
			}
		default:
			return aig.InvalidEdge, fmt.Errorf("dch: unbound node %d", id)
		}
		value = resolveRepresentative(result, value)
		copies[id] = value
		if repCopy != aig.InvalidEdge {
			member, head := aig.Aig_Regular(value), aig.Aig_Regular(repCopy)
			memberKind, headKind := result.Node(member).Kind, result.Node(head).Kind
			if head.NodeID() < member.NodeID() && (memberKind == aig.And || memberKind == aig.ExclusiveOr) && (headKind == aig.And || headKind == aig.ExclusiveOr) {
				aig.Aig_ObjCreateRepr(result, head, member)
				if result.Node(member).Refs == 0 && !tfi.check(result, member, head) && (!skipRedundantSupport || !Dch_ObjCheckSuppRed(result, member, head)) {
					tail := head
					for next := aig.Aig_ObjEquiv(result, tail); next != aig.InvalidEdge; next = aig.Aig_ObjEquiv(result, tail) {
						tail = next
					}
					aig.Aig_ObjSetEquiv(result, tail, member)
					tfi.link(result, tail, member)
				}
			}
		}
		return phaseEdge(value, edge.Complemented()), nil
	}
	for _, node := range source.LiveEdges() {
		kind := source.Node(node).Kind
		if kind == aig.And || kind == aig.ExclusiveOr || kind == aig.Buffer {
			if _, err := copyEdge(node); err != nil {
				return nil, err
			}
		}
	}
	for _, output := range source.Outputs() {
		driver, err := copyEdge(output)
		if err != nil {
			return nil, err
		}
		result.AddOutput(resolveRepresentative(result, driver))
	}
	aig.Aig_ManReprStop(result)
	// DFS duplicates choices with their correct topological order, retaining
	// alternatives even though they are not ordinary output-fanin nodes.
	result = aig.Aig_ManDupDfs(result)
	if err := copyMetadata(result, source); err != nil {
		return nil, err
	}
	if err := CheckChoices(result, skipRedundantSupport); err != nil {
		return nil, fmt.Errorf("dch: invalid derived choices: %w", err)
	}
	return result, nil
}
func Dch_DeriveChoiceAig(g *aig.Graph, skip int) (*aig.Graph, error) {
	return DeriveChoiceAig(g, skip != 0)
}
func Dch_DeriveChoiceCountReprs(g *aig.Graph) int {
	n := 0
	for _, edge := range g.LiveEdges() {
		if repr := aig.Aig_ObjRepr(g, edge); repr != aig.InvalidEdge {
			assert(repr.NodeID() < edge.NodeID(), "representative precedes node")
			n++
		}
	}
	return n
}
func Dch_DeriveChoiceCountEquivs(g *aig.Graph) int {
	n := 0
	for _, edge := range g.LiveEdges() {
		if equiv := aig.Aig_ObjEquiv(g, edge); equiv != aig.InvalidEdge {
			assert(equiv.NodeID() < edge.NodeID(), "equivalent node precedes node")
			n++
		}
	}
	return n
}
