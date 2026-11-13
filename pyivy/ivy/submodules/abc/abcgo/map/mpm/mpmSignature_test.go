package mpm

import (
	"math/rand"
	"testing"
)

func TestSignatureInvarianceUnderInputPermutations(t *testing.T) {
	r := rand.New(rand.NewSource(951))
	for trial := 0; trial < 100; trial++ {
		truth := word(r.Uint64())
		var original [6]int32
		Ifd_ComputeSignature(truth, original[:])
		for v := int32(0); v < 6; v++ {
			var got [6]int32
			Ifd_ComputeSignature(Extra_Truth6ChangePhase(truth, v), got[:])
			if got != original {
				t.Fatalf("trial %d phase %d", trial, v)
			}
			if v < 5 {
				Ifd_ComputeSignature(Extra_Truth6SwapAdjacent(truth, v), got[:])
				if got != original {
					t.Fatalf("trial %d swap %d", trial, v)
				}
			}
		}
	}
}
