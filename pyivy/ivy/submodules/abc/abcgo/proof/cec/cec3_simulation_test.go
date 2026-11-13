package cec

import (
	"math/rand"
	"testing"
)

func TestCec3SimulationPrimitives(t *testing.T) {
	rng := rand.New(rand.NewSource(50291))
	for words := 1; words <= 5; words++ {
		sim := newCec3Simulation(5, words)
		for trial := 0; trial < 100; trial++ {
			for i := 0; i < words; i++ {
				sim.object(1)[i], sim.object(2)[i] = rng.Uint64(), rng.Uint64()
			}
			for flags := 0; flags < 4; flags++ {
				sim.andFanins(3, 1, 2, flags&1 != 0, flags&2 != 0)
				for bit := 0; bit < words*64; bit++ {
					x := ((sim.object(1)[bit>>6] >> uint(bit&63)) & 1) != 0
					y := ((sim.object(2)[bit>>6] >> uint(bit&63)) & 1) != 0
					got := ((sim.object(3)[bit>>6] >> uint(bit&63)) & 1) != 0
					want := (x != (flags&1 != 0)) && (y != (flags&2 != 0))
					if got != want {
						t.Fatalf("words=%d trial=%d flags=%d bit=%d", words, trial, flags, bit)
					}
				}
			}
			sim.copyDriver(4, 3, true)
			if !sim.equivalent(3, 4) {
				t.Fatal("complement signatures were not equivalent")
			}
			sim.setInputBit(4, words*64-1, sim.object(4)[words-1]>>63 == 0)
			if sim.equivalent(3, 4) {
				t.Fatal("different signatures were equivalent")
			}
		}
	}
}
