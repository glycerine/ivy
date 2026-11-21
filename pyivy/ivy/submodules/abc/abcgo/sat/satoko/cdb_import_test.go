package satoko

import (
	"reflect"
	"testing"
)

func TestBulkImportPreservesWatchOrderAndClauseOwnership(t *testing.T) {
	records := make([][]int32, 0, 700)
	for i := range 700 {
		x := int32(i%127+1) * 2
		clause := []int32{x, 0}
		if i%4 == 0 {
			clause = []int32{x, 0, 258, 260, 262, 264}
		}
		records = append(records, clause)
	}
	single := New()
	for _, clause := range records {
		if !single.AddClause32(clause) {
			t.Fatal("sequential clause rejected")
		}
	}
	bulk := NewFromClauses32(100000000, records)
	if bulk == nil || bulk.importing != nil || bulk.VariableCount() != single.VariableCount() {
		t.Fatal("bulk import changed live variables or retained import scratch")
	}
	if !reflect.DeepEqual(bulk.watches, single.watches) || !reflect.DeepEqual(bulk.clauses, single.clauses) {
		t.Fatal("bulk import changed watch or clause order")
	}
	for _, clause := range records {
		for i := range clause {
			clause[i] = 1
		}
	}
	for _, s := range []*Solver{single, bulk} {
		s.Bookmark()
		if !s.AddClause32([]int32{1}) || s.Solve() != SATOKO_SAT {
			t.Fatal("unexpected conflict after extending solver")
		}
		s.Rollback()
	}
	if bulk.ClauseCount() != len(records) || !reflect.DeepEqual(bulk.clauses, single.clauses) {
		t.Fatal("rollback failed to restore imported records")
	}
	if got, want := bulk.SolveAssumptions([]int{1, 3}), single.SolveAssumptions([]int{1, 3}); got != want || got != SATOKO_UNSAT {
		t.Fatalf("assumption results %d and %d", got, want)
	}
}

func TestBulkImportStopsAtFirstConflict(t *testing.T) {
	if NewFromClauses32(1000000, [][]int32{{0}, {1}, {-1}}) != nil {
		t.Fatal("accepted contradictory units")
	}
	for _, hint := range []int{-1, 0, 1000000} {
		s := NewFromClauses32(hint, [][]int32{{8, 2, 8}})
		if s == nil || s.VariableCount() != 5 || s.Stats().OriginalLiterals != 2 {
			t.Fatal("hint changed formula", hint)
		}
	}
}

func BenchmarkImportClauseDistribution(b *testing.B) {
	for _, tc := range []struct {
		name               string
		clauses, variables int
	}{
		{"one", 1, 2},
		{"sparse", 2048, 4096},
		{"shared", 2048, 71},
	} {
		clauses := make([][]int32, tc.clauses)
		for i := range clauses {
			left := int32(2*i%tc.variables) * 2
			right := int32((2*i+1)%tc.variables) * 2
			clauses[i] = []int32{right, left}
		}
		b.Run(tc.name+"/sequential", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				s := New()
				for _, clause := range clauses {
					if !s.AddClause32(clause) {
						b.Fatal("clause rejected")
					}
				}
				s.Destroy()
			}
		})
		b.Run(tc.name+"/bulk", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				s := NewFromClauses32(tc.variables, clauses)
				if s == nil {
					b.Fatal("clause import failed")
				}
				s.Destroy()
			}
		})
	}
}
