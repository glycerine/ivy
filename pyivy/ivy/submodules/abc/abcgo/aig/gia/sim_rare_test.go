package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRarePatternImpact(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	patterns := vec.Vec_WrdAlloc(2)
	vec.Vec_WrdPush(patterns, 1)
	vec.Vec_WrdPush(patterns, ^uint64(0))
	original := vec.Vec_WrdStart(2)
	p.VSimsPi = original
	rare := Gia_SimCollectRare(p, patterns, 2)
	if vec.Vec_IntSize(rare) != 2 || vec.Vec_IntEntry(rare, 0) != x|1 || vec.Vec_IntEntry(rare, 1) != 1 {
		t.Fatalf("wrong rare signals %v", vec.Vec_IntArray(rare))
	}
	pat := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(pat, 0)
	vec.Vec_IntPush(pat, 1)
	impact := Gia_SimQualityImpact(p, pat, rare)
	best := Gia_SimCollectBest(impact)
	if best == nil || vec.Vec_IntSize(best) != 1 || vec.Vec_IntEntry(best, 0) != 0 || vec.Vec_FltEntry(impact, 0) != 0.5 {
		t.Fatal("wrong rare-signal impact")
	}
	if Gia_ManPatGetOneQuo(p, 2, patterns, 1, 0) != 0.5 || p.VSimsPi != original {
		t.Fatal("wrong quality or input restoration")
	}
}
