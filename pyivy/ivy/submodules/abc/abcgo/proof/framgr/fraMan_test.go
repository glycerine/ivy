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
	if p.NFramesAll != 2 || p.NPatWords != 1 {
		t.Fatalf("frames, words = %d, %d", p.NFramesAll, p.NPatWords)
	}
	if Fra_ObjFraig(p, input, 1) != aig.InvalidEdge {
		t.Fatal("unmapped frame is not invalid")
	}
	p.PManFraig = Fra_ManPrepareComb(p)
	if p.PManFraig.RegisterCount() != 1 || p.PManFraig.OutputCount() != 0 {
		t.Fatal("register metadata was not set during construction")
	}
	Fra_ManFinalizeComb(p)
	if p.PManFraig.OutputCount() != graph.OutputCount() || p.PManFraig.RegisterCount() != 1 {
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
	p.PMemFanins[1] = []aig.Edge{2}
	p.PMemSatNums[1] = 4
	capacity := p.NMemAlloc
	Fra_ManClean(p, 1)
	if p.NMemAlloc != capacity || p.PMemFanins[1] != nil || p.PMemSatNums[1] != 0 {
		t.Fatal("manager maps not cleared")
	}
}

// Fra_ManClean in fraMan.c reallocates with 5000 spare slots when the
// requested node count grows, and zeroes both old and newly allocated slots.
func TestManagerCleanGrowthDiscardsAllOldMappings(t *testing.T) {
	p := &Fra_Man_t{}
	Fra_ManClean(p, 2)
	oldCapacity := p.NMemAlloc
	p.PMemFanins[0] = []aig.Edge{aig.Edge(2)}
	p.PMemSatNums[0] = 7
	p.PMemFanins[oldCapacity-1] = []aig.Edge{aig.Edge(4)}
	p.PMemSatNums[oldCapacity-1] = 9
	wantSize := oldCapacity + 1
	Fra_ManClean(p, wantSize)
	if p.NMemAlloc != wantSize+5000 || int32(len(p.PMemFanins)) != p.NMemAlloc || int32(len(p.PMemSatNums)) != p.NMemAlloc {
		t.Fatalf("unexpected grown mapping capacity %d", p.NMemAlloc)
	}
	for _, i := range []int32{0, oldCapacity - 1, oldCapacity, p.NMemAlloc - 1} {
		if p.PMemFanins[i] != nil || p.PMemSatNums[i] != 0 {
			t.Fatalf("mapping slot %d retained old data", i)
		}
	}
}
