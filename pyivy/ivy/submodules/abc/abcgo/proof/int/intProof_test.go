package int

import "testing"

// A=(x OR y), B=(!x AND !y). Both proof procedures must derive an
// interpolant satisfying A => I and I => !B for every shared assignment.
func TestMiniSatProofTwoSharedVariables(t *testing.T) {
	proof := &fixtureM114p{
		roots:  [][]int32{{0, 2}, {1}, {3}},
		chains: []M114pChain{{Clauses: []int32{0, 1}, Pivots: []int32{0}}, {Clauses: []int32{3, 2}, Pivots: []int32{1}}},
	}
	for _, pudlak := range []bool{false, true} {
		graph, err := InterpolateM114p(proof, []int32{0, 1, 1}, []int32{0, 1}, pudlak)
		if err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 4; mask++ {
			x, y := mask&1 != 0, mask&2 != 0
			out, err := graph.Evaluate([]bool{x, y})
			if err != nil || len(out) != 1 || out[0] != (x || y) {
				t.Fatalf("pudlak=%v x=%v y=%v: %v %v", pudlak, x, y, out, err)
			}
		}
	}
}

// A=(a AND (!a OR x)), B=(b AND (!b OR !x)). The source
// combines A-local pivots with OR, B-local pivots with AND, and shared
// pivots with either AND (McMillan) or a polarity-sensitive mux (Pudlak).
func TestMiniSatLocalAndSharedResolutionBranches(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		last := []int32{4, 5}
		if reverse {
			last = []int32{5, 4}
		}
		proof := &fixtureM114p{
			roots: [][]int32{{0}, {1, 2}, {4}, {5, 3}},
			chains: []M114pChain{
				{Clauses: []int32{0, 1}, Pivots: []int32{0}},
				{Clauses: []int32{2, 3}, Pivots: []int32{2}},
				{Clauses: last, Pivots: []int32{1}},
			},
		}
		for _, pudlak := range []bool{false, true} {
			g, err := InterpolateM114p(proof, []int32{0, 0, 1, 1}, []int32{-1, 0, -2}, pudlak)
			if err != nil {
				t.Fatalf("reverse=%v pudlak=%v: %v", reverse, pudlak, err)
			}
			for _, x := range []bool{false, true} {
				out, err := g.Evaluate([]bool{x})
				if err != nil || len(out) != 1 || out[0] != x {
					t.Fatalf("reverse=%v pudlak=%v x=%v output=%v err=%v", reverse, pudlak, x, out, err)
				}
			}
		}
	}
}

func TestMiniSatLocalPivotOperators(t *testing.T) {
	tests := []struct {
		name           string
		roots          [][]int32
		parts, classes []int32
		chains         []M114pChain
		want           func(bool, bool) bool
	}{
		{"A-local-or", [][]int32{{0, 2}, {1, 4}, {3}, {5}}, []int32{0, 0, 1, 1}, []int32{-1, 0, 1}, []M114pChain{
			{Clauses: []int32{0, 1}, Pivots: []int32{0}},
			{Clauses: []int32{4, 2}, Pivots: []int32{1}},
			{Clauses: []int32{5, 3}, Pivots: []int32{2}},
		}, func(x, y bool) bool { return x || y }},
		{"B-local-and", [][]int32{{2}, {4}, {0, 3}, {1, 5}}, []int32{0, 0, 1, 1}, []int32{-2, 0, 1}, []M114pChain{
			{Clauses: []int32{0, 2}, Pivots: []int32{1}},
			{Clauses: []int32{1, 3}, Pivots: []int32{2}},
			{Clauses: []int32{4, 5}, Pivots: []int32{0}},
		}, func(x, y bool) bool { return x && y }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			proof := &fixtureM114p{roots: tc.roots, chains: tc.chains}
			for _, pudlak := range []bool{false, true} {
				g, err := InterpolateM114p(proof, tc.parts, tc.classes, pudlak)
				if err != nil {
					t.Fatal(err)
				}
				for mask := 0; mask < 4; mask++ {
					x, y := mask&1 != 0, mask&2 != 0
					out, err := g.Evaluate([]bool{x, y})
					if err != nil || out[0] != tc.want(x, y) {
						t.Fatalf("pudlak=%v x=%v y=%v output=%v err=%v", pudlak, x, y, out, err)
					}
				}
			}
		})
	}
}

// In the source map, a shared SAT variable may be numbered after another
// shared input even when the earlier input is unused by the refutation.
func TestMiniSatSharedRootAnnotationUsesMappedInputOrder(t *testing.T) {
	for _, negated := range []bool{false, true} {
		first, second := int32(0), int32(1)
		if negated {
			first, second = second, first
		}
		proof := &fixtureM114p{roots: [][]int32{{first}, {second}}, chains: []M114pChain{{Clauses: []int32{0, 1}, Pivots: []int32{0}}}}
		for _, pudlak := range []bool{false, true} {
			g, err := InterpolateM114p(proof, []int32{0, 1}, []int32{1, 0}, pudlak)
			if err != nil {
				t.Fatal(err)
			}
			if g.InputCount() != 2 {
				t.Fatalf("inputs=%d, want 2", g.InputCount())
			}
			for mask := 0; mask < 4; mask++ {
				input0, input1 := mask&1 != 0, mask&2 != 0
				out, err := g.Evaluate([]bool{input0, input1})
				want := input1 != negated
				if err != nil || len(out) != 1 || out[0] != want {
					t.Fatalf("negated=%v pudlak=%v input=%d output=%v err=%v", negated, pudlak, mask, out, err)
				}
			}
		}
	}
}
