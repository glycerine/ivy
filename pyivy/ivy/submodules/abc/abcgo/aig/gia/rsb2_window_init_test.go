package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRsb2ReconvergentWindow(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, a, c)
	root := Gia_ManAppendAnd(p, x, y)
	Gia_ManAppendCo(p, root)
	paths := vec.Vec_IntStart(int(p.nObjs))
	win := Gia_RsbWindowInit(p, paths, root>>1, 8)
	if win == nil {
		t.Fatal("missed reconvergence")
	}
	inputs := Gia_RsbCreateWindowInputs(p, win)
	if vec.Vec_IntSize(inputs) != 3 {
		t.Fatalf("inputs %v", vec.Vec_IntArray(inputs))
	}
	p2 := Gia_ManStart(8)
	u, v := Gia_ManAppendCi(p2), Gia_ManAppendCi(p2)
	z := Gia_ManAppendAnd(p2, u, v)
	Gia_ManAppendCo(p2, z)
	if Gia_RsbWindowInit(p2, vec.Vec_IntStart(int(p2.nObjs)), z>>1, 8) != nil {
		t.Fatal("unexpected reconvergence")
	}
}

func TestRsb2ComputeWindow(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, a, c)
	root := Gia_ManAppendAnd(p, x, y)
	Gia_ManAppendCo(p, root)
	levels := vec.Vec_WecStart(int(Gia_ManLevelNum(p) + 1))
	paths := vec.Vec_IntStart(int(p.nObjs))
	Gia_ManStaticFanoutStart(p)
	defer Gia_ManStaticFanoutStop(p)
	var win, ins *Vec_Int_t
	if Gia_RsbWindowCompute(p, root>>1, 3, 8, levels, paths, &win, &ins) != 1 || vec.Vec_IntSize(ins) != 3 {
		t.Fatal("failed to compute window")
	}
	refs := vec.Vec_IntStart(int(p.nObjs))
	outs := Gia_RsbFindOutputs(p, win, ins, refs)
	q := Gia_RsbDeriveGiaFromWindows(p, win, ins, outs)
	if Gia_ManVerifyTwoTruths(p, q) != 1 {
		t.Fatal("computed window changed function")
	}
}

func TestRsb2GrowAlternativeWindow(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	root := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, root)
	Gia_ManStaticFanoutStart(p)
	defer Gia_ManStaticFanoutStop(p)
	levels := vec.Vec_WecStart(int(Gia_ManLevelNum(p) + 1))
	inputs := vec.Vec_IntAlloc(3)
	vec.Vec_IntPush(inputs, x>>1)
	vec.Vec_IntPush(inputs, c>>1)
	window := vec.Vec_IntAlloc(8)
	Gia_WinCreateFromCut(p, root>>1, inputs, levels, window)
	Gia_RsbWindowGrow(p, levels, window, inputs, 3)
	if vec.Vec_IntSize(inputs) != 3 {
		t.Fatalf("inputs %v", vec.Vec_IntArray(inputs))
	}
	for _, id := range vec.Vec_IntArray(inputs) {
		if Gia_ObjIsAnd(Gia_ManObj(p, id)) != 0 {
			t.Fatalf("internal node %d remains in cut", id)
		}
	}
}

func TestRsb2EnumerateWindows(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, a, c)
	root := Gia_ManAppendAnd(p, x, y)
	Gia_ManAppendCo(p, root)
	Gia_RsbEnumerateWindows(p, 3, 8)
}
