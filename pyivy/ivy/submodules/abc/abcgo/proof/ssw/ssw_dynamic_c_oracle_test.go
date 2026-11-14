package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// The compiled ABC fixture returns one latch and one input/output in both
// global and local simulation modes, with aggressive solver recycling.
func TestDynamicLatchCorrespondenceMatchesC(t *testing.T) {
	for _, local := range []int32{0, 1} {
		g := aig.NewGraph()
		x, a, b := g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(a, b))
		g.AddOutput(x)
		g.AddOutput(x)
		aig.Aig_ManSetRegNum(g, 2)
		pars := new(Ssw_Pars_t)
		Ssw_ManSetDefaultParamsLcorr(pars)
		pars.FDynamic = 1
		pars.FLocalSim = local
		pars.NSatVarMax2 = 1
		pars.NRecycleCalls2 = 0
		out := Ssw_LatchCorrespondence(g, pars)
		if out == nil || out.RegisterCount() != 1 || out.InputCount() != 2 || out.OutputCount() != 2 {
			t.Fatalf("local=%d unexpected interface", local)
		}
		for mask := 0; mask < 16; mask++ {
			inputs := make([]bool, 4)
			for i := range inputs {
				inputs[i] = mask&(1<<i) != 0
			}
			before, after := sswTrace(t, g, inputs), sswTrace(t, out, inputs)
			for i := range before {
				if before[i] != after[i] {
					t.Fatalf("local=%d trace=%v original=%v result=%v", local, inputs, before, after)
				}
			}
		}
	}
}
