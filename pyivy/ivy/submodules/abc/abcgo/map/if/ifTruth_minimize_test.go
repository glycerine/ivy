package if_

import (
	"math/rand"
	"slices"
	"testing"
)

// ifTruthMinimize (word-level, as Abc_TtMinBase) agrees with the bit-level
// reference on random functions of random supports.
func TestTruthMinimizeMatchesBitLevel(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for iter := 0; iter < 3000; iter++ {
		n := 1 + rng.Intn(8)
		words := ifTruthWords(n)
		// a random function of a random subset of the n variables
		var supp []int
		for v := 0; v < n; v++ {
			if rng.Intn(2) == 0 {
				supp = append(supp, v)
			}
		}
		table := make([]uint64, words)
		inner := uint64(rng.Int63())<<1 | uint64(rng.Intn(2))
		for a := 0; a < 1<<uint(n); a++ {
			idx := 0
			for i, v := range supp {
				idx |= (a >> uint(v) & 1) << uint(i)
			}
			if inner>>uint(idx%64)&1 != 0 {
				ifTruthSet(table, a)
			}
		}
		if n < 6 { // replicate the meaningful bits over the word
			for shift := uint(1) << uint(n); shift < 64; shift <<= 1 {
				table[0] |= table[0] << shift
			}
		}
		leaves := make([]int32, n)
		for i := range leaves {
			leaves[i] = int32(10 + 3*i)
		}
		wantT, wantL := ifTruthMinimizeBits(table, leaves)
		gotT, gotL := ifTruthMinimize(table, leaves)
		if !slices.Equal(gotL, wantL) || !slices.Equal(gotT, wantT) {
			t.Fatalf("n=%d supp=%v: got %x %v, want %x %v", n, supp, gotT, gotL, wantT, wantL)
		}
	}
}
