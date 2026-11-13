package int

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

// A two-latch shift register requires a three-frame trace to reach the bad
// output from the all-zero initial state.
func TestInterpolationTwoRegisterCounterexample(t *testing.T) {
	graph := aig.NewGraph()
	input, first, second := graph.NewInput(), graph.NewInput(), graph.NewInput()
	graph.AddOutput(second)
	graph.AddOutput(input)
	graph.AddOutput(first)
	aig.Aig_ManSetRegNum(graph, 2)
	var opts Inter_ManParams_t
	Inter_ManSetDefaultParams(&opts)
	opts.NFramesMax = 6
	var frame int32
	if got := Inter_ManPerformInterpolation(graph, &opts, &frame); got != 0 || frame != 2 {
		t.Fatalf("got %d frame %d", got, frame)
	}
	cex := graph.SequentialCounterexample()
	if cex == nil || cex.IFrame != 2 {
		t.Fatalf("bad witness: %+v", cex)
	}
	state := []bool{false, false}
	for f := int32(0); f <= cex.IFrame; f++ {
		x := util.Abc_InfoHasBit(cex.PData, 2+f)
		out, err := graph.Evaluate([]bool{x, state[0], state[1]})
		if err != nil {
			t.Fatal(err)
		}
		if f == cex.IFrame && !out[0] {
			t.Fatalf("witness does not violate property: %v", out)
		}
		state = []bool{out[1], out[2]}
	}
}
