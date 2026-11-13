package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	abcmain "github.com/glycerine/abcgo/base/main"
	"testing"
)

func TestSecSaveUnsolvedCopiesGraph(t *testing.T) {
	graph := aig.NewGraph()
	x := graph.NewInput()
	graph.AddOutput(x)
	secSaveUnsolved(graph)
	got, ok := abcmain.Abc_FrameReadSave1().(*aig.Graph)
	if !ok || got == nil || got == graph {
		t.Fatal("expected independent saved graph")
	}
	out, err := got.Evaluate([]bool{true})
	if err != nil || len(out) != 1 || !out[0] {
		t.Fatalf("saved graph: %v %v", out, err)
	}
}
