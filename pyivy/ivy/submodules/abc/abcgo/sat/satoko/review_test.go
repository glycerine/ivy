package satoko

import (
	"bytes"
	"reflect"
	"testing"
)

func TestInvalidConeDoesNotPartiallyMark(t *testing.T) {
	s := New()
	s.SetVariables(2)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("invalid cone accepted")
			}
		}()
		s.MarkCone([]int{0, 2}, true)
	}()
	if s.coneActive || s.marks[0] || s.marks[1] {
		t.Fatal("invalid cone changed selection")
	}
}

func TestInvalidAssumptionsPreserveContext(t *testing.T) {
	s := New()
	s.SetVariables(1)
	s.PushAssumption(0)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected invalid assumption panic")
			}
		}()
		s.SolveAssumptions([]int{3})
	}()
	if !reflect.DeepEqual(s.assumptions, []int{0}) {
		t.Fatal("invalid request replaced original assumptions", s.assumptions)
	}
	if s.Solve() != SATOKO_SAT || s.ModelValue(0) != 1 {
		t.Fatal("original context corrupted")
	}
}
func TestTemporarySolveOptionsRestoredAfterCallbackPanic(t *testing.T) {
	s := New()
	s.SetVariables(1)
	s.PushAssumption(0)
	s.opts.ConflictLimit = 77
	s.SetStopFunc(func(int) int { panic("callback") })
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected callback panic")
			}
		}()
		s.SolveAssumptionsLimit([]int{1}, 3)
	}()
	if s.opts.ConflictLimit != 77 || !reflect.DeepEqual(s.assumptions, []int{0}) {
		t.Fatal("temporary solve context leaked", s.opts.ConflictLimit, s.assumptions)
	}
	s.SetStopFunc(nil)
	if s.Solve() != SATOKO_SAT {
		t.Fatal("solver unusable after recovered callback")
	}
}
func TestMinimizeFinalCoreMatchesResult(t *testing.T) {
	s := New()
	s.SetVariables(3)
	s.AddClause([]int{0, 2})
	out, result := s.MinimizeAssumptions([]int{4, 1, 3}, 0)
	if result != SATOKO_UNSAT {
		t.Fatal(result)
	}
	want := make([]int, len(out))
	for i, lit := range out {
		want[i] = lit ^ 1
	}
	if !reflect.DeepEqual(s.FinalConflict(), want) {
		t.Fatal("minimization returned stale core", s.FinalConflict(), want)
	}
}
func TestZeroBasedDumpHasNoTerminator(t *testing.T) {
	s := New()
	s.AddClause([]int{0})
	var b bytes.Buffer
	if err := s.WriteDIMACS(&b, false, true); err != nil {
		t.Fatal(err)
	}
	if b.String() != "p cnf 1 1\n0 \n" {
		t.Fatal(b.String())
	}
}

func TestRollbackRestoresLiteralStatistics(t *testing.T) {
	s := New()
	s.SetVariables(2)
	s.AddClause([]int{0, 2})
	s.Bookmark()
	s.AddClause([]int{1, 2})
	s.Rollback()
	if s.stats.OriginalLiterals != 2 || s.stats.LearntLiterals != 0 {
		t.Fatal("rolled-back clauses counted in statistics", s.stats)
	}
}
func TestResetPreservesCancellationSettings(t *testing.T) {
	s := New()
	s.SetVariables(1)
	stop := 1
	s.SetStopPointer(&stop)
	s.SetRunID(14)
	s.Reset()
	s.SetVariables(1)
	if s.stop != &stop || s.runID != 14 || s.Solve() != SATOKO_UNDEC {
		t.Fatal("reset lost cancellation settings")
	}
}

func TestRollbackReleasesClausesAndPreservesUnitStatistics(t *testing.T) {
	s := New()
	s.SetVariables(6)
	s.AddClause([]int{0})
	s.AddClause([]int{2, 4})
	s.addStored([]int{6}, true, 1)
	s.addStored([]int{8, 10}, true, 2)
	before := *s.Stats()
	s.Bookmark()
	firstRemoved := len(s.clauses)
	s.AddClause([]int{3, 5, 7})
	s.Rollback()
	if got := s.Stats(); got.OriginalLiterals != before.OriginalLiterals || got.LearntLiterals != before.LearntLiterals {
		t.Errorf("restored original and learnt literals = %d, %d; want %d, %d", got.OriginalLiterals, got.LearntLiterals, before.OriginalLiterals, before.LearntLiterals)
	}
	for _, c := range s.clauses[:cap(s.clauses)][firstRemoved:] {
		if c != nil {
			t.Fatal("clause array retains a rolled-back clause")
		}
	}
}

func TestClauseInputWidthsAndOwnership(t *testing.T) {
	for _, size := range []int{2, 5, 6, 16, 17, 129} {
		for _, compactInput := range []bool{false, true} {
			input := make([]int, size+1)
			input32 := make([]int32, len(input))
			want := make([]int, size)
			for i := range size {
				want[i] = 2 * i
				input[i] = 2 * (size - 1 - i)
				input32[i] = int32(input[i])
			}
			input[size], input32[size] = input[0], input32[0]
			s := New()
			if compactInput {
				if !s.AddClause32(input32) {
					t.Fatal("rejected clause")
				}
			} else if !s.AddClause(input) {
				t.Fatal("rejected clause")
			}
			for i := range input {
				if int(input32[i]) != input[i] {
					t.Fatal("insertion changed caller's literal order")
				}
				input[i], input32[i] = 1, 1
			}
			if s.VariableCount() != size || s.Stats().OriginalLiterals != int64(size) || !reflect.DeepEqual(s.clauses[0].lits, want) {
				t.Fatalf("size %d compact input %t: stored %v, want %v", size, compactInput, s.clauses[0].lits, want)
			}
			for _, lit := range want[:len(want)-1] {
				if !s.AddClause([]int{lit ^ 1}) {
					t.Fatal("unexpected unit conflict")
				}
			}
			if s.Solve() != SATOKO_SAT || s.ModelValue(size-1) != 1 {
				t.Fatal("stored clause did not force its last variable")
			}
		}
	}
}

func TestAddClause32ValidatesBeforeAddingVariables(t *testing.T) {
	s := New()
	defer func() {
		if recover() == nil || s.VariableCount() != 0 || len(s.clauses) != 0 {
			t.Fatal("invalid literal partially changed the solver")
		}
	}()
	s.AddClause32([]int32{120, 2, -1})
}
