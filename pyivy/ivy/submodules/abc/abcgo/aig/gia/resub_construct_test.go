package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestResubConstructAndInsert(t *testing.T) {
	g := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	node := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, node)
	objects := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(objects, node>>1)
	funcs := vec.Vec_WecAlloc(1)
	gates := vec.Vec_WecPushLevel(funcs)
	// Reverse variable order encodes XOR; the top literal complements it.
	vec.Vec_IntPush(gates, b)
	vec.Vec_IntPush(gates, a)
	vec.Vec_IntPush(gates, 2*g.nObjs+1)
	result := Gia_ManInsertFromGates(g, objects, funcs)
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		got, err := result.Evaluate(inputs)
		want := inputs[0] == inputs[1]
		if err != nil || got[0] != want {
			t.Fatalf("%v -> %v want %v (%v)", inputs, got, want, err)
		}
	}
}
func TestResubConstructDirectVariable(t *testing.T) {
	funcs := vec.Vec_WecAlloc(1)
	f := vec.Vec_WecPushLevel(funcs)
	vec.Vec_IntPush(f, 3)
	g := Gia_ManConstructFromGates(funcs, 2)
	got, err := g.Evaluate([]bool{true})
	if err != nil || got[0] {
		t.Fatalf("direct variable %v %v", got, err)
	}
}
