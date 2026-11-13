package cnf

import (
	"testing"

	"github.com/glycerine/abcgo/sat/bsat"
)

func TestSolverLoadingPropagatesRootUnits(t *testing.T) {
	// Cnf_DataWriteIntoSolverInt in cnfMan.c calls sat_solver_simplify
	// before returning. The last unit makes the two preceding clauses
	// demand opposite values of the second variable.
	data := &Cnf_Dat_t{Variables: 2, Clauses: [][]int32{{1, 2}, {1, 3}, {0}}}
	if solver := Cnf_DataWriteIntoSolver(data, 1, 0); solver != nil {
		t.Fatal("root propagation should reject the inconsistent formula")
	}
}

func TestSolverLoadingSourceRootCases(t *testing.T) {
	for _, tc := range []struct {
		name    string
		clauses [][]int32
	}{
		{"empty clause", [][]int32{{}}},
		{"contradictory units", [][]int32{{0}, {1}}},
		{"queued contradictory implications", [][]int32{{1, 2}, {1, 3}, {0}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := &Cnf_Dat_t{Variables: 2, Clauses: tc.clauses}
			if solver := Cnf_DataWriteIntoSolver(data, 1, 0); solver != nil {
				t.Fatal("C loader returns NULL for a root contradiction")
			}
		})
	}
}

func TestSolverLoadingRootImplicationsAndRepeatedSimplify(t *testing.T) {
	// x, !x or y, !y or z: root propagation sets all three variables.
	data := &Cnf_Dat_t{Variables: 3, Clauses: [][]int32{{1, 2}, {3, 4}, {0}}}
	solver := Cnf_DataWriteIntoSolver(data, 1, 0)
	if solver == nil || !solver.Simplify() || !solver.Simplify() {
		t.Fatal("satisfiable root implications must survive repeated simplification")
	}
	if got := solver.Solve(bsat.Negative(2)); got != bsat.Unsatisfiable {
		t.Fatalf("negated implied variable = %v, want UNSAT", got)
	}
	if got := solver.Solve(); got != bsat.Satisfiable {
		t.Fatalf("unconstrained status = %v, want SAT", got)
	}
	for variable := int32(0); variable < 3; variable++ {
		if value, ok := solver.Value(variable); !ok || !value {
			t.Fatalf("model variable %d = %v (available=%v), want true", variable, value, ok)
		}
	}
}
