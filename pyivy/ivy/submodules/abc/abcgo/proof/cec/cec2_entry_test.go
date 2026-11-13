package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec2EntryPoint(t *testing.T) {
	var opts Cec2_Par_t
	Cec2_SetDefaultParams(&opts)
	if opts.nSimWords != 12 || opts.nSimRounds != 4 || opts.nItersMax != 10 || opts.nConfLimit != 1000 || opts.fUseCones != 1 {
		t.Fatalf("defaults %+v", opts)
	}
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, x)
	var out *Gia_Man_t
	if got := Cec2_ManPerformSweeping(g, &opts, &out); got != 1 || out == nil {
		t.Fatalf("non-miter result %d %v", got, out)
	}
	for _, v := range []bool{false, true} {
		values, err := out.Evaluate([]bool{v})
		if err != nil || values[0] != v {
			t.Fatalf("input %v output %v err %v", v, values, err)
		}
	}
	opts.fIsMiter = 1
	var sentinel *Gia_Man_t = out
	if got := Cec2_ManPerformSweeping(g, &opts, &sentinel); got != 0 || g.SequentialCounterexample() == nil {
		t.Fatalf("miter result %d", got)
	}
	if sentinel != out {
		t.Fatal("early miter failure overwrote output")
	}
}
