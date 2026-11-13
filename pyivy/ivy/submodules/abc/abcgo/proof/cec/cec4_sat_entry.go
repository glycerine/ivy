package cec

import (
	"time"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose2"
)

func cec4EnsureCNF(p *Cec4_Man_t) *cec4CNF {
	if p == nil || p.pNew == nil || p.pPars == nil {
		panic("cec4: missing SAT manager")
	}
	if p.cnf == nil {
		p.cnf = newCec4CNF(p.pNew, p.pPars.jType)
		p.pSat = p.cnf.solver
	} else if p.cnf.graph != p.pNew {
		panic("cec4: SAT manager graph changed")
	}
	return p.cnf
}

// Cec4_ObjGetCnfVar adds the object's transitive fanin cone on demand.
func Cec4_ObjGetCnfVar(p *Cec4_Man_t, object int32) int32 {
	c := cec4EnsureCNF(p)
	old := c.gates
	v := c.variable(object)
	for i := range old {
		p.nGates[i] += c.gates[i] - old[i]
	}
	return v
}

// Cec4_ManSatSolverRecycle discards learned clauses and object-to-variable IDs.
func Cec4_ManSatSolverRecycle(p *Cec4_Man_t) {
	c := cec4EnsureCNF(p)
	c.recycle()
	p.pSat = c.solver
	p.nCallsSince = 0
	p.nRecycles++
}

// Cec4_ManSolveTwo compares two regular GIA objects under the given expected
// phase. It leaves a counterexample model in p.pSat when SAT is returned.
func Cec4_ManSolveTwo(p *Cec4_Man_t, first, second, phase int32, easy *int32, verbose int32) int32 {
	if easy == nil || first == second || phase < 0 || phase > 1 {
		panic("cec4: invalid SAT comparison")
	}
	if first > second {
		first, second = second, first
	}
	if first < 0 {
		panic("cec4: invalid SAT object")
	}
	*easy = 0
	c := cec4EnsureCNF(p)
	p.nCallsSince++
	if p.nCallsSince > p.pPars.nCallsRecycle && p.pPars.nSatVarMax != 0 && int32(len(c.ids)) > p.pPars.nSatVarMax {
		Cec4_ManSatSolverRecycle(p)
		c = p.cnf
	}
	limit := p.pPars.nBTLimit
	if p.vFails != nil && (vec.Vec_BitGetEntry(p.vFails, int(first)) != 0 || vec.Vec_BitGetEntry(p.vFails, int(second)) != 0) {
		limit /= 10
		if limit < 1 {
			limit = 1
		}
	}
	start := time.Now()
	a := Cec4_ObjGetCnfVar(p, first)
	b := Cec4_ObjGetCnfVar(p, second)
	if c.jType > 0 {
		c.solver.SetNewRound()
		c.solver.MarkCone(glucose2.Var(a))
		c.solver.MarkCone(glucose2.Var(b))
		if c.needsFullCone {
			for v := 0; v < c.solver.NVars(); v++ {
				c.solver.MarkCone(glucose2.Var(v))
			}
		}
	}
	p.timeCnf += time.Since(start).Nanoseconds()
	solve := func(x, y int32) (int32, int32) {
		if limit > 0 {
			c.solver.SetConfBudget(int64(limit))
		} else {
			c.solver.BudgetOff()
		}
		before := c.solver.Conflicts
		status := c.solver.SolveABCLiterals([]int32{x, y})
		return status, int32(c.solver.Conflicts - before)
	}
	status, conflicts := solve(a<<1|1, b<<1|phase)
	firstConflicts := conflicts
	maxConflicts := conflicts
	if status == glucose2.GLUCOSE_UNSAT && first != 0 {
		status, conflicts = solve(a<<1, b<<1|(phase^1))
		maxConflicts = conflicts
		if verbose != 0 && status == glucose2.GLUCOSE_UNSAT {
			if firstConflicts > maxConflicts {
				maxConflicts = firstConflicts
			}
			conflicts += firstConflicts
		}
	}
	if verbose != 0 && (status == glucose2.GLUCOSE_SAT || status == glucose2.GLUCOSE_UNSAT) {
		index := 0
		if status == glucose2.GLUCOSE_UNSAT {
			index = 1
		}
		if conflicts == 0 {
			p.nConflicts[index][0]++
			*easy = 1
		}
		p.nConflicts[index][1] += conflicts
		if maxConflicts > p.nConflicts[index][2] {
			p.nConflicts[index][2] = maxConflicts
		}
	}
	return status
}
