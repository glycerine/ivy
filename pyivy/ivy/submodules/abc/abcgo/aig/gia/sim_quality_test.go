package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSimulationPatternQuality(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	prior := vec.Vec_WrdStart(2)
	p.VSimsPi = prior
	pat := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(pat, 1)
	vec.Vec_IntPush(pat, 1)
	result := Gia_SimQualityOne(p, pat, 1)
	if vec.Vec_IntEntry(result, 0) != 2 || p.VSimsPi != prior {
		t.Fatal("incorrect output sensitivity or input restoration")
	}
	words := vec.Vec_WrdStart(2)
	Gia_ManPatUpdateOne(p, words, 7, 1, pat)
	got := Gia_ManPatCollectOne(p, words, 7, 1)
	if vec.Vec_IntEntry(got, 0) != 1 || vec.Vec_IntEntry(got, 1) != 1 {
		t.Fatal("pattern round trip failed")
	}
}
