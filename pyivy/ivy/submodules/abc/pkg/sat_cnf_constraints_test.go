package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func cnfConstraintGraph(t testing.TB, n int) *AigManager {
	t.Helper()
	g := NewAigManager()
	x, _ := g.CreateCI()
	y, _ := g.CreateCI()
	outputs := []AigLit{x, y.Not(), 0, 1}
	for _, l := range outputs[:n] {
		g.CreateCO(l)
	}
	return g
}
func TestCNFConstraintsNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_constraints.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Kind, Outputs int
		A, B, C       SatVar
		Clauses       [][]SatLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 36 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		var clauses [][]SatLit
		var e error
		if row.Kind == 2 {
			clauses, e = CNFXORClauses(row.A, row.B, row.C)
		} else {
			g := cnfConstraintGraph(t, row.Outputs)
			c, err := DeriveAigCNF4(g, row.Outputs, AigCNFDeriveOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if row.Kind == 0 {
				var clause []SatLit
				clause, e = c.OutputORClause(g, 3)
				clauses = [][]SatLit{clause}
			} else {
				clauses, e = c.OutputANDClauses(g, 3)
			}
		}
		if e != nil || !equalSatClauses(clauses, row.Clauses) {
			t.Fatal(row, clauses, e)
		}
	}
}
func TestCNFConstraintsSemantics(t *testing.T) {
	for a := SatVar(1); a <= 3; a++ {
		for b := SatVar(1); b <= 3; b++ {
			for c := SatVar(1); c <= 3; c++ {
				s := NewSatSolver()
				s.SetVariableCount(4)
				ok, e := AddCNFXOR(s, a, b, c)
				if e != nil || !ok {
					t.Fatal(ok, e)
				}
				for assignment := 0; assignment < 8; assignment++ {
					var assumptions []SatLit
					for v := 1; v <= 3; v++ {
						l := SatLit(v) << 1
						if assignment>>uint(v-1)&1 == 0 {
							l = l.Not()
						}
						assumptions = append(assumptions, l)
					}
					answer, e := s.Solve(assumptions, SatLimits{})
					want := SatUnsatisfiable
					if ((assignment>>uint(a-1))^(assignment>>uint(b-1)))&1 == (assignment>>uint(c-1))&1 {
						want = SatSatisfiable
					}
					if e != nil || answer.Status != want {
						t.Fatal(a, b, c, assignment, answer, e)
					}
				}
			}
		}
	}
	for n := 0; n <= 4; n++ {
		for _, all := range []bool{false, true} {
			g := cnfConstraintGraph(t, n)
			c, e := DeriveAigCNF4(g, n, AigCNFDeriveOptions{})
			if e != nil {
				t.Fatal(e)
			}
			phase, e := c.TransformPolarity(g, true, CNFTransformOptions{})
			if e != nil {
				t.Fatal(e)
			}
			s := NewSatSolver()
			ok, e := phase.LoadInto(s, 3)
			if e != nil || !ok {
				t.Fatal(ok, e)
			}
			if all {
				ok, e = phase.AddOutputAND(s, g, 3)
			} else {
				ok, e = phase.AddOutputOR(s, g, 3)
			}
			if e != nil {
				t.Fatal(e)
			}
			for assignment := 0; assignment < 4; assignment++ {
				var assumptions []SatLit
				for i, l := range phase.inputs {
					l += 6
					if assignment>>uint(i)&1 == 0 {
						l = l.Not()
					}
					assumptions = append(assumptions, l)
				}
				values := []bool{assignment&1 != 0, assignment&2 == 0, true, false}
				want := all
				for _, v := range values[:n] {
					if all {
						want = want && v
					} else {
						want = want || v
					}
				}
				answer, e := s.Solve(assumptions, SatLimits{})
				expected := SatUnsatisfiable
				if want {
					expected = SatSatisfiable
				}
				if e != nil || answer.Status != expected {
					t.Fatal(n, all, assignment, ok, answer, e)
				}
			}
		}
	}
	if _, e := CNFXORClauses(0, 1, 2); e == nil {
		t.Fatal("zero XOR variable")
	}
	g := cnfConstraintGraph(t, 2)
	asserted, e := DeriveAigCNF4(g, 0, AigCNFDeriveOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = asserted.OutputORClause(g, 0); e == nil {
		t.Fatal("unassigned CO")
	}
}

func TestCNFOutputORSolver2(t *testing.T) {
	g := cnfConstraintGraph(t, 2)
	cnf, err := DeriveAigCNF4(g, 2, AigCNFDeriveOptions{})
	if err != nil {
		t.Fatal(err)
	}
	frame, err := cnf.UnrollFrames(g, 1, false, CNFFrameOptions{Base: 3})
	if err != nil {
		t.Fatal(err)
	}
	solver, ok, err := frame.NewSolver2(SatSolver2Options{})
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	before := solver.ClauseCount()
	if ok, err = cnf.AddOutputORSolver2(solver, g, 3, 91); err != nil || !ok || solver.ClauseCount() != before+1 {
		t.Fatal(ok, err, solver.ClauseCount(), before)
	}
	stored, err := solver.Clause(before + 1)
	want, _ := cnf.OutputORClause(g, 3)
	slices.Sort(want)
	if err != nil || stored.ID != 91 || !slices.Equal(stored.Literals, want) {
		t.Fatal(stored, want, err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := make([]SatLit, len(cnf.inputs))
		for index, literal := range cnf.inputs {
			literal += 6
			if assignment>>index&1 == 0 {
				literal = literal.Not()
			}
			inputs[index] = literal
		}
		result, err := solver.Solve(inputs, SatLimits{})
		wantStatus := SatUnsatisfiable
		if assignment != 2 { // outputs are x and !y
			wantStatus = SatSatisfiable
		}
		if err != nil || result.Status != wantStatus {
			t.Fatal(assignment, result, err)
		}
	}
	emptyGraph := cnfConstraintGraph(t, 0)
	emptyCNF, _ := DeriveAigCNF4(emptyGraph, 0, AigCNFDeriveOptions{})
	before = solver.ClauseCount()
	if ok, err := emptyCNF.AddOutputORSolver2(solver, emptyGraph, 0, 1); ok || !errors.Is(err, ErrDimension) || solver.ClauseCount() != before {
		t.Fatal(ok, err, solver.ClauseCount(), before)
	}
}
func FuzzCNFXORConstraints(f *testing.F) {
	f.Add(uint8(1), uint8(2), uint8(3))
	f.Fuzz(func(t *testing.T, x, y, z uint8) {
		a, b, c := SatVar(1+x%7), SatVar(1+y%7), SatVar(1+z%7)
		clauses, e := CNFXORClauses(a, b, c)
		if e != nil {
			t.Fatal(e)
		}
		for assignment := 0; assignment < 128; assignment++ {
			sat := true
			for _, clause := range clauses {
				ok := false
				for _, l := range clause {
					v := assignment>>uint(l.Var()-1)&1 != 0
					if l&1 != 0 {
						v = !v
					}
					ok = ok || v
				}
				sat = sat && ok
			}
			want := ((assignment>>uint(a-1))^(assignment>>uint(b-1)))&1 == (assignment>>uint(c-1))&1
			if sat != want {
				t.Fatal(a, b, c, assignment)
			}
		}
	})
}
