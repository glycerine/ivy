package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestTreePhasePropagationOnDetectedAdder(t *testing.T) {
	p := gia.Gia_ManStart(64)
	in := []int32{gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)}
	out := make([]int32, 2)
	Acec_InsertFadd(p, in, out)
	gia.Gia_ManAppendCo(p, out[0])
	gia.Gia_ManAppendCo(p, out[1])
	adds := Ree_ManComputeCuts(p, nil, 0)
	box := int32(-1)
	for i := 0; i+5 < vec.Vec_IntSize(adds); i += 6 {
		if vec.Vec_IntEntry(adds, i+3) == out[0]>>1 && vec.Vec_IntEntry(adds, i+4) == out[1]>>1 {
			box = int32(i / 6)
			break
		}
	}
	if box < 0 {
		t.Fatal("full adder missing")
	}
	boxes := vec.Vec_WecStart(1)
	vec.Vec_WecPush(boxes, 0, box)
	m := Acec_TreeCarryMap(p, adds, boxes)
	visited := vec.Vec_BitStart(vec.Vec_IntSize(adds) / 6)
	Acec_TreePhases_rec(p, adds, m, out[1]>>1, 1, visited)
	if Acec_SignBit2(adds, box, 4) != 1 {
		t.Fatal("carry phase missing")
	}
	for k := int32(0); k < 3; k++ {
		if Acec_SignBit2(adds, box, k) != Acec_SignBit(adds, box, k)^1 {
			t.Fatalf("input phase %d", k)
		}
	}
	if vec.Vec_BitEntry(visited, int(box)) != 1 {
		t.Fatal("adder was not visited")
	}
}
