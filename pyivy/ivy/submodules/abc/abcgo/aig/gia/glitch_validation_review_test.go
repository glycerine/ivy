package gia

import (
	"math"
	"os/exec"
	"strings"
	"testing"
)

func TestGlitchRejectsNaNTransitionProbability(t *testing.T) {
	exe := compileGiaSourceOracle(t, "glitch_review.c")
	g := glitchReviewGraph{cis: 1, nodes: []glitchReviewNode{{[]int32{0}, 0xaaaaaaaaaaaaaaaa}}, outputs: []int32{1}}
	file := g.write(t, t.TempDir())
	for _, tc := range []struct {
		name, mode string
		set        func(*Gli_Man_t, float32)
	}{{"combinational", "0", Gli_ManSetPiRandom}, {"sequential", "1", Gli_ManSetPiRandomSeq}} {
		t.Run(tc.name, func(t *testing.T) {
			source, err := exec.Command(exe, file, tc.mode, "1", "0", "NaN", "0").CombinedOutput()
			if err == nil || !strings.Contains(string(source), "PiTransProb < 1.0") {
				t.Fatalf("original probability precondition changed: %v %s", err, source)
			}
			p := g.build()
			defer Gli_ManStop(p)
			cexReviewMustPanic(t, func() { tc.set(p, float32(math.NaN())) })
		})
	}
}

func TestGlitchValidationAndLifetime(t *testing.T) {
	for _, prob := range []float32{-1, 0, 1, 2, float32(math.Inf(-1)), float32(math.Inf(1))} {
		for _, set := range []func(*Gli_Man_t, float32){Gli_ManSetPiRandom, Gli_ManSetPiRandomSeq} {
			p := glitchReviewBuffer()
			cexReviewMustPanic(t, func() { set(p, prob) })
			Gli_ManStop(p)
		}
	}
	for _, args := range [][3]int32{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}} {
		cexReviewMustPanic(t, func() { Gli_ManAlloc(args[0], args[1], args[2]) })
	}
	p := glitchReviewBuffer()
	defer Gli_ManStop(p)
	for _, args := range [][2]int32{{-1, 0}, {0, -1}} {
		cexReviewMustPanic(t, func() { Gli_ObjAlloc(p, args[0], args[1]) })
	}
	// Every kind of indexed port rejects both negative and out-of-range indices.
	for _, get := range []func(*Gli_Man_t, int32) *Gli_Obj_t{Gli_ManPi, Gli_ManPo, Gli_ManRo, Gli_ManRi} {
		for _, id := range []int32{-1, 5} {
			cexReviewMustPanic(t, func() { get(p, id) })
		}
	}
	for _, set := range []func(*Gli_Man_t, int32){Gli_ManSetPiFromSaved, Gli_ManSetDataSaved} {
		cexReviewMustPanic(t, func() { set(p, 0) })
		p.pSimInfoPrev = make([]uint32, Gli_ManCiNum(p))
		for _, bit := range []int32{-1, 32} {
			cexReviewMustPanic(t, func() { set(p, bit) })
		}
		p.pSimInfoPrev = nil
	}
	Gli_ManFinalize(p)
	cexReviewMustPanic(t, func() { Gli_ObjAddFanin(p.objects[1], p.objects[0]) })
	p.objects[1].fPhase2 ^= 1
	cexReviewMustPanic(t, func() { Gli_ManVerify(p) })
	p.objects[1].fPhase2 ^= 1
	p.objects[1].nSwitches = 1
	cexReviewMustPanic(t, func() { Gli_ManVerify(p) })
	incomplete := Gli_ManAlloc(1, 0, 1)
	Gli_ManCreateCi(incomplete, 1)
	cexReviewMustPanic(t, func() { Gli_ManFinalize(incomplete) })
	Gli_ManStop(incomplete)
	wide := Vec_IntAlloc(17)
	for i := 0; i < 17; i++ {
		Vec_IntPush(wide, 0)
	}
	cexReviewMustPanic(t, func() { Gli_ManCreateNode(p, wide, 0, []word{0}) })
	Gli_ManStop(nil)
	Gli_ManStop(p)
	if p.objects != nil || p.vCis != nil || p.vCos != nil || p.pSimInfoPrev != nil || Gli_ManObjNum(p) != 0 {
		t.Fatal("stop retains manager graph or snapshots")
	}
}

