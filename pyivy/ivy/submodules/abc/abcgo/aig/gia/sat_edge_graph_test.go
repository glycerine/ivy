package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSatEdgeGraph(t *testing.T) {
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
	input := vec.Vec_IntAlloc(2)
	vec.Vec_IntPushTwo(input, x>>1, y>>1)
	Gia_ManEdgeFromArray(p, input)
	polars, skip := vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
	edges := Seg_ManCountIntEdges(p, polars, skip, 1)
	e := vec.Vec_IntArray(edges)
	if len(e) != 2 || e[0] != x>>1 || e[1] != y>>1 || vec.Vec_IntSize(polars) != 1 || vec.Vec_IntSize(skip) != 1 {
		t.Fatalf("edges %v polar %v skip %v", e, vec.Vec_IntArray(polars), vec.Vec_IntArray(skip))
	}
	adj := Seg_ManCollectObjEdges(edges, p.nObjs)
	if vec.Vec_IntSize(vec.Vec_WecEntry(adj, int(x>>1))) != 1 || vec.Vec_IntEntry(vec.Vec_WecEntry(adj, int(y>>1)), 0) != 0 {
		t.Fatal("incorrect incident edges")
	}
}

func TestSatEdgeManager(t *testing.T) {
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
	man := Seg_ManAlloc(p, 0)
	if man.nVars != 1 || man.DelayMax != 2 || vec.Vec_IntEntry(man.vNvars, int(y>>1)) != 2 || man.pSat.NumVars() != 3 {
		t.Fatalf("invalid manager: vars %d delay %d total %d", man.nVars, man.DelayMax, man.pSat.NumVars())
	}
	Seg_ManCreateCnf(man, 0, 0)
	if man.pSat.Solve(bsat.Positive(0), bsat.Negative(1)) != bsat.Unsatisfiable {
		t.Fatal("selected edge must imply first delay level")
	}
	if man.pSat.Solve(bsat.Negative(0), bsat.Negative(2)) != bsat.Unsatisfiable {
		t.Fatal("unselected edge must imply second delay level")
	}
	if man.pSat.Solve(bsat.Positive(0), bsat.Positive(1)) != bsat.Satisfiable {
		t.Fatal("valid edge choice rejected")
	}
	Seg_ManClean(man)
	if man.pSat.NumVars() != 1 || vec.Vec_IntSize(man.vEdges) != 0 {
		t.Fatal("manager reset failed")
	}
	Seg_ManStop(man)
}

func TestSatEdgeComputeDelay(t *testing.T) {
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
	Seg_ManComputeDelay(p, 0, 0, 0, 0)
	if p.VEdge1 == nil || Gia_ObjCheckEdge(p, x>>1, y>>1) == 0 {
		t.Fatal("expected the edge that minimizes delay")
	}
}
