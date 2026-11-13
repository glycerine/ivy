package saig

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

type unfold2Classes struct{ typeI, typeII []aig.Edge }

// filterUnfold2Candidates classifies the three induction obligations with an
// incremental SAT solver. Unknown results do not qualify as a proof.
func filterUnfold2Candidates(g *aig.Graph, candidates *vec.Vec_Vec_t, conflicts, propagations, verbose int32) (unfold2Classes, error) {
	var classes unfold2Classes
	if g == nil || candidates == nil || conflicts < 0 || propagations < 0 {
		return classes, fmt.Errorf("saig: invalid induction filter arguments")
	}
	frames := Saig_ManCreateIndMiter2(g, candidates)
	// Keep each induction output as a query variable; the source passes
	// the full CO count so the CNF does not assert all outputs at once.
	formula := cnf.Cnf_DeriveSimple(frames, aig.Aig_ManCoNum(frames))
	if formula == nil {
		return classes, fmt.Errorf("saig: cannot encode induction miter")
	}
	solver := cnf.Cnf_DataWriteIntoSolver(formula, 1, 0)
	if solver == nil {
		return classes, fmt.Errorf("saig: inconsistent induction miter encoding")
	}
	solver.SetLimits(int64(conflicts), int64(propagations))
	outputs := frames.Outputs()
	counter := 0
	solve := func(index int) bsat.Status {
		id := outputs[index].NodeID()
		if id >= len(formula.VarNums) || formula.VarNums[id] < 0 {
			panic("saig: missing induction output variable")
		}
		return solver.Solve(bsat.Positive(formula.VarNums[id]))
	}
	for level := 0; level < vec.Vec_VecSize(candidates); level++ {
		nodes := vec.Vec_VecEntry(candidates, level)
		kept := 0
		for _, item := range vec.Vec_PtrArray(nodes) {
			e := item.(aig.Edge)
			status := solve(counter * 3)
			counter++
			kind := 0
			if status == bsat.Unsatisfiable {
				kind = 1
				classes.typeI = append(classes.typeI, e)
			} else if solve(counter*3-2) == bsat.Satisfiable && solve(counter*3-1) == bsat.Unsatisfiable {
				kind = 2
				classes.typeII = append(classes.typeII, e)
			}
			if kind != 0 {
				vec.Vec_PtrWriteEntry(nodes, kept, e)
				kept++
				if verbose != 0 {
					fmt.Printf("type %d: %d\n", kind, e.NodeID())
				}
			}
		}
		vec.Vec_PtrShrink(nodes, kept)
	}
	return classes, nil
}
