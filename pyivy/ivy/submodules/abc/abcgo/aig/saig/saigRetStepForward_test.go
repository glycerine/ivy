package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestRetimeNodeForwardPreservesNextState(t *testing.T) {
	for _, invert := range []bool{false, true} {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		lo0, lo1 := g.NewInput(), g.NewInput()
		if invert {
			lo0, lo1 = lo0.Not(), lo1.Not()
		}
		node := g.AND(lo0, lo1)
		g.AddOutput(node)
		g.AddOutput(a)
		g.AddOutput(b)
		aig.Aig_ManSetRegNum(g, 2)
		aig.Aig_ManIncrementTravId(g)
		aig.Aig_ObjSetTravIdCurrent(g, aig.Aig_Regular(lo0))
		result := Saig_ManRetimeNodeFwd(g, node, 0)
		if result == aig.InvalidEdge || g.RegisterCount() != 3 {
			t.Fatalf("invert=%v: result=%v regs=%d", invert, result, g.RegisterCount())
		}
		for x := 0; x < 4; x++ {
			inputs := []bool{x&1 != 0, x&2 != 0, false, false, false}
			outputs, err := g.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			want := inputs[0] && inputs[1]
			if invert {
				want = inputs[0] || inputs[1]
			}
			if outputs[len(outputs)-1] != want {
				t.Fatalf("invert=%v inputs=%v: new latch input=%v want %v", invert, inputs, outputs[len(outputs)-1], want)
			}
		}
		if result.Complemented() != invert {
			t.Fatalf("invert=%v result complement=%v", invert, result.Complemented())
		}
	}
}
