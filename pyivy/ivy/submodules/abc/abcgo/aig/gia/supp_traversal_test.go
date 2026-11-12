package gia

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestMin2GatherSupportOverlap(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	cis, nodes := vec.Vec_IntAlloc(4), vec.Vec_IntAlloc(4)
	Gia_ManIncrementTravId(p)
	Gia_ManIncrementTravId(p)
	if Gia_Min2ManGatherSupp_rec(p, x>>1, cis, nodes) != 0 {
		t.Fatal("first cone overlaps")
	}
	if !reflect.DeepEqual(vec.Vec_IntArray(cis), []int32{a >> 1, b >> 1}) || !reflect.DeepEqual(vec.Vec_IntArray(nodes), []int32{x >> 1}) {
		t.Fatal("wrong first cone")
	}
	vec.Vec_IntClear(cis)
	vec.Vec_IntClear(nodes)
	Gia_ManIncrementTravId(p)
	if Gia_Min2ManGatherSupp_rec(p, y>>1, cis, nodes) != 1 {
		t.Fatal("missing overlap")
	}
	if !reflect.DeepEqual(vec.Vec_IntArray(cis), []int32{c >> 1}) || !reflect.DeepEqual(vec.Vec_IntArray(nodes), []int32{y >> 1}) {
		t.Fatal("wrong second cone")
	}
}
