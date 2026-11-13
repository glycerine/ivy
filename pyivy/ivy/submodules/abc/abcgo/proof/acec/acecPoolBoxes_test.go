package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCollectAdderBoxChain(t *testing.T) {
	p := gia.Gia_ManStart(20)
	for i := 0; i < 11; i++ {
		gia.Gia_ManAppendCi(p)
	}
	adds := vec.Vec_IntAlloc(12)
	for _, x := range []int32{1, 2, 3, 4, 5, 0, 5, 6, 7, 8, 9, 0} {
		vec.Vec_IntPush(adds, x)
	}
	xors := vec.Vec_IntAlloc(4)
	for _, x := range []int32{10, 9, 6, 7} {
		vec.Vec_IntPush(xors, x)
	}
	sets := Acec_ManCollectBoxSets(p, adds, xors)
	if vec.Vec_WecSize(sets) != 5 {
		t.Fatalf("levels %d", vec.Vec_WecSize(sets))
	}
	boxes := vec.Vec_IntArray(vec.Vec_WecEntry(sets, 2))
	if len(boxes) != 2 || boxes[0] != 0 || boxes[1] != 1 {
		t.Fatalf("boxes %v", boxes)
	}
	ins := vec.Vec_IntArray(vec.Vec_WecEntry(sets, 3))
	outs := vec.Vec_IntArray(vec.Vec_WecEntry(sets, 4))
	if len(ins) != 10 || len(outs) != 6 {
		t.Fatalf("inputs %v outputs %v", ins, outs)
	}
	for i := 3; i < len(ins); i += 2 {
		if ins[i] < ins[i-2] {
			t.Fatalf("unsorted input ranks %v", ins)
		}
	}
}