func TestGlitchBorrowedTruthAndIndependentManagers(t *testing.T) {
	truth := []word{0xaaaaaaaaaaaaaaaa}
	build := func() *Gli_Man_t {
		p := Gli_ManAlloc(2, 0, 1)
		input := Gli_ManCreateCi(p, 1)
		pins := Vec_IntAlloc(1)
		Vec_IntPush(pins, input)
		Gli_ManCreateNode(p, pins, 0, truth)
		Gli_ManFinalize(p)
		return p
	}
	p, q := build(), build()
	defer Gli_ManStop(q)
	truth[0] = 0x5555555555555555
	if Gli_NodeComputeValue(p.objects[1]) != 1 || Gli_NodeComputeValue(q.objects[1]) != 1 {
		t.Fatal("source borrowed truth was copied or lost")
	}
	p.objects[0].fPhase = 1
	if Gli_NodeComputeValue(p.objects[1]) != 0 || Gli_NodeComputeValue(q.objects[1]) != 1 {
		t.Fatal("graphs share mutable state")
	}
	Gli_ManStop(p)
	truth[0] = 0xaaaaaaaaaaaaaaaa
	if Gli_NodeComputeValue(q.objects[1]) != 0 {
		t.Fatal("stopping another manager damaged borrowed truth")
	}
}

func TestGlitchSourceRegisterAccessorOffsetDefect(t *testing.T) {
	exe := compileGiaSourceOracle(t, "glitch_review.c")
	g := glitchReviewGraph{cis: 4, regs: 1, nodes: []glitchReviewNode{{[]int32{0}, 0xaaaaaaaaaaaaaaaa}}, outputs: []int32{0, 1, 4}}
	source, err := exec.Command(exe, g.write(t, t.TempDir()), "6", "0", "0", "0.5", "0").CombinedOutput()
	if err != nil {
		t.Fatalf("source accessor: %v %s", err, source)
	}
	if string(source) != "ports ro1/3 ri6/7\n" {
		t.Fatalf("source offset defect changed: %s", source)
	}
	p := g.build()
	defer Gli_ManStop(p)
	if Gli_ManRo(p, 0) != p.objects[3] || Gli_ManRi(p, 0) != p.objects[7] {
		t.Fatal("Go should retain correct PI/PO offsets used by source simulation macros")
	}
}

func TestGlitchGraphAccessorsAndGrowingHints(t *testing.T) {
	// Go's allocation sizes are hints: growing graphs remain safe, unlike C's
	// fixed packed buffer. Exercise classification and indexed link access too.
	p := Gli_ManAlloc(0, 1, 0)
	defer Gli_ManStop(p)
	a, b := Gli_ManCreateCi(p, 1), Gli_ManCreateCi(p, 1)
	pins := Vec_IntAlloc(2)
	Vec_IntPush(pins, a)
	Vec_IntPush(pins, b)
	n := Gli_ManCreateNode(p, pins, 2, []word{8})
	co0, co1 := Gli_ManCreateCo(p, n), Gli_ManCreateCo(p, n)
	Gli_ManFinalize(p)
	if Gli_ManObjNum(p) != 5 || Gli_ManNodeNum(p) != 1 || Gli_ManPiNum(p) != 1 || Gli_ManPoNum(p) != 1 || Gli_ManRegNum(p) != 1 {
		t.Fatal("manager counts disagree")
	}
	if Gli_ManCi(p, 0) != p.objects[a] || Gli_ManCo(p, 1) != p.objects[co1] || Gli_ManRo(p, 0) != p.objects[b] || Gli_ManRi(p, 0) != p.objects[co1] {
		t.Fatal("port accessor mismatch")
	}
	for i, o := range p.objects {
		isCi, isNode, isCo := i < 2, i == 2, i > 2
		if (Gli_ObjIsTerm(o) != 0) != (isCi || isCo) || (Gli_ObjIsCi(o) != 0) != isCi || (Gli_ObjIsCo(o) != 0) != isCo || (Gli_ObjIsNode(o) != 0) != isNode {
			t.Fatalf("object%d classification mismatch", i)
		}
		if Gli_ObjFaninNum(o) != int32(len(o.fanins)) || Gli_ObjFanoutNum(o) != int32(len(o.fanouts)) {
			t.Fatal("link count mismatch")
		}
	}
	if Gli_ObjFanin(p.objects[n], 0) != p.objects[a] || Gli_ObjFanin(p.objects[n], 1) != p.objects[b] || Gli_ObjFanout(p.objects[n], 0) != p.objects[co0] || Gli_ObjFanout(p.objects[n], 1) != p.objects[co1] {
		t.Fatal("link ordering mismatch")
	}
}
