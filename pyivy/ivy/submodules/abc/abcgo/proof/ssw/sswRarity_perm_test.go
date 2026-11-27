package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// TestRarRandomPermFlopFollowsGlibcRand: Ssw_RarRandomPermFlop seeds
// srand(1) and swaps with rand() % size, so the permutation is glibc's.
func TestRarRandomPermFlopFollowsGlibcRand(t *testing.T) {
	const flops, unused = 9, 3
	util.C_Srand(1)
	want := make([]int32, flops)
	for i := range want {
		want[i] = int32(i)
	}
	for i := 0; i < flops; i++ {
		k := int(util.C_Rand()) % flops
		want[i], want[k] = want[k], want[i]
	}
	for i := 0; i < unused; i++ {
		k := int(util.C_Rand()) % len(want)
		want = append(want, -1)
		want[len(want)-1], want[k] = want[k], want[len(want)-1]
	}
	got := vec.Vec_IntArray(Ssw_RarRandomPermFlop(flops, unused))
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("permutation %v, want %v", got, want)
		}
	}
}
