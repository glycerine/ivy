package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCoreCollectionAndDuplication(t *testing.T) {
	p := gia.Gia_ManStart(32)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	sum, carry := gia.Gia_ManAppendAnd(p, a, b), gia.Gia_ManAppendAnd(p, b, c)
	gia.Gia_ManAppendCo(p, sum)
	gia.Gia_ManAppendCo(p, carry)
	adds := vec.Vec_IntAlloc(6)
	for _, v := range []int32{a >> 1, b >> 1, c >> 1, sum >> 1, carry >> 1, 0} {
		vec.Vec_IntPush(adds, v)
	}
	order := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(order, 0)
	nodes := Gia_PolynCoreCollect(p, adds, order)
	if vec.Vec_IntSize(nodes) != 2 {
		t.Fatalf("nodes %v", vec.Vec_IntArray(nodes))
	}
	leaves := vec.Vec_IntAlloc(3)
	for _, v := range []int32{a >> 1, b >> 1, c >> 1} {
		vec.Vec_IntPush(leaves, v)
	}
	for _, mode := range []int32{0, 1} {
		dst := Gia_PolynCoreDupTree(p, nil, leaves, nodes, mode)
		for i := 0; i < 8; i++ {
			in := []bool{i&1 != 0, i&2 != 0, i&4 != 0}
			want, e1 := p.Evaluate(in)
			got, e2 := dst.Evaluate(in)
			if e1 != nil || e2 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
				t.Fatalf("mode %d input %v got %v want %v", mode, in, got, want)
			}
		}
	}
}
