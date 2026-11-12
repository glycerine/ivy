package mfs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/satoko"
)

func encodeGraph(g *aig.Graph) ([][]int, int) {
	var c [][]int
	add := func(xs ...int) { c = append(c, xs) }
	add(0)
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		z := int(e)
		a, b := int(n.Fanins[0]), int(n.Fanins[1])
		switch n.Kind {
		case aig.And:
			add(a^1, b^1, z)
			add(a, z^1)
			add(b, z^1)
		case aig.ExclusiveOr:
			add(a, b, z^1)
			add(a^1, b^1, z^1)
			add(a, b^1, z)
			add(a^1, b, z)
		case aig.Buffer, aig.Output:
			add(a, z^1)
			add(a^1, z)
		}
	}
	return c, int(aig.Aig_ManObjNumMax(g))
}
func graphLiteral(g *aig.Graph, output int) int { return int(g.Node(g.Outputs()[output]).Fanins[0]) }
func Abc_MfsSatAddXor(s *satoko.Solver, a, b, z int) {
	s.AddClause([]int{a, b, z ^ 1})
	s.AddClause([]int{a ^ 1, b ^ 1, z ^ 1})
	s.AddClause([]int{a, b ^ 1, z})
	s.AddClause([]int{a ^ 1, b, z})
}
func newSolver(clauses [][]int, vars int) *satoko.Solver {
	s := satoko.New()
	s.SetVariables(vars)
	for _, c := range clauses {
		s.AddClause(c)
	}
	return s
}
func addWindowConstraints(p *Mfs_Man_t, s *satoko.Solver, offset int) {
	for o := 0; o <= p.CareOutputs; o++ {
		s.AddClause([]int{graphLiteral(p.Window, o) + 2*offset})
	}
	if p.Params.OneHotness {
		inputs := p.Window.Inputs()
		for i, a := range inputs {
			for _, b := range inputs[i+1:] {
				s.AddClause([]int{int(a) ^ 1 + 2*offset, int(b) ^ 1 + 2*offset})
			}
		}
	}
}

// Two independently constrained window copies have opposite target values.
// Each projection variable is the XOR of one divisor across the copies.
func Abc_MfsCreateSolverResub(p *Mfs_Man_t, candidates []int, invert bool) *satoko.Solver {
	p.clauses, p.vars = encodeGraph(p.Window)
	s := newSolver(p.clauses, 2*p.vars+len(p.Divisors))
	for _, c := range p.clauses {
		copy := make([]int, len(c))
		for i, l := range c {
			copy[i] = l + 2*p.vars
		}
		s.AddClause(copy)
	}
	addWindowConstraints(p, s, 0)
	addWindowConstraints(p, s, p.vars)
	f := graphLiteral(p.Window, 1+p.CareOutputs)
	if invert {
		f ^= 1
	}
	s.AddClause([]int{f})
	s.AddClause([]int{(f ^ 1) + 2*p.vars})
	p.projection = make([]int, len(p.Divisors))
	for i := range p.Divisors {
		a := graphLiteral(p.Window, 2+p.CareOutputs+i)
		z := (2*p.vars + i) << 1
		p.projection[i] = z
		Abc_MfsSatAddXor(s, a, a+2*p.vars, z)
	}
	for _, i := range candidates {
		s.AddClause([]int{p.projection[i] ^ 1})
	}
	p.Solver = s
	return s
}
func solve(p *Mfs_Man_t, s *satoko.Solver, assumptions []int) int {
	p.Stats.SATCalls++
	status := s.SolveAssumptionsLimit(assumptions, p.Params.BTLimit)
	if status == satoko.SATOKO_UNDEC {
		p.Stats.Timeouts++
	}
	return status
}

// Interpolation is exact functional separation by projected SAT enumeration.
// It returns a truth table, not a resolution proof trace. Unreachable projected
// minterms are zero; every on/off care minterm is separated before any update.
func Abc_NtkMfsInterplate(p *Mfs_Man_t, candidates []int) []uint32 {
	if len(candidates) > MFS_FANIN_MAX {
		panic("mfs: too many projected inputs")
	}
	s := newSolver(p.clauses, p.vars)
	addWindowConstraints(p, s, 0)
	s.AddClause([]int{graphLiteral(p.Window, 1+p.CareOutputs)})
	truth := make([]uint32, truthWords(len(candidates)))
	vars := make([]int, len(candidates))
	for i, c := range candidates {
		vars[i] = graphLiteral(p.Window, 2+p.CareOutputs+c)
	}
	for {
		status := solve(p, s, nil)
		if status == satoko.SATOKO_UNDEC {
			return nil
		}
		if status == satoko.SATOKO_UNSAT {
			break
		}
		mint := 0
		block := make([]int, len(vars))
		for i, l := range vars {
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
		setBit(truth, mint)
		if !s.AddClause(block) {
			break
		}
	}
	return normalizeTruth(truth, len(candidates))
}
func Abc_NtkMfsInterplateEval(p *Mfs_Man_t, candidates []int) int {
	truth := Abc_NtkMfsInterplate(p, candidates)
	if truth == nil {
		return -1
	}
	return gateCount(truthGraph(truth, len(candidates)))
}
