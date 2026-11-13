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

func TestCec4SimulateMuxComplements(t *testing.T) {
	for mask := int32(0); mask < 8; mask++ {
		g := gia.Gia_ManStart(10)
		g.EnableMuxes()
		c, a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		m := gia.Gia_ManAppendMuxReal(g, c^(mask&1), a^((mask>>1)&1), b^((mask>>2)&1))
		sim := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
		sim.object(c >> 1)[0] = 0xaaaaaaaaaaaaaaaa
		sim.object(a >> 1)[0] = 0xcccccccccccccccc
		sim.object(b >> 1)[0] = 0xf0f0f0f0f0f0f0f0
		cec4SimulateNode(sim, g, m>>1)
		for bit := 0; bit < 64; bit++ {
			control := (bit&1 != 0) != (mask&1 != 0)
			yes := (bit&2 != 0) != (mask&2 != 0)
			no := (bit&4 != 0) != (mask&4 != 0)
			want := control && yes || !control && no
			if got := sim.object(m >> 1)[0]>>uint(bit)&1 != 0; got != want {
				t.Fatalf("mask %d bit %d got %v want %v", mask, bit, got, want)
			}
		}
	}
}
