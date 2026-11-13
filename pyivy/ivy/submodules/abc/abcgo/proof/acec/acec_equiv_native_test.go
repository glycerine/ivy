package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestEquivalenceClassesDistinguishNativeXorAndAnd(t *testing.T) {
	makeGraph := func(xor bool) (*Gia_Man_t, int32) {
		p := gia.Gia_ManStart(16)
		a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
		var lit int32
		if xor {
			lit = gia.Gia_ManAppendXorReal(p, a, b)
		} else {
			lit = gia.Gia_ManAppendAnd(p, a, b)
		}
		gia.Gia_ManAppendCo(p, lit)
		return p, lit >> 1
	}
	a, ai := makeGraph(true)
	b, bi := makeGraph(false)
	var ma, mb *Vec_Int_t
	Acec_ComputeEquivClasses(a, b, &ma, &mb)
	x, y := vec.Vec_IntEntry(ma, int(ai)), vec.Vec_IntEntry(mb, int(bi))
	if x>>1 == y>>1 {
		t.Fatalf("XOR and AND put in same class: %d %d", x, y)
	}
	c, ci := makeGraph(true)
	a, ai = makeGraph(true)
	Acec_ComputeEquivClasses(a, c, &ma, &mb)
	x, y = vec.Vec_IntEntry(ma, int(ai)), vec.Vec_IntEntry(mb, int(ci))
	if x != y {
		t.Fatalf("identical XORs assigned different classes: %d %d", x, y)
	}
}
