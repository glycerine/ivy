package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestCbaReasonRetainsTraversalForShrinking(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	root := g.AND(x, y)
	g.AddOutput(root)
	c := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(c.PData, 0)
	p := &Saig_ManCba_t{pFrames: g, pCex: c, vMapPiF2A: vec.Vec_IntAllocArrayCopy([]int32{0, 0, 1, 0}, 4)}
	if got := vec.Vec_IntArray(Saig_ManCbaFindReason(p)); !reflect.DeepEqual(got, []int32{1}) {
		t.Fatalf("reasons %v", got)
	}
	if !p.visited[root.NodeID()] || !p.visited[y.NodeID()] || p.visited[x.NodeID()] {
		t.Fatalf("incorrect traversal marks %v", p.visited)
	}
}
