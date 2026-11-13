package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// The compiled ABC fixture ssw_latch_default.c returns zero registers, one
// primary input, and one false primary output for two identical latches.
func TestLatchCorrespondenceDefaultMatchesC(t *testing.T) {
	g := aig.NewGraph()
	x, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.XOR(a, b))
	g.AddOutput(x)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 2)
	out := Ssw_LatchCorrespondence(g, nil)
	if out == nil || out.RegisterCount() != 0 || out.InputCount() != 1 || out.OutputCount() != 1 {
		t.Fatalf("unexpected correspondence interface: %v", out)
	}
	for _, value := range []bool{false, true} {
		values, err := out.Evaluate([]bool{value})
		if err != nil || len(values) != 1 || values[0] {
			t.Fatalf("input %t output %v err %v", value, values, err)
		}
	}
}
