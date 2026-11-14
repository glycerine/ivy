package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGiaRetimeForwardClasses(t *testing.T) {
	g := Gia_ManStart(12)
	pi := Gia_ManAppendCi(g)
	r0, r1 := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, r0, r1)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, pi)
	Gia_ManAppendCo(g, pi)
	Gia_ManSetRegNum(g, 2)
	g.VFlopClasses = vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(g.VFlopClasses, 7)
	vec.Vec_IntPush(g.VFlopClasses, 7)
	var fixed, moves int32
	next := Gia_ManRetimeForwardOne(g, &fixed, &moves)
	if moves != 1 || Gia_ManRegNum(next) != 1 || vec.Vec_IntEntry(next.VFlopClasses, 0) != 7 {
		t.Fatalf("moves %d regs %d classes %v", moves, Gia_ManRegNum(next), vec.Vec_IntArray(next.VFlopClasses))
	}
}

func TestGiaRetimeKeepsDifferentClasses(t *testing.T) {
	g := Gia_ManStart(12)
	pi := Gia_ManAppendCi(g)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, pi)
	Gia_ManAppendCo(g, pi)
	Gia_ManSetRegNum(g, 2)
	g.VFlopClasses = vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(g.VFlopClasses, 1)
	vec.Vec_IntPush(g.VFlopClasses, 2)
	var fixed, moves int32
	next := Gia_ManRetimeForwardOne(g, &fixed, &moves)
	if moves != 0 || Gia_ManRegNum(next) != 2 || vec.Vec_IntSize(next.VFlopClasses) != 2 {
		t.Fatalf("moves %d regs %d", moves, Gia_ManRegNum(next))
	}
}
