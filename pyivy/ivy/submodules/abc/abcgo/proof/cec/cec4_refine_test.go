package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4RefineSplitsAllSimulationSignatures(t *testing.T) {
	g := gia.Gia_ManStart(12)
	for i := 0; i < 5; i++ {
		gia.Gia_ManAppendCi(g)
	}
	g.StartRepresentatives()
	for id := int32(1); id < 6; id++ {
		gia.Gia_ObjSetRepr(g, id, 0)
		gia.Gia_ObjSetNext(g, id-1, id)
	}
	gia.Gia_ObjSetNext(g, 5, -1)
	s := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
	signatures := []uint64{0, 2, 4, ^uint64(2), 6, ^uint64(4)}
	for id, v := range signatures {
		s.object(int32(id))[0] = v
	}
	cec4RefineClass(s, g, 0)
	for _, tc := range []struct{ id, repr int32 }{{1, gia.GIA_VOID}, {2, gia.GIA_VOID}, {3, 1}, {4, gia.GIA_VOID}, {5, 2}} {
		if got := gia.Gia_ObjRepr(g, tc.id); got != tc.repr {
			t.Errorf("repr(%d)=%d want %d", tc.id, got, tc.repr)
		}
	}
	if gia.Gia_ObjNext(g, 1) != 3 || gia.Gia_ObjNext(g, 2) != 5 {
		t.Fatal("class members were not relinked")
	}
}
