package llb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSweepQuantificationMaskPreservesLatchVariables(t *testing.T) {
	g := aig.NewGraph()
	pi, latch := g.NewInput(), g.NewInput()
	g.AddOutput(latch)
	li := g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	n := int(aig.Aig_ManObjNumMax(g))
	order := vec.Vec_IntStartFull(n)
	vec.Vec_IntWriteEntry(order, pi.NodeID(), 1)
	vec.Vec_IntWriteEntry(order, latch.NodeID(), 3)
	vec.Vec_IntWriteEntry(order, li.NodeID(), 2)
	for _, add := range []int32{0, 1} {
		mask := Llb_Nonlin4SweepVars2Q(g, order, add)
		if vec.Vec_IntSize(mask) != n {
			t.Fatalf("mask size = %d, want %d", vec.Vec_IntSize(mask), n)
		}
		for i := 0; i < n; i++ {
			want := int32(1)
			if i == 3 || (add != 0 && i == 2) {
				want = 0
			}
			if got := vec.Vec_IntEntry(mask, i); got != want {
				t.Errorf("add %d index %d = %d, want %d", add, i, got, want)
			}
		}
	}
}
