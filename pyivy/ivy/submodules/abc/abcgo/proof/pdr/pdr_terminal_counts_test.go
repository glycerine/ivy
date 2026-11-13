package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// ABC's Saig_ManForEachPo uses nTruePos, even when physical COs differ.
func TestPdrUsesStoredTrueOutputCount(t *testing.T) {
	for _, incremental := range []bool{false, true} {
		g := aig.NewGraph()
		g.AddOutput(g.Const0())
		g.AddOutput(g.Const1())
		aig.Aig_ManSetRegNum(g, 0)
		g.SetTruePoCount(1)
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.fSolveAll, pars.fSilent, pars.fNotVerbose = 1, 1, 1
		var got int32
		if incremental {
			got = IPdr_ManSolve(g, &pars)
		} else {
			got = Pdr_ManSolve(g, &pars)
		}
		if got != 1 || pars.nProveOuts != 1 || pars.nFailOuts != 0 || vec.Vec_IntSize(pars.vOutMap) != 1 || vec.Vec_IntEntry(pars.vOutMap, 0) != 1 {
			t.Fatalf("incremental=%v result=%d proved=%d failed=%d map=%v", incremental, got, pars.nProveOuts, pars.nFailOuts, vec.Vec_IntArray(pars.vOutMap))
		}
	}
}
