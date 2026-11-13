package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
)

// cec2CNF builds clauses on demand using the frontier traversal of
// Gia_ObjGetCnfVar. SAT variable numbers are private to this solver instance.
type cec2CNF struct {
	graph   *gia.Gia_Man_t
	solver  *satoko.Solver
	ids     map[int32]int
	objects []int32
}

func newCec2CNF(graph *gia.Gia_Man_t, solver *satoko.Solver) *cec2CNF {
	return &cec2CNF{graph: graph, solver: solver, ids: make(map[int32]int)}
}

func (c *cec2CNF) variable(root int32) int {
	var frontier []int32
	add := func(id int32) int {
		if v, ok := c.ids[id]; ok {
			return v
		}
		v := c.solver.AddVariable(false)
		c.ids[id] = v
		c.objects = append(c.objects, id)
		if id == 0 {
			if !c.solver.AddClause([]int{2*v + 1}) {
				panic("cec: inconsistent constant")
			}
		} else if gia.Gia_ObjIsAnd(gia.Gia_ManObj(c.graph, id)) != 0 {
			frontier = append(frontier, id)
		}
		return v
	}
	result := add(root)
	for i := 0; i < len(frontier); i++ {
		id := frontier[i]
		obj := gia.Gia_ManObj(c.graph, id)
		out := 2 * c.ids[id]
		if gia.Gia_ObjMark0(obj) != 0 {
			var then, otherwise int32
			control := gia.Gia_ObjRecognizeMuxLits(c.graph, obj, &then, &otherwise)
			literals := []int32{control, then, otherwise}
			sat := [3]int{}
			for k, lit := range literals {
				sat[k] = 2*add(lit>>1) + int(lit&1)
			}
			if !cec2AddMuxClauses(c.solver, out, sat[0], sat[1], sat[2]) {
				panic("cec: inconsistent mux CNF")
			}
		} else {
			fanins := cec2CollectSuper(c.graph, id<<1, true)
			lits := make([]int, len(fanins))
			for k, lit := range fanins {
				lits[k] = 2*add(lit>>1) + int(lit&1)
			}
			if !cec2AddSuperClauses(c.solver, out, lits) {
				panic("cec: inconsistent supergate CNF")
			}
		}
	}
	return result
}

// solveTwo follows Cec2_ManSolveTwo: find a witness to the predicted phase,
// first with object 0 low, then (except for constant 0) with object 0 high.
// SAT leaves its witness available through the solver's model.
func (c *cec2CNF) solveTwo(first, second int32, phase int, limit int64) int {
	if first > second {
		first, second = second, first
	}
	if first == second || phase < 0 || phase > 1 {
		panic("cec: invalid SAT comparison")
	}
	a, b := c.variable(first), c.variable(second)
	result := c.solver.SolveAssumptionsLimit([]int{2*a + 1, 2*b + phase}, limit)
	if result == satoko.SATOKO_UNSAT && first != 0 {
		result = c.solver.SolveAssumptionsLimit([]int{2 * a, 2*b + (phase ^ 1)}, limit)
	}
	return result
}
