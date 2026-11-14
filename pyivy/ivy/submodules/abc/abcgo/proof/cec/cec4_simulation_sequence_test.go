package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4RegisterOutputSimulationCopiesRegisterInput(t *testing.T) {
	g := gia.Gia_ManStart(10)
	pi, ro := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	p := Cec4_ManCreate(g, &Cec_ParFra_t{NWords: 1})
	Cec4_ManSimAlloc(p, 1)
	ri := gia.Gia_ObjId(g, gia.Gia_ManCo(g, 1))
	Cec4_ObjSim(p, ri)[0] = 0x9876543210
	Cec4_ObjSimRo(p, ro>>1)
	if got := Cec4_ObjSim(p, ro>>1)[0]; got != 0x9876543210 {
		t.Fatalf("register output simulation %#x", got)
	}
	Cec4_ManSimulateCis(p)
	for i := int32(0); i < gia.Gia_ManCiNum(g); i++ {
		id := gia.Gia_ObjId(g, gia.Gia_ManCi(g, i))
		if Cec4_ObjSim(p, id)[0]&1 != 0 {
			t.Fatal("random simulation lost zero pattern")
		}
	}
}
