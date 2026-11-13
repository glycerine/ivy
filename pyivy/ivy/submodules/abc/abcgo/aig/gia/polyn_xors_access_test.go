package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestPolynXorsAccessor(t *testing.T) {
	p := Gia_ManStart(4)
	if p.PolynXors() != nil {
		t.Fatal("unexpected XOR data")
	}
	x := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(x, 1)
	p.SetPolynXors(x)
	if p.PolynXors() != x {
		t.Fatal("XOR data not retained")
	}
	p.SetPolynXors(nil)
	if p.PolynXors() != nil {
		t.Fatal("XOR data not cleared")
	}
}
