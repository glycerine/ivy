package bsat

import "testing"

func solveSevenPigeons(t testing.TB) Statistics {
	s := NewSolver()
	defer s.Close()
	s.SetNumVars(42)
	for p := int32(0); p < 7; p++ {
		var row [6]Literal
		for h := int32(0); h < 6; h++ {
			row[h] = Positive(6*p + h)
		}
		if !s.AddClause(row[:]...) {
			t.Fatal("inconsistent pigeon clauses")
		}
	}
	for h := int32(0); h < 6; h++ {
		for p := int32(0); p < 7; p++ {
			for q := p + 1; q < 7; q++ {
				if !s.AddClause(Negative(6*p+h), Negative(6*q+h)) {
					t.Fatal("inconsistent hole clauses")
				}
			}
		}
	}
	if s.Solve() != Unsatisfiable {
		t.Fatal("expected seven pigeons in six holes to be UNSAT")
	}
	return s.Statistics()
}

func BenchmarkPigeonhole7Into6(b *testing.B) {
	b.ReportAllocs()
	var stats Statistics
	for b.Loop() {
		stats = solveSevenPigeons(b)
	}
	b.ReportMetric(float64(stats.Conflicts), "conflicts/op")
	b.ReportMetric(float64(stats.Decisions), "decisions/op")
	b.ReportMetric(float64(stats.Propagations), "propagations/op")
}

func TestSevenPigeonsNativeSearch(t *testing.T) {
	// From test_vectors/perf_pigeonhole.c with the same clause ordering.
	stats := solveSevenPigeons(t)
	if stats.Conflicts != 604 || stats.Decisions != 788 || stats.Propagations != 7409 {
		t.Fatalf("native search counts 604/788/7409, Go %d/%d/%d", stats.Conflicts, stats.Decisions, stats.Propagations)
	}
}
