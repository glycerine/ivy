package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func temporalFixture() *aig.Graph {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(g.XOR(pi, lo))
	aig.Aig_ManSetRegNum(g, 1)
	return g
}
func TestTemporalFrames(t *testing.T) {
	q := Saig_ManTemporFrames(temporalFixture(), 3)
	if q == nil || q.InputCount() != 3 || q.OutputCount() != 1 {
		t.Fatalf("frames %v", q)
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != (in[0] != in[1] != in[2]) {
			t.Fatalf("%v -> %v (%v)", in, out, err)
		}
	}
}
func TestTemporalDecompose(t *testing.T) {
	q := Saig_ManTemporDecompose(temporalFixture(), 2)
	if q == nil || q.InputCount() != 5 || q.OutputCount() != 3 || q.RegisterCount() != 2 {
		t.Fatalf("decomposition %v", q)
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 32; bits++ {
		in := make([]bool, 5)
		for i := range in {
			in[i] = bits&(1<<i) != 0
		}
		out, err := q.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		state := in[1] != in[2]
		if in[3] {
			state = in[4]
		}
		if out[0] != state || !out[1] || out[2] != (in[0] != state) {
			t.Fatalf("%v -> %v", in, out)
		}
	}
}
