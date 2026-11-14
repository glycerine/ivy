package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestPatternPackPartialAssignments(t *testing.T) {
	values, care := vec.Vec_WrdStart(2), vec.Vec_WrdStart(2)
	if Gia_ManSimBitPackOne(1, values, care, 0, []int32{0}) != 1 {
		t.Fatal("first pattern not allocated")
	}
	if Gia_ManSimBitPackOne(1, values, care, 1, []int32{2}) != 0 {
		t.Fatal("compatible pattern not merged")
	}
	if Gia_ManSimBitPackOne(1, values, care, 1, []int32{1}) != 1 {
		t.Fatal("conflict not separated")
	}
	if vec.Vec_WrdEntry(values, 0)&3 != 1 || vec.Vec_WrdEntry(values, 1)&1 != 1 {
		t.Fatal("incorrect assignments")
	}
}
func TestPatternPackFullWordBoundary(t *testing.T) {
	p := Gia_ManStart(12)
	for i := 0; i < 6; i++ {
		Gia_ManAppendCi(p)
	}
	p.VSimsPi = vec.Vec_WrdStart(0)
	store := vec.Vec_IntAlloc(64 * 8)
	for value := 0; value < 64; value++ {
		vec.Vec_IntPush(store, 0)
		vec.Vec_IntPush(store, 6)
		for i := 0; i < 6; i++ {
			lit := int32(2 * i)
			if value&(1<<i) == 0 {
				lit++
			}
			vec.Vec_IntPush(store, lit)
		}
	}
	result := Gia_ManSimBitPacking(p, store, 64, 0)
	if vec.Vec_WrdSize(result) != 12 {
		t.Fatalf("expected two words per input, got %d", vec.Vec_WrdSize(result))
	}
	data := vec.Vec_WrdArray(result)
	for bit := 0; bit < 64; bit++ {
		for i := 0; i < 6; i++ {
			if (data[2*i]>>bit)&1 != uint64((bit>>i)&1) {
				t.Fatalf("pattern %d variable %d", bit, i)
			}
		}
	}
}
