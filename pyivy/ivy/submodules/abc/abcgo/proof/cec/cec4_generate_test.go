package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4GeneratePatternsDisprovesCandidate(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, c)
	gia.Gia_ManAppendCo(g, y)
	gia.Gia_ManSetPhase(g)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, y>>1, x>>1)
	gia.Gia_ObjSetNext(g, x>>1, y>>1)
	gia.Gia_ObjSetNext(g, y>>1, -1)
	pack := newCec4PatternPacker(int(gia.Gia_ManObjNum(g)), 1)
	count, _ := cec4GeneratePatterns(g, pack, newCec4Candidates(g), 2000)
	if count == 0 {
		t.Fatal("no distinguishing pattern generated")
	}
	cec4SimulateGraph(pack.values, g)
	if pack.values.equivalent(x>>1, y>>1) {
		t.Fatal("packed patterns did not distinguish candidates")
	}
}
