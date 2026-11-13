package cec

import "testing"

func TestCec4SimulationHashMatchesCWordOrder(t *testing.T) {
	primes := [16]uint32{1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147}
	for seed := uint64(0); seed < 1000; seed++ {
		words := []uint64{seed * 0x9e3779b97f4a7c15, ^seed, seed<<33 | seed}
		var h uint32
		for i, w := range words {
			for j := 0; j < 2; j++ {
				x := uint32(w >> uint(32*j))
				if words[0]&1 != 0 {
					x = ^x
				}
				h ^= x * primes[(2*i+j)&15]
			}
		}
		want := int32(h % 127)
		if got := Cec4_ManSimHashKey(words, int32(len(words)), 127); got != want {
			t.Fatalf("seed %d: got %d want %d", seed, got, want)
		}
	}
}
