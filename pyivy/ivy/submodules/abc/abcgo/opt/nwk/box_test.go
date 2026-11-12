package nwk

import (
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestTimingAcrossBoxAndQueueInsertion(t *testing.T) {
	p := Nwk_ManAlloc()
	pi := Nwk_ManCreateCi(p, 0)
	pre := Nwk_ManCreateNode(p, 1, 0)
	Nwk_ObjAddFanin(pre, pi)
	pre.Function = hop.Hop_IthVar(p.Hop, 0)
	boxIn := Nwk_ManCreateCo(p)
	Nwk_ObjAddFanin(boxIn, pre)
	boxOut := Nwk_ManCreateCi(p, 0)
	post := Nwk_ManCreateNode(p, 1, 0)
	Nwk_ObjAddFanin(post, boxOut)
	post.Function = hop.Hop_IthVar(p.Hop, 0)
	po := Nwk_ManCreateCo(p)
	Nwk_ObjAddFanin(po, post)
	p.Timing = tim.Tim_ManStart(2, 2)
	tim.Tim_ManSetDelayTables(p.Timing, [][]float32{{0, 1, 1, 3}})
	tim.Tim_ManCreateBox(p.Timing, 0, 1, 1, 1, 0, 0)
	if Nwk_ManLevel(p) != 3 || Nwk_ManVerifyTopoOrder(p) != 1 {
		t.Fatal("box levels/order")
	}
	if delay := Nwk_ManDelayTraceLut(p); delay != 5 {
		t.Fatalf("box delay %g", delay)
	}
	if pi.Required != 0 || boxOut.Required != 4 {
		t.Fatalf("box required pi%g boxout%g", pi.Required, boxOut.Required)
	}
	v := vec.Vec_PtrAlloc(0)
	Nwk_NodeUpdateAddToQueue(v, post, -1, 1)
	Nwk_NodeUpdateAddToQueue(v, pre, -1, 1)
	if vectorNodes(v)[0] != pre {
		t.Fatal("queue ordering")
	}
	if Nwk_ObjLevelNew(boxOut) != 2 {
		t.Fatal("box incremental level")
	}
	pre.Level = 9
	Nwk_ManUpdateLevel(pre)
	if po.Level != 3 {
		t.Fatal("box level propagation")
	}
}
func TestDuplicateAdjacencyPreservesMultiplicity(t *testing.T) {
	p := Nwk_ManAlloc()
	a, b := Nwk_ManCreateCi(p, 0), Nwk_ManCreateCi(p, 0)
	n := Nwk_ManCreateNode(p, 2, 0)
	Nwk_ObjAddFanin(n, a)
	Nwk_ObjAddFanin(n, a)
	Nwk_ObjPatchFanin(n, a, b)
	if len(a.Fanouts) != 1 || len(b.Fanouts) != 1 || n.Fanins[1] != a {
		t.Fatal("duplicate patch corrupted adjacency")
	}
	Nwk_ObjDeleteFanin(n, a)
	if len(a.Fanouts) != 0 || len(n.Fanins) != 1 {
		t.Fatal("duplicate deletion")
	}
}
