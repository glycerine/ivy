package cec

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestSpecRefineWithoutCandidates(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a)
	calls := 0
	run := func(*aig.Graph, int32, int32, int32, int32) *util.Abc_Cex_t { calls++; return nil }
	if Gia_CommandSpecI(p, 5, 100, 0, 0, 0, run) != 0 {
		t.Fatal("missing classes accepted")
	}
	p.StartRepresentatives()
	if Gia_CommandSpecI(p, 5, 100, 0, 0, 0, run) != 1 || calls != 0 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestSpecRefineInvokesBMC(t *testing.T) {
	p := gia.Gia_ManStart(10)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManAppendAnd(p, a, b)
	y := gia.Gia_ManAppendAnd(p, x, c)
	gia.Gia_ManAppendCo(p, y)
	p.StartRepresentatives()
	gia.Gia_ObjSetRepr(p, x>>1, 0)
	calls := 0
	run := func(g *aig.Graph, start, frames, conflicts, verbose int32) *util.Abc_Cex_t {
		calls++
		if g.OutputCount() == 0 || start != 0 || frames != 7 || conflicts != 42 || verbose != 1 {
			t.Fatalf("invalid callback args: %d %d %d %d", start, frames, conflicts, verbose)
		}
		return nil
	}
	if Gia_CommandSpecI(p, 7, 42, 0, 0, 1, run) != 1 || calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
}
