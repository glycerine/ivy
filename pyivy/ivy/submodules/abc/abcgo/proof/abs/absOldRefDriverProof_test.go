package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLegacyRefinementReportsProofOnlyForProvedBackends(t *testing.T) {
	for _, mode := range []string{"bdd", "pdr"} {
		t.Run(mode, func(t *testing.T) {
			g := aig.NewGraph()
			lo := g.NewInput()
			g.AddOutput(lo)
			g.AddOutput(g.Const0())
			aig.Aig_ManSetRegNum(g, 1)
			flops := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
			abstract := saig.Saig_ManDupAbstraction(g, flops)
			var bdd, pdr int32
			if mode == "bdd" {
				bdd = 1
			} else {
				pdr = 1
			}
			var status, frames int32
			if next := Saig_ManCexRefine(g, abstract, flops, 3, 0, bdd, pdr, 0, nil, &status, &frames); next != nil || status != 1 {
				t.Fatalf("next %v status %d", next, status)
			}
		})
	}
}
