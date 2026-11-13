package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestEquivalenceSweepRequiresResult(t *testing.T) {
	makeGraph := func() *Gia_Man_t {
		p := gia.Gia_ManStart(8)
		a := gia.Gia_ManAppendCi(p)
		gia.Gia_ManAppendCo(p, a)
		return p
	}
	a, b := makeGraph(), makeGraph()
	var m0, m1 *Vec_Int_t
	called := false
	defer func() {
		if recover() == nil || !called || m0 == nil || m1 == nil {
			t.Fatalf("expected source sweep precondition after maps, called=%v maps=%v,%v", called, m0, m1)
		}
	}()
	acecComputeEquivClasses(a, b, &m0, &m1, func(*gia.Gia_Man_t, int32, int32) *gia.Gia_Man_t { called = true; return nil })
}
