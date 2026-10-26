package abc

// Stand-alone tests for satSolver2 gate-clause construction.

import (
	"errors"
	"reflect"
	"testing"
)

func satSolver2GateSatisfied(clauses []SatSolver2ClauseSpec, assignment int) bool {
	for _, clause := range clauses {
		satisfied := false
		for _, literal := range clause.Literals {
			value := assignment&(1<<literal.Var()) != 0
			satisfied = satisfied || value != literal.Negative()
		}
		if !satisfied {
			return false
		}
	}
	return true
}

func TestSatSolver2GateClausesExactOrderAndMetadata(t *testing.T) {
	and, err := SatSolver2AndClauses(2, 0, 1, true, false, true, 41)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]SatLit{{5, 1}, {5, 2}, {4, 0, 3}}
	if len(and) != len(want) {
		t.Fatalf("clause count = %d", len(and))
	}
	for i, clause := range and {
		if !reflect.DeepEqual(clause.Literals, want[i]) || !clause.PartA || clause.ID != 41 {
			t.Fatalf("clause %d = %+v, want %v", i, clause, want[i])
		}
	}
	xor, err := SatSolver2XorClauses(0, 1, 2, false, false, -1)
	if err != nil {
		t.Fatal(err)
	}
	if got := [][]SatLit{xor[0].Literals, xor[1].Literals, xor[2].Literals, xor[3].Literals}; !reflect.DeepEqual(got, [][]SatLit{{1, 3, 5}, {1, 2, 4}, {0, 3, 4}, {0, 2, 5}}) {
		t.Fatalf("xor clauses = %v", got)
	}
	clone := CloneSatSolver2ClauseSpecs(and)
	clone[0].Literals[0] = 0
	if and[0].Literals[0] != 5 {
		t.Fatal("clone aliases source")
	}
}

func TestSatSolver2GateClauseSemantics(t *testing.T) {
	for _, complement := range []bool{false, true} {
		constant, _ := SatSolver2ConstantClauses(0, complement, false, 1)
		for assignment := 0; assignment < 2; assignment++ {
			if got := satSolver2GateSatisfied(constant, assignment); got != ((assignment&1 != 0) != complement) {
				t.Fatalf("constant complement=%v assignment=%d got=%v", complement, assignment, got)
			}
		}
		buffer, _ := SatSolver2BufferClauses(0, 1, complement, false, 1)
		constraint, _ := SatSolver2ConstraintClauses(0, 1, complement, false, 1)
		for assignment := 0; assignment < 4; assignment++ {
			a := assignment&1 != 0
			b := assignment&2 != 0
			if got := satSolver2GateSatisfied(buffer, assignment); got != (a == (b != complement)) {
				t.Fatalf("buffer complement=%v assignment=%d got=%v", complement, assignment, got)
			}
			if got := satSolver2GateSatisfied(constraint, assignment); got != (a != complement) {
				t.Fatalf("constraint complement=%v assignment=%d got=%v", complement, assignment, got)
			}
		}
		for _, complement0 := range []bool{false, true} {
			and, _ := SatSolver2AndClauses(2, 0, 1, complement0, complement, false, 1)
			for assignment := 0; assignment < 8; assignment++ {
				a := assignment&1 != 0
				b := assignment&2 != 0
				output := assignment&4 != 0
				want := output == ((a != complement0) && (b != complement))
				if got := satSolver2GateSatisfied(and, assignment); got != want {
					t.Fatalf("and complements=%v/%v assignment=%d got=%v want=%v", complement0, complement, assignment, got, want)
				}
			}
		}
		xor, _ := SatSolver2XorClauses(0, 1, 2, complement, false, 1)
		for assignment := 0; assignment < 8; assignment++ {
			a := assignment&1 != 0
			b := assignment&2 != 0
			c := assignment&4 != 0
			want := a == (b != c != complement)
			if got := satSolver2GateSatisfied(xor, assignment); got != want {
				t.Fatalf("xor complement=%v assignment=%d got=%v want=%v", complement, assignment, got, want)
			}
		}
	}
}

func TestAddSatSolver2ClauseSpecsAtomicAndSolving(t *testing.T) {
	clauses, _ := SatSolver2AndClauses(2, 0, 1, false, false, true, 9)
	solver := NewSatSolver()
	if ok, err := AddSatSolver2ClauseSpecs(solver, clauses); err != nil || !ok {
		t.Fatalf("load = %v, %v", ok, err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		assumptions := make([]SatLit, 3)
		for variable := range assumptions {
			assumptions[variable], _ = MakeSatLit(SatVar(variable), assignment&(1<<variable) == 0)
		}
		result, err := solver.Solve(assumptions, SatLimits{})
		if err != nil || (result.Status == SatSatisfiable) != satSolver2GateSatisfied(clauses, assignment) {
			t.Fatalf("assignment=%d status=%d err=%v", assignment, result.Status, err)
		}
	}
	before := solver.VariableCount()
	bad := CloneSatSolver2ClauseSpecs(clauses)
	bad = append(bad, SatSolver2ClauseSpec{ID: 0, Literals: []SatLit{0}})
	if ok, err := AddSatSolver2ClauseSpecs(solver, bad); ok || !errors.Is(err, ErrDimension) || solver.VariableCount() != before {
		t.Fatalf("atomic invalid load = %v, %v vars=%d", ok, err, solver.VariableCount())
	}
	if ok, err := AddSatSolver2ClauseSpecs(nil, clauses); ok || !errors.Is(err, ErrDimension) {
		t.Fatalf("nil solver = %v, %v", ok, err)
	}
}

func TestSatSolver2GateClauseErrors(t *testing.T) {
	if _, err := SatSolver2ConstantClauses(-1, false, false, 1); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative variable: %v", err)
	}
	if _, err := SatSolver2BufferClauses(0, 1, false, false, 0); !errors.Is(err, ErrDimension) {
		t.Fatalf("zero ID: %v", err)
	}
	if _, err := SatSolver2AndClauses(int(MaxLiteralID)+1, 0, 1, false, false, false, 1); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("large variable: %v", err)
	}
}

func FuzzSatSolver2GateClauses(f *testing.F) {
	for _, seed := range []byte{0, 1, 7, 31} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, mask byte) {
		complement0 := mask&1 != 0
		complement1 := mask&2 != 0
		partA := mask&4 != 0
		and, err := SatSolver2AndClauses(2, 0, 1, complement0, complement1, partA, int(mask)+1)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			a := assignment&1 != 0
			b := assignment&2 != 0
			output := assignment&4 != 0
			want := output == ((a != complement0) && (b != complement1))
			if got := satSolver2GateSatisfied(and, assignment); got != want {
				t.Fatalf("mask=%d assignment=%d got=%v want=%v", mask, assignment, got, want)
			}
		}
		xor, err := SatSolver2XorClauses(0, 1, 2, mask&8 != 0, partA, int(mask)+1)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			a := assignment&1 != 0
			b := assignment&2 != 0
			c := assignment&4 != 0
			want := a == (b != c != (mask&8 != 0))
			if got := satSolver2GateSatisfied(xor, assignment); got != want {
				t.Fatalf("xor mask=%d assignment=%d got=%v want=%v", mask, assignment, got, want)
			}
		}
	})
}
