package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRarePatternImprovement(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	patterns := vec.Vec_WrdAlloc(2)
	vec.Vec_WrdPush(patterns, 0)
	vec.Vec_WrdPush(patterns, ^uint64(0))
	p.VSimsPi = patterns
	Gia_ManPatRareImprove(p, 2, 0)
	if p.VSimsPi != patterns || vec.Vec_WrdEntry(patterns, 0) == 0 {
		t.Fatal("failed to improve or restore patterns")
	}
}
