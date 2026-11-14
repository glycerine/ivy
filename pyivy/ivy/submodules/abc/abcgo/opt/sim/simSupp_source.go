package sim

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
)

// SourceSupportSolver proves whether toggling input for output has a
// counterexample. A satisfiable result returns that input assignment as words.
type SourceSupportSolver func(input, output int) (satisfiable bool, model []uint32)

// Sim_ComputeFunSuppSource follows simSupp.c's round, SAT, and FIFO schedule.
// It recreates the glibc rand stream privately; it does not mutate libc's
// process-global RNG. With a nil solver it uses the package's exact
// two-cofactor solver; C compatibility supplies ABC's FRAIG callback.
func Sim_ComputeFunSuppSource(g *aig.Graph, verbose bool) Info {
	return Sim_ComputeFunSuppSourceWithSolver(g, verbose, nil)
}

func Sim_ComputeFunSuppSourceWithSolver(g *aig.Graph, verbose bool, solver SourceSupportSolver) Info {
	p := Sim_ManStart(g, false)
	rng := newSourceRand(0xABC)
	sourceAssignRandom(p, rng)
	sourceComputeSuppRound(p, false)
	Sim_ComputeSuppSetTargets(p)
	if verbose {
		fmt.Printf("Number of support targets after simulation = %5d.\n", targetCount(p))
	}
	if targetCount(p) == 0 {
		return p.Functional
	}

	sourceAssignRandom(p, rng)
	nSolved := sourceComputeSuppRound(p, true)
	if targetCount(p) == 0 {
		return p.Functional
	}
	if verbose {
		fmt.Printf("Targets = %5d.   Solved = %5d.  Fifo = %5d.\n", targetCount(p), nSolved, len(p.Patterns))
	}

	for targetCount(p) > 0 {
		sourceSolveTargetsUsingSat(p, p.Words/SIM_NUM_WORDS(len(p.Inputs)), solver)
		sourceAssignFromFifo(p)
		nSolved = sourceComputeSuppRound(p, true)
		if verbose {
			fmt.Printf("Targets = %5d.   Solved = %5d.  Fifo = %5d.  SAT runs = %3d.\n", targetCount(p), nSolved, len(p.Patterns), p.SATRuns)
		}
	}
	return p.Functional
}

func sourceComputeSuppRound(p *Sim_Man_t, useTargets bool) int {
	Sim_UtilSimulate(p, false)
	n := 0
	for i := range p.Inputs {
		if useTargets && len(p.Targets[i]) == 0 {
			continue
		}
		n += sourceComputeSuppRoundNode(p, i, useTargets)
	}
	return n
}

func sourceComputeSuppRoundNode(p *Sim_Man_t, input int, useTargets bool) int {
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

	count := 0
	firstPattern := true
	check := func(output int) {
		e := p.Outputs[output]
		if !affected[e.NodeID()] || Sim_UtilInfoCompare(p, e) {
			return
		}
		if !Sim_HasBit(p.Functional[output], input) {
			Sim_SetBit(p.Functional[output], input)
			count++
		}
		if !useTargets {
			return
		}
		if !firstPattern && len(p.Patterns) > 1000 {
			return
		}
		// C walks each target vector in reverse order and saves its first
		// differing simulation lane as a counterexample pattern.
		for lane := 0; lane < p.Words*32; lane++ {
			if Sim_HasBit(p.Sim0[e.NodeID()], lane) == Sim_HasBit(p.Sim1[e.NodeID()], lane) {
				continue
			}
			pat := Sim_ManPatAlloc(p)
			pat.Input, pat.Output = input, output
			for i, ci := range p.Inputs {
				if Sim_HasBit(p.Sim0[ci.NodeID()], lane) {
					Sim_SetBit(pat.Data, i)
				}
			}
			p.Patterns = append(p.Patterns, pat)
			firstPattern = false
			break
		}
	}

	if useTargets {
		targets := p.Targets[input]
		remaining := targets[:0]
		for i := len(targets) - 1; i >= 0; i-- {
			output := targets[i]
			if !affected[p.Outputs[output].NodeID()] || Sim_UtilInfoCompare(p, p.Outputs[output]) {
				remaining = append(remaining, output)
				continue
			}
			check(output)
		}
		// Reversing into the backing slice above would alter the source order;
		// restore ascending order for the next source-style reverse traversal.
		reverseInts(remaining)
		p.Targets[input] = remaining
	} else {
		for output := range p.Outputs {
			check(output)
		}
	}
	return count
}

