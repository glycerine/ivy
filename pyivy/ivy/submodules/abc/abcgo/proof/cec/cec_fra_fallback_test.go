package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestFraFallbackTruthTables(t *testing.T) {
	for _, unequal := range []bool{false, true} {
		g := gia.Gia_ManStart(24)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, a, b), c)
		y := gia.Gia_ManAppendAnd(g, a, gia.Gia_ManAppendAnd(g, b, c))
		gia.Gia_ManAppendCo(g, x)
		gia.Gia_ManAppendCo(g, y)
		gia.Gia_ManAppendCo(g, x)
		if unequal {
			gia.Gia_ManAppendCo(g, a)
		} else {
			gia.Gia_ManAppendCo(g, y)
		}
		status, output, err := cecVerifyFra(g, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !unequal {
			if status != 1 || output != -1 || g.CombinationalCounterexample() != nil {
				t.Fatalf("equivalent: %d %d", status, output)
			}
			continue
		}
		if status != 0 || output != 1 || g.CombinationalCounterexample() == nil {
			t.Fatalf("different: %d %d", status, output)
		}
		w := g.CombinationalCounterexample()
		in := []bool{w.PData[0]&1 != 0, w.PData[0]&2 != 0, w.PData[0]&4 != 0}
		got, e := g.Evaluate(in)
		if e != nil || got[2] == got[3] {
			t.Fatalf("invalid witness %v output %v error %v", in, got, e)
		}
	}
}

func TestFraFallbackClearsStaleWitness(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, x)
	if err := g.SetCombinationalCounterexample([]bool{true, true}, 0); err != nil {
		t.Fatal(err)
	}
	status, _, err := cecVerifyFra(g, 0)
	if err != nil || status != 1 || g.CombinationalCounterexample() != nil {
		t.Fatalf("status %d error %v stale witness %+v", status, err, g.CombinationalCounterexample())
	}
}

func TestFraFallbackNativeAndWideWitness(t *testing.T) {
	g := gia.Gia_ManStart(60)
	g.EnableMuxes()
	inputs := make([]int32, 40)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(g)
	}
	x := gia.Gia_ManAppendXorReal(g, inputs[35], inputs[2])
	mux := gia.Gia_ManAppendMuxReal(g, inputs[38], x, inputs[39])
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, mux)
	status, output, err := cecVerifyFra(g, 0)
	if err != nil || status != 0 || output != 0 {
		t.Fatalf("status %d output %d error %v", status, output, err)
	}
	w := g.CombinationalCounterexample()
	if w == nil || w.NPis != 40 {
		t.Fatalf("witness %+v", w)
	}
	in := make([]bool, 40)
	for i := range in {
		in[i] = w.PData[i>>5]&(uint32(1)<<uint(i&31)) != 0
	}
	got, e := g.Evaluate(in)
	if e != nil || got[0] == got[1] {
		t.Fatalf("witness %v outputs %v error %v", in, got, e)
	}
}
