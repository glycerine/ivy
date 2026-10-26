package abc

// Go-only truth tests for the ordered gate clauses in satSolver3.h.

import (
	"errors"
	"testing"
)

func satSolver3ClausesSatisfied(clauses [][]SatLit, assignment int) bool {
	for _, clause := range clauses {
		satisfied := false
		for _, literal := range clause {
			value := assignment&(1<<literal.Var()) != 0
			if value != literal.Negative() {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return false
		}
	}
	return true
}

func TestSatSolver3GateClausesExhaustive(t *testing.T) {
	for complement := 0; complement < 2; complement++ {
		comp := complement != 0
		constant, err := SatSolver3ConstantClauses(0, comp)
		if err != nil || len(constant) != 1 {
			t.Fatal(err, constant)
		}
		for assignment := 0; assignment < 2; assignment++ {
			want := (assignment != 0) != comp
			if satSolver3ClausesSatisfied(constant, assignment) != want {
				t.Fatal("constant", comp, assignment)
			}
		}

		buffer, _ := SatSolver3BufferClauses(0, 1, comp)
		enabled, _ := SatSolver3EnabledBufferClauses(0, 1, 2, comp)
		for assignment := 0; assignment < 8; assignment++ {
			a := assignment&1 != 0
			b := assignment&2 != 0
			enable := assignment&4 != 0
			want := a == (b != comp)
			if satSolver3ClausesSatisfied(buffer, assignment) != want {
				t.Fatal("buffer", comp, assignment)
			}
			if satSolver3ClausesSatisfied(enabled, assignment) != (!enable || want) {
				t.Fatal("enabled buffer", comp, assignment)
			}
		}
	}

	for flags := 0; flags < 8; flags++ {
		complement0 := flags&1 != 0
		complement1 := flags&2 != 0
		complementOutput := flags&4 != 0
		clauses, err := SatSolver3AndClauses(0, 1, 2, complement0, complement1, complementOutput)
		if err != nil || len(clauses) != 3 {
			t.Fatal(err, clauses)
		}
		for assignment := 0; assignment < 8; assignment++ {
			output := assignment&1 != 0
			input0 := assignment&2 != 0
			input1 := assignment&4 != 0
			want := output == (((input0 != complement0) && (input1 != complement1)) != complementOutput)
			if satSolver3ClausesSatisfied(clauses, assignment) != want {
				t.Fatal("and", flags, assignment)
			}
		}
	}

	for complement := 0; complement < 2; complement++ {
		clauses, _ := SatSolver3XorClauses(0, 1, 2, complement != 0)
		for assignment := 0; assignment < 8; assignment++ {
			output := assignment&1 != 0
			input0 := assignment&2 != 0
			input1 := assignment&4 != 0
			want := output == ((input0 != input1) != (complement != 0))
			if satSolver3ClausesSatisfied(clauses, assignment) != want {
				t.Fatal("xor", complement, assignment)
			}
		}
	}

	for flags := 0; flags < 8; flags++ {
		cc, ct, ce := flags&1 != 0, flags&2 != 0, flags&4 != 0
		clauses, _ := SatSolver3MuxClauses(0, 1, 2, 3, cc, ct, ce, false)
		if len(clauses) != 6 {
			t.Fatal(len(clauses))
		}
		for assignment := 0; assignment < 16; assignment++ {
			output := assignment&1 != 0
			control := assignment&2 != 0
			thenValue := assignment&4 != 0
			elseValue := assignment&8 != 0
			selected := elseValue != ce
			if control != cc {
				selected = thenValue != ct
			}
			if satSolver3ClausesSatisfied(clauses, assignment) != (output == selected) {
				t.Fatal("mux", flags, assignment)
			}
		}
	}
	equalMux, _ := SatSolver3MuxClauses(0, 1, 2, 2, false, false, false, false)
	if len(equalMux) != 4 {
		t.Fatal(len(equalMux))
	}
	complementedMux, _ := SatSolver3MuxClauses(0, 1, 2, 3, false, false, false, true)
	if complementedMux[0][2].Negative() {
		t.Fatal("source first mux clause order changed", complementedMux[0])
	}

	mux4, _ := SatSolver3Mux4Clauses(0, 1, 2, 3, 4, 5, 6)
	for assignment := 0; assignment < 1<<7; assignment++ {
		output := assignment&1 != 0
		selector := (assignment >> 1) & 3
		selected := assignment&(1<<(3+selector)) != 0
		if satSolver3ClausesSatisfied(mux4, assignment) != (output == selected) {
			t.Fatal("mux4", assignment)
		}
	}

	xorAnd, _ := SatSolver3XorAndClauses(0, 1, 2, 3)
	for assignment := 0; assignment < 16; assignment++ {
		output := assignment&1 != 0
		a := assignment&2 != 0
		b := assignment&4 != 0
		c := assignment&8 != 0
		if satSolver3ClausesSatisfied(xorAnd, assignment) != (output == ((a != b) && c)) {
			t.Fatal("xor-and", assignment)
		}
	}

	for complement := 0; complement < 2; complement++ {
		constraint, _ := SatSolver3ConstraintClauses(0, 1, complement != 0)
		for assignment := 0; assignment < 4; assignment++ {
			value := assignment&1 != 0
			if satSolver3ClausesSatisfied(constraint, assignment) != (value != (complement != 0)) {
				t.Fatal("constraint", complement, assignment)
			}
		}
	}

	half, _ := SatSolver3HalfSorterClauses(0, 1, 2, 3)
	for assignment := 0; assignment < 16; assignment++ {
		a := assignment&1 != 0
		b := assignment&2 != 0
		input0 := assignment&4 != 0
		input1 := assignment&8 != 0
		want := (!input0 || a) && (!input1 || a) && (!input0 || !input1 || b)
		if satSolver3ClausesSatisfied(half, assignment) != want {
			t.Fatal("half sorter", assignment)
		}
	}
}

func TestSatSolver3GateOwnershipValidationAndAtomicLoad(t *testing.T) {
	clauses, err := SatSolver3BufferClauses(0, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	clone := CloneSatSolver3Clauses(clauses)
	clone[0][0] = clone[0][0].Not()
	if clauses[0][0] == clone[0][0] {
		t.Fatal("clone aliases source")
	}
	if _, err = SatSolver3BufferClauses(-1, 0, false); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	solver, _ := NewSatSolver3(SatSolver3Options{MaxVariables: 2})
	_ = solver.SetVariableCount(2)
	positive := satSolver3TestLiteral(t, 0, false)
	before := solver.OriginalClauseCount()
	if ok, loadErr := solver.AddClauses([][]SatLit{{positive}, {positive.Not()}}); ok || loadErr != nil {
		t.Fatal(ok, loadErr)
	}
	if solver.OriginalClauseCount() != before {
		t.Fatal("failed load mutated solver", solver.OriginalClauseCount(), before)
	}
	if ok, loadErr := solver.AddClauses(clauses); !ok || loadErr != nil || solver.OriginalClauseCount() != before+2 {
		t.Fatal(ok, loadErr, solver.OriginalClauseCount())
	}
}
