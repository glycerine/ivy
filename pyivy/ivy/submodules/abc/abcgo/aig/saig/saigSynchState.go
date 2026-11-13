package saig

import (
	"math/rand"

	"github.com/glycerine/abcgo/aig/aig"
)

// synchState simulates 16 independent ternary patterns per word. Each pair of
// bits encodes zero as 00, one as 01, and unknown as 11.
type synchState struct {
	graph  *aig.Graph
	words  [][]uint32
	random *rand.Rand
}

func newSynchState(g *aig.Graph, nwords int) *synchState {
	if g == nil || nwords <= 0 {
		panic("saig: invalid synchronizer simulation")
	}
	s := &synchState{graph: g, words: make([][]uint32, len(g.Nodes())), random: rand.New(rand.NewSource(1))}
	for i := range s.words {
		s.words[i] = make([]uint32, nwords)
	}
	for i := range s.words[0] {
		s.words[0][i] = 0x55555555
	}
	return s
}

func synchNot(w uint32) uint32 { return w ^ (^(w & (w >> 1)) & 0x55555555) }
func synchAnd(u, w uint32) uint32 {
	return (u & w) | ((((u & (u >> 1) & w &^ (w >> 1)) | (w & (w >> 1) & u &^ (u >> 1))) & 0x55555555) << 1)
}
func synchValue(v uint32) uint32 {
	switch v {
	case 0:
		return 0
	case 1:
		return 0x55555555
	case 3:
		return ^uint32(0)
	default:
		panic("saig: invalid ternary value")
	}
}

func (s *synchState) initRegisters(unknown bool) {
	for i := int32(0); i < Saig_ManRegNum(s.graph); i++ {
		e := Saig_ManLo(s.graph, i)
		value := uint32(0)
		if unknown {
			value = 3
		} else if s.graph.Node(e).MarkA {
			value = 1
		}
		for w := range s.words[e.NodeID()] {
			s.words[e.NodeID()][w] = synchValue(value)
		}
	}
}
func (s *synchState) randomInputs() {
	for i := int32(0); i < Saig_ManPiNum(s.graph); i++ {
		e := aig.Aig_ManCi(s.graph, i)
		for w := range s.words[e.NodeID()] {
			s.words[e.NodeID()][w] = s.random.Uint32() & 0x55555555
		}
	}
}
func (s *synchState) givenInputs(values []byte) {
	if len(values) != int(Saig_ManPiNum(s.graph)) {
		panic("saig: incorrect input pattern length")
	}
	for i, v := range values {
		e := aig.Aig_ManCi(s.graph, int32(i))
		for w := range s.words[e.NodeID()] {
			s.words[e.NodeID()][w] = synchValue(uint32(v))
		}
	}
}
func (s *synchState) simulate() {
	g := s.graph
	done := make([]bool, len(s.words))
	active := make([]bool, len(s.words))
	var visit func(aig.Edge, int) uint32
	for word := range s.words[0] {
		clear(done)
		clear(active)
		visit = func(e aig.Edge, w int) uint32 {
			id := e.NodeID()
			if !done[id] {
				if active[id] {
					panic("saig: cycle in synchronization simulation")
				}
				active[id] = true
				n := g.Node(e)
				switch n.Kind {
				case aig.And:
					s.words[id][w] = synchAnd(visit(n.Fanins[0], w), visit(n.Fanins[1], w))
				case aig.ExclusiveOr:
					a, b := visit(n.Fanins[0], w), visit(n.Fanins[1], w)
					s.words[id][w] = synchNot(synchAnd(synchNot(synchAnd(a, synchNot(b))), synchNot(synchAnd(synchNot(a), b))))
				case aig.Buffer, aig.Output:
					s.words[id][w] = visit(n.Fanins[0], w)
				case aig.Input, aig.Constant:
				default:
					panic("saig: unsupported synchronization object")
				}
				done[id], active[id] = true, false
			}
			v := s.words[id][w]
			if e.Complemented() {
				v = synchNot(v)
			}
			return v
		}
		for i := int32(0); i < Saig_ManRegNum(g); i++ {
			visit(Saig_ManLi(g, i), word)
		}
	}
}
func (s *synchState) transfer() {
	for i := int32(0); i < Saig_ManRegNum(s.graph); i++ {
		li, lo := Saig_ManLi(s.graph, i), Saig_ManLo(s.graph, i)
		copy(s.words[lo.NodeID()], s.words[li.NodeID()])
	}
}
func (s *synchState) bestPattern() (int, int) {
	counts := make([]int, len(s.words[0])*16)
	for i := int32(0); i < Saig_ManRegNum(s.graph); i++ {
		e := Saig_ManLi(s.graph, i)
		for w, value := range s.words[e.NodeID()] {
			for bit := 0; bit < 16; bit++ {
				if value>>uint(bit*2)&3 == 3 {
					counts[w*16+bit]++
				}
			}
		}
	}
	best := 0
	for i := 1; i < len(counts); i++ {
		if counts[i] < counts[best] {
			best = i
		}
	}
	return best, counts[best]
}
func (s *synchState) savePattern(pattern int) ([]byte, int) {
	if pattern < 0 || pattern >= 16*len(s.words[0]) {
		panic("saig: invalid synchronization pattern")
	}
	word, shift := pattern/16, uint(2*(pattern%16))
	inputs := make([]byte, Saig_ManPiNum(s.graph))
	for i := range inputs {
		e := aig.Aig_ManCi(s.graph, int32(i))
		inputs[i] = byte(s.words[e.NodeID()][word] >> shift & 3)
	}
	unknown := 0
	for i := int32(0); i < Saig_ManRegNum(s.graph); i++ {
		li, lo := Saig_ManLi(s.graph, i), Saig_ManLo(s.graph, i)
		v := s.words[li.NodeID()][word] >> shift & 3
		if v == 3 {
			unknown++
		}
		for w := range s.words[lo.NodeID()] {
			s.words[lo.NodeID()][w] = synchValue(v)
		}
	}
	return inputs, unknown
}
