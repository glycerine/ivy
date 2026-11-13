package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/glucose2"
)

// cec4CNF owns the SAT IDs and the incremental Glucose2 instance used for
// Cec4's simple on-demand CNF. IDs and models never use the GIA scratch fields.
type cec4CNF struct {
	graph         *gia.Gia_Man_t
	solver        *glucose2.SimpSolver
	jType         int32
	ids           map[int32]int32
	objects       []int32 // SAT variable to object ID, or -1 for an auxiliary
	inputs        map[int32]int32
	needsFullCone bool
	gates         [2]int32
}

func newCec4CNF(graph *gia.Gia_Man_t, jType int32) *cec4CNF {
	if graph == nil || jType < 0 || jType > 2 {
		panic("cec4: invalid SAT manager")
	}
	s := glucose2.NewABCSolver()
	s.JustificationMode = int(jType)
	return &cec4CNF{graph: graph, solver: s, jType: jType, ids: make(map[int32]int32), inputs: make(map[int32]int32)}
}

func (c *cec4CNF) newVar(id int32) int32 {
	v := int32(c.solver.NewVar(true, true))
	c.objects = append(c.objects, id)
	if id >= 0 {
		c.ids[id] = v
	}
	return v
}

// gate installs the same clause/circuit split as Cec4_ObjGetCnfVar: mode 0
// uses clauses, mode 1 uses both, and mode 2 uses circuit reasoning.
func (c *cec4CNF) gate(out, a, b int32, xor bool) {
	av, bv := a>>1, b>>1
	// Glucose's circuit representation requires distinct fanins.
	useClauses := c.jType < 2 || av == bv
	if av == bv {
		c.needsFullCone = true
	}
	if useClauses {
		var ok bool
		if xor {
			ok = c.solver.AddABCXor(out, av, bv, (a^b)&1 != 0)
		} else {
			ok = c.solver.AddABCAnd(out, av, bv, a&1 != 0, b&1 != 0, false)
		}
		if !ok {
			panic("cec4: inconsistent gate encoding")
		}
	}
	if c.jType > 0 && av != bv {
		if xor {
			c.gates[1]++
		} else {
			c.gates[0]++
		}
		if (!xor && a > b) || (xor && a < b) {
			a, b = b, a
		}
		c.solver.SetVarFaninLits(glucose2.Var(out), glucose2.Lit(a), glucose2.Lit(b))
	}
}

func (c *cec4CNF) variable(id int32) int32 {
	if v, ok := c.ids[id]; ok {
		return v
	}
	if id < 0 || id >= gia.Gia_ManObjNum(c.graph) {
		panic("cec4: invalid graph object")
	}
	a, b, third, kind := c.graph.Gate(id)
	if id == 0 {
		v := c.newVar(id)
		if !c.solver.AddABCClause([]int32{v<<1 | 1}) {
			panic("cec4: inconsistent constant")
		}
		return v
	}
	if kind == 1 {
		v := c.newVar(id)
		c.inputs[id] = v
		return v
	}
	if kind == 5 || kind == 6 {
		c.needsFullCone = true
		v := c.variable(a >> 1)
		out := c.newVar(id)
		lit := v<<1 | a&1
		if !c.solver.AddABCClause([]int32{out<<1 | 1, lit}) || !c.solver.AddABCClause([]int32{out << 1, lit ^ 1}) {
			panic("cec4: inconsistent buffer")
		}
		return out
	}
	if kind != 2 && kind != 3 && kind != 4 {
		panic("cec4: unsupported graph gate")
	}
	la := c.variable(a>>1)<<1 | a&1
	lb := c.variable(b>>1)<<1 | b&1
	if kind == 4 {
		c.needsFullCone = true
		lc := c.variable(third>>1)<<1 | third&1
		x := c.newVar(-1)
		c.gate(x, la, lb, false)
		y := c.newVar(-1)
		c.gate(y, la^1, lc, false)
		out := c.newVar(-1)
		c.gate(out, x<<1|1, y<<1|1, false)
		// The final AND computes the complement of the mux.
		v := c.newVar(id)
		if !c.solver.AddABCClause([]int32{v<<1 | 1, out<<1 | 1}) || !c.solver.AddABCClause([]int32{v << 1, out << 1}) {
			panic("cec4: inconsistent mux output")
		}
		return v
	}
	out := c.newVar(id)
	c.gate(out, la, lb, kind == 3)
	return out
}

// solveTwo returns Glucose2's SAT/UNSAT/UNDEC status; a SAT result leaves
// the witness in solver.Model. phase is the expected relation's complement.
func (c *cec4CNF) solveTwo(first, second, phase, limit int32) int32 {
	if first > second {
		first, second = second, first
	}
	if first < 0 || first == second || phase < 0 || phase > 1 {
		panic("cec4: invalid SAT comparison")
	}
	a, b := c.variable(first), c.variable(second)
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
	solve := func(x, y int32) int32 {
		if limit > 0 {
			c.solver.SetConfBudget(int64(limit))
		} else {
			c.solver.BudgetOff()
		}
		return c.solver.SolveABCLiterals([]int32{x, y})
	}
	status := solve(a<<1|1, b<<1|phase)
	if status == glucose2.GLUCOSE_UNSAT && first != 0 {
		status = solve(a<<1, b<<1|(phase^1))
	}
	return status
}

func (c *cec4CNF) recycle() {
	c.solver = glucose2.NewABCSolver()
	c.solver.JustificationMode = int(c.jType)
	clear(c.ids)
	clear(c.inputs)
	c.objects = c.objects[:0]
	c.needsFullCone = false
	c.gates = [2]int32{}
}
