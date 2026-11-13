package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestInductionRewritePreservesFrameInterface(t *testing.T) {
	source := aig.NewGraph()
	input, latch := source.NewInput(), source.NewInput()
	source.AddOutput(input)
	aig.Aig_ManSetRegNum(source, 1)
	p := Fra_ManStart(source, &Fra_Par_t{NFramesK: 1})
	p.pManFraig = Fra_FramesWithClasses(p)
	Fra_FraigInductionRewrite(p)
	if p.pManFraig.RegisterCount() != 1 || p.pManFraig.OutputCount() != 1 {
		t.Fatal("rewritten interface changed")
	}
	if Fra_ObjFraig(p, input, 1) == aig.InvalidEdge || Fra_ObjFraig(p, latch, 1) == aig.InvalidEdge {
		t.Fatal("missing rewritten frame mapping")
	}
	if Fra_ObjFraig(p, latch, 0) != aig.InvalidEdge {
		t.Fatal("stale old frame mapping")
	}
}

func TestFramesAddMoreTracksLatchCopies(t *testing.T) {
	graph := aig.NewGraph()
	input, latch := graph.NewInput(), graph.NewInput()
	node := graph.AND(input, latch)
	graph.AddOutput(input)
	aig.Aig_ManSetRegNum(graph, 1)
	Fra_FramesAddMore(graph, 1)
	if aig.Aig_ObjCopy(graph, latch) != input {
		t.Fatal("latch next-state copy missing")
	}
	if aig.Aig_ObjCopy(graph, node) != input {
		t.Fatal("node copy does not use next state")
	}
}
