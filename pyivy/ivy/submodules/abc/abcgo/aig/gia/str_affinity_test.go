package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestStructuredDelayAndAffinity(t *testing.T) {
	if Str_Delay2(1, 1, 4) != 2 || Str_Delay2(3, 3, 4) != 18 || Str_Delay3(2, 2, 2, 4) != 19 {
		t.Fatal("incorrect packed delay")
	}
	var matrix [256]uint64
	matrix[2] = uint64(1) << 57
	original := matrix
	transpose64(matrix[:])
	transpose64(matrix[:])
	if matrix != original {
		t.Fatal("transpose did not round trip")
	}
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, a, c)
	z := Gia_ManAppendAnd(p, b, c)
	vec.Vec_IntFill(&p.vCopies, int(p.nObjs), 0)
	super := vec.Vec_IntAlloc(6)
	for _, lit := range []int32{x, y, z} {
		vec.Vec_IntPush(super, lit)
	}
	delays := vec.Vec_IntStart(3)
	if got := Str_ManVectorAffinity(p, super, delays, matrix[:], 12); got != 3 || vec.Vec_IntSize(super) != 3 {
		t.Fatalf("affinity count %d or changed supergate", got)
	}
}
