package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestShrinkDivisorTruths(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b^1)
	divs := vec.Vec_IntAlloc(3)
	for _, id := range []int32{a >> 1, b >> 1, x >> 1} {
		vec.Vec_IntPush(divs, id)
	}
	truth := vec.Vec_WrdStart(int(p.nObjs))
	out := vec.Vec_WrdAlloc(3)
	Shr_ManComputeTruths(p, 2, divs, out, truth)
	if got := vec.Vec_WrdEntry(out, 2); got != 0x2222222222222222 {
		t.Fatalf("truth %x", got)
	}
}
