package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCexMinCubeOutputs(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	v := vec.Vec_VecAlloc(3)
	vec.Vec_VecExpandInt(v, 2)
	vec.Vec_IntPush(vec.Vec_VecEntryInt(v, 1), 0)
	vec.Vec_IntPush(vec.Vec_VecEntryInt(v, 2), 1)
	q := Saig_ManCexMinDupWithCubes(g, v)
	if q == nil || q.OutputCount() != 3 || q.RegisterCount() != 1 {
		t.Fatal("incorrect interface")
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	for _, x := range []bool{false, true} {
		out, err := q.Evaluate([]bool{x, false})
		if err != nil || out[0] != x || out[1] == x || out[2] != x {
			t.Fatalf("%v -> %v (%v)", x, out, err)
		}
	}
}
