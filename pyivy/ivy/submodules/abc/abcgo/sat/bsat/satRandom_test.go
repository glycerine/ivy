package bsat

import (
	"math"
	"testing"
)

// The floating-point recurrence in src/sat/bsat/satSolver.c also accepts
// fractional seeds through ABC's seeded constructor.
func TestRandomMatchesNativeRecurrence(t *testing.T) {
	for _, seed := range []float64{0, math.Copysign(0, -1), 1, 42, 91648253, 2147483646, 2147483647, 0.25, 1.125, -17.75, 3e9} {
		have, want := seed, seed
		for i := 0; i < 10000; i++ {
			want *= 1389796
			want -= float64(int32(want/2147483647)) * 2147483647
			value := drand(&have)
			if math.Float64bits(have) != math.Float64bits(want) || math.Float64bits(value) != math.Float64bits(want/2147483647) {
				t.Fatalf("seed %g step %d: state %g draw %g, want state %g", seed, i, have, value, want)
			}
		}
	}
}

func BenchmarkSATDecisionRandom(b *testing.B) {
	seed := 91648253.0
	for b.Loop() {
		drand(&seed)
	}
}

func TestSolverRandomDeterministic(t *testing.T) {
	a, b := 91648253.0, 91648253.0
	for i := 0; i < 10000; i++ {
		x, y := drand(&a), drand(&b)
		if x != y || x < 0 || x >= 1 {
			t.Fatal(i, x, y)
		}
	}
	a = 91648253
	for i := 0; i < 10000; i++ {
		x := irand(&a, 17)
		if x < 0 || x >= 17 {
			t.Fatal(x)
		}
	}
}
