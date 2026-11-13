package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSequentialCounterexamplePatternsPreserveFrames(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b, r := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, r)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManSetRegNum(g, 1)
	_ = b
	cex := util.Abc_CexAlloc(1, 2, 2)
	cex.IFrame = 1
	cex.PData[0] = 2 | 16
	info := vec.Vec_PtrAllocSimInfo(7, 2)
	Cec_ManSeqDeriveInfoFromCex(info, g, cex)
	if got := cecInfo(info, 0); got[0] != 0 || got[1] != 0 {
		t.Fatalf("initial state %v", got)
	}
	for i, want := range []uint32{1, 0, 0, 1} {
		words := cecInfo(info, int32(i+1))
		if words[0]&1 != 0 || words[0]>>1&1 != want {
			t.Fatalf("frame input %d = %#x, want witness bit %d and zero pattern", i, words[0], want)
		}
	}
	cex.PData[0] |= 1
	Cec_ManSeqDeriveInfoInitRandom(info, g, cex)
	if got := cecInfo(info, 0); got[0] != ^uint32(0) || got[1] != ^uint32(0) {
		t.Fatalf("initial state not replicated: %v", got)
	}
	cex.PData[0] &^= 1
	Cec_ManSeqDeriveInfoInitRandom(info, g, cex)
	if got := cecInfo(info, 0); got[0] != 0 || got[1] != 0 {
		t.Fatalf("zero initial state not replicated: %v", got)
	}
}

func TestSequentialResimulationPropagatesRegisterAcrossFrames(t *testing.T) {
	for frames := 1; frames <= 2; frames++ {
		g := gia.Gia_ManStart(8)
		pi, ro := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, ro)
		gia.Gia_ManAppendCo(g, pi)
		gia.Gia_ManSetRegNum(g, 1)
		g.StartRepresentatives()
		info := vec.Vec_PtrAllocSimInfo(1+frames, 1)
		cecInfo(info, 1)[0] = 2 // The second packed pattern sets the first-frame input.
		got := Cec_ManSeqResimulateInfo(g, info, nil, 1)
		want := int32(0)
		if frames == 2 {
			want = 1
		}
		if got != want {
			t.Fatalf("frames %d: status %d, want %d", frames, got, want)
		}
	}
}

func TestSequentialResimulationWithoutPrimaryInputs(t *testing.T) {
	g := gia.Gia_ManStart(8)
	ro := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManSetRegNum(g, 1)
	g.StartRepresentatives()
	info := vec.Vec_PtrAllocSimInfo(1, 1)
	cecInfo(info, 0)[0] = 2
	if got := Cec_ManSeqResimulateInfo(g, info, nil, 1); got != 1 {
		t.Fatalf("zero-input miter status %d", got)
	}
}

func TestSequentialResimulationDetectsConstantOneMiter(t *testing.T) {
	g := gia.Gia_ManStart(8)
	ro := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, ro^1)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManSetRegNum(g, 1)
	g.StartRepresentatives()
	info := vec.Vec_PtrAllocSimInfo(1, 1)
	if got := Cec_ManSeqResimulateInfo(g, info, nil, 1); got != 1 {
		t.Fatalf("constant-one miter status %d", got)
	}
}
