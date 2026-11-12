package cgt

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"math/bits"
	"math/rand"
)

func evaluateNodes(g *aig.Graph, inputs []bool) []bool {
	if len(inputs) != g.InputCount() {
		panic("cgt: wrong simulation input count")
	}
	values := make([]bool, len(g.Nodes()))
	done := make([]bool, len(values))
	for i, e := range g.Inputs() {
		values[e.NodeID()] = inputs[i]
		done[e.NodeID()] = true
	}
	var visit func(aig.Edge) bool
	visit = func(e aig.Edge) bool {
		id := e.NodeID()
		if done[id] {
			return values[id] != (e&1 != 0)
		}
		n := g.Node(e)
		var v bool
		switch n.Kind {
		case aig.Constant:
			v = true
		case aig.And:
			v = visit(n.Fanins[0]) && visit(n.Fanins[1])
		case aig.ExclusiveOr:
			v = visit(n.Fanins[0]) != visit(n.Fanins[1])
		case aig.Buffer, aig.Output:
			v = visit(n.Fanins[0])
		default:
			panic("cgt: unbound simulation input")
		}
		values[id], done[id] = v, true
		return v != (e&1 != 0)
	}
	for _, e := range g.LiveEdges() {
		visit(e)
	}
	return values
}
func Cgt_ManCollectFanoutPos(g *aig.Graph, root aig.Edge) []aig.Edge {
	index := fanoutIndex(g)
	seen := make(map[int]bool)
	var outs []aig.Edge
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		if seen[e.NodeID()] {
			return
		}
		seen[e.NodeID()] = true
		if g.Node(e).Kind == aig.Output {
			outs = append(outs, e)
			return
		}
		for _, out := range index[e.NodeID()] {
			visit(out)
		}
	}
	visit(root)
	return outs
}
func Cgt_ManCheckGateComplete(g *aig.Graph, all Gates, gate aig.Edge, fanouts []aig.Edge) bool {
	_, _, pos := layout(g)
	outputs := g.Outputs()
	positions := make(map[int]int)
	for i, e := range outputs {
		positions[e.NodeID()] = i
	}
	for _, out := range fanouts {
		index, ok := positions[out.NodeID()]
		if !ok || index < pos {
			return false
		}
		found := false
		for _, candidate := range all[index-pos] {
			found = found || candidate == gate
		}
		if !found {
			return false
		}
	}
	return true
}
func Cgt_ManCompleteGates(g *aig.Graph, all Gates, odcMax int, verbose bool) []aig.Edge {
	seen := make(map[aig.Edge]bool)
	var complete []aig.Edge
	for _, set := range all {
		for _, gate := range set {
			if seen[gate] {
				continue
			}
			seen[gate] = true
			if Cgt_ManCheckGateComplete(g, all, gate, Cgt_ManCollectFanoutPos(g, regular(gate))) {
				complete = append(complete, gate)
			}
		}
	}
	return complete
}
func sequentialSamples(g *aig.Graph) [][]uint32 {
	const frames = 32
	const lanes = 32
	regs, pis, pos := layout(g)
	states := make([][]bool, lanes)
	for i := range states {
		states[i] = make([]bool, regs)
	}
	samples := make([][]uint32, len(g.Nodes()))
	for i := range samples {
		samples[i] = make([]uint32, frames)
	}
	rng := rand.New(rand.NewSource(1))
	outs := g.Outputs()
	for frame := 0; frame < frames; frame++ {
		for lane := 0; lane < lanes; lane++ {
			in := make([]bool, g.InputCount())
			for i := 0; i < pis; i++ {
				in[i] = rng.Intn(2) != 0
			}
			copy(in[pis:], states[lane])
			values := evaluateNodes(g, in)
			for id, v := range values {
				if v {
					samples[id][frame] |= 1 << uint(lane)
				}
			}
			for r := 0; r < regs; r++ {
				e := outs[pos+r]
				states[lane][r] = values[e.NodeID()] != (e&1 != 0)
			}
		}
	}
	return samples
}
func sampleWord(samples [][]uint32, gate aig.Edge, word int) uint32 {
	x := samples[gate.NodeID()][word]
	if gate&1 != 0 {
		return ^x
	}
	return x
}
func Cgt_ManComputeCoverage(g *aig.Graph, gates Gates) float32 {
	if len(gates) == 0 {
		return 0
	}
	samples := sequentialSamples(g)
	total := 0
	for _, set := range gates {
		for word := 0; word < 32; word++ {
			var active uint32
			for _, gate := range set {
				active |= sampleWord(samples, gate, word)
			}
			total += bits.OnesCount32(active)
		}
	}
	return float32(100*total) / float32(1024*len(gates))
}
func Cgt_ManDecideSimple(g *aig.Graph, all Gates, odcMax int, verbose bool) Gates {
	chosen := make(Gates, len(all))
	samples := sequentialSamples(g)
	for r, set := range all {
		best := 0
		for _, gate := range set {
			score := 0
			for word := 0; word < 32; word++ {
				score += bits.OnesCount32(sampleWord(samples, gate, word))
			}
			if score > best {
				best = score
				chosen[r] = []aig.Edge{gate}
			}
		}
	}
	if verbose {
		fmt.Printf("CGT simple coverage: %.2f%%\n", Cgt_ManComputeCoverage(g, chosen))
	}
	return chosen
}
func Cgt_ManDecideArea(g *aig.Graph, all Gates, odcMax int, verbose bool) Gates {
	complete := Cgt_ManCompleteGates(g, all, odcMax, verbose)
	allowed := make(map[aig.Edge]bool)
	phase := make(map[int]aig.Edge)
	for _, gate := range complete {
		if _, seen := phase[gate.NodeID()]; !seen {
			phase[gate.NodeID()] = gate
			allowed[gate] = true
		}
	}
	selected := make(Gates, len(all))
	for r, set := range all {
		for _, gate := range set {
			if allowed[gate] {
				selected[r] = append(selected[r], gate)
			}
		}
	}
	if verbose {
		fmt.Printf("CGT area coverage: %.2f%%\n", Cgt_ManComputeCoverage(g, selected))
	}
	return selected
}
