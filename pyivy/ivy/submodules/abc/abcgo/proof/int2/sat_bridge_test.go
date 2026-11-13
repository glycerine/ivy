package int2

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSATBridgeCumulativeGlobalLimits(t *testing.T) {
	a, b := sat_solver_new(), sat_solver_new()
	for _, s := range []*sat_solver{a, b} {
		sat_solver_setnvars(s, 3)
		sat_solver_addclause(s, []int32{1, 2})
		sat_solver_addclause(s, []int32{3, 4})
	}
	for _, assumption := range []int32{0, 1, 0} {
		global := int64(1)
		got := sat_solver_solve(a, []int32{assumption}, 0, 0, 0, global)
		want := b.SolveWithResourceLimits([]bsat.Literal{bsat.Literal(assumption)}, 0, 0, 0, global)
		if got != int32(want) || a.PropagationCount() != b.PropagationCount() {
			t.Fatalf("assumption %d: bridge %d/%d, direct %d/%d", assumption, got, a.PropagationCount(), want, b.PropagationCount())
		}
	}
}

func TestSATBridgeAndGateAllInputPolarities(t *testing.T) {
	for complement0 := int32(0); complement0 <= 1; complement0++ {
		for complement1 := int32(0); complement1 <= 1; complement1++ {
			for complementOut := int32(0); complementOut <= 1; complementOut++ {
				solver := sat_solver_new()
				sat_solver_setnvars(solver, 3)
				if sat_solver_add_and(solver, 2, 0, 1, complement0, complement1, complementOut) != 3 {
					t.Fatal("AND gate clauses were rejected")
				}
				for x := int32(0); x <= 1; x++ {
					for y := int32(0); y <= 1; y++ {
						for z := int32(0); z <= 1; z++ {
							want := (x^complement0)&(y^complement1) ^ complementOut
							assumptions := []int32{sat_literal(0, 1^x), sat_literal(1, 1^y), sat_literal(2, 1^z)}
							status := sat_solver_solve(solver, assumptions, 0, 0, 0, 0)
							if (status == l_True) != (z == want) {
								t.Fatalf("phase %d/%d/%d inputs %d,%d,%d: status %d", complement0, complement1, complementOut, x, y, z, status)
							}
						}
					}
				}
			}
		}
	}
}

func TestSATBridgeBufferComplement(t *testing.T) {
	for complement := int32(0); complement <= 1; complement++ {
		solver := sat_solver_new()
		sat_solver_setnvars(solver, 2)
		if got := sat_solver_add_buffer(solver, 1, 0, complement); got != 2 {
			t.Fatalf("buffer clauses = %d", got)
		}
		for input := int32(0); input <= 1; input++ {
			for output := int32(0); output <= 1; output++ {
				status := sat_solver_solve(solver, []int32{sat_literal(0, 1^input), sat_literal(1, 1^output)}, 0, 0, 0, 0)
				if (status == l_True) != (output == (input ^ complement)) {
					t.Fatalf("complement %d input %d output %d status %d", complement, input, output, status)
				}
			}
		}
	}
}

func TestSATBridgeLoadsClausesAndDetectsContradiction(t *testing.T) {
	// x0 => x1 and x1 => x0, with x0 required true.
	solver := sat_solver_from_clauses(2, [][]int32{{1, 2}, {3, 0}, {0}})
	if solver == nil {
		t.Fatal("satisfiable CNF rejected")
	}
	if got := sat_solver_solve(solver, []int32{sat_literal(1, 0)}, 0, 0, 0, 0); got != l_True {
		t.Fatalf("x1=true status = %d, want satisfiable", got)
	}
	if got := sat_solver_solve(solver, []int32{sat_literal(1, 1)}, 0, 0, 0, 0); got != l_False {
		t.Fatalf("x1=false status = %d, want unsatisfiable", got)
	}
	if got := sat_solver_from_clauses(1, [][]int32{{0}, {1}}); got != nil {
		t.Fatal("contradictory unit clauses should reject solver construction")
	}
}

func TestSATBridgeClauseExpandsVariableRange(t *testing.T) {
	solver := sat_solver_new()
	if sat_solver_addclause(solver, []int32{sat_literal(4, 0)}) != 1 {
		t.Fatal("unit clause for a newly introduced variable was rejected")
	}
	if got := sat_solver_nvars(solver); got != 5 {
		t.Fatalf("variable count = %d, want 5", got)
	}
	if got := sat_solver_solve(solver, []int32{sat_literal(4, 1)}, 0, 0, 0, 0); got != l_False {
		t.Fatalf("contradicting assignment status = %d, want unsatisfiable", got)
	}
}
