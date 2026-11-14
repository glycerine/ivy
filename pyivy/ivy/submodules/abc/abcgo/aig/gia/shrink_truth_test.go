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

func TestShrinkManagerDivisors(t *testing.T) {
	original := Gia_ManStart(10)
	Gia_ManAppendCi(original)
	Gia_ManAppendCi(original)
	p := Shr_ManAlloc(original)
	a, b := Gia_ManAppendCi(p.pNew), Gia_ManAppendCi(p.pNew)
	x := Gia_ManHashAnd(p.pNew, a, b)
	Gia_ObjSetAndLevel(p.pNew, Gia_ManObj(p.pNew, x>>1))
	Shr_ManAddFanout(p, a>>1, x>>1)
	Shr_ManAddFanout(p, b>>1, x>>1)
	leaves := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(leaves, a>>1)
	vec.Vec_IntPush(leaves, b>>1)
	if n := Shr_ManCollectDivisors(p, leaves, 8, 10); n != 3 || vec.Vec_IntEntry(p.vDivs, 2) != x>>1 {
		t.Fatalf("divisors %v", vec.Vec_IntArray(p.vDivs))
	}
	Gia_ManAppendCo(p.pNew, x)
	result := Shr_ManFree(p)
	if Gia_ManCoNum(result) != 1 {
		t.Fatal("shrink cleanup failed")
	}
}

func TestShrinkBidecomposition(t *testing.T) {
	g := Gia_ManStart(10)
	Gia_ManAppendCi(g)
	Gia_ManAppendCi(g)
	p := Shr_ManAlloc(g)
	a, b := Gia_ManAppendCi(p.pNew), Gia_ManAppendCi(p.pNew)
	leaves := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(leaves, a>>1)
	vec.Vec_IntPush(leaves, b>>1)
	lit := Shr_ObjPerformBidec(p, p.pManDec, p.pNew, leaves, 0x2222222222222222, ^word(0))
	truth := Shr_ManComputeTruth6(p.pNew, Gia_ManObj(p.pNew, lit>>1), leaves, p.vTruths)
	if lit&1 != 0 {
		truth = ^truth
	}
	if truth != 0x2222222222222222 {
		t.Fatalf("truth %x", truth)
	}
	Gia_ManAppendCo(p.pNew, lit)
	Shr_ManFree(p)
}

func TestShrink6MappedTruth(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c^1)
	Gia_ManAppendCo(p, y)
	p.VMapping = vec.Vec_IntStart(int(p.nObjs))
	vec.Vec_IntWriteEntry(p.VMapping, int(y>>1), int32(p.nObjs))
	vec.Vec_IntPush(p.VMapping, 3)
	for _, id := range []int32{a >> 1, b >> 1, c >> 1} {
		vec.Vec_IntPush(p.VMapping, id)
	}
	vec.Vec_IntPush(p.VMapping, y>>1)
	q := Gia_ManMapShrink6(p, 10, 0, 0)
	if q == nil || Gia_ManVerifyTwoTruths(p, q) != 1 {
		t.Fatal("six-input shrink changed function")
	}
}
