package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
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

func TestSblWindowCNF(t *testing.T) {
	g := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	g.vMapping2 = vec.Vec_WecStart(int(g.nObjs))
	p := Sbl_ManAlloc(g, 4)
	p.LutSize = 4
	for _, id := range []int32{a >> 1, b >> 1, c >> 1} {
		vec.Vec_IntPush(p.vLeaves, id)
	}
	vec.Vec_IntPush(p.vAnds, x>>1)
	vec.Vec_IntPush(p.vAnds, y>>1)
	Sbl_ManComputeCuts(p)
	Sbl_ManCreateCnf(p)
	direct := Sbl_ManFindCut(p, 4, 7, 0, 0, 0)
	via := Sbl_ManFindCut(p, 4, 4, 0, 1, 0)
	if direct < 0 || via < 0 {
		t.Fatal("missing cuts")
	}
	if p.pSat.Solve(bsat.Positive(p.FirstVar+via), bsat.Negative(0)) != bsat.Unsatisfiable {
		t.Fatal("selected cut must require its internal fanin")
	}
	if p.pSat.Solve(bsat.Positive(p.FirstVar+direct), bsat.Negative(1)) != bsat.Unsatisfiable {
		t.Fatal("selected cut must require its root")
	}
	if p.pSat.Solve(bsat.Positive(p.FirstVar+direct), bsat.Positive(1)) != bsat.Satisfiable {
		t.Fatal("valid cut rejected")
	}
}

func TestSblTiming(t *testing.T) {
	g := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	g.vMapping2 = vec.Vec_WecStart(int(g.nObjs))
	for _, fan := range []int32{a >> 1, b >> 1} {
		vec.Vec_WecPush(g.vMapping2, int(x>>1), fan)
	}
	for _, fan := range []int32{x >> 1, c >> 1} {
		vec.Vec_WecPush(g.vMapping2, int(y>>1), fan)
	}
	p := Sbl_ManAlloc(g, 4)
	if Sbl_ManCreateTiming(p, 0) != 2 || vec.Vec_IntEntry(p.vArrs, int(x>>1)) != 1 || vec.Vec_IntEntry(p.vReqs, int(x>>1)) != 1 || Sbl_ManCriticalFanin(p, y>>1, Gia_ObjLutFanins2(g, y>>1)) != x>>1 {
		t.Fatal("incorrect arrival or required times")
	}
}

func TestSblUpdateMapping(t *testing.T) {
	g := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	g.vMapping2 = vec.Vec_WecStart(int(g.nObjs))
	for _, fan := range []int32{a >> 1, b >> 1} {
		vec.Vec_WecPush(g.vMapping2, int(x>>1), fan)
	}
	for _, fan := range []int32{x >> 1, c >> 1} {
		vec.Vec_WecPush(g.vMapping2, int(y>>1), fan)
	}
	g.pLutRefs = make([]int32, g.nObjs)
	for _, id := range []int32{a >> 1, b >> 1, x >> 1, c >> 1} {
		g.pLutRefs[id]++
	}
	p := Sbl_ManAlloc(g, 4)
	p.LutSize = 4
	for _, id := range []int32{a >> 1, b >> 1, c >> 1} {
		vec.Vec_IntPush(p.vLeaves, id)
	}
	vec.Vec_IntPush(p.vAnds, x>>1)
	vec.Vec_IntPush(p.vAnds, y>>1)
	vec.Vec_IntPush(p.vRoots, y>>1)
	Sbl_ManComputeCuts(p)
	cut := Sbl_ManFindCut(p, 4, 7, 0, 0, 0)
	vec.Vec_IntPush(p.vSolCur, cut)
	vec.Vec_IntPush(p.vSolBest, cut)
	Sbl_ManGetCurrentMapping(p)
	if vec.Vec_IntSize(vec.Vec_WecEntry(p.vWindow, 1)) != 3 {
		t.Fatal("window mapping not decoded")
	}
	Sbl_ManUpdateMapping(p)
	if Gia_ObjIsLut2(g, x>>1) != 0 || Gia_ObjLutSize2(g, y>>1) != 3 || g.pLutRefs[x>>1] != 0 || g.pLutRefs[a>>1] != 1 {
		t.Fatal("mapping or references not updated")
	}
}

