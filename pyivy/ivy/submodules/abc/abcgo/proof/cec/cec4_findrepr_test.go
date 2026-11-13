package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4FindReprSelectsEarlierLiveMatch(t *testing.T) {
	g := gia.Gia_ManStart(8)
	for i := 0; i < 4; i++ {
		gia.Gia_ManAppendCi(g)
	}
	g.StartRepresentatives()
	for id := int32(2); id <= 4; id++ {
		gia.Gia_ObjSetRepr(g, id, 1)
		gia.Gia_ObjSetNext(g, id-1, id)
	}
	gia.Gia_ObjSetNext(g, 4, -1)
	s := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
	s.object(1)[0] = 2
	s.object(2)[0] = 4
	s.object(3)[0] = ^uint64(4)
	s.object(4)[0] = 4
	if got := cec4FindRepr(s, g, 4); got != 2 {
		t.Fatalf("first matching candidate = %d", got)
	}
	gia.Gia_ObjSetFailed(g, 2)
	if got := cec4FindRepr(s, g, 4); got != 3 {
		t.Fatalf("matching candidate after failure = %d", got)
	}
	gia.Gia_ObjSetFailed(g, 3)
	if got := cec4FindRepr(s, g, 4); got != -1 {
		t.Fatalf("expected no representative, got %d", got)
	}
}
