package ivy

import (
	"github.com/glycerine/abcgo/bool/dec"
	"testing"
)

func TestCountRewriteWithMissingIntermediateAND(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	root := Ivy_And(p, a, c)
	Ivy_ManIncrementTravId(p)
	g := dec.Dec_GraphCreate(3)
	g.SetLeafSignal(0, a)
	g.SetLeafSignal(1, b)
	g.SetLeafSignal(2, c)
	inner := dec.Dec_GraphAddNodeAnd(g, dec.Dec_EdgeCreate(0, 0), dec.Dec_EdgeCreate(1, 1))
	dec.Dec_GraphSetRoot(g, dec.Dec_GraphAddNodeAnd(g, inner, dec.Dec_EdgeCreate(2, 0)))
	if got := Ivy_GraphToNetworkCount(p, root, g, 2, 2); got != 2 {
		t.Fatalf("count %d, want two missing ANDs", got)
	}
	if got := dec.Dec_GraphNodeFunc(g, int32(inner.Node)); got != nil {
		t.Errorf("missing AND retained payload %v", got)
	}
	if got := Ivy_GraphToNetworkCount(p, root, g, 1, 2); got != -1 {
		t.Fatalf("over-budget count %d, want rejection", got)
	}
}
