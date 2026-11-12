package sim

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/bits"
)

// Sim_ComputeStrSupp returns node-ID-indexed structural supports, including COs.
func Sim_ComputeStrSupp(g *aig.Graph) Info {
	order := graphOrder(g)
	words := SIM_NUM_WORDS(len(g.Inputs()))
	r := Sim_UtilInfoAlloc(int(aig.Aig_ManObjNumMax(g)), words, true)
	for i, e := range g.Inputs() {
		Sim_SetBit(r[e.NodeID()], i)
	}
	for _, e := range order {
		n := g.Node(e)
		switch n.Kind {
		case aig.Buffer, aig.Output:
			copy(r[e.NodeID()], r[n.Fanins[0].NodeID()])
		case aig.And, aig.ExclusiveOr:
			for w := 0; w < words; w++ {
				r[e.NodeID()][w] = r[n.Fanins[0].NodeID()][w] | r[n.Fanins[1].NodeID()][w]
			}
		}
	}
	return r
}

// Sim_ComputeFunSupp returns CO-indexed exact functional supports. Simulation
// establishes positive dependencies; SAT removes the remaining false targets.
func Sim_ComputeFunSupp(g *aig.Graph, verbose bool) Info {
	p := Sim_ManStart(g, false)
	Sim_UtilAssignRandom(p)
	Sim_ComputeSuppRound(p, false)
	Sim_ComputeSuppSetTargets(p)
	Sim_UtilAssignRandom(p)
	Sim_ComputeSuppRound(p, true)
	for targetCount(p) > 0 {
		Sim_SolveTargetsUsingSat(p, max(1, p.Words/max(1, SIM_NUM_WORDS(len(p.Inputs)))))
		if len(p.Patterns) > 0 {
			Sim_UtilAssignFromFifo(p)
			Sim_ComputeSuppRound(p, true)
		}
	}
	if verbose {
		Sim_ManPrintStats(p)
	}
	return p.Functional
}
func targetCount(p *Sim_Man_t) int {
	n := 0
	for _, t := range p.Targets {
		n += len(t)
	}
	return n
}
func Sim_UtilAssignRandom(p *Sim_Man_t) {
	for _, e := range p.Inputs {
		fillRandom(p.Sim0[e.NodeID()], p.rng)
	}
}
func Sim_ComputeSuppRound(p *Sim_Man_t, useTargets bool) int {
	Sim_UtilSimulate(p, false)
	n := 0
	for i := p.Input; i < len(p.Inputs); i++ {
		if !useTargets || len(p.Targets[i]) > 0 {
			n += Sim_ComputeSuppRoundNode(p, i, useTargets)
		}
	}
	return n
}
func Sim_ComputeSuppRoundNode(p *Sim_Man_t, input int, useTargets bool) int {
	// Only the input's transitive fanout is recomputed. Unaffected fanins read
	// the baseline image, matching the source's mixed-image simulation.
	affected := make([]bool, len(p.Sim0))
	affected[p.Inputs[input].NodeID()] = true
	Sim_UtilInfoFlip(p, p.Inputs[input])
	for _, e := range p.Order {
		node := p.Graph.Node(e)
		if node.Kind == aig.Constant || node.Kind == aig.Input {
			continue
		}
		left := affected[node.Fanins[0].NodeID()]
		right := false
		if node.Kind == aig.And || node.Kind == aig.ExclusiveOr {
			right = affected[node.Fanins[1].NodeID()]
		}
		if left || right {
			affected[e.NodeID()] = true
			Sim_UtilSimulateNode(p, e, true, left, right)
		}
	}

	n := 0
	remaining := p.Targets[input][:0]
	visit := func(o int) bool {
		e := p.Outputs[o]
		if !affected[e.NodeID()] {
			return false
		}
		a, b := p.Sim0[e.NodeID()], p.Sim1[e.NodeID()]
		if Sim_UtilInfoIsEqual(a, b, p.Words) {
			return false
		}
		if !Sim_HasBit(p.Functional[o], input) {
			Sim_SetBit(p.Functional[o], input)
			n++
		}
		return true
	}
	if useTargets {
		for _, o := range p.Targets[input] {
			if !visit(o) {
				remaining = append(remaining, o)
			}
		}
		p.Targets[input] = remaining
	} else {
		for o := range p.Outputs {
			visit(o)
		}
	}
	return n
}
func Sim_ComputeSuppSetTargets(p *Sim_Man_t) {
	for i := range p.Targets {
		p.Targets[i] = p.Targets[i][:0]
	}
	for o, e := range p.Outputs {
		for _, i := range Sim_UtilInfoDetectNews(p.Functional[o], p.Structural[e.NodeID()], SIM_NUM_WORDS(len(p.Inputs))) {
			p.Targets[i] = append(p.Targets[i], o)
		}
	}
}
func Sim_SolveTargetsUsingSat(p *Sim_Man_t, limit int) {
	if limit <= 0 {
		panic("nonpositive target batch")
	}
	found := 0
	for i := len(p.Inputs) - 1; i >= 0; i-- {
		for len(p.Targets[i]) > 0 {
			o := p.Targets[i][len(p.Targets[i])-1]
			p.Targets[i] = p.Targets[i][:len(p.Targets[i])-1]
			p.SATRuns++
			different, pat := cofactorDifference(p.Graph, o, i, -1)
			if different {
				p.SAT++
				Sim_SetBit(p.Functional[o], i)
				p.Patterns = append(p.Patterns, &Sim_Pat_t{Input: i, Output: o, Data: pat})
				found++
				if found == limit {
					return
				}
			} else {
				p.UNSAT++
			}
		}
	}
}

// Counterexamples are expanded to distance-one neighborhoods. If CI count
// exceeds the batch width, only fitting flips are inserted (never past a row).
func Sim_UtilAssignFromFifo(p *Sim_Man_t) {
	Sim_UtilAssignRandom(p)
	block := max(1, SIM_NUM_WORDS(len(p.Inputs)))
	for offset := 0; offset < p.Words && len(p.Patterns) > 0; offset += block {
		last := len(p.Patterns) - 1
		pat := p.Patterns[last]
		p.Patterns = p.Patterns[:last]
		end := min(p.Words, offset+block)
		for i, e := range p.Inputs {
			row := p.Sim0[e.NodeID()]
			Sim_UtilSetConst(row[offset:end], end-offset, Sim_HasBit(pat.Data, i))
			if i < 32*(end-offset) {
				Sim_XorBit(row[offset:end], i)
			}
		}
	}
}
func Sim_SolveSuppModelVerify(g *aig.Graph, model []bool, input, output int) bool {
	if len(model) != g.InputCount() {
		panic("model width mismatch")
	}
	a, err := g.Evaluate(model)
	if err != nil {
		panic(err)
	}
	other := append([]bool(nil), model...)
	other[input] = !other[input]
	b, err := g.Evaluate(other)
	if err != nil {
		panic(err)
	}
	return a[output] != b[output]
}

// firstDifference is shared by source-oracle tests and witness extraction.
func firstDifference(a, b []uint32) int {
	for w := range a {
		if x := a[w] ^ b[w]; x != 0 {
			return 32*w + bits.TrailingZeros32(x)
		}
	}
	return -1
}
