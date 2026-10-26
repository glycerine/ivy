package abc

// Stand-alone tests for the satUtil solver diagnostics and snapshots.

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSatSolverDIMACSExactAndFile(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	y := satStoreTestLit(t, 1, false)
	z := satStoreTestLit(t, 2, false)
	solver := NewSatSolver()
	if ok, err := solver.AddClause([]SatLit{x, y.Not()}); err != nil || !ok {
		t.Fatalf("add binary: ok=%v err=%v", ok, err)
	}
	if ok, err := solver.AddClause([]SatLit{z}); err != nil || !ok {
		t.Fatalf("add unit: ok=%v err=%v", ok, err)
	}
	standard, err := SatSolverDIMACS(solver, []SatLit{x.Not()}, SatSolverDIMACSOptions{IncrementVariables: true})
	if err != nil {
		t.Fatal(err)
	}
	want := "p cnf 3 3\n1 -2 0\n3 0\n-1 0\n\n"
	if string(standard) != want {
		t.Fatalf("standard snapshot = %q, want %q", standard, want)
	}
	raw, err := SatSolverDIMACS(solver, []SatLit{x.Not()}, SatSolverDIMACSOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRaw := "p cnf 3 3\n0 -1 \n2 \n-0 \n\n"
	if string(raw) != wantRaw {
		t.Fatalf("raw snapshot = %q, want %q", raw, wantRaw)
	}
	formula, err := ParseCNFDIMACS(standard, CNFDIMACSReadOptions{})
	if err != nil || formula.variables != 3 || len(formula.clauses) != 3 {
		t.Fatalf("round trip = %+v, err=%v", formula, err)
	}
	path := filepath.Join(t.TempDir(), "solver.cnf")
	if err = WriteSatSolverDIMACSFile(solver, path, []SatLit{x.Not()}, SatSolverDIMACSOptions{IncrementVariables: true}); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(path)
	if err != nil || !reflect.DeepEqual(written, standard) {
		t.Fatalf("file = %q, err=%v", written, err)
	}
	if after, _ := SatSolverDIMACS(solver, []SatLit{x.Not()}, SatSolverDIMACSOptions{IncrementVariables: true}); !reflect.DeepEqual(after, standard) {
		t.Fatal("serialization mutated solver")
	}
}

func TestSatSolverDIMACSInconsistentAndClauseFormat(t *testing.T) {
	solver := NewSatSolver()
	if err := solver.SetVariableCount(1); err != nil {
		t.Fatal(err)
	}
	if ok, err := solver.AddClause(nil); err != nil || ok {
		t.Fatalf("empty clause: ok=%v err=%v", ok, err)
	}
	data, err := SatSolverDIMACS(solver, nil, SatSolverDIMACSOptions{IncrementVariables: true})
	if err != nil || string(data) != "p cnf 1 1\n0\n\n" {
		t.Fatalf("inconsistent = %q, err=%v", data, err)
	}
	formula, err := ParseCNFDIMACS(data, CNFDIMACSReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	check := NewSatSolver()
	if ok, loadErr := formula.LoadInto(check, 0); loadErr != nil || ok {
		t.Fatalf("empty round trip: ok=%v err=%v", ok, loadErr)
	}
	x := satStoreTestLit(t, 0, false)
	if text, formatErr := FormatSatSolverClauseDIMACS([]SatLit{x, x.Not()}, true); formatErr != nil || text != "1 -1 0\n" {
		t.Fatalf("clause format = %q, err=%v", text, formatErr)
	}
	if _, formatErr := FormatSatSolverClauseDIMACS([]SatLit{SatLit(math.MaxUint32)}, true); !errors.Is(formatErr, ErrInvalidLiteral) {
		t.Fatalf("invalid format: %v", formatErr)
	}
}

func TestSatSolverStatisticsMemoryAndModel(t *testing.T) {
	stats := SatStats{Restarts: 1, Conflicts: 2, Decisions: 3, Propagations: 4}
	wantFirst := "starts        :                1\nconflicts     :                2\ndecisions     :                3\npropagations  :                4\n"
	if got := FormatSatSolverStatistics(stats, false); got != wantFirst {
		t.Fatalf("first stats = %q", got)
	}
	wantSecond := "starts        :          1\nconflicts     :          2\ndecisions     :          3\npropagations  :          4\n"
	if got := FormatSatSolverStatistics(stats, true); got != wantSecond {
		t.Fatalf("second stats = %q", got)
	}
	for _, value := range []uint64{0, 1, 1<<30 - 1, 1 << 30, 1<<32 + 17, 1<<53 - 1} {
		if got := SatLegacyCounterFloat(value); got != float64(value) {
			t.Fatalf("counter(%d) = %.0f", value, got)
		}
	}
	if bytes, err := SatSolver2VariableBytes(8, false); err != nil || bytes != 64 {
		t.Fatalf("64-bit integer activity bytes = %d, err=%v", bytes, err)
	}
	if bytes, err := SatSolver2VariableBytes(4, false); err != nil || bytes != 56 {
		t.Fatalf("32-bit integer activity bytes = %d, err=%v", bytes, err)
	}
	if bytes, err := SatSolver2VariableBytes(8, true); err != nil || bytes != 68 {
		t.Fatalf("64-bit float activity bytes = %d, err=%v", bytes, err)
	}
	if NativeSatSolver2VariableBytes() != 64 {
		t.Fatalf("native bytes = %d", NativeSatSolver2VariableBytes())
	}
	if _, err := SatSolver2VariableBytes(3, false); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad pointer size: %v", err)
	}

	solver := NewSatSolver()
	x := satStoreTestLit(t, 0, false)
	y := satStoreTestLit(t, 1, false)
	_, _ = solver.AddClause([]SatLit{x})
	_, _ = solver.AddClause([]SatLit{y.Not()})
	result, err := solver.Solve(nil, SatLimits{})
	if err != nil {
		t.Fatal(err)
	}
	model, err := SatSolverModel(result, []SatVar{1, 0})
	if err != nil || !reflect.DeepEqual(model, []int{0, 1}) {
		t.Fatalf("model = %v, err=%v", model, err)
	}
	if _, err = SatSolverModel(SatResult{Status: SatUnsatisfiable}, nil); !errors.Is(err, ErrInvariant) {
		t.Fatalf("non-SAT model: %v", err)
	}
	if _, err = SatSolverModel(result, []SatVar{2}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("bad model variable: %v", err)
	}
}

func TestSatSolverDIMACSErrorsLimitsCancellationAndDisabledDouble(t *testing.T) {
	solver := NewSatSolver()
	x := satStoreTestLit(t, 0, false)
	_, _ = solver.AddClause([]SatLit{x})
	full, err := SatSolverDIMACS(solver, nil, SatSolverDIMACSOptions{IncrementVariables: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = SatSolverDIMACS(solver, nil, SatSolverDIMACSOptions{IncrementVariables: true, MaxBytes: len(full) - 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit: %v", err)
	}
	if _, err = SatSolverDIMACS(solver, nil, SatSolverDIMACSOptions{IncrementVariables: true, Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %v", err)
	}
	if _, err = SatSolverDIMACS(solver, []SatLit{satStoreTestLit(t, 1, false)}, SatSolverDIMACSOptions{}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("bad assumption: %v", err)
	}
	if _, err = SatSolverDIMACS(nil, nil, SatSolverDIMACSOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("nil solver: %v", err)
	}
	if _, err = SatSolverDIMACS(solver, nil, SatSolverDIMACSOptions{MaxBytes: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative limit: %v", err)
	}
	if err = WriteSatSolverDIMACSFile(solver, "", nil, SatSolverDIMACSOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("empty path: %v", err)
	}
	if err = WriteSatSolverDIMACSFile(solver, t.TempDir(), nil, SatSolverDIMACSOptions{}); err == nil {
		t.Fatal("directory write succeeded")
	}
	before, _ := SatSolverDIMACS(solver, nil, SatSolverDIMACSOptions{IncrementVariables: true})
	if err = solver.DoubleClauses(0); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("double clauses: %v", err)
	}
	after, _ := SatSolverDIMACS(solver, nil, SatSolverDIMACSOptions{IncrementVariables: true})
	if !reflect.DeepEqual(before, after) {
		t.Fatal("disabled doubling mutated solver")
	}
}

func FuzzSatSolverDIMACSRoundTrip(f *testing.F) {
	for _, seed := range []struct {
		variables int
		data      []byte
	}{{1, []byte{0}}, {3, []byte{0, 3, 6, 9}}, {8, []byte{1, 2, 4, 8, 16}}} {
		f.Add(seed.variables, seed.data)
	}
	f.Fuzz(func(t *testing.T, variables int, data []byte) {
		variables = 1 + int(uint(variables)%8)
		solver := NewSatSolver()
		clauses := make([][]SatLit, 0)
		for offset := 0; offset < len(data) && len(clauses) < 24; {
			length := 1 + int(data[offset]%3)
			offset++
			clause := make([]SatLit, 0, length)
			for i := 0; i < length && offset < len(data); i, offset = i+1, offset+1 {
				literal, err := MakeSatLit(SatVar(int(data[offset])%variables), data[offset]&0x80 != 0)
				if err != nil {
					t.Fatal(err)
				}
				clause = append(clause, literal)
			}
			clauses = append(clauses, clause)
			if ok, err := solver.AddClause(clause); err != nil {
				t.Fatal(err)
			} else if !ok {
				break
			}
		}
		if err := solver.SetVariableCount(variables); err != nil {
			t.Fatal(err)
		}
		assumptions := make([]SatLit, 0, 2)
		for i := 0; i < len(data) && i < 2; i++ {
			literal, _ := MakeSatLit(SatVar(int(data[len(data)-1-i])%variables), data[len(data)-1-i]&0x40 != 0)
			assumptions = append(assumptions, literal)
		}
		original, err := solver.Solve(assumptions, SatLimits{})
		if err != nil {
			t.Fatal(err)
		}
		text, err := SatSolverDIMACS(solver, assumptions, SatSolverDIMACSOptions{IncrementVariables: true, MaxBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		formula, err := ParseCNFDIMACS(text, CNFDIMACSReadOptions{})
		if err != nil {
			t.Fatalf("parse %q: %v", text, err)
		}
		replay := NewSatSolver()
		ok, err := formula.LoadInto(replay, 0)
		if err != nil {
			t.Fatal(err)
		}
		status := SatUnsatisfiable
		if ok {
			result, solveErr := replay.Solve(nil, SatLimits{})
			if solveErr != nil {
				t.Fatal(solveErr)
			}
			status = result.Status
		}
		if status != original.Status {
			t.Fatalf("status mismatch: original=%v roundtrip=%v text=%q clauses=%v assumptions=%v", original.Status, status, text, clauses, assumptions)
		}
	})
}
