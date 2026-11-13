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
	status, _ := c.solveTwoWithCone(first, second, phase, limit, false)
	return status
}

// cone collects the SAT variables and CI indices reachable from both roots,
// matching Cec2_ManCollect_rec without storing traversal marks on the graph.
func (c *cec2CNF) cone(first, second int32) (variables []int, inputs map[int32]int) {
	seen := make(map[int32]bool)
	inputs = make(map[int32]int)
	var visit func(int32)
	visit = func(id int32) {
		if seen[id] {
			return
		}
		seen[id] = true
		v, ok := c.ids[id]
		if ok {
			variables = append(variables, v)
		}
		if id == 0 {
			return
		}
		obj := gia.Gia_ManObj(c.graph, id)
		if gia.Gia_ObjIsAnd(obj) != 0 {
			visit(gia.Gia_ObjFaninLit0(obj, id) >> 1)
			visit(gia.Gia_ObjFaninLit1(obj, id) >> 1)
		} else if gia.Gia_ObjIsCi(obj) != 0 {
			if !ok {
				panic("cec: cone input has no SAT variable")
			}
			inputs[gia.Gia_ObjCioId(obj)] = v
		} else {
			panic("cec: unsupported object in SAT cone")
		}
	}
	visit(first)
	visit(second)
	return
}

// solveTwoWithCone preserves the source's optional decision-cone restriction.
// The returned input map identifies the SAT model values needed for a pattern.
func (c *cec2CNF) solveTwoWithCone(first, second int32, phase int, limit int64, useCone bool) (int, map[int32]int) {
	if first > second {
		first, second = second, first
	}
	if first == second || phase < 0 || phase > 1 {
		panic("cec: invalid SAT comparison")
	}
	a, b := c.variable(first), c.variable(second)
	variables, inputs := c.cone(first, second)
	if useCone {
		c.solver.MarkCone(variables, true)
		defer c.solver.MarkCone(variables, false)
	}
	status := c.solver.SolveAssumptionsLimit([]int{2*a + 1, 2*b + phase}, limit)
	if status == satoko.SATOKO_UNSAT && first != 0 {
		status = c.solver.SolveAssumptionsLimit([]int{2 * a, 2*b + (phase ^ 1)}, limit)
	}
	return status, inputs
}

// rollback discards a non-cone query's clauses and their SAT ID mappings.
// The caller must bookmark an empty solver before constructing any query.
func (c *cec2CNF) rollback() {
	c.solver.Rollback()
	clear(c.ids)
	c.objects = c.objects[:0]
}
