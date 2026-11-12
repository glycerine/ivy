package ssc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
)

func Ssc_ManStartSolver(p *Ssc_Man_t) {
	p.solver = satoko.New()
	p.solver.AddVariable(false)
	p.solver.AddClause([]int{1})
	p.inputs = make([]int, len(p.Aig.InputLiterals()))
	for i := range p.inputs {
		p.inputs[i] = p.solver.AddVariable(false) << 1
	}
	for side, g := range []*gia.Gia_Man_t{p.Aig, p.Care} {
		p.caches[side] = make([]int, g.ObjectCount())
		for i := range p.caches[side] {
			p.caches[side][i] = -1
		}
		p.caches[side][0] = 0
		for i, lit := range g.InputLiterals() {
			p.caches[side][lit>>1] = p.inputs[i]
		}
	}
	p.inconsistent = false
	p.callsSince = 0
	for _, lit := range p.Care.OutputLiterals() {
		value := p.encode(1, lit)
		if !p.solver.AddClause([]int{value ^ 1}) {
			p.inconsistent = true
		}
	}
}

// encode emits exact local gate clauses on demand. Output/buffer nodes alias
// their input; XOR and MUX are encoded directly rather than treated as ANDs.
func (p *Ssc_Man_t) encode(side int, lit int32) int {
	id := lit >> 1
	if p.caches[side][id] >= 0 {
		return p.caches[side][id] ^ int(lit&1)
	}
	g := p.Aig
	if side == 1 {
		g = p.Care
	}
	a, b, c, kind := g.Gate(id)
	x := p.encode(side, a)
	if kind == 5 || kind == 6 {
		p.caches[side][id] = x
		return x ^ int(lit&1)
	}
	if kind < 2 || kind > 4 {
		panic("ssc: invalid CNF node")
	}
	fanins := []int{x, p.encode(side, b)}
	if kind == 4 {
		fanins = append(fanins, p.encode(side, c))
	}
	y := p.solver.AddVariable(false) << 1
	p.caches[side][id] = y
	for assignment := 0; assignment < 1<<len(fanins); assignment++ {
		v := func(i int) bool { return assignment&(1<<i) != 0 }
		result := v(0) && v(1)
		if kind == 3 {
			result = v(0) != v(1)
		} else if kind == 4 {
			result = v(0) && v(1) || !v(0) && v(2)
		}
		clause := make([]int, len(fanins)+1)
		for i, f := range fanins {
			clause[i] = f
			if v(i) {
				clause[i] ^= 1
			}
		}
		clause[len(fanins)] = y
		if !result {
			clause[len(fanins)] ^= 1
		}
		if !p.solver.AddClause(clause) {
			p.inconsistent = true
		}
	}
	return y ^ int(lit&1)
}
func Ssc_ManFindPivotSat(p *Ssc_Man_t) ([]bool, int) {
	if p.inconsistent {
		return nil, satoko.SATOKO_UNSAT
	}
	status := p.solver.SolveAssumptionsLimit(nil, int64(p.Params.BTLimit))
	if status != satoko.SATOKO_SAT {
		return nil, status
	}
	return p.pattern(), status
}
func (p *Ssc_Man_t) pattern() []bool {
	v := make([]bool, len(p.inputs))
	for i, lit := range p.inputs {
		v[i] = p.solver.ModelValue(lit>>1) == 1
	}
	return v
}
func Ssc_ManCollectSatPattern(p *Ssc_Man_t) []bool { return p.pattern() }

// Return codes retain SAT's convention: UNSAT proves equivalence; SAT provides
// a care-valid distinguishing pattern; UNKNOWN never permits a merge.
func Ssc_ManCheckEquivalence(p *Ssc_Man_t, repr, node int32, complement int32) int {
	if p.Params.SatVarMax > 0 && p.solver.VariableCount() > int(p.Params.SatVarMax) && p.callsSince > int(p.Params.CallsRecycle) {
		Ssc_ManStartSolver(p)
		p.Stats.Recycles++
	}
	p.callsSince++
	p.Stats.Calls++
	a, b := p.encode(0, repr<<1), p.encode(0, node<<1)^int(complement&1)
	status := satoko.SATOKO_UNSAT
	for _, assume := range [][]int{{a, b ^ 1}, {a ^ 1, b}} {
		result := p.solver.SolveAssumptionsLimit(assume, int64(p.Params.BTLimit))
		if result == satoko.SATOKO_SAT {
			p.Pattern = p.pattern()
			p.Stats.Counterexamples++
			return result
		}
		if result != satoko.SATOKO_UNSAT {
			status = result
		}
	}
	if status == satoko.SATOKO_UNSAT {
		p.Stats.Proofs++
	} else {
		p.Stats.Undecided++
	}
	return status
}
func Ssc_ManSatSolverRecycle(p *Ssc_Man_t) { Ssc_ManStartSolver(p); p.Stats.Recycles++ }
