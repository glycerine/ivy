package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestStructuralXorTree(t *testing.T) {
	p := gia.Gia_ManStart(64)
	gia.Gia_ManHashStart(p)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManHashXor(p, a, b)
	y := gia.Gia_ManHashXor(p, x, c)
	roots := Acec_StructDetectXorRoots(p)
	if vec.Vec_IntSize(roots) != 1 || vec.Vec_IntEntry(roots, 0) != y>>1 {
		t.Fatalf("roots: %v", roots)
	}
	ranks := Acec_StructAssignRanks(p, roots)
	for _, lit := range []int32{a, b, c, x, y} {
		if got := vec.Vec_IntEntry(ranks, int(lit>>1)); got != 0 {
			t.Errorf("rank %d = %d", lit>>1, got)
		}
	}
	leaves := Acec_FindTreeLeaves(p, roots, ranks)
	level := vec.Vec_WecEntry(leaves, 0)
	if vec.Vec_IntSize(level) != 3 {
		t.Fatalf("leaf count = %d", vec.Vec_IntSize(level))
	}
	got := map[int32]bool{}
	for i := 0; i < vec.Vec_IntSize(level); i++ {
		got[vec.Vec_IntEntry(level, i)] = true
	}
	for _, lit := range []int32{a, b, c} {
		if !got[lit>>1] {
			t.Errorf("missing leaf %d", lit>>1)
		}
	}
	shadows := Acec_FindShadows(p, ranks)
	for _, lit := range []int32{a, b, c} {
		if vec.Vec_IntEntry(shadows, int(lit>>1)) != -1 {
			t.Error("CI shadow must be unset")
		}
	}
}

func TestStructuralXorSharedInputs(t *testing.T) {
	p := gia.Gia_ManStart(64)
	gia.Gia_ManHashStart(p)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManHashXor(p, a, b)
	y := gia.Gia_ManHashXor(p, a, c)
	roots := Acec_StructDetectXorRoots(p)
	if vec.Vec_IntSize(roots) != 2 {
		t.Fatalf("roots = %d", vec.Vec_IntSize(roots))
	}
	ranks := Acec_StructAssignRanks(p, roots)
	if got := vec.Vec_IntEntry(ranks, int(a>>1)); got != -1 {
		t.Errorf("shared input rank = %d", got)
	}
	if vec.Vec_IntEntry(ranks, int(x>>1)) == vec.Vec_IntEntry(ranks, int(y>>1)) {
		t.Fatal("distinct roots assigned same rank")
	}
	leaves := Acec_FindTreeLeaves(p, roots, ranks)
	for i := 0; i < 2; i++ {
		level := vec.Vec_WecEntry(leaves, i)
		if vec.Vec_IntSize(level) != 2 {
			t.Fatalf("tree %d leaf count = %d", i, vec.Vec_IntSize(level))
		}
	}
}
