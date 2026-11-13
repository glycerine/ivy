package mpm

import (
	"math/rand"
	"testing"
)

func TestTruthSixSwapAndPhase(t *testing.T) {
	r := rand.New(rand.NewSource(341))
	for trial := 0; trial < 100; trial++ {
		table := word(r.Uint64())
		for v := int32(0); v < 6; v++ {
			phase := Extra_Truth6ChangePhase(table, v)
			for i := 0; i < 64; i++ {
				if (phase>>uint(i))&1 != (table>>uint(i^(1<<v)))&1 {
					t.Fatalf("phase variable %d bit %d", v, i)
				}
			}
			if v < 5 {
				swapped := Extra_Truth6SwapAdjacent(table, v)
				for i := 0; i < 64; i++ {
					j := i
					if ((i>>v)^(i>>(v+1)))&1 != 0 {
						j ^= 3 << v
					}
					if (swapped>>uint(i))&1 != (table>>uint(j))&1 {
						t.Fatalf("swap variable %d bit %d", v, i)
					}
				}
			}
		}
	}
}
