package saig

import "github.com/glycerine/abcgo/aig/aig"

const SAIG_WORDS = 16

// strSimWords keeps the 16 consecutive simulation words for each object.
// Storage is local to one matching run rather than sharing AIG object data.
type strSimWords struct {
	graph *aig.Graph
	words [][SAIG_WORDS]uint32
}

func newStrSimWords(g *aig.Graph) *strSimWords {
	s := &strSimWords{graph: g, words: make([][SAIG_WORDS]uint32, len(g.Nodes()))}
	for i := range s.words[0] {
		s.words[0][i] = ^uint32(0)
	}
	return s
}

// frame evaluates unmatched gates and latch inputs for one timeframe. Matched
// gates are preloaded by the caller, like primary inputs and latch outputs.
func (s *strSimWords) frame(frame int, matched []bool) {
	if frame < 0 || frame >= SAIG_WORDS {
		panic("saig: invalid simulation frame")
	}
	g := s.graph
	done := make([]bool, len(s.words))
	active := make([]bool, len(s.words))
	var value func(aig.Edge) uint32
	value = func(edge aig.Edge) uint32 {
		id := edge.NodeID()
		if !done[id] {
			if active[id] {
				panic("saig: combinational simulation cycle")
			}
			active[id] = true
			node := g.Node(edge)
			if (id >= len(matched) || !matched[id]) && node.Kind != aig.Input && node.Kind != aig.Constant {
				switch node.Kind {
				case aig.And:
					s.words[id][frame] = value(node.Fanins[0]) & value(node.Fanins[1])
				case aig.ExclusiveOr:
					s.words[id][frame] = value(node.Fanins[0]) ^ value(node.Fanins[1])
				case aig.Buffer, aig.Output:
					s.words[id][frame] = value(node.Fanins[0])
				default:
					panic("saig: unsupported simulation object")
				}
			}
			active[id], done[id] = false, true
		}
		word := s.words[id][frame]
		if edge.Complemented() {
			word = ^word
		}
		return word
	}
	for _, edge := range g.LiveEdges() {
		node := g.Node(edge)
		if node.Kind != aig.Input && node.Kind != aig.Constant {
			value(edge)
		}
	}
	if frame+1 < SAIG_WORDS {
		for i := int32(0); i < Saig_ManRegNum(g); i++ {
			li, lo := Saig_ManLi(g, i), Saig_ManLo(g, i)
			s.words[lo.NodeID()][frame+1] = s.words[li.NodeID()][frame]
		}
	}
}