func TestSblEvaluateMapping(t *testing.T) {
	g := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	g.vMapping2 = vec.Vec_WecStart(int(g.nObjs))
	for _, fan := range []int32{a >> 1, b >> 1, c >> 1} {
		vec.Vec_WecPush(g.vMapping2, int(y>>1), fan)
	}
	p := Sbl_ManAlloc(g, 4)
	p.LutSize = 4
	for _, id := range []int32{a >> 1, b >> 1, c >> 1} {
		vec.Vec_IntPush(p.vLeaves, id)
	}
	vec.Vec_IntPush(p.vAnds, x>>1)
	vec.Vec_IntPush(p.vAnds, y>>1)
	vec.Vec_IntPush(p.vRoots, y>>1)
	Sbl_ManComputeCuts(p)
	direct := Sbl_ManFindCut(p, 4, 7, 0, 0, 0)
	first := Sbl_ManFindCut(p, 3, 3, 0, 0, 0)
	via := Sbl_ManFindCut(p, 4, 4, 0, 1, 0)
	vec.Vec_IntPush(p.vSolCur, direct)
	if Sbl_ManEvaluateMapping(p, 1) != 1 {
		t.Fatal("matching delay rejected")
	}
	vec.Vec_IntClear(p.vSolCur)
	vec.Vec_IntPush(p.vSolCur, first)
	vec.Vec_IntPush(p.vSolCur, via)
	if Sbl_ManEvaluateMapping(p, 1) != 0 || vec.Vec_IntSize(p.vPath) != 2 {
		t.Fatalf("critical path %v", vec.Vec_IntArray(p.vPath))
	}
}

func TestSblWholeGraphWindow(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x^1)
	p := Sbl_ManAlloc(g, 4)
	Sbl_ManWindow(p)
	if vec.Vec_IntSize(p.vLeaves) != 2 || vec.Vec_IntSize(p.vAnds) != 1 || vec.Vec_IntSize(p.vRoots) != 1 || vec.Vec_IntEntry(p.vRoots, 0) != x>>1 {
		t.Fatal("invalid whole graph window")
	}
}

func TestSblSearchWindow(t *testing.T) {
	g := Gia_ManStart(32)
	inputs := make([]int32, 12)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(g)
	}
	var roots []int32
	current := Gia_ManAppendAnd(g, inputs[0], inputs[1])
	roots = append(roots, current)
	for i := 2; i < len(inputs); i++ {
		current = Gia_ManAppendAnd(g, current, inputs[i])
		roots = append(roots, current)
	}
	Gia_ManAppendCo(g, current)
	g.VMapping = vec.Vec_IntStart(int(g.nObjs))
	for i, lit := range roots {
		id := lit >> 1
		left := inputs[0] >> 1
		if i > 0 {
			left = roots[i-1] >> 1
		}
		right := inputs[i+1] >> 1
		vec.Vec_IntWriteEntry(g.VMapping, int(id), int32(vec.Vec_IntSize(g.VMapping)))
		vec.Vec_IntPush(g.VMapping, 2)
		vec.Vec_IntPush(g.VMapping, left)
		vec.Vec_IntPush(g.VMapping, right)
		vec.Vec_IntPush(g.VMapping, id)
	}
	Gia_ManComputeOneWinStart(g, 16, 0)
	p := Sbl_ManAlloc(g, 16)
	p.LutSize = 4
	got := Sbl_ManTestSat(p, current>>1)
	if got != 2 || p.nImproved != 1 {
		t.Fatalf("expected improved window, got %d", got)
	}
	Gia_ManComputeOneWin(g, -1, nil, nil, nil, nil)
}

func TestGiaManLutSat(t *testing.T) {
	g := Gia_ManStart(32)
	inputs := make([]int32, 12)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(g)
	}
	var roots []int32
	current := Gia_ManAppendAnd(g, inputs[0], inputs[1])
	roots = append(roots, current)
	for i := 2; i < len(inputs); i++ {
		current = Gia_ManAppendAnd(g, current, inputs[i])
		roots = append(roots, current)
	}
	Gia_ManAppendCo(g, current)
	g.VMapping = vec.Vec_IntStart(int(g.nObjs))
	for i, lit := range roots {
		id := lit >> 1
		left := inputs[0] >> 1
		if i > 0 {
			left = roots[i-1] >> 1
		}
		vec.Vec_IntWriteEntry(g.VMapping, int(id), int32(vec.Vec_IntSize(g.VMapping)))
		vec.Vec_IntPush(g.VMapping, 2)
		vec.Vec_IntPush(g.VMapping, left)
		vec.Vec_IntPush(g.VMapping, inputs[i+1]>>1)
		vec.Vec_IntPush(g.VMapping, id)
	}
	before := Gia_ManLutNum(g)
	Gia_ManLutSat(g, 4, 16, 1, 0, 0, 0, 0, 0, 0, 0)
	if g.VMapping == nil || g.vMapping2 != nil || Gia_ManLutNum(g) >= before {
		t.Fatalf("mapping did not improve: %d -> %d", before, Gia_ManLutNum(g))
	}
}
