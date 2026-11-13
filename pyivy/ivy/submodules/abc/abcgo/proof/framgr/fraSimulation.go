package framgr

// Bit-parallel FRA simulation primitives, adapted from fraSim.c. A node's
// internal words are stored modulo its AIG phase, matching the source's
// equivalence-class hashing convention. Output words hold actual values.

import (
	"math/bits"

	"github.com/glycerine/abcgo/aig/aig"
)

type fraSimulation struct {
	graph                                                        *aig.Graph
	prefixFrames, frames, wordsPerFrame, wordsTotal, prefixWords int
	words                                                        []uint32
	rounds                                                       int
	nonconstantOutput                                            bool
}

func newFraSimulation(graph *aig.Graph, prefix, frames, wordsPerFrame int) *fraSimulation {
	if graph == nil || prefix < 0 || frames <= 0 || wordsPerFrame <= 0 {
		panic("fra: invalid simulation dimensions")
	}
	totalFrames := prefix + frames
	s := &fraSimulation{graph: graph, prefixFrames: prefix, frames: totalFrames, wordsPerFrame: wordsPerFrame,
		wordsTotal: totalFrames * wordsPerFrame, prefixWords: prefix * wordsPerFrame}
	s.words = make([]uint32, len(graph.Nodes())*s.wordsTotal)
	return s
}

func (s *fraSimulation) nodeWords(id int) []uint32 {
	if id < 0 || id >= len(s.words)/s.wordsTotal {
		panic("fra: invalid simulation node")
	}
	return s.words[id*s.wordsTotal : (id+1)*s.wordsTotal]
}
func (s *fraSimulation) frameWords(id, frame int) []uint32 {
	if frame < 0 || frame >= s.frames {
		panic("fra: invalid simulation frame")
	}
	words := s.nodeWords(id)
	return words[frame*s.wordsPerFrame : (frame+1)*s.wordsPerFrame]
}

var fraSimulationPrimes = [...]uint32{
	1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459, 1499, 1559, 1607, 1657, 1709, 1759,
	1823, 1877, 1933, 1997, 2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543, 2609, 2663,
	2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089, 3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557,
	3613, 3671, 3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243, 4289, 4363, 4441, 4493,
	4549, 4621, 4663, 4729, 4793, 4871, 4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
	5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073, 6131, 6199, 6257, 6301, 6353, 6397,
	6481, 6563, 6619, 6689, 6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309, 7393, 7477,
	7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933, 8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
}

