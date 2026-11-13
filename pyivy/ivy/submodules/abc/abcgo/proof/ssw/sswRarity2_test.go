package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRarity2CountsAndSelectsRareLatchPatterns(t *testing.T) {
	g := aig.NewGraph()
	x, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, a))
	li0 := g.AddOutput(a)
	g.AddOutput(b)
	aig.Aig_ManSetRegNum(g, 2)
	p := sswRar2Start(g, 1, 1, 1, false)
	p.inits = vec.Vec_IntAlloc(2)
	Ssw_SmlSimInfo(p.sim, li0)[0] = 1
	p.updateCounters()
	if p.rarity[0] != 31 || p.rarity[1] != 1 || p.rarity[2] != 32 {
		t.Fatalf("incorrect group counts: %v", p.rarity)
	}
	p.transferPatterns()
	got := vec.Vec_IntArray(p.inits)
	if len(got) != 2 || got[0] != 1 || got[1] != 0 {
		t.Fatalf("rare pattern not selected: %v", got)
	}
}
func TestRarity2StartingStateReplaysPriorFrames(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	util.Abc_InfoSetBit(c.PData, 1)
	got := sswRar2StartingState(g, c)
	if vec.Vec_IntSize(got) != 1 || vec.Vec_IntEntry(got, 0) != 1 {
		t.Fatalf("incorrect final-frame state %v", vec.Vec_IntArray(got))
	}
}
func TestRarity2SimulationFindsReachableFailure(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, state))
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	if got := Ssw_RarSimulate2(g, 2, 1, 1, 1, 0, 0); got != 0 {
		t.Fatalf("expected reachable failure, got %d", got)
	}
}
func TestRarity2FilterKeepsReachableConstant(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, state))
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	if got := Ssw_RarSignalFilter2(g, 2, 1, 1, 2, 0, nil, 1, 0); got != -1 {
		t.Fatalf("filter should return unknown, got %d", got)
	}
	if repr := aig.Aig_ObjRepr(g, state); repr != g.Const1() {
		t.Fatalf("constant latch class missing: %v", repr)
	}
}

func TestRarity2GiaRoundTripClasses(t *testing.T) {
	g := gia.Gia_ManStart(16)
	x, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, x, state)
	gia.Gia_ManAppendCo(g, gate)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	if got := Ssw_RarSignalFilterGia2(g, 2, 1, 1, 1, 0, nil, 1, 0); got != -1 {
		t.Fatalf("got %d", got)
	}
	if !g.HasRepresentatives() {
		t.Fatal("GIA representative mapping not restored")
	}
}
