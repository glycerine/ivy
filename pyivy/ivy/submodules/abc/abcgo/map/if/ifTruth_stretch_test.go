package if_

import (
	"math/rand"
	"slices"
	"testing"
)

func TestTruthStretchMatchesBitProjection(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for it := 0; it < 20000; it++ {
		nTarget := 1 + r.Intn(12)
		target := make([]int32, nTarget)
		for i := range target {
			target[i] = int32(10*i + r.Intn(5))
		}
		var source []int32
		for _, id := range target {
			if r.Intn(3) != 0 {
				source = append(source, id)
			}
		}
		if len(source) == 0 {
			source = append(source, target[r.Intn(nTarget)])
			slices.Sort(source)
		}
		table := make([]uint64, ifTruthWords(len(source)))
		for i := range table {
			table[i] = r.Uint64() // small tables are deliberately not replicated
		}
		complement := r.Intn(2) == 1
		got := ifTruthProject(table, source, target, 0, 0, complement)
		want := ifTruthProjectBits(table, source, target, 0, 0, complement)
		if !slices.Equal(got, want) {
			t.Fatalf("source %v target %v: got %x want %x", source, target, got, want)
		}
	}
}
