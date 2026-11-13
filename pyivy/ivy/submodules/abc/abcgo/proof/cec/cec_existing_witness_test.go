package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCec2AndCec3PreserveExistingWitnessResult(t *testing.T) {
	for _, engine := range []int{2, 3} {
		g := gia.Gia_ManStart(4)
		x := gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, x)
		cex := util.Abc_CexAlloc(0, 1, 1)
		cex.PData[0] = 1
		if err := g.SetCecCombinationalWitness(cex); err != nil {
			t.Fatal(err)
		}
		var status int32
		if engine == 2 {
			var pars Cec2_Par_t
			Cec2_SetDefaultParams(&pars)
			status = Cec2_ManPerformSweeping(g, &pars, nil)
		} else {
			var pars Cec3_Par_t
			Cec3_SetDefaultParams(&pars)
			status = Cec3_ManPerformSweeping(g, &pars, nil)
		}
		if status != 0 || g.SequentialCounterexample() == nil {
			t.Fatalf("engine %d lost existing witness result", engine)
		}
	}
}
