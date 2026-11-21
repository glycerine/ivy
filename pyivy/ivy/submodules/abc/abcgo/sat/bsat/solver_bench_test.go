package bsat

import "testing"

func solvePigeons(t testing.TB, holes int32) Statistics {
	s := NewSolver()
	defer s.Close()
	s.SetNumVars(holes * (holes + 1))
	for p := int32(0); p <= holes; p++ {
		var row [7]Literal
		for h := int32(0); h < holes; h++ {
			row[h] = Positive(holes*p + h)
		}
		if !s.AddClause(row[:holes]...) {
			t.Fatal("inconsistent pigeon clauses")
		}
	}
	for h := int32(0); h < holes; h++ {
		for p := int32(0); p <= holes; p++ {
			for q := p + 1; q <= holes; q++ {
				if !s.AddClause(Negative(holes*p+h), Negative(holes*q+h)) {
					t.Fatal("inconsistent hole clauses")
				}
			}
		}
	}
	if s.Solve() != Unsatisfiable {
		t.Fatalf("expected %d pigeons in %d holes to be UNSAT", holes+1, holes)
	}
	return s.Statistics()
}

func BenchmarkPigeonhole7Into6(b *testing.B) { benchmarkPigeons(b, 6) }
func BenchmarkPigeonhole8Into7(b *testing.B) { benchmarkPigeons(b, 7) }

func benchmarkPigeons(b *testing.B, holes int32) {
	b.ReportAllocs()
	var stats Statistics
	for b.Loop() {
		stats = solvePigeons(b, holes)
	}
	b.ReportMetric(float64(stats.Conflicts), "conflicts/op")
	b.ReportMetric(float64(stats.Decisions), "decisions/op")
	b.ReportMetric(float64(stats.Propagations), "propagations/op")
}

func TestSevenPigeonsNativeSearch(t *testing.T) {
	// From test_vectors/perf_pigeonhole.c with the same clause ordering.
	stats := solvePigeons(t, 6)
	if stats.Conflicts != 604 || stats.Decisions != 788 || stats.Propagations != 7409 {
		t.Fatalf("native search counts 604/788/7409, Go %d/%d/%d", stats.Conflicts, stats.Decisions, stats.Propagations)
	}
}

func TestEightPigeonsNativeSearch(t *testing.T) {
	stats := solvePigeons(t, 7)
	if stats.Conflicts != 3378 || stats.Decisions != 4221 || stats.Propagations != 48104 {
		t.Fatalf("native search counts 3378/4221/48104, Go %d/%d/%d", stats.Conflicts, stats.Decisions, stats.Propagations)
	}
}
