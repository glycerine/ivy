package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSpeedupChoicePreservesFunction(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	dst := Gia_ManDup(p)
	Gia_ManHashStart(dst)
	dst.pReprs = make([]Gia_Rpr_t, 128)
	dst.pNexts = make([]int32, 128)
	for id := range dst.pReprs {
		dst.pReprs[id].iRepr = GIA_VOID
	}
	leaves := vec.Vec_IntAlloc(3)
	for _, lit := range []int32{a, b, c} {
		vec.Vec_IntPush(leaves, lit>>1)
	}
	times := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(times, a>>1)
	Gia_ManSpeedupObj(dst, p, Gia_ManObj(p, y>>1), leaves, times)
	Gia_ManHashStop(dst)
	choice := Gia_ObjNext(dst, y>>1)
	if choice == 0 {
		t.Fatal("expected a generated choice")
	}
	for mask := 0; mask < 8; mask++ {
		vals, _, err := dst.evaluateObjects([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
		if err != nil || vals[choice] != vals[y>>1] {
			t.Fatalf("mask %d choice differs", mask)
		}
	}
}
