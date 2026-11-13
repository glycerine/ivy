package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestEquivalenceFailureDoesNotPublishMaps(t *testing.T) {
	makeGraph := func() *Gia_Man_t {
		p := gia.Gia_ManStart(8)
		a := gia.Gia_ManAppendCi(p)
		gia.Gia_ManAppendCo(p, a)
		return p
	}
	a, b := makeGraph(), makeGraph()
	m0, m1 := vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
	called := false
	ok := acecComputeEquivClasses(a, b, &m0, &m1, func(*gia.Gia_Man_t, int32, int32) *gia.Gia_Man_t { called = true; return nil })
	if ok || !called || m0 != nil || m1 != nil {
		t.Fatalf("ok=%v called=%v maps=%v,%v", ok, called, m0, m1)
	}
}
