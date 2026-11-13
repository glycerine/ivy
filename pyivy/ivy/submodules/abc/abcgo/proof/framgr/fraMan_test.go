package framgr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestManagerLifecycleRetainsFrameMapping(t *testing.T) {
	graph := aig.NewGraph()
	input, latch := graph.NewInput(), graph.NewInput()
	graph.AddOutput(input)
	graph.AddOutput(latch)
	aig.Aig_ManSetRegNum(graph, 1)
	var pars Fra_Par_t
	Fra_ParamsDefaultSeq(&pars)
	p := Fra_ManStart(graph, &pars)
	if p.nFramesAll != 2 || p.nPatWords != 1 {
		t.Fatalf("frames, words = %d, %d", p.nFramesAll, p.nPatWords)
	}
	if Fra_ObjFraig(p, input, 1) != aig.InvalidEdge {
		t.Fatal("unmapped frame is not invalid")
	}
	p.pManFraig = Fra_ManPrepareComb(p)
	if p.pManFraig.RegisterCount() != 1 || p.pManFraig.OutputCount() != 0 {
		t.Fatal("register metadata was not set during construction")
	}
	Fra_ManFinalizeComb(p)
	if p.pManFraig.OutputCount() != graph.OutputCount() || p.pManFraig.RegisterCount() != 1 {
		t.Fatal("invalid completed interface")
	}
	copied := Fra_ObjFraig(p, input, 0)
	Fra_ObjSetFraig(p, input, 1, copied.Not())
	mapping := Fra_ManStop(p)
	if mapping[input.NodeID()*2] != copied || mapping[input.NodeID()*2+1] != copied.Not() {
		t.Fatal("frame mapping lost on shutdown")
	}
}

func TestManagerCleanResetsSATMaps(t *testing.T) {
	p := &Fra_Man_t{}
	Fra_ManClean(p, 2)
	p.pMemFanins[1] = []aig.Edge{2}
	p.pMemSatNums[1] = 4
	capacity := p.nMemAlloc
	Fra_ManClean(p, 1)
	if p.nMemAlloc != capacity || p.pMemFanins[1] != nil || p.pMemSatNums[1] != 0 {
		t.Fatal("manager maps not cleared")
	}
}
