package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCexMinFrameTerms(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	frames := Saig_ManCexMinCollectFrameTerms(g, c)
	if vec.Vec_VecSize(frames) != 2 {
		t.Fatal("frame count")
	}
	first, last := vec.Vec_IntArray(vec.Vec_VecEntryInt(frames, 0)), vec.Vec_IntArray(vec.Vec_VecEntryInt(frames, 1))
	if len(first) != 1 || first[0] != int32(pi.NodeID()) || len(last) != 1 || last[0] != int32(lo.NodeID()) {
		t.Fatalf("first %v last %v", first, last)
	}
}
