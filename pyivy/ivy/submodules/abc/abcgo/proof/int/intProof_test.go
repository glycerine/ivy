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
