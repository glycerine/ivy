package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestRarityRandomSeedAdvancesStream(t *testing.T) {
	for _, seed := range []int32{0, 1, 2, 19, 100} {
		aig.Aig_ManRandom(1)
		for i := int32(0); i < seed; i++ {
			aig.Aig_ManRandom(0)
		}
		want := aig.Aig_ManRandom(0)
		Ssw_RarManPrepareRandom(seed)
		if got := aig.Aig_ManRandom(0); got != want {
			t.Fatalf("seed=%d got=%d want=%d", seed, got, want)
		}
	}
}
