package int_test

import (
	"github.com/glycerine/abcgo/aig/aig"
	interp "github.com/glycerine/abcgo/proof/int"
	"testing"
)

func TestExternalInterpolationOptions(t *testing.T) {
	var options interp.Inter_ManParams_t
	interp.Inter_ManSetDefaultParams(&options)
	if options.NFramesK != 1 || options.FCheckKstep != 1 || options.IFrameMax != -1 {
		t.Fatalf("unexpected defaults: %+v", options)
	}
	options.NFramesMax = 1
	options.FVerbose = 0
	graph := aig.NewGraph()
	graph.NewInput()
	state := graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(state)
	aig.Aig_ManSetRegNum(graph, 1)
	var frame int32
	if got := interp.Inter_ManPerformInterpolation(graph, &options, &frame); got != -1 || frame != -1 {
		t.Fatalf("got %d frame %d", got, frame)
	}
}
