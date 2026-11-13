package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4SimulateXORComplements(t *testing.T) {
	for aInv := int32(0); aInv < 2; aInv++ {
		for bInv := int32(0); bInv < 2; bInv++ {
			g := gia.Gia_ManStart(8)
			a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			x := gia.Gia_ManAppendXorReal(g, a^aInv, b^bInv)
			out := gia.Gia_ManAppendCo(g, x^1)
			sim := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
			sim.object(a >> 1)[0] = 0xaaaaaaaaaaaaaaaa
			sim.object(b >> 1)[0] = 0xcccccccccccccccc
			cec4SimulateNode(sim, g, x>>1)
			cec4SimulateNode(sim, g, out>>1)
			for bit := 0; bit < 64; bit++ {
				av, bv := bit&1 != 0, bit&2 != 0
				want := (av != (aInv != 0)) != (bv != (bInv != 0))
				if got := sim.object(x >> 1)[0]>>uint(bit)&1 != 0; got != want {
					t.Fatalf("bit %d complements %d,%d: %v, want %v", bit, aInv, bInv, got, want)
				}
				if got := sim.object(out >> 1)[0]>>uint(bit)&1 != 0; got == want {
					t.Fatalf("output did not complement XOR")
				}
			}
		}
	}
}
