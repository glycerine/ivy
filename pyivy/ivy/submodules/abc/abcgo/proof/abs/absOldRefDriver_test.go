package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestLegacyRefinementRunsAllCheckingBackends(t *testing.T) {
	for _, mode := range []string{"bmc", "bdd", "pdr"} {
		t.Run(mode, func(t *testing.T) {
			g := aig.NewGraph()
			x, lo0, lo1 := g.NewInput(), g.NewInput(), g.NewInput()
			g.AddOutput(g.AND(x, g.OR(lo0, lo1)))
			g.AddOutput(lo0)
			g.AddOutput(lo1)
			aig.Aig_ManSetRegNum(g, 2)
			flops := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
			abstract := saig.Saig_ManDupAbstraction(g, flops)
			var bdd, pdr int32
			if mode == "bdd" {
				bdd = 1
			}
			if mode == "pdr" {
				pdr = 1
			}
			var status, framesDone int32
			next := Saig_ManCexRefine(g, abstract, flops, 3, 0, bdd, pdr, 0, nil, &status, &framesDone)
			if next == nil {
				t.Fatalf("backend %s failed to refine (status %d)", mode, status)
			}
			if got := vec.Vec_IntArray(flops); !reflect.DeepEqual(got, []int32{0, 1}) {
				t.Fatalf("flops %v", got)
			}
		})
	}
}

func TestLegacyAbstractionDriverStopsAfterSafeBound(t *testing.T) {
	g := aig.NewGraph()
	lo := g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 1)
	params := Gia_ParAbs_t{nFramesBmc: 2, nRatio: 0}
	got := Saig_ManCexAbstractionFlops(g, &params)
	if got == nil || !reflect.DeepEqual(vec.Vec_IntArray(got), []int32{0}) {
		t.Fatalf("flops %v", got)
	}
}
