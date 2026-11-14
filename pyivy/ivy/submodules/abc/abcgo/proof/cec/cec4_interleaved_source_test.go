package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// Compiled ABC reports status=0 old_c=1 old_gate=1 pattern=3,4,7.
// The source replays reduced CI object IDs directly into original simulation.
func TestCec4ReplayUsesSourceReducedIDsWithInterleavedInputs(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	c := gia.Gia_ManAppendCi(g)
	y := gia.Gia_ManAppendAnd(g, a, c)
	gia.Gia_ManSetPhase(g)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, y>>1, x>>1)
	p := Cec4_ManCreate(g, &Cec_ParFra_t{JType: 0, NBTLimit: 100, NWords: 1, NItersMax: 1})
	p.pNew = Cec4_ManStartNew(g)
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, x>>1), gia.Gia_ManAppendAnd(p.pNew, 2, 4))
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, y>>1), gia.Gia_ManAppendAnd(p.pNew, 2, 6))
	Cec4_ManSimAlloc(p, 1)
	p.simState.sim.setInputBit(c>>1, 1, true)
	if got := Cec4_ManSweepNode(p, y>>1, x>>1); got != 0 {
		t.Fatalf("sweep = %d, want source SAT result", got)
	}
	if p.simState.sim.object(c >> 1)[0]&2 == 0 || p.simState.sim.object(x >> 1)[0]&2 == 0 {
		t.Fatal("source writes reduced CI ID 3 into original object 3, leaving CI 4 unchanged")
	}
}
