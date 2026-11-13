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
