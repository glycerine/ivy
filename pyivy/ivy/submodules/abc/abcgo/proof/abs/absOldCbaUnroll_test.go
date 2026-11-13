package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestCbaUnrollRecordsLatchInput(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	li := g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(0, 1, 2)
	c.IFrame = 1
	util.Abc_InfoSetBit(c.PData, 0)
	var mapping *vec.Vec_Int_t
	frames := vec.Vec_VecAlloc(2)
	graph := Saig_ManCbaUnrollWithCex(g, c, 0, &mapping, &frames)
	if got := vec.Vec_IntArray(mapping); !reflect.DeepEqual(got, []int32{0, 0}) {
		t.Fatalf("mapping %v", got)
	}
	entries := vec.Vec_IntArray(vec.Vec_VecEntryInt(frames, 0))
	if len(entries) != 2 || entries[0] != int32(li.NodeID()) || entries[1] != int32(graph.Inputs()[0]) {
		t.Fatalf("register/frame map %v", entries)
	}
}