func sourceSolveTargetsUsingSat(p *Sim_Man_t, limit int, solver SourceSupportSolver) {
	// C passes nSimWords/nSuppWords, but its Counter == 1 return is active and
	// the `Counter == Limit` condition is commented out. Keep the passed source
	// batch value for parity while intentionally stopping at one SAT witness.
	_ = limit
	p.SATRuns = 0
	for input := range p.Targets {
		for len(p.Targets[input]) > 0 {
			last := len(p.Targets[input]) - 1
			output := p.Targets[input][last]
			p.SATRuns++
			different, model := false, []uint32(nil)
			if solver == nil {
				different, model = cofactorDifference(p.Graph, output, input, -1)
			} else {
				different, model = solver(input, output)
			}
			if !different {
				p.UNSAT++
				p.Targets[input] = append(p.Targets[input][:last], p.Targets[input][last+1:]...)
				continue
			}
			p.SAT++
			Sim_SetBit(p.Functional[output], input)
			p.Patterns = append(p.Patterns, &Sim_Pat_t{Input: input, Output: output, Data: model})
			p.Input = input
			return // C stops at the first SAT target; Limit is commented out.
		}
	}
}

func sourceAssignFromFifo(p *Sim_Man_t) {
	for offset := 0; offset < p.Words && len(p.Patterns) > 0; {
		last := len(p.Patterns) - 1
		pat := p.Patterns[last]
		p.Patterns = p.Patterns[:last]
		words := SIM_NUM_WORDS(len(p.Inputs))
		end := min(offset+words, p.Words)
		for i, e := range p.Inputs {
			row := p.Sim0[e.NodeID()]
			value := uint32(0)
			if Sim_HasBit(pat.Data, i) {
				value = ^uint32(0)
			}
			for w := offset; w < end; w++ {
				row[w] = value
			}
			if i >= 32*(p.Words-offset) {
				panic("sim: source FIFO flip would exceed simulation row")
			}
			Sim_XorBit(row[offset:], i)
		}
		Sim_ManPatFree(p, pat)
		if end == p.Words {
			break
		}
		offset = end
	}
}

func reverseInts(values []int) {
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
}

// sourceRand reproduces the glibc rand()/srand() stream used by the C oracle.
// ABC's C macro consumes three rand calls per output word.
type sourceRand struct {
	state [31]uint32
	index uint64
}

func newSourceRand(seed uint32) *sourceRand {
	r := &sourceRand{}
	if seed == 0 {
		seed = 1
	}
	word := int64(int32(seed))
	if word < 0 {
		word = -word
	}
	r.state[0] = uint32(word)
	for i := 1; i < 31; i++ {
		hi, lo := word/127773, word%127773
		word = 16807*lo - 2836*hi
		if word < 0 {
			word += 2147483647
		}
		r.state[i] = uint32(word)
	}
	for i := 0; i < 310; i++ {
		r.Uint31()
	}
	return r
}

func (r *sourceRand) Uint31() uint32 {
	i := r.index % 31
	front := (i + 3) % 31
	value := r.state[front] + r.state[i]
	r.state[front] = value
	r.index++
	return value >> 1
}

func sourceAssignRandom(p *Sim_Man_t, rng *sourceRand) {
	for _, e := range p.Inputs {
		row := p.Sim0[e.NodeID()]
		for i := range row {
			row[i] = (rng.Uint31() << 24) ^ (rng.Uint31() << 12) ^ rng.Uint31()
		}
	}
}
