package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestResubManagerVerify(t *testing.T) {
	a, b := word(0xaaaaaaaaaaaaaaaa), word(0xcccccccccccccccc)
	on := a & b
	divs := vec.Vec_PtrAlloc(4)
	for _, v := range []word{^on, on, a, b} {
		vec.Vec_PtrPush(divs, []word{v})
	}
	p := Gia_ResbAlloc(1)
	Gia_ResbInit(p, divs, 1, 2, 10, 0, 1, 0, 0, 0)
	top := Gia_ManResubAddNode(p, 4, 6, 0)
	vec.Vec_IntPush(p.vGates, top)
	result := []word{0}
	if Gia_ManResubVerify(p, result) != 1 || result[0] != on {
		t.Fatalf("invalid AND result %x", result)
	}
	vec.Vec_IntWriteEntry(p.vGates, vec.Vec_IntSize(p.vGates)-1, top^1)
	if Gia_ManResubVerify(p, nil) != 0 {
		t.Fatal("incorrect function accepted")
	}
	Gia_ResbFree(p)
}
