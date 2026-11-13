package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCexTargetEnlarge(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, lo := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, lo)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	q := Bmc_CexTargetEnlarge(g, 2)
	if q == nil || gia.Gia_ManCiNum(q) != 3 || gia.Gia_ManCoNum(q) != 1 {
		t.Fatal("wrong target interface")
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != in[1] {
			t.Fatalf("%v -> %v (%v)", in, out, err)
		}
	}
}
func TestCexTargetQuantifiesFrameInputs(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, lo := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, lo)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	q := Bmc_CexTarget(g, 1)
	if q == nil {
		t.Fatal("missing target")
	}
	if gia.Gia_ManRegNum(q) != 1 {
		t.Fatalf("regs %d", gia.Gia_ManRegNum(q))
	}
}
