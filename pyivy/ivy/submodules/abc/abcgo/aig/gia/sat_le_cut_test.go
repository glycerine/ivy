package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSleCutMergeAndContainment(t *testing.T) {
	a := []int32{Sle_CutSetSizeSign(2, (1<<1)|(1<<3)), 1, 3}
	b := []int32{Sle_CutSetSizeSign(2, (1<<3)|(1<<5)), 3, 5}
	out := make([]int32, 5)
	if Sle_CutMergeOrder(a, b, out, 3) != 1 || Sle_CutSize(out) != 3 || out[1] != 1 || out[2] != 3 || out[3] != 5 {
		t.Fatalf("merge %v", out)
	}
	if Sle_CutMergeOrder(a, b, out, 2) != 0 {
		t.Fatal("oversized cut accepted")
	}
	if Sle_SetCutIsContainedOrder(out, a) != 1 || Sle_SetCutIsContainedOrder(a, b) != 0 {
		t.Fatal("bad cut containment")
	}
	set := vec.Vec_IntStart(1)
	Sle_SetAddCut(set, out)
	if Sle_SetLastCutIsContained(set, a) != 0 {
		t.Fatal("superset contains candidate")
	}
	Sle_SetAddCut(set, a)
	if Sle_CutIsUsed(vec.Vec_IntArray(set)[1:]) != 0 || Sle_SetLastCutIsContained(set, out) != 1 {
		t.Fatal("dominance filtering failed")
	}
}

func TestSleComputeCuts(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	cuts := Sle_ManComputeCuts(p, 3, 0)
	list := sleList(cuts, y>>1, false)
	found := false
	for _, cut := range list {
		leaves := Sle_CutLeaves(cut)
		if len(leaves) == 3 && leaves[0] == a>>1 && leaves[1] == b>>1 && leaves[2] == c>>1 {
			found = true
		}
	}
	if !found || Sle_ManComputeDelay(p, cuts) != 1 {
		t.Fatalf("cuts %v delay %d", list, Sle_ManComputeDelay(p, cuts))
	}
	if got := Sle_ManComputeDelay(p, Sle_ManComputeCuts(p, 2, 0)); got != 2 {
		t.Fatalf("2-LUT delay %d", got)
	}
}

func TestSleCutFanins(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	cuts := Sle_ManComputeCuts(p, 3, 0)
	mask := Sle_ManInternalNodeMask(p)
	fanins := Sle_ManCollectCutFanins(p, cuts, mask)
	got := vec.Vec_IntArray(vec.Vec_WecEntry(fanins, int(y>>1)))
	if len(got) != 1 || got[0] != x>>1 {
		t.Fatalf("cut fanins %v", got)
	}
	var piOnly, internal bool
	for _, cut := range sleList(cuts, y>>1, false) {
		if Sle_ManCutHasPisOnly(cut, mask) != 0 {
			piOnly = true
		} else {
			internal = true
		}
	}
	if !piOnly || !internal {
		t.Fatal("expected both primary-input and internal cuts")
	}
}

func TestSleManagerVariables(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	m := Sle_ManAlloc(p, 0, 0)
	Sle_ManMarkupVariables(m)
	if m.nLevels != 1 || m.nNodeVars != p.nObjs || m.nCutVars != 3 || m.nEdgeVars != 1 || m.nDelayVars != 2 || m.nVarsTotal != p.nObjs+6 {
		t.Fatalf("bad variable counts: %+v", m)
	}
	if Sle_ManCheckContained([]int32{1, 3}, 2, []int32{3, 2, 1}, 3) != 1 || Sle_ManCheckContained([]int32{1, 4}, 2, []int32{3, 2, 1}, 3) != 0 {
		t.Fatal("bad LUT containment")
	}
	Sle_ManStop(m)
}

func TestSleInitialMapping(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	p.VMapping = vec.Vec_IntStart(int(p.nObjs))
	for _, lut := range []struct {
		id int32
		in []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, c >> 1}}} {
		vec.Vec_IntWriteEntry(p.VMapping, int(lut.id), int32(vec.Vec_IntSize(p.VMapping)))
		vec.Vec_IntPush(p.VMapping, int32(len(lut.in)))
		for _, fan := range lut.in {
			vec.Vec_IntPush(p.VMapping, fan)
		}
		vec.Vec_IntPush(p.VMapping, lut.id)
	}
	edges := vec.Vec_IntAlloc(2)
	vec.Vec_IntPushTwo(edges, x>>1, y>>1)
	Gia_ManEdgeFromArray(p, edges)
	m := Sle_ManAlloc(p, 0, 0)
	Sle_ManMarkupVariables(m)
	Sle_ManDeriveInit(m)
	polar := vec.Vec_IntArray(m.vPolars)
	for _, expected := range []int32{x >> 1, y >> 1, vec.Vec_IntEntry(m.vCutFirst, int(x>>1)), vec.Vec_IntEntry(m.vEdgeFirst, int(y>>1))} {
		if vec.Vec_IntFind(m.vPolars, expected) < 0 {
			t.Fatalf("missing %d in %v", expected, polar)
		}
	}
}

func TestSleCNFCover(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	m := Sle_ManAlloc(p, 0, 0)
	Sle_ManMarkupVariables(m)
	Sle_ManDeriveCnf(m, 0, 0)
	if m.pSat.Solve() != bsat.Satisfiable {
		t.Fatal("mapping formula should be satisfiable")
	}
	if m.pSat.Solve(bsat.Negative(y>>1)) != bsat.Unsatisfiable {
		t.Fatal("output driver must be mapped")
	}
	base := vec.Vec_IntEntry(m.vCutFirst, int(y>>1))
	if m.pSat.Solve(bsat.Positive(base), bsat.Positive(base+1)) != bsat.Unsatisfiable {
		t.Fatal("two cuts cannot both be selected")
	}
}

func TestSleDeriveResult(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	m := Sle_ManAlloc(p, 0, 0)
	Sle_ManMarkupVariables(m)
	Sle_ManDeriveCnf(m, 0, 1)
	if m.pSat.Solve() != bsat.Satisfiable {
		t.Fatal("expected satisfiable mapping")
	}
	edges, mapping := vec.Vec_IntAlloc(4), vec.Vec_IntAlloc(16)
	Sle_ManDeriveResult(m, edges, mapping)
	p.VMapping = mapping
	if Gia_ObjIsLut(p, y>>1) == 0 || Gia_ObjLutSize(p, y>>1) < 2 {
		t.Fatal("output LUT not mapped")
	}
	if old := m.nEdgeClas2; Sle_ManAddEdgeConstraints(m, 1) != 0 || m.nEdgeClas2 != old {
		t.Fatal("unexpected edge violations")
	}
}

func TestSleExplore(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	Sle_ManExplore(p, 0, 0, 1, 0, 0)
	if p.VMapping == nil || Gia_ObjIsLut(p, y>>1) == 0 {
		t.Fatal("SAT mapping not produced")
	}
	leaves := Gia_ObjLutFanins(p, y>>1)
	if len(leaves) != 3 || leaves[0] != a>>1 || leaves[1] != b>>1 || leaves[2] != c>>1 {
		t.Fatalf("unexpected LUT leaves %v", leaves)
	}
}
