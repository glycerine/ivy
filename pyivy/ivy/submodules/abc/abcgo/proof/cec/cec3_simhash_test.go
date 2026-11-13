package cec

import (
	"math/rand"
	"testing"
)

func TestCec3SimulationHashPhaseAndWords(t *testing.T) {
	rng := rand.New(rand.NewSource(10248))
	primes := []uint32{1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147}
	for n := 1; n <= 20; n++ {
		for trial := 0; trial < 50; trial++ {
			words := make([]uint64, n)
			for i := range words {
				words[i] = rng.Uint64()
			}
			var expected uint32
			for i := 0; i < 2*n; i++ {
				half := uint32(words[i/2] >> uint(32*(i%2)))
				if words[0]&1 != 0 {
					half = ^half
				}
				expected ^= half * primes[i%16]
			}
			got := Cec3_ManSimHashKey(words, int32(n), 997)
			if got != int32(expected%997) {
				t.Fatalf("n=%d trial=%d got=%d want=%d", n, trial, got, expected%997)
			}
			for i := range words {
				words[i] = ^words[i]
			}
			if complement := Cec3_ManSimHashKey(words, int32(n), 997); complement != got {
				t.Fatal("complement changed normalized hash")
			}
		}
	}
}
