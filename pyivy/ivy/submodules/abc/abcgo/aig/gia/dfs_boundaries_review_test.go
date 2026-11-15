package gia

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestDfsSupportThresholdIncludesDanglingNodes(t *testing.T) {
	p := Gia_ManStart(64)
	defer Gia_ManStop(p)
	inputs := make([]int32, 18)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(p)
	}
	root := inputs[0]
	for i := 1; i < len(inputs); i++ {
		root = Gia_ManAppendAnd(p, root, inputs[i])
		if got := Gia_ManSuppSizeOne(p, Gia_ObjFromLit(p, root)); got != int32(i+1) {
			t.Fatalf("prefix %d: support %d, want %d", i, got, i+1)
		}
	}
	Gia_ManAppendCo(p, root)
	Gia_ManAppendAnd(p, inputs[0], inputs[1]) // deliberately disconnected
	text := hashReviewStdout(t, func() {
		if got := Gia_ManSuppSizeTest(p); got != 16 {
			t.Errorf("small-support count = %d, want 16", got)
		}
	})
	if !strings.HasPrefix(text, "Nodes with small support 16 (out of 18)\n") {
		t.Fatalf("threshold report: %q", text)
	}
}

func TestDfsEmptyAndConstantLevels(t *testing.T) {
	for _, outputs := range []bool{false, true} {
		t.Run(fmt.Sprint(outputs), func(t *testing.T) {
			p := Gia_ManStart(8)
			defer Gia_ManStop(p)
			Gia_ManAppendCi(p) // unused input must not appear in reverse levels
			if outputs {
				Gia_ManAppendCo(p, 0)
				Gia_ManAppendCo(p, 1)
			}
			forward := Gia_ManLevelize(p)
			if vec.Vec_VecSize(forward) != 1 || vec.Vec_VecSizeSize(forward) != 0 {
				t.Fatal("logic-free graph must have one empty forward bucket")
			}
			if outputs {
				// Source orders COs by their drivers, ignoring CO level values.
				Gia_ObjSetLevelId(p, 2, -7)
				Gia_ObjSetLevelId(p, 3, 1000)
				assertIDs(t, Gia_ManOrderReverse(p), 2, 3, 0, 1)
			} else {
				assertIDs(t, Gia_ManOrderReverse(p), 0, 1)
			}
			reverse := Gia_ManLevelizeR(p)
			if outputs {
				if Vec_WecSize(reverse) != 2 {
					t.Fatal("constant-driven output sets reverse depth one")
				}
				assertIDs(t, Vec_WecEntry(reverse, 0), 2, 3)
				assertIDs(t, Vec_WecEntry(reverse, 1))
			} else if Vec_WecSize(reverse) != 1 || Vec_WecSizeSize(reverse) != 0 {
				t.Fatal("output-free graph must have one empty reverse bucket")
			}
			assertIDs(t, Gia_ManCollectAndsAll(p))
			assertIDs(t, Gia_ManCollectSeq(p, nil, 0), 0)
			if Gia_ManSuppSizeOne(p, Gia_ManConst0(p)) != 0 {
				t.Fatal("constant has empty support")
			}
		})
	}
}

func TestDfsTraversalGrowthAndRollover(t *testing.T) {
	p := Gia_ManStart(8)
	defer Gia_ManStop(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	if Gia_ManSuppSizeOne(p, Gia_ObjFromLit(p, x)) != 2 {
		t.Fatal("initial support")
	}
	oldSize := len(p.pTravIds)
	for p.nObjs <= int32(oldSize+5) {
		x = Gia_ManAppendAnd(p, x, b)
	}
	if Gia_ManSuppSizeOne(p, Gia_ObjFromLit(p, x)) != 2 || len(p.pTravIds) < int(p.nObjs) || p.nTravIdsAlloc != int32(len(p.pTravIds)) {
		t.Fatal("traversal storage did not grow with the graph")
	}
	for i := range p.pTravIds {
		p.pTravIds[i] = 1 // every stale slot would collide without rollover clearing
	}
	p.nTravIds = int32(^uint32(0) >> 1)
	v := Vec_IntAlloc(2)
	Gia_ManCollectCis(p, []int32{x >> 1}, 1, v)
	assertIDs(t, v, b>>1, a>>1)
	if p.nTravIds != 1 || Gia_ObjIsTravIdCurrentId(p, 0) == 0 {
		t.Fatal("rollover must clear stale marks and premark the constant")
	}
}

func TestDfsAllWalksHandleDeepCones(t *testing.T) {
	p := Gia_ManStart(22000)
	defer Gia_ManStop(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := a
	const depth = 20000
	for i := 0; i < depth; i++ {
		root = Gia_ManAppendAnd(p, root, b)
	}
	Gia_ManAppendCo(p, root)
	co := p.nObjs - 1
	work := Vec_IntAlloc(depth)
	Gia_ManCollectCis(p, []int32{co}, 1, work)
	assertIDs(t, work, b>>1, a>>1)
	if Gia_ManSuppSize(p, []int32{co, co, 0}, 3) != 2 {
		t.Fatal("deep support must count shared/constant roots once")
	}
	nodes := Gia_ManCollectNodesCis(p, []int32{co}, 1)
	if Vec_IntSize(nodes) != depth+2 {
		t.Fatal("deep nodes/CIs lost objects")
	}
	Gia_ManIncrementTravId(p)
	Gia_ManCollectAnds(p, []int32{co}, 1, work, nil)
	if Vec_IntSize(work) != depth {
		t.Fatal("deep internal collection lost nodes")
	}
	roots := Vec_IntAlloc(1)
	Vec_IntPush(roots, root>>1)
	Gia_ManCollectTfi(p, roots, work)
	if Vec_IntSize(work) != depth {
		t.Fatal("deep TFI lost nodes")
	}
	Gia_ManStaticFanoutStart(p)
	Vec_IntWriteEntry(roots, 0, 3) // first AND
	Gia_ManCollectTfo(p, roots, work)
	if Vec_IntSize(work) != depth || Vec_IntEntry(work, 0) != root>>1 || Vec_IntEntryLast(work) != 3 {
		t.Fatal("deep TFO is not reverse topological")
	}
	sequence := Gia_ManCollectSeq(p, []int32{0}, 1)
	if Vec_IntSize(sequence) != depth+4 || Vec_IntEntry(sequence, 0) != 0 || Vec_IntEntryLast(sequence) != co {
		t.Fatal("deep sequential collection lost/order-changed objects")
	}
}
