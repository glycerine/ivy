package saig

import (
	"math/rand"

	"github.com/glycerine/abcgo/aig/aig"
)

// strSimMatcher keeps cross-graph correspondence separate from each graph's
// internal representative table: equal edge IDs in different graphs are not
// the same object.
type strSimMatcher struct {
	sim    [2]*strSimWords
	pair   [2]map[aig.Edge]aig.Edge
	random *rand.Rand
}

func newStrSimMatcher(a, b *aig.Graph) *strSimMatcher {
	if a == nil || b == nil || Saig_ManPiNum(a) != Saig_ManPiNum(b) {
		panic("saig: incompatible simulation inputs")
	}
	m := &strSimMatcher{sim: [2]*strSimWords{newStrSimWords(a), newStrSimWords(b)}, pair: [2]map[aig.Edge]aig.Edge{make(map[aig.Edge]aig.Edge), make(map[aig.Edge]aig.Edge)}, random: rand.New(rand.NewSource(1))}
	m.link(a.Const1(), b.Const1())
	for i := int32(0); i < Saig_ManPiNum(a); i++ {
		m.link(aig.Aig_ManCi(a, i), aig.Aig_ManCi(b, i))
	}
	return m
}

func (m *strSimMatcher) link(a, b aig.Edge) {
	m.pair[0][a], m.pair[1][b] = b, a
}

func (m *strSimMatcher) round() {
	matched := [2][]bool{make([]bool, len(m.sim[0].words)), make([]bool, len(m.sim[1].words))}
	// Iterate graph order so that random signatures are deterministic.
	for _, a := range m.sim[0].graph.LiveEdges() {
		b, ok := m.pair[0][a]
		if !ok {
			continue
		}
		matched[0][a.NodeID()], matched[1][b.NodeID()] = true, true
		kind := m.sim[0].graph.Node(a).Kind
		if kind != aig.Input && kind != aig.And && kind != aig.ExclusiveOr {
			continue
		}
		for i := 0; i < SAIG_WORDS; i++ {
			word := m.random.Uint32()
			m.sim[0].words[a.NodeID()][i], m.sim[1].words[b.NodeID()][i] = word, word
		}
	}
	for frame := 0; frame < SAIG_WORDS; frame++ {
		m.sim[0].frame(frame, matched[0])
		m.sim[1].frame(frame, matched[1])
	}
}

// detectUnique pairs only signatures occurring once in each graph and with
// equal object kinds, excluding all-zero and all-one signatures as ABC does.
func (m *strSimMatcher) detectUnique() int {
	type candidate struct {
		edge  aig.Edge
		count int
	}
	var indexes [2]map[[SAIG_WORDS]uint32]candidate
	zero := [SAIG_WORDS]uint32{}
	one := zero
	for i := range one {
		one[i] = ^uint32(0)
	}
	for side := 0; side < 2; side++ {
		indexes[side] = make(map[[SAIG_WORDS]uint32]candidate)
		s := m.sim[side]
		for _, edge := range s.graph.LiveEdges() {
			kind := s.graph.Node(edge).Kind
			if kind != aig.Input && kind != aig.And && kind != aig.ExclusiveOr {
				continue
			}
			if _, ok := m.pair[side][edge]; ok {
				continue
			}
			signature := s.words[edge.NodeID()]
			if signature == zero || signature == one {
				continue
			}
			entry := indexes[side][signature]
			entry.edge, entry.count = edge, entry.count+1
			indexes[side][signature] = entry
		}
	}
	count := 0
	for signature, a := range indexes[0] {
		b := indexes[1][signature]
		if a.count != 1 || b.count != 1 || m.sim[0].graph.Node(a.edge).Kind != m.sim[1].graph.Node(b.edge).Kind {
			continue
		}
		m.link(a.edge, b.edge)
		count++
	}
	return count
}

// pruneInconsistent removes pairs whose immediate structural predecessors do
// not correspond. The decision is simultaneous, matching the original pass.
func (m *strSimMatcher) pruneInconsistent() int {
	var remove []aig.Edge
	for a, b := range m.pair[0] {
		ga, gb := m.sim[0].graph, m.sim[1].graph
		na, nb := ga.Node(a), gb.Node(b)
		if na.Kind == aig.And || na.Kind == aig.ExclusiveOr {
			if na.Kind != nb.Kind {
				remove = append(remove, a)
				continue
			}
			for i := 0; i < 2; i++ {
				match, ok := m.pair[0][aig.Aig_Regular(na.Fanins[i])]
				if !ok || match != aig.Aig_Regular(nb.Fanins[i]) {
					remove = append(remove, a)
					break
				}
			}
		} else if Saig_ObjIsLo(ga, a) {
			if !Saig_ObjIsLo(gb, b) {
				remove = append(remove, a)
				continue
			}
			fa := aig.Aig_ObjFanin0(ga, Saig_ObjLoToLi(ga, a))
			fb := aig.Aig_ObjFanin0(gb, Saig_ObjLoToLi(gb, b))
			match, ok := m.pair[0][fa]
			if !ok || match != fb {
				remove = append(remove, a)
			}
		}
	}
	for _, a := range remove {
		b := m.pair[0][a]
		delete(m.pair[0], a)
		delete(m.pair[1], b)
	}
	return len(remove)
}

// boundary returns matched objects adjacent to an unmatched object, treating
// each latch input and its corresponding output as adjacent.
func (m *strSimMatcher) boundary(side int) map[aig.Edge]bool {
	g, pairs := m.sim[side].graph, m.pair[side]
	fanouts := make(map[aig.Edge][]aig.Edge)
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		var count int
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			count = 2
		case aig.Buffer, aig.Output:
			count = 1
		}
		for i := 0; i < count; i++ {
			f := aig.Aig_Regular(n.Fanins[i])
			fanouts[f] = append(fanouts[f], e)
		}
	}
	result := make(map[aig.Edge]bool)
	add := func(e aig.Edge) {
		if _, ok := pairs[e]; ok {
			result[e] = true
		}
	}
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		if n.Kind != aig.Input && n.Kind != aig.And && n.Kind != aig.ExclusiveOr {
			continue
		}
		if _, ok := pairs[e]; ok {
			continue
		}
		if Saig_ObjIsLo(g, e) {
			f := aig.Aig_ObjFanin0(g, Saig_ObjLoToLi(g, e))
			if f != g.Const1() {
				add(f)
			}
		}
		if n.Kind == aig.And || n.Kind == aig.ExclusiveOr {
			add(aig.Aig_Regular(n.Fanins[0]))
			add(aig.Aig_Regular(n.Fanins[1]))
		}
		for _, f := range fanouts[e] {
			if Saig_ObjIsPo(g, f) {
				continue
			}
			if Saig_ObjIsLi(g, f) {
				f = Saig_ObjLiToLo(g, f)
			}
			add(f)
		}
	}
	return result
}

func (m *strSimMatcher) extend(distance int) {
	for i := 0; i < distance; i++ {
		boundary := [2]map[aig.Edge]bool{m.boundary(0), m.boundary(1)}
		for side := 0; side < 2; side++ {
			for edge := range boundary[side] {
				other, ok := m.pair[side][edge]
				if !ok || Saig_ObjIsPi(m.sim[1-side].graph, other) {
					continue
				}
				delete(m.pair[side], edge)
				delete(m.pair[1-side], other)
			}
		}
	}
}
