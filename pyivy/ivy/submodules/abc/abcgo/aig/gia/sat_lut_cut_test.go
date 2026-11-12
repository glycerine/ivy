package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSblCutFeasibleAndDominance(t *testing.T) {
	if Sbl_CutIsFeasible(15, 0, 0, 0, 4) != 1 || Sbl_CutIsFeasible(31, 0, 0, 0, 4) != 0 || Sbl_CutIsFeasible(63, 0, 0, 0, 6) != 1 || Sbl_CutIsFeasible(127, 0, 0, 0, 6) != 0 {
		t.Fatal("cut capacity")
	}
	a, b, c, d := vec.Vec_WrdAlloc(2), vec.Vec_WrdAlloc(2), vec.Vec_WrdAlloc(2), vec.Vec_WrdAlloc(2)
	if Sbl_CutPushUncontained(a, b, c, d, 7, 0, 0, 0) != 0 || Sbl_CutPushUncontained(a, b, c, d, 15, 0, 0, 0) != 1 || Sbl_CutPushUncontained(a, b, c, d, 3, 0, 0, 0) != 0 {
		t.Fatal("cut insertion")
	}
	if vec.Vec_WrdSize(a) != 1 || vec.Vec_WrdEntry(a, 0) != 3 {
		t.Fatal("dominated cut retained")
	}
}

func TestSblMergeCuts(t *testing.T) {
	left := [4][]word{{1, 2}, {0, 0}, {0, 0}, {0, 0}}
	right := [4][]word{{4, 6}, {0, 0}, {0, 0}, {0, 0}}
	result := sblMergeCuts(left, right, 4)
	got := vec.Vec_WrdArray(result[0])
	if len(got) != 2 || got[0] != 5 || got[1] != 6 {
		t.Fatalf("minimal merged cuts %v", got)
	}
	for i := 1; i < 4; i++ {
		if vec.Vec_WrdSize(result[i]) != 2 {
			t.Fatal("unaligned cut masks")
		}
	}
}

func TestSblManagerReset(t *testing.T) {
	g := Gia_ManStart(8)
	Gia_ManAppendCi(g)
	p := Sbl_ManAlloc(g, 4)
	if p.Power2 != 4 || p.FirstVar != p.pSat.NumVars() {
		t.Fatal("invalid window manager")
	}
	vec.Vec_IntPush(p.vLeaves, 1)
	vec.Vec_WrdPush(p.vCutsI1, 1)
	p.pSat.NewVariable()
	Sbl_ManClean(p)
	if vec.Vec_IntSize(p.vLeaves) != 0 || vec.Vec_WrdSize(p.vCutsI1) != 0 || p.pSat.NumVars() != p.FirstVar {
		t.Fatal("window manager reset failed")
	}
	Sbl_ManStop(p)
}

func TestSblManagerCutMerge(t *testing.T) {
	p := Sbl_ManAlloc(Gia_ManStart(4), 4)
	p.LutSize = 4
	for i := int32(0); i < 2; i++ {
		vec.Vec_IntPush(p.vCutsStart, i)
		vec.Vec_IntPush(p.vCutsNum, 1)
		vec.Vec_IntPush(p.vCutsObj, -1)
		vec.Vec_WrdPush(p.vCutsI1, word(1)<<uint(i))
		vec.Vec_WrdPush(p.vCutsI2, 0)
		vec.Vec_WrdPush(p.vCutsN1, 0)
		vec.Vec_WrdPush(p.vCutsN2, 0)
	}
	Sbl_ManComputeCutsOne(p, 0, 1, 0)
	if Sbl_ManFindCut(p, 2, 3, 0, 0, 0) < 0 || Sbl_ManFindCut(p, 2, 0, 0, 1, 0) < 0 {
		t.Fatal("missing merged or unit cut")
	}
}

func TestSblComputeWindowCuts(t *testing.T) {
	g := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	g.vMapping2 = vec.Vec_WecStart(int(g.nObjs))
	vec.Vec_WecPush(g.vMapping2, int(y>>1), a>>1)
	vec.Vec_WecPush(g.vMapping2, int(y>>1), b>>1)
	vec.Vec_WecPush(g.vMapping2, int(y>>1), c>>1)
	p := Sbl_ManAlloc(g, 4)
	p.LutSize = 4
	for _, id := range []int32{a >> 1, b >> 1, c >> 1} {
		vec.Vec_IntPush(p.vLeaves, id)
	}
	vec.Vec_IntPush(p.vAnds, x>>1)
	vec.Vec_IntPush(p.vAnds, y>>1)
	vec.Vec_IntPush(p.vRoots, y>>1)
	if n := Sbl_ManComputeCuts(p); n != 8 {
		t.Fatalf("got %d cuts", n)
	}
	if Sbl_ManFindCut(p, 4, 7, 0, 0, 0) < 0 || vec.Vec_IntSize(p.vRootVars) != 1 || vec.Vec_IntSize(p.vSolInit) != 1 || vec.Vec_IntSize(p.vPolar) != 2 {
		t.Fatal("missing window cut or initial mapping")
	}
	if Gia_ManObj(g, y>>1).Value != ^uint32(0) {
		t.Fatal("temporary object values not cleared")
	}
}
