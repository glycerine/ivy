package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestFramesNonInitial(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.XOR(pi, lo)
	g.AddOutput(lo)
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 1)
	q := Saig_ManFramesNonInitial(g, 2)
	defer Saig_ManStopMap1(g)
	if q == nil || q.InputCount() != 3 || q.OutputCount() != 3 {
		t.Fatalf("frame graph %v", q)
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	if Saig_ManGetMap1(g, pi, 0) != int32(q.Inputs()[1].NodeID()) || Saig_ManGetMap1(g, pi, 1) != int32(q.Inputs()[2].NodeID()) || Saig_ManGetMap1(g, lo, 0) != int32(q.Inputs()[0].NodeID()) {
		t.Fatal("incorrect input maps")
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != in[0] || out[1] != (in[0] != in[1]) || out[2] != (in[0] != in[1] != in[2]) {
			t.Fatalf("%v -> %v (%v)", in, out, err)
		}
	}
}
