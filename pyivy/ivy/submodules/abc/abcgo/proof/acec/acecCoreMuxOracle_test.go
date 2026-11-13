package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// The MUX control is itself a gate, so collecting only the two data fanins
// would omit part of the cone needed to duplicate the arithmetic core.
func TestCoreCollectionNativeMuxControlCone(t *testing.T) {
	p := gia.Gia_ManStart(32)
	p.EnableMuxes()
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	control := gia.Gia_ManAppendAnd(p, a, b)
	sum := gia.Gia_ManAppendMuxReal(p, control, c, a)
	carry := gia.Gia_ManAppendAnd(p, b, c)
	gia.Gia_ManAppendCo(p, sum)
	gia.Gia_ManAppendCo(p, carry)
	adds := vec.Vec_IntAlloc(6)
	for _, v := range []int32{a >> 1, b >> 1, c >> 1, sum >> 1, carry >> 1, 0} {
		vec.Vec_IntPush(adds, v)
	}
	order := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(order, 0)
	leaves := vec.Vec_IntAlloc(3)
	for _, v := range []int32{a >> 1, b >> 1, c >> 1} {
		vec.Vec_IntPush(leaves, v)
	}
	nodes := Gia_PolynCoreCollect(p, adds, order)
	for _, mode := range []int32{0, 1} {
		dst := Gia_PolynCoreDupTree(p, nil, leaves, nodes, mode)
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			want, err := p.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			got, err := dst.Evaluate(inputs)
			if err != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
				t.Fatalf("mode %d inputs %v got %v want %v err %v", mode, inputs, got, want, err)
			}
		}
	}
}

func TestCoreCollectionConstantFanin(t *testing.T) {
	p := gia.Gia_ManStart(16)
	p.EnableMuxes()
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	sum := gia.Gia_ManAppendMuxReal(p, 0, a, b)
	carry := gia.Gia_ManAppendAnd(p, b, c)
	gia.Gia_ManAppendCo(p, sum)
	gia.Gia_ManAppendCo(p, carry)
	adds := vec.Vec_IntAlloc(6)
	for _, v := range []int32{a >> 1, b >> 1, c >> 1, sum >> 1, carry >> 1, 0} {
		vec.Vec_IntPush(adds, v)
	}
	order := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(order, 0)
	leaves := vec.Vec_IntAlloc(3)
	for _, v := range []int32{a >> 1, b >> 1, c >> 1} {
		vec.Vec_IntPush(leaves, v)
	}
	nodes := Gia_PolynCoreCollect(p, adds, order)
	dst := Gia_PolynCoreDupTree(p, nil, leaves, nodes, 0)
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		want, err := p.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := dst.Evaluate(inputs)
		if err != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("inputs %v got %v want %v err %v", inputs, got, want, err)
		}
	}
}
