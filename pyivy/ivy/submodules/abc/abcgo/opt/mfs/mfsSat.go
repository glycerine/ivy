package mfs

import "github.com/glycerine/abcgo/sat/satoko"

// Abc_NtkMfsSolveSat enumerates fanin projections under observability and care.
// A timeout leaves a partial set but returns false, preventing its use in a rewrite.
func Abc_NtkMfsSolveSat(p *Mfs_Man_t, target int) bool {
	clauses, vars := encodeGraph(p.Window)
	s := newSolver(clauses, vars)
	addWindowConstraints(p, s, 0)
	fs := p.Network.Nodes[target].Fanins
	p.Care = make([]uint32, truthWords(len(fs)))
	projection := make([]int, len(fs))
	for i := range fs {
		projection[i] = graphLiteral(p.Window, 1+p.CareOutputs+i)
	}
	for {
		remaining := p.Params.BTLimit
		if remaining > 0 {
			remaining -= s.Stats().AllConflicts
			if remaining <= 0 {
				p.Stats.Timeouts++
				return false
			}
		}
		p.Stats.SATCalls++
		status := s.SolveAssumptionsLimit(nil, remaining)
		if status == satoko.SATOKO_UNDEC {
			p.Stats.Timeouts++
			return false
		}
		if status == satoko.SATOKO_UNSAT {
			break
		}
		mint := 0
		block := make([]int, len(projection))
		for i, l := range projection {
			value := s.ModelValue(l>>1) == 1
			if l&1 != 0 {
				value = !value
			}
			block[i] = l
			if value {
				mint |= 1 << uint(i)
				block[i] ^= 1
			}
		}
		setBit(p.Care, mint)
		if !s.AddClause(block) {
			break
		}
	}
	p.Care = normalizeTruth(p.Care, len(fs))
	return true
}
func Abc_NtkAddOneHotness(p *Mfs_Man_t) bool {
	inputs := p.Window.Inputs()
	for i, a := range inputs {
		for _, b := range inputs[i+1:] {
			if !p.Solver.AddClause([]int{int(a) ^ 1, int(b) ^ 1}) {
				return false
			}
		}
	}
	return true
}
