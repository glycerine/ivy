package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestPolynomialOrderSimpleAnd(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	n := gia.Gia_ManAppendAnd(p, a, b)
	gia.Gia_ManAppendCo(p, n)
	order := Gia_PolynReorder(p, 0, 0)
	if got := vec.Vec_IntEntry(order, int(a>>1)); got != 1 {
		t.Errorf("a order %d", got)
	}
	if got := vec.Vec_IntEntry(order, int(b>>1)); got != 2 {
		t.Errorf("b order %d", got)
	}
	if got := vec.Vec_IntEntry(order, int(n>>1)); got != 3 {
		t.Errorf("AND order %d", got)
	}
	if p.PolynXors() == nil || vec.Vec_IntSize(p.PolynXors()) != 0 {
		t.Fatal("expected stored empty XOR collection")
	}
}

func TestPolynomialOrderNativeXor(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManAppendXorReal(p, a, b)
	gia.Gia_ManAppendCo(p, x)
	order := Gia_PolynReorder(p, 0, 0)
	if got := vec.Vec_IntEntry(order, int(x>>1)); got != 3 {
		t.Fatalf("XOR order %d", got)
	}
	if p.PolynXors() == nil || vec.Vec_IntSize(p.PolynXors()) != 1 || vec.Vec_IntEntry(p.PolynXors(), 0) != x>>1 {
		t.Fatal("XOR root collection incorrect")
	}
}
