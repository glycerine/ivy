package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCexMinReasonsAcrossFrames(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	c.PData[0] = 2
	for _, test := range []struct {
		pi          int32
		first, last []int32
	}{{0, nil, []int32{0}}, {1, []int32{0}, nil}} {
		reasons := Saig_ManCexMinComputeReason(g, c, test.pi)
		for f, want := range [][]int32{test.first, test.last} {
			got := vec.Vec_IntArray(vec.Vec_VecEntryInt(reasons, f))
			if len(got) != len(want) {
				t.Fatalf("pi=%d frame=%d got %v want %v", test.pi, f, got, want)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("pi=%d frame=%d got %v want %v", test.pi, f, got, want)
				}
			}
		}
	}
}
func TestCexMinChoosesOneFalseAndInput(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b).Not())
	c := util.Abc_CexAlloc(0, 2, 1)
	reasons := Saig_ManCexMinComputeReason(g, c, 1)
	got := vec.Vec_IntArray(vec.Vec_VecEntryInt(reasons, 0))
	if len(got) != 1 || got[0]&1 != 1 {
		t.Fatalf("reasons %v", got)
	}
}