// Hash, IsConstant and Equal provide the callbacks for classPartition.
func (s *fraSimulation) Hash(id int) uint64 {
	var hash uint32
	words := s.nodeWords(id)
	for i := s.prefixWords; i < s.wordsTotal; i++ {
		hash ^= words[i] * fraSimulationPrimes[i&127]
	}
	return uint64(hash)
}
func (s *fraSimulation) IsConstant(id int) bool {
	for _, word := range s.nodeWords(id)[s.prefixWords:] {
		if word != 0 {
			return false
		}
	}
	return true
}
func (s *fraSimulation) Equal(a, b int) bool {
	x, y := s.nodeWords(a), s.nodeWords(b)
	for i := s.prefixWords; i < s.wordsTotal; i++ {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}
func (s *fraSimulation) DifferenceWeight(a, b int) int {
	x, y := s.nodeWords(a), s.nodeWords(b)
	count := 0
	for i := s.prefixWords; i < s.wordsTotal; i++ {
		count += bits.OnesCount32(x[i] ^ y[i])
	}
	return count
}
func (s *fraSimulation) CountOnes(id int) int {
	count := 0
	for _, word := range s.nodeWords(id) {
		count += bits.OnesCount32(word)
	}
	return count
}
func (s *fraSimulation) assignFrame(id, frame int, value bool) {
	var word uint32
	if value {
		word = ^uint32(0)
	}
	words := s.frameWords(id, frame)
	for i := range words {
		words[i] = word
	}
}
func (s *fraSimulation) initialize(initialState bool) {
	inputs := s.graph.Inputs()
	registers := int(aig.Aig_ManRegNum(s.graph))
	if initialState && (registers == 0 || registers >= len(inputs)) {
		panic("fra: invalid sequential initial state")
	}
	for i, input := range inputs {
		if initialState && i >= len(inputs)-registers {
			s.assignFrame(input.NodeID(), 0, false)
			continue
		}
		words := s.nodeWords(input.NodeID())
		for j := range words {
			words[j] = aig.Aig_ManRandom(0)
		}
	}
}

// assignDistance1 packs a counterexample and, for one frame, flips one input
// in each successive bit, following Fra_SmlAssignDist1.
func (s *fraSimulation) assignDistance1(pattern []uint32) {
	inputs := s.graph.Inputs()
	registers := int(aig.Aig_ManRegNum(s.graph))
	trueInputs := len(inputs) - registers
	required := len(inputs)
	if s.frames > 1 {
		required = trueInputs*s.frames + registers
	}
	if len(pattern) < (required+31)/32 {
		panic("fra: short simulation pattern")
	}
	bit := func(i int) bool { return pattern[i/32]&(uint32(1)<<uint(i%32)) != 0 }
	if s.frames == 1 {
		for i, input := range inputs {
			s.assignFrame(input.NodeID(), 0, bit(i))
		}
		limit := min(len(inputs), s.wordsTotal*32-1)
		for i := 0; i < limit; i++ {
			words := s.nodeWords(inputs[i].NodeID())
			words[(i+1)/32] ^= uint32(1) << uint((i+1)%32)
		}
		return
	}
	for frame := 0; frame < s.frames; frame++ {
		for i := 0; i < trueInputs; i++ {
			s.assignFrame(inputs[i].NodeID(), frame, bit(trueInputs*frame+i))
		}
	}
	for i := 0; i < registers; i++ {
		s.assignFrame(inputs[trueInputs+i].NodeID(), 0, bit(trueInputs*s.frames+i))
	}
}

// initializeGiven loads input-major values from consecutive test patterns and
// pads the last machine word with the final pattern, as in fraSim.c.
func (s *fraSimulation) initializeGiven(patterns []byte) {
	inputs := s.graph.Inputs()
	if aig.Aig_ManRegNum(s.graph) != 0 || len(inputs) == 0 || len(patterns) == 0 || len(patterns)%len(inputs) != 0 {
		panic("fra: invalid combinational simulation patterns")
	}
	count := len(patterns) / len(inputs)
	if count > s.wordsTotal*32 {
		panic("fra: too many simulation patterns")
	}
	for _, value := range patterns {
		if value > 1 {
			panic("fra: invalid simulation pattern bit")
		}
	}
	for i, input := range inputs {
		words := s.nodeWords(input.NodeID())
		clear(words)
		for k := 0; k < s.wordsTotal*32; k++ {
			index := min(k, count-1)
			if patterns[index*len(inputs)+i] != 0 {
				words[k/32] |= uint32(1) << uint(k%32)
			}
		}
	}
}

func (s *fraSimulation) hasNonconstantOutput() bool {
	outputs := s.graph.Outputs()
	primaryOutputs := len(outputs) - int(aig.Aig_ManRegNum(s.graph))
	for _, output := range outputs[:primaryOutputs] {
		if !s.IsConstant(output.NodeID()) {
			return true
		}
	}
	return false
}

// simulateOne performs node evaluation and latch transfer in frame order.
func (s *fraSimulation) simulateOne() {
	nodes := s.graph.Nodes()
	outputs := s.graph.Outputs()
	registers := int(aig.Aig_ManRegNum(s.graph))
	for frame := 0; frame < s.frames; frame++ {
		for _, edge := range s.graph.LiveEdges() {
			id := edge.NodeID()
			node := nodes[id]
			if node.Kind == aig.And || node.Kind == aig.ExclusiveOr || node.Kind == aig.Buffer {
				s.simulateNode(edge, frame)
			}
		}
		for _, out := range outputs {
			node := nodes[out.NodeID()]
			src := s.frameWords(node.Fanins[0].NodeID(), frame)
			dst := s.frameWords(out.NodeID(), frame)
			flip := nodes[node.Fanins[0].NodeID()].Phase != node.Fanins[0].Complemented()
			for i := range dst {
				dst[i] = src[i]
				if flip {
					dst[i] = ^dst[i]
				}
			}
		}
		if frame+1 < s.frames {
			for i := 0; i < registers; i++ {
				li := aig.Aig_ManLi(s.graph, int32(i))
				lo := aig.Aig_ManLo(s.graph, int32(i))
				copy(s.frameWords(lo.NodeID(), frame+1), s.frameWords(li.NodeID(), frame))
			}
		}
	}
	s.rounds++
}

// simulateNode updates one phase-normalized AIG node in a single frame.
func (s *fraSimulation) simulateNode(edge aig.Edge, frame int) {
	nodes := s.graph.Nodes()
	node := nodes[edge.NodeID()]
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
		panic("fra: cannot simulate non-node")
	}
	dst := s.frameWords(edge.NodeID(), frame)
	a := s.frameWords(node.Fanins[0].NodeID(), frame)
	var b []uint32
	if node.Kind != aig.Buffer {
		b = s.frameWords(node.Fanins[1].NodeID(), frame)
	}
	for i := range dst {
		left := a[i]
		if nodes[node.Fanins[0].NodeID()].Phase != node.Fanins[0].Complemented() {
			left = ^left
		}
		value := left
		if b != nil {
			right := b[i]
			if nodes[node.Fanins[1].NodeID()].Phase != node.Fanins[1].Complemented() {
				right = ^right
			}
			if node.Kind == aig.And {
				value = left & right
			} else {
				value = left ^ right
			}
		}
		if node.Phase {
			value = ^value
		}
		dst[i] = value
	}
}
