package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestIncrementalSimulation(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManIncrSimStart(p, 1, 12)
	defer Gia_ManIncrSimStop(p)
	Gia_ManIncrSimUpdate(p)
	assignment := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(assignment, a)
	vec.Vec_IntPush(assignment, b^1)
	Gia_ManIncrSimSet(p, assignment)
	Gia_ManIncrSimCheckEqual(p, x, a)
	if giaSimBit(Gia_ManBuiltInData(p, x>>1), 0) != 0 {
		t.Fatal("expected a AND b to be zero")
	}
	vec.Vec_IntClear(assignment)
	vec.Vec_IntPush(assignment, a)
	vec.Vec_IntPush(assignment, b)
	Gia_ManIncrSimSet(p, assignment)
	Gia_ManIncrSimCheckOver(p, x, b)
	if giaSimBit(Gia_ManBuiltInData(p, x>>1), 1) != 1 {
		t.Fatal("updated cone not resimulated")
	}
	y := Gia_ManAppendAnd(p, x, b^1)
	Gia_ManIncrSimCheckOver(p, y, a)
	if Gia_ManBuiltInData(p, y>>1)[0] != 0 {
		t.Fatal("new node was not correctly simulated")
	}
}
